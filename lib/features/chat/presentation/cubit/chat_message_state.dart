import 'package:equatable/equatable.dart';
import 'package:sophia_ai/features/chat/domain/entities/chat_message.dart';

class ChatMessageState extends Equatable {
  final List<ChatMessage> messages;
  final bool isListening; // Para la animación de ondas de voz (screen3)
  final bool isTyping; // Para mostrar "Sophia is typing..."

  const ChatMessageState({
    this.messages = const [],
    this.isListening = false,
    this.isTyping = false,
  });

  ChatMessageState copyWith({
    List<ChatMessage>? messages,
    bool? isListening,
    bool? isTyping,
  }) {
    return ChatMessageState(
      messages: messages ?? this.messages,
      isListening: isListening ?? this.isListening,
      isTyping: isTyping ?? this.isTyping,
    );
  }

  @override
  List<Object> get props => [messages, isListening, isTyping];
}
