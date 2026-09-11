import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:go_router/go_router.dart';
import 'package:sophia_ai/core/di/service_locator.dart';
import 'package:sophia_ai/core/models/models.dart';
import 'package:sophia_ai/core/theme/design_tokens.dart';
import 'package:sophia_ai/core/widgets/motion/motion_widgets.dart';
import 'package:sophia_ai/core/widgets/neon_wrapper.dart';
import '../../domain/reminders_repository.dart';
import '../cubit/reminder_detail_cubit.dart';
import '../cubit/reminders_cubit.dart';
import '../cubit/reminders_state.dart';

class RemindersPage extends StatelessWidget {
  const RemindersPage({super.key, this.cubit, this.autoLoad = true});
  final RemindersCubit? cubit;
  final bool autoLoad;
  @override
  Widget build(BuildContext context) => NeonWrapper(
    child: BlocProvider.value(
      value: cubit ?? sl<RemindersCubit>()
        ..load(),
      child: Scaffold(
        backgroundColor: context.colors.surface.withValues(alpha: 0),
        appBar: AppBar(
          title: const Text('Reminders'),
          actions: [
            TactileButton(
              semanticLabel: 'Actualizar recordatorios',
              onPressed: () => context.read<RemindersCubit>().load(),
              child: const Padding(
                padding: EdgeInsets.all(SophiaSpace.sm),
                child: Icon(Icons.refresh),
              ),
            ),
            TactileButton(
              semanticLabel: 'Crear recordatorio',
              onPressed: () => _openEditor(context),
              child: const Padding(
                padding: EdgeInsets.all(SophiaSpace.sm),
                child: Icon(Icons.add),
              ),
            ),
            const SizedBox(width: SophiaSpace.xs),
          ],
        ),
        body: const _Body(),
      ),
    ),
  );
  void _openEditor(BuildContext context) => showSophiaSheet(
    context: context,
    builder: (_) => _ReminderEditor(cubit: context.read<RemindersCubit>()),
  );
}

class _Body extends StatelessWidget {
  const _Body();
  @override
  Widget build(BuildContext context) =>
      BlocBuilder<RemindersCubit, RemindersState>(
        builder: (context, state) {
          if (state.isLoading && state.reminders.isEmpty) {
            return const MotionSwap(
              child: Padding(
                key: ValueKey('reminders-loading'),
                padding: EdgeInsets.all(SophiaSpace.lg),
                child: Column(
                  children: [
                    ContentSkeleton(),
                    SizedBox(height: SophiaSpace.md),
                    ContentSkeleton(),
                  ],
                ),
              ),
            );
          }
          if (state.errorMessage != null && state.reminders.isEmpty) {
            return MotionSwap(
              child: Center(
                key: const ValueKey('reminders-error'),
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Text(
                      state.errorMessage!,
                      style: TextStyle(color: context.colors.softInk),
                    ),
                    TactileButton(
                      onPressed: () => context.read<RemindersCubit>().load(),
                      child: const Padding(
                        padding: EdgeInsets.all(SophiaSpace.sm),
                        child: Text('Reintentar'),
                      ),
                    ),
                  ],
                ),
              ),
            );
          }
          if (state.reminders.isEmpty) {
            return MotionSwap(
              child: Center(
                key: const ValueKey('reminders-empty'),
                child: Text(
                  'No tienes reminders todavía',
                  style: TextStyle(color: context.colors.softInk),
                ),
              ),
            );
          }
          return MotionSwap(
            child: RefreshIndicator(
              key: const ValueKey('reminders-content'),
              onRefresh: context.read<RemindersCubit>().load,
              child: ListView(
                padding: const EdgeInsets.all(SophiaSpace.lg),
                children: [
                  Text(
                    'Simple recurrence: none, daily, weekly, monthly',
                    style: Theme.of(context).textTheme.labelSmall?.copyWith(
                      color: context.colors.accent,
                    ),
                  ),
                  const SizedBox(height: SophiaSpace.sm),
                  ...state.reminders.map((r) => _ReminderTile(r)),
                ],
              ),
            ),
          );
        },
      );
}

class _ReminderTile extends StatelessWidget {
  const _ReminderTile(this.reminder);
  final Reminder reminder;
  @override
  Widget build(BuildContext context) => Container(
    margin: const EdgeInsets.only(bottom: SophiaSpace.sm),
    padding: const EdgeInsets.all(SophiaSpace.md),
    decoration: BoxDecoration(
      color: context.colors.elevated,
      border: Border.all(color: context.colors.line),
      borderRadius: BorderRadius.circular(SophiaRadius.card),
    ),
    child: Row(
      children: [
        Expanded(
          child: TactileButton(
            onPressed: () => context.push('/reminders/${reminder.id}'),
            child: Padding(
              padding: const EdgeInsets.symmetric(vertical: SophiaSpace.xs),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(reminder.title),
                  Text(
                    '${_when(reminder.scheduledAt)} · ${reminder.recurrenceRule.name}',
                    style: Theme.of(context).textTheme.labelSmall?.copyWith(
                      color: context.colors.accent,
                    ),
                  ),
                ],
              ),
            ),
          ),
        ),
        TactileButton(
          semanticLabel: 'Cancelar ${reminder.title}',
          onPressed: () => context.read<RemindersCubit>().cancel(reminder.id),
          child: const Padding(
            padding: EdgeInsets.all(SophiaSpace.xs),
            child: Icon(Icons.close),
          ),
        ),
        TactileButton(
          semanticLabel: 'Archivar ${reminder.title}',
          onPressed: () => context.read<RemindersCubit>().archive(reminder.id),
          child: const Padding(
            padding: EdgeInsets.all(SophiaSpace.xs),
            child: Icon(Icons.archive_outlined),
          ),
        ),
      ],
    ),
  );
}

