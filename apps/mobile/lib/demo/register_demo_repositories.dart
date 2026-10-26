import 'package:sophia_ai/core/di/service_locator.dart';
import 'package:sophia_core/features/actions/domain/action_proposals_repository.dart';
import 'package:sophia_core/features/auth/domain/auth_repository.dart';
import 'package:sophia_core/features/beliefs/domain/beliefs_repository.dart';
import 'package:sophia_core/features/chat/domain/conversations_repository.dart';
import 'package:sophia_core/features/contexts/domain/contexts_repository.dart';
import 'package:sophia_core/features/reminders/domain/reminders_repository.dart';
import 'package:sophia_core/features/settings/domain/settings_services_repository.dart';
import 'package:sophia_core/features/system/domain/system_repository.dart';
import 'package:sophia_core/features/users/domain/user_repository.dart';
import 'package:sophia_core/demo/demo_repositories.dart';
import 'package:sophia_core/demo/demo_store.dart';

void registerDemoRepositories() {
  final store = DemoStore.instance;
  sl.registerSingleton<ConversationsRepository>(
    DemoConversationsRepository(store),
  );
  sl.registerSingleton<ActionProposalsRepository>(
    DemoActionProposalsRepository(store),
  );
  sl.registerSingleton<RemindersRepository>(DemoRemindersRepository(store));
  sl.registerSingleton<BeliefsRepository>(DemoBeliefsRepository(store));
  sl.registerSingleton<ContextsRepository>(DemoContextsRepository());
  sl.registerSingleton<UserRepository>(DemoUserRepository(store));
  sl.registerSingleton<AuthRepository>(DemoAuthRepository(store));
  sl.registerSingleton<SystemRepository>(DemoSystemRepository());
  sl.registerSingleton<SettingsServicesRepository>(DemoSettingsRepository());
}
