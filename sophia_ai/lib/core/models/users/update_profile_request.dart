import 'package:equatable/equatable.dart';

/// PATCH `/users/me/profile` — flat partial body (no `{profile: ...}` wrapper).
///
/// Null fields are omitted. Never sends `user_id`.
class UpdateProfileRequest extends Equatable {
  const UpdateProfileRequest({
    this.displayName,
    this.preferredName,
    this.timezone,
    this.locale,
    this.avatarUrl,
  });

  final String? displayName;
  final String? preferredName;
  final String? timezone;
  final String? locale;
  final String? avatarUrl;

  /// Partial map: only non-null fields.
  Map<String, dynamic> toJson() {
    return <String, dynamic>{
      if (displayName != null) 'display_name': displayName,
      if (preferredName != null) 'preferred_name': preferredName,
      if (timezone != null) 'timezone': timezone,
      if (locale != null) 'locale': locale,
      if (avatarUrl != null) 'avatar_url': avatarUrl,
    };
  }

  bool get isEmpty => toJson().isEmpty;

  @override
  List<Object?> get props => [
    displayName,
    preferredName,
    timezone,
    locale,
    avatarUrl,
  ];
}
