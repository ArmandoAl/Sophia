# Notification Delivery Attempts

Fecha: 2026-07-03

## Objetivo

`notification_delivery_attempts` registra metadata operacional de intentos de delivery sin almacenar tokens completos ni payloads sensibles.

## Coleccion

```text
notification_delivery_attempts
```

## Campos

| Campo | Descripcion |
| --- | --- |
| `id` | ID del attempt. |
| `user_id` | Usuario propietario del reminder/token. |
| `reminder_id` | Reminder asociado o ID de prueba. |
| `provider` | `fcm` o `noop`. |
| `status` | `delivered`, `failed` o `skipped`. |
| `dry_run` | Indica si se evito envio real. |
| `error_type` | Codigo operativo seguro, sin cuerpo completo de error sensible. |
| `created_at` | Timestamp del attempt. |

Error types used by FCM paths:

- `auth_error`
- `invalid_token`
- `quota`
- `network`
- `dry_run_success`

## Seguridad

No se guardan:

- device tokens completos,
- cuerpos FCM completos,
- reminder descriptions,
- access tokens,
- credenciales,
- secretos.

## Escritura

El Reminder Worker registra attempts para reminders procesados. El comando `cmd/tools/test-fcm-delivery` tambien registra un attempt para pruebas controladas.

## Indices Esperados

- `user_id asc, created_at desc`
- `reminder_id asc, created_at desc`
- `provider asc, status asc, created_at desc`

## Limitaciones

- Todavia no hay endpoint publico/admin para consultar attempts.
- No hay limpieza/retencion automatica de attempts.
- Provider-specific retry policy sigue pendiente.
