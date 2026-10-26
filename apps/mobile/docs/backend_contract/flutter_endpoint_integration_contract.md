# Flutter Endpoint Integration Contract

Fecha: 2026-07-05

Este documento es el contrato funcional de integracion entre Sofia Backend v0.1 y Flutter. Refleja el estado actual del backend: endpoints existentes, seguridad esperada, cuerpos JSON y recomendaciones de arquitectura Flutter.

## Convenciones Globales

- `base_url`: variable de ambiente/app config de Flutter.
- Auth: `Authorization: Bearer <token>` para todos los endpoints marcados con auth.
- JSON: enviar `Content-Type: application/json`; el backend responde JSON.
- Fechas: strings RFC3339, por ejemplo `2026-07-05T15:00:00Z`.
- Errores: formato general `{ "message": "..." }`.
- Estados comunes: `400` validacion/body invalido, `401` sin token/token invalido/credenciales invalidas, `403` feature deshabilitada o policy, `404` recurso inexistente o de otro usuario, `409` conflicto/regla de negocio, `429` rate limit, `500` error interno.
- Ownership: Flutter nunca debe enviar `user_id` para operar datos del usuario. El backend lo toma del JWT.
- Paginacion: listas principales pueden aceptar `limit` y `cursor`, y responder `next_cursor`.
- Estados de integracion: `listo` significa usable para v0.1; `parcial` significa usable con feature flag o UX defensiva; `no usar todavia` significa existe para tooling/admin o futuro.

## No Conectar Todavia

- Gemini real como dependencia obligatoria: usar AI Runtime fake por default; Gemini es opcional y diagnostico.
- FCM real visible para usuarios: device tokens y dry-run existen, pero push real debe seguir desactivado por default.
- Delete executor: solo existe delete request/control operativo; no ofrecer borrado fisico inmediato en UI normal.
- Migraciones legacy: no exponer comandos ni colecciones staging al cliente.
- Hardware/smart home: fuera de v0.1.
- RAG/embeddings: memory textual esta lista; busqueda semantica no.

## Auth

| Endpoint | Auth | Proposito | Request | Response esperada | Errores | Flutter sugerido | Estado | Seguridad |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `POST /auth/register` | No | Crear cuenta y defaults de profile/settings. | `{ "name": "Armando", "email": "a@example.test", "password": "Password123" }` | `201 { "id": "...", "name": "...", "email": "..." }` | `400` campos invalidos/password debil, `409` email duplicado, `429` rate limit | `RegisterScreen`, `AuthRepository.register`, `AuthCubit` | listo | No guardar password; email se normaliza; password minimo 8. |
| `POST /auth/login` | No | Autenticar y obtener JWT. | `{ "email": "a@example.test", "password": "Password123" }` | `200 { "token": "...", "user": { "id": "...", "name": "...", "email": "..." } }` | `400`, `401 { "message": "invalid credentials" }`, `429` | `LoginScreen`, `AuthRepository.login`, `AuthCubit/SessionCubit` | listo | No loggear token; manejar expiracion JWT con logout/relogin. |

## Users, Profile, Settings, Privacy

