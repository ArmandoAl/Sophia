package domain

import "context"

type ConversationRepository interface {
	Create(ctx context.Context, conversation *Conversation) error
	FindByID(ctx context.Context, userID, conversationID string) (*Conversation, error)
	List(ctx context.Context, filter ListFilter) ([]*Conversation, error)
	Update(ctx context.Context, conversation *Conversation) error
	Archive(ctx context.Context, userID, conversationID string) (*Conversation, error)
}

type ConversationMessageRepository interface {
	Create(ctx context.Context, message *ConversationMessage) error
	List(ctx context.Context, userID, conversationID string, limit int, cursor string) ([]*ConversationMessage, error)
}

type ConversationLister interface {
	List(ctx context.Context, filter ListFilter) ([]*Conversation, error)
}

type ConversationMessageLister interface {
	List(ctx context.Context, userID, conversationID string, limit int, cursor string) ([]*ConversationMessage, error)
}
