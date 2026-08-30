# Migration Apply Checklist

Fecha: 2026-07-01

## Estado

No ejecutar apply todavia.

## Comando de validacion previa

```sh
go run ./cmd/tools/migrate-firestore-legacy -validate-only=true -target-mode=modular-only -limit=20 -output-dir=docs
```

Debe generar:

- `docs/migration_preview.json`
- `docs/migration_preview.md`

## Checklist obligatorio antes de apply

- [ ] Confirmar backup/export de Firestore.
- [ ] Ejecutar `-validate-only=true`.
- [ ] Confirmar que `conflicts` esta vacio o solo contiene severidad `medium` revisada.
- [ ] Confirmar que no hay conflictos `high`.
- [ ] Confirmar que no hay conflictos `critical`.
- [ ] Revisar manualmente `users/user_001`.
- [ ] No aplicar `users` auth si falta `password_hash`.
- [ ] Confirmar que `users` no aparece como target en `modular-only`.
- [ ] Confirmar que `activities` y `reminders` no aparecen como target en `modular-only`.
- [ ] Confirmar timezone canonico por usuario.
- [ ] Confirmar si fechas de mayo 2026 deben importarse, archivarse o ignorarse.
- [ ] Confirmar que `activity_001` y `reminder_001` se revisaron primero en staging.
- [ ] No usar `-target-mode=in-place` hasta tener backup y decision explicita.
- [ ] Confirmar que `user_profiles/user_001` no existe.
- [ ] Confirmar que `ai_settings/user_001` no existe.
- [ ] Ejecutar primero en emulator o proyecto clonado.
- [ ] Revisar logs despues de dry-run.

## Comando de apply permitido solo despues del checklist

```sh
go run ./cmd/tools/migrate-firestore-legacy -dry-run=false -apply=true -target-mode=modular-only -limit=20 -output-dir=docs
```

Notas:

- `modular-only` no genera targets `manual_review`.
- `modular-only` no genera targets `users`, `activities` ni `reminders`.
- Conflictos `high` o `critical` bloquean apply.

## Condiciones para detener

Detener si:

- aparece un target `users/*` que no fue revisado,
- aparece conflicto `high` o `critical`,
- aparece timezone mixto sin decision,
- aparecen reminders en pasado que podrian disparar workers futuros,
- aparece target existente en `user_profiles` o `ai_settings`,
- hay mas documentos de los esperados por `-limit`.

## Rollback

No hay rollback automatico todavia.

Antes de apply, preparar:

- backup Firestore,
- lista exacta de documentos destino,
- plan manual de revert por documento,
- ventana de mantenimiento,
- verificacion post-apply con exporter.
