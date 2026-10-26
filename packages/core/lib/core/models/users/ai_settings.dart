import 'package:equatable/equatable.dart';

import '../enums/autonomy_level.dart';
import '../enums/proactivity_level.dart';

/// Flat AI settings (GET/PATCH `/users/me/ai-settings`, nested in MeResponse).
///
/// Free strings: [assistantName], [communicationStyle], [preferredLanguage].
/// Enums only: [proactivityLevel], [autonomyLevel].
class AiSettings extends Equatable {
  const AiSettings({
    required this.userId,
    required this.assistantName,
    required this.communicationStyle,
    required this.proactivityLevel,
    required this.autonomyLevel,
    required this.memoryEnabled,
    required this.remindersEnabled,
    required this.planningEnabled,
    required this.preferredLanguage,
    required this.createdAt,
    required this.updatedAt,
  });

  final String userId;
  final String assistantName;
  final String communicationStyle;
  final ProactivityLevel proactivityLevel;
  final AutonomyLevel autonomyLevel;
  final bool memoryEnabled;
  final bool remindersEnabled;
  final bool planningEnabled;
  final String preferredLanguage;
  final DateTime createdAt;
  final DateTime updatedAt;

  factory AiSettings.fromJson(Map<String, dynamic> json) {
    return AiSettings(
      userId: json['user_id'] as String,
      assistantName: json['assistant_name'] as String,
      communicationStyle: json['communication_style'] as String,
      proactivityLevel: ProactivityLevel.fromJson(
        json['proactivity_level'] as String,
      ),
      autonomyLevel: AutonomyLevel.fromJson(json['autonomy_level'] as String),
      memoryEnabled: json['memory_enabled'] as bool,
      remindersEnabled: json['reminders_enabled'] as bool,
      planningEnabled: json['planning_enabled'] as bool,
      preferredLanguage: json['preferred_language'] as String,
      createdAt: DateTime.parse(json['created_at'] as String),
      updatedAt: DateTime.parse(json['updated_at'] as String),
    );
  }

  Map<String, dynamic> toJson() => {
    'user_id': userId,
    'assistant_name': assistantName,
    'communication_style': communicationStyle,
    'proactivity_level': proactivityLevel.value,
    'autonomy_level': autonomyLevel.value,
    'memory_enabled': memoryEnabled,
    'reminders_enabled': remindersEnabled,
    'planning_enabled': planningEnabled,
    'preferred_language': preferredLanguage,
    'created_at': createdAt.toUtc().toIso8601String(),
    'updated_at': updatedAt.toUtc().toIso8601String(),
  };

  @override
  List<Object?> get props => [
    userId,
    assistantName,
    communicationStyle,
    proactivityLevel,
    autonomyLevel,
    memoryEnabled,
    remindersEnabled,
    planningEnabled,
    preferredLanguage,
    createdAt,
    updatedAt,
  ];
}
