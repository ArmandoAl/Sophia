# Firestore Indexes

Sprint 12 hardens the indexes needed by the current modular backend before adding a real LLM runtime.

These indexes assume the production repositories keep filtering by `user_id` first and use cursor pagination ordered by `created_at desc, id desc` for high-volume collections.

## Activities

Collection: `activities`

| Query | Composite index |
| --- | --- |
| List user activities | `user_id asc, created_at desc, id desc` |
| Filter by status | `user_id asc, status asc, created_at desc, id desc` |
| Filter by type | `user_id asc, type asc, created_at desc, id desc` |
| Filter by priority | `user_id asc, priority asc, created_at desc, id desc` |
| Filter by tag | `user_id asc, tags array, created_at desc, id desc` |

Notes:
- Mixed filters such as status + type + priority may require additional composite indexes once product usage settles.
- `from` and `to` are still applied in application filtering; moving them into Firestore range queries should be paired with explicit date indexes.

## Reminders

Collection: `reminders`

| Query | Composite index |
| --- | --- |
| List user reminders | `user_id asc, created_at desc, id desc` |
| Filter by status | `user_id asc, status asc, created_at desc, id desc` |
| Filter by activity | `user_id asc, activity_id asc, created_at desc, id desc` |
| Due reminders | `user_id asc, status asc, next_run_at asc` |
| Worker due scan | `status asc, next_run_at asc` |
| Worker due scan with lease visibility | `status asc, next_run_at asc` |

Notes:
- `ListDue` and worker due scans use Firestore range filtering on `next_run_at`.
- Worker claim/lease uses `reminders/{id}` inside a Firestore transaction after the candidate scan. `processing_by`, `processing_until` and `last_attempt_at` are written on the document but are not currently used as query predicates.
- Simple recurrence reuses `status + next_run_at`; RRULE/exception queries may need new indexes later.

## Memories

Collection: `memories`

| Query | Composite index |
| --- | --- |
| List active memories | `user_id asc, status asc, created_at desc, id desc` |
| Filter by type | `user_id asc, status asc, type asc, created_at desc, id desc` |
| Filter by importance | `user_id asc, status asc, importance asc, created_at desc, id desc` |
| Filter by tag | `user_id asc, status asc, tags array, created_at desc, id desc` |

Notes:
- Basic search remains textual and in-process. Semantic search/RAG will need a separate vector strategy.
- Deleted memories must remain excluded from list/search paths.

## Insights

Collections: `mood_entries`, `activity_outcomes`, `reflections`

| Query | Composite index |
| --- | --- |
| Mood by date | `user_id asc, occurred_at desc` |
| Mood by activity | `user_id asc, activity_id asc, occurred_at desc` |
| Outcomes by activity | `user_id asc, activity_id asc, created_at desc` |
| Reflections by date | `user_id asc, occurred_at desc` |
| Reflections by activity | `user_id asc, activity_id asc, occurred_at desc` |

Notes:
- Summary ranges should eventually use Firestore range predicates instead of scanning recent entries.

## AI Action Proposals

Collection: `ai_action_proposals`

| Query | Composite index |
| --- | --- |
| List user proposals | `user_id asc, created_at desc` |
| Filter by status | `user_id asc, status asc, created_at desc` |
| Audit by tool | `user_id asc, tool_name asc, created_at desc` |

Notes:
- Execution safety depends more on transactional writes and audit logs than indexes.
- Expiration sweeps will need `status asc, expires_at asc` once a worker exists.

## Notification Delivery Attempts

Collection: `notification_delivery_attempts`

| Query | Composite index |
| --- | --- |
| Attempts by user | `user_id asc, created_at desc` |
| Attempts by reminder | `reminder_id asc, created_at desc` |
| Attempts by provider/status | `provider asc, status asc, created_at desc` |

Notes:
- Current code only writes attempts. These indexes are expected for future operational dashboards and audits.
- Attempts intentionally exclude full device tokens and notification payloads.
# Conversation V1 indexes

- `v1_conversations`: `user_id ASC, updated_at DESC, id DESC`
- `v1_conversations`: `user_id ASC, status ASC, updated_at DESC, id DESC`
- `v1_conversation_messages`: `user_id ASC, conversation_id ASC, created_at ASC, id ASC`

Do not deploy automatically from the application.
