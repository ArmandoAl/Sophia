# Flutter F0 API Test Evidence

Fecha: 2026-07-10  
Sprint: F0 — Flutter Contract Verification (test evidence)

Fuentes inspeccionadas (solo lectura):

| Fuente | Path | Tipo |
| --- | --- | --- |
| HTTP handler suite | `sofia-backend/internal/server/routes_test.go` | Automated `httptest` |
| Runtime app tests | `internal/ai/runtime/application/runtime_test.go` | Unit/app |
| Notifications service | `internal/notifications/application/service_test.go` | Unit |
| Domain/unit misc | jwt, recurrence, jsonschema, delete_executor, … | Unit |
| Local app E2E | `cmd/tools/run-local-e2e` + `docs/local_e2e_report.md` | App E2E (**no HTTP mux**) |
| Postman | `sophia_ai/docs/backend_contract/postman_collection.json` | Manual smoke |
| Smoke / runbook | `api_smoke_tests.md`, `manual_test_runbook_v0_1.md` | Manual |
| Pagination doc | `sofia-backend/docs/sprint_12_report.md` | Doc (ejemplo `items` incorrecto) |

**Solo auditoría.** No se modificaron tests, backend ni Flutter. No se ejecutaron requests a producción.

### Niveles de evidencia

| Level | Significado |
| --- | --- |
| **E2E_VERIFIED** | E2E automatizado **sobre HTTP** con asserts de contrato wire |
| **HANDLER_TESTED** | `routes_test.go` ejercita el endpoint vía HTTP |
| **UNIT_ONLY** | Solo service/domain (sin JSON HTTP) |
| **SMOKE_DOCUMENTED** | Postman/smoke/runbook documentan el request |
| **IMPLEMENTATION_ONLY** | Inferido de handler/DTO; sin test HTTP ni smoke del endpoint |
| **NO_EVIDENCE** | Sin test ni doc |

Nivel primario = el más fuerte para contrato Flutter (wire). Smoke y app E2E son suplemento.

---

## 1. Executive Summary

**Endpoints Flutter auditados: 50**

| Evidence Level (primario) | Conteo | Lectura |
| --- | ---: | --- |
| E2E_VERIFIED (HTTP) | **0** | No hay suite E2E HTTP automatizada |
| HANDLER_TESTED | **43** | `routes_test.go` con status y/o body parcial |
| UNIT_ONLY | **4** | 3× device-tokens (service) + runtime success (HTTP solo 401 en routes) |
| IMPLEMENTATION_ONLY | **3** | `PATCH /reminders/{id}`, `GET /insights/outcomes`, `GET /ai/action-proposals` |
| SMOKE_DOCUMENTED (solo) | **0** | Postman complementa HANDLER; no es nivel único |
| NO_EVIDENCE | **0** | Todos existen en código |

### Hallazgos críticos

1. **Cero asserts de `next_cursor` / `cursor`** en tests Go. Paginación = implementación + doc (sprint 12 con wrapper `items` falso).
2. Wrappers **flat** (profile/activity/etc.) están *parcialmente* demostrados por decode en tests — suficiente si F1 ignora docs de envelope.
3. **`POST /ai/runtime/message` success** no está en `routes_test` (solo 401); success = unit + app E2E + Postman.
4. **Notifications** sin tests HTTP; solo unit + docs.
5. **`GET /insights/outcomes`**, **`GET /ai/action-proposals`**, **`PATCH /reminders/{id}`** sin success HTTP.
6. **429** solo en login/register rate-limit tests.
7. App E2E valida negocio sin validar JSON HTTP.

---

## 2. Endpoint Evidence Matrix

Columnas Success / Auth / Validation / Ownership / Pagination: **YES** | **PARTIAL** | **NO** | **N/A**

