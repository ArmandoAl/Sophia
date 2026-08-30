import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:sophia_ai/core/models/models.dart';
import 'package:sophia_ai/core/network/api_exception.dart';
import '../../domain/reminders_repository.dart';

class ReminderDetailState {
  const ReminderDetailState({this.reminder, this.loading = false, this.error});
  final Reminder? reminder;
  final bool loading;
  final String? error;
}

class ReminderDetailCubit extends Cubit<ReminderDetailState> {
  ReminderDetailCubit(this._repo) : super(const ReminderDetailState());
  final RemindersRepository _repo;
  Future<void> load(String id) async {
    emit(const ReminderDetailState(loading: true));
    try {
      emit(ReminderDetailState(reminder: await _repo.getById(id)));
    } catch (e) {
      emit(
        ReminderDetailState(
          error: e is ApiException
              ? e.message
              : 'No se pudo cargar el reminder',
        ),
      );
    }
  }
}
