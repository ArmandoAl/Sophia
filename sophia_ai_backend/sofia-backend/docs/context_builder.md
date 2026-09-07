# Context Builder

Fecha: 2026-07-01

## Proposito

El Context Builder inicial prepara un resumen compacto para futuros modelos sin enviar todo el historial ni contenido sensible completo.

## Fuentes actuales

Para el usuario autenticado:

- auth user,
- user profile,
- AI settings,
- recent activities,
- due reminders,
- insights summary,
- relevant memories via busqueda textual simple.

## Limites

El limite inicial por seccion es `5`.

Secciones limitadas:

- `recent_activities`
- `due_reminders`
- `relevant_memories`

Insights se resume en metricas agregadas:

- average energy,
- average stress,
- average focus,
- total outcomes,
- total reflections.

## Privacidad

El builder no acepta `user_id` desde body. Todas las consultas usan el `user_id` autenticado.

No incluye contenido completo de memory. Para memories expone resumen/titulo, tipo, estado, importancia y tags.

## AISettings

| Setting | Efecto |
| --- | --- |
| `memory_enabled=false` | Excluye memories del contexto. |
| `reminders_enabled=false` | Excluye due reminders. |
| `planning_enabled=false` | No bloquea lectura de activities hoy, pero bloquea planning tools. |

## Riesgos

- Todavia no hay token budget real.
- La busqueda textual de memory es basica.
- Falta clasificacion de sensibilidad por campo/contenido.
- Falta auditoria de contexto enviado a providers externos.

## Proximo paso recomendado

Antes de integrar LLM real, convertir `ContextSummary` en un prompt/context package versionado con:

- limites por tokens,
- redaccion por sensibilidad,
- visibility de memory,
- trazas auditables,
- snapshots de eval.
