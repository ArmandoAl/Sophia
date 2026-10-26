/// Configuración pública de la compilación (`--dart-define`); nada de aquí es secreto.
///
/// - `DEMO_MODE=true`: repositorios simulados de `sophia_core` (sin red).
/// - `SOFIA_API_BASE_URL=https://…`: API real; el backend debe permitir el origen del preview (CORS).
const bool kDemoMode = bool.fromEnvironment('DEMO_MODE', defaultValue: false);