| Endpoint | Auth | Proposito | Request | Response esperada | Errores | Flutter sugerido | Estado | Seguridad |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `GET /users/me` | Si | Obtener auth user + profile + AI settings. | N/A | `{ "user": {...}, "profile": {...}, "ai_settings": {...} }` | `401`, `404` si defaults faltan | `HomeBootstrap/ProfileScreen`, `UserRepository.getMe`, `SessionCubit/UserCubit` | listo | Fuente unica para usuario autenticado; no exponer `password_hash`. |
| `PATCH /users/me/profile` | Si | Actualizar perfil personal. | `{ "display_name": "...", "preferred_name": "...", "timezone": "America/Tijuana", "locale": "es-MX", "avatar_url": "https://..." }` | `200 { "profile": {...} }` | `400` timezone/locale vacios, `401` | `EditProfileScreen`, `UserRepository.updateProfile`, `UserProfileCubit` | listo | No aceptar `user_id`; validar UX antes de enviar. |
| `GET /users/me/ai-settings` | Si | Leer configuracion de Sofia para el usuario. | N/A | `200 { "ai_settings": {...} }` | `401`, `404` | `AssistantSettingsScreen`, `UserRepository.getAiSettings`, `AiSettingsCubit` | listo | Controla memory/reminders/planning; Flutter debe respetar toggles. |
| `PATCH /users/me/ai-settings` | Si | Actualizar estilo/autonomia/toggles. | `{ "assistant_name": "Sofía", "communication_style": "concise", "proactivity_level": "medium", "autonomy_level": "manual", "memory_enabled": true, "reminders_enabled": true, "planning_enabled": true, "preferred_language": "es" }` | `200 { "ai_settings": {...} }` | `400` enums invalidos, `401` | `AssistantSettingsScreen`, `UserRepository.updateAiSettings`, `AiSettingsCubit` | listo | Autonomia no habilita ejecucion automatica; proposal-only sigue vigente. |
| `POST /users/me/onboarding/complete` | Si | Marcar onboarding completado. | `{}` o body vacio | `200 { "profile": {...} }` | `401`, `404` | `OnboardingFlow`, `UserRepository.completeOnboarding`, `OnboardingCubit` | listo | Solo afecta al usuario autenticado. |
| `GET /users/me/export` | Si | Exportar datos propios. | N/A | `200 { "user": {...}, "profile": {...}, "ai_settings": {...}, "activities": [], "reminders": [], "insights": {...}, "memories": [], "action_proposals": [] }` | `401` | `PrivacyExportScreen`, `PrivacyRepository.exportMe`, `PrivacyCubit` | listo | Solo datos propios; cuidado con mostrar contenido sensible en pantalla. |
| `POST /users/me/delete-request` | Si | Crear solicitud controlada de borrado. | `{ "reason": "optional" }` o `{}` | `202/201 { "delete_request": { "id": "...", "status": "pending", ... } }` | `401`, `409` solicitud existente | `DeleteAccountRequestScreen`, `PrivacyRepository.createDeleteRequest`, `PrivacyCubit` | parcial | No borra inmediatamente; no prometer borrado instantaneo. |

## Activities

| Endpoint | Auth | Proposito | Request | Response esperada | Errores | Flutter sugerido | Estado | Seguridad |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `POST /activities` | Si | Crear activity: task, habit, routine, goal, event o checklist. | `{ "type": "task", "title": "...", "description": "...", "priority": "medium", "start_at": null, "due_at": "...", "timezone": "America/Tijuana", "tags": [], "has_reminder": false, "reminder_at": null, "recurrence_rule": "", "ai_manageable": false, "ai_notes": "" }` | `201 { "activity": {...} }` | `400`, `401`, `403` planning disabled si aplica | `ActivityEditorScreen`, `ActivitiesRepository.create`, `ActivityEditorCubit` | listo | No enviar `user_id`; reminder automatico solo si `has_reminder=true` y permitido. |
| `GET /activities` | Si | Listar activities propias. | Query `status,type,priority,from,to,tag,limit,cursor` | `200 { "activities": [...], "next_cursor": "..." }` | `400` query invalida, `401` | `ActivitiesListScreen`, `ActivitiesRepository.list`, `ActivitiesCubit` | listo | Lista filtrada por JWT. |
| `GET /activities/{id}` | Si | Obtener detalle propio. | N/A | `200 { "activity": {...} }` | `401`, `404` | `ActivityDetailScreen`, `ActivitiesRepository.getById`, `ActivityDetailCubit` | listo | `404` tambien representa recurso de otro usuario. |
| `PATCH /activities/{id}` | Si | Actualizar campos parciales. | Cualquier subset de create + `status`. | `200 { "activity": {...} }` | `400`, `401`, `404` | `ActivityEditorScreen`, `ActivitiesRepository.update`, `ActivityEditorCubit` | listo | No cambiar ownership desde Flutter. |
| `POST /activities/{id}/complete` | Si | Completar activity. | `{}` | `200 { "activity": {...} }` | `401`, `404`, `409` estado invalido | `ActivityDetailScreen`, `ActivitiesRepository.complete`, `ActivityDetailCubit` | listo | No crea outcome automaticamente; Flutter puede sugerir registrar outcome. |
| `POST /activities/{id}/cancel` | Si | Cancelar activity. | `{}` | `200 { "activity": {...} }` | `401`, `404`, `409` | `ActivityDetailScreen`, `ActivitiesRepository.cancel`, `ActivityDetailCubit` | listo | Cancela reminders pendientes asociados segun backend. |
| `POST /activities/{id}/archive` | Si | Soft archive de activity. | `{}` | `200 { "activity": {...} }` | `401`, `404`, `409` | `ActivityArchiveAction`, `ActivitiesRepository.archive`, `ActivitiesCubit` | listo | No borrar fisicamente en cliente. |

