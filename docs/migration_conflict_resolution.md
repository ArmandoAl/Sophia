# Migration Conflict Resolution

Fecha: 2026-07-01

## Conflictos originales

| Target | Severidad | Problema | Resolucion |
| --- | --- | --- | --- |
| `activities/activity_001` | medium | `merge` sobre coleccion legacy/destino podia sobrescribir campos normalizados. | Cambiado a staging por defecto: `migration_activities_preview/activity_001` con accion `create`. |
| `reminders/reminder_001` | medium | `merge` sobre coleccion legacy/destino podia sobrescribir campos normalizados. | Cambiado a staging por defecto: `migration_reminders_preview/reminder_001` con accion `create`. |
| `users/user_001` | critical | Falta `password_hash`; no se puede crear o sobrescribir auth user automaticamente. | Se mantiene `manual_review` y bloquea apply. |

## Estrategia nueva

El migrador ahora soporta:

```sh
-target-mode=staging|in-place
```

Default:

```sh
-target-mode=staging
```

En staging:

- legacy `activities` se normaliza hacia `migration_activities_preview`,
- legacy `reminders` se normaliza hacia `migration_reminders_preview`,
- no hay merge sobre colecciones productivas,
- el preview sirve para inspeccion humana y validacion de shape.

En in-place:

- `activities` conserva destino `activities` con `merge`,
- `reminders` conserva destino `reminders` con `merge`,
- debe solicitarse explicitamente,
- no debe usarse hasta resolver conflictos y backup.

## Estado despues de staging

Conflictos eliminados:

- `activities/activity_001`
- `reminders/reminder_001`

Conflicto restante:

- `users/user_001`: critical, manual_review.

Esto es intencional. El usuario legacy no debe escribirse automaticamente porque no tiene `password_hash`.

## Condiciones de apply

Apply sigue bloqueado si:

- existe un target `manual_review` critical sin revision explicita,
- existe conflicto `high` o `critical`,
- se usa `target-mode=in-place` sin haberlo solicitado explicitamente,
- `user_profiles` o `ai_settings` ya existen y aparecen como conflictos `high`.

## Comandos seguros

Validacion staging:

```sh
go run ./cmd/tools/migrate-firestore-legacy -validate-only=true -target-mode=staging -limit=20 -output-dir=docs
```

Preview in-place solo para comparar, sin apply:

```sh
go run ./cmd/tools/migrate-firestore-legacy -validate-only=true -target-mode=in-place -limit=20 -output-dir=docs
```

No ejecutar `-apply=true` todavia.
