# Sofía — Prompts de ejecución para Cursor

Informe completo: https://claude.ai/code/artifact/a07dc45c-7ed0-4803-a404-b5f1d7de3a69

8 pasos secuenciales. **No los reordenes**: el paso 5 sobre el 1 sin hacer produce colecciones vacías.
Cada prompt es autocontenido — cópialo entero. Entre paso y paso, verifica y commitea.

---

## Contexto común (pégalo al inicio de cada sesión de Cursor)

```
Trabajo en Sofía, un asistente personal con IA.

Repos:
- Backend Go 1.24 + Firestore: /Users/armandoalvarado/Documents/Sophia/sophia_ai_backend/sofia-backend
- Frontend Flutter: /Users/armandoalvarado/Documents/Sophia/sophia_ai

Arquitectura del backend — respétala estrictamente:
- Capas por módulo: internal/<módulo>/{domain,application,infrastructure,interfaces/http}
- domain = tipos + reglas + interfaces de repositorio. Sin dependencias de infraestructura.
- application = casos de uso. Depende de interfaces de domain, nunca de Firestore.
- infrastructure = DOS implementaciones de cada repositorio, siempre:
    memory_repository.go     (en memoria, la que usan los tests)
    firestore_repository.go  (producción)
  Si añades un método a una interfaz de repositorio, implementa AMBAS o los tests no compilan.
- El cableado vive en internal/app/modules.go y internal/app/repositories.go
- La configuración por variables de entorno vive en internal/config/config.go — sigue ese patrón
  exacto para cualquier flag nueva (nombre, default, parseo).

Reglas para todo lo que escribas:
- CERO dependencias nuevas. Solo stdlib y lo que ya está en go.mod / pubspec.yaml.
- Cambios quirúrgicos: toca solo lo que el paso pide. No refactorices de paso.
- No añadas abstracciones especulativas: nada de interfaces con una sola implementación,
  ni factories, ni config para valores que no cambian.
- Comentarios en el mismo idioma y densidad que el código que rodea el cambio.
- Al terminar, ejecuta `go vet ./... && go test ./...` y arregla lo que rompas.
- Dime explícitamente qué archivos tocaste y qué NO hiciste.
```

---

## Paso 1 — Desconectar el dry-run

**Esfuerzo:** 1 línea + tests. **Sin esto, ningún otro paso sirve para nada.**

```
En internal/conversations/application/service.go, el método SendMessage llama al runtime
con DryRun: true hardcodeado (línea ~93). Eso hace que HandleMessage salte el bloque
`if !request.DryRun` y NUNCA persista las propuestas en ai_action_proposals.
Resultado: el frontend recibe propuestas sin ID y su botón Confirm da 404.

Cambia DryRun a false para que las propuestas se persistan y tengan ID real.

Requisitos:
1. Verifica en internal/ai/runtime/application/runtime.go que con DryRun:false se llama a
   s.actions.CreateActionProposal y que el ID del proposal se propaga a
   domain.ActionProposalOutput.ID. Confirma que ese ID llega a la respuesta HTTP en
   internal/conversations/interfaces/http/responses.go — si no está mapeado, mapéalo.
2. CreateActionProposal puede fallar por validación de JSON-Schema o por política de
   autonomía (ErrAutonomyNotAllowed). Hoy HandleMessage hace `return nil, err` ante ese
   fallo, lo que tumbaría el mensaje entero del chat. Cámbialo para que una propuesta que
   falla al persistirse se cuente como rechazada (RejectedActionsCount++) y se omita,
   sin abortar la respuesta. El usuario debe seguir recibiendo su assistant_message.
3. Añade un test en internal/conversations que verifique que tras SendMessage las
   propuestas devueltas traen ID no vacío. Usa los repositorios en memoria.

No cambies nada más.
```

**Verifica tú:** `go test ./...` verde, y que el chat sigue respondiendo si una propuesta falla.

---

## Paso 2 — La señal de tres estados

**Esfuerzo:** ~1 día. Es la señal más valiosa de todo el sistema.

