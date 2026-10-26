# Flutter F0 HTTP Contract Verification

Fecha: 2026-07-10  
Sprint: F0 — Flutter Contract Verification  
Alcance: endpoints públicos documentados para Flutter vs implementación HTTP real en `sophia_ai_backend/sofia-backend`  
Fuente de verdad: `internal/server/routes.go`, handlers en `internal/*/interfaces/http`, DTOs request/response, `internal/platform/httpjson`, tests en `internal/server/routes_test.go`  
Documentación de referencia (no autoridad): `docs/backend_contract/flutter_endpoint_integration_contract.md`

**Esta actividad es solo auditoría.** No se modificó código productivo, handlers Go, modelos Flutter, Firestore ni migraciones.

---

## 1. Executive Summary

**Total de endpoints auditados: 50**

| Estado | Conteo | Lectura |
| --- | ---: | --- |
| CONFIRMED | 14 | Path/método/auth/status y forma de respuesta alineados con docs (detalle menor aceptable) |
| DOCUMENTATION_MISMATCH | 33 | Implementados, pero docs incorrectos en wrapper, status, errores, query o campos |
| IMPLEMENTATION_AMBIGUITY | 3 | Comportamiento o serialización real ambigua / peligrosa para generar DTOs |
| NOT_IMPLEMENTED | 0 | Todos los paths documentados están registrados |
| INTERNAL_ONLY | 0 | Ninguno de la lista Flutter es solo tooling interno |

### Hallazgo central (BLOCKER para F1)

La documentación Flutter asume wrappers del tipo `{"activity": {...}}`, `{"profile": {...}}`, `{"action_proposal": {...}}`.  
La implementación HTTP real, en la mayoría de creates/gets/actions, **devuelve el recurso en la raíz** (objeto plano). Las listas sí usan wrappers (`activities`, `reminders`, `memories`, etc.).

Generar modelos Flutter desde la documentación actual produciría parsers rotos.

### Otros hallazgos de alto impacto

1. **`GET /users/me/export`** serializa entidades de dominio sin DTO HTTP: structs sin `json` tags → claves **PascalCase** en activities/reminders/insights/memory/profile/ai_settings embebidos; `user` y `action_proposals` sí tienen tags snake_case. Forma distinta a docs (`insights` anidado vs listas planas).
2. **Outcomes list** usa `"outcomes"`, no `"activity_outcomes"`.
3. **Device tokens GET** devuelve **array JSON raíz**, no `{"device_tokens":[...]}`.
4. **Feature disabled** (memory/reminders) responde **409**, no 403 como documenta el contrato.
5. **Delete-request**: siempre **202** + objeto plano; **no hay 409** por solicitud existente.
6. **404 de path inválido** en resources usa `http.NotFound` (texto/HTML estándar), no `{"message":"..."}`.

---

## 2. Route Registration Verification

Fuente: `internal/server/routes.go`. Auth = middleware `Authn` (JWT Bearer). Auth endpoints usan rate limit opcional.

