# Reminders Domain

Fecha: 2026-06-30

## Proposito

`Reminder` es el dominio responsable de representar recordatorios programados para un usuario. En Sprint 4 no envia notificaciones reales ni corre un scheduler automatico; deja el backend preparado para consultar recordatorios vencidos y para que un worker futuro los procese.

## Entidad Reminder

Campos:

- `id`
- `user_id`
- `activity_id`
- `title`
- `description`
- `status`
- `scheduled_at`
- `timezone`
- `recurrence_rule`
- `next_run_at`
- `last_run_at`
- `delivered_at`
- `failed_at`
- `failure_reason`
- `retry_count`
- `max_retries`
- `source`
- `created_by`
- `created_at`
- `updated_at`

## Enums

ReminderStatus:

- `pending`
- `due`
- `delivered`
- `failed`
- `canceled`
- `archived`

ReminderSource:

- `user`
- `activity`
- `ai_suggested`
- `ai_created`

CreatedBy:

- `user`
- `sofia`

## Defaults

Al crear un reminder:

- `status`: `pending`
- `next_run_at`: igual a `scheduled_at`
- `source`: `user`, salvo reminders creados desde Activity que usan `activity`
- `created_by`: `user`
- `max_retries`: `3` si no se envia

## Validaciones

- `title` requerido.
- `scheduled_at` requerido y RFC3339 en HTTP.
- `timezone` requerido.
- `activity_id`, si se envia, debe pertenecer al usuario autenticado.
- No se acepta `user_id` desde body.
- Campos desconocidos se rechazan por `DecodeJSON`.
- Si `AISettings.reminders_enabled=false`, crear reminders responde `409`.

## Endpoints

Todos requieren JWT.

### POST /reminders

Request:

```json
{
  "activity_id": "optional-activity-id",
  "title": "Call mom",
  "description": "Short reminder note",
  "scheduled_at": "2026-07-01T09:00:00Z",
  "timezone": "America/Tijuana",
  "recurrence_rule": "",
  "max_retries": 3
}
```

Response `201`:

```json
{
  "id": "reminder-id",
  "user_id": "user-id",
  "activity_id": "optional-activity-id",
  "title": "Call mom",
  "status": "pending",
  "scheduled_at": "2026-07-01T09:00:00Z",
  "timezone": "America/Tijuana",
  "next_run_at": "2026-07-01T09:00:00Z",
  "retry_count": 0,
  "max_retries": 3,
  "source": "user",
  "created_by": "user"
}
```

### GET /reminders

Query params:

- `status`
- `activity_id`
- `from`
- `to`
- `limit`

Response:

```json
{
  "reminders": []
}
```

### GET /reminders/due

Lista reminders propios con:

- `status=pending`
- `next_run_at <= now`

Query params:

- `limit`

### GET /reminders/{id}

Devuelve un reminder propio. Si pertenece a otro usuario responde `404`.

### PATCH /reminders/{id}

Permite actualizar:

- `title`
- `description`
- `scheduled_at`
- `timezone`
- `recurrence_rule`
- `max_retries`

### POST /reminders/{id}/cancel

Marca `status=canceled`.

### POST /reminders/{id}/archive

Soft delete: marca `status=archived`.

## Integracion con Activities

- `POST /activities` crea un Reminder asociado cuando `has_reminder=true` y `reminder_at` esta definido.
- El Reminder asociado usa `activity_id`, `source=activity`, mismo `title`, `description`, `timezone` y `recurrence_rule`.
- `POST /activities/{id}/cancel` cancela reminders pendientes asociados.
- `POST /activities/{id}/archive` cancela reminders pendientes asociados.
- `POST /activities/{id}/complete` no altera reminders asociados todavia.

## Persistencia

Memory:

- `reminders/infrastructure.InMemoryReminderRepository`

Firestore:

- `reminders/infrastructure.FirestoreReminderRepository`
- Coleccion: `reminders`

Campos Firestore principales:

- `id`
- `user_id`
- `activity_id`
- `title`
- `description`
- `status`
- `scheduled_at`
- `timezone`
- `recurrence_rule`
- `next_run_at`
- `last_run_at`
- `delivered_at`
- `failed_at`
- `failure_reason`
- `retry_count`
- `max_retries`
- `source`
- `created_by`
- `created_at`
- `updated_at`

## Limitaciones

- No hay push notifications.
- No hay scheduler automatico ni goroutine de background.
- No hay expansion de `recurrence_rule`.
- No hay transacciones cross-repository entre Activity y Reminder.
- No hay paginacion por cursor.
- Firestore puede requerir indices para filtros compuestos.
