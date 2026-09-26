import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:sophia_ai/core/di/service_locator.dart';
import 'package:sophia_ai/core/config/demo_mode.dart';
import 'package:sophia_ai/core/config/feature_flags.dart';
import 'package:sophia_ai/core/router/app_router.dart';
import 'package:sophia_ai/core/storage/token_storage.dart';
import 'package:sophia_ai/core/theme/app_theme.dart';
import 'package:sophia_ai/features/session/presentation/cubit/session_cubit.dart';
import 'package:sophia_ai/demo/demo_repositories.dart';
import 'package:sophia_ai/demo/demo_store.dart';

Future<void> main() async {
  WidgetsFlutterBinding.ensureInitialized();
  if (demoMode) {
    registerDemoRepositories();
    await initDependencies(
      tokenStorage: MemoryTokenStorage(),
      featureFlags: FeatureFlags.defaults.copyWith(notificationsEnabled: false),
    );
    final sessionCubit = sl<SessionCubit>()..applyMe(DemoStore.instance.me);
    runApp(SophiaApp(sessionCubit: sessionCubit, isDemo: true));
    return;
  }
  await initDependencies();
  final sessionCubit = sl<SessionCubit>();
  unawaited(sessionCubit.bootstrap());
  runApp(SophiaApp(sessionCubit: sessionCubit));
}

class SophiaApp extends StatelessWidget {
  const SophiaApp({
    super.key,
    required this.sessionCubit,
    this.router,
    this.isDemo = false,
  });

  final SessionCubit sessionCubit;
  final RouterConfig<Object>? router;
  final bool isDemo;

  @override
  Widget build(BuildContext context) {
    return BlocProvider.value(
      value: sessionCubit,
      child: MaterialApp.router(
        title: 'SOPHIA AI',
        debugShowCheckedModeBanner: false,
        theme: AppTheme.darkTheme,
        darkTheme: AppTheme.darkTheme,
        themeMode: ThemeMode.system,
        builder: (context, child) => isDemo
            ? Column(
                children: [
                  MaterialBanner(
                    content: const Text(
                      'Demo local · respuestas y acciones simuladas.',
                    ),
                    leading: const Icon(Icons.science_outlined),
                    actions: [const SizedBox.shrink()],
                  ),
                  Expanded(child: child ?? const SizedBox.shrink()),
                ],
              )
            : child ?? const SizedBox.shrink(),
        routerConfig: router ?? AppRouter.create(sessionCubit: sessionCubit),
      ),
    );
  }
}