| Method | Documented Path | Actual Path | Auth | Status |
| --- | --- | --- | --- | --- |
| POST | `/auth/register` | `/auth/register` | No | CONFIRMED |
| POST | `/auth/login` | `/auth/login` | No | CONFIRMED |
| GET | `/users/me` | `/users/me` | JWT | CONFIRMED |
| PATCH | `/users/me/profile` | `/users/me/profile` | JWT | CONFIRMED (path); wrapper mismatch abajo |
| GET | `/users/me/ai-settings` | `/users/me/ai-settings` | JWT | CONFIRMED (path) |
| PATCH | `/users/me/ai-settings` | `/users/me/ai-settings` | JWT | CONFIRMED (path) |
| POST | `/users/me/onboarding/complete` | `/users/me/onboarding/complete` | JWT | CONFIRMED (path) |
| GET | `/users/me/export` | `/users/me/export` | JWT | CONFIRMED (path) |
| POST | `/users/me/delete-request` | `/users/me/delete-request` | JWT | CONFIRMED (path) |
| POST | `/activities` | `/activities` | JWT | CONFIRMED (path) |
| GET | `/activities` | `/activities` | JWT | CONFIRMED (path) |
| GET | `/activities/{id}` | `/activities/{id}` via `/activities/` | JWT | CONFIRMED (path) |
| PATCH | `/activities/{id}` | `/activities/{id}` | JWT | CONFIRMED (path) |
| POST | `/activities/{id}/complete` | `/activities/{id}/complete` | JWT | CONFIRMED (path) |
| POST | `/activities/{id}/cancel` | `/activities/{id}/cancel` | JWT | CONFIRMED (path) |
| POST | `/activities/{id}/archive` | `/activities/{id}/archive` | JWT | CONFIRMED (path) |
| POST | `/reminders` | `/reminders` | JWT | CONFIRMED (path) |
| GET | `/reminders` | `/reminders` | JWT | CONFIRMED (path) |
| GET | `/reminders/due` | `/reminders/due` | JWT | CONFIRMED (path) |
| GET | `/reminders/{id}` | `/reminders/{id}` | JWT | CONFIRMED (path) |
| PATCH | `/reminders/{id}` | `/reminders/{id}` | JWT | CONFIRMED (path) |
| POST | `/reminders/{id}/cancel` | `/reminders/{id}/cancel` | JWT | CONFIRMED (path) |
| POST | `/reminders/{id}/archive` | `/reminders/{id}/archive` | JWT | CONFIRMED (path) |
| POST/GET | `/insights/mood` | `/insights/mood` | JWT | CONFIRMED (path) |
| POST/GET | `/insights/outcomes` | `/insights/outcomes` | JWT | CONFIRMED (path) |
| POST/GET | `/insights/reflections` | `/insights/reflections` | JWT | CONFIRMED (path) |
| GET | `/insights/summary` | `/insights/summary` | JWT | CONFIRMED (path) |
| POST/GET | `/memory` | `/memory` | JWT | CONFIRMED (path) |
| GET/PATCH/DELETE | `/memory/{id}` | `/memory/{id}` | JWT | CONFIRMED (path) |
| POST | `/memory/{id}/archive` | `/memory/{id}/archive` | JWT | CONFIRMED (path) |
| POST | `/memory/search` | `/memory/search` | JWT | CONFIRMED (path) |
| POST/GET | `/notifications/device-tokens` | `/notifications/device-tokens` | JWT | CONFIRMED (path) |
| DELETE | `/notifications/device-tokens/{id}` | `/notifications/device-tokens/{id}` | JWT | CONFIRMED (path) |
| GET | `/tools` | `/tools` | JWT | CONFIRMED (path) |
| GET | `/tools/{name}` | `/tools/{name}` | JWT | CONFIRMED (path) |
| POST/GET | `/ai/action-proposals` | `/ai/action-proposals` | JWT | CONFIRMED (path) |
| GET | `/ai/action-proposals/{id}` | `/ai/action-proposals/{id}` | JWT | CONFIRMED (path) |
| POST | `.../confirm\|reject\|execute` | mismos paths | JWT | CONFIRMED (path) |
| POST | `/ai/runtime/message` | `/ai/runtime/message` | JWT | CONFIRMED (path) |
| GET | `/health` | `/health` | No | CONFIRMED |

**Nota de routing:** Go `ServeMux` registra colecciones y prefijos (`/activities/`, `/reminders/`, etc.). Paths mal formados llaman `http.NotFound` (no JSON de `httpjson`).

---

## 3. Endpoint Contract Matrix

Convenciones observadas en código:

- Decode: `DisallowUnknownFields`; body único; límite `requestBodyMax`.
- Omitido en create (string/`bool` no-pointer): zero value (`""`, `false`, `0`, `null` slice).
- PATCH: campos `*T` — omitido = no cambia; `null` JSON deserializa a `nil` pointer = no cambia (igual que omitido para pointers). Excepción: fechas en activities usan `**time.Time` para distinguir clear vs omit en algunos paths.
- Fechas request: RFC3339 (`time.RFC3339`).
- Fechas response HTTP DTO: `time.Time` / `*time.Time` → encoding/json RFC3339; `omitempty` omite nil/zero según tag.
- Error estándar: `{"message":"..."}` vía `httpjson.WriteError` (salvo `http.NotFound` nativo).

### 3.1 Auth

| Endpoint | Request DTO | Success Status | Response Wrapper | Nullable/Omitted Risks | Verification Status |
| --- | --- | --- | --- | --- | --- |
| `POST /auth/register` | `RegisterRequest`: `name`, `email`, `password` (todos string) | **201** | **Directo** `{id,name,email}` | Campos vacíos → 400 dominio; unknown fields → 400 body | CONFIRMED |
| `POST /auth/login` | `LoginRequest`: `email`, `password` | **200** | **Directo** `{token,user:{id,name,email}}` | Credenciales inválidas → 401 `invalid credentials` | CONFIRMED |

Diferencias: password mínimo 8 (`ErrWeakPassword`). Rate limit puede devolver 429 (middleware). Register no devuelve token (hay que login).

### 3.2 Users / Privacy

