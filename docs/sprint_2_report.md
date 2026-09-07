# Sprint 2 Report - Users Domain

Fecha: 2026-06-29

## Auditoria inicial

Se revisaron:

- `docs/sprint_0_5_report.md`
- `docs/sprint_1_report.md`
- `docs/architecture_state.md`
- Estado del grafo codebase-memory
- Implementacion actual de auth/infrastructure

Hallazgos:

- Auth estaba separado en domain/application/infrastructure/interfaces.
- Existian repositorios auth en memoria y Firestore.
- `GET /users/me` era minimo y devolvia solo auth user.
- No existian `internal/users/domain`, `internal/users/application` ni `internal/users/infrastructure`.
- Register no creaba perfil ni settings de Sofia.
- Firestore estaba listo para compartir cliente, pero users no tenia colecciones propias.

## Que cambio

- Se creo `internal/users/domain`.
- Se agrego `UserProfile`.
- Se agrego `AISettings`.
- Se crearon contratos `UserProfileRepository` y `AISettingsRepository`.
- Se implementaron repositorios en memoria y Firestore.
- Se agrego service de users con casos de uso:
  - `GetMe`
  - `UpdateProfile`
  - `GetAISettings`
  - `UpdateAISettings`
  - `CompleteOnboarding`
  - `ProvisionDefaults`
- `POST /auth/register` ahora crea profile y AI settings default mediante hook inyectado desde `cmd/sofia`.
- `GET /users/me` ahora devuelve auth user + profile + AI settings.
- Se agregaron endpoints profile/settings/onboarding.
- Se actualizaron tests HTTP.

## Endpoints

| Metodo | Ruta | Auth | Descripcion |
| --- | --- | --- | --- |
| `GET` | `/users/me` | Si | Devuelve auth user, profile y AI settings. |
| `PATCH` | `/users/me/profile` | Si | Actualiza profile del usuario autenticado. |
| `GET` | `/users/me/ai-settings` | Si | Devuelve AI settings. |
| `PATCH` | `/users/me/ai-settings` | Si | Actualiza AI settings. |
| `POST` | `/users/me/onboarding/complete` | Si | Completa onboarding. |

## Defaults

Profile:

- `timezone`: `America/Tijuana`
- `locale`: `es-MX`
- `onboarding_completed`: `false`

AISettings:

- `assistant_name`: `Sofía`
- `communication_style`: `warm`
- `proactivity_level`: `medium`
- `autonomy_level`: `suggestive`
- `memory_enabled`: `true`
- `reminders_enabled`: `true`
- `planning_enabled`: `true`
- `preferred_language`: `es`

## Validaciones

- `timezone` requerido/no vacio.
- `locale` requerido/no vacio.
- `assistant_name` requerido/no vacio.
- `preferred_language` requerido/no vacio.
- `proactivity_level`: `low`, `medium`, `high`.
- `autonomy_level`: `manual`, `suggestive`, `semi_autonomous`.
- Campos desconocidos se rechazan por `DecodeJSON`.

## Decisiones arquitectonicas

- Auth no importa users.
- El provisioning se inyecta en auth handler como hook post-register desde composition root.
- Users depende de `auth/domain.UserRepository` solo para leer el auth user en `GetMe`.
- Los repositorios Firestore de users viven en `internal/users/infrastructure`.
- No se implementaron Activities, Reminders, Memory ni IA.

## Pruebas ejecutadas

Comando:

```sh
env GOCACHE=/Users/armandoalvarado/Documents/Sophia/sophia_ai_backend/sofia-backend/.cache/go-build go test ./...
```

Casos cubiertos:

- Register exitoso.
- Register email duplicado.
- Register password debil.
- Login exitoso devuelve token.
- Login usuario inexistente devuelve `401`.
- Login password incorrecto devuelve `401`.
- Register crea profile default.
- Register crea AI settings default.
- `GET /users/me` devuelve auth user + profile + settings.
- `PATCH /users/me/profile` actualiza campos validos.
- `PATCH /users/me/profile` rechaza timezone vacio.
- `GET /users/me/ai-settings` devuelve settings.
- `PATCH /users/me/ai-settings` actualiza campos validos.
- `PATCH /users/me/ai-settings` rechaza autonomy level invalido.
- Usuario no autenticado recibe `401`.
- Health check y CORS basico siguen pasando.

## Deuda tecnica

- Provisioning de auth user + profile + AI settings no es transaccional en Firestore.
- Si falla el hook post-register, el auth user puede quedar creado sin profile/settings.
- Falta validar timezone contra IANA.
- Falta validar locale/preferred_language contra BCP 47.
- Falta test con Firestore emulator.
- Falta endpoint para reparar/provisionar usuarios legacy de forma administrativa.

## Riesgos para proximos sprints

- Antes de Activities/Reminders conviene decidir si `timezone` sera obligatorio para scheduling real.
- Memory debe respetar `memory_enabled`.
- Reminders debe respetar `reminders_enabled`.
- Planning debe respetar `planning_enabled` y `autonomy_level`.
