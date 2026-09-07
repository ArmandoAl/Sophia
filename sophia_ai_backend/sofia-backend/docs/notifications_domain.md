# Notifications Domain

Sprint 21 adds a minimal notifications module for device tokens. It does not send push notifications by itself.

## Module

```text
internal/notifications/domain
internal/notifications/application
internal/notifications/infrastructure
internal/notifications/interfaces/http
```

## Entity

`DeviceToken`

| Field | Description |
| --- | --- |
| `id` | Device token record ID. |
| `user_id` | Owner from JWT context. |
| `platform` | `ios`, `android` or `web`. |
| `token` | Full FCM token, stored server-side only. |
| `enabled` | Whether token can be used for delivery. |
| `created_at` | Creation timestamp. |
| `updated_at` | Last update timestamp. |
| `last_seen_at` | Last registration/seen timestamp. |

## Endpoints

All endpoints require JWT.

### `POST /notifications/device-tokens`

Request:

```json
{
  "platform": "ios",
  "token": "fcm-device-token"
}
```

Response returns redacted token only:

```json
{
  "id": "...",
  "user_id": "...",
  "platform": "ios",
  "token_redacted": "abcd...7890",
  "enabled": true,
  "created_at": "...",
  "updated_at": "...",
  "last_seen_at": "..."
}
```

### `GET /notifications/device-tokens`

Lists only the authenticated user's enabled device tokens. Tokens are redacted.

### `DELETE /notifications/device-tokens/{id}`

Deletes only a token owned by the authenticated user.

## Security

- `user_id` is never accepted from request body.
- Full tokens are not returned by API responses.
- Cross-user delete returns not found.
- FCM delivery uses token records only server-side.

## Storage

Collection:

`device_tokens`

Expected indexes:

- `user_id asc, enabled asc`

Delivery attempts are stored separately in:

`notification_delivery_attempts`

See `docs/notification_delivery_attempts.md`.
