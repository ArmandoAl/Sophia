# Activities Domain

Fecha: 2026-06-30

## Por que Activity y no Task

Sofia necesita planificar mas que tareas simples. Una Activity es el concepto superior para representar:

- `task`
- `habit`
- `routine`
- `goal`
- `event`
- `checklist`

Esto permite construir planificacion, recordatorios, rutinas y sugerencias futuras sin duplicar modelos ni forzar todo a comportarse como una task.

## Entidad Activity

Campos:

- `id`
- `user_id`
- `type`
- `title`
- `description`
- `status`
- `priority`
- `start_at`
- `due_at`
- `completed_at`
- `canceled_at`
- `archived_at`
- `timezone`
- `tags`
- `source`
- `created_by`
- `created_at`
- `updated_at`
- `has_reminder`
- `reminder_at`
- `recurrence_rule`
- `ai_manageable`
- `ai_notes`

## Enums

ActivityType:

- `task`
- `habit`
- `routine`
- `goal`
- `event`
- `checklist`

ActivityStatus:

- `pending`
- `active`
- `completed`
- `canceled`
- `archived`

ActivityPriority:

- `low`
- `medium`
- `high`
- `urgent`

ActivitySource:

- `user`
- `ai_suggested`
- `ai_created`
- `imported`

CreatedBy:

- `user`
- `sofia`

## Defaults

Al crear una activity:

- `type`: `task` si no se envia.
- `status`: `pending`.
- `priority`: `medium` si no se envia.
- `source`: `user`.
- `created_by`: `user`.
- `ai_manageable`: `false` si no se envia.

## Validaciones

- `title` requerido.
- `type` permitido.
- `status` permitido.
- `priority` permitido.
- `timezone` requerido.
- Fechas deben venir en RFC3339.
- `reminder_at` solo se permite si `has_reminder=true`.
- `recurrence_rule` se guarda como string, no se ejecuta.
- Campos desconocidos se rechazan.

## Endpoints

### POST /activities

Request:

```json
{
  "type": "task",
  "title": "Plan sprint",
  "description": "Prepare activities domain",
  "priority": "high",
  "start_at": "2026-07-01T09:00:00Z",
  "due_at": "2026-07-01T18:00:00Z",
  "timezone": "America/Tijuana",
  "tags": ["work", "sofia"],
  "has_reminder": true,
  "reminder_at": "2026-07-01T08:45:00Z",
  "recurrence_rule": "",
  "ai_manageable": false,
  "ai_notes": ""
}
```

Response `201`:

```json
{
  "id": "activity-id",
  "user_id": "user-id",
  "type": "task",
  "title": "Plan sprint",
  "status": "pending",
  "priority": "high",
  "timezone": "America/Tijuana",
  "tags": ["work", "sofia"],
  "source": "user",
  "created_by": "user",
  "has_reminder": true,
  "reminder_at": "2026-07-01T08:45:00Z",
  "ai_manageable": false
}
```

Nota Sprint 4:

- Si `has_reminder=true` y `reminder_at` viene definido, el backend crea automaticamente un `Reminder` asociado con `source=activity`.
- Si `AISettings.reminders_enabled=false`, crear una Activity con reminder asociado responde `409`.
- La creacion de Activity y Reminder todavia no es transaccional entre repositorios; si falla el reminder despues de crear la Activity, se devuelve error y queda deuda tecnica documentada para introducir unidad de trabajo/transacciones.

### GET /activities

Query params:

- `status`
- `type`
- `priority`
- `from`
- `to`
- `tag`
- `limit`

Ejemplo:

```text
GET /activities?status=pending&type=habit&priority=low&tag=health&limit=20
```

Response:

```json
{
  "activities": []
}
```

### GET /activities/{id}

Devuelve una activity propia. Si pertenece a otro usuario responde `404`.

### PATCH /activities/{id}

Permite actualizar:

- `type`
- `title`
- `description`
- `status`
- `priority`
- `start_at`
- `due_at`
- `timezone`
- `tags`
- `has_reminder`
- `reminder_at`
- `recurrence_rule`
- `ai_manageable`
- `ai_notes`

### POST /activities/{id}/complete

Marca `status=completed` y setea `completed_at`.

Comportamiento esperado con reminders:

- Por ahora no modifica reminders asociados.
- En un sprint futuro se debe decidir si completar una activity cancela reminders pendientes, los marca delivered, o conserva historial para recurrencia.

### POST /activities/{id}/cancel

Marca `status=canceled` y setea `canceled_at`.

Desde Sprint 4 cancela reminders pendientes asociados a la Activity.

### POST /activities/{id}/archive

Soft delete: marca `status=archived` y setea `archived_at`.

Desde Sprint 4 cancela reminders pendientes asociados a la Activity.

## Seguridad

- Todos los endpoints requieren JWT.
- No se acepta `user_id` en bodies.
- `user_id` se obtiene del token.
- Repositorios filtran por `user_id`.
- No hay acceso cruzado entre usuarios.

## Firestore

Coleccion:

- `activities`

Campos principales:

- `user_id`
- `type`
- `status`
- `priority`
- `tags`
- `start_at`
- `due_at`
- `created_at`

Nota: filtros compuestos pueden requerir indices Firestore cuando se usen en entorno real.

## Pendiente para Reminder Engine

Este sprint solo guarda:

- `has_reminder`
- `reminder_at`
- `recurrence_rule`

No implementa:

- scheduler
- worker
- notificaciones
- reintentos
- expansion de recurrencias
- timezone execution semantics
