import '../../../core/models/auth/auth_user.dart';
import '../../../core/models/auth/login_request.dart';
import '../../../core/models/auth/login_response.dart';
import '../../../core/models/auth/register_request.dart';

abstract interface class AuthRepository {
  Future<AuthUser> register(RegisterRequest request);

  Future<LoginResponse> login(LoginRequest request);

  /// Clears the locally stored JWT. No backend call.
  Future<void> logout();
}
