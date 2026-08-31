abstract class ActionProposalsRepository {
  Future<void> confirm(String proposalId);

  Future<void> reject(String proposalId);

  Future<void> execute(String proposalId);
}
