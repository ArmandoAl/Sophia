# Sofia Flutter Canonical API Contract v0.1

Fecha: 2026-07-10  
Sprint: F0 cerrado como consolidación documental  
Estado: **OPERATIVE REFERENCE** para implementación Flutter F1+

Este documento **sustituye como referencia operativa** a estimaciones iniciales en `flutter_models_needed.md` y a envelopes incorrectos en `flutter_endpoint_integration_contract.md`.  
Los audits F0 y docs históricos **permanecen** como evidencia; no borrarlos.

---

## 1. Contract Authority

### Precedencia (obligatoria)

1. Comportamiento demostrado por E2E/integration **HTTP** (hoy: casi ausente; `routes_test.go` es el proxy más fuerte).
2. Handler/route real (`sofia-backend/internal/server/routes.go` + `*/interfaces/http`).
3. DTO/struct + validación de dominio.
4. Tests unitarios / app-layer E2E (`run-local-e2e` — **no** pinnea JSON wire).
5. Documentación previa (`flutter_*`, Postman, runbooks) — **menor autoridad**.

### Contradicciones

Si docs ≠ handlers: **gana el handler**.  
Si handler ≠ test HTTP: marcar **PARTIAL** / **NEEDS_BACKEND_CLARIFICATION**.  
Si serialización ambigua (export): **BLOCKED** para codegen.

### Fuentes F0

| Doc | Rol |
| --- | --- |
| `docs/audits/flutter_contract_f0_http_verification.md` | Paths, wrappers, status codes |
| `docs/audits/flutter_contract_f0_schema_inventory.md` | Campos, tipos, enums |
| `docs/audits/flutter_contract_f0_test_evidence.md` | Qué está demostrado |

---

## 2. Global HTTP Rules

| Regla | Contrato canónico |
| --- | --- |
| **base URL** | Configurable (`AppConfig.baseUrl`). Local tipico `http://localhost:8080`. Sin path prefix. |
| **JSON** | `Content-Type: application/json`. Backend `DisallowUnknownFields` en decode → no enviar keys extras. |
| **Authorization** | `Authorization: Bearer <jwt>` en todos los endpoints Auth=Yes. Auth register/login/health: sin Bearer. |
| **RFC3339** | Request dates como **string** RFC3339. Response `time.Time` → string RFC3339. DeviceToken timestamps HTTP son **string** ya formateados. |
| **Errors** | Body JSON estándar: `{ "message": "<string>" }`. Excepción: path basura en resource routes puede ser `http.NotFound` **no JSON**. |
| **Ownership** | **Nunca** enviar `user_id` en body. Ownership = JWT. Cross-user → 404 (o 403 en algunos insights). |
| **Pagination** | Listas: `limit` (default backend 50), `cursor` (id del último item) en activities/reminders/memory. Response: `next_cursor` omitempty. **Sin tests** → usar con cautela; no asumir opaque tokens. Action-proposals: solo `limit`/`status`, **sin** cursor. |
| **Nullability** | `*time.Time` + `omitempty` → clave ausente o null → `DateTime?`. Strings `omitempty` → ausente si vacío. |
| **PATCH omitted** | Campos no cambiados: **omitir la key** del JSON (no enviar `null` salvo que se quiera el mismo efecto que omit en pointers Go). Cliente: `includeIfNull: false` / build map manual. |
| **Create zero-values** | Bool omit → `false`; string omit → `""`; int omit → `0` (p.ej. `max_retries` 0 → default 3 en reminders). |
| **Feature disabled** | Memory/reminders disabled → **409**, no 403. |
| **Rate limit** | Auth login/register pueden → **429**. |

---

## 3. Canonical Endpoint Matrix

Status:

- **READY** — path/shape/status listos para cliente (handler + evidencia razonable)
- **PARTIAL** — usable con flags / sin asserts completos
- **BLOCKED** — no implementar cliente hasta clarificar
- **INTERNAL_ONLY** — no UI usuario final (tools/debug)

