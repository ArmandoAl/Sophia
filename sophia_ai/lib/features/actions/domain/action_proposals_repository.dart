abstract class ActionProposalsRepository {
  Future<void> confirm(
    String proposalId, {
    Map<String, dynamic>? correctedInput,
    int? decisionLatencyMs,
  });

  Future<void> reject(
    String proposalId, {
    String? rejectionReason,
    int? decisionLatencyMs,
  });

  Future<void> execute(String proposalId);
}
