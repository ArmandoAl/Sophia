# Conversation ingestion format

`POST /ingestion/conversations` accepts normalized JSON. Native WhatsApp, Instagram, and Facebook exports must be converted before calling this endpoint.

```json
{
  "conversations": [
    {
      "source": "whatsapp",
      "external_id": "chat-123",
      "participants": ["Armando", "María"],
      "messages": [
        {
          "author": "Armando",
          "text": "Prefiero reuniones por la mañana.",
          "sent_at": "2026-09-09T16:00:00Z"
        }
      ]
    }
  ]
}
```

`source` must be `whatsapp`, `instagram`, `facebook`, or `manual`. `external_id` must be stable within a source; resending it returns the original batch instead of enqueuing duplicate work. Participants and messages must be non-empty, and every message requires a non-empty author, text, and RFC 3339 timestamp.

The endpoint rejects request bodies larger than 900 KiB so a batch remains within Firestore's document-size envelope.
