# Flutter Screens Mapping

Fecha: 2026-07-05

Mapa sugerido de pantallas, repositorios y Cubit/BLoC para Sofia Flutter v0.1.

| Pantalla | Trabajo de usuario | Endpoints | Repository | Cubit/BLoC | Fase | Estado |
| --- | --- | --- | --- | --- | --- | --- |
| `SplashSessionScreen` | Verificar sesion local y health opcional. | `GET /health`, `GET /users/me` | `SystemRepository`, `UserRepository` | `SessionCubit` | 1 | listo |
| `LoginScreen` | Iniciar sesion. | `POST /auth/login` | `AuthRepository` | `AuthCubit` | 1 | listo |
| `RegisterScreen` | Crear cuenta. | `POST /auth/register` | `AuthRepository` | `AuthCubit` | 1 | listo |
| `OnboardingFlow` | Capturar preferencias iniciales. | `PATCH /users/me/profile`, `PATCH /users/me/ai-settings`, `POST /users/me/onboarding/complete` | `UserRepository` | `OnboardingCubit` | 1 | listo |
| `ProfileScreen` | Ver/editar perfil. | `GET /users/me`, `PATCH /users/me/profile` | `UserRepository` | `UserProfileCubit` | 1 | listo |
| `AssistantSettingsScreen` | Configurar Sofia. | `GET /users/me/ai-settings`, `PATCH /users/me/ai-settings` | `UserRepository` | `AiSettingsCubit` | 1 | listo |
| `DiagnosticsScreen` | Ver estado basico backend. | `GET /health` | `SystemRepository` | `HealthCubit` | 1 | listo |
| `ActivitiesListScreen` | Ver y filtrar activities. | `GET /activities` | `ActivitiesRepository` | `ActivitiesCubit` | 2 | listo |
| `ActivityDetailScreen` | Ver, completar, cancelar o archivar. | `GET /activities/{id}`, action endpoints | `ActivitiesRepository` | `ActivityDetailCubit` | 2 | listo |
| `ActivityEditorScreen` | Crear/editar activity. | `POST /activities`, `PATCH /activities/{id}` | `ActivitiesRepository` | `ActivityEditorCubit` | 2 | listo |
| `RemindersScreen` | Ver reminders. | `GET /reminders` | `RemindersRepository` | `RemindersCubit` | 2 | listo |
| `ReminderDetailScreen` | Ver/cancelar/archivar reminder. | `GET /reminders/{id}`, action endpoints | `RemindersRepository` | `ReminderDetailCubit` | 2 | listo |
| `ReminderEditorScreen` | Crear/editar reminder recurrente simple. | `POST /reminders`, `PATCH /reminders/{id}` | `RemindersRepository` | `ReminderEditorCubit` | 2 | listo |
| `DueRemindersPanel` | Ver reminders vencidos. | `GET /reminders/due` | `RemindersRepository` | `DueRemindersCubit` | 2 | parcial |
| `MoodCheckInScreen` | Registrar estado del usuario. | `POST /insights/mood` | `InsightsRepository` | `MoodCubit` | 3 | listo |
| `MoodHistoryScreen` | Ver historial de mood. | `GET /insights/mood` | `InsightsRepository` | `MoodCubit` | 3 | listo |
| `ActivityOutcomeScreen` | Registrar resultado tras completar activity. | `POST /insights/outcomes` | `InsightsRepository` | `OutcomeCubit` | 3 | listo |
| `OutcomesHistoryScreen` | Ver resultados historicos. | `GET /insights/outcomes` | `InsightsRepository` | `OutcomeCubit` | 3 | listo |
| `ReflectionEditorScreen` | Crear reflexion. | `POST /insights/reflections` | `InsightsRepository` | `ReflectionCubit` | 3 | listo |
| `ReflectionsScreen` | Listar reflexiones. | `GET /insights/reflections` | `InsightsRepository` | `ReflectionCubit` | 3 | listo |
| `InsightsDashboardScreen` | Ver resumen no clinico. | `GET /insights/summary` | `InsightsRepository` | `InsightsSummaryCubit` | 3 | listo |
| `MemoryListScreen` | Ver memorias explicitas. | `GET /memory` | `MemoryRepository` | `MemoryCubit` | 3 | listo |
| `MemoryDetailScreen` | Ver/archivar/eliminar memoria. | `GET /memory/{id}`, `POST /memory/{id}/archive`, `DELETE /memory/{id}` | `MemoryRepository` | `MemoryDetailCubit` | 3 | listo |
| `MemoryEditorScreen` | Crear/editar memoria. | `POST /memory`, `PATCH /memory/{id}` | `MemoryRepository` | `MemoryEditorCubit` | 3 | listo |
| `MemorySearchScreen` | Busqueda textual. | `POST /memory/search` | `MemoryRepository` | `MemorySearchCubit` | 3 | listo |
| `AssistantChatScreen` | Mensaje a runtime proposal-only. | `POST /ai/runtime/message` | `AiRuntimeRepository` | `AssistantRuntimeCubit` | 4 | parcial |
| `AssistantApprovalsScreen` | Ver proposals pendientes. | `GET /ai/action-proposals` | `AiActionsRepository` | `ActionProposalsCubit` | 4 | parcial |
| `ActionProposalDetailScreen` | Confirmar/rechazar/ejecutar explicitamente. | `GET /ai/action-proposals/{id}`, confirm/reject/execute | `AiActionsRepository` | `ActionProposalDetailCubit` | 4 | parcial |
| `DeveloperToolsScreen` | Inspeccionar tool registry. | `GET /tools`, `GET /tools/{name}` | `ToolsRepository` | `ToolsCubit` | 4 | parcial |
| `NotificationSettingsScreen` | Registrar y administrar device tokens. | device-token endpoints | `NotificationsRepository` | `NotificationsCubit` | 5 | parcial |
| `PrivacyExportScreen` | Descargar/ver export personal. | `GET /users/me/export` | `PrivacyRepository` | `PrivacyCubit` | 1/3 | listo |
| `DeleteAccountRequestScreen` | Solicitar borrado controlado. | `POST /users/me/delete-request` | `PrivacyRepository` | `PrivacyCubit` | 5+ | parcial |

## Pantallas Que Deben Permanecer Ocultas o Feature-Flagged

| Pantalla/Feature | Motivo |
| --- | --- |
| Gemini provider settings obligatorios | Gemini es opcional; fake provider debe seguir funcionando. |
| Push notification live test | FCM real no debe enviar visible por default. |
| Delete executor/admin apply | No hay auth admin UI segura para v0.1. |
| Migration tools | Son comandos internos, no API publica. |
| Hardware/smart home | Fuera del backend runtime actual. |
| RAG/semantic memory | Embeddings no implementados. |

## Navegacion Recomendada

- `UnauthenticatedStack`: login/register.
- `OnboardingStack`: profile basics + AI settings + complete onboarding.
- `MainShell`: activities, reminders, insights, memory, assistant, settings.
- `SettingsStack`: profile, assistant settings, notifications, privacy export, diagnostics.

## Estado Global Recomendado

- `SessionState`: token, auth user, profile, AI settings.
- `FeatureFlagsState`: derivado de AI settings y config local.
- `SyncState`: por modulo, con loading/error/lastUpdated.
- `PrivacyState`: export/delete request separado para evitar mezclar datos sensibles con dashboard.
