package worker

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	actionsdomain "github.com/armandoalvarado/sofia-backend/internal/ai/actions/domain"
	actionsinfra "github.com/armandoalvarado/sofia-backend/internal/ai/actions/infrastructure"
	runtimedomain "github.com/armandoalvarado/sofia-backend/internal/ai/runtime/domain"
	authdomain "github.com/armandoalvarado/sofia-backend/internal/auth/domain"
	authinfra "github.com/armandoalvarado/sofia-backend/internal/auth/infrastructure"
	learningapp "github.com/armandoalvarado/sofia-backend/internal/learning/application"
	learningdomain "github.com/armandoalvarado/sofia-backend/internal/learning/domain"
	learninginfra "github.com/armandoalvarado/sofia-backend/internal/learning/infrastructure"
	privacydomain "github.com/armandoalvarado/sofia-backend/internal/privacy/domain"
	privacyinfra "github.com/armandoalvarado/sofia-backend/internal/privacy/infrastructure"
	usersapp "github.com/armandoalvarado/sofia-backend/internal/users/application"
	usersdomain "github.com/armandoalvarado/sofia-backend/internal/users/domain"
	usersinfra "github.com/armandoalvarado/sofia-backend/internal/users/infrastructure"
	"github.com/google/uuid"
)

func TestRunDailySilentDayDoesNotCallModel(t *testing.T) {
	env := newWorkerTestEnv(t)
	day := time.Date(2026, 9, 6, 12, 0, 0, 0, time.UTC)
	if err := env.worker.RunDaily(context.Background(), env.userID, day); err != nil {
		t.Fatal(err)
	}
	if env.model.callCount() != 0 {
		t.Fatalf("silent day spent %d model calls, want 0", env.model.callCount())
	}
	if _, err := env.learning.FindDailySummary(context.Background(), env.userID, "2026-09-06"); err != learningdomain.ErrDailySummaryNotFound {
		t.Fatalf("silent day should not persist a summary, err=%v", err)
	}
}

func TestRunDailyReinforceKeepsBeliefInCore(t *testing.T) {
	env := newWorkerTestEnv(t)
	ctx := context.Background()
	day := time.Date(2026, 9, 6, 12, 0, 0, 0, time.UTC)
	belief, err := env.learning.UpsertBelief(ctx, env.userID, "Prefiere reuniones a las 9am", learningdomain.CategorySchedule)
	if err != nil {
		t.Fatal(err)
	}
	start := belief.Confidence
	env.mustProposal(t, "create_reminder", actionsdomain.FeedbackApprovedDirect, day)
	env.model.body = marshalSynthesis(t, synthesisOutput{
		Reinforced: []reinforcedBelief{{BeliefID: belief.ID, Evidence: "approved create_reminder without edits", ConfidenceDelta: 0.1}},
	})

	if err := env.worker.RunDaily(ctx, env.userID, day); err != nil {
		t.Fatal(err)
	}
	updated, err := env.learning.GetBelief(ctx, env.userID, belief.ID)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Confidence <= start {
		t.Fatalf("confidence did not rise: start=%v got=%v", start, updated.Confidence)
	}
	if updated.PromptSlot != learningdomain.PromptSlotCore {
		t.Fatalf("reinforced belief left core: slot=%q", updated.PromptSlot)
	}
	if env.model.callCount() != 1 {
		t.Fatalf("expected one model call, got %d", env.model.callCount())
	}
	assertDecisionsOnly(t, env.model.lastRequest())
}

func TestRunDailyRepeatedContradictionLeavesCore(t *testing.T) {
	env := newWorkerTestEnv(t)
	ctx := context.Background()
	day := time.Date(2026, 9, 6, 12, 0, 0, 0, time.UTC)

	target, err := env.learning.UpsertBelief(ctx, env.userID, "Cena siempre a las seis en punto", learningdomain.CategorySchedule)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 8; i++ {
		if _, err := env.learning.ContradictBelief(ctx, env.userID, target.ID); err != nil {
			t.Fatal(err)
		}
	}
	for i := 0; i < 30; i++ {
		statement := strings.Repeat("alfa"+strconv.Itoa(i)+" ", 40) + "nodo" + strconv.Itoa(i)
		if _, err := env.learning.UpsertBelief(ctx, env.userID, statement, learningdomain.CategoryWorkStyle); err != nil {
			t.Fatal(err)
		}
	}
	env.mustProposal(t, "create_reminder", actionsdomain.FeedbackApprovedCorrected, day)
	env.model.body = marshalSynthesis(t, synthesisOutput{
		Contradicted: []contradictedBelief{{BeliefID: target.ID, Evidence: "corrected reminder time", Note: "schedule changed"}},
	})

	if err := env.worker.RunDaily(ctx, env.userID, day); err != nil {
		t.Fatal(err)
	}
	updated, err := env.learning.GetBelief(ctx, env.userID, target.ID)
	if err != nil {
		t.Fatal(err)
	}
	if updated.PromptSlot != learningdomain.PromptSlotSituational {
		t.Fatalf("contradicted belief should leave core, slot=%q", updated.PromptSlot)
	}
	version, err := env.learning.GetActivePromptVersion(ctx, env.userID)
	if err != nil || version == nil {
		t.Fatalf("expected compiled prompt, err=%v", err)
	}
	if strings.Contains(version.Content, target.Statement) {
		t.Fatal("contradicted belief leaked into the rebuilt prompt base")
	}
}

