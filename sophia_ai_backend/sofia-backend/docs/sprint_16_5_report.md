# Sprint 16.5 Report: Legacy Migration Closure

## Objetivo

Cerrar la fase de migracion legacy sin aplicar cambios peligrosos.

No se aplico migracion real. No se sobrescribio `users`. No se borraron colecciones legacy. No se integro hardware ni ejecucion automatica.

## Inputs revisados

- `docs/migration_preview_staging.md`
- `docs/migration_preview_staging.json`
- `docs/migration_conflict_resolution.md`
- `docs/migration_apply_checklist.md`
- `docs/data_model_reconciliation.md`
- `docs/architecture_state.md`

## Cambios implementados

- Agregado `-target-mode=modular-only`.
- `modular-only` solo genera:
  - `user_profiles`
  - `ai_settings`
- `modular-only` excluye:
  - `users`
  - `activities`
  - `reminders`
  - staging activities/reminders.
- Apply in-place de legacy `activities/reminders` queda bloqueado.
- Tests agregados para modular-only, manual_review, staging e in-place blocking.

## Decision final

| Target | Decision |
| --- | --- |
| `users/user_001` | Manual review permanente; no auto-apply. |
| `user_profiles/user_001` | Candidato a apply seguro con `modular-only` si no existe. |
| `ai_settings/user_001` | Candidato a apply seguro con `modular-only` si no existe. |
| `migration_activities_preview/activity_001` | Solo staging/inspeccion. |
| `migration_reminders_preview/reminder_001` | Solo staging/inspeccion. |

## Comando recomendado

Validate-only:

```sh
go run ./cmd/tools/migrate-firestore-legacy -validate-only=true -target-mode=modular-only -limit=20 -output-dir=docs
```

Apply minimo solo despues de backup y validate-only limpio:

```sh
go run ./cmd/tools/migrate-firestore-legacy -dry-run=false -apply=true -target-mode=modular-only -limit=20 -output-dir=docs
```

## Tests

Ejecutado:

```sh
go test ./...
```

Resultado: suite completa pasando.

## Riesgos restantes

- `user_profiles` y `ai_settings` pueden tener defaults inferidos.
- Hay que validar existencia real antes de apply.
- Activities/reminders legacy quedan pendientes de decision manual futura.
