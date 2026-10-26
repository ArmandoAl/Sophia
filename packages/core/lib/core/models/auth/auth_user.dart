import 'package:equatable/equatable.dart';

/// Auth user shape shared by register, login, and `/users/me`.
///
/// Register/login omit [role]. GET `/users/me` includes it.
class AuthUser extends Equatable {
  const AuthUser({
    required this.id,
    required this.name,
    required this.email,
    this.role,
  });

  final String id;
  final String name;
  final String email;

  /// Present on `/users/me`; absent on register/login responses.
  final String? role;

  factory AuthUser.fromJson(Map<String, dynamic> json) {
    return AuthUser(
      id: json['id'] as String,
      name: json['name'] as String,
      email: json['email'] as String,
      role: json['role'] as String?,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'name': name,
      'email': email,
      if (role != null) 'role': role,
    };
  }

  @override
  List<Object?> get props => [id, name, email, role];
}