```
Hoy una propuesta de IA solo puede ser confirmed o rejected. Necesito distinguir tres
reacciones del usuario: aprobado directo, aprobado con corrección, y rechazado (con motivo).

BACKEND (sofia-backend):

1. En internal/ai/actions/domain/proposal.go añade a AIActionProposal:
     ConversationID    string          // qué conversación la generó
     Feedback          string          // approved_direct|approved_corrected|rejected|ignored|expired
     CorrectedInput    json.RawMessage // el JSON final tras la edición del usuario
     CorrectionDelta   []string        // nombres de los campos que cambiaron
     RejectionReason   string          // wrong_time|not_needed|wrong_person|other|""
     DecidedAt         *time.Time
     DecisionLatencyMS int64
   Añade las constantes Feedback* y un IsValidFeedback(), siguiendo el estilo de
   IsValidStatus/IsValidRiskLevel que ya están en ese archivo.

2. Confirm() acepta ahora un input opcional corregido:
     func (p *AIActionProposal) Confirm(correctedInput json.RawMessage) error
   - Si correctedInput es nil o json-igual a ProposedInput → Feedback = approved_direct
   - Si difiere → Feedback = approved_corrected, guarda CorrectedInput y calcula
     CorrectionDelta comparando las claves de primer nivel de ambos objetos JSON.
   - En ambos casos DecidedAt = now y DecisionLatencyMS = now - CreatedAt en milisegundos.
   Reject(reason string) hace lo análogo con Feedback = rejected.

3. La ejecución debe usar el input corregido cuando exista. En
   internal/ai/actions/application/usecases.go, ExecuteConfirmedActionProposal y los
   executors de registerDefaultExecutors deben leer CorrectedInput si no está vacío, y
   ProposedInput en caso contrario. Valida el input corregido contra el JSON-Schema de la
   tool ANTES de guardarlo — un input corregido inválido debe devolver
   ErrInvalidProposedInput, no ejecutarse.

4. Los handlers de confirm/reject en internal/ai/actions/interfaces/http/handler.go deben
   aceptar un body JSON opcional:
     POST .../confirm  { "corrected_input": {...}, "decision_latency_ms": 4200 }
     POST .../reject   { "rejection_reason": "wrong_time" }
   Body vacío o ausente debe seguir funcionando igual que hoy (retrocompatible).
   Si el cliente manda decision_latency_ms, úsalo; si no, calcúlalo del servidor.

5. Persiste todos los campos nuevos en AMBOS repositorios de
   internal/ai/actions/infrastructure/.

6. Tests: confirmación directa, confirmación con corrección (verifica CorrectionDelta),
   rechazo con motivo, y corrected_input que no valida contra el schema.

FRONTEND (sophia_ai):

7. En lib/core/widgets/action_proposal_card.dart, pasa de dos botones a tres:
   "No" (rechazar), "Ajustar" (editar), "Hacerlo" (confirmar).
   - "Ajustar" abre un diálogo con los campos del proposed_input editables como formulario
     y al guardar llama a confirm con corrected_input.
   - "No" abre un selector rápido de motivo (wrong_time / not_needed / wrong_person / otro)
     y llama a reject con rejection_reason.
   - Mide el tiempo desde que la tarjeta se muestra hasta que se pulsa un botón y envíalo
     como decision_latency_ms.
   Respeta el estilo visual existente de la tarjeta (glassmorphism, colores actuales).

8. Actualiza lib/features/actions/ (domain + data) para que confirm y reject acepten los
   parámetros nuevos.

9. Las propuestas deben poder confirmarse UNA A UNA, no en lote. Hoy
   lib/features/chat/presentation/pages/chat_page.dart itera todas las acciones con un
   solo botón; si Sofía propone 3 cosas y 2 son buenas, hay que rechazar las 3. Cámbialo a
   una fila de botones por acción.
```

**Verifica tú:** que el flujo confirmar→ejecutar sigue creando la actividad/recordatorio real.

---

## Paso 3 — Historial conversacional

**Esfuerzo:** ~2 h. Arregla que Sofía no entienda "mejor a las 4".

