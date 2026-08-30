import 'package:flutter_test/flutter_test.dart';
import 'package:sophia_ai/core/models/models.dart';
import 'package:sophia_ai/core/network/api_exception.dart';
import 'package:sophia_ai/core/storage/token_storage.dart';
import 'package:sophia_ai/features/auth/domain/auth_repository.dart';
import 'package:sophia_ai/features/auth/presentation/cubit/auth_cubit.dart';
import 'package:sophia_ai/features/auth/presentation/cubit/auth_state.dart';
import 'package:sophia_ai/features/session/presentation/cubit/session_cubit.dart';
import 'package:sophia_ai/features/session/presentation/cubit/session_state.dart';
import 'package:sophia_ai/features/system/domain/system_repository.dart';
import 'package:sophia_ai/features/system/presentation/cubit/health_cubit.dart';
import 'package:sophia_ai/features/system/presentation/cubit/health_state.dart';
import 'package:sophia_ai/features/users/domain/user_repository.dart';
import 'package:sophia_ai/features/users/presentation/cubit/ai_settings_cubit.dart';
import 'package:sophia_ai/features/users/presentation/cubit/ai_settings_state.dart';
import 'package:sophia_ai/features/users/presentation/cubit/onboarding_cubit.dart';
import 'package:sophia_ai/features/users/presentation/cubit/onboarding_state.dart';
import 'package:sophia_ai/features/users/presentation/cubit/user_profile_cubit.dart';
import 'package:sophia_ai/features/users/presentation/cubit/user_profile_state.dart';

final _createdAt = DateTime.utc(2026, 7, 10, 18);
final _updatedAt = DateTime.utc(2026, 7, 11, 9, 30);

AuthUser sampleUser({String? role}) => AuthUser(
  id: 'usr_1',
  name: 'Armando',
  email: 'armando@example.com',
  role: role,
);

UserProfile sampleProfile({
  bool onboardingCompleted = false,
  String displayName = 'Armando',
}) {
  return UserProfile(
    id: 'usr_1',
    userId: 'usr_1',
    displayName: displayName,
    preferredName: 'Armando',
    timezone: 'America/Tijuana',
    locale: 'es-MX',
    onboardingCompleted: onboardingCompleted,
    createdAt: _createdAt,
    updatedAt: _updatedAt,
  );
}

AiSettings sampleSettings({
  bool memoryEnabled = true,
  bool remindersEnabled = true,
  bool planningEnabled = true,
}) {
  return AiSettings(
    userId: 'usr_1',
    assistantName: 'Sofía',
    communicationStyle: 'warm',
    proactivityLevel: ProactivityLevel.medium,
    autonomyLevel: AutonomyLevel.suggestive,
    memoryEnabled: memoryEnabled,
    remindersEnabled: remindersEnabled,
    planningEnabled: planningEnabled,
    preferredLanguage: 'es',
    createdAt: _createdAt,
    updatedAt: _updatedAt,
  );
}

MeResponse sampleMe({UserProfile? profile, AiSettings? settings}) {
  return MeResponse(
    user: sampleUser(role: 'user'),
    profile: profile ?? sampleProfile(),
    aiSettings: settings ?? sampleSettings(),
  );
}

class FakeAuthRepository implements AuthRepository {
  FakeAuthRepository({this.tokenStorage});

  final MemoryTokenStorage? tokenStorage;
  Object? registerError;
  Object? loginError;
  AuthUser? registerResult;
  LoginResponse? loginResult;
  int logoutCalls = 0;

  @override
  Future<AuthUser> register(RegisterRequest request) async {
    if (registerError != null) throw registerError!;
    return registerResult ?? sampleUser();
  }

  @override
  Future<LoginResponse> login(LoginRequest request) async {
    if (loginError != null) throw loginError!;
    final response =
        loginResult ?? LoginResponse(token: 'jwt-valid', user: sampleUser());
    await tokenStorage?.saveToken(response.token);
    return response;
  }