| Endpoint | Request DTO | Success Status | Response Wrapper | Nullable/Omitted Risks | Verification Status |
| --- | --- | --- | --- | --- | --- |
| `GET /users/me` | N/A | **200** | `{user, profile, ai_settings}` | `avatar_url` omitempty si vacío | CONFIRMED |
| `PATCH /users/me/profile` | pointers: `display_name`, `preferred_name`, `timezone`, `locale`, `avatar_url` | **200** | **Directo profile** (NO `{"profile":...}`) | timezone/locale `""` → 400; omitido = no change | DOCUMENTATION_MISMATCH |
| `GET /users/me/ai-settings` | N/A | **200** | **Directo ai_settings** | — | DOCUMENTATION_MISMATCH |
| `PATCH /users/me/ai-settings` | pointers bool/string de settings | **200** | **Directo ai_settings** | enums inválidos → 400; omitido = no change | DOCUMENTATION_MISMATCH |
| `POST /users/me/onboarding/complete` | **body ignorado** (no decode) | **200** | **Directo profile** | Cualquier body no leído | DOCUMENTATION_MISMATCH |
| `GET /users/me/export` | N/A | **200** | Ver §4 / ambigüedad | Domain dump PascalCase + shape distinta a docs | IMPLEMENTATION_AMBIGUITY |
| `POST /users/me/delete-request` | `{reason}` opcional; body vacío permitido si `ContentLength==0` | **202** | **Directo** `DeleteRequest` (NO wrapper) | Docs 201/409 inexistentes; permite múltiples pending | DOCUMENTATION_MISMATCH |

`user` en `/users/me`: `{id,name,email,role}` (sin password).  
`profile`: snake_case DTO con `created_at`/`updated_at` RFC3339.  
`ai_settings`: enums `proactivity_level`=`low|medium|high`; `autonomy_level`=`manual|suggestive|semi_autonomous`.

Export real (`UserDataExport`):

```json
{
  "generated_at": "...",
  "user": { /* auth User snake_case + role/timestamps; password omitted */ },
  "profile": { /* domain UserProfile → likely PascalCase keys */ },
  "ai_settings": { /* domain AISettings → likely PascalCase */ },
  "activities": [ /* domain Activity → PascalCase */ ],
  "reminders": [ /* PascalCase */ ],
  "mood_entries": [...],
  "activity_outcomes": [...],
  "reflections": [...],
  "memories": [...],
  "action_proposals": [ /* snake_case tags */ ]
}
```

Docs esperaban `insights: {...}` anidado y wrappers homogéneos — **incorrecto**.

### 3.3 Activities

| Endpoint | Request DTO | Success Status | Response Wrapper | Nullable/Omitted Risks | Verification Status |
| --- | --- | --- | --- | --- | --- |
| `POST /activities` | `createActivityRequest` | **201** | **Directo activity** | `start_at`/`due_at`/`reminder_at` null/omit → nil; `timezone` requerido; defaults type=`task`, priority=`medium` | DOCUMENTATION_MISMATCH |
| `GET /activities` | query filters | **200** | `{activities, next_cursor?}` | `next_cursor` = último id si len>0 | CONFIRMED |
| `GET /activities/{id}` | N/A | **200** | **Directo** | 404 not found / cross-user | DOCUMENTATION_MISMATCH |
| `PATCH /activities/{id}` | update pointers + `**time` dates | **200** | **Directo** | null vs omit en fechas es sutil (`**time.Time`) | DOCUMENTATION_MISMATCH |
| `POST .../complete` | no body | **200** | **Directo** | Docs 409 estado inválido: **no** mapeado en handler (complete no valida transición conflict) | DOCUMENTATION_MISMATCH |
| `POST .../cancel` | no body | **200** | **Directo** | idem | DOCUMENTATION_MISMATCH |
| `POST .../archive` | no body | **200** | **Directo** | idem | DOCUMENTATION_MISMATCH |

Enums dominio: type `task|habit|routine|goal|event|checklist`; status `pending|active|completed|canceled|archived`; priority `low|medium|high|urgent`.

Docs `403 planning disabled`: **no existe** en activities.  
`has_reminder` + reminders disabled → **409** (`ErrRemindersDisabled`), no 403.

Activity response omitempty: `start_at`, `due_at`, `completed_at`, `canceled_at`, `archived_at`, `reminder_at`, `recurrence_rule`, `ai_notes`.

### 3.4 Reminders

| Endpoint | Request DTO | Success Status | Response Wrapper | Nullable/Omitted Risks | Verification Status |
| --- | --- | --- | --- | --- | --- |
| `POST /reminders` | create: `activity_id`, `title`, `description`, `scheduled_at` (req RFC3339), `timezone`, recurrence fields, `max_retries` | **201** | **Directo reminder** | reminders disabled → **409**; activity ajena → **403** | DOCUMENTATION_MISMATCH |
| `GET /reminders` | query | **200** | `{reminders, next_cursor?}` | — | CONFIRMED |
| `GET /reminders/due` | `limit?` | **200** | `{reminders, next_cursor?}` | mismo list wrapper | CONFIRMED |
| `GET/PATCH /reminders/{id}` | update pointers | **200** | **Directo** | — | DOCUMENTATION_MISMATCH |
| `POST .../cancel\|archive` | no body | **200** | **Directo** | — | DOCUMENTATION_MISMATCH |

