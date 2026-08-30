# Technical Debt Ledger

Fecha: 2026-06-30

## Critica antes de LLM

1. **Privacy/export/delete definitivo**
   - Memory, reflections, mood y proposed_input pueden contener datos sensibles.
   - `DELETE /memory/{id}` es soft delete, no borrado fisico.
   - Falta exportacion y eliminacion verificable por usuario.

2. **AI Runtime policy**
   - `ai/actions` tiene reglas basicas, pero no policy engine formal.
   - Falta separar propuestas generadas por LLM vs usuario/sistema.
   - Falta auditoria de prompts/contexto/tool calls.

3. **JSON Schema validation real**
   - `ToolDefinition.input_schema` es documentativo.
   - El executor valida de forma parcial con structs Go.

4. **Context Builder**
   - No existe ensamblador de contexto con memory visibility, importance, recency y token budget.

5. **Firestore legacy reconciliation**
   - Firestore real contiene colecciones legacy/conceptuales fuera del contrato actual.
   - `users`, `activities` y `reminders` existen con schema parcial distinto.
   - `ai_memory`, `conversations` y embeddings no deben importarse hasta cerrar privacidad/RAG.

## Alta prioridad

- Firestore emulator tests para auth/users/activities/reminders/insights/memory/tools/actions.
- Transacciones para:
  - register + profile + AISettings,
  - activity + reminder,
  - action proposal execution + domain write.
- Paginacion por cursor en listados.
- Indices Firestore definidos y documentados.
- Validacion timezone IANA.
- Validacion locale BCP 47.
- Expiracion automatica de action proposals.
- Worker de reminders con locking/idempotencia.

## Media prioridad

- Recurrence real para reminders.
- Push notifications via FCM/APNs u otro canal.
- Versionado de tools.
- Auditoria de cambios en memory y tools.
- Unificar responses DTO para evitar exponer structs de dominio directamente en tools/actions.
- Seguir vigilando crecimiento de `internal/app`; si aumenta, partir builders por contexto sin introducir DI framework pesado.

## Baja prioridad / limpieza

- Remover o documentar placeholders vacios:
  - `internal/app`
  - `internal/http`
  - `internal/domain/activity`
- Revisar carpetas vacias heredadas:
  - `internal/services`
  - `internal/infrastructure`
  - `internal/shared`
- Actualizar codebase-memory exclusions: el index actual excluye `internal/tools`, lo que reduce visibilidad futura.

## Riesgos acumulados

- La suite es principalmente integration-style en `internal/server/routes_test.go`; hay pocos tests unitarios por modulo.
- Firestore behavior real puede divergir del repositorio memory.
- No hay rate limiting ni proteccion brute-force para auth.
- No hay refresh tokens ni revocacion JWT.
- No hay migraciones/versionado de modelo documental.
- Migraciones Firestore todavia no tienen checkpoint/resume ni transacciones cross-collection.
