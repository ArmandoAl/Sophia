# Sofia V1 Readiness Audit

Fecha: 2026-07-14

Scope ejecutado: auditoria read-only enfocada en `sophia_ai/`, contrastada contra el backend Go real en `sophia_ai_backend/sofia-backend/`. No se modifico codigo productivo.

## Veredicto

Sofia no esta lista para una V1 funcional de conversacion real.

La V1 minima viable mas cercana hoy es:

1. Backend Go local o deployado.
2. Flutter `sophia_ai` apuntando a `SOFIA_API_BASE_URL`.
3. Login/registro.
4. Sesion con bearer token.
5. Entrada a `/chat`.
6. Mensaje a Sofia por `POST /ai/runtime/message`.
7. Respuesta de modelo real con Gemini.
8. Persistencia en Firestore.
9. Capacidad adicional: reminders, porque backend tiene rutas, repositorios, tests y worker; el front solo tiene UI mock.

Brecha principal: el backend ya tiene runtime de IA, reminders y persistencia seleccionable, pero el front no consume chat/reminders reales. En `sophia_ai`, la integracion real cubre auth, sesion, `/users/me`, profile, AI settings, onboarding y health. Chat, reminders, smart home, dashboard legacy e insight siguen siendo prototipo local o UI historica.

## Evidencia De Frontend
Actúa como Senior Go Backend Engineer y Software Architect del proyecto Sofía.

Vamos a implementar el Sprint V1.1 — Conversation Persistence.

Antes de modificar código, revisa:

- docs/audits/v1_readiness_audit.md
- docs/architecture_state.md, si existe
- docs/flutter_endpoint_integration_contract.md
- docs/flutter_canonical_api_contract_v0_1.md, si existe
- docs/ai_runtime.md
- docs/ai_runtime_safety.md
- internal/ai/runtime
- internal/app
- internal/server/routes.go
- internal/database
- los repositorios Firestore e in-memory existentes

El objetivo es cerrar el blocker principal de Sofía V1:

Actualmente existe POST /ai/runtime/message, pero no existe persistencia canónica de conversaciones y mensajes ni recuperación de historial.

==================================================
OBJETIVO FUNCIONAL
==================================================

Implementar un bounded context vertical `conversations` que permita:

1. Crear conversaciones.
2. Listar conversaciones del usuario autenticado.
3. Obtener una conversación propia.
4. Listar sus mensajes en orden cronológico.
5. Enviar un mensaje a una conversación.
6. Persistir el mensaje del usuario.
7. Invocar el AI Runtime existente.
8. Persistir la respuesta de Sofía.
9. Devolver ambos mensajes al cliente.
10. Mantener el runtime proposal-only.
11. No ejecutar action proposals automáticamente.
12. Soportar repositories memory y Firestore.

La conversación debe servir como historial básico de V1.

==================================================
MODELO MÍNIMO
==================================================

Define como mínimo:

Conversation:
- id
- user_id
- title
- status: active | archived
- created_at
- updated_at
- last_message_at

ConversationMessage:
- id
- conversation_id
- user_id
- role: user | assistant | system
- content
- runtime_request_id opcional
- created_at

No agregues todavía:

- attachments
- audio
- imágenes
- branches
- embeddings
- vector search
- summaries automáticos
- RAG
- multimodal
- streaming
- token usage si el provider actual no lo expone

==================================================
ARQUITECTURA
==================================================

Mantén el patrón actual del backend:

internal/conversations/
  domain/
  application/
  infrastructure/
  interfaces/http/

Los use cases deben depender de interfaces de dominio, no de Firestore.

Implementa:

- InMemoryConversationRepository
- FirestoreConversationRepository
- InMemoryConversationMessageRepository
- FirestoreConversationMessageRepository

Puedes elegir entre:

A. Colecciones raíz:
- conversations
- conversation_messages

o

B. conversations/{conversation_id}/messages

Selecciona la opción que mejor encaje con los repositorios y consultas actuales.

Documenta la decisión.

Todas las consultas deben validar ownership por user_id.

==================================================
INTEGRACIÓN CON AI RUNTIME
==================================================

