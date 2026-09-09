package application_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	activitiesapp "github.com/armandoalvarado/sofia-backend/internal/activities/application"
	activitiesdomain "github.com/armandoalvarado/sofia-backend/internal/activities/domain"
	activitiesinfra "github.com/armandoalvarado/sofia-backend/internal/activities/infrastructure"
	actionsapp "github.com/armandoalvarado/sofia-backend/internal/ai/actions/application"
	actionsdomain "github.com/armandoalvarado/sofia-backend/internal/ai/actions/domain"
	actionsinfra "github.com/armandoalvarado/sofia-backend/internal/ai/actions/infrastructure"
	runtimeapp "github.com/armandoalvarado/sofia-backend/internal/ai/runtime/application"
	runtimedomain "github.com/armandoalvarado/sofia-backend/internal/ai/runtime/domain"
	runtimeinfra "github.com/armandoalvarado/sofia-backend/internal/ai/runtime/infrastructure"
	authdomain "github.com/armandoalvarado/sofia-backend/internal/auth/domain"
	authinfra "github.com/armandoalvarado/sofia-backend/internal/auth/infrastructure"
	insightsapp "github.com/armandoalvarado/sofia-backend/internal/insights/application"
	insightsinfra "github.com/armandoalvarado/sofia-backend/internal/insights/infrastructure"
	learningdomain "github.com/armandoalvarado/sofia-backend/internal/learning/domain"
	memoryapp "github.com/armandoalvarado/sofia-backend/internal/memory/application"
	memorydomain "github.com/armandoalvarado/sofia-backend/internal/memory/domain"
	memoryinfra "github.com/armandoalvarado/sofia-backend/internal/memory/infrastructure"
	remindersapp "github.com/armandoalvarado/sofia-backend/internal/reminders/application"
	remindersdomain "github.com/armandoalvarado/sofia-backend/internal/reminders/domain"
	remindersinfra "github.com/armandoalvarado/sofia-backend/internal/reminders/infrastructure"
	toolsapp "github.com/armandoalvarado/sofia-backend/internal/tools/application"
	toolsdomain "github.com/armandoalvarado/sofia-backend/internal/tools/domain"
	toolsinfra "github.com/armandoalvarado/sofia-backend/internal/tools/infrastructure"
	usersapp "github.com/armandoalvarado/sofia-backend/internal/users/application"
	usersdomain "github.com/armandoalvarado/sofia-backend/internal/users/domain"
	usersinfra "github.com/armandoalvarado/sofia-backend/internal/users/infrastructure"
)

