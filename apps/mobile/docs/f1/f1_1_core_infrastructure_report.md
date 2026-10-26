# F1.1 — Core HTTP Infrastructure Report

Fecha: 2026-07-11  
Proyecto: `sophia_ai` (Flutter)  
Sprint: F1.1 — infraestructura HTTP/config únicamente  
Fuente de verdad: `docs/backend_contract/flutter_canonical_api_contract_v0_1.md`

---

## Resumen

Se implementó la capa base de red y configuración para integración con Sofia Backend v0.1:

- `AppConfig` (base URL por ambiente / `--dart-define`)
- `FeatureFlags` (defaults conservadores)
- `ApiClient` + `ApiException` (HTTP JSON, Bearer, errores canónicos)
- `TokenStorage` (seguro + memoria para tests)
- Registro explícito en GetIt existente
- Unit tests con `MockClient` (sin backend real)

**No** se implementaron Auth/Users models, repositories, pantallas, refresh token, ni cambios al backend Go.

---

## Archivos creados

| Archivo | Rol |
| --- | --- |
| `lib/core/config/app_config.dart` | Base URL, timeout, normalización, `resolve()` |
| `lib/core/config/feature_flags.dart` | Flags locales F futuras |
| `lib/core/network/api_client.dart` | Cliente HTTP reusable |
| `lib/core/network/api_exception.dart` | Error tipado status/message/rawBody |
| `lib/core/storage/token_storage.dart` | `TokenStorage` + Secure + Memory |
| `test/core/network/api_client_test.dart` | Unit tests F1.1 |
| `docs/f1/f1_1_core_infrastructure_report.md` | Este reporte |

## Archivos modificados

| Archivo | Cambio |
| --- | --- |
| `pubspec.yaml` | + `http`, + `flutter_secure_storage` |
| `lib/core/di/service_locator.dart` | Registro de `AppConfig`, `FeatureFlags`, `TokenStorage`, `ApiClient` |

### Formato incidental (`dart format .`)

`dart format .` también reformateó archivos UI preexistentes (sin cambio semántico):

- `lib/core/router/page_transitions.dart`
- `lib/core/widgets/action_proposal_card.dart`
- `lib/core/widgets/neon_button.dart`
- `lib/core/widgets/schedule_conflicts_card.dart`
- `lib/features/chat/domain/entities/chat_message.dart`
- `lib/features/settings/presentation/cubit/settings_state.dart`
- `lib/features/settings/presentation/pages/settings_page.dart`

---

## Arquitectura final

```
lib/core/
  config/
    app_config.dart          # SOFIA_API_BASE_URL / default localhost:8080
    feature_flags.dart       # gates locales (todas off por default)
  network/
    api_client.dart          # GET/POST/PATCH/DELETE + JSON + Bearer
    api_exception.dart       # 401/409/429 helpers
  storage/
    token_storage.dart       # SecureTokenStorage | MemoryTokenStorage
  di/
    service_locator.dart     # composición explícita GetIt (ya existía)
```

### Decisiones de naming / rutas

| Decisión | Motivo |
| --- | --- |
| Rutas bajo `lib/core/{config,network,storage}` | Coinciden con el layout pedido y con `lib/core/` existente (`di`, `router`, `theme`). |
| Paquete `http` (no Dio) | El proyecto no tenía cliente HTTP. `http` es suficiente, liviano, y `MockClient` cubre tests. Evita un segundo stack. |
| GetIt sin injectable | Ya había `get_it` con composición explícita. No se introdujo DI codegen. |
| `TokenStorage` abstracto | Permite `SecureTokenStorage` en app y `MemoryTokenStorage` en tests sin platform channels. |

---

## Dependencias añadidas

| Paquete | Versión | Uso |
| --- | --- | --- |
| `http` | ^1.5.0 | Cliente HTTP + `MockClient` en tests |
| `flutter_secure_storage` | ^9.2.4 | Persistencia JWT |

### Implicaciones web / mobile

- **iOS/Android:** Keychain / EncryptedSharedPreferences (apropiado para JWT).
- **Web:** el plugin usa almacenamiento del navegador (menos seguro que mobile). Aceptable para desarrollo local; no tratar como vault. Preferir HTTPS y sesiones cortas en producción web.
- **No** se hardcodearon secretos ni URLs de producción.

---

