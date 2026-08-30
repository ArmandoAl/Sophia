package infrastructure

import (
	"context"
	"errors"
	"sort"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/armandoalvarado/sofia-backend/internal/conversations/domain"
	"github.com/armandoalvarado/sofia-backend/internal/database"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	conversationsCollection = database.V1ConversationsCollection
	messagesCollection      = database.V1ConversationMessagesCollection
)

type FirestoreConversationRepository struct {
	client  *firestore.Client
	timeout time.Duration
}

type FirestoreConversationMessageRepository struct {
	client  *firestore.Client
	timeout time.Duration
}

type firestoreConversation struct {
	ID            string     `firestore:"id"`
	UserID        string     `firestore:"user_id"`
	Title         string     `firestore:"title"`
	Status        string     `firestore:"status"`
	CreatedAt     time.Time  `firestore:"created_at"`
	UpdatedAt     time.Time  `firestore:"updated_at"`
	LastMessageAt *time.Time `firestore:"last_message_at,omitempty"`
}

type firestoreMessage struct {
	ID               string    `firestore:"id"`
	ConversationID   string    `firestore:"conversation_id"`
	UserID           string    `firestore:"user_id"`
	Role             string    `firestore:"role"`
	Content          string    `firestore:"content"`
	RuntimeRequestID string    `firestore:"runtime_request_id"`
	CreatedAt        time.Time `firestore:"created_at"`
}

func NewFirestoreConversationRepository(client *firestore.Client) *FirestoreConversationRepository {
	return &FirestoreConversationRepository{client: client, timeout: 5 * time.Second}
}

func NewFirestoreConversationMessageRepository(client *firestore.Client) *FirestoreConversationMessageRepository {
	return &FirestoreConversationMessageRepository{client: client, timeout: 5 * time.Second}
}

func (r *FirestoreConversationRepository) Create(ctx context.Context, conversation *domain.Conversation) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	_, err := r.client.Collection(conversationsCollection).Doc(conversation.ID).Set(ctx, conversationToDocument(conversation))
	return err
}

func (r *FirestoreConversationRepository) FindByID(ctx context.Context, userID, conversationID string) (*domain.Conversation, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	doc, err := r.client.Collection(conversationsCollection).Doc(conversationID).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, nil
		}
		return nil, err
	}
	conversation, err := documentToConversation(doc)
	if err != nil {
		return nil, err
	}
	if conversation.UserID != userID {
		return nil, nil
	}
	return conversation, nil
}

func (r *FirestoreConversationRepository) List(ctx context.Context, filter domain.ListFilter) ([]*domain.Conversation, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	query := r.client.Collection(conversationsCollection).Where("user_id", "==", filter.UserID)
	if filter.Status != "" {
		query = query.Where("status", "==", filter.Status)
	}
	query = query.OrderBy("updated_at", firestore.Desc).OrderBy("id", firestore.Desc)
	if filter.Cursor != "" {
		cursorDoc, err := r.client.Collection(conversationsCollection).Doc(filter.Cursor).Get(ctx)
		if err == nil {
			var cursor firestoreConversation
			if err := cursorDoc.DataTo(&cursor); err == nil && cursor.UserID == filter.UserID {
				query = query.StartAfter(cursor.UpdatedAt, cursor.ID)
			}
		} else if status.Code(err) != codes.NotFound {
			return nil, err
		}
	}
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	iter := query.Documents(ctx)
	defer iter.Stop()
	result := make([]*domain.Conversation, 0)
	for {
		doc, err := iter.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, err
		}
		conversation, err := documentToConversation(doc)
		if err != nil {
			return nil, err
		}
		result = append(result, conversation)
	}
	return result, nil
}

func (r *FirestoreConversationRepository) Update(ctx context.Context, conversation *domain.Conversation) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	_, err := r.client.Collection(conversationsCollection).Doc(conversation.ID).Set(ctx, conversationToDocument(conversation))
	return err
}

