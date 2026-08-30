import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:go_router/go_router.dart';
import 'package:sophia_ai/core/widgets/neon_wrapper.dart';
import 'package:sophia_ai/core/widgets/sophia_card.dart';
import 'package:sophia_ai/features/session/presentation/cubit/session_cubit.dart';
import 'package:sophia_ai/features/session/presentation/cubit/session_state.dart';

/// Settings hub for F1 account surfaces (profile, assistant, logout).
class SettingsPage extends StatelessWidget {
  const SettingsPage({super.key});

  @override
  Widget build(BuildContext context) {
    return NeonWrapper(
      child: Scaffold(
        backgroundColor: Colors.transparent,
        body: SafeArea(
          child: Padding(
            padding: const EdgeInsets.all(24),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  'Settings',
                  style: Theme.of(context).textTheme.headlineMedium?.copyWith(
                    fontWeight: FontWeight.bold,
                  ),
                ),
                const SizedBox(height: 8),
                BlocBuilder<SessionCubit, SessionState>(
                  builder: (context, state) {
                    final email = state is SessionAuthenticated
                        ? state.user.email
                        : '';
                    return Text(
                      email.isEmpty
                          ? 'Account & assistant'
                          : 'Signed in as $email',
                      style: Theme.of(
                        context,
                      ).textTheme.bodyMedium?.copyWith(color: Colors.grey),
                    );
                  },
                ),
                const SizedBox(height: 32),
                Expanded(
                  child: ListView(
                    children: [
                      SophiaCard(
                        child: Column(
                          children: [
                            ListTile(
                              key: const Key('settings_profile'),
                              leading: const Icon(Icons.person_outline),
                              title: const Text('Profile'),
                              subtitle: const Text('Name, timezone, locale'),
                              trailing: const Icon(Icons.chevron_right),
                              onTap: () => context.go('/profile'),
                            ),
                            const Divider(color: Colors.white10, height: 1),
                            ListTile(
                              key: const Key('settings_assistant'),
                              leading: const Icon(Icons.smart_toy_outlined),
                              title: const Text('Assistant settings'),
                              subtitle: const Text(
                                'Sofía preferences (proposal-only)',
                              ),
                              trailing: const Icon(Icons.chevron_right),
                              onTap: () => context.go('/assistant-settings'),
                            ),
                            const Divider(color: Colors.white10, height: 1),
                            ListTile(
                              key: const Key('settings_diagnostics'),
                              leading: const Icon(Icons.monitor_heart_outlined),
                              title: const Text('Diagnostics'),
                              subtitle: const Text('Backend health'),
                              trailing: const Icon(Icons.chevron_right),
                              onTap: () => context.go('/dashboard'),
                            ),
                          ],
                        ),
                      ),
                      const SizedBox(height: 24),
                      SophiaCard(
                        child: ListTile(
                          key: const Key('settings_logout'),
                          leading: Icon(
                            Icons.logout,
                            color: Theme.of(context).colorScheme.error,
                          ),
                          title: Text(
                            'Sign out',
                            style: TextStyle(
                              color: Theme.of(context).colorScheme.error,
                            ),
                          ),
                          onTap: () => context.read<SessionCubit>().logout(),
                        ),
                      ),
                    ],
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}