**Importante:** HTTP `reminderResponse` **no expone** `recurrence_interval`, `recurrence_until`, `recurrence_count`, `recurrence_occurrences`, `processing_*`, `last_attempt_at` aunque existen en dominio. Docs/`flutter_models_needed.md` listan esos campos → mismatch de modelo.

Recurrence rule dominio: `none|daily|weekly|monthly`.

### 3.5 Insights

| Endpoint | Request DTO | Success Status | Response Wrapper | Nullable/Omitted Risks | Verification Status |
| --- | --- | --- | --- | --- | --- |
| `POST /insights/mood` | mood create; `occurred_at` RFC3339 req | **201** | **Directo mood_entry** | levels 1..5; mood enum; activity ajena → 403 | DOCUMENTATION_MISMATCH |
| `GET /insights/mood` | `from,to,limit,activity_id` | **200** | `{mood_entries:[...]}` **sin cursor** | — | CONFIRMED |
| `POST /insights/outcomes` | outcome create; `activity_id` req | **201** | **Directo** | — | DOCUMENTATION_MISMATCH |
| `GET /insights/outcomes` | query | **200** | `{outcomes:[...]}` **NO** `activity_outcomes` | — | DOCUMENTATION_MISMATCH |
| `POST /insights/reflections` | reflection; `occurred_at` req; `content` req | **201** | **Directo** | — | DOCUMENTATION_MISMATCH |
| `GET /insights/reflections` | `from,to,limit,activity_id` | **200** | `{reflections:[...]}` | Docs `tag` query: **no implementado** | DOCUMENTATION_MISMATCH |
| `GET /insights/summary` | `from,to,limit` (limit vía mismo filter) | **200** | Directo summary fields | — | CONFIRMED |

Mood: `very_low|low|neutral|good|great`.  
Outcome result: `completed|partially_completed|skipped|failed|rescheduled`.

### 3.6 Memory

| Endpoint | Request DTO | Success Status | Response Wrapper | Nullable/Omitted Risks | Verification Status |
| --- | --- | --- | --- | --- | --- |
| `POST /memory` | createMemoryRequest | **201** | **Directo** | memory disabled → **409** (docs 403) | DOCUMENTATION_MISMATCH |
| `GET /memory` | type,tag,importance,status,limit,cursor | **200** | `{memories, next_cursor?}` | — | CONFIRMED |
| `GET/PATCH /memory/{id}` | update pointers + status | **200** | **Directo** | — | DOCUMENTATION_MISMATCH |
| `POST /memory/{id}/archive` | no body | **200** | **Directo** | — | DOCUMENTATION_MISMATCH |
| `DELETE /memory/{id}` | N/A | **200** | **Directo memory** (status deleted), no 204 | Docs 200/204 ambiguo | DOCUMENTATION_MISMATCH |
| `POST /memory/search` | query,type,tag,importance,limit | **200** | `{memories, next_cursor?}` | query/filter requerido | CONFIRMED |

Enums memory alineados con docs (type/source/visibility/status/confidence/importance).

### 3.7 Notifications

| Endpoint | Request DTO | Success Status | Response Wrapper | Nullable/Omitted Risks | Verification Status |
| --- | --- | --- | --- | --- | --- |
| `POST /notifications/device-tokens` | `platform`, `token` | **201** | **Directo** deviceTokenResponse | platform `ios|android|web`; timestamps como **string** formateado | CONFIRMED (shape cercana a docs) |
| `GET /notifications/device-tokens` | N/A | **200** | **Array raíz** `[...]` | Docs wrapper `device_tokens` | DOCUMENTATION_MISMATCH |
| `DELETE .../{id}` | N/A | **200** | `{"status":"deleted"}` | Docs 204 o `message`; actual `status` | DOCUMENTATION_MISMATCH |

Timestamps: no `time.Time` JSON nativo; `Format("2006-01-02T15:04:05Z07:00")` → string.

### 3.8 Tools

| Endpoint | Request DTO | Success Status | Response Wrapper | Nullable/Omitted Risks | Verification Status |
| --- | --- | --- | --- | --- | --- |
| `GET /tools` | N/A | **200** | `{tools:[ToolDefinition...]}` | — | CONFIRMED |
| `GET /tools/{name}` | N/A | **200** | **Directo ToolDefinition** | Docs `{tool:{...}}` | DOCUMENTATION_MISMATCH |

### 3.9 AI Action Proposals

| Endpoint | Request DTO | Success Status | Response Wrapper | Nullable/Omitted Risks | Verification Status |
| --- | --- | --- | --- | --- | --- |
| `POST /ai/action-proposals` | tool_name, proposed_input (raw JSON), reason, risk_level, expires_at? | **201** | **Directo proposal** | expires_at RFC3339 opcional | DOCUMENTATION_MISMATCH |
| `GET /ai/action-proposals` | `status?`, `limit?` (default 50); **sin cursor** | **200** | `{action_proposals:[...]}` | Docs cursor: no | CONFIRMED (list) / mismatch cursor |
| `GET .../{id}` | N/A | **200** | **Directo** | — | DOCUMENTATION_MISMATCH |
| `POST .../confirm\|reject\|execute` | no body | **200** | **Directo** | 409 transitions; 403 tool/autonomy | DOCUMENTATION_MISMATCH |

