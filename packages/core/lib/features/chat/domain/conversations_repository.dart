import '../../../core/models/models.dart';

abstract class ConversationsRepository {
  Future<Conversation> createConversation(CreateConversationRequest request);
  Future<ConversationListResponse> listConversations({
    String? status,
    int? limit,
    String? cursor,
  });
  Future<Conversation> getConversation(String id);
  Future<ConversationMessagesResponse> listMessages(
    String conversationId, {
    int? limit,
    String? cursor,
  });
  Future<SendConversationMessageResponse> sendMessage(
    String conversationId,
    SendConversationMessageRequest request,
  );
  Future<Conversation> archiveConversation(String id);
}
