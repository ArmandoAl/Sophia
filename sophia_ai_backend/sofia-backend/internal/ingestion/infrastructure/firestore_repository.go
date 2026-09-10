package infrastructure

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"sort"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/armandoalvarado/sofia-backend/internal/ingestion/domain"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	ingestionBatchesCollection     = "ingestion_batches"
	ingestionExternalIDsCollection = "ingestion_external_ids"
)

var _ domain.BatchRepository = (*FirestoreBatchRepository)(nil)

type FirestoreBatchRepository struct {
	client  *firestore.Client
	timeout time.Duration
}

func NewFirestoreBatchRepository(client *firestore.Client) *FirestoreBatchRepository {
	return &FirestoreBatchRepository{client: client, timeout: 10 * time.Second}
}

func (r *FirestoreBatchRepository) Create(ctx context.Context, batch *domain.Batch) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	err := r.client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		if err := tx.Create(r.client.Collection(ingestionBatchesCollection).Doc(batch.ID), batch); err != nil {
			return err
		}
		for _, conversation := range batch.Conversations {
			ref := r.client.Collection(ingestionExternalIDsCollection).Doc(externalMarkerID(batch.UserID, conversation.Source, conversation.ExternalID))
			if err := tx.Create(ref, map[string]any{"batch_id": batch.ID, "user_id": batch.UserID, "source": conversation.Source, "external_id": conversation.ExternalID}); err != nil {
				return err
			}
		}
		return nil
	})
	if status.Code(err) == codes.AlreadyExists {
		return domain.ErrExternalIDExists
	}
	return err
}

func (r *FirestoreBatchRepository) ListByUser(ctx context.Context, userID string) ([]*domain.Batch, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	iter := r.client.Collection(ingestionBatchesCollection).Where("user_id", "==", userID).Documents(ctx)
	defer iter.Stop()
	batches := make([]*domain.Batch, 0)
	for {
		doc, err := iter.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, err
		}
		var batch domain.Batch
		if err := doc.DataTo(&batch); err != nil {
			return nil, err
		}
		batches = append(batches, &batch)
	}
	sort.Slice(batches, func(i, j int) bool { return batches[i].CreatedAt.After(batches[j].CreatedAt) })
	return batches, nil
}

func (r *FirestoreBatchRepository) FindByID(ctx context.Context, userID, batchID string) (*domain.Batch, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	doc, err := r.client.Collection(ingestionBatchesCollection).Doc(batchID).Get(ctx)
	if status.Code(err) == codes.NotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var batch domain.Batch
	if err := doc.DataTo(&batch); err != nil {
		return nil, err
	}
	if batch.UserID != userID {
		return nil, nil
	}
	return &batch, nil
}

func (r *FirestoreBatchRepository) FindByExternalID(ctx context.Context, userID, source, externalID string) (*domain.Batch, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	doc, err := r.client.Collection(ingestionExternalIDsCollection).Doc(externalMarkerID(userID, source, externalID)).Get(ctx)
	if status.Code(err) == codes.NotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	batchID, _ := doc.Data()["batch_id"].(string)
	return r.FindByID(ctx, userID, batchID)
}

func (r *FirestoreBatchRepository) ClaimNext(ctx context.Context, workerID string, now, leaseUntil time.Time) (*domain.Batch, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	ref, err := r.firstEligible(ctx, now)
	if err != nil || ref == nil {
		return nil, err
	}
	var claimed *domain.Batch
	err = r.client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		doc, err := tx.Get(ref)
		if err != nil {
			return err
		}
		var batch domain.Batch
		if err := doc.DataTo(&batch); err != nil {
			return err
		}
		if batch.Status != domain.StatusPending && (batch.Status != domain.StatusProcessing || batch.ProcessingUntil.After(now)) {
			return nil
		}
		batch.Status, batch.ProcessingBy, batch.ProcessingUntil, batch.UpdatedAt = domain.StatusProcessing, workerID, leaseUntil, now
		if err := tx.Set(ref, &batch); err != nil {
			return err
		}
		claimed = &batch
		return nil
	})
	return claimed, err
}

func (r *FirestoreBatchRepository) firstEligible(ctx context.Context, now time.Time) (*firestore.DocumentRef, error) {
	queries := []firestore.Query{
		r.client.Collection(ingestionBatchesCollection).Where("status", "==", domain.StatusPending).OrderBy("created_at", firestore.Asc).Limit(1),
		r.client.Collection(ingestionBatchesCollection).Where("status", "==", domain.StatusProcessing).Where("processing_until", "<=", now).OrderBy("processing_until", firestore.Asc).Limit(1),
	}
	for _, query := range queries {
		iter := query.Documents(ctx)
		doc, err := iter.Next()
		iter.Stop()
		if err == nil {
			return doc.Ref, nil
		}
		if !errors.Is(err, iterator.Done) {
			return nil, err
		}
	}
	return nil, nil
}

func (r *FirestoreBatchRepository) Update(ctx context.Context, batch *domain.Batch) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	_, err := r.client.Collection(ingestionBatchesCollection).Doc(batch.ID).Set(ctx, batch)
	return err
}

func externalMarkerID(userID, source, externalID string) string {
	sum := sha256.Sum256([]byte(userID + "\x00" + source + "\x00" + externalID))
	return hex.EncodeToString(sum[:])
}
