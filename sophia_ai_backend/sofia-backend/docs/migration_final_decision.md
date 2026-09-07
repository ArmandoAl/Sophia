# Migration Final Decision

Fecha: 2026-07-01

## Decision

La fase de migracion legacy queda cerrada sin aplicar cambios peligrosos.

No se debe ejecutar migracion destructiva. No se debe sobrescribir `users`. No se deben migrar `activities` ni `reminders` legacy hacia colecciones productivas por defecto.

## Decision por target

| Target | Decision | Motivo |
| --- | --- | --- |
| `users/user_001` | Mantener `manual_review`; no escribir automaticamente. | Falta `password_hash`; sobrescribir auth romperia seguridad/autenticacion. |
| `user_profiles/user_001` | Candidato seguro para `modular-only` si el documento no existe. | Se deriva de legacy user sin credenciales. Debe validar existencia antes de apply. |
| `ai_settings/user_001` | Candidato seguro para `modular-only` si el documento no existe. | Defaults documentados; no toca credenciales. Debe validar existencia antes de apply. |
| `migration_activities_preview/activity_001` | Mantener como staging/inspeccion. No promover automaticamente. | Fechas en pasado, timezone default, `user_id` extraido de DocumentRef. |
| `migration_reminders_preview/reminder_001` | Mantener como staging/inspeccion. No promover automaticamente. | Fecha en pasado y podria afectar futuro worker de reminders. |

## Nuevo modo seguro

Se agrega:

```sh
-target-mode=modular-only
```

Este modo solo genera targets:

- `user_profiles`
- `ai_settings`

Excluye:

- `users`
- `activities`
- `reminders`
- `migration_activities_preview`
- `migration_reminders_preview`

## Estado de cierre

La migracion legacy queda dividida asi:

- Auth users: manual review permanente.
- Profile/settings: apply seguro posible con `modular-only`, despues de validate-only real.
- Activities/reminders: conservar en staging; decision futura manual.
- Legacy collections: no borrar.

## No hacer

- No correr `target-mode=in-place` para activities/reminders.
- No aplicar `users`.
- No borrar colecciones legacy.
- No activar workers/push/hardware ligados a datos migrados.
