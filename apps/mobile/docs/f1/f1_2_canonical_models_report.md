# F1.2 — Canonical Models Report

Fecha: 2026-07-11  
Proyecto: `sophia_ai` (Flutter)  
Sprint: F1.2 — modelos/enums Auth/Users/System  
Fuente de verdad: `docs/backend_contract/flutter_canonical_api_contract_v0_1.md`

---

## Resumen

Se implementaron los modelos y enums canónicos de F1 (Core/System, Auth, Users) con serialización manual snake_case, fixtures JSON y tests de contrato.

**No** se implementaron: repositories, UI, `UserDataExport`, `DeleteRequest`, Activities/Reminders/Insights/Memory/Notifications/AI, `PagedResponse`, refresh token.

---

## Decisión de serialización

| Opción | Decisión |
| --- | --- |
| `json_serializable` / Freezed | **No usados** |
| Motivo | El proyecto no los tenía. PATCH parcial (omit nulls + bool `false` explícito) y enums `unknown+raw` son más claros a mano. Evita `build_runner` en este sprint. |
| Alternativa | `fromJson` / `toJson` manual + `Equatable` |

`dart run build_runner build` **no aplica**.

---

## ApiError

**No se creó** un modelo `ApiError` separado.

`ApiException` (F1.1) ya cubre `{ "message": "..." }`, body no JSON, y `statusCode` / `rawBody`. Duplicarlo no aporta valor en F1.2.

---

## Modelos creados

| Modelo | Path | Shape |
| --- | --- | --- |
| `HealthResponse` | `lib/core/models/system/health_response.dart` | flat |
| `AuthUser` | `lib/core/models/auth/auth_user.dart` | flat; `role?` |
| `RegisterRequest` | `lib/core/models/auth/register_request.dart` | request |
| `LoginRequest` | `lib/core/models/auth/login_request.dart` | request |
| `LoginResponse` | `lib/core/models/auth/login_response.dart` | `{token,user}` |
| `UserProfile` | `lib/core/models/users/user_profile.dart` | flat |
| `AiSettings` | `lib/core/models/users/ai_settings.dart` | flat |
| `MeResponse` | `lib/core/models/users/me_response.dart` | `{user,profile,ai_settings}` |
| `UpdateProfileRequest` | `lib/core/models/users/update_profile_request.dart` | partial flat |
| `UpdateAiSettingsRequest` | `lib/core/models/users/update_ai_settings_request.dart` | partial flat |

Barrel: `lib/core/models/models.dart`

---

## Enums creados

| Tipo | Valores wire | Unknown strategy |
| --- | --- | --- |
| `ProactivityLevel` | `low`, `medium`, `high` | parse → `isKnown: false` + raw `value`; `toJson()` lanza si unknown |
| `AutonomyLevel` | `manual`, `suggestive`, `semi_autonomous` | igual |

Implementados como clases value (no `enum` Dart puro) para conservar raw desconocido sin crash en responses.

Paths:

- `lib/core/models/enums/proactivity_level.dart`
- `lib/core/models/enums/autonomy_level.dart`

---

## Wrappers reales respetados

| Endpoint / uso | Wrapper | Cumplido |
| --- | --- | --- |
| GET `/health` | **ninguno** (flat) | sí |
| POST register | flat `AuthUser` | sí |
| POST login | `{ token, user }` | sí |
| GET `/users/me` | `{ user, profile, ai_settings }` | sí |
| GET/PATCH ai-settings | flat `AiSettings` | sí |
| PATCH profile | flat fields (no `{profile:{}}`) | sí |
| PATCH ai-settings | flat fields (no `{ai_settings:{}}`) | sí |

No se inventaron wrappers `data` / `items`.

---

## Nullability

| Campo | Decisión |
| --- | --- |
| `AuthUser.role` | `String?` — ausente en register/login; presente en `/me` |
| `UserProfile.avatarUrl` | `String?` — omitempty backend; `""` tratado como null |
| Demás campos response F1 | non-null según contrato |
| PATCH request fields | todos `?`; null = omitir key |
| Fechas | `DateTime` vía `DateTime.parse` (RFC3339) |

---

## Comportamiento PATCH

### `UpdateProfileRequest.toJson()`

- Solo incluye keys no-null: `display_name`, `preferred_name`, `timezone`, `locale`, `avatar_url`
- **Nunca** envía `user_id`
- **Nunca** envuelve en `{ "profile": ... }`

### `UpdateAiSettingsRequest.toJson()`

- Solo keys no-null
- Bools: `false` explícito **sí** se serializa (`memory_enabled: false`)
- Enums: solo valores known (`toJson()`); unknown → `StateError`
- **Nunca** envía `user_id` ni wrapper `{ "ai_settings": ... }`

---

## JSON fixtures

Directorio: `test/fixtures/f1/`

| Fixture | Uso |
| --- | --- |
| `health_response.json` | Health flat |
| `auth_register_response.json` | AuthUser sin role |
| `auth_login_response.json` | LoginResponse |
| `users_me_response.json` | MeResponse completo |
| `profile_without_avatar.json` | Profile sin `avatar_url` |
| `profile_with_avatar.json` | Profile con avatar + RFC3339 fraccional |
| `ai_settings_defaults.json` | Defaults Sofía / warm / medium / suggestive |

---

## Tests ejecutados

Archivo: `test/core/models/canonical_models_test.dart`

Cobertura pedida:

- [x] register sin role
- [x] login response
- [x] users/me completo
- [x] profile sin / con avatar_url
- [x] AI settings defaults
- [x] proactivity low/medium/high
- [x] autonomy manual/suggestive/semi_autonomous
- [x] PATCH profile omite nulls
- [x] PATCH settings permite `false` explícito
- [x] fechas RFC3339
- [x] enum desconocido → comportamiento controlado

Comando: `fvm flutter test`  
Resultado: **34/34 passed** (18 F1.1 + 16 F1.2)

---

## `flutter analyze`

Exit 0. Solo infos **preexistentes** UI:

- `neon_button.dart` — `Color.red/green/blue` deprecated
- `settings_page.dart` — `activeColor` deprecated

Ningún issue en `lib/core/models/**`.

---

## Errores o desviaciones del contrato

Ninguna desviación funcional respecto al contrato canónico F1.

Notas menores (documentadas, no bloqueantes):

1. Enums como clases value en lugar de `enum` Dart — requerido para unknown+raw.
2. `AiSettings.toJson()` en responses usa `.value` (incluye unknown raw si se re-serializa localmente); requests usan `.toJson()` que rechaza unknown.
3. `SystemHealth` UI mock en `lib/features/dashboard/domain/system_health.dart` **no** se tocó; el canónico es `HealthResponse`.

---

## Archivos tocados (resumen)

**Creados:** modelos bajo `lib/core/models/**`, fixtures `test/fixtures/f1/**`, test `test/core/models/canonical_models_test.dart`, este reporte.

**Modificados:** ninguno de infra F1.1 / UI / DI (solo format incidental si `dart format` tocó archivos previos).

---

## Próximo paso sugerido

F1.3: repositories Auth / Users / System + SessionCubit, consumiendo estos modelos vía `ApiClient`.