## Reminders

| Endpoint | Auth | Proposito | Request | Response esperada | Errores | Flutter sugerido | Estado | Seguridad |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `POST /reminders` | Si | Crear reminder propio, opcionalmente asociado a activity. | `{ "activity_id": "...", "title": "...", "description": "...", "scheduled_at": "...", "timezone": "America/Tijuana", "recurrence_rule": "none", "recurrence_interval": 1, "recurrence_until": "", "recurrence_count": 0, "max_retries": 3 }` | `201 { "reminder": {...} }` | `400`, `401`, `403` reminders disabled, `404` activity no propia | `ReminderEditorScreen`, `RemindersRepository.create`, `ReminderEditorCubit` | listo | No enviar descripcion sensible completa si se usara push. |
| `GET /reminders` | Si | Listar reminders propios. | Query `activity_id,status,from,to,limit,cursor` | `200 { "reminders": [...], "next_cursor": "..." }` | `400`, `401` | `RemindersScreen`, `RemindersRepository.list`, `RemindersCubit` | listo | Filtrado por JWT. |
| `GET /reminders/due` | Si | Ver reminders due del usuario. | Query `limit` | `200 { "reminders": [...] }` | `400`, `401` | `DueRemindersPanel`, `RemindersRepository.listDue`, `DueRemindersCubit` | parcial | Util para UI/debug; delivery real lo maneja worker. |
| `GET /reminders/{id}` | Si | Obtener reminder propio. | N/A | `200 { "reminder": {...} }` | `401`, `404` | `ReminderDetailScreen`, `RemindersRepository.getById`, `ReminderDetailCubit` | listo | `404` para cross-user. |
| `PATCH /reminders/{id}` | Si | Actualizar reminder. | Subset de create sin `activity_id`. | `200 { "reminder": {...} }` | `400`, `401`, `404`, `409` | `ReminderEditorScreen`, `RemindersRepository.update`, `ReminderEditorCubit` | listo | Recurrence simple: `none,daily,weekly,monthly`. |
| `POST /reminders/{id}/cancel` | Si | Cancelar reminder. | `{}` | `200 { "reminder": {...} }` | `401`, `404`, `409` | `ReminderDetailScreen`, `RemindersRepository.cancel`, `ReminderDetailCubit` | listo | No cancelar reminders de otro usuario. |
| `POST /reminders/{id}/archive` | Si | Archivar reminder. | `{}` | `200 { "reminder": {...} }` | `401`, `404`, `409` | `ReminderArchiveAction`, `RemindersRepository.archive`, `RemindersCubit` | listo | Soft archive. |

## Insights

