import 'models.dart';

abstract interface class BeliefsRepository {
  Future<List<Belief>> list({String? scope, String? scopeKey});
  Future<void> retire(String id);
  Future<Belief> updateStatement(String id, String statement);
  Future<PromptVersion?> getPromptVersion();
  Future<List<LearningSummary>> listSummaries({int limit = 7});
}
