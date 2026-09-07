# Flutter Integration Roadmap

Fecha: 2026-07-05

Este roadmap ordena la integracion Flutter para Sofia Backend v0.1 sin activar dependencias peligrosas ni features incompletas.

## Fase 1: Identidad, Profile y AI Settings

Objetivo: permitir login funcional, bootstrap de sesion y configuracion inicial de Sofia.

Endpoints:

- `POST /auth/register`
- `POST /auth/login`
- `GET /users/me`
- `PATCH /users/me/profile`
- `GET /users/me/ai-settings`
- `PATCH /users/me/ai-settings`
- `POST /users/me/onboarding/complete`
- `GET /health`

Flutter:

- `AuthRepository`, `UserRepository`, `SystemRepository`
- `AuthCubit`, `SessionCubit`, `UserProfileCubit`, `AiSettingsCubit`, `OnboardingCubit`
- Pantallas: login, register, onboarding, profile, assistant settings, diagnostics.

Criterio de salida:

- Token guardado de forma segura.
- `GET /users/me` hidrata estado global.
- Settings respetan `memory_enabled`, `reminders_enabled`, `planning_enabled`.
- Expired token produce logout/relogin.

## Fase 2: Activities y Reminders

Objetivo: construir el loop productivo central.

Endpoints:

- `POST /activities`
- `GET /activities`
- `GET /activities/{id}`
- `PATCH /activities/{id}`
- `POST /activities/{id}/complete`
- `POST /activities/{id}/cancel`
- `POST /activities/{id}/archive`
- `POST /reminders`
- `GET /reminders`
- `GET /reminders/due`
- `GET /reminders/{id}`
- `PATCH /reminders/{id}`
- `POST /reminders/{id}/cancel`
- `POST /reminders/{id}/archive`

Flutter:

- `ActivitiesRepository`, `RemindersRepository`
- `ActivitiesCubit`, `ActivityDetailCubit`, `ActivityEditorCubit`, `RemindersCubit`, `ReminderDetailCubit`, `ReminderEditorCubit`
- Pantallas: activity list/detail/editor, reminder list/detail/editor, due reminders panel.

Criterio de salida:

- No se envia `user_id`.
- Filtros basicos funcionan.
- Recurrence simple se representa como `none`, `daily`, `weekly`, `monthly`.
- El UI no promete push real todavia.

## Fase 3: Insights y Memory

Objetivo: habilitar tracking personal, reflexiones y memoria textual explicita.

Endpoints:

- `POST /insights/mood`
- `GET /insights/mood`
- `POST /insights/outcomes`
- `GET /insights/outcomes`
- `POST /insights/reflections`
- `GET /insights/reflections`
- `GET /insights/summary`
- `POST /memory`
- `GET /memory`
- `GET /memory/{id}`
- `PATCH /memory/{id}`
- `POST /memory/{id}/archive`
- `DELETE /memory/{id}`
- `POST /memory/search`

Flutter:

- `InsightsRepository`, `MemoryRepository`
- `MoodCubit`, `OutcomeCubit`, `ReflectionCubit`, `InsightsSummaryCubit`, `MemoryCubit`, `MemorySearchCubit`
- Pantallas: mood check-in, insights dashboard, reflections, memory list/detail/editor/search.

Criterio de salida:

- `memory_enabled=false` deshabilita UI de creacion/busqueda.
- Deleted memories no aparecen en listas ni busquedas.
- Copys evitan lenguaje clinico/diagnostico.

## Fase 4: AI Runtime y Action Proposals

Objetivo: probar interaccion assistant proposal-only sin ejecucion automatica.

Endpoints:

- `POST /ai/runtime/message`
- `POST /ai/action-proposals`
- `GET /ai/action-proposals`
- `GET /ai/action-proposals/{id}`
- `POST /ai/action-proposals/{id}/confirm`
- `POST /ai/action-proposals/{id}/reject`
- `POST /ai/action-proposals/{id}/execute`
- `GET /tools`
- `GET /tools/{name}`

Flutter:

- `AiRuntimeRepository`, `AiActionsRepository`, `ToolsRepository`
- `AssistantRuntimeCubit`, `ActionProposalsCubit`, `ActionProposalDetailCubit`, `ToolsCubit`
- Pantallas: assistant chat, approvals inbox, proposal detail, developer tools.

Criterio de salida:

- `dry_run=true` para smoke inicial.
- Ninguna propuesta se ejecuta sin accion explicita.
- `execute` queda feature-flagged.
- Gemini real no es dependencia obligatoria.

## Fase 5: Notifications y Device Tokens

Objetivo: preparar registro de dispositivos sin activar push visible por default.

Endpoints:

- `POST /notifications/device-tokens`
- `GET /notifications/device-tokens`
- `DELETE /notifications/device-tokens/{id}`

Flutter:

- `NotificationsRepository`
- `NotificationsCubit`
- Pantallas: notification settings, debug device tokens.

Criterio de salida:

- Tokens nunca se muestran completos.
- Registro queda condicionado a permisos del OS.
- FCM real sigue detras de config backend `FCM_ENABLED` y `FCM_DRY_RUN`.

## No Conectar Todavia

- Gemini real como requisito para abrir la app.
- FCM real visible para usuarios finales.
- Delete executor o borrado fisico automatico.
- Migraciones legacy o comandos internos.
- Hardware/smart home.
- RAG/embeddings o busqueda semantica.

## Orden Recomendado de Releases Flutter

1. Alpha local con memory driver: Fase 1 + Fase 2.
2. Alpha Firestore personal: Fase 1 + Fase 2 + export.
3. Beta privada: Fase 3.
4. Beta assistant: Fase 4 en `dry_run` y approvals.
5. Beta notifications: Fase 5 con dry-run/diagnostico antes de push real.
