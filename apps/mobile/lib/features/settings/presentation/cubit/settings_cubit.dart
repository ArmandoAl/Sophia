import 'package:flutter_bloc/flutter_bloc.dart';
import 'settings_state.dart';

class SettingsCubit extends Cubit<SettingsState> {
  SettingsCubit() : super(const SettingsState());

  void toggleNotifications(bool value) {
    emit(state.copyWith(notificationsEnabled: value));
  }

  void toggleDarkMode(bool value) {
    emit(state.copyWith(darkModeEnabled: value));
  }

  void toggleVoiceAssistant(bool value) {
    emit(state.copyWith(voiceAssistantEnabled: value));
  }

  void changeLanguage(String language) {
    emit(state.copyWith(language: language));
  }

  void changeVoiceVolume(double volume) {
    emit(state.copyWith(voiceVolume: volume));
  }
}
