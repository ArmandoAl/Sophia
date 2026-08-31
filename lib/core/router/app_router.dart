import 'package:go_router/go_router.dart';
import 'package:sophia_ai/core/router/go_router_refresh_stream.dart';
import 'package:sophia_ai/core/router/session_redirect.dart';
import 'package:sophia_ai/core/widgets/main_wrapper.dart';
import 'package:sophia_ai/features/auth/presentation/pages/login_screen.dart';
import 'package:sophia_ai/features/auth/presentation/pages/register_screen.dart';
import 'package:sophia_ai/features/chat/presentation/pages/chat_page.dart';
import 'package:sophia_ai/features/session/presentation/cubit/session_cubit.dart';
import 'package:sophia_ai/features/session/presentation/pages/splash_session_screen.dart';
import 'package:sophia_ai/features/settings/presentation/pages/settings_page.dart';
import 'package:sophia_ai/features/system/presentation/pages/diagnostics_screen.dart';
import 'package:sophia_ai/features/users/presentation/pages/assistant_settings_screen.dart';
import 'package:sophia_ai/features/users/presentation/pages/onboarding_flow_screen.dart';
import 'package:sophia_ai/features/users/presentation/pages/profile_screen.dart';
import 'package:sophia_ai/features/reminders/presentation/pages/reminders_page.dart';

class AppRouter {
  AppRouter._();

  static GoRouter create({required SessionCubit sessionCubit}) {
    return GoRouter(
      initialLocation: '/splash',
      refreshListenable: GoRouterRefreshStream(sessionCubit.stream),
      redirect: (context, state) => resolveSessionRedirect(
        session: sessionCubit.state,
        location: state.matchedLocation,
      ),
      routes: [
        GoRoute(
          path: '/splash',
          name: 'splash',
          builder: (context, state) => const SplashSessionScreen(),
        ),
        GoRoute(
          path: '/login',
          name: 'login',
          builder: (context, state) => const LoginScreen(),
        ),
        GoRoute(
          path: '/register',
          name: 'register',
          builder: (context, state) => const RegisterScreen(),
        ),
        GoRoute(
          path: '/onboarding',
          name: 'onboarding',
          builder: (context, state) => const OnboardingFlowScreen(),
        ),
        GoRoute(
          path: '/profile',
          name: 'profile',
          builder: (context, state) => const ProfileScreen(),
        ),
        GoRoute(
          path: '/assistant-settings',
          name: 'assistant-settings',
          builder: (context, state) => const AssistantSettingsScreen(),
        ),
        GoRoute(
          path: '/reminders',
          name: 'reminders',
          builder: (context, state) => const RemindersPage(),
        ),
        StatefulShellRoute.indexedStack(
          builder: (context, state, navigationShell) {
            return MainWrapper(navigationShell: navigationShell);
          },
          branches: [
            StatefulShellBranch(
              routes: [
                GoRoute(
                  path: '/chat',
                  name: 'chat',
                  builder: (context, state) => const ChatPage(),
                ),
              ],
            ),
            StatefulShellBranch(
              routes: [
                GoRoute(
                  path: '/dashboard',
                  name: 'dashboard',
                  builder: (context, state) => const DiagnosticsScreen(),
                ),
              ],
            ),
            StatefulShellBranch(
              routes: [
                GoRoute(
                  path: '/settings',
                  name: 'settings',
                  builder: (context, state) => const SettingsPage(),
                ),
              ],
            ),
          ],
        ),
      ],
    );
  }
}
