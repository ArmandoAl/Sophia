import 'dart:async';
import 'dart:convert';

import 'package:http/http.dart' as http;

import '../config/app_config.dart';
import '../storage/token_storage.dart';
import 'api_exception.dart';

typedef JsonMap = Map<String, dynamic>;

/// Reusable HTTP client for Sofia Backend v0.1.
///
/// Contract rules:
/// - No generic `{data}` / `{items}` wrappers assumed.
/// - Errors usually `{ "message": "..." }`; non-JSON bodies are tolerated.
/// - Bearer token attached when [TokenStorage] has a value.
class ApiClient {
  ApiClient({
    required AppConfig config,
    required TokenStorage tokenStorage,
    http.Client? httpClient,
  }) : _config = config,
       _tokenStorage = tokenStorage,
       _http = httpClient ?? http.Client(),
       _ownsHttpClient = httpClient == null;

  final AppConfig _config;
  final TokenStorage _tokenStorage;
  final http.Client _http;
  final bool _ownsHttpClient;

  AppConfig get config => _config;

  void close() {
    if (_ownsHttpClient) {
      _http.close();
    }
  }

  Future<dynamic> get(
    String path, {
    Map<String, String>? queryParameters,
    bool authenticated = true,
  }) {
    return _send(
      method: 'GET',
      path: path,
      queryParameters: queryParameters,
      authenticated: authenticated,
    );
  }

  Future<dynamic> post(
    String path, {
    Object? body,
    Map<String, String>? queryParameters,
    bool authenticated = true,
  }) {
    return _send(
      method: 'POST',
      path: path,
      body: body,
      queryParameters: queryParameters,
      authenticated: authenticated,
    );
  }

  Future<dynamic> patch(
    String path, {
    Object? body,
    Map<String, String>? queryParameters,
    bool authenticated = true,
  }) {
    return _send(
      method: 'PATCH',
      path: path,
      body: body,
      queryParameters: queryParameters,
      authenticated: authenticated,
    );
  }

  Future<dynamic> delete(
    String path, {
    Object? body,
    Map<String, String>? queryParameters,
    bool authenticated = true,
  }) {
    return _send(
      method: 'DELETE',
      path: path,
      body: body,
      queryParameters: queryParameters,
      authenticated: authenticated,
    );
  }

  Future<dynamic> _send({
    required String method,
    required String path,
    Object? body,
    Map<String, String>? queryParameters,
    required bool authenticated,
  }) async {
    final uri = _config.resolve(path, queryParameters);
    final headers = <String, String>{
      'Accept': 'application/json',
      'Content-Type': 'application/json',
    };

    if (authenticated) {
      final token = await _tokenStorage.readToken();
      if (token != null && token.isNotEmpty) {
        headers['Authorization'] = 'Bearer $token';
      }
    }

    final encodedBody = body == null ? null : jsonEncode(body);

    late final http.Response response;
    try {
      response = await _http
          .send(
            http.Request(method, uri)
              ..headers.addAll(headers)
              ..body = encodedBody ?? '',
          )
          .then(http.Response.fromStream)
          .timeout(_config.requestTimeout);
    } on TimeoutException {
      throw ApiException(
        statusCode: 0,
        message: 'Request timed out after ${_config.requestTimeout.inSeconds}s',
        method: method,
        uri: uri,
      );
    } on http.ClientException catch (e) {
      throw ApiException(
        statusCode: 0,
        message: e.message,
        rawBody: e,
        method: method,
        uri: uri,
      );
    }

    return _decodeResponse(response, method: method, uri: uri);
  }

  dynamic _decodeResponse(
    http.Response response, {
    required String method,
    required Uri uri,
  }) {
    final status = response.statusCode;
    final raw = response.body;

    if (status >= 200 && status < 300) {
      if (raw.trim().isEmpty) {
        return null;
      }
      try {
        return jsonDecode(raw);
      } on FormatException {
        throw ApiException(
          statusCode: status,
          message: 'Invalid JSON in success response',
          rawBody: raw,
          method: method,
          uri: uri,
        );
      }
    }

    throw _exceptionFromErrorBody(
      statusCode: status,
      rawBody: raw,
      method: method,
      uri: uri,
    );
  }

  ApiException _exceptionFromErrorBody({
    required int statusCode,
    required String rawBody,
    required String method,
    required Uri uri,
  }) {
    String message = _defaultMessageForStatus(statusCode);
    Object? parsed = rawBody;

    final trimmed = rawBody.trim();
    if (trimmed.isNotEmpty) {
      try {
        final decoded = jsonDecode(trimmed);
        parsed = decoded;
        if (decoded is Map<String, dynamic>) {
          final m = decoded['message'];
          if (m is String && m.isNotEmpty) {
            message = m;
          }
        }
      } on FormatException {
        // Non-JSON error bodies (e.g. plain http.NotFound text).
        message = trimmed;
        parsed = rawBody;
      }
    }

    return ApiException(
      statusCode: statusCode,
      message: message,
      rawBody: parsed,
      method: method,
      uri: uri,
    );
  }

  String _defaultMessageForStatus(int statusCode) {
    switch (statusCode) {
      case 401:
        return 'Session invalid or expired';
      case 403:
        return 'Forbidden';
      case 404:
        return 'Not found';
      case 409:
        return 'Conflict or feature disabled';
      case 429:
        return 'Rate limit exceeded';
      default:
        if (statusCode >= 500) {
          return 'Server error';
        }
        return 'Request failed';
    }
  }
}