| Endpoint | Auth | Proposito | Request | Response esperada | Errores | Flutter sugerido | Estado | Seguridad |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `POST /insights/mood` | Si | Registrar mood/productividad. | `{ "activity_id": "...", "mood": "good", "energy_level": 4, "stress_level": 2, "focus_level": 4, "note": "...", "occurred_at": "...", "timezone": "America/Tijuana" }` | `201 { "mood_entry": {...} }` | `400`, `401`, `404` activity no propia | `MoodCheckInScreen`, `InsightsRepository.createMood`, `MoodCubit` | listo | No es diagnostico medico; tratar notas como sensibles. |
| `GET /insights/mood` | Si | Listar mood entries propias. | Query `from,to,limit` si disponible | `200 { "mood_entries": [...] }` | `400`, `401` | `MoodHistoryScreen`, `InsightsRepository.listMood`, `MoodCubit` | listo | Datos privados. |
| `POST /insights/outcomes` | Si | Registrar resultado de activity. | `{ "activity_id": "...", "result": "completed", "completion_quality": 4, "satisfaction_level": 5, "difficulty_level": 2, "note": "..." }` | `201 { "activity_outcome": {...} }` | `400`, `401`, `404` | `ActivityOutcomeScreen`, `InsightsRepository.createOutcome`, `OutcomeCubit` | listo | Activity debe pertenecer al usuario. |
| `GET /insights/outcomes` | Si | Listar outcomes propios. | Query `activity_id,from,to,limit` si disponible | `200 { "activity_outcomes": [...] }` | `400`, `401` | `OutcomesHistoryScreen`, `InsightsRepository.listOutcomes`, `OutcomeCubit` | listo | No cruzar usuarios. |
| `POST /insights/reflections` | Si | Crear reflexion textual. | `{ "activity_id": "...", "title": "...", "content": "...", "sentiment": "optional", "tags": [], "occurred_at": "..." }` | `201 { "reflection": {...} }` | `400`, `401`, `404` | `ReflectionEditorScreen`, `InsightsRepository.createReflection`, `ReflectionCubit` | listo | `content` puede ser sensible; no enviarlo a LLM por default. |
| `GET /insights/reflections` | Si | Listar reflexiones propias. | Query `tag,from,to,limit` si disponible | `200 { "reflections": [...] }` | `400`, `401` | `ReflectionsScreen`, `InsightsRepository.listReflections`, `ReflectionCubit` | listo | Solo usuario autenticado. |
| `GET /insights/summary` | Si | Resumen simple de insights. | Query `from,to` | `200 { "average_energy": 0, "average_stress": 0, "average_focus": 0, "total_outcomes": 0, "total_reflections": 0 }` | `400`, `401` | `InsightsDashboardScreen`, `InsightsRepository.getSummary`, `InsightsSummaryCubit` | listo | No interpretar como analisis clinico. |

## Memory

