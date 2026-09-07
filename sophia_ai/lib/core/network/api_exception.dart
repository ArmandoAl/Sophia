/// Normalized HTTP failure for Sofia Backend.
///
/// Backend success/error JSON usually uses `{ "message": "..." }`, but some
/// invalid paths return non-JSON bodies (`http.NotFound`).
class ApiException implements Exception {
  const ApiException({
    required this.statusCode,
    required this.message,
    this.rawBody,
    this.method,
    this.uri,
  });

  final int statusCode;
  final String message;
  final Object? rawBody;
  final String? method;
  final Uri? uri;

  bool get isBadRequest => statusCode == 400;
  bool get isUnauthorized => statusCode == 401;
  bool get isForbidden => statusCode == 403;
  bool get isNotFound => statusCode == 404;
  bool get isConflict => statusCode == 409;
  bool get isRateLimited => statusCode == 429;
  bool get isServerError => statusCode >= 500;

  @override
  String toString() =>
      'ApiException($statusCode${method != null ? ' $method' : ''}'
      '${uri != null ? ' $uri' : ''}: $message)';
}
