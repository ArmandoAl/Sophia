import 'package:flutter/material.dart';
import 'core/di/service_locator.dart';
import 'core/theme/app_theme.dart';
import 'core/router/app_router.dart'; // Importar el router

void main() async {
  WidgetsFlutterBinding.ensureInitialized();
  await initDependencies();
  runApp(const SophiaApp());
}

class SophiaApp extends StatelessWidget {
  const SophiaApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp.router(
      // Cambiar a .router
      title: 'SOPHIA AI',
      debugShowCheckedModeBanner: false,
      theme: AppTheme.darkTheme,
      routerConfig: AppRouter.router, // Conectar GoRouter
    );
  }
}
