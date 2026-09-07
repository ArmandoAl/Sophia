# Sprint 27 Report - Safe FCM Delivery Test

Fecha: 2026-07-03

## Objetivo

Preparar una prueba end-to-end segura de Firebase Cloud Messaging sin enviar notificaciones reales por default.

## Cambios

- Nuevo comando:
  - `cmd/tools/test-fcm-delivery`
- Nueva coleccion:
  - `notification_delivery_attempts`
- Nuevo repositorio:
  - memory delivery attempts,
  - Firestore delivery attempts.
- Worker registra attempts para reminders procesados.
- `DeviceTokenRepository` agrega `FindByID` para pruebas internas por `device_token_id`.

## Comando

```sh
go run ./cmd/tools/test-fcm-delivery -user-id=<user_id> -dry-run=true
```

o:

```sh
go run ./cmd/tools/test-fcm-delivery -device-token-id=<device_token_id> -dry-run=true
```

`dry-run=true` es default. Si no hay `FCM_PROJECT_ID`, el comando usa `dry-run-local` solo en dry-run para evitar requerir credenciales reales.

## Attempt Fields

- `id`
- `user_id`
- `reminder_id`
- `provider`
- `status`
- `dry_run`
- `error_type`
- `created_at`

## Seguridad

- No se imprimen tokens completos.
- No se guardan payloads completos.
- Reminder description no se envia.
- Device token ID con `user_id` ajeno se rechaza como not found.
- Real network send requiere `-dry-run=false` y configuracion FCM explicita.

## Tests

Cobertura agregada:

- dry-run no manda red real,
- attempt se registra,
- token se redacta,
- usuario no accede tokens de otro.

Comando requerido:

```sh
env GOCACHE=/Users/armandoalvarado/Documents/Sophia/sophia_ai_backend/sofia-backend/.cache/go-build go test ./...
```

Resultado: PASS.

## Riesgos Pendientes

- No hay UI/admin endpoint para consultar attempts.
- No hay politica de retencion/limpieza de attempts.
- Falta clasificacion fina de errores FCM para invalid token, auth, quota y retryable failures.
