import 'package:equatable/equatable.dart';

import 'auth_user.dart';

/// POST `/auth/login` success body.
class LoginResponse extends Equatable {
  const LoginResponse({required this.token, required this.user});

  final String token;
  final AuthUser user;

  factory LoginResponse.fromJson(Map<String, dynamic> json) {
    return LoginResponse(
      token: json['token'] as String,
      user: AuthUser.fromJson(json['user'] as Map<String, dynamic>),
    );
  }

  Map<String, dynamic> toJson() => {'token': token, 'user': user.toJson()};

  @override
  List<Object?> get props => [token, user];
}
