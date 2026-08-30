class CreateConversationRequest {
  const CreateConversationRequest({this.title});

  final String? title;

  Map<String, dynamic> toJson() => {
    if (title != null && title!.trim().isNotEmpty) 'title': title!.trim(),
  };
}

class SendConversationMessageRequest {
  const SendConversationMessageRequest({required this.content});

  final String content;

  Map<String, dynamic> toJson() => {'content': content};
}
