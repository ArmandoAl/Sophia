# Sprint V1.1 Report - Conversation Persistence

Fecha: 2026-07-14

## Implementado

- Nuevo bounded context `internal/conversations`.
- Repositorios in-memory y Firestore para conversaciones y mensajes.
- Endpoints:
  - `POST /conversations`
  - `GET /conversations`
  - `GET /conversations/{id}`
  - `GET /conversations/{id}/messages`
  - `POST /conversations/{id}/messages`
  - `POST /conversations/{id}/archive`
- Wiring en `internal/app` y `internal/server/routes.go`.
- Integracion con `RuntimeService` existente.
- Envio de mensaje usa `dry_run=true` para mantener proposal-only.
- Si el runtime falla, no se crea assistant falsa.

## Tests

Agregados en `internal/server/routes_test.go`:

- crear conversacion autenticada;
- listar solo conversaciones propias;
- bloquear acceso cross-user;
- enviar mensaje y persistir user + assistant;
- recuperar mensajes en orden cronologico;
- mantener dry-run sin ejecucion automatica;
- fallo de provider conserva solo mensaje user;
- endpoint sin JWT retorna `401`.

Validacion:

```sh
env GOCACHE=/Users/armandoalvarado/Documents/Sophia/sophia_ai_backend/sofia-backend/.cache/go-build go test ./...
```

Resultado: pasa.

## Decisiones

- Colecciones raiz V1 `v1_conversations` y `v1_conversation_messages` (actualizadas en Sprint V1.1.5 para evitar la colección legacy `conversations`).
- No se agrego UnitOfWork generico.
- No se agrego streaming ni token usage.
- No se llamo Gemini real en tests.

## Pendiente Frontend

- Agregar models/repository/cubit para conversations en `sophia_ai`.
- Reemplazar `ChatMessageCubit` mock por `POST /conversations/{id}/messages`.
- Cargar historial desde `GET /conversations/{id}/messages`.
