# Firestore Transactions

Sprint 12 starts hardening critical write paths. The key principle is that multi-document consistency belongs in Firestore infrastructure or an explicit unit-of-work boundary, not in domain use cases.

## Current Status

| Flow | Status | Notes |
| --- | --- | --- |
| Register + user_profile + ai_settings | Implemented for Firestore | `FirestoreUserRepository.SaveWithDefaults` writes `users`, `user_profiles`, and `ai_settings` in one Firestore transaction. The in-memory path keeps the existing post-register provisioning hook. |
| Activity + reminder | Not fully atomic yet | The current activity service writes the activity, then calls the reminder bridge. This can leave an activity without its reminder if the second write fails. |
| Action proposal execution + domain write | Not fully atomic yet | Execution currently performs the domain write first and updates the proposal afterward. A crash between writes can leave the proposal stale. |

## Implemented: Registration Defaults

When Firestore is the active repository, registration now uses a transactional repository capability:

- Checks duplicate email.
- Writes `users/{user_id}`.
- Writes `user_profiles/{user_id}` with default profile values.
- Writes `ai_settings/{user_id}` with default AI settings.

The composition root disables the legacy post-register hook for repositories that implement this transactional capability, preventing duplicate provisioning.

## Required Next Step: Shared Firestore Unit Of Work

The remaining flows cross module boundaries:

- `activities` calls `reminders` through a bridge.
- `ai/actions` can call activities, reminders, or memory.

Making those atomic requires one of these approaches:

1. Add a Firestore unit-of-work abstraction that can be passed through repositories participating in a single transaction.
2. Add narrow infrastructure-level methods for specific cross-document transactions, such as `CreateActivityWithReminder`.
3. Move execution into an outbox pattern where proposal execution is idempotent and eventually consistent.

The recommended path is:

- Short term: add narrow Firestore transaction methods for `CreateActivityWithReminder` and `ExecuteCreateActivityProposal`.
- Medium term: introduce an outbox for AI action execution and external side effects.
- Long term: make all autonomous execution idempotent with audit-log correlation IDs.

## Production Risks

- Activity reminder creation is still a consistency risk during partial outages.
- AI action execution can produce stale proposal state if the process exits after the domain write.
- Duplicate email protection still relies on query checking. A canonical email document or unique-key collection would be stronger under high concurrency.
