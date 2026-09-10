import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:flutter/services.dart';

import '../../../../core/config/feature_flags.dart';
import '../../../../core/di/service_locator.dart';
import '../../../../core/theme/design_tokens.dart';
import '../../../../core/widgets/motion/motion_widgets.dart';
import '../cubit/settings_services_cubit.dart';

class NotificationsPage extends StatelessWidget {
  const NotificationsPage({super.key});

  @override
  Widget build(BuildContext context) => BlocProvider(
    create: (_) => sl<NotificationsCubit>()..load(),
    child: const _NotificationsView(),
  );
}

class _NotificationsView extends StatelessWidget {
  const _NotificationsView();

  @override
  Widget build(BuildContext context) {
    final enabled = sl<FeatureFlags>().notificationsEnabled;
    return Scaffold(
      appBar: AppBar(title: const Text('Notificaciones')),
      body: BlocBuilder<NotificationsCubit, NotificationsState>(
        builder: (context, state) {
          if (state.loading && state.devices.isEmpty) {
            return const MotionSwap(
              child: Padding(
                key: ValueKey('notifications-loading'),
                padding: EdgeInsets.all(SophiaSpace.lg),
                child: Column(
                  children: [
                    ContentSkeleton(height: SophiaSpace.xxl),
                    SizedBox(height: SophiaSpace.lg),
                    ContentSkeleton(),
                  ],
                ),
              ),
            );
          }
          return MotionSwap(
            child: ListView(
              key: const ValueKey('notifications-content'),
              padding: const EdgeInsets.all(SophiaSpace.lg),
              children: [
                SwitchListTile(
                  value: state.currentDeviceId != null,
                  onChanged: enabled
                      ? (value) async {
                          HapticFeedback.lightImpact();
                          if (value) {
                            await context.read<NotificationsCubit>().register();
                          } else if (state.currentDeviceId case final id?) {
                            await context.read<NotificationsCubit>().remove(id);
                          }
                        }
                      : null,
                  title: const Text('Notificaciones en este dispositivo'),
                  subtitle: Text(
                    enabled
                        ? 'Registra o quita este dispositivo.'
                        : 'Esta función todavía está desactivada.',
                  ),
                ),
                if (state.error != null)
                  Padding(
                    padding: const EdgeInsets.only(top: SophiaSpace.sm),
                    child: Text(
                      state.error!,
                      style: TextStyle(color: context.colors.critical),
                    ),
                  ),
                const SizedBox(height: SophiaSpace.lg),
                Text(
                  'Dispositivos registrados',
                  style: Theme.of(context).textTheme.titleMedium,
                ),
                const SizedBox(height: SophiaSpace.xs),
                if (!state.loading && state.devices.isEmpty)
                  const Text('No hay dispositivos registrados.'),
                for (final device in state.devices)
                  ListTile(
                    leading: const Icon(Icons.phone_iphone),
                    title: Text(device.platform),
                    subtitle: Text('Alta: ${_date(device.createdAt)}'),
                    trailing: IconButton(
                      tooltip: 'Quitar dispositivo',
                      icon: const Icon(Icons.close),
                      onPressed: () =>
                          context.read<NotificationsCubit>().remove(device.id),
                    ),
                  ),
              ],
            ),
          );
        },
      ),
    );
  }

  String _date(DateTime value) =>
      '${value.day.toString().padLeft(2, '0')}/${value.month.toString().padLeft(2, '0')}/${value.year}';
}
