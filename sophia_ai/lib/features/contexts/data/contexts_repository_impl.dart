import '../../../core/network/api_client.dart';
import '../../../core/network/require_json_map.dart';
import '../domain/contexts_repository.dart';
import '../domain/models.dart';

class ContextsRepositoryImpl implements ContextsRepository {
  ContextsRepositoryImpl(this._api);
  final ApiClient _api;
  UserContext _context(dynamic body, String operation) =>
      UserContext.fromJson(requireJsonMap(body, context: operation));
  @override
  Future<List<UserContext>> list() async {
    final body = requireJsonMap(
      await _api.get('/contexts'),
      context: 'GET /contexts',
    );
    return (body['contexts'] as List<dynamic>? ?? const [])
        .map((value) => UserContext.fromJson(value as Map<String, dynamic>))
        .toList();
  }

  @override
  Future<UserContext> create({
    required String kind,
    required String slug,
    required String label,
    required List<String> aliases,
  }) async => _context(
    await _api.post(
      '/contexts',
      body: {'kind': kind, 'slug': slug, 'label': label, 'aliases': aliases},
    ),
    'POST /contexts',
  );
  @override
  Future<UserContext> update(
    String id, {
    required String label,
    required List<String> aliases,
  }) async => _context(
    await _api.patch(
      '/contexts/$id',
      body: {'label': label, 'aliases': aliases},
    ),
    'PATCH /contexts/{id}',
  );
  @override
  Future<void> archive(String id) async {
    await _api.post('/contexts/$id/archive');
  }
}
