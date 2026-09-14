import 'package:flutter_bloc/flutter_bloc.dart';

import '../../../../core/models/users/ai_settings.dart';
import '../../../../core/models/users/update_ai_settings_request.dart';
import '../../../../core/network/api_exception.dart';
import '../../../session/presentation/cubit/session_cubit.dart';
import '../../../session/presentation/cubit/session_state.dart';
import '../../domain/user_repository.dart';
import 'ai_settings_state.dart';

/// AI preference settings. Does **not** enable autonomous execution.
class AiSettingsCubit extends Cubit<AiSettingsState> {
  AiSettingsCubit({
    required UserRepository userRepository,
    required SessionCubit sessionCubit,
  }) : _userRepository = userRepository,
       _sessionCubit = sessionCubit,
       super(const AiSettingsInitial());

  final UserRepository _userRepository;
  final SessionCubit _sessionCubit;

  void loadFromSession() {
    final session = _sessionCubit.state;
    if (session is SessionAuthenticated) {
      emit(AiSettingsLoaded(session.aiSettings));
    }
  }

  Future<void> refresh() async {
    emit(const AiSettingsLoading());
    try {
      final settings = await _userRepository.getAiSettings();
      _sessionCubit.applyAiSettings(settings);
      emit(AiSettingsLoaded(settings));
    } on ApiException catch (e) {
      if (e.isUnauthorized) {
        await _sessionCubit.logout();
      }
      emit(AiSettingsFailure(message: e.message, exception: e));
    } catch (e) {
      emit(AiSettingsFailure(message: e.toString()));
    }
  }

  Future<void> update(UpdateAiSettingsRequest request) async {
    final previous = _currentSettings;
    if (previous != null) {
      emit(AiSettingsLoaded(previous, saving: true));
    } else {
      emit(const AiSettingsLoading());
    }

    try {
      final updated = await _userRepository.updateAiSettings(request);
      _sessionCubit.applyAiSettings(updated);
      emit(AiSettingsLoaded(updated));
    } on ApiException catch (e) {
      if (e.isUnauthorized) {
        await _sessionCubit.logout();
      }
      emit(
        AiSettingsFailure(message: e.message, exception: e, previous: previous),
      );
    } catch (e) {
      emit(AiSettingsFailure(message: e.toString(), previous: previous));
    }
  }

  AiSettings? get _currentSettings {
    final s = state;
    if (s is AiSettingsLoaded) return s.settings;
    if (s is AiSettingsFailure) return s.previous;
    final session = _sessionCubit.state;
    if (session is SessionAuthenticated) return session.aiSettings;
    return null;
  }

  @override
  void emit(AiSettingsState state) {
    if (isClosed) return;
    super.emit(state);
  }
}
