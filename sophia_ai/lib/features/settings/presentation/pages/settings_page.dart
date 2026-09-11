import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:go_router/go_router.dart';

import '../../../../core/theme/design_tokens.dart';
import '../../../../core/widgets/motion/motion_widgets.dart';
import '../../../session/presentation/cubit/session_cubit.dart';
import '../../../session/presentation/cubit/session_state.dart';

class SettingsPage extends StatelessWidget {
  const SettingsPage({super.key});

  @override
  Widget build(BuildContext context) => Scaffold(
    appBar: AppBar(
      title: const Text('Ajustes'),
      actions: [
        TactileButton(
          key: const Key('settings_logout'),
          semanticLabel: 'Cerrar sesión',
          onPressed: context.read<SessionCubit>().logout,
          child: const Padding(
            padding: EdgeInsets.all(SophiaSpace.sm),
            child: Icon(Icons.logout),
          ),
        ),
        const SizedBox(width: SophiaSpace.xs),
      ],
    ),
    body: ListView(
      padding: const EdgeInsets.all(SophiaSpace.lg),
      children: [
        BlocBuilder<SessionCubit, SessionState>(
          builder: (context, state) {
            final email = state is SessionAuthenticated ? state.user.email : '';
            return Text(
              email.isEmpty ? 'Tu cuenta y Sofía' : email,
              style: TextStyle(color: context.colors.softInk),
            );
          },
        ),
        const SizedBox(height: SophiaSpace.xl),
        _Section(
          title: 'Cuenta',
          children: [
            _row(
              context,
              key: const Key('settings_profile'),
              icon: Icons.person_outline,
              title: 'Perfil',
              subtitle: 'Nombre, zona horaria e idioma',
              route: '/profile',
            ),
            _row(
              context,
              key: const Key('settings_assistant'),
              icon: Icons.tune,
              title: 'Ajustes del asistente',
              subtitle: 'Cómo quieres que te acompañe',
              route: '/assistant-settings',
            ),
          ],
        ),
        const SizedBox(height: SophiaSpace.lg),
        _Section(
          title: 'Aprendizaje',
          children: [
            _row(
              context,
              key: const Key('settings_beliefs'),
              icon: Icons.psychology_outlined,
              title: 'Lo que Sofía sabe de ti',
              subtitle: 'Revisa y corrige lo que ha entendido',
              route: '/beliefs',
            ),
            _row(
              context,
              key: const Key('settings_contexts'),
              icon: Icons.people_outline,
              title: 'Contextos',
              subtitle: 'Personas y modos en los que actúas distinto',
              route: '/contexts',
            ),
            _row(
              context,
              key: const Key('settings_entities'),
              icon: Icons.diversity_3_outlined,
              title: 'Personas y grupos',
              subtitle: 'Revisa lo que recuerdas de tu gente',
              route: '/entities',
            ),
            _row(
              context,
              icon: Icons.forum_outlined,
              title: 'Importar conversaciones',
              subtitle: 'Revisa lotes procesados por API',
              route: '/import-conversations',
            ),
          ],
        ),
        const SizedBox(height: SophiaSpace.lg),
        _Section(
          title: 'Control',
          children: [
            _row(
              context,
              icon: Icons.notifications_outlined,
              title: 'Notificaciones',
              route: '/notifications',
            ),
            _row(
              context,
              icon: Icons.lock_outline,
              title: 'Privacidad y datos',
              route: '/privacy-data',
            ),
            _row(
              context,
              key: const Key('settings_diagnostics'),
              icon: Icons.monitor_heart_outlined,
              title: 'Diagnóstico',
              subtitle: 'Estado del backend y del aprendizaje',
              route: '/diagnostics',
            ),
          ],
        ),
        const SizedBox(height: SophiaSpace.xl),
        TactileButton(
          onPressed: context.read<SessionCubit>().logout,
          child: Container(
            padding: const EdgeInsets.all(SophiaSpace.md),
            decoration: BoxDecoration(
              border: Border.all(color: context.colors.line),
              borderRadius: BorderRadius.circular(SophiaRadius.control),
            ),
            child: Row(
              children: [
                Icon(Icons.logout, color: context.colors.softInk),
                const SizedBox(width: SophiaSpace.sm),
                Text(
                  'Cerrar sesión',
                  style: TextStyle(color: context.colors.softInk),
                ),
              ],
            ),
          ),
        ),
      ],
    ),
  );

  Widget _row(
    BuildContext context, {
    Key? key,
    required IconData icon,
    required String title,
    String? subtitle,
    required String route,
  }) => TactileButton(
    key: key,
    onPressed: () => context.push(route),
    child: Padding(
      padding: const EdgeInsets.symmetric(vertical: SophiaSpace.sm),
      child: Row(
        children: [
          Icon(icon, color: context.colors.softInk),
          const SizedBox(width: SophiaSpace.md),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(title),
                if (subtitle != null)
                  Text(
                    subtitle,
                    style: Theme.of(context).textTheme.bodySmall?.copyWith(
                      color: context.colors.softInk,
                    ),
                  ),
              ],
            ),
          ),
          Icon(Icons.chevron_right, color: context.colors.muted),
        ],
      ),
    ),
  );
}

class _Section extends StatelessWidget {
  const _Section({required this.title, required this.children});
  final String title;
  final List<Widget> children;

  @override
  Widget build(BuildContext context) => Column(
    crossAxisAlignment: CrossAxisAlignment.start,
    children: [
      Padding(
        padding: const EdgeInsets.only(
          left: SophiaSpace.xs,
          bottom: SophiaSpace.xs,
        ),
        child: Text(
          title,
          style: Theme.of(
            context,
          ).textTheme.labelLarge?.copyWith(color: context.colors.softInk),
        ),
      ),
      Container(
        padding: const EdgeInsets.symmetric(horizontal: SophiaSpace.md),
        decoration: BoxDecoration(
          color: context.colors.elevated,
          borderRadius: BorderRadius.circular(SophiaRadius.card),
          border: Border.all(color: context.colors.line),
        ),
        child: Column(
          children: [
            for (var i = 0; i < children.length; i++) ...[
              children[i],
              if (i < children.length - 1)
                Divider(color: context.colors.line, height: 1),
            ],
          ],
        ),
      ),
    ],
  );
}
