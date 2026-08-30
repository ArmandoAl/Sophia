# Sprint V1.1.5 Report

## Resultado

PASS en código y tests. No se accedió a Firestore productivo.

## Riesgo y decisión

El snapshot documenta `conversations` legacy con schema distinto y subcolección `messages`. El repository V1 anterior usaba esa raíz y toleraba documentos incompletos. Se eligió `v1_conversations` y `v1_conversation_messages`.

## Cambios

- Nombres centralizados en `internal/database/collections.go`.
- Repositories Firestore actualizados; memory repositories y HTTP contract intactos.
- Export extendido con conversaciones y mensajes propios.
- Delete executor extendido con ambas colecciones V1 y conteos.
- Legacy preservado sin lectura, escritura, migración o borrado.

## Índices

- `v1_conversations`: `user_id ASC, updated_at DESC, id DESC`.
- `v1_conversations`: `user_id ASC, status ASC, updated_at DESC, id DESC`.
- `v1_conversation_messages`: `user_id ASC, conversation_id ASC, created_at ASC, id ASC`.

## Verificación

`go test ./...` pasa. Flutter no requiere cambios porque endpoints y modelos HTTP no cambiaron. No se ejecutaron migraciones ni Firestore real.

Riesgos restantes: los índices deben crearse manualmente antes de queries compuestas en producción; el export conserva su límite actual y no es paginado.
