package infrastructure

import (
	"context"
	"errors"
	"sort"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/armandoalvarado/sofia-backend/internal/insights/domain"
	"google.golang.org/api/iterator"
)

const (
	moodEntriesCollection      = "mood_entries"
	activityOutcomesCollection = "activity_outcomes"
	reflectionsCollection      = "reflections"
)

type FirestoreMoodEntryRepository struct {
	client  *firestore.Client
	timeout time.Duration
}

type FirestoreActivityOutcomeRepository struct {
	client  *firestore.Client
	timeout time.Duration
}

type FirestoreReflectionRepository struct {
	client  *firestore.Client
	timeout time.Duration
}

func NewFirestoreMoodEntryRepository(client *firestore.Client) *FirestoreMoodEntryRepository {
	return &FirestoreMoodEntryRepository{client: client, timeout: 5 * time.Second}
}

func NewFirestoreActivityOutcomeRepository(client *firestore.Client) *FirestoreActivityOutcomeRepository {
	return &FirestoreActivityOutcomeRepository{client: client, timeout: 5 * time.Second}
}

func NewFirestoreReflectionRepository(client *firestore.Client) *FirestoreReflectionRepository {
	return &FirestoreReflectionRepository{client: client, timeout: 5 * time.Second}
}

func (r *FirestoreMoodEntryRepository) Create(ctx context.Context, entry *domain.MoodEntry) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	_, err := r.client.Collection(moodEntriesCollection).Doc(entry.ID).Set(ctx, moodEntryToDocument(entry))
	return err
}

func (r *FirestoreMoodEntryRepository) List(ctx context.Context, filter domain.ListFilter) ([]*domain.MoodEntry, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	query := r.client.Collection(moodEntriesCollection).Where("user_id", "==", filter.UserID)
	if filter.ActivityID != "" {
		query = query.Where("activity_id", "==", filter.ActivityID)
	}
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}

	iter := query.Documents(ctx)
	defer iter.Stop()

	result := make([]*domain.MoodEntry, 0)
	for {
		doc, err := iter.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, err
		}
		var stored firestoreMoodEntry
		if err := doc.DataTo(&stored); err != nil {
			return nil, err
		}
		entry := stored.toDomain()
		if !matchesInsightFilter(entry.UserID, entry.ActivityID, entry.OccurredAt, filter) {
			continue
		}
		result = append(result, entry)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].OccurredAt.After(result[j].OccurredAt)
	})
	if filter.Limit > 0 && len(result) > filter.Limit {
		result = result[:filter.Limit]
	}
	return result, nil
}

func (r *FirestoreActivityOutcomeRepository) Create(ctx context.Context, outcome *domain.ActivityOutcome) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	_, err := r.client.Collection(activityOutcomesCollection).Doc(outcome.ID).Set(ctx, activityOutcomeToDocument(outcome))
	return err
}

func (r *FirestoreActivityOutcomeRepository) List(ctx context.Context, filter domain.ListFilter) ([]*domain.ActivityOutcome, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	query := r.client.Collection(activityOutcomesCollection).Where("user_id", "==", filter.UserID)
	if filter.ActivityID != "" {
		query = query.Where("activity_id", "==", filter.ActivityID)
	}
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}

	iter := query.Documents(ctx)
	defer iter.Stop()

	result := make([]*domain.ActivityOutcome, 0)
	for {
		doc, err := iter.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, err
		}
		var stored firestoreActivityOutcome
		if err := doc.DataTo(&stored); err != nil {
			return nil, err
		}
		outcome := stored.toDomain()
		if !matchesInsightFilter(outcome.UserID, outcome.ActivityID, outcome.CreatedAt, filter) {
			continue
		}
		result = append(result, outcome)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.After(result[j].CreatedAt)
	})
	if filter.Limit > 0 && len(result) > filter.Limit {
		result = result[:filter.Limit]
	}
	return result, nil
}

func (r *FirestoreReflectionRepository) Create(ctx context.Context, reflection *domain.Reflection) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	_, err := r.client.Collection(reflectionsCollection).Doc(reflection.ID).Set(ctx, reflectionToDocument(reflection))
	return err
}

func (r *FirestoreReflectionRepository) List(ctx context.Context, filter domain.ListFilter) ([]*domain.Reflection, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	query := r.client.Collection(reflectionsCollection).Where("user_id", "==", filter.UserID)
	if filter.ActivityID != "" {
		query = query.Where("activity_id", "==", filter.ActivityID)
	}
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}

	iter := query.Documents(ctx)
	defer iter.Stop()

	result := make([]*domain.Reflection, 0)
	for {
		doc, err := iter.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, err
		}
		var stored firestoreReflection
		if err := doc.DataTo(&stored); err != nil {
			return nil, err
		}
		reflection := stored.toDomain()
		if !matchesInsightFilter(reflection.UserID, reflection.ActivityID, reflection.OccurredAt, filter) {
			continue
		}
		result = append(result, reflection)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].OccurredAt.After(result[j].OccurredAt)
	})
	if filter.Limit > 0 && len(result) > filter.Limit {
		result = result[:filter.Limit]
	}
	return result, nil
}

