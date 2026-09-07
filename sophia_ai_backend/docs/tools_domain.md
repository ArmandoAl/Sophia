# Tools Domain

Fecha: 2026-06-30

## Proposito

`tools` registra las herramientas que Sofia puede proponer usar. Sprint 7 no integra LLM real ni llamadas externas; solo define el registry consultable para tool calling seguro.

## ToolDefinition

Campos:

- `id`
- `name`
- `description`
- `category`
- `input_schema`
- `output_schema`
- `requires_confirmation`
- `enabled`
- `min_autonomy_level`
- `created_at`
- `updated_at`

## Tools iniciales

- `create_activity`
- `update_activity`
- `complete_activity`
- `create_reminder`
- `cancel_reminder`
- `create_memory`
- `search_memory`
- `create_reflection`

## Endpoints

Todos requieren JWT.

### GET /tools

Lista tools registradas.

### GET /tools/{name}

Devuelve una tool por nombre.

## Seguridad

- `enabled=false` bloquea uso por action proposals.
- `requires_confirmation` determina si una proposal debe confirmarse antes de ejecutar.
- `min_autonomy_level` deja preparada la politica de autonomia.
- En Sprint 7 ninguna tool se ejecuta automaticamente.

## Persistencia

- Memory: `InMemoryToolDefinitionRepository`
- Firestore: `FirestoreToolDefinitionRepository`
- Coleccion: `tool_definitions`

## Deuda tecnica

- Validacion real contra JSON Schema.
- Versionado de tools.
- Separar tools internas de tools externas.
- Permisos por rol.
- Auditoria de cambios al registry.