| Endpoint | Auth | Proposito | Request | Response esperada | Errores | Flutter sugerido | Estado | Seguridad |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `POST /memory` | Si | Crear memoria explicita. | `{ "type": "preference", "title": "...", "content": "...", "summary": "...", "source": "user_manual", "source_id": "", "confidence": "high", "importance": "medium", "tags": [], "visibility": "assistant_context" }` | `201 { "memory": {...} }` | `400`, `401`, `403` memory disabled | `MemoryEditorScreen`, `MemoryRepository.create`, `MemoryEditorCubit` | listo | Solo crear si `memory_enabled=true`; contenido sensible. |
| `GET /memory` | Si | Listar memorias propias no eliminadas. | Query `type,tag,importance,status,limit,cursor` | `200 { "memories": [...], "next_cursor": "..." }` | `400`, `401`, `403` | `MemoryListScreen`, `MemoryRepository.list`, `MemoryCubit` | listo | No mostrar `deleted`; visibility importa para AI context. |
| `GET /memory/{id}` | Si | Obtener memoria propia. | N/A | `200 { "memory": {...} }` | `401`, `403`, `404` | `MemoryDetailScreen`, `MemoryRepository.getById`, `MemoryDetailCubit` | listo | Cross-user debe verse como `404`. |
| `PATCH /memory/{id}` | Si | Actualizar memoria. | Subset de create + `status`. | `200 { "memory": {...} }` | `400`, `401`, `403`, `404` | `MemoryEditorScreen`, `MemoryRepository.update`, `MemoryEditorCubit` | listo | Evitar cambiar a visibility amplia sin confirmacion UX. |
| `POST /memory/{id}/archive` | Si | Archivar memoria. | `{}` | `200 { "memory": {...} }` | `401`, `403`, `404` | `MemoryDetailScreen`, `MemoryRepository.archive`, `MemoryDetailCubit` | listo | Soft archive. |
| `DELETE /memory/{id}` | Si | Soft delete de memoria. | N/A | `200/204` o `{ "memory": {...status:"deleted"} }` | `401`, `403`, `404` | `MemoryDetailScreen`, `MemoryRepository.delete`, `MemoryDetailCubit` | listo | No aparece en busquedas posteriores. |
| `POST /memory/search` | Si | Busqueda textual basica. | `{ "query": "texto", "type": "preference", "tag": "x", "importance": "high", "limit": 20 }` | `200 { "memories": [...] }` | `400`, `401`, `403` | `MemorySearchScreen`, `MemoryRepository.search`, `MemorySearchCubit` | listo | No embeddings/RAG; no incluir deleted. |

## Notifications

| Endpoint | Auth | Proposito | Request | Response esperada | Errores | Flutter sugerido | Estado | Seguridad |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `POST /notifications/device-tokens` | Si | Registrar token FCM del dispositivo. | `{ "platform": "ios", "token": "fcm-device-token" }` | `201 { "id": "...", "user_id": "...", "platform": "ios", "token_redacted": "abcd...7890", "enabled": true, "created_at": "...", "updated_at": "...", "last_seen_at": "..." }` | `400`, `401` | `NotificationSettingsScreen`, `NotificationsRepository.registerToken`, `NotificationsCubit` | parcial | No loggear token; FCM real no activo por default. |
| `GET /notifications/device-tokens` | Si | Listar tokens propios redactados. | N/A | `200 { "device_tokens": [...] }` | `401` | `NotificationSettingsScreen`, `NotificationsRepository.listTokens`, `NotificationsCubit` | parcial | Solo tokens del usuario; respuesta redacted. |
| `DELETE /notifications/device-tokens/{id}` | Si | Eliminar token propio. | N/A | `204` o `{ "message": "deleted" }` | `401`, `404` | `NotificationSettingsScreen`, `NotificationsRepository.deleteToken`, `NotificationsCubit` | parcial | `404` para tokens ajenos. |

## Tools

| Endpoint | Auth | Proposito | Request | Response esperada | Errores | Flutter sugerido | Estado | Seguridad |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `GET /tools` | Si | Listar tool registry disponible. | N/A | `200 { "tools": [...] }` | `401` | `DeveloperToolsScreen`, `ToolsRepository.list`, `ToolsCubit` | parcial | UI final no necesita exponerlo al usuario normal. |
| `GET /tools/{name}` | Si | Obtener definicion de tool. | N/A | `200 { "tool": { "name": "...", "input_schema": {...}, ... } }` | `401`, `404` | `DeveloperToolsScreen`, `ToolsRepository.get`, `ToolsCubit` | parcial | No usar schemas para saltarse policy. |

## AI Action Proposals

