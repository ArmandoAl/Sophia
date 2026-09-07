# Auditoria del backend de Sofia

Fecha de auditoria: 2026-06-28  
Rol: Senior Go Backend Engineer y Architecture Auditor  
Alcance: lectura del repositorio completo, con foco en `sophia_ai_backend/sofia-backend`. No se implementaron cambios de codigo.

## 1. Estructura actual del repositorio

El repositorio contiene tres bloques principales:

| Ruta | Responsabilidad detectada | Estado |
| --- | --- | --- |
| `sophia_ai/` | Aplicacion Flutter cliente. Contiene pantallas de chat, dashboard, reminders, smart home, settings e insight. | Implementacion visual/frontend en progreso. |
| `sophia_ai_backend/sofia-backend/` | Backend Go de Sofia. Contiene autenticacion basica con registro/login en memoria. | Backend inicial, no persistente. |
| `sophia_db/` | Documento HTML interactivo con propuesta visual de arquitectura de datos/colecciones. | Documento de diseno, no migraciones ejecutables. |

### Backend Go

Archivos importantes:

| Archivo | Responsabilidad |
| --- | --- |
| `sophia_ai_backend/sofia-backend/cmd/sofia/main.go` | Punto de entrada real del backend. Inicializa repositorio en memoria, casos de uso y handlers HTTP. Expone `/auth/register` y `/auth/login` en `:8080`. |
| `sophia_ai_backend/sofia-backend/go.mod` | Modulo Go `github.com/armandoalvarado/sofia-backend`, Go `1.24.5`. |
| `internal/auth/domain/user.go` | Entidad `User`. |
| `internal/auth/domain/repository.go` | Interfaz `UserRepository`. |
| `internal/auth/domain/errors.go` | Errores de dominio/auth. |
| `internal/auth/aplication/register_user.go` | Caso de uso de registro. Nota: el paquete esta escrito como `aplication`, con una sola `p`. |
| `internal/auth/aplication/login_user.go` | Caso de uso de login. |
| `internal/auth/infrastructure/memory_repository.go` | Repositorio en memoria para usuarios. |
| `internal/auth/interfaces/http/hanlders.go` | Handler HTTP de auth. Nota: el archivo esta escrito como `hanlders.go`. |
| `internal/auth/interfaces/http/requests.go` | DTOs de request de auth. |
| `internal/auth/interfaces/http/responses.go` | DTOs de response de auth. |
| `internal/app/app.go` | Placeholder vacio: solo declara `package app`. |
| `internal/config/config.go` | Placeholder vacio: solo declara `package config`. |
| `internal/http/server.go` | Placeholder vacio: solo declara `package http`. |
| `internal/http/routes.go` | Placeholder vacio: solo declara `package http`. |
| `internal/http/middleware.go` | Placeholder vacio: solo declara `package http`. |
| `internal/domain/activity/*.go` | Paquete `activity` vacio, sin entidades ni contratos implementados. |
| `sophia_ai_backend/sophia-ai-486821-b188dd536e22.json` | Archivo JSON de credenciales/proyecto Google/Firebase aparente. Riesgo: secreto versionado o local dentro del repo. |

### Frameworks y librerias usadas

Backend:

- HTTP: libreria estandar `net/http`.
- JSON: libreria estandar `encoding/json`.
- Password hashing: `golang.org/x/crypto/bcrypt`.
- IDs: `github.com/google/uuid`.
- Dependencias indirectas de Google Cloud/Firestore aparecen en `go.mod`, incluyendo `cloud.google.com/go/firestore`, pero no hay uso real en el codigo Go actual.
- No se detecta framework HTTP como Gin, Echo, Fiber o Chi.
- No se detecta ORM.
- No se detecta sistema de migraciones.

Frontend:

- Flutter/Dart, con estructura por features y cubits.

### Punto de entrada de la app

El backend entra por:

```text
sophia_ai_backend/sofia-backend/cmd/sofia/main.go
```

Comportamiento actual:

- Crea `InMemoryUserRepository`.
- Crea `RegisterUserUseCase` y `LoginUserUseCase`.
- Crea `AuthHandler`.
- Registra rutas con `http.HandleFunc`.
- Escucha en `:8080`.
- Ignora el error retornado por `http.ListenAndServe`.

### Configuracion actual