```
El runtime recibe un único mensaje aislado: RuntimeRequest solo tiene UserID, Message,
DryRun y RequestID. El modelo nunca ve los turnos anteriores, así que Sofía no puede
entender una corrección conversacional ("mejor a las 4", "no, el jueves").

1. En internal/ai/runtime/domain/runtime.go añade a RuntimeRequest:
     ConversationID string
     History        []Turn
   con  type Turn struct { Role string `json:"role"`; Content string `json:"content"` }

2. En internal/conversations/application/service.go, SendMessage debe cargar los últimos
   10 mensajes con s.messages.List antes de llamar a HandleMessage y pasarlos como History,
   en orden cronológico ascendente (el más antiguo primero), truncando cada Content a 500
   caracteres. Excluye el mensaje del usuario que se acaba de guardar: ya viaja en Message.

3. Propaga History al ModelRequest en HandleMessage y añádelo al payload que construyen
   AMBOS clientes:
     internal/ai/runtime/infrastructure/deepseek/client.go  (buildUserContent)
     internal/ai/runtime/infrastructure/gemini/client.go
   Va como campo "conversation_history" del JSON, justo antes de "user_message".

4. Contabiliza su coste: el historial debe pasar por el mismo presupuesto de tokens que
   el resto del contexto (mira fitItems y approximateTokens en
   internal/ai/runtime/application/runtime.go). Reserva 1250 tokens para historial y no
   lo dejes fuera del cómputo.

5. Test: dos mensajes seguidos en la misma conversación; el segundo ModelRequest debe
   incluir el primero en History.
```

---

## Paso 4 — Recall real, sin vector DB

**Esfuerzo:** ~1 día. Elimina el escaneo O(n) por mensaje.

```
La búsqueda de memoria está rota de dos formas y ambas hay que arreglarlas juntas:

(a) internal/ai/runtime/application/runtime.go:150 pasa el MENSAJE COMPLETO del usuario
    como query, y internal/memory/infrastructure/memory_repository.go:176 hace
    strings.Contains del título contra esa frase entera. Nunca coincide.
(b) internal/memory/infrastructure/firestore_repository.go:145 (SearchBasic) construye la
    query SIN .Limit(), así que lee la colección entera del usuario en cada mensaje, y
    solo después filtra. Coste O(n) para recall cero.

Arréglalo indexando en escritura en lugar de escanear en lectura. Sin embeddings, sin
dependencias nuevas.

1. Crea internal/memory/domain/search.go con:
     func ExtractTerms(text string, max int) []string
   Minúsculas, split por no-alfanumérico (usa strings.FieldsFunc + unicode), descarta
   términos de menos de 4 runas y stopwords en español e inglés, deduplica preservando
   orden, corta en max. Define la lista de stopwords como una var package-level (map para
   lookup O(1)); incluye las comunes del español: para, pero, como, este, esta, esto,
   cuando, donde, porque, tengo, quiero, puedes, sobre, desde, hasta, entre...

2. Añade el campo SearchTerms []string a domain.Memory y a las structs de persistencia de
   AMBOS repositorios (tag firestore:"search_terms"). Se calcula en NewMemory y en Apply
   como ExtractTerms(Title + " " + Summary + " " + Content + " " + strings.Join(Tags," "), 40).
   Nunca lo calcules en lectura.

3. Añade a MemoryRepository:
     SearchByTerms(ctx context.Context, userID string, terms []string, limit int) ([]*Memory, error)
   Implementación Firestore: where user_id ==, status == active,
   search_terms array-contains-any terms (corta terms a 30, que es el límite de Firestore),
   .Limit(limit*4); después puntúa cada resultado por el número de términos que coinciden
   y devuelve los `limit` mejores. Implementación en memoria: mismo comportamiento.

4. En el ContextBuilder (runtime.go), sustituye la llamada a SearchMemory por
   ExtractTerms(message, 12) + SearchByTerms. Si ExtractTerms devuelve vacío, no busques.

5. Arregla dos bugs adyacentes en el mismo archivo:
   - redactText (~línea 333) devuelve la cadena literal "[REDACTED]" para CUALQUIER texto
     que contenga "@", destruyendo el recuerdo entero. Cámbialo para redactar solo el email
     con una regexp ([\w.+-]+@[\w-]+\.[\w.-]+ → "[email]") y conservar el resto del texto.
   - fitItems reparte el presupuesto por orden de llegada, así que actividades y
     recordatorios lo agotan y las memorias llegan vacías. Sustitúyelo por un reparto
     explícito por cuotas: beliefs/memorias 40%, historial 25%, actividades 25%,
     recordatorios 10%; la sobra de una sección pasa a la siguiente.
   - InsightsSummary se añade fuera del presupuesto (~línea 143), así que UsedApproxTokens
     miente. Hazlo pasar por el mismo cómputo.

6. Añade a firestore.indexes.json el índice compuesto que requiere la nueva query.

7. Añade una migración en cmd/tools/ que recorra las memorias existentes y les rellene
   search_terms. Sigue el patrón de cmd/tools/migrate-firestore-legacy/main.go.

8. Test: crea memorias, busca con una frase larga en lenguaje natural que contenga un par
   de sus palabras clave, verifica que las recupera (hoy devolvería cero).
```

