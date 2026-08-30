package infrastructure

import (
	"context"
	"sync"

	"github.com/armandoalvarado/sofia-backend/internal/notifications/domain"
)

type InMemoryDeviceTokenRepository struct {
	mu     sync.Mutex
	tokens map[string]*domain.DeviceToken
}

func NewInMemoryDeviceTokenRepository() *InMemoryDeviceTokenRepository {
	return &InMemoryDeviceTokenRepository{tokens: map[string]*domain.DeviceToken{}}
}

func (r *InMemoryDeviceTokenRepository) Save(ctx context.Context, token *domain.DeviceToken) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tokens[token.ID] = cloneDeviceToken(token)
	return nil
}

func (r *InMemoryDeviceTokenRepository) FindByID(ctx context.Context, tokenID string) (*domain.DeviceToken, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	token, ok := r.tokens[tokenID]
	if !ok {
		return nil, domain.ErrDeviceTokenNotFound
	}
	return cloneDeviceToken(token), nil
}

func (r *InMemoryDeviceTokenRepository) ListByUser(ctx context.Context, userID string) ([]*domain.DeviceToken, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	result := []*domain.DeviceToken{}
	for _, token := range r.tokens {
		if token.UserID == userID && token.Enabled {
			result = append(result, cloneDeviceToken(token))
		}
	}
	return result, nil
}

func (r *InMemoryDeviceTokenRepository) Delete(ctx context.Context, userID, tokenID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	token, ok := r.tokens[tokenID]
	if !ok || token.UserID != userID {
		return domain.ErrDeviceTokenNotFound
	}
	delete(r.tokens, tokenID)
	return nil
}

func cloneDeviceToken(token *domain.DeviceToken) *domain.DeviceToken {
	if token == nil {
		return nil
	}
	cp := *token
	return &cp
}

type InMemoryDeliveryAttemptRepository struct {
	mu       sync.Mutex
	attempts map[string]*domain.DeliveryAttempt
}

func NewInMemoryDeliveryAttemptRepository() *InMemoryDeliveryAttemptRepository {
	return &InMemoryDeliveryAttemptRepository{attempts: map[string]*domain.DeliveryAttempt{}}
}

func (r *InMemoryDeliveryAttemptRepository) Create(ctx context.Context, attempt *domain.DeliveryAttempt) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	cp := *attempt
	r.attempts[attempt.ID] = &cp
	return nil
}

func (r *InMemoryDeliveryAttemptRepository) ListAll() []*domain.DeliveryAttempt {
	r.mu.Lock()
	defer r.mu.Unlock()
	result := make([]*domain.DeliveryAttempt, 0, len(r.attempts))
	for _, attempt := range r.attempts {
		cp := *attempt
		result = append(result, &cp)
	}
	return result
}
