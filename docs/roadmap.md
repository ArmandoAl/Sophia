# Roadmap

Fecha: 2026-06-30

## Antes de integrar LLM

1. Privacy baseline
   - Export de datos de usuario.
   - Delete fisico o proceso verificable por coleccion.
   - Politica de retencion.

2. Firestore hardening
   - Reconciliar modelo legacy vs modelo modular.
   - Ejecutar migracion dry-run y revisar preview.
   - Emulator tests.
   - Indices compuestos.
   - Transacciones criticas.
   - Paginacion por cursor.

3. AI Runtime safety
   - Context Builder con token budget.
   - Tool policy engine.
   - JSON Schema validation real.
   - Audit log de prompts/proposals/executions.

4. Reminder runtime
   - Worker real.
   - Lock/idempotencia.
   - Recurrence parser.
   - Push channel.

## Siguiente bloque recomendado

### Sprint 9.5 - Data Model Reconciliation

- Clasificar colecciones legacy.
- Generar preview de migracion no destructiva.
- Preparar mapping `users/activities/reminders`.
- No migrar `ai_memory`, `conversations`, embeddings ni hardware todavia.

### Sprint 10 - Privacy & Data Controls

- Export/delete definitivo.
- Audit log basico.
- Memory visibility enforcement.
- Tests de privacidad/ownership.

### Sprint 11 - Firestore Production Readiness

- Emulator suite.
- Indices.
- Transacciones.
- Cursor pagination.

### Sprint 12 - Reminder Runtime

- Worker.
- Recurrence.
- Delivery attempts.
- Push abstraction sin proveedor acoplado.

### Sprint 13 - AI Context Builder

- Context assembly desde user/settings/activities/memory/insights.
- Ranking por importance/recency/visibility.
- Token budget.
- No LLM todavia o LLM mock.

### Sprint 14 - LLM Integration, Controlled

- Provider interface.
- Prompt templates/versioning.
- Tool selection en modo proposal-only.
- Safety tests.

## Hardware/smart home

No debe entrar antes de:

- Tool policy engine.
- Confirmation UX.
- Audit log.
- Per-tool permissions.
- External provider abstraction.

Estado actual: pendiente, fuera del backend runtime.
