# Sprint 20 Report - Gemini Evals & Metrics

## Summary

Sprint 20 hardens AI Runtime evaluations for optional real Gemini runs without enabling production, automatic execution, hardware integrations or real data writes.

## What Changed

- `cmd/tools/run-ai-evals` still runs fake evals every time.
- Gemini evals run only when `GEMINI_API_KEY` exists.
- Missing Gemini credentials produce a clear skipped report and do not fail CI.
- Reports are separated:
  - `docs/ai_eval_report.md`
  - `docs/gemini_eval_report.md`
- Versioned eval dataset now includes profiles:
  - `safety`
  - `planning`
  - `memory`
  - `reminders`
  - `injection`
- Eval reports now include metrics:
  - pass/fail,
  - provider,
  - latency_ms,
  - rejected_actions_count,
  - proposed_actions_count,
  - schema_failures_count,
  - policy_failures_count.

## Safety Behavior

- No proposals are executed.
- No hardware or external product actions are called.
- Fake evals use in-memory repositories.
- Gemini is optional and skipped when `GEMINI_API_KEY` is absent.
- Reports redact/truncate inputs and do not include API keys.

## Current Run

Fake provider:

- 12 passed
- 0 failed

Gemini provider:

- skipped because `GEMINI_API_KEY` was not set.

## Commands Run

```sh
env GOCACHE=/Users/armandoalvarado/Documents/Sophia/sophia_ai_backend/sofia-backend/.cache/go-build go test ./...
```

```sh
env GOCACHE=/Users/armandoalvarado/Documents/Sophia/sophia_ai_backend/sofia-backend/.cache/go-build go run ./cmd/tools/run-ai-evals -provider=auto -output-dir=docs -dataset=docs/evals/ai_runtime_cases.json
```

## Remaining Work

- Run Gemini evals with a temporary API key before any real rollout.
- Add provider-specific quality grading beyond safety pass/fail.
- Add structured JSON eval output if external dashboards need ingestion.
- Add rejection reason codes from policy/schema layers instead of current aggregate classification.
