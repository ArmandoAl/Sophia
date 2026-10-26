import '../../../core/models/users/ai_settings.dart';
import '../../../core/models/users/me_response.dart';
import '../../../core/models/users/update_ai_settings_request.dart';
import '../../../core/models/users/update_profile_request.dart';
import '../../../core/models/users/user_profile.dart';
import '../../../core/network/api_client.dart';
import '../../../core/network/require_json_map.dart';
import '../domain/user_repository.dart';

class UserRepositoryImpl implements UserRepository {
  UserRepositoryImpl(this._api);

  final ApiClient _api;

  @override
  Future<MeResponse> getMe() async {
    final body = await _api.get('/users/me');
    return MeResponse.fromJson(requireJsonMap(body, context: 'GET /users/me'));
  }

  @override
  Future<UserProfile> updateProfile(UpdateProfileRequest request) async {
    final body = await _api.patch('/users/me/profile', body: request.toJson());
    return UserProfile.fromJson(
      requireJsonMap(body, context: 'PATCH /users/me/profile'),
    );
  }

  @override
  Future<AiSettings> getAiSettings() async {
    final body = await _api.get('/users/me/ai-settings');
    return AiSettings.fromJson(
      requireJsonMap(body, context: 'GET /users/me/ai-settings'),
    );
  }

  @override
  Future<AiSettings> updateAiSettings(UpdateAiSettingsRequest request) async {
    final body = await _api.patch(
      '/users/me/ai-settings',
      body: request.toJson(),
    );
    return AiSettings.fromJson(
      requireJsonMap(body, context: 'PATCH /users/me/ai-settings'),
    );
  }

  @override
  Future<UserProfile> completeOnboarding() async {
    // Backend ignores body; send empty object for Content-Type consistency.
    final body = await _api.post(
      '/users/me/onboarding/complete',
      body: const <String, dynamic>{},
    );
    return UserProfile.fromJson(
      requireJsonMap(body, context: 'POST /users/me/onboarding/complete'),
    );
  }
}
