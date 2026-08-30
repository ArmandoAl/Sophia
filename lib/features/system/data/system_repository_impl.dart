import '../../../core/models/system/health_response.dart';
import '../../../core/network/api_client.dart';
import '../../../core/network/require_json_map.dart';
import '../domain/system_repository.dart';

class SystemRepositoryImpl implements SystemRepository {
  SystemRepositoryImpl(this._api);

  final ApiClient _api;

  @override
  Future<HealthResponse> getHealth() async {
    final body = await _api.get('/health', authenticated: false);
    return HealthResponse.fromJson(
      requireJsonMap(body, context: 'GET /health'),
    );
  }
}
