import 'package:equatable/equatable.dart';

/// Flat user profile (PATCH profile, onboarding, nested in MeResponse).
class UserProfile extends Equatable {
  const UserProfile({
    required this.id,
    required this.userId,
    required this.displayName,
    required this.preferredName,
    required this.timezone,
    required this.locale,
    this.avatarUrl,
    required this.onboardingCompleted,
    required this.createdAt,
    required this.updatedAt,
  });

  final String id;
  final String userId;
  final String displayName;
  final String preferredName;
  final String timezone;
  final String locale;
  final String? avatarUrl;
  final bool onboardingCompleted;
  final DateTime createdAt;
  final DateTime updatedAt;

  factory UserProfile.fromJson(Map<String, dynamic> json) {
    final avatar = json['avatar_url'] as String?;
    return UserProfile(
      id: json['id'] as String,
      userId: json['user_id'] as String,
      displayName: json['display_name'] as String,
      preferredName: json['preferred_name'] as String,
      timezone: json['timezone'] as String,
      locale: json['locale'] as String,
      avatarUrl: (avatar == null || avatar.isEmpty) ? null : avatar,
      onboardingCompleted: json['onboarding_completed'] as bool,
      createdAt: DateTime.parse(json['created_at'] as String),
      updatedAt: DateTime.parse(json['updated_at'] as String),
    );
  }

  Map<String, dynamic> toJson() => {
    'id': id,
    'user_id': userId,
    'display_name': displayName,
    'preferred_name': preferredName,
    'timezone': timezone,
    'locale': locale,
    if (avatarUrl != null && avatarUrl!.isNotEmpty) 'avatar_url': avatarUrl,
    'onboarding_completed': onboardingCompleted,
    'created_at': createdAt.toUtc().toIso8601String(),
    'updated_at': updatedAt.toUtc().toIso8601String(),
  };

  @override
  List<Object?> get props => [
    id,
    userId,
    displayName,
    preferredName,
    timezone,
    locale,
    avatarUrl,
    onboardingCompleted,
    createdAt,
    updatedAt,
  ];
}
