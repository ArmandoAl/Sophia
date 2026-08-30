# Manual Test Runbook v0.1

Fecha: 2026-07-05

## Objetivo

Guia final para que Armando pruebe Sofia Backend v0.1 manualmente de forma segura, primero con memory driver y despues, si aplica, con Firestore controlado.

## Reglas de Seguridad

- Usar datos sinteticos.
- No pegar API keys, tokens reales ni credenciales en reportes.
- Mantener `AI_RUNTIME_PROPOSAL_ONLY=true`.
- Mantener `AI_MODEL_PROVIDER=fake` para la prueba principal.
- Mantener `REMINDER_DELIVERY_PROVIDER=noop`.
- Mantener `FCM_DRY_RUN=true`.
- No ejecutar migraciones con `-apply=true`.
- No ejecutar delete executor contra Firestore real sin revisar requests aprobadas.

## Antes de Empezar

Desde la raiz del backend:

```sh
cd /Users/armandoalvarado/Documents/Sophia/sophia_ai_backend/sofia-backend
```

Validar tests:

```sh
env GOCACHE=/Users/armandoalvarado/Documents/Sophia/sophia_ai_backend/sofia-backend/.cache/go-build go test ./...
```

Resultado esperado:

```text
PASS
```

## Arrancar Backend con Memory

Este es el modo recomendado para la prueba manual principal.

```sh
ENV=development \
PORT=8080 \
PERSISTENCE_DRIVER=memory \
AI_MODEL_PROVIDER=fake \
AI_RUNTIME_PROPOSAL_ONLY=true \
REMINDER_DELIVERY_PROVIDER=noop \
FCM_DRY_RUN=true \
go run ./cmd/sofia
```

Validar health:

```sh
curl http://localhost:8080/health
```

Resultado esperado:

```json
{
  "status": "ok",
  "environment": "development",
  "firestore": "disabled"
}
```

## Arrancar Backend con Firestore

Usar solo despues de que memory pase.

```sh
ENV=development \
PORT=8080 \
PERSISTENCE_DRIVER=firestore \
FIRESTORE_PROJECT_ID=<project_id> \
FIRESTORE_DATABASE_ID=default \
GOOGLE_APPLICATION_CREDENTIALS=<path_to_credentials> \
AI_MODEL_PROVIDER=fake \
AI_RUNTIME_PROPOSAL_ONLY=true \
REMINDER_DELIVERY_PROVIDER=noop \
FCM_DRY_RUN=true \
go run ./cmd/sofia
```

Si usas emulator:

```sh
FIRESTORE_EMULATOR_HOST=127.0.0.1:8080
```

Resultado esperado de health con Firestore real/emulator configurado:

```json
{
  "status": "ok",
  "environment": "development",
  "firestore": "ok"
}
```

Si Firestore no esta configurado correctamente, detener y revisar `docs/environment.md`.

## Importar Postman Collection

Archivo:

```text
docs/postman_collection.json
```

Pasos:

1. Abrir Postman.
2. Import.
3. Seleccionar `docs/postman_collection.json`.
4. Confirmar variable `base_url=http://localhost:8080`.
5. No llenar `token`, `user_id`, `activity_id`, `reminder_id` ni `memory_id`; la coleccion los actualiza automaticamente.

## Orden Recomendado de Pruebas

Ejecutar en este orden:

1. `Health`
2. `Register`
3. `Login`
4. `Users Me`
5. `Update Profile`
6. `Get AI Settings`
7. `Update AI Settings`
8. `Create Activity`
9. `List Activities`
10. `Create Reminder`
11. `List Reminders`
12. `Create Mood`
13. `Create Memory`
14. `AI Runtime Fake Dry Run`
15. `Create Action Proposal`
16. `Confirm Action Proposal`
17. `Create Action Proposal To Reject`
18. `Reject Action Proposal`
19. `Export User Data`

## Resultados Esperados

