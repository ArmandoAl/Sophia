# Sprint 5 Report - Mood, Outcomes & Reflections

Fecha: 2026-06-30

## Auditoria inicial

Se revisaron:

- `docs/architecture_state.md`
- `docs/sprint_4_report.md`
- `docs/activities_domain.md`
- `docs/users_domain.md`
- Estado real de `cmd/sofia`, `internal/server`, `internal/activities` y `internal/reminders`

Estado real antes del sprint:

- Auth con JWT estaba operativo.
- Users tenia profile y AI settings.
- Activities era el dominio superior de planificacion.
- Reminder Engine existia con memory/Firestore y endpoint due.
- No existia modulo `internal/insights`.
- No habia tracking de mood, outcomes ni reflections.

Riesgos detectados:

- Insights registra datos personales sensibles aunque no sean clinicos.
- Firestore puede requerir indices para filtros por usuario, activity y rango.
- Summary debe mantenerse estadistico/simple para no aparentar diagnostico.

Decisiones:

- Crear `internal/insights` como modulo vertical.
- Separar entidades `MoodEntry`, `ActivityOutcome` y `Reflection`.
- Usar repositorios separados para cada entidad.
- Validar ownership de `activity_id` contra `ActivityRepository`.
- Mantener summary como promedios y conteos simples.
- No implementar Memory, IA ni analisis psicologico clinico.

## Que cambio

- Se creo `internal/insights/domain`.
- Se creo `internal/insights/application`.
- Se creo `internal/insights/infrastructure`.
- Se creo `internal/insights/interfaces/http`.
- Se agregaron entidades:
  - `MoodEntry`
  - `ActivityOutcome`
  - `Reflection`
- Se agregaron contratos:
  - `MoodEntryRepository`
  - `ActivityOutcomeRepository`
  - `ReflectionRepository`
- Se implementaron repositorios memory y Firestore.
- Se agregaron use cases:
  - `CreateMoodEntry`
  - `ListMoodEntries`
  - `CreateActivityOutcome`
  - `ListActivityOutcomes`
  - `CreateReflection`
  - `ListReflections`
  - `GetInsightsSummary`
- Se conecto Insights en `cmd/sofia/main.go`.
- Se agregaron rutas en `internal/server/routes.go`.
- Se agregaron tests HTTP end-to-end.

## Archivos modificados

- `sophia_ai_backend/sofia-backend/cmd/sofia/main.go`
- `sophia_ai_backend/sofia-backend/internal/server/routes.go`
- `sophia_ai_backend/sofia-backend/internal/server/routes_test.go`
- `sophia_ai_backend/sofia-backend/internal/insights/domain/insights.go`
- `sophia_ai_backend/sofia-backend/internal/insights/application/usecases.go`
- `sophia_ai_backend/sofia-backend/internal/insights/infrastructure/memory_repositories.go`
- `sophia_ai_backend/sofia-backend/internal/insights/infrastructure/firestore_repositories.go`
- `sophia_ai_backend/sofia-backend/internal/insights/interfaces/http/handler.go`
- `sophia_ai_backend/sofia-backend/internal/insights/interfaces/http/requests.go`
- `sophia_ai_backend/sofia-backend/internal/insights/interfaces/http/responses.go`
- `docs/architecture_state.md`
- `docs/insights_domain.md`
- `docs/sprint_5_report.md`

## Endpoints

| Metodo | Ruta | Auth | Descripcion |
| --- | --- | --- | --- |
| `POST` | `/insights/mood` | Si | Crea mood entry. |
| `GET` | `/insights/mood` | Si | Lista mood entries propias. |
| `POST` | `/insights/outcomes` | Si | Crea outcome asociado a Activity propia. |
| `GET` | `/insights/outcomes` | Si | Lista outcomes propios. |
| `POST` | `/insights/reflections` | Si | Crea reflection. |
| `GET` | `/insights/reflections` | Si | Lista reflections propias. |
| `GET` | `/insights/summary` | Si | Devuelve promedios energy/stress/focus y conteos. |

## Validaciones

- `mood` debe ser permitido.
- `result` debe ser permitido.
- Niveles numericos deben estar entre 1 y 5.
- `occurred_at` requerido para mood/reflection y debe venir en RFC3339.
- `timezone` requerido para mood.
- `content` requerido para reflection.
- `activity_id` requerido para outcome.
- Si `activity_id` existe, debe pertenecer al usuario autenticado.
- No se acepta `user_id` desde body.
- Campos desconocidos se rechazan por `DecodeJSON`.

## Privacidad

- Los endpoints requieren JWT.
- El usuario solo puede listar y crear datos propios.
- No hay diagnostico, perfil psicologico ni interpretacion clinica.
- Los datos se tratan como tracking personal/productivo.

## Pruebas ejecutadas

Comando:

```sh
env GOCACHE=/Users/armandoalvarado/Documents/Sophia/sophia_ai_backend/sofia-backend/.cache/go-build go test ./...
```

Resultado:

- Suite completa pasando.

Casos agregados/cubiertos:

- Crear mood.
- Rechazar mood invalido.
- Rechazar niveles fuera de rango.
- Crear outcome asociado a activity propia.
- Impedir outcome sobre activity de otro usuario.
- Crear reflection.
- Listar solo datos propios.
- Summary basico.
- Endpoints sin auth devuelven `401`.
- Regresion de auth/users/activities/reminders/health/CORS.

## Deuda tecnica

- Agregar paginacion por cursor.
- Agregar tests contra Firestore emulator.
- Definir indices Firestore para filtros compuestos.
- Validar timezone contra base IANA.
- Evaluar `occurred_at` para `ActivityOutcome`.
- Agregar export/delete de datos de insights por privacidad.
- Definir estrategia de retencion de datos.

## Riesgos

- Los datos de mood/reflections pueden ser sensibles.
- Summary simple puede ser malinterpretado como insight inteligente si UI no lo presenta bien.
- Sin paginacion real, listados grandes dependeran solo de `limit`.
- Firestore puede requerir indices al combinar filtros en produccion.

## Recomendaciones Sprint 6

- Implementar Memory solo despues de definir politicas de privacidad.
- Agregar export/delete de datos personales antes de usar insights para IA.
- Agregar Firestore emulator tests para insights.
- Definir si outcomes deben crear eventos historicos de productividad.
