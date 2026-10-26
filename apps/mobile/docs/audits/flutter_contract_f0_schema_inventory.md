# Flutter F0 Schema Inventory

Fecha: 2026-07-10  
Sprint: F0 — Flutter Contract Verification (schema / type-system)  
Fuente de verdad: DTOs HTTP y domain structs en `sophia_ai_backend/sofia-backend`  
Referencias (no autoridad): `docs/backend_contract/flutter_models_needed.md`, `flutter_endpoint_integration_contract.md`, `docs/audits/flutter_contract_f0_http_verification.md`

**Solo auditoría.** No se modificó backend, Flutter, Firestore, tags JSON ni se generó Dart.

---

## 1. Executive Summary

Inventario de tipos que Flutter debe modelar según **responses/requests HTTP reales**, no según la documentación Flutter.

### Hallazgos clave

1. **Response envelopes:** la mayoría de recursos salen **en la raíz** (sin `{"activity":...}`). Las listas usan claves nombradas (`activities`, `reminders`, `memories`, `outcomes`, …). **No existe** wrapper genérico `items`.
2. **Enums documentados ≈ backend** en valores string (coinciden en casi todos). Excepciones de contrato:
   - Runtime `mode`: docs `dry_run|proposal` → real `dry_run|ready`.
   - `communication_style` y `sentiment` / `preferred_language` / `assistant_name`: **string libre** (no enum).
3. **Reminder HTTP DTO** es un **subset** del dominio: no expone `recurrence_interval`, `recurrence_until`, `recurrence_count`, `recurrence_occurrences`, `processing_by`, `processing_until`, `last_attempt_at`.
4. **DeviceToken timestamps** en HTTP son **string** formateados, no `time.Time` JSON nativo.
5. **UserDataExport** es **LOW confidence**: mezcla domain structs sin tags (PascalCase probable) + shapes distintas a docs.
6. **AuthUser** tiene **tres shapes** según endpoint (register/login vs `/users/me` vs export).

### Conteos canónicos recomendados

| Categoría | Cantidad |
| --- | ---: |
| Modelos Dart canónicos (feature-ready, excl. export ambiguo) | **38** |
| Modelos ambiguos / diferir | **3** (`UserDataExport`, generic `PagedResponse<items>`, Reminder “full domain”) |
| Enums Dart canónicos | **22** (+ 1 opcional `RuntimeMode`) |

---

## 2. Canonical Model Inventory

| Flutter Model | Backend Source | Confidence | Safe To Generate |
| --- | --- | --- | --- |
| `ApiError` | `httpjson.ErrorResponse` | HIGH | YES |
| `HealthResponse` | `server.HealthResponse` | HIGH | YES |
| `ActivityListResponse` | `listActivitiesResponse` | HIGH | YES |
| `ReminderListResponse` | `remindersListResponse` | HIGH | YES |
| `MemoryListResponse` | `memoriesListResponse` | HIGH | YES |
| `MoodEntriesListResponse` | `moodEntriesListResponse` | HIGH | YES |
| `ActivityOutcomesListResponse` | `activityOutcomesListResponse` (`outcomes`) | HIGH | YES |
| `ReflectionsListResponse` | `reflectionsListResponse` | HIGH | YES |
| `ToolsListResponse` | `map[string]any{"tools":...}` | HIGH | YES |
| `ActionProposalsListResponse` | `{"action_proposals":...}` | HIGH | YES |
| `DeviceTokenDeleteResponse` | `map[string]string{"status":"deleted"}` | HIGH | YES |
| `AuthUser` (login/register) | `UserResponse` | HIGH | YES |
| `MeUser` / `AuthUser` con `role` | `userResponse` in users HTTP | HIGH | YES |
| `LoginResponse` | `LoginResponse` | HIGH | YES |
| `UserProfile` | `profileResponse` | HIGH | YES |
| `AiSettings` | `aiSettingsResponse` | HIGH | YES |
| `MeResponse` | `meResponse` | HIGH | YES |
| `UpdateProfileRequest` | `updateProfileRequest` | HIGH | YES |
| `UpdateAiSettingsRequest` | `updateAISettingsRequest` | HIGH | YES |
| `DeleteRequest` | `privacy.domain.DeleteRequest` | HIGH | YES |
| `DeleteRequestBody` | `deleteRequestBody` | HIGH | YES |
| `Activity` | `activityResponse` | HIGH | YES |
| `CreateActivityRequest` | `createActivityRequest` | HIGH | YES |
| `UpdateActivityRequest` | `updateActivityRequest` | HIGH | YES |
| `Reminder` (HTTP) | `reminderResponse` | HIGH | YES |
| `CreateReminderRequest` | `createReminderRequest` | HIGH | YES |
| `UpdateReminderRequest` | `updateReminderRequest` | HIGH | YES |
| `MoodEntry` | `moodEntryResponse` | HIGH | YES |
| `CreateMoodEntryRequest` | `createMoodEntryRequest` | HIGH | YES |
| `ActivityOutcome` | `activityOutcomeResponse` | HIGH | YES |
| `CreateActivityOutcomeRequest` | `createActivityOutcomeRequest` | HIGH | YES |
| `Reflection` | `reflectionResponse` | HIGH | YES |
| `CreateReflectionRequest` | `createReflectionRequest` | HIGH | YES |
| `InsightsSummary` | `summaryResponse` | HIGH | YES |
| `Memory` | `memoryResponse` | HIGH | YES |
| `CreateMemoryRequest` | `createMemoryRequest` | HIGH | YES |
| `UpdateMemoryRequest` | `updateMemoryRequest` | HIGH | YES |
| `SearchMemoryRequest` | `searchMemoryRequest` | HIGH | YES |
| `DeviceToken` | `deviceTokenResponse` | HIGH | YES |
| `RegisterDeviceTokenRequest` | `registerDeviceTokenRequest` | HIGH | YES |
| `ToolDefinition` | `tools.domain.ToolDefinition` | HIGH | YES |
| `AiActionProposal` | `actions.domain.AIActionProposal` | HIGH | YES |
| `CreateActionProposalRequest` | `createProposalRequest` | HIGH | YES |
| `RuntimeMessageRequest` | `messageRequest` | HIGH | YES |
| `RuntimeResponse` | `runtime.domain.RuntimeResponse` | HIGH | YES |
| `RuntimeProposedAction` | `ActionProposalOutput` | HIGH | YES |
| `RuntimeContextSummary` (+ nested) | `ContextSummary` et al. | MEDIUM | YES (nested) |
| `RuntimeObservability` | `RuntimeObservability` | HIGH | YES |
| `PagedResponse<T>` con `items` | — | LOW | **NO** |
| `UserDataExport` | `privacy.application.UserDataExport` | LOW | **NO** |
| `Reminder` full-domain (docs) | domain `Reminder` | LOW | **NO** (usar HTTP subset) |

