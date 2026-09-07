package infrastructure

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/armandoalvarado/sofia-backend/internal/reminders/domain"
)

type InMemoryReminderRepository struct {
	mu        sync.Mutex
	reminders map[string]*domain.Reminder
}

func NewInMemoryReminderRepository() *InMemoryReminderRepository {
	return &InMemoryReminderRepository{reminders: make(map[string]*domain.Reminder)}
}

func (r *InMemoryReminderRepository) Create(ctx context.Context, reminder *domain.Reminder) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.reminders[reminder.ID] = cloneReminder(reminder)
	return nil
}

func (r *InMemoryReminderRepository) FindByID(ctx context.Context, userID, reminderID string) (*domain.Reminder, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	reminder, ok := r.reminders[reminderID]
	if !ok || reminder.UserID != userID {
		return nil, nil
	}
	return cloneReminder(reminder), nil
}

func (r *InMemoryReminderRepository) List(ctx context.Context, filter domain.ListFilter) ([]*domain.Reminder, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	result := make([]*domain.Reminder, 0)
	for _, reminder := range r.reminders {
		if !matchesFilter(reminder, filter) {
			continue
		}
		result = append(result, cloneReminder(reminder))
	}
	sortReminders(result)
	result = remindersAfterCursor(result, filter.Cursor)
	if filter.Limit > 0 && len(result) > filter.Limit {
		result = result[:filter.Limit]
	}
	return result, nil
}

func remindersAfterCursor(reminders []*domain.Reminder, cursor string) []*domain.Reminder {
	if cursor == "" {
		return reminders
	}
	for i, reminder := range reminders {
		if reminder.ID == cursor {
			return reminders[i+1:]
		}
	}
	return reminders
}

func (r *InMemoryReminderRepository) Update(ctx context.Context, reminder *domain.Reminder) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.reminders[reminder.ID] = cloneReminder(reminder)
	return nil
}

func (r *InMemoryReminderRepository) Cancel(ctx context.Context, userID, reminderID string) (*domain.Reminder, error) {
	return r.change(ctx, userID, reminderID, func(reminder *domain.Reminder) error {
		return reminder.MarkCanceled()
	})
}

func (r *InMemoryReminderRepository) Archive(ctx context.Context, userID, reminderID string) (*domain.Reminder, error) {
	return r.change(ctx, userID, reminderID, func(reminder *domain.Reminder) error {
		return reminder.MarkArchived()
	})
}

func (r *InMemoryReminderRepository) ListDue(ctx context.Context, userID string, now time.Time, limit int) ([]*domain.Reminder, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	result := make([]*domain.Reminder, 0)
	for _, reminder := range r.reminders {
		if reminder.UserID != userID || !reminder.IsDue(now) {
			continue
		}
		result = append(result, cloneReminder(reminder))
	}
	sortReminders(result)
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

func (r *InMemoryReminderRepository) ListDueAll(ctx context.Context, now time.Time, limit int) ([]*domain.Reminder, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	result := make([]*domain.Reminder, 0)
	for _, reminder := range r.reminders {
		if !reminder.IsDue(now) {
			continue
		}
		result = append(result, cloneReminder(reminder))
	}
	sortReminders(result)
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

func (r *InMemoryReminderRepository) ClaimDue(ctx context.Context, userID, reminderID, workerID string, now, leaseUntil time.Time) (*domain.Reminder, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	reminder, ok := r.reminders[reminderID]
	if !ok || reminder.UserID != userID {
		return nil, domain.ErrReminderNotFound
	}
	cp := cloneReminder(reminder)
	if err := cp.Claim(workerID, now, leaseUntil); err != nil {
		return nil, err
	}
	r.reminders[reminderID] = cloneReminder(cp)
	return cp, nil
}

func (r *InMemoryReminderRepository) MarkDelivered(ctx context.Context, userID, reminderID string) (*domain.Reminder, error) {
	return r.change(ctx, userID, reminderID, func(reminder *domain.Reminder) error {
		return reminder.MarkDelivered()
	})
}

func (r *InMemoryReminderRepository) MarkFailed(ctx context.Context, userID, reminderID, reason string) (*domain.Reminder, error) {
	return r.change(ctx, userID, reminderID, func(reminder *domain.Reminder) error {
		return reminder.MarkFailed(reason)
	})
}

func (r *InMemoryReminderRepository) change(ctx context.Context, userID, reminderID string, mutate func(*domain.Reminder) error) (*domain.Reminder, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	reminder, ok := r.reminders[reminderID]
	if !ok || reminder.UserID != userID {
		return nil, domain.ErrReminderNotFound
	}
	cp := cloneReminder(reminder)
	if err := mutate(cp); err != nil {
		return nil, err
	}
	r.reminders[reminderID] = cloneReminder(cp)
	return cp, nil
}

func matchesFilter(reminder *domain.Reminder, filter domain.ListFilter) bool {
	if filter.UserID != "" && reminder.UserID != filter.UserID {
		return false
	}
	if filter.ActivityID != "" && reminder.ActivityID != filter.ActivityID {
		return false
	}
	if filter.Status != "" && reminder.Status != filter.Status {
		return false
	}
	if filter.From != nil && reminder.ScheduledAt.Before(*filter.From) {
		return false
	}
	if filter.To != nil && reminder.ScheduledAt.After(*filter.To) {
		return false
	}
	return true
}

func sortReminders(reminders []*domain.Reminder) {
	sort.Slice(reminders, func(i, j int) bool {
		return reminders[i].ScheduledAt.Before(reminders[j].ScheduledAt)
	})
}

func cloneReminder(reminder *domain.Reminder) *domain.Reminder {
	if reminder == nil {
		return nil
	}
	cp := *reminder
	cp.RecurrenceUntil = cloneTime(reminder.RecurrenceUntil)
	cp.LastRunAt = cloneTime(reminder.LastRunAt)
	cp.DeliveredAt = cloneTime(reminder.DeliveredAt)
	cp.FailedAt = cloneTime(reminder.FailedAt)
	cp.ProcessingUntil = cloneTime(reminder.ProcessingUntil)
	cp.LastAttemptAt = cloneTime(reminder.LastAttemptAt)
	return &cp
}

func cloneTime(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	cp := *value
	return &cp
}
