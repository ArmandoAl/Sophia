class CreateConversationRequest {
  const CreateConversationRequest({this.title});

  final String? title;

  Map<String, dynamic> toJson() => {
    if (title != null && title!.trim().isNotEmpty) 'title': title!.trim(),
  };
}

class SendConversationMessageRequest {
  const SendConversationMessageRequest({
    required this.content,
    this.activeContext,
  });

  final String content;
  final String? activeContext;

  Map<String, dynamic> toJson() => {
    'content': content,
    if (activeContext != null && activeContext!.isNotEmpty)
      'active_context': activeContext,
  };
}
