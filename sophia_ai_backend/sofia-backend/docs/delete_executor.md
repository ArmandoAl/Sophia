# Delete Executor

Sprint 18 adds a controlled physical delete executor for approved user deletion requests.

## Command

```sh
go run ./cmd/tools/process-delete-requests -limit=20 -output-dir=docs
```

The command requires:

- `PERSISTENCE_DRIVER=firestore`
- `FIRESTORE_PROJECT_ID`
- optional `FIRESTORE_DATABASE_ID`
- local credentials or emulator configuration as documented in `docs/environment.md`

It writes:

- `docs/delete_requests_report.json`

## Request Lifecycle

Delete request statuses:

- `pending`
- `approved`
- `processing`
- `completed`
- `failed`
- `rejected`

The public endpoint `POST /users/me/delete-request` still only creates `pending` requests. Sprint 18 does not add an admin HTTP endpoint because there is no secure admin auth pattern yet.

## Processing Rules

The executor only processes requests with status `approved`.

For each approved request:

1. Mark request as `processing`.
2. Delete documents where `user_id` matches the request user from known modular collections.
3. Register audit log.
4. Mark request as `completed`.
5. If any delete fails, mark request as `failed` and audit the failure.

## Target Collections

Deleted by default:

- `user_profiles`
- `ai_settings`
- `activities`
- `reminders`
- `mood_entries`
- `activity_outcomes`
- `reflections`
- `memories`
- `ai_action_proposals`

Not deleted by default:

- `users`
- `audit_logs`
- legacy hardware/location/body collections
- legacy/conceptual AI collections

## Safety Constraints

- No automatic deletion occurs when a user creates a delete request.
- Requests must be approved out of band before the CLI processes them.
- Deletes are scoped by `user_id`.
- Auth users are retained until there is an explicit account credential deletion policy.
- Audit logs are retained by default for compliance/debuggability.

## Production Gaps

- Admin approval workflow.
- Two-person approval for destructive operations.
- Firestore transaction/lease around request claiming for multi-operator runs.
- Export-before-delete package.
- Definitive auth user deletion strategy.
