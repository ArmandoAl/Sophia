import '../../../core/models/users/ai_settings.dart';
import '../../../core/models/users/me_response.dart';
import '../../../core/models/users/update_ai_settings_request.dart';
import '../../../core/models/users/update_profile_request.dart';
import '../../../core/models/users/user_profile.dart';

abstract interface class UserRepository {
  Future<MeResponse> getMe();

  Future<UserProfile> updateProfile(UpdateProfileRequest request);

  Future<AiSettings> getAiSettings();

  Future<AiSettings> updateAiSettings(UpdateAiSettingsRequest request);

  Future<UserProfile> completeOnboarding();
}
