# Architecture State

Fecha: 2026-07-03

## Estado consolidado

Sofia AI Backend tiene una arquitectura modular vertical basada en bounded contexts. El backend es usable para auth, perfil de usuario, activities, reminders persistidos con recurrencia simple, reminder worker, device tokens/FCM opcional, insights, memory textual, privacy controls y propuestas de accion controladas. El AI Runtime existe con providers `fake` y Gemini en modo proposal-only; todavia no esta listo para autonomia ni produccion prolongada sin evals reales con API key.

## Estructura viva

```text
cmd/sofia
internal/config
internal/app
internal/database
internal/platform/httpjson
internal/server
internal/auth
internal/users
internal/activities
internal/reminders
internal/insights
internal/memory
internal/conversations
internal/notifications
internal/privacy
internal/tools
internal/ai/actions
internal/ai/evals
```

Cada modulo principal sigue el patron:

```text
domain
application
infrastructure
interfaces/http
```

## Bounded contexts

| Contexto | Modulo | Responsabilidad | Estado |
| --- | --- | --- | --- |
| Identity | `auth` | Registro, login, credenciales, JWT | usable |
| User Model | `users` | Profile, onboarding, AISettings | usable |
| Planning | `activities` | Activity como unidad superior | usable |
| Scheduling | `reminders` | Recordatorios persistidos, due query, recurrencia simple y worker runtime | parcial usable |
| Personal Tracking | `insights` | Mood, outcomes, reflections | usable |
| Memory | `memory` | Memoria explicita textual | usable textual |
| Conversations | `conversations` | Historial canonico V1: conversaciones y mensajes persistidos | usable |
| Notifications | `notifications` | Device tokens y delivery FCM opcional para reminders | opcional/deshabilitado por default |
| Privacy | `privacy` | Export de usuario, delete requests, delete executor CLI y audit log basico | parcial |
| Tooling | `tools` | Registry de tools | preparatorio |
| AI Control Plane | `ai/actions` | Propuestas y ejecucion explicita controlada | parcial |
| AI Runtime | `ai/runtime` | Context builder, tool selection, safety policy, provider interface y Gemini proposal-only | proposal-only |
| AI Evaluation | `ai/evals` | Suite de evals de seguridad/runtime para fake y Gemini opcional | fake usable; Gemini real pendiente sin API key |

## Dependencias entre modulos

- `cmd/sofia` carga config, construye `internal/app.Container` y arranca el servidor.
- `internal/app` compone repositorios, services, handlers y rutas HTTP sin framework DI.
- `server` conoce handlers HTTP, pero no infraestructura.
- `auth` no depende de users.
- `users` depende de auth repository para `GetMe`.
- `activities` tiene un bridge hacia reminders para crear/cancelar reminders asociados.
- `reminders` depende de activities repository para validar ownership y de users service para `reminders_enabled`.
- `notifications` registra device tokens por usuario; el reminder worker puede usar FCM como provider opcional.
- `insights` depende de activities repository para validar ownership.
- `memory` depende de users service para `memory_enabled`.
- `conversations` depende por interfaz del AI Runtime existente; persiste user message, invoca runtime en `dry_run=true` y persiste assistant message sin ejecutar tools.
- `privacy` orquesta servicios existentes para exportar datos del usuario autenticado, registrar audit logs y procesar borrados fisicos aprobados por CLI interno.
- `ai/actions` depende de tools, users, activities, reminders y memory para ejecucion minima.
- `ai/runtime` depende por interfaces de users, activities, reminders, insights, memory, tools y ai/actions; no depende de Firestore ni de producto. Gemini vive en infraestructura como `ModelClient`.
- `ai/evals` compone un entorno in-memory sintetico para validar AI Runtime sin tocar Firestore ni ejecutar propuestas.

## Acoplamientos aceptables hoy

- `internal/app` como composition root explicito: aceptable mientras mantenga constructores simples.
- `activities -> ReminderBridge`: acoplamiento controlado por interfaz.
- `ai/actions -> domain services`: necesario para executor minimo, pero debe moverse a policy/executor registry si crece.

## Acoplamientos/deuda a vigilar

