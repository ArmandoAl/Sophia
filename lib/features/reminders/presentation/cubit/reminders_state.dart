import 'package:equatable/equatable.dart';

class TaskItem {
  final String id;
  final String title;
  final bool isCompleted;
  final bool isPriority;
  final String? aiSuggestion; // Texto verde si es sugerencia de IA

  const TaskItem({
    required this.id,
    required this.title,
    this.isCompleted = false,
    this.isPriority = false,
    this.aiSuggestion,
  });

  TaskItem copyWith({bool? isCompleted}) {
    return TaskItem(
      id: id,
      title: title,
      isPriority: isPriority,
      aiSuggestion: aiSuggestion,
      isCompleted: isCompleted ?? this.isCompleted,
    );
  }
}

class RemindersState extends Equatable {
  final List<TaskItem> tasks;
  const RemindersState(this.tasks);
  @override
  List<Object> get props => [tasks];
}
