# Sofía — Fase 3: prompts para Codex

Continúa desde `refactor_prompts_codex_fase2.md` (pasos 9-13, completados y verificados).
Runbook de infraestructura: `runbook_despliegue.md`

Orden: **14 → 15 → 16**. El 14 es de seguridad operativa y va antes de desplegar a Cloud Run.

Usa el mismo bloque de **Contexto común** de la fase 2.

---

## Paso 14 — Arreglos menores y una trampa de despliegue

Cosas pequeñas encontradas al verificar las fases 1 y 2. Todas juntas en un commit.

```
Cinco arreglos menores. Ninguno cambia comportamiento salvo el primero, que evita un
fallo silencioso en producción.

1. FIRESTORE_DATABASE_ID sin valor por defecto es una trampa de despliegue.
   Hoy, si PERSISTENCE_DRIVER=firestore y FIRESTORE_DATABASE_ID está vacío, el código
   llama a firestore.NewClient(), que apunta a la base implícita "(default)". Este
   proyecto NO tiene esa base: tiene una base NOMBRADA "default". El resultado es que
   el binario arranca bien y falla en la primera consulta con
   "The database (default) does not exist".
   En local eso es un error visible. En un job de Cloud Run programado de madrugada, es
   un job que muere cada noche sin que nadie lo note.

   Arréglalo fallando al arrancar, en la validación de config.Load():
   si cfg.PersistenceDriver == "firestore" y cfg.FirestoreDatabaseID está vacío,
   devuelve un error claro que diga que FIRESTORE_DATABASE_ID es obligatorio y que la
   base implícita "(default)" no se asume.
   NO le pongas un valor por defecto: es específico de un proyecto y esconder el problema
   es justo lo que causó el fallo. Que sea explícito.

   Añade el test correspondiente en config_test.go.

2. internal/learning/domain/belief.go declara ErrDailySummaryExists como alias de
   ErrDailySummaryAlreadyExists. Nadie lo usa: bórralo.

3. Los errores de daily summary (ErrDailySummaryAlreadyExists y compañía) están
   declarados en belief.go. Muévelos a daily_summary.go, que es su sitio.

4. internal/memory/infrastructure/firestore_repository.go, SearchBasic (~línea 142)
   construye la query sin .Limit(), así que lee la colección entera del usuario y recorta
   después en memoria. Ya no está en el camino caliente (solo lo usa GET /memory/search),
   pero sigue siendo O(n) por llamada. Añade .Limit() a la query como ya hace
   SearchByTerms, respetando el filter.Limit y aplicando un tope por defecto si viene 0.

5. Revisa que docs/environment.md liste TODAS las variables que hoy lee config.go,
   incluida FIRESTORE_DATABASE_ID con una nota de que es obligatoria con Firestore.
   Compara contra config.go una por una; hay varias de la fase 2 que pueden faltar.
```

---

## Paso 15 — Poder ver lo que Sofía aprendió

**El hueco más importante que queda.** El sistema ya acumula un perfil psicológico del
usuario y no hay ninguna forma de verlo.

```
Sofía acumula creencias sobre el usuario, recompila un prompt base cada día y decide
autonomía a partir de ello. Pero no existe ni un endpoint para leer nada de eso: el
usuario no puede ver qué cree Sofía sobre él, ni yo puedo depurar por qué actuó.

Esto no es una feature de transparencia: una corrección directa del usuario sobre una
creencia equivocada vale más que veinte inferencias, así que es el canal de aprendizaje
más denso que puede tener el sistema. Y sin observabilidad, el worker nocturno es una
caja negra que gasta dinero.

Añade endpoints de LECTURA (y dos de escritura muy acotados) en el módulo learning,
siguiendo el patrón de internal/memory/interfaces/http/.

1. GET /learning/beliefs
   Lista las creencias del usuario autenticado. Filtros por query string:
     ?scope=global|person|mode   ?scope_key=...   ?category=...
     ?trust_tier=1|2|3           ?prompt_slot=core|situational
     ?status=active|superseded|retired  (default: active)
     ?limit= &cursor=
   Cada creencia devuelve: id, statement, category, scope, scope_key, confidence,
   decayed_confidence (calculada al momento, NO la cruda), evidence_count,
   contradiction_count, trust_tier, prompt_slot, first_observed_at, last_reinforced_at.
   Ordena por beliefValue descendente: lo que más gobierna el comportamiento va primero.

2. DELETE /learning/beliefs/{id}
   Retira una creencia (status = retired). NO la borra: el historial se conserva.
   Es cómo el usuario dice "eso no es cierto sobre mí".

3. PATCH /learning/beliefs/{id}
   Solo permite editar el campo statement. Al editarlo:
     - recalcula search_terms y el embedding
     - NO toca confidence ni evidence_count: corregir la redacción no es evidencia nueva
   Cualquier otro campo en el body se rechaza con 400. No abras la puerta a que el
   cliente escriba confidence a mano.

4. GET /learning/prompt-version
   Devuelve la PromptVersion activa: version, content, token_count, created_at,
   belief_ids y quality. Es lo que de verdad se le manda al modelo cada día.

5. GET /learning/summaries
   Lista los daily summaries, más recientes primero, con ?limit= &cursor=.
   Incluye stats, delta_vs_previous y synthesis_cost. Con esto se ve de un vistazo si el
   worker corrió, qué aprendió y cuánto costó.

6. GET /learning/summaries/{date}   (formato 2026-09-09)
   Un summary concreto, con todo el detalle.

7. Reglas transversales, no las saltes:
   - Todo va detrás de r.Authn y filtra SIEMPRE por el userID del token. Ningún endpoint
     acepta un user_id del cliente.
   - Ningún endpoint devuelve el campo embedding: son cientos de floats que no le sirven
     a nadie en la UI y engordan la respuesta.
   - Registra en el audit log las dos operaciones de escritura (retirada y edición).

8. Tests: filtros combinados, que un usuario no ve creencias de otro, que PATCH con un
   campo no permitido devuelve 400, y que retirar saca la creencia del prompt base en la
   siguiente recompilación.
```