Status: `proposed|confirmed|rejected|executed|failed|expired`. Risk: `low|medium|high`.

### 3.10 AI Runtime

| Endpoint | Request DTO | Success Status | Response Wrapper | Nullable/Omitted Risks | Verification Status |
| --- | --- | --- | --- | --- | --- |
| `POST /ai/runtime/message` | `message`, `dry_run` (bool; omit→false) | **200** | **Directo RuntimeResponse** | invalid message 400; unsafe 409 | CONFIRMED |

`RuntimeResponse`: `mode`, `request_id`, `context_summary`, `available_tools`, `proposed_actions`, `assistant_message`, `observability`.

### 3.11 System

| Endpoint | Request DTO | Success Status | Response Wrapper | Nullable/Omitted Risks | Verification Status |
| --- | --- | --- | --- | --- | --- |
| `GET /health` | N/A | **200** | Directo `{status,environment,firestore}` | Si firestore=`error` → `status`=`degraded` (no siempre `ok`) | CONFIRMED |

---

## 4. Response Wrapper Inventory

### Patrón A — Objeto recurso en la raíz (sin wrapper)

Usado por:

- Auth register user; login (login tiene `token`+`user`, no wrapper de entidad única extra)
- Profile PATCH / onboarding complete (profile fields en raíz)
- AI settings GET/PATCH (settings fields en raíz)
- Activity create/get/patch/actions
- Reminder create/get/patch/actions
- Mood/outcome/reflection create
- Memory create/get/patch/archive/delete
- Device token create
- Tool get by name
- Action proposal create/get/confirm/reject/execute
- Runtime message
- Health
- Delete-request create

Ejemplo real:

```json
{ "id": "...", "user_id": "...", "title": "...", "status": "pending", ... }
```

**Docs incorrectos** cuando dicen `{"activity":{...}}`, `{"profile":{...}}`, `{"memory":{...}}`, `{"action_proposal":{...}}`, `{"tool":{...}}`, `{"delete_request":{...}}`.

### Patrón B — Lista con clave de colección (+ cursor opcional)

| Wrapper | Endpoints |
| --- | --- |
| `{"activities":[...],"next_cursor":"..."}` | GET `/activities` |
| `{"reminders":[...],"next_cursor":"..."}` | GET `/reminders`, GET `/reminders/due` |
| `{"memories":[...],"next_cursor":"..."}` | GET `/memory`, POST `/memory/search` |
| `{"mood_entries":[...]}` | GET `/insights/mood` (sin cursor) |
| `{"outcomes":[...]}` | GET `/insights/outcomes` |
| `{"reflections":[...]}` | GET `/insights/reflections` |
| `{"tools":[...]}` | GET `/tools` |
| `{"action_proposals":[...]}` | GET `/ai/action-proposals` (sin cursor) |

`next_cursor`: presente con `omitempty` cuando hay al menos un item; valor = ID del último elemento (no opaque token firmado).

### Patrón C — Aggregate / me

```json
{ "user": {...}, "profile": {...}, "ai_settings": {...} }
```

Solo `GET /users/me`.

### Patrón D — Array JSON raíz

```json
[ { "id": "...", "token_redacted": "...", ... } ]
```

Solo `GET /notifications/device-tokens`.

### Patrón E — Status map

```json
{ "status": "deleted" }
```

`DELETE /notifications/device-tokens/{id}`.

### Patrón F — Error

```json
{ "message": "..." }
```

Casi todos los errores de handler/`httpjson`.

### Patrón G — Export (especial)

Ver §3.2 — mezcla de claves y casing; **no** usar docs Flutter como spec.

### No encontrado

- `{"items":[...]}` — **no** usado en API pública auditada.
- `{"activity":{...}}` wrapper — **no** usado.

---

## 5. Error Contract Verification

### Formato dominante

```json
{ "message": "<string>" }
```

Implementado en `internal/platform/httpjson/httpjson.go` (`ErrorResponse.Message` con tag `json:"message"`).

### Mensajes / códigos observables

| Situación | Status | message típico |
| --- | ---: | --- |
| Body JSON inválido / unknown field | 400 | `invalid request body` (o `err.Error()` en algunos handlers notifications/runtime) |
| Validación dominio | 400 | texto de `errors.New` del dominio |
| Sin/invalid JWT | 401 | **siempre** `invalid credentials` (también login fallido) |
| Activity ajena en reminders/insights | 403 | `activity does not belong to the authenticated user` |
| Tool disabled / autonomy | 403 | dominio AI actions |
| Recurso no encontrado (handler) | 404 | `activity not found`, `reminder not found`, etc. |
| Path resource inválido | 404 | **`http.NotFound`** — **no** JSON `message` |
| Memory/reminders disabled | **409** | `memory is disabled...` / `reminders are disabled...` |
| Proposal transition / unsafe runtime | 409 | dominio |
| Rate limit auth | 429 | (rate limiter middleware) |
| Panic/unexpected | 500 | `internal server error` |
| Method no permitido | 405 | `method not allowed` |