| Domain | Method | Path | Auth | Request | Success | Response | Status |
| --- | --- | --- | --- | --- | --- | --- | --- |
| system | GET | `/health` | No | — | 200 | flat `HealthResponse` | READY |
| auth | POST | `/auth/register` | No | `RegisterRequest` | 201 | flat `AuthUser` (no role) | READY |
| auth | POST | `/auth/login` | No | `LoginRequest` | 200 | `LoginResponse` | READY |
| users | GET | `/users/me` | Yes | — | 200 | `MeResponse` | READY |
| users | PATCH | `/users/me/profile` | Yes | `UpdateProfileRequest` | 200 | flat `UserProfile` | READY |
| users | GET | `/users/me/ai-settings` | Yes | — | 200 | flat `AiSettings` | READY |
| users | PATCH | `/users/me/ai-settings` | Yes | `UpdateAiSettingsRequest` | 200 | flat `AiSettings` | READY |
| users | POST | `/users/me/onboarding/complete` | Yes | body ignored | 200 | flat `UserProfile` | READY |
| privacy | GET | `/users/me/export` | Yes | — | 200 | **ambiguous shape** | **BLOCKED** |
| privacy | POST | `/users/me/delete-request` | Yes | `DeleteRequestBody?` | 202 | flat `DeleteRequest` | PARTIAL |
| activities | POST | `/activities` | Yes | `CreateActivityRequest` | 201 | flat `Activity` | READY |
| activities | GET | `/activities` | Yes | query filters | 200 | `{activities,next_cursor?}` | PARTIAL (cursor untested) |
| activities | GET | `/activities/{id}` | Yes | — | 200 | flat `Activity` | READY |
| activities | PATCH | `/activities/{id}` | Yes | `UpdateActivityRequest` | 200 | flat `Activity` | READY |
| activities | POST | `/activities/{id}/complete` | Yes | `{}`/empty | 200 | flat `Activity` | READY |
| activities | POST | `/activities/{id}/cancel` | Yes | empty | 200 | flat `Activity` | READY |
| activities | POST | `/activities/{id}/archive` | Yes | empty | 200 | flat `Activity` | READY |
| reminders | POST | `/reminders` | Yes | `CreateReminderRequest` | 201 | flat `Reminder` | READY |
| reminders | GET | `/reminders` | Yes | query | 200 | `{reminders,next_cursor?}` | PARTIAL |
| reminders | GET | `/reminders/due` | Yes | `limit?` | 200 | `{reminders,next_cursor?}` | READY |
| reminders | GET | `/reminders/{id}` | Yes | — | 200 | flat `Reminder` | PARTIAL |
| reminders | PATCH | `/reminders/{id}` | Yes | `UpdateReminderRequest` | 200 | flat `Reminder` | PARTIAL (no HTTP test) |
| reminders | POST | `/reminders/{id}/cancel` | Yes | empty | 200 | flat `Reminder` | READY |
| reminders | POST | `/reminders/{id}/archive` | Yes | empty | 200 | flat `Reminder` | READY |
| insights | POST | `/insights/mood` | Yes | `CreateMoodEntryRequest` | 201 | flat `MoodEntry` | READY |
| insights | GET | `/insights/mood` | Yes | query | 200 | `{mood_entries}` | READY |
| insights | POST | `/insights/outcomes` | Yes | `CreateActivityOutcomeRequest` | 201 | flat `ActivityOutcome` | READY |
| insights | GET | `/insights/outcomes` | Yes | query | 200 | `{outcomes}` | PARTIAL (list untested) |
| insights | POST | `/insights/reflections` | Yes | `CreateReflectionRequest` | 201 | flat `Reflection` | READY |
| insights | GET | `/insights/reflections` | Yes | query | 200 | `{reflections}` | READY |
| insights | GET | `/insights/summary` | Yes | query | 200 | flat `InsightsSummary` | READY |
| memory | POST | `/memory` | Yes | `CreateMemoryRequest` | 201 | flat `Memory` | READY |
| memory | GET | `/memory` | Yes | query | 200 | `{memories,next_cursor?}` | PARTIAL |
| memory | GET | `/memory/{id}` | Yes | — | 200 | flat `Memory` | READY |
| memory | PATCH | `/memory/{id}` | Yes | `UpdateMemoryRequest` | 200 | flat `Memory` | READY |
| memory | POST | `/memory/{id}/archive` | Yes | empty | 200 | flat `Memory` | READY |
| memory | DELETE | `/memory/{id}` | Yes | — | 200 | flat `Memory` (status deleted) | READY |
| memory | POST | `/memory/search` | Yes | `SearchMemoryRequest` | 200 | `{memories,next_cursor?}` | READY |
| notifications | POST | `/notifications/device-tokens` | Yes | `RegisterDeviceTokenRequest` | 201 | flat `DeviceToken` | PARTIAL |
| notifications | GET | `/notifications/device-tokens` | Yes | — | 200 | **JSON array** `[DeviceToken]` | PARTIAL |
| notifications | DELETE | `/notifications/device-tokens/{id}` | Yes | — | 200 | `{status:"deleted"}` | PARTIAL |
| tools | GET | `/tools` | Yes | — | 200 | `{tools:[...]}` | INTERNAL_ONLY |
| tools | GET | `/tools/{name}` | Yes | — | 200 | flat `ToolDefinition` | INTERNAL_ONLY |
| ai | POST | `/ai/action-proposals` | Yes | `CreateActionProposalRequest` | 201 | flat `AiActionProposal` | PARTIAL |
| ai | GET | `/ai/action-proposals` | Yes | `status?`,`limit?` | 200 | `{action_proposals}` | PARTIAL |
| ai | GET | `/ai/action-proposals/{id}` | Yes | — | 200 | flat `AiActionProposal` | PARTIAL |
| ai | POST | `.../confirm` | Yes | empty | 200 | flat `AiActionProposal` | PARTIAL |
| ai | POST | `.../reject` | Yes | empty | 200 | flat `AiActionProposal` | PARTIAL |
| ai | POST | `.../execute` | Yes | empty | 200 | flat `AiActionProposal` | PARTIAL (feature-flag) |
| ai | POST | `/ai/runtime/message` | Yes | `RuntimeMessageRequest` | 200 | flat `RuntimeResponse` | PARTIAL |

---

## 4. Canonical Models

Convención Dart: `json_serializable` + `fieldRename: FieldRename.snake` **o** `@JsonKey(name: ...)` explícito.  
Ownership: **server** = no enviar en create/update; **client** = editable; **read** = solo response.

### 4.1 `ApiError` (body) + client wrapper

JSON body: `{ "message": "..." }`  
Client `ApiException`: `statusCode`, `message`, `rawBody?`.

| Dart Field | JSON Key | Dart Type | Nullable | Ownership |
| --- | --- | --- | --- | --- |
| message | `message` | `String` | no | server |

### 4.2 `HealthResponse`

