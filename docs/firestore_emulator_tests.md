# Firestore Emulator Tests

Sprint 12 adds an emulator-backed integration test suite in `internal/database/firestore_emulator_test.go`.

## Scope

The suite verifies the production Firestore client path without requiring real cloud credentials:

- Firestore bootstrap with existing config.
- Activities repository create/list with cursor-ready ordering.
- Reminders repository create/list.
- Memory repository create/search/delete-soft behavior.
- Deleted memories are excluded from basic search.

## Running

Start the emulator separately:

```bash
firebase emulators:start --only firestore
```

Then run:

```bash
ENV=test \
PERSISTENCE_DRIVER=firestore \
FIRESTORE_PROJECT_ID=demo-sofia \
FIRESTORE_EMULATOR_HOST=127.0.0.1:8080 \
go test ./internal/database -run TestFirestoreEmulatorRepositories
```

If `FIRESTORE_PROJECT_ID` or `FIRESTORE_EMULATOR_HOST` is missing, the test skips. This keeps `go test ./...` deterministic on machines without the emulator.

## Database IDs

If the project uses a named database, set:

```bash
FIRESTORE_DATABASE_ID=default
```

The backend uses the configured database ID when the Firestore client supports it.

## Privacy

The emulator suite creates synthetic test data only. It must not depend on real user documents or production credentials.

## Remaining Gaps

- Add emulator coverage for user registration transactions.
- Add emulator coverage for activity + reminder transactional behavior once a shared Firestore unit of work is introduced.
- Add emulator coverage for action proposal execution once it can commit proposal state and domain writes atomically.
