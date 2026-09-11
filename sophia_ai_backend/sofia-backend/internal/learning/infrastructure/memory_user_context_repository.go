package infrastructure

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/armandoalvarado/sofia-backend/internal/learning/domain"
)

var _ domain.UserContextRepository = (*InMemoryUserContextRepository)(nil)

type InMemoryUserContextRepository struct {
	mu       sync.Mutex
	contexts map[string]*domain.UserContext
}

var _ domain.EntityCandidateRepository = (*InMemoryEntityCandidateRepository)(nil)

type InMemoryEntityCandidateRepository struct {
	mu         sync.Mutex
	candidates map[string]*domain.EntityCandidate
}

func NewInMemoryUserContextRepository() *InMemoryUserContextRepository {
	return &InMemoryUserContextRepository{contexts: make(map[string]*domain.UserContext)}
}

func NewInMemoryEntityCandidateRepository() *InMemoryEntityCandidateRepository {
	return &InMemoryEntityCandidateRepository{candidates: make(map[string]*domain.EntityCandidate)}
}

func (r *InMemoryEntityCandidateRepository) UpsertMention(_ context.Context, userID, name, normalizedName, sampleContext string, now time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := userID + "\x00" + normalizedName
	value := r.candidates[key]
	if value == nil {
		value = &domain.EntityCandidate{UserID: userID, Name: name, NormalizedName: normalizedName, FirstSeenAt: now}
	}
	value.MentionCount++
	value.LastSeenAt = now
	if sampleContext != "" && len(value.SampleContexts) < 3 && !containsString(value.SampleContexts, sampleContext) {
		value.SampleContexts = append(value.SampleContexts, sampleContext)
	}
	r.candidates[key] = cloneEntityCandidate(value)
	return nil
}

func (r *InMemoryEntityCandidateRepository) List(_ context.Context, userID string) ([]*domain.EntityCandidate, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	result := make([]*domain.EntityCandidate, 0)
	for _, value := range r.candidates {
		if value.UserID == userID {
			result = append(result, cloneEntityCandidate(value))
		}
	}
	sort.SliceStable(result, func(i, j int) bool { return result[i].LastSeenAt.After(result[j].LastSeenAt) })
	return result, nil
}

func (r *InMemoryEntityCandidateRepository) Delete(_ context.Context, userID, normalizedName string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.candidates, userID+"\x00"+normalizedName)
	return nil
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

func (r *InMemoryUserContextRepository) FindByID(_ context.Context, userID, contextID string) (*domain.UserContext, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	value, ok := r.contexts[contextID]
	if !ok || value.UserID != userID {
		return nil, domain.ErrUserContextNotFound
	}
	return cloneUserContext(value), nil
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

func (r *InMemoryUserContextRepository) Merge(_ context.Context, userID, sourceID string, target *domain.UserContext) (*domain.UserContext, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	value, ok := r.contexts[sourceID]
	if !ok || value.UserID != userID {
		return nil, domain.ErrUserContextNotFound
	}
	cp := cloneUserContext(value)
	if err := cp.Merge(target); err != nil {
		return nil, err
	}
	r.contexts[sourceID] = cloneUserContext(cp)
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

func cloneEntityCandidate(value *domain.EntityCandidate) *domain.EntityCandidate {
	cp := *value
	cp.SampleContexts = cloneStrings(value.SampleContexts)
	return &cp
}

func containsString(values []string, candidate string) bool {
	for _, value := range values {
		if value == candidate {
			return true
		}
	}
	return false
}
