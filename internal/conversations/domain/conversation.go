package domain

import (
	"errors"
	"strings"
	"time"
)

const (
	StatusActive   = "active"
	StatusArchived = "archived"

	RoleUser      = "user"
	RoleAssistant = "assistant"
	RoleSystem    = "system"

	DefaultListLimit = 50
	MaxListLimit     = 100
)

var (
	ErrConversationNotFound = errors.New("conversation not found")
	ErrInvalidTitle         = errors.New("conversation title is invalid")
	ErrInvalidContent       = errors.New("message content is required")
	ErrInvalidRole          = errors.New("message role is invalid")
	ErrInvalidLimit         = errors.New("limit is invalid")
)

type Conversation struct {
	ID            string     `json:"id"`
	UserID        string     `json:"user_id"`
	Title         string     `json:"title"`
	Status        string     `json:"status"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	LastMessageAt *time.Time `json:"last_message_at,omitempty"`
}

type ConversationMessage struct {
	ID               string    `json:"id"`
	ConversationID   string    `json:"conversation_id"`
	UserID           string    `json:"user_id"`
	Role             string    `json:"role"`
	Content          string    `json:"content"`
	RuntimeRequestID string    `json:"runtime_request_id,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
}

type CreateConversationInput struct {
	Title string
}

type ListFilter struct {
	UserID string
	Status string
	Limit  int
	Cursor string
}

func NewConversation(id, userID string, input CreateConversationInput) (*Conversation, error) {
	title := strings.TrimSpace(input.Title)
	if title == "" {
		title = "New conversation"
	}
	if len([]rune(title)) > 120 {
		return nil, ErrInvalidTitle
	}
	now := time.Now().UTC()
	return &Conversation{
		ID:        id,
		UserID:    strings.TrimSpace(userID),
		Title:     title,
		Status:    StatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func NewMessage(id, conversationID, userID, role, content, runtimeRequestID string) (*ConversationMessage, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil, ErrInvalidContent
	}
	if !validRole(role) {
		return nil, ErrInvalidRole
	}
	return &ConversationMessage{
		ID:               id,
		ConversationID:   strings.TrimSpace(conversationID),
		UserID:           strings.TrimSpace(userID),
		Role:             role,
		Content:          content,
		RuntimeRequestID: strings.TrimSpace(runtimeRequestID),
		CreatedAt:        time.Now().UTC(),
	}, nil
}

func (c *Conversation) Touch(at time.Time) {
	when := at.UTC()
	c.LastMessageAt = &when
	c.UpdatedAt = when
}

func (c *Conversation) Archive() {
	c.Status = StatusArchived
	c.UpdatedAt = time.Now().UTC()
}

func ValidateListFilter(filter ListFilter) error {
	if filter.Limit < 0 || filter.Limit > MaxListLimit {
		return ErrInvalidLimit
	}
	return nil
}

func validRole(role string) bool {
	switch role {
	case RoleUser, RoleAssistant, RoleSystem:
		return true
	default:
		return false
	}
}
