package infrastructure

import (
	"context"
	"sort"
	"sync"

	"github.com/armandoalvarado/sofia-backend/internal/learning/domain"
	memorydomain "github.com/armandoalvarado/sofia-backend/internal/memory/domain"
)

var _ domain.BeliefRepository = (*InMemoryBeliefRepository)(nil)

type InMemoryBeliefRepository struct {
	mu      sync.Mutex
	beliefs map[string]*domain.Belief
}

func NewInMemoryBeliefRepository() *InMemoryBeliefRepository {
	return &InMemoryBeliefRepository{beliefs: make(map[string]*domain.Belief)}
}

func (r *InMemoryBeliefRepository) Create(ctx context.Context, belief *domain.Belief) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.beliefs[belief.ID] = cloneBelief(belief)
	return nil
}

func (r *InMemoryBeliefRepository) Update(ctx context.Context, belief *domain.Belief) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.beliefs[belief.ID]; !ok {
		return domain.ErrBeliefNotFound
	}
	r.beliefs[belief.ID] = cloneBelief(belief)
	return nil
}

func (r *InMemoryBeliefRepository) FindByID(ctx context.Context, userID, beliefID string) (*domain.Belief, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	belief, ok := r.beliefs[beliefID]
	if !ok || belief.UserID != userID {
		return nil, nil
	}
	return cloneBelief(belief), nil
}

func (r *InMemoryBeliefRepository) ListActive(ctx context.Context, userID string, limit int) ([]*domain.Belief, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	result := make([]*domain.Belief, 0)
	for _, belief := range r.beliefs {
		if belief.UserID != userID || belief.Status != domain.StatusActive {
			continue
		}
		result = append(result, cloneBelief(belief))
	}
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].Confidence == result[j].Confidence {
			return result[i].LastReinforcedAt.After(result[j].LastReinforcedAt)
		}
		return result[i].Confidence > result[j].Confidence
	})
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

func (r *InMemoryBeliefRepository) ListActiveByScope(ctx context.Context, userID, scope, scopeKey string, limit int) ([]*domain.Belief, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	result := make([]*domain.Belief, 0)
	for _, belief := range r.beliefs {
		if belief.UserID != userID || belief.Status != domain.StatusActive || belief.EffectiveScope() != scope || belief.ScopeKey != scopeKey {
			continue
		}
		result = append(result, cloneBelief(belief))
	}
	sort.SliceStable(result, func(i, j int) bool {
		return result[i].Confidence > result[j].Confidence
	})
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

func (r *InMemoryBeliefRepository) SearchByTerms(ctx context.Context, userID string, terms []string, limit int) ([]*domain.Belief, error) {
	terms = memorydomain.CapSearchTerms(terms)
	if len(terms) == 0 {
		return []*domain.Belief{}, nil
	}
	if limit <= 0 {
		limit = domain.DefaultSearchLimit()
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	candidates := make([]*domain.Belief, 0)
	for _, belief := range r.beliefs {
		if belief.UserID != userID || belief.Status != domain.StatusActive {
			continue
		}
		candidates = append(candidates, cloneBelief(belief))
	}
	return domain.RankBeliefsByTermMatches(candidates, terms, limit), nil
}

func (r *InMemoryBeliefRepository) SetPromptSlot(ctx context.Context, userID, beliefID, slot string) (*domain.Belief, error) {
	if !domain.IsValidPromptSlot(slot) {
		return nil, domain.ErrInvalidPromptSlot
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	belief, ok := r.beliefs[beliefID]
	if !ok || belief.UserID != userID {
		return nil, domain.ErrBeliefNotFound
	}
	cp := cloneBelief(belief)
	cp.PromptSlot = slot
	if err := cp.Validate(); err != nil {
		return nil, err
	}
	r.beliefs[beliefID] = cloneBelief(cp)
	return cp, nil
}

func (r *InMemoryBeliefRepository) RetireByBatchID(ctx context.Context, userID, batchID string) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	count := 0
	for id, belief := range r.beliefs {
		if belief.UserID != userID || belief.BatchID != batchID || belief.Status == domain.StatusRetired {
			continue
		}
		cp := cloneBelief(belief)
		cp.Status = domain.StatusRetired
		r.beliefs[id] = cp
		count++
	}
	return count, nil
}

func cloneBelief(belief *domain.Belief) *domain.Belief {
	if belief == nil {
		return nil
	}
	cp := *belief
	cp.SearchTerms = cloneStrings(belief.SearchTerms)
	cp.Embedding = cloneFloat32s(belief.Embedding)
	cp.LastContradictedAt = cloneTime(belief.LastContradictedAt)
	return &cp
}
