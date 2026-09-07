# Sprint 12 Report: Firestore Production Hardening

## Goal

Harden Firestore before connecting a real AI runtime or LLM provider.

No product endpoints, LLM integrations, notifications, or destructive migrations were added.

## What Changed

- Added Firestore emulator integration coverage in `internal/database/firestore_emulator_test.go`.
- Added cursor pagination support for activities, reminders, and memories.
- Added `next_cursor` to list responses for those modules.
- Added Firestore cursor queries ordered by `created_at desc, id desc`.
- Added transactional Firestore registration defaults for `users`, `user_profiles`, and `ai_settings`.
- Documented Firestore indexes, emulator usage, and transaction status.

## Pagination

The urgent high-volume list endpoints now accept an optional `cursor` query parameter:

- `GET /activities?limit=50&cursor=<activity_id>`
- `GET /reminders?limit=50&cursor=<reminder_id>`
- `GET /memory?limit=50&cursor=<memory_id>`

Responses include:

```json
{
  "items": [],
  "next_cursor": "last_seen_document_id"
}
```

This keeps the public API backward compatible because `cursor` and `next_cursor` are optional.

## Transactions

Implemented:

- Firestore registration transaction: auth user + profile + AI settings.

Documented but not yet implemented:

- Activity + reminder atomic creation.
- Action proposal execution + domain write atomic update.

Those two flows cross bounded contexts and should be implemented with a shared unit of work or narrow Firestore transaction methods. Adding partial transactions now would create misleading safety.

## Emulator Tests

Added an emulator suite that skips unless Firestore emulator configuration is present:

- `FIRESTORE_PROJECT_ID`
- `FIRESTORE_EMULATOR_HOST`
- `PERSISTENCE_DRIVER=firestore`

The suite uses synthetic documents only.

## Indexes

Required indexes are documented in `docs/firestore_indexes.md` for:

- activities
- reminders
- memories
- mood entries, outcomes, reflections
- ai action proposals

## Risks

- `activity + reminder` is still not atomic.
- AI proposal execution is still not atomic.
- Firestore uniqueness for email should move to a canonical unique-key document for stronger concurrency protection.
- Some range filters still happen in application code and should move closer to Firestore queries before large datasets.

## Tests

Executed:

```bash
go test ./...
```

The Firestore emulator test is included in the suite but skips when the emulator is not configured.
