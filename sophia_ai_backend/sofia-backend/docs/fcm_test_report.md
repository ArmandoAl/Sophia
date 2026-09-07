# FCM Test Report

Generated at: `2026-07-04T06:12:30Z`

- provider: `fcm`
- status: `failed`
- dry_run: `true`
- fcm_validate_only: `true`
- visible_notification: `false`
- project_id_valid: `false`
- device_token_exists: `false`
- user_id: ``
- device_token_id: ``
- token_redacted: ``
- reminder_id: ``
- attempt_id: ``
- error_type: `invalid_token`
- error: `user-id or device-token-id is required`

Full device tokens are intentionally excluded from this report.

## Interpretacion

Este entorno no incluyo `user-id` ni `device-token-id`, por lo que no se intento validar contra FCM real.

Para dry-run real usar:

```sh
env GOCACHE=/Users/armandoalvarado/Documents/Sophia/sophia_ai_backend/sofia-backend/.cache/go-build go run ./cmd/tools/test-fcm-delivery -device-token-id=<device_token_id> -dry-run=true
```

Con `dry-run=true`, FCM usa `validate_only=true` y no debe mostrar notificacion visible.
