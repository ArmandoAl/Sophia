# Sofía — Fase 2: prompts para Codex

Continúa desde `refactor_prompts_cursor.md` (pasos 1-8, ya completados y verificados).
Informe original: https://claude.ai/code/artifact/a07dc45c-7ed0-4803-a404-b5f1d7de3a69

Orden recomendado: **9 → 10 → 11 → 12 → 13**. El 9 es urgente y bloquea al 13.

---

## Contexto común (pégalo al inicio de cada sesión de Codex)

```
Trabajo en Sofía, un asistente personal con IA. Monorepo:
- Backend Go 1.24 + Firestore: sophia_ai_backend/sofia-backend
- Frontend Flutter: sophia_ai

Arquitectura del backend — respétala estrictamente:
- Capas por módulo: internal/<módulo>/{domain,application,infrastructure,interfaces/http}
- domain = tipos + reglas + interfaces de repositorio. Sin dependencias de infraestructura.
- application = casos de uso. Depende de interfaces de domain, nunca de Firestore.
- infrastructure = DOS implementaciones de cada repositorio, siempre:
    memory_*.go      (en memoria, la que usan los tests)
    firestore_*.go   (producción)
  Si añades un método a una interfaz de repositorio, implementa AMBAS o los tests no compilan.
- El cableado vive en internal/app/modules.go y internal/app/repositories.go
- La config por variables de entorno vive en internal/config/config.go — sigue ese patrón
  exacto (nombre, default, parseo, validación).

Reglas para todo lo que escribas:
- CERO dependencias nuevas salvo que el paso lo autorice explícitamente.
- Cambios quirúrgicos: toca solo lo que el paso pide. No refactorices de paso.
- Nada de abstracciones especulativas: ni interfaces con una sola implementación,
  ni factories, ni config para valores que no cambian.
- Comentarios en el mismo idioma y densidad que el código que rodea el cambio.
- Al terminar: `go vet ./... && go test ./...` y arregla lo que rompas.
- Dime qué archivos tocaste y qué NO hiciste.
```

---

## Paso 9 — Arreglar el doble conteo del worker

**URGENTE.** Corrompe los beliefs en silencio. El paso 13 lo empeora.

```
Hay un bug de idempotencia en internal/learning/worker/worker.go, en RunDaily.

El guard está bien puesto al principio (FindDailySummary → si existe, return nil), pero el
orden de escrituras deja una ventana de corrupción:

    applyDelta(...)                          ← muta los beliefs (EvidenceCount++, confianza)
    RecompilePromptBaseFromSummary(...)
    if version.TokenCount > promptTokenCeiling { return errors.New(...) }   ← sale AQUÍ
    CreateDailySummary(...)                  ← el marcador se escribe AL FINAL

Si algo falla entre applyDelta y CreateDailySummary, los beliefs ya cambiaron pero no hay
marcador. Cloud Scheduler reintenta, el guard no encuentra summary, y applyDelta corre por
segunda vez. Belief.Reinforce() hace EvidenceCount++ sin ninguna condición.

No es hipotético: la propia comprobación del techo de tokens sale DESPUÉS de mutar los
beliefs, así que un prompt que se pase del techo garantiza el doble conteo en cada reintento.

Impacto: beliefValue() usa log1p(EvidenceCount), así que la evidencia inflada asciende
beliefs al slot "core" por delante de otros mejor sustentados. Corrupción acumulativa y
silenciosa.

ARRÉGLALO reservando el día ANTES de mutar nada:

1. Divide la escritura del summary en dos fases:
   - Justo después de validar la salida del modelo y ANTES de applyDelta, crea el
     DailySummary con los stats y el coste de síntesis que ya conoces, y los campos de
     delta y prompt version vacíos. Usa el Create que ya existe: se apoya en el ID
     determinista DailySummaryID(userID, date) y en Firestore .Create(), que falla
     atómicamente si el documento ya existe. Ese fallo ES el lease: si otro proceso ya
     reservó el día, sal sin hacer nada.
   - Después corre applyDelta y la recompilación.
   - Al final, actualiza el summary con DeltaVsPrevious, PromptVersionBefore y
     PromptVersionAfter. Necesitarás un método Update en DailySummaryRepository:
     impleméntalo en las DOS implementaciones.

2. Trata el error "ya existe" de forma explícita. Si Firestore devuelve AlreadyExists,
   devuelve nil (no es un error: significa que el día ya se procesó). Define un
   ErrDailySummaryAlreadyExists en domain y mapéalo en el repositorio Firestore.

3. La comprobación del techo de tokens ya no debe abortar dejando estado a medias.
   Si el prompt recompilado se pasa del techo, registra un audit log y sigue: el techo
   ya se aplica dentro de RecompilePromptBase truncando por valor, así que superarlo es
   un síntoma a observar, no una razón para dejar el día a medio procesar.

4. Tests que deben pasar:
   - Un fallo simulado en la recompilación deja el summary creado, y un segundo RunDaily
     del mismo día NO vuelve a aplicar el delta (verifica EvidenceCount).
   - Dos RunDaily concurrentes del mismo usuario y día aplican el delta una sola vez.
   - El test TestRunDailySameDayIsIdempotent existente sigue pasando.
```

