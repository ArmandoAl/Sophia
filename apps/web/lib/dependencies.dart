import 'package:sophia_core/core/config/app_config.dart';
import 'package:sophia_core/core/network/api_client.dart';
import 'package:sophia_core/core/storage/token_storage.dart';
import 'package:sophia_core/demo/demo_repositories.dart';
import 'package:sophia_core/demo/demo_store.dart';
import 'package:sophia_core/features/auth/data/auth_repository_impl.dart';
import 'package:sophia_core/features/auth/domain/auth_repository.dart';
import 'package:sophia_core/features/auth/presentation/cubit/auth_cubit.dart';
import 'package:sophia_core/features/chat/data/conversations_repository_impl.dart';
import 'package:sophia_core/features/chat/domain/conversations_repository.dart';
import 'package:sophia_core/features/chat/presentation/cubit/chat_message_cubit.dart';
import 'package:sophia_core/features/session/presentation/cubit/session_cubit.dart';
import 'package:sophia_core/features/users/data/user_repository_impl.dart';
import 'package:sophia_core/features/users/domain/user_repository.dart';

import 'config.dart';
import 'platform/browser_token_storage.dart';

/// Raíz de composición de la web: mismos cubits y repositorios que la app móvil, con almacenamiento del navegador.
class Dependencies {
  Dependencies._({
    required this.tokenStorage,
    required this.session,
    required this.auth,
    required this.conversations,
    required this.demo,
    this.baseUrl,
  });

  final TokenStorage tokenStorage;
  final SessionCubit session;
  final AuthCubit auth;
  final ConversationsRepository conversations;
  final bool demo;
  final String? baseUrl;

  ChatMessageCubit newChat() => ChatMessageCubit(repository: conversations);

  factory Dependencies.create() {
    final TokenStorage tokens = BrowserTokenStorage();
    late final AuthRepository authRepository;
    late final UserRepository userRepository;
    late final ConversationsRepository conversations;
    String? baseUrl;
    if (kDemoMode) {
      final store = DemoStore.instance;
      authRepository = DemoAuthRepository(store);
      userRepository = DemoUserRepository(store);
      conversations = DemoConversationsRepository(store);
    } else {
      final config = AppConfig.fromEnvironment();
      baseUrl = config.baseUrl;
      final api = ApiClient(config: config, tokenStorage: tokens);
      authRepository = AuthRepositoryImpl(api: api, tokenStorage: tokens);
      userRepository = UserRepositoryImpl(api);
      conversations = ConversationsRepositoryImpl(api);
    }
    final session = SessionCubit(tokenStorage: tokens, userRepository: userRepository, authRepository: authRepository);
    final auth = AuthCubit(authRepository: authRepository, sessionCubit: session);
    return Dependencies._(
      tokenStorage: tokens,
      session: session,
      auth: auth,
      conversations: conversations,
      demo: kDemoMode,
      baseUrl: baseUrl,
    );
  }
}