No dupliques Gemini ni el ModelClient.

Reutiliza el RuntimeService existente.

El nuevo flujo debe ser:

1. Validar JWT.
2. Validar ownership de conversation.
3. Persistir mensaje user.
4. Invocar AI Runtime.
5. Persistir assistant_message.
6. Retornar resultado.

Para esta V1:

- utiliza `dry_run=true` o una variante interna equivalente que no persista proposals automáticamente, salvo que el contrato actual indique otra cosa;
- nunca ejecutes tools automáticamente;
- no confirmes proposals;
- si el runtime devuelve proposed_actions, pueden devolverse como metadata, pero no deben ejecutarse.

Define el comportamiento en caso de fallo del provider:

- el mensaje del usuario puede conservarse;
- no se debe inventar una respuesta de Sofía;
- devolver error HTTP controlado;
- documentar si se persiste un mensaje assistant de error o no;
- preferir no persistir una respuesta falsa.

==================================================
ENDPOINTS
==================================================

Implementa como mínimo:

POST /conversations
GET /conversations
GET /conversations/{id}
GET /conversations/{id}/messages
POST /conversations/{id}/messages
POST /conversations/{id}/archive

Todos requieren JWT.

Request para crear conversación:

{
  "title": "optional"
}

Request para enviar mensaje:

{
  "content": "Hola Sofía"
}

Respuesta del envío:

{
  "conversation": {...},
  "user_message": {...},
  "assistant_message": {...},
  "proposed_actions": [],
  "runtime_request_id": "..."
}

Adapta wrappers a las convenciones reales del backend.

No aceptes user_id desde body.

Agrega `limit` y `cursor` únicamente si se pueden implementar siguiendo el patrón actual sin sobrecomplicar el sprint. Para mensajes, mantener orden cronológico consistente.

==================================================
PERSISTENCIA Y CONSISTENCIA
==================================================

Para Firestore:

- documenta índices esperados;
- evita que un usuario consulte conversaciones ajenas;
- asegura que message.conversation_id corresponda a una conversación propia;
- usa timestamps consistentes;
- si es razonable, utiliza transacción o batch para actualizar `last_message_at`;
- no introduzcas una abstracción de UnitOfWork genérica grande.

Para memory:

- comportamiento equivalente;
- seguro para tests;
- concurrency-safe.

==================================================
WIRING
==================================================

Conecta el módulo en:

- internal/app/repositories.go
- internal/app/modules.go
- internal/app/handlers.go
- internal/server/routes.go

Adapta las rutas reales según el estilo del proyecto.

No metas lógica de negocio en handlers.

==================================================
TESTS
==================================================

Agrega como mínimo:

1. Crear conversación autenticada.
2. Listar solo conversaciones propias.
3. No acceder conversación de otro usuario.
4. Crear mensaje user.
5. Invocar FakeModelClient.
6. Persistir respuesta assistant.
7. Recuperar mensajes en orden.
8. No ejecutar acciones automáticamente.
9. Provider error no crea respuesta assistant falsa.
10. Endpoint sin JWT retorna 401.
11. Firestore repository compila.
12. Regresión completa con `go test ./...`.

No llames Gemini real en tests.

==================================================
DOCUMENTACIÓN
==================================================

Genera:

- docs/conversations_domain.md
- docs/conversations_api.md
- docs/sprint_v1_1_report.md

Actualiza `docs/architecture_state.md` si existe y está vigente.

El reporte debe incluir:

- auditoría inicial;
- decisiones;
- archivos creados/modificados;
- endpoints;
- modelos;
- persistencia;
- interacción con AI Runtime;
- errores;
- pruebas;
- deuda técnica;
- riesgos;
- resultado de `go test ./...`.

==================================================
RESTRICCIONES
==================================================

- No modificar Flutter.
- No agregar RAG.
- No agregar embeddings.
- No integrar hardware.
- No activar autonomía.
- No ejecutar Gemini real.
- No ejecutar migraciones.
- No modificar colecciones legacy.
- No guardar API keys.
- No realizar commits.
- No cambiar el comportamiento proposal-only.
- No tocar credenciales salvo documentar el riesgo.

