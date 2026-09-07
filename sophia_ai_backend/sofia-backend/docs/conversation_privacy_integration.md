# Conversation Privacy Integration

`GET /users/me/export` incluye `conversations` y `conversation_messages` del namespace V1, filtrados por el `user_id` autenticado. No incluye documentos legacy ni datos de otros usuarios.

El delete executor borra `v1_conversations` y `v1_conversation_messages` para requests aprobadas, reportando conteos por colección. No toca `conversations`, su subcolección `messages`, `audit_logs`, auth users ni hardware.

El contenido conversacional puede ser sensible. El audit log registra conteos, nunca contenido.