func (r *FirestoreConversationRepository) Archive(ctx context.Context, userID, conversationID string) (*domain.Conversation, error) {
	conversation, err := r.FindByID(ctx, userID, conversationID)
	if err != nil {
		return nil, err
	}
	if conversation == nil {
		return nil, domain.ErrConversationNotFound
	}
	conversation.Archive()
	if err := r.Update(ctx, conversation); err != nil {
		return nil, err
	}
	return conversation, nil
}

func (r *FirestoreConversationMessageRepository) Create(ctx context.Context, message *domain.ConversationMessage) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	_, err := r.client.Collection(messagesCollection).Doc(message.ID).Set(ctx, messageToDocument(message))
	return err
}

func (r *FirestoreConversationMessageRepository) List(ctx context.Context, userID, conversationID string, limit int, cursor string) ([]*domain.ConversationMessage, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	query := r.client.Collection(messagesCollection).
		Where("user_id", "==", userID).
		Where("conversation_id", "==", conversationID).
		OrderBy("created_at", firestore.Asc).
		OrderBy("id", firestore.Asc)
	if cursor != "" {
		cursorDoc, err := r.client.Collection(messagesCollection).Doc(cursor).Get(ctx)
		if err == nil {
			var stored firestoreMessage
			if err := cursorDoc.DataTo(&stored); err == nil && stored.UserID == userID && stored.ConversationID == conversationID {
				query = query.StartAfter(stored.CreatedAt, stored.ID)
			}
		} else if status.Code(err) != codes.NotFound {
			return nil, err
		}
	}
	if limit > 0 {
		query = query.Limit(limit)
	}
	iter := query.Documents(ctx)
	defer iter.Stop()
	result := make([]*domain.ConversationMessage, 0)
	for {
		doc, err := iter.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, err
		}
		message, err := documentToMessage(doc)
		if err != nil {
			return nil, err
		}
		result = append(result, message)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].CreatedAt.Equal(result[j].CreatedAt) {
			return result[i].ID < result[j].ID
		}
		return result[i].CreatedAt.Before(result[j].CreatedAt)
	})
	return result, nil
}

func conversationToDocument(conversation *domain.Conversation) firestoreConversation {
	return firestoreConversation{
		ID:            conversation.ID,
		UserID:        conversation.UserID,
		Title:         conversation.Title,
		Status:        conversation.Status,
		CreatedAt:     conversation.CreatedAt,
		UpdatedAt:     conversation.UpdatedAt,
		LastMessageAt: conversation.LastMessageAt,
	}
}

func documentToConversation(doc *firestore.DocumentSnapshot) (*domain.Conversation, error) {
	var stored firestoreConversation
	if err := doc.DataTo(&stored); err != nil {
		return nil, err
	}
	return &domain.Conversation{
		ID:            stored.ID,
		UserID:        stored.UserID,
		Title:         stored.Title,
		Status:        stored.Status,
		CreatedAt:     stored.CreatedAt,
		UpdatedAt:     stored.UpdatedAt,
		LastMessageAt: stored.LastMessageAt,
	}, nil
}

func messageToDocument(message *domain.ConversationMessage) firestoreMessage {
	return firestoreMessage{
		ID:               message.ID,
		ConversationID:   message.ConversationID,
		UserID:           message.UserID,
		Role:             message.Role,
		Content:          message.Content,
		RuntimeRequestID: message.RuntimeRequestID,
		CreatedAt:        message.CreatedAt,
	}
}

func documentToMessage(doc *firestore.DocumentSnapshot) (*domain.ConversationMessage, error) {
	var stored firestoreMessage
	if err := doc.DataTo(&stored); err != nil {
		return nil, err
	}
	return &domain.ConversationMessage{
		ID:               stored.ID,
		ConversationID:   stored.ConversationID,
		UserID:           stored.UserID,
		Role:             stored.Role,
		Content:          stored.Content,
		RuntimeRequestID: stored.RuntimeRequestID,
		CreatedAt:        stored.CreatedAt,
	}, nil
}
