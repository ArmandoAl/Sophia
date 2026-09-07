# Sprint 3 Report - Activities Domain

Fecha: 2026-06-29

## Auditoria inicial

Se revisaron:

- `docs/sprint_0_5_report.md`
- `docs/sprint_1_report.md`
- `docs/sprint_2_report.md`
- `docs/architecture_state.md`
- Estado del grafo codebase-memory
- Implementacion de auth, users e infraestructura

Estado real antes del sprint:

- Auth con JWT estaba operativo.
- Users domain existia con profile y AI settings.
- Firestore bootstrap y repositorios memory/Firestore existian para auth/users.
- No existia modulo `internal/activities`.
- Existia un placeholder viejo `internal/domain/activity`, no usado por la arquitectura modular actual.
- No habia endpoints de planificacion.

Riesgos detectados:

- Firestore puede requerir indices compuestos para filtros de activities.
- No hay paginacion por cursor.
- No existe Reminder Engine; campos de reminder/recurrence debian quedar pasivos.
- El placeholder `internal/domain/activity` puede confundir en el futuro.

Decisiones necesarias:

- Usar Activity como concepto superior en lugar de Task.
- Crear modulo vertical `internal/activities`.
- Usar soft delete por `status=archived` y `archived_at`.
- Mantener ownership estricto por JWT/context.
- No implementar scheduler ni IA.

## Que cambio

- Se creo `internal/activities/domain`.
- Se creo `internal/activities/application`.
- Se creo `internal/activities/infrastructure`.
- Se creo `internal/activities/interfaces/http`.
- Se agrego entidad `Activity`.
- Se agrego `ActivityRepository`.
- Se implementaron repositorios memory y Firestore.
- Se agregaron use cases:
  - `CreateActivity`
  - `GetActivity`
  - `ListActivities`
  - `UpdateActivity`
  - `CompleteActivity`
  - `CancelActivity`
  - `ArchiveActivity`
- Se agregaron endpoints `/activities`.
- Se conecto Activities en `cmd/sofia` y `internal/server/routes.go`.
- Se agregaron tests end-to-end HTTP con JWT.

## Endpoints

| Metodo | Ruta | Auth | Descripcion |
| --- | --- | --- | --- |
| `POST` | `/activities` | Si | Crea Activity. |
| `GET` | `/activities` | Si | Lista Activities propias. |
| `GET` | `/activities/{id}` | Si | Lee Activity propia. |
| `PATCH` | `/activities/{id}` | Si | Actualiza Activity propia. |
| `POST` | `/activities/{id}/complete` | Si | Completa Activity. |
| `POST` | `/activities/{id}/cancel` | Si | Cancela Activity. |
| `POST` | `/activities/{id}/archive` | Si | Archiva Activity. |

## Filtros

`GET /activities` soporta:

- `status`
- `type`
- `priority`
- `from`
- `to`
- `tag`
- `limit`

`limit` default: `50`.

## Validaciones

- `title` requerido.
- `type` debe ser permitido.
- `status` debe ser permitido.
- `priority` debe ser permitido.
- `timezone` requerido.
- Fechas en RFC3339.
- `reminder_at` requiere `has_reminder=true`.
- `recurrence_rule` se guarda, no se ejecuta.
- Campos desconocidos se rechazan.

## Seguridad

- Todos los endpoints requieren JWT.
- `user_id` nunca se acepta en body.
- El user ID sale del token.
- Repositorios usan `user_id` en `FindByID`, `List`, `Complete`, `Cancel`, `Archive`.
- Intentar leer actividad de otro usuario devuelve `404`.

## Pruebas ejecutadas

Comando:

```sh
env GOCACHE=/Users/armandoalvarado/Documents/Sophia/sophia_ai_backend/sofia-backend/.cache/go-build go test ./...
```

Casos cubiertos:

- Crear activity exitosamente.
- Crear activity sin title falla.
- Crear activity con type invalido falla.
- Listar solo activities del usuario autenticado.
- Get activity existente.
- No acceder a activity de otro usuario.
- Update activity.
- Complete activity.
- Cancel activity.
- Archive activity.
- Filtros basicos por status/type/priority/tag.
- Regresion de auth/users/health/CORS.

## Deuda tecnica

- Agregar indices Firestore para filtros compuestos.
- Agregar paginacion por cursor.
- Agregar tests contra Firestore emulator.
- Decidir si archived debe excluirse por default de listados futuros.
- Validar timezone contra IANA.
- Limpiar placeholder heredado `internal/domain/activity`.

## Pendiente para Reminder Engine

Queda para otro sprint:

- Scheduler.
- Worker.
- Notificaciones.
- Reintentos.
- Expansion de `recurrence_rule`.
- Semantica completa de timezone para ejecucion.
- Integracion con `reminders_enabled` de AISettings.
