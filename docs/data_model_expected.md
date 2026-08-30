# Expected Data Model

Fecha: 2026-06-30

Este documento describe las colecciones Firestore esperadas por el backend actual. No es un export de Firestore.

## Colecciones

### `users`

| Aspecto | Detalle |
| --- | --- |
| Proposito | Credenciales/auth user. |
| Owner | `internal/auth`. |
| Campos principales | `id`, `name`, `email`, `password_hash`, `role`, `created_at`, `updated_at`. |
| Indices probables | `email` unico/logico. |
| Riesgos Firestore | Unicidad de email no garantizada sin transaccion/constraint externo. |
| Relaciones logicas | 1:1 con `user_profiles`, `ai_settings`; 1:N con todos los datos de usuario. |

### `user_profiles`

| Aspecto | Detalle |
| --- | --- |
| Proposito | Perfil personal/configuracion basica. |
| Owner | `internal/users`. |
| Campos principales | `id`, `user_id`, `display_name`, `preferred_name`, `timezone`, `locale`, `avatar_url`, `onboarding_completed`, timestamps. |
| Indices probables | `user_id`. |
| Riesgos Firestore | Falta validacion IANA/BCP 47. |
| Relaciones logicas | 1:1 con `users`. |

### `ai_settings`

| Aspecto | Detalle |
| --- | --- |
| Proposito | Preferencias de Sofia por usuario. |
| Owner | `internal/users`. |
| Campos principales | `user_id`, `assistant_name`, `communication_style`, `proactivity_level`, `autonomy_level`, `memory_enabled`, `reminders_enabled`, `planning_enabled`, `preferred_language`, timestamps. |
| Indices probables | `user_id`. |
| Riesgos Firestore | Settings faltantes se autocrean en algunos flows; consistencia depende de repos. |
| Relaciones logicas | Controla `memory`, `reminders`, `ai/actions`. |

### `activities`

| Aspecto | Detalle |
| --- | --- |
| Proposito | Planificacion: task, habit, routine, goal, event, checklist. |
| Owner | `internal/activities`. |
| Campos principales | `id`, `user_id`, `type`, `title`, `status`, `priority`, fechas, `timezone`, `tags`, `source`, `created_by`, reminder fields, AI fields. |
| Indices probables | `user_id`, `user_id+status`, `user_id+type`, `user_id+priority`, `user_id+tags`, rangos por `start_at/due_at/created_at`. |
| Riesgos Firestore | Filtros compuestos pueden requerir indices; no hay cursor pagination. |
| Relaciones logicas | Puede crear `reminders`; outcomes/reflections/mood pueden referenciarla. |

### `reminders`

| Aspecto | Detalle |
| --- | --- |
| Proposito | Recordatorios programados. |
| Owner | `internal/reminders`. |
| Campos principales | `id`, `user_id`, `activity_id`, `title`, `status`, `scheduled_at`, `timezone`, `next_run_at`, retry/delivery fields, `source`, `created_by`, timestamps. |
| Indices probables | `user_id+status`, `user_id+activity_id`, `user_id+status+next_run_at`. |
| Riesgos Firestore | Due query/worker futuro necesita indices e idempotencia. |
| Relaciones logicas | Opcionalmente asociado a `activities`. |

### `mood_entries`

| Aspecto | Detalle |
| --- | --- |
| Proposito | Registro de mood/energia/estres/foco. |
| Owner | `internal/insights`. |
| Campos principales | `id`, `user_id`, `activity_id`, `mood`, `energy_level`, `stress_level`, `focus_level`, `note`, `occurred_at`, `timezone`, timestamps. |
| Indices probables | `user_id+occurred_at`, `user_id+activity_id`. |
| Riesgos Firestore | Datos sensibles; falta export/delete definitivo. |
| Relaciones logicas | Puede referenciar `activities`. |

### `activity_outcomes`

| Aspecto | Detalle |
| --- | --- |
| Proposito | Resultado de activity. |
| Owner | `internal/insights`. |
| Campos principales | `id`, `user_id`, `activity_id`, `result`, `completion_quality`, `satisfaction_level`, `difficulty_level`, `note`, timestamps. |
| Indices probables | `user_id+activity_id`, `user_id+created_at`. |
| Riesgos Firestore | Falta `occurred_at`; hoy usa `created_at` para rango. |
| Relaciones logicas | Requiere `activity_id` propia. |

### `reflections`

| Aspecto | Detalle |
| --- | --- |
| Proposito | Reflexiones personales/productivas. |
| Owner | `internal/insights`. |
| Campos principales | `id`, `user_id`, `activity_id`, `title`, `content`, `sentiment`, `tags`, `occurred_at`, timestamps. |
| Indices probables | `user_id+occurred_at`, `user_id+activity_id`, `tags`. |
| Riesgos Firestore | Muy sensible; puede alimentar memory en futuro. |
| Relaciones logicas | Puede referenciar `activities`; futuro source de `memories`. |

### `memories`

| Aspecto | Detalle |
| --- | --- |
| Proposito | Memoria explicita consultable. |
| Owner | `internal/memory`. |
| Campos principales | `id`, `user_id`, `type`, `title`, `content`, `summary`, `source`, `source_id`, `confidence`, `importance`, `tags`, `visibility`, `status`, timestamps, `last_accessed_at`. |
| Indices probables | `user_id+status`, `user_id+type`, `user_id+importance`, `user_id+tags`, `user_id+visibility`. |
| Riesgos Firestore | Soft delete no borra fisicamente; busqueda textual no escala/semantica. |
| Relaciones logicas | Puede referenciar sources futuras con `source/source_id`. |

### `tool_definitions`

| Aspecto | Detalle |
| --- | --- |
| Proposito | Registry de tools disponibles. |
| Owner | `internal/tools`. |
| Campos principales | `id`, `name`, `description`, `category`, `input_schema`, `output_schema`, `requires_confirmation`, `enabled`, `min_autonomy_level`, timestamps. |
| Indices probables | `name`, `enabled`, `category`. |
| Riesgos Firestore | Seed en startup; schemas guardados como string. |
| Relaciones logicas | Referenciado por `ai_action_proposals.tool_name`. |

### `ai_action_proposals`

| Aspecto | Detalle |
| --- | --- |
| Proposito | Propuestas de accion AI controladas por usuario. |
| Owner | `internal/ai/actions`. |
| Campos principales | `id`, `user_id`, `tool_name`, `status`, `proposed_input`, `reason`, `risk_level`, confirmation/execution timestamps, result/error. |
| Indices probables | `user_id+status`, `user_id+created_at`, `tool_name`. |
| Riesgos Firestore | `proposed_input` puede ser sensible; falta expiracion automatica. |
| Relaciones logicas | Tool registry; puede crear activities/reminders/memories. |
# Conversation V1 namespace

`v1_conversations` and `v1_conversation_messages` are separate from legacy `conversations` and its `messages` subcollection.
