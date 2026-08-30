# App Composition

Sprint 23 moves application wiring out of `cmd/sofia/main.go` into `internal/app` without adding a DI framework.

## Entry Point

`cmd/sofia/main.go` now only:

1. loads config,
2. builds the app container,
3. starts the HTTP server,
4. closes resources on exit.

## Package

```text
internal/app
```

Files:

- `container.go`: top-level `Build(config.Config)` orchestration.
- `repositories.go`: memory/Firestore repository selection.
- `modules.go`: application services, bridges, AI runtime and model provider wiring.
- `handlers.go`: HTTP handler construction and `server.Routes`.
- `workers.go`: documented boundary for future shared worker/provider builders.

## Repository Registration

Repositories are registered in:

```text
internal/app/repositories.go
```

Rules:

- `PERSISTENCE_DRIVER=memory` builds in-memory repositories.
- `PERSISTENCE_DRIVER=firestore` bootstraps Firestore once and builds Firestore repositories.
- Firestore close handling lives in `Repositories.Close`.

## Module Registration

Application services are registered in:

```text
internal/app/modules.go
```

This file wires:

- auth token service,
- users,
- activities,
- reminders,
- insights,
- memory,
- notifications,
- tools,
- AI action proposals,
- privacy,
- AI runtime.

Cross-module bridges remain explicit:

- activities uses reminder bridge,
- AI actions uses privacy audit recorder,
- AI runtime uses privacy audit recorder.

## Handler Registration

HTTP handlers and routes are registered in:

```text
internal/app/handlers.go
```

Public endpoint paths still live in `internal/server/routes.go`. Sprint 23 does not change routes or behavior.

## Model Provider

`BuildModelClient` remains explicit:

- fake provider by default,
- Gemini only when configured.

No LLM provider is constructed inside handlers.

## Worker Composition

Standalone worker commands still compose their own process-specific dependencies under `cmd/workers/*`.

`internal/app/workers.go` exists as a documented boundary for future shared worker builders if duplication becomes real.

## Principles

- no Wire/Fx/Dig,
- no reflection DI,
- explicit structs and constructor calls,
- domain modules unchanged,
- public API unchanged.
