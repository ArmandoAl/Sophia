import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:go_router/go_router.dart';

import '../../../../core/di/service_locator.dart';
import '../../../../core/models/reminders/reminder.dart';
import '../../../../core/theme/design_tokens.dart';
import '../../../../core/theme/motion.dart';
import '../../../../core/widgets/motion/motion_widgets.dart';
import '../../../reminders/presentation/cubit/reminders_cubit.dart';
import '../../../reminders/presentation/cubit/reminders_state.dart';

class DashboardPage extends StatelessWidget {
  const DashboardPage({super.key, this.cubit, this.autoLoad = true});

  final RemindersCubit? cubit;
  final bool autoLoad;

  @override
  Widget build(BuildContext context) => BlocProvider(
    create: (_) {
      final value = cubit ?? sl<RemindersCubit>();
      if (autoLoad) value.load();
      return value;
    },
    child: const _DashboardView(),
  );
}

class _DashboardView extends StatelessWidget {
  const _DashboardView();

  @override
  Widget build(BuildContext context) => Scaffold(
    appBar: AppBar(title: const Text('Hoy')),
    body: BlocBuilder<RemindersCubit, RemindersState>(
      builder: (context, state) {
        if (state.isLoading && state.reminders.isEmpty) {
          return const _DashboardLoading();
        }
        if (state.errorMessage != null && state.reminders.isEmpty) {
          return _DashboardError(message: state.errorMessage!);
        }

        final now = DateTime.now();
        final today = state.reminders
            .where((item) => DateUtils.isSameDay(item.nextRunAt.toLocal(), now))
            .toList();
        final later = state.reminders
            .where(
              (item) => !DateUtils.isSameDay(item.nextRunAt.toLocal(), now),
            )
            .take(3)
            .toList();

        return Align(
          alignment: Alignment.topCenter,
          child: ConstrainedBox(
            constraints: const BoxConstraints(
              maxWidth: SophiaSize.contentMaxWidth,
            ),
            child: ListView(
              padding: const EdgeInsets.all(SophiaSpace.lg),
              children: [
                StaggeredEntry(
                  index: 0,
                  child: _TodaySection(reminders: today),
                ),
                if (later.isNotEmpty) ...[
                  const SizedBox(height: SophiaSpace.xl),
                  StaggeredEntry(
                    index: 1,
                    child: Text(
                      'Después',
                      style: Theme.of(context).textTheme.titleMedium,
                    ),
                  ),
                  const SizedBox(height: SophiaSpace.sm),
                  StaggeredEntry(
                    index: 2,
                    child: _ReminderList(reminders: later),
                  ),
                ],
                const SizedBox(height: SophiaSpace.lg),
                StaggeredEntry(
                  index: later.isEmpty ? 1 : 3,
                  child: Align(
                    alignment: Alignment.centerLeft,
                    child: TactileButton(
                      semanticLabel: 'Ver todos los recordatorios',
                      onPressed: () => context.push('/reminders'),
                      child: Padding(
                        padding: const EdgeInsets.symmetric(
                          vertical: SophiaSpace.sm,
                        ),
                        child: Text(
                          'Ver todos los recordatorios',
                          style: Theme.of(context).textTheme.labelLarge
                              ?.copyWith(color: context.colors.accent),
                        ),
                      ),
                    ),
                  ),
                ),
              ],
            ),
          ),
        );
      },
    ),
  );
}

class _TodaySection extends StatelessWidget {
  const _TodaySection({required this.reminders});

  final List<Reminder> reminders;

