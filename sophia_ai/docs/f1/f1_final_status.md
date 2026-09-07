# F1 — Final Status

Fecha: 2026-07-11  
Proyecto Flutter: `sophia_ai`  
Contrato: `docs/backend_contract/flutter_canonical_api_contract_v0_1.md`

## Veredicto

**F1 cliente: PASS (funcional en código + tests mock).**  
Comunicación HTTP real contra backend local: **no verificada en esta sesión** → ver plan manual.

F2+ **no iniciado**.

---

## Tabla de requisitos

| Requisito | Estado | Evidencia | Archivo relacionado |
| --- | --- | --- | --- |
| AppConfig + dart-define baseUrl | PASS | default localhost:8080, normalize | `lib/core/config/app_config.dart` |
| ApiClient GET/POST/PATCH/DELETE + Bearer | PASS | unit tests MockClient | `lib/core/network/api_client.dart` |
| ApiException 401/409/429/non-JSON | PASS | tests F1.1 | `lib/core/network/api_exception.dart` |
| TokenStorage seguro | PASS | Secure + Memory; no logs | `lib/core/storage/token_storage.dart` |
| FeatureFlags conservadores | PASS | all false defaults | `lib/core/config/feature_flags.dart` |
| Modelos Auth/Users/Health canónicos | PASS | fixtures + tests | `lib/core/models/**` |
| Enums Proactivity/Autonomy unknown+raw | PASS | tests | `lib/core/models/enums/**` |
| PATCH omit nulls / bool false | PASS | Update*Request tests | `update_*_request.dart` |
| SystemRepository `/health` | PASS | repo tests | `features/system/**` |
| AuthRepository register/login/logout | PASS | token policy tests | `features/auth/**` |
| UserRepository me/profile/settings/onboarding | PASS | flat bodies tests | `features/users/**` |
| SessionCubit bootstrap | PASS | cubit tests | `features/session/**` |
| AuthCubit login/register | PASS | no auto-login register | `auth_cubit.dart` |
| Profile/AiSettings/Onboarding/Health cubits | PASS | sync session tests | `features/users|system/**` |
| Splash + redirect navigation | PASS | `session_redirect` + UI tests | `session_redirect.dart`, screens |
| LoginScreen errores 401/429 | PASS | friendly copy + tests | `login_screen.dart` |
| RegisterScreen + confirm password | PASS | 409 copy; → login | `register_screen.dart` |
| OnboardingFlow mínimo | PASS | profile+AI+complete | `onboarding_flow_screen.dart` |
| ProfileScreen | PASS | update → session | `profile_screen.dart` |
| AssistantSettingsScreen | PASS | bools off; proposal-only copy | `assistant_settings_screen.dart` |
| DiagnosticsScreen real `/health` | PASS | no CPU/RAM as health | `diagnostics_screen.dart` |
| MainShell autenticado | PASS | GoRouter shell | `app_router.dart` |
| No Activities F2 | PASS | no repos/UI Activities | — |
| No UserDataExport / PagedResponse items | PASS | audit | — |
| No Firestore SDK / Gemini obligatorio | PASS | audit | — |
| flutter analyze (F1 clean) | PASS | 0 errors; 3 infos preexistentes UI | neon_button |
| flutter test | PASS | 82/82 | — |
| flutter build web + dart-define | PASS | build/web | — |
| E2E real vs backend memory | PARTIAL | comandos documentados; no ejecutado aquí | `f1_manual_test_plan.md` |
| CORS Flutter Web ↔ Go | BLOCKED/UNKNOWN | no reproducido; no cambios Go | manual plan |

---

## Entregables F1

| Doc | Path |
| --- | --- |
| F1.1 infra | `docs/f1/f1_1_core_infrastructure_report.md` |
| F1.2 models | `docs/f1/f1_2_canonical_models_report.md` |
| F1.3 repos | `docs/f1/f1_3_repositories_report.md` |
| F1.4 state | `docs/f1/f1_4_state_management_report.md` |
| F1.5 UI | `docs/f1/f1_5_ui_integration_report.md` |
| Manual plan | `docs/f1/f1_manual_test_plan.md` |
| Este status | `docs/f1/f1_final_status.md` |

---

## Confirmación F2

**F2 no está conectada.** No hay ActivitiesRepository, pantallas de activities/reminders de contrato, ni navegación F2. Chat/Smart Home del shell son prototipo UI previo, no integración backend F2/F4.
