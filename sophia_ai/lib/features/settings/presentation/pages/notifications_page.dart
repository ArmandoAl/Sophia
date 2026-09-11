import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

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
                    ContentSkeleton(),
                    SizedBox(height: SophiaSpace.lg),
                    ContentSkeleton(),
                  ],
                ),
              ),
            );
          }
          return MotionSwap(
            child: Align(
              key: const ValueKey('notifications-content'),
              alignment: Alignment.topCenter,
              child: ConstrainedBox(
                constraints: const BoxConstraints(
                  maxWidth: SophiaSize.contentMaxWidth,
                ),
                child: ListView(
                  padding: const EdgeInsets.all(SophiaSpace.lg),
                  children: [
                    Text(
                      'Elige con calma dónde puede avisarte Sofía.',
                      style: TextStyle(color: context.colors.softInk),
                    ),
                    const SizedBox(height: SophiaSpace.lg),
                    Text(
                      'Este dispositivo',
                      style: Theme.of(context).textTheme.labelLarge,
                    ),
                    const SizedBox(height: SophiaSpace.xs),
                    Container(
                      padding: const EdgeInsets.all(SophiaSpace.md),
                      decoration: BoxDecoration(
                        color: context.colors.elevated,
                        border: Border.all(color: context.colors.line),
                        borderRadius: BorderRadius.circular(SophiaRadius.card),
                      ),
                      child: Row(
                        children: [
                          Expanded(
                            child: Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                const Text(
                                  'Notificaciones en este dispositivo',
                                ),
                                const SizedBox(height: SophiaSpace.xxs),
                                Text(
                                  enabled
                                      ? 'Puedes cambiarlo cuando quieras.'
                                      : 'Esta función todavía está desactivada.',
                                  style: Theme.of(context).textTheme.bodySmall,
                                ),
                              ],
                            ),
                          ),
                          const SizedBox(width: SophiaSpace.md),
                          SophiaSwitch(
                            value: state.currentDeviceId != null,
                            semanticLabel: 'Notificaciones en este dispositivo',
                            onChanged: enabled
                                ? (value) async {
                                    if (value) {
                                      await context
                                          .read<NotificationsCubit>()
                                          .register();
                                    } else if (state.currentDeviceId
                                        case final id?) {
                                      await context
                                          .read<NotificationsCubit>()
                                          .remove(id);
                                    }
                                  }
                                : null,
                          ),
                        ],
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
                    const SizedBox(height: SophiaSpace.xl),
                    Text(
                      'Dispositivos registrados',
                      style: Theme.of(context).textTheme.labelLarge,
                    ),
                    const SizedBox(height: SophiaSpace.xs),
                    Container(
                      padding: const EdgeInsets.symmetric(
                        horizontal: SophiaSpace.md,
                      ),
                      decoration: BoxDecoration(
                        color: context.colors.elevated,
                        border: Border.all(color: context.colors.line),
                        borderRadius: BorderRadius.circular(SophiaRadius.card),
                      ),
                      child: state.devices.isEmpty
                          ? Padding(
                              padding: const EdgeInsets.symmetric(
                                vertical: SophiaSpace.lg,
                              ),
                              child: Text(
                                'No hay dispositivos registrados.',
                                style: TextStyle(color: context.colors.softInk),
                              ),
                            )
                          : Column(
                              children: [
                                for (
                                  var index = 0;
                                  index < state.devices.length;
                                  index++
                                ) ...[
                                  _DeviceRow(
                                    platform: state.devices[index].platform,
                                    createdAt: state.devices[index].createdAt,
                                    onRemove: () => context
                                        .read<NotificationsCubit>()
                                        .remove(state.devices[index].id),
                                  ),
                                  if (index < state.devices.length - 1)
                                    Divider(color: context.colors.line),
                                ],
                              ],
                            ),
                    ),
                  ],
                ),
              ),
            ),
          );
        },
      ),
    );
  }
}

class _DeviceRow extends StatelessWidget {
  const _DeviceRow({
    required this.platform,
    required this.createdAt,
    required this.onRemove,
  });

  final String platform;
  final DateTime createdAt;
  final VoidCallback onRemove;

  @override
  Widget build(BuildContext context) => Padding(
    padding: const EdgeInsets.symmetric(vertical: SophiaSpace.sm),
    child: Row(
      children: [
        Icon(Icons.devices_outlined, color: context.colors.softInk),
        const SizedBox(width: SophiaSpace.md),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(platform),
              Text(
                'Alta: ${_date(createdAt)}',
                style: SophiaType.dataLabel(context),
              ),
            ],
          ),
        ),
        TactileButton(
          semanticLabel: 'Quitar dispositivo $platform',
          onPressed: onRemove,
          child: const Padding(
            padding: EdgeInsets.all(SophiaSpace.sm),
            child: Icon(Icons.close),
          ),
        ),
      ],
    ),
  );
}

String _date(DateTime value) =>
    '${value.day.toString().padLeft(2, '0')}/${value.month.toString().padLeft(2, '0')}/${value.year}';
