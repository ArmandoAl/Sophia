import 'package:equatable/equatable.dart';

import '../../../../core/models/auth/auth_user.dart';
import '../../../../core/models/users/ai_settings.dart';
import '../../../../core/models/users/user_profile.dart';

/// Global session lifecycle for Sofia F1.
sealed class SessionState extends Equatable {
  const SessionState();

  @override
  List<Object?> get props => [];
}

final class SessionInitial extends SessionState {
  const SessionInitial();
}

final class SessionLoading extends SessionState {
  const SessionLoading();
}

final class SessionUnauthenticated extends SessionState {
  const SessionUnauthenticated();
}

final class SessionAuthenticated extends SessionState {
  const SessionAuthenticated({
    required this.user,
    required this.profile,
    required this.aiSettings,
  });

  final AuthUser user;
  final UserProfile profile;
  final AiSettings aiSettings;

  SessionAuthenticated copyWith({
    AuthUser? user,
    UserProfile? profile,
    AiSettings? aiSettings,
  }) {
    return SessionAuthenticated(
      user: user ?? this.user,
      profile: profile ?? this.profile,
      aiSettings: aiSettings ?? this.aiSettings,
    );
  }

  @override
  List<Object?> get props => [user, profile, aiSettings];
}

/// Recoverable failure. [tokenRetained] is true for non-401 errors during bootstrap.
final class SessionFailure extends SessionState {
  const SessionFailure({
    required this.message,
    this.error,
    this.tokenRetained = true,
  });

  final String message;
  final Object? error;
  final bool tokenRetained;

  @override
  List<Object?> get props => [message, error, tokenRetained];
}
