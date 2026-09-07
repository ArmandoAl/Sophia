# v0.1 Test Results Template

Fecha de prueba:

Tester:

Entorno:

- [ ] Memory
- [ ] Firestore emulator
- [ ] Firestore real controlado

Commit/version:

## Configuracion

| Variable | Valor usado |
| --- | --- |
| `ENV` |  |
| `PORT` |  |
| `PERSISTENCE_DRIVER` |  |
| `FIRESTORE_PROJECT_ID` |  |
| `FIRESTORE_DATABASE_ID` |  |
| `AI_MODEL_PROVIDER` |  |
| `AI_RUNTIME_PROPOSAL_ONLY` |  |
| `REMINDER_DELIVERY_PROVIDER` |  |
| `FCM_DRY_RUN` |  |

No pegar secretos ni paths sensibles completos.

## Preflight

| Check | Resultado | Notas |
| --- | --- | --- |
| `go test ./...` | pass/fail |  |
| Backend arranca | pass/fail |  |
| `/health` responde | pass/fail |  |

## Postman Smoke Tests

| Orden | Request | Resultado | Notas |
| --- | --- | --- | --- |
| 1 | Health | pass/fail |  |
| 2 | Register | pass/fail |  |
| 3 | Login | pass/fail |  |
| 4 | Users Me | pass/fail |  |
| 5 | Update Profile | pass/fail |  |
| 6 | Get AI Settings | pass/fail |  |
| 7 | Update AI Settings | pass/fail |  |
| 8 | Create Activity | pass/fail |  |
| 9 | List Activities | pass/fail |  |
| 10 | Create Reminder | pass/fail |  |
| 11 | List Reminders | pass/fail |  |
| 12 | Create Mood | pass/fail |  |
| 13 | Create Memory | pass/fail |  |
| 14 | AI Runtime Fake Dry Run | pass/fail |  |
| 15 | Create Action Proposal | pass/fail |  |
| 16 | Confirm Action Proposal | pass/fail |  |
| 17 | Create Action Proposal To Reject | pass/fail |  |
| 18 | Reject Action Proposal | pass/fail |  |
| 19 | Export User Data | pass/fail |  |

## E2E Local Command

Comando:

```sh
go run ./cmd/tools/run-local-e2e -output-dir=docs
```

Resultado:

- [ ] pass
- [ ] fail
- [ ] blocked

Campos esperados:

| Campo | Valor |
| --- | --- |
| `proposal_only_preserved` | true/false |
| `no_automatic_ai_execution` | true/false |
| `explicit_execution_created_activity` | true/false |

## Errores Observados

| Hora | Paso | Error | Severidad | Accion |
| --- | --- | --- | --- | --- |
|  |  |  | low/medium/high/blocker |  |

## Seguridad

| Check | Resultado | Notas |
| --- | --- | --- |
| No se pegaron secretos en reportes | pass/fail |  |
| AI Runtime siguio proposal-only | pass/fail |  |
| FCM no envio push real | pass/fail |  |
| No se ejecutaron migraciones destructivas | pass/fail |  |
| No se ejecuto delete executor peligroso | pass/fail |  |

## Decision Final

- [ ] PASS
- [ ] FAIL
- [ ] BLOCKED

Resumen:

Pendientes antes del siguiente sprint:

