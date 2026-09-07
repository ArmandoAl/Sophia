# Sprint 7 Report - Tool Registry & AI Action Proposals

Fecha: 2026-06-30

## Auditoria inicial

Se revisaron:

- `docs/architecture_state.md`
- `docs/sprint_6_report.md`
- `docs/memory_domain.md`
- `docs/activities_domain.md`
- `docs/reminders_domain.md`
- `docs/users_domain.md`

Estado real antes del sprint:

- Auth, users, activities, reminders, insights y memory estaban operativos.
- No existia registry de tools.
- No existian action proposals.
- No habia sistema de confirmacion/ejecucion controlada.

## Que cambio

- Se creo `internal/tools`.
- Se creo `internal/ai/actions`.
- Se agrego entidad `ToolDefinition`.
- Se agrego entidad `AIActionProposal`.
- Se implementaron repos memory/Firestore.
- Se seedearon tools default en startup.
- Se agregaron endpoints `/tools`.
- Se agregaron endpoints `/ai/action-proposals`.
- Se implemento ejecucion explicita minima:
  - `create_activity`
  - `create_reminder`
  - `create_memory`

## Endpoints

Tools:

- `GET /tools`
- `GET /tools/{name}`

AI Actions:

- `POST /ai/action-proposals`
- `GET /ai/action-proposals`
- `GET /ai/action-proposals/{id}`
- `POST /ai/action-proposals/{id}/confirm`
- `POST /ai/action-proposals/{id}/reject`
- `POST /ai/action-proposals/{id}/execute`

## Reglas de autonomia

- `manual`: proposals si, ejecucion requiere confirmacion.
- `suggestive`: proposals si, cambios de datos requieren confirmacion.
- `semi_autonomous`: preparado para low risk sin confirmacion, pero Sprint 7 mantiene ejecucion explicita por endpoint.

## Seguridad

- Todos los endpoints requieren JWT.
- Ownership por `user_id` del token.
- No se ejecutan proposals de otro usuario.
- Tool debe existir y estar enabled.
- No hay LLM real.
- No hay llamadas externas.
- No hay ejecucion automatica.
- Errores de ejecucion quedan registrados como `failed`.

## Pruebas ejecutadas

Comando:

```sh
env GOCACHE=/Users/armandoalvarado/Documents/Sophia/sophia_ai_backend/sofia-backend/.cache/go-build go test ./...
```

Resultado:

- Suite completa pasando.

Casos cubiertos:

- Seed/list tools.
- Crear proposal.
- Rechazar proposal.
- Confirmar proposal.
- Impedir acceso cruzado.
- Ejecutar `create_activity` confirmado.
- Ejecutar `create_reminder` confirmado.
- Ejecutar `create_memory` confirmado.
- Bloquear `create_memory` si `memory_enabled=false`.
- Respetar confirmacion basica para `manual`.

## Deuda tecnica

- Validacion JSON Schema real.
- Versionado de tools.
- Executor para tools pendientes.
- Policies mas finas por autonomia/riesgo.
- Expiracion automatica de proposals.
- Auditoria de proposals generadas por IA.
- AI Context Builder.
- Integracion LLM real y tool selection.

## Riesgos

- `proposed_input` puede contener datos sensibles.
- Ejecutar tools futuras sin policy granular seria riesgoso.
- Firestore puede requerir indices por `user_id/status`.
- JSON Schema aun es documentativo, no validacion formal.
