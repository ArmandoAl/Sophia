import 'package:bloc/bloc.dart';
import 'package:sophia_core/core/models/models.dart';
import 'package:sophia_core/core/network/api_exception.dart';
import 'package:sophia_core/core/config/demo_mode.dart';
import 'package:sophia_core/features/chat/domain/conversations_repository.dart';
import 'package:sophia_core/features/chat/domain/entities/chat_message.dart';
import 'package:sophia_core/features/chat/presentation/cubit/chat_message_state.dart';
import 'package:sophia_core/demo/demo_store.dart';

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
          suggestedReplies: list.suggestedReplies,
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

    emit(
      state.copyWith(
        isTyping: true,
        clearError: true,
        suggestedReplies: const [],
      ),
    );
    try {
      final response = await _repository.sendMessage(
        conversation.id,
        SendConversationMessageRequest(
          content: content,
          activeContext: state.manualContext,
        ),
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
          activeEntity: response.activeEntity,
          contextChanged: response.contextChanged,
          suggestedReplies: response.suggestedReplies,
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

  void setManualContext(EntityReference entity) {
    emit(
      state.copyWith(
        activeEntity: entity,
        manualContext: entity.scopeKey,
        contextChanged: true,
      ),
    );
  }

  void useAutomaticContext() {
    emit(
      state.copyWith(
        clearActiveEntity: true,
        clearManualContext: true,
        contextChanged: false,
      ),
    );
  }

  void demoProposalRejected() {
    if (!demoMode) return;
    _appendDemoMessage(
      'Entiendo. ¿Qué te gustaría cambiar para que te funcione mejor?',
      DemoStore.instance.retryReplies,
    );
  }

  void demoProposalApproved(String title) {
    if (!demoMode) return;
    _appendDemoMessage(
      'Listo. “$title” ya aparece en Actividades de esta demo.',
      const [],
    );
  }

  void _appendDemoMessage(String text, List<String> replies) {
    final message = ChatMessage(
      id: 'demo-local-${DateTime.now().microsecondsSinceEpoch}',
      text: text,
      isUser: false,
      timestamp: DateTime.now(),
    );
    emit(
      state.copyWith(
        messages: [message, ...state.messages],
        suggestedReplies: replies,
      ),
    );
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
                      'proposed_input': action.proposedInput,
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

  @override
  void emit(ChatMessageState state) {
    if (isClosed) return;
    super.emit(state);
  }
}
