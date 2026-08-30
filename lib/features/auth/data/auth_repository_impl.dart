import '../../../core/models/auth/auth_user.dart';
import '../../../core/models/auth/login_request.dart';
import '../../../core/models/auth/login_response.dart';
import '../../../core/models/auth/register_request.dart';
import '../../../core/network/api_client.dart';
import '../../../core/network/require_json_map.dart';
import '../../../core/storage/token_storage.dart';
import '../domain/auth_repository.dart';

class AuthRepositoryImpl implements AuthRepository {
  AuthRepositoryImpl({
    required ApiClient api,
    required TokenStorage tokenStorage,
  }) : _api = api,
       _tokenStorage = tokenStorage;

  final ApiClient _api;
  final TokenStorage _tokenStorage;

  @override
  Future<AuthUser> register(RegisterRequest request) async {
    final body = await _api.post(
      '/auth/register',
      body: request.toJson(),
      authenticated: false,
    );
    return AuthUser.fromJson(
      requireJsonMap(body, context: 'POST /auth/register'),
    );
  }

  @override
  Future<LoginResponse> login(LoginRequest request) async {
    final body = await _api.post(
      '/auth/login',
      body: request.toJson(),
      authenticated: false,
    );
    final response = LoginResponse.fromJson(
      requireJsonMap(body, context: 'POST /auth/login'),
    );
    await _tokenStorage.saveToken(response.token);
    return response;
  }

  @override
  Future<void> logout() => _tokenStorage.clearToken();
}
