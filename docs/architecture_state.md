# Architecture State

Fecha: 2026-06-30

## Estado actual

Sprint 7 agrega `tools` y `ai/actions` como base segura para tool calling. Sofia ya puede registrar herramientas disponibles, crear propuestas de accion estructuradas, confirmarlas/rechazarlas y ejecutar explicitamente un subconjunto minimo de acciones confirmadas.

El backend queda organizado asi:

```text
sophia_ai_backend/sofia-backend
  /cmd/sofia
  /internal/config
  /internal/database
  /internal/platform/httpjson
  /internal/server
  /internal/auth
    /context
    /domain
    /application
    /infrastructure
      /jwt
    /interfaces/http
  /internal/users
    /domain
    /application
    /infrastructure
    /interfaces/http
  /internal/activities
    /domain
    /application
    /infrastructure
    /interfaces/http
  /internal/reminders
    /domain
    /application
    /infrastructure
    /interfaces/http
  /internal/insights
    /domain
    /application
    /infrastructure
    /interfaces/http
  /internal/memory
    /domain
    /application
    /infrastructure
    /interfaces/http
  /internal/tools
    /domain
    /application
    /infrastructure
    /interfaces/http
  /internal/ai/actions
    /domain
    /application
    /infrastructure
    /interfaces/http
```

## Modulos

| Modulo | Responsabilidad |
| --- | --- |
| `auth` | Credenciales, password hash, email, role, JWT y login/register. |
| `users` | Perfil, onboarding y configuracion inicial de Sofia por usuario. |
| `activities` | Planificacion personal: task, habit, routine, goal, event, checklist. |
| `reminders` | Reminder Engine minimo: recordatorios propios, due query, estado de entrega/cancelacion y asociacion con activities. |
| `insights` | Tracking de mood, outcomes y reflections con summary simple por rango. |
| `memory` | Memoria explicita del asistente, busqueda basica y soft delete. |
| `tools` | Registry de herramientas disponibles para propuestas de accion. |
| `ai/actions` | Propuestas de accion, confirmacion/rechazo y ejecucion explicita controlada. |

## Persistencia

Auth:

- `auth/infrastructure.InMemoryUserRepository`
- `auth/infrastructure.FirestoreUserRepository`
- Coleccion Firestore: `users`

Users:

- `users/infrastructure.InMemoryUserProfileRepository`
- `users/infrastructure.InMemoryAISettingsRepository`
- `users/infrastructure.FirestoreUserProfileRepository`
- `users/infrastructure.FirestoreAISettingsRepository`
- Colecciones Firestore: `user_profiles`, `ai_settings`

Activities:

- `activities/infrastructure.InMemoryActivityRepository`
- `activities/infrastructure.FirestoreActivityRepository`
- Coleccion Firestore: `activities`

Reminders:

- `reminders/infrastructure.InMemoryReminderRepository`
- `reminders/infrastructure.FirestoreReminderRepository`
- Coleccion Firestore: `reminders`

Insights:

- `insights/infrastructure.InMemoryMoodEntryRepository`
- `insights/infrastructure.InMemoryActivityOutcomeRepository`
- `insights/infrastructure.InMemoryReflectionRepository`
- `insights/infrastructure.FirestoreMoodEntryRepository`
- `insights/infrastructure.FirestoreActivityOutcomeRepository`
- `insights/infrastructure.FirestoreReflectionRepository`
- Colecciones Firestore: `mood_entries`, `activity_outcomes`, `reflections`

Memory:

- `memory/infrastructure.InMemoryMemoryRepository`
- `memory/infrastructure.FirestoreMemoryRepository`
- Coleccion Firestore: `memories`

Tools:

- `tools/infrastructure.InMemoryToolDefinitionRepository`
- `tools/infrastructure.FirestoreToolDefinitionRepository`
- Coleccion Firestore: `tool_definitions`

AI Actions:

- `ai/actions/infrastructure.InMemoryAIActionProposalRepository`
- `ai/actions/infrastructure.FirestoreAIActionProposalRepository`
- Coleccion Firestore: `ai_action_proposals`

Los use cases dependen de contratos de dominio, no de Firestore.

## Seguridad

- Todos los endpoints `/users/me*`, `/activities*`, `/reminders*`, `/insights*`, `/memory*`, `/tools*` y `/ai/action-proposals*` requieren JWT.
- El `user_id` nunca se acepta desde body.
- Ownership se deriva del token.
- Repositorios de activities, reminders, insights y memory siempre filtran por `user_id`.
- No se expone `password_hash`.
- `DecodeJSON` rechaza campos desconocidos.
- Si `AISettings.reminders_enabled=false`, no se permite crear reminders nuevos.
- Si `AISettings.memory_enabled=false`, no se permite crear, listar, leer, actualizar, borrar ni buscar memorias.
- Insights no acepta `user_id` desde body y no hace diagnostico clinico.
- Memory usa soft delete y no acepta `user_id` desde body.
- AI Actions no ejecuta nada automaticamente; solo registra proposals y ejecuta por endpoint explicito.
- No se ejecutan proposals de otro usuario.

