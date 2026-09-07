# Context Budget

Sprint 13 adds an approximate context budget to the AI Runtime ContextBuilder.

## Purpose

The backend should not send unlimited user data to a model provider. Context must be:

- scoped to the authenticated user,
- redacted,
- bounded,
- ranked by usefulness,
- safe for future prompt construction.

## Current Budget

Default approximate budget:

- `1200` approximate tokens

Approximation:

- about one token per four runes,
- plus small per-item overhead.

This is intentionally simple until a real provider tokenizer is selected.

## Included Context

ContextBuilder may include:

- user summary with email redacted,
- profile summary,
- AI settings,
- recent activities,
- due reminders when reminders are enabled,
- insights summary,
- assistant-visible active memories when memory is enabled.

## Memory Rules

Memories are included only when:

- `memory_enabled=true`,
- memory status is `active`,
- memory visibility is `assistant_context`.

Relevant memories are ranked by:

1. importance,
2. recency via `updated_at`.

## Redaction

The current redaction layer:

- replaces email-like text with `[REDACTED]`,
- truncates long text snippets,
- normalizes whitespace.

This is basic prompt hygiene, not a full DLP system.

## Response Field

Runtime responses include:

```json
{
  "context_summary": {
    "token_budget": {
      "max_approx_tokens": 1200,
      "used_approx_tokens": 42
    }
  }
}
```

## Remaining Work

- Use provider-specific tokenizer once Gemini/OpenAI decision is made.
- Add per-section budgets.
- Add stronger PII redaction.
- Add configurable max context budget by environment.
