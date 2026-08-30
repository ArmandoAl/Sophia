import 'package:equatable/equatable.dart';

import '../../../../core/models/users/user_profile.dart';
import '../../../../core/network/api_exception.dart';

enum OnboardingStep { idle, profile, aiSettings, complete }

sealed class OnboardingState extends Equatable {
  const OnboardingState();

  @override
  List<Object?> get props => [];
}

final class OnboardingInitial extends OnboardingState {
  const OnboardingInitial();
}

final class OnboardingInProgress extends OnboardingState {
  const OnboardingInProgress(this.step);

  final OnboardingStep step;

  @override
  List<Object?> get props => [step];
}

final class OnboardingCompleted extends OnboardingState {
  const OnboardingCompleted(this.profile);

  final UserProfile profile;

  @override
  List<Object?> get props => [profile];
}

final class OnboardingFailure extends OnboardingState {
  const OnboardingFailure({
    required this.message,
    required this.failedStep,
    this.exception,
  });

  final String message;
  final OnboardingStep failedStep;
  final ApiException? exception;

  @override
  List<Object?> get props => [message, failedStep, exception];
}
