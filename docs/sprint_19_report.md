# Sprint 19 Report - Gemini Runtime Evals & Safe Observability

## Summary

Sprint 19 hardens AI Runtime before prolonged real-model usage. It keeps the system proposal-only, keeps Gemini optional, and adds safe observability without logging full prompts or context.

## What Changed

- Added versioned eval dataset:
  - `docs/evals/ai_runtime_cases.json`
- Updated `cmd/tools/run-ai-evals`:
  - always runs fake provider evals,
  - runs Gemini evals only when `GEMINI_API_KEY` exists,
  - writes `docs/ai_eval_report.md`,
  - writes `docs/gemini_eval_report.md`.
- Added prompt injection eval case.
- Added runtime observability:
  - `request_id`,
  - provider latency,
  - provider error type,
  - planned actions count,
  - proposals count,
  - rejected actions count.
- Added audit metadata for provider/evaluation stages without full prompt logging.

## Eval Cases

Dataset cases cover:

- create reminder,
- create activity,
- create memory,
- reject injected `user_id`,
- reject extra fields,
- reject unknown tool,
- memory disabled,
- reminders disabled,
- planning disabled,
- basic prompt injection,
- memory privacy,
- no automatic execution.

## Gemini Behavior

Gemini is not required in CI. In this run, `GEMINI_API_KEY` was not set, so the Gemini eval report was generated as skipped.

## Safety Invariants

- No automatic execution.
- `dry_run=true` does not persist proposals.
- `dry_run=false` can create proposals but does not execute them.
- Unsafe model output is filtered by tool selection, policy, safety and schema validation.
- Full prompts are not logged in audit metadata.

## Tests Added

- versioned dataset evals pass with fake provider,
- prompt injection case exists in dataset,
- Gemini eval command skips without API key,
- runtime observability is populated,
- audit metadata does not include full prompt,
- proposal-only behavior remains intact.

## Commands Run

```sh
env GOCACHE=/Users/armandoalvarado/Documents/Sophia/sophia_ai_backend/sofia-backend/.cache/go-build go test ./...
```

```sh
env GOCACHE=/Users/armandoalvarado/Documents/Sophia/sophia_ai_backend/sofia-backend/.cache/go-build go run ./cmd/tools/run-ai-evals -provider=auto -output-dir=docs -dataset=docs/evals/ai_runtime_cases.json
```

## Remaining Work

- Run Gemini evals with a temporary API key before production use.
- Add policy rejection reason codes.
- Add provider name to observability.
- Add metrics sink/dashboard.
- Add stronger DLP before any prompt logging.
