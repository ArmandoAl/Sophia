// Package application keeps episode extraction separate from belief synthesis.
// This path may read raw conversations, but it only creates episodes: the synthesis
// worker still never reads raw conversations, and episodes never enter the base prompt.
package application

import (
	"context"
	"encoding/json"
	"sort"
	"strings"
	"time"

	runtimedomain "github.com/armandoalvarado/sofia-backend/internal/ai/runtime/domain"
	"github.com/armandoalvarado/sofia-backend/internal/learning/domain"
	memorydomain "github.com/armandoalvarado/sofia-backend/internal/memory/domain"
	"github.com/armandoalvarado/sofia-backend/internal/platform/jsonschema"
	"github.com/google/uuid"
)

const episodeOutputSchema = `{
  "type":"object","additionalProperties":false,"required":["episodes"],
  "properties":{"episodes":{"type":"array","items":{"type":"object","additionalProperties":false,
    "required":["summary","entity_ids","topics","salience"],
    "properties":{"summary":{"type":"string"},"entity_ids":{"type":"array","items":{"type":"string"}},
    "topics":{"type":"array","items":{"type":"string"}},"salience":{"type":"number"},
    "follow_up_at":{"type":"string"}}}}}
}`

type extractedEpisodes struct {
	Episodes []struct {
		Summary    string   `json:"summary"`
		EntityIDs  []string `json:"entity_ids"`
		Topics     []string `json:"topics"`
		Salience   float64  `json:"salience"`
		FollowUpAt *string  `json:"follow_up_at"`
	} `json:"episodes"`
}

type OpenThread struct {
	ID        string
	Source    string
	EntityIDs []string
	Summary   string
	DueAt     time.Time
	belief    *domain.Belief
	episode   *domain.Episode
}

func (s *Service) SetEpisodeStore(repo domain.EpisodeRepository, model runtimedomain.ModelClient, minSalience float64, dailyLimit int) {
	s.episodes, s.episodeModel = repo, model
	if minSalience <= 0 {
		minSalience = domain.EpisodeMinSalience
	}
	if dailyLimit <= 0 {
		dailyLimit = domain.DefaultEpisodeDailyLimit
	}
	s.episodeMinSalience, s.episodeDailyLimit = minSalience, dailyLimit
}

func (s *Service) ExtractEpisodes(ctx context.Context, userID, conversationID, source string, occurredAt time.Time, turns []runtimedomain.Turn, entityIDs []string) ([]*domain.Episode, error) {
	if len(entityIDs) == 0 || s.episodes == nil || s.episodeModel == nil {
		return nil, nil
	}
	existing, err := s.episodes.ListByConversation(ctx, userID, conversationID)
	if err != nil || len(existing) > 0 {
		return existing, err
	}
	if occurredAt.IsZero() {
		occurredAt = time.Now().UTC()
	}
	day := time.Date(occurredAt.Year(), occurredAt.Month(), occurredAt.Day(), 0, 0, 0, 0, occurredAt.Location())
	count, err := s.episodes.CountBetween(ctx, userID, day, day.AddDate(0, 0, 1))
	if err != nil || count >= s.episodeDailyLimit {
		return nil, err
	}
	payload, err := json.Marshal(map[string]any{
		"instructions":  "Extract only concrete memorable episodes from this conversation. Use 1-2 sentence summaries. Set follow_up_at only for an unresolved decision or future event.",
		"output_schema": json.RawMessage(episodeOutputSchema), "allowed_entity_ids": entityIDs, "conversation": turns,
	})
	if err != nil {
		return nil, err
	}
	response, err := s.episodeModel.Generate(ctx, runtimedomain.ModelRequest{UserID: userID, Message: string(payload), Task: runtimedomain.TaskExtract})
	if err != nil {
		return nil, err
	}
	if err := jsonschema.Validate(json.RawMessage(episodeOutputSchema), json.RawMessage(response.AssistantMessage)); err != nil {
		return nil, err
	}
	var output extractedEpisodes
	if err := json.Unmarshal([]byte(response.AssistantMessage), &output); err != nil {
		return nil, err
	}
	allowed := make(map[string]bool, len(entityIDs))
	for _, id := range entityIDs {
		allowed[id] = true
	}
	remaining := s.episodeDailyLimit - count
	created := make([]*domain.Episode, 0, min(remaining, len(output.Episodes)))
	for _, value := range output.Episodes {
		if len(created) >= remaining || value.Salience < s.episodeMinSalience {
			continue
		}
		ids := make([]string, 0, len(value.EntityIDs))
		for _, id := range value.EntityIDs {
			if allowed[id] {
				ids = append(ids, id)
			}
		}
		if len(ids) == 0 {
			continue
		}
		episode := &domain.Episode{
			ID: uuid.NewString(), UserID: userID, OccurredAt: occurredAt.UTC(), Summary: strings.TrimSpace(value.Summary),
			EntityIDs: ids, Topics: value.Topics, Salience: value.Salience, Source: source,
			ConversationID: conversationID, Status: domain.EpisodeStatusActive,
		}
		if episode.Source == "" {
			episode.Source = domain.EpisodeSourceConversation
		}
		if value.FollowUpAt != nil {
			parsed, err := time.Parse(time.RFC3339, *value.FollowUpAt)
			if err != nil {
				return nil, domain.ErrInvalidEpisode
			}
			episode.FollowUpAt = &parsed
		}
		episode.SearchTerms = memorydomain.ExtractTerms(strings.Join(append([]string{episode.Summary}, episode.Topics...), " "), 40)
		episode.Embedding = s.embedStatement(ctx, userID, episode.Summary)
		if err := episode.Validate(); err != nil {
			return nil, err
		}
		if err := s.episodes.Create(ctx, episode); err != nil {
			return nil, err
		}
		created = append(created, episode)
	}
	return created, nil
}

