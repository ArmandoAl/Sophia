package http

import "github.com/armandoalvarado/sofia-backend/internal/conversations/domain"

type createConversationRequest struct {
	Title string `json:"title"`
}

type sendMessageRequest struct {
	Content string `json:"content"`
}

func (r createConversationRequest) toDomain() domain.CreateConversationInput {
	return domain.CreateConversationInput{Title: r.Title}
}
