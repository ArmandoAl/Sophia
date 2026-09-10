# Environment

## Core

| Variable | Default | Required | Description |
| --- | --- | --- | --- |
| `ENV` | `development` | no | Runtime mode: `development`, `test` or `production`. |
| `PORT` | `8080` | no | HTTP server port. |
| `JWT_SECRET` | insecure development default | production | JWT signing secret. Must be explicit and secure in production. |
| `JWT_ACCESS_TOKEN_TTL` | `24h` | no | Access token lifetime as Go duration. |
| `AUTH_RATE_LIMIT_REQUESTS` | `10` | no | Max auth requests per IP+route within the rate window. |
| `AUTH_RATE_LIMIT_WINDOW` | `1m` | no | Auth rate limit window as Go duration. |
| `CORS_ALLOWED_ORIGINS` | empty | no | Comma-separated allowed origins. |
| `REQUEST_BODY_LIMIT_BYTES` | `1048576` | no | Max JSON request body size. |

## Persistence

| Variable | Default | Required | Description |
| --- | --- | --- | --- |
| `PERSISTENCE_DRIVER` | `memory` | production | `memory` or `firestore`. Production requires `firestore`. |
| `FIRESTORE_PROJECT_ID` | empty | when Firestore is enabled | Google Cloud/Firebase project ID. |
| `GOOGLE_CLOUD_PROJECT` | `FIRESTORE_PROJECT_ID` | when embeddings are enabled | Google Cloud project used by Vertex AI. Firestore also uses it when `FIRESTORE_PROJECT_ID` is empty. |
| `FIRESTORE_DATABASE_ID` | empty | when Firestore is enabled | Explicit Firestore database ID. The implicit `"(default)"` database is not assumed; this project uses the named database `default`. |
| `GOOGLE_APPLICATION_CREDENTIALS` | empty | local Firestore with service account | Path to local Google credentials. Do not commit credential JSON files. |
| `FIRESTORE_EMULATOR_HOST` | empty | emulator only | Firestore emulator host, for example `localhost:8081`. |

## HTTP Timeouts

| Variable | Default | Description |
| --- | --- | --- |
| `HTTP_READ_TIMEOUT` | `10s` | Request read timeout. |
| `HTTP_WRITE_TIMEOUT` | `10s` | Response write timeout. |
| `HTTP_IDLE_TIMEOUT` | `60s` | Keep-alive idle timeout. |
| `HTTP_READ_HEADER_TIMEOUT` | `5s` | Header read timeout. |

## AI Runtime

| Variable | Default | Required | Description |
| --- | --- | --- | --- |
| `AI_RUNTIME_ENABLED` | `true` | no | Enables the AI runtime wiring. When false, the app uses the fake provider path. |
| `AI_RUNTIME_PROPOSAL_ONLY` | `true` | yes | Must remain true. `false` is rejected because autonomous execution is not supported yet. |
| `CONTEXT_TOKEN_BUDGET` | `4600` | no | Maximum approximate tokens assembled for one runtime request. |
| `AUTONOMY_THRESHOLD` | `0.85` | no | Confidence threshold used by autonomy decisions. |
| `AI_MODEL_PROVIDER` | `fake` | no | `fake` or `deepseek`. |
| `DEEPSEEK_API_KEY` | empty | when `AI_MODEL_PROVIDER=deepseek` | DeepSeek API key. Never commit or log this value. |
| `DEEPSEEK_MODEL` | `deepseek-v4-pro` | when `AI_MODEL_PROVIDER=deepseek` | Default model for tasks without an override. |
| `DEEPSEEK_MODEL_PLAN` | `DEEPSEEK_MODEL` | no | Optional model override for planning. |
| `DEEPSEEK_MODEL_SYNTHESIZE` | `DEEPSEEK_MODEL` | no | Optional model override for daily synthesis. |
| `DEEPSEEK_MODEL_EXTRACT` | `DEEPSEEK_MODEL` | no | Optional model override for extraction. |
| `DEEPSEEK_BASE_URL` | DeepSeek API | no | Optional API base URL override. |

## Embeddings

| Variable | Default | Required | Description |
| --- | --- | --- | --- |
| `EMBEDDINGS_ENABLED` | `false` | no | Enables Vertex AI embeddings for semantic belief deduplication. |
| `EMBEDDINGS_MODEL` | empty | when embeddings are enabled | Vertex AI text embedding model ID. |
| `BELIEF_DEDUPE_THRESHOLD` | `0.85` | no | Minimum cosine similarity for beliefs in the same category. |

## Ingestion Worker

| Variable | Default | Required | Description |
| --- | --- | --- | --- |
| `INGESTION_MAX_TOKENS_PER_BATCH` | `50000` | no | Hard input plus output token budget for one imported batch. |

## Synthesis Worker

| Variable | Default | Required | Description |
| --- | --- | --- | --- |
| `SYNTHESIS_WORKER_ENABLED` | `false` | no | Enables the standalone daily synthesis worker. |
| `SYNTHESIS_WORKER_ID` | host-derived | no | Stable worker identity used for synthesis leases. |
| `SYNTHESIS_WORKER_INTERVAL` | `1m` | no | Poll interval as a positive Go duration. |
| `SYNTHESIS_WORKER_LEASE` | `2m` | no | Lease duration for one daily synthesis run. |
| `SYNTHESIS_RUN_HOUR_LOCAL` | `3` | no | Local hour, from 0 through 23, when daily synthesis runs. |

## Reminder Worker

| Variable | Default | Required | Description |
| --- | --- | --- | --- |
| `REMINDER_WORKER_ENABLED` | `false` | no | Enables the standalone reminder worker command. |
| `REMINDER_WORKER_ID` | host-derived | no | Stable worker identity used for claim/lease metadata. Set explicitly in multi-instance deployments. |
| `REMINDER_WORKER_INTERVAL` | `30s` | no | Poll interval for due reminders. Must be a positive Go duration. |
| `REMINDER_WORKER_BATCH_SIZE` | `50` | no | Max due reminders processed per worker tick. Must be a positive integer. |
| `REMINDER_WORKER_LEASE_DURATION` | `2m` | no | Processing lease duration for one delivery attempt. Must exceed expected provider latency. |
| `REMINDER_DELIVERY_PROVIDER` | `noop` | no | Reminder delivery provider: `noop` or `fcm`. |

## FCM

| Variable | Default | Required | Description |
| --- | --- | --- | --- |
| `FCM_ENABLED` | `false` | no | Enables FCM provider behavior. When false, FCM delivery is a safe no-op. |
| `FCM_DRY_RUN` | `true` | no | Prevents real network sends when true. |
| `FCM_PROJECT_ID` | empty | when FCM is enabled | Firebase/Google project ID used for FCM v1. |

Real FCM delivery also uses `GOOGLE_APPLICATION_CREDENTIALS` or Application Default Credentials.

## Firestore Named Databases

For projects using a named database, set:

```sh
export PERSISTENCE_DRIVER=firestore
export FIRESTORE_PROJECT_ID=sophia-ai-60e87
export FIRESTORE_DATABASE_ID=default
```

With `PERSISTENCE_DRIVER=firestore`, `FIRESTORE_DATABASE_ID` must be set. The backend fails during configuration loading instead of falling back to `firestore.NewClient` and the implicit `"(default)"` database.
