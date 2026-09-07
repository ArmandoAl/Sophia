# Local Runbook

Fecha: 2026-07-04

## Prerrequisitos

- Go instalado.
- Repo en `sofia-backend`.
- No guardar credenciales JSON dentro del repo.
- Para Firestore real o emulator, configurar variables explicitas.

## Validacion Rapida

```sh
env GOCACHE=/Users/armandoalvarado/Documents/Sophia/sophia_ai_backend/sofia-backend/.cache/go-build go test ./...
```

## Arrancar API con Memory

```sh
ENV=development \
PORT=8080 \
PERSISTENCE_DRIVER=memory \
AI_MODEL_PROVIDER=fake \
AI_RUNTIME_PROPOSAL_ONLY=true \
go run ./cmd/sofia
```

Health check:

```sh
curl http://localhost:8080/health
```

## Arrancar API con Firestore

```sh
ENV=development \
PORT=8080 \
PERSISTENCE_DRIVER=firestore \
FIRESTORE_PROJECT_ID=<project_id> \
FIRESTORE_DATABASE_ID=default \
GOOGLE_APPLICATION_CREDENTIALS=<path_to_credentials> \
AI_MODEL_PROVIDER=fake \
AI_RUNTIME_PROPOSAL_ONLY=true \
go run ./cmd/sofia
```

Si usas emulator:

```sh
FIRESTORE_EMULATOR_HOST=127.0.0.1:8080
```

## Registrar Usuario y Login

Register:

```sh
curl -X POST http://localhost:8080/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"name":"Armando","email":"armando@example.com","password":"password123"}'
```

Login:

```sh
curl -X POST http://localhost:8080/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"armando@example.com","password":"password123"}'
```

Usar el token devuelto:

```sh
curl http://localhost:8080/users/me \
  -H 'Authorization: Bearer <token>'
```

## AI Runtime Fake

```sh
curl -X POST http://localhost:8080/ai/runtime/message \
  -H 'Authorization: Bearer <token>' \
  -H 'Content-Type: application/json' \
  -d '{"message":"Crea una tarea para revisar el roadmap mañana","dry_run":true}'
```

Esperado:

- provider fake,
- proposal-only,
- no ejecucion automatica.

## Reminder Worker Noop

```sh
ENV=development \
PERSISTENCE_DRIVER=memory \
REMINDER_WORKER_ENABLED=true \
REMINDER_WORKER_ID=local-worker-1 \
REMINDER_DELIVERY_PROVIDER=noop \
REMINDER_WORKER_INTERVAL=30s \
go run ./cmd/workers/reminders
```

## Reminder Worker Lease Validation

```sh
go run ./cmd/tools/test-reminder-worker-lease -output-dir=docs
```

Revisar:

```text
docs/reminder_worker_operational_report.md
```

## AI Evals Fake

```sh
go run ./cmd/tools/run-ai-evals
```

Revisar:

```text
docs/ai_eval_report.md
```

## Gemini Diagnostic

```sh
go run ./cmd/tools/test-gemini-provider
```

Sin `GEMINI_API_KEY`, el resultado esperado es `skipped`.

## FCM Dry-Run

Requiere Firestore, FCM project y device token existente:

```sh
PERSISTENCE_DRIVER=firestore \
FIRESTORE_PROJECT_ID=<project_id> \
FCM_PROJECT_ID=<project_id> \
GOOGLE_APPLICATION_CREDENTIALS=<path_to_credentials> \
go run ./cmd/tools/test-fcm-delivery -device-token-id=<device_token_id> -dry-run=true
```

Revisar:

```text
docs/fcm_test_report.md
```

Debe mantener `visible_notification=false`.

## Firestore Snapshot

```sh
PERSISTENCE_DRIVER=firestore \
FIRESTORE_PROJECT_ID=<project_id> \
GOOGLE_APPLICATION_CREDENTIALS=<path_to_credentials> \
go run ./cmd/tools/export-firestore-schema -limit=20 -output-dir=docs -redact=true
```

Archivos esperados:

- `docs/database_snapshot.json`
- `docs/database_schema.md`
- `docs/database_collections_summary.md`

## Migracion Legacy Segura

Validate-only:

```sh
PERSISTENCE_DRIVER=firestore \
FIRESTORE_PROJECT_ID=<project_id> \
go run ./cmd/tools/migrate-firestore-legacy \
  -validate-only=true \
  -target-mode=modular-only \
  -limit=20 \
  -output-dir=docs
```

No ejecutar `-apply=true` sin backup y checklist manual.

## Delete Executor

Solo procesa requests `approved`:

```sh
PERSISTENCE_DRIVER=firestore \
FIRESTORE_PROJECT_ID=<project_id> \
go run ./cmd/tools/process-delete-requests -limit=20 -output-dir=docs
```

No ejecutarlo contra Firestore real sin revisar antes las delete requests aprobadas. Este comando no tiene modo dry-run.

## Seguridad Local

- No usar secretos reales en comandos que queden en historial compartido.
- No imprimir tokens completos.
- No commitear credenciales.
- Mantener `AI_RUNTIME_PROPOSAL_ONLY=true`.
- Mantener `FCM_DRY_RUN=true` salvo prueba controlada.
