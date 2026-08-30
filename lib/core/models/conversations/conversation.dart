import 'package:equatable/equatable.dart';

class Conversation extends Equatable {
  const Conversation({
    required this.id,
    required this.userId,
    required this.title,
    required this.status,
    required this.createdAt,
    required this.updatedAt,
    this.lastMessageAt,
  });

  final String id;
  final String userId;
  final String title;
  final String status;
  final DateTime createdAt;
  final DateTime updatedAt;
  final DateTime? lastMessageAt;

  factory Conversation.fromJson(Map<String, dynamic> json) {
    return Conversation(
      id: json['id'] as String,
      userId: json['user_id'] as String,
      title: json['title'] as String? ?? '',
      status: json['status'] as String? ?? 'active',
      createdAt: DateTime.parse(json['created_at'] as String),
      updatedAt: DateTime.parse(json['updated_at'] as String),
      lastMessageAt: json['last_message_at'] == null
          ? null
          : DateTime.parse(json['last_message_at'] as String),
    );
  }

  Map<String, dynamic> toJson() => {
    'id': id,
    'user_id': userId,
    'title': title,
    'status': status,
    'created_at': createdAt.toUtc().toIso8601String(),
    'updated_at': updatedAt.toUtc().toIso8601String(),
    if (lastMessageAt != null)
      'last_message_at': lastMessageAt!.toUtc().toIso8601String(),
  };

  @override
  List<Object?> get props => [
    id,
    userId,
    title,
    status,
    createdAt,
    updatedAt,
    lastMessageAt,
  ];
}