==================================================
ENTREGA FINAL
==================================================

Al terminar responde con:

1. Estado del sprint.
2. Endpoints agregados.
3. Colecciones agregadas.
4. Cantidad de tests agregados.
5. Resultado de go test ./...
6. Riesgos pendientes.
7. Ruta de los documentos generados.

Root Flutter real: `sophia_ai/`

Entry point:

- `lib/main.dart`
- `lib/core/di/service_locator.dart`
- `lib/core/router/app_router.dart`

Integracion backend real:

- `lib/core/config/app_config.dart`: default `http://localhost:8080`, override por `SOFIA_API_BASE_URL`.
- `lib/core/network/api_client.dart`: cliente HTTP con JSON, bearer token y errores.
- `lib/core/storage/token_storage.dart`: almacenamiento de token.
- `lib/features/auth/data/auth_repository_impl.dart`: consume `/auth/register` y `/auth/login`.
- `lib/features/users/data/user_repository_impl.dart`: consume `/users/me`, `/users/me/profile`, `/users/me/ai-settings`, `/users/me/onboarding/complete`.
- `lib/features/system/data/system_repository_impl.dart`: health backend.

Estado de navegacion:

- `AppRouter` inicia en `/splash`.
- Auth redirect real usa `SessionCubit`.
- Rutas principales: `/login`, `/register`, `/onboarding`, `/profile`, `/assistant-settings`, `/chat`, `/smart-home`, `/dashboard`, `/settings`.

Partes no backend-backed:

- `lib/features/chat/presentation/cubit/chat_message_cubit.dart`: carga mensajes iniciales hardcoded y simula respuesta con `Future.delayed`.
- `lib/features/reminders/presentation/cubit/reminders_cubit.dart`: emite tareas hardcoded.
- `smart_home`, dashboard legacy, settings legacy e insight contienen UI/prototipo sin contrato backend V1.

Tests frontend existentes:

- `test/core/network/api_client_test.dart`
- `test/core/models/canonical_models_test.dart`
- `test/features/repositories_f1_test.dart`
- `test/features/cubits_f1_test.dart`
- `test/features/f1_ui_navigation_test.dart`

## Evidencia De Backend

Root backend real: `sophia_ai_backend/sofia-backend/`

Entry point:

- `cmd/sofia/main.go`
- `internal/app/container.go`
- `internal/app/repositories.go`
- `internal/app/modules.go`
- `internal/server/routes.go`

Rutas reales relevantes:

- `GET /health`
- `POST /auth/register`
- `POST /auth/login`
- `GET /users/me`
- `PATCH /users/me/profile`
- `GET/PATCH /users/me/ai-settings`
- `POST /users/me/onboarding/complete`
- `/activities`, `/activities/`
- `/reminders`, `/reminders/due`, `/reminders/`
- `/memory`, `/memory/search`, `/memory/`
- `/tools`, `/tools/`
- `/ai/action-proposals`, `/ai/action-proposals/`
- `POST /ai/runtime/message`

Persistencia real:

- `PERSISTENCE_DRIVER=memory` por defecto.
- `PERSISTENCE_DRIVER=firestore` activa repositorios Firestore.
- Produccion exige Firestore.
- Se encontro `sophia_ai_backend/sophia-backend-key.json` fuera del modulo auditado; no se imprimio contenido.

Runtime de IA:

- `AI_MODEL_PROVIDER=fake` por defecto.
- `AI_MODEL_PROVIDER=gemini` requiere `GEMINI_API_KEY`.
- `AI_RUNTIME_PROPOSAL_ONLY=false` no esta soportado.
- `POST /ai/runtime/message` existe, construye contexto, llama provider, valida propuestas y puede persistir action proposals si no es dry-run.
- No hay evidencia de persistencia de conversaciones como hilo/historial canonico. El runtime persiste auditoria/propuestas, no una coleccion de conversaciones consumible por el front.

Capacidad adicional mas cercana:

- `reminders` esta mas cerca que chat desde el punto de vista backend: rutas, repositorios memory/firestore, tests, worker y entrega opcional noop/FCM.
- La brecha esta en Flutter: no hay models/repos/cubit backend-backed para reminders; el cubit actual es hardcoded.

