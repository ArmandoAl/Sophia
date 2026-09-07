# AI Runtime Safety

Fecha: 2026-07-01

## Principio

Sofía AI Runtime sigue siendo proposal-only.

El modelo puede sugerir acciones estructuradas, pero no puede:

- ejecutar acciones automaticamente,
- confirmar propuestas,
- saltarse ownership,
- inyectar `user_id`,
- usar tools no disponibles,
- escribir campos no permitidos por schema.

## Capas de defensa

| Capa | Responsabilidad |
| --- | --- |
| ContextBuilder | Construye contexto limitado, redacted y scoped al usuario autenticado. |
| ToolSelector | Expone solo tools habilitadas por settings/autonomia. |
| PolicyEngine | Evalua autonomy, risk, tool category y confirmation. |
| SafetyPolicy | Rechaza planned actions inseguras antes de crear proposals. |
| JSON Schema validation | Rechaza proposed_input con campos faltantes, tipos invalidos o campos extra. |
| AI Action Proposals | Persiste acciones solo como `proposed`; ejecucion requiere endpoint explicito. |
| Evals | Verifican regresiones de seguridad antes de uso prolongado. |

## Evals actuales

El comando:

```sh
go run ./cmd/tools/run-ai-evals -provider=auto -output-dir=docs -dataset=docs/evals/ai_runtime_cases.json
```

genera:

```text
docs/ai_eval_report.md
docs/gemini_eval_report.md
docs/gemini_eval_report.json
```

`ai_eval_report.md` always contains fake-provider results. `gemini_eval_report.md` and `gemini_eval_report.json` contain Gemini results only when `GEMINI_API_KEY` exists; otherwise they are marked `skipped` and CI remains green.

Profiles:

- `safety`
- `planning`
- `memory`
- `reminders`
- `injection`

Metrics per case:

- pass/fail,
- latency_ms,
- proposed actions,
- rejected actions,
- schema failures,
- policy failures.

Casos cubiertos:

- crear activity desde mensaje natural,
- crear reminder desde mensaje natural,
- crear memory desde mensaje natural,
- buscar memoria sin exponer contenido sensible,
- rechazar `user_id` inyectado por modelo,
- rechazar tool inexistente,
- rechazar campos extra por schema,
- bloquear memory si `memory_enabled=false`,
- bloquear reminder si `reminders_enabled=false`,
- bloquear planning si `planning_enabled=false`,
- prompt injection basico,
- asegurar que no ejecuta automaticamente.

## Gemini opcional

Gemini no es obligatorio en CI.

Para correr evals con Gemini:

```sh
AI_MODEL_PROVIDER=gemini \
GEMINI_API_KEY=... \
GEMINI_MODEL=gemini-1.5-flash \
go run ./cmd/tools/run-ai-evals -provider=gemini -output-dir=docs
```

Para usar Gemini solo si hay API key:

```sh
go run ./cmd/tools/run-ai-evals -provider=auto -output-dir=docs
```

## Privacidad

Los reportes no deben guardar:

- prompts completos sensibles,
- tokens,
- secrets,
- password hashes,
- contenido completo de memorias.

Los inputs del reporte se redacted/truncan de forma basica. El JSON incluye metadata de provider/model/status/latency y deja usage/cost ausente cuando el provider no devuelve usage metadata.

## Bloqueo de uso prolongado

Antes de permitir uso real prolongado:

- `go test ./...` debe pasar,
- `run-ai-evals` debe pasar con fake,
- se recomienda correr evals con Gemini usando una API key temporal,
- cualquier FAIL debe bloquear rollout.
