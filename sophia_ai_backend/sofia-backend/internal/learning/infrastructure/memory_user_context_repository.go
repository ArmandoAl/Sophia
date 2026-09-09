package infrastructure

import (
	"context"
	"sort"
	"sync"

	"github.com/armandoalvarado/sofia-backend/internal/learning/domain"
)

var _ domain.UserContextRepository = (*InMemoryUserContextRepository)(nil)

type InMemoryUserContextRepository struct {
	mu       sync.Mutex
	contexts map[string]*domain.UserContext
}

func NewInMemoryUserContextRepository() *InMemoryUserContextRepository {
	return &InMemoryUserContextRepository{contexts: make(map[string]*domain.UserContext)}
}

func (r *InMemoryUserContextRepository) Create(_ context.Context, value *domain.UserContext) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.contexts[value.ID] = cloneUserContext(value)
	return nil
}

func (r *InMemoryUserContextRepository) List(_ context.Context, userID string) ([]*domain.UserContext, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	result := make([]*domain.UserContext, 0)
	for _, value := range r.contexts {
		if value.UserID == userID {
			result = append(result, cloneUserContext(value))
		}
	}
	sort.SliceStable(result, func(i, j int) bool { return result[i].UpdatedAt.After(result[j].UpdatedAt) })
	return result, nil
}

func (r *InMemoryUserContextRepository) FindBySlug(_ context.Context, userID, kind, slug string) (*domain.UserContext, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, value := range r.contexts {
		if value.UserID == userID && value.Kind == kind && value.Slug == slug {
			return cloneUserContext(value), nil
		}
	}
	return nil, nil
}

func (r *InMemoryUserContextRepository) Update(_ context.Context, userID, contextID string, update domain.UserContextUpdate) (*domain.UserContext, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	value, ok := r.contexts[contextID]
	if !ok || value.UserID != userID {
		return nil, domain.ErrUserContextNotFound
	}
	cp := cloneUserContext(value)
	if err := cp.Apply(update); err != nil {
		return nil, err
	}
	r.contexts[contextID] = cloneUserContext(cp)
	return cp, nil
}

func (r *InMemoryUserContextRepository) Archive(_ context.Context, userID, contextID string) (*domain.UserContext, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	value, ok := r.contexts[contextID]
	if !ok || value.UserID != userID {
		return nil, domain.ErrUserContextNotFound
	}
	cp := cloneUserContext(value)
	cp.Archive()
	r.contexts[contextID] = cloneUserContext(cp)
	return cp, nil
}

func cloneUserContext(value *domain.UserContext) *domain.UserContext {
	if value == nil {
		return nil
	}
	cp := *value
	cp.Aliases = cloneStrings(value.Aliases)
	return &cp
}
