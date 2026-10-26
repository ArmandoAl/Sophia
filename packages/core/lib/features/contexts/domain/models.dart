class UserContext {
  const UserContext({
    required this.id,
    required this.kind,
    required this.slug,
    required this.label,
    required this.aliases,
    required this.active,
    this.relationship = '',
    this.createdFrom = 'explicit',
    this.status = 'active',
    this.mergedInto = '',
    this.threadsMuted = false,
  });
  final String id, kind, slug, label;
  final List<String> aliases;
  final bool active;
  final String relationship, createdFrom, status, mergedInto;
  final bool threadsMuted;
  String get scopeKey => '$kind:$slug';
  factory UserContext.fromJson(Map<String, dynamic> json) => UserContext(
    id: json['id'] as String,
    kind: json['kind'] as String,
    slug: json['slug'] as String,
    label: json['label'] as String,
    aliases: (json['aliases'] as List<dynamic>? ?? const []).cast<String>(),
    active: json['active'] as bool? ?? true,
    relationship: json['relationship'] as String? ?? '',
    createdFrom: json['created_from'] as String? ?? 'explicit',
    status: json['status'] as String? ?? 'active',
    mergedInto: json['merged_into'] as String? ?? '',
    threadsMuted: json['threads_muted'] as bool? ?? false,
  );
}

class Episode {
  const Episode({
    required this.id,
    required this.occurredAt,
    required this.summary,
    required this.topics,
  });

  final String id, summary;
  final DateTime occurredAt;
  final List<String> topics;

  factory Episode.fromJson(Map<String, dynamic> json) => Episode(
    id: json['id'] as String,
    occurredAt: DateTime.parse(json['occurred_at'] as String).toLocal(),
    summary: json['summary'] as String,
    topics: (json['topics'] as List<dynamic>? ?? const []).cast<String>(),
  );
}
