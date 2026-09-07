import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:get_it/get_it.dart';
import 'package:sophia_ai/core/models/models.dart';
import 'package:sophia_ai/core/network/api_exception.dart';
import 'package:sophia_ai/core/router/session_redirect.dart';
import 'package:sophia_ai/core/storage/token_storage.dart';
import 'package:sophia_ai/core/theme/app_theme.dart';
import 'package:sophia_ai/features/auth/domain/auth_repository.dart';
import 'package:sophia_ai/features/auth/presentation/cubit/auth_cubit.dart';
import 'package:sophia_ai/features/auth/presentation/pages/login_screen.dart';
import 'package:sophia_ai/features/auth/presentation/pages/register_screen.dart';
import 'package:sophia_ai/features/session/presentation/cubit/session_cubit.dart';
import 'package:sophia_ai/features/session/presentation/cubit/session_state.dart';
import 'package:sophia_ai/features/session/presentation/pages/splash_session_screen.dart';
import 'package:sophia_ai/features/settings/presentation/pages/settings_page.dart';
import 'package:sophia_ai/features/system/domain/system_repository.dart';
import 'package:sophia_ai/features/system/presentation/cubit/health_cubit.dart';
import 'package:sophia_ai/features/system/presentation/pages/diagnostics_screen.dart';
import 'package:sophia_ai/features/users/domain/user_repository.dart';
import 'package:sophia_ai/features/users/presentation/cubit/ai_settings_cubit.dart';
import 'package:sophia_ai/features/users/presentation/cubit/onboarding_cubit.dart';
import 'package:sophia_ai/features/users/presentation/cubit/user_profile_cubit.dart';
import 'package:sophia_ai/features/users/presentation/pages/assistant_settings_screen.dart';
import 'package:sophia_ai/features/users/presentation/pages/onboarding_flow_screen.dart';
import 'package:sophia_ai/features/users/presentation/pages/profile_screen.dart';

final _t0 = DateTime.utc(2026, 7, 10, 18);
final _t1 = DateTime.utc(2026, 7, 11, 9, 30);

AuthUser _user() => const AuthUser(
  id: 'usr_1',
  name: 'Armando',
  email: 'armando@example.com',
  role: 'user',
);

UserProfile _profile({
  bool onboardingCompleted = false,
  String name = 'Armando',
}) {
  return UserProfile(
    id: 'usr_1',
    userId: 'usr_1',
    displayName: name,
    preferredName: name,
    timezone: 'America/Tijuana',
    locale: 'es-MX',
    onboardingCompleted: onboardingCompleted,
    createdAt: _t0,
    updatedAt: _t1,
  );
}

AiSettings _settings({bool memoryEnabled = true}) {
  return AiSettings(
    userId: 'usr_1',
    assistantName: 'Sofía',
    communicationStyle: 'warm',
    proactivityLevel: ProactivityLevel.medium,
    autonomyLevel: AutonomyLevel.suggestive,
    memoryEnabled: memoryEnabled,
    remindersEnabled: true,
    planningEnabled: true,
    preferredLanguage: 'es',
    createdAt: _t0,
    updatedAt: _t1,
  );
}

MeResponse _me({UserProfile? profile, AiSettings? settings}) => MeResponse(
  user: _user(),
  profile: profile ?? _profile(),
  aiSettings: settings ?? _settings(),
);

class _FakeAuth implements AuthRepository {
  _FakeAuth(this.tokens);
  final MemoryTokenStorage tokens;
  Object? loginError;
  Object? registerError;

  @override
  Future<LoginResponse> login(LoginRequest request) async {
    if (loginError != null) throw loginError!;
    await tokens.saveToken('jwt-test');
    return LoginResponse(token: 'jwt-test', user: _user());
  }

  @override
  Future<AuthUser> register(RegisterRequest request) async {
    if (registerError != null) throw registerError!;
    return _user();
  }

  @override
  Future<void> logout() => tokens.clearToken();
}

class _FakeUser implements UserRepository {
  MeResponse me = _me();
  UserProfile? updatedProfile;
  AiSettings? updatedSettings;

  @override
  Future<MeResponse> getMe() async => me;