**Verifica tú:** ejecuta la migración antes de desplegar, o las memorias viejas serán invisibles.

---

## Paso 5 — Dónde vive el entendimiento

**Esfuerzo:** ~2 días.

```
Necesito dos colecciones nuevas que son la base del aprendizaje a largo plazo. Crea un
módulo nuevo internal/learning/ siguiendo exactamente la estructura de internal/memory/.

1. internal/learning/domain/belief.go — un "belief" es una creencia destilada sobre el
   usuario, la unidad atómica del entendimiento:

     type Belief struct {
       ID, UserID           string
       Statement            string    // "Prefiere agendar reuniones antes de las 11am"
       Category             string    // schedule|communication|priorities|work_style|personal|constraint
       Confidence           float64   // 0..1
       EvidenceCount        int
       ContradictionCount   int
       FirstObservedAt      time.Time
       LastReinforcedAt     time.Time
       LastContradictedAt   *time.Time
       Status               string    // active|superseded|retired
       SupersededBy         string    // encadenado, NUNCA se borra
       SupersedeReason      string
       PromptSlot           string    // core|situational
       TokenCost            int
       SearchTerms          []string  // reutiliza ExtractTerms del paso 4
     }

   Métodos en domain, con la lógica de evolución:
     Reinforce()          → EvidenceCount++, Confidence sube hacia 1 asintóticamente,
                            LastReinforcedAt = now
     Contradict()         → ContradictionCount++, Confidence baja con peso 2.5x el del
                            refuerzo (cambiar de opinión debe ser barato)
     Supersede(newID, reason)
     DecayedConfidence(now) → Confidence * pow(0.5, díasDesdeLastReinforcedAt/90)
                            Media vida de 90 días: un hábito no reforzado deja de gobernar
                            el comportamiento por sí solo.

   TokenCost se calcula al crear/actualizar, con la misma heurística que approximateTokens.

2. internal/learning/domain/prompt_version.go — el prompt base como dato versionado:

     type PromptVersion struct {
       ID, UserID  string
       Version     int
       Content     string
       TokenCount  int
       BeliefIDs   []string   // auditoría: qué beliefs lo compusieron
       GeneratedFromSummaryID string
       CreatedAt   time.Time
       Active      bool       // EXACTAMENTE uno activo por usuario
       Quality     PromptQuality  // approval_rate, correction_rate, sample_size
     }

3. internal/learning/domain/daily_summary.go — append-only, un doc por usuario y día:
   campos Date (string "2006-01-02"), GeneratedAt, Observations, DeltaVsPrevious
   {Reinforced, Contradicted, Novel}, Stats {ProposalsTotal, ApprovedDirect,
   ApprovedCorrected, Rejected, Ignored, MedianDecisionSeconds}, SynthesisCost
   {InputTokens, OutputTokens, Model, LatencyMS}, PromptVersionBefore/After.
   NUNCA se sobrescribe: el historial completo es el activo.

4. Interfaces de repositorio en domain, con AMBAS implementaciones en
   internal/learning/infrastructure/ (memory_ y firestore_). Métodos mínimos:
   - BeliefRepository: Create, Update, FindByID, ListActive, SearchByTerms, SetPromptSlot
   - PromptVersionRepository: CreateActive (desactiva la anterior en la MISMA transacción
     Firestore — usa RunTransaction, no dos escrituras sueltas), GetActive, ListVersions
   - DailySummaryRepository: Create, FindByDate, ListRecent

5. Casos de uso en internal/learning/application/ con la garantía anti-duplicados:
   UpsertBelief(userID, statement, category) busca primero por SearchByTerms un belief
   activo con solapamiento alto de términos; si lo encuentra lo refuerza en lugar de crear
   uno nuevo. Esto es lo que impide que "prefiere reuniones por la mañana" se guarde 40
   veces y contamine el prompt.

6. Cablea todo en internal/app/repositories.go y modules.go.

7. Añade los índices a firestore.indexes.json:
   user_beliefs (user_id ASC, status ASC, confidence DESC)
   daily_summaries (user_id ASC, date DESC)
   ai_action_proposals (user_id ASC, created_at ASC)
   ai_action_proposals (user_id ASC, tool_name ASC, feedback ASC)

8. Tests de la lógica de evolución: refuerzo sube la confianza, contradicción la baja más
   rápido, el decaimiento reduce la confianza de un belief viejo, y UpsertBelief no crea
   duplicados para el mismo enunciado en otras palabras.

NO escribas todavía el worker ni la síntesis. Solo el esquema y los repositorios.
```

