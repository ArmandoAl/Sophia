import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:sophia_ai/core/models/models.dart';
import 'package:sophia_ai/core/network/api_exception.dart';
import '../../domain/reminders_repository.dart';
import 'reminders_state.dart';

class RemindersCubit extends Cubit<RemindersState> {
  RemindersCubit({required RemindersRepository repository})
    : _repository = repository,
      super(const RemindersState());
  final RemindersRepository _repository;
  Future<void> load() async {
    emit(state.copyWith(isLoading: true, clearError: true));
    try {
      final result = await _repository.list(status: 'pending');
      emit(state.copyWith(reminders: result.reminders, isLoading: false));
    } catch (e) {
      emit(state.copyWith(isLoading: false, errorMessage: _error(e)));
    }
  }

  Future<bool> create(CreateReminderRequest request) =>
      _submit(() => _repository.create(request));
  Future<bool> cancel(String id) => _submit(() => _repository.cancel(id));
  Future<bool> archive(String id) => _submit(() => _repository.archive(id));
  Future<bool> update(String id, UpdateReminderRequest request) =>
      _submit(() => _repository.update(id, request));
  Future<bool> _submit(Future<Reminder> Function() action) async {
    emit(state.copyWith(isSubmitting: true, clearError: true));
    try {
      await action();
      await load();
      emit(state.copyWith(isSubmitting: false));
      return true;
    } catch (e) {
      emit(state.copyWith(isSubmitting: false, errorMessage: _error(e)));
      return false;
    }
  }

  String _error(Object e) =>
      e is ApiException ? e.message : 'No se pudieron cargar los reminders';
}
