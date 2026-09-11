package application

import (
	"context"
	"testing"
	"time"

	"github.com/armandoalvarado/sofia-backend/internal/learning/domain"
	learninginfra "github.com/armandoalvarado/sofia-backend/internal/learning/infrastructure"
)

type countingContextRepository struct {
	domain.UserContextRepository
	lists int
}

func (r *countingContextRepository) List(ctx context.Context, userID string) ([]*domain.UserContext, error) {
	r.lists++
	return r.UserContextRepository.List(ctx, userID)
}

func TestEntityResolverMatchesWholeWordsOnly(t *testing.T) {
	svc := newLearningService()
	if _, err := svc.CreateUserContext(context.Background(), "user-1", domain.UserContextCreate{Kind: domain.ScopePerson, Slug: "diana", Label: "Diana"}); err != nil {
		t.Fatal(err)
	}
	matches, err := svc.ResolveEntities(context.Background(), "user-1", "ana dijo que vendría")
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 0 {
		t.Fatalf("ana activated Diana: %+v", matches)
	}
	message := "Hablé con Diana ayer"
	matches, err = svc.ResolveEntities(context.Background(), "user-1", message)
	if err != nil || len(matches) != 1 || message[matches[0].Start:matches[0].End] != "Diana" {
		t.Fatalf("resolver returned wrong source position: %+v err=%v", matches, err)
	}
}

func TestEntityResolverCachesCatalogPerUser(t *testing.T) {
	base := learninginfra.NewInMemoryUserContextRepository()
	value, err := domain.NewUserContext("diana", "user-1", domain.UserContextCreate{Kind: domain.ScopePerson, Slug: "diana", Label: "Diana"})
	if err != nil {
		t.Fatal(err)
	}
	if err := base.Create(context.Background(), value); err != nil {
		t.Fatal(err)
	}
	repo := &countingContextRepository{UserContextRepository: base}
	resolver := NewEntityResolver(repo, nil, time.Minute)
	for range 2 {
		if _, err := resolver.Resolve(context.Background(), "user-1", "Hablé con Diana"); err != nil {
			t.Fatal(err)
		}
	}
	if repo.lists != 1 {
		t.Fatalf("catalog reads=%d, want 1 within TTL", repo.lists)
	}
}

func TestEntityResolverMarksHomonymsAmbiguous(t *testing.T) {
	svc := newLearningService()
	for _, input := range []domain.UserContextCreate{
		{Kind: domain.ScopePerson, Slug: "diana-hermana", Label: "Diana", Relationship: "hermana"},
		{Kind: domain.ScopePerson, Slug: "diana-trabajo", Label: "Diana", Relationship: "compañera"},
	} {
		if _, err := svc.CreateUserContext(context.Background(), "user-1", input); err != nil {
			t.Fatal(err)
		}
	}
	matches, err := svc.ResolveEntities(context.Background(), "user-1", "Hablé con Diana ayer")
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 2 || !matches[0].Ambiguous || !matches[1].Ambiguous || matches[0].Start != matches[1].Start {
		t.Fatalf("homonyms were not returned as ambiguous candidates: %+v", matches)
	}
}

func TestCandidatePromotesOnThirdMention(t *testing.T) {
	svc := newLearningService()
	candidates := learninginfra.NewInMemoryEntityCandidateRepository()
	svc.SetEntityCandidateRepository(candidates, 3)
	ctx := context.Background()
	for range 2 {
		if _, err := svc.ResolveEntities(ctx, "user-1", "Hablé con Lucía ayer"); err != nil {
			t.Fatal(err)
		}
	}
	if promoted, err := svc.PromoteEntityCandidates(ctx, "user-1", time.Now().UTC()); err != nil || promoted != 0 {
		t.Fatalf("candidate with two mentions promoted=%d err=%v", promoted, err)
	}
	if _, err := svc.ResolveEntities(ctx, "user-1", "Volví a ver a Lucía hoy"); err != nil {
		t.Fatal(err)
	}
	if promoted, err := svc.PromoteEntityCandidates(ctx, "user-1", time.Now().UTC()); err != nil || promoted != 1 {
		t.Fatalf("candidate with three mentions promoted=%d err=%v", promoted, err)
	}
	values, err := svc.ListEntities(ctx, "user-1", domain.ScopePerson, domain.ContextStatusPendingReview)
	if err != nil || len(values) != 1 || values[0].Slug != "lucia" || values[0].EffectiveCreatedFrom() != domain.CreatedFromInferred {
		t.Fatalf("unexpected promoted entity: %+v err=%v", values, err)
	}
	if _, err := svc.UpsertEntityBelief(ctx, "user-1", values[0].ID, "Le gusta bailar", domain.CategoryPersonal, domain.FactKindTrait, nil, nil); err != nil {
		t.Fatalf("pending entity did not accept facts: %v", err)
	}
	if loaded, err := svc.GetEntityContext(ctx, "user-1", values[0].ID, 0); err != nil || len(loaded.Beliefs) != 1 {
		t.Fatalf("pending entity context was not loadable: %+v err=%v", loaded, err)
	}
}