### Desviaciones del contrato documental

1. Docs: feature disabled → 403. Real memory/reminders → **409**.
2. Docs: activity complete invalid state → 409. Real: no mapeo de conflicto de transición en handler activities.
3. Docs: delete-request duplicate → 409. Real: **no** implementado.
4. 401 genérico `invalid credentials` también para token ausente (no distingue “missing token”).
5. Auth package define `ErrorResponse` local con campo Go `Error` taggeado `json:"message"` — no usado por `httpjson` write path; irrelevante si se usa solo `httpjson`.

---

## 6. Query Parameter Verification

Defaults de list: **limit=50** si omitido o 0 (application layer).

### GET `/activities`

| Parámetro | Tipo | Req | Default | Validación |
| --- | --- | --- | --- | --- |
| `status` | string | no | vacío | enum status si no vacío |
| `type` | string | no | vacío | enum type |
| `priority` | string | no | vacío | enum priority |
| `from` | RFC3339 | no | nil | parse error → 400/500 path según handler |
| `to` | RFC3339 | no | nil | from≤to |
| `tag` | string | no | vacío | lowercased |
| `limit` | int | no | 50 | >0 (0→50); inválido → ErrInvalidLimit |
| `cursor` | string | no | vacío | id cursor |

### GET `/reminders`

| Parámetro | Tipo | Req | Default | Validación |
| --- | --- | --- | --- | --- |
| `activity_id` | string | no | vacío | — |
| `status` | string | no | vacío | enum |
| `from`/`to` | RFC3339 | no | nil | rango |
| `limit` | int | no | 50 | positivo |
| `cursor` | string | no | vacío | — |

### GET `/reminders/due`

| Parámetro | Tipo | Req | Default | Validación |
| --- | --- | --- | --- | --- |
| `limit` | int | no | 50 | positivo |

### Insights GET (`/mood`, `/outcomes`, `/reflections`, `/summary`)

| Parámetro | Tipo | Req | Default | Validación |
| --- | --- | --- | --- | --- |
| `activity_id` | string | no | vacío | — |
| `from`/`to` | RFC3339 | no | nil | rango |
| `limit` | int | no | 50 | positivo |
| `tag` | — | — | — | **Documentado para reflections; NO existe en ListFilter HTTP** |

### GET `/memory`

| Parámetro | Tipo | Req | Default | Validación |
| --- | --- | --- | --- | --- |
| `type` | string | no | vacío | — |
| `tag` | string | no | vacío | lowercased |
| `importance` | string | no | vacío | — |
| `status` | string | no | vacío | — |
| `limit` | int | no | 50 | positivo |
| `cursor` | string | no | vacío | — |

### GET `/ai/action-proposals`

| Parámetro | Tipo | Req | Default | Validación |
| --- | --- | --- | --- | --- |
| `status` | string | no | vacío | — |
| `limit` | int | no | 50 | Atoi o 400 `invalid limit` |
| `cursor` | — | — | — | **Documentado; NO implementado** |

---

## 7. Documentation Mismatches

| Severity | Endpoint | Documented | Actual | Flutter Impact |
| --- | --- | --- | --- | --- |
| BLOCKER | Casi todos los creates/gets de recurso | Wrapper `{"activity"\|"profile"\|"memory"\|...: {...}}` | Objeto en raíz | Parsers generados desde docs fallan |
| BLOCKER | `GET /users/me/export` | `insights` anidado + snake_case homogéneo | Listas planas + domain JSON PascalCase mixto | No generar `UserDataExport` aún |
| HIGH | `GET /insights/outcomes` | `activity_outcomes` | `outcomes` | Lista outcomes rota |
| HIGH | `GET /notifications/device-tokens` | `{device_tokens:[...]}` | `[...]` | Lista tokens rota |
| HIGH | `DELETE .../device-tokens/{id}` | 204 o `{message:deleted}` | 200 `{status:deleted}` | Delete ack incorrecto |
| HIGH | Memory/Reminders disabled | 403 | 409 | Feature-flag UX/error mapping incorrecto |
| HIGH | `PATCH` profile / AI settings / onboarding | Wrapped | Flat | Profile/settings models mal anidados |
| HIGH | Reminder model docs | Incluye recurrence_interval/until/count/processing_* | HTTP response omite esos campos | Modelo Reminder sobredimensionado / null forever |
| MEDIUM | `POST /users/me/delete-request` | 201/202 + wrapper + 409 duplicate | Solo 202 flat; sin 409 | Privacy flow incorrecto |
| MEDIUM | `GET /tools/{name}` | `{tool:{...}}` | Flat tool | Tool detail parse |
| MEDIUM | Action proposal single | `{action_proposal:{...}}` | Flat | Proposal detail parse |
| MEDIUM | `DELETE /memory/{id}` | 200/204 ambiguo | 200 + memory body | Esperar body |
| MEDIUM | Activities planning disabled | 403 | No existe | Dead error branch |
| MEDIUM | Activity complete/cancel 409 | Documentado | No conflict mapping | Dead branch |
| MEDIUM | Reflections `tag` query | Documentado | No implementado | Filtro UI inútil |
| MEDIUM | Action proposals `cursor` | Documentado | No implementado | Paginación falsa |
| LOW | `GET /health` | status siempre ok-ish | `degraded` si firestore error | Diagnostics copy |
| LOW | 401 message | genérico | siempre `invalid credentials` | UX copy |
| LOW | Resource 404 path | JSON message | net/http NotFound | Error parser |

