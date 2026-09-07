# Memory Domain

Fecha: 2026-07-03

## Estado Actual

`internal/memory` implementa memoria explicita, textual y consultable por usuario. No implementa embeddings, vector search ni RAG completo.

La memoria esta protegida por JWT, ownership por `user_id` desde contexto y `AISettings.memory_enabled`. Si `memory_enabled=false`, crear, listar, consultar o buscar memorias devuelve error.

## Entidad

`Memory` contiene:

- `id`
- `user_id`
- `type`
- `title`
- `content`
- `summary`
- `source`
- `source_id`
- `confidence`
- `importance`
- `tags`
- `visibility`
- `status`
- `created_at`
- `updated_at`
- `last_accessed_at`

## Enums

Types:

- `preference`
- `fact`
- `episodic`
- `semantic`
- `instruction`
- `relationship`
- `project`
- `system_note`

Sources:

- `user_manual`
- `conversation`
- `activity`
- `reflection`
- `imported`
- `ai_inferred`

Visibility:

- `private`
- `assistant_context`
- `archived`

Status:

- `active`
- `archived`
- `deleted`

Confidence:

- `low`
- `medium`
- `high`

Importance:

- `low`
- `medium`
- `high`
- `critical`

## Endpoints

Todos requieren JWT.

| Metodo | Ruta | Descripcion |
| --- | --- | --- |
| `POST` | `/memory` | Crea una memoria explicita. |
| `GET` | `/memory` | Lista memorias propias con filtros basicos. |
| `GET` | `/memory/{id}` | Obtiene una memoria propia y actualiza `last_accessed_at`. |
| `PATCH` | `/memory/{id}` | Actualiza campos permitidos. |
| `POST` | `/memory/{id}/archive` | Archiva una memoria. |
| `DELETE` | `/memory/{id}` | Hace soft delete. |
| `POST` | `/memory/search` | Busca por texto/filtros basicos. |

## Filtros y Busqueda

`GET /memory` soporta:

- `type`
- `tag`
- `importance`
- `status`
- `limit`
- `cursor`

`POST /memory/search` soporta:

- `query`
- `type`
- `tag`
- `importance`
- `limit`

La busqueda actual es textual simple sobre `title`, `content` y `summary`. Las memorias con `status=deleted` no aparecen en listados ni busquedas.

## Privacidad

- `user_id` nunca se acepta desde el body.
- `deleted` es soft delete.
- `visibility=assistant_context` es la unica visibilidad que el ContextBuilder debe considerar para contexto del modelo.
- El AI Runtime no debe enviar contenido completo sensible; usa resumen/titulo/metadatos limitados.

## Pendiente

- Embeddings.
- Vector search.
- RAG completo.
- Importacion segura de `ai_memory` legacy.
- Export/delete definitivo de credenciales auth asociado a privacidad completa.
