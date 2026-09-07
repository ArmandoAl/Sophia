# Runtime Observability

Sprint 19 adds safe observability to AI Runtime without logging full prompts or model context.

## Response Fields

`POST /ai/runtime/message` now includes:

```json
{
  "request_id": "...",
  "observability": {
    "provider_latency_ms": 12,
    "provider_error_type": "",
    "planned_actions_count": 1,
    "proposals_count": 1,
    "rejected_actions_count": 0
  }
}
```

## Audit Metadata

Runtime audit events include:

- `request_id`
- provider latency in milliseconds,
- provider error type when applicable,
- planned actions count,
- persisted proposal count,
- rejected actions count,
- dry-run flag,
- token/context counts.

The audit metadata intentionally does not include:

- full prompt text,
- full context payload,
- raw model response,
- API keys,
- tokens,
- secrets.

The only message-derived field currently recorded is a coarse length fingerprint such as `len:42`.

## Events

| Event | Purpose |
| --- | --- |
| `ai_runtime_context_built` | Context was built with bounded counts and token estimate. |
| `ai_runtime_model_request_planned` | Provider call is about to happen; records tool count and dry-run. |
| `ai_runtime_provider_failed` | Provider returned an error. |
| `ai_runtime_proposals_generated` | Provider returned planned actions. |
| `ai_runtime_proposals_evaluated` | Safety/policy filtering finished. |

## Privacy Notes

Observability is operational, not forensic prompt logging. Full prompts and raw context should remain out of logs until a stronger redaction and retention policy exists.

## Remaining Work

- Add structured provider names.
- Add dashboards/metrics sink.
- Add per-policy rejection reason codes.
- Add trace propagation from HTTP request headers if needed.