| Endpoint | Evidence Level | Success | Auth | Validation | Ownership | Pagination | Source |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `POST /auth/register` | HANDLER_TESTED | YES | N/A | YES | N/A | N/A | routes + Postman + app E2E |
| `POST /auth/login` | HANDLER_TESTED | YES | N/A | YES | N/A | N/A | routes + Postman; **429 YES** |
| `GET /users/me` | HANDLER_TESTED | YES | YES | N/A | N/A | N/A | routes + Postman |
| `PATCH /users/me/profile` | HANDLER_TESTED | YES | PARTIAL | YES | N/A | N/A | routes + Postman (flat fields asserted) |
| `GET /users/me/ai-settings` | HANDLER_TESTED | YES | PARTIAL | N/A | N/A | N/A | routes + Postman |
| `PATCH /users/me/ai-settings` | HANDLER_TESTED | YES | PARTIAL | YES | N/A | N/A | routes + Postman |
| `POST /users/me/onboarding/complete` | HANDLER_TESTED | YES | PARTIAL | N/A | N/A | N/A | routes |
| `GET /users/me/export` | HANDLER_TESTED | PARTIAL | YES | N/A | YES | N/A | routes + Postman + app E2E; casing no asertada |
| `POST /users/me/delete-request` | HANDLER_TESTED | YES | PARTIAL | NO | N/A | N/A | routes 202; no 409 duplicate |
| `POST /activities` | HANDLER_TESTED | YES | PARTIAL | YES | N/A | N/A | routes + Postman + app E2E |
| `GET /activities` | HANDLER_TESTED | YES | PARTIAL | PARTIAL | YES | **NO** | routes + Postman `limit`; **no cursor** |
| `GET /activities/{id}` | HANDLER_TESTED | YES | PARTIAL | N/A | YES | N/A | routes |
| `PATCH /activities/{id}` | HANDLER_TESTED | YES | PARTIAL | PARTIAL | PARTIAL | N/A | routes |
| `POST .../complete` | HANDLER_TESTED | YES | PARTIAL | NO | PARTIAL | N/A | routes |
| `POST .../cancel` | HANDLER_TESTED | YES | PARTIAL | NO | PARTIAL | N/A | routes |
| `POST .../archive` | HANDLER_TESTED | YES | PARTIAL | NO | PARTIAL | N/A | routes |
| `POST /reminders` | HANDLER_TESTED | YES | PARTIAL | YES | PARTIAL | N/A | routes + Postman + app E2E; **409 disabled** |
| `GET /reminders` | HANDLER_TESTED | YES | PARTIAL | NO | YES | **NO** | routes + Postman |
| `GET /reminders/due` | HANDLER_TESTED | YES | PARTIAL | NO | PARTIAL | N/A | routes |
| `GET /reminders/{id}` | HANDLER_TESTED | PARTIAL | PARTIAL | N/A | YES | N/A | cross-user 404; success indirect |
| `PATCH /reminders/{id}` | IMPLEMENTATION_ONLY | NO | NO | NO | NO | N/A | sin routes/Postman |
| `POST /reminders/{id}/cancel` | HANDLER_TESTED | YES | PARTIAL | NO | PARTIAL | N/A | routes |
| `POST /reminders/{id}/archive` | HANDLER_TESTED | YES | PARTIAL | NO | PARTIAL | N/A | routes |
| `POST /insights/mood` | HANDLER_TESTED | YES | YES | YES | N/A | N/A | routes + Postman |
| `GET /insights/mood` | HANDLER_TESTED | YES | YES | NO | YES | N/A | routes (`mood_entries`) |
| `POST /insights/outcomes` | HANDLER_TESTED | YES | PARTIAL | PARTIAL | YES | N/A | routes (403 cross-user) |
| `GET /insights/outcomes` | IMPLEMENTATION_ONLY | NO | PARTIAL | NO | NO | N/A | solo 401 en loop auth |
| `POST /insights/reflections` | HANDLER_TESTED | YES | PARTIAL | PARTIAL | N/A | N/A | routes |
| `GET /insights/reflections` | HANDLER_TESTED | YES | YES | NO | YES | N/A | routes |
| `GET /insights/summary` | HANDLER_TESTED | YES | YES | NO | N/A | N/A | routes |
| `POST /memory` | HANDLER_TESTED | YES | YES | PARTIAL | N/A | N/A | routes + Postman + app E2E; **409** |
| `GET /memory` | HANDLER_TESTED | YES | YES | NO | YES | **NO** | routes |
| `GET /memory/{id}` | HANDLER_TESTED | YES | PARTIAL | N/A | YES | N/A | routes |
| `PATCH /memory/{id}` | HANDLER_TESTED | YES | PARTIAL | PARTIAL | PARTIAL | N/A | routes |
| `POST /memory/{id}/archive` | HANDLER_TESTED | YES | PARTIAL | NO | PARTIAL | N/A | routes |
| `DELETE /memory/{id}` | HANDLER_TESTED | YES | PARTIAL | NO | PARTIAL | N/A | routes |
| `POST /memory/search` | HANDLER_TESTED | YES | YES | PARTIAL | PARTIAL | N/A | routes |
| `POST /notifications/device-tokens` | UNIT_ONLY | PARTIAL | NO | PARTIAL | PARTIAL | N/A | service_test; no HTTP |
| `GET /notifications/device-tokens` | UNIT_ONLY | PARTIAL | NO | NO | YES | N/A | service list; array HTTP untested |
| `DELETE /notifications/device-tokens/{id}` | UNIT_ONLY | PARTIAL | NO | NO | YES | N/A | service; `{status:deleted}` untested HTTP |
| `GET /tools` | HANDLER_TESTED | YES | PARTIAL | N/A | N/A | N/A | routes (`tools`) |
| `GET /tools/{name}` | HANDLER_TESTED | YES | PARTIAL | N/A | N/A | N/A | routes |
| `POST /ai/action-proposals` | HANDLER_TESTED | YES | PARTIAL | PARTIAL | N/A | N/A | routes + Postman + app E2E |
| `GET /ai/action-proposals` | IMPLEMENTATION_ONLY | NO | NO | NO | NO | NO | sin list HTTP test |
| `GET /ai/action-proposals/{id}` | HANDLER_TESTED | PARTIAL | PARTIAL | N/A | YES | N/A | cross-user + helpers |
| `POST .../confirm` | HANDLER_TESTED | YES | PARTIAL | PARTIAL | PARTIAL | N/A | routes + Postman |
| `POST .../reject` | HANDLER_TESTED | YES | PARTIAL | PARTIAL | PARTIAL | N/A | routes + Postman |
| `POST .../execute` | HANDLER_TESTED | YES | PARTIAL | YES | PARTIAL | N/A | routes + app E2E |
| `POST /ai/runtime/message` | UNIT_ONLY | PARTIAL | YES | PARTIAL | PARTIAL | N/A | routes **401 only**; unit + Postman + app E2E |
| `GET /health` | HANDLER_TESTED | YES | N/A | N/A | N/A | N/A | routes + Postman + runbook |