  @override
  Widget build(BuildContext context) {
    final hasAttention = reminders.isNotEmpty;
    return Container(
      padding: const EdgeInsets.all(SophiaSpace.lg),
      decoration: BoxDecoration(
        color: hasAttention
            ? context.colors.attention.withValues(alpha: SophiaOpacity.faint)
            : context.colors.elevated,
        borderRadius: BorderRadius.circular(SophiaRadius.card),
        border: Border.all(
          color: hasAttention ? context.colors.attention : context.colors.line,
        ),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text('PARA HOY', style: Theme.of(context).textTheme.labelMedium),
          const SizedBox(height: SophiaSpace.sm),
          Row(
            crossAxisAlignment: CrossAxisAlignment.end,
            children: [
              TweenAnimationBuilder<int>(
                tween: IntTween(begin: 0, end: reminders.length),
                duration: SophiaMotion.resolve(context, SophiaMotion.short),
                curve: SophiaMotion.contentCurve,
                builder: (_, value, _) =>
                    Text('$value', style: SophiaType.data(context)),
              ),
              const SizedBox(width: SophiaSpace.sm),
              Expanded(
                child: Text(
                  hasAttention
                      ? reminders.length == 1
                            ? 'recordatorio que merece tu atención'
                            : 'recordatorios que merecen tu atención'
                      : 'Nada pendiente por ahora',
                  style: Theme.of(context).textTheme.titleLarge,
                ),
              ),
            ],
          ),
          if (hasAttention) ...[
            const SizedBox(height: SophiaSpace.lg),
            _ReminderList(reminders: reminders),
          ] else ...[
            const SizedBox(height: SophiaSpace.xs),
            Text(
              'Puedes seguir con tu día. Sofía te avisará cuando haya algo que revisar.',
              style: Theme.of(
                context,
              ).textTheme.bodyMedium?.copyWith(color: context.colors.softInk),
            ),
          ],
        ],
      ),
    );
  }
}

class _ReminderList extends StatelessWidget {
  const _ReminderList({required this.reminders});

  final List<Reminder> reminders;

  @override
  Widget build(BuildContext context) => Column(
    children: [
      for (var index = 0; index < reminders.length; index++) ...[
        _ReminderRow(reminder: reminders[index]),
        if (index < reminders.length - 1)
          Divider(height: SophiaSpace.md, color: context.colors.line),
      ],
    ],
  );
}

class _ReminderRow extends StatelessWidget {
  const _ReminderRow({required this.reminder});

  final Reminder reminder;

  @override
  Widget build(BuildContext context) {
    final time = reminder.nextRunAt.toLocal();
    final hour = time.hour.toString().padLeft(2, '0');
    final minute = time.minute.toString().padLeft(2, '0');
    return TactileButton(
      semanticLabel: 'Abrir ${reminder.title}',
      onPressed: () => context.push('/reminders/${reminder.id}'),
      child: Padding(
        padding: const EdgeInsets.symmetric(vertical: SophiaSpace.xs),
        child: Row(
          children: [
            Text('$hour:$minute', style: SophiaType.dataLabel(context)),
            const SizedBox(width: SophiaSpace.md),
            Expanded(
              child: Text(
                reminder.title,
                style: Theme.of(context).textTheme.bodyMedium,
              ),
            ),
            Icon(Icons.chevron_right, color: context.colors.muted),
          ],
        ),
      ),
    );
  }
}

class _DashboardLoading extends StatelessWidget {
  const _DashboardLoading();

  @override
  Widget build(BuildContext context) => const Padding(
    padding: EdgeInsets.all(SophiaSpace.lg),
    child: Column(
      children: [
        ContentSkeleton(),
        SizedBox(height: SophiaSpace.lg),
        ContentSkeleton(),
      ],
    ),
  );
}

class _DashboardError extends StatelessWidget {
  const _DashboardError({required this.message});

  final String message;

  @override
  Widget build(BuildContext context) => Center(
    child: Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        Text(message, style: TextStyle(color: context.colors.softInk)),
        const SizedBox(height: SophiaSpace.sm),
        TactileButton(
          semanticLabel: 'Reintentar',
          onPressed: context.read<RemindersCubit>().load,
          child: const Padding(
            padding: EdgeInsets.all(SophiaSpace.sm),
            child: Text('Reintentar'),
          ),
        ),
      ],
    ),
  );
}
