class Belief {
  const Belief({
    required this.id,
    required this.statement,
    required this.category,
    required this.scope,
    required this.scopeKey,
    required this.confidence,
    required this.decayedConfidence,
    required this.evidenceCount,
    required this.contradictionCount,
    required this.trustTier,
    required this.promptSlot,
    this.status = 'active',
    this.subjectType = 'user',
    this.subjectId = '',
    this.factKind = '',
    this.validUntil,
    this.followUpAt,
    this.sensitive = false,
  });
  final String id, statement, category, scope, scopeKey, promptSlot;
  final String status, subjectType, subjectId, factKind;
  final double confidence, decayedConfidence;
  final int evidenceCount, contradictionCount, trustTier;
  final DateTime? validUntil, followUpAt;
  final bool sensitive;
  factory Belief.fromJson(Map<String, dynamic> json) => Belief(
    id: json['id'] as String,
    statement: json['statement'] as String,
    category: json['category'] as String,
    scope: json['scope'] as String,
    scopeKey: json['scope_key'] as String? ?? '',
    confidence: (json['confidence'] as num).toDouble(),
    decayedConfidence: (json['decayed_confidence'] as num).toDouble(),
    evidenceCount: json['evidence_count'] as int,
    contradictionCount: json['contradiction_count'] as int,
    trustTier: json['trust_tier'] as int,
    promptSlot: json['prompt_slot'] as String? ?? '',
    status: json['status'] as String? ?? 'active',
    subjectType: json['subject_type'] as String? ?? 'user',
    subjectId: json['subject_id'] as String? ?? '',
    factKind: json['fact_kind'] as String? ?? '',
    validUntil: DateTime.tryParse(
      json['valid_until'] as String? ?? '',
    )?.toLocal(),
    followUpAt: DateTime.tryParse(
      json['follow_up_at'] as String? ?? '',
    )?.toLocal(),
    sensitive: json['sensitive'] as bool? ?? false,
  );
}

class PromptVersion {
  const PromptVersion({
    required this.version,
    required this.content,
    required this.tokenCount,
  });
  final int version, tokenCount;
  final String content;
  factory PromptVersion.fromJson(Map<String, dynamic> json) => PromptVersion(
    version: json['version'] as int,
    content: json['content'] as String,
    tokenCount: json['token_count'] as int,
  );
}

class LearningSummary {
  const LearningSummary({
    required this.date,
    required this.approved,
    required this.corrected,
    required this.rejected,
    required this.inputTokens,
    required this.outputTokens,
  });
  final String date;
  final int approved, corrected, rejected, inputTokens, outputTokens;
  factory LearningSummary.fromJson(Map<String, dynamic> json) {
    final stats = json['stats'] as Map<String, dynamic>? ?? const {};
    final cost = json['synthesis_cost'] as Map<String, dynamic>? ?? const {};
    return LearningSummary(
      date: json['date'] as String,
      approved: stats['approved_direct'] as int? ?? 0,
      corrected: stats['approved_corrected'] as int? ?? 0,
      rejected: stats['rejected'] as int? ?? 0,
      inputTokens: cost['input_tokens'] as int? ?? 0,
      outputTokens: cost['output_tokens'] as int? ?? 0,
    );
  }
}
