# Sprint 10 Report - Migration Preview Validation

Fecha: 2026-07-01

## Objetivo

Validar el preview de migracion Firestore legacy hacia modelo modular y preparar condiciones seguras para apply, sin aplicar cambios reales.

## Archivos revisados

- `docs/migration_preview.json`
- `docs/migration_preview.md`
- `docs/data_model_reconciliation.md`
- `docs/firestore_migration_plan.md`
- `docs/database_schema.md`
- `docs/architecture_state.md`

## Hallazgos

- `users/user_001` esta correctamente en `manual_review`.
- `user_profiles/user_001` y `ai_settings/user_001` son creates razonables, pero requieren conflict check.
- `activities/activity_001` y `reminders/reminder_001` usan merge sobre colecciones que tambien son legacy.
- Hay defaults inventados en settings, timezone y timestamps.
- Hay timezone mixto: perfil usa `America/Mexico_City`, activities/reminders usan `America/Tijuana`.
- Las fechas migradas de activity/reminder estan en mayo de 2026 y por tanto quedan en pasado.
- `user_id` se extrae desde DocumentRef legacy y debe validarse por documento.

## Cambios implementados

En `cmd/tools/migrate-firestore-legacy`:

- Se agrego `-validate-only`.
- Se agrego `-allow-manual-review`.
- Se agrego deteccion de conflictos contra documentos destino existentes.
- Apply queda bloqueado si hay `manual_review` sin flag explicito.
- Apply queda bloqueado si hay conflictos `high` o `critical`.

En `internal/platform/firestoremigration`:

- Se agrego estructura `Conflict`.
- El preview ahora puede incluir `conflicts`.
- Se agregaron warnings mas detallados.
- Se agrego `ValidateApplyAllowed`.

## Documentacion generada

- `docs/sprint_10_report.md`
- `docs/migration_validation.md`
- `docs/migration_apply_checklist.md`

## Tests

Se agregaron/actualizaron tests para:

- warnings en activity/reminder mappings,
- bloqueo por manual_review,
- bloqueo por conflicto high.

Comando ejecutado:

```sh
env GOCACHE=/Users/armandoalvarado/Documents/Sophia/sophia_ai_backend/sofia-backend/.cache/go-build go test ./...
```

Resultado: suite completa pasando.

## Decision

No aplicar migracion todavia.

Antes de cualquier apply real, ejecutar:

```sh
go run ./cmd/tools/migrate-firestore-legacy -validate-only=true -limit=20 -output-dir=docs
```

Y revisar `conflicts`, timezone, fechas pasadas y targets manual_review.
