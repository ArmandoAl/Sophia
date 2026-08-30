# Sprint 8 Report - Firestore Schema Exporter & Database Snapshot

Fecha: 2026-07-01

## Objetivo

Crear una herramienta interna para inspeccionar Firestore y generar artefactos legibles por humanos y agentes de IA sin modificar logica de producto ni exponer datos sensibles completos.

## Entregables

- `cmd/tools/export-firestore-schema/main.go`
- `internal/platform/firestoreexporter/exporter.go`
- `internal/platform/firestoreexporter/exporter_test.go`
- `docs/firestore_exporter.md`
- `docs/sprint_8_report.md`

Cuando se ejecuta con Firestore configurado, el comando genera:

- `docs/database_snapshot.json`
- `docs/database_schema.md`
- `docs/database_collections_summary.md`

## Comando

```sh
go run ./cmd/tools/export-firestore-schema -limit=20 -output-dir=docs -redact=true
```

## Configuracion usada

El comando reutiliza `internal/config.Load()` y `internal/database.NewFirestore()`.

Variables relevantes:

- `ENV`
- `PERSISTENCE_DRIVER`
- `FIRESTORE_PROJECT_ID`
- `GOOGLE_APPLICATION_CREDENTIALS`
- `FIRESTORE_EMULATOR_HOST`

Si `PERSISTENCE_DRIVER` no es `firestore`, el comando falla con un mensaje claro. Para una ejecucion explicita en local se puede usar `-force`, pero siempre requiere `FIRESTORE_PROJECT_ID`.

## Que inspecciona

- Colecciones raiz.
- Documentos muestreados por coleccion.
- Campos detectados.
- Tipos inferidos.
- Campos opcionales en la muestra.
- Arrays, maps y timestamps.
- Nombres de subcolecciones encontrados en documentos muestreados.

## Redaccion

El exportador no exporta valores completos para campos sensibles. Redacta por nombre de campo cuando detecta patrones como:

- `password_hash`
- `token`
- `secret`
- `email`
- `content`
- `proposed_input`
- `execution_result`
- `note`
- `reflection`

La meta es auditar estructura, no datos personales.

## Comparacion contra modelo esperado

El comando lee `docs/data_model_expected.md` y compara:

- colecciones esperadas vs existentes,
- colecciones faltantes,
- colecciones inesperadas,
- campos faltantes evidentes,
- campos inesperados evidentes.

La comparacion es heuristica porque el modelo esperado vive en Markdown.

## Pruebas agregadas

Se agregaron tests unitarios para:

- inferencia de tipos,
- redaccion de campos sensibles,
- inferencia de schema desde documentos simulados,
- comparacion expected vs actual,
- render basico de Markdown,
- parseo del modelo esperado.

## Pruebas ejecutadas

```sh
env GOCACHE=/Users/armandoalvarado/Documents/Sophia/sophia_ai_backend/sofia-backend/.cache/go-build go test ./...
```

Resultado: suite completa pasando.

## Ejecucion del exportador

Se intento ejecutar:

```sh
env GOCACHE=/Users/armandoalvarado/Documents/Sophia/sophia_ai_backend/sofia-backend/.cache/go-build go run ./cmd/tools/export-firestore-schema -limit=20 -output-dir=docs -redact=true
```

Resultado:

```text
PERSISTENCE_DRIVER="memory"; set PERSISTENCE_DRIVER=firestore or pass -force for explicit schema export
```

No se generaron `database_snapshot.json`, `database_schema.md` ni `database_collections_summary.md` en este entorno porque Firestore no esta configurado como persistence driver activo. Esto evita producir snapshots falsos o conectarse a un proyecto ambiguo.

## Limitaciones

- No requiere Firestore real para tests.
- No valida indices reales en Firestore.
- No inspecciona profundamente schemas de mapas anidados.
- No detecta todos los campos si no aparecen en la muestra.
- `required_in_sample` no equivale a required global.
- Subcolecciones solo se detectan dentro de documentos muestreados.

## Riesgos

- `database_snapshot.json` puede contener valores no sensibles pero aun contextuales.
- Campos sensibles con nombres no previstos podrian no redactarse.
- Los IDs de documentos se exportan como `_document_id` para auditoria estructural.

## Uso recomendado

1. Ejecutar contra emulator o entorno controlado.
2. Revisar manualmente `database_snapshot.json`.
3. Compartir preferentemente `database_schema.md` y `database_collections_summary.md`.
4. No commitear credenciales ni exports no revisados.
