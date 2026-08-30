# Conversations Domain

Fecha: 2026-07-14

## Objetivo

`internal/conversations` agrega el historial canonico minimo para Sofia V1.1.

Permite crear conversaciones, listar conversaciones propias, obtener una conversacion, listar mensajes cronologicos y enviar un mensaje que persiste:

1. mensaje del usuario;
2. llamada al AI Runtime existente en `dry_run=true`;
3. mensaje assistant real devuelto por el runtime.

No ejecuta tools ni confirma action proposals.

## Modelo

`Conversation`

- `id`
- `user_id`
- `title`
- `status`: `active | archived`
- `created_at`
- `updated_at`
- `last_message_at`

`ConversationMessage`

- `id`
- `conversation_id`
- `user_id`
- `role`: `user | assistant | system`
- `content`
- `runtime_request_id`
- `created_at`

## Persistencia

Desde Sprint V1.1.5 las colecciones V1 son `v1_conversations` y `v1_conversation_messages`. La colección legacy `conversations` y su subcolección `messages` quedan fuera del runtime.

Se eligieron colecciones raiz:

- `v1_conversations`
- `v1_conversation_messages`

Razon: encaja con los repositorios existentes del backend, que usan colecciones raiz por bounded context (`activities`, `reminders`, `memories`, `ai_action_proposals`). Tambien simplifica queries por `user_id` y evita acoplar el repositorio de mensajes a paths anidados.

## Ownership

Todos los reads/writes reciben `user_id` desde JWT/context. El body nunca acepta `user_id`.

`FindByID` devuelve `nil` cuando el documento no existe o pertenece a otro usuario; el service lo traduce a `conversation not found`.

## Fallo Del Provider

Si el AI Runtime falla:

- el mensaje del usuario ya persistido se conserva;
- no se persiste ningun mensaje assistant falso;
- el handler devuelve error HTTP controlado;
- el cliente puede mostrar estado de retry sobre el ultimo mensaje user.

## Indices Firestore Esperados

- `v1_conversations`: `user_id ASC, updated_at DESC, id DESC`
- `v1_conversations`: `user_id ASC, status ASC, updated_at DESC, id DESC`
- `v1_conversation_messages`: `user_id ASC, conversation_id ASC, created_at ASC, id ASC`

## Fuera De Scope

- streaming;
- attachments/audio/imagenes;
- branches;
- embeddings/vector search;
- summaries automaticos;
- token usage/cost;
- ejecucion automatica de tools.