| Dart Field | JSON Key | Dart Type | Nullable | Ownership |
| --- | --- | --- | --- | --- |
| status | `status` | `String` | no | server (`ok`\|`degraded`) |
| environment | `environment` | `String` | no | server |
| firestore | `firestore` | `String` | no | server (`ok`\|`disabled`\|`error`) |

### 4.3 `AuthUser`

Usar un modelo con campos opcionales según endpoint.

| Dart Field | JSON Key | Dart Type | Nullable | Ownership |
| --- | --- | --- | --- | --- |
| id | `id` | `String` | no | server |
| name | `name` | `String` | no | client (register) / server |
| email | `email` | `String` | no | client (register/login) |
| role | `role` | `String?` | yes | server (solo `/users/me`) |

Register/login: sin `role`. Nunca password en response.

### 4.4 `LoginResponse`

| Dart Field | JSON Key | Dart Type | Nullable | Ownership |
| --- | --- | --- | --- | --- |
| token | `token` | `String` | no | server (secure storage) |
| user | `user` | `AuthUser` | no | server |

### 4.5 `UserProfile`

Response de PATCH profile, onboarding, y nested en `MeResponse`.

| Dart Field | JSON Key | Dart Type | Nullable | Ownership |
| --- | --- | --- | --- | --- |
| id | `id` | `String` | no | server |
| userId | `user_id` | `String` | no | server (read) |
| displayName | `display_name` | `String` | no | client |
| preferredName | `preferred_name` | `String` | no | client |
| timezone | `timezone` | `String` | no | client |
| locale | `locale` | `String` | no | client |
| avatarUrl | `avatar_url` | `String?` | yes | client |
| onboardingCompleted | `onboarding_completed` | `bool` | no | server |
| createdAt | `created_at` | `DateTime` | no | server |
| updatedAt | `updated_at` | `DateTime` | no | server |

### 4.6 `AiSettings`

| Dart Field | JSON Key | Dart Type | Nullable | Ownership |
| --- | --- | --- | --- | --- |
| userId | `user_id` | `String` | no | server |
| assistantName | `assistant_name` | `String` | no | client (free string) |
| communicationStyle | `communication_style` | `String` | no | client (free string) |
| proactivityLevel | `proactivity_level` | `ProactivityLevel` | no | client |
| autonomyLevel | `autonomy_level` | `AutonomyLevel` | no | client |
| memoryEnabled | `memory_enabled` | `bool` | no | client |
| remindersEnabled | `reminders_enabled` | `bool` | no | client |
| planningEnabled | `planning_enabled` | `bool` | no | client |
| preferredLanguage | `preferred_language` | `String` | no | client (free) |
| createdAt | `created_at` | `DateTime` | no | server |
| updatedAt | `updated_at` | `DateTime` | no | server |

### 4.7 `MeResponse`

| Dart Field | JSON Key | Dart Type | Nullable | Ownership |
| --- | --- | --- | --- | --- |
| user | `user` | `AuthUser` | no | server |
| profile | `profile` | `UserProfile` | no | server |
| aiSettings | `ai_settings` | `AiSettings` | no | server |

### 4.8 `DeleteRequest`

| Dart Field | JSON Key | Dart Type | Nullable | Ownership |
| --- | --- | --- | --- | --- |
| id | `id` | `String` | no | server |
| userId | `user_id` | `String` | no | server |
| status | `status` | `DeleteRequestStatus` | no | server |
| reason | `reason` | `String?` | yes | client |
| createdAt | `created_at` | `DateTime` | no | server |
| updatedAt | `updated_at` | `DateTime` | no | server |

### 4.9 `Activity`

| Dart Field | JSON Key | Dart Type | Nullable | Ownership |
| --- | --- | --- | --- | --- |
| id | `id` | `String` | no | server |
| userId | `user_id` | `String` | no | server |
| type | `type` | `ActivityType` | no | client |
| title | `title` | `String` | no | client |
| description | `description` | `String` | no | client |
| status | `status` | `ActivityStatus` | no | both (PATCH/actions) |
| priority | `priority` | `ActivityPriority` | no | client |
| startAt | `start_at` | `DateTime?` | yes | client |
| dueAt | `due_at` | `DateTime?` | yes | client |
| completedAt | `completed_at` | `DateTime?` | yes | server |
| canceledAt | `canceled_at` | `DateTime?` | yes | server |
| archivedAt | `archived_at` | `DateTime?` | yes | server |
| timezone | `timezone` | `String` | no | client |
| tags | `tags` | `List<String>` | no | client |
| source | `source` | `ActivitySource` | no | server |
| createdBy | `created_by` | `CreatedBy` | no | server |
| createdAt | `created_at` | `DateTime` | no | server |
| updatedAt | `updated_at` | `DateTime` | no | server |
| hasReminder | `has_reminder` | `bool` | no | client |
| reminderAt | `reminder_at` | `DateTime?` | yes | client |
| recurrenceRule | `recurrence_rule` | `String?` | yes | client (free string on activity) |
| aiManageable | `ai_manageable` | `bool` | no | client |
| aiNotes | `ai_notes` | `String?` | yes | client |

### 4.10 `ActivityListResponse`

| Dart Field | JSON Key | Dart Type | Nullable | Ownership |
| --- | --- | --- | --- | --- |
| activities | `activities` | `List<Activity>` | no | server |
| nextCursor | `next_cursor` | `String?` | yes | server |

### 4.11 `Reminder` (HTTP subset — canónico)

**No incluir** processing locks ni recurrence_interval/until/count/occurrences (no en HTTP response).

