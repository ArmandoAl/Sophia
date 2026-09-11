import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:sophia_ai/core/config/app_config.dart';
import 'package:sophia_ai/core/di/service_locator.dart';
import 'package:sophia_ai/core/models/models.dart';
import 'package:sophia_ai/core/network/api_client.dart';
import 'package:sophia_ai/core/network/api_exception.dart';
import 'package:sophia_ai/core/storage/token_storage.dart';
import 'package:sophia_ai/features/chat/data/conversations_repository_impl.dart';
import 'package:sophia_ai/features/chat/domain/conversations_repository.dart';
import 'package:sophia_ai/features/chat/presentation/cubit/chat_message_cubit.dart';
import 'package:sophia_ai/features/chat/presentation/pages/chat_page.dart';

final _createdAt = DateTime.utc(2026, 7, 15, 12);
final _updatedAt = DateTime.utc(2026, 7, 15, 12, 1);

Map<String, dynamic> conversationJson({String id = 'conv_1'}) => {
  'id': id,
  'user_id': 'usr_1',
  'title': 'Sofia chat',
  'status': 'active',
  'created_at': _createdAt.toIso8601String(),
  'updated_at': _updatedAt.toIso8601String(),
  'last_message_at': _updatedAt.toIso8601String(),
};

Map<String, dynamic> messageJson({
  String id = 'msg_1',
  String role = 'user',
  String content = 'Hola',
}) => {
  'id': id,
  'conversation_id': 'conv_1',
  'user_id': 'usr_1',
  'role': role,
  'content': content,
  'created_at': _createdAt.toIso8601String(),
  if (role == 'assistant') 'runtime_request_id': 'req_1',
};

Map<String, dynamic> sendResponseJson() => {
  'conversation': conversationJson(),
  'user_message': messageJson(),
  'assistant_message': messageJson(
    id: 'msg_2',
    role: 'assistant',
    content: 'Hola, soy Sophia.',
  ),
  'proposed_actions': <Map<String, dynamic>>[],
  'runtime_request_id': 'req_1',
  'active_entity': {
    'id': 'entity_1',
    'scope_key': 'person:diana',
    'label': 'Diana',
    'relationship': 'hermana',
  },
  'context_changed': true,
};