---

## Paso 10 — Solo DeepSeek, y un modelo por tarea

```
Dos cambios que van juntos: eliminar Gemini y hacer el modelo configurable por tarea.

PARTE A — Eliminar Gemini como proveedor.

Archivos con referencias (9):
  internal/ai/runtime/infrastructure/gemini/client.go  (+ client_test.go)
  cmd/tools/test-gemini-provider/main.go               (+ main_test.go)
  cmd/tools/run-ai-evals/main.go                       (+ main_test.go)
  internal/app/modules.go
  internal/config/config.go                            (+ config_test.go)

1. Borra el paquete internal/ai/runtime/infrastructure/gemini/ y el comando
   cmd/tools/test-gemini-provider/ completos.
2. Quita de config.go: GeminiAPIKey, GeminiModel, defaultGeminiModel y sus validaciones.
   El validador de AIModelProvider debe aceptar solo "fake" y "deepseek".
3. En modules.go, BuildModelClient se queda con dos ramas: "fake" y "deepseek".
4. En cmd/tools/run-ai-evals, quita la selección de proveedor Gemini.
5. NO toques la interfaz runtimedomain.ModelClient. Sigue teniendo dos implementaciones
   reales (deepseek y FakeModelClient) y es lo que permite que los tests corran sin red.

PARTE B — Un modelo por tarea.

Hoy config.go guarda un solo DeepSeekModel para todo. Pero ModelRequest ya tiene el campo
Task (internal/ai/runtime/domain/runtime.go:135) y ambos clientes ramifican sobre él.
Falta que la tarea elija también el modelo.

6. Sustituye DeepSeekModel por un mapa tarea → modelo, poblado desde el entorno:
     DEEPSEEK_MODEL           (default, obligatorio: se usa para cualquier tarea sin override)
     DEEPSEEK_MODEL_PLAN      (opcional)
     DEEPSEEK_MODEL_SYNTHESIZE(opcional)
     DEEPSEEK_MODEL_EXTRACT   (opcional)
   Si una variable específica está vacía, se cae al default. No inventes IDs de modelo:
   el default actual es el valor de DEEPSEEK_MODEL y todo lo demás es opt-in.

7. Define TaskPlan = "plan" junto al TaskSynthesize que ya existe, y haz que el runtime
   lo ponga explícitamente en el ModelRequest del camino caliente (hoy va vacío).

8. El cliente DeepSeek elige el modelo con request.Task al construir el payload, y sigue
   reportando en Usage el modelo realmente usado (ya lo hace).

9. Test: un ModelRequest con Task="synthesize" y DEEPSEEK_MODEL_SYNTHESIZE puesto usa ese
   modelo; sin la variable, cae al default. Verifica también que Usage.Model refleja el
   modelo real, no el default.

10. Actualiza docs/environment.md con las variables nuevas y quita las de Gemini.
```

---

## Paso 11 — Deduplicación semántica de beliefs

> **Decidido:** embeddings con Vertex AI. DeepSeek no ofrece API de embeddings, pero son un
> servicio aparte y mucho más barato que la generación, y ya estás en GCP. DeepSeek sigue
> siendo el único proveedor de **generación**.