No existe configuracion funcional por ambiente.

Observaciones:

- `internal/config/config.go` esta vacio.
- Puerto `:8080` esta hardcodeado.
- No hay lectura de variables de entorno.
- No hay configuracion de base de datos.
- No hay configuracion de CORS, timeouts, logging, JWT/secrets ni Firebase/Firestore.
- Existe un JSON de credenciales Google/Firebase fuera del modulo Go, pero el backend no lo consume.

## 2. Arquitectura detectada

El backend intenta aproximarse a una arquitectura modular/limpia para `auth`, pero todavia esta en una fase muy inicial.

### Separacion por capas

Hay separacion parcial dentro de `internal/auth`:

| Capa | Ruta | Estado |
| --- | --- | --- |
| Dominio | `internal/auth/domain` | Existe: entidad `User`, repositorio, errores, interfaz `UserService`. |
| Aplicacion/casos de uso | `internal/auth/aplication` | Existe: registro y login. |
| Infraestructura | `internal/auth/infrastructure` | Existe: repositorio en memoria. |
| Interfaces HTTP | `internal/auth/interfaces/http` | Existe: handlers y DTOs. |

La separacion aun no esta consolidada a nivel aplicacion completa:

- `cmd/sofia/main.go` arma dependencias y rutas directamente.
- `internal/app`, `internal/config` e `internal/http` existen como placeholders, pero no participan.
- `internal/domain/activity` existe vacio y duplicaria conceptualmente el dominio modular de `internal/auth/domain`.

### Handlers/controllers

Si existen:

- `AuthHandler.Register`
- `AuthHandler.Login`

Problemas detectados:

- El paquete se llama `http`, lo cual puede confundirse con `net/http` en imports.
- Archivo `hanlders.go` tiene typo.
- No se setea `Content-Type: application/json`.
- No hay limite de tamano del body.
- No hay validaciones semanticas.
- No se responde body en `405 Method Not Allowed`.

### Services/usecases

Si existen casos de uso:

- `RegisterUserUseCase.Execute(name, email, password)`
- `LoginUserUseCase.Execute(email, password)`

Existe tambien una interfaz `UserService`, pero no se usa.

### Repositories

Si existe contrato:

- `UserRepository.Save(user *User) error`
- `UserRepository.FindByEmail(email string) (*User, error)`

Implementacion actual:

- `InMemoryUserRepository`, protegido con `sync.Mutex`, indexado por email.

Limitaciones:

- Los datos se pierden al reiniciar.
- No hay busqueda por ID.
- No hay actualizacion, borrado ni paginacion.
- No hay persistencia real.
- `FindByEmail` retorna `(nil, nil)` cuando no encuentra usuario, lo cual obliga a los casos de uso a distinguir ausencia por `nil`.

### Models/entities

Implementado:

- `auth/domain.User` con `ID`, `Name`, `Email`, `Password`, `Role`, `CreatedAt`, `UpdatedAt`.

No implementado:

- Tasks/activities.
- Reminders.
- Mood/outcomes.
- Memory.
- Agents.
- Tools.
- Conversations.
- Job queue.
- Presence logs.

### DTOs

Implementados:

- `RegisterRequest`
- `LoginRequest`
- `UserResponse`
- `ErrorResponse`

Limitaciones:

- No hay DTOs versionados.
- No hay DTO de token/sesion.
- `ErrorResponse` tiene campo Go `Error`, pero tag JSON `message`; el handler lo llena como `ErrorResponse{Error: ...}` y sale como `{"message":"..."}`.

### Middlewares

No hay middlewares funcionales.

Existe:

- `internal/http/middleware.go`, vacio.

Faltan:

- Logging.
- Recovery/panic handling.
- CORS.
- Request ID.
- Auth/JWT.
- Rate limiting basico para auth.
- Timeouts.
- Body size limit.

### Migraciones

No se detectan migraciones.

No existen:

- Directorio `migrations`.
- Archivos `.sql`.
- Herramientas como Goose, Atlas, Migrate, Ent, SQLC o GORM migrations.

### Configuracion de base de datos

No hay conexion funcional de base de datos.

Senales relevantes:

