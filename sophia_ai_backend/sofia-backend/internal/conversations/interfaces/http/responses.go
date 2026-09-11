package http

import (
	runtimedomain "github.com/armandoalvarado/sofia-backend/internal/ai/runtime/domain"
	"github.com/armandoalvarado/sofia-backend/internal/conversations/application"
	"github.com/armandoalvarado/sofia-backend/internal/conversations/domain"
)

type conversationResponse struct {
	Conversation *domain.Conversation `json:"conversation"`
}

type listConversationsResponse struct {
	Conversations []*domain.Conversation `json:"conversations"`
	NextCursor    string                 `json:"next_cursor,omitempty"`
}

type listMessagesResponse struct {
	Messages   []*domain.ConversationMessage `json:"messages"`
	NextCursor string                        `json:"next_cursor,omitempty"`
}

type sendMessageResponse struct {
	Conversation     *domain.Conversation                 `json:"conversation"`
	UserMessage      *domain.ConversationMessage          `json:"user_message"`
	AssistantMessage *domain.ConversationMessage          `json:"assistant_message"`
	ProposedActions  []runtimedomain.ActionProposalOutput `json:"proposed_actions"`
	RuntimeRequestID string                               `json:"runtime_request_id"`
	ActiveEntity     *runtimedomain.EntityReference       `json:"active_entity,omitempty"`
	ContextChanged   bool                                 `json:"context_changed"`
}

func listConversationsFromDomain(conversations []*domain.Conversation, limit int) listConversationsResponse {
	resp := listConversationsResponse{Conversations: conversations}
	if limit > 0 && len(conversations) == limit {
		resp.NextCursor = conversations[len(conversations)-1].ID
	}
	return resp
}

func listMessagesFromDomain(messages []*domain.ConversationMessage, limit int) listMessagesResponse {
	resp := listMessagesResponse{Messages: messages}
	if limit > 0 && len(messages) == limit {
		resp.NextCursor = messages[len(messages)-1].ID
	}
	return resp
}

func sendMessageFromResult(result *application.SendMessageResult) sendMessageResponse {
	return sendMessageResponse{
		Conversation:     result.Conversation,
		UserMessage:      result.UserMessage,
		AssistantMessage: result.AssistantMessage,
		ProposedActions:  result.ProposedActions,
		RuntimeRequestID: result.RuntimeRequestID,
		ActiveEntity:     result.ActiveEntity,
		ContextChanged:   result.ContextChanged,
	}
}
