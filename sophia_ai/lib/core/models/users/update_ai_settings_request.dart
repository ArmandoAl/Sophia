import 'package:equatable/equatable.dart';

import '../enums/autonomy_level.dart';
import '../enums/proactivity_level.dart';

/// PATCH `/users/me/ai-settings` — flat partial body (no wrapper).
///
/// Null fields are omitted. Explicit `false` for bools is serialized.
/// Never sends `user_id`.
class UpdateAiSettingsRequest extends Equatable {
  const UpdateAiSettingsRequest({
    this.assistantName,
    this.communicationStyle,
    this.proactivityLevel,
    this.autonomyLevel,
    this.memoryEnabled,
    this.remindersEnabled,
    this.planningEnabled,
    this.preferredLanguage,
  });

  final String? assistantName;
  final String? communicationStyle;
  final ProactivityLevel? proactivityLevel;
  final AutonomyLevel? autonomyLevel;
  final bool? memoryEnabled;
  final bool? remindersEnabled;
  final bool? planningEnabled;
  final String? preferredLanguage;

  /// Partial map: only non-null fields (bools include explicit `false`).
  Map<String, dynamic> toJson() {
    return <String, dynamic>{
      if (assistantName != null) 'assistant_name': assistantName,
      if (communicationStyle != null) 'communication_style': communicationStyle,
      if (proactivityLevel != null)
        'proactivity_level': proactivityLevel!.toJson(),
      if (autonomyLevel != null) 'autonomy_level': autonomyLevel!.toJson(),
      if (memoryEnabled != null) 'memory_enabled': memoryEnabled,
      if (remindersEnabled != null) 'reminders_enabled': remindersEnabled,
      if (planningEnabled != null) 'planning_enabled': planningEnabled,
      if (preferredLanguage != null) 'preferred_language': preferredLanguage,
    };
  }

  bool get isEmpty => toJson().isEmpty;

  @override
  List<Object?> get props => [
    assistantName,
    communicationStyle,
    proactivityLevel,
    autonomyLevel,
    memoryEnabled,
    remindersEnabled,
    planningEnabled,
    preferredLanguage,
  ];
}
