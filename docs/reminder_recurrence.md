# Reminder Recurrence

Sprint 22 adds simple reminder recurrence. This is intentionally not a full RRULE/calendar engine.

## Supported Rules

`recurrence_rule` supports:

- `none`
- `daily`
- `weekly`
- `monthly`

Empty `recurrence_rule` is treated as `none`.

## Fields

| Field | Description |
| --- | --- |
| `recurrence_rule` | `none`, `daily`, `weekly`, `monthly`. |
| `recurrence_interval` | Positive interval for recurrent reminders. Defaults to `1` when recurrence is enabled. |
| `recurrence_until` | Optional RFC3339 stop boundary. |
| `recurrence_count` | Optional max number of delivered occurrences. |
| `recurrence_occurrences` | Count of successful delivered occurrences. |
| `next_run_at` | Next due time used by worker scans. |

## Worker Behavior

After successful delivery:

- non-recurring reminders become `delivered`,
- recurring reminders increment `recurrence_occurrences`,
- if count/until is reached, the reminder becomes `delivered`,
- otherwise the reminder remains `pending` and `next_run_at` moves forward.

This prevents duplicate notifications in repeated worker runs for the same due window because the reminder is no longer due after `next_run_at` moves to the future.

## Date Math

The implementation uses Go `time.AddDate`:

- daily: `AddDate(0, 0, interval)`
- weekly: `AddDate(0, 0, 7*interval)`
- monthly: `AddDate(0, interval, 0)`

Timezone is still required for reminders, but this sprint does not implement timezone-local DST-aware calendar rules.

## Validation

- `timezone` is required.
- recurrent reminders require positive `recurrence_interval`.
- `recurrence_count` cannot be negative.
- unsupported recurrence rules are rejected.

## Not Implemented

- arbitrary RRULE,
- business-day schedules,
- timezone-local DST adjustment rules,
- exceptions/skipped occurrences,
- multi-instance worker lease.