Auth PARTIAL = middleware compartido; no siempre hay 401 dedicado por ruta.

---

## 3. Response Contract Evidence

| Endpoint / grupo | Asserted in test | Observed smoke/E2E | Inferred handler |
| --- | --- | --- | --- |
| Register `{id,name,email}` | YES (routes) | Postman | — |
| Login `{token,user}` | YES | Postman | — |
| `GET /users/me` `{user,profile,ai_settings}` | YES (nested decode) | Postman | — |
| Profile PATCH **flat** | YES (`display_name` root) | Postman (no assert CI) | Confirma mismatch docs |
| AI settings GET/PATCH flat | YES parcial | Postman | — |
| Onboarding profile flat | YES (`onboarding_completed`) | — | — |
| Export | PARTIAL (activities/memories titles only) | Postman + app E2E counts | PascalCase risk **no** tested |
| Delete-request 202 body | PARTIAL (status code + audit) | — | shape from domain tags |
| Activity create/get **flat** | YES (`id`,`title`,…) | Postman | — |
| `GET /activities` `{activities}` | YES | Postman | `next_cursor` **not** asserted |
| Reminder create/list **flat** / `{reminders}` | YES | Postman | cursor **not** asserted |
| Mood create flat; list `mood_entries` | YES | Postman create | — |
| Outcomes create flat | YES | — | list key `outcomes` **untested** |
| Reflections list `reflections` | YES | — | — |
| Summary flat averages | YES | — | — |
| Memory CRUD/search `{memories}` | YES | Postman create | cursor **not** asserted |
| Device tokens | NO HTTP | docs only | handler: array + `{status:deleted}` |
| Tools `{tools}` / flat by name | YES | — | — |
| Proposal create/actions **flat** | YES (status fields) | Postman | list `{action_proposals}` **untested** |
| Runtime response | Unit asserts mode/actions | Postman + app E2E | HTTP handler untested success |
| Health | YES | runbook curl | — |

---

## 4. Critical Flutter Integration Gaps

Solo gaps que pueden romper cliente:

