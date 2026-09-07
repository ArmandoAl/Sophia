import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:sophia_ai/core/di/service_locator.dart';
import 'package:sophia_ai/core/router/app_router.dart';
import 'package:sophia_ai/core/theme/app_theme.dart';
import 'package:sophia_ai/features/session/presentation/cubit/session_cubit.dart';

Future<void> main() async {
  WidgetsFlutterBinding.ensureInitialized();
  await initDependencies();
  final sessionCubit = sl<SessionCubit>();
  unawaited(sessionCubit.bootstrap());
  runApp(SophiaApp(sessionCubit: sessionCubit));
}

class SophiaApp extends StatelessWidget {
  const SophiaApp({super.key, required this.sessionCubit, this.router});

  final SessionCubit sessionCubit;
  final RouterConfig<Object>? router;

  @override
  Widget build(BuildContext context) {
    return BlocProvider.value(
      value: sessionCubit,
      child: MaterialApp.router(
        title: 'SOPHIA AI',
        debugShowCheckedModeBanner: false,
        theme: AppTheme.darkTheme,
        routerConfig: router ?? AppRouter.create(sessionCubit: sessionCubit),
      ),
    );
  }
}
