# Sprint 15 Report: Gemini Proposal-Only Provider

## Goal

Integrate Gemini as a real AI Runtime `ModelClient` while keeping Sofía proposal-only.

No automatic execution was added. No hardware, push notifications, or product-domain behavior was added.

## What Changed

- Added Gemini provider in `internal/ai/runtime/infrastructure/gemini`.
- Added AI provider config:
  - `AI_MODEL_PROVIDER=fake|gemini`
  - `GEMINI_API_KEY`
  - `GEMINI_MODEL`
  - `AI_RUNTIME_ENABLED`
  - `AI_RUNTIME_PROPOSAL_ONLY`
- Wired `cmd/sofia` to select fake or Gemini provider.
- Kept fake provider as default.
- Enforced proposal-only config; `AI_RUNTIME_PROPOSAL_ONLY=false` is rejected.
- Added Gemini structured-output parser.
- Added tests with fake HTTP transport, no real API calls.
- Added runtime test proving invalid model actions do not create proposals.

## Runtime Behavior

`POST /ai/runtime/message` continues to work with:

- fake provider,
- Gemini provider.

`dry_run=true`:

- builds context,
- calls selected model provider,
- validates model output,
- returns proposed actions,
- persists nothing.

`dry_run=false`:

- builds context,
- calls selected model provider,
- validates model output,
- creates safe `AIActionProposal` records,
- does not execute actions.

## Safety Controls

Gemini output is not trusted. It must pass:

- selected tool availability,
- policy engine,
- safety policy,
- JSON schema validation,
- ownership rules,
- proposal-only flow.

The prompt explicitly instructs Gemini not to include `user_id`, emails, tokens, secrets, passwords, or owner fields in `proposed_input`. The backend still validates this separately through schemas and safety policy.

## Privacy

The context sent to Gemini is compacted and redacted:

- user email is `[REDACTED]`,
- deleted memories are excluded,
- only active assistant-context memories are included,
- long text is truncated,
- no full memory corpus is sent.

## Tests

Executed:

```sh
go test ./...
```

Result: passing.

## Remaining Work

- Add retries with backoff for transient Gemini errors.
- Add model provider timeouts in config.
- Add eval fixtures for planning quality and unsafe proposal rejection.
- Review default `GEMINI_MODEL` before production.
- Add stronger PII redaction before increasing context limits.
