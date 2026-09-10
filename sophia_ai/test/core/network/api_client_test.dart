import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:sophia_ai/core/config/app_config.dart';
import 'package:sophia_ai/core/config/feature_flags.dart';
import 'package:sophia_ai/core/network/api_client.dart';
import 'package:sophia_ai/core/network/api_exception.dart';
import 'package:sophia_ai/core/storage/token_storage.dart';

void main() {
  group('AppConfig', () {
    test('default baseUrl is localhost:8080 without trailing slash', () {
      final config = AppConfig(
        baseUrl: AppConfig.normalizeBaseUrl(AppConfig.defaultBaseUrl),
      );
      expect(config.baseUrl, 'http://localhost:8080');
    });

    test('fromEnvironment uses default when dart-define is empty', () {
      final config = AppConfig.fromEnvironment();
      expect(config.baseUrl, 'http://localhost:8080');
    });

    test('normalizeBaseUrl strips trailing slashes', () {
      expect(
        AppConfig.normalizeBaseUrl('http://localhost:8080/'),
        'http://localhost:8080',
      );
      expect(
        AppConfig.normalizeBaseUrl('http://localhost:8080///'),
        'http://localhost:8080',
      );
    });

    test('resolve joins path without double slash', () {
      final config = AppConfig(baseUrl: 'http://localhost:8080');
      expect(
        config.resolve('/health').toString(),
        'http://localhost:8080/health',
      );
      expect(
        config.resolve('health').toString(),
        'http://localhost:8080/health',
      );
    });

    test('constructor override acts as dart-define substitute in tests', () {
      final config = AppConfig(
        baseUrl: AppConfig.normalizeBaseUrl('http://127.0.0.1:9090/'),
      );
      expect(config.baseUrl, 'http://127.0.0.1:9090');
    });
  });

  group('FeatureFlags', () {
    test('defaults enable beta action execution only', () {
      const flags = FeatureFlags.defaults;
      expect(flags.aiRuntimeEnabled, isFalse);
      expect(flags.aiActionExecutionEnabled, isTrue);
      expect(flags.notificationsEnabled, isFalse);
      expect(flags.smartHomeEnabled, isFalse);
    });
  });

  group('MemoryTokenStorage', () {
    test('read/save/clear round-trip', () async {
      final storage = MemoryTokenStorage();
      expect(await storage.readToken(), isNull);
      await storage.saveToken('jwt-abc');
      expect(await storage.readToken(), 'jwt-abc');
      await storage.clearToken();
      expect(await storage.readToken(), isNull);
    });
  });

  group('ApiClient', () {
    late MemoryTokenStorage tokens;

    setUp(() {
      tokens = MemoryTokenStorage();
    });

    ApiClient buildClient(MockClient mock) {
      return ApiClient(
        config: const AppConfig(baseUrl: 'http://localhost:8080'),
        tokenStorage: tokens,
        httpClient: mock,
      );
    }

    test('attaches Bearer header when token exists', () async {
      http.BaseRequest? captured;
      final client = buildClient(
        MockClient((request) async {
          captured = request;
          return http.Response(
            '{"ok":true}',
            200,
            headers: {'content-type': 'application/json'},
          );
        }),
      );

      await tokens.saveToken('secret-token');
      await client.get('/users/me');

      expect(captured!.headers['Authorization'], 'Bearer secret-token');
      expect(captured!.headers['Content-Type'], contains('application/json'));
    });

    test('omits Authorization when no token', () async {
      http.BaseRequest? captured;
      final client = buildClient(
        MockClient((request) async {
          captured = request;
          return http.Response('{"status":"ok"}', 200);
        }),
      );

      await client.get('/health', authenticated: false);

      expect(captured!.headers.containsKey('Authorization'), isFalse);
    });

    test('parses JSON object root', () async {
      final client = buildClient(
        MockClient(
          (_) async => http.Response('{"status":"ok","version":"0.1"}', 200),
        ),
      );

      final body =
          await client.get('/health', authenticated: false)
              as Map<String, dynamic>;
      expect(body['status'], 'ok');
      expect(body['version'], '0.1');
    });

    test('parses JSON array root', () async {
      final client = buildClient(
        MockClient(
          (_) async => http.Response(
            '[{"id":"1"},{"id":"2"}]',
            200,
            headers: {'content-type': 'application/json'},
          ),
        ),
      );

      final body =
          await client.get('/notifications/device-tokens') as List<dynamic>;
      expect(body, hasLength(2));
      expect((body[0] as Map)['id'], '1');
    });

    test('DELETE with empty body returns null', () async {
      final client = buildClient(
        MockClient((request) async {
          expect(request.method, 'DELETE');
          return http.Response('', 204);
        }),
      );

      final body = await client.delete('/notifications/device-tokens/abc');
      expect(body, isNull);
    });

    test('error JSON {"message":"..."} becomes ApiException.message', () async {
      final client = buildClient(
        MockClient(
          (_) async => http.Response('{"message":"invalid credentials"}', 401),
        ),
      );

      try {
        await client.post(
          '/auth/login',
          body: {'email': 'a', 'password': 'b'},
          authenticated: false,
        );
        fail('expected ApiException');
      } on ApiException catch (e) {
        expect(e.statusCode, 401);
        expect(e.message, 'invalid credentials');
        expect(e.isUnauthorized, isTrue);
        expect(e.rawBody, isA<Map>());
      }
    });

    test('non-JSON error body is tolerated', () async {
      final client = buildClient(
        MockClient((_) async => http.Response('404 page not found', 404)),
      );

      try {
        await client.get('/garbage');
        fail('expected ApiException');
      } on ApiException catch (e) {
        expect(e.statusCode, 404);
        expect(e.message, '404 page not found');
        expect(e.rawBody, '404 page not found');
      }
    });

    test('status 401 maps unauthorized', () async {
      final client = buildClient(
        MockClient(
          (_) async => http.Response('{"message":"unauthorized"}', 401),
        ),
      );

      await expectLater(
        client.get('/users/me'),
        throwsA(
          isA<ApiException>()
              .having((e) => e.statusCode, 'statusCode', 401)
              .having((e) => e.isUnauthorized, 'isUnauthorized', true),
        ),
      );
    });

    test('status 409 maps conflict / feature disabled', () async {
      final client = buildClient(
        MockClient(
          (_) async =>
              http.Response('{"message":"memory feature disabled"}', 409),
        ),
      );

      await expectLater(
        client.get('/memory'),
        throwsA(
          isA<ApiException>()
              .having((e) => e.statusCode, 'statusCode', 409)
              .having((e) => e.isConflict, 'isConflict', true)
              .having((e) => e.message, 'message', 'memory feature disabled'),
        ),
      );
    });

    test('status 429 maps rate limit', () async {
      final client = buildClient(
        MockClient(
          (_) async => http.Response('{"message":"too many requests"}', 429),
        ),
      );

      await expectLater(
        client.post('/auth/register', body: {}, authenticated: false),
        throwsA(
          isA<ApiException>()
              .having((e) => e.statusCode, 'statusCode', 429)
              .having((e) => e.isRateLimited, 'isRateLimited', true),
        ),
      );
    });

    test('supports POST PATCH DELETE methods', () async {
      final methods = <String>[];
      final client = buildClient(
        MockClient((request) async {
          methods.add(request.method);
          return http.Response('{"ok":true}', 200);
        }),
      );

      await client.post('/x', body: {'a': 1});
      await client.patch('/x', body: {'a': 2});
      await client.delete('/x');

      expect(methods, ['POST', 'PATCH', 'DELETE']);
    });
  });
}