  @override
  Future<UserProfile> updateProfile(UpdateProfileRequest request) async {
    final next =
        updatedProfile ??
        _profile(
          onboardingCompleted: me.profile.onboardingCompleted,
          name: request.displayName ?? me.profile.displayName,
        );
    me = MeResponse(user: me.user, profile: next, aiSettings: me.aiSettings);
    return next;
  }

  @override
  Future<AiSettings> getAiSettings() async => me.aiSettings;

  @override
  Future<AiSettings> updateAiSettings(UpdateAiSettingsRequest request) async {
    final next =
        updatedSettings ??
        _settings(
          memoryEnabled: request.memoryEnabled ?? me.aiSettings.memoryEnabled,
        );
    me = MeResponse(user: me.user, profile: me.profile, aiSettings: next);
    return next;
  }

  @override
  Future<UserProfile> completeOnboarding() async {
    final next = _profile(
      onboardingCompleted: true,
      name: me.profile.displayName,
    );
    me = MeResponse(user: me.user, profile: next, aiSettings: me.aiSettings);
    return next;
  }
}

class _FakeSystem implements SystemRepository {
  HealthResponse health = const HealthResponse(
    status: 'ok',
    environment: 'local',
    firestore: 'ok',
  );
  Object? error;

  @override
  Future<HealthResponse> getHealth() async {
    if (error != null) throw error!;
    return health;
  }
}

class _ThrowingUser extends _FakeUser {
  @override
  Future<MeResponse> getMe() async {
    throw const ApiException(statusCode: 500, message: 'boom');
  }
}

