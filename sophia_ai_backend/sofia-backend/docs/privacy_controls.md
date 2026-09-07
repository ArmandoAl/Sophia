# Privacy Controls

Fecha: 2026-07-01

## Objetivo

Agregar controles minimos de privacidad antes de integrar un LLM real.

## Endpoints

### `GET /users/me/export`

Requiere JWT.

Exporta solo datos del usuario autenticado:

- auth user
- user profile
- AI settings
- activities
- reminders
- mood entries
- activity outcomes
- reflections
- active memories
- AI action proposals

No acepta `user_id` por body/query/path. El `user_id` viene del JWT.

### `POST /users/me/delete-request`

Requiere JWT.

Crea una solicitud de borrado fisico controlado. No borra datos inmediatamente.

Request opcional:

```json
{
  "reason": "I want my data deleted"
}
```

Response:

```json
{
  "id": "...",
  "user_id": "...",
  "status": "pending",
  "reason": "I want my data deleted",
  "created_at": "...",
  "updated_at": "..."
}
```

## Memory Enforcement

- `deleted` no aparece en `GET /memory`.
- `deleted` no aparece en `POST /memory/search`.
- `deleted` no aparece en `GET /users/me/export`.
- `ai/runtime` solo incluye memories con `visibility=assistant_context` como contexto del asistente.

## Audit

Se registra audit log para:

- export de datos,
- solicitud de borrado,
- ejecucion de borrado completada,
- ejecucion de borrado fallida.

La metadata se redacta/trunca antes de persistirse.

## Delete Executor

El borrado fisico controlado existe como comando interno:

```sh
go run ./cmd/tools/process-delete-requests -limit=20 -output-dir=docs
```

Solo procesa solicitudes con status `approved`. Las solicitudes creadas por usuarios siguen naciendo como `pending`.

Colecciones borradas por defecto:

- `user_profiles`
- `ai_settings`
- `activities`
- `reminders`
- `mood_entries`
- `activity_outcomes`
- `reflections`
- `memories`
- `ai_action_proposals`

No borra por defecto:

- `users`
- `audit_logs`
- colecciones legacy/hardware.

## Limitaciones

- No hay workflow administrativo HTTP para aprobar delete requests.
- No hay export paginado/cursor.
- No hay firma/checksum del export.
- No hay retencion configurada para audit logs.
- No hay endpoint para consultar audit logs; por ahora es infraestructura interna.

## Antes de LLM real

Falta:

- policy de retencion,
- export firmado/versionado,
- redaccion fuerte para prompts/contexto,
- auditoria de prompts y tool calls.