| Endpoint | Auth | Proposito | Request | Response esperada | Errores | Flutter sugerido | Estado | Seguridad |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `POST /ai/action-proposals` | Si | Crear proposal estructurada manual/runtime. | `{ "tool_name": "create_activity", "proposed_input": {...}, "reason": "...", "risk_level": "low" }` | `201 { "action_proposal": {...status:"proposed"} }` | `400` schema/tool, `401`, `403/409` policy | `ActionProposalDebugScreen`, `AiActionsRepository.create`, `ActionProposalsCubit` | parcial | No aceptar `user_id` en input; schema validation activa. |
| `GET /ai/action-proposals` | Si | Listar proposals propias. | Query `status,limit,cursor` si disponible | `200 { "action_proposals": [...] }` | `401` | `AssistantApprovalsScreen`, `AiActionsRepository.list`, `ActionProposalsCubit` | parcial | Mostrar razon/riesgo antes de confirmar. |
| `GET /ai/action-proposals/{id}` | Si | Detalle de proposal propia. | N/A | `200 { "action_proposal": {...} }` | `401`, `404` | `ActionProposalDetailScreen`, `AiActionsRepository.get`, `ActionProposalDetailCubit` | parcial | Cross-user como `404`. |
| `POST /ai/action-proposals/{id}/confirm` | Si | Confirmar proposal. | `{}` | `200 { "action_proposal": {...status:"confirmed"} }` | `401`, `404`, `409` estado invalido | `ActionProposalDetailScreen`, `AiActionsRepository.confirm`, `ActionProposalDetailCubit` | parcial | Confirmar no implica ejecucion automatica. |
| `POST /ai/action-proposals/{id}/reject` | Si | Rechazar proposal. | `{}` | `200 { "action_proposal": {...status:"rejected"} }` | `401`, `404`, `409` | `ActionProposalDetailScreen`, `AiActionsRepository.reject`, `ActionProposalDetailCubit` | parcial | Mantener audit trail. |
| `POST /ai/action-proposals/{id}/execute` | Si | Ejecutar explicitamente proposal confirmada. | `{}` | `200 { "action_proposal": {...status:"executed", "execution_result": {...}} }` | `401`, `403`, `404`, `409`, `500` | `ActionProposalDetailScreen`, `AiActionsRepository.execute`, `ActionProposalDetailCubit` | parcial | Solo ejecucion explicita; no auto-run. Usar feature flag. |

## AI Runtime

| Endpoint | Auth | Proposito | Request | Response esperada | Errores | Flutter sugerido | Estado | Seguridad |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `POST /ai/runtime/message` | Si | Enviar mensaje a runtime proposal-only. | `{ "message": "Recuérdame estudiar mañana a las 8", "dry_run": true }` | `200 { "mode": "dry_run|proposal", "context_summary": {...}, "available_tools": [...], "proposed_actions": [...], "assistant_message": "...", "request_id": "..." }` | `400` mensaje invalido, `401`, `409` action insegura | `AssistantChatScreen`, `AiRuntimeRepository.sendMessage`, `AssistantRuntimeCubit` | parcial | Default fake; Gemini opcional; no guardar prompts sensibles; no ejecutar automaticamente. |

## System

| Endpoint | Auth | Proposito | Request | Response esperada | Errores | Flutter sugerido | Estado | Seguridad |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `GET /health` | No | Diagnostico basico backend/persistencia. | N/A | `200 { "status": "ok", "environment": "...", "firestore": "ok|disabled|error" }` | `405` metodo invalido | `DiagnosticsScreen`, `SystemRepository.health`, `HealthCubit` | listo | No incluir secretos ni detalles internos sensibles. |

## Recomendacion de Cliente Flutter

- Crear una capa `ApiClient` con interceptores para bearer token, refresh pendiente/no implementado, JSON snake_case y errores normalizados.
- Separar repositorios por modulo: `AuthRepository`, `UserRepository`, `ActivitiesRepository`, `RemindersRepository`, `InsightsRepository`, `MemoryRepository`, `NotificationsRepository`, `ToolsRepository`, `AiActionsRepository`, `AiRuntimeRepository`, `SystemRepository`.
- Usar feature flags locales para `aiRuntime`, `actionProposalsExecute`, `notifications`, `privacyDeleteRequest` y cualquier provider real.
- Modelar `ApiException(statusCode, message, code?)` para mapear `401` a cierre de sesion, `403` a feature disabled, `409` a conflicto recuperable y `429` a retry/backoff.