- `go.mod` incluye dependencias indirectas de Firestore/Google Cloud.
- `sophia_db/db.html` documenta una arquitectura NoSQL estilo colecciones.
- `sophia_ai_backend/sophia-ai-486821-b188dd536e22.json` sugiere intencion de usar Google/Firebase.
- El codigo actual usa exclusivamente memoria local.

## 3. Estado de base de datos

### Motor usado

Motor real en runtime: ninguno.

El backend actualmente usa:

```text
InMemoryUserRepository -> map[string]*domain.User
```

Motor aspiracional/documentado:

- `sophia_db/db.html` modela colecciones NoSQL, compatibles conceptualmente con Firestore.
- `go.mod` contiene dependencia indirecta `cloud.google.com/go/firestore`.

Conclusion: Firestore parece ser la direccion prevista, pero aun no esta integrado.

### Como se conecta

No hay conexion de base de datos implementada.

No existe:

- Cliente Firestore.
- Inicializacion con credentials.
- Configuracion de proyecto.
- Health check de DB.
- Manejo de contexto/timeouts.

### Tablas/modelos existentes

En codigo Go:

| Entidad | Estado |
| --- | --- |
| `User` | Implementada en dominio auth. |
| `Activity` | No implementada; paquete vacio. |

En documento `sophia_db/db.html`, se describen 11 colecciones:

| Coleccion | Proposito documentado |
| --- | --- |
| `users` | Perfil maestro del usuario y politicas de IA. |
| `locations` | Espacios fisicos monitorizados. |
| `bodies` | Nodos de hardware/IA en el hogar. |
| `activities` | Tareas, recurrencia y decisiones de IA. |
| `ai_memory` | Memoria episodica/semantica vectorizada. |
| `knowledge_files` | Archivos procesados por IA. |
| `conversations` | Historial multimodal usuario-IA. |
| `ai_actions_log` | Bitacora de razonamiento y ejecucion. |
| `tool_calls` | Detalle de llamadas a herramientas externas. |
| `job_queue` | Cola de tareas programadas y reintentos. |
| `presence_logs` | Sensores y deteccion de usuarios. |

Estas colecciones son documentacion, no schema ejecutable.

### Migraciones existentes

No existen migraciones reales.

Para Firestore, probablemente no se usarian migraciones SQL tradicionales, pero si convendria tener:

- Versionado de schema logico.
- Scripts de bootstrap.
- Indices Firestore declarados.
- Reglas de seguridad Firebase/Firestore.
- Seeds para desarrollo.

### Entidades ya implementadas

Implementada:

- Usuario basico para autenticacion.

Parcial:

- Auth register/login sin tokens.

No implementadas:

- Perfil IA del usuario.
- Preferencias/timezone.
- Tareas/activities.
- Reminders.
- Mood/outcomes.
- Memoria.
- Agentes.
- Herramientas internas.
- Planificacion.
- Conversaciones.
- Job queue.

### Que falta para soportar Sofia

Para usuarios:

- Persistencia real.
- Indices unicos por email.
- Password policy.
- Login con token/sesion.
- Refresh/revocation.
- Recuperacion de cuenta.
- Perfil y preferencias.
- Separar `password_hash` del JSON publico.

Para tareas:

- Entidad `Task` o `Activity`.
- Estados: pending, active, completed, canceled, snoozed.
- Prioridad, fechas, recurrencia, ownership por usuario.
- Repositorio persistente.
- Endpoints CRUD.

Para recordatorios:

- Entidad `Reminder`.
- Programacion `run_at`, timezone, recurrence.
- Worker/job queue.
- Estado de entrega y reintentos.

Para memoria:

- Entidad `Memory`.
- Tipos: episodic, semantic, preference, fact.
- Embeddings/vector store o estrategia Firestore + proveedor vectorial.
- Politicas de retencion/consentimiento.
- Endpoints de ingestion y busqueda.

Para agentes:

- Entidad `Agent`.
- Tool registry.
- Action log.
- Job queue.
- Trazabilidad de decisiones.
- Permisos por usuario.

## 4. APIs existentes

