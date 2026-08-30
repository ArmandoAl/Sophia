# Sprint 25 Report - Gemini Eval Artifacts

Fecha: 2026-07-03

## Objetivo

Ejecutar y endurecer evaluaciones del AI Runtime con Gemini opcional, sin activar produccion, sin ejecucion automatica y sin guardar prompts completos.

## Cambios

- `cmd/tools/run-ai-evals` ahora genera `docs/gemini_eval_report.json`.
- El reporte JSON registra:
  - provider name,
  - model,
  - latency total,
  - status,
  - resumen pass/fail,
  - proposed/rejected/schema/policy counts,
  - metadata de privacidad.
- Si no existe `GEMINI_API_KEY`, Gemini queda `skipped` de forma explicita y CI no falla.
- Los reportes siguen sin guardar:
  - prompts completos,
  - contexto completo,
  - raw model responses,
  - API keys,
  - tokens,
  - secrets.

## Costos

El schema del reporte queda preparado para `usage` y `cost`.

Estado actual: el Runtime/ModelClient no expone usage metadata de Gemini hacia el runner de evals. Por eso `usage` y `cost` se omiten cuando no hay datos confiables. No se inventan tokens ni costos.

## Ejecucion

Comando ejecutado:

```sh
env GOCACHE=/Users/armandoalvarado/Documents/Sophia/sophia_ai_backend/sofia-backend/.cache/go-build go run ./cmd/tools/run-ai-evals -provider=auto -output-dir=docs -dataset=docs/evals/ai_runtime_cases.json
```

Resultado:

- fake provider: ejecutado y PASS.
- Gemini provider: `skipped`.
- Motivo: `GEMINI_API_KEY not set`.

Archivos generados:

- `docs/ai_eval_report.md`
- `docs/gemini_eval_report.md`
- `docs/gemini_eval_report.json`

## Tests

Comando requerido:

```sh
env GOCACHE=/Users/armandoalvarado/Documents/Sophia/sophia_ai_backend/sofia-backend/.cache/go-build go test ./...
```

Resultado: PASS.

## Riesgos Restantes

- Falta una corrida real con `GEMINI_API_KEY` temporal.
- Falta exponer usage metadata real desde provider/runtime si se quieren costos aproximados confiables.
- Los evals actuales validan seguridad/proposal-only; no miden calidad conversacional profunda ni costo/calidad por modelo.
