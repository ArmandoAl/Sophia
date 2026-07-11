import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:sophia_ai/features/reminders/presentation/cubit/reminders_state.dart';

class RemindersCubit extends Cubit<RemindersState> {
  RemindersCubit() : super(const RemindersState([])) {
    loadTasks();
  }

  void loadTasks() {
    emit(
      const RemindersState([
        TaskItem(id: '1', title: "Pay Electricity Bill", isPriority: true),
        TaskItem(id: '2', title: "Submit Project Proposal", isPriority: true),
        TaskItem(
          id: '3',
          title: "Buy milk",
          aiSuggestion: "AI Suggested based on location",
        ),
        TaskItem(id: '4', title: "Call Mom"),
        TaskItem(id: '5', title: "Book flight for vacation", isCompleted: true),
      ]),
    );
  }

  void toggleTask(String id) {
    final newTasks = state.tasks.map((task) {
      if (task.id == id) return task.copyWith(isCompleted: !task.isCompleted);
      return task;
    }).toList();
    emit(RemindersState(newTasks));
  }
}