| Metodo HTTP | Ruta | Handler | Descripcion | Estado | Request esperado | Response esperado |
| --- | --- | --- | --- | --- | --- | --- |
| `POST` | `/auth/register` | `AuthHandler.Register` | Registra usuario en repositorio en memoria, hashea password con bcrypt y retorna usuario publico. | Implementado parcial | `{"name":"Armando","email":"a@example.com","password":"secret"}` | `201 {"id":"uuid","name":"Armando","email":"a@example.com"}` |
| Cualquier otro metodo | `/auth/register` | `AuthHandler.Register` | Rechaza metodos distintos de POST. | Implementado parcial | N/A | `405` sin body definido |
| `POST` | `/auth/login` | `AuthHandler.Login` | Busca usuario por email en memoria, compara password con bcrypt y retorna usuario publico. | Implementado parcial | `{"email":"a@example.com","password":"secret"}` | `200 {"id":"uuid","name":"Armando","email":"a@example.com"}` |
| Cualquier otro metodo | `/auth/login` | `AuthHandler.Login` | Rechaza metodos distintos de POST. | Implementado parcial | N/A | `405` sin body definido |

Errores comunes:

| Caso | Status actual | Body |
| --- | --- | --- |
| JSON invalido | `400` | `{"message":"Invalid request body"}` |
| Usuario duplicado | `409` | `{"message":"user already exists"}` |
| Credenciales invalidas segun handler | `401` | `{"message":"invalid credentials"}` |
| Password incorrecto | `500` actual, por desalineacion de errores | `{"message":"invalid password"}` |
| Usuario no encontrado | `500` actual, por desalineacion de errores | `{"message":"user not found"}` |

Nota critica: `handleError` solo mapea `ErrUserAlreadyExists` y `ErrInvalidCredentials`. Pero `LoginUserUseCase` retorna `ErrUserNotFound` o `ErrInvalidPassword`, por lo que login fallido termina como `500 Internal Server Error`. Esto deberia ser `401 Unauthorized` con error generico.

No existen endpoints para:

- Health check.
- Users profile.
- Tasks/activities.
- Reminders.
- Mood/outcomes.
- Memory.
- Agents.
- Tools.
- Conversations.

## 5. Calidad tecnica

### Organizacion del codigo

Fortalezas:

- Hay una intencion clara de separar dominio, aplicacion, infraestructura e interfaces en auth.
- El dominio no depende de HTTP.
- Los casos de uso dependen de una interfaz de repositorio, no de una implementacion concreta.
- Passwords se hashean con bcrypt.

Debilidades:

- La arquitectura solo esta aplicada a auth.
- Hay paquetes vacios que sugieren estructura futura, pero todavia no aportan.
- `main.go` concentra wiring, rutas y configuracion de servidor.
- Hay duplicidad conceptual entre `internal/auth/domain` e `internal/domain/activity`.
- El paquete `aplication` tiene typo.
- El archivo `hanlders.go` tiene typo.

### Consistencia de nombres

Problemas:

- `aplication` deberia ser `application`.
- `hanlders.go` deberia ser `handlers.go`.
- Mezcla de `internal/auth/domain` y `internal/domain/activity` genera dos estilos de modularizacion.
- `ErrInvalidCredentials`, `ErrInvalidPassword` y `ErrUserNotFound` compiten para representar fallos de login.
- `ai_memory` usa campo `userId` en el HTML, mientras otras colecciones usan `user_id`.

### Manejo de errores

Estado actual: basico e inconsistente.

Riesgos:

- Login fallido retorna `500` para usuario no encontrado/password invalido.
- `http.ListenAndServe` ignora errores.
- No hay logging de errores internos.
- No hay envoltura de errores ni diferenciacion clara entre errores de dominio, validacion e infraestructura.
- `RegisterUserUseCase` trata cualquier error de `FindByEmail` como error interno, pero el repositorio en memoria usa `(nil, nil)` para not found.

### Validaciones

No hay validaciones de negocio o entrada.

Faltan:

- Email requerido y formato.
- Password requerido, longitud minima y politica.
- Name requerido.
- Normalizacion de email.
- Rechazo de campos vacios.
- Limite de tamano de request.

### Configuracion por ambiente

No existe.

Faltan:

- `PORT`.
- `ENV`.
- `DATABASE_URL` o `FIRESTORE_PROJECT_ID`.
- Ruta/uso seguro de credenciales.
- CORS origins.
- JWT secret/private key.
- Log level.
- Timeouts.

### Seguridad basica

Fortalezas:

- Passwords no se guardan en claro; se usa bcrypt.
- `UserResponse` no expone password.

