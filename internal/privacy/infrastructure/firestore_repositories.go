package infrastructure

import (
	"context"
	"errors"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/armandoalvarado/sofia-backend/internal/privacy/domain"
	"google.golang.org/api/iterator"
)

const (
	auditLogsCollection      = "audit_logs"
	deleteRequestsCollection = "delete_requests"
)

type FirestoreAuditLogRepository struct {
	client  *firestore.Client
	timeout time.Duration
}

type FirestoreDeleteRequestRepository struct {
	client  *firestore.Client
	timeout time.Duration
}

func NewFirestoreAuditLogRepository(client *firestore.Client) *FirestoreAuditLogRepository {
	return &FirestoreAuditLogRepository{client: client, timeout: 5 * time.Second}
}

func NewFirestoreDeleteRequestRepository(client *firestore.Client) *FirestoreDeleteRequestRepository {
	return &FirestoreDeleteRequestRepository{client: client, timeout: 5 * time.Second}
}

func (r *FirestoreAuditLogRepository) Create(ctx context.Context, log *domain.AuditLog) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	_, err := r.client.Collection(auditLogsCollection).Doc(log.ID).Set(ctx, auditLogToDocument(log))
	return err
}

func (r *FirestoreAuditLogRepository) ListByUser(ctx context.Context, userID string, limit int) ([]*domain.AuditLog, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	query := r.client.Collection(auditLogsCollection).Where("user_id", "==", userID)
	if limit > 0 {
		query = query.Limit(limit)
	}
	iter := query.Documents(ctx)
	defer iter.Stop()
	result := []*domain.AuditLog{}
	for {
		doc, err := iter.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, err
		}
		var stored firestoreAuditLog
		if err := doc.DataTo(&stored); err != nil {
			return nil, err
		}
		result = append(result, stored.toDomain())
	}
	return result, nil
}

func (r *FirestoreDeleteRequestRepository) Create(ctx context.Context, request *domain.DeleteRequest) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	_, err := r.client.Collection(deleteRequestsCollection).Doc(request.ID).Set(ctx, deleteRequestToDocument(request))
	return err
}

func (r *FirestoreDeleteRequestRepository) ListByUser(ctx context.Context, userID string, limit int) ([]*domain.DeleteRequest, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	query := r.client.Collection(deleteRequestsCollection).Where("user_id", "==", userID)
	if limit > 0 {
		query = query.Limit(limit)
	}
	iter := query.Documents(ctx)
	defer iter.Stop()
	result := []*domain.DeleteRequest{}
	for {
		doc, err := iter.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, err
		}
		var stored firestoreDeleteRequest
		if err := doc.DataTo(&stored); err != nil {
			return nil, err
		}
		result = append(result, stored.toDomain())
	}
	return result, nil
}

func (r *FirestoreDeleteRequestRepository) ListByStatus(ctx context.Context, status string, limit int) ([]*domain.DeleteRequest, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	query := r.client.Collection(deleteRequestsCollection).Where("status", "==", status).OrderBy("created_at", firestore.Asc)
	if limit > 0 {
		query = query.Limit(limit)
	}
	return r.collect(ctx, query)
}

func (r *FirestoreDeleteRequestRepository) Update(ctx context.Context, request *domain.DeleteRequest) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	_, err := r.client.Collection(deleteRequestsCollection).Doc(request.ID).Set(ctx, deleteRequestToDocument(request))
	return err
}

func (r *FirestoreDeleteRequestRepository) collect(ctx context.Context, query firestore.Query) ([]*domain.DeleteRequest, error) {
	iter := query.Documents(ctx)
	defer iter.Stop()
	result := []*domain.DeleteRequest{}
	for {
		doc, err := iter.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, err
		}
		var stored firestoreDeleteRequest
		if err := doc.DataTo(&stored); err != nil {
			return nil, err
		}
		result = append(result, stored.toDomain())
	}
	return result, nil
}

type firestoreAuditLog struct {
	ID           string         `firestore:"id"`
	UserID       string         `firestore:"user_id"`
	Action       string         `firestore:"action"`
	ResourceType string         `firestore:"resource_type"`
	ResourceID   string         `firestore:"resource_id"`
	CreatedAt    time.Time      `firestore:"created_at"`
	Metadata     map[string]any `firestore:"metadata"`
}

type firestoreDeleteRequest struct {
	ID        string    `firestore:"id"`
	UserID    string    `firestore:"user_id"`
	Status    string    `firestore:"status"`
	Reason    string    `firestore:"reason"`
	CreatedAt time.Time `firestore:"created_at"`
	UpdatedAt time.Time `firestore:"updated_at"`
}

func auditLogToDocument(log *domain.AuditLog) firestoreAuditLog {
	return firestoreAuditLog{ID: log.ID, UserID: log.UserID, Action: log.Action, ResourceType: log.ResourceType, ResourceID: log.ResourceID, CreatedAt: log.CreatedAt, Metadata: log.Metadata}
}

func (l firestoreAuditLog) toDomain() *domain.AuditLog {
	return &domain.AuditLog{ID: l.ID, UserID: l.UserID, Action: l.Action, ResourceType: l.ResourceType, ResourceID: l.ResourceID, CreatedAt: l.CreatedAt, Metadata: l.Metadata}
}

func deleteRequestToDocument(request *domain.DeleteRequest) firestoreDeleteRequest {
	return firestoreDeleteRequest{ID: request.ID, UserID: request.UserID, Status: request.Status, Reason: request.Reason, CreatedAt: request.CreatedAt, UpdatedAt: request.UpdatedAt}
}

func (r firestoreDeleteRequest) toDomain() *domain.DeleteRequest {
	return &domain.DeleteRequest{ID: r.ID, UserID: r.UserID, Status: r.Status, Reason: r.Reason, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt}
}