| Dart Field | JSON Key | Dart Type | Nullable | Ownership |
| --- | --- | --- | --- | --- |
| id | `id` | `String` | no | server |
| userId | `user_id` | `String` | no | server |
| activityId | `activity_id` | `String?` | yes | client (create) |
| title | `title` | `String` | no | client |
| description | `description` | `String` | no | client |
| status | `status` | `ReminderStatus` | no | server |
| scheduledAt | `scheduled_at` | `DateTime` | no | client |
| timezone | `timezone` | `String` | no | client |
| recurrenceRule | `recurrence_rule` | `RecurrenceRule?` | yes | client |
| nextRunAt | `next_run_at` | `DateTime` | no | server |
| lastRunAt | `last_run_at` | `DateTime?` | yes | server |
| deliveredAt | `delivered_at` | `DateTime?` | yes | server |
| failedAt | `failed_at` | `DateTime?` | yes | server |
| failureReason | `failure_reason` | `String?` | yes | server |
| retryCount | `retry_count` | `int` | no | server |
| maxRetries | `max_retries` | `int` | no | client/server |
| source | `source` | `ReminderSource` | no | server |
| createdBy | `created_by` | `CreatedBy` | no | server |
| createdAt | `created_at` | `DateTime` | no | server |
| updatedAt | `updated_at` | `DateTime` | no | server |

### 4.12 `ReminderListResponse`

| Dart Field | JSON Key | Dart Type | Nullable | Ownership |
| --- | --- | --- | --- | --- |
| reminders | `reminders` | `List<Reminder>` | no | server |
| nextCursor | `next_cursor` | `String?` | yes | server |

### 4.13 `MoodEntry`

| Dart Field | JSON Key | Dart Type | Nullable | Ownership |
| --- | --- | --- | --- | --- |
| id | `id` | `String` | no | server |
| userId | `user_id` | `String` | no | server |
| activityId | `activity_id` | `String?` | yes | client |
| mood | `mood` | `Mood` | no | client |
| energyLevel | `energy_level` | `int` | no | client (1..5) |
| stressLevel | `stress_level` | `int` | no | client |
| focusLevel | `focus_level` | `int` | no | client |
| note | `note` | `String?` | yes | client |
| occurredAt | `occurred_at` | `DateTime` | no | client |
| timezone | `timezone` | `String` | no | client |
| createdAt | `created_at` | `DateTime` | no | server |
| updatedAt | `updated_at` | `DateTime` | no | server |

### 4.14 `MoodEntriesListResponse`

| Dart Field | JSON Key | Dart Type | Nullable | Ownership |
| --- | --- | --- | --- | --- |
| moodEntries | `mood_entries` | `List<MoodEntry>` | no | server |

### 4.15 `ActivityOutcome`

| Dart Field | JSON Key | Dart Type | Nullable | Ownership |
| --- | --- | --- | --- | --- |
| id | `id` | `String` | no | server |
| userId | `user_id` | `String` | no | server |
| activityId | `activity_id` | `String` | no | client |
| result | `result` | `OutcomeResult` | no | client |
| completionQuality | `completion_quality` | `int` | no | client 1..5 |
| satisfactionLevel | `satisfaction_level` | `int` | no | client |
| difficultyLevel | `difficulty_level` | `int` | no | client |
| note | `note` | `String?` | yes | client |
| createdAt | `created_at` | `DateTime` | no | server |
| updatedAt | `updated_at` | `DateTime` | no | server |

### 4.16 `ActivityOutcomesListResponse`

| Dart Field | JSON Key | Dart Type | Nullable | Ownership |
| --- | --- | --- | --- | --- |
| outcomes | `outcomes` | `List<ActivityOutcome>` | no | server |

**NOT** `activity_outcomes`.

### 4.17 `Reflection`

| Dart Field | JSON Key | Dart Type | Nullable | Ownership |
| --- | --- | --- | --- | --- |
| id | `id` | `String` | no | server |
| userId | `user_id` | `String` | no | server |
| activityId | `activity_id` | `String?` | yes | client |
| title | `title` | `String?` | yes | client |
| content | `content` | `String` | no | client |
| sentiment | `sentiment` | `String?` | yes | client (free) |
| tags | `tags` | `List<String>` | no | client |
| occurredAt | `occurred_at` | `DateTime` | no | client |
| createdAt | `created_at` | `DateTime` | no | server |
| updatedAt | `updated_at` | `DateTime` | no | server |

### 4.18 `ReflectionsListResponse`

| Dart Field | JSON Key | Dart Type | Nullable | Ownership |
| --- | --- | --- | --- | --- |
| reflections | `reflections` | `List<Reflection>` | no | server |

### 4.19 `InsightsSummary`

| Dart Field | JSON Key | Dart Type | Nullable | Ownership |
| --- | --- | --- | --- | --- |
| averageEnergy | `average_energy` | `double` | no | server |
| averageStress | `average_stress` | `double` | no | server |
| averageFocus | `average_focus` | `double` | no | server |
| totalOutcomes | `total_outcomes` | `int` | no | server |
| totalReflections | `total_reflections` | `int` | no | server |

### 4.20 `Memory`

