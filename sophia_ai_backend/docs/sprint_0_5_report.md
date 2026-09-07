# Sprint 0.5 Report - Backend Foundation

Fecha: 2026-06-28

## Que cambio

- Se corrigio `internal/auth/aplication` a `internal/auth/application`.
- Se corrigio `hanlders.go` a `handlers.go`.
- Se agrego configuracion real en `internal/config`.
- Se agrego servidor modular en `internal/server`.
- Se agregaron utilidades HTTP JSON en `internal/platform/httpjson`.
- Se corrigieron errores de login para responder `401 {"message":"invalid credentials"}` en usuario inexistente o password incorrecto.
- Se agrego JWT basico con HMAC SHA-256.
- Se agrego middleware auth con `Authorization: Bearer <token>`.
- Se agrego endpoint protegido `GET /users/me`.
- Se agregaron validaciones minimas de register/login.
- Se agregaron tests de integracion HTTP para el sprint.
- Se documento el riesgo del JSON de credenciales Google/Firebase y se agrego patron a `.gitignore`.

## Archivos modificados o creados

Principales:

- `sophia_ai_backend/sofia-backend/cmd/sofia/main.go`
- `sophia_ai_backend/sofia-backend/internal/config/config.go`
- `sophia_ai_backend/sofia-backend/internal/server/server.go`
- `sophia_ai_backend/sofia-backend/internal/server/routes.go`
- `sophia_ai_backend/sofia-backend/internal/server/middleware.go`
- `sophia_ai_backend/sofia-backend/internal/platform/httpjson/httpjson.go`
- `sophia_ai_backend/sofia-backend/internal/auth/application/register_user.go`
- `sophia_ai_backend/sofia-backend/internal/auth/application/login_user.go`
- `sophia_ai_backend/sofia-backend/internal/auth/context/context.go`
- `sophia_ai_backend/sofia-backend/internal/auth/domain/errors.go`
- `sophia_ai_backend/sofia-backend/internal/auth/domain/repository.go`
- `sophia_ai_backend/sofia-backend/internal/auth/domain/user.go`
- `sophia_ai_backend/sofia-backend/internal/auth/infrastructure/memory_repository.go`
- `sophia_ai_backend/sofia-backend/internal/auth/infrastructure/jwt/jwt.go`
- `sophia_ai_backend/sofia-backend/internal/auth/interfaces/http/handlers.go`
- `sophia_ai_backend/sofia-backend/internal/auth/interfaces/http/requests.go`
- `sophia_ai_backend/sofia-backend/internal/auth/interfaces/http/responses.go`
- `sophia_ai_backend/sofia-backend/internal/users/interfaces/http/handler.go`
- `sophia_ai_backend/sofia-backend/internal/server/routes_test.go`
- `.gitignore`
- `docs/security_notes.md`
- `docs/architecture_state.md`
- `docs/sprint_0_5_report.md`

## Endpoints disponibles

### POST /auth/register

Request:

```json
{
  "name": "Armando",
  "email": "armando@example.com",
  "password": "strong-password"
}
```

Response `201`:

```json
{
  "id": "...",
  "name": "Armando",
  "email": "armando@example.com"
}
```

### POST /auth/login

Request:

```json
{
  "email": "armando@example.com",
  "password": "strong-password"
}
```

Response `200`:

```json
{
  "token": "...",
  "user": {
    "id": "...",
    "name": "Armando",
    "email": "armando@example.com"
  }
}
```

### GET /users/me

Headers:

```text
Authorization: Bearer <token>
```

Response `200`:

```json
{
  "id": "...",
  "name": "Armando",
  "email": "armando@example.com"
}
```

## Variables de entorno

| Variable | Default | Requerida | Descripcion |
| --- | --- | --- | --- |
| `ENV` | `development` | No | Ambiente de ejecucion. |
| `PORT` | `8080` | No | Puerto HTTP. |
| `JWT_SECRET` | `development-insecure-jwt-secret` en development | Si en production | Secreto HMAC para JWT. |
| `CORS_ALLOWED_ORIGINS` | vacio | No | Lista separada por comas de origins permitidos. |
| `REQUEST_BODY_LIMIT_BYTES` | `1048576` | No | Limite maximo de body JSON. |

## Pruebas ejecutadas

Comando:

```sh
env GOCACHE=/Users/armandoalvarado/Documents/Sophia/sophia_ai_backend/sofia-backend/.cache/go-build go test ./...
```

Resultado:

```text
ok github.com/armandoalvarado/sofia-backend/internal/server
```

Casos cubiertos:

- Register exitoso.
- Register email duplicado.
- Register password debil.
- Login exitoso devuelve token.
- Login usuario inexistente devuelve `401`.
- Login password incorrecto devuelve `401`.
- `GET /users/me` sin token devuelve `401`.
- `GET /users/me` con token valido devuelve usuario.

## Decisiones arquitectonicas

- No se agrego framework HTTP externo.
- No se implemento Firestore.
- Se mantuvo repositorio en memoria y se extendio con `FindByID`.
- Se centralizo JSON para evitar respuestas inconsistentes.
- Login devuelve siempre error generico para credenciales invalidas.
- JWT se implemento con libreria estandar para minimizar dependencias.
- `users` queda minimo y apoyado en `UserRepository` hasta que se implemente persistencia real.

## Deuda tecnica restante

- Persistencia real.
- Refresh/revocacion de tokens.
- Health/readiness endpoints.
- Logger estructurado.
- Tests unitarios por caso de uso.
- Configuracion de despliegue.
- Limpieza futura de placeholders `internal/app`, `internal/http` e `internal/domain/activity`.
- Validacion de formato de email, no solo requerido/lowercase/trim.

## Riesgos

- El JSON de credenciales Google/Firebase detectado debe salir del repo y puede requerir rotacion.
- JWT sin revocacion implica que un token emitido vive hasta expirar.
- El repositorio en memoria no es durable ni compartido entre procesos.
- El secreto default de development es inseguro y no debe llegar a production.