## Comportamiento del ApiClient

- Métodos: `get`, `post`, `patch`, `delete`
- Headers: `Accept` + `Content-Type: application/json`
- Bearer automático si `TokenStorage.readToken()` no es null/vacío y `authenticated: true`
- Timeout: `AppConfig.requestTimeout` (default 30s) → `ApiException(statusCode: 0)`
- Success 2xx:
  - body vacío → `null`
  - JSON objeto o lista raíz → `dynamic` parseado (sin wrappers `data`/`items`)
- Errors:
  - JSON `{ "message": "..." }` → `ApiException.message`
  - body no JSON → message = texto crudo
  - helpers: `isUnauthorized` (401), `isConflict` (409), `isRateLimited` (429)

---

## Estrategia de token storage

```dart
abstract class TokenStorage {
  Future<String?> readToken();
  Future<void> saveToken(String token);
  Future<void> clearToken();
}
```

- Producción DI: `SecureTokenStorage`
- Tests: `MemoryTokenStorage`
- Clave interna: `sofia_access_token`
- **No** se imprime el token en logs
- **No** hay refresh token (fuera de alcance F1.1)

---

## Feature flags

| Flag | Default | Notas |
| --- | --- | --- |
| `aiRuntimeEnabled` | `false` | F4 |
| `aiActionExecutionEnabled` | `false` | Obligatorio off |
| `notificationsEnabled` | `false` | F5 |
| `privacyDeleteRequestEnabled` | `false` | F5+ |
| `smartHomeEnabled` | `false` | Fuera de Backend v0.1 |

Ninguna feature gated se implementó en este sprint.

---

## AppConfig

- Default: `http://localhost:8080`
- Override: `--dart-define=SOFIA_API_BASE_URL=...`
- `normalizeBaseUrl` elimina trailing `/`
- `resolve(path)` evita dobles slash

Ejemplo:

```bash
fvm flutter run -d chrome --dart-define=SOFIA_API_BASE_URL=http://localhost:8080
```

---

## Tests ejecutados

Archivo: `test/core/network/api_client_test.dart`

Cobertura F1.1:

- [x] baseUrl default
- [x] override vía constructor / normalize (proxy de dart-define en tests)
- [x] Bearer con token
- [x] request sin Authorization
- [x] error JSON `message`
- [x] error body no JSON
- [x] 401 / 409 / 429
- [x] JSON array raíz
- [x] JSON objeto raíz
- [x] DELETE body vacío → `null`
- [x] POST/PATCH/DELETE methods
- [x] FeatureFlags defaults
- [x] MemoryTokenStorage round-trip

Comando: `fvm flutter test`  
Resultado: **18/18 passed**

---

## Resultados `flutter analyze`

Comando: `fvm flutter analyze`  
Exit: **0** (solo `info`)

### Introducidos por este sprint

Ninguno. Los archivos nuevos de `lib/core/config|network|storage` no reportan issues.

### Preexistentes (UI)

| Severidad | Ubicación | Regla |
| --- | --- | --- |
| info | `lib/core/widgets/neon_button.dart` | `deprecated_member_use` (`Color.red/green/blue`) |
| info | `lib/features/settings/presentation/pages/settings_page.dart` | `deprecated_member_use` (`activeColor`) |

---

## Resultados `flutter pub get` / `dart format`

- `fvm flutter pub get`: OK (deps resueltas)
- `fvm dart format .`: OK (9 archivos formateados, ver lista incidental arriba)

---

## Deuda / bloqueos

1. **`String.fromEnvironment` no se puede override en runtime** — el test de dart-define valida el default vacío; el override se prueba vía constructor + `normalizeBaseUrl` (mismo path de producción).
2. **Export privacy** sigue **BLOCKED** en contrato canónico — no afecta F1.1.
3. **Web token storage** menos seguro — documentado; evaluar cookie httpOnly / session strategy en fases posteriores si hay target web productivo.
4. **CORS** no validado aquí — al conectar Auth en F1.2 contra `localhost:8080` desde Flutter Web habrá que confirmar headers CORS del backend.
5. **Pantallas / Auth / Users** deliberadamente no conectados — siguiente sprint.

---

## Próximo paso sugerido (fuera de este prompt)

F1.2: modelos Auth/Users + repositories + SessionCubit según contrato canónico, sin AI runtime ni Smart Home.
