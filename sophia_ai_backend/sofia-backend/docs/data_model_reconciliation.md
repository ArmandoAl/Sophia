# Data Model Reconciliation

Fecha: 2026-07-01

## Diagnostico

Firestore real contiene un modelo anterior/conceptual. El backend actual ya espera un modelo modular por bounded context.

Colecciones reales detectadas:

- `users`
- `activities`
- `reminders`
- `ai_memory`
- `ai_actions_log`
- `ai_policy`
- `bodies`
- `conversations`
- `daily_summaries`
- `locations`
- `message_embeddings`

Colecciones esperadas por el backend actual:

- `users`
- `user_profiles`
- `ai_settings`
- `activities`
- `reminders`
- `mood_entries`
- `activity_outcomes`
- `reflections`
- `memories`
- `tool_definitions`
- `ai_action_proposals`

## Clasificacion

| Coleccion real | Clasificacion | Motivo |
| --- | --- | --- |
| `users` | `migrate_now` | Existe, pero mezcla auth, perfil y settings AI legacy. Debe separarse en `users`, `user_profiles`, `ai_settings`. |
| `activities` | `migrate_now` | Existe con schema legacy parcial; debe normalizarse a Activity actual. |
| `reminders` | `migrate_now` | Existe con schema legacy parcial; debe normalizarse a Reminder actual. |
| `ai_memory` | `migrate_later` | Fuente candidata para `memories`, pero requiere reglas de privacidad/calidad antes de importar. |
| `ai_actions_log` | `migrate_later` | Puede alimentar auditoria o `ai_action_proposals`, pero no debe convertirse sin validar semantica. |
| `ai_policy` | `migrate_later` | Candidato para policy engine futuro. |
| `conversations` | `migrate_later` | Candidato para conversations/RAG futuro. No existe modulo runtime actual. |
| `message_embeddings` | `migrate_later` | Candidato para RAG/semantic memory futuro. No hay embeddings runtime. |
| `daily_summaries` | `ignore_for_now` | Podria alimentar insights/memory futuro, pero no hay entidad destino estable. |
| `bodies` | `legacy` | Relacionado con hardware/body runtime futuro. Fuera del backend actual. |
| `locations` | `legacy` | Relacionado con hardware/location context futuro. Fuera del backend actual. |

## Mapping

| Legacy | Nuevo destino | Estado |
| --- | --- | --- |
| `users` | `users` | manual review porque falta `password_hash`. |
| `users` | `user_profiles` | migracion segura por `-target-mode=modular-only`. |
| `users` | `ai_settings` | migracion segura por `-target-mode=modular-only`. |
| `activities` | `migration_activities_preview` por defecto; `activities` solo con `-target-mode=in-place` | staging preparado para evitar sobrescritura de coleccion legacy/destino. |
| `reminders` | `migration_reminders_preview` por defecto; `reminders` solo con `-target-mode=in-place` | staging preparado para evitar sobrescritura de coleccion legacy/destino. |
| `ai_memory` | `memories` | futuro, requiere privacy review. |
| `ai_actions_log` | audit/action proposals futuro | futuro. |
| `ai_policy` | policy engine futuro | futuro. |
| `conversations` | conversations/RAG futuro | futuro. |
| `message_embeddings` | vector/RAG futuro | futuro. |
| `bodies` | hardware futuro | no integrar ahora. |
| `locations` | hardware/location futuro | no integrar ahora. |

## Reglas de reconciliacion

- No borrar colecciones legacy.
- No migrar credenciales si no existe `password_hash`.
- No mezclar hardware en runtime actual.
- No importar memory/conversations hasta tener privacy/export/delete.
- No ejecutar migracion real sin `-apply=true` y `-dry-run=false`.
- Si se aplica algo, usar `-target-mode=modular-only` para limitar escritura a `user_profiles` y `ai_settings`.

## Riesgos

- `users` legacy no puede convertirse automaticamente en auth user completo.
- `activities.user_id` y `reminders.user_id` son referencias Firestore legacy; el migrador extrae el ID.
- `activities` y `reminders` ya no migran in-place por defecto; staging evita merge sobre colecciones productivas.
- `ai_memory` puede contener datos sensibles; requiere proceso dedicado.