---

## 3. Field Matrix

Leyenda columnas: **ReqC** = required en create/request; **ReqR** = presente en response exitoso típico; **Null** = puede ser JSON null; **Omit** = puede ausentarse (`omitempty` o no enviado); **Def** = default backend.

### 3.1 `ApiError`

Fuente: `internal/platform/httpjson/httpjson.go` → `ErrorResponse`

| JSON Key | Go Type | JSON Type | Dart Type | ReqC | ReqR | Null | Omit | Default | Validation |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `message` | `string` | string | `String` | — | Y | N | N | — | texto libre |

Nota: `statusCode` / `rawBody` son **client-side** (no vienen en body). Modelar en Flutter aparte del JSON.

### 3.2 `HealthResponse`

Fuente: `internal/server/health.go`

| JSON Key | Go Type | JSON Type | Dart Type | ReqC | ReqR | Null | Omit | Default | Validation |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `status` | `string` | string | `String` | — | Y | N | N | `ok` o `degraded` | si firestore=`error` → `degraded` |
| `environment` | `string` | string | `String` | — | Y | N | N | config | — |
| `firestore` | `string` | string | `String` | — | Y | N | N | `ok\|disabled\|error` | — |

### 3.3 Pagination (no modelo `items`)

Patrón real: lista nombrada + `next_cursor` opcional.

| JSON Key | Go Type | JSON Type | Dart Type | Notes |
| --- | --- | --- | --- | --- |
| `<collection>` | `[]T` | array | `List<T>` | siempre presente (puede ser `[]`) |
| `next_cursor` | `string` | string | `String?` | `omitempty`; = último `id` si len>0 |

Collections: `activities`, `reminders`, `memories`. Insights lists **sin** cursor. Action proposals list **sin** cursor.

### 3.4 Auth — `AuthUser` / `LoginResponse`

#### Register/Login user — `internal/auth/interfaces/http/responses.go` `UserResponse`

| JSON Key | Go Type | JSON Type | Dart Type | ReqC | ReqR | Null | Omit | Default | Validation |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `id` | `string` | string | `String` | — | Y | N | N | uuid | — |
| `name` | `string` | string | `String` | Y (register) | Y | N | N | — | non-empty |
| `email` | `string` | string | `String` | Y | Y | N | N | normalized | email rules |
| `password` | — | — | — | Y register/login | **never** | — | — | — | min 8 |

#### Login — `LoginResponse`

| JSON Key | Go Type | JSON Type | Dart Type | ReqR |
| --- | --- | --- | --- | --- |
| `token` | `string` | string | `String` | Y |
| `user` | `UserResponse` | object | `AuthUser` | Y |

#### Me user — `internal/users/interfaces/http/responses.go` `userResponse`

| JSON Key | Go Type | JSON Type | Dart Type | ReqR | Omit |
| --- | --- | --- | --- | --- | --- |
| `id` | `string` | string | `String` | Y | N |
| `name` | `string` | string | `String` | Y | N |
| `email` | `string` | string | `String` | Y | N |
| `role` | `string` | string | `String` | Y | N |

Recomendación: un `AuthUser` con `role` opcional; register/login no lo envían.

Password: domain `User.Password` tiene `json:"-"` — **nunca** en JSON.

### 3.5 `UserProfile`

Fuente HTTP: `profileResponse` — `internal/users/interfaces/http/responses.go`  
Request PATCH: `updateProfileRequest` — `requests.go`

#### Response

| JSON Key | Go Type | JSON Type | Dart Type | ReqR | Null | Omit | Default | Validation |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `id` | `string` | string | `String` | Y | N | N | =userID | — |
| `user_id` | `string` | string | `String` | Y | N | N | JWT user | **no enviar en write** |
| `display_name` | `string` | string | `String` | Y | N | N | register name | — |
| `preferred_name` | `string` | string | `String` | Y | N | N | =display | — |
| `timezone` | `string` | string | `String` | Y | N | N | `America/Tijuana` | non-empty |
| `locale` | `string` | string | `String` | Y | N | N | `es-MX` | non-empty |
| `avatar_url` | `string` | string | `String?` | N | N* | Y | `""` | omitempty si vacío |
| `onboarding_completed` | `bool` | bool | `bool` | Y | N | N | `false` | — |
| `created_at` | `time.Time` | string RFC3339 | `DateTime` | Y | N | N | server | — |
| `updated_at` | `time.Time` | string RFC3339 | `DateTime` | Y | N | N | server | — |

\*empty string omitido, no null tipico.

#### UpdateProfileRequest (todos pointer → omit = no change)