## Endpoints actuales

| Metodo | Ruta | Protegido | Descripcion |
| --- | --- | --- | --- |
| `GET` | `/health` | No | Estado basico del servicio y Firestore. |
| `POST` | `/auth/register` | No | Registra auth user y crea defaults de users. |
| `POST` | `/auth/login` | No | Autentica y devuelve JWT. |
| `GET` | `/users/me` | Si | Devuelve auth user, profile y AI settings. |
| `PATCH` | `/users/me/profile` | Si | Actualiza perfil. |
| `GET` | `/users/me/ai-settings` | Si | Lee settings de Sofia. |
| `PATCH` | `/users/me/ai-settings` | Si | Actualiza settings de Sofia. |
| `POST` | `/users/me/onboarding/complete` | Si | Marca onboarding completado. |
| `POST` | `/activities` | Si | Crea activity. |
| `GET` | `/activities` | Si | Lista activities del usuario autenticado. |
| `GET` | `/activities/{id}` | Si | Lee activity propia. |
| `PATCH` | `/activities/{id}` | Si | Actualiza activity propia. |
| `POST` | `/activities/{id}/complete` | Si | Completa activity propia. |
| `POST` | `/activities/{id}/cancel` | Si | Cancela activity propia. |
| `POST` | `/activities/{id}/archive` | Si | Archiva activity propia. |
| `POST` | `/reminders` | Si | Crea reminder propio. |
| `GET` | `/reminders` | Si | Lista reminders propios. |
| `GET` | `/reminders/due` | Si | Lista reminders pendientes con `next_run_at <= now`. |
| `GET` | `/reminders/{id}` | Si | Lee reminder propio. |
| `PATCH` | `/reminders/{id}` | Si | Actualiza reminder propio. |
| `POST` | `/reminders/{id}/cancel` | Si | Cancela reminder propio. |
| `POST` | `/reminders/{id}/archive` | Si | Archiva reminder propio. |
| `POST` | `/insights/mood` | Si | Crea mood entry propia. |
| `GET` | `/insights/mood` | Si | Lista mood entries propias. |
| `POST` | `/insights/outcomes` | Si | Crea outcome asociado a una activity propia. |
| `GET` | `/insights/outcomes` | Si | Lista outcomes propios. |
| `POST` | `/insights/reflections` | Si | Crea reflection propia. |
| `GET` | `/insights/reflections` | Si | Lista reflections propias. |
| `GET` | `/insights/summary` | Si | Devuelve promedios simples y conteos por rango. |
| `POST` | `/memory` | Si | Crea memory propia. |
| `GET` | `/memory` | Si | Lista memories propias activas por default. |
| `GET` | `/memory/{id}` | Si | Lee memory propia y actualiza `last_accessed_at`. |
| `PATCH` | `/memory/{id}` | Si | Actualiza memory propia. |
| `POST` | `/memory/{id}/archive` | Si | Archiva memory propia. |
| `DELETE` | `/memory/{id}` | Si | Soft delete de memory propia. |
| `POST` | `/memory/search` | Si | Busqueda basica por texto, type, tag e importance. |
| `GET` | `/tools` | Si | Lista herramientas registradas. |
| `GET` | `/tools/{name}` | Si | Consulta una herramienta por nombre. |
| `POST` | `/ai/action-proposals` | Si | Crea propuesta de accion. |
| `GET` | `/ai/action-proposals` | Si | Lista propuestas propias. |
| `GET` | `/ai/action-proposals/{id}` | Si | Lee propuesta propia. |
| `POST` | `/ai/action-proposals/{id}/confirm` | Si | Confirma propuesta propia. |
| `POST` | `/ai/action-proposals/{id}/reject` | Si | Rechaza propuesta propia. |
| `POST` | `/ai/action-proposals/{id}/execute` | Si | Ejecuta propuesta confirmada soportada. |

## Limites conocidos

- No hay notificaciones push reales.
- No hay worker automatico en background; `ListDueReminders` es invocable por servicio y endpoint.
- `recurrence_rule` se guarda, pero no se expande ni ejecuta.
- Completar una activity aun no cambia reminders asociados; cancelar o archivar si cancela reminders pendientes asociados.
- No hay agentes autonomos.
- Memory no tiene embeddings, vector search ni RAG todavia.
- No hay LLM real integrado.
- No hay llamadas a APIs externas desde tools.
- Tools soportados para ejecucion real minima: `create_activity`, `create_reminder`, `create_memory`.
- Insights no genera interpretaciones clinicas, recomendaciones terapeuticas ni inferencias de personalidad.
- Firestore puede requerir indices para filtros compuestos de activities.
- No hay paginacion por cursor; solo `limit`.
- `internal/app`, `internal/http` e `internal/domain/activity` siguen como placeholders heredados.