func TestRunDailyRecompiledPromptNeverExceedsCeiling(t *testing.T) {
	env := newWorkerTestEnv(t)
	ctx := context.Background()
	day := time.Date(2026, 9, 6, 12, 0, 0, 0, time.UTC)
	for i := 0; i < 40; i++ {
		statement := strings.Repeat("beta"+strconv.Itoa(i)+" ", 40) + "hilo" + strconv.Itoa(i)
		if _, err := env.learning.UpsertBelief(ctx, env.userID, statement, learningdomain.CategoryPersonal); err != nil {
			t.Fatal(err)
		}
	}
	env.mustProposal(t, "create_activity", actionsdomain.FeedbackApprovedDirect, day)
	if err := env.worker.RunDaily(ctx, env.userID, day); err != nil {
		t.Fatal(err)
	}
	version, err := env.learning.GetActivePromptVersion(ctx, env.userID)
	if err != nil || version == nil {
		t.Fatalf("expected compiled prompt, err=%v", err)
	}
	if version.TokenCount > learningdomain.PromptTokenCeiling {
		t.Fatalf("TokenCount = %d exceeds ceiling %d", version.TokenCount, learningdomain.PromptTokenCeiling)
	}
	if learningdomain.ApproximateTokens(version.Content) > learningdomain.PromptTokenCeiling {
		t.Fatalf("content tokens exceed ceiling: %d", learningdomain.ApproximateTokens(version.Content))
	}
}

