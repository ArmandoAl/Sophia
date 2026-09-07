package infrastructure

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/armandoalvarado/sofia-backend/internal/conversations/domain"
)

type InMemoryConversationRepository struct {
	mu            sync.Mutex
	conversations map[string]*domain.Conversation
}

type InMemoryConversationMessageRepository struct {
	mu       sync.Mutex
	messages map[string]*domain.ConversationMessage
}

func NewInMemoryConversationRepository() *InMemoryConversationRepository {
	return &InMemoryConversationRepository{conversations: make(map[string]*domain.Conversation)}
}

func NewInMemoryConversationMessageRepository() *InMemoryConversationMessageRepository {
	return &InMemoryConversationMessageRepository{messages: make(map[string]*domain.ConversationMessage)}
}

func (r *InMemoryConversationRepository) Create(ctx context.Context, conversation *domain.Conversation) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.conversations[conversation.ID] = cloneConversation(conversation)
	return nil
}

func (r *InMemoryConversationRepository) FindByID(ctx context.Context, userID, conversationID string) (*domain.Conversation, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	conversation, ok := r.conversations[conversationID]
	if !ok || conversation.UserID != userID {
		return nil, nil
	}
	return cloneConversation(conversation), nil
}

func (r *InMemoryConversationRepository) List(ctx context.Context, filter domain.ListFilter) ([]*domain.Conversation, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	result := make([]*domain.Conversation, 0)
	for _, conversation := range r.conversations {
		if filter.UserID != "" && conversation.UserID != filter.UserID {
			continue
		}
		if filter.Status != "" && conversation.Status != filter.Status {
			continue
		}
		result = append(result, cloneConversation(conversation))
	}
	sortConversations(result)
	result = conversationsAfterCursor(result, filter.Cursor)
	if filter.Limit > 0 && len(result) > filter.Limit {
		result = result[:filter.Limit]
	}
	return result, nil
}

func (r *InMemoryConversationRepository) Update(ctx context.Context, conversation *domain.Conversation) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.conversations[conversation.ID] = cloneConversation(conversation)
	return nil
}

func (r *InMemoryConversationRepository) Archive(ctx context.Context, userID, conversationID string) (*domain.Conversation, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	conversation, ok := r.conversations[conversationID]
	if !ok || conversation.UserID != userID {
		return nil, domain.ErrConversationNotFound
	}
	cp := cloneConversation(conversation)
	cp.Archive()
	r.conversations[conversationID] = cloneConversation(cp)
	return cp, nil
}

func (r *InMemoryConversationMessageRepository) Create(ctx context.Context, message *domain.ConversationMessage) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.messages[message.ID] = cloneMessage(message)
	return nil
}

func (r *InMemoryConversationMessageRepository) List(ctx context.Context, userID, conversationID string, limit int, cursor string) ([]*domain.ConversationMessage, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	result := make([]*domain.ConversationMessage, 0)
	for _, message := range r.messages {
		if message.UserID != userID || message.ConversationID != conversationID {
			continue
		}
		result = append(result, cloneMessage(message))
	}
	sortMessages(result)
	result = messagesAfterCursor(result, cursor)
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

func sortConversations(conversations []*domain.Conversation) {
	sort.Slice(conversations, func(i, j int) bool {
		left := conversations[i].UpdatedAt
		right := conversations[j].UpdatedAt
		if left.Equal(right) {
			return conversations[i].ID > conversations[j].ID
		}
		return left.After(right)
	})
}

func sortMessages(messages []*domain.ConversationMessage) {
	sort.Slice(messages, func(i, j int) bool {
		if messages[i].CreatedAt.Equal(messages[j].CreatedAt) {
			return messages[i].ID < messages[j].ID
		}
		return messages[i].CreatedAt.Before(messages[j].CreatedAt)
	})
}

func conversationsAfterCursor(conversations []*domain.Conversation, cursor string) []*domain.Conversation {
	if cursor == "" {
		return conversations
	}
	for i, conversation := range conversations {
		if conversation.ID == cursor {
			return conversations[i+1:]
		}
	}
	return conversations
}

func messagesAfterCursor(messages []*domain.ConversationMessage, cursor string) []*domain.ConversationMessage {
	if cursor == "" {
		return messages
	}
	for i, message := range messages {
		if message.ID == cursor {
			return messages[i+1:]
		}
	}
	return messages
}

func cloneConversation(conversation *domain.Conversation) *domain.Conversation {
	if conversation == nil {
		return nil
	}
	cp := *conversation
	cp.LastMessageAt = cloneTime(conversation.LastMessageAt)
	return &cp
}

func cloneMessage(message *domain.ConversationMessage) *domain.ConversationMessage {
	if message == nil {
		return nil
	}
	cp := *message
	return &cp
}

func cloneTime(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	cp := *value
	return &cp
}
