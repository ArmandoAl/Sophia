import 'package:equatable/equatable.dart';

import '../../../../core/models/users/ai_settings.dart';
import '../../../../core/network/api_exception.dart';

sealed class AiSettingsState extends Equatable {
  const AiSettingsState();

  @override
  List<Object?> get props => [];
}

final class AiSettingsInitial extends AiSettingsState {
  const AiSettingsInitial();
}

final class AiSettingsLoading extends AiSettingsState {
  const AiSettingsLoading();
}

final class AiSettingsLoaded extends AiSettingsState {
  const AiSettingsLoaded(this.settings, {this.saving = false});

  final AiSettings settings;
  final bool saving;

  bool get memoryEnabled => settings.memoryEnabled;
  bool get remindersEnabled => settings.remindersEnabled;
  bool get planningEnabled => settings.planningEnabled;

  @override
  List<Object?> get props => [settings, saving];
}

final class AiSettingsFailure extends AiSettingsState {
  const AiSettingsFailure({
    required this.message,
    this.exception,
    this.previous,
  });

  final String message;
  final ApiException? exception;
  final AiSettings? previous;

  @override
  List<Object?> get props => [message, exception, previous];
}