func TestRunDailySameDayIsIdempotent(t *testing.T) {
	env := newWorkerTestEnv(t)
	ctx := context.Background()
	day := time.Date(2026, 9, 6, 12, 0, 0, 0, time.UTC)
	env.mustProposal(t, "create_reminder", actionsdomain.FeedbackApprovedDirect, day)

	if err := env.worker.RunDaily(ctx, env.userID, day); err != nil {
		t.Fatal(err)
	}
	if err := env.worker.RunDaily(ctx, env.userID, day); err != nil {
		t.Fatal(err)
	}
	if env.model.callCount() != 1 {
		t.Fatalf("expected a single model call, got %d", env.model.callCount())
	}
	summaries, err := env.learning.ListRecentDailySummaries(ctx, env.userID, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(summaries) != 1 {
		t.Fatalf("expected one daily summary, got %d", len(summaries))
	}
}

func TestRunDailyUnknownNovelScopeFallsBackToGlobal(t *testing.T) {
	env := newWorkerTestEnv(t)
	ctx := context.Background()
	day := time.Date(2026, 9, 6, 12, 0, 0, 0, time.UTC)
	env.mustProposal(t, "create_activity", actionsdomain.FeedbackApprovedDirect, day)
	env.model.body = marshalSynthesis(t, synthesisOutput{Novel: []novelBelief{{
		Statement: "Hablar con cercanía", Category: learningdomain.CategoryCommunication, Confidence: 0.7,
		Scope: learningdomain.ScopePerson, ScopeKey: "person:unknown",
	}}})
	if err := env.worker.RunDaily(ctx, env.userID, day); err != nil {
		t.Fatal(err)
	}
	beliefs, err := env.learning.ListActiveBeliefs(ctx, env.userID, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(beliefs) != 1 || beliefs[0].EffectiveScope() != learningdomain.ScopeGlobal || beliefs[0].ScopeKey != "" {
		t.Fatalf("unknown context did not fall back to global: %+v", beliefs)
	}
}

func TestRunDailyRecompileFailureDoesNotReapplyDelta(t *testing.T) {
	env := newWorkerTestEnv(t)
	ctx := context.Background()
	day := time.Date(2026, 9, 6, 12, 0, 0, 0, time.UTC)
	belief, err := env.learning.UpsertBelief(ctx, env.userID, "Prefiere reuniones a las 9am", learningdomain.CategorySchedule)
	if err != nil {
		t.Fatal(err)
	}
	env.mustProposal(t, "create_reminder", actionsdomain.FeedbackApprovedDirect, day)
	env.model.body = marshalSynthesis(t, synthesisOutput{
		Reinforced: []reinforcedBelief{{BeliefID: belief.ID, Evidence: "approved reminder", ConfidenceDelta: 0.1}},
	})
	env.learning = learningapp.NewService(env.beliefs, failingPromptVersionRepository{env.prompts}, env.summaries)
	env.worker.deps.Learning = env.learning

	if err := env.worker.RunDaily(ctx, env.userID, day); !errors.Is(err, errRecompile) {
		t.Fatalf("expected recompile failure, got %v", err)
	}
	summary, err := env.learning.FindDailySummary(ctx, env.userID, "2026-09-06")
	if err != nil {
		t.Fatalf("reserved summary missing after recompile failure: %v", err)
	}
	if len(summary.DeltaVsPrevious.Reinforced) != 0 {
		t.Fatal("failed run should leave the reserved summary incomplete")
	}
	if err := env.worker.RunDaily(ctx, env.userID, day); err != nil {
		t.Fatal(err)
	}
	updated, err := env.learning.GetBelief(ctx, env.userID, belief.ID)
	if err != nil {
		t.Fatal(err)
	}
	if updated.EvidenceCount != 2 {
		t.Fatalf("EvidenceCount = %d, want 2", updated.EvidenceCount)
	}
	if env.model.callCount() != 1 {
		t.Fatalf("expected a single model call, got %d", env.model.callCount())
	}
}

func TestRunDailyConcurrentAppliesDeltaOnce(t *testing.T) {
	env := newWorkerTestEnv(t)
	ctx := context.Background()
	day := time.Date(2026, 9, 6, 12, 0, 0, 0, time.UTC)
	belief, err := env.learning.UpsertBelief(ctx, env.userID, "Prefiere reuniones a las 9am", learningdomain.CategorySchedule)
	if err != nil {
		t.Fatal(err)
	}
	env.mustProposal(t, "create_reminder", actionsdomain.FeedbackApprovedDirect, day)
	env.model.body = marshalSynthesis(t, synthesisOutput{
		Reinforced: []reinforcedBelief{{BeliefID: belief.ID, Evidence: "approved reminder", ConfidenceDelta: 0.1}},
	})
	var ready sync.WaitGroup
	ready.Add(2)
	release := make(chan struct{})
	env.model.ready = &ready
	env.model.release = release
	other := New(env.worker.deps, env.worker.options)

	errs := make(chan error, 2)
	go func() { errs <- env.worker.RunDaily(ctx, env.userID, day) }()
	go func() { errs <- other.RunDaily(ctx, env.userID, day) }()
	ready.Wait()
	close(release)
	for range 2 {
		if err := <-errs; err != nil {
			t.Fatal(err)
		}
	}
	updated, err := env.learning.GetBelief(ctx, env.userID, belief.ID)
	if err != nil {
		t.Fatal(err)
	}
	if updated.EvidenceCount != 2 {
		t.Fatalf("EvidenceCount = %d, want 2", updated.EvidenceCount)
	}
}

type workerTestEnv struct {
	userID    string
	worker    *Worker
	learning  *learningapp.Service
	model     *scriptedModelClient
	proposals *actionsinfra.InMemoryAIActionProposalRepository
	beliefs   *learninginfra.InMemoryBeliefRepository
	prompts   *learninginfra.InMemoryPromptVersionRepository
	summaries *learninginfra.InMemoryDailySummaryRepository
}

func newWorkerTestEnv(t *testing.T) *workerTestEnv {
	t.Helper()
	authRepo := authinfra.NewInMemoryUserRepository()
	profiles := usersinfra.NewInMemoryUserProfileRepository()
	settings := usersinfra.NewInMemoryAISettingsRepository()
	users := usersapp.NewService(authRepo, profiles, settings)
	user := authdomain.NewUser("user-1", "Sofia Tester", "tester@example.com", "secret")
	if err := authRepo.Save(user); err != nil {
		t.Fatal(err)
	}
	if err := users.ProvisionDefaults(user); err != nil {
		t.Fatal(err)
	}
	timezone := "UTC"
	if _, err := users.UpdateProfile(user.ID, usersdomain.ProfileUpdate{Timezone: &timezone}); err != nil {
		t.Fatal(err)
	}

	beliefs := learninginfra.NewInMemoryBeliefRepository()
	prompts := learninginfra.NewInMemoryPromptVersionRepository()
	summaries := learninginfra.NewInMemoryDailySummaryRepository()
	learning := learningapp.NewService(beliefs, prompts, summaries)
	learning.SetContextRepository(learninginfra.NewInMemoryUserContextRepository())
	proposals := actionsinfra.NewInMemoryAIActionProposalRepository()
	model := &scriptedModelClient{body: `{"reinforced":[],"contradicted":[],"novel":[]}`, usage: runtimedomain.Usage{InputTokens: 12, OutputTokens: 8, Model: "fake"}}
	audit := privacyinfra.NewInMemoryAuditLogRepository()
	runner := New(Deps{
		Proposals: proposals,
		Learning:  learning,
		Users:     users,
		UserIDs:   authRepo,
		Model:     model,
		Audit:     auditAdapter{repo: audit},
	}, Options{WorkerID: "test-synth", Now: func() time.Time { return time.Date(2026, 9, 7, 3, 0, 0, 0, time.UTC) }})
	return &workerTestEnv{
		userID: user.ID, worker: runner, learning: learning, model: model, proposals: proposals,
		beliefs: beliefs, prompts: prompts, summaries: summaries,
	}
}

func (env *workerTestEnv) mustProposal(t *testing.T, tool, feedback string, day time.Time) {
	t.Helper()
	proposal, err := actionsdomain.NewAIActionProposal(uuid.New().String(), env.userID, actionsdomain.ProposalCreate{
		ToolName:      tool,
		ProposedInput: json.RawMessage(`{"title":"original"}`),
		Reason:        "test decision",
		RiskLevel:     actionsdomain.RiskMedium,
	}, true)
	if err != nil {
		t.Fatal(err)
	}
	switch feedback {
	case actionsdomain.FeedbackApprovedDirect:
		if err := proposal.Confirm(nil); err != nil {
			t.Fatal(err)
		}
	case actionsdomain.FeedbackApprovedCorrected:
		if err := proposal.Confirm(json.RawMessage(`{"title":"corrected"}`)); err != nil {
			t.Fatal(err)
		}
	case actionsdomain.FeedbackRejected:
		if err := proposal.Reject("too late"); err != nil {
			t.Fatal(err)
		}
	}
	proposal.CreatedAt = day
	if err := env.proposals.Create(context.Background(), proposal); err != nil {
		t.Fatal(err)
	}
}

type scriptedModelClient struct {
	mu      sync.Mutex
	calls   []runtimedomain.ModelRequest
	body    string
	usage   runtimedomain.Usage
	ready   *sync.WaitGroup
	release <-chan struct{}
}

func (c *scriptedModelClient) Generate(_ context.Context, request runtimedomain.ModelRequest) (runtimedomain.ModelResponse, error) {
	c.mu.Lock()
	c.calls = append(c.calls, request)
	body, usage, ready, release := c.body, c.usage, c.ready, c.release
	c.mu.Unlock()
	if ready != nil {
		ready.Done()
	}
	if release != nil {
		<-release
	}
	return runtimedomain.ModelResponse{AssistantMessage: body, Usage: usage}, nil
}

func (c *scriptedModelClient) callCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.calls)
}

