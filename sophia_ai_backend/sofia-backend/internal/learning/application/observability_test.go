package application

import (
	"context"
	"testing"

	"github.com/armandoalvarado/sofia-backend/internal/learning/domain"
	"github.com/armandoalvarado/sofia-backend/internal/learning/infrastructure"
)

func TestListBeliefsCombinesFiltersAndIsolatesUsers(t *testing.T) {
	ctx := context.Background()
	repo := infrastructure.NewInMemoryBeliefRepository()
	service := NewService(repo, infrastructure.NewInMemoryPromptVersionRepository(), infrastructure.NewInMemoryDailySummaryRepository())
	match, err := service.UpsertBeliefWithTrust(ctx, "user-1", "hablar formal", domain.CategoryCommunication, domain.ScopePerson, "person:cliente", domain.TrustStated, 0.7, "")
	if err != nil {
		t.Fatal(err)
	}
	match.PromptSlot = domain.PromptSlotSituational
	if err := repo.Update(ctx, match); err != nil {
		t.Fatal(err)
	}
	_, _ = service.UpsertBeliefWithTrust(ctx, "user-1", "hablar casual", domain.CategoryCommunication, domain.ScopePerson, "person:maria", domain.TrustStated, 0.7, "")
	_, _ = service.UpsertBeliefWithTrust(ctx, "user-2", "hablar formal", domain.CategoryCommunication, domain.ScopePerson, "person:cliente", domain.TrustStated, 0.7, "")
	page, err := service.ListBeliefs(ctx, "user-1", BeliefListFilter{Scope: domain.ScopePerson, ScopeKey: "person:cliente", Category: domain.CategoryCommunication, TrustTier: domain.TrustStated, PromptSlot: domain.PromptSlotSituational, Status: domain.StatusActive, Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Beliefs) != 1 || page.Beliefs[0].ID != match.ID {
		t.Fatalf("unexpected filtered beliefs: %+v", page.Beliefs)
	}
}

func TestRetiredBeliefLeavesNextPromptCompilation(t *testing.T) {
	ctx := context.Background()
	repo := infrastructure.NewInMemoryBeliefRepository()
	service := NewService(repo, infrastructure.NewInMemoryPromptVersionRepository(), infrastructure.NewInMemoryDailySummaryRepository())
	removed, _ := service.UpsertBelief(ctx, "user-1", "prefiere mañanas", domain.CategorySchedule)
	kept, _ := service.UpsertBelief(ctx, "user-1", "respuestas breves", domain.CategoryCommunication)
	if _, err := service.RecompilePromptBase(ctx, "user-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.RetireBelief(ctx, "user-1", removed.ID); err != nil {
		t.Fatal(err)
	}
	version, err := service.RecompilePromptBase(ctx, "user-1")
	if err != nil {
		t.Fatal(err)
	}
	if version == nil || len(version.BeliefIDs) != 1 || version.BeliefIDs[0] != kept.ID {
		t.Fatalf("retired belief remained in prompt: %+v", version)
	}
}

func TestRetiringOnlyBeliefClearsActivePrompt(t *testing.T) {
	ctx := context.Background()
	repo := infrastructure.NewInMemoryBeliefRepository()
	prompts := infrastructure.NewInMemoryPromptVersionRepository()
	service := NewService(repo, prompts, infrastructure.NewInMemoryDailySummaryRepository())
	belief, _ := service.UpsertBelief(ctx, "user-1", "prefiere mañanas", domain.CategorySchedule)
	if _, err := service.RecompilePromptBase(ctx, "user-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.RetireBelief(ctx, "user-1", belief.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := service.RecompilePromptBase(ctx, "user-1"); err != nil {
		t.Fatal(err)
	}
	active, err := service.GetActivePromptVersion(ctx, "user-1")
	if err != nil || active != nil {
		t.Fatalf("active prompt = %+v, err = %v; want nil", active, err)
	}
}
