# v0.1 Checklist

Fecha: 2026-07-04

## Release Gate

| Item | Estado | Evidencia |
| --- | --- | --- |
| API HTTP corre | ready | `go run ./cmd/sofia` con config local. |
| Memory driver corre | ready | `PERSISTENCE_DRIVER=memory` es default y cubierto por tests. |
| Firestore driver corre | ready con config | `PERSISTENCE_DRIVER=firestore`, `FIRESTORE_PROJECT_ID`, `FIRESTORE_DATABASE_ID` si aplica. |
| AI Runtime fake corre | ready | `AI_MODEL_PROVIDER=fake` por default. |
| Gemini diagnostic documentado | ready | `docs/gemini_provider_diagnostic.md`, `docs/gemini_troubleshooting.md`. |
| Worker noop corre | ready | `REMINDER_DELIVERY_PROVIDER=noop`, `REMINDER_WORKER_ENABLED=true`. |
| FCM dry-run documentado | ready | `docs/fcm_real_dry_run.md`, `docs/fcm_test_report.md`. |
| Migrations safe/no destructive | ready | Migrador dry-run/staging/modular-only; ver `docs/safe_apply_plan.md`. |
| Evals fake pasan | ready | `docs/ai_eval_report.md`: 12 passed, 0 failed. |

## Comandos Minimos

### Tests

```sh
env GOCACHE=/Users/armandoalvarado/Documents/Sophia/sophia_ai_backend/sofia-backend/.cache/go-build go test ./...
```

### API local memory

```sh
ENV=development \
PERSISTENCE_DRIVER=memory \
AI_MODEL_PROVIDER=fake \
AI_RUNTIME_PROPOSAL_ONLY=true \
go run ./cmd/sofia
```

### Worker noop

```sh
ENV=development \
PERSISTENCE_DRIVER=memory \
REMINDER_WORKER_ENABLED=true \
REMINDER_DELIVERY_PROVIDER=noop \
go run ./cmd/workers/reminders
```

### AI evals fake

```sh
go run ./cmd/tools/run-ai-evals
```

### Gemini diagnostic

```sh
go run ./cmd/tools/test-gemini-provider
```

Sin `GEMINI_API_KEY`, el resultado esperado es `skipped`.

### FCM dry-run

```sh
go run ./cmd/tools/test-fcm-delivery -device-token-id=<device_token_id> -dry-run=true
```

### Reminder worker lease validation

```sh
go run ./cmd/tools/test-reminder-worker-lease -output-dir=docs
```

## No-Go

No promover v0.1 si:

- `AI_RUNTIME_PROPOSAL_ONLY=false`.
- `FCM_DRY_RUN=false` sin dry-run real validado.
- `REMINDER_DELIVERY_PROVIDER=fcm` en entorno no controlado.
- Migrador legacy se ejecuta con `-apply=true` sin checklist manual.
- Firestore real no tiene backup antes de cualquier apply.

