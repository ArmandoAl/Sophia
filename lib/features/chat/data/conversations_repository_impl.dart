import '../../../core/models/models.dart';
import '../../../core/network/api_client.dart';
import '../../../core/network/require_json_map.dart';
import '../domain/conversations_repository.dart';

class ConversationsRepositoryImpl implements ConversationsRepository {
  ConversationsRepositoryImpl(this._api);

  final ApiClient _api;

  @override
  Future<Conversation> createConversation(
    CreateConversationRequest request,
  ) async {
    final body = await _api.post('/conversations', body: request.toJson());
    final map = requireJsonMap(body, context: 'POST /conversations');
    return Conversation.fromJson(map['conversation'] as Map<String, dynamic>);
  }

  @override
  Future<ConversationListResponse> listConversations({
    String? status,
    int? limit,
    String? cursor,
  }) async {
    final body = await _api.get(
      '/conversations',
      queryParameters: _query(status: status, limit: limit, cursor: cursor),
    );
    return ConversationListResponse.fromJson(
      requireJsonMap(body, context: 'GET /conversations'),
    );
  }

  @override
  Future<Conversation> getConversation(String id) async {
    final body = await _api.get('/conversations/$id');
    final map = requireJsonMap(body, context: 'GET /conversations/{id}');
    return Conversation.fromJson(map['conversation'] as Map<String, dynamic>);
  }

  @override
  Future<ConversationMessagesResponse> listMessages(
    String conversationId, {
    int? limit,
    String? cursor,
  }) async {
    final body = await _api.get(
      '/conversations/$conversationId/messages',
      queryParameters: _query(limit: limit, cursor: cursor),
    );
    return ConversationMessagesResponse.fromJson(
      requireJsonMap(body, context: 'GET /conversations/{id}/messages'),
    );
  }

  @override
  Future<SendConversationMessageResponse> sendMessage(
    String conversationId,
    SendConversationMessageRequest request,
  ) async {
    final body = await _api.post(
      '/conversations/$conversationId/messages',
      body: request.toJson(),
    );
    return SendConversationMessageResponse.fromJson(
      requireJsonMap(body, context: 'POST /conversations/{id}/messages'),
    );
  }

  @override
  Future<Conversation> archiveConversation(String id) async {
    final body = await _api.post('/conversations/$id/archive');
    final map = requireJsonMap(
      body,
      context: 'POST /conversations/{id}/archive',
    );
    return Conversation.fromJson(map['conversation'] as Map<String, dynamic>);
  }

  Map<String, String>? _query({String? status, int? limit, String? cursor}) {
    final query = <String, String>{};
    if (status != null && status.isNotEmpty) query['status'] = status;
    if (limit != null) query['limit'] = '$limit';
    if (cursor != null && cursor.isNotEmpty) query['cursor'] = cursor;
    return query.isEmpty ? null : query;
  }
}
