import '../../../core/network/api_client.dart';
import '../domain/action_proposals_repository.dart';

class ActionProposalsRepositoryImpl implements ActionProposalsRepository {
  ActionProposalsRepositoryImpl(this._api);

  final ApiClient _api;

  @override
  Future<void> confirm(
    String proposalId, {
    Map<String, dynamic>? correctedInput,
    int? decisionLatencyMs,
  }) async {
    await _api.post(
      '/ai/action-proposals/$proposalId/confirm',
      body: _optionalBody(
        correctedInput: correctedInput,
        decisionLatencyMs: decisionLatencyMs,
      ),
    );
  }

  @override
  Future<void> reject(
    String proposalId, {
    String? rejectionReason,
    int? decisionLatencyMs,
  }) async {
    await _api.post(
      '/ai/action-proposals/$proposalId/reject',
      body: _optionalBody(
        rejectionReason: rejectionReason,
        decisionLatencyMs: decisionLatencyMs,
      ),
    );
  }

  @override
  Future<void> execute(String proposalId) async {
    await _api.post('/ai/action-proposals/$proposalId/execute');
  }

  Map<String, dynamic>? _optionalBody({
    Map<String, dynamic>? correctedInput,
    String? rejectionReason,
    int? decisionLatencyMs,
  }) {
    final body = <String, dynamic>{};
    if (correctedInput != null) {
      body['corrected_input'] = correctedInput;
    }
    if (rejectionReason != null && rejectionReason.isNotEmpty) {
      body['rejection_reason'] = rejectionReason;
    }
    if (decisionLatencyMs != null) {
      body['decision_latency_ms'] = decisionLatencyMs;
    }
    return body.isEmpty ? null : body;
  }
}