func TestContextBuilderCollectsOnlyUserData(t *testing.T) {
	env := newRuntimeTestEnv(t)
	ctx := context.Background()
	userA := env.createUser(t, "user-a", "a@example.com")
	userB := env.createUser(t, "user-b", "b@example.com")

	ownActivity, err := env.activities.CreateActivity(ctx, userA.ID, activitiesdomain.ActivityCreate{Title: "Own activity", Timezone: "America/Tijuana"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = env.activities.CreateActivity(ctx, userB.ID, activitiesdomain.ActivityCreate{Title: "Other activity", Timezone: "America/Tijuana"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = env.memories.CreateMemory(ctx, userA.ID, memorydomain.MemoryCreate{Title: "Coffee preference", Content: "Likes coffee before planning", Tags: []string{"planning"}})
	if err != nil {
		t.Fatal(err)
	}

	summary, err := env.contextBuilder.Build(ctx, userA.ID, "coffee")
	if err != nil {
		t.Fatal(err)
	}
	if summary.User.ID != userA.ID {
		t.Fatalf("unexpected user summary: %+v", summary.User)
	}
	if len(summary.RecentActivities) != 1 || summary.RecentActivities[0].ID != ownActivity.ID {
		t.Fatalf("context leaked or missed activities: %+v", summary.RecentActivities)
	}
	if len(summary.RelevantMemories) != 1 || summary.RelevantMemories[0].Title == "" {
		t.Fatalf("expected relevant own memory, got %+v", summary.RelevantMemories)
	}
}

func TestContextBuilderLoadsPromptBaseWithoutMixingIntoDynamicContext(t *testing.T) {
	env := newRuntimeTestEnv(t)
	ctx := context.Background()
	user := env.createUser(t, "user-a", "a@example.com")
	env.contextBuilder.SetPromptBaseReader(stubPromptBaseReader{content: "Prefer morning meetings."})

	summary, err := env.contextBuilder.Build(ctx, user.ID, "plan the day")
	if err != nil {
		t.Fatal(err)
	}
	if summary.PromptBase != "Prefer morning meetings." {
		t.Fatalf("expected prompt base, got %q", summary.PromptBase)
	}
	raw, err := json.Marshal(summary)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "Prefer morning meetings.") {
		t.Fatal("prompt base leaked into the dynamic context payload")
	}
}

func TestContextBuilderColdStartWithoutPromptVersion(t *testing.T) {
	env := newRuntimeTestEnv(t)
	user := env.createUser(t, "user-a", "a@example.com")
	summary, err := env.contextBuilder.Build(context.Background(), user.ID, "hello")
	if err != nil {
		t.Fatal(err)
	}
	if summary.PromptBase != "" {
		t.Fatalf("cold start should have empty prompt base, got %q", summary.PromptBase)
	}
}

type stubPromptBaseReader struct {
	content string
}

func (s stubPromptBaseReader) ActivePromptContent(context.Context, string) (string, error) {
	return s.content, nil
}

func TestContextBuilderRecallsMemoryFromNaturalLanguage(t *testing.T) {
	env := newRuntimeTestEnv(t)
	ctx := context.Background()
	user := env.createUser(t, "user-a", "a@example.com")

	_, err := env.memories.CreateMemory(ctx, user.ID, memorydomain.MemoryCreate{
		Title:   "Coffee preference",
		Content: "Likes coffee before planning the day",
		Tags:    []string{"planning"},
	})
	if err != nil {
		t.Fatal(err)
	}

	summary, err := env.contextBuilder.Build(ctx, user.ID, "Hey can you please remind me what I like about coffee when I am planning my morning?")
	if err != nil {
		t.Fatal(err)
	}
	if len(summary.RelevantMemories) != 1 {
		t.Fatalf("expected natural-language recall to find the coffee memory, got %+v", summary.RelevantMemories)
	}
}

func TestContextBuilderRedactsEmailWithoutDroppingMemory(t *testing.T) {
	env := newRuntimeTestEnv(t)
	ctx := context.Background()
	user := env.createUser(t, "user-a", "a@example.com")

	_, err := env.memories.CreateMemory(ctx, user.ID, memorydomain.MemoryCreate{
		Title:   "Coffee contact",
		Content: "Likes coffee",
		Summary: "Talk to alice@example.com about coffee before planning",
	})
	if err != nil {
		t.Fatal(err)
	}

	summary, err := env.contextBuilder.Build(ctx, user.ID, "coffee planning")
	if err != nil {
		t.Fatal(err)
	}
	if len(summary.RelevantMemories) != 1 {
		t.Fatalf("expected redacted memory, got %+v", summary.RelevantMemories)
	}
	title := summary.RelevantMemories[0].Title
	if title == "[REDACTED]" {
		t.Fatal("email redaction wiped the whole memory text")
	}
	if strings.Contains(title, "alice@example.com") {
		t.Fatalf("email was not redacted: %q", title)
	}
	if !strings.Contains(title, "[email]") || !strings.Contains(title, "Talk to") {
		t.Fatalf("expected email placeholder with remaining text, got %q", title)
	}
}

func TestContextBuilderCountsInsightsInTokenBudget(t *testing.T) {
	env := newRuntimeTestEnv(t)
	ctx := context.Background()
	user := env.createUser(t, "user-a", "a@example.com")

	summary, err := env.contextBuilder.Build(ctx, user.ID, "para")
	if err != nil {
		t.Fatal(err)
	}
	if summary.TokenBudget.UsedApproxTokens <= 0 {
		t.Fatalf("insights should count toward UsedApproxTokens, got %+v", summary.TokenBudget)
	}
	if summary.TokenBudget.UsedApproxTokens > summary.TokenBudget.MaxApproxTokens {
		t.Fatalf("token budget exceeded: %+v", summary.TokenBudget)
	}
}

func TestContextBuilderKeepsMemoriesAlongsideActivities(t *testing.T) {
	env := newRuntimeTestEnv(t)
	ctx := context.Background()
	user := env.createUser(t, "user-a", "a@example.com")

	longTitle := strings.Repeat("activity planning notes ", 8)
	for i := 0; i < 5; i++ {
		if _, err := env.activities.CreateActivity(ctx, user.ID, activitiesdomain.ActivityCreate{Title: longTitle, Timezone: "America/Tijuana"}); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := env.memories.CreateMemory(ctx, user.ID, memorydomain.MemoryCreate{
		Title:   "Coffee preference",
		Content: "Likes coffee before planning",
	}); err != nil {
		t.Fatal(err)
	}

	summary, err := env.contextBuilder.Build(ctx, user.ID, "Hey can you please remind me what I like about coffee when I am planning my morning?")
	if err != nil {
		t.Fatal(err)
	}
	if len(summary.RecentActivities) == 0 {
		t.Fatal("expected activities in context")
	}
	if len(summary.RelevantMemories) != 1 {
		t.Fatalf("activities should not starve memories, got memories=%+v activities=%d used=%d", summary.RelevantMemories, len(summary.RecentActivities), summary.TokenBudget.UsedApproxTokens)
	}
}

func TestContextBuilderSetsCurrentDateTime(t *testing.T) {
	t.Run("profile timezone", func(t *testing.T) {
		env := newRuntimeTestEnv(t)
		user := env.createUser(t, "user-timezone", "timezone@example.com")
		timezone := "America/Mexico_City"
		if _, err := env.users.UpdateProfile(user.ID, usersdomain.ProfileUpdate{Timezone: &timezone}); err != nil {
			t.Fatal(err)
		}

		summary, err := env.contextBuilder.Build(context.Background(), user.ID, "mañana")
		if err != nil {
			t.Fatal(err)
		}
		current, err := time.Parse(time.RFC3339, summary.CurrentDateTime)
		if err != nil {
			t.Fatalf("current_datetime is not RFC3339: %q: %v", summary.CurrentDateTime, err)
		}
		location, err := time.LoadLocation(timezone)
		if err != nil {
			t.Fatal(err)
		}
		_, gotOffset := current.Zone()
		_, wantOffset := current.In(location).Zone()
		if gotOffset != wantOffset {
			t.Fatalf("current_datetime offset = %d, want %d for %s", gotOffset, wantOffset, timezone)
		}
	})

	t.Run("missing profile falls back to UTC", func(t *testing.T) {
		env := newRuntimeTestEnv(t)
		user := env.createUser(t, "user-utc", "utc@example.com")
		builder := runtimeapp.NewContextBuilder(
			profilelessUserReader{delegate: env.users},
			env.activities,
			env.reminders,
			env.insights,
			env.memories,
			5,
			4600,
		)

		summary, err := builder.Build(context.Background(), user.ID, "hoy")
		if err != nil {
			t.Fatal(err)
		}
		current, err := time.Parse(time.RFC3339, summary.CurrentDateTime)
		if err != nil {
			t.Fatalf("current_datetime is not RFC3339: %q: %v", summary.CurrentDateTime, err)
		}
		_, offset := current.Zone()
		if offset != 0 {
			t.Fatalf("current_datetime offset = %d, want UTC", offset)
		}
	})
}

func TestSettingsGateContextAndTools(t *testing.T) {
	env := newRuntimeTestEnv(t)
	ctx := context.Background()
	user := env.createUser(t, "user-a", "a@example.com")

	memoryEnabled := false
	remindersEnabled := false
	planningEnabled := false
	if _, err := env.users.UpdateAISettings(user.ID, usersdomain.AISettingsUpdate{
		MemoryEnabled:    &memoryEnabled,
		RemindersEnabled: &remindersEnabled,
		PlanningEnabled:  &planningEnabled,
	}); err != nil {
		t.Fatal(err)
	}

	summary, err := env.contextBuilder.Build(ctx, user.ID, "remember coffee")
	if err != nil {
		t.Fatal(err)
	}
	if summary.MemoryIncluded || summary.RemindersIncluded || len(summary.RelevantMemories) != 0 {
		t.Fatalf("disabled context was included: %+v", summary)
	}

	tools, err := env.toolSelector.SelectTools(ctx, summary, "remember coffee")
	if err != nil {
		t.Fatal(err)
	}
	for _, tool := range tools {
		if tool.Category == toolsdomain.CategoryMemory || tool.Category == toolsdomain.CategoryReminders || tool.Category == toolsdomain.CategoryActivities {
			t.Fatalf("tool should have been gated by settings: %+v", tool)
		}
	}
}

func TestRuntimeDryRunDoesNotCreateProposal(t *testing.T) {
	env := newRuntimeTestEnv(t)
	ctx := context.Background()
	user := env.createUser(t, "user-a", "a@example.com")

	resp, err := env.runtime.HandleMessage(ctx, runtimedomain.RuntimeRequest{
		UserID:  user.ID,
		Message: "Recuérdame estudiar mañana",
		DryRun:  true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Mode != runtimedomain.ModeDryRun || len(resp.ProposedActions) != 1 {
		t.Fatalf("unexpected dry_run response: %+v", resp)
	}

	proposals, err := env.actions.ListActionProposals(ctx, actionsdomain.ListFilter{UserID: user.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(proposals) != 0 {
		t.Fatalf("dry_run created proposals: %+v", proposals)
	}
}

func TestRuntimeCreatesSafeProposalWithoutExecuting(t *testing.T) {
	env := newRuntimeTestEnv(t)
	ctx := context.Background()
	user := env.createUser(t, "user-a", "a@example.com")

	resp, err := env.runtime.HandleMessage(ctx, runtimedomain.RuntimeRequest{
		UserID:  user.ID,
		Message: "Recuérdame estudiar mañana",
		DryRun:  false,
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Mode != runtimedomain.ModeReady || len(resp.ProposedActions) != 1 || resp.ProposedActions[0].ID == "" {
		t.Fatalf("unexpected proposal response: %+v", resp)
	}

	proposals, err := env.actions.ListActionProposals(ctx, actionsdomain.ListFilter{UserID: user.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(proposals) != 1 || proposals[0].Status != actionsdomain.StatusProposed {
		t.Fatalf("expected one proposed action, got %+v", proposals)
	}
	reminders, err := env.reminders.ListReminders(ctx, remindersdomain.ListFilter{UserID: user.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(reminders) != 0 {
		t.Fatalf("runtime executed automatically: %+v", reminders)
	}
	if resp.RequestID == "" || resp.Observability.ProviderLatencyMS < 0 || resp.Observability.ProposalsCount != 1 || resp.Observability.RejectedActionsCount != 0 {
		t.Fatalf("missing runtime observability: %+v", resp.Observability)
	}
}

func TestRuntimeAutoExecutesStrongRecentHistory(t *testing.T) {
	env := newRuntimeTestEnv(t)
	ctx := context.Background()
	user := env.createUser(t, "user-a", "a@example.com")
	level := usersdomain.AutonomySemiAutonomous
	if _, err := env.users.UpdateAISettings(user.ID, usersdomain.AISettingsUpdate{AutonomyLevel: &level}); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 36; i++ {
		proposal, err := env.actions.CreateActionProposal(ctx, user.ID, actionsdomain.ProposalCreate{
			ToolName:      toolsdomain.ToolCreateReminder,
			ProposedInput: json.RawMessage(`{"title":"habit","scheduled_at":"2026-09-08T15:00:00Z","timezone":"UTC"}`),
			RiskLevel:     actionsdomain.RiskMedium,
		})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := env.actions.ConfirmActionProposal(ctx, user.ID, proposal.ID, nil, nil); err != nil {
			t.Fatal(err)
		}
	}
	for i := 0; i < 4; i++ {
		proposal, err := env.actions.CreateActionProposal(ctx, user.ID, actionsdomain.ProposalCreate{
			ToolName:      toolsdomain.ToolCreateReminder,
			ProposedInput: json.RawMessage(`{"title":"skip","scheduled_at":"2026-09-08T15:00:00Z","timezone":"UTC"}`),
			RiskLevel:     actionsdomain.RiskMedium,
		})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := env.actions.RejectActionProposal(ctx, user.ID, proposal.ID, "no", nil); err != nil {
			t.Fatal(err)
		}
	}
	safety := runtimeapp.NewSafetyPolicy()
	safety.SetAutonomy(env.actions, learningdomain.DefaultAutonomyThreshold)
	env.runtime = runtimeapp.NewRuntimeService(env.contextBuilder, env.toolSelector, runtimeapp.NewPlanner(runtimeinfra.NewFakeModelClient()), safety, env.actions)

	resp, err := env.runtime.HandleMessage(ctx, runtimedomain.RuntimeRequest{
		UserID:  user.ID,
		Message: "Recuérdame estudiar mañana",
		DryRun:  false,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.ProposedActions) != 1 || resp.ProposedActions[0].Status != actionsdomain.StatusExecuted {
		t.Fatalf("expected auto-executed reminder, got %+v", resp.ProposedActions)
	}
	reminders, err := env.reminders.ListReminders(ctx, remindersdomain.ListFilter{UserID: user.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(reminders) != 1 {
		t.Fatalf("expected one reminder after auto-execute, got %+v", reminders)
	}
}

func TestRuntimeInvalidModelActionDoesNotCreateProposal(t *testing.T) {
	env := newRuntimeTestEnv(t)
	ctx := context.Background()
	user := env.createUser(t, "user-a", "a@example.com")

	planner := runtimeapp.NewPlanner(invalidActionModel{})
	env.runtime = runtimeapp.NewRuntimeService(env.contextBuilder, env.toolSelector, planner, runtimeapp.NewSafetyPolicy(), env.actions)

	resp, err := env.runtime.HandleMessage(ctx, runtimedomain.RuntimeRequest{
		UserID:  user.ID,
		Message: "Create a reminder with unsafe fields",
		DryRun:  false,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.ProposedActions) != 0 {
		t.Fatalf("expected unsafe model action to be skipped, got %+v", resp.ProposedActions)
	}
	if resp.Observability.RejectedActionsCount != 1 {
		t.Fatalf("expected rejected action count, got %+v", resp.Observability)
	}
	proposals, err := env.actions.ListActionProposals(ctx, actionsdomain.ListFilter{UserID: user.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(proposals) != 0 {
		t.Fatalf("unsafe model action created proposals: %+v", proposals)
	}
}

func TestHandleMessagePassesHistoryToModelRequest(t *testing.T) {
	env := newRuntimeTestEnv(t)
	ctx := context.Background()
	user := env.createUser(t, "user-a", "a@example.com")
	recorder := &recordingPlanner{inner: runtimeapp.NewPlanner(runtimeinfra.NewFakeModelClient())}
	env.runtime = runtimeapp.NewRuntimeService(env.contextBuilder, env.toolSelector, recorder, runtimeapp.NewSafetyPolicy(), env.actions)

	if _, err := env.runtime.HandleMessage(ctx, runtimedomain.RuntimeRequest{
		UserID:  user.ID,
		Message: "mejor a las 4",
		DryRun:  true,
		History: []runtimedomain.Turn{{Role: "user", Content: "Recuérdame estudiar mañana"}},
	}); err != nil {
		t.Fatal(err)
	}
	if len(recorder.requests) != 1 {
		t.Fatalf("expected 1 model request, got %d", len(recorder.requests))
	}
	history := recorder.requests[0].History
	if len(history) != 1 || history[0].Content != "Recuérdame estudiar mañana" {
		t.Fatalf("expected prior turn in ModelRequest.History, got %+v", history)
	}
	if recorder.requests[0].Context.TokenBudget.MaxApproxTokens != 4600 {
		t.Fatalf("history should fit into the existing budget, got %+v", recorder.requests[0].Context.TokenBudget)
	}
	if recorder.requests[0].Context.TokenBudget.UsedApproxTokens <= 0 {
		t.Fatalf("history was not counted in token budget: %+v", recorder.requests[0].Context.TokenBudget)
	}
}

func TestRuntimeAuditMetadataDoesNotIncludeFullPrompt(t *testing.T) {
	env := newRuntimeTestEnv(t)
	ctx := context.Background()
	user := env.createUser(t, "user-a", "a@example.com")
	audit := &recordingAuditRecorder{}
	env.runtime.SetAuditRecorder(audit)
	prompt := "Recuérdame estudiar mañana y este texto no debe aparecer completo en audit metadata"

	if _, err := env.runtime.HandleMessage(ctx, runtimedomain.RuntimeRequest{
		UserID:  user.ID,
		Message: prompt,
		DryRun:  true,
	}); err != nil {
		t.Fatal(err)
	}

	raw, _ := json.Marshal(audit.metadata)
	if strings.Contains(string(raw), prompt) {
		t.Fatalf("audit metadata leaked full prompt: %s", raw)
	}
	if !strings.Contains(string(raw), "request_id") || !strings.Contains(string(raw), "provider_latency_ms") ||
		!strings.Contains(string(raw), "input_tokens") || !strings.Contains(string(raw), "output_tokens") ||
		!strings.Contains(string(raw), "cached_input_tokens") {
		t.Fatalf("expected observability metadata, got %s", raw)
	}
}

type invalidActionModel struct{}

func (invalidActionModel) Generate(context.Context, runtimedomain.ModelRequest) (runtimedomain.ModelResponse, error) {
	input, _ := json.Marshal(map[string]any{"user_id": "other-user", "title": "Unsafe"})
	return runtimedomain.ModelResponse{
		AssistantMessage: "Unsafe proposal should be filtered.",
		PlannedActions: []runtimedomain.PlannedAction{{
			ToolName:      toolsdomain.ToolCreateReminder,
			ProposedInput: input,
			Reason:        "contains forbidden fields and misses required scheduled_at/timezone",
			RiskLevel:     actionsdomain.RiskMedium,
		}},
	}, nil
}

type recordingPlanner struct {
	inner    runtimedomain.Planner
	requests []runtimedomain.ModelRequest
}

func (p *recordingPlanner) Plan(ctx context.Context, request runtimedomain.ModelRequest) (runtimedomain.ModelResponse, error) {
	copied := request
	if request.History != nil {
		copied.History = append([]runtimedomain.Turn(nil), request.History...)
	}
	p.requests = append(p.requests, copied)
	return p.inner.Plan(ctx, request)
}

type recordingAuditRecorder struct {
	metadata []map[string]any
}

func (r *recordingAuditRecorder) RecordAuditLog(ctx context.Context, userID, action, resourceType, resourceID string, metadata map[string]any) error {
	r.metadata = append(r.metadata, metadata)
	return nil
}

type runtimeTestEnv struct {
	users          *usersapp.Service
	activities     *activitiesapp.Service
	reminders      *remindersapp.Service
	insights       *insightsapp.Service
	memories       *memoryapp.Service
	actions        *actionsapp.Service
	contextBuilder *runtimeapp.ContextBuilder
	toolSelector   *runtimeapp.ToolSelector
	runtime        *runtimeapp.RuntimeService
	authRepo       *authinfra.InMemoryUserRepository
	profiles       *usersinfra.InMemoryUserProfileRepository
	settings       *usersinfra.InMemoryAISettingsRepository
}

type profilelessUserReader struct {
	delegate *usersapp.Service
}

func (r profilelessUserReader) GetMe(userID string) (*usersapp.Me, error) {
	me, err := r.delegate.GetMe(userID)
	if err != nil {
		return nil, err
	}
	withoutProfile := *me
	withoutProfile.Profile = nil
	return &withoutProfile, nil
}

func newRuntimeTestEnv(t *testing.T) *runtimeTestEnv {
	t.Helper()
	ctx := context.Background()
	authRepo := authinfra.NewInMemoryUserRepository()
	profiles := usersinfra.NewInMemoryUserProfileRepository()
	settings := usersinfra.NewInMemoryAISettingsRepository()
	activityRepo := activitiesinfra.NewInMemoryActivityRepository()
	reminderRepo := remindersinfra.NewInMemoryReminderRepository()
	moodRepo := insightsinfra.NewInMemoryMoodEntryRepository()
	outcomeRepo := insightsinfra.NewInMemoryActivityOutcomeRepository()
	reflectionRepo := insightsinfra.NewInMemoryReflectionRepository()
	memoryRepo := memoryinfra.NewInMemoryMemoryRepository()
	toolRepo := toolsinfra.NewInMemoryToolDefinitionRepository()
	actionRepo := actionsinfra.NewInMemoryAIActionProposalRepository()

	users := usersapp.NewService(authRepo, profiles, settings)
	activities := activitiesapp.NewService(activityRepo)
	reminders := remindersapp.NewService(reminderRepo, activityRepo, users)
	insights := insightsapp.NewService(moodRepo, outcomeRepo, reflectionRepo, activityRepo)
	memories := memoryapp.NewService(memoryRepo, users, nil)
	tools := toolsapp.NewService(toolRepo)
	if err := tools.SeedDefaultTools(ctx); err != nil {
		t.Fatal(err)
	}
	actions := actionsapp.NewService(actionRepo, tools, users, activities, reminders, memories)
	activities.SetReminderBridge(reminders)

	contextBuilder := runtimeapp.NewContextBuilder(users, activities, reminders, insights, memories, 5, 4600)
	toolSelector := runtimeapp.NewToolSelector(tools)
	planner := runtimeapp.NewPlanner(runtimeinfra.NewFakeModelClient())
	runtime := runtimeapp.NewRuntimeService(contextBuilder, toolSelector, planner, runtimeapp.NewSafetyPolicy(), actions)

	return &runtimeTestEnv{
		users:          users,
		activities:     activities,
		reminders:      reminders,
		insights:       insights,
		memories:       memories,
		actions:        actions,
		contextBuilder: contextBuilder,
		toolSelector:   toolSelector,
		runtime:        runtime,
		authRepo:       authRepo,
		profiles:       profiles,
		settings:       settings,
	}
}

func (e *runtimeTestEnv) createUser(t *testing.T, id, email string) *authdomain.User {
	t.Helper()
	user := authdomain.NewUser(id, "Test User", email, "hash")
	if err := e.authRepo.Save(user); err != nil {
		t.Fatal(err)
	}
	if err := e.users.ProvisionDefaults(user); err != nil {
		t.Fatal(err)
	}
	return user
}

func TestRuntimeListsDueRemindersInContext(t *testing.T) {
	env := newRuntimeTestEnv(t)
	ctx := context.Background()
	user := env.createUser(t, "user-a", "a@example.com")

	_, err := env.reminders.CreateReminder(ctx, user.ID, remindersdomain.ReminderCreate{
		Title:       "Due reminder",
		ScheduledAt: time.Now().Add(-time.Hour),
		Timezone:    "America/Tijuana",
	})
	if err != nil {
		t.Fatal(err)
	}

	summary, err := env.contextBuilder.Build(ctx, user.ID, "status")
	if err != nil {
		t.Fatal(err)
	}
	if len(summary.DueReminders) != 1 || summary.DueReminders[0].Title != "Due reminder" {
		t.Fatalf("expected due reminder in context, got %+v", summary.DueReminders)
	}
}
