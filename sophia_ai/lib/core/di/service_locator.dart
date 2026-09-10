import 'package:get_it/get_it.dart';
import 'package:http/http.dart' as http;

import '../../features/actions/data/action_proposals_repository_impl.dart';
import '../../features/actions/domain/action_proposals_repository.dart';
import '../../features/auth/data/auth_repository_impl.dart';
import '../../features/auth/domain/auth_repository.dart';
import '../../features/auth/presentation/cubit/auth_cubit.dart';
import '../../features/beliefs/data/beliefs_repository_impl.dart';
import '../../features/beliefs/domain/beliefs_repository.dart';
import '../../features/beliefs/presentation/cubit/beliefs_cubit.dart';
import '../../features/chat/data/conversations_repository_impl.dart';
import '../../features/chat/domain/conversations_repository.dart';
import '../../features/chat/presentation/cubit/chat_message_cubit.dart';
import '../../features/contexts/data/contexts_repository_impl.dart';
import '../../features/contexts/domain/contexts_repository.dart';
import '../../features/contexts/presentation/cubit/contexts_cubit.dart';
import '../../features/dashboard/presentation/cubit/dashboard_cubit.dart';
import '../../features/session/presentation/cubit/session_cubit.dart';
import '../../features/settings/presentation/cubit/settings_cubit.dart';
import '../../features/settings/data/settings_services_repository_impl.dart';
import '../../features/settings/domain/settings_services_repository.dart';
import '../../features/settings/presentation/cubit/settings_services_cubit.dart';
import '../../features/system/data/system_repository_impl.dart';
import '../../features/system/domain/system_repository.dart';
import '../../features/system/presentation/cubit/health_cubit.dart';
import '../../features/system/presentation/cubit/learning_diagnostics_cubit.dart';
import '../../features/users/data/user_repository_impl.dart';
import '../../features/users/domain/user_repository.dart';
import '../../features/users/presentation/cubit/ai_settings_cubit.dart';
import '../../features/users/presentation/cubit/onboarding_cubit.dart';
import '../../features/users/presentation/cubit/user_profile_cubit.dart';
import '../../features/reminders/data/reminders_repository_impl.dart';
import '../../features/reminders/domain/reminders_repository.dart';
import '../../features/reminders/presentation/cubit/reminders_cubit.dart';
import '../config/app_config.dart';
import '../config/feature_flags.dart';
import '../network/api_client.dart';
import '../storage/token_storage.dart';

final sl = GetIt.instance;

