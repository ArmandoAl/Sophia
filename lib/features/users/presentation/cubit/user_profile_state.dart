import 'package:equatable/equatable.dart';

import '../../../../core/models/users/user_profile.dart';
import '../../../../core/network/api_exception.dart';

sealed class UserProfileState extends Equatable {
  const UserProfileState();

  @override
  List<Object?> get props => [];
}

final class UserProfileInitial extends UserProfileState {
  const UserProfileInitial();
}

final class UserProfileLoading extends UserProfileState {
  const UserProfileLoading();
}

final class UserProfileLoaded extends UserProfileState {
  const UserProfileLoaded(this.profile, {this.saving = false});

  final UserProfile profile;
  final bool saving;

  @override
  List<Object?> get props => [profile, saving];
}

final class UserProfileFailure extends UserProfileState {
  const UserProfileFailure({
    required this.message,
    this.exception,
    this.previous,
  });

  final String message;
  final ApiException? exception;
  final UserProfile? previous;

  @override
  List<Object?> get props => [message, exception, previous];
}
