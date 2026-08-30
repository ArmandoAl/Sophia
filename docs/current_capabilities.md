# Current Capabilities

Fecha: 2026-07-03

## Que Puede Hacer Sofia Hoy

Sofía ya puede registrar usuarios, autenticar con JWT y mantener un perfil de usuario con configuracion inicial de la asistente: nombre, idioma, estilo de comunicacion, niveles de autonomia y switches para memoria, reminders y planning.

Puede gestionar activities como unidad amplia de planificacion: tareas, habitos, rutinas, objetivos, eventos y checklists. Las activities son privadas por usuario y pueden completarse, cancelarse o archivarse.

Puede crear y listar reminders. Tambien puede consultar reminders due, asociarlos a activities y cancelarlos cuando una activity se cancela o archiva. Ya existe recurrencia simple para reminders: `none`, `daily`, `weekly` y `monthly`.

Existe un reminder worker ejecutable por separado. Puede buscar reminders due, reclamarlos con lease, procesarlos de forma idempotente y usar delivery `noop` o FCM. Con Firestore, el claim transaccional permite correr multiples instancias evitando procesamiento paralelo del mismo reminder.

Sofía puede registrar device tokens para notificaciones y tiene FCM como primer provider real opcional. Por seguridad, FCM no envia notificaciones reales salvo configuracion explicita.

Puede registrar insights personales: mood entries, outcomes de activities, reflections y summary basico. Esto es tracking personal/productivo, no diagnostico clinico.

Puede guardar memoria textual explicita, buscarla por texto/filtros y aplicar soft delete. El ContextBuilder solo debe usar memorias activas con visibilidad `assistant_context`.

Puede exportar datos propios del usuario, crear delete requests y procesar borrado fisico controlado mediante comando interno solo para requests aprobadas. Audit logs se conservan por default.

Puede listar tools internas, crear AI action proposals y ejecutar explicitamente algunas acciones confirmadas. El AI Runtime puede usar provider `fake` o Gemini, pero siempre en modo proposal-only.

## Que Esta Apagado Por Seguridad

- Worker de reminders: `REMINDER_WORKER_ENABLED=false` por default.
- Delivery real FCM: `REMINDER_DELIVERY_PROVIDER=noop`, `FCM_ENABLED=false`, `FCM_DRY_RUN=true` por default.
- Ejecucion automatica de IA: no existe en runtime.
- Gemini: no es default; `AI_MODEL_PROVIDER=fake` es el modo seguro.
- Evals Gemini reales: se saltan si no hay `GEMINI_API_KEY`.
- Borrado fisico: no ocurre al crear `/users/me/delete-request`; requiere aprobacion y comando interno.
- Migracion legacy: no aplica cambios destructivos por default.

## Que Requiere Configuracion

Firestore real requiere:

- `PERSISTENCE_DRIVER=firestore`
- `FIRESTORE_PROJECT_ID`
- `FIRESTORE_DATABASE_ID` si se usa database nombrada
- `GOOGLE_APPLICATION_CREDENTIALS` para desarrollo local cuando aplique
- `FIRESTORE_EMULATOR_HOST` para emulator

FCM requiere:

- `REMINDER_DELIVERY_PROVIDER=fcm`
- `FCM_ENABLED=true`
- `FCM_DRY_RUN=false` solo despues de pruebas
- `FCM_PROJECT_ID`
- credenciales via ADC

Gemini requiere:

- `AI_MODEL_PROVIDER=gemini`
- `GEMINI_API_KEY`
- `GEMINI_MODEL`
- `AI_RUNTIME_ENABLED=true`
- `AI_RUNTIME_PROPOSAL_ONLY=true`

El worker requiere:

- `REMINDER_WORKER_ENABLED=true`
- `REMINDER_WORKER_ID`
- `REMINDER_WORKER_INTERVAL`
- `REMINDER_WORKER_BATCH_SIZE`
- `REMINDER_WORKER_LEASE_DURATION`

## Que Falta Para Produccion

- Validacion operacional del worker multi-instancia con Firestore emulator/produccion.
- Rollout real de FCM con preferencias de usuario, limpieza de tokens invalidos y auditoria de intentos.
- Evals reales con Gemini API key y criterio de calidad/costo antes de uso prolongado.
- Rate limiting y hardening adicional de auth.
- Recurrencia compleja tipo RRULE, excepciones, DST y validacion IANA fuerte.
- Paginacion por cursor consistente en todos los listados grandes.
- Politica final de privacidad: borrado de auth user, retencion de audit logs y export firmado.
- RAG semantico con embeddings y controles de privacidad.
- Migracion legacy revisada/aprobada sin sobrescribir datos.
- Observabilidad externa para runtime, worker y delivery.
# Update 2026-07-16

Conversation persistence uses isolated V1 Firestore namespaces. Privacy export and approved delete execution cover only those namespaces.
# Release Gate V1.4: local memory E2E passes for auth, conversations/history and reminders. Firestore/Gemini real validation remains externally blocked.