---

## Paso 6 — El corazón: síntesis diaria

**Esfuerzo:** ~3 días. Es el mecanismo que creías tener.

```
Ahora el job diario que hace evolucionar a Sofía. Crea cmd/workers/synthesis/main.go
siguiendo el patrón de cmd/workers/reminders/main.go (config, señales de apagado,
lease de worker, logger, flag de habilitación por env var).

REGLA INNEGOCIABLE: el summarizer NO lee conversaciones crudas. Lee DECISIONES.
Una conversación de 40 turnos sobre qué cenar produce cero señal si no hubo propuesta;
una sola corrección de horario produce señal fuerte. Esto es lo que impide que el ruido
diario contamine la síntesis. Si en algún momento te tienta pasarle los mensajes al
modelo, no lo hagas.

1. internal/learning/worker/worker.go con RunDaily(ctx, userID, day):
   a) Carga las propuestas del día vía un método nuevo
      ListByDateRange(ctx, userID, from, to) en el repositorio de ai_action_proposals
      (impleméntalo en memoria y en Firestore).
   b) Si no hay ninguna propuesta, retorna sin gastar un token. Días silenciosos son gratis.
   c) Extrae las señales: correcciones (con su CorrectionDelta), rechazos (con motivo),
      y aprobaciones directas agrupadas por tool_name.
   d) Carga los beliefs activos: son el "histórico previo" contra el que contrastar.
   e) UNA sola llamada al modelo, con salida estructurada obligatoria:
        { "reinforced":   [{"belief_id","evidence","confidence_delta"}],
          "contradicted": [{"belief_id","evidence","note"}],
          "novel":        [{"statement","category","confidence"}] }
      Reutiliza el runtimedomain.ModelClient que ya existe; no crees un cliente nuevo.
      Valida la salida con internal/platform/jsonschema antes de aplicarla.
   f) Aplica el delta: Reinforce, Contradict, y UpsertBelief para los novel.
      Si un belief queda contradicho y existe uno nuevo que lo reemplaza, encadénalos
      con Supersede en lugar de borrar.
   g) Persiste el DailySummary con su SynthesisCost real (tokens de entrada y salida
      leídos de la respuesta del proveedor — si el cliente actual no los expone, añádelo).
   h) Llama a RecompilePromptBase.

2. RecompilePromptBase(ctx, userID) — el punto clave de toda la arquitectura:

   const promptTokenCeiling = 1000   // techo DURO

   El prompt base NO se acumula ni se edita: se RECONSTRUYE desde cero a partir de los
   beliefs activos, cada día. Eso hace estructuralmente imposible que crezca sin límite.
   - Ordena los beliefs por valor:
       confidence * log1p(evidence_count) / (1 + contradiction_count)
     usando DecayedConfidence, no la confianza cruda.
   - Rellena hasta el techo. Lo que cabe se marca PromptSlot="core"; lo que no cabe NO se
     pierde: se marca "situational" y se recuperará por búsqueda cuando sea relevante.
   - Crea una PromptVersion nueva con CreateActive.

3. Carga del prompt base en el runtime: el ContextBuilder debe leer la PromptVersion
   activa del usuario e incluirla en el ModelRequest como campo propio (no la mezcles con
   el contexto dinámico: en el paso 8 tiene que quedar en la parte cacheable). Si el
   usuario no tiene ninguna versión activa todavía, el sistema debe funcionar igual que
   hoy — arranque en frío sin errores.

4. Cuarentena antirrigidez: si la tasa de aprobación de una tool cae más de 15 puntos en
   7 días, registra un audit log y baja su autonomía efectiva a "suggestive". Sirve para
   detectar que el usuario cambió de hábitos antes de acumular 30 rechazos.

5. Config en internal/config/config.go, con el mismo patrón que las de reminders:
   SYNTHESIS_WORKER_ENABLED, SYNTHESIS_WORKER_ID, SYNTHESIS_WORKER_LEASE,
   SYNTHESIS_RUN_HOUR_LOCAL (hora local del usuario a la que cierra el día).
   El corte del día debe usar la timezone del perfil del usuario, no UTC.

6. Añade el target de build al Dockerfile: hoy solo compila ./cmd/sofia. Necesito poder
   construir también ./cmd/workers/synthesis.

7. Tests con un ModelClient fake (ya existe runtimeinfra.NewFakeModelClient):
   - día sin propuestas → cero llamadas al modelo
   - refuerzo sube confianza y el belief sigue en core
   - contradicción repetida saca al belief del prompt base
   - el prompt recompilado NUNCA supera promptTokenCeiling
   - dos ejecuciones del mismo día son idempotentes (no duplican el summary)
```

