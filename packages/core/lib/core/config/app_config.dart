/// Application configuration for Sofia Backend integration (F1).
///
/// Override base URL at build/run time:
/// `flutter run --dart-define=SOFIA_API_BASE_URL=http://localhost:8080`
class AppConfig {
  const AppConfig({
    required this.baseUrl,
    this.requestTimeout = defaultTimeout,
  });

  // Producción por defecto: un build sin --dart-define apunta a la API real.
  // Para desarrollo local: --dart-define=SOFIA_API_BASE_URL=http://localhost:8080
  static const String defaultBaseUrl = 'https://sofia-backend-m737njuj6q-uc.a.run.app';
  static const Duration defaultTimeout = Duration(seconds: 30);
  static const String baseUrlDefineKey = 'SOFIA_API_BASE_URL';

  final String baseUrl;
  final Duration requestTimeout;

  /// Reads `--dart-define=SOFIA_API_BASE_URL=...` when present.
  factory AppConfig.fromEnvironment({Duration timeout = defaultTimeout}) {
    const raw = String.fromEnvironment(baseUrlDefineKey, defaultValue: '');
    final resolved = raw.trim().isEmpty ? defaultBaseUrl : raw;
    return AppConfig(
      baseUrl: normalizeBaseUrl(resolved),
      requestTimeout: timeout,
    );
  }

  /// Removes trailing slashes so path joins never produce `//`.
  static String normalizeBaseUrl(String url) {
    var normalized = url.trim();
    while (normalized.endsWith('/')) {
      normalized = normalized.substring(0, normalized.length - 1);
    }
    return normalized;
  }

  Uri resolve(String path, [Map<String, String>? queryParameters]) {
    final normalizedPath = path.startsWith('/') ? path : '/$path';
    return Uri.parse(
      '$baseUrl$normalizedPath',
    ).replace(queryParameters: queryParameters);
  }
}
