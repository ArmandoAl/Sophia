package infrastructure

import (
	"context"
	"sort"
	"sync"

	"github.com/armandoalvarado/sofia-backend/internal/learning/domain"
)

var _ domain.DailySummaryRepository = (*InMemoryDailySummaryRepository)(nil)

type InMemoryDailySummaryRepository struct {
	mu        sync.Mutex
	summaries map[string]*domain.DailySummary
}

func NewInMemoryDailySummaryRepository() *InMemoryDailySummaryRepository {
	return &InMemoryDailySummaryRepository{summaries: make(map[string]*domain.DailySummary)}
}

func (r *InMemoryDailySummaryRepository) Create(ctx context.Context, summary *domain.DailySummary) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.summaries[summary.ID]; exists {
		return domain.ErrDailySummaryExists
	}
	for _, existing := range r.summaries {
		if existing.UserID == summary.UserID && existing.Date == summary.Date {
			return domain.ErrDailySummaryExists
		}
	}
	r.summaries[summary.ID] = cloneDailySummary(summary)
	return nil
}

func (r *InMemoryDailySummaryRepository) FindByDate(ctx context.Context, userID, date string) (*domain.DailySummary, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if summary, ok := r.summaries[domain.DailySummaryID(userID, date)]; ok && summary.UserID == userID {
		return cloneDailySummary(summary), nil
	}
	for _, summary := range r.summaries {
		if summary.UserID == userID && summary.Date == date {
			return cloneDailySummary(summary), nil
		}
	}
	return nil, nil
}

func (r *InMemoryDailySummaryRepository) ListRecent(ctx context.Context, userID string, limit int) ([]*domain.DailySummary, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	result := make([]*domain.DailySummary, 0)
	for _, summary := range r.summaries {
		if summary.UserID != userID {
			continue
		}
		result = append(result, cloneDailySummary(summary))
	}
	sort.SliceStable(result, func(i, j int) bool {
		return result[i].Date > result[j].Date
	})
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

func cloneDailySummary(summary *domain.DailySummary) *domain.DailySummary {
	if summary == nil {
		return nil
	}
	cp := *summary
	cp.Observations = cloneStrings(summary.Observations)
	cp.DeltaVsPrevious = domain.DailyDelta{
		Reinforced:   cloneStrings(summary.DeltaVsPrevious.Reinforced),
		Contradicted: cloneStrings(summary.DeltaVsPrevious.Contradicted),
		Novel:        cloneStrings(summary.DeltaVsPrevious.Novel),
	}
	return &cp
}