**Verifica tú:** el punto 7, último test. La idempotencia importa porque Cloud Scheduler reintenta.

---

## Paso 7 — Probabilidad y autonomía

**Esfuerzo:** ~1 día.

```
Ahora el motor que decide cuándo Sofía actúa sin preguntar. Heurística explicable,
NO machine learning: con un usuario y ~1000 decisiones al año cualquier modelo entrenado
sobreajusta, y la explicabilidad vale más que la precisión — Sofía tiene que poder decir
"lo hice porque las últimas 36 de 40 veces lo aprobaste".

1. Crea internal/learning/domain/scoring.go:

   type ToolHistory struct {
     Total, ApprovedDirect, ApprovedCorrected, Rejected int
     DaysSinceLastSample float64
   }

   func PredictApproval(h ToolHistory, beliefs []ScoredBelief) (float64, []string)
     - Base con suavizado de Laplace: (ApprovedDirect + 1) / (Total + 2)
       para que 1-de-1 no dé 100%.
     - Las correcciones cuentan como medio éxito: la intención era buena.
       base += 0.5 * ApprovedCorrected / (Total + 2)
     - Beliefs que apoyan: +0.10 * confidence. Que contradicen: -0.25 * confidence.
       La asimetría es intencional.
     - Decaimiento por recencia, media vida 60 días, empujando hacia 0.5 (incertidumbre):
       base = 0.5 + (base - 0.5) * pow(0.5, DaysSinceLastSample/60)
     - Clamp a [0,1]. Devuelve también los IDs de los beliefs usados, para auditoría.

   func ShouldAutoExecute(p float64, h ToolHistory, reversible bool, threshold float64) bool
     return p >= threshold && h.Total >= 10 && reversible
     Tres condiciones, no una: umbral + volumen mínimo de evidencia + reversibilidad.

2. Marca la reversibilidad en las tools. En internal/tools/domain añade Reversible bool a
   ToolDefinition y ponlo en el seed de SeedDefaultTools: crear actividad/recordatorio/
   memoria son reversibles; borrar y cualquier cosa que salga al exterior, no.
   Una acción irreversible NUNCA se autoejecuta, por muy alta que sea la probabilidad.

3. Persiste el scoring para poder auditarlo. Al crear una propuesta guarda en
   ai_action_proposals: predicted_approval, prediction_basis []string,
   prediction_model_version ("heuristic_v1"), autonomy_mode_used
   ("proposed"|"auto_executed"). Sin esto no puedes depurar por qué Sofía hizo algo.

4. Conecta el motor al runtime. Hoy SafetyPolicy.CanExecute en
   internal/ai/runtime/application/runtime.go devuelve ErrUnsafeActionProposal siempre
   e incondicionalmente. Sustitúyelo por:
     - calcula PredictApproval para la tool y el usuario
     - si ShouldAutoExecute y el AutonomyLevel del usuario es semi_autonomous → permite
     - en cualquier otro caso → propone, como hoy
   El umbral por defecto es 0.85; hazlo configurable con AUTONOMY_THRESHOLD en config.go.

5. Limpia el código muerto que confunde: autonomyAllows e isPlanningTool en runtime.go
   (~líneas 395 y 421) están declaradas y no se llaman desde ningún sitio. O las usas en
   esta lógica o las borras.

6. Un endpoint para que el usuario vea el razonamiento:
   GET /ai/action-proposals/{id}/explanation → devuelve predicted_approval, los beliefs
   que la sustentaron y el histórico de la tool. Sin esto, la autonomía es una caja negra.

7. Tests obligatorios — estos son los que impiden que Sofía haga tonterías:
   - arranque en frío (Total=0) → probabilidad cercana a 0.5 y NUNCA autoejecuta
   - histórico fuerte y reciente (36/40, 1 día) → >0.85 y sí autoejecuta
   - acción irreversible → nunca autoejecuta, ni con probabilidad 0.99
   - un belief contradictorio fuerte baja del umbral
   - histórico de 8 meses sin uso → vuelve a la incertidumbre (antirrigidez)
```

