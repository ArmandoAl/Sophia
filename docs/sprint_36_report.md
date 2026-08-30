# Sprint 36 Report: Flutter Endpoint Integration Contract

Fecha: 2026-07-05

## Objetivo

Crear una mega auditoria funcional de endpoints existentes para usarla como contrato de integracion con Flutter, sin modificar endpoints ni logica de producto.

## Fuentes Revisadas

- `docs/postman_collection.json`
- `docs/api_smoke_tests.md`
- `docs/manual_test_runbook_v0_1.md`
- `docs/local_e2e_report.md`
- `docs/current_capabilities.md`
- `docs/system_overview.md`
- `docs/architecture_state.md`
- `docs/auth_hardening.md`
- `docs/notifications_domain.md`
- `docs/reminder_worker.md`
- `docs/ai_runtime_safety.md`
- `internal/server/routes.go`
- handlers/requests HTTP principales para activities, reminders, memory y AI runtime.

## Documentos Creados

- `docs/flutter_endpoint_integration_contract.md`
- `docs/flutter_integration_roadmap.md`
- `docs/flutter_models_needed.md`
- `docs/flutter_screens_mapping.md`
- `docs/sprint_36_report.md`

## Cobertura

Se documento el contrato de integracion para:

- Auth: register/login.
- Users: `me`, profile, AI settings, onboarding, export, delete request.
- Activities: CRUD operativo y acciones complete/cancel/archive.
- Reminders: CRUD operativo, due, recurrence simple y acciones.
- Insights: mood, outcomes, reflections y summary.
- Memory: CRUD textual, archive, soft delete y search basico.
- Notifications: device tokens redacted.
- Tools: registry consultable.
- AI Actions: proposal lifecycle y ejecucion explicita.
- AI Runtime: message proposal-only.
- System: health.

## Decisiones Para Flutter

- Usar repositories por modulo y Cubit/BLoC por flujo.
- Tratar AI Runtime y AI Actions como `parcial` y feature-flagged.
- Tratar Notifications como `parcial`: device tokens estan listos, FCM real no debe activarse por default.
- Tratar delete request como flujo controlado, no borrado inmediato.
- Usar `GET /users/me` como bootstrap principal de sesion.
- No enviar `user_id` desde Flutter en operaciones de datos del usuario.

## Que No Debe Conectarse Todavia

- Gemini real como dependencia obligatoria.
- FCM real visible para usuarios.
- Delete executor.
- Migraciones legacy.
- Hardware/smart home.
- RAG/embeddings.

## Cambios de Codigo

No se hicieron cambios de codigo de producto. Este sprint fue documental/auditoria.

## Pruebas

Ejecutado correctamente:

```sh
go test ./...
```

Resultado: PASS.

## Riesgos

- Algunos responses exactos pueden necesitar ajuste fino al generar cliente tipado desde pruebas reales, especialmente wrappers (`activity`, `reminder`, listas) si Flutter usa modelos estrictos.
- AI Runtime y Action Proposals deben mantenerse proposal-only hasta terminar validaciones con Gemini real.
- Notifications requiere permisos de OS y FCM config; no debe bloquear el uso principal de la app.