---

## Paso 16 — Las pantallas que faltan en Flutter

Depende del paso 15.

```
El backend ya expone /contexts (paso 12) y /learning/* (paso 15), pero la app Flutter no
tiene pantalla para ninguno de los dos. Añádelas siguiendo la arquitectura existente:
lib/features/<feature>/{data,domain,presentation}, con Cubit para el estado, y el
ApiClient de lib/core/network. Respeta el estilo visual actual (mira
lib/core/widgets/action_proposal_card.dart como referencia de tono).

PANTALLA A — "Lo que Sofía sabe de ti"

1. Nueva feature lib/features/beliefs/. Lista las creencias de GET /learning/beliefs,
   agrupadas por categoría, ordenadas por influencia.
2. Cada creencia muestra: el enunciado, una barra o indicador de confianza, y de dónde
   salió (nivel de confianza: "porque lo decidiste", "porque lo dijiste", "deducido de
   tus conversaciones"). Traduce trust_tier a lenguaje humano, no muestres el número.
3. Marca visualmente cuáles están en el prompt base (prompt_slot = core): son las que
   de verdad gobiernan el comportamiento hoy.
4. Dos acciones por creencia:
     - "No es cierto"  → DELETE, con confirmación
     - "Corregir"      → editar el texto y PATCH
   Tras cualquiera de las dos, refresca la lista.
5. Filtro por contexto en la cabecera: Todas / Global / y una entrada por cada contexto
   que devuelva GET /contexts.

PANTALLA B — Contextos

6. Nueva feature lib/features/contexts/. CRUD contra /contexts: listar, crear, editar,
   archivar. Campos: kind (person|mode), slug, label, aliases.
7. Al crear, genera el slug automáticamente desde el label (minúsculas, sin acentos,
   guiones) pero deja editarlo.
8. Explica en la pantalla, en una línea, para qué sirve: que Sofía se comporte distinto
   según con quién estés o qué estés haciendo.

PANTALLA C — Estado del aprendizaje (dentro de la de diagnóstico que ya existe)

9. En lib/features/system/presentation/pages/diagnostics_screen.dart añade una sección
   con GET /learning/prompt-version y GET /learning/summaries?limit=7:
     - versión activa del prompt y su token_count
     - los últimos 7 días: fecha, propuestas aprobadas/corregidas/rechazadas, y coste
   Sirve para ver de un vistazo si el worker nocturno está corriendo.

10. Navegación: A y B cuelgan de la pantalla de ajustes que ya existe
    (lib/features/settings/presentation/pages/settings_page.dart).

11. Al terminar corre `fvm dart analyze lib` (NO `flutter analyze`: toma un lock global
    y se cuelga si hay otro build de Flutter corriendo en la máquina). Debe salir
    "No issues found!".
```

---

# Lo que tienes que hacer tú

## Ya completado

- [x] Credenciales ADC renovadas, quota project correcto, `.zshrc` limpio
- [x] Índices de Firestore desplegados
- [x] Backfills — nada que migrar, base verificada vacía
- [x] Backend probado contra Firestore real (usuario creado y confirmado)

## Pendiente, en orden

**1 · Commit y push** — sigue sin hacerse, y es lo único que te permite volver atrás.
```bash
cd ~/Documents/Sophia
git add -A && git commit -m "Fase 2 completa: idempotencia, DeepSeek, embeddings, contextos, ingesta"
git push
```

**2 · Reglas deny-all de Firestore** — paso 3 del runbook. Hazlo ahora que ya confirmaste
que el backend escribe bien: si algo se rompe después, sabes que fueron las reglas.

**3 · Habilitar Vertex AI** — paso 4 del runbook.
```bash
gcloud services enable aiplatform.googleapis.com --project=sophia-ai-60e87
```

**4 · Alerta de presupuesto** — paso 6 del runbook. **Antes** de encender nada automático.
Y pon también un límite de gasto en el panel de DeepSeek, que se factura aparte.

**5 · Worker de síntesis en Cloud Run** — paso 8 del runbook. Tres cosas que no puedes
saltarte ahí:
   - `FIRESTORE_DATABASE_ID=default` en las variables del job (sin esto muere cada noche)
   - Los secretos en Secret Manager, nunca en `--set-env-vars`
   - La service account del job necesita `roles/datastore.user`, `roles/aiplatform.user`
     y `roles/secretmanager.secretAccessor` — tus permisos no se heredan
   Ejecútalo **a mano** y comprueba que termina en `Succeeded` antes de programarlo.

**6 · Definir tus 3 o 4 contextos** — paso 9 del runbook. Más fácil desde la app cuando
esté el paso 16; por ahora con curl.

**7 · Usar Sofía a diario, 2-3 semanas.** No es relleno: sin decisiones tuyas reales no
hay creencias de nivel 1, y sin nivel 1 nada asciende ni se vuelve autónomo.

**8 · Limpiar los usuarios de prueba.** El registro creó 3 usuarios en la base. Decide
con cuál te quedas y borra los otros antes de empezar a acumular datos reales asociados.

**9 · La ingesta de conversaciones, al final.** Después de las 2-3 semanas de uso, y
después de pensar el tema de privacidad: importar tus chats procesa mensajes de gente que
no dio su consentimiento.
