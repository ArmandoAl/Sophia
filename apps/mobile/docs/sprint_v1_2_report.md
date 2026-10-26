# Sprint V1.2 Report - Flutter Real Chat

Fecha: 2026-07-15

## Estado Inicial

El chat Flutter usaba mensajes hardcodeados, `Future.delayed` y respuestas locales por keywords. El input glass visible no enviaba mensajes.

## Implementado

- Modelos manuales:
  - `Conversation`
  - `ConversationMessage`
  - `CreateConversationRequest`
  - `SendConversationMessageRequest`
  - `ConversationListResponse`
  - `ConversationMessagesResponse`
  - `SendConversationMessageResponse`
  - `RuntimeProposedAction`
- `ConversationsRepository` + `ConversationsRepositoryImpl`.
- DI en `service_locator.dart`.
- `ChatMessageCubit` real:
  - carga o crea conversacion activa;
  - carga historial persistido;
  - envia mensajes al backend;
  - agrega solo mensajes devueltos por backend;
  - maneja errores sin respuesta mock.
- `ChatPage` adaptada:
  - loading;
  - empty state;
  - error/retry;
  - sending indicator;
  - input real;
  - avatar local.

## Limpieza Del Mock

El path productivo ya no contiene:

- mensajes iniciales hardcodeados;
- respuestas por keyword;
- `Future.delayed` para simular IA.

Los widgets ricos existentes se conservaron para render defensivo.

## Tests

Agregado `test/features/chat_real_test.dart`:

- parseo de modelos;
- repository create/list/send;
- cubit crea conversacion si no existe;
- cubit carga historial;
- cubit envia y agrega respuesta backend;
- cubit maneja error de provider;
- widget muestra historial;
- widget muestra error.

Validacion:

```sh
fvm flutter analyze
fvm flutter test
```

Resultado: ambos pasan.

## Pendiente

- UI para elegir/listar conversaciones antiguas.
- Retry de envio puntual por mensaje.
- Confirm/reject/execute de action proposals, con flujo explicito y seguro.
