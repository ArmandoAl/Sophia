# Environment

Fecha: 2026-06-28

## Modos soportados

| ENV | Uso | Persistencia permitida |
| --- | --- | --- |
| `development` | Desarrollo local. | `memory` o `firestore`. |
| `test` | Tests automatizados. | `memory` recomendado. |
| `production` | Runtime real. | `firestore` requerido. |

## Variables

| Variable | Default | Requerida | Descripcion |
| --- | --- | --- | --- |
| `ENV` | `development` | No | Debe ser `development`, `test` o `production`. |
| `PORT` | `8080` | No | Puerto HTTP. |
| `JWT_SECRET` | Default inseguro solo fuera de production | Si en production | Secreto HMAC para JWT. |
| `PERSISTENCE_DRIVER` | `memory` | Si en production como `firestore` | Selecciona `memory` o `firestore`. |
| `FIRESTORE_PROJECT_ID` | vacio | Si `PERSISTENCE_DRIVER=firestore` | Proyecto Google Cloud/Firestore. |
| `GOOGLE_APPLICATION_CREDENTIALS` | vacio | No | Ruta local para ADC en development. No debe apuntar a archivo versionado. |
| `FIRESTORE_EMULATOR_HOST` | vacio | No | Host del emulator para dev/test, por ejemplo `localhost:8081`. |
| `CORS_ALLOWED_ORIGINS` | vacio | No | Lista separada por comas de origins permitidos. |
| `REQUEST_BODY_LIMIT_BYTES` | `1048576` | No | Limite de body JSON. |
| `HTTP_READ_TIMEOUT` | `10s` | No | Timeout de lectura. |
| `HTTP_WRITE_TIMEOUT` | `10s` | No | Timeout de escritura. |
| `HTTP_IDLE_TIMEOUT` | `60s` | No | Timeout de conexiones idle. |
| `HTTP_READ_HEADER_TIMEOUT` | `5s` | No | Timeout de lectura de headers. |

## Desarrollo con memoria

```sh
ENV=development \
PERSISTENCE_DRIVER=memory \
JWT_SECRET=local-secret \
go run ./cmd/sofia
```

Este modo no requiere Firestore y `/health` reporta:

```json
{
  "status": "ok",
  "environment": "development",
  "firestore": "disabled"
}
```

## Desarrollo con Firestore

Con Application Default Credentials:

```sh
ENV=development \
PERSISTENCE_DRIVER=firestore \
FIRESTORE_PROJECT_ID=sophia-dev \
GOOGLE_APPLICATION_CREDENTIALS=/path/outside/repo/service-account.json \
JWT_SECRET=local-secret \
go run ./cmd/sofia
```

Con emulator:

```sh
ENV=development \
PERSISTENCE_DRIVER=firestore \
FIRESTORE_PROJECT_ID=sophia-dev \
FIRESTORE_EMULATOR_HOST=localhost:8081 \
JWT_SECRET=local-secret \
go run ./cmd/sofia
```

## Production

Production debe definir al menos:

```sh
ENV=production
PERSISTENCE_DRIVER=firestore
FIRESTORE_PROJECT_ID=<project-id>
JWT_SECRET=<strong-secret>
```

No se debe usar el default de development para `JWT_SECRET`.

## Seguridad

- No guardar service account JSON dentro del repo.
- Si un JSON de credenciales ya fue versionado o compartido, rotarlo.
- Preferir credenciales administradas por el entorno de despliegue.
- No loggear tokens, passwords ni payloads sensibles.
