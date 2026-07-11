# Flutter Backend Integration Audit

Fecha: 2026-07-10  
Alcance: `sophia_ai` Flutter vs Sofia Backend v0.1  
Fuentes: `docs/backend_contract/*` + inventario de `lib/`

## Veredicto

El proyecto Flutter actual es un **prototipo UI** (glassmorphism/neon) con Cubits locales y datos mock. **No hay capa de red, auth, repositorios ni modelos alineados al contrato**. La integración con Backend v0.1 requiere construir la infraestructura de cliente desde cero y reutilizar solo shells/widgets de presentación.

## Estado actual del proyecto

| Área | Estado | Evidencia |
| --- | --- | --- |
| HTTP / `ApiClient` | Ausente | Sin `dio`/`http` en `pubspec.yaml`; sin cliente en `lib/` |
| Auth / JWT | Ausente | Carpeta `lib/features/auth/` vacía; sin login/register |
| Secure token storage | Ausente | Sin `flutter_secure_storage` ni similar |
| Repositories | Ausentes | Solo comentario en `service_locator.dart` |
| Modelos API (snake_case) | Ausentes | Entidades locales no coinciden con contrato |
| `json_serializable` | Ausente | No está en dependencias |
| Feature flags | Ausentes | No hay gating de AI/notifications/delete |
| Navegación auth-aware | Ausente | `GoRouter` entra directo a `/chat` sin splash/sesión |
| DI (GetIt) | Parcial | Solo `DashboardCubit` y `SettingsCubit` registrados |

### Dependencias relevantes (`pubspec.yaml`)

Presentes: `bloc`, `flutter_bloc`, `equatable`, `get_it`, `go_router`, `google_fonts`, `flutter_animate`, `glassmorphism`.

Faltantes para v0.1: `dio` (o `http`), `flutter_secure_storage`, `json_annotation`/`json_serializable`/`build_runner`, config de `base_url` (env/flavors).

### Arquitectura observada

```
lib/
  core/          # theme, router, di, widgets visuales
  features/
    auth/        # vacío
    chat/        # UI + mock ChatMessageCubit
    dashboard/   # System diagnostics mock (CPU/RAM)
    insight/     # WeeklyInsightPage (no en router)
    reminders/   # RemindersPage (no en router)
    settings/    # toggles locales
    smart_home/  # fuera de Backend v0.1
```

Patrón: `presentation/cubit` + ocasional `domain/entities`. No hay `data/` ni `domain/repositories`.

## Inventario de pantallas existentes

| Pantalla actual | En router | Datos | Mapeo contrato | Reutilizable |
| --- | --- | --- | --- | --- |
| `ChatPage` | Sí (`/chat`) | Mock local | `AssistantChatScreen` (Fase 4) | **Sí** — shell UI + `ActionProposalCard` |
| `SmartHomePage` | Sí (`/smart-home`) | Mock local | Fuera de v0.1 | **No conectar** — ocultar/flag |
| `DashboardPage` | Sí (`/dashboard`) | Mock CPU/RAM | `DiagnosticsScreen` (Fase 1) | **Parcial** — layout; remapear a `GET /health` |
| `SettingsPage` | Sí (`/settings`) | Estado local | Hub de `SettingsStack` | **Sí** — shell; reemplazar toggles por profile/AI/privacy |
| `RemindersPage` | **No** | `TaskItem` mock | `ActivitiesListScreen` / `RemindersScreen` (Fase 2) | **Parcial** — lista/checkbox UI |
| `WeeklyInsightPage` | **No** | Accuracy + decisions mock | Mezcla insights + approvals | **Parcial** — cards Approve/Reject → proposals |

### Widgets core reutilizables

| Widget | Uso en integración |
| --- | --- |
| `MainWrapper` | Shell principal; hay que reordenar tabs |
| `NeonWrapper`, `SophiaCard` | Look & feel; mantener |
| `ActionProposalCard` | Base visual para AI proposals (adaptar a `AiActionProposal`) |
| `VoiceVisualizer` | Cosmético; no bloquea backend |
| `ScheduleConflictCard`, `MapLocationCard` | Demo chat; no mapear a API v0.1 |
| `DeviceCard`, `CircularLightControl`, `ActiveDevicesTable` | Smart home; no conectar |

## Cubits existentes vs contrato

| Cubit actual | Rol hoy | Destino recomendado |
| --- | --- | --- |
| `ChatMessageCubit` | Respuestas mock por keywords | Sustituir/envolver con `AssistantRuntimeCubit` (Fase 4) |
| `DashboardCubit` | CPU/RAM fake | Reemplazar por `HealthCubit` + `SystemRepository` |
| `SettingsCubit` | Toggles UI locales | Dividir en `UserProfileCubit`, `AiSettingsCubit`, `NotificationsCubit` |
| `RemindersCubit` | `TaskItem` local | Reemplazar por `ActivitiesCubit` / `RemindersCubit` de contrato |
| `InsightCubit` | Accuracy + decisions | Separar: `InsightsSummaryCubit` vs `ActionProposalsCubit` |
| `SmartHomeCubit` | Dispositivos mock | Mantener aislado; no cablear a backend |

## Modelos existentes vs contrato

