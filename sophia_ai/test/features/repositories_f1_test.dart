import 'dart:convert';
import 'dart:io';

import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:sophia_ai/core/config/app_config.dart';
import 'package:sophia_ai/core/models/models.dart';
import 'package:sophia_ai/core/network/api_client.dart';
import 'package:sophia_ai/core/network/api_exception.dart';
import 'package:sophia_ai/core/storage/token_storage.dart';
import 'package:sophia_ai/features/auth/data/auth_repository_impl.dart';
import 'package:sophia_ai/features/system/data/system_repository_impl.dart';
import 'package:sophia_ai/features/users/data/user_repository_impl.dart';

String fixture(String name) =>
    File('test/fixtures/f1/$name').readAsStringSync();

Map<String, dynamic> fixtureMap(String name) =>
    jsonDecode(fixture(name)) as Map<String, dynamic>;

void main() {
  late MemoryTokenStorage tokens;
  const config = AppConfig(baseUrl: 'http://localhost:8080');

  setUp(() {
    tokens = MemoryTokenStorage();
  });

  ApiClient clientFor(MockClient mock) {
    return ApiClient(config: config, tokenStorage: tokens, httpClient: mock);
  }

  group('SystemRepository', () {
    test('health parse correcto', () async {
      final repo = SystemRepositoryImpl(
        clientFor(
          MockClient(
            (_) async => http.Response(fixture('health_response.json'), 200),
          ),
        ),
      );

      final health = await repo.getHealth();
      expect(health.status, 'ok');
      expect(health.environment, 'local');
      expect(health.firestore, 'ok');
    });

    test('firestore disabled', () async {
      final repo = SystemRepositoryImpl(
        clientFor(
          MockClient(
            (_) async => http.Response(
              jsonEncode({
                'status': 'ok',
                'environment': 'local',
                'firestore': 'disabled',
              }),
              200,
            ),
          ),
        ),
      );

      final health = await repo.getHealth();
      expect(health.firestore, 'disabled');
      expect(health.isOk, isTrue);
    });

    test('degraded when firestore error', () async {
      final repo = SystemRepositoryImpl(
        clientFor(
          MockClient(
            (_) async => http.Response(
              jsonEncode({
                'status': 'degraded',
                'environment': 'local',
                'firestore': 'error',
              }),
              200,
            ),
          ),
        ),
      );

      final health = await repo.getHealth();
      expect(health.isDegraded, isTrue);
      expect(health.firestore, 'error');
    });
  });

  group('AuthRepository', () {
    test('register 201 returns AuthUser without token side-effect', () async {
      final repo = AuthRepositoryImpl(
        api: clientFor(
          MockClient(
            (_) async =>
                http.Response(fixture('auth_register_response.json'), 201),
          ),
        ),
        tokenStorage: tokens,
      );

      final user = await repo.register(
        const RegisterRequest(
          name: 'Armando',
          email: 'armando@example.com',
          password: 'password1',
        ),
      );

      expect(user.id, 'usr_reg_001');
      expect(user.role, isNull);
      expect(await tokens.readToken(), isNull);
    });

    test('register 409 propagates ApiException', () async {
      final repo = AuthRepositoryImpl(
        api: clientFor(
          MockClient(
            (_) async => http.Response(
              jsonEncode({'message': 'email already registered'}),
              409,
            ),
          ),
        ),
        tokenStorage: tokens,
      );

      await expectLater(
        repo.register(
          const RegisterRequest(
            name: 'A',
            email: 'a@b.com',
            password: 'password1',
          ),
        ),
        throwsA(
          isA<ApiException>()
              .having((e) => e.statusCode, 'status', 409)
              .having((e) => e.isConflict, 'conflict', true)
              .having((e) => e.message, 'message', 'email already registered'),
        ),
      );
    });

    test('register 429 propagates rate limit', () async {
      final repo = AuthRepositoryImpl(
        api: clientFor(
          MockClient(
            (_) async => http.Response(
              jsonEncode({'message': 'too many requests'}),
              429,
            ),
          ),
        ),
        tokenStorage: tokens,
      );

      await expectLater(
        repo.register(
          const RegisterRequest(
            name: 'A',
            email: 'a@b.com',
            password: 'password1',
          ),
        ),
        throwsA(
          isA<ApiException>().having((e) => e.isRateLimited, '429', true),
        ),
      );
    });

    test('login 200 guarda token', () async {
      final repo = AuthRepositoryImpl(
        api: clientFor(
          MockClient(
            (_) async =>
                http.Response(fixture('auth_login_response.json'), 200),
          ),
        ),
        tokenStorage: tokens,
      );

      final login = await repo.login(
        const LoginRequest(email: 'armando@example.com', password: 'password1'),
      );

      expect(login.user.email, 'armando@example.com');
      expect(await tokens.readToken(), login.token);
      expect(await tokens.readToken(), startsWith('eyJ'));
    });

    test('login 401 no guarda token', () async {
      await tokens.saveToken('stale-token');
      final repo = AuthRepositoryImpl(
        api: clientFor(
          MockClient(
            (_) async => http.Response(
              jsonEncode({'message': 'invalid credentials'}),
              401,
            ),
          ),
        ),
        tokenStorage: tokens,
      );

      await expectLater(
        repo.login(const LoginRequest(email: 'a@b.com', password: 'wrongpass')),
        throwsA(
          isA<ApiException>()
              .having((e) => e.isUnauthorized, '401', true)
              .having((e) => e.message, 'message', 'invalid credentials'),
        ),
      );
      expect(await tokens.readToken(), 'stale-token');
    });

    test('logout limpia token', () async {
      await tokens.saveToken('jwt-to-clear');
      final repo = AuthRepositoryImpl(
        api: clientFor(MockClient((_) async => http.Response('{}', 200))),
        tokenStorage: tokens,
      );

      await repo.logout();
      expect(await tokens.readToken(), isNull);
    });
  });

  group('UserRepository', () {
    test('getMe parsea wrapper correcto', () async {
      await tokens.saveToken('jwt-test');
      final repo = UserRepositoryImpl(
        clientFor(
          MockClient(
            (_) async => http.Response(fixture('users_me_response.json'), 200),
          ),
        ),
      );

      final me = await repo.getMe();
      expect(me.user.role, 'user');
      expect(me.profile.displayName, 'Armando');
      expect(me.aiSettings.assistantName, 'Sofía');
    });

    test('update profile envía body plano y omite nulls', () async {
      await tokens.saveToken('jwt-test');
      http.BaseRequest? captured;
      final profileJson = fixtureMap('profile_with_avatar.json');

      final repo = UserRepositoryImpl(
        clientFor(
          MockClient((request) async {
            captured = request;
            return http.Response(jsonEncode(profileJson), 200);
          }),
        ),
      );

      final updated = await repo.updateProfile(
        const UpdateProfileRequest(
          displayName: 'Nuevo',
          preferredName: null,
          timezone: 'America/Mexico_City',
        ),
      );

      expect(captured!.method, 'PATCH');
      expect(captured!.url.path, '/users/me/profile');
      final sent = jsonDecode((captured as http.Request).body) as Map;
      expect(sent, {
        'display_name': 'Nuevo',
        'timezone': 'America/Mexico_City',
      });
      expect(sent.containsKey('user_id'), isFalse);
      expect(sent.containsKey('profile'), isFalse);
      expect(updated.displayName, 'Armando');
    });

    test('get AI settings plano', () async {
      await tokens.saveToken('jwt-test');
      final repo = UserRepositoryImpl(
        clientFor(
          MockClient(
            (_) async =>
                http.Response(fixture('ai_settings_defaults.json'), 200),
          ),
        ),
      );

      final settings = await repo.getAiSettings();
      expect(settings.proactivityLevel, ProactivityLevel.medium);
      expect(settings.memoryEnabled, isTrue);
    });

    test('update AI settings permite booleans false', () async {
      await tokens.saveToken('jwt-test');
      http.BaseRequest? captured;
      final settingsJson = fixtureMap('ai_settings_defaults.json');
      settingsJson['memory_enabled'] = false;
      settingsJson['reminders_enabled'] = false;

      final repo = UserRepositoryImpl(
        clientFor(
          MockClient((request) async {
            captured = request;
            return http.Response(jsonEncode(settingsJson), 200);
          }),
        ),
      );

      final updated = await repo.updateAiSettings(
        const UpdateAiSettingsRequest(
          memoryEnabled: false,
          remindersEnabled: false,
        ),
      );

      final sent = jsonDecode((captured as http.Request).body) as Map;
      expect(sent['memory_enabled'], isFalse);
      expect(sent['reminders_enabled'], isFalse);
      expect(sent.containsKey('ai_settings'), isFalse);
      expect(sent.containsKey('user_id'), isFalse);
      expect(updated.memoryEnabled, isFalse);
      expect(updated.remindersEnabled, isFalse);
    });

    test('complete onboarding sin wrapper', () async {
      await tokens.saveToken('jwt-test');
      http.BaseRequest? captured;
      final profile = fixtureMap('profile_with_avatar.json');
      profile['onboarding_completed'] = true;

      final repo = UserRepositoryImpl(
        clientFor(
          MockClient((request) async {
            captured = request;
            return http.Response(jsonEncode(profile), 200);
          }),
        ),
      );

      final result = await repo.completeOnboarding();
      expect(captured!.method, 'POST');
      expect(captured!.url.path, '/users/me/onboarding/complete');
      final sent = jsonDecode((captured as http.Request).body) as Map;
      expect(sent, isEmpty);
      expect(result.onboardingCompleted, isTrue);
      expect(result.toJson().containsKey('profile'), isFalse);
    });

    test('bearer token se envía automáticamente', () async {
      await tokens.saveToken('secret-jwt');
      http.BaseRequest? captured;

      final repo = UserRepositoryImpl(
        clientFor(
          MockClient((request) async {
            captured = request;
            return http.Response(fixture('users_me_response.json'), 200);
          }),
        ),
      );

      await repo.getMe();
      expect(captured!.headers['Authorization'], 'Bearer secret-jwt');
    });
  });
}
