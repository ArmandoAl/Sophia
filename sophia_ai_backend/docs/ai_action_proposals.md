# AI Action Proposals

Fecha: 2026-06-30

## Proposito

`ai/actions` permite que Sofia registre propuestas de accion estructuradas. No ejecuta acciones autonomas sin control. En Sprint 7 la ejecucion ocurre solo si el usuario llama explicitamente `POST /ai/action-proposals/{id}/execute`.

## Entidad AIActionProposal

Campos:

- `id`
- `user_id`
- `tool_name`
- `status`
- `proposed_input`
- `reason`
- `risk_level`
- `requires_confirmation`
- `expires_at`
- `created_at`
- `updated_at`
- `confirmed_at`
- `rejected_at`
- `executed_at`
- `execution_result`
- `execution_error`

## Status

- `proposed`
- `confirmed`
- `rejected`
- `executed`
- `failed`
- `expired`

## RiskLevel

- `low`
- `medium`
- `high`

## Ciclo de vida

1. Crear proposal con `POST /ai/action-proposals`.
2. Confirmar o rechazar.
3. Ejecutar solo via endpoint explicito.
4. Registrar `execution_result` o `execution_error`.

## Reglas de autonomia

`manual`:

- Puede crear proposals.
- Requiere confirmacion para todo.

`suggestive`:

- Puede crear proposals.
- Requiere confirmacion para cambios de datos.

`semi_autonomous`:

- Puede ejecutar herramientas low risk si `requires_confirmation=false`.
- En Sprint 7 se mantiene ejecucion explicita por endpoint.

## Endpoints

Todos requieren JWT.

### POST /ai/action-proposals

Request:

```json
{
  "tool_name": "create_activity",
  "risk_level": "low",
  "reason": "User asked to plan this task",
  "proposed_input": {
    "title": "Plan sprint",
    "timezone": "America/Tijuana"
  }
}
```

### GET /ai/action-proposals

Query params:

- `status`
- `limit`

### GET /ai/action-proposals/{id}

Lee una proposal propia.

### POST /ai/action-proposals/{id}/confirm

Marca `status=confirmed`.

### POST /ai/action-proposals/{id}/reject

Marca `status=rejected`.

### POST /ai/action-proposals/{id}/execute

Ejecuta proposal confirmada si la tool esta habilitada y soportada.

## Ejecucion Sprint 7

Implementado:

- `create_activity`
- `create_reminder`
- `create_memory`

Pendiente:

- `update_activity`
- `complete_activity`
- `cancel_reminder`
- `search_memory`
- `create_reflection`

Si una ejecucion soportada falla por validacion o settings, la proposal queda `failed` con `execution_error`.

## Seguridad

- Ownership por JWT.
- No se puede leer/confirmar/rechazar/ejecutar proposal de otro usuario.
- No se aceptan acciones peligrosas.
- No hay llamadas externas.
- No hay modelo LLM real.
- No hay ejecucion automatica en background.

## Falta para AI Context Builder

- Construir contexto desde user/profile/settings, activities, memory e insights.
- Presupuesto de tokens.
- Politicas de visibility/importance para memory.
- Tool selection real por LLM.
- Validacion formal de JSON Schema.
- Auditoria de propuestas generadas por IA.
