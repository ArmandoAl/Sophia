# Conversation Firestore Namespace

Sprint V1.1.5 confirma una colision: la colección legacy `conversations` existe con `body_id`, `session_id`, `timestamp`, `user_id` como referencia y subcolección `messages`.

El bounded context V1 usa exclusivamente:

- `v1_conversations`
- `v1_conversation_messages`

Los nombres están centralizados en `internal/database/collections.go`. Las colecciones legacy no se leen, escriben, migran ni borran. Ownership, orden cronológico y cursores permanecen iguales.
