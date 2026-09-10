package application

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"testing"

	"github.com/armandoalvarado/sofia-backend/internal/learning/domain"
	learninginfra "github.com/armandoalvarado/sofia-backend/internal/learning/infrastructure"
)

type fakeEmbeddingProvider struct {
	vectors map[string][]float32
	err     error
}

func (f fakeEmbeddingProvider) EmbedText(_ context.Context, text string) ([]float32, error) {
	if f.err != nil {
		return nil, f.err
	}
	return append([]float32(nil), f.vectors[text]...), nil
}

type fakeAuditRecorder struct {
	count int
}

func (f *fakeAuditRecorder) RecordAuditLog(context.Context, string, string, string, string, map[string]any) error {
	f.count++
	return nil
}

func TestUpsertBeliefDeduplicatesSemanticParaphrases(t *testing.T) {
	firstStatement := "Tengo una junta temprano"
	secondStatement := "Morning meeting works best"
	svc := newLearningService()
	svc.SetEmbeddingProvider(fakeEmbeddingProvider{vectors: map[string][]float32{
		firstStatement:  {1, 0},
		secondStatement: {0.99, 0.01},
	}}, 0.85)

	first, err := svc.UpsertBelief(context.Background(), "user-1", firstStatement, domain.CategorySchedule)
	if err != nil {
		t.Fatal(err)
	}
	second, err := svc.UpsertBelief(context.Background(), "user-1", secondStatement, domain.CategorySchedule)
	if err != nil {
		t.Fatal(err)
	}
	if first.ID != second.ID || second.EvidenceCount != 2 {
		t.Fatalf("expected one reinforced belief, got first=%s second=%s evidence=%d", first.ID, second.ID, second.EvidenceCount)
	}
}

func TestUpsertBeliefDoesNotDeduplicateAcrossCategories(t *testing.T) {
	firstStatement := "Tengo una junta temprano"
	secondStatement := "Morning meeting works best"
	svc := newLearningService()
	svc.SetEmbeddingProvider(fakeEmbeddingProvider{vectors: map[string][]float32{
		firstStatement:  {1, 0},
		secondStatement: {1, 0},
	}}, 0.85)

	first, err := svc.UpsertBelief(context.Background(), "user-1", firstStatement, domain.CategorySchedule)
	if err != nil {
		t.Fatal(err)
	}
	second, err := svc.UpsertBelief(context.Background(), "user-1", secondStatement, domain.CategoryCommunication)
	if err != nil {
		t.Fatal(err)
	}
	if first.ID == second.ID {
		t.Fatal("beliefs from different categories were deduplicated")
	}
}

func TestUpsertBeliefDegradesToLexicalDedupeWhenEmbeddingFails(t *testing.T) {
	svc := newLearningService()
	svc.SetEmbeddingProvider(fakeEmbeddingProvider{err: errors.New("provider unavailable")}, 0.85)
	audit := &fakeAuditRecorder{}
	svc.SetAuditRecorder(audit)

	first, err := svc.UpsertBelief(context.Background(), "user-1", "Prefiere reuniones por la mañana", domain.CategorySchedule)
	if err != nil {
		t.Fatal(err)
	}
	second, err := svc.UpsertBelief(context.Background(), "user-1", "Le gusta tener reuniones por la mañana", domain.CategorySchedule)
	if err != nil {
		t.Fatal(err)
	}
	if first.ID != second.ID || second.EvidenceCount != 2 {
		t.Fatalf("expected lexical dedupe after degradation, got first=%s second=%s evidence=%d", first.ID, second.ID, second.EvidenceCount)
	}
	if audit.count != 1 {
		t.Fatalf("expected one degradation audit log, got %d", audit.count)
	}
}

func TestUpsertBeliefAuditsDisabledAndEmptyEmbeddings(t *testing.T) {
	for _, test := range []struct {
		name     string
		provider *fakeEmbeddingProvider
	}{
		{name: "disabled"},
		{name: "empty", provider: &fakeEmbeddingProvider{vectors: map[string][]float32{}}},
	} {
		t.Run(test.name, func(t *testing.T) {
			svc := newLearningService()
			if test.provider != nil {
				svc.SetEmbeddingProvider(test.provider, 0.85)
			}
			audit := &fakeAuditRecorder{}
			svc.SetAuditRecorder(audit)
			if _, err := svc.UpsertBelief(context.Background(), "user-1", "Trabaja temprano", domain.CategoryWorkStyle); err != nil {
				t.Fatal(err)
			}
			if audit.count != 1 {
				t.Fatalf("expected one degradation audit log, got %d", audit.count)
			}
		})
	}
}

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

