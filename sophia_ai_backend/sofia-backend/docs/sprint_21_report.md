# Sprint 21 Report - FCM Notification Provider

## Summary

Sprint 21 adds device token management and an optional FCM delivery provider for the reminder worker. No real notifications are sent by default.

## What Changed

- Added `internal/notifications` module:
  - domain,
  - application,
  - memory/Firestore repositories,
  - HTTP handlers.
- Added endpoints:
  - `POST /notifications/device-tokens`
  - `GET /notifications/device-tokens`
  - `DELETE /notifications/device-tokens/{id}`
- Added FCM reminder delivery provider:
  - dry-run safe,
  - injected sender for tests,
  - HTTP FCM v1 sender for real delivery.
- Added worker provider selection:
  - `REMINDER_DELIVERY_PROVIDER=noop|fcm`
- Added FCM config:
  - `FCM_ENABLED`
  - `FCM_DRY_RUN`
  - `FCM_PROJECT_ID`

## Safety

- FCM is disabled by default.
- Worker remains disabled by default.
- FCM dry-run is true by default.
- If FCM provider is selected while disabled, delivery is a no-op and reminders are not failed.
- Reminder descriptions are not sent.
- Reminder titles are truncated.
- Device tokens are redacted in API responses and tests.

## Tests

Covered:

- register device token,
- list only own device tokens,
- delete only own device token,
- FCM dry-run with injected sender,
- worker provider wiring,
- token redaction.

## Verification

```sh
env GOCACHE=/Users/armandoalvarado/Documents/Sophia/sophia_ai_backend/sofia-backend/.cache/go-build go test ./...
```

Result: passing.

## Remaining Work

- Delivery attempt audit logs.
- FCM error classification and token cleanup.
- User-level notification preferences.
- Firestore emulator coverage for `device_tokens`.
