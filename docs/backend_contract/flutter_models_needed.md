# Flutter Models Needed

Fecha: 2026-07-05

Todos los modelos deben serializar/deserializar snake_case y tratar timestamps como `DateTime?` desde RFC3339.

## Core API

| Modelo | Campos principales | Notas |
| --- | --- | --- |
| `ApiError` | `statusCode`, `message`, `rawBody?` | Normalizar errores `{message}`. |
| `PagedResponse<T>` | `items`, `nextCursor?` | Adaptar por endpoint: `activities`, `reminders`, `memories`. |
| `HealthResponse` | `status`, `environment`, `firestore` | Para diagnostics. |

## Auth y Users

| Modelo | Campos principales | Notas |
| --- | --- | --- |
| `AuthUser` | `id`, `name`, `email`, `role?`, `createdAt?`, `updatedAt?` | Nunca `password_hash`. |
| `LoginResponse` | `token`, `user` | Guardar token en storage seguro. |
| `UserProfile` | `id`, `userId`, `displayName`, `preferredName`, `timezone`, `locale`, `avatarUrl?`, `onboardingCompleted`, `createdAt`, `updatedAt` | Profile rico. |
| `AiSettings` | `userId`, `assistantName`, `communicationStyle`, `proactivityLevel`, `autonomyLevel`, `memoryEnabled`, `remindersEnabled`, `planningEnabled`, `preferredLanguage`, `createdAt`, `updatedAt` | Gate de features. |
| `MeResponse` | `user`, `profile`, `aiSettings` | Bootstrap principal. |
| `UserDataExport` | `user`, `profile`, `aiSettings`, `activities`, `reminders`, `insights`, `memories`, `actionProposals` | Datos sensibles; no cache largo. |
| `DeleteRequest` | `id`, `userId`, `status`, `reason?`, `createdAt`, `updatedAt` | No implica borrado inmediato. |

Enums:

- `ProactivityLevel`: `low`, `medium`, `high`
- `AutonomyLevel`: `manual`, `suggestive`, `semi_autonomous`
- `DeleteRequestStatus`: `pending`, `approved`, `processing`, `completed`, `failed`, `rejected`

## Activities

| Modelo | Campos principales |
| --- | --- |
| `Activity` | `id`, `userId`, `type`, `title`, `description`, `status`, `priority`, `startAt?`, `dueAt?`, `completedAt?`, `canceledAt?`, `archivedAt?`, `timezone`, `tags`, `source`, `createdBy`, `createdAt`, `updatedAt`, `hasReminder`, `reminderAt?`, `recurrenceRule`, `aiManageable`, `aiNotes?` |
| `CreateActivityRequest` | `type`, `title`, `description`, `priority`, `startAt?`, `dueAt?`, `timezone`, `tags`, `hasReminder`, `reminderAt?`, `recurrenceRule`, `aiManageable`, `aiNotes?` |
| `UpdateActivityRequest` | nullable subset de create + `status` |
| `ActivityListResponse` | `activities`, `nextCursor?` |

Enums:

- `ActivityType`: `task`, `habit`, `routine`, `goal`, `event`, `checklist`
- `ActivityStatus`: `pending`, `active`, `completed`, `canceled`, `archived`
- `ActivityPriority`: `low`, `medium`, `high`, `urgent`
- `ActivitySource`: `user`, `ai_suggested`, `ai_created`, `imported`
- `CreatedBy`: `user`, `sofia`

## Reminders

| Modelo | Campos principales |
| --- | --- |
| `Reminder` | `id`, `userId`, `activityId?`, `title`, `description`, `status`, `scheduledAt`, `timezone`, `recurrenceRule`, `recurrenceInterval`, `recurrenceUntil?`, `recurrenceCount`, `recurrenceOccurrences`, `nextRunAt`, `lastRunAt?`, `deliveredAt?`, `failedAt?`, `failureReason?`, `retryCount`, `maxRetries`, `source`, `createdBy`, `processingBy?`, `processingUntil?`, `lastAttemptAt?`, `createdAt`, `updatedAt` |
| `CreateReminderRequest` | `activityId?`, `title`, `description`, `scheduledAt`, `timezone`, `recurrenceRule`, `recurrenceInterval`, `recurrenceUntil?`, `recurrenceCount`, `maxRetries` |
| `UpdateReminderRequest` | nullable subset de create sin `activityId` |
| `ReminderListResponse` | `reminders`, `nextCursor?` |

Enums:

- `ReminderStatus`: `pending`, `due`, `delivered`, `failed`, `canceled`, `archived`
- `ReminderSource`: `user`, `activity`, `ai_suggested`, `ai_created`
- `RecurrenceRule`: `none`, `daily`, `weekly`, `monthly`

