import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:sophia_ai/core/widgets/main_wrapper.dart';
import 'package:sophia_ai/features/dashboard/presentation/pages/dashboard_page.dart';
import 'package:sophia_ai/features/smart_home/presentation/pages/smart_home_page.dart';
import 'package:sophia_ai/features/chat/presentation/pages/chat_page.dart';
import 'package:sophia_ai/features/settings/presentation/pages/settings_page.dart';

final GlobalKey<NavigatorState> _rootNavigatorKey = GlobalKey<NavigatorState>();
final GlobalKey<NavigatorState> _shellNavigatorKey =
    GlobalKey<NavigatorState>();

class AppRouter {
  static final GoRouter router = GoRouter(
    navigatorKey: _rootNavigatorKey,
    initialLocation: '/chat',
    routes: [
      // ShellRoute mantiene el estado de las tabs (no recarga al cambiar)
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
                path: '/smart-home',
                name: 'smart-home',
                builder: (context, state) => const SmartHomePage(),
              ),
            ],
          ),
          // Rama 0: Dashboard
          StatefulShellBranch(
            navigatorKey: _shellNavigatorKey,
            routes: [
              GoRoute(
                path: '/dashboard',
                name: 'dashboard',
                builder: (context, state) => const DashboardPage(),
              ),
            ],
          ),
          // Rama 1: Smart Home (Living Room, etc)

          // Rama 2: Chat (Sophia AI)

          // Rama 3: Settings
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
