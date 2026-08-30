# Legacy Collections

Fecha: 2026-07-01

## Resumen

Estas colecciones existen en Firestore real pero no son el contrato primario del backend actual.

## Colecciones

### `ai_memory`

Clasificacion: `migrate_later`

Contiene memoria inferida/conceptual con `statement`, `type`, `confidence`, refs a body/message y counters. Candidato futuro para `memories`, pero requiere privacy review, deduplicacion y reglas de visibility.

### `ai_actions_log`

Clasificacion: `migrate_later`

Contiene acciones AI previas con `action_type`, `approved`, `executed`, `reasoning`, `target_refs`. Candidato para audit log o historico de action proposals, no para ejecucion directa.

### `ai_policy`

Clasificacion: `migrate_later`

Contiene politicas conceptuales por action type. Candidato para policy engine futuro.

### `bodies`

Clasificacion: `legacy`

Parece representar cuerpos/dispositivos/contexto fisico con subcoleccion `extensions`. Fuera del runtime backend actual. No migrar todavia.

### `conversations`

Clasificacion: `migrate_later`

Contiene sesiones y subcoleccion `messages`. Candidato para modulo conversations/RAG futuro.

### `daily_summaries`

Clasificacion: `ignore_for_now`

Podria alimentar insights o memory, pero no hay entidad destino estable.

### `locations`

Clasificacion: `legacy`

Parece representar ubicaciones fisicas/geohash/privacy. Relacionado con hardware/location context futuro. No integrar ahora.

### `message_embeddings`

Clasificacion: `migrate_later`

Contiene vectores ligados a mensajes. Candidato para RAG futuro cuando exista estrategia de embeddings.

## Politica

- No borrar.
- No leer desde runtime actual.
- No exponer en endpoints actuales.
- No migrar a memory sin privacy/export/delete.
- No integrar hardware/location hasta que exista modulo dedicado y consentimiento explicito.
