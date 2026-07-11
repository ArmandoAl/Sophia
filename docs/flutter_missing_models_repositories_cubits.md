# Flutter Missing Models, Repositories & Cubits

Fecha: 2026-07-10  
Contrato de referencia: `docs/backend_contract/flutter_models_needed.md`, `flutter_screens_mapping.md`.

Leyenda de estado:

- **Falta** — no existe en el proyecto
- **Parcial** — hay UI/cubit local no alineado al API
- **Existe (UI)** — pantalla/widget reutilizable sin data layer
- **Fuera de v0.1** — no implementar contra backend ahora

---

## 1. Infraestructura compartida (faltante)

| Pieza | Estado | Notas |
| --- | --- | --- |
| `ApiClient` | Falta | Bearer, JSON, base_url |
| `ApiException` / `ApiError` | Falta | Mapear `{ message }` + statusCode |
| `PagedResponse<T>` | Falta | `next_cursor` |
| Token secure storage | Falta | |
| Feature flags store | Falta | aiRuntime, execute, notifications, deleteRequest, smartHome |
| `json_serializable` setup | Falta | `fieldRename: snake` |

---

## 2. Modelos

### Core API

| Modelo | Estado | Fase |
| --- | --- | --- |
| `ApiError` | Falta | 1 |
| `PagedResponse<T>` | Falta | 1–2 |
| `HealthResponse` | Falta | 1 |
| `SystemHealth` (cpu/ram) | Parcial / incorrecto | Reemplazar; no es el contrato |

### Auth y Users

| Modelo | Estado | Fase |
| --- | --- | --- |
| `AuthUser` | Falta | 1 |
| `LoginResponse` | Falta | 1 |
| `UserProfile` | Falta | 1 |
| `AiSettings` | Falta | 1 |
| `MeResponse` | Falta | 1 |
| `UserDataExport` | Falta | 1/3 |
| `DeleteRequest` | Falta | 5+ |
| Enums `ProactivityLevel`, `AutonomyLevel`, `DeleteRequestStatus` | Falta | 1 / 5 |

### Activities

| Modelo | Estado | Fase |
| --- | --- | --- |
| `Activity` | Falta | 2 |
| `CreateActivityRequest` / `UpdateActivityRequest` | Falta | 2 |
| `ActivityListResponse` | Falta | 2 |
| Enums type/status/priority/source/`CreatedBy` | Falta | 2 |
| `TaskItem` (local) | Parcial | No reutilizar como Activity |

### Reminders

| Modelo | Estado | Fase |
| --- | --- | --- |
| `Reminder` | Falta | 2 |
| `CreateReminderRequest` / `UpdateReminderRequest` | Falta | 2 |
| `ReminderListResponse` | Falta | 2 |
| Enums status/source/`RecurrenceRule` | Falta | 2 |

### Insights

| Modelo | Estado | Fase |
| --- | --- | --- |
| `MoodEntry` | Falta | 3 |
| `ActivityOutcome` | Falta | 3 |
| `Reflection` | Falta | 3 |
| `InsightsSummary` | Falta | 3 |
| Enums `Mood`, `OutcomeResult` | Falta | 3 |
| `DecisionItem` (local) | Parcial | Mejor mapear a proposals, no a mood |

### Memory

| Modelo | Estado | Fase |
| --- | --- | --- |
| `Memory` + create/update/search/list | Falta | 3 |
| Enums type/source/visibility/status/confidence/importance | Falta | 3 |

### Notifications

| Modelo | Estado | Fase |
| --- | --- | --- |
| `DeviceToken` | Falta | 5 |
| `RegisterDeviceTokenRequest` | Falta | 5 |
| `DevicePlatform` | Falta | 5 |

### Tools / AI Actions / Runtime

