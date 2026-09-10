class DeviceToken {
  const DeviceToken({
    required this.id,
    required this.platform,
    required this.createdAt,
    required this.tokenRedacted,
  });

  final String id;
  final String platform;
  final DateTime createdAt;
  final String tokenRedacted;

  factory DeviceToken.fromJson(Map<String, dynamic> json) => DeviceToken(
    id: json['id'] as String,
    platform: json['platform'] as String,
    createdAt: DateTime.parse(json['created_at'] as String),
    tokenRedacted: json['token_redacted'] as String? ?? '',
  );
}

class IngestionBatch {
  const IngestionBatch({
    required this.id,
    required this.status,
    required this.beliefsCreated,
    required this.createdAt,
  });

  final String id;
  final String status;
  final int beliefsCreated;
  final DateTime createdAt;

  factory IngestionBatch.fromJson(Map<String, dynamic> json) => IngestionBatch(
    id: json['id'] as String,
    status: json['status'] as String,
    beliefsCreated: json['beliefs_created'] as int? ?? 0,
    createdAt: DateTime.parse(json['created_at'] as String),
  );
}
