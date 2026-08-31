import '../../../core/network/api_client.dart';
import '../domain/action_proposals_repository.dart';

class ActionProposalsRepositoryImpl implements ActionProposalsRepository {
  ActionProposalsRepositoryImpl(this._api);

  final ApiClient _api;

  @override
  Future<void> confirm(String proposalId) async {
    await _api.post('/ai/action-proposals/$proposalId/confirm');
  }

  @override
  Future<void> reject(String proposalId) async {
    await _api.post('/ai/action-proposals/$proposalId/reject');
  }

  @override
  Future<void> execute(String proposalId) async {
    await _api.post('/ai/action-proposals/$proposalId/execute');
  }
}