| Modelo | Estado | Fase |
| --- | --- | --- |
| `ToolDefinition` | Falta | 4 |
| `AiActionProposal` | Falta | 4 |
| `RuntimeMessageRequest` / `RuntimeResponse` / `RuntimeProposedAction` | Falta | 4 |
| Enums `ProposalStatus`, `RiskLevel` | Falta | 4 |
| `ChatMessage` (local) | Parcial | Mantener para UI; hidratar desde runtime |
| `ProposedAction` (widget) | Parcial | Adaptar a proposal API |

### Fuera de v0.1

| Modelo | Estado |
| --- | --- |
| `SmartDevice` | Fuera de v0.1 — no cablear |

---

## 3. Repositories

| Repository (contrato) | Estado | Fase | Endpoints clave |
| --- | --- | --- | --- |
| `AuthRepository` | Falta | 1 | register, login |
| `UserRepository` | Falta | 1 | me, profile, ai-settings, onboarding |
| `SystemRepository` | Falta | 1 | health |
| `PrivacyRepository` | Falta | 1/5 | export, delete-request |
| `ActivitiesRepository` | Falta | 2 | CRUD + complete/cancel/archive |
| `RemindersRepository` | Falta | 2 | CRUD + due/cancel/archive |
| `InsightsRepository` | Falta | 3 | mood, outcomes, reflections, summary |
| `MemoryRepository` | Falta | 3 | CRUD + search + archive/delete |
| `AiRuntimeRepository` | Falta | 4 | runtime/message |
| `AiActionsRepository` | Falta | 4 | proposals CRUD actions |
| `ToolsRepository` | Falta | 4 | tools list/get |
| `NotificationsRepository` | Falta | 5 | device-tokens |

**Hoy:** cero implementaciones; solo comentario en `service_locator.dart`.

---

## 4. Cubits / BLoCs

### Faltantes (contrato)

| Cubit | Fase | Pantalla destino |
| --- | --- | --- |
| `SessionCubit` | 1 | Splash / global |
| `AuthCubit` | 1 | Login / Register |
| `UserProfileCubit` | 1 | Profile |
| `AiSettingsCubit` | 1 | Assistant settings |
| `OnboardingCubit` | 1 | Onboarding |
| `HealthCubit` | 1 | Diagnostics |
| `PrivacyCubit` | 1/5 | Export / delete request |
| `ActivitiesCubit` | 2 | Activities list |
| `ActivityDetailCubit` | 2 | Activity detail |
| `ActivityEditorCubit` | 2 | Activity editor |
| `RemindersCubit` (API) | 2 | Reminders list |
| `ReminderDetailCubit` | 2 | Reminder detail |
| `ReminderEditorCubit` | 2 | Reminder editor |
| `DueRemindersCubit` | 2 | Due panel (parcial) |
| `MoodCubit` | 3 | Mood check-in / history |
| `OutcomeCubit` | 3 | Outcomes |
| `ReflectionCubit` | 3 | Reflections |
| `InsightsSummaryCubit` | 3 | Insights dashboard |
| `MemoryCubit` / `MemoryDetailCubit` / `MemoryEditorCubit` / `MemorySearchCubit` | 3 | Memory flows |
| `AssistantRuntimeCubit` | 4 | Assistant chat |
| `ActionProposalsCubit` | 4 | Approvals inbox |
| `ActionProposalDetailCubit` | 4 | Proposal detail |
| `ToolsCubit` | 4 | Developer tools |
| `NotificationsCubit` | 5 | Notification settings |

### Existentes a reemplazar o adaptar

| Cubit actual | Acción |
| --- | --- |
| `ChatMessageCubit` | Adaptar → `AssistantRuntimeCubit` (Fase 4); hasta entonces mock + flag |
| `DashboardCubit` | Reemplazar → `HealthCubit` |
| `SettingsCubit` | Reducir a preferencias locales UX; datos de cuenta → cubits de user/AI |
| `RemindersCubit` + `TaskItem` | Reemplazar → Activities/Reminders de contrato |
| `InsightCubit` + `DecisionItem` | Dividir → InsightsSummary vs ActionProposals |
| `SmartHomeCubit` | Aislar; Fuera de v0.1 |

