package infrastructure

import (
	"context"
	"sort"
	"sync"

	"github.com/armandoalvarado/sofia-backend/internal/learning/domain"
)

var _ domain.PromptVersionRepository = (*InMemoryPromptVersionRepository)(nil)

type InMemoryPromptVersionRepository struct {
	mu       sync.Mutex
	versions map[string]*domain.PromptVersion
}

func NewInMemoryPromptVersionRepository() *InMemoryPromptVersionRepository {
	return &InMemoryPromptVersionRepository{versions: make(map[string]*domain.PromptVersion)}
}

func (r *InMemoryPromptVersionRepository) CreateActive(ctx context.Context, version *domain.PromptVersion) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	next := 1
	for _, existing := range r.versions {
		if existing.UserID != version.UserID {
			continue
		}
		existing.Active = false
		if existing.Version >= next {
			next = existing.Version + 1
		}
	}
	version.Version = next
	version.Active = true
	if version.TokenCount == 0 {
		version.TokenCount = domain.ApproximateTokens(version.Content)
	}
	r.versions[version.ID] = clonePromptVersion(version)
	return nil
}

func (r *InMemoryPromptVersionRepository) GetActive(ctx context.Context, userID string) (*domain.PromptVersion, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, version := range r.versions {
		if version.UserID == userID && version.Active {
			return clonePromptVersion(version), nil
		}
	}
	return nil, nil
}

func (r *InMemoryPromptVersionRepository) ListVersions(ctx context.Context, userID string, limit int) ([]*domain.PromptVersion, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	result := make([]*domain.PromptVersion, 0)
	for _, version := range r.versions {
		if version.UserID != userID {
			continue
		}
		result = append(result, clonePromptVersion(version))
	}
	sort.SliceStable(result, func(i, j int) bool {
		return result[i].Version > result[j].Version
	})
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

func clonePromptVersion(version *domain.PromptVersion) *domain.PromptVersion {
	if version == nil {
		return nil
	}
	cp := *version
	cp.BeliefIDs = cloneStrings(version.BeliefIDs)
	return &cp
}
