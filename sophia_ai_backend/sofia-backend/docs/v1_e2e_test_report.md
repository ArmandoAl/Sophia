# V1 E2E Test Report

## Commands

- `env GOCACHE=... go test ./...`
- `env GOCACHE=... go vet ./...`
- `env GOCACHE=... go build -o /tmp/sofia-v1-gate ./cmd/sofia`
- `fvm flutter analyze`
- `fvm flutter test`
- `fvm flutter build web`
- `go run ./cmd/tools/test-gemini-provider` with empty key
- `go run ./cmd/tools/run-ai-evals` with empty key

## Results

Backend tests/vet/build: PASS. Flutter analyze/tests/build: PASS, 97 Flutter tests. Web build emitted only WASM compatibility warnings from `flutter_secure_storage_web`.

Memory smoke: PASS with synthetic account. Verified health, register, login, `/users/me`, create/send/recover conversation messages, create/list reminder. Response assistant was non-empty and history contained two messages.

Firestore: BLOCKED by external configuration. Gemini: BLOCKED/skipped because `GEMINI_API_KEY` is absent.
