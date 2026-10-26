import 'dart:async';

import 'package:jaspr/dom.dart';
import 'package:jaspr/jaspr.dart';
import 'package:sophia_core/features/auth/presentation/cubit/auth_state.dart';
import 'package:sophia_core/features/chat/presentation/cubit/chat_message_cubit.dart';
import 'package:sophia_core/features/session/presentation/cubit/session_state.dart';

import 'dependencies.dart';
import 'pages/chat_page.dart';
import 'pages/login_page.dart';

Component _t(String value) => Component.text(value);

/// Raíz de la interfaz web: puerta de sesión (SessionCubit del core) y conversación (ChatMessageCubit del core).
class App extends StatefulComponent {
  const App({super.key});

  @override
  State<App> createState() => _AppState();
}

class _AppState extends State<App> {
  late final Dependencies deps;
  ChatMessageCubit? _chat;
  final _subscriptions = <StreamSubscription<Object?>>[];

  @override
  void initState() {
    super.initState();
    deps = Dependencies.create();
    _subscriptions.add(deps.session.stream.listen(_onSession));
    _subscriptions.add(deps.auth.stream.listen((_) => _refresh()));
    deps.session.bootstrap();
  }

  @override
  void dispose() {
    for (final subscription in _subscriptions) {
      subscription.cancel();
    }
    _chat?.close();
    deps.session.close();
    deps.auth.close();
    super.dispose();
  }

  void _refresh() {
    if (mounted) setState(() {});
  }

  void _onSession(SessionState state) {
    if (state is SessionAuthenticated && _chat == null) {
      final chat = deps.newChat();
      _chat = chat;
      _subscriptions.add(chat.stream.listen((_) => _refresh()));
      chat.load();
    } else if (state is! SessionAuthenticated && state is! SessionLoading && _chat != null) {
      _chat!.close();
      _chat = null;
    }
    _refresh();
  }

  Future<void> _logout() async {
    await deps.session.logout();
    deps.auth.reset();
  }

  @override
  Component build(BuildContext context) {
    final session = deps.session.state;
    return div(classes: 'app', [
      if (deps.demo)
        p(classes: 'banner', [_t('Demo local · respuestas y acciones simuladas.')])
      else if (deps.baseUrl != null)
        p(classes: 'banner', [_t('API: ${deps.baseUrl}')]),
      switch (session) {
        SessionAuthenticated() when _chat != null => ChatPage(
          assistantName: session.aiSettings.assistantName,
          userName: session.profile.preferredName,
          state: _chat!.state,
          onSend: _chat!.sendMessage,
          onRetry: _chat!.retry,
          onLogout: _logout,
        ),
        SessionInitial() || SessionLoading() || SessionAuthenticated() => main_(id: 'contenido', [
          p(classes: 'hint', attributes: {'role': 'status'}, [_t('Cargando…')]),
        ]),
        SessionFailure() => main_(id: 'contenido', [
          p(classes: 'alert', attributes: {'role': 'alert'}, [
            span([_t(session.message)]),
            button([_t('Reintentar')], classes: 'btn', onClick: deps.session.bootstrap),
          ]),
        ]),
        SessionUnauthenticated() => LoginPage(
          demo: deps.demo,
          loading: deps.auth.state is AuthLoading,
          error: switch (deps.auth.state) {
            AuthFailure(:final message) => message,
            _ => null,
          },
          onSubmit: (email, password) => deps.auth.login(email: email, password: password),
        ),
      },
    ]);
  }
}