| JSON Key | Go Type | JSON Type | Dart Type | ReqC | Null | Omit |
| --- | --- | --- | --- | --- | --- | --- |
| `display_name` | `*string` | string\|null | `String?` | N | Y* | Y |
| `preferred_name` | `*string` | string\|null | `String?` | N | Y* | Y |
| `timezone` | `*string` | string\|null | `String?` | N | Y* | Y |
| `locale` | `*string` | string\|null | `String?` | N | Y* | Y |
| `avatar_url` | `*string` | string\|null | `String?` | N | Y* | Y |

\*JSON `null` → Go nil → no change (igual que omit). `""` en timezone/locale → 400.

### 3.6 `AiSettings`

Fuente: `aiSettingsResponse` / `updateAISettingsRequest`

| JSON Key | Go Type | JSON Type | Dart Type | ReqR | Omit | Default | Validation |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `user_id` | `string` | string | `String` | Y | N | JWT | read-only |
| `assistant_name` | `string` | string | `String` | Y | N | `Sofía` | non-empty; **free string** |
| `communication_style` | `string` | string | `String` | Y | N | `warm` | **free string** (no enum) |
| `proactivity_level` | `string` | string | `ProactivityLevel` | Y | N | `medium` | enum |
| `autonomy_level` | `string` | string | `AutonomyLevel` | Y | N | `suggestive` | enum |
| `memory_enabled` | `bool` | bool | `bool` | Y | N | `true` | — |
| `reminders_enabled` | `bool` | bool | `bool` | Y | N | `true` | — |
| `planning_enabled` | `bool` | bool | `bool` | Y | N | `true` | — |
| `preferred_language` | `string` | string | `String` | Y | N | `es` | non-empty; free |
| `created_at` | `time.Time` | RFC3339 | `DateTime` | Y | N | server | — |
| `updated_at` | `time.Time` | RFC3339 | `DateTime` | Y | N | server | — |

Update request: mismos campos como `*string` / `*bool`.

### 3.7 `MeResponse`

| JSON Key | Go Type | JSON Type | Dart Type | ReqR |
| --- | --- | --- | --- | --- |
| `user` | `userResponse` | object | `AuthUser` | Y |
| `profile` | `profileResponse` | object | `UserProfile` | Y |
| `ai_settings` | `aiSettingsResponse` | object | `AiSettings` | Y |

### 3.8 `DeleteRequest`

Fuente: `internal/privacy/domain/privacy.go` (serializado directo)

| JSON Key | Go Type | JSON Type | Dart Type | ReqR | Omit | Default |
| --- | --- | --- | --- | --- | --- | --- |
| `id` | `string` | string | `String` | Y | N | uuid |
| `user_id` | `string` | string | `String` | Y | N | JWT |
| `status` | `string` | string | `DeleteRequestStatus` | Y | N | `pending` |
| `reason` | `string` | string | `String?` | N | Y | `""` omitempty |
| `created_at` | `time.Time` | RFC3339 | `DateTime` | Y | N | server |
| `updated_at` | `time.Time` | RFC3339 | `DateTime` | Y | N | server |

Body create: `{ "reason": string }` opcional.

### 3.9 `UserDataExport` — LOW / DO NOT GENERATE YET

Fuente: `internal/privacy/application/service.go` `UserDataExport`

| JSON Key | Go embedded | Risk |
| --- | --- | --- |
| `generated_at` | `time.Time` | OK RFC3339 |
| `user` | `*auth.User` (tags snake; password `-`) | OK-ish |
| `profile` | `*users.UserProfile` **sin json tags** | PascalCase risk |
| `ai_settings` | `*users.AISettings` **sin tags** | PascalCase risk |
| `activities` | `[]*activities.Activity` **sin tags** | PascalCase risk |
| `reminders` | `[]*reminders.Reminder` **sin tags** | PascalCase + internal fields |
| `mood_entries` | insights domain **sin tags** | PascalCase |
| `activity_outcomes` | insights domain | PascalCase |
| `reflections` | insights domain | PascalCase |
| `memories` | memory domain **sin tags** | PascalCase |
| `action_proposals` | `AIActionProposal` **con tags** | snake_case OK |

Docs `insights: {...}` **no existe**.

### 3.10 `Activity`

Fuente: `internal/activities/interfaces/http/responses.go` `activityResponse`

| JSON Key | Go Type | JSON Type | Dart Type | ReqR | Null | Omit | Default | Validation |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `id` | `string` | string | `String` | Y | N | N | uuid | — |
| `user_id` | `string` | string | `String` | Y | N | N | JWT | read-only |
| `type` | `string` | string | `ActivityType` | Y | N | N | `task` | enum |
| `title` | `string` | string | `String` | Y | N | N | — | required |
| `description` | `string` | string | `String` | Y | N | N | `""` | — |
| `status` | `string` | string | `ActivityStatus` | Y | N | N | `pending` | enum |
| `priority` | `string` | string | `ActivityPriority` | Y | N | N | `medium` | enum |
| `start_at` | `*time.Time` | string\|omit | `DateTime?` | N | Y | Y | nil | RFC3339 |
| `due_at` | `*time.Time` | string\|omit | `DateTime?` | N | Y | Y | nil | RFC3339 |
| `completed_at` | `*time.Time` | string\|omit | `DateTime?` | N | Y | Y | nil | server |
| `canceled_at` | `*time.Time` | string\|omit | `DateTime?` | N | Y | Y | nil | server |
| `archived_at` | `*time.Time` | string\|omit | `DateTime?` | N | Y | Y | nil | server |
| `timezone` | `string` | string | `String` | Y | N | N | — | required |
| `tags` | `[]string` | array | `List<String>` | Y | N* | N | `[]` | — |
| `source` | `string` | string | `ActivitySource` | Y | N | N | `user` | enum |
| `created_by` | `string` | string | `CreatedBy` | Y | N | N | `user` | enum |
| `created_at` | `time.Time` | RFC3339 | `DateTime` | Y | N | N | server | — |
| `updated_at` | `time.Time` | RFC3339 | `DateTime` | Y | N | N | server | — |
| `has_reminder` | `bool` | bool | `bool` | Y | N | N | `false` | — |
| `reminder_at` | `*time.Time` | string\|omit | `DateTime?` | N | Y | Y | nil | requires has_reminder |
| `recurrence_rule` | `string` | string\|omit | `String?` | N | N | Y | `""` | free/activity-level string |
| `ai_manageable` | `bool` | bool | `bool` | Y | N | N | `false` | — |
| `ai_notes` | `string` | string\|omit | `String?` | N | N | Y | `""` | — |