```
UpsertBelief (internal/learning/application/usecases.go) deduplica con
HasHighTermOverlap, que es solapamiento léxico. Falla con paráfrasis: "junta",
"reunión", "meeting", "me junto con" no comparten términos. Cuando llegue la ingesta
de conversaciones (paso 13) el volumen de entrada se multiplica y esto creará
cuasi-duplicados que contaminan el prompt base.

Añade deduplicación semántica por embeddings.

1. Ya existe la interfaz internal/memory/application/EmbeddingProvider (EmbedText), hoy
   inyectada como nil en modules.go. Impleméntala contra Vertex AI text embeddings, en
   internal/ai/embeddings/infrastructure/vertex/. Usa la librería de Google Cloud que ya
   está en go.mod si sirve; si no, HTTP directo con la stdlib. Config nueva:
     EMBEDDINGS_ENABLED   (bool, default false)
     EMBEDDINGS_MODEL     (string)
     GOOGLE_CLOUD_PROJECT (ya deberías tenerlo para Firestore)

2. Añade Embedding []float32 a domain.Belief y persístelo en AMBOS repositorios.
   Fírestore guarda arrays de float64: convierte al leer y escribir.

3. En UpsertBelief, cuando el proveedor esté disponible:
   - calcula el embedding del statement entrante
   - trae candidatos con SearchByTerms como hoy (el filtro léxico sigue siendo el
     primer corte barato; no lo quites)
   - además, si el léxico no encuentra nada, trae los beliefs activos de la misma
     categoría (con límite) y compara por coseno
   - coseno >= 0.85 Y misma categoría → refuerza el existente en vez de crear
   - el umbral va en config: BELIEF_DEDUPE_THRESHOLD (default 0.85)

4. DEGRADACIÓN OBLIGATORIA: si EMBEDDINGS_ENABLED es false, si el proveedor falla, o si
   devuelve un vector vacío, UpsertBelief debe seguir funcionando exactamente como hoy
   (solo léxico). Nunca falles un upsert por un problema de embeddings. Registra un
   audit log cuando degrade.

5. Comando de backfill en cmd/tools/backfill-belief-embeddings/, siguiendo el patrón de
   cmd/tools/backfill-memory-search-terms/. Debe ser reanudable y procesar por lotes.

6. Tests con un EmbeddingProvider falso y determinista (no llames a la red):
   - dos frases parafraseadas con vectores cercanos → un solo belief, evidencia 2
   - dos frases de categorías distintas con vectores cercanos → dos beliefs
   - proveedor que devuelve error → comportamiento idéntico al de hoy, sin fallo
```

## Paso 12 — Contexto por persona y por modo

