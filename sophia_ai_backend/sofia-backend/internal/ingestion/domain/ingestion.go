package domain

import (
	"context"
	"errors"
	"strings"
	"time"
)

const (
	SourceWhatsApp  = "whatsapp"
	SourceInstagram = "instagram"
	SourceFacebook  = "facebook"
	SourceManual    = "manual"

	StatusPending    = "pending"
	StatusProcessing = "processing"
	StatusDone       = "done"
	StatusFailed     = "failed"
)

var (
	ErrInvalidBatch     = errors.New("invalid ingestion batch")
	ErrBatchNotFound    = errors.New("ingestion batch not found")
	ErrExternalIDExists = errors.New("conversation external id already exists")
)

type RawMessage struct {
	Author string    `json:"author" firestore:"author"`
	Text   string    `json:"text" firestore:"text"`
	SentAt time.Time `json:"sent_at" firestore:"sent_at"`
}

type RawConversation struct {
	Source       string       `json:"source" firestore:"source"`
	ExternalID   string       `json:"external_id" firestore:"external_id"`
	Participants []string     `json:"participants" firestore:"participants"`
	Messages     []RawMessage `json:"messages" firestore:"messages"`
}

type Batch struct {
	ID                string            `json:"id" firestore:"id"`
	UserID            string            `json:"user_id" firestore:"user_id"`
	Conversations     []RawConversation `json:"conversations,omitempty" firestore:"conversations"`
	Status            string            `json:"status" firestore:"status"`
	ConversationCount int               `json:"conversation_count" firestore:"conversation_count"`
	MessageCount      int               `json:"message_count" firestore:"message_count"`
	WindowsProcessed  int               `json:"windows_processed" firestore:"windows_processed"`
	BeliefsCreated    int               `json:"beliefs_created" firestore:"beliefs_created"`
	InputTokens       int               `json:"input_tokens" firestore:"input_tokens"`
	OutputTokens      int               `json:"output_tokens" firestore:"output_tokens"`
	FailureReason     string            `json:"failure_reason,omitempty" firestore:"failure_reason"`
	ProcessingBy      string            `json:"-" firestore:"processing_by"`
	ProcessingUntil   time.Time         `json:"-" firestore:"processing_until"`
	CreatedAt         time.Time         `json:"created_at" firestore:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at" firestore:"updated_at"`
}

type BatchRepository interface {
	Create(ctx context.Context, batch *Batch) error
	FindByID(ctx context.Context, userID, batchID string) (*Batch, error)
	FindByExternalID(ctx context.Context, userID, source, externalID string) (*Batch, error)
	ClaimNext(ctx context.Context, workerID string, now, leaseUntil time.Time) (*Batch, error)
	Update(ctx context.Context, batch *Batch) error
}

func ValidateConversations(values []RawConversation) error {
	if len(values) == 0 {
		return ErrInvalidBatch
	}
	seen := make(map[string]struct{}, len(values))
	for i := range values {
		conversation := &values[i]
		conversation.Source = strings.ToLower(strings.TrimSpace(conversation.Source))
		conversation.ExternalID = strings.TrimSpace(conversation.ExternalID)
		key := conversation.Source + "\x00" + conversation.ExternalID
		if !IsValidSource(conversation.Source) || conversation.ExternalID == "" || len(conversation.Participants) == 0 || len(conversation.Messages) == 0 {
			return ErrInvalidBatch
		}
		if _, ok := seen[key]; ok {
			return ErrInvalidBatch
		}
		seen[key] = struct{}{}
		for j := range conversation.Participants {
			conversation.Participants[j] = strings.TrimSpace(conversation.Participants[j])
			if conversation.Participants[j] == "" {
				return ErrInvalidBatch
			}
		}
		for j := range conversation.Messages {
			message := &conversation.Messages[j]
			message.Author = strings.TrimSpace(message.Author)
			message.Text = strings.TrimSpace(message.Text)
			if message.Author == "" || message.Text == "" || message.SentAt.IsZero() {
				return ErrInvalidBatch
			}
			message.SentAt = message.SentAt.UTC()
		}
	}
	return nil
}

func IsValidSource(value string) bool {
	switch value {
	case SourceWhatsApp, SourceInstagram, SourceFacebook, SourceManual:
		return true
	default:
		return false
	}
}
