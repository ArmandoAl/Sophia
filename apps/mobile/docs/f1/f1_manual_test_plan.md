# F1 — Manual Test Plan

Fecha: 2026-07-11  
Objetivo: validar Flutter Web ↔ Sofia Backend local (memory / fake AI).

**Nota:** este plan no se ejecutó automáticamente en el sprint F1.5. Completar y anotar resultados abajo.

---

## 1. Backend local (memory)

Desde el repo Go (`sofia-backend` / path local del equipo):

```bash
ENV=development \
PORT=8080 \
PERSISTENCE_DRIVER=memory \
AI_MODEL_PROVIDER=fake \
AI_RUNTIME_PROPOSAL_ONLY=true \
go run ./cmd/sofia
```

Esperado: proceso escuchando en `:8080`. Smoke:

```bash
curl -s http://localhost:8080/health
```

Esperado JSON flat: `status`, `environment`, `firestore`.

---

## 2. Flutter Web

Desde `sophia_ai`:

```bash
fvm flutter pub get
fvm flutter run -d chrome \
  --dart-define=SOFIA_API_BASE_URL=http://localhost:8080
```

Build estático (opcional):

```bash
fvm flutter build web \
  --dart-define=SOFIA_API_BASE_URL=http://localhost:8080
```

---

## 3. Flujo manual

| # | Paso | Acción UI | Esperado |
| --- | --- | --- | --- |
| 1 | Splash | Abrir app sin token | Splash breve → Login |
| 2 | Register | name, email, password≥8, confirm | Success → Login; **sin** sesión automática |
| 3 | Register 409 | mismo email | “already registered” |
| 4 | Login bad | password incorrecta | “Invalid email or password” (401) |
| 5 | Login OK | credenciales válidas | Si onboarding incompleto → Onboarding; si completo → Chat shell |
| 6 | users/me | implícito post-login | Session autenticada (Settings muestra email) |
| 7 | Onboarding | profile + AI settings → Finish | POST complete; entra a main |
| 8 | Profile | Settings → Profile → editar → Save | Datos persisten; session actualizada |
| 9 | AI settings | desactivar Memory → Save | `memory_enabled: false` aceptado |
| 10 | Diagnostics | System tab | status / environment / firestore (no CPU/RAM) |
| 11 | Health degraded | (si aplica) firestore error | UI degraded |
| 12 | Logout | Settings → Sign out | Login; token limpiado |
| 13 | Rate limit | muchos login/register | Mensaje 429 |

---

## 4. Resultados esperados (contrato)

- Bodies planos (sin `{profile:{}}` / `{ai_settings:{}}` en PATCH)
- Login response `{token,user}`; token **nunca** visible en UI/logs
- Register response AuthUser sin token
- Me wrapper solo en GET `/users/me`
- Onboarding no marca completo si falla un paso intermedio
- Autonomy copy no promete auto-ejecución

---

## 5. Plantilla de errores

```
Fecha:
Paso #:
Request (método/path):
Status:
Body (sin secrets):
Consola Flutter / Network:
¿CORS?: sí/no
Repro:
```

Si CORS bloquea `localhost:8080` desde Chrome: **documentar aquí antes de cambiar Go**.

---

## 6. Confirmación F2

Al cerrar esta prueba manual, confirmar:

- [ ] No se usaron endpoints `/activities` ni `/reminders` desde la app F1
- [ ] Chat UI no llama `/ai/runtime/message`
- [ ] Smart Home no habla con backend Sofia v0.1
- [ ] F2 sigue **no conectada**

---

## Resultados de esta sesión (automatizado)

| Check | Resultado |
| --- | --- |
| `flutter test` | 82 passed |
| `flutter analyze` | 0 errors (3 infos preexistentes neon_button) |
| `flutter build web` + dart-define | OK |
| Backend levantado + E2E manual | **No ejecutado** |
