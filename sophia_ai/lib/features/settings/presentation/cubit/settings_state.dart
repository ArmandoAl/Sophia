import 'package:equatable/equatable.dart';

class SettingsState extends Equatable {
  final bool notificationsEnabled;
  final bool darkModeEnabled;
  final bool voiceAssistantEnabled;
  final String language;
  final double voiceVolume;

  const SettingsState({
    this.notificationsEnabled = true,
    this.darkModeEnabled = true,
    this.voiceAssistantEnabled = true,
    this.language = 'English',
    this.voiceVolume = 0.8,
  });

  SettingsState copyWith({
    bool? notificationsEnabled,
    bool? darkModeEnabled,
    bool? voiceAssistantEnabled,
    String? language,
    double? voiceVolume,
  }) {
    return SettingsState(
      notificationsEnabled: notificationsEnabled ?? this.notificationsEnabled,
      darkModeEnabled: darkModeEnabled ?? this.darkModeEnabled,
      voiceAssistantEnabled:
          voiceAssistantEnabled ?? this.voiceAssistantEnabled,
      language: language ?? this.language,
      voiceVolume: voiceVolume ?? this.voiceVolume,
    );
  }

  @override
  List<Object> get props => [
    notificationsEnabled,
    darkModeEnabled,
    voiceAssistantEnabled,
    language,
    voiceVolume,
  ];
}
