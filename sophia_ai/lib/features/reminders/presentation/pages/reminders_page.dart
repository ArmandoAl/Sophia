import 'package:flutter/material.dart';
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
            IconButton(
              onPressed: () => context.read<RemindersCubit>().load(),
              icon: const Icon(Icons.refresh),
            ),
            IconButton(
              onPressed: () => _openEditor(context),
              icon: const Icon(Icons.add),
            ),
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
                      style: const TextStyle(color: Colors.white70),
                    ),
                    TextButton(
                      onPressed: () => context.read<RemindersCubit>().load(),
                      child: const Text('Reintentar'),
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
                  const Text(
                    'Simple recurrence: none, daily, weekly, monthly',
                    style: TextStyle(color: Colors.cyanAccent, fontSize: 12),
                  ),
                  const SizedBox(height: 12),
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
  Widget build(BuildContext context) => Card(
    color: const Color(0xFF151B24),
    child: ListTile(
      onTap: () => context.push('/reminders/${reminder.id}'),
      title: Text(reminder.title, style: const TextStyle(color: Colors.white)),
      subtitle: Text(
        '${_when(reminder.scheduledAt)} · ${reminder.recurrenceRule.name}',
        style: const TextStyle(color: Colors.cyanAccent),
      ),
      trailing: PopupMenuButton<String>(
        onSelected: (v) => v == 'cancel'
            ? context.read<RemindersCubit>().cancel(reminder.id)
            : context.read<RemindersCubit>().archive(reminder.id),
        itemBuilder: (_) => const [
          PopupMenuItem(value: 'cancel', child: Text('Cancelar')),
          PopupMenuItem(value: 'archive', child: Text('Archivar')),
        ],
      ),
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
            padding: const EdgeInsets.all(24),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(r.title, style: Theme.of(context).textTheme.headlineSmall),
                const SizedBox(height: 12),
                Text(r.description),
                const SizedBox(height: 12),
                Text('Programado: ${_when(r.scheduledAt)}'),
                Text('Recurrencia: ${r.recurrenceRule.name}'),
                Text('Estado: ${r.status.name}'),
                const SizedBox(height: 24),
                const Text(
                  'El reminder se guarda en Sofia. Las notificaciones push reales todavía no están activas.',
                  style: TextStyle(color: Colors.white70),
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
      left: 20,
      right: 20,
      top: 20,
      bottom: MediaQuery.viewInsetsOf(context).bottom + 20,
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
        ListTile(
          title: Text('Fecha: ${_when(scheduled)}'),
          trailing: const Icon(Icons.event),
          onTap: () async {
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
        ),
        DropdownButtonFormField(
          initialValue: recurrence,
          items: RecurrenceRule.values
              .map((r) => DropdownMenuItem(value: r, child: Text(r.name)))
              .toList(),
          onChanged: (r) => setState(() => recurrence = r!),
          decoration: const InputDecoration(labelText: 'Recurrence'),
        ),
        const SizedBox(height: 12),
        FilledButton.icon(
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
          icon: const Icon(Icons.save),
          label: const Text('Guardar reminder'),
        ),
        const SizedBox(height: 8),
        const Text(
          'Push real puede no estar activo todavía.',
          style: TextStyle(color: Colors.white70),
        ),
      ],
    ),
  );
}
