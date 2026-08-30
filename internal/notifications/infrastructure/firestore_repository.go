package infrastructure

import (
	"context"
	"errors"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/armandoalvarado/sofia-backend/internal/notifications/domain"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const deviceTokensCollection = "device_tokens"
const deliveryAttemptsCollection = "notification_delivery_attempts"

type FirestoreDeviceTokenRepository struct {
	client  *firestore.Client
	timeout time.Duration
}

func NewFirestoreDeviceTokenRepository(client *firestore.Client) *FirestoreDeviceTokenRepository {
	return &FirestoreDeviceTokenRepository{client: client, timeout: 5 * time.Second}
}

func (r *FirestoreDeviceTokenRepository) Save(ctx context.Context, token *domain.DeviceToken) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	_, err := r.client.Collection(deviceTokensCollection).Doc(token.ID).Set(ctx, deviceTokenToDocument(token))
	return err
}

func (r *FirestoreDeviceTokenRepository) FindByID(ctx context.Context, tokenID string) (*domain.DeviceToken, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	snap, err := r.client.Collection(deviceTokensCollection).Doc(tokenID).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, domain.ErrDeviceTokenNotFound
		}
		return nil, err
	}
	var stored firestoreDeviceToken
	if err := snap.DataTo(&stored); err != nil {
		return nil, err
	}
	return stored.toDomain(), nil
}

func (r *FirestoreDeviceTokenRepository) ListByUser(ctx context.Context, userID string) ([]*domain.DeviceToken, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	iter := r.client.Collection(deviceTokensCollection).
		Where("user_id", "==", userID).
		Where("enabled", "==", true).
		Documents(ctx)
	defer iter.Stop()

	result := []*domain.DeviceToken{}
	for {
		doc, err := iter.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, err
		}
		var stored firestoreDeviceToken
		if err := doc.DataTo(&stored); err != nil {
			return nil, err
		}
		result = append(result, stored.toDomain())
	}
	return result, nil
}

func (r *FirestoreDeviceTokenRepository) Delete(ctx context.Context, userID, tokenID string) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	doc := r.client.Collection(deviceTokensCollection).Doc(tokenID)
	snap, err := doc.Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return domain.ErrDeviceTokenNotFound
		}
		return err
	}
	var stored firestoreDeviceToken
	if err := snap.DataTo(&stored); err != nil {
		return err
	}
	if stored.UserID != userID {
		return domain.ErrDeviceTokenNotFound
	}
	_, err = doc.Delete(ctx)
	return err
}

type firestoreDeviceToken struct {
	ID         string    `firestore:"id"`
	UserID     string    `firestore:"user_id"`
	Platform   string    `firestore:"platform"`
	Token      string    `firestore:"token"`
	Enabled    bool      `firestore:"enabled"`
	CreatedAt  time.Time `firestore:"created_at"`
	UpdatedAt  time.Time `firestore:"updated_at"`
	LastSeenAt time.Time `firestore:"last_seen_at"`
}

func deviceTokenToDocument(token *domain.DeviceToken) firestoreDeviceToken {
	return firestoreDeviceToken{
		ID:         token.ID,
		UserID:     token.UserID,
		Platform:   token.Platform,
		Token:      token.Token,
		Enabled:    token.Enabled,
		CreatedAt:  token.CreatedAt,
		UpdatedAt:  token.UpdatedAt,
		LastSeenAt: token.LastSeenAt,
	}
}

func (t firestoreDeviceToken) toDomain() *domain.DeviceToken {
	return &domain.DeviceToken{
		ID:         t.ID,
		UserID:     t.UserID,
		Platform:   t.Platform,
		Token:      t.Token,
		Enabled:    t.Enabled,
		CreatedAt:  t.CreatedAt,
		UpdatedAt:  t.UpdatedAt,
		LastSeenAt: t.LastSeenAt,
	}
}

type FirestoreDeliveryAttemptRepository struct {
	client  *firestore.Client
	timeout time.Duration
}

func NewFirestoreDeliveryAttemptRepository(client *firestore.Client) *FirestoreDeliveryAttemptRepository {
	return &FirestoreDeliveryAttemptRepository{client: client, timeout: 5 * time.Second}
}

func (r *FirestoreDeliveryAttemptRepository) Create(ctx context.Context, attempt *domain.DeliveryAttempt) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	_, err := r.client.Collection(deliveryAttemptsCollection).Doc(attempt.ID).Set(ctx, firestoreDeliveryAttempt{
		ID:         attempt.ID,
		UserID:     attempt.UserID,
		ReminderID: attempt.ReminderID,
		Provider:   attempt.Provider,
		Status:     attempt.Status,
		DryRun:     attempt.DryRun,
		ErrorType:  attempt.ErrorType,
		CreatedAt:  attempt.CreatedAt,
	})
	return err
}

type firestoreDeliveryAttempt struct {
	ID         string    `firestore:"id"`
	UserID     string    `firestore:"user_id"`
	ReminderID string    `firestore:"reminder_id"`
	Provider   string    `firestore:"provider"`
	Status     string    `firestore:"status"`
	DryRun     bool      `firestore:"dry_run"`
	ErrorType  string    `firestore:"error_type"`
	CreatedAt  time.Time `firestore:"created_at"`
}