Riesgos:

- No hay tokens ni sesiones, por lo que login no autentica llamadas futuras.
- No hay TLS/configuracion de reverse proxy documentada.
- No hay rate limit en auth.
- No hay CORS.
- No hay CSRF strategy si se usan cookies en el futuro.
- No hay validacion de input.
- No hay proteccion contra enumeracion de usuarios: errores de login distinguen usuario inexistente/password invalido internamente y hoy salen como 500.
- Archivo de credenciales Google parece estar dentro del repo.
- Entidad `User` serializa `Password` con tag JSON `password`; aunque `UserResponse` no lo expone, es un footgun.

### Testabilidad

Fortalezas:

- Los casos de uso son testeables por depender de `UserRepository`.
- El repositorio en memoria puede servir como fake inicial.

Debilidades:

- No hay tests.
- `main.go` no esta separado en una funcion de construccion del servidor.
- Handlers no comparten helpers de JSON/errores.
- No hay interfaces para clock/ID generation; `time.Now()` y `uuid.New()` estan hardcodeados.
- No hay configuracion inyectable.

### Riesgos actuales

| Riesgo | Severidad | Detalle |
| --- | --- | --- |
| Datos no persistentes | Alta | Usuarios desaparecen al reiniciar proceso. |
| Credenciales Google en repo | Alta | Posible secreto sensible versionado/local. Debe revisarse inmediatamente. |
| Login fallido como 500 | Alta | Mala semantica API, ruido operacional y potencial fuga de comportamiento. |
| Sin autenticacion real post-login | Alta | No hay JWT/sesion; no se pueden proteger recursos. |
| Sin validaciones | Alta | Se aceptan usuarios con campos vacios o emails invalidos. |
| Sin configuracion por ambiente | Media | Dificulta despliegue, testing y seguridad. |
| Sin DB ni migraciones/schema operativo | Alta | No soporta tareas, memoria, reminders ni agentes. |
| Estructura futura ambigua | Media | Paquetes vacios pueden consolidarse mal si no se define convencion ahora. |
| Sin tests | Media | Cambios futuros pueden romper auth silenciosamente. |

## 6. Recomendacion arquitectonica

Recomiendo evolucionar hacia una arquitectura limpia y modular, pero sin hacer un refactor grande inmediato. El codigo actual ya contiene una semilla buena en `internal/auth`; conviene formalizarla y extenderla dominio por dominio.

### Principios objetivo

- `cmd/api` solo arranca la app.
- `internal/config` carga configuracion.
- `internal/server` construye servidor HTTP con timeouts y rutas.
- Cada dominio tiene entidad, repositorio, casos de uso y handlers.
- Infraestructura implementa contratos, no invade dominio.
- HTTP DTOs no son entidades de dominio.
- Firestore o el motor elegido vive detras de repositorios.
- Auth produce identidad verificable para proteger endpoints.

### Estructura objetivo adaptada

Propuesta:

```text
/cmd/api
/internal/config
/internal/server
/internal/platform/httpjson
/internal/platform/clock
/internal/auth
  /domain
  /application
  /infrastructure
  /interfaces/http
/internal/users
  /domain
  /application
  /infrastructure/firestore
  /interfaces/http
/internal/tasks
  /domain
  /application
  /infrastructure/firestore
  /interfaces/http
/internal/reminders
  /domain
  /application
  /infrastructure/firestore
  /interfaces/http
/internal/mood
  /domain
  /application
  /infrastructure/firestore
  /interfaces/http
/internal/memory
  /domain
  /application
  /infrastructure/firestore
  /interfaces/http
/internal/agents
  /domain
  /application
  /infrastructure
  /interfaces/http
/internal/tools
/internal/database
/docs
/tests
```

Comparado con la estructura sugerida originalmente:

- Mantendria modulos verticales (`auth`, `tasks`, `memory`) porque el repo ya inicio asi y escala mejor para Sofia.
- Usaria `internal/database` solo para clientes compartidos, transacciones y bootstrap.
- Evitaria un `/internal/repositories` global al inicio; puede volverse un cajon generico. Mejor repositorios por modulo.
- `internal/services` global no seria necesario todavia; los casos de uso en `application` son mas explicitos.
- `internal/ai`, `internal/agents`, `internal/memory` si merecen modulos propios por el producto Sofia.