\*nil slice encodes as `null` en Go si no se normaliza; handler usa `append` copy → suele ser `[]`.

#### CreateActivityRequest

| JSON Key | Go Type | ReqC | Default if omit | Notes |
| --- | --- | --- | --- | --- |
| `type` | `string` | N | `task` | enum |
| `title` | `string` | **Y** | — | required |
| `description` | `string` | N | `""` | — |
| `priority` | `string` | N | `medium` | enum |
| `start_at` | `*string` | N | nil | RFC3339 or empty→nil |
| `due_at` | `*string` | N | nil | — |
| `timezone` | `string` | **Y** | — | required |
| `tags` | `[]string` | N | nil/`[]` | — |
| `has_reminder` | `bool` | N | `false` | — |
| `reminder_at` | `*string` | N | nil | — |
| `recurrence_rule` | `string` | N | `""` | — |
| `ai_manageable` | `bool` | N | `false` | — |
| `ai_notes` | `string` | N | `""` | — |

#### UpdateActivityRequest

Todos `*T` / `**time.Time` para dates; omit/null pointer = no change. Incluye `status`.

#### ActivityListResponse

`activities: List<Activity>`, `next_cursor: String?`

### 3.11 `Reminder` (HTTP subset)

Fuente: `internal/reminders/interfaces/http/responses.go`

| JSON Key | Go Type | JSON Type | Dart Type | ReqR | Omit | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| `id` | `string` | string | `String` | Y | N | — |
| `user_id` | `string` | string | `String` | Y | N | read-only |
| `activity_id` | `string` | string\|omit | `String?` | N | Y | omitempty |
| `title` | `string` | string | `String` | Y | N | required create |
| `description` | `string` | string | `String` | Y | N | — |
| `status` | `string` | string | `ReminderStatus` | Y | N | default pending |
| `scheduled_at` | `time.Time` | RFC3339 | `DateTime` | Y | N | **required create** |
| `timezone` | `string` | string | `String` | Y | N | required |
| `recurrence_rule` | `string` | string\|omit | `RecurrenceRule?` | N | Y | default none |
| `next_run_at` | `time.Time` | RFC3339 | `DateTime` | Y | N | =scheduled initially |
| `last_run_at` | `*time.Time` | string\|omit | `DateTime?` | N | Y | — |
| `delivered_at` | `*time.Time` | string\|omit | `DateTime?` | N | Y | — |
| `failed_at` | `*time.Time` | string\|omit | `DateTime?` | N | Y | — |
| `failure_reason` | `string` | string\|omit | `String?` | N | Y | — |
| `retry_count` | `int` | number | `int` | Y | N | — |
| `max_retries` | `int` | number | `int` | Y | N | default 3 si 0 |
| `source` | `string` | string | `ReminderSource` | Y | N | default user |
| `created_by` | `string` | string | `CreatedBy` | Y | N | default user |
| `created_at` | `time.Time` | RFC3339 | `DateTime` | Y | N | — |
| `updated_at` | `time.Time` | RFC3339 | `DateTime` | Y | N | — |

**NO en HTTP response (sí en domain):** `recurrence_interval`, `recurrence_until`, `recurrence_count`, `recurrence_occurrences`, `processing_by`, `processing_until`, `last_attempt_at`.

#### CreateReminderRequest

| JSON Key | Go Type | ReqC | Default | Validation |
| --- | --- | --- | --- | --- |
| `activity_id` | `string` | N | `""` | must own if set |
| `title` | `string` | **Y** | — | required |
| `description` | `string` | N | `""` | — |
| `scheduled_at` | `string` | **Y** | — | RFC3339 required |
| `timezone` | `string` | **Y** | — | required |
| `recurrence_rule` | `string` | N | `none` | enum |
| `recurrence_interval` | `int` | N | 0 | positive if recurring |
| `recurrence_until` | `string` | N | empty→nil | RFC3339 |
| `recurrence_count` | `int` | N | 0 | ≥0 |
| `max_retries` | `int` | N | 0→3 | — |

Create acepta recurrence fields aunque **response no los devuelve** → Flutter puede enviarlos pero no rehidratarlos del GET.

### 3.12 Insights

#### MoodEntry — `moodEntryResponse`

| JSON Key | Go Type | Dart Type | ReqR | Omit | Validation |
| --- | --- | --- | --- | --- | --- |
| `id` | `string` | `String` | Y | N | — |
| `user_id` | `string` | `String` | Y | N | — |
| `activity_id` | `string` | `String?` | N | Y | optional |
| `mood` | `string` | `Mood` | Y | N | enum |
| `energy_level` | `int` | `int` | Y | N | 1..5 |
| `stress_level` | `int` | `int` | Y | N | 1..5 |
| `focus_level` | `int` | `int` | Y | N | 1..5 |
| `note` | `string` | `String?` | N | Y | — |
| `occurred_at` | `time.Time` | `DateTime` | Y | N | RFC3339 req create |
| `timezone` | `string` | `String` | Y | N | required |
| `created_at`/`updated_at` | `time.Time` | `DateTime` | Y | N | — |

