# Gemini Provider

Sprint 15 adds Gemini as a real `ModelClient` for the AI Runtime.

Gemini is proposal-only in this backend. It can suggest structured actions, but it cannot execute, confirm, or bypass policy.

## Configuration

```sh
AI_RUNTIME_ENABLED=true
AI_RUNTIME_PROPOSAL_ONLY=true
AI_MODEL_PROVIDER=gemini
GEMINI_API_KEY=...
GEMINI_MODEL=gemini-2.5-flash
```

Defaults:

- `AI_MODEL_PROVIDER=fake`
- `GEMINI_MODEL=gemini-1.5-flash`
- `AI_RUNTIME_ENABLED=true`
- `AI_RUNTIME_PROPOSAL_ONLY=true`

`GEMINI_API_KEY` is required only when `AI_MODEL_PROVIDER=gemini`.

Supported/tested config values:

- `GEMINI_MODEL=gemini-2.5-flash`
- `GEMINI_MODEL=gemini-1.5-flash`

## Location

```text
internal/ai/runtime/infrastructure/gemini
```

The provider implements:

```go
type ModelClient interface {
    Generate(ctx context.Context, request ModelRequest) (ModelResponse, error)
}
```

## Structured Output

Gemini is asked to return JSON only:

```json
{
  "assistant_message": "string",
  "proposed_actions": [
    {
      "tool_name": "create_reminder",
      "proposed_input": {},
      "reason": "string",
      "risk_level": "low|medium|high",
      "requires_confirmation": true
    }
  ]
}
```

The backend then converts valid actions into internal `PlannedAction` values.

## Safety Path

Every Gemini planned action still passes through:

- `ToolSelector`
- `PolicyEngine`
- `SafetyPolicy`
- `ToolDefinition.input_schema` validation
- AI action proposal creation rules

With `dry_run=true`, no proposal is persisted.

With `dry_run=false`, a safe proposal may be created, but it is not executed.

## Privacy

Before reaching Gemini, context is already compacted by ContextBuilder:

- email is redacted,
- deleted memories are excluded,
- only `visibility=assistant_context` memories are included,
- long snippets are truncated,
- memory content is represented as summary/title-level context.

The provider does not log prompts or raw model responses.

## Tests

Gemini tests use an in-memory fake HTTP transport. They do not require an API key and do not call Google.

Covered:

- missing API key fails,
- valid structured response parses,
- invalid structured response fails,
- HTTP error returns typed provider error,
- 401 is classified as `unauthorized`,
- 403 is classified as `permission_denied` or `quota`,
- 404 is classified as `model_not_found`,
- 429 is classified as `quota`,
- error body is redacted/truncated and API keys are never included.

## Diagnostic Command

```sh
go run ./cmd/tools/test-gemini-provider -output-dir=docs
```

Optional model override:

```sh
go run ./cmd/tools/test-gemini-provider -model=gemini-2.5-flash -output-dir=docs
```

The command writes:

```text
docs/gemini_provider_diagnostic.md
```

If `GEMINI_API_KEY` is missing, the report is marked `skipped` and does not fail.

## Limitations

- No retries yet.
- No provider-specific tokenizer yet.
- Gemini evals are optional and skipped when `GEMINI_API_KEY` is not set.
- Eval reports include safety metrics but do not grade answer quality yet.
- The default model should be reviewed before production rollout.
