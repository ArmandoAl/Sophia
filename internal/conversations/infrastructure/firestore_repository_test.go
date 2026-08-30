package infrastructure

import (
	"testing"

	"github.com/armandoalvarado/sofia-backend/internal/database"
)

func TestFirestoreRepositoriesUseV1Namespaces(t *testing.T) {
	if conversationsCollection != database.V1ConversationsCollection || conversationsCollection == "conversations" {
		t.Fatalf("unexpected conversation collection: %q", conversationsCollection)
	}
	if messagesCollection != database.V1ConversationMessagesCollection || messagesCollection == "conversation_messages" {
		t.Fatalf("unexpected message collection: %q", messagesCollection)
	}
}
