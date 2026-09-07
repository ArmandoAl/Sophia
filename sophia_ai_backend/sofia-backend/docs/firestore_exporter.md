# Firestore Schema Exporter

## Purpose

`cmd/tools/export-firestore-schema` is an internal CLI tool that inspects Firestore and writes human-readable and agent-readable database documentation.

It does not change product state, create business records, call LLMs or modify API endpoints.

## Command

```sh
go run ./cmd/tools/export-firestore-schema -limit=20 -output-dir=docs -redact=true
```

## Required Environment

The command reuses the backend configuration loader.

Required for normal execution:

- `ENV`
- `PERSISTENCE_DRIVER=firestore`
- `FIRESTORE_PROJECT_ID`

Optional for local development:

- `FIRESTORE_DATABASE_ID`
- `GOOGLE_APPLICATION_CREDENTIALS`
- `FIRESTORE_EMULATOR_HOST`

If `PERSISTENCE_DRIVER` is not `firestore`, the command exits with a clear error. For explicit local inspection it can be run with `-force`, but `FIRESTORE_PROJECT_ID` is still required.

If your Firebase Console shows a named database such as `default`, set:

```sh
export FIRESTORE_DATABASE_ID=default
```

When this variable is present, the backend/exporter uses `firestore.NewClientWithDatabase`. When it is absent, it keeps the SDK default database behavior.

## Flags

| Flag | Default | Description |
| --- | ---: | --- |
| `-limit` | `20` | Maximum documents sampled per root collection. |
| `-output-dir` | `docs` | Directory where output files are written. |
| `-redact` | `true` | Redacts sensitive fields in sample documents. |
| `-force` | `false` | Allows running when `PERSISTENCE_DRIVER` is not `firestore`. |

## Generated Files

| File | Purpose |
| --- | --- |
| `docs/database_snapshot.json` | Machine-readable snapshot with collections, inferred fields, sample counts and redacted sample documents. |
| `docs/database_schema.md` | Human-readable schema summary with types, optional fields, sensitive fields and index recommendations. |
| `docs/database_collections_summary.md` | Compact status table comparing actual collections against `docs/data_model_expected.md`. |

## Redaction

The exporter redacts or truncates fields whose names suggest sensitive data, including:

- `password_hash`
- `password`
- `token`
- `secret`
- `email`
- `content`
- `proposed_input`
- `execution_result`
- `note`
- `reflection`

This is intentionally conservative. The output can still reveal structure, collection names, IDs, timestamps and non-sensitive values. Review generated files before sharing outside the engineering team.

## Expected Model Comparison

The tool reads `docs/data_model_expected.md` from the output directory and compares:

- expected collections vs actual root collections,
- unexpected collections,
- missing collections,
- evident field differences when expected fields are listed.

The comparison is heuristic because `data_model_expected.md` is documentation, not a strict schema contract.

## Firestore Emulator

For emulator usage:

```sh
export ENV=development
export PERSISTENCE_DRIVER=firestore
export FIRESTORE_PROJECT_ID=sofia-local
export FIRESTORE_DATABASE_ID=default
export FIRESTORE_EMULATOR_HOST=localhost:8081
go run ./cmd/tools/export-firestore-schema -limit=20 -output-dir=docs -redact=true
```

## Limitations

- Samples root collections only.
- Detects subcollection names only from sampled documents.
- Does not infer nested map schemas field-by-field.
- Does not validate Firestore indexes.
- Does not prove required fields globally; `required_in_sample` only means present in all sampled documents.
- Does not export full sensitive values.

## Sharing Guidance

Safe-ish to share internally after review:

- `database_schema.md`
- `database_collections_summary.md`

Review carefully before sharing:

- `database_snapshot.json`

Never share credentials, service account JSON files or unredacted exports.
