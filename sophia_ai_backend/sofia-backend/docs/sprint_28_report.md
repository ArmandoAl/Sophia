# Sprint 28 Report - Gemini Provider Diagnostics

Fecha: 2026-07-04

## Objetivo

Diagnosticar y corregir errores 403 del proveedor Gemini sin comprometer seguridad.

## Hallazgo Inicial

`docs/gemini_eval_report.md` y `docs/gemini_eval_report.json` muestran fallos del provider Gemini con:

```text
gemini request failed: status=403
```

Antes de este sprint, el provider no exponia una clasificacion suficiente para distinguir permission denied, quota, model not found u otros errores.

## Cambios

- `GeminiError` tipado en `internal/ai/runtime/infrastructure/gemini`.
- Captura de body de error redacted/truncado.
- Clasificacion de:
  - `unauthorized`
  - `permission_denied`
  - `model_not_found`
  - `quota`
  - `request_failed`
- Runtime observability ahora puede usar `ProviderErrorType()` si el provider lo expone.
- Nuevo comando:
  - `cmd/tools/test-gemini-provider`
- Docs:
  - `docs/gemini_provider_diagnostic.md`
  - `docs/gemini_troubleshooting.md`

## Config Gemini

Modelos documentados:

- `GEMINI_MODEL=gemini-2.5-flash`
- `GEMINI_MODEL=gemini-1.5-flash`

El backend mantiene default historico `gemini-1.5-flash`, pero el comando de diagnostico permite override:

```sh
go run ./cmd/tools/test-gemini-provider -model=gemini-2.5-flash -output-dir=docs
```

## Diagnostico Ejecutado

Comando:

```sh
env GOCACHE=/Users/armandoalvarado/Documents/Sophia/sophia_ai_backend/sofia-backend/.cache/go-build go run ./cmd/tools/test-gemini-provider -output-dir=docs
```

Resultado actual:

- status: `skipped`
- reason: `GEMINI_API_KEY not set`

No se hizo llamada real en este entorno.

## Seguridad

- No se imprime `GEMINI_API_KEY`.
- No se guardan prompts completos.
- No se guarda contexto completo.
- El body de error se redacted/trunca.

## Tests

Comando requerido:

```sh
env GOCACHE=/Users/armandoalvarado/Documents/Sophia/sophia_ai_backend/sofia-backend/.cache/go-build go test ./...
```

Resultado: PASS.

## Pendiente

- Ejecutar diagnostico con `GEMINI_API_KEY` real.
- Si persiste 403, revisar `error_type`:
  - `permission_denied`: permisos/API/restricciones de key.
  - `quota`: billing/cuota/rate limit.
  - `model_not_found`: probar otro `GEMINI_MODEL` y verificar disponibilidad.
