import 'package:equatable/equatable.dart';

import '../auth/auth_user.dart';
import 'ai_settings.dart';
import 'user_profile.dart';

/// GET `/users/me` — nested wrapper `{ user, profile, ai_settings }`.
class MeResponse extends Equatable {
  const MeResponse({
    required this.user,
    required this.profile,
    required this.aiSettings,
  });

  final AuthUser user;
  final UserProfile profile;
  final AiSettings aiSettings;

  factory MeResponse.fromJson(Map<String, dynamic> json) {
    return MeResponse(
      user: AuthUser.fromJson(json['user'] as Map<String, dynamic>),
      profile: UserProfile.fromJson(json['profile'] as Map<String, dynamic>),
      aiSettings: AiSettings.fromJson(
        json['ai_settings'] as Map<String, dynamic>,
      ),
    );
  }

  Map<String, dynamic> toJson() => {
    'user': user.toJson(),
    'profile': profile.toJson(),
    'ai_settings': aiSettings.toJson(),
  };

  @override
  List<Object?> get props => [user, profile, aiSettings];
}
