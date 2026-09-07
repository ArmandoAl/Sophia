# Memory Domain

Fecha: 2026-06-30

## Proposito

`memory` implementa memoria explicita y consultable para Sofia. En este sprint no hay embeddings, vector search, RAG completo ni agentes autonomos.

La memoria se considera informacion sensible del usuario. Todos los endpoints requieren JWT y el `user_id` siempre proviene del token.

## Entidad Memory

Campos:

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

## Tipos

MemoryType:

- `preference`
- `fact`
- `episodic`
- `semantic`
- `instruction`
- `relationship`
- `project`
- `system_note`

MemorySource:

- `user_manual`
- `conversation`
- `activity`
- `reflection`
- `imported`
- `ai_inferred`

MemoryVisibility:

- `private`
- `assistant_context`
- `archived`

MemoryStatus:

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

## Defaults

Al crear memory:

- `type`: `fact`
- `source`: `user_manual`
- `confidence`: `medium`
- `importance`: `medium`
- `visibility`: `assistant_context`
- `status`: `active`

## Validaciones

- `title` requerido.
- `content` requerido.
- `type` permitido.
- `source` permitido.
- `visibility` permitido.
- `status` permitido.
- `confidence` permitido.
- `importance` permitido.
- No se acepta `user_id` desde body.
- Campos desconocidos se rechazan por `DecodeJSON`.

## AISettings

Si `AISettings.memory_enabled=false`, el modulo bloquea:

- crear memory,
- listar memories,
- leer memory,
- actualizar memory,
- archivar memory,
- soft delete,
- buscar memory.

La respuesta HTTP actual es `409` con mensaje claro.

## Endpoints

Todos requieren JWT.

### POST /memory

Request:

```json
{
  "type": "preference",
  "title": "Prefers morning planning",
  "content": "The user prefers planning important work in the morning.",
  "summary": "Morning planning preference",
  "source": "user_manual",
  "source_id": "",
  "confidence": "high",
  "importance": "medium",
  "tags": ["planning", "routine"],
  "visibility": "assistant_context"
}
```

### GET /memory

Query params:

- `type`
- `tag`
- `importance`
- `status`
- `limit`

Por default lista `status=active`.

### GET /memory/{id}

Devuelve memory propia y actualiza `last_accessed_at`.

### PATCH /memory/{id}

Permite actualizar:

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

### POST /memory/{id}/archive

Marca `status=archived` y `visibility=archived`.

### DELETE /memory/{id}

Soft delete: marca `status=deleted`.

### POST /memory/search

Request:

```json
{
  "query": "planning",
  "type": "preference",
  "tag": "planning",
  "importance": "medium",
  "limit": 20
}
```

Busqueda basica:

- `query` busca en `title`, `content` y `summary`.
- `type` filtra por tipo.
- `tag` filtra por tag normalizado.
- `importance` filtra por importancia.
- Solo devuelve memories `active`.

No usa embeddings todavia.

## Relacion con Reflections

Sprint 6 no crea memories automaticamente desde reflections.

Pendiente:

- permitir `reflection_id` como origen validado,
- crear memory desde reflection propia,
- decidir si se copia `content`, `summary` o ambos,
- mantener trazabilidad con `source=reflection` y `source_id`.

## Embeddings y RAG

Se dejo preparada la interfaz:

- `internal/memory/application/embedding_provider.go`

No hay implementacion real todavia. La busqueda actual es textual y local al repositorio.

## Persistencia

Memory:

- `InMemoryMemoryRepository`
- `FirestoreMemoryRepository`

Firestore:

- Coleccion: `memories`

Campos principales:

- `user_id`
- `type`
- `source`
- `confidence`
- `importance`
- `tags`
- `visibility`
- `status`
- `created_at`
- `updated_at`
- `last_accessed_at`

## Privacidad

- Las memorias pueden contener preferencias, relaciones, instrucciones y hechos sensibles.
- Todos los accesos son por usuario autenticado.
- No hay acceso cruzado entre usuarios.
- No hay memoria global compartida.
- `DELETE /memory/{id}` es soft delete, no borrado fisico.

## Deuda tecnica

- Borrado fisico/retencion y exportacion de datos personales.
- Firestore indexes para filtros compuestos.
- Tests con Firestore emulator.
- Crear memory desde reflection.
- Embeddings y vector search.
- RAG/context assembly con presupuestos de tokens.
- Auditoria de acceso y cambios de memories sensibles.

## Riesgos

- Memory es de los dominios mas sensibles del producto.
- Soft delete no cumple por si solo con borrado definitivo.
- Busqueda textual puede no encontrar semantica relacionada.
- Futuro `ai_inferred` requiere explicabilidad y controles de aprobacion del usuario.
