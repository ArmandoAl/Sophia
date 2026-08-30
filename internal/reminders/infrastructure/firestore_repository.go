package infrastructure

import (
	"context"
	"errors"
	"sort"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/armandoalvarado/sofia-backend/internal/reminders/domain"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const remindersCollection = "reminders"

type FirestoreReminderRepository struct {
	client  *firestore.Client
	timeout time.Duration
}

type firestoreReminder struct {
	ID                    string     `firestore:"id"`
	UserID                string     `firestore:"user_id"`
	ActivityID            string     `firestore:"activity_id"`
	Title                 string     `firestore:"title"`
	Description           string     `firestore:"description"`
	Status                string     `firestore:"status"`
	ScheduledAt           time.Time  `firestore:"scheduled_at"`
	Timezone              string     `firestore:"timezone"`
	RecurrenceRule        string     `firestore:"recurrence_rule"`
	RecurrenceInterval    int        `firestore:"recurrence_interval"`
	RecurrenceUntil       *time.Time `firestore:"recurrence_until,omitempty"`
	RecurrenceCount       int        `firestore:"recurrence_count"`
	RecurrenceOccurrences int        `firestore:"recurrence_occurrences"`
	NextRunAt             time.Time  `firestore:"next_run_at"`
	LastRunAt             *time.Time `firestore:"last_run_at,omitempty"`
	DeliveredAt           *time.Time `firestore:"delivered_at,omitempty"`
	FailedAt              *time.Time `firestore:"failed_at,omitempty"`
	FailureReason         string     `firestore:"failure_reason"`
	RetryCount            int        `firestore:"retry_count"`
	MaxRetries            int        `firestore:"max_retries"`
	ProcessingBy          string     `firestore:"processing_by"`
	ProcessingUntil       *time.Time `firestore:"processing_until,omitempty"`
	LastAttemptAt         *time.Time `firestore:"last_attempt_at,omitempty"`
	Source                string     `firestore:"source"`
	CreatedBy             string     `firestore:"created_by"`
	CreatedAt             time.Time  `firestore:"created_at"`
	UpdatedAt             time.Time  `firestore:"updated_at"`
}

func NewFirestoreReminderRepository(client *firestore.Client) *FirestoreReminderRepository {
	return &FirestoreReminderRepository{client: client, timeout: 5 * time.Second}
}

func (r *FirestoreReminderRepository) Create(ctx context.Context, reminder *domain.Reminder) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	_, err := r.client.Collection(remindersCollection).Doc(reminder.ID).Set(ctx, reminderToDocument(reminder))
	return err
}

func (r *FirestoreReminderRepository) FindByID(ctx context.Context, userID, reminderID string) (*domain.Reminder, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	doc, err := r.client.Collection(remindersCollection).Doc(reminderID).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, nil
		}
		return nil, err
	}

	reminder, err := documentToReminder(doc)
	if err != nil {
		return nil, err
	}
	if reminder.UserID != userID {
		return nil, nil
	}
	return reminder, nil
}

func (r *FirestoreReminderRepository) List(ctx context.Context, filter domain.ListFilter) ([]*domain.Reminder, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	query := r.client.Collection(remindersCollection).Where("user_id", "==", filter.UserID)
	if filter.Status != "" {
		query = query.Where("status", "==", filter.Status)
	}
	if filter.ActivityID != "" {
		query = query.Where("activity_id", "==", filter.ActivityID)
	}
	query = query.OrderBy("created_at", firestore.Desc).OrderBy("id", firestore.Desc)
	if filter.Cursor != "" {
		cursorDoc, err := r.client.Collection(remindersCollection).Doc(filter.Cursor).Get(ctx)
		if err == nil {
			var cursor firestoreReminder
			if err := cursorDoc.DataTo(&cursor); err == nil && cursor.UserID == filter.UserID {
				query = query.StartAfter(cursor.CreatedAt, cursor.ID)
			}
		} else if status.Code(err) != codes.NotFound {
			return nil, err
		}
	}
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	return r.collect(ctx, query, filter)
}

