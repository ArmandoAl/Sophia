import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:sophia_ai/core/widgets/neon_wrapper.dart';
import '../cubit/reminders_cubit.dart';
import '../cubit/reminders_state.dart';

class RemindersPage extends StatelessWidget {
  const RemindersPage({super.key});

  @override
  Widget build(BuildContext context) {
    return NeonWrapper(
      child: BlocProvider(
        create: (_) => RemindersCubit(),
        child: Scaffold(
          backgroundColor: Colors.transparent,
          appBar: AppBar(
            title: const Text("Smart Reminders"),
            actions: [
              IconButton(icon: const Icon(Icons.add), onPressed: () {}),
            ],
          ),
          body: const _RemindersBody(),
        ),
      ),
    );
  }
}

class _RemindersBody extends StatelessWidget {
  const _RemindersBody();

  @override
  Widget build(BuildContext context) {
    return BlocBuilder<RemindersCubit, RemindersState>(
      builder: (context, state) {
        final highPriority = state.tasks.where((t) => t.isPriority).toList();
        final upcoming = state.tasks.where((t) => !t.isPriority).toList();

        return ListView(
          padding: const EdgeInsets.all(24),
          children: [
            if (highPriority.isNotEmpty) ...[
              const _SectionHeader(
                title: "High Priority",
                color: Colors.redAccent,
              ),
              ...highPriority.map((task) => _TaskTile(task: task)),
              const SizedBox(height: 24),
            ],

            const _SectionHeader(title: "Upcoming", color: Colors.cyanAccent),
            ...upcoming.map((task) => _TaskTile(task: task)),
          ],
        );
      },
    );
  }
}

class _SectionHeader extends StatelessWidget {
  final String title;
  final Color color;

  const _SectionHeader({required this.title, required this.color});

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 12),
      child: Row(
        children: [
          Container(width: 20, height: 2, color: color),
          const SizedBox(width: 8),
          Text(
            title,
            style: const TextStyle(
              fontWeight: FontWeight.bold,
              fontSize: 16,
              color: Colors.white,
            ),
          ),
        ],
      ),
    );
  }
}

class _TaskTile extends StatelessWidget {
  final TaskItem task;

  const _TaskTile({required this.task});

  @override
  Widget build(BuildContext context) {
    final isDone = task.isCompleted;

    return Container(
      margin: const EdgeInsets.only(bottom: 12),
      padding: const EdgeInsets.all(4),
      decoration: BoxDecoration(
        color: const Color(0xFF151B24),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: Colors.white.withValues(alpha: 0.05)),
      ),
      child: ListTile(
        leading: Transform.scale(
          scale: 1.2,
          child: Checkbox(
            value: isDone,
            activeColor: Colors.cyanAccent,
            checkColor: Colors.black,
            side: const BorderSide(color: Colors.grey),
            shape: RoundedRectangleBorder(
              borderRadius: BorderRadius.circular(4),
            ),
            onChanged: (_) =>
                context.read<RemindersCubit>().toggleTask(task.id),
          ),
        ),
        title: Text(
          task.title,
          style: TextStyle(
            color: isDone ? Colors.grey : Colors.white,
            decoration: isDone ? TextDecoration.lineThrough : null,
          ),
        ),
        subtitle: task.aiSuggestion != null
            ? Padding(
                padding: const EdgeInsets.only(top: 4.0),
                child: Row(
                  children: [
                    const Icon(
                      Icons.auto_awesome,
                      size: 12,
                      color: Colors.greenAccent,
                    ),
                    const SizedBox(width: 4),
                    Text(
                      task.aiSuggestion!,
                      style: const TextStyle(
                        color: Colors.greenAccent,
                        fontSize: 12,
                      ),
                    ),
                  ],
                ),
              )
            : null,
      ),
    );
  }
}
