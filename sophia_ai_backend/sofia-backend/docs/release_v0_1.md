# Sofia Backend v0.1

Fecha: 2026-07-04

## Objetivo de la release

Sofia Backend v0.1 es una version operable personal del backend modular. Esta pensada para correr localmente o contra Firestore controlado, validar flujos base y preparar la siguiente etapa sin activar autonomia real.

Esta release no es una version de produccion publica.

## Incluye

- API HTTP con `cmd/sofia`.
- Auth con JWT.
- Users/profile/AI settings.
- Activities.
- Reminders con recurrencia simple.
- Reminder worker opcional con provider `noop` por default.
- Device tokens y FCM opcional en dry-run documentado.
- Insights.
- Memory textual.
- Privacy export, delete request y delete executor interno.
- Tool registry.
- AI action proposals.
- AI Runtime proposal-only con provider `fake` por default.
- Gemini provider documentado y diagnosticable, no default.
- Firestore exporter y migrador legacy seguro/no destructivo.
- Evals fake para AI Runtime.

## No Incluye

- Autonomia real.
- Ejecucion automatica de acciones de IA.
- LLM obligatorio.
- Push notifications reales por default.
- Hardware/smart home.
- RAG/embeddings.
- Migracion destructiva de datos legacy.
- Garantia multi-instancia productiva sin validacion adicional con emulator/produccion.

## Modos soportados

### Local memory

Modo recomendado para desarrollo personal rapido:

```sh
ENV=development \
PERSISTENCE_DRIVER=memory \
AI_MODEL_PROVIDER=fake \
AI_RUNTIME_PROPOSAL_ONLY=true \
go run ./cmd/sofia
```

### Firestore controlado

Modo para persistencia real:

```sh
ENV=development \
PERSISTENCE_DRIVER=firestore \
FIRESTORE_PROJECT_ID=<project_id> \
FIRESTORE_DATABASE_ID=default \
GOOGLE_APPLICATION_CREDENTIALS=<path> \
AI_MODEL_PROVIDER=fake \
AI_RUNTIME_PROPOSAL_ONLY=true \
go run ./cmd/sofia
```

Usar `FIRESTORE_DATABASE_ID` solo si aplica. No versionar credenciales.

## Procesos

| Proceso | Comando | Estado v0.1 |
| --- | --- | --- |
| API HTTP | `go run ./cmd/sofia` | Operable |
| Reminder worker | `go run ./cmd/workers/reminders` | Operable con `noop`, deshabilitado por default |
| Firestore exporter | `go run ./cmd/tools/export-firestore-schema` | Operable con Firestore configurado |
| Legacy migrator | `go run ./cmd/tools/migrate-firestore-legacy` | Seguro por default, dry-run/staging/modular-only |
| Delete executor | `go run ./cmd/tools/process-delete-requests` | Interno, solo approved requests |
| AI eval runner | `go run ./cmd/tools/run-ai-evals` | Fake siempre, Gemini opcional |
| Gemini diagnostic | `go run ./cmd/tools/test-gemini-provider` | Skips seguro sin API key |
| FCM dry-run | `go run ./cmd/tools/test-fcm-delivery` | Validate-only documentado |
| Worker lease test | `go run ./cmd/tools/test-reminder-worker-lease` | Simulacion deterministica |

## Criterio de release

La release se considera lista para uso personal cuando:

- `go test ./...` pasa.
- API HTTP arranca con memory driver.
- Firestore driver arranca con configuracion valida o emulator.
- AI Runtime fake responde sin ejecutar acciones.
- Worker arranca en modo `noop` sin enviar notificaciones.
- Migraciones permanecen en modo seguro/no destructivo.
- Evals fake pasan.

## Verificacion ejecutada

```sh
env GOCACHE=/Users/armandoalvarado/Documents/Sophia/sophia_ai_backend/sofia-backend/.cache/go-build go test ./...
```

Resultado: PASS.
