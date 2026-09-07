# FCM Delivery

Sprint 21 adds Firebase Cloud Messaging as the first real notification provider for the reminder worker.

## Defaults

FCM is safe by default:

- `REMINDER_DELIVERY_PROVIDER=noop`
- `FCM_ENABLED=false`
- `FCM_DRY_RUN=true`

This means no real push notification is sent unless FCM is explicitly selected and enabled.

## Configuration

```sh
REMINDER_DELIVERY_PROVIDER=fcm
FCM_ENABLED=true
FCM_DRY_RUN=true
FCM_PROJECT_ID=your-firebase-project
GOOGLE_APPLICATION_CREDENTIALS=/path/to/service-account.json
```

Set `FCM_DRY_RUN=false` only after validating credentials, device token registration, and safety behavior.

## Payload

FCM payload is intentionally minimal:

```json
{
  "data": {
    "reminder_id": "...",
    "type": "reminder_due"
  },
  "notification": {
    "title": "Truncated reminder title"
  }
}
```

The reminder description is not sent. The title is truncated to 80 runes.

## Token Safety

- Full device tokens are stored server-side only.
- API responses return `token_redacted`.
- Logs/reports must never include complete FCM tokens.

## Worker Behavior

When `REMINDER_DELIVERY_PROVIDER=fcm` and `FCM_ENABLED=false`, the provider behaves as a safe no-op and does not fail reminders.

When `FCM_ENABLED=true` and `FCM_DRY_RUN=true`, the HTTP sender uses FCM `validate_only=true`. This validates credentials, project ID and device token shape/access without showing a notification.

When `FCM_ENABLED=true` and `FCM_DRY_RUN=false`, the HTTP FCM v1 sender uses Google Application Default Credentials.

The reminder worker records one `notification_delivery_attempts` document per processed reminder attempt. The attempt stores provider/status/dry-run/error metadata, not full device tokens or reminder descriptions.

## Safe Test Command

Use the internal CLI to exercise FCM safely:

```sh
go run ./cmd/tools/test-fcm-delivery -user-id=<user_id> -dry-run=true
```

or:

```sh
go run ./cmd/tools/test-fcm-delivery -device-token-id=<device_token_id> -dry-run=true
```

`-dry-run=true` is the default. The command writes `docs/fcm_test_report.md` and never prints full device tokens.

## Remaining Work

- Token invalidation cleanup based on FCM errors.
- User notification preferences.
- Provider-specific retry policy using classified errors.