```
Sofía debe comportarse distinto según el contexto: no habla igual conmigo trabajando que
de chill, ni igual cuando estoy con mi pareja que con un cliente. Hoy todos los beliefs
son globales.

NO crees un módulo nuevo. Son dos campos en Belief más un catálogo pequeño.

1. En internal/learning/domain/belief.go añade:
     Scope    string   // "global" | "person" | "mode"
     ScopeKey string   // "" para global; "person:<slug>" o "mode:<slug>" si no
   Constantes ScopeGlobal/ScopePerson/ScopeMode y un IsValidScope(), al estilo de los
   IsValid* que ya hay. Un belief sin Scope se trata como global (retrocompatible con
   todo lo ya guardado: no hace falta migración).

2. Catálogo de contextos, en el mismo módulo learning:
     type UserContext struct {
       ID, UserID string
       Kind       string    // "person" | "mode"
       Slug       string    // "maria", "work", "pareja"
       Label      string    // "María", "Trabajo"
       Aliases    []string  // "mari", "mary" — para resolver menciones
       Active     bool
       CreatedAt, UpdatedAt time.Time
     }
   Repositorio con las dos implementaciones y CRUD mínimo: Create, List, FindBySlug,
   Update, Archive. Endpoints REST bajo /contexts siguiendo el patrón de /memory.

3. Resolución del contexto activo en el ContextBuilder. TRES fuentes, en este orden de
   prioridad, y de momento SOLO las dos primeras:
   a) explícita: el cliente manda active_context en el request (nuevo campo opcional
      en RuntimeRequest y en el body de POST /conversations/{id}/messages)
   b) estructural: si hay una actividad en curso cuyo título o tags mencionan el slug o
      un alias de un contexto, úsalo
   c) inferida por el modelo — NO LA IMPLEMENTES AHORA. Equivocarse de contexto es peor
      que no personalizar. Deja el hueco documentado y nada más.

4. Dónde entran los beliefs con scope — esto es lo importante y no lo cambies:
   - El PREFIJO CACHEABLE (BuildPromptPrefix) lleva SOLO los beliefs globales. Debe
     seguir siendo byte-idéntico durante todo el día, o se pierde el caché de DeepSeek.
   - El SUFIJO VARIABLE (BuildPromptSuffix) lleva el fragmento del contexto activo, en
     un campo nuevo "active_context" con el label y sus beliefs.
   Si metes los beliefs de contexto en el prefijo, cada cambio de contexto invalida la
   caché. No lo hagas.

5. RecompilePromptBase mantiene el techo de 1000 tokens SOLO para los globales. Añade un
   techo separado y más pequeño para el fragmento por contexto:
     const ContextFragmentTokenCeiling = 300
   Se calcula en tiempo de request desde los beliefs de ese ScopeKey, ordenados por
   beliefValue como los globales.

6. UpsertBelief debe deduplicar DENTRO del mismo (Scope, ScopeKey). "Le hablo formal" con
   scope person:cliente y con scope person:maria son beliefs distintos, no duplicados.

7. Índice nuevo en firestore.indexes.json:
   user_beliefs (user_id ASC, scope ASC, scope_key ASC, status ASC, confidence DESC)

8. El worker de síntesis debe poder asignar scope: amplía el schema de salida
   synthesisOutputSchema para que los beliefs "novel" puedan traer scope y scope_key, y
   valida que el scope_key exista en el catálogo antes de aplicarlo. Un scope_key
   desconocido se ignora y el belief se guarda como global.

9. Tests:
   - un belief con scope person:maria NO aparece en el prefijo cacheable
   - con active_context = person:maria, sí aparece en el sufijo
   - dos prefijos consecutivos con contextos distintos son byte-idénticos
   - el fragmento de contexto nunca supera ContextFragmentTokenCeiling
   - deduplicación: mismo enunciado con scope distinto crea dos beliefs
```

---

## Paso 13 — Ingesta de conversaciones externas

> Depende de 9, 11 y 12. **No lo lances antes.**

