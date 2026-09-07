import '../conversations/conversation.dart';
import '../conversations/conversation_message.dart';

class ConversationListResponse {
  const ConversationListResponse({
    required this.conversations,
    this.nextCursor,
  });

  final List<Conversation> conversations;
  final String? nextCursor;

  factory ConversationListResponse.fromJson(Map<String, dynamic> json) {
    return ConversationListResponse(
      conversations: ((json['conversations'] as List?) ?? const [])
          .map((item) => Conversation.fromJson(item as Map<String, dynamic>))
          .toList(),
      nextCursor: json['next_cursor'] as String?,
    );
  }
}

class ConversationMessagesResponse {
  const ConversationMessagesResponse({required this.messages, this.nextCursor});

  final List<ConversationMessage> messages;
  final String? nextCursor;

  factory ConversationMessagesResponse.fromJson(Map<String, dynamic> json) {
    return ConversationMessagesResponse(
      messages: ((json['messages'] as List?) ?? const [])
          .map(
            (item) =>
                ConversationMessage.fromJson(item as Map<String, dynamic>),
          )
          .toList(),
      nextCursor: json['next_cursor'] as String?,
    );
  }
}

class RuntimeProposedAction {
  const RuntimeProposedAction({
    this.id,
    required this.toolName,
    this.status,
    required this.proposedInput,
    required this.reason,
    required this.riskLevel,
    required this.requiresConfirmation,
  });

  final String? id;
  final String toolName;
  final String? status;
  final Map<String, dynamic> proposedInput;
  final String reason;
  final String riskLevel;
  final bool requiresConfirmation;

  factory RuntimeProposedAction.fromJson(Map<String, dynamic> json) {
    return RuntimeProposedAction(
      id: json['id'] as String?,
      toolName: json['tool_name'] as String? ?? '',
      status: json['status'] as String?,
      proposedInput:
          (json['proposed_input'] as Map?)?.cast<String, dynamic>() ??
          const <String, dynamic>{},
      reason: json['reason'] as String? ?? '',
      riskLevel: json['risk_level'] as String? ?? 'medium',
      requiresConfirmation: json['requires_confirmation'] as bool? ?? true,
    );
  }
}

class SendConversationMessageResponse {
  const SendConversationMessageResponse({
    required this.conversation,
    required this.userMessage,
    required this.assistantMessage,
    required this.proposedActions,
    this.runtimeRequestId,
  });

  final Conversation conversation;
  final ConversationMessage userMessage;
  final ConversationMessage assistantMessage;
  final List<RuntimeProposedAction> proposedActions;
  final String? runtimeRequestId;

  factory SendConversationMessageResponse.fromJson(Map<String, dynamic> json) {
    return SendConversationMessageResponse(
      conversation: Conversation.fromJson(
        json['conversation'] as Map<String, dynamic>,
      ),
      userMessage: ConversationMessage.fromJson(
        json['user_message'] as Map<String, dynamic>,
      ),
      assistantMessage: ConversationMessage.fromJson(
        json['assistant_message'] as Map<String, dynamic>,
      ),
      proposedActions: ((json['proposed_actions'] as List?) ?? const [])
          .map(
            (item) =>
                RuntimeProposedAction.fromJson(item as Map<String, dynamic>),
          )
          .toList(),
      runtimeRequestId: json['runtime_request_id'] as String?,
    );
  }
}
