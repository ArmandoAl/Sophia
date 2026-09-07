package infrastructure

import (
	"context"
	"errors"
	"sort"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/armandoalvarado/sofia-backend/internal/activities/domain"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const activitiesCollection = "activities"

type FirestoreActivityRepository struct {
	client  *firestore.Client
	timeout time.Duration
}

type firestoreActivity struct {
	ID             string     `firestore:"id"`
	UserID         string     `firestore:"user_id"`
	Type           string     `firestore:"type"`
	Title          string     `firestore:"title"`
	Description    string     `firestore:"description"`
	Status         string     `firestore:"status"`
	Priority       string     `firestore:"priority"`
	StartAt        *time.Time `firestore:"start_at,omitempty"`
	DueAt          *time.Time `firestore:"due_at,omitempty"`
	CompletedAt    *time.Time `firestore:"completed_at,omitempty"`
	CanceledAt     *time.Time `firestore:"canceled_at,omitempty"`
	ArchivedAt     *time.Time `firestore:"archived_at,omitempty"`
	Timezone       string     `firestore:"timezone"`
	Tags           []string   `firestore:"tags"`
	Source         string     `firestore:"source"`
	CreatedBy      string     `firestore:"created_by"`
	CreatedAt      time.Time  `firestore:"created_at"`
	UpdatedAt      time.Time  `firestore:"updated_at"`
	HasReminder    bool       `firestore:"has_reminder"`
	ReminderAt     *time.Time `firestore:"reminder_at,omitempty"`
	RecurrenceRule string     `firestore:"recurrence_rule"`
	AIManageable   bool       `firestore:"ai_manageable"`
	AINotes        string     `firestore:"ai_notes"`
}

func NewFirestoreActivityRepository(client *firestore.Client) *FirestoreActivityRepository {
	return &FirestoreActivityRepository{client: client, timeout: 5 * time.Second}
}

func (r *FirestoreActivityRepository) Create(ctx context.Context, activity *domain.Activity) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	_, err := r.client.Collection(activitiesCollection).Doc(activity.ID).Set(ctx, activityToDocument(activity))
	return err
}

func (r *FirestoreActivityRepository) FindByID(ctx context.Context, userID, activityID string) (*domain.Activity, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	doc, err := r.client.Collection(activitiesCollection).Doc(activityID).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, nil
		}
		return nil, err
	}

	activity, err := documentToActivity(doc)
	if err != nil {
		return nil, err
	}
	if activity.UserID != userID {
		return nil, nil
	}
	return activity, nil
}

func (r *FirestoreActivityRepository) List(ctx context.Context, filter domain.ListFilter) ([]*domain.Activity, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	query := r.client.Collection(activitiesCollection).Where("user_id", "==", filter.UserID)
	if filter.Status != "" {
		query = query.Where("status", "==", filter.Status)
	}
	if filter.Type != "" {
		query = query.Where("type", "==", filter.Type)
	}
	if filter.Priority != "" {
		query = query.Where("priority", "==", filter.Priority)
	}
	query = query.OrderBy("created_at", firestore.Desc).OrderBy("id", firestore.Desc)
	if filter.Cursor != "" {
		cursorDoc, err := r.client.Collection(activitiesCollection).Doc(filter.Cursor).Get(ctx)
		if err == nil {
			var cursor firestoreActivity
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

	iter := query.Documents(ctx)
	defer iter.Stop()

	result := make([]*domain.Activity, 0)
	for {
		doc, err := iter.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, err
		}
		activity, err := documentToActivity(doc)
		if err != nil {
			return nil, err
		}
		if !matchesFilter(activity, filter) {
			continue
		}
		result = append(result, activity)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.After(result[j].CreatedAt)
	})
	if filter.Limit > 0 && len(result) > filter.Limit {
		result = result[:filter.Limit]
	}
	return result, nil
}

func (r *FirestoreActivityRepository) Update(ctx context.Context, activity *domain.Activity) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	_, err := r.client.Collection(activitiesCollection).Doc(activity.ID).Set(ctx, activityToDocument(activity))
	return err
}

func (r *FirestoreActivityRepository) Archive(ctx context.Context, userID, activityID string) (*domain.Activity, error) {
	return r.change(ctx, userID, activityID, func(activity *domain.Activity) error {
		return activity.MarkArchived()
	})
}

func (r *FirestoreActivityRepository) Complete(ctx context.Context, userID, activityID string) (*domain.Activity, error) {
	return r.change(ctx, userID, activityID, func(activity *domain.Activity) error {
		return activity.MarkCompleted()
	})
}

func (r *FirestoreActivityRepository) Cancel(ctx context.Context, userID, activityID string) (*domain.Activity, error) {
	return r.change(ctx, userID, activityID, func(activity *domain.Activity) error {
		return activity.MarkCanceled()
	})
}

func (r *FirestoreActivityRepository) change(ctx context.Context, userID, activityID string, mutate func(*domain.Activity) error) (*domain.Activity, error) {
	activity, err := r.FindByID(ctx, userID, activityID)
	if err != nil {
		return nil, err
	}
	if activity == nil {
		return nil, domain.ErrActivityNotFound
	}
	if err := mutate(activity); err != nil {
		return nil, err
	}
	if err := r.Update(ctx, activity); err != nil {
		return nil, err
	}
	return activity, nil
}

func activityToDocument(activity *domain.Activity) firestoreActivity {
	return firestoreActivity{
		ID:             activity.ID,
		UserID:         activity.UserID,
		Type:           activity.Type,
		Title:          activity.Title,
		Description:    activity.Description,
		Status:         activity.Status,
		Priority:       activity.Priority,
		StartAt:        activity.StartAt,
		DueAt:          activity.DueAt,
		CompletedAt:    activity.CompletedAt,
		CanceledAt:     activity.CanceledAt,
		ArchivedAt:     activity.ArchivedAt,
		Timezone:       activity.Timezone,
		Tags:           append([]string(nil), activity.Tags...),
		Source:         activity.Source,
		CreatedBy:      activity.CreatedBy,
		CreatedAt:      activity.CreatedAt,
		UpdatedAt:      activity.UpdatedAt,
		HasReminder:    activity.HasReminder,
		ReminderAt:     activity.ReminderAt,
		RecurrenceRule: activity.RecurrenceRule,
		AIManageable:   activity.AIManageable,
		AINotes:        activity.AINotes,
	}
}

func documentToActivity(doc *firestore.DocumentSnapshot) (*domain.Activity, error) {
	var stored firestoreActivity
	if err := doc.DataTo(&stored); err != nil {
		return nil, err
	}
	return stored.toDomain(), nil
}

func (a firestoreActivity) toDomain() *domain.Activity {
	return &domain.Activity{
		ID:             a.ID,
		UserID:         a.UserID,
		Type:           a.Type,
		Title:          a.Title,
		Description:    a.Description,
		Status:         a.Status,
		Priority:       a.Priority,
		StartAt:        a.StartAt,
		DueAt:          a.DueAt,
		CompletedAt:    a.CompletedAt,
		CanceledAt:     a.CanceledAt,
		ArchivedAt:     a.ArchivedAt,
		Timezone:       a.Timezone,
		Tags:           append([]string(nil), a.Tags...),
		Source:         a.Source,
		CreatedBy:      a.CreatedBy,
		CreatedAt:      a.CreatedAt,
		UpdatedAt:      a.UpdatedAt,
		HasReminder:    a.HasReminder,
		ReminderAt:     a.ReminderAt,
		RecurrenceRule: a.RecurrenceRule,
		AIManageable:   a.AIManageable,
		AINotes:        a.AINotes,
	}
}
