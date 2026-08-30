# Sprint 13 Report: AI Safety Runtime Hardening

## Goal

Strengthen AI Runtime safety before connecting Gemini, OpenAI, or another real model provider.

No real LLM integration was added. No automatic execution was added. Sofía remains proposal-only.

## What Changed

- Added local JSON Schema validation in `internal/platform/jsonschema`.
- Validated `ToolDefinition.input_schema` in runtime safety checks.
- Validated `ToolDefinition.input_schema` before creating AI action proposals.
- Hardened default tool schemas with `additionalProperties:false` and enums.
- Added an explicit runtime policy engine.
- Added executor registry for AI action execution.
- Added audit logging for AI runtime and action proposal lifecycle events.
- Added approximate context token budget.
- Improved context memory filtering by visibility, status, importance, and recency.
- Redacted email from runtime context summary.

## Audit Events

New audit actions:

- `ai_runtime_context_built`
- `ai_runtime_model_request_planned`
- `ai_runtime_proposals_generated`
- `ai_action_proposal_generated`
- `ai_action_proposal_confirmed`
- `ai_action_proposal_rejected`
- `ai_action_proposal_executed`
- `ai_action_proposal_failed`

Audit metadata is routed through the existing privacy audit service, which redacts sensitive metadata.

## Executor Registry

`ai/actions` no longer relies on a growing execution switch as the main extension point.

Current registered executors:

- `create_activity`
- `create_reminder`
- `create_memory`

Execution is still explicit through the existing action proposal endpoint and does not happen automatically from AI Runtime.

## Policy Status

Implemented:

- autonomy-level gating,
- feature flag gating from AI settings,
- risk/confirmation checks,
- tool schema validation,
- proposal-only runtime behavior.

Still pending:

- persisted policy decisions,
- policy denial reason codes,
- model-provider specific safety filters,
- stronger DLP/redaction.

## Tests

Added tests for:

- JSON schema validation,
- rejected missing/extra/wrong-type schema inputs,
- AI action lifecycle audit logs.

Executed:

```bash
go test ./...
```

Result: passing.

## Remaining Risks Before LLM

- The schema validator supports only the subset Sofía currently needs.
- Existing persisted tool definitions may need reseeding or migration to receive hardened schemas.
- Context redaction is basic and should not be treated as complete PII detection.
- Tool risk is still derived in code instead of stored as tool metadata.
