# Functionality Matrix

Fecha: 2026-06-30

## Resumen ejecutivo

| Modulo | Entidades | Endpoints | Repositorios | Casos de uso | Estado | Dependencias externas |
| --- | --- | --- | --- | --- | --- | --- |
| `auth` | `User` | `POST /auth/register`, `POST /auth/login` | Memory, Firestore | Register, Login | usable | bcrypt, uuid, JWT secret, Firestore opcional |
| `users` | `UserProfile`, `AISettings` | `/users/me`, `/users/me/profile`, `/users/me/ai-settings`, `/users/me/onboarding/complete` | Memory, Firestore | GetMe, UpdateProfile, Get/UpdateAISettings, CompleteOnboarding | usable | Firestore opcional |
| `activities` | `Activity` | `/activities`, `/activities/{id}`, `/complete`, `/cancel`, `/archive` | Memory, Firestore | Create, Get, List, Update, Complete, Cancel, Archive | usable | Firestore opcional |
| `reminders` | `Reminder` | `/reminders`, `/reminders/due`, `/reminders/{id}`, `/cancel`, `/archive` | Memory, Firestore | Create, Get, List, Update, Cancel, Archive, ListDue, MarkDelivered, MarkFailed | parcial | Firestore opcional; worker/push pendientes |
| `insights` | `MoodEntry`, `ActivityOutcome`, `Reflection` | `/insights/mood`, `/insights/outcomes`, `/insights/reflections`, `/insights/summary` | Memory, Firestore | Create/List mood, outcome, reflection; summary | usable | Firestore opcional |
| `memory` | `Memory` | `/memory`, `/memory/{id}`, `/memory/search`, `/archive`, `DELETE` | Memory, Firestore | Create, Get, List, Update, Archive, Delete, Search | usable textual | Firestore opcional; embeddings/RAG pendientes |
| `tools` | `ToolDefinition` | `GET /tools`, `GET /tools/{name}` | Memory, Firestore | SeedDefaultTools, List, Get | preparatorio | Firestore opcional |
| `ai/actions` | `AIActionProposal` | `/ai/action-proposals`, `/confirm`, `/reject`, `/execute` | Memory, Firestore | Create, List, Get, Confirm, Reject, ExecuteConfirmed | parcial | Firestore opcional; AI Runtime/LLM pendiente |

## Clasificacion de features

### Listas para usarse

- Auth basico con registro/login/JWT.
- Users profile y AI settings.
- Activities CRUD operacional con ownership.
- Reminders CRUD basico sin notificaciones.
- Insights tracking: mood, outcomes, reflections, summary simple.
- Memory textual explicita con busqueda basica.

### Preparadas pero no completas

- Reminder Engine: tiene entidad, due query y estados, pero no worker real, recurrence ni push.
- AI Action Proposals: registra, confirma, rechaza y ejecuta subset minimo; no hay AI Runtime.
- Tool Registry: seed y consulta, pero sin versionado ni validacion JSON Schema real.
- Firestore repositories: implementados, pero faltan emulator tests e indices formales.

### Requieren servicios externos

- Push notifications: FCM/APNs o canal equivalente.
- LLM/tool calling real: proveedor LLM, AI runtime, prompt/context builder.
- Embeddings/RAG: proveedor embeddings, vector store o Firestore vector search si se adopta.
- Hardware/smart home: Google Home/Home Assistant/Matter/etc. No existe integracion runtime.

### Solo arquitectura o preparatorio

- `internal/memory/application.EmbeddingProvider`.
- Tools no ejecutados: `update_activity`, `complete_activity`, `cancel_reminder`, `search_memory`, `create_reflection`.
- Placeholders heredados vacios: `internal/app`, `internal/http`, `internal/domain/activity`.

## Diagnostico por bounded context

| Bounded context | Salud | Comentario |
| --- | --- | --- |
| Identity/Auth | buena | Separado de users, pero register + defaults no es transaccional. |
| User Personalization | buena | AISettings ya controla memory/reminders/autonomy. |
| Planning | buena | Activity como agregado superior evita proliferacion de Task/Habit/Routine. |
| Scheduling | parcial | Reminders persiste y consulta due, pero falta engine real. |
| Personal Insights | buena | Adecuado para tracking, sin diagnostico clinico. |
| Memory | parcial-buena | Textual usable; falta privacidad avanzada, export/delete y semantic search. |
| Tooling/AI Actions | preparatorio | Buen control humano; falta runtime, policy engine y schema validation. |