String _when(DateTime d) =>
    '${d.toLocal().day}/${d.toLocal().month} ${d.toLocal().hour.toString().padLeft(2, '0')}:${d.toLocal().minute.toString().padLeft(2, '0')}';

class ReminderDetailPage extends StatelessWidget {
  const ReminderDetailPage({super.key, required this.id});
  final String id;
  @override
  Widget build(BuildContext context) => BlocProvider(
    create: (_) => ReminderDetailCubit(sl<RemindersRepository>())..load(id),
    child: Scaffold(
      appBar: AppBar(title: const Text('Reminder detail')),
      body: BlocBuilder<ReminderDetailCubit, ReminderDetailState>(
        builder: (_, s) {
          if (s.loading) {
            return const MotionSwap(
              child: Padding(
                key: ValueKey('reminder-detail-loading'),
                padding: EdgeInsets.all(SophiaSpace.lg),
                child: ContentSkeleton(height: 180),
              ),
            );
          }
          if (s.error != null) return Center(child: Text(s.error!));
          final r = s.reminder!;
          return Padding(
            padding: const EdgeInsets.all(SophiaSpace.lg),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(r.title, style: Theme.of(context).textTheme.headlineSmall),
                const SizedBox(height: SophiaSpace.sm),
                Text(r.description),
                const SizedBox(height: SophiaSpace.sm),
                Text('Programado: ${_when(r.scheduledAt)}'),
                Text('Recurrencia: ${r.recurrenceRule.name}'),
                Text('Estado: ${r.status.name}'),
                const SizedBox(height: SophiaSpace.lg),
                Text(
                  'El reminder se guarda en Sofia. Las notificaciones push reales todavía no están activas.',
                  style: TextStyle(color: context.colors.softInk),
                ),
              ],
            ),
          );
        },
      ),
    ),
  );
}

class _ReminderEditor extends StatefulWidget {
  const _ReminderEditor({required this.cubit});
  final RemindersCubit cubit;
  @override
  State<_ReminderEditor> createState() => _ReminderEditorState();
}

class _ReminderEditorState extends State<_ReminderEditor> {
  final title = TextEditingController();
  final description = TextEditingController();
  DateTime scheduled = DateTime.now().add(const Duration(hours: 1));
  RecurrenceRule recurrence = RecurrenceRule.none;
  @override
  Widget build(BuildContext context) => Padding(
    padding: EdgeInsets.only(
      left: SophiaSpace.lg,
      right: SophiaSpace.lg,
      top: SophiaSpace.lg,
      bottom: MediaQuery.viewInsetsOf(context).bottom + SophiaSpace.lg,
    ),
    child: Wrap(
      children: [
        TextField(
          controller: title,
          decoration: const InputDecoration(labelText: 'Título'),
        ),
        TextField(
          controller: description,
          decoration: const InputDecoration(labelText: 'Descripción'),
        ),
        TactileButton(
          onPressed: () async {
            final d = await showDatePicker(
              context: context,
              firstDate: DateTime.now(),
              lastDate: DateTime.now().add(const Duration(days: 3650)),
              initialDate: scheduled,
            );
            if (d != null) {
              setState(
                () => scheduled = DateTime(
                  d.year,
                  d.month,
                  d.day,
                  scheduled.hour,
                  scheduled.minute,
                ),
              );
            }
          },
          child: Padding(
            padding: const EdgeInsets.symmetric(vertical: SophiaSpace.sm),
            child: Row(
              children: [
                Expanded(child: Text('Fecha: ${_when(scheduled)}')),
                const Icon(Icons.event),
              ],
            ),
          ),
        ),
        DropdownButtonFormField(
          initialValue: recurrence,
          items: RecurrenceRule.values
              .map((r) => DropdownMenuItem(value: r, child: Text(r.name)))
              .toList(),
          onChanged: (r) {
            HapticFeedback.selectionClick();
            setState(() => recurrence = r!);
          },
          decoration: const InputDecoration(labelText: 'Recurrence'),
        ),
        const SizedBox(height: SophiaSpace.sm),
        TactileButton(
          onPressed: () async {
            if (title.text.trim().isEmpty) return;
            final navigator = Navigator.of(context);
            final ok = await widget.cubit.create(
              CreateReminderRequest(
                title: title.text.trim(),
                description: description.text.trim(),
                scheduledAt: scheduled.toUtc().toIso8601String(),
                timezone: DateTime.now().timeZoneName,
                recurrenceRule: recurrence.name,
              ),
            );
            if (ok && mounted) navigator.pop();
          },
          child: Container(
            padding: const EdgeInsets.all(SophiaSpace.sm),
            alignment: Alignment.center,
            decoration: BoxDecoration(
              color: context.colors.accent,
              borderRadius: BorderRadius.circular(SophiaRadius.control),
            ),
            child: Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                Icon(Icons.save, color: context.colors.surface),
                const SizedBox(width: SophiaSpace.xs),
                Text(
                  'Guardar reminder',
                  style: TextStyle(color: context.colors.surface),
                ),
              ],
            ),
          ),
        ),
        const SizedBox(height: SophiaSpace.xs),
        Text(
          'Push real puede no estar activo todavía.',
          style: TextStyle(color: context.colors.softInk),
        ),
      ],
    ),
  );
}
