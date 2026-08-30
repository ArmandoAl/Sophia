import 'package:equatable/equatable.dart';

class ConversationMessage extends Equatable {
  const ConversationMessage({
    required this.id,
    required this.conversationId,
    required this.userId,
    required this.role,
    required this.content,
    required this.createdAt,
    this.runtimeRequestId,
  });

  final String id;
  final String conversationId;
  final String userId;
  final String role;
  final String content;
  final String? runtimeRequestId;
  final DateTime createdAt;

  bool get isUser => role == 'user';
  bool get isAssistant => role == 'assistant';

  factory ConversationMessage.fromJson(Map<String, dynamic> json) {
    final runtimeRequestId = json['runtime_request_id'] as String?;
    return ConversationMessage(
      id: json['id'] as String,
      conversationId: json['conversation_id'] as String,
      userId: json['user_id'] as String,
      role: json['role'] as String,
      content: json['content'] as String? ?? '',
      runtimeRequestId: runtimeRequestId == null || runtimeRequestId.isEmpty
          ? null
          : runtimeRequestId,
      createdAt: DateTime.parse(json['created_at'] as String),
    );
  }

  Map<String, dynamic> toJson() => {
    'id': id,
    'conversation_id': conversationId,
    'user_id': userId,
    'role': role,
    'content': content,
    if (runtimeRequestId != null) 'runtime_request_id': runtimeRequestId,
    'created_at': createdAt.toUtc().toIso8601String(),
  };

  @override
  List<Object?> get props => [
    id,
    conversationId,
    userId,
    role,
    content,
    runtimeRequestId,
    createdAt,
  ];
}