| Dart Field | JSON Key | Dart Type | Nullable | Ownership |
| --- | --- | --- | --- | --- |
| id | `id` | `String` | no | server |
| userId | `user_id` | `String` | no | server |
| type | `type` | `MemoryType` | no | client |
| title | `title` | `String` | no | client |
| content | `content` | `String` | no | client |
| summary | `summary` | `String?` | yes | client |
| source | `source` | `MemorySource` | no | client |
| sourceId | `source_id` | `String?` | yes | client |
| confidence | `confidence` | `Confidence` | no | client |
| importance | `importance` | `Importance` | no | client |
| tags | `tags` | `List<String>` | no | client |
| visibility | `visibility` | `MemoryVisibility` | no | client |
| status | `status` | `MemoryStatus` | no | both |
| createdAt | `created_at` | `DateTime` | no | server |
| updatedAt | `updated_at` | `DateTime` | no | server |
| lastAccessedAt | `last_accessed_at` | `DateTime?` | yes | server |

### 4.21 `MemoryListResponse`

| Dart Field | JSON Key | Dart Type | Nullable | Ownership |
| --- | --- | --- | --- | --- |
| memories | `memories` | `List<Memory>` | no | server |
| nextCursor | `next_cursor` | `String?` | yes | server |

### 4.22 `DeviceToken`

| Dart Field | JSON Key | Dart Type | Nullable | Ownership |
| --- | --- | --- | --- | --- |
| id | `id` | `String` | no | server |
| userId | `user_id` | `String` | no | server |
| platform | `platform` | `DevicePlatform` | no | client |
| tokenRedacted | `token_redacted` | `String` | no | server |
| enabled | `enabled` | `bool` | no | server |
| createdAt | `created_at` | `String` | no | server (preformatted) |
| updatedAt | `updated_at` | `String` | no | server |
| lastSeenAt | `last_seen_at` | `String` | no | server |

### 4.23 `DeviceTokenDeleteResponse`

| Dart Field | JSON Key | Dart Type | Nullable | Ownership |
| --- | --- | --- | --- | --- |
| status | `status` | `String` | no | server (`deleted`) |

### 4.24 `ToolDefinition`

| Dart Field | JSON Key | Dart Type | Nullable | Ownership |
| --- | --- | --- | --- | --- |
| id | `id` | `String` | no | server |
| name | `name` | `String` | no | server |
| description | `description` | `String` | no | server |
| category | `category` | `String` | no | server |
| inputSchema | `input_schema` | `Map<String,dynamic>` | no | server |
| outputSchema | `output_schema` | `Map<String,dynamic>?` | yes | server |
| requiresConfirmation | `requires_confirmation` | `bool` | no | server |
| enabled | `enabled` | `bool` | no | server |
| minAutonomyLevel | `min_autonomy_level` | `AutonomyLevel` | no | server |
| createdAt | `created_at` | `DateTime` | no | server |
| updatedAt | `updated_at` | `DateTime` | no | server |

### 4.25 `ToolsListResponse`

| Dart Field | JSON Key | Dart Type | Nullable | Ownership |
| --- | --- | --- | --- | --- |
| tools | `tools` | `List<ToolDefinition>` | no | server |

### 4.26 `AiActionProposal`

| Dart Field | JSON Key | Dart Type | Nullable | Ownership |
| --- | --- | --- | --- | --- |
| id | `id` | `String` | no | server |
| userId | `user_id` | `String` | no | server |
| toolName | `tool_name` | `String` | no | client |
| status | `status` | `ProposalStatus` | no | server |
| proposedInput | `proposed_input` | `Map<String,dynamic>` | no | client |
| reason | `reason` | `String` | no | client |
| riskLevel | `risk_level` | `RiskLevel` | no | client |
| requiresConfirmation | `requires_confirmation` | `bool` | no | server |
| expiresAt | `expires_at` | `DateTime?` | yes | client/server |
| createdAt | `created_at` | `DateTime` | no | server |
| updatedAt | `updated_at` | `DateTime` | no | server |
| confirmedAt | `confirmed_at` | `DateTime?` | yes | server |
| rejectedAt | `rejected_at` | `DateTime?` | yes | server |
| executedAt | `executed_at` | `DateTime?` | yes | server |
| executionResult | `execution_result` | `Map<String,dynamic>?` | yes | server |
| executionError | `execution_error` | `String?` | yes | server |

### 4.27 `ActionProposalsListResponse`

| Dart Field | JSON Key | Dart Type | Nullable | Ownership |
| --- | --- | --- | --- | --- |
| actionProposals | `action_proposals` | `List<AiActionProposal>` | no | server |

### 4.28 `RuntimeResponse` (+ nested mínimos)

| Dart Field | JSON Key | Dart Type | Nullable | Ownership |
| --- | --- | --- | --- | --- |
| mode | `mode` | `RuntimeMode` | no | server (`dry_run`\|`ready`) |
| requestId | `request_id` | `String` | no | server |
| contextSummary | `context_summary` | `Map<String,dynamic>` o typed | no | server |
| availableTools | `available_tools` | `List<Map<String,dynamic>>` / typed | no | server |
| proposedActions | `proposed_actions` | `List<RuntimeProposedAction>` | no | server |
| assistantMessage | `assistant_message` | `String` | no | server |
| observability | `observability` | `RuntimeObservability` | no | server |

### 4.29 `RuntimeProposedAction`

| Dart Field | JSON Key | Dart Type | Nullable | Ownership |
| --- | --- | --- | --- | --- |
| id | `id` | `String?` | yes | server |
| toolName | `tool_name` | `String` | no | server |
| status | `status` | `String` | no | server |
| proposedInput | `proposed_input` | `Map<String,dynamic>` | no | server |
| reason | `reason` | `String` | no | server |
| riskLevel | `risk_level` | `RiskLevel` | no | server |
| requiresConfirmation | `requires_confirmation` | `bool` | no | server |

### 4.30 `RuntimeObservability`

