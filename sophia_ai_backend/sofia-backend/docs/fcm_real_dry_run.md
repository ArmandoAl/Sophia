# FCM Real Dry-Run

Fecha: 2026-07-04

## Objetivo

Validar FCM end-to-end sin enviar una notificacion visible al usuario.

## Comando

Por device token:

```sh
env GOCACHE=/Users/armandoalvarado/Documents/Sophia/sophia_ai_backend/sofia-backend/.cache/go-build go run ./cmd/tools/test-fcm-delivery -device-token-id=<device_token_id> -dry-run=true
```

Por usuario:

```sh
env GOCACHE=/Users/armandoalvarado/Documents/Sophia/sophia_ai_backend/sofia-backend/.cache/go-build go run ./cmd/tools/test-fcm-delivery -user-id=<user_id> -dry-run=true
```

## Requisitos

- `PERSISTENCE_DRIVER=firestore`
- `FIRESTORE_PROJECT_ID`
- `FIRESTORE_DATABASE_ID` si aplica
- `FCM_PROJECT_ID`
- `GOOGLE_APPLICATION_CREDENTIALS` o Application Default Credentials
- Un `device_tokens/{id}` existente y enabled

## Seguridad

- `-dry-run=true` es default.
- El sender HTTP usa FCM `validate_only=true`.
- `visible_notification=false` en el reporte.
- No se imprime el token completo.
- No se guarda el payload completo.

## Reporte

El comando escribe:

```text
docs/fcm_test_report.md
```

Campos relevantes:

- `status`
- `dry_run`
- `fcm_validate_only`
- `visible_notification`
- `project_id_valid`
- `device_token_exists`
- `token_redacted`
- `error_type`

## Estado Local Actual

La ultima ejecucion local no incluyo `user-id` ni `device-token-id`, por lo que el reporte quedo en `failed` con `error_type=invalid_token` y no intento FCM real.
