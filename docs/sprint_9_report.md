# Sprint 9 Report - AI Runtime Architecture Foundation

Fecha: 2026-07-01

## Objetivo

Implementar la base interna del AI Runtime sin integrar LLM real, sin llamar Gemini/OpenAI, sin ejecutar acciones automaticamente y sin agregar autonomia peligrosa.

## Documentacion revisada

- `docs/architecture_state.md`
- `docs/functionality_matrix.md`
- `docs/data_model_expected.md`
- `docs/tools_domain.md` historico
- `docs/ai_action_proposals.md` historico
- `docs/memory_domain.md` historico

Nota: `docs/database_schema.md` y `docs/database_collections_summary.md` no existen en este entorno porque Sprint 8 no pudo ejecutar el export real con `PERSISTENCE_DRIVER=memory`. La herramienta existe, pero requiere Firestore/emulator configurado.

## Cambios principales

Se creo:

- `internal/ai/runtime/domain`
- `internal/ai/runtime/application`
- `internal/ai/runtime/infrastructure`
- `internal/ai/runtime/interfaces/http`

Se agrego endpoint:

- `POST /ai/runtime/message`

Se actualizo:

- `cmd/sofia/main.go`
- `internal/server/routes.go`
- `internal/server/routes_test.go`
- `docs/architecture_state.md`

## Componentes implementados

| Componente | Estado | Detalle |
| --- | --- | --- |
| `ContextBuilder` | inicial | Reune user, profile, settings, activities, due reminders, insights summary y memories relevantes. |
| `ToolSelector` | inicial | Filtra tools por enabled, autonomia y settings. |
| `SafetyPolicy` | inicial | Bloquea ejecucion automatica y valida proposals seguras. |
| `Planner` | inicial | Coordina `ModelClient`. |
| `ModelClient` | interfaz | Provider-neutral. |
| `FakeModelClient` | implementado | Simula propuestas basicas sin LLM externo. |
| `RuntimeService` | inicial | Orquesta todo y crea proposals si `dry_run=false`. |

## Endpoint

### `POST /ai/runtime/message`

Request:

```json
{
  "message": "Recuérdame estudiar mañana a las 8",
  "dry_run": true
}
```

Response:

```json
{
  "mode": "dry_run",
  "context_summary": {},
  "available_tools": [],
  "proposed_actions": [],
  "assistant_message": "AI Runtime is ready, but no model provider is configured yet."
}
```

`context_summary`, `available_tools` y `proposed_actions` se llenan segun datos reales disponibles, settings y heuristicas del fake provider.

## Seguridad

- Requiere JWT.
- `user_id` viene del token.
- No acepta `user_id` desde body.
- `dry_run=true` no persiste proposals.
- `dry_run=false` solo crea proposals, no ejecuta acciones.
- `SafetyPolicy.CanExecute` siempre bloquea ejecucion automatica.
- Memory/reminders/planning respetan AISettings.

## Tests agregados

- runtime sin auth devuelve 401.
- context builder reune datos propios.
- no incluye datos de otro usuario.
- `memory_enabled=false` excluye memory.
- `reminders_enabled=false` excluye reminder tools/context.
- `planning_enabled=false` excluye planning tools.
- `dry_run` no crea proposal.
- `dry_run=false` crea proposal segura.
- no ejecuta automaticamente.

## Pruebas ejecutadas

```sh
env GOCACHE=/Users/armandoalvarado/Documents/Sophia/sophia_ai_backend/sofia-backend/.cache/go-build go test ./...
```

Resultado: suite completa pasando.

## Deuda tecnica

- `FakeModelClient` es heuristico; no representa calidad de un LLM real.
- Falta validacion JSON Schema formal.
- Falta policy engine versionado por tool/riesgo/autonomia.
- Falta token budget real para contexto.
- Falta observabilidad/redaccion de prompts.
- Falta provider real con timeouts/retries/config.

## Recomendacion para Sprint 10

Antes de integrar Gemini/OpenAI:

1. Ejecutar Firestore exporter contra emulator o proyecto real controlado.
2. Crear AI Context Builder versionado con budget de tokens.
3. Agregar JSON Schema validation para tool inputs.
4. Crear policy engine auditable.
5. Agregar trazas redacted para runtime.
6. Mantener modo proposal-only al integrar el primer LLM.
