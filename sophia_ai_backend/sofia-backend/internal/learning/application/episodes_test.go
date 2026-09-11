package application

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	runtimedomain "github.com/armandoalvarado/sofia-backend/internal/ai/runtime/domain"
	"github.com/armandoalvarado/sofia-backend/internal/learning/domain"
	learninginfra "github.com/armandoalvarado/sofia-backend/internal/learning/infrastructure"
)

type episodeModel struct {
	body  string
	calls int
}

func (m *episodeModel) Generate(context.Context, runtimedomain.ModelRequest) (runtimedomain.ModelResponse, error) {
	m.calls++
	return runtimedomain.ModelResponse{AssistantMessage: m.body}, nil
}

func newEpisodeService(model *episodeModel) (*Service, *learninginfra.InMemoryEpisodeRepository) {
	service := newLearningService()
	repo := learninginfra.NewInMemoryEpisodeRepository()
	service.SetEpisodeStore(repo, model, domain.EpisodeMinSalience, 5)
	return service, repo
}

func TestEpisodeExtractionSkipsConversationsWithoutEntities(t *testing.T) {
	model := &episodeModel{body: `{"episodes":[]}`}
	service, _ := newEpisodeService(model)
	episodes, err := service.ExtractEpisodes(context.Background(), "user", "conversation", domain.EpisodeSourceConversation, time.Now(), []runtimedomain.Turn{{Role: "user", Content: "Hola"}}, nil)
	if err != nil || len(episodes) != 0 || model.calls != 0 {
		t.Fatalf("episodes=%v calls=%d err=%v", episodes, model.calls, err)
	}
}

func TestEpisodeExtractionRejectsLowSalience(t *testing.T) {
	model := &episodeModel{body: `{"episodes":[{"summary":"Diana comentó algo menor.","entity_ids":["diana"],"topics":[],"salience":0.39}]}`}
	service, repo := newEpisodeService(model)
	episodes, err := service.ExtractEpisodes(context.Background(), "user", "conversation", domain.EpisodeSourceConversation, time.Now(), nil, []string{"diana"})
	if err != nil || len(episodes) != 0 {
		t.Fatalf("episodes=%v err=%v", episodes, err)
	}
	stored, err := repo.SearchByEntity(context.Background(), "user", "diana", 5)
	if err != nil || len(stored) != 0 {
		t.Fatalf("low-salience episode persisted: %v err=%v", stored, err)
	}
}

func TestSearchEpisodesReturnsOnlyEntityAndFitsBudget(t *testing.T) {
	model := &episodeModel{}
	service, repo := newEpisodeService(model)
	now := time.Now().UTC()
	for index, entityID := range []string{"diana", "samira", "diana", "diana"} {
		summary := strings.Repeat(fmt.Sprintf("recuerdo%d ", index), 100)
		episode := &domain.Episode{ID: fmt.Sprintf("e%d", index), UserID: "user", OccurredAt: now.Add(-time.Duration(index) * time.Hour), Summary: summary, EntityIDs: []string{entityID}, Salience: 0.8, Source: domain.EpisodeSourceConversation, ConversationID: fmt.Sprintf("c%d", index), Status: domain.EpisodeStatusActive}
		if err := repo.Create(context.Background(), episode); err != nil {
			t.Fatal(err)
		}
	}
	values, err := service.SearchEpisodes(context.Background(), "user", "diana", nil, 5)
	if err != nil {
		t.Fatal(err)
	}
	tokens := 0
	for _, episode := range values {
		if episode.EntityIDs[0] != "diana" {
			t.Fatalf("foreign episode returned: %+v", episode)
		}
		tokens += domain.ApproximateTokens(episode.Summary)
	}
	if tokens > domain.EpisodeContextTokenBudget {
		t.Fatalf("episode context tokens=%d", tokens)
	}
}

func TestOpenThreadsRespectDueSensitiveAndRetakenRules(t *testing.T) {
	service, _ := newEpisodeService(&episodeModel{})
	entity, err := service.CreateUserContext(context.Background(), "user", domain.UserContextCreate{Kind: domain.ScopePerson, Slug: "diana", Label: "Diana"})
	if err != nil {
		t.Fatal(err)
	}
	now, valid := time.Now().UTC(), time.Now().UTC().Add(48*time.Hour)
	future, past := now.Add(time.Hour), now.Add(-time.Hour)
	if _, err := service.UpsertEntityBelief(context.Background(), "user", entity.ID, "Cambiará de trabajo", domain.CategoryPersonal, domain.FactKindState, &valid, &future); err != nil {
		t.Fatal(err)
	}
	if _, err := service.UpsertEntityBelief(context.Background(), "user", entity.ID, "Está enferma", domain.CategoryPersonal, domain.FactKindState, &valid, &past); err != nil {
		t.Fatal(err)
	}
	due, err := service.UpsertEntityBelief(context.Background(), "user", entity.ID, "Decidirá si acepta la oferta", domain.CategoryPersonal, domain.FactKindState, &valid, &past)
	if err != nil {
		t.Fatal(err)
	}
	threads, err := service.GetOpenThreads(context.Background(), "user", 3)
	if err != nil || len(threads) != 1 || threads[0].ID != due.ID {
		t.Fatalf("threads=%+v err=%v", threads, err)
	}
	if err := service.MarkOpenThreadRetaken(context.Background(), threads[0]); err != nil {
		t.Fatal(err)
	}
	threads, err = service.GetOpenThreads(context.Background(), "user", 3)
	if err != nil || len(threads) != 0 {
		t.Fatalf("retaken/future/sensitive thread returned: %+v err=%v", threads, err)
	}
	if _, err := service.UpsertEntityBelief(context.Background(), "user", entity.ID, "Confirmará la mudanza", domain.CategoryPersonal, domain.FactKindState, &valid, &past); err != nil {
		t.Fatal(err)
	}
	muted := true
	if _, err := service.UpdateUserContext(context.Background(), "user", entity.ID, domain.UserContextUpdate{ThreadsMuted: &muted}); err != nil {
		t.Fatal(err)
	}
	threads, err = service.GetOpenThreads(context.Background(), "user", 3)
	if err != nil || len(threads) != 0 {
		t.Fatalf("muted entity thread returned: %+v err=%v", threads, err)
	}
}