| Dart Field | JSON Key | Dart Type | Nullable | Ownership |
| --- | --- | --- | --- | --- |
| providerLatencyMs | `provider_latency_ms` | `int` | no | server |
| providerErrorType | `provider_error_type` | `String?` | yes | server |
| plannedActionsCount | `planned_actions_count` | `int` | no | server |
| proposalsCount | `proposals_count` | `int` | no | server |
| rejectedActionsCount | `rejected_actions_count` | `int` | no | server |

### 4.31 Explicitly NOT canonical

| Name | Reason |
| --- | --- |
| `PagedResponse` with `items` | Backend no usa `items` |
| `UserDataExport` | BLOCKED — domain dump / casing |
| Reminder “full domain” | Campos no expuestos en HTTP |

---

## 5. Canonical Request Models

### Auth

**`RegisterRequest`:** `name`, `email`, `password` (min 8) — todos required.  
**`LoginRequest`:** `email`, `password`.

### Users PATCH — omit unchanged keys

**`UpdateProfileRequest`** (all optional):  
`display_name`, `preferred_name`, `timezone`, `locale`, `avatar_url`  
→ omit keys not edited. Do not send `user_id`. Empty timezone/locale → 400.

**`UpdateAiSettingsRequest`** (all optional):  
`assistant_name`, `communication_style`, `proactivity_level`, `autonomy_level`, `memory_enabled`, `reminders_enabled`, `planning_enabled`, `preferred_language`.

**Onboarding complete:** no body required (handler ignores body).

**`DeleteRequestBody`:** optional `{ "reason": "..." }` or empty body when Content-Length 0.

### Activities

**`CreateActivityRequest`:**  
Required efectivo: `title`, `timezone`.  
Optional: `type` (default task), `description`, `priority` (default medium), `start_at`/`due_at`/`reminder_at` (RFC3339 strings or omit), `tags`, `has_reminder`, `recurrence_rule`, `ai_manageable`, `ai_notes`.  
Never: `user_id`, `id`, timestamps server.

**`UpdateActivityRequest`:** same fields as pointers + `status`; **omit unchanged**.

### Reminders

**`CreateReminderRequest`:**  
Required: `title`, `scheduled_at` (RFC3339), `timezone`.  
Optional: `activity_id`, `description`, `recurrence_rule`, `recurrence_interval`, `recurrence_until`, `recurrence_count`, `max_retries`.  
Note: recurrence_* may be accepted on create but **not returned** on GET — do not rely on round-trip.

**`UpdateReminderRequest`:** optional subset without `activity_id`; omit unchanged.

### Insights creates

**`CreateMoodEntryRequest`:** `mood`, `energy_level`, `stress_level`, `focus_level`, `occurred_at`, `timezone` required; `activity_id`, `note` optional.  
**`CreateActivityOutcomeRequest`:** `activity_id`, `result`, levels 1..5, `note?`.  
**`CreateReflectionRequest`:** `content`, `occurred_at` required; `activity_id`, `title`, `sentiment`, `tags` optional.

### Memory

**`CreateMemoryRequest`:** `title`, `content` required; defaults type=`fact`, source=`user_manual`, confidence/importance=`medium`, visibility=`assistant_context`.  
**`UpdateMemoryRequest`:** optional + `status`; omit unchanged.  
**`SearchMemoryRequest`:** `query` and/or filters; `limit?`.

### Notifications

**`RegisterDeviceTokenRequest`:** `platform`, `token` (full FCM token — never log).

### AI

**`CreateActionProposalRequest`:** `tool_name`, `proposed_input` (object), `reason`, `risk_level`, `expires_at?`.  
**`RuntimeMessageRequest`:** `message` required; **`dry_run` must be set explicitly** (omit → `false`). Prefer `dry_run: true` for smoke.

---

## 6. Canonical Enums

Estrategia unknown (cerrados):

- **Parse:** prefer `enum` + `@JsonValue` / `fromJson` que mapee unknown → **`unknown` sentinel** o conserve `raw` en campo paralelo.
- **Recomendación Sofia:** `Unknown` case + keep `rawValue` string; **no throw** en response parse (evita crash app).  
- **Request encode:** never send unknown; only backend-accepted values.

| Dart Enum | Backend values (exact) | Unknown strategy |
| --- | --- | --- |
| `ProactivityLevel` | `low`, `medium`, `high` | unknown+raw |
| `AutonomyLevel` | `manual`, `suggestive`, `semi_autonomous` | unknown+raw |
| `DeleteRequestStatus` | `pending`, `approved`, `processing`, `completed`, `failed`, `rejected` | unknown+raw |
| `ActivityType` | `task`, `habit`, `routine`, `goal`, `event`, `checklist` | unknown+raw |
| `ActivityStatus` | `pending`, `active`, `completed`, `canceled`, `archived` | unknown+raw |
| `ActivityPriority` | `low`, `medium`, `high`, `urgent` | unknown+raw |
| `ActivitySource` | `user`, `ai_suggested`, `ai_created`, `imported` | unknown+raw |
| `CreatedBy` | `user`, `sofia` | unknown+raw |
| `ReminderStatus` | `pending`, `due`, `delivered`, `failed`, `canceled`, `archived` | unknown+raw |
| `ReminderSource` | `user`, `activity`, `ai_suggested`, `ai_created` | unknown+raw |
| `RecurrenceRule` | `none`, `daily`, `weekly`, `monthly` | unknown+raw |
| `Mood` | `very_low`, `low`, `neutral`, `good`, `great` | unknown+raw |
| `OutcomeResult` | `completed`, `partially_completed`, `skipped`, `failed`, `rescheduled` | unknown+raw |
| `MemoryType` | `preference`, `fact`, `episodic`, `semantic`, `instruction`, `relationship`, `project`, `system_note` | unknown+raw |
| `MemorySource` | `user_manual`, `conversation`, `activity`, `reflection`, `imported`, `ai_inferred` | unknown+raw |
| `MemoryVisibility` | `private`, `assistant_context`, `archived` | unknown+raw |
| `MemoryStatus` | `active`, `archived`, `deleted` | unknown+raw |
| `Confidence` | `low`, `medium`, `high` | unknown+raw |
| `Importance` | `low`, `medium`, `high`, `critical` | unknown+raw |
| `DevicePlatform` | `ios`, `android`, `web` | unknown+raw |
| `ProposalStatus` | `proposed`, `confirmed`, `rejected`, `executed`, `failed`, `expired` | unknown+raw |
| `RiskLevel` | `low`, `medium`, `high` | unknown+raw |
| `RuntimeMode` | `dry_run`, `ready` | unknown+raw (**not** `proposal`) |

