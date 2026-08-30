import 'package:flutter_bloc/flutter_bloc.dart';

import '../../../../core/network/api_exception.dart';
import '../../domain/system_repository.dart';
import 'health_state.dart';

/// Backend `GET /health` status. Not related to UI mock CPU/RAM metrics.
class HealthCubit extends Cubit<HealthState> {
  HealthCubit({required SystemRepository systemRepository})
    : _systemRepository = systemRepository,
      super(const HealthInitial());

  final SystemRepository _systemRepository;

  Future<void> check() async {
    emit(const HealthLoading());
    try {
      final health = await _systemRepository.getHealth();
      if (health.isDegraded) {
        emit(HealthDegraded(health));
      } else {
        emit(HealthHealthy(health));
      }
    } on ApiException catch (e) {
      emit(HealthFailure(message: e.message, exception: e));
    } catch (e) {
      emit(HealthFailure(message: e.toString()));
    }
  }
}
