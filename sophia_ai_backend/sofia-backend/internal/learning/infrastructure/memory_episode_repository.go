package infrastructure

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/armandoalvarado/sofia-backend/internal/learning/domain"
)

var _ domain.EpisodeRepository = (*InMemoryEpisodeRepository)(nil)

type InMemoryEpisodeRepository struct {
	mu       sync.Mutex
	episodes map[string]*domain.Episode
}

func NewInMemoryEpisodeRepository() *InMemoryEpisodeRepository {
	return &InMemoryEpisodeRepository{episodes: make(map[string]*domain.Episode)}
}

func (r *InMemoryEpisodeRepository) Create(_ context.Context, episode *domain.Episode) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.episodes[episode.ID] = cloneEpisode(episode)
	return nil
}

func (r *InMemoryEpisodeRepository) Update(_ context.Context, episode *domain.Episode) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.episodes[episode.ID]; !ok {
		return domain.ErrInvalidEpisode
	}
	r.episodes[episode.ID] = cloneEpisode(episode)
	return nil
}

func (r *InMemoryEpisodeRepository) ListByConversation(_ context.Context, userID, conversationID string) ([]*domain.Episode, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	result := make([]*domain.Episode, 0)
	for _, episode := range r.episodes {
		if episode.UserID == userID && episode.ConversationID == conversationID {
			result = append(result, cloneEpisode(episode))
		}
	}
	return result, nil
}

func (r *InMemoryEpisodeRepository) CountBetween(_ context.Context, userID string, from, to time.Time) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	count := 0
	for _, episode := range r.episodes {
		if episode.UserID == userID && !episode.OccurredAt.Before(from) && episode.OccurredAt.Before(to) {
			count++
		}
	}
	return count, nil
}

func (r *InMemoryEpisodeRepository) SearchByEntity(_ context.Context, userID, entityID string, limit int) ([]*domain.Episode, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	result := make([]*domain.Episode, 0)
	for _, episode := range r.episodes {
		if episode.UserID == userID && episode.Status == domain.EpisodeStatusActive && containsString(episode.EntityIDs, entityID) {
			result = append(result, cloneEpisode(episode))
		}
	}
	sortEpisodes(result)
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

func (r *InMemoryEpisodeRepository) ListDueFollowUps(_ context.Context, userID string, now time.Time, limit int) ([]*domain.Episode, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	result := make([]*domain.Episode, 0)
	for _, episode := range r.episodes {
		if episode.UserID == userID && episode.Status == domain.EpisodeStatusActive && episode.FollowUpAt != nil && !episode.FollowUpAt.After(now) && episode.FollowedUpAt == nil {
			result = append(result, cloneEpisode(episode))
		}
	}
	sort.SliceStable(result, func(i, j int) bool { return result[i].FollowUpAt.Before(*result[j].FollowUpAt) })
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

func (r *InMemoryEpisodeRepository) ArchiveLowSalienceBefore(_ context.Context, before time.Time, highSalience float64) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	count := 0
	for id, episode := range r.episodes {
		if episode.Status == domain.EpisodeStatusActive && episode.Salience < highSalience && episode.OccurredAt.Before(before) {
			cp := cloneEpisode(episode)
			cp.Status = domain.EpisodeStatusArchived
			r.episodes[id] = cp
			count++
		}
	}
	return count, nil
}

func (r *InMemoryEpisodeRepository) ReassignEntity(_ context.Context, userID, sourceEntityID, targetEntityID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for id, episode := range r.episodes {
		if episode.UserID != userID || !containsString(episode.EntityIDs, sourceEntityID) {
			continue
		}
		cp := cloneEpisode(episode)
		for index, entityID := range cp.EntityIDs {
			if entityID == sourceEntityID {
				cp.EntityIDs[index] = targetEntityID
			}
		}
		r.episodes[id] = cp
	}
	return nil
}

func sortEpisodes(values []*domain.Episode) {
	sort.SliceStable(values, func(i, j int) bool {
		if values[i].OccurredAt.Equal(values[j].OccurredAt) {
			return values[i].Salience > values[j].Salience
		}
		return values[i].OccurredAt.After(values[j].OccurredAt)
	})
}

func cloneEpisode(value *domain.Episode) *domain.Episode {
	cp := *value
	cp.EntityIDs = cloneStrings(value.EntityIDs)
	cp.Topics = cloneStrings(value.Topics)
	cp.SearchTerms = cloneStrings(value.SearchTerms)
	cp.Embedding = cloneFloat32s(value.Embedding)
	cp.FollowUpAt = cloneTime(value.FollowUpAt)
	cp.FollowedUpAt = cloneTime(value.FollowedUpAt)
	return &cp
}
