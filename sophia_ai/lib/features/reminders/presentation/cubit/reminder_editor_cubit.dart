import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:sophia_ai/core/models/models.dart';
import 'package:sophia_ai/core/network/api_exception.dart';
import '../../domain/reminders_repository.dart';

class ReminderEditorState {
  const ReminderEditorState({
    this.submitting = false,
    this.success = false,
    this.error,
  });
  final bool submitting, success;
  final String? error;
}

class ReminderEditorCubit extends Cubit<ReminderEditorState> {
  ReminderEditorCubit(this._repo) : super(const ReminderEditorState());
  final RemindersRepository _repo;
  Future<void> save({
    String? id,
    required CreateReminderRequest create,
    UpdateReminderRequest? update,
  }) async {
    emit(const ReminderEditorState(submitting: true));
    try {
      if (id == null) {
        await _repo.create(create);
      } else {
        await _repo.update(id, update!);
      }
      emit(const ReminderEditorState(success: true));
    } catch (e) {
      emit(
        ReminderEditorState(
          error: e is ApiException
              ? e.message
              : 'No se pudo guardar el reminder',
        ),
      );
    }
  }
}
