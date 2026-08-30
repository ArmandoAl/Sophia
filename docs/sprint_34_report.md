# Sprint 34 Report - Auth Hardening

Fecha: 2026-07-05

## Objetivo

Endurecer auth antes de pruebas prolongadas sin implementar nuevas features de producto ni refresh tokens.

## Cambios

- Rate limiting basico para:
  - `POST /auth/register`
  - `POST /auth/login`
- Audit log para fallo de login:
  - `auth_login_failed`
- JWT expiration configurable:
  - `JWT_ACCESS_TOKEN_TTL`
- Configuracion nueva:
  - `AUTH_RATE_LIMIT_REQUESTS`
  - `AUTH_RATE_LIMIT_WINDOW`
- Documentacion de password policy actual.
- Refresh tokens documentados como pendiente.

## Seguridad

- No se loggea password.
- No se loggea token.
- Login failure audit usa metadata redacted/truncada.
- Rate limit es in-memory por proceso.

## Tests Agregados

- Login rate limit.
- Register rate limit.
- Expired token rejected.
- Auth hardening config.
- Login failure audit log.

## Deuda Tecnica

- Rate limit distribuido para multi-instancia.
- Lockout/backoff por cuenta/email.
- Refresh tokens con rotacion y revocacion.
- Password policy mas fuerte antes de produccion publica.

## Tests

Comando requerido:

```sh
env GOCACHE=/Users/armandoalvarado/Documents/Sophia/sophia_ai_backend/sofia-backend/.cache/go-build go test ./...
```

Resultado: PASS.