**Not enums (String):** `communication_style`, `assistant_name`, `preferred_language`, `sentiment`, activity `recurrence_rule`, tool `category`/`name`.

---

## 7. Response Wrappers

| Endpoint | Wrapper Key | List Key | Cursor Key |
| --- | --- | --- | --- |
| Most resource create/get/patch/actions | **none (flat root)** | — | — |
| `GET /users/me` | nested object | — | — |
| `GET /activities` | — | `activities` | `next_cursor` |
| `GET /reminders`, `/reminders/due` | — | `reminders` | `next_cursor` |
| `GET /memory`, `POST /memory/search` | — | `memories` | `next_cursor` |
| `GET /insights/mood` | — | `mood_entries` | — |
| `GET /insights/outcomes` | — | **`outcomes`** | — |
| `GET /insights/reflections` | — | `reflections` | — |
| `GET /tools` | — | `tools` | — |
| `GET /ai/action-proposals` | — | `action_proposals` | — |
| `GET /notifications/device-tokens` | **raw JSON array** | — | — |
| `DELETE .../device-tokens/{id}` | — | — | ack `status` |
| Errors | — | — | `message` |

**Forbidden assumption:** `{ "activity": {...} }`, `{ "profile": {...} }`, `{ "items": [...] }`, `{ "device_tokens": [...] }`, `{ "activity_outcomes": [...] }`.

---

## 8. Error Handling Contract

| Status | Flutter Semantic | Required Action |
| --- | --- | --- |
| **400** | Validation / bad body / unknown JSON field | Show `message`; keep form state; do not logout |
| **401** | Unauthenticated / invalid token / bad login | Clear token; navigate login; treat message as generic (`invalid credentials`) |
| **403** | Forbidden (e.g. activity not owned on some insights/reminders paths) | Show forbidden UX; do not treat as logout |
| **404** | Not found or cross-user resource | Show not found; do not reveal ownership |
| **409** | Conflict: duplicate email; **feature disabled** (memory/reminders); proposal transition; unsafe runtime | Map disabled toggles to settings UX; conflict recoverable |
| **429** | Rate limited (auth) | Backoff / retry later message |
| **500** | Server error | Generic error; log `message` without PII |
| **202** | Accepted (delete-request) | Success pending — not immediate delete |
| Non-JSON 404 | Bad path | Treat as transport/not-found; do not parse as `ApiError` |

---

## 9. Models Safe For Sprint F1

Exact set:

1. `ApiError` / `ApiException`  
2. `HealthResponse`  
3. `AuthUser`  
4. `RegisterRequest` / `LoginRequest`  
5. `LoginResponse`  
6. `UserProfile`  
7. `AiSettings`  
8. `MeResponse`  
9. `UpdateProfileRequest`  
10. `UpdateAiSettingsRequest`  
11. Enums: `ProactivityLevel`, `AutonomyLevel`  

Plus infra (not domain models): `AppConfig`, `TokenStorage`, `ApiClient`.

**Out of F1 models:** Activity/Reminder/Insights/Memory/AI/Notifications/Export/DeleteRequest UI models (DeleteRequest optional later).

---

## 10. Models Deferred To F2

- `Activity`, `CreateActivityRequest`, `UpdateActivityRequest`, `ActivityListResponse`  
- `Reminder` (HTTP), `CreateReminderRequest`, `UpdateReminderRequest`, `ReminderListResponse`  
- Enums: Activity*, Reminder*, `RecurrenceRule`, `CreatedBy`  
- Pagination helpers (cursor) — implement defensively  

---

## 11. Models Deferred To F3

- `MoodEntry`, `CreateMoodEntryRequest`, `MoodEntriesListResponse`  
- `ActivityOutcome`, `CreateActivityOutcomeRequest`, `ActivityOutcomesListResponse`  
- `Reflection`, `CreateReflectionRequest`, `ReflectionsListResponse`  
- `InsightsSummary`  
- `Memory`, create/update/search, `MemoryListResponse`  
- Enums Mood, OutcomeResult, Memory*, Confidence, Importance  

---

## 12. Models Deferred To F4

- `AiActionProposal`, `CreateActionProposalRequest`, `ActionProposalsListResponse`  
- `RuntimeMessageRequest`, `RuntimeResponse`, `RuntimeProposedAction`, `RuntimeObservability`  
- `ToolDefinition`, `ToolsListResponse` (debug)  
- Enums `ProposalStatus`, `RiskLevel`, `RuntimeMode`  
- Feature flags: `dry_run=true` default client; execute gated  

