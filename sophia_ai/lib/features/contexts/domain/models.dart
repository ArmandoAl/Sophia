class UserContext {
  const UserContext({
    required this.id,
    required this.kind,
    required this.slug,
    required this.label,
    required this.aliases,
    required this.active,
  });
  final String id, kind, slug, label;
  final List<String> aliases;
  final bool active;
  String get scopeKey => '$kind:$slug';
  factory UserContext.fromJson(Map<String, dynamic> json) => UserContext(
    id: json['id'] as String,
    kind: json['kind'] as String,
    slug: json['slug'] as String,
    label: json['label'] as String,
    aliases: (json['aliases'] as List<dynamic>? ?? const []).cast<String>(),
    active: json['active'] as bool? ?? true,
  );
}