  @override
  Future<void> logout() async {
    logoutCalls++;
    await tokenStorage?.clearToken();
  }
}

class FakeUserRepository implements UserRepository {
  Object? getMeError;
  MeResponse? me;
  Object? updateProfileError;
  UserProfile? updatedProfile;
  Object? getAiSettingsError;
  AiSettings? aiSettings;
  Object? updateAiSettingsError;
  AiSettings? updatedAiSettings;
  Object? completeOnboardingError;
  UserProfile? completedProfile;
  int completeCalls = 0;

  @override
  Future<MeResponse> getMe() async {
    if (getMeError != null) throw getMeError!;
    return me ?? sampleMe();
  }

  @override
  Future<UserProfile> updateProfile(UpdateProfileRequest request) async {
    if (updateProfileError != null) throw updateProfileError!;
    return updatedProfile ??
        sampleProfile(displayName: request.displayName ?? 'Armando');
  }

  @override
  Future<AiSettings> getAiSettings() async {
    if (getAiSettingsError != null) throw getAiSettingsError!;
    return aiSettings ?? sampleSettings();
  }

  @override
  Future<AiSettings> updateAiSettings(UpdateAiSettingsRequest request) async {
    if (updateAiSettingsError != null) throw updateAiSettingsError!;
    return updatedAiSettings ??
        sampleSettings(
          memoryEnabled: request.memoryEnabled ?? true,
          remindersEnabled: request.remindersEnabled ?? true,
          planningEnabled: request.planningEnabled ?? true,
        );
  }

  @override
  Future<UserProfile> completeOnboarding() async {
    completeCalls++;
    if (completeOnboardingError != null) throw completeOnboardingError!;
    return completedProfile ?? sampleProfile(onboardingCompleted: true);
  }
}

class FakeSystemRepository implements SystemRepository {
  Object? error;
  HealthResponse? health;

  @override
  Future<HealthResponse> getHealth() async {
    if (error != null) throw error!;
    return health ??
        const HealthResponse(
          status: 'ok',
          environment: 'local',
          firestore: 'ok',
        );
  }
}