func (c *scriptedModelClient) lastRequest() runtimedomain.ModelRequest {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.calls[len(c.calls)-1]
}

var errRecompile = errors.New("recompile failed")

type failingPromptVersionRepository struct {
	learningdomain.PromptVersionRepository
}

func (failingPromptVersionRepository) CreateActive(context.Context, *learningdomain.PromptVersion) error {
	return errRecompile
}

func marshalSynthesis(t *testing.T, output synthesisOutput) string {
	t.Helper()
	if output.Reinforced == nil {
		output.Reinforced = []reinforcedBelief{}
	}
	if output.Contradicted == nil {
		output.Contradicted = []contradictedBelief{}
	}
	if output.Novel == nil {
		output.Novel = []novelBelief{}
	}
	raw, err := json.Marshal(output)
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

func assertDecisionsOnly(t *testing.T, request runtimedomain.ModelRequest) {
	t.Helper()
	if request.Task != runtimedomain.TaskSynthesize {
		t.Fatalf("expected synthesize task, got %q", request.Task)
	}
	if request.PromptBase != "" {
		t.Fatal("synthesis must not send prompt_base; priors belong in the decisions payload")
	}
	if len(request.History) != 0 {
		t.Fatal("synthesis must not receive conversation history")
	}
	if strings.Contains(request.Message, "conversation_history") || strings.Contains(request.Message, "user_message") {
		t.Fatal("synthesis prompt included raw conversation fields")
	}
	if !strings.Contains(request.Message, `"decisions"`) {
		t.Fatal("synthesis prompt missing decisions payload")
	}
}

type auditAdapter struct {
	repo *privacyinfra.InMemoryAuditLogRepository
}

func (a auditAdapter) RecordAuditLog(ctx context.Context, userID, action, resourceType, resourceID string, metadata map[string]any) error {
	return a.repo.Create(ctx, &privacydomain.AuditLog{
		ID:           uuid.New().String(),
		UserID:       userID,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		CreatedAt:    time.Now().UTC(),
		Metadata:     metadata,
	})
}