Future<void> initDependencies({
  AppConfig? config,
  FeatureFlags? featureFlags,
  TokenStorage? tokenStorage,
  http.Client? httpClient,
}) async {
  final appConfig = config ?? AppConfig.fromEnvironment();
  final flags = featureFlags ?? FeatureFlags.defaults;
  final tokens = tokenStorage ?? SecureTokenStorage();

  if (!sl.isRegistered<AppConfig>()) {
    sl.registerSingleton<AppConfig>(appConfig);
  }
  if (!sl.isRegistered<FeatureFlags>()) {
    sl.registerSingleton<FeatureFlags>(flags);
  }
  if (!sl.isRegistered<TokenStorage>()) {
    sl.registerSingleton<TokenStorage>(tokens);
  }
  if (!sl.isRegistered<ApiClient>()) {
    sl.registerLazySingleton<ApiClient>(
      () => ApiClient(
        config: sl<AppConfig>(),
        tokenStorage: sl<TokenStorage>(),
        httpClient: httpClient,
      ),
    );
  }

  if (!sl.isRegistered<SystemRepository>()) {
    sl.registerLazySingleton<SystemRepository>(
      () => SystemRepositoryImpl(sl<ApiClient>()),
    );
  }
  if (!sl.isRegistered<AuthRepository>()) {
    sl.registerLazySingleton<AuthRepository>(
      () => AuthRepositoryImpl(
        api: sl<ApiClient>(),
        tokenStorage: sl<TokenStorage>(),
      ),
    );
  }
  if (!sl.isRegistered<UserRepository>()) {
    sl.registerLazySingleton<UserRepository>(
      () => UserRepositoryImpl(sl<ApiClient>()),
    );
  }
  if (!sl.isRegistered<ConversationsRepository>()) {
    sl.registerLazySingleton<ConversationsRepository>(
      () => ConversationsRepositoryImpl(sl<ApiClient>()),
    );
  }
  if (!sl.isRegistered<ActionProposalsRepository>()) {
    sl.registerLazySingleton<ActionProposalsRepository>(
      () => ActionProposalsRepositoryImpl(sl<ApiClient>()),
    );
  }
  if (!sl.isRegistered<RemindersRepository>()) {
    sl.registerLazySingleton<RemindersRepository>(
      () => RemindersRepositoryImpl(sl<ApiClient>()),
    );
  }
  if (!sl.isRegistered<BeliefsRepository>()) {
    sl.registerLazySingleton<BeliefsRepository>(
      () => BeliefsRepositoryImpl(sl<ApiClient>()),
    );
  }
  if (!sl.isRegistered<ContextsRepository>()) {
    sl.registerLazySingleton<ContextsRepository>(
      () => ContextsRepositoryImpl(sl<ApiClient>()),
    );
  }
  if (!sl.isRegistered<SettingsServicesRepository>()) {
    sl.registerLazySingleton<SettingsServicesRepository>(
      () => SettingsServicesRepositoryImpl(sl<ApiClient>()),
    );
  }

  // Session is the global auth source of truth (singleton).
  if (!sl.isRegistered<SessionCubit>()) {
    sl.registerLazySingleton(
      () => SessionCubit(
        tokenStorage: sl<TokenStorage>(),
        userRepository: sl<UserRepository>(),
        authRepository: sl<AuthRepository>(),
      ),
    );
  }

  if (!sl.isRegistered<AuthCubit>()) {
    sl.registerFactory(
      () => AuthCubit(
        authRepository: sl<AuthRepository>(),
        sessionCubit: sl<SessionCubit>(),
      ),
    );
  }
  if (!sl.isRegistered<UserProfileCubit>()) {
    sl.registerFactory(
      () => UserProfileCubit(
        userRepository: sl<UserRepository>(),
        sessionCubit: sl<SessionCubit>(),
      ),
    );
  }
  if (!sl.isRegistered<AiSettingsCubit>()) {
    sl.registerFactory(
      () => AiSettingsCubit(
        userRepository: sl<UserRepository>(),
        sessionCubit: sl<SessionCubit>(),
      ),
    );
  }
  if (!sl.isRegistered<OnboardingCubit>()) {
    sl.registerFactory(
      () => OnboardingCubit(
        userRepository: sl<UserRepository>(),
        sessionCubit: sl<SessionCubit>(),
      ),
    );
  }
  if (!sl.isRegistered<HealthCubit>()) {
    sl.registerFactory(
      () => HealthCubit(systemRepository: sl<SystemRepository>()),
    );
  }
  if (!sl.isRegistered<ChatMessageCubit>()) {
    sl.registerFactory(
      () => ChatMessageCubit(repository: sl<ConversationsRepository>()),
    );
  }
  if (!sl.isRegistered<RemindersCubit>()) {
    sl.registerFactory(
      () => RemindersCubit(repository: sl<RemindersRepository>()),
    );
  }
  if (!sl.isRegistered<BeliefsCubit>()) {
    sl.registerFactory(() => BeliefsCubit(sl<BeliefsRepository>()));
  }
  if (!sl.isRegistered<ContextsCubit>()) {
    sl.registerFactory(() => ContextsCubit(sl<ContextsRepository>()));
  }
  if (!sl.isRegistered<LearningDiagnosticsCubit>()) {
    sl.registerFactory(() => LearningDiagnosticsCubit(sl<BeliefsRepository>()));
  }
  sl.registerFactory(() => PrivacyCubit(sl<SettingsServicesRepository>()));
  sl.registerFactory(
    () => NotificationsCubit(sl<SettingsServicesRepository>()),
  );
  sl.registerFactory(() => IngestionCubit(sl<SettingsServicesRepository>()));

  // Legacy UI prototype cubits (not backend-backed).
  if (!sl.isRegistered<DashboardCubit>()) {
    sl.registerFactory(() => DashboardCubit());
  }
  if (!sl.isRegistered<SettingsCubit>()) {
    sl.registerFactory(() => SettingsCubit());
  }
}
