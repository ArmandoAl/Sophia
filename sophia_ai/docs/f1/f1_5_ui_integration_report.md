# F1.5 — UI Integration Report

Fecha: 2026-07-11  
Proyecto: `sophia_ai` (Flutter)  
Sprint: F1.5 — pantallas + navegación session-aware  
Fuente: `docs/backend_contract/flutter_canonical_api_contract_v0_1.md`

---

## Resumen

F1 queda funcionalmente cerrado en cliente: bootstrap de sesión, auth, onboarding, profile, assistant settings, diagnostics y shell autenticado, sin Activities ni F2–F5.

---

## Pantallas

| Pantalla | Ruta | Cubit |
| --- | --- | --- |
| `SplashSessionScreen` | `/splash` | `SessionCubit` |
| `LoginScreen` | `/login` | `AuthCubit` |
| `RegisterScreen` | `/register` | `AuthCubit` |
| `OnboardingFlowScreen` | `/onboarding` | `OnboardingCubit` |
| `ProfileScreen` | `/profile` | `UserProfileCubit` |
| `AssistantSettingsScreen` | `/assistant-settings` | `AiSettingsCubit` |
| `DiagnosticsScreen` | `/dashboard` | `HealthCubit` |
| `SettingsPage` (hub) | `/settings` | `SessionCubit` |
| Main shell | `/chat`, `/smart-home`, … | existente |

Smart Home / Chat UI prototype permanecen visibles en shell; **no** usan backend F2/F4.

---

## Bootstrap y navegación

`main.dart` → `initDependencies()` → `SessionCubit.bootstrap()` → `BlocProvider` + `AppRouter.create`.

Redirect puro: `lib/core/router/session_redirect.dart`

| Estado | Destino |
| --- | --- |
| initial/loading | `/splash` |
| failure + tokenRetained | `/splash` (retry / sign out local) |
| unauthenticated | `/login` |
| authenticated + `!onboardingCompleted` | `/onboarding` |
| authenticated + onboarding done | main (`/chat`); bloquea auth/onboarding |

---

## Decisiones de producto

### Register
**Sin auto-login.** Tras register → snackbar + navega a login (si hay GoRouter). Alineado con F1.4: backend no devuelve JWT en register.

### Autonomy UI
`semi_autonomous` se etiqueta como preferencia futura/controlada. Copy: proposal-only; no se presenta “ejecución autónoma activa”.

### Diagnostics
Reemplaza CPU/RAM mock en `/dashboard` por `HealthResponse` (`status`, `environment`, `firestore`).

---

## Estilos

Reutiliza `NeonWrapper`, `NeonButton`, `SophiaCard`, `AuthScaffold`, tema dark existente. Sin rewrite visual.

---

## Local backend (documentado; no verificado E2E en este sprint)

Backend (desde repo Go):

```bash
ENV=development \
PORT=8080 \
PERSISTENCE_DRIVER=memory \
AI_MODEL_PROVIDER=fake \
AI_RUNTIME_PROPOSAL_ONLY=true \
go run ./cmd/sofia
```

Flutter Web:

```bash
fvm flutter run -d chrome \
  --dart-define=SOFIA_API_BASE_URL=http://localhost:8080
```

**CORS:** no se modificó el backend. Si Flutter Web falla por CORS, reproducir y documentar antes de tocar Go.

**Comunicación real backend↔app:** no afirmada aquí (backend no levantado en esta sesión). `flutter build web` con dart-define **sí** compiló.

---

## Tests

`test/features/f1_ui_navigation_test.dart` (+ cubits/repos/models previos)

- redirect rules (unit)
- splash failure/retry
- login 401 / success → onboarding|/chat rules
- register sin token
- onboarding complete
- profile / AI settings update
- logout
- diagnostics ok/degraded

`fvm flutter test` → **82/82 passed**

---

## Analyze / build

- `flutter analyze`: solo infos preexistentes `neon_button` Color channels
- `flutter build web --dart-define=SOFIA_API_BASE_URL=http://localhost:8080` → **OK** (wasm dry-run warnings de `flutter_secure_storage_web`)

---

## Auditoría rápida F1

| Check | Estado |
| --- | --- |
| Token no logueado | PASS |
| No `user_id` en PATCH bodies | PASS (Update*Request) |
| Sin wrappers `data`/`items` | PASS |
| Sin `UserDataExport` | PASS |
| Sin `PagedResponse` genérico | PASS |
| Sin Activities/F2–F5 wiring | PASS |
| Sin Firestore SDK negocio | PASS |
| Sin Gemini obligatorio | PASS |

---

## Deuda UI

- Chat / Smart Home siguen siendo prototipo visual
- Settings legacy (voice/dark) removidos del hub; no son contrato backend
- SessionFailure recoverable requiere retry manual en splash
- Web token storage = browser storage (F1.1)
