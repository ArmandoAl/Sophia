import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:sophia_core/core/storage/token_storage.dart';

/// Platform-backed secure storage.
///
/// - **iOS/Android**: Keychain / Keystore via `flutter_secure_storage`.
/// - **Web**: falls back to browser storage (less secure than mobile). Prefer
///   short-lived sessions and HTTPS; do not treat web storage as vault-grade.
class SecureTokenStorage implements TokenStorage {
  SecureTokenStorage({FlutterSecureStorage? storage})
    : _storage = storage ?? const FlutterSecureStorage();

  static const _tokenKey = 'sofia_access_token';

  final FlutterSecureStorage _storage;

  @override
  Future<String?> readToken() => _storage.read(key: _tokenKey);

  @override
  Future<void> saveToken(String token) =>
      _storage.write(key: _tokenKey, value: token);

  @override
  Future<void> clearToken() => _storage.delete(key: _tokenKey);
}
