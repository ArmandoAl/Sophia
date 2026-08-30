enum MessageType {
  text,
  scheduleConflict,
  mapLocation,
  audioWave,
  actionProposal,
}

class ChatMessage {
  final String id;
  final String text;
  final bool isUser;
  final MessageType type;
  final DateTime timestamp;
  final Map<String, dynamic>?
  metadata; // Para datos extra (ej: coordenadas del mapa, acciones propuestas)

  const ChatMessage({
    required this.id,
    required this.text,
    required this.isUser,
    this.type = MessageType.text,
    required this.timestamp,
    this.metadata,
  });
}