func (s *Service) SearchEpisodes(ctx context.Context, userID, entityID string, terms []string, limit int) ([]*domain.Episode, error) {
	if s.episodes == nil {
		return nil, nil
	}
	if limit <= 0 || limit > domain.DefaultEpisodeSearchLimit {
		limit = domain.DefaultEpisodeSearchLimit
	}
	values, err := s.episodes.SearchByEntity(ctx, userID, entityID, limit*4)
	if err != nil {
		return nil, err
	}
	if text := strings.TrimSpace(strings.Join(terms, " ")); text != "" {
		if query := s.embedStatement(ctx, userID, text); len(query) > 0 {
			sort.SliceStable(values, func(i, j int) bool {
				return cosineSimilarity(values[i].Embedding, query) > cosineSimilarity(values[j].Embedding, query)
			})
		}
	}
	result, tokens := make([]*domain.Episode, 0, limit), 0
	for _, episode := range values {
		cost := domain.ApproximateTokens(episode.Summary)
		if len(result) >= limit || tokens+cost > domain.EpisodeContextTokenBudget {
			break
		}
		result, tokens = append(result, episode), tokens+cost
	}
	return result, nil
}

func (s *Service) GetOpenThreads(ctx context.Context, userID string, limit int) ([]OpenThread, error) {
	if limit <= 0 || limit > 3 {
		limit = 3
	}
	now := time.Now().UTC()
	beliefs, err := s.beliefs.ListDueFollowUps(ctx, userID, now, limit*4)
	if err != nil {
		return nil, err
	}
	var episodes []*domain.Episode
	if s.episodes != nil {
		episodes, err = s.episodes.ListDueFollowUps(ctx, userID, now, limit*4)
		if err != nil {
			return nil, err
		}
	}
	threads := make([]OpenThread, 0, len(beliefs)+len(episodes))
	for _, belief := range beliefs {
		if belief.FactKind == domain.FactKindState && (belief.ValidUntil == nil || !belief.ValidUntil.After(now)) {
			continue
		}
		muted, err := s.entityThreadsMuted(ctx, userID, []string{belief.SubjectID})
		if err != nil {
			return nil, err
		}
		if belief.Sensitive || muted {
			continue
		}
		threads = append(threads, OpenThread{ID: belief.ID, Source: "belief", EntityIDs: []string{belief.SubjectID}, Summary: belief.Statement, DueAt: *belief.FollowUpAt, belief: belief})
	}
	for _, episode := range episodes {
		muted, err := s.entityThreadsMuted(ctx, userID, episode.EntityIDs)
		if err != nil {
			return nil, err
		}
		if muted {
			continue
		}
		threads = append(threads, OpenThread{ID: episode.ID, Source: "episode", EntityIDs: episode.EntityIDs, Summary: episode.Summary, DueAt: *episode.FollowUpAt, episode: episode})
	}
	sort.SliceStable(threads, func(i, j int) bool { return threads[i].DueAt.Before(threads[j].DueAt) })
	result, tokens := make([]OpenThread, 0, limit), 0
	for _, thread := range threads {
		cost := domain.ApproximateTokens(thread.Summary)
		if len(result) >= limit || tokens+cost > domain.OpenThreadTokenBudget {
			break
		}
		result, tokens = append(result, thread), tokens+cost
	}
	return result, nil
}

func (s *Service) MarkOpenThreadRetaken(ctx context.Context, thread OpenThread) error {
	now := time.Now().UTC()
	if thread.belief != nil {
		thread.belief.FollowedUpAt = &now
		return s.beliefs.Update(ctx, thread.belief)
	}
	if thread.episode != nil && s.episodes != nil {
		thread.episode.FollowedUpAt = &now
		return s.episodes.Update(ctx, thread.episode)
	}
	return nil
}

func (s *Service) ArchiveStaleEpisodes(ctx context.Context, now time.Time) (int, error) {
	if s.episodes == nil {
		return 0, nil
	}
	return s.episodes.ArchiveLowSalienceBefore(ctx, now.Add(-domain.EpisodeRetention), domain.EpisodeHighSalience)
}

func (s *Service) entityThreadsMuted(ctx context.Context, userID string, ids []string) (bool, error) {
	for _, id := range ids {
		if id == "" || s.contexts == nil {
			continue
		}
		entity, err := s.contexts.FindByID(ctx, userID, id)
		if err != nil {
			return false, err
		}
		if entity != nil && (!entity.CanLoadContext() || entity.ThreadsMuted) {
			return true, nil
		}
	}
	return false, nil
}