func (r *FirestoreReminderRepository) Update(ctx context.Context, reminder *domain.Reminder) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	_, err := r.client.Collection(remindersCollection).Doc(reminder.ID).Set(ctx, reminderToDocument(reminder))
	return err
}

func (r *FirestoreReminderRepository) Cancel(ctx context.Context, userID, reminderID string) (*domain.Reminder, error) {
	return r.change(ctx, userID, reminderID, func(reminder *domain.Reminder) error {
		return reminder.MarkCanceled()
	})
}

func (r *FirestoreReminderRepository) Archive(ctx context.Context, userID, reminderID string) (*domain.Reminder, error) {
	return r.change(ctx, userID, reminderID, func(reminder *domain.Reminder) error {
		return reminder.MarkArchived()
	})
}

func (r *FirestoreReminderRepository) ListDue(ctx context.Context, userID string, now time.Time, limit int) ([]*domain.Reminder, error) {
	filter := domain.ListFilter{UserID: userID, Status: domain.StatusPending, Limit: limit}
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	query := r.client.Collection(remindersCollection).
		Where("user_id", "==", userID).
		Where("status", "==", domain.StatusPending).
		Where("next_run_at", "<=", now).
		OrderBy("next_run_at", firestore.Asc)
	if limit > 0 {
		query = query.Limit(limit)
	}
	return r.collect(ctx, query, filter)
}

func (r *FirestoreReminderRepository) ListDueAll(ctx context.Context, now time.Time, limit int) ([]*domain.Reminder, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	query := r.client.Collection(remindersCollection).
		Where("status", "==", domain.StatusPending).
		Where("next_run_at", "<=", now).
		OrderBy("next_run_at", firestore.Asc)
	if limit > 0 {
		query = query.Limit(limit)
	}
	return r.collect(ctx, query, domain.ListFilter{Status: domain.StatusPending, Limit: limit})
}

func (r *FirestoreReminderRepository) ClaimDue(ctx context.Context, userID, reminderID, workerID string, now, leaseUntil time.Time) (*domain.Reminder, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	docRef := r.client.Collection(remindersCollection).Doc(reminderID)
	var claimed *domain.Reminder
	err := r.client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		doc, err := tx.Get(docRef)
		if err != nil {
			if status.Code(err) == codes.NotFound {
				return domain.ErrReminderNotFound
			}
			return err
		}
		reminder, err := documentToReminder(doc)
		if err != nil {
			return err
		}
		if reminder.UserID != userID {
			return domain.ErrReminderNotFound
		}
		if err := reminder.Claim(workerID, now, leaseUntil); err != nil {
			return err
		}
		if err := tx.Set(docRef, reminderToDocument(reminder)); err != nil {
			return err
		}
		claimed = reminder
		return nil
	})
	if err != nil {
		return nil, err
	}
	return claimed, nil
}

func (r *FirestoreReminderRepository) MarkDelivered(ctx context.Context, userID, reminderID string) (*domain.Reminder, error) {
	return r.change(ctx, userID, reminderID, func(reminder *domain.Reminder) error {
		return reminder.MarkDelivered()
	})
}

func (r *FirestoreReminderRepository) MarkFailed(ctx context.Context, userID, reminderID, reason string) (*domain.Reminder, error) {
	return r.change(ctx, userID, reminderID, func(reminder *domain.Reminder) error {
		return reminder.MarkFailed(reason)
	})
}

func (r *FirestoreReminderRepository) change(ctx context.Context, userID, reminderID string, mutate func(*domain.Reminder) error) (*domain.Reminder, error) {
	reminder, err := r.FindByID(ctx, userID, reminderID)
	if err != nil {
		return nil, err
	}
	if reminder == nil {
		return nil, domain.ErrReminderNotFound
	}
	if err := mutate(reminder); err != nil {
		return nil, err
	}
	if err := r.Update(ctx, reminder); err != nil {
		return nil, err
	}
	return reminder, nil
}

