package application

import (
	"context"
	"errors"
	"time"

	"github.com/armandoalvarado/sofia-backend/internal/ingestion/domain"
	"github.com/google/uuid"
)

type BeliefRetirer interface {
	RetireBeliefsByBatch(ctx context.Context, userID, batchID string) (int, error)
}

type Service struct {
	batches domain.BatchRepository
	beliefs BeliefRetirer
}

func NewService(batches domain.BatchRepository, beliefs BeliefRetirer) *Service {
	return &Service{batches: batches, beliefs: beliefs}
}

func (s *Service) Enqueue(ctx context.Context, userID string, conversations []domain.RawConversation) (*domain.Batch, error) {
	if userID == "" || domain.ValidateConversations(conversations) != nil {
		return nil, domain.ErrInvalidBatch
	}
	for _, conversation := range conversations {
		existing, err := s.batches.FindByExternalID(ctx, userID, conversation.Source, conversation.ExternalID)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			return existing, nil
		}
	}
	now := time.Now().UTC()
	batch := &domain.Batch{ID: uuid.New().String(), UserID: userID, Conversations: conversations, Status: domain.StatusPending, ConversationCount: len(conversations), CreatedAt: now, UpdatedAt: now}
	for _, conversation := range conversations {
		batch.MessageCount += len(conversation.Messages)
	}
	if err := s.batches.Create(ctx, batch); err != nil {
		if errors.Is(err, domain.ErrExternalIDExists) {
			first := conversations[0]
			return s.batches.FindByExternalID(ctx, userID, first.Source, first.ExternalID)
		}
		return nil, err
	}
	return batch, nil
}

func (s *Service) RetireBatch(ctx context.Context, userID, batchID string) (int, error) {
	batch, err := s.batches.FindByID(ctx, userID, batchID)
	if err != nil {
		return 0, err
	}
	if batch == nil {
		return 0, domain.ErrBatchNotFound
	}
	return s.beliefs.RetireBeliefsByBatch(ctx, userID, batchID)
}
