package infrastructure

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/armandoalvarado/sofia-backend/internal/ingestion/domain"
)

var _ domain.BatchRepository = (*InMemoryBatchRepository)(nil)

type InMemoryBatchRepository struct {
	mu      sync.Mutex
	batches map[string]*domain.Batch
}

func NewInMemoryBatchRepository() *InMemoryBatchRepository {
	return &InMemoryBatchRepository{batches: make(map[string]*domain.Batch)}
}

func (r *InMemoryBatchRepository) Create(ctx context.Context, batch *domain.Batch) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, existing := range r.batches {
		for _, old := range existing.Conversations {
			for _, value := range batch.Conversations {
				if existing.UserID == batch.UserID && old.Source == value.Source && old.ExternalID == value.ExternalID {
					return domain.ErrExternalIDExists
				}
			}
		}
	}
	r.batches[batch.ID] = cloneBatch(batch)
	return nil
}

func (r *InMemoryBatchRepository) FindByID(ctx context.Context, userID, batchID string) (*domain.Batch, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	batch := r.batches[batchID]
	if batch == nil || batch.UserID != userID {
		return nil, nil
	}
	return cloneBatch(batch), nil
}

func (r *InMemoryBatchRepository) FindByExternalID(ctx context.Context, userID, source, externalID string) (*domain.Batch, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, batch := range r.batches {
		if batch.UserID != userID {
			continue
		}
		for _, conversation := range batch.Conversations {
			if conversation.Source == source && conversation.ExternalID == externalID {
				return cloneBatch(batch), nil
			}
		}
	}
	return nil, nil
}

func (r *InMemoryBatchRepository) ClaimNext(ctx context.Context, workerID string, now, leaseUntil time.Time) (*domain.Batch, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	eligible := make([]*domain.Batch, 0)
	for _, batch := range r.batches {
		if batch.Status == domain.StatusPending || (batch.Status == domain.StatusProcessing && !batch.ProcessingUntil.After(now)) {
			eligible = append(eligible, batch)
		}
	}
	sort.Slice(eligible, func(i, j int) bool { return eligible[i].CreatedAt.Before(eligible[j].CreatedAt) })
	if len(eligible) == 0 {
		return nil, nil
	}
	batch := cloneBatch(eligible[0])
	batch.Status = domain.StatusProcessing
	batch.ProcessingBy = workerID
	batch.ProcessingUntil = leaseUntil
	batch.UpdatedAt = now
	r.batches[batch.ID] = cloneBatch(batch)
	return batch, nil
}

func (r *InMemoryBatchRepository) Update(ctx context.Context, batch *domain.Batch) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.batches[batch.ID] == nil {
		return domain.ErrBatchNotFound
	}
	r.batches[batch.ID] = cloneBatch(batch)
	return nil
}

func cloneBatch(batch *domain.Batch) *domain.Batch {
	if batch == nil {
		return nil
	}
	cp := *batch
	cp.Conversations = make([]domain.RawConversation, len(batch.Conversations))
	for i, conversation := range batch.Conversations {
		cp.Conversations[i] = conversation
		cp.Conversations[i].Participants = append([]string(nil), conversation.Participants...)
		cp.Conversations[i].Messages = append([]domain.RawMessage(nil), conversation.Messages...)
	}
	return &cp
}