## Insights

| Modelo | Campos principales |
| --- | --- |
| `MoodEntry` | `id`, `userId`, `activityId?`, `mood`, `energyLevel`, `stressLevel`, `focusLevel`, `note`, `occurredAt`, `timezone`, `createdAt`, `updatedAt` |
| `ActivityOutcome` | `id`, `userId`, `activityId`, `result`, `completionQuality`, `satisfactionLevel`, `difficultyLevel`, `note`, `createdAt`, `updatedAt` |
| `Reflection` | `id`, `userId`, `activityId?`, `title`, `content`, `sentiment?`, `tags`, `occurredAt`, `createdAt`, `updatedAt` |
| `InsightsSummary` | `averageEnergy`, `averageStress`, `averageFocus`, `totalOutcomes`, `totalReflections` |

Enums:

- `Mood`: `very_low`, `low`, `neutral`, `good`, `great`
- `OutcomeResult`: `completed`, `partially_completed`, `skipped`, `failed`, `rescheduled`

Numeric ranges:

- `energyLevel`, `stressLevel`, `focusLevel`, `completionQuality`, `satisfactionLevel`, `difficultyLevel`: `1..5`

## Memory

| Modelo | Campos principales |
| --- | --- |
| `Memory` | `id`, `userId`, `type`, `title`, `content`, `summary`, `source`, `sourceId?`, `confidence`, `importance`, `tags`, `visibility`, `status`, `createdAt`, `updatedAt`, `lastAccessedAt?` |
| `CreateMemoryRequest` | `type`, `title`, `content`, `summary`, `source`, `sourceId?`, `confidence`, `importance`, `tags`, `visibility` |
| `UpdateMemoryRequest` | nullable subset de create + `status` |
| `SearchMemoryRequest` | `query`, `type?`, `tag?`, `importance?`, `limit` |
| `MemoryListResponse` | `memories`, `nextCursor?` |

Enums:

- `MemoryType`: `preference`, `fact`, `episodic`, `semantic`, `instruction`, `relationship`, `project`, `system_note`
- `MemorySource`: `user_manual`, `conversation`, `activity`, `reflection`, `imported`, `ai_inferred`
- `MemoryVisibility`: `private`, `assistant_context`, `archived`
- `MemoryStatus`: `active`, `archived`, `deleted`
- `Confidence`: `low`, `medium`, `high`
- `Importance`: `low`, `medium`, `high`, `critical`

## Notifications

| Modelo | Campos principales | Notas |
| --- | --- | --- |
| `DeviceToken` | `id`, `userId`, `platform`, `tokenRedacted`, `enabled`, `createdAt`, `updatedAt`, `lastSeenAt` | Full token solo se envia en POST; no se guarda en UI logs. |
| `RegisterDeviceTokenRequest` | `platform`, `token` | Token viene de Firebase Messaging en Flutter. |

Enums:

- `DevicePlatform`: `ios`, `android`, `web`

## Tools, AI Actions y Runtime

| Modelo | Campos principales | Notas |
| --- | --- | --- |
| `ToolDefinition` | `id`, `name`, `description`, `category`, `inputSchema`, `outputSchema?`, `requiresConfirmation`, `enabled`, `minAutonomyLevel`, `createdAt`, `updatedAt` | JSON schema como `Map<String, dynamic>`. |
| `AiActionProposal` | `id`, `userId`, `toolName`, `status`, `proposedInput`, `reason`, `riskLevel`, `requiresConfirmation`, `expiresAt?`, `createdAt`, `updatedAt`, `confirmedAt?`, `rejectedAt?`, `executedAt?`, `executionResult?`, `executionError?` | Mostrar riesgo antes de confirm/execute. |
| `RuntimeMessageRequest` | `message`, `dryRun` | Default `dryRun=true` para primeras pruebas. |
| `RuntimeResponse` | `mode`, `requestId?`, `contextSummary`, `availableTools`, `proposedActions`, `assistantMessage`, `observability?` | No asumir Gemini. |
| `RuntimeProposedAction` | `toolName`, `proposedInput`, `reason`, `riskLevel`, `requiresConfirmation` | Nunca ejecutar automaticamente. |

Enums:

- `ProposalStatus`: `proposed`, `confirmed`, `rejected`, `executed`, `failed`, `expired`
- `RiskLevel`: `low`, `medium`, `high`

## Serializacion Recomendable

- Usar `json_serializable` o equivalente con `fieldRename: FieldRename.snake`.
- Mantener `Map<String, dynamic>` para schemas, proposed input y execution result.
- Los requests de PATCH deben omitir campos `null` cuando el usuario no los cambio.
- Tratar `404` en recursos propios como "no encontrado o sin permisos", sin revelar ownership.
