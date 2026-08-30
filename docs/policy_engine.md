# AI Runtime Policy Engine

Sprint 13 introduces an explicit policy engine before connecting Gemini, OpenAI, or another model provider.

## Goals

- Keep Sofía proposal-only.
- Prevent automatic execution.
- Gate tools by user settings and autonomy.
- Validate tool input before creating proposals.
- Keep policy decisions outside model prompts.

## Inputs

Policy decisions use:

- authenticated `user_id`
- `AISettings.autonomy_level`
- `AISettings.memory_enabled`
- `AISettings.reminders_enabled`
- `AISettings.planning_enabled`
- tool category
- tool risk level
- `ToolDefinition.requires_confirmation`
- `ToolDefinition.min_autonomy_level`
- `ToolDefinition.input_schema`

## Tool Selection Rules

A tool is available only when:

- it is enabled,
- the user's autonomy level meets `min_autonomy_level`,
- its feature area is enabled,
- non-confirmed high-risk execution is not allowed.

Current feature gates:

- memory tools require `memory_enabled=true`
- reminder tools require `reminders_enabled=true`
- activity/planning tools require `planning_enabled=true`

## Proposal Rules

A planned action can become a proposal only when:

- `user_id` matches the authenticated context,
- the tool was selected by policy,
- the proposed input is valid JSON,
- the proposed input matches `ToolDefinition.input_schema`,
- write actions remain proposals,
- execution remains explicit and endpoint-driven.

## Execution Rules

Sprint 13 does not add automatic execution.

The `SafetyPolicy.CanExecute` path still denies automatic runtime execution. Existing `/ai/action-proposals/{id}/execute` behavior remains explicit and confirmation-aware.

## Remaining Work

Before real LLM integration:

- add policy reason codes for observability,
- persist denied proposal attempts if useful,
- add stricter field constraints such as lengths and RFC3339 formats,
- add tool-level risk configuration instead of deriving risk in code.