type firestoreMoodEntry struct {
	ID          string    `firestore:"id"`
	UserID      string    `firestore:"user_id"`
	ActivityID  string    `firestore:"activity_id"`
	Mood        string    `firestore:"mood"`
	EnergyLevel int       `firestore:"energy_level"`
	StressLevel int       `firestore:"stress_level"`
	FocusLevel  int       `firestore:"focus_level"`
	Note        string    `firestore:"note"`
	OccurredAt  time.Time `firestore:"occurred_at"`
	Timezone    string    `firestore:"timezone"`
	CreatedAt   time.Time `firestore:"created_at"`
	UpdatedAt   time.Time `firestore:"updated_at"`
}

type firestoreActivityOutcome struct {
	ID                string    `firestore:"id"`
	UserID            string    `firestore:"user_id"`
	ActivityID        string    `firestore:"activity_id"`
	Result            string    `firestore:"result"`
	CompletionQuality int       `firestore:"completion_quality"`
	SatisfactionLevel int       `firestore:"satisfaction_level"`
	DifficultyLevel   int       `firestore:"difficulty_level"`
	Note              string    `firestore:"note"`
	CreatedAt         time.Time `firestore:"created_at"`
	UpdatedAt         time.Time `firestore:"updated_at"`
}

type firestoreReflection struct {
	ID         string    `firestore:"id"`
	UserID     string    `firestore:"user_id"`
	ActivityID string    `firestore:"activity_id"`
	Title      string    `firestore:"title"`
	Content    string    `firestore:"content"`
	Sentiment  string    `firestore:"sentiment"`
	Tags       []string  `firestore:"tags"`
	OccurredAt time.Time `firestore:"occurred_at"`
	CreatedAt  time.Time `firestore:"created_at"`
	UpdatedAt  time.Time `firestore:"updated_at"`
}

func moodEntryToDocument(entry *domain.MoodEntry) firestoreMoodEntry {
	return firestoreMoodEntry{
		ID:          entry.ID,
		UserID:      entry.UserID,
		ActivityID:  entry.ActivityID,
		Mood:        entry.Mood,
		EnergyLevel: entry.EnergyLevel,
		StressLevel: entry.StressLevel,
		FocusLevel:  entry.FocusLevel,
		Note:        entry.Note,
		OccurredAt:  entry.OccurredAt,
		Timezone:    entry.Timezone,
		CreatedAt:   entry.CreatedAt,
		UpdatedAt:   entry.UpdatedAt,
	}
}

func activityOutcomeToDocument(outcome *domain.ActivityOutcome) firestoreActivityOutcome {
	return firestoreActivityOutcome{
		ID:                outcome.ID,
		UserID:            outcome.UserID,
		ActivityID:        outcome.ActivityID,
		Result:            outcome.Result,
		CompletionQuality: outcome.CompletionQuality,
		SatisfactionLevel: outcome.SatisfactionLevel,
		DifficultyLevel:   outcome.DifficultyLevel,
		Note:              outcome.Note,
		CreatedAt:         outcome.CreatedAt,
		UpdatedAt:         outcome.UpdatedAt,
	}
}

func reflectionToDocument(reflection *domain.Reflection) firestoreReflection {
	return firestoreReflection{
		ID:         reflection.ID,
		UserID:     reflection.UserID,
		ActivityID: reflection.ActivityID,
		Title:      reflection.Title,
		Content:    reflection.Content,
		Sentiment:  reflection.Sentiment,
		Tags:       append([]string(nil), reflection.Tags...),
		OccurredAt: reflection.OccurredAt,
		CreatedAt:  reflection.CreatedAt,
		UpdatedAt:  reflection.UpdatedAt,
	}
}

func (m firestoreMoodEntry) toDomain() *domain.MoodEntry {
	return &domain.MoodEntry{
		ID:          m.ID,
		UserID:      m.UserID,
		ActivityID:  m.ActivityID,
		Mood:        m.Mood,
		EnergyLevel: m.EnergyLevel,
		StressLevel: m.StressLevel,
		FocusLevel:  m.FocusLevel,
		Note:        m.Note,
		OccurredAt:  m.OccurredAt,
		Timezone:    m.Timezone,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

func (o firestoreActivityOutcome) toDomain() *domain.ActivityOutcome {
	return &domain.ActivityOutcome{
		ID:                o.ID,
		UserID:            o.UserID,
		ActivityID:        o.ActivityID,
		Result:            o.Result,
		CompletionQuality: o.CompletionQuality,
		SatisfactionLevel: o.SatisfactionLevel,
		DifficultyLevel:   o.DifficultyLevel,
		Note:              o.Note,
		CreatedAt:         o.CreatedAt,
		UpdatedAt:         o.UpdatedAt,
	}
}

func (r firestoreReflection) toDomain() *domain.Reflection {
	return &domain.Reflection{
		ID:         r.ID,
		UserID:     r.UserID,
		ActivityID: r.ActivityID,
		Title:      r.Title,
		Content:    r.Content,
		Sentiment:  r.Sentiment,
		Tags:       append([]string(nil), r.Tags...),
		OccurredAt: r.OccurredAt,
		CreatedAt:  r.CreatedAt,
		UpdatedAt:  r.UpdatedAt,
	}
}
