# Sprint 1 Report - Infrastructure

Fecha: 2026-06-28

## Auditoria inicial

Antes de modificar codigo se revisaron:

- Estado del grafo del repo.
- `docs/sprint_0_5_report.md`.
- `docs/architecture_state.md`.
- Estructura actual en `sophia_ai_backend/sofia-backend`.

Hallazgos:

- Sprint 0.5 estaba implementado: config, server modular, JSON helpers, JWT, middleware auth, `/users/me` y tests.
- Persistencia seguia en memoria.
- No existia bootstrap Firestore.
- No habia selector de persistencia por ambiente.
- No existia `/health`.
- CORS estaba implementado, pero sin test dedicado.
- Timeouts HTTP existian, pero estaban hardcodeados en `internal/server`.
- El JSON de credenciales Google/Firebase seguia presente en `sophia_ai_backend`, documentado como riesgo y no usado.
- No hay repo Git inicializado en la raiz del workspace.

## Que cambio

- Se agrego bootstrap Firestore en `internal/database/firestore.go`.
- Se agrego repositorio Firestore para auth users.
- Se agrego seleccion de persistencia `memory`/`firestore` desde config.
- Se agrego `/health`.
- Se agregaron timeouts configurables.
- Se reforzo config para `development`, `test` y `production`.
- Se agrego logging basico de startup, Firestore bootstrap, ListenAndServe y requests fallidos.
- Se agregaron tests de config, health, CORS, auth con memoria y mapeo Firestore sin nube.
- Se actualizo `docs/architecture_state.md`.
- Se creo `docs/environment.md`.

## Arquitectura de infraestructura

```text
cmd/sofia
  config.Load()
  buildUserRepository()
    memory -> auth/infrastructure.InMemoryUserRepository
    firestore -> database.NewFirestore + auth/infrastructure.FirestoreUserRepository
  server.New()

auth/application
  depende solo de auth/domain.UserRepository

auth/infrastructure
  contiene implementaciones memory/firestore

database
  contiene cliente Firestore compartible por infraestructura futura
```

Firestore no entra en los use cases.

## Variables de entorno

Ver tambien `docs/environment.md`.

| Variable | Descripcion |
| --- | --- |
| `ENV` | `development`, `test` o `production`. |
| `PORT` | Puerto HTTP. |
| `JWT_SECRET` | Requerido y seguro en production. |
| `PERSISTENCE_DRIVER` | `memory` o `firestore`. |
| `FIRESTORE_PROJECT_ID` | Requerido si `PERSISTENCE_DRIVER=firestore`. |
| `GOOGLE_APPLICATION_CREDENTIALS` | Opcional para ADC local. |
| `FIRESTORE_EMULATOR_HOST` | Opcional para emulator. |
| `CORS_ALLOWED_ORIGINS` | Origins permitidos separados por coma. |
| `REQUEST_BODY_LIMIT_BYTES` | Limite de body JSON. |
| `HTTP_READ_TIMEOUT` | Timeout de lectura. |
| `HTTP_WRITE_TIMEOUT` | Timeout de escritura. |
| `HTTP_IDLE_TIMEOUT` | Timeout idle. |
| `HTTP_READ_HEADER_TIMEOUT` | Timeout de headers. |

## Memoria vs Firestore

Memoria:

- Default en development/test.
- No requiere credenciales.
- No persiste reinicios.
- Ideal para tests y desarrollo rapido.

Firestore:

- Activado con `PERSISTENCE_DRIVER=firestore`.
- Requiere `FIRESTORE_PROJECT_ID`.
- Usa Application Default Credentials del entorno.
- Puede usar emulator mediante `FIRESTORE_EMULATOR_HOST`.
- En production es obligatorio.

## Estado de Firestore

Implementado:

- Bootstrap de cliente.
- Cierre de cliente en shutdown normal del proceso.
- Repositorio `users`.
- Campos `id`, `name`, `email`, `password_hash`, `role`, `created_at`, `updated_at`.
- Mapeo dominio/documento.

No implementado:

- Indices declarativos.
- Tests contra emulator.
- Health check con ping periodico.
- Transacciones para unicidad fuerte de email.

## Endpoints disponibles

| Metodo | Ruta | Protegido | Estado |
| --- | --- | --- | --- |
| `GET` | `/health` | No | Implementado. |
| `POST` | `/auth/register` | No | Implementado compatible Sprint 0.5. |
| `POST` | `/auth/login` | No | Implementado compatible Sprint 0.5. |
| `GET` | `/users/me` | Si | Implementado compatible Sprint 0.5. |

## Pruebas ejecutadas

Comando:

```sh
env GOCACHE=/Users/armandoalvarado/Documents/Sophia/sophia_ai_backend/sofia-backend/.cache/go-build go test ./...
```

Resultado:

```text
ok github.com/armandoalvarado/sofia-backend/internal/auth/infrastructure
ok github.com/armandoalvarado/sofia-backend/internal/config
ok github.com/armandoalvarado/sofia-backend/internal/server
```

Nota:

- La primera corrida requirio permisos elevados porque Go necesitaba escribir locks/checksums en el module cache global para compilar dependencias Firestore/OpenTelemetry.
- `go mod tidy` actualizo `go.mod`/`go.sum` con dependencias transitivas faltantes.

## Riesgos

- El archivo JSON de credenciales detectado debe salir del repo y posiblemente rotarse.
- La unicidad de email en Firestore se maneja por query previa; no es transaccional todavia.
- Firestore no tiene tests contra emulator en este sprint.
- JWT sigue sin refresh/revocacion.
- No hay graceful shutdown.

## Deuda tecnica

- Agregar readiness con verificacion activa de Firestore.
- Declarar indices Firestore.
- Implementar transaccion o documento indice por email para evitar carrera en duplicados.
- Agregar graceful shutdown con signals.
- Limpiar placeholders heredados `internal/app`, `internal/http`, `internal/domain/activity`.
- Agregar logger estructurado cuando el backend crezca.

## Recomendaciones para Sprint 2

- Implementar Users/Profile minimo o iniciar Tasks segun prioridad de producto.
- Si Sprint 2 toca datos persistentes, primero definir estrategia de IDs, ownership y timestamps.
- Agregar tests con Firestore emulator antes de depender de Firestore en flujos mas amplios.
- Introducir health/readiness separado antes de despliegue real.
