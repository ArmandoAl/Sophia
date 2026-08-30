# Sprint 32 Report - API Smoke Test Collection

Fecha: 2026-07-05

## Objetivo

Crear una coleccion Postman/HTTP para probar manualmente Sofia Backend v0.1 sin agregar features ni modificar logica de producto.

## Entregables

- `docs/api_smoke_tests.md`
- `docs/postman_collection.json`
- `docs/sprint_32_report.md`

## Flujos Cubiertos

- Health.
- Register.
- Login.
- Users/me.
- Update profile.
- Get/update AI settings.
- Create/list activities.
- Create/list reminders.
- Create mood.
- Create memory.
- AI Runtime fake dry_run.
- Create action proposal.
- Confirm action proposal.
- Create second action proposal for reject path.
- Reject action proposal.
- Export user data.

## Variables Postman

- `base_url`
- `token`
- `user_id`
- `activity_id`
- `reminder_id`
- `memory_id`
- `action_proposal_id`
- `rejected_action_proposal_id`

## Seguridad

- La coleccion usa datos sinteticos.
- No contiene tokens reales.
- No contiene credenciales reales.
- El email se genera con dominio `example.test`.
- AI Runtime se prueba con `dry_run=true`.

## Tests

Comando requerido:

```sh
env GOCACHE=/Users/armandoalvarado/Documents/Sophia/sophia_ai_backend/sofia-backend/.cache/go-build go test ./...
```

Resultado: PASS.
