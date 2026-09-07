# Sprint 6 Report - Memory Foundation

Fecha: 2026-06-30

## Auditoria inicial

Se revisaron:

- `docs/architecture_state.md`
- `docs/sprint_5_report.md`
- `docs/users_domain.md`
- `docs/activities_domain.md`
- `docs/insights_domain.md`
- Estado real de `cmd/sofia`, `internal/server`, `internal/users` e `internal/insights`

Estado real antes del sprint:

- Auth con JWT estaba operativo.
- Users tenia AISettings con `memory_enabled`.
- Activities, reminders e insights existian como modulos verticales.
- No existia modulo `internal/memory`.
- No habia memoria explicita, busqueda ni soft delete de memorias.

Riesgos detectados:

- Memory puede contener datos altamente sensibles.
- El backend aun no tiene export/delete definitivo de datos personales.
- RAG/embeddings prematuros agregarian complejidad y riesgos de privacidad.

Decisiones:

- Crear `internal/memory` como modulo vertical.
- Implementar memoria explicita y consultable sin embeddings.
- Crear interfaz `EmbeddingProvider` sin implementacion real.
- Respetar `AISettings.memory_enabled` como gate de todo el modulo.
- Usar soft delete (`status=deleted`) para `DELETE`.
- Mantener busqueda basica por texto y filtros.

## Que cambio

- Se creo `internal/memory/domain`.
- Se creo `internal/memory/application`.
- Se creo `internal/memory/infrastructure`.
- Se creo `internal/memory/interfaces/http`.
- Se agrego entidad `Memory`.
- Se agrego contrato `MemoryRepository`.
- Se implementaron:
  - `InMemoryMemoryRepository`
  - `FirestoreMemoryRepository`
- Se agregaron use cases:
  - `CreateMemory`
  - `GetMemory`
  - `ListMemories`
  - `UpdateMemory`
  - `ArchiveMemory`
  - `DeleteMemory`
  - `SearchMemory`
- Se agrego `EmbeddingProvider` como interfaz futura.
- Se conecto Memory en `cmd/sofia/main.go`.
- Se agregaron rutas en `internal/server/routes.go`.
- Se agregaron tests HTTP end-to-end.

## Archivos modificados

- `sophia_ai_backend/sofia-backend/cmd/sofia/main.go`
- `sophia_ai_backend/sofia-backend/internal/server/routes.go`
- `sophia_ai_backend/sofia-backend/internal/server/routes_test.go`
- `sophia_ai_backend/sofia-backend/internal/memory/domain/memory.go`
- `sophia_ai_backend/sofia-backend/internal/memory/application/usecases.go`
- `sophia_ai_backend/sofia-backend/internal/memory/application/embedding_provider.go`
- `sophia_ai_backend/sofia-backend/internal/memory/infrastructure/memory_repository.go`
- `sophia_ai_backend/sofia-backend/internal/memory/infrastructure/firestore_repository.go`
- `sophia_ai_backend/sofia-backend/internal/memory/interfaces/http/handler.go`
- `sophia_ai_backend/sofia-backend/internal/memory/interfaces/http/requests.go`
- `sophia_ai_backend/sofia-backend/internal/memory/interfaces/http/responses.go`
- `docs/architecture_state.md`
- `docs/memory_domain.md`
- `docs/sprint_6_report.md`

## Endpoints

| Metodo | Ruta | Auth | Descripcion |
| --- | --- | --- | --- |
| `POST` | `/memory` | Si | Crea memory propia. |
| `GET` | `/memory` | Si | Lista memories propias. |
| `GET` | `/memory/{id}` | Si | Lee memory propia y actualiza `last_accessed_at`. |
| `PATCH` | `/memory/{id}` | Si | Actualiza memory propia. |
| `POST` | `/memory/{id}/archive` | Si | Archiva memory propia. |
| `DELETE` | `/memory/{id}` | Si | Soft delete de memory propia. |
| `POST` | `/memory/search` | Si | Busqueda basica por texto/filtros. |

## Busqueda

`POST /memory/search` soporta:

- `query`: texto en `title`, `content`, `summary`.
- `type`
- `tag`
- `importance`
- `limit`

No usa embeddings todavia.

## AISettings

Si `memory_enabled=false`, se bloquea:

- create,
- list,
- get,
- update,
- archive,
- delete,
- search.

Respuesta HTTP: `409`.

## Pruebas ejecutadas

Comando:

```sh
env GOCACHE=/Users/armandoalvarado/Documents/Sophia/sophia_ai_backend/sofia-backend/.cache/go-build go test ./...
```

Resultado:

- Suite completa pasando.

Casos agregados/cubiertos:

- Crear memory.
- Bloquear memory si `memory_enabled=false`.
- Listar solo memorias propias.
- Get memory propia.
- Impedir acceso cruzado.
- Actualizar memory.
- Archive.
- Delete soft.
- Search basico por texto/tag/type.
- Endpoints sin auth devuelven `401`.
- Regresion de auth/users/activities/reminders/insights/health/CORS.

## Integracion con Reflections

No se implemento creacion automatica desde `reflection_id` en Sprint 6.

Queda documentado como pendiente para evitar acoplar Memory con Insights sin definir:

- flujo de consentimiento,
- formato de resumen,
- trazabilidad,
- politica de edicion/borrado cruzado.

## Deuda tecnica

- Crear memory desde reflection propia.
- Implementar embeddings y vector search.
- Implementar RAG/context assembly.
- Agregar tests contra Firestore emulator.
- Definir indices Firestore para filtros compuestos.
- Agregar exportacion y borrado fisico de datos personales.
- Auditar accesos a memorias sensibles.
- Definir aprobacion de `ai_inferred`.

## Riesgos

- Memory puede almacenar informacion muy sensible.
- Soft delete no equivale a cumplimiento de borrado definitivo.
- Search textual no cubre semantica.
- Futuro RAG puede exponer contexto excesivo si no hay controles por visibility/importancia.

## Recomendaciones Sprint 7

- Privacy controls: export/delete definitivo.
- Crear memories desde reflections con consentimiento explicito.
- Preparar indices Firestore.
- Disenar context assembly antes de meter embeddings.