| Gap | Riesgo Flutter |
| --- | --- |
| Docs wrappers vs tests flat (profile/activity/…) | **Deserialización fallida** si se genera desde docs |
| Export shape/casing no asertada | **Modelo incompatible** / parse frágil |
| Outcomes list key `outcomes` sin test de lista | **Deserialización** si se usa `activity_outcomes` |
| Device-tokens GET array vs `{device_tokens}` | **Deserialización** (F5) |
| Device delete `{status:deleted}` sin HTTP test | Ack incorrecto |
| `next_cursor` / `cursor` sin tests | **Paginación rota** o infinite loop |
| Runtime HTTP success no en routes_test | Chat F4: shape/`mode=ready` no pinneado en HTTP |
| 401 siempre `invalid credentials` | **Logout** agresivo OK, pero copy confuso |
| 409 memory/reminders disabled (tested) vs docs 403 | **UX conflicto** incorrecta si se mapea 403 |
| PATCH reminder sin evidencia | Editor F2 puede sorprender |
| GET proposals list sin evidencia | Inbox F4 |
| Rate limit 429 solo auth | Otros endpoints: handling especulativo |

---

## 5. 401 / 403 / 404 / 409 / 429 Evidence

| Status | Endpoints (evidencia) | Evidence | Flutter Handling Confidence |
| --- | --- | --- | --- |
| **401** | login bad creds; `/users/me` no token; insights/memory auth loops; export; runtime no token; assertUnauthorized helper | HANDLER_TESTED | **HIGH** — logout/relogin seguro; mensaje siempre `invalid credentials` |
| **403** | outcome cross-user activity; (reminders activity not allowed en handler map) | HANDLER_TESTED (outcomes) | **MEDIUM** — no matriz completa por recurso |
| **404** | activity/reminder/memory/proposal cross-user o missing | HANDLER_TESTED | **HIGH** para recursos testeados; path basura = non-JSON `NotFound` **untested** as contract |
| **409** | register duplicate; reminders disabled; memory disabled; proposal transition/execute autonomy | HANDLER_TESTED | **HIGH** para esos casos; **LOW** para delete-request duplicate (no existe) |
| **429** | login + register rate limit tests | HANDLER_TESTED | **HIGH** solo auth; **N/A** resto |
| **202** | delete-request | HANDLER_TESTED | **HIGH** status; body PARTIAL |
| **400** | weak password; empty timezone; invalid activity type; invalid mood/level; invalid scheduled_at; invalid autonomy | HANDLER_TESTED | **HIGH** para casos cubiertos |

---

## 6. Pagination Evidence

| Endpoint | `limit` query | `cursor` query | `next_cursor` response | Fin de página (sin next) |
| --- | --- | --- | --- | --- |
| `GET /activities` | SMOKE (Postman `limit=20`); handler acepta; **no assert test** | IMPLEMENTATION_ONLY | IMPLEMENTATION_ONLY (omitempty en DTO) | IMPLEMENTATION_ONLY (omit si vacío / len==0) |
| `GET /reminders` | SMOKE `limit=20` | IMPLEMENTATION_ONLY | IMPLEMENTATION_ONLY | IMPLEMENTATION_ONLY |
| `GET /memory` | default 50 en code | IMPLEMENTATION_ONLY | IMPLEMENTATION_ONLY | IMPLEMENTATION_ONLY |
| `GET /ai/action-proposals` | code `limit` only | **N/A** (no cursor en handler) | **N/A** | N/A |

**Confirmación:** no hay test Go que cree >limit items, pase `cursor`, y aserte `next_cursor` igual al último id, ni que la última página omita `next_cursor`.

Sprint 12 documenta paginación pero muestra `"items": []` — **contradice** implementación (`activities`/`reminders`/`memories`).

---

## 7. Recommended Contract Tests Before Flutter Integration

*No implementar aquí — solo propuesta.*

### BLOCKER_BEFORE_F1

1. **Contract snapshot tests** (httptest): assert JSON keys de `GET /users/me`, `PATCH /profile`, `PATCH /ai-settings`, login/register — **flat vs wrapped** pinneado.
2. **401 matrix** mínima: un request sin token a `/users/me` + un endpoint protegido F1; assert `{"message":"invalid credentials"}`.
3. **Export warning test** o skip export en F1: si se expone export, assert al menos keys top-level reales (`mood_entries`, no `insights`) y documentar casing — o marcar export fuera de F1.
4. **Golden decode** de Activity/Reminder create response flat (ya casi cubierto; añadir `require.JSONEq` o struct completo).