func (r *FirestoreReminderRepository) collect(ctx context.Context, query firestore.Query, filter domain.ListFilter) ([]*domain.Reminder, error) {
	iter := query.Documents(ctx)
	defer iter.Stop()

	result := make([]*domain.Reminder, 0)
	for {
		doc, err := iter.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, err
		}
		reminder, err := documentToReminder(doc)
		if err != nil {
			return nil, err
		}
		if !matchesFilter(reminder, filter) {
			continue
		}
		result = append(result, reminder)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].ScheduledAt.Before(result[j].ScheduledAt)
	})
	if filter.Limit > 0 && len(result) > filter.Limit {
		result = result[:filter.Limit]
	}
	return result, nil
}

func reminderToDocument(reminder *domain.Reminder) firestoreReminder {
	return firestoreReminder{
		ID:                    reminder.ID,
		UserID:                reminder.UserID,
		ActivityID:            reminder.ActivityID,
		Title:                 reminder.Title,
		Description:           reminder.Description,
		Status:                reminder.Status,
		ScheduledAt:           reminder.ScheduledAt,
		Timezone:              reminder.Timezone,
		RecurrenceRule:        reminder.RecurrenceRule,
		RecurrenceInterval:    reminder.RecurrenceInterval,
		RecurrenceUntil:       reminder.RecurrenceUntil,
		RecurrenceCount:       reminder.RecurrenceCount,
		RecurrenceOccurrences: reminder.RecurrenceOccurrences,
		NextRunAt:             reminder.NextRunAt,
		LastRunAt:             reminder.LastRunAt,
		DeliveredAt:           reminder.DeliveredAt,
		FailedAt:              reminder.FailedAt,
		FailureReason:         reminder.FailureReason,
		RetryCount:            reminder.RetryCount,
		MaxRetries:            reminder.MaxRetries,
		ProcessingBy:          reminder.ProcessingBy,
		ProcessingUntil:       reminder.ProcessingUntil,
		LastAttemptAt:         reminder.LastAttemptAt,
		Source:                reminder.Source,
		CreatedBy:             reminder.CreatedBy,
		CreatedAt:             reminder.CreatedAt,
		UpdatedAt:             reminder.UpdatedAt,
	}
}

func documentToReminder(doc *firestore.DocumentSnapshot) (*domain.Reminder, error) {
	var stored firestoreReminder
	if err := doc.DataTo(&stored); err != nil {
		return nil, err
	}
	return stored.toDomain(), nil
}

func (r firestoreReminder) toDomain() *domain.Reminder {
	return &domain.Reminder{
		ID:                    r.ID,
		UserID:                r.UserID,
		ActivityID:            r.ActivityID,
		Title:                 r.Title,
		Description:           r.Description,
		Status:                r.Status,
		ScheduledAt:           r.ScheduledAt,
		Timezone:              r.Timezone,
		RecurrenceRule:        r.RecurrenceRule,
		RecurrenceInterval:    r.RecurrenceInterval,
		RecurrenceUntil:       r.RecurrenceUntil,
		RecurrenceCount:       r.RecurrenceCount,
		RecurrenceOccurrences: r.RecurrenceOccurrences,
		NextRunAt:             r.NextRunAt,
		LastRunAt:             r.LastRunAt,
		DeliveredAt:           r.DeliveredAt,
		FailedAt:              r.FailedAt,
		FailureReason:         r.FailureReason,
		RetryCount:            r.RetryCount,
		MaxRetries:            r.MaxRetries,
		ProcessingBy:          r.ProcessingBy,
		ProcessingUntil:       r.ProcessingUntil,
		LastAttemptAt:         r.LastAttemptAt,
		Source:                r.Source,
		CreatedBy:             r.CreatedBy,
		CreatedAt:             r.CreatedAt,
		UpdatedAt:             r.UpdatedAt,
	}
}