| Modelo local | Coincide con contrato | Acción |
| --- | --- | --- |
| `ChatMessage` | No (`RuntimeResponse` / proposals) | Mantener UI; mapear desde runtime |
| `SystemHealth` (cpu/ram) | No (`HealthResponse`) | Reemplazar |
| `TaskItem` | No (`Activity` / `Reminder`) | Reemplazar |
| `DecisionItem` | No (`AiActionProposal`) | Reemplazar o adaptar |
| `SmartDevice` | N/A (fuera de scope) | No tocar en v0.1 |
| Auth/User/Profile/AiSettings/Activity/Reminder/Mood/Memory/… | **No existen** | Crear según `flutter_models_needed.md` |

## Navegación actual vs recomendada

**Actual** (`AppRouter`):

- `initialLocation: /chat`
- Shell: Chat → Smart Home → System → Settings
- Sin stacks unauthenticated/onboarding
- `RemindersPage` e `WeeklyInsightPage` huérfanas

**Contrato recomendado**:

- `UnauthenticatedStack`: login/register
- `OnboardingStack`: profile + AI settings + complete
- `MainShell`: activities, reminders, insights, memory, assistant, settings
- `SettingsStack`: profile, assistant settings, notifications, privacy, diagnostics

### Cambios de navegación necesarios

1. Splash/sesión: token local → `GET /users/me` o login.
2. Redirects por `SessionCubit` (auth / onboarding / main).
3. Reemplazar tab Smart Home por Activities (o Reminders) en v0.1.
4. Mover Smart Home detrás de feature flag / ruta oculta.
5. Registrar rutas hijas: profile, AI settings, activity detail/editor, etc.
6. Cablear `RemindersPage` / insights al shell o a Settings.

## Gaps por fase del roadmap

### Fase 1 — Identidad, Profile, AI Settings

| Requerido | Existe |
| --- | --- |
| Login/Register/Onboarding/Profile/AssistantSettings/Diagnostics | No (salvo shell Settings + Dashboard parcial) |
| `AuthRepository`, `UserRepository`, `SystemRepository` | No |
| `AuthCubit`, `SessionCubit`, `UserProfileCubit`, `AiSettingsCubit`, `OnboardingCubit`, `HealthCubit` | No |
| Token seguro + logout en 401 | No (Sign Out es stub) |

### Fase 2 — Activities y Reminders

| Requerido | Existe |
| --- | --- |
| List/detail/editor activities y reminders | Solo `RemindersPage` mock (huérfana) |
| Repos/cubits de contrato | No |
| Distinción Activity vs Reminder | UI mezcla ambos en `TaskItem` |

### Fase 3 — Insights y Memory

| Requerido | Existe |
| --- | --- |
| Mood/outcomes/reflections/memory screens | No |
| `WeeklyInsightPage` | UI demo; no es mood/summary API |

### Fase 4 — AI Runtime y Proposals

| Requerido | Existe |
| --- | --- |
| Chat UI | Sí (`ChatPage`) |
| Runtime real / dry_run | No (mock keywords) |
| Approvals inbox | Parcial en `WeeklyInsightPage` + `ActionProposalCard` |
| Confirm ≠ execute automático | TODOs en chat; proposals son smart-home demo |

### Fase 5 — Notifications / Privacy delete

| Requerido | Existe |
| --- | --- |
| Device tokens / FCM | Toggle local “Notifications” sin API |
| Export / delete-request | No |

## Riesgos de romper la UI actual

| Riesgo | Severidad | Mitigación |
| --- | --- | --- |
| Quitar Smart Home del shell | Alta percepción de cambio | Flag `smartHomeEnabled=false`; ruta oculta |
| Chat deja de “responder” al cablear runtime | Media | Mantener mock detrás de flag hasta Fase 4; o dry_run fake |
| Remapear Dashboard CPU/RAM → health | Baja | Misma página, nuevos campos |
| Settings pierde toggles voice/volume | Baja | Mover a sección “Local UX”; AI settings al backend |
| `RemindersPage` TaskItem → Activity/Reminder | Media | Nueva capa de modelos; adaptar tiles |
| `ActionProposalCard` asume acciones smart-home | Media | Generalizar a `toolName` + `proposedInput` + risk |
| Inyectar auth rompe `initialLocation: /chat` | Alta | Introducir splash + redirects sin borrar shell |
| GetIt incompleto + Cubits creados en `BlocProvider` | Media | Unificar registro DI al cablear repos |
| Input de chat glass no llama a Cubit | Baja (bug actual) | `_GlassChatInputArea` no envía mensajes; `_ChatInputArea` sí pero no se usa |

## Lo que NO conectar (confirmado)

- Smart home / hardware
- Gemini obligatorio
- FCM push visible
- Delete executor / borrado físico inmediato
- Migraciones legacy
- RAG / embeddings

## Conclusión de auditoría

Hay **buena base visual** (theme, shell, chat, settings, cards) y **cero infraestructura de backend**. La primera fase debe priorizar `ApiClient` + auth/sesión + profile/AI settings + health, sin tocar smart home ni runtime real. Las pantallas de reminders/insights existentes son plantillas UI, no implementaciones del contrato.
