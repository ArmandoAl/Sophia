import 'dart:convert';
import 'dart:io';

import 'package:flutter_test/flutter_test.dart';
import 'package:sophia_ai/core/models/models.dart';

Map<String, dynamic> loadFixture(String name) {
  final file = File('test/fixtures/f1/$name');
  return jsonDecode(file.readAsStringSync()) as Map<String, dynamic>;
}

void main() {
  group('HealthResponse', () {
    test('parses flat health fixture', () {
      final health = HealthResponse.fromJson(
        loadFixture('health_response.json'),
      );
      expect(health.status, 'ok');
      expect(health.environment, 'local');
      expect(health.firestore, 'ok');
      expect(health.isOk, isTrue);
    });
  });

  group('AuthUser / Register', () {
    test('register response without role', () {
      final user = AuthUser.fromJson(
        loadFixture('auth_register_response.json'),
      );
      expect(user.id, 'usr_reg_001');
      expect(user.name, 'Armando');
      expect(user.email, 'armando@example.com');
      expect(user.role, isNull);
      expect(user.toJson().containsKey('role'), isFalse);
    });
  });

  group('LoginResponse', () {
    test('parses token + user without role', () {
      final login = LoginResponse.fromJson(
        loadFixture('auth_login_response.json'),
      );
      expect(login.token, startsWith('eyJ'));
      expect(login.user.id, 'usr_login_001');
      expect(login.user.role, isNull);
      expect(login.toJson()['user'], isA<Map<String, dynamic>>());
      expect(
        (login.toJson()['user'] as Map<String, dynamic>).containsKey('role'),
        isFalse,
      );
    });
  });

  group('UserProfile', () {
    test('profile without avatar_url', () {
      final profile = UserProfile.fromJson(
        loadFixture('profile_without_avatar.json'),
      );
      expect(profile.avatarUrl, isNull);
      expect(profile.onboardingCompleted, isFalse);
      expect(profile.timezone, 'America/Tijuana');
      expect(profile.createdAt.toUtc(), DateTime.utc(2026, 7, 10, 18));
      expect(profile.toJson().containsKey('avatar_url'), isFalse);
    });

    test('profile with avatar_url and RFC3339 fractions', () {
      final profile = UserProfile.fromJson(
        loadFixture('profile_with_avatar.json'),
      );
      expect(profile.avatarUrl, 'https://cdn.example.com/a.png');
      expect(profile.onboardingCompleted, isTrue);
      expect(profile.createdAt.toUtc().millisecond, 123);
      expect(profile.updatedAt.toUtc().millisecond, 456);
    });
  });

  group('AiSettings', () {
    test('defaults fixture', () {
      final settings = AiSettings.fromJson(
        loadFixture('ai_settings_defaults.json'),
      );
      expect(settings.assistantName, 'Sofía');
      expect(settings.communicationStyle, 'warm');
      expect(settings.preferredLanguage, 'es');
      expect(settings.proactivityLevel, ProactivityLevel.medium);
      expect(settings.autonomyLevel, AutonomyLevel.suggestive);
      expect(settings.memoryEnabled, isTrue);
      expect(settings.remindersEnabled, isTrue);
      expect(settings.planningEnabled, isTrue);
    });
  });

  group('MeResponse', () {
    test('parses nested user/profile/ai_settings wrapper', () {
      final me = MeResponse.fromJson(loadFixture('users_me_response.json'));
      expect(me.user.role, 'user');
      expect(me.profile.avatarUrl, isNotNull);
      expect(me.aiSettings.proactivityLevel, ProactivityLevel.high);
      expect(me.aiSettings.autonomyLevel, AutonomyLevel.semiAutonomous);
      expect(me.aiSettings.remindersEnabled, isFalse);

      final encoded = me.toJson();
      expect(encoded.keys, containsAll(['user', 'profile', 'ai_settings']));
      expect(encoded.containsKey('data'), isFalse);
    });
  });

  group('ProactivityLevel / AutonomyLevel', () {
    test('known proactivity values', () {
      expect(ProactivityLevel.fromJson('low'), ProactivityLevel.low);
      expect(ProactivityLevel.fromJson('medium'), ProactivityLevel.medium);
      expect(ProactivityLevel.fromJson('high'), ProactivityLevel.high);
      expect(ProactivityLevel.low.toJson(), 'low');
      expect(ProactivityLevel.medium.toJson(), 'medium');
      expect(ProactivityLevel.high.toJson(), 'high');
    });

    test('known autonomy values', () {
      expect(AutonomyLevel.fromJson('manual'), AutonomyLevel.manual);
      expect(AutonomyLevel.fromJson('suggestive'), AutonomyLevel.suggestive);
      expect(
        AutonomyLevel.fromJson('semi_autonomous'),
        AutonomyLevel.semiAutonomous,
      );
      expect(AutonomyLevel.semiAutonomous.toJson(), 'semi_autonomous');
    });

    test('unknown enum values keep raw and do not throw on parse', () {
      final p = ProactivityLevel.fromJson('ultra');
      expect(p.isKnown, isFalse);
      expect(p.value, 'ultra');
      expect(() => p.toJson(), throwsStateError);

      final a = AutonomyLevel.fromJson('full_auto');
      expect(a.isKnown, isFalse);
      expect(a.value, 'full_auto');
      expect(() => a.toJson(), throwsStateError);
    });

    test('AiSettings tolerates unknown enum on response parse', () {
      final json = loadFixture('ai_settings_defaults.json');
      json['proactivity_level'] = 'experimental';
      json['autonomy_level'] = 'chaos';
      final settings = AiSettings.fromJson(json);
      expect(settings.proactivityLevel.isKnown, isFalse);
      expect(settings.proactivityLevel.value, 'experimental');
      expect(settings.autonomyLevel.isKnown, isFalse);
      expect(settings.autonomyLevel.value, 'chaos');
    });
  });

  group('UpdateProfileRequest PATCH', () {
    test('omits nulls and never sends user_id', () {
      const req = UpdateProfileRequest(
        displayName: 'Nuevo',
        preferredName: null,
        timezone: 'America/Mexico_City',
      );
      final map = req.toJson();
      expect(map, {'display_name': 'Nuevo', 'timezone': 'America/Mexico_City'});
      expect(map.containsKey('preferred_name'), isFalse);
      expect(map.containsKey('locale'), isFalse);
      expect(map.containsKey('avatar_url'), isFalse);
      expect(map.containsKey('user_id'), isFalse);
      expect(map.containsKey('profile'), isFalse);
    });

    test('empty request yields empty map', () {
      expect(const UpdateProfileRequest().toJson(), isEmpty);
    });
  });

  group('UpdateAiSettingsRequest PATCH', () {
    test('allows explicit false and omits nulls', () {
      const req = UpdateAiSettingsRequest(
        memoryEnabled: false,
        remindersEnabled: false,
        planningEnabled: true,
        assistantName: null,
        proactivityLevel: ProactivityLevel.low,
      );
      final map = req.toJson();
      expect(map['memory_enabled'], isFalse);
      expect(map['reminders_enabled'], isFalse);
      expect(map['planning_enabled'], isTrue);
      expect(map['proactivity_level'], 'low');
      expect(map.containsKey('assistant_name'), isFalse);
      expect(map.containsKey('user_id'), isFalse);
      expect(map.containsKey('ai_settings'), isFalse);
    });

    test('rejects encoding unknown enum in request', () {
      final unknown = ProactivityLevel.fromJson('nope');
      final req = UpdateAiSettingsRequest(proactivityLevel: unknown);
      expect(req.toJson, throwsStateError);
    });
  });

  group('RegisterRequest / LoginRequest', () {
    test('serialize auth bodies', () {
      expect(
        const RegisterRequest(
          name: 'A',
          email: 'a@b.com',
          password: 'password1',
        ).toJson(),
        {'name': 'A', 'email': 'a@b.com', 'password': 'password1'},
      );
      expect(
        const LoginRequest(email: 'a@b.com', password: 'password1').toJson(),
        {'email': 'a@b.com', 'password': 'password1'},
      );
    });
  });
}
