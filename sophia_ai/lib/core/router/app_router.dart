import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:sophia_ai/core/theme/design_tokens.dart';
import 'package:sophia_ai/core/theme/motion.dart';
import 'package:sophia_ai/core/router/go_router_refresh_stream.dart';
import 'package:sophia_ai/core/router/session_redirect.dart';
import 'package:sophia_ai/core/widgets/main_wrapper.dart';
import 'package:sophia_ai/core/widgets/motion/motion_widgets.dart';
import 'package:sophia_ai/features/auth/presentation/pages/login_screen.dart';
import 'package:sophia_ai/features/auth/presentation/pages/register_screen.dart';
import 'package:sophia_ai/features/beliefs/presentation/pages/beliefs_page.dart';
import 'package:sophia_ai/features/chat/presentation/pages/chat_page.dart';
import 'package:sophia_ai/features/contexts/presentation/pages/contexts_page.dart';
import 'package:sophia_ai/features/dashboard/presentation/pages/dashboard_page.dart';
import 'package:sophia_ai/features/session/presentation/cubit/session_cubit.dart';
import 'package:sophia_ai/features/session/presentation/pages/splash_session_screen.dart';
import 'package:sophia_ai/features/settings/presentation/pages/settings_page.dart';
import 'package:sophia_ai/features/settings/presentation/pages/privacy_data_page.dart';
import 'package:sophia_ai/features/settings/presentation/pages/notifications_page.dart';
import 'package:sophia_ai/features/settings/presentation/pages/import_conversations_page.dart';
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
          pageBuilder: (context, state) =>
              _detail(context, state, const SplashSessionScreen()),
        ),
        GoRoute(
          path: '/login',
          name: 'login',
          pageBuilder: (context, state) =>
              _detail(context, state, const LoginScreen()),
        ),
        GoRoute(
          path: '/register',
          name: 'register',
          pageBuilder: (context, state) =>
              _detail(context, state, const RegisterScreen()),
        ),
        GoRoute(
          path: '/onboarding',
          name: 'onboarding',
          pageBuilder: (context, state) =>
              _detail(context, state, const OnboardingFlowScreen()),
        ),
        GoRoute(
          path: '/profile',
          name: 'profile',
          pageBuilder: (context, state) =>
              _detail(context, state, const ProfileScreen()),
        ),
        GoRoute(
          path: '/assistant-settings',
          name: 'assistant-settings',
          pageBuilder: (context, state) =>
              _detail(context, state, const AssistantSettingsScreen()),
        ),
        GoRoute(
          path: '/reminders',
          name: 'reminders',
          pageBuilder: (context, state) =>
              _detail(context, state, const RemindersPage()),
        ),
        GoRoute(
          path: '/reminders/:id',
          name: 'reminder-detail',
          pageBuilder: (context, state) => _detail(
            context,
            state,
            ReminderDetailPage(id: state.pathParameters['id']!),
          ),
        ),
        GoRoute(
          path: '/beliefs',
          name: 'beliefs',
          pageBuilder: (context, state) =>
              _detail(context, state, const BeliefsPage()),
        ),
        GoRoute(
          path: '/contexts',
          name: 'contexts',
          pageBuilder: (context, state) =>
              _detail(context, state, const ContextsPage()),
        ),
        GoRoute(
          path: '/privacy-data',
          pageBuilder: (context, state) =>
              _detail(context, state, const PrivacyDataPage()),
        ),
        GoRoute(
          path: '/notifications',
          pageBuilder: (context, state) =>
              _detail(context, state, const NotificationsPage()),
        ),
        GoRoute(
          path: '/import-conversations',
          pageBuilder: (context, state) =>
              _detail(context, state, const ImportConversationsPage()),
        ),
        GoRoute(
          path: '/diagnostics',
          pageBuilder: (context, state) =>
              _detail(context, state, const DiagnosticsScreen()),
        ),
        StatefulShellRoute(
          navigatorContainerBuilder: (context, shell, children) =>
              LateralBranchContainer(
                currentIndex: shell.currentIndex,
                children: children,
              ),
          pageBuilder: (context, state, navigationShell) => _detail(
            context,
            state,
            MainWrapper(navigationShell: navigationShell),
          ),
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
                  builder: (context, state) => const DashboardPage(),
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

  static Page<void> _detail(
    BuildContext context,
    GoRouterState state,
    Widget child,
  ) => CustomTransitionPage<void>(
    key: state.pageKey,
    transitionDuration: SophiaMotion.resolve(context, SophiaMotion.medium),
    reverseTransitionDuration: SophiaMotion.resolve(
      context,
      SophiaMotion.medium,
    ),
    child: SophiaSheetBackdrop(child: child),
    transitionsBuilder: (context, animation, secondary, child) {
      final incoming = CurvedAnimation(
        parent: animation,
        curve: SophiaMotion.structuralCurve,
      );
      final outgoing = CurvedAnimation(
        parent: secondary,
        curve: SophiaMotion.structuralCurve,
      );
      return AnimatedBuilder(
        animation: Listenable.merge([animation, secondary]),
        child: child,
        builder: (_, child) => Stack(
          children: [
            Positioned.fill(
              child: Opacity(
                opacity: incoming.value,
                child: Transform.translate(
                  offset: Offset(
                    SophiaMotion.hierarchicalEnterOffset *
                            (1 - incoming.value) -
                        SophiaMotion.hierarchicalExitOffset * outgoing.value,
                    0,
                  ),
                  child: child,
                ),
              ),
            ),
            if (outgoing.value > 0)
              Positioned.fill(
                child: IgnorePointer(
                  child: ColoredBox(
                    color: context.colors.scrim.withValues(
                      alpha:
                          SophiaMotion.hierarchicalScrimOpacity *
                          outgoing.value,
                    ),
                  ),
                ),
              ),
          ],
        ),
      );
    },
  );
}
