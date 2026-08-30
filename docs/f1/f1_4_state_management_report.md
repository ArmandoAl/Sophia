# F1.4 — State Management Report

Fecha: 2026-07-11  
Proyecto: `sophia_ai` (Flutter)  
Sprint: F1.4 — Session / Auth / Profile / AI settings / Onboarding / Health cubits  
Fuente de verdad: `docs/backend_contract/flutter_canonical_api_contract_v0_1.md`

---

## Resumen

Se implementó la capa de estado F1 con **SessionCubit como única fuente de verdad** de sesión autenticada. Auth y feature cubits delegan hidratación/propagación a Session. UI no conectada (salvo comentarios de aislamiento en dashboard mock).

---

## Cubits creados / adaptados

| Cubit | Path | Rol |
| --- | --- | --- |
| `SessionCubit` | `lib/features/session/presentation/cubit/` | Bootstrap, hydrate, logout, apply profile/settings |
| `AuthCubit` | `lib/features/auth/presentation/cubit/` | register / login form state |
| `UserProfileCubit` | `lib/features/users/presentation/cubit/` | load / update profile → session |
| `AiSettingsCubit` | `lib/features/users/presentation/cubit/` | load / update settings → session |
| `OnboardingCubit` | `lib/features/users/presentation/cubit/` | profile → settings → complete |
| `HealthCubit` | `lib/features/system/presentation/cubit/` | `GET /health` |
| `DashboardCubit` | existente | **aislado** como mock CPU/RAM (no backend) |

---

## Estados definidos

### SessionState
`initial` → `loading` → `unauthenticated` | `authenticated(user, profile, aiSettings)` | `failure(message, tokenRetained)`

### AuthState
`initial` | `loading` | `loginSuccess` | `registerSuccess(user)` | `failure`

### UserProfileState / AiSettingsState
`initial` | `loading` | `loaded` | `failure(previous?)`

### OnboardingState
`initial` | `inProgress(step)` | `completed(profile)` | `failure(failedStep)`  
Steps: `profile` | `aiSettings` | `complete`

### HealthState
`initial` | `loading` | `healthy` | `degraded` | `failure`

---

## Flujo de bootstrap

```
App start
  → SessionCubit.bootstrap()
    → TokenStorage.readToken()
    → sin token → SessionUnauthenticated
    → con token → UserRepository.getMe()
        → 200 → SessionAuthenticated
        → 401 → clearToken → SessionUnauthenticated
        → otro error → SessionFailure(tokenRetained: true)
```

`main.dart` aún **no** llama `bootstrap()` (UI wiring = F1.5+). DI ya registra el singleton.

---

## Flujo login / logout

### Login
1. `AuthCubit.login` → `AuthRepository.login` (guarda JWT)
2. `SessionCubit.establishFromStoredToken()` → `getMe`
3. Si session autenticada → `AuthLoginSuccess`
4. Si session failure / unauthenticated → `AuthFailure` (sin duplicar user en Auth)

### Register (decisión documentada)
**Sin auto-login.** Register → `AuthRegisterSuccess(user)` sin token. El usuario debe hacer login después. Motivo: el backend no devuelve token en register; auto-login ocultaría fallos de login y duplicaría política de sesión.

### Logout
`SessionCubit.logout()` → `AuthRepository.logout()` (`clearToken`) → `SessionUnauthenticated`.

---

## Sincronización con SessionCubit

| Cubit | Propagación |
| --- | --- |
| Auth (login) | `establishFromStoredToken` / `applyMe` |
| UserProfile | `applyProfile` en success |
| AiSettings | `applyAiSettings` en success |
| Onboarding | `applyProfile` / `applyAiSettings` por paso; complete solo si backend OK |

Feature availability en `AiSettingsLoaded`: `memoryEnabled`, `remindersEnabled`, `planningEnabled` (derivados de settings; **no** activan ejecución autónoma ni FeatureFlags de runtime).

---

## DI actualizado

`lib/core/di/service_locator.dart` (GetIt existente):

- Singletons: `AppConfig`, `FeatureFlags`, `TokenStorage`, `ApiClient`, repositories, **`SessionCubit`**
- Factories: `AuthCubit`, `UserProfileCubit`, `AiSettingsCubit`, `OnboardingCubit`, `HealthCubit`
- Legacy factories: `DashboardCubit`, `SettingsCubit` (UI mock)

---

## SystemHealth aislado

- `SystemHealth` (CPU/RAM) documentado como **prototype UI only**
- `DashboardCubit` comenta que no es `GET /health`
- Salud real backend: `HealthCubit` + `HealthResponse`

---

## Tests

Archivo: `test/features/cubits_f1_test.dart` (fakes de repositories)

| Área | Casos |
| --- | --- |
| Session | sin token, getMe OK, 401 limpia, 500 retiene, logout |
| Auth | login OK, 401, 429, register OK, register 409 |
| Profile | update → session; failure conserva session |
| AI settings | update + toggle true→false en session |
| Onboarding | success; fallo intermedio; fallo complete |
| Health | ok, degraded, network failure |

`fvm flutter test` → **68/68 passed**

---

## `flutter analyze`

Exit 0. Solo infos UI preexistentes (`neon_button`, `settings_page`).  
Ningún issue en cubits F1.4.

---

## Riesgos / deuda

1. **`bootstrap()` no invocado en `main`** — pendiente wiring UI / splash.
2. **SessionFailure recoverable** — UI futura debe ofrecer retry sin forzar logout.
3. **401 en Profile/AI settings** llama `session.logout()` — agresivo pero correcto para sesión inválida.
4. **Dashboard mock** sigue visible en UI hasta F1.5 diagnostics.
5. **SettingsCubit legacy** (notifications/darkMode) no es `AiSettingsCubit`.
6. No hay `BlocProvider` en el árbol de widgets todavía.

---

## Próximo paso sugerido

F1.5: conectar router/splash a `SessionCubit.bootstrap`, pantallas login/register/onboarding/settings/diagnostics mínimas, sin Activities.
