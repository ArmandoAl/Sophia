import 'api_exception.dart';

/// Defensive cast of ApiClient decoded bodies to a JSON object map.
Map<String, dynamic> requireJsonMap(dynamic body, {required String context}) {
  if (body is Map<String, dynamic>) return body;
  if (body is Map) return Map<String, dynamic>.from(body);
  throw ApiException(
    statusCode: 200,
    message: 'Unexpected response shape for $context',
    rawBody: body,
  );
}