```
Quiero que Sofía me conozca desde el día uno, importando conversaciones externas
(WhatsApp, Instagram, Facebook). Esto choca con una regla del worker diario, y la forma
de resolverlo es que sea OTRO camino, no el mismo.

REGLA QUE NO SE TOCA: el worker de síntesis diaria (internal/learning/worker) sigue
leyendo SOLO decisiones. No le añadas conversaciones. Lo que construyes aquí es un
pipeline separado, con su propio job, su propio presupuesto y su propio nivel de confianza.

PARTE A — Niveles de confianza (hazlo primero, es la defensa).

1. En domain.Belief añade:
     TrustTier int   // 1 = decisiones, 2 = declaraciones explícitas, 3 = inferido de chats
   Constantes TrustDecision=1, TrustStated=2, TrustInferred=3. Los beliefs existentes sin
   valor se tratan como nivel 1 (vienen de decisiones reales).

2. Techo de confianza por nivel, aplicado dentro de Belief.Reinforce():
     nivel 1 → 1.0     nivel 2 → 0.8     nivel 3 → 0.5
   Un belief de nivel 3 NUNCA supera 0.5 por sí solo. Si una decisión real (nivel 1) lo
   corrobora, sube su TrustTier a 1 y el techo se levanta. Ese ascenso es el único camino.

3. Consecuencia buscada: como beliefValue() multiplica por la confianza, un belief sacado
   de un chat jamás desplaza del prompt base a uno sostenido por decisiones. Puede
   informar la búsqueda situacional, pero no gobierna el comportamiento hasta que lo
   confirmes con un acto. Escribe esto como comentario en el código.

PARTE B — El importador.

4. Nuevo módulo internal/ingestion/, con la estructura de capas de siempre.

5. Formato de entrada: un JSON normalizado, no el formato nativo de cada red. Define
     type RawConversation struct {
       Source      string    // "whatsapp" | "instagram" | "facebook" | "manual"
       ExternalID  string    // para idempotencia
       Participants []string
       Messages    []RawMessage  // {Author, Text, SentAt}
     }
   La conversión desde los export de cada red es responsabilidad de un script aparte que
   NO tienes que escribir ahora. Documenta el formato en docs/ingestion_format.md.

6. Endpoint POST /ingestion/conversations que acepta un lote y lo encola. Límite duro de
   tamaño por petición y validación estricta. Guarda los lotes en una colección
   ingestion_batches con status pending|processing|done|failed y contadores.

7. Worker cmd/workers/ingestion/, patrón de cmd/workers/synthesis:
   - procesa un lote por vez, con lease
   - trocea los mensajes en ventanas (p.ej. 50 mensajes) y por cada ventana hace UNA
     llamada al modelo con Task = "extract" (defínelo junto a TaskPlan y TaskSynthesize)
   - el prompt de extract pide SOLO hechos y preferencias observables, con salida
     estructurada validada por JSON-Schema:
       { "beliefs": [ {"statement","category","scope","scope_key","confidence"} ] }
   - usa el modelo de DEEPSEEK_MODEL_EXTRACT (paso 10): es extracción mecánica, va con
     el modelo barato
   - cada belief resultante entra por UpsertBelief con TrustTier = 3
   - registra el coste en el batch (tokens de entrada y salida por ventana)

8. CUARENTENA Y REVERSIBILIDAD, no negociable:
   - todo belief creado por ingesta guarda BatchID
   - endpoint DELETE /ingestion/batches/{id} que retira (status=retired) todos los
     beliefs de ese lote de una vez. Si una importación sale mal, se deshace entera.
   - los beliefs de nivel 3 arrancan con PromptSlot="situational", nunca "core"

9. Presupuesto: variable INGESTION_MAX_TOKENS_PER_BATCH. Si un lote lo supera, el worker
   para, marca el batch como failed con el motivo, y no sigue gastando.

10. Tests:
    - un lote de 120 mensajes se procesa en 3 ventanas
    - los beliefs creados tienen TrustTier 3 y no superan 0.5 de confianza
    - un belief de nivel 3 no entra al prompt base aunque tenga mucha evidencia
    - una decisión real que corrobora un belief nivel 3 lo asciende a nivel 1
    - DELETE del batch retira todos sus beliefs y ninguno más
    - reprocesar el mismo ExternalID no duplica nada
```

---

# Lo que tienes que hacer tú

Sigue pendiente de la fase 1 (nada de esto lo puede hacer Codex):

1. `firebase deploy --only firestore:indexes` — tras los pasos 12 y 13 hay índices nuevos
2. Correr `backfill-memory-search-terms` (y `backfill-belief-embeddings` si eliges 11-A)
3. Reglas de seguridad de Firestore para `user_beliefs`, `prompt_versions`,
   `daily_summaries`, y las nuevas `user_contexts` e `ingestion_batches`
4. Cloud Run Job + Cloud Scheduler para el worker de síntesis
5. Alerta de presupuesto en GCP antes de encender nada automático
6. Usar Sofía a diario: sin decisiones reales no hay nivel 1, y sin nivel 1 los beliefs
   importados nunca ascienden

Nuevo de esta fase:

7. **Habilitar la API de Vertex AI embeddings** en tu proyecto de GCP
8. Definir tus contextos iniciales (paso 12): qué personas y qué modos. Empieza con
   pocos y reales — 3 o 4. Un catálogo de 20 contextos que no usas es ruido
9. Exportar tus conversaciones de WhatsApp/Instagram y convertirlas al formato de
    `docs/ingestion_format.md`. El script de conversión es tuyo: depende del formato
    exacto que te dé cada app y de qué chats quieres incluir
10. **Revisar el aviso de privacidad.** Importar conversaciones significa procesar datos
    de terceros que no consintieron: las personas con las que hablaste. Aunque sea para
    uso propio, si esto va a TestFlight con más usuarios, cambia lo que tienes que declarar