Create request: mismos campos de entrada (sin id/user/timestamps).

#### ActivityOutcome — `activityOutcomeResponse`

| JSON Key | Dart Type | ReqR | Omit | Validation |
| --- | --- | --- | --- | --- |
| `id` | `String` | Y | N | — |
| `user_id` | `String` | Y | N | — |
| `activity_id` | `String` | Y | N | required |
| `result` | `OutcomeResult` | Y | N | enum |
| `completion_quality` | `int` | Y | N | 1..5 |
| `satisfaction_level` | `int` | Y | N | 1..5 |
| `difficulty_level` | `int` | Y | N | 1..5 |
| `note` | `String?` | N | Y | — |
| `created_at`/`updated_at` | `DateTime` | Y | N | — |

List wrapper key: **`outcomes`** (no `activity_outcomes`).

#### Reflection — `reflectionResponse`

| JSON Key | Dart Type | ReqR | Omit | Notes |
| --- | --- | --- | --- | --- |
| `id` | `String` | Y | N | — |
| `user_id` | `String` | Y | N | — |
| `activity_id` | `String?` | N | Y | — |
| `title` | `String?` | N | Y | — |
| `content` | `String` | Y | N | required |
| `sentiment` | `String?` | N | Y | **free string** |
| `tags` | `List<String>` | Y | N | — |
| `occurred_at` | `DateTime` | Y | N | RFC3339 req create |
| `created_at`/`updated_at` | `DateTime` | Y | N | — |

#### InsightsSummary — `summaryResponse`

| JSON Key | Go Type | Dart Type |
| --- | --- | --- |
| `average_energy` | `float64` | `double` |
| `average_stress` | `float64` | `double` |
| `average_focus` | `float64` | `double` |
| `total_outcomes` | `int` | `int` |
| `total_reflections` | `int` | `int` |

### 3.13 Memory

Fuente: `memoryResponse` / requests — `internal/memory/interfaces/http/`

| JSON Key | Go Type | Dart Type | ReqR | Omit | Default create |
| --- | --- | --- | --- | --- | --- |
| `id` | `string` | `String` | Y | N | uuid |
| `user_id` | `string` | `String` | Y | N | JWT |
| `type` | `string` | `MemoryType` | Y | N | `fact` |
| `title` | `string` | `String` | Y | N | required |
| `content` | `string` | `String` | Y | N | required |
| `summary` | `string` | `String?` | N | Y | `""` |
| `source` | `string` | `MemorySource` | Y | N | `user_manual` |
| `source_id` | `string` | `String?` | N | Y | `""` |
| `confidence` | `string` | `Confidence` | Y | N | `medium` |
| `importance` | `string` | `Importance` | Y | N | `medium` |
| `tags` | `[]string` | `List<String>` | Y | N | `[]` |
| `visibility` | `string` | `MemoryVisibility` | Y | N | `assistant_context` |
| `status` | `string` | `MemoryStatus` | Y | N | `active` |
| `created_at`/`updated_at` | `time.Time` | `DateTime` | Y | N | server |
| `last_accessed_at` | `*time.Time` | `DateTime?` | N | Y | nil |

Search request: `query`, `type`, `tag`, `importance`, `limit` (query/filter required by domain).

### 3.14 Notifications

#### DeviceToken — `deviceTokenResponse` (HTTP)

| JSON Key | Go Type | JSON Type | Dart Type | Notes |
| --- | --- | --- | --- | --- |
| `id` | `string` | string | `String` | — |
| `user_id` | `string` | string | `String` | — |
| `platform` | `string` | string | `DevicePlatform` | enum |
| `token_redacted` | `string` | string | `String` | never full token |
| `enabled` | `bool` | bool | `bool` | default true |
| `created_at` | `string` | **string** | `String` or parse `DateTime` | **not** raw time.Time encode |
| `updated_at` | `string` | **string** | idem | Format layout RFC3339-like |
| `last_seen_at` | `string` | **string** | idem | — |

Register: `platform`, `token` (full token solo en POST).

Delete ack: `{ "status": "deleted" }` → `DeviceTokenDeleteResponse`.

List GET: `List<DeviceToken>` **array raíz**.

### 3.15 `ToolDefinition`

Fuente: `internal/tools/domain/tool.go`

| JSON Key | Go Type | Dart Type | Omit |
| --- | --- | --- | --- |
| `id` | `string` | `String` | N |
| `name` | `string` | `String` | N |
| `description` | `string` | `String` | N |
| `category` | `string` | `String` | N |
| `input_schema` | `json.RawMessage` | `Map<String,dynamic>` | N |
| `output_schema` | `json.RawMessage` | `Map<String,dynamic>?` | Y |
| `requires_confirmation` | `bool` | `bool` | N |
| `enabled` | `bool` | `bool` | N |
| `min_autonomy_level` | `string` | `AutonomyLevel` | N |
| `created_at`/`updated_at` | `time.Time` | `DateTime` | N |

List: `{ "tools": [ToolDefinition] }`. Get by name: **flat** ToolDefinition.

### 3.16 `AiActionProposal`

Fuente: `internal/ai/actions/domain/proposal.go`

| JSON Key | Go Type | Dart Type | Omit |
| --- | --- | --- | --- |
| `id` | `string` | `String` | N |
| `user_id` | `string` | `String` | N |
| `tool_name` | `string` | `String` | N |
| `status` | `string` | `ProposalStatus` | N |
| `proposed_input` | `json.RawMessage` | `Map<String,dynamic>` | N |
| `reason` | `string` | `String` | N |
| `risk_level` | `string` | `RiskLevel` | N |
| `requires_confirmation` | `bool` | `bool` | N |
| `expires_at` | `*time.Time` | `DateTime?` | Y |
| `created_at`/`updated_at` | `time.Time` | `DateTime` | N |
| `confirmed_at`/`rejected_at`/`executed_at` | `*time.Time` | `DateTime?` | Y |
| `execution_result` | `json.RawMessage` | `Map<String,dynamic>?` | Y |
| `execution_error` | `string` | `String?` | Y |