---

## Paso 8 — Prefijo cacheable

**Esfuerzo:** ~1 día. Recorta ~45 % del coste variable.

```
El prompt base sintetizado es idéntico en todos los requests del mismo día (solo cambia
cuando el worker lo recompila de madrugada), así que es el candidato perfecto para prompt
caching. Hoy no se puede cachear nada porque buildUserContent() mete instrucciones, tools,
prompt base y contexto dinámico en un único blob JSON.

1. Elimina la duplicación primero: el systemPrompt() de
   internal/ai/runtime/infrastructure/deepseek/client.go:148 está copiado literal en
   gemini/client.go:153. Extrae las instrucciones estáticas a un solo sitio
   (internal/ai/runtime/domain/) y que ambos clientes lo consuman. Con el prompt base
   dinámico, esta duplicación garantizaría divergencia entre proveedores.

2. Reestructura el payload de AMBOS clientes en dos bloques estrictamente separados:

   PREFIJO ESTABLE (byte-idéntico durante todo el día → cacheable):
     - instrucciones de sistema
     - prompt base del usuario (la PromptVersion activa del paso 6)
     - definiciones de las tools disponibles

   SUFIJO VARIABLE (cambia en cada request):
     - historial conversacional
     - beliefs situacionales recuperados
     - estado (actividades, recordatorios)
     - mensaje del usuario

   El prefijo va PRIMERO y su serialización tiene que ser determinista: ordena las tools
   por nombre y usa un orden de claves estable. Si un solo byte cambia entre requests,
   el caché no acierta y todo esto no sirve de nada.

3. Activa el caching del proveedor sobre ese prefijo. Consulta la documentación vigente del
   proveedor configurado en AI_MODEL_PROVIDER antes de implementarlo — el mecanismo difiere
   (unos cachean automáticamente por prefijo, otros necesitan un marcador explícito).
   Dime cuál aplicaste y qué requiere.

4. Instrumenta el coste real, que hoy no se mide en ninguna parte. Los clientes ya parsean
   la respuesta del proveedor; extrae de ahí los tokens de entrada, de salida y los
   cacheados, y regístralos en el audit log de "ai_runtime_proposals_generated" junto a la
   latencia que ya se guarda. Sin esto no puedes saber si el caché funciona.

5. Haz configurable el presupuesto de contexto: DefaultTokenBudget está hardcodeado a 1200
   en internal/ai/runtime/application/runtime.go:26. Muévelo a config.go como
   CONTEXT_TOKEN_BUDGET con default 4600, que es el objetivo del plan.

6. Test: dos requests seguidos del mismo usuario producen prefijos byte-idénticos.
   Compáralos con un hash en el test — es la única forma de que esto no se rompa en
   silencio con un cambio futuro.
```