void main() {
  late MemoryTokenStorage tokens;
  late FakeAuthRepository authRepo;
  late FakeUserRepository userRepo;
  late SessionCubit session;

  setUp(() {
    tokens = MemoryTokenStorage();
    authRepo = FakeAuthRepository(tokenStorage: tokens);
    userRepo = FakeUserRepository();
    session = SessionCubit(
      tokenStorage: tokens,
      userRepository: userRepo,
      authRepository: authRepo,
    );
  });

  tearDown(() async {
    await session.close();
  });

  group('SessionCubit', () {
    test('sin token → unauthenticated', () async {
      await session.bootstrap();
      expect(session.state, isA<SessionUnauthenticated>());
    });

    test('token válido + getMe → authenticated', () async {
      await tokens.saveToken('jwt');
      userRepo.me = sampleMe();
      await session.bootstrap();
      final state = session.state as SessionAuthenticated;
      expect(state.user.email, 'armando@example.com');
      expect(state.profile.displayName, 'Armando');
      expect(state.aiSettings.assistantName, 'Sofía');
    });

    test('token inválido 401 limpia storage', () async {
      await tokens.saveToken('bad-jwt');
      userRepo.getMeError = const ApiException(
        statusCode: 401,
        message: 'unauthorized',
      );
      await session.bootstrap();
      expect(session.state, isA<SessionUnauthenticated>());
      expect(await tokens.readToken(), isNull);
    });

    test('error 500 no borra token', () async {
      await tokens.saveToken('still-valid');
      userRepo.getMeError = const ApiException(
        statusCode: 500,
        message: 'boom',
      );
      await session.bootstrap();
      final failure = session.state as SessionFailure;
      expect(failure.tokenRetained, isTrue);
      expect(failure.message, 'boom');
      expect(await tokens.readToken(), 'still-valid');
    });

    test('logout limpia token y sesión', () async {
      await tokens.saveToken('jwt');
      userRepo.me = sampleMe();
      await session.bootstrap();
      await session.logout();
      expect(session.state, isA<SessionUnauthenticated>());
      expect(await tokens.readToken(), isNull);
      expect(authRepo.logoutCalls, 1);
    });
  });

  group('AuthCubit', () {
    late AuthCubit auth;

    setUp(() {
      auth = AuthCubit(authRepository: authRepo, sessionCubit: session);
    });

    tearDown(() async {
      await auth.close();
    });

    test('login success hidrata session', () async {
      userRepo.me = sampleMe();
      await auth.login(email: 'a@b.com', password: 'password1');
      expect(auth.state, isA<AuthLoginSuccess>());
      expect(session.state, isA<SessionAuthenticated>());
      expect(await tokens.readToken(), 'jwt-valid');
    });

    test('login invalid credentials', () async {
      authRepo.loginError = const ApiException(
        statusCode: 401,
        message: 'invalid credentials',
      );
      await auth.login(email: 'a@b.com', password: 'wrong');
      final failure = auth.state as AuthFailure;
      expect(failure.exception?.isUnauthorized, isTrue);
      expect(session.state, isA<SessionInitial>());
      expect(await tokens.readToken(), isNull);
    });

    test('login rate limited', () async {
      authRepo.loginError = const ApiException(
        statusCode: 429,
        message: 'too many requests',
      );
      await auth.login(email: 'a@b.com', password: 'password1');
      expect((auth.state as AuthFailure).exception?.isRateLimited, isTrue);
    });

    test('register success without auto-login', () async {
      await auth.register(
        name: 'Armando',
        email: 'a@b.com',
        password: 'password1',
      );
      expect(auth.state, isA<AuthRegisterSuccess>());
      expect(session.state, isA<SessionInitial>());
      expect(await tokens.readToken(), isNull);
    });

    test('register duplicate 409', () async {
      authRepo.registerError = const ApiException(
        statusCode: 409,
        message: 'email already registered',
      );
      await auth.register(name: 'A', email: 'a@b.com', password: 'password1');
      expect((auth.state as AuthFailure).exception?.isConflict, isTrue);
    });
  });

  group('UserProfileCubit', () {
    test('update success propaga session', () async {
      await tokens.saveToken('jwt');
      userRepo.me = sampleMe();
      await session.bootstrap();

      final cubit = UserProfileCubit(
        userRepository: userRepo,
        sessionCubit: session,
      );
      cubit.loadFromSession();
      userRepo.updatedProfile = sampleProfile(displayName: 'Nuevo');

      await cubit.update(const UpdateProfileRequest(displayName: 'Nuevo'));

      expect((cubit.state as UserProfileLoaded).profile.displayName, 'Nuevo');
      expect(
        (session.state as SessionAuthenticated).profile.displayName,
        'Nuevo',
      );
      await cubit.close();
    });

    test('update failure conserva estado anterior en session', () async {
      await tokens.saveToken('jwt');
      userRepo.me = sampleMe(profile: sampleProfile(displayName: 'Original'));
      await session.bootstrap();

      final cubit = UserProfileCubit(
        userRepository: userRepo,
        sessionCubit: session,
      );
      cubit.loadFromSession();
      userRepo.updateProfileError = const ApiException(
        statusCode: 400,
        message: 'invalid timezone',
      );

      await cubit.update(const UpdateProfileRequest(timezone: ''));

      final failure = cubit.state as UserProfileFailure;
      expect(failure.previous?.displayName, 'Original');
      expect(
        (session.state as SessionAuthenticated).profile.displayName,
        'Original',
      );
      await cubit.close();
    });
  });

  group('AiSettingsCubit', () {
    test('update success y toggle true→false en session', () async {
      await tokens.saveToken('jwt');
      userRepo.me = sampleMe(settings: sampleSettings(memoryEnabled: true));
      await session.bootstrap();

      final cubit = AiSettingsCubit(
        userRepository: userRepo,
        sessionCubit: session,
      );
      cubit.loadFromSession();
      expect((cubit.state as AiSettingsLoaded).memoryEnabled, isTrue);

      userRepo.updatedAiSettings = sampleSettings(memoryEnabled: false);
      await cubit.update(const UpdateAiSettingsRequest(memoryEnabled: false));

      final loaded = cubit.state as AiSettingsLoaded;
      expect(loaded.memoryEnabled, isFalse);
      expect(loaded.remindersEnabled, isTrue);
      expect(
        (session.state as SessionAuthenticated).aiSettings.memoryEnabled,
        isFalse,
      );
      await cubit.close();
    });
  });

  group('OnboardingCubit', () {
    test('profile + settings + complete success', () async {
      await tokens.saveToken('jwt');
      userRepo.me = sampleMe();
      await session.bootstrap();

      final cubit = OnboardingCubit(
        userRepository: userRepo,
        sessionCubit: session,
      );

      await cubit.run(
        profile: const UpdateProfileRequest(displayName: 'Armando'),
        aiSettings: const UpdateAiSettingsRequest(assistantName: 'Sofía'),
      );

      expect(cubit.state, isA<OnboardingCompleted>());
      expect(
        (session.state as SessionAuthenticated).profile.onboardingCompleted,
        isTrue,
      );
      expect(userRepo.completeCalls, 1);
      await cubit.close();
    });

    test('fallo en paso intermedio no marca completo', () async {
      await tokens.saveToken('jwt');
      userRepo.me = sampleMe();
      await session.bootstrap();

      final cubit = OnboardingCubit(
        userRepository: userRepo,
        sessionCubit: session,
      );
      userRepo.updateAiSettingsError = const ApiException(
        statusCode: 400,
        message: 'bad settings',
      );

      await cubit.run(
        profile: const UpdateProfileRequest(displayName: 'Armando'),
        aiSettings: const UpdateAiSettingsRequest(assistantName: 'X'),
      );

      final failure = cubit.state as OnboardingFailure;
      expect(failure.failedStep, OnboardingStep.aiSettings);
      expect(userRepo.completeCalls, 0);
      expect(
        (session.state as SessionAuthenticated).profile.onboardingCompleted,
        isFalse,
      );
      await cubit.close();
    });

    test('fallo en complete no emite OnboardingCompleted', () async {
      await tokens.saveToken('jwt');
      userRepo.me = sampleMe();
      await session.bootstrap();

      final cubit = OnboardingCubit(
        userRepository: userRepo,
        sessionCubit: session,
      );
      userRepo.completeOnboardingError = const ApiException(
        statusCode: 500,
        message: 'server',
      );

      await cubit.run();

      expect(cubit.state, isA<OnboardingFailure>());
      expect(cubit.state, isNot(isA<OnboardingCompleted>()));
      expect(
        (session.state as SessionAuthenticated).profile.onboardingCompleted,
        isFalse,
      );
      await cubit.close();
    });
  });

  group('HealthCubit', () {
    test('ok', () async {
      final system = FakeSystemRepository();
      final cubit = HealthCubit(systemRepository: system);
      await cubit.check();
      expect(cubit.state, isA<HealthHealthy>());
      await cubit.close();
    });

    test('degraded', () async {
      final system = FakeSystemRepository()
        ..health = const HealthResponse(
          status: 'degraded',
          environment: 'local',
          firestore: 'error',
        );
      final cubit = HealthCubit(systemRepository: system);
      await cubit.check();
      expect(cubit.state, isA<HealthDegraded>());
      await cubit.close();
    });

    test('network failure', () async {
      final system = FakeSystemRepository()
        ..error = const ApiException(
          statusCode: 0,
          message: 'connection failed',
        );
      final cubit = HealthCubit(systemRepository: system);
      await cubit.check();
      expect((cubit.state as HealthFailure).message, 'connection failed');
      await cubit.close();
    });
  });
}
