# Sprint 11 Report - Privacy Controls

Fecha: 2026-07-01

## Objetivo

Implementar controles minimos de privacidad antes de integrar un LLM real.

## Cambios implementados

Se creo modulo:

```text
internal/privacy/domain
internal/privacy/application
internal/privacy/infrastructure
internal/privacy/interfaces/http
```

Endpoints:

- `GET /users/me/export`
- `POST /users/me/delete-request`

Infraestructura:

- `audit_logs`
- `delete_requests`
- repositorios memory/firestore

## Export de usuario

Incluye solo datos del usuario autenticado:

- user
- profile
- AI settings
- activities
- reminders
- mood entries
- activity outcomes
- reflections
- active memories
- action proposals

No acepta `user_id` externo.

## Delete Request

`POST /users/me/delete-request` crea una solicitud con status `pending`.

No borra datos fisicamente todavia.

## Audit Log

Se registra audit log para:

- export de datos,
- delete request.

La metadata se redacta/trunca antes de persistirse.

## Memory Hardening

- Memories `deleted` quedan fuera de search/list/export.
- AI Runtime solo incluye memories con `visibility=assistant_context`.

## Tests

Se agregaron tests para:

- export solo datos propios,
- memory deleted no aparece en search/export,
- audit log se crea,
- export sin auth devuelve 401.

Comando ejecutado:

```sh
env GOCACHE=/Users/armandoalvarado/Documents/Sophia/sophia_ai_backend/sofia-backend/.cache/go-build go test ./...
```

Resultado: suite completa pasando.

## Deuda restante

- Delete fisico verificable por coleccion.
- Workflow administrativo de aprobacion/rechazo de delete requests.
- Export paginado/firmado.
- Audit log de prompts/tool calls cuando exista LLM.
- Politica de retencion.
