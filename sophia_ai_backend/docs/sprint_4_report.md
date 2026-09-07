# Sprint 4 Report - Reminder Engine

Fecha: 2026-06-30

## Auditoria inicial

Se revisaron:

- `docs/architecture_state.md`
- `docs/sprint_3_report.md`
- `docs/activities_domain.md`
- `docs/users_domain.md`
- Estado del codigo en `cmd/sofia`, `internal/server`, `internal/users` e `internal/activities`

Estado real antes del sprint:

- Auth con JWT estaba operativo.
- Users tenia profile y AI settings, incluyendo `reminders_enabled`.
- Activities existia como dominio modular con repositorios memory/Firestore.
- Activities tenia campos pasivos `has_reminder`, `reminder_at` y `recurrence_rule`.
- No existia modulo `internal/reminders`.
- No existia worker, endpoint due ni repositorio de reminders.

Riesgos detectados:

- La integracion Activity -> Reminder necesita transaccion en persistencia real para evitar creacion parcial.
- Firestore puede requerir indices adicionales para filtros compuestos.
- `recurrence_rule` todavia no tiene parser ni expansion.
- No hay infraestructura de notificaciones.

Decisiones:

- Crear `internal/reminders` como modulo vertical igual a users/activities.
- Mantener `ReminderRepository` como contrato de dominio.
- Implementar memory y Firestore sin contaminar use cases con Firestore.
- Conectar Activities con Reminders mediante una interfaz pequena en `activities/application`.
- No correr worker en background; exponer `ListDueReminders` y `GET /reminders/due`.
- Responder `409` cuando `AISettings.reminders_enabled=false`.

## Que cambio

- Se creo `internal/reminders/domain`.
- Se creo `internal/reminders/application`.
- Se creo `internal/reminders/infrastructure`.
- Se creo `internal/reminders/interfaces/http`.
- Se agrego entidad `Reminder`.
- Se agrego contrato `ReminderRepository`.
- Se implementaron repositorios:
  - `InMemoryReminderRepository`
  - `FirestoreReminderRepository`
- Se agregaron use cases:
  - `CreateReminder`
  - `GetReminder`
  - `ListReminders`
  - `UpdateReminder`
  - `CancelReminder`
  - `ArchiveReminder`
  - `ListDueReminders`
  - `MarkReminderDelivered`
  - `MarkReminderFailed`
- Se agrego integracion con Activities:
  - crear reminder asociado al crear activity con `has_reminder=true` y `reminder_at`
  - cancelar reminders pendientes al cancelar/archive activity
- Se conecto Reminders en `cmd/sofia/main.go`.
- Se agregaron rutas en `internal/server/routes.go`.
- Se agregaron tests HTTP end-to-end.

## Archivos modificados

- `sophia_ai_backend/sofia-backend/cmd/sofia/main.go`
- `sophia_ai_backend/sofia-backend/internal/server/routes.go`
- `sophia_ai_backend/sofia-backend/internal/server/routes_test.go`
- `sophia_ai_backend/sofia-backend/internal/activities/application/usecases.go`
- `sophia_ai_backend/sofia-backend/internal/activities/interfaces/http/handler.go`
- `sophia_ai_backend/sofia-backend/internal/reminders/domain/reminder.go`
- `sophia_ai_backend/sofia-backend/internal/reminders/application/usecases.go`
- `sophia_ai_backend/sofia-backend/internal/reminders/infrastructure/memory_repository.go`
- `sophia_ai_backend/sofia-backend/internal/reminders/infrastructure/firestore_repository.go`
- `sophia_ai_backend/sofia-backend/internal/reminders/interfaces/http/handler.go`
- `sophia_ai_backend/sofia-backend/internal/reminders/interfaces/http/requests.go`
- `sophia_ai_backend/sofia-backend/internal/reminders/interfaces/http/responses.go`
- `docs/architecture_state.md`
- `docs/activities_domain.md`
- `docs/reminders_domain.md`
- `docs/sprint_4_report.md`

## Endpoints

| Metodo | Ruta | Auth | Descripcion |
| --- | --- | --- | --- |
| `POST` | `/reminders` | Si | Crea reminder propio. |
| `GET` | `/reminders` | Si | Lista reminders propios. |
| `GET` | `/reminders/due` | Si | Lista reminders pendientes con `next_run_at <= now`. |
| `GET` | `/reminders/{id}` | Si | Lee reminder propio. |
| `PATCH` | `/reminders/{id}` | Si | Actualiza reminder propio. |
| `POST` | `/reminders/{id}/cancel` | Si | Cancela reminder propio. |
| `POST` | `/reminders/{id}/archive` | Si | Archiva reminder propio. |

## Variables de entorno

No se agregaron variables nuevas.

Se reutiliza la configuracion existente:

- `ENV`
- `PORT`
- `JWT_SECRET`
- `CORS_ALLOWED_ORIGINS`
- `REQUEST_BODY_LIMIT_BYTES`
- `PERSISTENCE_DRIVER`
- `FIRESTORE_PROJECT_ID`
- `GOOGLE_APPLICATION_CREDENTIALS`
- `FIRESTORE_EMULATOR_HOST`

## Pruebas ejecutadas

Comando:

```sh
env GOCACHE=/Users/armandoalvarado/Documents/Sophia/sophia_ai_backend/sofia-backend/.cache/go-build go test ./...
```

Resultado:

- Suite completa pasando.

Casos agregados/cubiertos:

- Crear reminder exitoso.
- Crear reminder sin title falla.
- Crear reminder con `scheduled_at` invalido falla.
- Listar solo reminders del usuario autenticado.
- No acceder reminder de otro usuario.
- Cancelar reminder.
- Archivar reminder.
- Listar due reminders.
- Crear activity con `has_reminder` crea reminder asociado.
- Cancelar activity cancela reminders pendientes asociados.
- `reminders_enabled=false` bloquea creacion.
- Regresion de auth/users/activities/health/CORS.

## Deuda tecnica

- Implementar transacciones o unidad de trabajo para Activity + Reminder.
- Agregar worker real con ciclo controlado, shutdown graceful y locks/idempotencia.
- Agregar notificaciones push/email/local cuando exista canal.
- Parsear y expandir `recurrence_rule`.
- Definir semantica de reminders al completar activity.
- Agregar paginacion por cursor.
- Agregar tests contra Firestore emulator.
- Definir indices Firestore para filtros compuestos.
- Validar timezone contra base IANA.

## Riesgos

- Sin transaccion, una falla al crear reminder asociado puede dejar una Activity creada sin reminder.
- El endpoint `/reminders/due` esta protegido por JWT de usuario; un worker de sistema futuro necesitara auth interna distinta.
- Firestore puede necesitar indices segun filtros reales.
- Recurrence sin expansion puede generar expectativas funcionales si clientes lo presentan como activo.

## Recomendaciones Sprint 5

- Implementar Reminder Worker real con locking/idempotencia.
- Definir canal de notificacion inicial.
- Decidir comportamiento de reminders al completar activities.
- Introducir tests de Firestore emulator para reminders.
- Preparar modelo de recurrence antes de Memory/Agents para evitar retrabajo.
