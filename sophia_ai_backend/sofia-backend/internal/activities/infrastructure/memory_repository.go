package infrastructure

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/armandoalvarado/sofia-backend/internal/activities/domain"
)

type InMemoryActivityRepository struct {
	mu         sync.Mutex
	activities map[string]*domain.Activity
}

func NewInMemoryActivityRepository() *InMemoryActivityRepository {
	return &InMemoryActivityRepository{activities: make(map[string]*domain.Activity)}
}

func (r *InMemoryActivityRepository) Create(ctx context.Context, activity *domain.Activity) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	cp := cloneActivity(activity)
	r.activities[activity.ID] = cp
	return nil
}

func (r *InMemoryActivityRepository) FindByID(ctx context.Context, userID, activityID string) (*domain.Activity, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	activity, ok := r.activities[activityID]
	if !ok || activity.UserID != userID {
		return nil, nil
	}
	return cloneActivity(activity), nil
}

func (r *InMemoryActivityRepository) List(ctx context.Context, filter domain.ListFilter) ([]*domain.Activity, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	result := make([]*domain.Activity, 0)
	for _, activity := range r.activities {
		if !matchesFilter(activity, filter) {
			continue
		}
		result = append(result, cloneActivity(activity))
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.After(result[j].CreatedAt)
	})
	result = activitiesAfterCursor(result, filter.Cursor)

	if filter.Limit > 0 && len(result) > filter.Limit {
		result = result[:filter.Limit]
	}
	return result, nil
}

func activitiesAfterCursor(activities []*domain.Activity, cursor string) []*domain.Activity {
	if cursor == "" {
		return activities
	}
	for i, activity := range activities {
		if activity.ID == cursor {
			return activities[i+1:]
		}
	}
	return activities
}

func (r *InMemoryActivityRepository) Update(ctx context.Context, activity *domain.Activity) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.activities[activity.ID] = cloneActivity(activity)
	return nil
}

func (r *InMemoryActivityRepository) Archive(ctx context.Context, userID, activityID string) (*domain.Activity, error) {
	return r.change(ctx, userID, activityID, func(activity *domain.Activity) error {
		return activity.MarkArchived()
	})
}

func (r *InMemoryActivityRepository) Complete(ctx context.Context, userID, activityID string) (*domain.Activity, error) {
	return r.change(ctx, userID, activityID, func(activity *domain.Activity) error {
		return activity.MarkCompleted()
	})
}

func (r *InMemoryActivityRepository) Cancel(ctx context.Context, userID, activityID string) (*domain.Activity, error) {
	return r.change(ctx, userID, activityID, func(activity *domain.Activity) error {
		return activity.MarkCanceled()
	})
}

func (r *InMemoryActivityRepository) change(ctx context.Context, userID, activityID string, mutate func(*domain.Activity) error) (*domain.Activity, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	activity, ok := r.activities[activityID]
	if !ok || activity.UserID != userID {
		return nil, domain.ErrActivityNotFound
	}

	cp := cloneActivity(activity)
	if err := mutate(cp); err != nil {
		return nil, err
	}
	r.activities[activityID] = cloneActivity(cp)
	return cp, nil
}

func matchesFilter(activity *domain.Activity, filter domain.ListFilter) bool {
	if activity.UserID != filter.UserID {
		return false
	}
	if filter.Status != "" && activity.Status != filter.Status {
		return false
	}
	if filter.Type != "" && activity.Type != filter.Type {
		return false
	}
	if filter.Priority != "" && activity.Priority != filter.Priority {
		return false
	}
	if filter.Tag != "" && !hasTag(activity.Tags, filter.Tag) {
		return false
	}
	if filter.From != nil && !activityInRangeAfter(activity, *filter.From) {
		return false
	}
	if filter.To != nil && !activityInRangeBefore(activity, *filter.To) {
		return false
	}
	return true
}

func hasTag(tags []string, tag string) bool {
	for _, existing := range tags {
		if existing == tag {
			return true
		}
	}
	return false
}

func activityInRangeAfter(activity *domain.Activity, from time.Time) bool {
	activityTime := rangeTime(activity)
	return activityTime.Equal(from) || activityTime.After(from)
}

func activityInRangeBefore(activity *domain.Activity, to time.Time) bool {
	activityTime := rangeTime(activity)
	return activityTime.Equal(to) || activityTime.Before(to)
}

func rangeTime(activity *domain.Activity) time.Time {
	if activity.StartAt != nil {
		return *activity.StartAt
	}
	if activity.DueAt != nil {
		return *activity.DueAt
	}
	return activity.CreatedAt
}

func cloneActivity(activity *domain.Activity) *domain.Activity {
	if activity == nil {
		return nil
	}

	cp := *activity
	cp.Tags = append([]string(nil), activity.Tags...)
	cp.StartAt = cloneTime(activity.StartAt)
	cp.DueAt = cloneTime(activity.DueAt)
	cp.CompletedAt = cloneTime(activity.CompletedAt)
	cp.CanceledAt = cloneTime(activity.CanceledAt)
	cp.ArchivedAt = cloneTime(activity.ArchivedAt)
	cp.ReminderAt = cloneTime(activity.ReminderAt)
	return &cp
}

func cloneTime(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	cp := *value
	return &cp
}
