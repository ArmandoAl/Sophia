# Sprint 14 Report: Migration Conflict Resolution

## Objetivo

Resolver conflictos del validate-only antes de cualquier apply real.

No se aplicaron cambios reales en Firestore. No se borraron colecciones legacy. No se tocaron endpoints publicos. No se integro LLM ni hardware.

## Hallazgos

El preview anterior tenia tres conflictos:

- `activities/activity_001`: medium, merge podia sobrescribir campos.
- `reminders/reminder_001`: medium, merge podia sobrescribir campos.
- `users/user_001`: critical, manual_review porque falta `password_hash`.

## Cambios implementados

- Agregado flag `-target-mode=staging|in-place`.
- Default cambiado a `staging`.
- `activities` legacy ahora apunta por defecto a `migration_activities_preview`.
- `reminders` legacy ahora apunta por defecto a `migration_reminders_preview`.
- `users` se mantiene como `manual_review`.
- `user_profiles` y `ai_settings` siguen preparandose como `create`.
- El comando genera tambien:
  - `docs/migration_preview_staging.json`
  - `docs/migration_preview_staging.md`
- Tests actualizados para staging default e in-place explicito.

## Preview staging

El preview staging evita conflictos sobre colecciones productivas:

| Target | Accion | Estado |
| --- | --- | --- |
| `migration_activities_preview/activity_001` | create | seguro para inspeccion staging |
| `migration_reminders_preview/reminder_001` | create | seguro para inspeccion staging |
| `user_profiles/user_001` | create | preparado, requiere validar existencia real |
| `ai_settings/user_001` | create | preparado, requiere validar existencia real |
| `users/user_001` | manual_review | bloqueado por falta de `password_hash` |

Conflicto restante:

- `users/user_001`: critical, esperado.

## Nota sobre regeneracion

Se intento ejecutar:

```sh
go run ./cmd/tools/migrate-firestore-legacy -validate-only=true -target-mode=staging -limit=20 -output-dir=docs
```

El entorno local respondio `PERSISTENCE_DRIVER="memory"`, por lo que no se conecto a Firestore ni se escribio nada real. Los archivos staging fueron regenerados offline desde el ultimo preview validate-only existente.

## Riesgos restantes

- `users/user_001` no puede migrarse automaticamente sin credenciales validas.
- Timezone sigue inconsistente: profile usa `America/Mexico_City`; activities/reminders default a `America/Tijuana`.
- Fechas legacy de activities/reminders estan en el pasado.
- Antes de cualquier apply debe ejecutarse validate-only real contra Firestore con `PERSISTENCE_DRIVER=firestore`.

## Tests

Ejecutado:

```sh
go test ./...
```

Resultado: suite completa pasando.
