# Guía para agentes

- Desarrollo en servidor (comandos, preview, túnel): `docs/SERVER_DEVELOPMENT.md`. Paridad móvil ↔ web: `docs/web_parity.md`. Guía de la migración: `docs/FLUTTER_JASPR_SERVER_BASE.md`.
- `packages/core` debe seguir siendo Dart puro (sólo `bloc`, `equatable`, `http`); hay una prueba que lo exige. `flutter_bloc`, GetIt y GoRouter se quedan en `apps/mobile`.
- Rama de trabajo `migration/jaspr-server-20261026`; `demo` y `main` son líneas distintas y no se fusionan.
- La web no debe apuntar al backend de producción en pruebas: usa el modo demo (`DEMO_MODE=true`). `AppConfig.defaultBaseUrl` apunta a Cloud Run.
- No versionar secretos ni `config/`; sin `push --force`.