### Estado global recomendado (aún faltante)

| Estado | Contenido |
| --- | --- |
| `SessionState` | token, auth user, profile, ai settings |
| `FeatureFlagsState` | derivado de AI settings + config local |
| `SyncState` | loading/error/lastUpdated por módulo |
| `PrivacyState` | export/delete separado del dashboard |

---

## 5. Pantallas: reutilizables vs faltantes

### Reutilizables (Existe UI)

| Actual | Destino contrato | Esfuerzo |
| --- | --- | --- |
| `ChatPage` + `ActionProposalCard` | `AssistantChatScreen` | Medio (data + map proposals) |
| `SettingsPage` | Hub `SettingsStack` | Bajo–medio |
| `DashboardPage` | `DiagnosticsScreen` | Medio (campos distintos) |
| `RemindersPage` | Activities o Reminders list | Medio–alto (modelo distinto) |
| `WeeklyInsightPage` | Approvals / insights shell | Medio |
| `MainWrapper` | MainShell | Medio (tabs) |

### Faltantes (crear)

| Pantalla contrato | Fase |
| --- | --- |
| `SplashSessionScreen` | 1 |
| `LoginScreen` | 1 |
| `RegisterScreen` | 1 |
| `OnboardingFlow` | 1 |
| `ProfileScreen` / `EditProfileScreen` | 1 |
| `AssistantSettingsScreen` | 1 |
| `ActivityDetailScreen` / `ActivityEditorScreen` | 2 |
| `ReminderDetailScreen` / `ReminderEditorScreen` | 2 |
| `DueRemindersPanel` | 2 |
| `MoodCheckInScreen` / `MoodHistoryScreen` | 3 |
| `ActivityOutcomeScreen` / `OutcomesHistoryScreen` | 3 |
| `ReflectionEditorScreen` / `ReflectionsScreen` | 3 |
| `InsightsDashboardScreen` | 3 |
| `MemoryList/Detail/Editor/Search` | 3 |
| `AssistantApprovalsScreen` / `ActionProposalDetailScreen` | 4 |
| `DeveloperToolsScreen` | 4 |
| `NotificationSettingsScreen` | 5 |
| `PrivacyExportScreen` | 1/3 |
| `DeleteAccountRequestScreen` | 5+ |

### Ocultar / no conectar

| Feature | Motivo |
| --- | --- |
| `SmartHomePage` | Fuera de Backend v0.1 |
| Gemini settings obligatorios | Provider fake default |
| Push live test | FCM off por default |
| Delete executor admin | No UI segura v0.1 |

---

## 6. Matriz rápida Fase 1 (primera implementación)

Todo lo siguiente está **Falta** y es el set mínimo a crear:

```
core/network/
  api_client.dart
  api_exception.dart
core/storage/
  token_storage.dart
core/config/
  app_config.dart
  feature_flags.dart

features/auth/
  data/models/     AuthUser, LoginResponse, ...
  data/repositories/ auth_repository.dart
  presentation/    login, register, auth_cubit

features/session/  (o users/)
  MeResponse, UserProfile, AiSettings
  user_repository.dart
  session_cubit, user_profile_cubit, ai_settings_cubit, onboarding_cubit

features/system/
  HealthResponse
  system_repository.dart
  health_cubit
  diagnostics (adaptar dashboard)
```

---

## 7. Resumen numérico

| Categoría | Requeridos por contrato v0.1 | Existen alineados | Faltan / a reemplazar |
| --- | --- | --- | --- |
| Modelos API | ~35+ (incl. requests/enums) | 0 | Todos |
| Repositories | 11 | 0 | 11 |
| Cubits de contrato | ~25 | 0 | ~25 |
| Pantallas de contrato | ~30 | ~5 shells UI | ~25 nuevas o adaptaciones fuertes |

**Conclusión:** el gap es casi total en data/domain. La UI existente cubre ~15–20% del shell visual de Fases 1/2/4; no reduce el trabajo de modelos/repos/cubits de Fase 1.
