import '../../../core/network/api_client.dart';
import '../../../core/network/require_json_map.dart';
import '../../beliefs/domain/models.dart';
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

  @override
  Future<List<UserContext>> listEntities({String? kind, String? status}) async {
    final query = <String, String>{
      if (kind != null) 'kind': kind,
      if (status != null) 'status': status,
    };
    final body = requireJsonMap(
      await _api.get(
        '/entities',
        queryParameters: query.isEmpty ? null : query,
      ),
      context: 'GET /entities',
    );
    return (body['entities'] as List<dynamic>? ?? const [])
        .map((value) => UserContext.fromJson(value as Map<String, dynamic>))
        .toList();
  }

  @override
  Future<UserContext> createEntity({
    required String kind,
    required String slug,
    required String label,
    required String relationship,
    required List<String> aliases,
  }) async => _context(
    await _api.post(
      '/entities',
      body: {
        'kind': kind,
        'slug': slug,
        'label': label,
        'relationship': relationship,
        'aliases': aliases,
      },
    ),
    'POST /entities',
  );

  @override
  Future<EntityDetails> getEntity(String id) async {
    final body = requireJsonMap(
      await _api.get('/entities/$id'),
      context: 'GET /entities/{id}',
    );
    return EntityDetails(
      entity: UserContext.fromJson(body['entity'] as Map<String, dynamic>),
      facts: (body['facts'] as List<dynamic>? ?? const [])
          .map((value) => Belief.fromJson(value as Map<String, dynamic>))
          .toList(),
      userBeliefs: (body['user_beliefs'] as List<dynamic>? ?? const [])
          .map((value) => Belief.fromJson(value as Map<String, dynamic>))
          .toList(),
      episodes: (body['episodes'] as List<dynamic>? ?? const [])
          .map((value) => Episode.fromJson(value as Map<String, dynamic>))
          .toList(),
    );
  }

  @override
  Future<UserContext> updateEntity(
    String id,
    Map<String, dynamic> changes,
  ) async => _context(
    await _api.patch('/entities/$id', body: changes),
    'PATCH /entities/{id}',
  );

  @override
  Future<UserContext> mergeEntity(String id, String into) async => _context(
    await _api.post('/entities/$id/merge', body: {'into': into}),
    'POST /entities/{id}/merge',
  );

  @override
  Future<void> archiveEntity(String id) => _api.delete('/entities/$id');
}
