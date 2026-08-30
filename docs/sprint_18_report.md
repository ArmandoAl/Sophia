# Sprint 18 Report - Controlled Physical Delete

## Summary

Sprint 18 adds a controlled, verifiable physical delete executor for approved privacy delete requests. It does not delete automatically on request creation and does not add UI, LLM integration or hardware behavior.

## What Changed

- Added delete request statuses:
  - `pending`
  - `approved`
  - `processing`
  - `completed`
  - `failed`
  - `rejected`
- Extended `DeleteRequestRepository` with:
  - `ListByStatus`
  - `Update`
- Added `internal/privacy/application.DeleteExecutor`.
- Added Firestore user-data deleter scoped by `user_id`.
- Added CLI:
  - `cmd/tools/process-delete-requests`
- Added JSON report output:
  - `docs/delete_requests_report.json`

## Delete Scope

Deleted collections:

- `user_profiles`
- `ai_settings`
- `activities`
- `reminders`
- `mood_entries`
- `activity_outcomes`
- `reflections`
- `memories`
- `ai_action_proposals`

Excluded collections:

- `users`
- `audit_logs`
- legacy hardware/body/location data
- legacy AI/conversation/RAG collections

## Audit

The executor records audit logs for:

- `user_delete_processed`
- `user_delete_failed`

## Endpoint Decision

No admin endpoint was added. The backend does not yet have a secure admin auth pattern, so Sprint 18 keeps execution as an internal CLI.

## Tests Added

- pending requests are not processed,
- approved requests are processed,
- audit log is generated,
- another user's data is not deleted.

## Risks

- Firestore deletes are not yet claimed with a distributed lease; avoid concurrent executor runs.
- Auth user deletion still needs an explicit product/security decision.
- Audit log retention policy is still basic.
- Legacy collections remain intentionally out of scope.
