# API Smoke Tests v0.1

Fecha: 2026-07-05

## Objetivo

Esta guia acompaña la coleccion Postman `docs/postman_collection.json` para probar manualmente Sofia Backend v0.1 sin datos sensibles reales.

## Variables

La coleccion usa estas variables:

| Variable | Valor sugerido | Uso |
| --- | --- | --- |
| `base_url` | `http://localhost:8080` | URL base de la API. |
| `token` | auto | JWT obtenido en login. |
| `user_id` | auto | ID del usuario registrado/login. |
| `activity_id` | auto | ID de la activity creada. |
| `reminder_id` | auto | ID del reminder creado. |
| `memory_id` | auto | ID de la memory creada. |

La coleccion tambien guarda `action_proposal_id` y `rejected_action_proposal_id` para encadenar confirm/reject.

## Preparacion

Arrancar API en modo local seguro:

```sh
ENV=development \
PORT=8080 \
PERSISTENCE_DRIVER=memory \
AI_MODEL_PROVIDER=fake \
AI_RUNTIME_PROPOSAL_ONLY=true \
go run ./cmd/sofia
```

Importar en Postman:

```text
docs/postman_collection.json
```

## Flujo Principal

Ejecutar requests en orden:

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

## Datos Sinteticos

- Email: `sofia.smoke.<timestamp>@example.test`
- Password: `SmokePass123!`
- Activity: revisar el checklist v0.1.
- Reminder: revisar status local.
- Memory: preferencia ficticia de trabajo.

No usar emails, tokens, nombres, notas ni contenido personal real.

## Resultados Esperados

| Request | Resultado esperado |
| --- | --- |
| Health | `200`, status `ok`. |
| Register | `201`, guarda `user_id`. |
| Login | `200`, guarda `token` y `user_id`. |
| Users Me | `200`, devuelve user/profile/settings. |
| Update Profile | `200`, actualiza timezone/locale/nombres. |
| AI Settings | `200`, mantiene memory/reminders/planning enabled. |
| Create Activity | `201`, guarda `activity_id`. |
| List Activities | `200`, incluye lista `activities`. |
| Create Reminder | `201`, guarda `reminder_id`. |
| List Reminders | `200`, incluye lista `reminders`. |
| Create Mood | `201`, crea mood sintético. |
| Create Memory | `201`, guarda `memory_id`. |
| AI Runtime Fake Dry Run | `200`, no ejecuta acciones. |
| Create Action Proposal | `201`, guarda proposal. |
| Confirm Action Proposal | `200`, status confirmado. |
| Reject Action Proposal | `200`, status rechazado. |
| Export User Data | `200`, exporta solo datos propios. |

## Notas QA

- Si `Create Memory` devuelve `409`, revisar que `memory_enabled=true` en AI settings.
- Si `Create Reminder` devuelve `409`, revisar que `reminders_enabled=true`.
- Si AI Runtime intenta ejecutar acciones automaticamente, es regresion: v0.1 debe ser proposal-only.
- La coleccion no prueba Gemini real, FCM real, worker ni Firestore migration.