## Checks Ejecutados

Frontend:

- `fvm flutter analyze`
- Resultado: falla por 3 infos de lint/deprecacion en `lib/core/widgets/neon_button.dart` (`Color.red`, `Color.green`, `Color.blue`). No hay errores.

Frontend tests:

- `fvm flutter test`
- Resultado: pasan 82 tests.

Backend:

- `env GOCACHE=.../.cache/go-build go test ./...`
- Resultado: pasa.

Nota: `flutter analyze` y `flutter test` resolvieron/descargaron paquetes antes de ejecutar. No se instalaron dependencias nuevas manualmente.

## Gap V1 Minima

| Requisito V1 | Estado real | Brecha |
|---|---|---|
| Levantar backend | Parcialmente listo | Configurar env real, Firestore y secretos fuera de repo |
| Levantar Flutter | Listo con warnings | Limpiar 3 infos si se quiere analyze verde |
| Registrar/login | Backend + front conectados | Registro no deja sesion; login posterior requerido |
| Mantener sesion | Implementado | Validar storage por plataforma |
| Entrar a app | Implementado | Flujo depende de `/users/me` y onboarding |
| Abrir conversacion | UI existe | No esta conectada al backend |
| Enviar mensaje | UI simula | Falta repository/cubit a `/ai/runtime/message` |
| API IA real | Backend soporta Gemini | Falta env `AI_MODEL_PROVIDER=gemini`, `GEMINI_API_KEY`, credito y prueba real |
| Recibir respuesta | Backend soporta | Front no parsea respuesta runtime |
| Persistir conversacion | No evidenciado | Falta contrato/coleccion/historial de conversaciones |
| Recuperar historial | No evidenciado | Falta endpoint y consumer |
| Capacidad adicional | Backend reminders fuerte | Front reminders hardcoded |

## Camino Minimo A V1

1. Front chat: reemplazar `ChatMessageCubit` mock por repository minimo que llame `POST /ai/runtime/message`.
2. Definir contrato Flutter para `RuntimeResponse` solo con campos usados: `assistant_message`, `proposed_actions`, `request_id`, `mode`.
3. Backend: confirmar si V1 acepta "historial basico" como nueva persistencia de mensajes o si se agrega endpoint de conversaciones. Hoy no esta cubierto.
4. Backend env: correr con `PERSISTENCE_DRIVER=firestore`, `AI_MODEL_PROVIDER=gemini`, `GEMINI_API_KEY`, `FIRESTORE_PROJECT_ID`, JWT seguro.
5. Front reminders: agregar repository minimo para `GET/POST /reminders` y conectar `RemindersCubit`.
6. E2E local: login, `/users/me`, chat runtime con Gemini, crear/listar reminder.

## Clasificacion

IMPLEMENTED_AND_CONNECTED:

- Auth HTTP client.
- Login token storage.
- Session bootstrap.
- `/users/me`.
- Profile.
- AI settings.
- Onboarding complete.
- Health diagnostics.

IMPLEMENTED_BACKEND_ONLY:

- Activities.
- Reminders.
- Memory.
- Tools.
- AI action proposals.
- AI runtime message.
- Privacy export/delete request.
- Notification device tokens.

FRONTEND_MOCK_ONLY:

- Chat messages and assistant responses.
- Reminders task list.
- Smart home.
- Rich cards for maps/schedule conflicts/action proposals.

DOCUMENTED_OR_HISTORICAL_ONLY_FOR_V1:

- Conversation persistence and history recovery.
- Full autonomous execution.
- Smart home/presence/hardware.
- Advanced RAG/embeddings.
- Public production push delivery.

## Decision

La V1 mas barata no debe intentar "todo Sofia". Debe ser:

- auth + session existente;
- chat real via `/ai/runtime/message`;
- Gemini configurado;
- Firestore activo;
- reminders como capacidad adicional;
- historial minimo definido y probado.

Sin conectar chat y sin una historia persistida recuperable, el proyecto aun no cumple la V1 descrita.
