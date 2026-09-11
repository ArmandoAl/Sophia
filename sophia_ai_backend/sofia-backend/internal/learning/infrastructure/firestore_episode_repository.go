package infrastructure

import (
	"context"
	"errors"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/armandoalvarado/sofia-backend/internal/learning/domain"
	"google.golang.org/api/iterator"
)

const episodesCollection = "episodes"

var _ domain.EpisodeRepository = (*FirestoreEpisodeRepository)(nil)

type FirestoreEpisodeRepository struct {
	client  *firestore.Client
	timeout time.Duration
}

type firestoreEpisode struct {
	ID             string     `firestore:"id"`
	UserID         string     `firestore:"user_id"`
	OccurredAt     time.Time  `firestore:"occurred_at"`
	Summary        string     `firestore:"summary"`
	EntityIDs      []string   `firestore:"entity_ids"`
	Topics         []string   `firestore:"topics"`
	Salience       float64    `firestore:"salience"`
	Source         string     `firestore:"source"`
	ConversationID string     `firestore:"conversation_id"`
	SearchTerms    []string   `firestore:"search_terms"`
	Embedding      []float64  `firestore:"embedding"`
	FollowUpAt     *time.Time `firestore:"follow_up_at,omitempty"`
	FollowedUpAt   *time.Time `firestore:"followed_up_at,omitempty"`
	Status         string     `firestore:"status"`
}

func NewFirestoreEpisodeRepository(client *firestore.Client) *FirestoreEpisodeRepository {
	return &FirestoreEpisodeRepository{client: client, timeout: 5 * time.Second}
}

func (r *FirestoreEpisodeRepository) Create(ctx context.Context, episode *domain.Episode) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	_, err := r.client.Collection(episodesCollection).Doc(episode.ID).Create(ctx, episodeToDocument(episode))
	return err
}

func (r *FirestoreEpisodeRepository) Update(ctx context.Context, episode *domain.Episode) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	_, err := r.client.Collection(episodesCollection).Doc(episode.ID).Set(ctx, episodeToDocument(episode))
	return err
}

func (r *FirestoreEpisodeRepository) ListByConversation(ctx context.Context, userID, conversationID string) ([]*domain.Episode, error) {
	return r.collect(ctx, r.client.Collection(episodesCollection).Where("user_id", "==", userID).Where("conversation_id", "==", conversationID))
}

func (r *FirestoreEpisodeRepository) CountBetween(ctx context.Context, userID string, from, to time.Time) (int, error) {
	values, err := r.collect(ctx, r.client.Collection(episodesCollection).Where("user_id", "==", userID).Where("occurred_at", ">=", from).Where("occurred_at", "<", to))
	return len(values), err
}

func (r *FirestoreEpisodeRepository) SearchByEntity(ctx context.Context, userID, entityID string, limit int) ([]*domain.Episode, error) {
	query := r.client.Collection(episodesCollection).
		Where("user_id", "==", userID).
		Where("entity_ids", "array-contains", entityID).
		Where("status", "==", domain.EpisodeStatusActive).
		OrderBy("occurred_at", firestore.Desc).
		OrderBy("salience", firestore.Desc)
	if limit > 0 {
		query = query.Limit(limit)
	}
	return r.collect(ctx, query)
}

func (r *FirestoreEpisodeRepository) ListDueFollowUps(ctx context.Context, userID string, now time.Time, limit int) ([]*domain.Episode, error) {
	query := r.client.Collection(episodesCollection).
		Where("user_id", "==", userID).
		Where("status", "==", domain.EpisodeStatusActive).
		Where("follow_up_at", "<=", now).
		OrderBy("follow_up_at", firestore.Asc)
	values, err := r.collect(ctx, query)
	if err != nil {
		return nil, err
	}
	result := values[:0]
	for _, episode := range values {
		if episode.FollowedUpAt == nil {
			result = append(result, episode)
		}
	}
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

func (r *FirestoreEpisodeRepository) ArchiveLowSalienceBefore(ctx context.Context, before time.Time, highSalience float64) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	iter := r.client.Collection(episodesCollection).Where("occurred_at", "<", before).Documents(ctx)
	defer iter.Stop()
	count := 0
	for {
		doc, err := iter.Next()
		if errors.Is(err, iterator.Done) {
			return count, nil
		}
		if err != nil {
			return count, err
		}
		var stored firestoreEpisode
		if err := doc.DataTo(&stored); err != nil {
			return count, err
		}
		if stored.Status != domain.EpisodeStatusActive || stored.Salience >= highSalience {
			continue
		}
		if _, err := doc.Ref.Update(ctx, []firestore.Update{{Path: "status", Value: domain.EpisodeStatusArchived}}); err != nil {
			return count, err
		}
		count++
	}
}

func (r *FirestoreEpisodeRepository) ReassignEntity(ctx context.Context, userID, sourceEntityID, targetEntityID string) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	iter := r.client.Collection(episodesCollection).Where("user_id", "==", userID).Where("entity_ids", "array-contains", sourceEntityID).Documents(ctx)
	defer iter.Stop()
	for {
		doc, err := iter.Next()
		if errors.Is(err, iterator.Done) {
			return nil
		}
		if err != nil {
			return err
		}
		var stored firestoreEpisode
		if err := doc.DataTo(&stored); err != nil {
			return err
		}
		for index, entityID := range stored.EntityIDs {
			if entityID == sourceEntityID {
				stored.EntityIDs[index] = targetEntityID
			}
		}
		if _, err := doc.Ref.Update(ctx, []firestore.Update{{Path: "entity_ids", Value: stored.EntityIDs}}); err != nil {
			return err
		}
	}
}

func (r *FirestoreEpisodeRepository) collect(ctx context.Context, query firestore.Query) ([]*domain.Episode, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	iter := query.Documents(ctx)
	defer iter.Stop()
	result := make([]*domain.Episode, 0)
	for {
		doc, err := iter.Next()
		if errors.Is(err, iterator.Done) {
			return result, nil
		}
		if err != nil {
			return nil, err
		}
		var stored firestoreEpisode
		if err := doc.DataTo(&stored); err != nil {
			return nil, err
		}
		result = append(result, stored.toDomain())
	}
}

func episodeToDocument(value *domain.Episode) firestoreEpisode {
	return firestoreEpisode{
		ID: value.ID, UserID: value.UserID, OccurredAt: value.OccurredAt, Summary: value.Summary,
		EntityIDs: cloneStrings(value.EntityIDs), Topics: cloneStrings(value.Topics), Salience: value.Salience,
		Source: value.Source, ConversationID: value.ConversationID, SearchTerms: cloneStrings(value.SearchTerms),
		Embedding: float32sToFloat64s(value.Embedding), FollowUpAt: cloneTime(value.FollowUpAt),
		FollowedUpAt: cloneTime(value.FollowedUpAt), Status: value.Status,
	}
}

func (m firestoreEpisode) toDomain() *domain.Episode {
	return &domain.Episode{
		ID: m.ID, UserID: m.UserID, OccurredAt: m.OccurredAt, Summary: m.Summary,
		EntityIDs: cloneStrings(m.EntityIDs), Topics: cloneStrings(m.Topics), Salience: m.Salience,
		Source: m.Source, ConversationID: m.ConversationID, SearchTerms: cloneStrings(m.SearchTerms),
		Embedding: float64sToFloat32s(m.Embedding), FollowUpAt: cloneTime(m.FollowUpAt),
		FollowedUpAt: cloneTime(m.FollowedUpAt), Status: m.Status,
	}
}
