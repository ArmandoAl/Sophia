# System Overview

Fecha: 2026-07-03

## Resumen

Sofia AI Backend es un backend Go modular para una IA personal. Hoy soporta identidad, perfil, configuracion de Sofía, activities, reminders, insights, memory textual, privacidad basica, tools, action proposals y AI Runtime proposal-only.

El sistema esta diseñado para crecer hacia IA con tool calling, pero mantiene barreras claras: no hay ejecucion autonoma, no hay LLM obligatorio y no hay notificaciones reales por default.

## Procesos

| Proceso | Comando | Estado |
| --- | --- | --- |
| API HTTP | `cmd/sofia` | Principal. Carga config, construye `internal/app` y arranca server. |
| Reminder worker | `cmd/workers/reminders` | Opcional. Deshabilitado por default. Usa claim/lease para Firestore multi-instancia. |
| Firestore exporter | `cmd/tools/export-firestore-schema` | Herramienta interna de snapshot/schema. |
| Legacy migrator | `cmd/tools/migrate-firestore-legacy` | Dry-run/staging/modular-only; no destructivo por default. |
| Delete executor | `cmd/tools/process-delete-requests` | Procesa solo delete requests aprobadas. |
| AI eval runner | `cmd/tools/run-ai-evals` | Fake siempre; Gemini solo si hay API key. |

## Composition Root

`cmd/sofia/main.go` es intencionalmente pequeño:

1. carga `internal/config`,
2. llama `app.Build(cfg)`,
3. difiere `Close()`,
4. arranca `ListenAndServe`.

`internal/app` ya no es placeholder. Es el composition root real del backend:

- `container.go`: orquestacion de build.
- `repositories.go`: repositorios memory/Firestore.
- `modules.go`: servicios, bridges, AI runtime y providers.
- `handlers.go`: handlers HTTP y rutas.
- `workers.go`: frontera documentada para builders futuros.

## Modulos Principales

| Modulo | Responsabilidad |
| --- | --- |
| `auth` | Registro, login, bcrypt, JWT. |
| `users` | Profile, AI settings, onboarding, `/users/me`. |
| `activities` | Activity como unidad de planning. |
| `reminders` | Reminder CRUD, due query, recurrence simple, worker support. |
| `notifications` | Device tokens y provider FCM opcional. |
| `insights` | Mood, outcomes, reflections, summary. |
| `memory` | Memoria textual explicita y busqueda basica. |
| `privacy` | User export, delete requests, audit logs. |
| `tools` | Tool registry. |
| `ai/actions` | AI action proposals y ejecucion explicita controlada. |
| `ai/runtime` | Context builder, tool selection, safety, fake/Gemini provider. |
| `ai/evals` | Evaluaciones de seguridad/runtime. |

## Persistencia

El backend soporta `PERSISTENCE_DRIVER=memory|firestore`.

Firestore es infraestructura, no dominio. Los use cases dependen de contratos de repositorio. Firestore real todavia conserva colecciones legacy, por lo que las migraciones siguen en modo seguro/no destructivo.

## Seguridad Por Default

- JWT en endpoints privados.
- Ownership por `user_id` del token.
- `user_id` no se acepta desde body en modulos de usuario.
- `memory_enabled`, `reminders_enabled` y `planning_enabled` gobiernan capacidades.
- AI Runtime es proposal-only.
- Gemini no es provider default.
- FCM esta deshabilitado y dry-run por default.
- Reminder worker esta deshabilitado por default.
- Delete fisico requiere request aprobada y comando interno.

## Fronteras Importantes

- No hay autonomia real de IA.
- No hay scheduler complejo; el reminder worker tiene claim/lease Firestore pero requiere validacion operacional multi-instancia.
- No hay push real activado por default.
- No hay RAG semantico.
- No hay hardware/smart home en runtime actual.