### BEFORE_F2

5. **Pagination contract:** activities + reminders + memory con `limit=1`, assert `next_cursor`, segunda página con `cursor`, última sin `next_cursor`.
6. **`PATCH /reminders/{id}`** success + validation.
7. **`GET /reminders/{id}`** success assert (no solo cross-user).
8. Filters invalid query → 400.

### BEFORE_F3

9. **`GET /insights/outcomes`** success assert key **`outcomes`**.
10. Mood/reflection omitempty / nullable fields golden.
11. Memory search empty query → 400.

### BEFORE_F4

12. **`POST /ai/runtime/message` HTTP success** assert `mode` ∈ {`dry_run`,`ready`}, keys `proposed_actions`, `assistant_message`.
13. **`GET /ai/action-proposals`** list wrapper `action_proposals`.
14. Proposal confirm/reject/execute status machine already partial — añadir assert body flat.

### NICE_TO_HAVE

15. Notifications HTTP: register → list **array** → delete `{status:deleted}` + 401.
16. 429 message body shape.
17. Non-JSON 404 path garbage.
18. CORS (ya hay test) no crítico Flutter mobile.

---

## 8. Final Verdict

### 1. ¿Podemos iniciar F1 Core/Auth/Users?

**Sí, con condiciones:**

- Generar modelos desde **DTOs Go / evidencia de tests** (flat profile/settings, `MeResponse` nested), **no** desde wrappers de la doc Flutter.
- Tratar **export** como fuera de alcance F1 o LOW confidence.
- Aceptar que paginación no está pinneada (F1 casi no la necesita).

### 2. ¿Qué endpoints F1 tienen evidencia suficiente?

| Endpoint | ¿Suficiente para F1? |
| --- | --- |
| `POST /auth/register` | YES |
| `POST /auth/login` | YES (+ 429) |
| `GET /users/me` | YES |
| `PATCH /users/me/profile` | YES (flat) |
| `GET/PATCH /users/me/ai-settings` | YES |
| `POST /users/me/onboarding/complete` | YES |
| `GET /health` | YES |
| `GET /users/me/export` | **NO** (shape ambigua) |
| `POST /users/me/delete-request` | PARTIAL (status OK; body/409 débil) — diferir UI |

### 3. ¿Qué tests son BLOCKER_BEFORE_F1?

1. Pin de **response envelopes** auth/users (me + profile + ai-settings + login).  
2. Confirmación 401 message para session logout.  
3. Decisión explícita: **no generar UserDataExport** hasta test de shape — o test BLOCKER si F1 incluye export.

(Pagination no es blocker F1.)

### 4. ¿Es seguro generar modelos estrictos con `json_serializable`?

**Parcialmente sí para F1 auth/users/health/activity DTOs HTTP**, si:

- `explicitToJson` / campos required alineados a tests,
- `DateTime` desde RFC3339,
- **no** asumir wrappers documentados,
- PATCH requests usan omit-null,
- enums cerrados (ya validados en backend tests).

**No** es seguro ser estricto aún para: `UserDataExport`, device-token list envelope, outcomes list key sin test, Reminder campos domain-only, Runtime HTTP full payload, pagination `next_cursor`.

---

## Appendix — Evidence map quick refs

| Claim | Where |
| --- | --- |
| Handler suite | `internal/server/routes_test.go` (`TestRegister*` … `TestHealthCheck`) |
| 429 auth | `TestLoginRateLimit`, `TestRegisterRateLimit` |
| 409 reminders/memory disabled | `TestRemindersDisabledBlocksCreation`, `TestMemoryDisabledBlocksCreationAndSearch` |
| Ownership | `TestCannotAccessAnotherUsers*`, export ownership test |
| App E2E steps | `docs/local_e2e_report.md` |
| Postman flow | `api_smoke_tests.md` + `postman_collection.json` |
| Pagination doc (wrong `items`) | `docs/sprint_12_report.md` |
| Notifications unit | `internal/notifications/application/service_test.go` |
| Runtime unit | `internal/ai/runtime/application/runtime_test.go` |

*Fin evidencia F0.*
