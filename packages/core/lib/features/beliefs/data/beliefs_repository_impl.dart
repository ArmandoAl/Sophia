import '../../../core/network/api_client.dart';
import '../../../core/network/api_exception.dart';
import '../../../core/network/require_json_map.dart';
import '../domain/beliefs_repository.dart';
import '../domain/models.dart';

class BeliefsRepositoryImpl implements BeliefsRepository {
  BeliefsRepositoryImpl(this._api);
  final ApiClient _api;

  @override
  Future<List<Belief>> list({String? scope, String? scopeKey}) async {
    final query = <String, String>{};
    if (scope != null) query['scope'] = scope;
    if (scopeKey != null) query['scope_key'] = scopeKey;
    final body = requireJsonMap(
      await _api.get(
        '/learning/beliefs',
        queryParameters: query.isEmpty ? null : query,
      ),
      context: 'GET /learning/beliefs',
    );
    return (body['beliefs'] as List<dynamic>? ?? const [])
        .map((value) => Belief.fromJson(value as Map<String, dynamic>))
        .toList();
  }

  @override
  Future<void> retire(String id) async {
    await _api.delete('/learning/beliefs/$id');
  }

  @override
  Future<Belief> updateStatement(String id, String statement) async =>
      Belief.fromJson(
        requireJsonMap(
          await _api.patch(
            '/learning/beliefs/$id',
            body: {'statement': statement},
          ),
          context: 'PATCH /learning/beliefs/{id}',
        ),
      );

  @override
  Future<PromptVersion?> getPromptVersion() async {
    try {
      return PromptVersion.fromJson(
        requireJsonMap(
          await _api.get('/learning/prompt-version'),
          context: 'GET /learning/prompt-version',
        ),
      );
    } on ApiException catch (error) {
      if (error.statusCode == 404) return null;
      rethrow;
    }
  }

  @override
  Future<List<LearningSummary>> listSummaries({int limit = 7}) async {
    final body = requireJsonMap(
      await _api.get(
        '/learning/summaries',
        queryParameters: {'limit': '$limit'},
      ),
      context: 'GET /learning/summaries',
    );
    return (body['summaries'] as List<dynamic>? ?? const [])
        .map((value) => LearningSummary.fromJson(value as Map<String, dynamic>))
        .toList();
  }
}
