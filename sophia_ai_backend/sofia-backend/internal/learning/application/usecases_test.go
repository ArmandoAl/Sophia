package application

import (
	"context"
	"strconv"
	"strings"
	"testing"

	"github.com/armandoalvarado/sofia-backend/internal/learning/domain"
	learninginfra "github.com/armandoalvarado/sofia-backend/internal/learning/infrastructure"
)

func TestUpsertBeliefReinforcesParaphraseInsteadOfDuplicating(t *testing.T) {
	svc := newLearningService()
	ctx := context.Background()

	first, err := svc.UpsertBelief(ctx, "user-1", "Prefiere reuniones por la mañana", domain.CategorySchedule)
	if err != nil {
		t.Fatal(err)
	}
	second, err := svc.UpsertBelief(ctx, "user-1", "Le gusta tener reuniones por la mañana", domain.CategorySchedule)
	if err != nil {
		t.Fatal(err)
	}
	if first.ID != second.ID {
		t.Fatalf("expected paraphrase to reinforce existing belief, got %s vs %s", first.ID, second.ID)
	}
	if second.EvidenceCount != 2 {
		t.Fatalf("EvidenceCount = %d, want 2", second.EvidenceCount)
	}
	if second.Confidence <= first.Confidence {
		t.Fatalf("confidence should rise on reinforce: first=%v second=%v", first.Confidence, second.Confidence)
	}

	listed, err := svc.ListActiveBeliefs(ctx, "user-1", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(listed) != 1 {
		t.Fatalf("expected a single active belief, got %d", len(listed))
	}
}

func TestUpsertBeliefCreatesDistinctTopics(t *testing.T) {
	svc := newLearningService()
	ctx := context.Background()

	meetings, err := svc.UpsertBelief(ctx, "user-1", "Prefiere reuniones por la mañana", domain.CategorySchedule)
	if err != nil {
		t.Fatal(err)
	}
	coffee, err := svc.UpsertBelief(ctx, "user-1", "Prefiere café de grano por la tarde", domain.CategoryPersonal)
	if err != nil {
		t.Fatal(err)
	}
	if meetings.ID == coffee.ID {
		t.Fatal("different topics should not collapse into one belief")
	}
}

func TestCreateActivePromptVersionDeactivatesPrevious(t *testing.T) {
	svc := newLearningService()
	ctx := context.Background()

	first, err := svc.CreateActivePromptVersion(ctx, "user-1", domain.PromptVersionCreate{Content: "You know the user prefers morning meetings."})
	if err != nil {
		t.Fatal(err)
	}
	second, err := svc.CreateActivePromptVersion(ctx, "user-1", domain.PromptVersionCreate{Content: "Schedule deep work before 11am."})
	if err != nil {
		t.Fatal(err)
	}
	if !second.Active || second.Version != 2 {
		t.Fatalf("expected version 2 active, got %+v", second)
	}
	listed, err := svc.ListPromptVersions(ctx, "user-1", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(listed) != 2 {
		t.Fatalf("expected both versions retained, got %d", len(listed))
	}
	active, err := svc.GetActivePromptVersion(ctx, "user-1")
	if err != nil {
		t.Fatal(err)
	}
	if active.ID != second.ID {
		t.Fatalf("active version = %s, want %s", active.ID, second.ID)
	}
	var inactiveCount int
	for _, version := range listed {
		if !version.Active {
			inactiveCount++
			if version.ID != first.ID {
				t.Fatalf("deactivated version = %s, want %s", version.ID, first.ID)
			}
		}
	}
	if inactiveCount != 1 {
		t.Fatalf("expected exactly one inactive version, got %d", inactiveCount)
	}
}

func TestRecompilePromptBaseNeverExceedsCeiling(t *testing.T) {
	svc := newLearningService()
	ctx := context.Background()
	for i := 0; i < 40; i++ {
		statement := strings.Repeat("pref"+strconv.Itoa(i)+" ", 40) + "tema" + strconv.Itoa(i)
		if _, err := svc.UpsertBelief(ctx, "user-1", statement, domain.CategorySchedule); err != nil {
			t.Fatal(err)
		}
	}
	version, err := svc.RecompilePromptBase(ctx, "user-1")
	if err != nil {
		t.Fatal(err)
	}
	if version == nil {
		t.Fatal("expected a compiled prompt version")
	}
	if version.TokenCount > domain.PromptTokenCeiling {
		t.Fatalf("TokenCount = %d, exceeds ceiling %d", version.TokenCount, domain.PromptTokenCeiling)
	}
	if domain.ApproximateTokens(version.Content) > domain.PromptTokenCeiling {
		t.Fatalf("content tokens = %d, exceeds ceiling", domain.ApproximateTokens(version.Content))
	}

	beliefs, err := svc.ListActiveBeliefs(ctx, "user-1", 100)
	if err != nil {
		t.Fatal(err)
	}
	var core, situational int
	for _, belief := range beliefs {
		switch belief.PromptSlot {
		case domain.PromptSlotCore:
			core++
		case domain.PromptSlotSituational:
			situational++
		default:
			t.Fatalf("belief %s missing prompt slot", belief.ID)
		}
	}
	if core == 0 || situational == 0 {
		t.Fatalf("expected both core and situational slots, core=%d situational=%d", core, situational)
	}
}

func TestDailySummaryIsAppendOnly(t *testing.T) {
	svc := newLearningService()
	ctx := context.Background()

	if _, err := svc.CreateDailySummary(ctx, "user-1", domain.DailySummaryCreate{Date: "2026-09-07", Observations: []string{"first"}}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CreateDailySummary(ctx, "user-1", domain.DailySummaryCreate{Date: "2026-09-07", Observations: []string{"overwrite"}}); err != domain.ErrDailySummaryExists {
		t.Fatalf("expected ErrDailySummaryExists, got %v", err)
	}
	summary, err := svc.FindDailySummary(ctx, "user-1", "2026-09-07")
	if err != nil {
		t.Fatal(err)
	}
	if len(summary.Observations) != 1 || summary.Observations[0] != "first" {
		t.Fatalf("append-only history was overwritten: %+v", summary.Observations)
	}
}

func newLearningService() *Service {
	return NewService(
		learninginfra.NewInMemoryBeliefRepository(),
		learninginfra.NewInMemoryPromptVersionRepository(),
		learninginfra.NewInMemoryDailySummaryRepository(),
	)
}
