# Flutter Chat Integration

Fecha: 2026-07-15

## Estado

`ChatPage` ya usa el backend real de conversations.

Flujo productivo:

1. carga conversaciones activas;
2. toma la mas reciente;
3. si no hay conversaciones, crea una;
4. carga mensajes desde `GET /conversations/{id}/messages`;
5. envia por `POST /conversations/{id}/messages`;
6. muestra `user_message` y `assistant_message` devueltos por backend.

No se crean respuestas locales falsas.

## Archivos

- `lib/core/models/conversations/*`
- `lib/features/chat/domain/conversations_repository.dart`
- `lib/features/chat/data/conversations_repository_impl.dart`
- `lib/features/chat/presentation/cubit/chat_message_cubit.dart`
- `lib/features/chat/presentation/pages/chat_page.dart`

## Contrato

Endpoints usados:

- `POST /conversations`
- `GET /conversations?status=active&limit=1`
- `GET /conversations/{id}/messages`
- `POST /conversations/{id}/messages`
- `POST /conversations/{id}/archive`

`ApiClient` conserva el manejo de bearer token y base URL.

## UI

Se conserva la pantalla visual existente: shell neon, burbujas, input glass y cards de proposals.

Cambios:

- loading inicial;
- estado vacio;
- error con retry;
- indicador de envio;
- input real;
- avatar local sin dependencia de red;
- propuestas renderizadas como sugerencias, no ejecutables.

## Seguridad

Flutter no envia `user_id`.

Flutter no conoce provider fake/Gemini ni API keys.

Las `proposed_actions` se muestran como metadata defensiva. Confirm/execute queda fuera de V1.2.
