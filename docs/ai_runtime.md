# AI Runtime

Fecha: 2026-07-01

## Objetivo

`internal/ai/runtime` prepara la arquitectura para integrar Gemini, OpenAI u otro proveedor.

Sprint 15 agrega Gemini como primer provider real, manteniendo `fake` como default seguro.

El runtime actual:

- construye contexto limitado del usuario autenticado,
- selecciona tools disponibles segun settings/autonomia,
- aplica reglas minimas de seguridad,
- usa `ModelClient` generico con provider `fake` o `gemini`,
- responde `POST /ai/runtime/message`,
- puede crear `AIActionProposal` cuando `dry_run=false`,
- nunca ejecuta acciones automaticamente.

## Endpoint

### `POST /ai/runtime/message`

Requiere JWT.

Request:

```json
{
  "message": "Recuérdame estudiar mañana a las 8",
  "dry_run": true
}
```

Response con provider fake o Gemini:

```json
{
  "mode": "dry_run",
  "request_id": "...",
  "context_summary": {},
  "available_tools": [],
  "proposed_actions": [],
  "assistant_message": "AI Runtime is ready, but no model provider is configured yet.",
  "observability": {
    "provider_latency_ms": 0,
    "planned_actions_count": 0,
    "proposals_count": 0,
    "rejected_actions_count": 0
  }
}
```

En la practica `context_summary` y `available_tools` se llenan con datos reales limitados del usuario autenticado. `proposed_actions` puede contener propuestas simuladas del `FakeModelClient` o propuestas estructuradas de Gemini.

## Modulos

```text
internal/ai/runtime/domain
internal/ai/runtime/application
internal/ai/runtime/infrastructure
internal/ai/runtime/interfaces/http
```

## Componentes

| Componente | Responsabilidad | Estado |
| --- | --- | --- |
| `ContextBuilder` | Reunir contexto limitado del usuario | implementado inicial |
| `ToolSelector` | Filtrar tools por AISettings/autonomia | implementado inicial |
| `SafetyPolicy` | Bloquear ejecucion automatica y proposals inseguras | implementado inicial |
| `Planner` | Coordinar `ModelClient` | implementado inicial |
| `ModelClient` | Interfaz generica para proveedores futuros | implementado |
| `FakeModelClient` | Provider fake para pruebas/flujo sin LLM | implementado |
| `Gemini ModelClient` | Provider real proposal-only con structured output | implementado inicial |
| `RuntimeService` | Orquestar contexto, tools, planner, safety y proposals | implementado inicial |

## Reglas de seguridad

- `user_id` viene siempre del JWT/context.
- El request no acepta `user_id`.
- No hay ejecucion automatica.
- Acciones de escritura requieren proposal.
- Memory requiere `memory_enabled=true`.
- Reminders requiere `reminders_enabled=true`.
- Planning tools requieren `planning_enabled=true`.
- Tools deshabilitadas no se exponen.
- Proposals deben corresponder a una tool disponible para ese usuario.
- Gemini output no se confia: siempre pasa por ToolSelector, PolicyEngine, SafetyPolicy y JSON Schema validation.

## Integracion con Action Proposals

Con `dry_run=true`:

- se construye contexto,
- se seleccionan tools,
- se simulan proposals,
- no se persiste nada.

Con `dry_run=false`:

- se validan proposals con `SafetyPolicy`,
- se crea `AIActionProposal`,
- la accion queda en estado `proposed`,
- no se confirma ni ejecuta automaticamente.

## Config

```sh
AI_MODEL_PROVIDER=fake|gemini
GEMINI_API_KEY=...
GEMINI_MODEL=gemini-1.5-flash
AI_RUNTIME_ENABLED=true
AI_RUNTIME_PROPOSAL_ONLY=true
```

## Pendiente

- Ejecutar evals Gemini con API key temporal antes de produccion.
- Validacion de modelo/costo antes de produccion.
- Observabilidad externa/dashboard; no loggear prompts completos.
- Token budget con tokenizer real.
