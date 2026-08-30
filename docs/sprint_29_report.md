# Sprint 29 Report - FCM Real Dry-Run Validation

Fecha: 2026-07-04

## Objetivo

Validar FCM end-to-end con dry-run real, sin enviar notificacion visible.

## Cambios

- `HTTPFCMSender` ahora usa FCM `validate_only=true` cuando `dryRun=true`.
- `cmd/tools/test-fcm-delivery` valida:
  - target requerido (`user_id` o `device_token_id`),
  - `FCM_PROJECT_ID`,
  - existencia/ownership de device token,
  - resultado del provider.
- El reporte incluye:
  - `fcm_validate_only`,
  - `visible_notification`,
  - `project_id_valid`,
  - `device_token_exists`,
  - `error_type`.
- Se agrega clasificacion:
  - `auth_error`,
  - `invalid_token`,
  - `quota`,
  - `network`,
  - `dry_run_success`.

## Diagnostico Local

Comando ejecutado:

```sh
env GOCACHE=/Users/armandoalvarado/Documents/Sophia/sophia_ai_backend/sofia-backend/.cache/go-build go run ./cmd/tools/test-fcm-delivery -output-dir=docs
```

Resultado:

- `status=failed`
- `error_type=invalid_token`
- motivo: falta `user-id` o `device-token-id`
- `visible_notification=false`

No se intento FCM real porque no habia target de device token.

## Docs

- `docs/fcm_real_dry_run.md`
- `docs/fcm_error_handling.md`
- `docs/fcm_test_report.md`
- `docs/fcm_delivery.md`

## Tests

Comando requerido:

```sh
env GOCACHE=/Users/armandoalvarado/Documents/Sophia/sophia_ai_backend/sofia-backend/.cache/go-build go test ./...
```

Resultado: PASS.

## Pendiente

- Ejecutar dry-run real con `device_token_id` valido.
- Definir limpieza automatica de tokens invalidos.
- Implementar retry/backoff por tipo de error.