- `ai/actions` puede convertirse rapido en orquestador gigante si se agregan mas tools sin registry de executors.
- `internal/app` concentra wiring manual; vigilar crecimiento y partir builders por contexto si se vuelve dificil de leer.
- Tools y actions exponen structs de dominio con JSON tags; conviene DTOs dedicados si la API se estabiliza.
- Placeholders heredados vacios pueden confundir:
  - `internal/app`
  - `internal/http`
  - `internal/domain/activity`

## Seguridad actual

- JWT protege todos los endpoints privados.
- Ownership por `user_id` desde token.
- `user_id` no se acepta desde body en modulos actuales.
- `memory_enabled` bloquea memory.
- `reminders_enabled` bloquea reminder creation.
- AI Actions no ejecuta automaticamente.
- AI Runtime no ejecuta acciones automaticamente; en `dry_run=false` solo puede crear proposals seguras.
- Gemini output no se confia; pasa por ToolSelector, PolicyEngine, SafetyPolicy y JSON Schema validation.
- `cmd/tools/run-ai-evals` verifica regresiones de seguridad antes de uso prolongado.
- FCM esta deshabilitado y en dry-run por default; el provider real requiere configuracion explicita.
- Reminder worker esta deshabilitado por default y usa claim/lease transaccional en Firestore para evitar procesamiento paralelo multi-instancia; falta validacion operacional con multiples procesos reales.
- Soft delete existe en memory; falta delete definitivo.
- Privacy controls: export propio, delete request pendiente, executor fisico solo para requests aprobadas y audit log basico.

## Listo para usar

- Auth.
- Users/profile/settings.
- Activities.
- Reminders CRUD persistido.
- Recurrencia simple de reminders: `none`, `daily`, `weekly`, `monthly`.
- Reminder worker single-process con provider `noop` y FCM opcional.
- Device token management para notificaciones.
- Insights tracking.
- Memory textual.
- Conversations y mensajes cronologicos basicos.
- Export de datos de usuario y delete request controlado.

## Parcial/preparatorio

- Reminder Engine runtime.
- AI Action Proposals.
- AI Runtime foundation con provider Gemini proposal-only.
- AI Runtime evals con fake provider; Gemini opcional existe pero falta corrida real con API key en este entorno.
- Reminder Worker Runtime con polling configurable e idempotencia local.
- FCM delivery opcional, deshabilitado/dry-run por default.
- Tool Registry.
- Firestore production hardening.
- Data model reconciliation: Firestore real contiene colecciones legacy y el backend espera colecciones modulares nuevas. Existe migrador dry-run en `cmd/tools/migrate-firestore-legacy`.
- Migration closure: `target-mode=modular-only` permite preparar solo `user_profiles` y `ai_settings`; `users`, `activities` y `reminders` legacy quedan fuera de apply seguro.
- Memory semantica/RAG.

## No listo

- AI Runtime production-ready con evals.
- Ejecucion autonoma real.
- Streaming/multimodal/conversation summaries.
- Push notifications reales por default/produccion.
- Recurrencia compleja/RRULE.
- Worker multi-instancia con lease/transaccion Firestore.
- Hardware/smart home.
- Privacy export/delete definitivo.

## Data Model Reconciliation

Firestore real contiene colecciones legacy/conceptuales:

- `ai_memory`
- `ai_actions_log`
- `ai_policy`
- `bodies`
- `conversations`
- `daily_summaries`
- `locations`
- `message_embeddings`

Colecciones `users`, `activities` y `reminders` existen, pero con schema legacy/parcial. Sprint 9.5 agrega documentacion y herramienta dry-run para preparar migracion no destructiva hacia `user_profiles`, `ai_settings`, `activities` y `reminders` normalizados.

No se deben borrar colecciones legacy ni mezclar hardware/RAG en runtime actual.

## Tests

Release Gate V1.4 (2026-07-16): backend tests/vet/build and Flutter analyze/tests/web build pass; synthetic memory E2E passes. Firestore and Gemini remain externally blocked.

Ultima verificacion:

```sh
env GOCACHE=/Users/armandoalvarado/Documents/Sophia/sophia_ai_backend/sofia-backend/.cache/go-build go test ./...
```

Resultado: suite completa pasando.
