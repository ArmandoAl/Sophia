# FCM Error Handling

Fecha: 2026-07-04

## Error Types

| Error | Significado | Accion sugerida |
| --- | --- | --- |
| `auth_error` | Credenciales ADC, permisos o `FCM_PROJECT_ID` invalidos. | Revisar `GOOGLE_APPLICATION_CREDENTIALS`, ADC, proyecto y permisos Firebase Messaging. |
| `invalid_token` | Device token inexistente, deshabilitado, mal formado o no valido para el proyecto. | Re-registrar token en cliente, borrar tokens invalidos. |
| `quota` | Cuota o rate limit. | Revisar cuotas Firebase/Google Cloud y reintentar con backoff. |
| `network` | Error de red/timeout. | Reintentar con backoff; revisar conectividad. |
| `dry_run_success` | FCM validate-only paso sin notificacion visible. | El path FCM y credenciales parecen validos. |

## Redaccion

Los errores guardados o reportados no deben incluir:

- device token completo,
- access token,
- service account JSON,
- payload completo,
- reminder description.

## Worker

El worker registra `notification_delivery_attempts.error_type` usando la misma clasificacion basica.

## Limitaciones

- Aun falta politica automatica para borrar tokens invalidos.
- Aun falta retry/backoff diferenciado por error.
