package application

import (
	"context"
	"testing"

	"github.com/armandoalvarado/sofia-backend/internal/learning/domain"
	"github.com/armandoalvarado/sofia-backend/internal/learning/infrastructure"
)

func TestInferredBeliefStaysOutOfPromptUntilDecisionPromotesIt(t *testing.T) {
	ctx := context.Background()
	beliefs := infrastructure.NewInMemoryBeliefRepository()
	service := NewService(beliefs, infrastructure.NewInMemoryPromptVersionRepository(), infrastructure.NewInMemoryDailySummaryRepository())
	inferred, err := service.UpsertBeliefWithTrust(ctx, "user-1", "prefiere reuniones temprano", domain.CategorySchedule, domain.ScopeGlobal, "", domain.TrustInferred, 1, "batch-1")
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 100; i++ {
		inferred.Reinforce(domain.TrustInferred)
	}
	if err := beliefs.Update(ctx, inferred); err != nil {
		t.Fatal(err)
	}
	version, err := service.RecompilePromptBase(ctx, "user-1")
	if err != nil {
		t.Fatal(err)
	}
	if version != nil || inferred.Confidence > 0.5 {
		t.Fatalf("inferred belief reached prompt or exceeded ceiling: version=%+v confidence=%f", version, inferred.Confidence)
	}
	promoted, err := service.UpsertBelief(ctx, "user-1", "prefiere reuniones temprano", domain.CategorySchedule)
	if err != nil {
		t.Fatal(err)
	}
	if promoted.ID != inferred.ID || promoted.EffectiveTrustTier() != domain.TrustDecision || promoted.Confidence <= 0.5 {
		t.Fatalf("decision did not promote inferred belief: %+v", promoted)
	}
}