| Prueba | Status esperado | Validacion |
| --- | --- | --- |
| Health | `200` | `status=ok`. |
| Register | `201` | Se guarda `user_id`. |
| Login | `200` | Se guarda `token`. |
| Users Me | `200` | Devuelve user/profile/AI settings. |
| Update Profile | `200` | Actualiza timezone/locale/preferred name. |
| Get AI Settings | `200` | Devuelve settings actuales. |
| Update AI Settings | `200` | `memory_enabled`, `reminders_enabled`, `planning_enabled` quedan `true`. |
| Create Activity | `201` | Se guarda `activity_id`. |
| List Activities | `200` | Lista incluye activities del usuario. |
| Create Reminder | `201` | Se guarda `reminder_id`. |
| List Reminders | `200` | Lista incluye reminders del usuario. |
| Create Mood | `201` | Crea mood sintetico. |
| Create Memory | `201` | Se guarda `memory_id`. |
| AI Runtime Fake Dry Run | `200` | No ejecuta acciones ni crea documentos de dominio. |
| Create Action Proposal | `201` | Se guarda proposal. |
| Confirm Action Proposal | `200` | Proposal pasa a confirmado. |
| Create Action Proposal To Reject | `201` | Se guarda segundo proposal. |
| Reject Action Proposal | `200` | Proposal pasa a rechazado. |
| Export User Data | `200` | Export contiene datos propios del usuario. |

## Validacion E2E Local Automatizada

Ademas de Postman, correr:

```sh
env GOCACHE=/Users/armandoalvarado/Documents/Sophia/sophia_ai_backend/sofia-backend/.cache/go-build go run ./cmd/tools/run-local-e2e -output-dir=docs
```

Revisar:

```text
docs/local_e2e_report.md
```

Debe mostrar:

- `proposal_only_preserved=true`
- `no_automatic_ai_execution=true`
- `explicit_execution_created_activity=true`

## Errores Comunes

| Sintoma | Causa probable | Accion |
| --- | --- | --- |
| `connection refused` | Backend no esta corriendo o puerto distinto. | Arrancar `go run ./cmd/sofia` y revisar `PORT`. |
| `401 invalid credentials` en rutas privadas | Token no existe o expiro. | Ejecutar `Login` otra vez en Postman. |
| `409` al crear memory | `memory_enabled=false`. | Ejecutar `Update AI Settings`. |
| `409` al crear reminder | `reminders_enabled=false`. | Ejecutar `Update AI Settings`. |
| `429 rate limit exceeded` | Demasiados intentos auth. | Esperar `AUTH_RATE_LIMIT_WINDOW` o reiniciar backend local. |
| `Firestore project id required` | Firestore sin `FIRESTORE_PROJECT_ID`. | Exportar variable correcta. |
| `database default does not exist` | Database ID incorrecto. | Revisar `FIRESTORE_DATABASE_ID`, normalmente `default`. |
| Gemini skipped | No hay `GEMINI_API_KEY`. | Esperado en prueba v0.1 manual principal. |
| FCM no envia notificacion | `FCM_DRY_RUN=true` o provider noop. | Esperado; no activar push real en esta prueba. |

## Como Resetear Datos Locales

### Memory Driver

Los datos viven en memoria del proceso.

Reset:

1. Detener backend con `Ctrl+C`.
2. Arrancar de nuevo con `PERSISTENCE_DRIVER=memory`.
3. Reimportar o resetear variables de la coleccion si quieres IDs limpios.

En Postman:

- borrar `token`,
- borrar `user_id`,
- borrar `activity_id`,
- borrar `reminder_id`,
- borrar `memory_id`,
- borrar `action_proposal_id`,
- borrar `rejected_action_proposal_id`,
- borrar `smoke_email` para generar uno nuevo.

### Firestore Emulator

Si usas emulator, detener y limpiar datos desde la UI/emulator o reiniciar sin persistencia exportada.

### Firestore Real

No borrar manualmente colecciones productivas durante esta prueba.

Usar solo datos sinteticos con email `example.test`.

## Que NO Activar Todavia

- `AI_RUNTIME_PROPOSAL_ONLY=false`
- `AI_MODEL_PROVIDER=gemini` para la prueba principal
- `FCM_DRY_RUN=false`
- `REMINDER_DELIVERY_PROVIDER=fcm` en entorno no controlado
- `REMINDER_WORKER_ENABLED=true` contra Firestore real sin plan de prueba
- Migraciones con `-apply=true`
- Delete executor contra Firestore real sin revisar requests `approved`
- Hardware/smart home
- RAG/embeddings
- Notificaciones push reales

## Cierre de Prueba

Completar:

```text
docs/v0_1_test_results_template.md
```

Guardar evidencia no sensible:

- status de tests,
- capturas de respuestas sin tokens,
- errores observados,
- decision final: pass/fail/blocked.

