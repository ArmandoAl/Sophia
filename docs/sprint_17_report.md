# Sprint 17 Report - Reminder Worker Runtime

## Summary

Sprint 17 adds a standalone reminder worker runtime for processing due reminders. It does not add push notifications, hardware integration or autonomous AI execution.

## What Changed

- Added `internal/reminders/worker` with:
  - `Worker`,
  - `ReminderRepository` worker contract,
  - `ReminderDeliveryProvider`,
  - `NoopDeliveryProvider`.
- Added `cmd/workers/reminders`.
- Added worker configuration:
  - `REMINDER_WORKER_ENABLED`,
  - `REMINDER_WORKER_INTERVAL`,
  - `REMINDER_WORKER_BATCH_SIZE`.
- Added global due scans to reminder repositories:
  - memory repository: `ListDueAll`,
  - Firestore repository: `ListDueAll`.
- Updated Firestore due queries to filter with `next_run_at <= now`.
- Updated docs for environment and Firestore indexes.

## Behavior

The worker scans pending reminders that are due, re-reads each reminder before processing, delivers through the configured provider and marks the reminder as `delivered` or `failed`.

The included delivery provider is noop, so no external notification is sent.

## Idempotency

Implemented safeguards:

- only pending due reminders are scanned,
- each reminder is reloaded before delivery,
- non-pending or no-longer-due reminders are skipped,
- in-process duplicate processing is blocked,
- completed reminders are not processed again on later runs.

Remaining gap:

- multi-instance idempotency still needs a Firestore claim/lease transaction.

## Tests Added

- processes due reminders,
- skips reminders that are not due,
- respects `max_retries`,
- marks failed when provider fails,
- basic idempotency across repeated runs,
- worker configuration loading and validation.

## Risks

- Provider failures currently mark reminders as `failed`; future retry scheduling should distinguish transient from permanent failures.
- Firestore worker scans require the `status asc, next_run_at asc` index.
- Multi-worker deployments need a distributed claim mechanism before production scaling.

## Documentation

- `docs/reminder_worker.md`
- `docs/reminder_delivery.md`
- `docs/sprint_17_report.md`