### Camino incremental recomendado

1. Corregir errores funcionales de auth sin mover todo.
2. Crear `internal/config` real.
3. Extraer servidor/rutas desde `main.go`.
4. Definir persistencia Firestore o alternativa antes de crear mas dominios.
5. Implementar dominios nuevos con el mismo patron modular.

No recomiendo migrar todo a una arquitectura nueva en un solo PR. El backend es pequeno; el valor esta en establecer convenciones y crecer de forma disciplinada.

## 7. Roadmap tecnico: 5 sprints pequenos

### Sprint 1: Users/Auth persistente y seguro

Objetivo:

- Convertir auth actual en una base usable: validacion, persistencia, errores correctos y token/sesion inicial.

Archivos que tocaria:

- `cmd/sofia/main.go` o futuro `cmd/api/main.go`
- `internal/config/config.go`
- `internal/auth/domain/user.go`
- `internal/auth/domain/errors.go`
- `internal/auth/aplication/register_user.go`
- `internal/auth/aplication/login_user.go`
- `internal/auth/interfaces/http/hanlders.go`
- `internal/auth/interfaces/http/requests.go`
- `internal/auth/interfaces/http/responses.go`
- `internal/auth/infrastructure/*`
- `internal/database/*`

Entidades necesarias:

- `User`
- `Session` o `AuthToken` si se decide persistir sesiones

Endpoints:

- `POST /auth/register`
- `POST /auth/login`
- `GET /users/me`

Pruebas sugeridas:

- Registro exitoso.
- Email duplicado.
- Password debil.
- Login exitoso.
- Login con password incorrecto retorna `401`.
- Login con usuario inexistente retorna `401` generico.
- `GET /users/me` requiere auth.

Riesgos:

- Elegir mal estrategia de auth puede bloquear el resto.
- Si Firestore se integra sin abstraccion clara, contaminara casos de uso.
- Credenciales/secrets deben salir del repo y moverse a entorno seguro.

### Sprint 2: Tasks/Activities basicas

Objetivo:

- Implementar tareas personales como primer dominio productivo de Sofia.

Archivos que tocaria:

- `internal/tasks/domain/task.go`
- `internal/tasks/domain/repository.go`
- `internal/tasks/application/create_task.go`
- `internal/tasks/application/list_tasks.go`
- `internal/tasks/application/update_task.go`
- `internal/tasks/application/complete_task.go`
- `internal/tasks/infrastructure/firestore/task_repository.go`
- `internal/tasks/interfaces/http/handlers.go`
- `internal/tasks/interfaces/http/requests.go`
- `internal/tasks/interfaces/http/responses.go`
- `internal/server/routes.go`

Entidades necesarias:

- `Task`
- `TaskStatus`
- `TaskPriority`

Endpoints:

- `POST /tasks`
- `GET /tasks`
- `GET /tasks/{id}`
- `PATCH /tasks/{id}`
- `POST /tasks/{id}/complete`
- `DELETE /tasks/{id}`

Pruebas sugeridas:

- Crear tarea con usuario autenticado.
- Listar solo tareas del usuario.
- Validar titulo requerido.
- Completar tarea.
- No permitir acceso cruzado entre usuarios.

Riesgos:

- Confundir `activities` del schema HTML con `tasks`. Conviene decidir nombre canonico.
- Recurrencia no debe entrar completa en este sprint; solo campos preparados.

### Sprint 3: Reminders y scheduler minimo

Objetivo:

- Agregar recordatorios programables asociados a tareas o independientes.

Archivos que tocaria:

- `internal/reminders/domain/reminder.go`
- `internal/reminders/domain/repository.go`
- `internal/reminders/application/create_reminder.go`
- `internal/reminders/application/list_reminders.go`
- `internal/reminders/application/mark_sent.go`
- `internal/reminders/infrastructure/firestore/reminder_repository.go`
- `internal/reminders/interfaces/http/handlers.go`
- `internal/reminders/interfaces/http/requests.go`
- `internal/reminders/interfaces/http/responses.go`
- `internal/jobs` o `internal/reminders/worker`
- `internal/server/routes.go`

Entidades necesarias:

- `Reminder`
- `ReminderStatus`
- `RecurrenceRule` basica
- `Job` si se implementa cola minima