Create request: `tool_name`, `proposed_input`, `reason`, `risk_level`, `expires_at?` (string RFC3339).

### 3.17 AI Runtime

#### RuntimeMessageRequest — `messageRequest`

| JSON Key | Go Type | Dart Type | Default omit |
| --- | --- | --- | --- |
| `message` | `string` | `String` | required non-empty |
| `dry_run` | `bool` | `bool` | **`false`** si omit (docs sugerían true para smoke — client debe setear) |

#### RuntimeResponse — `internal/ai/runtime/domain/runtime.go`

| JSON Key | Go Type | Dart Type | Omit |
| --- | --- | --- | --- |
| `mode` | `string` | `RuntimeMode` (`dry_run`\|`ready`) | N |
| `request_id` | `string` | `String` | N (puede vacío?) |
| `context_summary` | `ContextSummary` | nested object / `Map` | N |
| `available_tools` | `[]ToolSummary` | `List<RuntimeToolSummary>` | N |
| `proposed_actions` | `[]ActionProposalOutput` | `List<RuntimeProposedAction>` | N |
| `assistant_message` | `string` | `String` | N |
| `observability` | `RuntimeObservability` | object | N |

#### RuntimeProposedAction — `ActionProposalOutput`

| JSON Key | Dart Type | Omit |
| --- | --- | --- |
| `id` | `String?` | Y (dry_run may omit) |
| `tool_name` | `String` | N |
| `status` | `String` / ProposalStatus | N |
| `proposed_input` | `Map<String,dynamic>` | N |
| `reason` | `String` | N |
| `risk_level` | `RiskLevel` | N |
| `requires_confirmation` | `bool` | N |

#### Nested runtime (MEDIUM)

- `ContextSummary`: typed nested + `insights_summary: Map<String,dynamic>`, `limits: Map<String,int>`
- `ToolSummary.input_schema`: `Map<String,dynamic>?`
- `RuntimeObservability.provider_error_type`: `String?` omitempty

---

## 4. Enum Verification Matrix

| Enum | Backend Values | Documented Values | Match | Source |
| --- | --- | --- | --- | --- |
| ProactivityLevel | `low`, `medium`, `high` | same | YES | `users/domain/ai_settings.go` `IsValidProactivityLevel` |
| AutonomyLevel | `manual`, `suggestive`, `semi_autonomous` | same | YES | same + tools domain |
| DeleteRequestStatus | `pending`, `approved`, `processing`, `completed`, `failed`, `rejected` | same | YES | `privacy/domain/privacy.go` consts (no HTTP validator switch) |
| ActivityType | `task`, `habit`, `routine`, `goal`, `event`, `checklist` | same | YES | `activities/domain` `IsValidType` |
| ActivityStatus | `pending`, `active`, `completed`, `canceled`, `archived` | same | YES | `IsValidStatus` |
| ActivityPriority | `low`, `medium`, `high`, `urgent` | same | YES | `IsValidPriority` |
| ActivitySource | `user`, `ai_suggested`, `ai_created`, `imported` | same | YES | `IsValidSource` |
| CreatedBy | `user`, `sofia` | same | YES | activities + reminders |
| ReminderStatus | `pending`, `due`, `delivered`, `failed`, `canceled`, `archived` | same | YES | `reminders/domain` |
| ReminderSource | `user`, `activity`, `ai_suggested`, `ai_created` | same | YES | — |
| RecurrenceRule | `none`, `daily`, `weekly`, `monthly` | same | YES | `IsValidRecurrence` (+ empty→none) |
| Mood | `very_low`, `low`, `neutral`, `good`, `great` | same | YES | `insights/domain` |
| OutcomeResult | `completed`, `partially_completed`, `skipped`, `failed`, `rescheduled` | same | YES | — |
| MemoryType | 8 values docs | same | YES | `memory/domain` |
| MemorySource | 6 values docs | same | YES | — |
| MemoryVisibility | `private`, `assistant_context`, `archived` | same | YES | — |
| MemoryStatus | `active`, `archived`, `deleted` | same | YES | — |
| Confidence | `low`, `medium`, `high` | same | YES | — |
| Importance | `low`, `medium`, `high`, `critical` | same | YES | — |
| DevicePlatform | `ios`, `android`, `web` | same | YES | `notifications/domain` |
| ProposalStatus | `proposed`, `confirmed`, `rejected`, `executed`, `failed`, `expired` | same | YES | `ai/actions/domain` |
| RiskLevel | `low`, `medium`, `high` | same | YES | — |
| RuntimeMode | `dry_run`, `ready` | docs: `dry_run\|proposal` | **NO** | `ai/runtime/domain` |

### String libre (NO enums)

| Field | Notes |
| --- | --- |
| `communication_style` | default `warm`; no IsValid* |
| `assistant_name` | non-empty only |
| `preferred_language` | non-empty only |
| `sentiment` (reflection) | optional free string |
| `activity.recurrence_rule` | string libre en activity (≠ Reminder RecurrenceRule enum) |
| Tool `category` / `name` | constrained by seed data, not closed enum in validator beyond non-empty |

Backend **no** acepta valores enum adicionales fuera de los `switch` (invalid → 400).

---

## 5. Timestamp Contract

### RFC3339 via `encoding/json` `time.Time` / `*time.Time`

Usar `DateTime` / `DateTime?` en Dart:

- Profile/AISettings `created_at`, `updated_at`
- Activity date fields (`start_at`, `due_at`, `*_at`, `created_at`, `updated_at`)
- Reminder `scheduled_at`, `next_run_at`, `*_at`, `created_at`, `updated_at`
- Insights `occurred_at`, `created_at`, `updated_at`
- Memory timestamps + `last_accessed_at?`
- DeleteRequest timestamps
- ToolDefinition / AiActionProposal timestamps + optional `expires_at`, `confirmed_at`, etc.
- Runtime nested no usa muchos timestamps
- Health: no timestamps
- Export `generated_at`

### Nullable / omitted (`*time.Time` + `omitempty`)

→ `DateTime?`, tolerar ausencia de clave:

Activity: `start_at`, `due_at`, `completed_at`, `canceled_at`, `archived_at`, `reminder_at`  
Reminder: `last_run_at`, `delivered_at`, `failed_at`  
Memory: `last_accessed_at`  
Proposal: `expires_at`, `confirmed_at`, `rejected_at`, `executed_at`

### Request timestamps as **string** RFC3339

Create/update envían strings; vacíos/`null` según pointer semantics.

### Special: DeviceToken HTTP timestamps

Campos `created_at`, `updated_at`, `last_seen_at` son **`string`** ya formateados en handler (`Format(...)`), no `time.Time` encoder.

Recomendación: `DateTime.parse` defensivo **o** mantener `String` si se quiere evitar ambigüedad de offset.

### Firestore

Timestamps se materializan a `time.Time` en repos antes de HTTP DTOs (salvo device token string path). Flutter no ve tipos Firestore.

---

## 6. Collection and Map Contract

### Usar `List<T>`

| JSON | T |
| --- | --- |
| `activities` | `Activity` |
| `reminders` | `Reminder` |
| `memories` | `Memory` |
| `mood_entries` | `MoodEntry` |
| `outcomes` | `ActivityOutcome` |
| `reflections` | `Reflection` |
| `tools` | `ToolDefinition` |
| `action_proposals` | `AiActionProposal` |
| `tags` | `String` |
| `available_tools` | `RuntimeToolSummary` |
| `proposed_actions` | `RuntimeProposedAction` |
| `recent_activities` / `due_reminders` / `relevant_memories` | `ItemSummary` |
| Device tokens GET | `List<DeviceToken>` (raíz) |

Nil vs empty: handlers suelen construir `make([]T, 0, n)` → JSON `[]`, no `null`. Cliente debe aceptar ambos.

### Usar `Map<String, dynamic>`

| Campo | Motivo |
| --- | --- |
| `input_schema` / `output_schema` | `json.RawMessage` JSON Schema |
| `proposed_input` | `json.RawMessage` tool-specific |
| `execution_result` | `json.RawMessage` |
| `context_summary.insights_summary` | `map[string]any` |
| `context_summary` (alternativa) | si no se modelan todos los nested types |
| `observability` parcial | opcional map si se simplifica |

### Usar `Map<String, int>`

| Campo |
| --- |
| `context_summary.limits` |

### No usar `List` genérico `items`

No hay `{"items":[...]}` en API pública.

---

## 7. Fields Flutter Must Never Send

Verificado en handlers/DTOs (ownership vía JWT; create requests no incluyen estos campos):

| Field / class | Evidence | Flutter rule |
| --- | --- | --- |
| `user_id` | No en create DTOs; se toma de JWT claims | Nunca enviar en body |
| `password` / hash | `User.Password` `json:"-"`; solo request auth plaintext | Nunca persistir/loggear; no en models de response |
| `id` en creates | Server uuid | No enviar en create (salvo no aplicable) |
| `created_at` / `updated_at` | Server-set | No enviar en create/update |
| Activity `completed_at` / `canceled_at` / `archived_at` | Solo actions server | No PATCH directo típico |
| Reminder `processing_by`, `processing_until`, `last_attempt_at` | Domain worker fields; **no en HTTP response** | No modelar en cliente HTTP; no enviar |
| Reminder `retry_count`, `next_run_at`, `delivered_at`, … | Response-only | No enviar en create (excepto max_retries permitido) |
| `source` / `created_by` en reminder create HTTP | Handler fuerza `SourceUser`/`CreatedByUser` | No están en create DTO HTTP |
| Device full `token` en logs/UI storage de list | List returns `token_redacted` only | No loggear token completo |
| `execution_result` / proposal status transitions | Solo via confirm/reject/execute endpoints | No inventar PATCH de status |
| Audit logs | No endpoint Flutter público en routes | No modelar para app user |
| Delete executor reports | CLI/worker only | INTERNAL |

---

## 8. Documentation Mismatches

| Severity | Model | Field/Enum | Documented | Actual | Flutter Impact |
| --- | --- | --- | --- | --- | --- |
| BLOCKER | Resource responses | envelope | `{activity\|profile\|...: {}}` | flat object | Wrong codegen wrappers |
| BLOCKER | `UserDataExport` | shape/casing | nested `insights` + snake | flat lists + domain dump | Do not generate |
| HIGH | `PagedResponse` | `items` | generic items | named keys | Wrong list models |
| HIGH | Outcomes list | key | `activity_outcomes` | `outcomes` | Parse fail |
| HIGH | `Reminder` | many fields | full domain | HTTP subset | Phantom fields always null |
| HIGH | DeviceToken list | envelope | `{device_tokens}` | raw array | Parse fail |
| HIGH | DeviceToken times | type | DateTime like others | preformatted string | Type mismatch |
| HIGH | RuntimeMode | values | `proposal` | `ready` | Wrong enum |
| MEDIUM | RuntimeMessageRequest | `dry_run` default | suggest true | omit→**false** | Accidental non-dry-run |
| MEDIUM | RuntimeProposedAction | fields | no id/status | has id?/status | Incomplete model |
| MEDIUM | AuthUser | role/timestamps | optional everywhere | role only on `/me`; timestamps on export user | Inconsistent shape |
| MEDIUM | Delete device | ack | `message`/`204` | `{status:deleted}` | Wrong ack model |
| LOW | communication_style | implied enum | examples only | free string | Over-constrain UI |
| LOW | Activity recurrence_rule | same as Reminder enum | docs blur | free string on activity | Wrong enum reuse |