---

## 8. Flutter Model Risks

**No generar todavía** (contrato real ambiguo o docs engañosos):

1. Cualquier modelo que asuma wrapper `activity` / `profile` / `ai_settings` / `reminder` / `mood_entry` / `activity_outcome` / `reflection` / `memory` / `action_proposal` / `tool` / `delete_request` en responses de create/get/action.
2. `UserDataExport` / `PrivacyExport` — serialización domain PascalCase + shape distinta.
3. `Reminder` completo según `flutter_models_needed.md` (campos HTTP ausentes).
4. `PagedResponse` genérico con `items` — el backend no usa `items`.
5. `DeviceTokenListResponse` con clave `device_tokens`.
6. Outcomes list model con clave `activity_outcomes`.
7. Error mappers que asuman 403 para memory/reminders disabled.
8. Delete-request conflict (409) handling.
9. Modelos que dependan de query `tag` en reflections o `cursor` en action-proposals.

---

## 9. Safe Models To Generate

Contrato **suficientemente confirmado** por handlers + tests HTTP (usar **código**, no docs wrappers):

| Modelo Flutter sugerido | Base real | Notas |
| --- | --- | --- |
| `ApiError` | `{message}` | Único error JSON estándar |
| `HealthResponse` | health.go | Incluir `degraded` |
| `AuthUser` (register/login user) | `{id,name,email}` | Login user sin role |
| `LoginResponse` | `{token,user}` | — |
| `MeResponse` | users responses | Único aggregate confirmado |
| `UserProfile` (HTTP DTO) | profileResponse | Para `/users/me` y flat PATCH response |
| `AiSettings` (HTTP DTO) | aiSettingsResponse | Flat GET/PATCH |
| `Activity` (HTTP DTO) | activityResponse | **Sin** wrapper; omitempty dates |
| `ActivityListResponse` | `{activities,next_cursor?}` | — |
| `CreateActivityRequest` / `UpdateActivityRequest` | requests.go | — |
| `Reminder` (HTTP subset) | reminderResponse | Solo campos del DTO HTTP |
| `ReminderListResponse` | `{reminders,next_cursor?}` | — |
| `CreateReminderRequest` / `UpdateReminderRequest` | requests.go | — |
| `MoodEntry` + `MoodEntriesListResponse` | responses | list key `mood_entries` |
| `ActivityOutcome` + list con key **`outcomes`** | responses | nombre clave crítico |
| `Reflection` + `ReflectionsListResponse` | responses | — |
| `InsightsSummary` | summaryResponse | — |
| `Memory` + `MemoryListResponse` | responses | — |
| `Create/Update/SearchMemoryRequest` | requests | — |
| `DeviceToken` | deviceTokenResponse | timestamps **String** |
| `RegisterDeviceTokenRequest` | request | — |
| `DeviceTokenDeleteResponse` | `{status:deleted}` | — |
| `ToolDefinition` | domain json tags | list via `{tools}` |
| `AiActionProposal` | domain json tags | flat en single; list `{action_proposals}` |
| `CreateActionProposalRequest` | handler request | + `expires_at` opcional |
| `RuntimeMessageRequest` / `RuntimeResponse` | runtime domain | `dry_run` default false |

Enums seguros (constantes dominio): Activity/Reminder/Memory/Mood/Outcome/Proposal/Risk/DevicePlatform/AI settings levels — alineados con docs en valores string.

---

## 10. Final Verdict

### 1. ¿El contrato documental es suficientemente confiable para generar modelos Flutter?

**No.** Es útil como mapa de rutas y intención, pero **no** como especificación de response shape. La divergencia sistemática de wrappers y varios mismatches de claves/status lo invalidan como fuente única para codegen F1.

### 2. ¿Qué dominios están completamente confirmados?

