package infrastructure

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/armandoalvarado/sofia-backend/internal/insights/domain"
)

type InMemoryMoodEntryRepository struct {
	mu      sync.Mutex
	entries map[string]*domain.MoodEntry
}

type InMemoryActivityOutcomeRepository struct {
	mu       sync.Mutex
	outcomes map[string]*domain.ActivityOutcome
}

type InMemoryReflectionRepository struct {
	mu          sync.Mutex
	reflections map[string]*domain.Reflection
}

func NewInMemoryMoodEntryRepository() *InMemoryMoodEntryRepository {
	return &InMemoryMoodEntryRepository{entries: make(map[string]*domain.MoodEntry)}
}

func NewInMemoryActivityOutcomeRepository() *InMemoryActivityOutcomeRepository {
	return &InMemoryActivityOutcomeRepository{outcomes: make(map[string]*domain.ActivityOutcome)}
}

func NewInMemoryReflectionRepository() *InMemoryReflectionRepository {
	return &InMemoryReflectionRepository{reflections: make(map[string]*domain.Reflection)}
}

func (r *InMemoryMoodEntryRepository) Create(ctx context.Context, entry *domain.MoodEntry) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.entries[entry.ID] = cloneMoodEntry(entry)
	return nil
}

func (r *InMemoryMoodEntryRepository) List(ctx context.Context, filter domain.ListFilter) ([]*domain.MoodEntry, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	result := make([]*domain.MoodEntry, 0)
	for _, entry := range r.entries {
		if !matchesInsightFilter(entry.UserID, entry.ActivityID, entry.OccurredAt, filter) {
			continue
		}
		result = append(result, cloneMoodEntry(entry))
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].OccurredAt.After(result[j].OccurredAt)
	})
	if filter.Limit > 0 && len(result) > filter.Limit {
		result = result[:filter.Limit]
	}
	return result, nil
}

func (r *InMemoryActivityOutcomeRepository) Create(ctx context.Context, outcome *domain.ActivityOutcome) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.outcomes[outcome.ID] = cloneActivityOutcome(outcome)
	return nil
}

func (r *InMemoryActivityOutcomeRepository) List(ctx context.Context, filter domain.ListFilter) ([]*domain.ActivityOutcome, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	result := make([]*domain.ActivityOutcome, 0)
	for _, outcome := range r.outcomes {
		if !matchesInsightFilter(outcome.UserID, outcome.ActivityID, outcome.CreatedAt, filter) {
			continue
		}
		result = append(result, cloneActivityOutcome(outcome))
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.After(result[j].CreatedAt)
	})
	if filter.Limit > 0 && len(result) > filter.Limit {
		result = result[:filter.Limit]
	}
	return result, nil
}

func (r *InMemoryReflectionRepository) Create(ctx context.Context, reflection *domain.Reflection) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.reflections[reflection.ID] = cloneReflection(reflection)
	return nil
}

func (r *InMemoryReflectionRepository) List(ctx context.Context, filter domain.ListFilter) ([]*domain.Reflection, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	result := make([]*domain.Reflection, 0)
	for _, reflection := range r.reflections {
		if !matchesInsightFilter(reflection.UserID, reflection.ActivityID, reflection.OccurredAt, filter) {
			continue
		}
		result = append(result, cloneReflection(reflection))
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].OccurredAt.After(result[j].OccurredAt)
	})
	if filter.Limit > 0 && len(result) > filter.Limit {
		result = result[:filter.Limit]
	}
	return result, nil
}

func matchesInsightFilter(userID, activityID string, occurredAt time.Time, filter domain.ListFilter) bool {
	if userID != filter.UserID {
		return false
	}
	if filter.ActivityID != "" && activityID != filter.ActivityID {
		return false
	}
	if filter.From != nil && occurredAt.Before(*filter.From) {
		return false
	}
	if filter.To != nil && occurredAt.After(*filter.To) {
		return false
	}
	return true
}

func cloneMoodEntry(entry *domain.MoodEntry) *domain.MoodEntry {
	if entry == nil {
		return nil
	}
	cp := *entry
	return &cp
}

func cloneActivityOutcome(outcome *domain.ActivityOutcome) *domain.ActivityOutcome {
	if outcome == nil {
		return nil
	}
	cp := *outcome
	return &cp
}

func cloneReflection(reflection *domain.Reflection) *domain.Reflection {
	if reflection == nil {
		return nil
	}
	cp := *reflection
	cp.Tags = append([]string(nil), reflection.Tags...)
	return &cp
}
