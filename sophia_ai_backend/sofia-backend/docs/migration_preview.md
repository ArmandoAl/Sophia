# Migration Preview

Generated at: `2026-07-01T09:56:58Z`

- Dry run: `true`
- Apply requested: `false`
- Limit: `20`
- Target mode: `staging`

| Target collection | Document ID | Action | Warnings |
| --- | --- | --- | --- |
| `migration_activities_preview` | `activity_001` | `create` | legacy activity id will be reused; target-mode=staging writes normalized preview outside production activities; timezone defaulted to America/Tijuana; created_at is migration time because legacy activity has no created_at; user_id extracted from legacy DocumentRef; activity start_at is in the past; activity due_at is in the past |
| `ai_settings` | `user_001` | `create` | ai settings are inferred from legacy user preferences; memory/reminders/planning default to enabled; preferred_language defaulted to es |
| `migration_reminders_preview` | `reminder_001` | `create` | legacy reminder id will be reused; target-mode=staging writes normalized preview outside production reminders; timezone defaulted to America/Tijuana; created_at is migration time because legacy reminder has no created_at; user_id extracted from legacy DocumentRef; scheduled_at is in the past |
| `user_profiles` | `user_001` | `create` | legacy user id will be reused as modular user_id |
| `users` | `user_001` | `manual_review` | legacy user id will be reused as modular user_id; password_hash is not present in legacy user; do not overwrite auth users automatically |

## Conflicts

| Collection | Document ID | Severity | Message |
| --- | --- | --- | --- |
| `users` | `user_001` | `critical` | manual review target exists and must not be auto-applied |

## Warnings

- target-mode=staging; activities/reminders are written to migration preview collections
- preview regenerated offline from last validate-only snapshot because local environment is not configured for Firestore