---

# Lo que tienes que hacer tú

Cursor no puede (o no debe) hacer esto:

### Antes de empezar

1. **Rama en cada repo.** Son dos repos separados y la raíz `Sophia/` no está bajo git.
   ```
   cd sophia_ai_backend/sofia-backend && git checkout -b feat/learning-loop
   cd sophia_ai && git checkout -b feat/learning-loop
   ```
   Commitea después de cada paso, no al final. Si el paso 6 sale mal, quieres poder volver al 5.

2. **Decide tres números que son producto, no código.** Cursor pondrá defaults; confírmalos o cámbialos:
   - **Umbral de autonomía** (`AUTONOMY_THRESHOLD`, propuesto 0.85). Cuánto riesgo aceptas de que Sofía haga algo que no querías.
   - **Retención de mensajes crudos** (propuesto 90 días). Ojo: interactúa con tu flujo de privacidad y borrado que ya existe en `internal/privacy/`.
   - **Hora de corte del día** (`SYNTHESIS_RUN_HOUR_LOCAL`). Si trabajas de noche, cortar a medianoche parte tu día en dos.

### Durante

3. **Verifica tras cada paso**, no confíes en el "listo" de Cursor:
   ```
   cd sophia_ai_backend/sofia-backend && go vet ./... && go test ./...
   ```

4. **Ejecuta la migración del paso 4** antes de desplegar, o las memorias existentes quedan invisibles (no tendrán `search_terms`).

### Infraestructura (solo tú tienes las credenciales)

5. **Desplegar índices de Firestore** — después de los pasos 4 y 5, y **antes** de que el código nuevo corra en producción, o las queries fallarán:
   ```
   firebase deploy --only firestore:indexes
   ```

6. **Reglas de seguridad de Firestore para las colecciones nuevas.** `firebase.json` no declara un archivo de `rules`. Comprueba cómo estás protegiendo `memories` y `ai_action_proposals` hoy y replícalo para `user_beliefs`, `prompt_versions` y `daily_summaries`. Son los datos más sensibles del sistema: es el perfil psicológico del usuario.

7. **Cloud Run Job + Cloud Scheduler para el worker de síntesis.** El paso 6 añade el target al Dockerfile, pero desplegar la imagen y programar la ejecución diaria es tuyo. Configúralo con reintentos — por eso el test de idempotencia del paso 6 importa.

8. **Alerta de presupuesto en GCP**, antes de encender el worker. El job diario consume tokens sin que nadie lo esté mirando; un bug de bucle se nota en la factura, no en la UI.

### Producto

9. **Usa Sofía a diario.** Esto no es un chiste ni relleno: el bucle no tiene nada que aprender sin decisiones reales tuyas. Necesitas ~10 propuestas por herramienta antes de que `ShouldAutoExecute` permita nada. Son unas 2-3 semanas de uso real. Ningún dato sintético sirve aquí — el sistema aprende de *tus* correcciones, y esas no se pueden falsear.

10. **Decide qué haces con `sophia_db/db.html`.** Describe un esquema que no existe (`ai_memory`, `ai_actions_log`, `tool_calls`, `bodies`, `job_queue`, `embedding_vector`). Es peligroso porque *parece* la fuente de verdad. Actualízalo al esquema real tras el paso 5, o bórralo. Mantenerlo desactualizado es la peor de las tres opciones.

11. **Revisa el aviso de privacidad.** A partir del paso 5, el sistema guarda creencias inferidas sobre el usuario, no solo datos que él escribió. Si esto va a TestFlight con usuarios reales, eso cambia lo que tienes que declarar. Y la pantalla "Lo que Sofía sabe de ti" del paso 2 deja de ser una feature bonita y pasa a ser el mecanismo de rectificación.
