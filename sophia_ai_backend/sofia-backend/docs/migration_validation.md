# Migration Validation

Fecha: 2026-07-01

## Preview revisado

Archivos revisados:

- `docs/migration_preview.json`
- `docs/migration_preview.md`
- `docs/data_model_reconciliation.md`
- `docs/firestore_migration_plan.md`
- `docs/database_schema.md`
- `docs/architecture_state.md`

Preview generado:

- `generated_at`: `2026-07-01T07:34:27Z`
- `dry_run`: `true`
- `apply`: `false`
- `limit`: `20`

## Validacion por target

| Target | Accion | Validacion | Riesgo |
| --- | --- | --- | --- |
| `users/user_001` | `manual_review` | Correcto. No debe aplicarse automaticamente porque falta `password_hash`. | Critico si se sobrescribe auth user. |
| `user_profiles/user_001` | `create` | Correcto como destino modular. Usa `timezone=America/Mexico_City` de legacy. | Conflicto si ya existe perfil modular. |
| `ai_settings/user_001` | `create` | Correcto como destino modular inicial. | Defaults inventados: memory/reminders/planning enabled, preferred_language `es`. |
| `activities/activity_001` | `merge` | Parcialmente seguro solo con revision. Normaliza schema legacy. | Misma coleccion destino/legacy; merge puede sobrescribir campos. |
| `reminders/reminder_001` | `merge` | Parcialmente seguro solo con revision. Normaliza schema legacy. | Misma coleccion destino/legacy; fecha queda en pasado. |

## Riesgos detectados

### Sobrescritura de datos existentes

- `activities` y `reminders` son colecciones legacy y destino al mismo tiempo.
- `merge` no borra documentos, pero puede sobrescribir campos normalizados si ya hubo escritura del backend nuevo.
- `user_profiles` y `ai_settings` son `create`; si existen, apply debe bloquearse.

### Defaults inventados

- `ai_settings.memory_enabled=true`
- `ai_settings.reminders_enabled=true`
- `ai_settings.planning_enabled=true`
- `ai_settings.preferred_language=es`
- `activities.timezone=America/Tijuana`
- `reminders.timezone=America/Tijuana`
- `activities.created_at` y `reminders.created_at` usan tiempo de migracion porque legacy no trae esos campos.

### Timezone inconsistente

- `user_profiles` conserva `America/Mexico_City`.
- `activities` y `reminders` default a `America/Tijuana`.
- Antes de apply se debe decidir timezone canonico por usuario.

### Fechas antiguas

- `activities.start_at` y `due_at`: `2026-05-03T07:00:00Z`.
- `reminders.scheduled_at` y `next_run_at`: `2026-05-03T07:00:00Z`.
- Para el 1 de julio de 2026, esas fechas estan en el pasado.

### IDs legacy

- Se reutilizan `user_001`, `activity_001`, `reminder_001`.
- Esto facilita trazabilidad, pero puede colisionar con datos creados por el backend nuevo.

### `user_id` extraido desde DocumentRef

- El preview transforma refs legacy a `user_001`.
- Correcto como estrategia inicial, pero debe validarse para todos los documentos antes de apply.

## Cambios de seguridad agregados al migrador

- Flag `-validate-only`.
- Reporte de conflictos cuando el target ya existe.
- Warnings mas detallados por defaults, fechas pasadas, IDs legacy y `DocumentRef`.
- Apply bloqueado si existen targets `manual_review` y no se pasa flag explicito.
- Apply bloqueado si hay conflictos `high` o `critical`.

## Veredicto

No esta listo para apply en produccion.

Siguiente paso seguro: ejecutar `-validate-only` contra Firestore real para regenerar preview con conflictos de existencia antes de evaluar cualquier `-apply=true`.
