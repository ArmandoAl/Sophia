package application

import (
	"context"
	"errors"
	"strings"

	runtimedomain "github.com/armandoalvarado/sofia-backend/internal/ai/runtime/domain"
	"github.com/armandoalvarado/sofia-backend/internal/conversations/domain"
	"github.com/google/uuid"
)

type Service struct {
	conversations domain.ConversationRepository
	messages      domain.ConversationMessageRepository
	runtime       runtimedomain.RuntimeService
}

type SendMessageResult struct {
	Conversation     *domain.Conversation
	UserMessage      *domain.ConversationMessage
	AssistantMessage *domain.ConversationMessage
	ProposedActions  []runtimedomain.ActionProposalOutput
	RuntimeRequestID string
}

func NewService(conversations domain.ConversationRepository, messages domain.ConversationMessageRepository, runtime runtimedomain.RuntimeService) *Service {
	return &Service{conversations: conversations, messages: messages, runtime: runtime}
}

func (s *Service) CreateConversation(ctx context.Context, userID string, input domain.CreateConversationInput) (*domain.Conversation, error) {
	conversation, err := domain.NewConversation(uuid.New().String(), userID, input)
	if err != nil {
		return nil, err
	}
	if err := s.conversations.Create(ctx, conversation); err != nil {
		return nil, err
	}
	return conversation, nil
}

func (s *Service) ListConversations(ctx context.Context, filter domain.ListFilter) ([]*domain.Conversation, error) {
	if filter.Limit == 0 {
		filter.Limit = domain.DefaultListLimit
	}
	if err := domain.ValidateListFilter(filter); err != nil {
		return nil, err
	}
	return s.conversations.List(ctx, filter)
}

func (s *Service) GetConversation(ctx context.Context, userID, conversationID string) (*domain.Conversation, error) {
	conversation, err := s.conversations.FindByID(ctx, userID, conversationID)
	if err != nil {
		return nil, err
	}
	if conversation == nil {
		return nil, domain.ErrConversationNotFound
	}
	return conversation, nil
}

func (s *Service) ListMessages(ctx context.Context, userID, conversationID string, limit int, cursor string) ([]*domain.ConversationMessage, error) {
	if _, err := s.GetConversation(ctx, userID, conversationID); err != nil {
		return nil, err
	}
	if limit == 0 {
		limit = domain.DefaultListLimit
	}
	if err := domain.ValidateListFilter(domain.ListFilter{Limit: limit}); err != nil {
		return nil, err
	}
	return s.messages.List(ctx, userID, conversationID, limit, cursor)
}

func (s *Service) SendMessage(ctx context.Context, userID, conversationID, content string) (*SendMessageResult, error) {
	conversation, err := s.GetConversation(ctx, userID, conversationID)
	if err != nil {
		return nil, err
	}

	userMessage, err := domain.NewMessage(uuid.New().String(), conversation.ID, userID, domain.RoleUser, content, "")
	if err != nil {
		return nil, err
	}
	if err := s.messages.Create(ctx, userMessage); err != nil {
		return nil, err
	}
	conversation.Touch(userMessage.CreatedAt)
	if err := s.conversations.Update(ctx, conversation); err != nil {
		return nil, err
	}

	history, err := s.runtimeHistory(ctx, userID, conversation.ID, userMessage.ID)
	if err != nil {
		return nil, err
	}

	runtimeResponse, err := s.runtime.HandleMessage(ctx, runtimedomain.RuntimeRequest{
		UserID:         userID,
		Message:        userMessage.Content,
		DryRun:         false,
		ConversationID: conversation.ID,
		History:        history,
	})
	if err != nil {
		return nil, err
	}
	if runtimeResponse == nil || runtimeResponse.AssistantMessage == "" {
		return nil, errors.New("runtime returned empty assistant message")
	}

	assistantMessage, err := domain.NewMessage(uuid.New().String(), conversation.ID, userID, domain.RoleAssistant, runtimeResponse.AssistantMessage, runtimeResponse.RequestID)
	if err != nil {
		return nil, err
	}
	if err := s.messages.Create(ctx, assistantMessage); err != nil {
		return nil, err
	}
	conversation.Touch(assistantMessage.CreatedAt)
	if err := s.conversations.Update(ctx, conversation); err != nil {
		return nil, err
	}

	return &SendMessageResult{
		Conversation:     conversation,
		UserMessage:      userMessage,
		AssistantMessage: assistantMessage,
		ProposedActions:  runtimeResponse.ProposedActions,
		RuntimeRequestID: runtimeResponse.RequestID,
	}, nil
}

func (s *Service) ArchiveConversation(ctx context.Context, userID, conversationID string) (*domain.Conversation, error) {
	return s.conversations.Archive(ctx, userID, conversationID)
}

const (
	runtimeHistoryLimit = 10
	runtimeHistoryChars = 500
)

func (s *Service) runtimeHistory(ctx context.Context, userID, conversationID, currentMessageID string) ([]runtimedomain.Turn, error) {
	stored, err := s.messages.List(ctx, userID, conversationID, 0, "")
	if err != nil {
		return nil, err
	}
	history := make([]runtimedomain.Turn, 0, runtimeHistoryLimit)
	for _, message := range stored {
		if message == nil || message.ID == currentMessageID {
			continue
		}
		history = append(history, runtimedomain.Turn{
			Role:    message.Role,
			Content: truncateChars(message.Content, runtimeHistoryChars),
		})
	}
	if len(history) > runtimeHistoryLimit {
		history = history[len(history)-runtimeHistoryLimit:]
	}
	return history, nil
}

func truncateChars(value string, limit int) string {
	runes := []rune(strings.TrimSpace(value))
	if limit <= 0 || len(runes) <= limit {
		return string(runes)
	}
	return string(runes[:limit])
}
