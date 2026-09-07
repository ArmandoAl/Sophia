# Reminders Domain

Reminders represent scheduled user notifications and reminder records.

## Entity

Core fields:

- `id`
- `user_id`
- `activity_id`
- `title`
- `description`
- `status`
- `scheduled_at`
- `timezone`
- `next_run_at`
- `last_run_at`
- `delivered_at`
- `failed_at`
- `failure_reason`
- `retry_count`
- `max_retries`
- `processing_by`
- `processing_until`
- `last_attempt_at`
- `source`
- `created_by`
- `created_at`
- `updated_at`

Recurrence fields:

- `recurrence_rule`
- `recurrence_interval`
- `recurrence_until`
- `recurrence_count`
- `recurrence_occurrences`

## Status

- `pending`
- `due`
- `delivered`
- `failed`
- `canceled`
- `archived`

The worker scans `pending` reminders where `next_run_at <= now`.

## Worker Claim/Lease

The worker claims due reminders before delivery:

- `processing_by` identifies the worker that currently owns the attempt.
- `processing_until` is the lease expiration.
- `last_attempt_at` records when the current/last attempt started.

Only `pending` reminders due at `next_run_at <= now` are claimable. If `processing_until` is still in the future, another worker must skip the reminder. If the lease expires, another worker can reclaim it. Delivered or failed reminders clear the active lease.

## Recurrence

Supported simple recurrence:

- `none`
- `daily`
- `weekly`
- `monthly`

After successful delivery, recurring reminders stay `pending` with a future `next_run_at` unless `recurrence_count` or `recurrence_until` stops the recurrence.

Full RRULE/calendar behavior is intentionally out of scope.

## Security

- All HTTP reminder endpoints require JWT.
- `user_id` comes from the authenticated context.
- Users cannot access reminders owned by another user.
- Reminder descriptions are not sent through FCM delivery.
