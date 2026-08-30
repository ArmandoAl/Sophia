# Sprint 26 Report - Reminder Worker Firestore Lease

Fecha: 2026-07-03

## Objetivo

Agregar claim/lease para que Reminder Worker pueda ejecutarse en multiples instancias contra Firestore sin entregar el mismo reminder en paralelo.

## Cambios

- `Reminder` agrega:
  - `processing_by`,
  - `processing_until`,
  - `last_attempt_at`.
- `ReminderRepository` agrega `ClaimDue`.
- Repositorio memory implementa claim atomico con mutex.
- Repositorio Firestore implementa claim con `RunTransaction`.
- Worker reclama cada reminder antes de llamar al delivery provider.
- `MarkDelivered` y `MarkFailed` limpian el lease activo.
- Config nueva:
  - `REMINDER_WORKER_ID`,
  - `REMINDER_WORKER_LEASE_DURATION`.

## Semantica

- Solo reminders `pending` y due son claimable.
- Si `processing_until` esta en el futuro, otro worker recibe skip.
- Si el lease expira, otro worker puede reclamar.
- Delivered/failed no se reclaman de nuevo.
- Provider failure deja el reminder `failed`, incrementa `retry_count`, conserva `last_attempt_at` y limpia lease activo.

## Archivos Principales

- `internal/reminders/domain/reminder.go`
- `internal/reminders/infrastructure/memory_repository.go`
- `internal/reminders/infrastructure/firestore_repository.go`
- `internal/reminders/worker/worker.go`
- `cmd/workers/reminders/main.go`
- `internal/config/config.go`

## Docs

- `docs/reminder_worker_lease.md`
- `docs/reminder_worker.md`
- `docs/reminders_domain.md`
- `docs/firestore_indexes.md`
- `docs/environment.md`

## Tests

Cobertura agregada:

- dos workers no procesan el mismo reminder,
- lease expirado permite retry,
- reminder delivered no se reclama otra vez,
- provider failure mantiene estado correcto y libera lease,
- config carga worker ID y lease duration.

Comando requerido:

```sh
env GOCACHE=/Users/armandoalvarado/Documents/Sophia/sophia_ai_backend/sofia-backend/.cache/go-build go test ./...
```

Resultado: PASS.

## Riesgos Restantes

- No hay heartbeat ni extension de lease para providers lentos.
- Retry scheduling sigue siendo basico.
- Se debe validar el comportamiento con Firestore emulator y dos procesos reales antes de produccion.
