import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:sophia_ai/features/chat/domain/entities/chat_message.dart';
import 'package:sophia_ai/features/chat/presentation/cubit/chat_message_state.dart';

// --- Cubit ---
class ChatMessageCubit extends Cubit<ChatMessageState> {
  ChatMessageCubit() : super(const ChatMessageState(messages: [])) {
    _loadInitialMessages();
  }

  void _loadInitialMessages() {
    // Simulamos historial previo basado en screen6
    final initialMsgs = [
      ChatMessage(
        id: '1',
        text: "Add a meeting with Dev Team at 3 PM.",
        isUser: true,
        timestamp: DateTime.now().subtract(const Duration(minutes: 5)),
      ),
      ChatMessage(
        id: '2',
        text: "Scheduling Conflict Detected",
        isUser: false,
        type: MessageType.scheduleConflict, // Renderizará la tarjeta amarilla
        timestamp: DateTime.now().subtract(const Duration(minutes: 4)),
      ),
      ChatMessage(
        id: '3',
        text: "Your car is parked here. Row C, Level 2.",
        isUser: false,
        type: MessageType.mapLocation,
        timestamp: DateTime.now().subtract(const Duration(minutes: 3)),
      ),

      // id: action['id'] as String,
      //     icon: _getIconData(action['icon'] as String),
      //     label: action['label'] as String,
      //     detail: action['detail'] as String,
      //     color: _getColorFromString(action['color'] as String?),
      ChatMessage(
        id: '4',
        text: "Proposal: Relax Mode",
        isUser: false,
        type: MessageType.actionProposal,
        timestamp: DateTime.now().subtract(const Duration(minutes: 2)),
        metadata: {
          'title': 'Proposal: Relax Mode',
          'actions': [
            {
              'id': '1',
              'icon': 'lightbulb',
              'label': 'Lights',
              'detail': 'Dimmed to 40%',
              'color': 'cyan',
            },
            {
              'id': '2',
              'icon': 'music_note',
              'label': 'Music',
              'detail': 'Jazz Playlist',
              'color': 'purple',
            },
            {
              'id': '3',
              'icon': 'ac_unit',
              'label': 'AC',
              'detail': 'Set to 22°C',
              'color': 'blue',
            },
          ],
        },
      ),
    ];
    emit(state.copyWith(messages: initialMsgs.reversed.toList()));
  }

  void sendMessage(String text) async {
    if (text.isEmpty) return;

    final userMsg = ChatMessage(
      id: DateTime.now().toString(),
      text: text,
      isUser: true,
      timestamp: DateTime.now(),
    );

    emit(
      state.copyWith(messages: [userMsg, ...state.messages], isTyping: true),
    );

    // Simular respuesta de IA
    await Future.delayed(const Duration(seconds: 2));

    // Lógica simple para simular respuestas ricas (Maps, Propuestas o Texto)
    ChatMessage aiMsg;

    final lowerText = text.toLowerCase();

    if (lowerText.contains("parked") || lowerText.contains("car")) {
      // Respuesta con mapa de ubicación
      aiMsg = ChatMessage(
        id: DateTime.now().toString(),
        text: "Your car is parked here. Row C, Level 2.",
        isUser: false,
        type: MessageType.mapLocation,
        timestamp: DateTime.now(),
      );
    } else if (lowerText.contains("casa") ||
        lowerText.contains("home") ||
        lowerText.contains("relax") ||
        lowerText.contains("relajar")) {
      // Respuesta con propuesta de acciones (modo relax)
      aiMsg = ChatMessage(
        id: DateTime.now().toString(),
        text: "Proposal: Relax Mode",
        isUser: false,
        type: MessageType.actionProposal,
        timestamp: DateTime.now(),
        metadata: {
          'title': 'Proposal: Relax Mode',
          'actions': [
            {
              'id': '1',
              'icon': 'lightbulb',
              'label': 'Lights',
              'detail': 'Dimmed to 40%',
              'color': 'cyan',
            },
            {
              'id': '2',
              'icon': 'music_note',
              'label': 'Music',
              'detail': 'Jazz Playlist',
              'color': 'purple',
            },
            {
              'id': '3',
              'icon': 'ac_unit',
              'label': 'AC',
              'detail': 'Set to 22°C',
              'color': 'blue',
            },
            {
              'id': '4',
              'icon': 'coffee',
              'label': 'Coffee Machine',
              'detail': 'Brewing espresso',
              'color': 'orange',
            },
          ],
        },
      );
    } else if (lowerText.contains("guest") ||
        lowerText.contains("invitados") ||
        lowerText.contains("visitors")) {
      // Respuesta con propuesta de acciones (modo invitados)
      aiMsg = ChatMessage(
        id: DateTime.now().toString(),
        text: "Proposal: Guest Mode",
        isUser: false,
        type: MessageType.actionProposal,
        timestamp: DateTime.now(),
        metadata: {
          'title': 'Proposal: Guest Mode',
          'actions': [
            {
              'id': '1',
              'icon': 'lightbulb',
              'label': 'Lights',
              'detail': 'Dimmed',
              'color': 'cyan',
            },
            {
              'id': '2',
              'icon': 'music_note',
              'label': 'Music',
              'detail': 'Jazz',
              'color': 'purple',
            },
            {
              'id': '3',
              'icon': 'ac_unit',
              'label': 'AC',
              'detail': '22°C',
              'color': 'blue',
            },
          ],
        },
      );
    } else {
      // Respuesta de texto simple
      aiMsg = ChatMessage(
        id: DateTime.now().toString(),
        text: "I've processed your request: \"$text\". Anything else?",
        isUser: false,
        timestamp: DateTime.now(),
      );
    }

    emit(state.copyWith(messages: [aiMsg, ...state.messages], isTyping: false));
  }

  void toggleListening() {
    emit(state.copyWith(isListening: !state.isListening));
  }
}