---

## 13. Models Deferred To F5

- `DeviceToken`, `RegisterDeviceTokenRequest`, `DeviceTokenDeleteResponse`  
- `DevicePlatform`  
- `DeleteRequest`, `DeleteRequestBody` (privacy)  
- **Still BLOCKED:** `UserDataExport` until backend clarification  

---

## 14. Blockers and Backend Clarifications

| Severity | Domain | Issue | Required Before |
| --- | --- | --- | --- |
| BLOCKER | privacy export | Domain structs without json tags → likely PascalCase mix; shape ≠ docs | F5 export UI / any export model |
| HIGH | docs vs wire | Historical docs use resource wrappers | Ignore old docs; this file wins |
| HIGH | pagination | `next_cursor` untested; sprint12 shows wrong `items` | F2 list infinite scroll |
| MEDIUM | outcomes list | Key `outcomes` untested HTTP success | F3 outcomes history |
| MEDIUM | action proposals list | List endpoint untested | F4 inbox |
| MEDIUM | PATCH reminder | No HTTP test | F2 reminder editor |
| MEDIUM | notifications HTTP | No handler tests; list is raw array | F5 |
| MEDIUM | runtime HTTP success | Only 401 in routes_test | F4 chat |
| LOW | delete-request | No 409 on duplicate; body assert weak | F5 privacy |
| LOW | non-JSON 404 | `http.NotFound` | Robust ApiClient |
| INFO | feature disabled | 409 not 403 | All phases with toggles |

**NEEDS_BACKEND_CLARIFICATION:** export serialization (DTO HTTP snake_case vs current dump).

---

## 15. F1 Implementation Readiness

| Element | Status |
| --- | --- |
| ApiClient | READY_TO_IMPLEMENT |
| ApiException | READY_TO_IMPLEMENT |
| AppConfig | READY_TO_IMPLEMENT |
| TokenStorage | READY_TO_IMPLEMENT |
| AuthUser | READY_TO_IMPLEMENT |
| LoginResponse | READY_TO_IMPLEMENT |
| UserProfile | READY_TO_IMPLEMENT |
| AiSettings | READY_TO_IMPLEMENT |
| MeResponse | READY_TO_IMPLEMENT |
| HealthResponse | READY_TO_IMPLEMENT |
| AuthRepository | READY_TO_IMPLEMENT |
| UserRepository | READY_TO_IMPLEMENT |
| SystemRepository | READY_TO_IMPLEMENT |
| SessionCubit | READY_TO_IMPLEMENT |
| UserDataExport model | **BLOCKED** |
| Privacy export repository | **BLOCKED** |
| Activity* models | Deferred F2 (not F1 blocker) |

F1 also needs: `RegisterRequest`, `LoginRequest`, `UpdateProfileRequest`, `UpdateAiSettingsRequest`, onboarding call (no body model required).

---

## 16. Final Sprint F0 Verdict

### 1. ¿Sprint F0 puede cerrarse?

**Sí.** HTTP verification, schema inventory, test evidence y este contrato canónico forman el paquete F0. Documentos históricos se conservan.

### 2. ¿F1 puede comenzar?

**Sí**, inmediatamente para Core/Auth/Users/Health/Session, usando **este** contrato (flat responses, `MeResponse` nested).

### 3. ¿Qué blockers existen?

- **Export** (`UserDataExport`) — BLOCKED.  
- No blockers duros para ApiClient + Auth + Me + Profile + AI settings + Health.  
- No generar wrappers documentados antiguos.

### 4. ¿Cuál es el set exacto de modelos F1?

```
ApiException (+ ApiError body)
HealthResponse
AuthUser
RegisterRequest
LoginRequest
LoginResponse
MeResponse
UserProfile
AiSettings
UpdateProfileRequest
UpdateAiSettingsRequest
ProactivityLevel
AutonomyLevel
```

Infra: `AppConfig`, `TokenStorage`, `ApiClient`, repos Auth/User/System, `SessionCubit` (+ Auth/Onboarding/Profile/AiSettings/Health cubits según plan).

---

## Appendix A — Agent implementation notes

1. Prefer `json_serializable` with explicit snake keys.  
2. PATCH: `Map` builders or `includeIfNull: false`.  
3. On 401 from any authenticated call → `SessionCubit.logout()`.  
4. Gate UI with `AiSettings.memoryEnabled` / `remindersEnabled` / `planningEnabled`; expect **409** if ignored.  
5. Do not implement smart home against this API.  
6. Runtime: always send `"dry_run": true` until F4 flag.  
7. When docs conflict with this file → **this file wins**.

## Appendix B — Supersedes (operative)

| Old guidance | Canonical replacement |
| --- | --- |
| `{ "activity": {...} }` responses | Flat `Activity` |
| `{ "profile": {...} }` on PATCH | Flat `UserProfile` |
| `{ "ai_settings": {...} }` on GET/PATCH settings | Flat `AiSettings` |
| `{ "items": [...] }` | Named list keys |
| `{ "device_tokens": [...] }` | Raw array |
| `{ "activity_outcomes": [...] }` | `{ "outcomes": [...] }` |
| Runtime mode `proposal` | `ready` |
| Feature disabled 403 | **409** for memory/reminders |
| Reminder full domain fields | HTTP subset only |
| `PagedResponse.items` | Per-domain list responses |
| `UserDataExport` ready | **BLOCKED** |

*Fin contrato canónico v0.1.*
