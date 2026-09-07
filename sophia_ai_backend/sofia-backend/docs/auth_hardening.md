# Auth Hardening

Fecha: 2026-07-05

## Objetivo

Endurecer auth antes de pruebas prolongadas de Sofia Backend v0.1, manteniendo el alcance simple y sin introducir refresh tokens todavia.

## Rate Limiting

Endpoints protegidos por rate limit en memoria por proceso:

- `POST /auth/register`
- `POST /auth/login`

Configuracion:

| Variable | Default | Descripcion |
| --- | --- | --- |
| `AUTH_RATE_LIMIT_REQUESTS` | `10` | Max requests permitidas por IP+ruta dentro de la ventana. |
| `AUTH_RATE_LIMIT_WINDOW` | `1m` | Ventana de rate limit como duracion Go. |

Respuesta al exceder limite:

```json
{
  "message": "rate limit exceeded"
}
```

Status:

```text
429 Too Many Requests
```

Limitaciones:

- Es in-memory y por proceso.
- No comparte estado entre instancias.
- No sustituye WAF/API gateway/rate limit distribuido.

## Login Failure Audit

Los fallos de login por credenciales invalidas registran audit log:

- `action=auth_login_failed`
- `resource_type=auth`
- `resource_id=login`

Metadata redacted:

- `email` pasa por redaccion del servicio de privacidad.
- `remote_ip`
- `user_agent` truncado.

No se registra:

- password,
- token,
- request body completo.

## JWT Expiration

Nuevo TTL configurable:

| Variable | Default | Descripcion |
| --- | --- | --- |
| `JWT_ACCESS_TOKEN_TTL` | `24h` | Tiempo de vida del access token. |

Tokens expirados son rechazados por middleware auth y responden `401`.

## Password Policy

Politica actual:

- `name` requerido.
- `email` requerido.
- `email` normalizado con trim + lowercase.
- `password` requerido.
- longitud minima: 8 caracteres.

Pendiente recomendado antes de produccion publica:

- complejidad configurable,
- bloqueo de passwords comunes,
- deteccion de credenciales filtradas,
- rate limit distribuido,
- MFA opcional,
- rotacion/invalidacion de sesiones.

## Refresh Tokens

Refresh tokens quedan documentados como pendiente y no se implementan en este sprint.

Recomendacion futura:

- refresh token opaco,
- almacenado hasheado,
- rotacion por uso,
- revocacion por usuario/device,
- expiracion absoluta,
- audit logs de refresh/revoke.