Endpoints:

- `POST /reminders`
- `GET /reminders`
- `GET /reminders/{id}`
- `PATCH /reminders/{id}`
- `DELETE /reminders/{id}`

Pruebas sugeridas:

- Crear reminder con `run_at`.
- Rechazar fechas invalidas.
- Respetar timezone del usuario.
- Listar reminders pendientes.
- Worker marca reminder como enviado o fallido.

Riesgos:

- Timezones y recurrencia pueden crecer rapido.
- Un scheduler en memoria no sirve para produccion si hay multiples instancias.
- Firestore requiere estrategia para consultar por `run_at` y `status` con indices.

### Sprint 4: Mood/outcomes

Objetivo:

- Registrar estado emocional, resultados y feedback del usuario para alimentar planificacion y memoria.

Archivos que tocaria:

- `internal/mood/domain/mood_entry.go`
- `internal/mood/domain/outcome.go`
- `internal/mood/domain/repository.go`
- `internal/mood/application/create_mood_entry.go`
- `internal/mood/application/create_outcome.go`
- `internal/mood/application/list_entries.go`
- `internal/mood/infrastructure/firestore/repository.go`
- `internal/mood/interfaces/http/handlers.go`
- `internal/mood/interfaces/http/requests.go`
- `internal/mood/interfaces/http/responses.go`
- `internal/server/routes.go`

Entidades necesarias:

- `MoodEntry`
- `Outcome`
- `EnergyLevel`
- `Sentiment`

Endpoints:

- `POST /mood`
- `GET /mood`
- `POST /outcomes`
- `GET /outcomes`
- `GET /outcomes/summary`

Pruebas sugeridas:

- Crear mood entry autenticado.
- Validar rangos numericos.
- Asociar outcome a task opcional.
- Consultar resumen por rango de fechas.
- Aislamiento por usuario.

Riesgos:

- Datos sensibles: requiere consentimiento, retencion y controles de privacidad.
- Nombres de campos deben alinearse con memoria futura para no duplicar conceptos.

### Sprint 5: Memory inicial

Objetivo:

- Implementar memoria basica consultable para Sofia, empezando sin sobreconstruir vector search si aun no es necesario.

Archivos que tocaria:

- `internal/memory/domain/memory.go`
- `internal/memory/domain/repository.go`
- `internal/memory/application/create_memory.go`
- `internal/memory/application/search_memory.go`
- `internal/memory/application/link_memory.go`
- `internal/memory/infrastructure/firestore/memory_repository.go`
- `internal/memory/interfaces/http/handlers.go`
- `internal/memory/interfaces/http/requests.go`
- `internal/memory/interfaces/http/responses.go`
- `internal/ai/embeddings/*` si se decide integrar embeddings
- `internal/server/routes.go`

Entidades necesarias:

- `Memory`
- `MemoryType`
- `MemorySource`
- `MemoryConfidence`
- `MemoryLink`

Endpoints:

- `POST /memory`
- `GET /memory`
- `GET /memory/{id}`
- `POST /memory/search`
- `DELETE /memory/{id}`

Pruebas sugeridas:

- Crear memoria con usuario autenticado.
- Buscar por texto basico.
- Filtrar por tipo/source.
- Borrar memoria.
- Validar que un usuario no vea memoria de otro.

Riesgos:

- Embeddings/vector store puede agregar complejidad prematura.
- Memoria contiene datos muy sensibles; requiere politicas de consentimiento y borrado desde el inicio.
- Si se mezclan conversaciones, knowledge files y memories sin limites claros, el modelo de datos se vuelve difuso.

## Conclusion

El backend de Sofia esta en etapa fundacional. Ya existe una buena semilla para auth con separacion parcial por dominio, casos de uso, repositorio e interfaz HTTP. Sin embargo, actualmente no hay persistencia real, configuracion, middlewares, tests, migraciones/schema operativo ni APIs para las capacidades centrales de Sofia.

La recomendacion principal es no hacer un gran refactor todavia. Primero hay que estabilizar auth, definir persistencia y convenciones de arquitectura. Despues conviene crecer por modulos verticales: users, tasks, reminders, mood/outcomes y memory. Esa ruta permite que Sofia avance hacia una IA personal/autonoma sin convertir el backend temprano en una arquitectura grande pero vacia.
