import 'package:sophia_core/core/storage/token_storage.dart';
import 'package:web/web.dart' as web;

/// JWT en `sessionStorage`: dura la pestaña. En la web no hay un almacén tipo Keychain; no es una bóveda.
class BrowserTokenStorage implements TokenStorage {
  static const _key = 'sofia_access_token';

  @override
  Future<String?> readToken() async => _safe<String?>(() => web.window.sessionStorage.getItem(_key));

  @override
  Future<void> saveToken(String token) async => _safe<void>(() => web.window.sessionStorage.setItem(_key, token));

  @override
  Future<void> clearToken() async => _safe<void>(() => web.window.sessionStorage.removeItem(_key));

  static T? _safe<T>(T Function() action) {
    try {
      return action();
    } on Object {
      return null;
    }
  }
}
