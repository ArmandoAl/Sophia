# F1.3 — Repositories Report

Fecha: 2026-07-11  
Proyecto: `sophia_ai` (Flutter)  
Sprint: F1.3 — System / Auth / User repositories  
Fuente de verdad: `docs/backend_contract/flutter_canonical_api_contract_v0_1.md`

---

## Resumen

Se implementaron repositories remotos F1 sobre `ApiClient` + modelos canónicos F1.2, con contratos abstractos, implementaciones HTTP, registro GetIt y tests con `MockClient`.

**No** se crearon Cubits, pantallas, refresh token, export, delete-request, ni módulos Activities+.

---

## Interfaces e implementaciones

| Contrato | Implementación | Ubicación |
| --- | --- | --- |
| `SystemRepository` | `SystemRepositoryImpl` | `lib/features/system/{domain,data}/` |
| `AuthRepository` | `AuthRepositoryImpl` | `lib/features/auth/{domain,data}/` |
| `UserRepository` | `UserRepositoryImpl` | `lib/features/users/{domain,data}/` |

Patrón: `abstract interface class` + `*Impl` remota. Errores HTTP **no** se envuelven en `Exception` genérica; propagan `ApiException` desde `ApiClient`.

Helper compartido: `lib/core/network/require_json_map.dart` (cast defensivo a `Map<String, dynamic>`).

DI (`lib/core/di/service_locator.dart`): lazy singletons de los tres repositories.

---

## Endpoints conectados

### SystemRepository

| Método | Path | Auth | Response |
| --- | --- | --- | --- |
| `getHealth()` | `GET /health` | No | flat `HealthResponse` |

### AuthRepository

| Método | Path | Auth | Request | Response / side-effect |
| --- | --- | --- | --- | --- |
| `register` | `POST /auth/register` | No | flat `RegisterRequest` | 201 flat `AuthUser` (sin token) |
| `login` | `POST /auth/login` | No | flat `LoginRequest` | `{token,user}` + **saveToken** |
| `logout` | — | — | — | `TokenStorage.clearToken()` local |

### UserRepository

| Método | Path | Auth | Request | Response |
| --- | --- | --- | --- | --- |
| `getMe` | `GET /users/me` | Bearer | — | `{user,profile,ai_settings}` |
| `updateProfile` | `PATCH /users/me/profile` | Bearer | flat partial | flat `UserProfile` |
| `getAiSettings` | `GET /users/me/ai-settings` | Bearer | — | flat `AiSettings` |
| `updateAiSettings` | `PATCH /users/me/ai-settings` | Bearer | flat partial | flat `AiSettings` |
| `completeOnboarding` | `POST /users/me/onboarding/complete` | Bearer | `{}` | flat `UserProfile` |

---

## Request / response reales respetados

- Register/login bodies planos; register **no** guarda token.
- Login response `{ token, user }` — único lugar que persiste JWT.
- Profile / AI settings PATCH: body plano, sin wrappers `{profile}` / `{ai_settings}`.
- PATCH omite nulls vía `Update*Request.toJson()`; bools `false` se envían.
- Onboarding: `{}`; response profile plano.
- `getMe` es el único wrapper nested F1 users.

---

## Política de token

| Evento | Acción |
| --- | --- |
| Login 200 | `TokenStorage.saveToken(token)` — **sin logs del token** |
| Login 401 / error | no escribe token (token previo intacto) |
| Register | no toca storage |
| Logout | `clearToken()` local only |
| User endpoints | Bearer automático vía `ApiClient` + storage |
| Refresh | **no implementado** |

---

## Errores mapeados

Propagados como `ApiException` (F1.1 + `isBadRequest` añadido en F1.3):

| Status | Helper | Uso típico |
| --- | --- | --- |
| 400 | `isBadRequest` | validación |
| 401 | `isUnauthorized` | credenciales / sesión |
| 409 | `isConflict` | email duplicado / conflicto |
| 429 | `isRateLimited` | rate limit auth |
| non-JSON | `message` = texto crudo | paths inválidos |

Repositories **no** hacen `catch (e) => Exception(...)`.

---

## Tests ejecutados

Archivo: `test/features/repositories_f1_test.dart`

### System
- [x] health parse correcto
- [x] firestore disabled
- [x] degraded / firestore error

### Auth
- [x] register 201
- [x] register 409
- [x] register 429
- [x] login 200 guarda token
- [x] login 401 no guarda token
- [x] logout limpia token

### Users
- [x] getMe wrapper
- [x] update profile body plano + omite nulls
- [x] get AI settings plano
- [x] update AI settings bools false
- [x] complete onboarding sin wrapper
- [x] bearer automático

Comando: `fvm flutter test`  
Resultado: **49/49 passed** (18 F1.1 + 16 F1.2 + 15 F1.3)

---

## `flutter analyze`

Exit 0. Solo infos UI **preexistentes** (`neon_button`, `settings_page`).  
Ningún issue en repositories nuevos.

---

## Inconsistencias detectadas

Ninguna contra el contrato canónico en estos endpoints.

Notas operativas (no bloqueos de código):

1. **Register no autentica** — tras register el cliente debe llamar login (o flujo UI futuro).
2. **Login 401 no limpia token previo** — deliberado; SessionCubit futuro puede decidir `clearToken` en 401 global.
3. **Onboarding envía `{}`** — backend ignora body; vacío es seguro.
4. **Dashboard `SystemHealth` mock** sigue desconectado de `SystemRepository` / `HealthResponse` (UI no tocada).

---

## Archivos creados / modificados

**Creados**

- `lib/features/system/domain/system_repository.dart`
- `lib/features/system/data/system_repository_impl.dart`
- `lib/features/auth/domain/auth_repository.dart`
- `lib/features/auth/data/auth_repository_impl.dart`
- `lib/features/users/domain/user_repository.dart`
- `lib/features/users/data/user_repository_impl.dart`
- `lib/core/network/require_json_map.dart`
- `test/features/repositories_f1_test.dart`
- `docs/f1/f1_3_repositories_report.md`

**Modificados**

- `lib/core/di/service_locator.dart` — registro repos
- `lib/core/network/api_exception.dart` — `isBadRequest`

---

## Próximo paso sugerido

F1.4: `SessionCubit` (+ Auth/Onboarding/Profile cubits según plan), sin conectar aún Activities ni AI runtime.
