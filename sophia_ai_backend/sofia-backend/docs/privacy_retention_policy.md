# Privacy Retention Policy

This document describes the current backend retention behavior before LLM production usage.

## User-Initiated Delete Request

`POST /users/me/delete-request` creates a pending delete request. It does not delete data immediately.

Pending requests require explicit approval before the internal executor can process them.

## Physical Delete Scope

Approved requests also delete `v1_conversations` and `v1_conversation_messages`, scoped by `user_id`. Legacy `conversations` data remains retained.

Approved delete requests physically delete modular user data from:

- profile/settings collections,
- activities and reminders,
- insights,
- memories,
- AI action proposals.

Deletion is scoped by `user_id`.

## Retained By Default

The following are retained by default:

- `users` auth records,
- `audit_logs`,
- legacy hardware/location/body collections,
- legacy AI/conversation/RAG collections.

Reasons:

- auth deletion needs a separate credential/account decision,
- audit logs are operational/compliance evidence,
- legacy collections need reconciliation before destructive operations.

## Audit

The executor writes audit logs for:

- completed delete execution,
- failed delete execution.

Audit metadata is redacted/truncated using the existing privacy audit redaction path.

## Future Requirements

- Admin approval endpoint or operator workflow.
- Export-before-delete confirmation.
- Auth user deletion policy.
- Audit retention window.
- Legal hold support.
- User-facing status for delete requests.