func TestScopedBeliefDoesNotEnterGlobalPromptBase(t *testing.T) {
	svc := newLearningService()
	ctx := context.Background()
	if _, err := svc.UpsertBelief(ctx, "user-1", "Prefiere respuestas breves", domain.CategoryCommunication); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.UpsertBelief(ctx, "user-1", "Con María usa un tono cariñoso", domain.CategoryCommunication, domain.ScopePerson, "person:maria"); err != nil {
		t.Fatal(err)
	}
	version, err := svc.RecompilePromptBase(ctx, "user-1")
	if err != nil {
		t.Fatal(err)
	}
	if version == nil || !strings.Contains(version.Content, "respuestas breves") {
		t.Fatalf("global belief missing from prompt base: %+v", version)
	}
	if strings.Contains(version.Content, "María") {
		t.Fatalf("scoped belief leaked into prompt base: %q", version.Content)
	}
}

func TestContextFragmentStaysWithinTokenCeiling(t *testing.T) {
	svc := newLearningService()
	ctx := context.Background()
	if _, err := svc.CreateUserContext(ctx, "user-1", domain.UserContextCreate{Kind: domain.ScopePerson, Slug: "maria", Label: "María"}); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 30; i++ {
		statement := strings.Repeat("detalle"+strconv.Itoa(i)+" ", 30)
		if _, err := svc.UpsertBelief(ctx, "user-1", statement, domain.CategoryPersonal, domain.ScopePerson, "person:maria"); err != nil {
			t.Fatal(err)
		}
	}
	statements, err := svc.ContextBeliefStatements(ctx, "user-1", "person:maria")
	if err != nil {
		t.Fatal(err)
	}
	if tokens := domain.ApproximateTokens(strings.Join(statements, "\n")); tokens > domain.ContextFragmentTokenCeiling {
		t.Fatalf("context fragment tokens = %d, ceiling = %d", tokens, domain.ContextFragmentTokenCeiling)
	}
}

func TestUpsertBeliefDoesNotDeduplicateAcrossScopes(t *testing.T) {
	svc := newLearningService()
	ctx := context.Background()
	global, err := svc.UpsertBelief(ctx, "user-1", "Le hablo formal", domain.CategoryCommunication)
	if err != nil {
		t.Fatal(err)
	}
	person, err := svc.UpsertBelief(ctx, "user-1", "Le hablo formal", domain.CategoryCommunication, domain.ScopePerson, "person:maria")
	if err != nil {
		t.Fatal(err)
	}
	if global.ID == person.ID {
		t.Fatal("beliefs from different scopes were deduplicated")
	}
}

func TestUserContextCRUD(t *testing.T) {
	svc := newLearningService()
	ctx := context.Background()
	value, err := svc.CreateUserContext(ctx, "user-1", domain.UserContextCreate{Kind: domain.ScopePerson, Slug: "maria", Label: "María", Aliases: []string{"Mari"}})
	if err != nil {
		t.Fatal(err)
	}
	listed, err := svc.ListUserContexts(ctx, "user-1")
	if err != nil || len(listed) != 1 {
		t.Fatalf("unexpected contexts: %+v err=%v", listed, err)
	}
	label := "María personal"
	updated, err := svc.UpdateUserContext(ctx, "user-1", value.ID, domain.UserContextUpdate{Label: &label})
	if err != nil || updated.Label != label {
		t.Fatalf("unexpected update: %+v err=%v", updated, err)
	}
	archived, err := svc.ArchiveUserContext(ctx, "user-1", value.ID)
	if err != nil || archived.Active {
		t.Fatalf("unexpected archive: %+v err=%v", archived, err)
	}
	resolved, err := svc.FindUserContextByScopeKey(ctx, "user-1", "person:maria")
	if err != nil || resolved != nil {
		t.Fatalf("archived context should not resolve: %+v err=%v", resolved, err)
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
	if _, err := svc.CreateDailySummary(ctx, "user-1", domain.DailySummaryCreate{Date: "2026-09-07", Observations: []string{"overwrite"}}); err != domain.ErrDailySummaryAlreadyExists {
		t.Fatalf("expected ErrDailySummaryAlreadyExists, got %v", err)
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
	service := NewService(
		learninginfra.NewInMemoryBeliefRepository(),
		learninginfra.NewInMemoryPromptVersionRepository(),
		learninginfra.NewInMemoryDailySummaryRepository(),
	)
	service.SetContextRepository(learninginfra.NewInMemoryUserContextRepository())
	return service
}
