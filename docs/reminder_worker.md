# Reminder Worker Runtime

Sprint 17 adds a standalone reminder worker runtime without hardware integration or AI autonomy. Sprint 21 adds FCM as an optional delivery provider, but real push delivery remains disabled by default.

## Command

```sh
go run ./cmd/workers/reminders
```

The command exits immediately unless `REMINDER_WORKER_ENABLED=true`.

## Configuration

| Variable | Default | Description |
| --- | --- | --- |
| `REMINDER_WORKER_ENABLED` | `false` | Enables the standalone worker. |
| `REMINDER_WORKER_ID` | host-derived | Stable worker identity used in reminder claims. |
| `REMINDER_WORKER_INTERVAL` | `30s` | Polling interval for due reminders. |
| `REMINDER_WORKER_BATCH_SIZE` | `50` | Max reminders processed per run. |
| `REMINDER_WORKER_LEASE_DURATION` | `2m` | Claim lease duration for a processing attempt. |
| `REMINDER_DELIVERY_PROVIDER` | `noop` | `noop` or `fcm`. |

The worker reuses the existing persistence configuration:

- `PERSISTENCE_DRIVER=memory|firestore`
- `FIRESTORE_PROJECT_ID`
- `FIRESTORE_DATABASE_ID`
- `GOOGLE_APPLICATION_CREDENTIALS`
- `FIRESTORE_EMULATOR_HOST`

## Processing Model

The worker:

1. Lists due reminders with `status=pending` and `next_run_at <= now`.
2. Claims each candidate with `processing_by`, `processing_until` and `last_attempt_at`.
3. Skips stale reminders that are no longer due, pending or claimable.
4. Skips concurrent duplicate processing inside the same process.
5. Marks successful non-recurring deliveries as `delivered`.
6. For recurring reminders, advances `next_run_at` and keeps the reminder pending until count/until limits are reached.
7. Marks failed deliveries as `failed` and increments `retry_count`.
8. Marks reminders already at `max_retries` as failed without calling the delivery provider.
9. Records `notification_delivery_attempts` metadata for processed attempts.

## Idempotency

The current idempotency layer is conservative:

- due scans only include pending reminders,
- each reminder is claimed before delivery,
- Firestore claims are transactional,
- expired leases allow another worker to retry,
- delivered or failed reminders are no longer picked up by later scans,
- recurring reminders are not picked up again until `next_run_at` is advanced into the past,
- an in-process lock prevents the same reminder from being processed twice concurrently by one worker process.

## Operational Lease Validation

Sprint 30 adds a deterministic operational harness:

```sh
go run ./cmd/tools/test-reminder-worker-lease -output-dir=docs
```

It creates synthetic reminders and runs simulated workers against the same lease contract used by runtime code. The generated report is:

```text
docs/reminder_worker_operational_report.md
```

Validated scenarios:

- active lease blocks a second worker,
- expired lease allows retry,
- recurring reminder advances `next_run_at`,
- failed reminder is not delivered twice.

This harness is safe to run locally and in CI because it does not send notifications and does not require production credentials. It complements, but does not replace, emulator-backed validation of Firestore transactions.

Remaining production work:

- operational validation with multiple deployed worker instances,
- lease duration tuning against real provider latency,
- retry scheduling that keeps retryable reminders pending instead of marking all provider failures final,
- RRULE/exception-based recurrence expansion beyond the simple built-in rules.

## Firestore Index

The worker due scan needs:

`status asc, next_run_at asc`

The claim itself is performed by document ID in a Firestore transaction and does not need a separate index.

User-scoped due checks still use:

`user_id asc, status asc, next_run_at asc`

## Delivery

The worker supports:

- `noop`
- `fcm`

FCM remains disabled and dry-run by default.

Delivery attempts are recorded without full tokens, payload bodies or reminder descriptions.

## Limits

The worker still does not send push notifications by default. Real FCM delivery requires explicit FCM configuration.