Future<SessionCubit> _wire({
  required MemoryTokenStorage tokens,
  required _FakeAuth auth,
  required _FakeUser user,
  required _FakeSystem system,
}) async {
  final sl = GetIt.instance;
  await sl.reset();
  sl.registerSingleton<TokenStorage>(tokens);
  sl.registerSingleton<AuthRepository>(auth);
  sl.registerSingleton<UserRepository>(user);
  sl.registerSingleton<SystemRepository>(system);
  sl.registerLazySingleton(
    () => SessionCubit(
      tokenStorage: tokens,
      userRepository: user,
      authRepository: auth,
    ),
  );
  sl.registerFactory(() => AuthCubit(authRepository: auth, sessionCubit: sl()));
  sl.registerFactory(
    () => UserProfileCubit(userRepository: user, sessionCubit: sl()),
  );
  sl.registerFactory(
    () => AiSettingsCubit(userRepository: user, sessionCubit: sl()),
  );
  sl.registerFactory(
    () => OnboardingCubit(userRepository: user, sessionCubit: sl()),
  );
  sl.registerFactory(() => HealthCubit(systemRepository: system));
  return sl<SessionCubit>();
}

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  late MemoryTokenStorage tokens;
  late _FakeAuth auth;
  late _FakeUser user;
  late _FakeSystem system;
  late SessionCubit session;

  setUp(() async {
    tokens = MemoryTokenStorage();
    auth = _FakeAuth(tokens);
    user = _FakeUser();
    system = _FakeSystem();
    session = await _wire(
      tokens: tokens,
      auth: auth,
      user: user,
      system: system,
    );
  });

  tearDown(() async {
    await session.close();
    await GetIt.instance.reset();
  });

  group('resolveSessionRedirect', () {
    test('sin token / unauthenticated → login', () {
      expect(
        resolveSessionRedirect(
          session: const SessionUnauthenticated(),
          location: '/splash',
        ),
        '/login',
      );
      expect(
        resolveSessionRedirect(
          session: const SessionUnauthenticated(),
          location: '/login',
        ),
        isNull,
      );
    });

    test('loading → splash', () {
      expect(
        resolveSessionRedirect(
          session: const SessionLoading(),
          location: '/chat',
        ),
        '/splash',
      );
    });

    test('authenticated incomplete → onboarding', () {
      final authState = SessionAuthenticated(
        user: _user(),
        profile: _profile(onboardingCompleted: false),
        aiSettings: _settings(),
      );
      expect(
        resolveSessionRedirect(session: authState, location: '/chat'),
        '/onboarding',
      );
      expect(
        resolveSessionRedirect(session: authState, location: '/onboarding'),
        isNull,
      );
    });

    test('authenticated complete → main from auth routes', () {
      final authState = SessionAuthenticated(
        user: _user(),
        profile: _profile(onboardingCompleted: true),
        aiSettings: _settings(),
      );
      expect(
        resolveSessionRedirect(session: authState, location: '/login'),
        '/chat',
      );
      expect(
        resolveSessionRedirect(session: authState, location: '/settings'),
        isNull,
      );
    });
  });

  testWidgets('splash shows loading then retry on failure', (tester) async {
    await tokens.saveToken('jwt');
    // Force recoverable bootstrap failure (token retained).
    user.me = _me();
    // Replace getMe via a throwing wrapper by temporarily using error path:
    final throwingUser = _ThrowingUser();
    await GetIt.instance.reset();
    session = await _wire(
      tokens: tokens,
      auth: auth,
      user: throwingUser,
      system: system,
    );

    await tester.pumpWidget(
      BlocProvider.value(
        value: session,
        child: MaterialApp(
          theme: AppTheme.darkTheme,
          home: const SplashSessionScreen(),
        ),
      ),
    );
    expect(find.text('Starting Sophia…'), findsOneWidget);

    await session.bootstrap();
    await tester.pump();
    expect(find.text('Retry'), findsOneWidget);
  });

  testWidgets('login error shows friendly 401 message', (tester) async {
    auth.loginError = const ApiException(
      statusCode: 401,
      message: 'invalid credentials',
    );
    await tester.pumpWidget(
      BlocProvider.value(
        value: session,
        child: MaterialApp(
          theme: AppTheme.darkTheme,
          home: const LoginScreen(),
        ),
      ),
    );
    await tester.enterText(find.byKey(const Key('login_email')), 'a@b.com');
    await tester.enterText(find.byKey(const Key('login_password')), 'bad');
    await tester.tap(find.byKey(const Key('login_submit')));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 50));
    expect(find.textContaining('Invalid email'), findsOneWidget);
  });

  testWidgets('login success hydrates session for onboarding redirect rule', (
    tester,
  ) async {
    user.me = _me(profile: _profile(onboardingCompleted: false));
    await tester.pumpWidget(
      BlocProvider.value(
        value: session,
        child: MaterialApp(
          theme: AppTheme.darkTheme,
          home: const LoginScreen(),
        ),
      ),
    );
    await tester.enterText(find.byKey(const Key('login_email')), 'a@b.com');
    await tester.enterText(
      find.byKey(const Key('login_password')),
      'password1',
    );
    await tester.tap(find.byKey(const Key('login_submit')));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 50));
    expect(session.state, isA<SessionAuthenticated>());
    expect(
      resolveSessionRedirect(session: session.state, location: '/login'),
      '/onboarding',
    );
  });

  testWidgets('login success with completed onboarding → /chat rule', (
    tester,
  ) async {
    user.me = _me(profile: _profile(onboardingCompleted: true));
    await tester.pumpWidget(
      BlocProvider.value(
        value: session,
        child: MaterialApp(
          theme: AppTheme.darkTheme,
          home: const LoginScreen(),
        ),
      ),
    );
    await tester.enterText(find.byKey(const Key('login_email')), 'a@b.com');
    await tester.enterText(
      find.byKey(const Key('login_password')),
      'password1',
    );
    await tester.tap(find.byKey(const Key('login_submit')));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 50));
    expect(
      resolveSessionRedirect(session: session.state, location: '/login'),
      '/chat',
    );
  });

  testWidgets('register success leaves no token', (tester) async {
    await tester.pumpWidget(
      BlocProvider.value(
        value: session,
        child: MaterialApp(
          theme: AppTheme.darkTheme,
          home: const RegisterScreen(),
        ),
      ),
    );
    await tester.enterText(find.byKey(const Key('register_name')), 'Armando');
    await tester.enterText(find.byKey(const Key('register_email')), 'a@b.com');
    await tester.enterText(
      find.byKey(const Key('register_password')),
      'password1',
    );
    await tester.enterText(
      find.byKey(const Key('register_confirm')),
      'password1',
    );
    await tester.tap(find.byKey(const Key('register_submit')));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 50));
    expect(await tokens.readToken(), isNull);
  });

  testWidgets('onboarding finish marks complete in session', (tester) async {
    session.applyMe(_me(profile: _profile(onboardingCompleted: false)));
    await tester.binding.setSurfaceSize(const Size(800, 1400));
    addTearDown(() => tester.binding.setSurfaceSize(null));
    await tester.pumpWidget(
      BlocProvider.value(
        value: session,
        child: MaterialApp(
          theme: AppTheme.darkTheme,
          home: const OnboardingFlowScreen(),
        ),
      ),
    );
    await tester.tap(find.byKey(const Key('onboarding_next')));
    await tester.pump();
    await tester.ensureVisible(find.byKey(const Key('onboarding_finish')));
    await tester.tap(find.byKey(const Key('onboarding_finish')));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 50));
    expect(
      (session.state as SessionAuthenticated).profile.onboardingCompleted,
      isTrue,
    );
    expect(
      resolveSessionRedirect(session: session.state, location: '/onboarding'),
      '/chat',
    );
  });

  testWidgets('profile update propagates to session', (tester) async {
    session.applyMe(_me(profile: _profile(onboardingCompleted: true)));
    user.updatedProfile = _profile(onboardingCompleted: true, name: 'Nuevo');
    await tester.pumpWidget(
      BlocProvider.value(
        value: session,
        child: MaterialApp(
          theme: AppTheme.darkTheme,
          home: const ProfileScreen(),
        ),
      ),
    );
    await tester.enterText(
      find.byKey(const Key('profile_display_name')),
      'Nuevo',
    );
    await tester.ensureVisible(find.byKey(const Key('profile_save')));
    await tester.tap(find.byKey(const Key('profile_save')));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 50));
    expect(
      (session.state as SessionAuthenticated).profile.displayName,
      'Nuevo',
    );
  });

  testWidgets('AI settings toggle memory false', (tester) async {
    session.applyMe(
      _me(
        profile: _profile(onboardingCompleted: true),
        settings: _settings(memoryEnabled: true),
      ),
    );
    user.updatedSettings = _settings(memoryEnabled: false);
    await tester.binding.setSurfaceSize(const Size(800, 1400));
    addTearDown(() => tester.binding.setSurfaceSize(null));
    await tester.pumpWidget(
      BlocProvider.value(
        value: session,
        child: MaterialApp(
          theme: AppTheme.darkTheme,
          home: const AssistantSettingsScreen(),
        ),
      ),
    );
    await tester.ensureVisible(find.byKey(const Key('ai_memory_switch')));
    await tester.tap(find.byKey(const Key('ai_memory_switch')));
    await tester.pump();
    await tester.ensureVisible(find.byKey(const Key('ai_settings_save')));
    await tester.tap(find.byKey(const Key('ai_settings_save')));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 50));
    expect(
      (session.state as SessionAuthenticated).aiSettings.memoryEnabled,
      isFalse,
    );
  });

  testWidgets('logout from settings clears token', (tester) async {
    await tokens.saveToken('jwt');
    session.applyMe(_me(profile: _profile(onboardingCompleted: true)));
    await tester.pumpWidget(
      BlocProvider.value(
        value: session,
        child: MaterialApp(
          theme: AppTheme.darkTheme,
          home: const SettingsPage(),
        ),
      ),
    );
    await tester.tap(find.byKey(const Key('settings_logout')));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 50));
    expect(session.state, isA<SessionUnauthenticated>());
    expect(await tokens.readToken(), isNull);
  });

  testWidgets('diagnostics healthy and degraded', (tester) async {
    await tester.pumpWidget(
      MaterialApp(theme: AppTheme.darkTheme, home: const DiagnosticsScreen()),
    );
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 50));
    expect(find.text('ok'), findsWidgets);

    system.health = const HealthResponse(
      status: 'degraded',
      environment: 'local',
      firestore: 'error',
    );
    await tester.tap(find.byKey(const Key('diagnostics_refresh')));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 50));
    expect(find.text('degraded'), findsOneWidget);
  });
}