---

## 9. Canonical Flutter Model Set

Agrupado por feature. **No generar Dart aquí** — lista final recomendada.

### core/

- `ApiError` (+ client `statusCode`)
- `HealthResponse`
- *(no `PagedResponse<items>`)*

### auth/

- `AuthUser`
- `LoginResponse`
- `RegisterRequest` (name, email, password)
- `LoginRequest`

### users/

- `MeResponse`
- `UserProfile`
- `AiSettings`
- `UpdateProfileRequest`
- `UpdateAiSettingsRequest`

### privacy/ (parcial)

- `DeleteRequest`
- `DeleteRequestBody`
- **Diferir** `UserDataExport`

### activities/

- `Activity`
- `CreateActivityRequest`
- `UpdateActivityRequest`
- `ActivityListResponse`

### reminders/

- `Reminder` (**HTTP fields only**)
- `CreateReminderRequest`
- `UpdateReminderRequest`
- `ReminderListResponse`

### insights/

- `MoodEntry`, `CreateMoodEntryRequest`, `MoodEntriesListResponse`
- `ActivityOutcome`, `CreateActivityOutcomeRequest`, `ActivityOutcomesListResponse` (key `outcomes`)
- `Reflection`, `CreateReflectionRequest`, `ReflectionsListResponse`
- `InsightsSummary`

### memory/

- `Memory`
- `CreateMemoryRequest`
- `UpdateMemoryRequest`
- `SearchMemoryRequest`
- `MemoryListResponse`

### notifications/

- `DeviceToken`
- `RegisterDeviceTokenRequest`
- `DeviceTokenDeleteResponse`
- *(list = `List<DeviceToken>` sin wrapper)*

### tools/

- `ToolDefinition`
- `ToolsListResponse`

### ai_actions/

- `AiActionProposal`
- `CreateActionProposalRequest`
- `ActionProposalsListResponse`

### ai_runtime/

- `RuntimeMessageRequest`
- `RuntimeResponse`
- `RuntimeProposedAction`
- `RuntimeObservability`
- `RuntimeContextSummary` (+ `UserSummary`, `ProfileSummary`, `AiSettingsSummary`, `ItemSummary`, `RuntimeToolSummary`, `TokenBudget`) — o colapsar context a `Map<String,dynamic>` en v1

### enums/ (22 + RuntimeMode)

ProactivityLevel, AutonomyLevel, DeleteRequestStatus, ActivityType, ActivityStatus, ActivityPriority, ActivitySource, CreatedBy, ReminderStatus, ReminderSource, RecurrenceRule, Mood, OutcomeResult, MemoryType, MemorySource, MemoryVisibility, MemoryStatus, Confidence, Importance, DevicePlatform, ProposalStatus, RiskLevel, **RuntimeMode**

---

## 10. Final Verdict

### 1. ¿Cuántos modelos Flutter canónicos deben existir?

**~38–45** tipos de datos (incluyendo list wrappers, requests y nested runtime).  
Conjunto mínimo F1 (auth/users/health/activities/reminders sin runtime nested completo): **~25**.

### 2. ¿Cuántos enums?

**22** cerrados alineados docs↔backend + **1** `RuntimeMode` (`dry_run`|`ready`) no documentado correctamente.  
Varios campos son **string libre**, no enums.

### 3. ¿Qué modelos siguen ambiguos?

1. `UserDataExport` (PascalCase / shape)  
2. Generic `PagedResponse.items` (no existe)  
3. `Reminder` “full domain” de docs (campos no expuestos por HTTP)  
4. Opcional: profundidad de tipado de `context_summary` vs `Map<String,dynamic>`

### 4. ¿Qué modelos pueden generarse inmediatamente?

Todos los marcados **Safe To Generate = YES** en §2, usando **DTOs HTTP Go** como fuente (flat resources + named list keys), **después** de alinear el contrato documental de envelopes (o ignorar docs de wrappers).

**No generar aún:** `UserDataExport`, `PagedResponse<items>`, Reminder con processing/recurrence_interval fields.

---

## Appendix — Source file index

| Area | Paths |
| --- | --- |
| Errors | `internal/platform/httpjson/httpjson.go` |
| Health | `internal/server/health.go` |
| Auth HTTP | `internal/auth/interfaces/http/{requests,responses,handlers}.go` |
| Users HTTP | `internal/users/interfaces/http/{requests,responses,handler}.go` |
| Users domain | `internal/users/domain/{profile,ai_settings}.go` |
| Privacy | `internal/privacy/domain/privacy.go`, `application/service.go` |
| Activities | `internal/activities/interfaces/http/*`, `domain/activity.go` |
| Reminders | `internal/reminders/interfaces/http/*`, `domain/reminder.go` |
| Insights | `internal/insights/interfaces/http/*`, `domain/insights.go` |
| Memory | `internal/memory/interfaces/http/*`, `domain/memory.go` |
| Notifications | `internal/notifications/interfaces/http/handler.go`, `domain/device_token.go` |
| Tools | `internal/tools/domain/tool.go`, `interfaces/http/handler.go` |
| AI Actions | `internal/ai/actions/domain/proposal.go`, `interfaces/http/handler.go` |
| AI Runtime | `internal/ai/runtime/domain/runtime.go`, `interfaces/http/handler.go` |

*Fin inventario F0 schema.*
