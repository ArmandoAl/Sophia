# Sprint 16 Report: AI Runtime Evaluations

## Objetivo

Crear una suite de evaluaciones para AI Runtime antes de permitir uso real prolongado.

No se agrego ejecucion automatica. No se integro hardware. No se agregaron features grandes de producto.

## Cambios implementados

- Nuevo modulo `internal/ai/evals`.
- Nuevo comando `cmd/tools/run-ai-evals`.
- Reporte generado en `docs/ai_eval_report.md`.
- Tests automatizados para evals criticas con fake provider.
- Documento de seguridad en `docs/ai_runtime_safety.md`.

## Casos cubiertos

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
- asegurar que no ejecuta automaticamente.

## Comando

Default fake provider:

```sh
go run ./cmd/tools/run-ai-evals -provider=fake -output-dir=docs
```

Gemini opcional:

```sh
go run ./cmd/tools/run-ai-evals -provider=gemini -output-dir=docs
```

Auto:

```sh
go run ./cmd/tools/run-ai-evals -provider=auto -output-dir=docs
```

`auto` usa Gemini solo si `GEMINI_API_KEY` esta presente; si no, usa fake.

## Resultado actual

Provider usado para el reporte:

```text
fake
```

Todos los casos actuales pasan.

## Pruebas

Ejecutado:

```sh
go test ./...
```

Resultado: suite completa pasando.

## Riesgos restantes

- Gemini debe evaluarse con API key real antes de rollout prolongado.
- Las evals actuales son unitarias/sinteticas, no miden calidad conversacional profunda.
- Falta dataset versionado de conversaciones reales redacted.
- Falta score historico para detectar degradaciones entre modelos.
