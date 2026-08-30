import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:sophia_ai/core/models/models.dart';
import 'package:sophia_ai/core/network/api_exception.dart';
import 'package:sophia_ai/features/chat/domain/conversations_repository.dart';
import 'package:sophia_ai/features/chat/domain/entities/chat_message.dart';
import 'package:sophia_ai/features/chat/presentation/cubit/chat_message_state.dart';

class ChatMessageCubit extends Cubit<ChatMessageState> {
  ChatMessageCubit({required ConversationsRepository repository})
    : _repository = repository,
      super(const ChatMessageState());

  final ConversationsRepository _repository;

  Future<void> load() async {
    emit(state.copyWith(isLoading: true, clearError: true));
    try {
      final list = await _repository.listConversations(
        status: 'active',
        limit: 1,
      );
      final conversation = list.conversations.isEmpty
          ? await _repository.createConversation(
              const CreateConversationRequest(title: 'Sofia chat'),
            )
          : list.conversations.first;
      final history = await _repository.listMessages(conversation.id);
      emit(
        state.copyWith(
          conversation: conversation,
          messages: history.messages
              .map(_messageFromBackend)
              .toList()
              .reversed
              .toList(),
          isLoading: false,
          clearError: true,
        ),
      );
    } catch (e) {
      emit(state.copyWith(isLoading: false, errorMessage: _messageForError(e)));
    }
  }

  Future<void> sendMessage(String text) async {
    final content = text.trim();
    if (content.isEmpty || state.isTyping) return;

    var conversation = state.conversation;
    if (conversation == null) {
      await load();
      conversation = state.conversation;
      if (conversation == null) return;
    }

    emit(state.copyWith(isTyping: true, clearError: true));
    try {
      final response = await _repository.sendMessage(
        conversation.id,
        SendConversationMessageRequest(content: content),
      );
      final newMessages = [
        _messageFromBackend(
          response.assistantMessage,
          response.proposedActions,
        ),
        _messageFromBackend(response.userMessage),
        ...state.messages,
      ];
      emit(
        state.copyWith(
          conversation: response.conversation,
          messages: newMessages,
          isTyping: false,
          clearError: true,
        ),
      );
    } catch (e) {
      emit(state.copyWith(isTyping: false, errorMessage: _messageForError(e)));
    }
  }

  Future<void> retry() => load();

  void toggleListening() {
    emit(state.copyWith(isListening: !state.isListening));
  }

  ChatMessage _messageFromBackend(
    ConversationMessage message, [
    List<RuntimeProposedAction> proposedActions = const [],
  ]) {
    final hasActions = !message.isUser && proposedActions.isNotEmpty;
    return ChatMessage(
      id: message.id,
      text: message.content,
      isUser: message.isUser,
      type: hasActions ? MessageType.actionProposal : MessageType.text,
      timestamp: message.createdAt,
      metadata: hasActions
          ? {
              'title': message.content.isEmpty ? 'Proposal' : message.content,
              'actions': proposedActions
                  .map(
                    (action) => {
                      'id': action.id ?? action.toolName,
                      'icon': 'check_circle',
                      'label': action.toolName,
                      'detail': action.reason,
                      'color': action.riskLevel == 'high' ? 'orange' : 'cyan',
                    },
                  )
                  .toList(),
            }
          : null,
    );
  }

  String _messageForError(Object error) {
    if (error is ApiException) return error.message;
    return 'Unable to load Sophia chat';
  }
}
