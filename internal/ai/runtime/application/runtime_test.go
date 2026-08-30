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
	if !strings.Contains(string(raw), "request_id") || !strings.Contains(string(raw), "provider_latency_ms") {
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
	memories       *memoryapp.Service
	actions        *actionsapp.Service
	contextBuilder *runtimeapp.ContextBuilder
	toolSelector   *runtimeapp.ToolSelector
	runtime        *runtimeapp.RuntimeService
	authRepo       *authinfra.InMemoryUserRepository
	profiles       *usersinfra.InMemoryUserProfileRepository
	settings       *usersinfra.InMemoryAISettingsRepository
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

	contextBuilder := runtimeapp.NewContextBuilder(users, activities, reminders, insights, memories, 5)
	toolSelector := runtimeapp.NewToolSelector(tools)
	planner := runtimeapp.NewPlanner(runtimeinfra.NewFakeModelClient())
	runtime := runtimeapp.NewRuntimeService(contextBuilder, toolSelector, planner, runtimeapp.NewSafetyPolicy(), actions)

	return &runtimeTestEnv{
		users:          users,
		activities:     activities,
		reminders:      reminders,
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
