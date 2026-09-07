# Firestore Migration Plan

Fecha: 2026-07-01

## Objetivo

Preparar una migracion no destructiva desde el modelo legacy hacia el modelo modular actual.

## Herramienta

```sh
go run ./cmd/tools/migrate-firestore-legacy -dry-run=true -apply=false -limit=20 -output-dir=docs
```

Por defecto:

- no escribe datos,
- genera `docs/migration_preview.json`,
- genera `docs/migration_preview.md`.

Para aplicar cambios no destructivos:

```sh
go run ./cmd/tools/migrate-firestore-legacy -dry-run=false -apply=true -limit=20 -output-dir=docs
```

## Configuracion requerida

- `PERSISTENCE_DRIVER=firestore`
- `FIRESTORE_PROJECT_ID`
- `FIRESTORE_DATABASE_ID` si se usa database nombrada, por ejemplo `default`
- `GOOGLE_APPLICATION_CREDENTIALS` o credenciales de entorno
- `FIRESTORE_EMULATOR_HOST` si se usa emulator

## Fase 1: Dry-run

1. Ejecutar preview con `-dry-run=true`.
2. Revisar `migration_preview.md`.
3. Revisar warnings de documentos con campos faltantes.
4. Confirmar que los IDs destino son correctos.
5. Confirmar que no se migran colecciones hardware/RAG.

## Fase 2: Migracion inicial preparada

La herramienta solo prepara/aplica:

- `users -> user_profiles`
- `users -> ai_settings`
- `activities legacy -> activities`
- `reminders legacy -> reminders`

`users -> users` queda como `manual_review` porque el modelo legacy no contiene `password_hash`.

## Fase 3: Futuro

Despues de privacy/export/delete y AI Runtime hardening:

- `ai_memory -> memories`
- `ai_actions_log -> audit log/action proposal history`
- `ai_policy -> policy engine`
- `conversations/message_embeddings -> conversations/RAG`

Despues de definir hardware runtime:

- `bodies`
- `locations`

## Garantias

- No borra datos.
- No borra colecciones legacy.
- No escribe nada por defecto.
- Con `-apply=true`, usa merge writes y salta targets de `manual_review`.

## Riesgos antes de aplicar

- Colecciones destino `activities` y `reminders` comparten nombre con legacy; usar emulator o backup antes de apply.
- Auth users requieren estrategia separada por `password_hash`.
- No hay transacciones cross-collection todavia.
- No hay checkpoint/resume para migraciones grandes.

## Recomendacion

Ejecutar primero contra Firestore emulator o contra un proyecto clonado. Para produccion, tomar backup/export de Firestore antes de cualquier `-apply=true`.
