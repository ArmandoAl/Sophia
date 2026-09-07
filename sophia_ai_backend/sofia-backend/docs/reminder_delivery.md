# Reminder Delivery

Sprint 17 introduces a delivery abstraction without real notification delivery.

## Interface

`internal/reminders/worker.ReminderDeliveryProvider` exposes:

```go
Deliver(ctx context.Context, reminder *domain.Reminder) error
```

The worker owns reminder state transitions. Providers only attempt delivery and return success or failure.

## Noop Provider

`NoopDeliveryProvider` is the initial implementation.

Behavior:

- returns success,
- does not call external services,
- does not send push, email, SMS or hardware commands,
- allows the worker lifecycle and idempotency paths to be tested safely.

## FCM Provider

Sprint 21 adds FCM as the first real provider, but it is disabled by default.

Safe defaults:

- `REMINDER_DELIVERY_PROVIDER=noop`
- `FCM_ENABLED=false`
- `FCM_DRY_RUN=true`

The FCM payload includes only:

- `reminder_id`
- truncated title
- `type=reminder_due`

Reminder descriptions and full device tokens are never included in delivery logs/reports.

## Future Providers

Future delivery providers may include:

- APNs,
- email,
- webhook integrations,
- in-app notification queues.

Before adding real providers, the backend needs:

- user notification preferences,
- device tokens,
- opt-in/opt-out controls,
- provider-specific retry classification,
- audit logging for delivery attempts,
- redaction of notification payloads in logs.

## Safety Notes

Providers must not log reminder descriptions, tokens or personal content in full. Failure reasons should be short and operational.
