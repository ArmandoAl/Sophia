import 'package:equatable/equatable.dart';
import 'package:sophia_ai/core/models/models.dart';
import 'package:sophia_ai/features/chat/domain/entities/chat_message.dart';

class ChatMessageState extends Equatable {
  final Conversation? conversation;
  final List<ChatMessage> messages;
  final bool isLoading;
  final bool isListening; // Para la animación de ondas de voz (screen3)
  final bool isTyping; // Para mostrar "Sophia is typing..."
  final String? errorMessage;

  const ChatMessageState({
    this.conversation,
    this.messages = const [],
    this.isLoading = false,
    this.isListening = false,
    this.isTyping = false,
    this.errorMessage,
  });

  ChatMessageState copyWith({
    Conversation? conversation,
    List<ChatMessage>? messages,
    bool? isLoading,
    bool? isListening,
    bool? isTyping,
    String? errorMessage,
    bool clearError = false,
  }) {
    return ChatMessageState(
      conversation: conversation ?? this.conversation,
      messages: messages ?? this.messages,
      isLoading: isLoading ?? this.isLoading,
      isListening: isListening ?? this.isListening,
      isTyping: isTyping ?? this.isTyping,
      errorMessage: clearError ? null : errorMessage ?? this.errorMessage,
    );
  }

  @override
  List<Object?> get props => [
    conversation,
    messages,
    isLoading,
    isListening,
    isTyping,
    errorMessage,
  ];
}
