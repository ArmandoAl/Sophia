import 'package:equatable/equatable.dart';

import '../../../../core/models/auth/auth_user.dart';
import '../../../../core/network/api_exception.dart';

sealed class AuthState extends Equatable {
  const AuthState();

  @override
  List<Object?> get props => [];
}

final class AuthInitial extends AuthState {
  const AuthInitial();
}

final class AuthLoading extends AuthState {
  const AuthLoading();
}

/// Login succeeded and [SessionCubit] was hydrated.
final class AuthLoginSuccess extends AuthState {
  const AuthLoginSuccess();
}

/// Register succeeded. No token — user must login (no auto-login).
final class AuthRegisterSuccess extends AuthState {
  const AuthRegisterSuccess(this.user);

  final AuthUser user;

  @override
  List<Object?> get props => [user];
}

final class AuthFailure extends AuthState {
  const AuthFailure({required this.message, this.exception});

  final String message;
  final ApiException? exception;

  int? get statusCode => exception?.statusCode;

  @override
  List<Object?> get props => [message, exception];
}