void main() {
  group('conversation models', () {
    test('parse Conversation', () {
      final conversation = Conversation.fromJson(conversationJson());
      expect(conversation.id, 'conv_1');
      expect(conversation.userId, 'usr_1');
      expect(conversation.lastMessageAt, _updatedAt);
    });

    test('parse ConversationMessage', () {
      final message = ConversationMessage.fromJson(
        messageJson(role: 'assistant', content: 'Lista.'),
      );
      expect(message.isAssistant, isTrue);
      expect(message.runtimeRequestId, 'req_1');
      expect(message.createdAt, _createdAt);
    });

    test('parse send response', () {
      final response = SendConversationMessageResponse.fromJson(
        sendResponseJson(),
      );
      expect(response.userMessage.content, 'Hola');
      expect(response.assistantMessage.content, 'Hola, soy Sophia.');
      expect(response.runtimeRequestId, 'req_1');
      expect(response.activeEntity?.label, 'Diana');
      expect(response.contextChanged, isTrue);
      expect(
        const SendConversationMessageRequest(
          content: '¿Cómo está?',
          activeContext: 'person:diana',
        ).toJson()['active_context'],
        'person:diana',
      );
    });
  });

  group('ConversationsRepositoryImpl', () {
    late MemoryTokenStorage tokens;

    setUp(() async {
      tokens = MemoryTokenStorage();
      await tokens.saveToken('jwt');
    });

    ApiClient clientFor(MockClient mock) => ApiClient(
      config: const AppConfig(baseUrl: 'http://localhost:8080'),
      tokenStorage: tokens,
      httpClient: mock,
    );

    test('create conversation', () async {
      http.BaseRequest? captured;
      final repo = ConversationsRepositoryImpl(
        clientFor(
          MockClient((request) async {
            captured = request;
            return http.Response(
              jsonEncode({'conversation': conversationJson()}),
              201,
            );
          }),
        ),
      );

      final conversation = await repo.createConversation(
        const CreateConversationRequest(title: 'Sofia chat'),
      );

      expect(captured!.url.path, '/conversations');
      expect(conversation.id, 'conv_1');
    });

    test('list messages', () async {
      final repo = ConversationsRepositoryImpl(
        clientFor(
          MockClient(
            (_) async => http.Response(
              jsonEncode({
                'messages': [
                  messageJson(),
                  messageJson(id: 'msg_2', role: 'assistant', content: 'Hola.'),
                ],
              }),
              200,
            ),
          ),
        ),
      );

      final response = await repo.listMessages('conv_1');
      expect(response.messages, hasLength(2));
      expect(response.messages.last.isAssistant, isTrue);
    });

    test('send message', () async {
      final repo = ConversationsRepositoryImpl(
        clientFor(
          MockClient(
            (_) async => http.Response(jsonEncode(sendResponseJson()), 201),
          ),
        ),
      );

      final response = await repo.sendMessage(
        'conv_1',
        const SendConversationMessageRequest(content: 'Hola'),
      );

      expect(response.assistantMessage.content, 'Hola, soy Sophia.');
    });
  });

  group('ChatMessageCubit', () {
    late FakeConversationsRepository repo;
    late ChatMessageCubit cubit;

    setUp(() {
      repo = FakeConversationsRepository();
      cubit = ChatMessageCubit(repository: repo);
    });

    tearDown(() async {
      await cubit.close();
    });

    test('creates conversation if none exists', () async {
      await cubit.load();
      expect(repo.createCalls, 1);
      expect(cubit.state.conversation?.id, 'conv_1');
    });

    test('loads history', () async {
      repo.listResult = ConversationListResponse(
        conversations: [sampleConversation()],
      );
      repo.messagesResult = ConversationMessagesResponse(
        messages: [
          sampleMessage(content: 'Hola'),
          sampleMessage(id: 'msg_2', role: 'assistant', content: 'Que tal'),
        ],
      );

      await cubit.load();

      expect(cubit.state.messages, hasLength(2));
      expect(cubit.state.messages.last.text, 'Hola');
    });

    test('sends message and appends backend response', () async {
      await cubit.load();
      await cubit.sendMessage('Hola');

      expect(repo.sendCalls, 1);
      expect(cubit.state.messages.first.text, 'Hola, soy Sophia.');
      expect(cubit.state.isTyping, isFalse);
    });

    test('handles provider error without mock response', () async {
      await cubit.load();
      repo.sendError = const ApiException(
        statusCode: 500,
        message: 'provider failed',
      );

      await cubit.sendMessage('Hola');

      expect(cubit.state.errorMessage, 'provider failed');
      expect(
        cubit.state.messages.any((m) => m.text.contains("I've processed")),
        isFalse,
      );
    });
  });

  group('ChatPage', () {
    tearDown(() async {
      await sl.reset();
    });

    testWidgets('shows history', (tester) async {
      final repo = FakeConversationsRepository()
        ..listResult = ConversationListResponse(
          conversations: [sampleConversation()],
        )
        ..messagesResult = ConversationMessagesResponse(
          messages: [sampleMessage(content: 'Persisted hello')],
        );
      sl.registerFactory(() => ChatMessageCubit(repository: repo));

      await tester.pumpWidget(const MaterialApp(home: ChatPage()));
      await tester.pump();
      await tester.pump();

      expect(find.text('Persisted hello'), findsOneWidget);
    });

    testWidgets('shows error', (tester) async {
      final repo = FakeConversationsRepository()
        ..listError = const ApiException(statusCode: 500, message: 'boom');
      sl.registerFactory(() => ChatMessageCubit(repository: repo));

      await tester.pumpWidget(const MaterialApp(home: ChatPage()));
      await tester.pump();
      await tester.pump();

      expect(find.text('boom'), findsOneWidget);
    });
  });
}

Conversation sampleConversation() => Conversation.fromJson(conversationJson());

ConversationMessage sampleMessage({
  String id = 'msg_1',
  String role = 'user',
  String content = 'Hola',
}) => ConversationMessage.fromJson(
  messageJson(id: id, role: role, content: content),
);

class FakeConversationsRepository implements ConversationsRepository {
  ConversationListResponse? listResult;
  ConversationMessagesResponse? messagesResult;
  Object? listError;
  Object? sendError;
  int createCalls = 0;
  int sendCalls = 0;

  @override
  Future<Conversation> createConversation(
    CreateConversationRequest request,
  ) async {
    createCalls++;
    return sampleConversation();
  }

  @override
  Future<ConversationListResponse> listConversations({
    String? status,
    int? limit,
    String? cursor,
  }) async {
    if (listError != null) throw listError!;
    return listResult ?? const ConversationListResponse(conversations: []);
  }

  @override
  Future<Conversation> getConversation(String id) async => sampleConversation();

  @override
  Future<ConversationMessagesResponse> listMessages(
    String conversationId, {
    int? limit,
    String? cursor,
  }) async {
    return messagesResult ?? const ConversationMessagesResponse(messages: []);
  }

  @override
  Future<SendConversationMessageResponse> sendMessage(
    String conversationId,
    SendConversationMessageRequest request,
  ) async {
    sendCalls++;
    if (sendError != null) throw sendError!;
    return SendConversationMessageResponse.fromJson(sendResponseJson());
  }

  @override
  Future<Conversation> archiveConversation(String id) async =>
      sampleConversation();
}
