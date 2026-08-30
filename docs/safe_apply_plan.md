# Safe Apply Plan

Fecha: 2026-07-01

## Objetivo

Permitir un apply minimo y reversible solo para documentos modulares derivados de legacy user:

- `user_profiles/user_001`
- `ai_settings/user_001`

No incluye:

- `users/user_001`
- `activities/activity_001`
- `reminders/reminder_001`

## Paso 1: Validate-only real

Ejecutar contra Firestore real:

```sh
PERSISTENCE_DRIVER=firestore \
FIRESTORE_PROJECT_ID=<project_id> \
go run ./cmd/tools/migrate-firestore-legacy \
  -validate-only=true \
  -target-mode=modular-only \
  -limit=20 \
  -output-dir=docs
```

Validar:

- no existe conflicto `high`,
- no existe conflicto `critical`,
- targets solo incluyen `user_profiles` y `ai_settings`,
- `users` no aparece como target,
- `activities` y `reminders` no aparecen como target.

## Paso 2: Backup

Antes de cualquier apply:

- exportar Firestore,
- guardar preview generado,
- guardar lista exacta de documentos destino,
- confirmar rollback manual.

## Paso 3: Apply minimo

Solo si validate-only sale limpio:

```sh
PERSISTENCE_DRIVER=firestore \
FIRESTORE_PROJECT_ID=<project_id> \
go run ./cmd/tools/migrate-firestore-legacy \
  -dry-run=false \
  -apply=true \
  -target-mode=modular-only \
  -limit=20 \
  -output-dir=docs
```

## Paso 4: Verificacion

Despues del apply:

- ejecutar exporter Firestore,
- confirmar `user_profiles/user_001`,
- confirmar `ai_settings/user_001`,
- confirmar que `users/user_001` no fue modificado,
- confirmar que `activities` y `reminders` no fueron modificados.

## Rollback manual

Si algo sale mal:

- borrar solo documentos creados en `user_profiles` y `ai_settings`,
- no tocar `users`,
- no tocar legacy collections.

## Bloqueos

Detener si:

- `target-mode` no es `modular-only`,
- aparece cualquier target `users`,
- aparece cualquier target `activities` o `reminders`,
- existe conflicto `high` o `critical`,
- no existe backup/export.