Ningún dominio está “completamente” limpio docs↔código. Los más cercanos a **usable desde código** (ignoring docs wrappers):

- **Auth** (register/login)
- **GET /users/me** + health
- **Activities list** + activity HTTP DTO (si se ignora wrapper documental)
- **Reminders list** + reminder HTTP DTO subset
- **Memory list/search**
- **Insights summary** + mood list key
- **Tools list**
- **AI runtime message**
- **Action proposals list** key

### 3. ¿Qué dominios necesitan corrección documental o backend?

| Dominio | Acción recomendada (fuera de F0) |
| --- | --- |
| Docs globales de wrappers | Reescribir contrato a “recurso en raíz” vs “lista con clave” |
| Privacy export | Backend: serializar vía HTTP DTOs snake_case **o** documentar PascalCase real |
| Insights outcomes list key | Docs → `outcomes` **o** backend renombrar (elegir uno) |
| Notifications list/delete | Docs → array raíz + `{status:deleted}` |
| Feature disabled codes | Docs 409 **o** backend 403 — alinear |
| Delete-request | Docs quitar 409/wrapper; o implementar unicidad |
| Reminder response fields | Docs reducir al DTO HTTP **o** ampliar response |
| Reflections tag / proposals cursor | Quitar de docs o implementar |

### 4. ¿Existe algún BLOCKER para Sprint F1?

**Sí — BLOCKER.**

No iniciar generación masiva de modelos Flutter ni repositories basados solo en `flutter_endpoint_integration_contract.md` / `flutter_models_needed.md` hasta:

1. Actualizar el contrato documental con wrappers reales, **o**
2. Congelar un “HTTP OpenAPI/snapshot” generado desde DTOs Go reales,

y resolver al menos: **export serialization**, **outcomes list key**, **device-tokens list shape**.

F1 puede proceder de forma segura **solo** para el subset de §9, tomando DTOs Go como fuente de verdad y tratando la documentación Flutter actual como desactualizada en response envelopes.

---

## Appendix A — Evidence map

| Área | Path |
| --- | --- |
| Routes | `sophia_ai_backend/sofia-backend/internal/server/routes.go` |
| Errors JSON | `.../internal/platform/httpjson/httpjson.go` |
| Auth | `.../internal/auth/interfaces/http/*` |
| Users | `.../internal/users/interfaces/http/*` |
| Privacy | `.../internal/privacy/interfaces/http/handler.go`, `.../application/service.go` |
| Activities | `.../internal/activities/interfaces/http/*` |
| Reminders | `.../internal/reminders/interfaces/http/*` |
| Insights | `.../internal/insights/interfaces/http/*` |
| Memory | `.../internal/memory/interfaces/http/*` |
| Notifications | `.../internal/notifications/interfaces/http/handler.go` |
| Tools | `.../internal/tools/interfaces/http/handler.go` |
| AI Actions | `.../internal/ai/actions/interfaces/http/handler.go` |
| AI Runtime | `.../internal/ai/runtime/interfaces/http/handler.go`, `.../domain/runtime.go` |
| HTTP tests | `.../internal/server/routes_test.go` |

## Appendix B — Status tally by endpoint (50)

**CONFIRMED (14):**  
`POST /auth/register`, `POST /auth/login`, `GET /users/me`, `GET /activities`, `GET /reminders`, `GET /reminders/due`, `GET /insights/mood`, `GET /insights/summary`, `GET /memory`, `POST /memory/search`, `POST /notifications/device-tokens`, `GET /tools`, `POST /ai/runtime/message`, `GET /health`

**IMPLEMENTATION_AMBIGUITY (3):**  
`GET /users/me/export`, (reminder HTTP field subset vs domain — counted under mismatches for endpoints but model risk), plus export-driven ambiguity; also treat **resource `http.NotFound` non-JSON** as cross-cutting ambiguity. Explicit endpoint ambiguity: **export**. Additional ambiguity flagged on **PATCH activities date null semantics** and **onboarding body ignored** — rolled into DOCUMENTATION_MISMATCH counts for those endpoints.

Recount for table §1:

- CONFIRMED: 14  
- DOCUMENTATION_MISMATCH: 33  
- IMPLEMENTATION_AMBIGUITY: 3 (`GET /users/me/export`; plus classify `GET /notifications/device-tokens` as mismatch not ambiguity; third: **Reminder response field gap** applied to model layer — for endpoint tally use: export, and note **DELETE memory status ambiguity** already mismatch).  

Precise IMPLEMENTATION_AMBIGUITY endpoints:

1. `GET /users/me/export`  
2. `PATCH /activities/{id}` (null vs omit on nested date pointers)  
3. `POST /users/me/onboarding/complete` (body contract undefined / ignored)

Remaining 33 = DOCUMENTATION_MISMATCH.  
NOT_IMPLEMENTED: 0. INTERNAL_ONLY: 0.  
14+33+3=50.

---

*Fin del informe F0.*
