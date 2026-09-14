import 'package:flutter_bloc/flutter_bloc.dart';

import '../../../../core/models/users/update_ai_settings_request.dart';
import '../../../../core/models/users/update_profile_request.dart';
import '../../../../core/network/api_exception.dart';
import '../../../session/presentation/cubit/session_cubit.dart';
import '../../domain/user_repository.dart';
import 'onboarding_state.dart';

/// Multi-step onboarding. Does not emit [OnboardingCompleted] unless
/// `POST /users/me/onboarding/complete` succeeds.
class OnboardingCubit extends Cubit<OnboardingState> {
  OnboardingCubit({
    required UserRepository userRepository,
    required SessionCubit sessionCubit,
  }) : _userRepository = userRepository,
       _sessionCubit = sessionCubit,
       super(const OnboardingInitial());

  final UserRepository _userRepository;
  final SessionCubit _sessionCubit;

  Future<void> run({
    UpdateProfileRequest? profile,
    UpdateAiSettingsRequest? aiSettings,
  }) async {
    try {
      if (profile != null && !profile.isEmpty) {
        emit(const OnboardingInProgress(OnboardingStep.profile));
        final updated = await _userRepository.updateProfile(profile);
        _sessionCubit.applyProfile(updated);
      }

      if (aiSettings != null && !aiSettings.isEmpty) {
        emit(const OnboardingInProgress(OnboardingStep.aiSettings));
        final updated = await _userRepository.updateAiSettings(aiSettings);
        _sessionCubit.applyAiSettings(updated);
      }

      emit(const OnboardingInProgress(OnboardingStep.complete));
      final completed = await _userRepository.completeOnboarding();
      _sessionCubit.applyProfile(completed);
      emit(OnboardingCompleted(completed));
    } on ApiException catch (e) {
      final step = switch (state) {
        OnboardingInProgress(:final step) => step,
        _ => OnboardingStep.idle,
      };
      emit(
        OnboardingFailure(message: e.message, failedStep: step, exception: e),
      );
    } catch (e) {
      final step = switch (state) {
        OnboardingInProgress(:final step) => step,
        _ => OnboardingStep.idle,
      };
      emit(OnboardingFailure(message: e.toString(), failedStep: step));
    }
  }

  void reset() => emit(const OnboardingInitial());

  @override
  void emit(OnboardingState state) {
    if (isClosed) return;
    super.emit(state);
  }
}
