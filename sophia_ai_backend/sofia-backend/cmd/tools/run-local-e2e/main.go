package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	activitiesdomain "github.com/armandoalvarado/sofia-backend/internal/activities/domain"
	actionsdomain "github.com/armandoalvarado/sofia-backend/internal/ai/actions/domain"
	runtimedomain "github.com/armandoalvarado/sofia-backend/internal/ai/runtime/domain"
	"github.com/armandoalvarado/sofia-backend/internal/app"
	authapp "github.com/armandoalvarado/sofia-backend/internal/auth/application"
	"github.com/armandoalvarado/sofia-backend/internal/config"
	memorydomain "github.com/armandoalvarado/sofia-backend/internal/memory/domain"
	remindersdomain "github.com/armandoalvarado/sofia-backend/internal/reminders/domain"
	usersdomain "github.com/armandoalvarado/sofia-backend/internal/users/domain"
)

type e2eStep struct {
	Name   string
	Status string
	Detail string
}

type e2eReport struct {
	GeneratedAt              time.Time
	UserID                   string
	ActivityID               string
	ReminderID               string
	MemoryID                 string
	ProposalID               string
	ExecutedProposalStatus   string
	ExportActivities         int
	ExportReminders          int
	ExportMemories           int
	ExportActionProposals    int
	RuntimeDryRunActions     int
	ProposalsBeforeDryRun    int
	ProposalsAfterDryRun     int
	ProposalOnlyPreserved    bool
	NoAutomaticAIExecution   bool
	ExplicitExecutionCreated bool
	Steps                    []e2eStep
}

func main() {
	outputDir := flag.String("output-dir", "docs", "directory where local_e2e_report.md is written")
	flag.Parse()

	report, err := runE2E(context.Background(), *outputDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "local e2e failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("wrote %s with %d steps\n", filepath.Join(*outputDir, "local_e2e_report.md"), len(report.Steps))
}

func runE2E(ctx context.Context, outputDir string) (*e2eReport, error) {
	report := &e2eReport{GeneratedAt: time.Now().UTC()}
	cfg := config.Config{
		Env:                   "test",
		Port:                  "8080",
		JWTSecret:             "local-e2e-insecure-secret",
		PersistenceDriver:     "memory",
		AIModelProvider:       "fake",
		AIRuntimeEnabled:      true,
		AIRuntimeProposalOnly: true,
		RequestBodyLimitBytes: 1 << 20,
		ReadTimeout:           10 * time.Second,
		WriteTimeout:          10 * time.Second,
		IdleTimeout:           time.Minute,
		ReadHeaderTimeout:     5 * time.Second,
	}

	repositories, err := app.BuildRepositories(cfg)
	if err != nil {
		return report, err
	}
	defer repositories.Close()
	modules, err := app.BuildModules(cfg, repositories)
	if err != nil {
		return report, err
	}

	register := authapp.NewRegisterUserUseCase(repositories.Users)
	user, err := register.Execute("Local E2E User", "local.e2e@example.test", "LocalE2E123!")
	if err != nil {
		return report, err
	}
	report.UserID = user.ID
	report.pass("register user", "auth user created in memory")

	if err := modules.Users.ProvisionDefaults(user); err != nil {
		return report, err
	}
	me, err := modules.Users.GetMe(user.ID)
	if err != nil {
		return report, err
	}
	if me.Profile == nil || me.AISettings == nil {
		return report, errors.New("profile/settings were not provisioned")
	}
	report.pass("create profile/settings", "default user profile and AI settings created")

	memoryEnabled, remindersEnabled, planningEnabled := true, true, true
	autonomy := usersdomain.AutonomyManual
	if _, err := modules.Users.UpdateAISettings(user.ID, usersdomain.AISettingsUpdate{
		MemoryEnabled:     &memoryEnabled,
		RemindersEnabled:  &remindersEnabled,
		PlanningEnabled:   &planningEnabled,
		AutonomyLevel:     &autonomy,
		PreferredLanguage: strPtr("es"),
	}); err != nil {
		return report, err
	}
	report.pass("configure AI settings", "memory/reminders/planning enabled with manual autonomy")

	now := time.Date(2026, 7, 5, 12, 0, 0, 0, time.UTC)
	activity, err := modules.Activities.CreateActivity(ctx, user.ID, activitiesdomain.ActivityCreate{
		Type:        activitiesdomain.TypeTask,
		Title:       "Local E2E activity",
		Description: "Synthetic local E2E activity",
		Priority:    activitiesdomain.PriorityMedium,
		DueAt:       timePtr(now.Add(24 * time.Hour)),
		Timezone:    "America/Tijuana",
		Tags:        []string{"e2e", "local"},
	})
	if err != nil {
		return report, err
	}
	report.ActivityID = activity.ID
	report.pass("create activity", activity.ID)

	recurrenceUntil := now.Add(14 * 24 * time.Hour)
	reminder, err := modules.Reminders.CreateReminder(ctx, user.ID, remindersdomain.ReminderCreate{
		ActivityID:         activity.ID,
		Title:              "Local E2E recurring reminder",
		Description:        "Synthetic recurring reminder",
		ScheduledAt:        now.Add(time.Hour),
		Timezone:           "America/Tijuana",
		RecurrenceRule:     remindersdomain.RecurrenceDaily,
		RecurrenceInterval: 1,
		RecurrenceUntil:    &recurrenceUntil,
		RecurrenceCount:    3,
		MaxRetries:         3,
		Source:             remindersdomain.SourceUser,
		CreatedBy:          remindersdomain.CreatedByUser,
	})
	if err != nil {
		return report, err
	}
	report.ReminderID = reminder.ID
	if reminder.RecurrenceRule != remindersdomain.RecurrenceDaily {
		return report, errors.New("recurring reminder was not created with daily recurrence")
	}
	report.pass("create recurring reminder", reminder.ID)

	memory, err := modules.Memory.CreateMemory(ctx, user.ID, memorydomain.MemoryCreate{
		Type:       memorydomain.TypePreference,
		Title:      "Local E2E preference",
		Content:    "Synthetic local E2E memory content.",
		Summary:    "Synthetic E2E memory",
		Source:     memorydomain.SourceUserManual,
		Confidence: memorydomain.ConfidenceHigh,
		Importance: memorydomain.ImportanceMedium,
		Tags:       []string{"e2e", "local"},
		Visibility: memorydomain.VisibilityAssistantContext,
	})
	if err != nil {
		return report, err
	}
	report.MemoryID = memory.ID
	report.pass("create memory", memory.ID)

	beforeDryRun, err := modules.Actions.ListActionProposals(ctx, actionsdomain.ListFilter{UserID: user.ID, Limit: 100})
	if err != nil {
		return report, err
	}
	report.ProposalsBeforeDryRun = len(beforeDryRun)
	runtimeResponse, err := modules.AIRuntime.HandleMessage(ctx, runtimedomain.RuntimeRequest{
		UserID:  user.ID,
		Message: "Crea una tarea sintetica para revisar el flujo local E2E",
		DryRun:  true,
	})
	if err != nil {
		return report, err
	}
	report.RuntimeDryRunActions = len(runtimeResponse.ProposedActions)
	afterDryRun, err := modules.Actions.ListActionProposals(ctx, actionsdomain.ListFilter{UserID: user.ID, Limit: 100})
	if err != nil {
		return report, err
	}
	report.ProposalsAfterDryRun = len(afterDryRun)
	report.ProposalOnlyPreserved = report.ProposalsBeforeDryRun == report.ProposalsAfterDryRun
	report.NoAutomaticAIExecution = report.ProposalOnlyPreserved
	if !report.ProposalOnlyPreserved {
		return report, errors.New("AI runtime dry_run persisted proposals")
	}
	report.pass("call AI runtime fake dry_run", fmt.Sprintf("actions=%d persisted_delta=0", report.RuntimeDryRunActions))
	report.pass("validate no automatic AI execution", "dry_run did not persist or execute proposals")
	report.pass("validate proposal-only", "runtime stayed proposal-only")

	proposalInput, err := json.Marshal(map[string]any{
		"type":        activitiesdomain.TypeTask,
		"title":       "Explicit proposal E2E activity",
		"description": "Created only after explicit proposal execution",
		"priority":    activitiesdomain.PriorityLow,
		"timezone":    "America/Tijuana",
	})
	if err != nil {
		return report, err
	}
	proposal, err := modules.Actions.CreateActionProposal(ctx, user.ID, actionsdomain.ProposalCreate{
		ToolName:      "create_activity",
		ProposedInput: proposalInput,
		Reason:        "Local E2E explicit execution validation",
		RiskLevel:     actionsdomain.RiskLow,
	})
	if err != nil {
		return report, err
	}
	report.ProposalID = proposal.ID
	report.pass("create proposal", proposal.ID)

	confirmed, err := modules.Actions.ConfirmActionProposal(ctx, user.ID, proposal.ID)
	if err != nil {
		return report, err
	}
	if confirmed.Status != actionsdomain.StatusConfirmed {
		return report, errors.New("proposal was not confirmed")
	}
	report.pass("confirm proposal", confirmed.Status)

	activitiesBeforeExecute, err := modules.Activities.ListActivities(ctx, activitiesdomain.ListFilter{UserID: user.ID, Limit: 100})
	if err != nil {
		return report, err
	}
	executed, err := modules.Actions.ExecuteConfirmedActionProposal(ctx, user.ID, proposal.ID)
	if err != nil {
		return report, err
	}
	report.ExecutedProposalStatus = executed.Status
	activitiesAfterExecute, err := modules.Activities.ListActivities(ctx, activitiesdomain.ListFilter{UserID: user.ID, Limit: 100})
	if err != nil {
		return report, err
	}
	report.ExplicitExecutionCreated = executed.Status == actionsdomain.StatusExecuted && len(activitiesAfterExecute) == len(activitiesBeforeExecute)+1
	if !report.ExplicitExecutionCreated {
		return report, fmt.Errorf("explicit execution did not create exactly one activity status=%s before=%d after=%d", executed.Status, len(activitiesBeforeExecute), len(activitiesAfterExecute))
	}
	report.pass("execute proposal explicitly", "created one activity after explicit execute")

	export, err := modules.Privacy.ExportUserData(ctx, user.ID)
	if err != nil {
		return report, err
	}
	report.ExportActivities = len(export.Activities)
	report.ExportReminders = len(export.Reminders)
	report.ExportMemories = len(export.Memories)
	report.ExportActionProposals = len(export.ActionProposals)
	if report.ExportActivities < 2 || report.ExportReminders < 1 || report.ExportMemories < 1 || report.ExportActionProposals < 1 {
		return report, fmt.Errorf("export missing expected data activities=%d reminders=%d memories=%d proposals=%d", report.ExportActivities, report.ExportReminders, report.ExportMemories, report.ExportActionProposals)
	}
	report.pass("export user data", fmt.Sprintf("activities=%d reminders=%d memories=%d proposals=%d", report.ExportActivities, report.ExportReminders, report.ExportMemories, report.ExportActionProposals))

	if err := writeReport(outputDir, report); err != nil {
		return report, err
	}
	return report, nil
}

func (r *e2eReport) pass(name, detail string) {
	r.Steps = append(r.Steps, e2eStep{Name: name, Status: "pass", Detail: detail})
}

func writeReport(outputDir string, report *e2eReport) error {
	if strings.TrimSpace(outputDir) == "" {
		outputDir = "docs"
	}
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return err
	}
	var b strings.Builder
	b.WriteString("# Local E2E Report\n\n")
	b.WriteString(fmt.Sprintf("Generated at: `%s`\n\n", report.GeneratedAt.Format(time.RFC3339)))
	b.WriteString("## Summary\n\n")
	b.WriteString(fmt.Sprintf("- user_id: `%s`\n", report.UserID))
	b.WriteString(fmt.Sprintf("- activity_id: `%s`\n", report.ActivityID))
	b.WriteString(fmt.Sprintf("- reminder_id: `%s`\n", report.ReminderID))
	b.WriteString(fmt.Sprintf("- memory_id: `%s`\n", report.MemoryID))
	b.WriteString(fmt.Sprintf("- proposal_id: `%s`\n", report.ProposalID))
	b.WriteString(fmt.Sprintf("- executed_proposal_status: `%s`\n", report.ExecutedProposalStatus))
	b.WriteString(fmt.Sprintf("- proposal_only_preserved: `%t`\n", report.ProposalOnlyPreserved))
	b.WriteString(fmt.Sprintf("- no_automatic_ai_execution: `%t`\n", report.NoAutomaticAIExecution))
	b.WriteString(fmt.Sprintf("- explicit_execution_created_activity: `%t`\n", report.ExplicitExecutionCreated))
	b.WriteString("\n## Export Counts\n\n")
	b.WriteString(fmt.Sprintf("- activities: `%d`\n", report.ExportActivities))
	b.WriteString(fmt.Sprintf("- reminders: `%d`\n", report.ExportReminders))
	b.WriteString(fmt.Sprintf("- memories: `%d`\n", report.ExportMemories))
	b.WriteString(fmt.Sprintf("- action_proposals: `%d`\n", report.ExportActionProposals))
	b.WriteString("\n## AI Runtime Safety\n\n")
	b.WriteString(fmt.Sprintf("- dry_run_actions_returned: `%d`\n", report.RuntimeDryRunActions))
	b.WriteString(fmt.Sprintf("- proposals_before_dry_run: `%d`\n", report.ProposalsBeforeDryRun))
	b.WriteString(fmt.Sprintf("- proposals_after_dry_run: `%d`\n", report.ProposalsAfterDryRun))
	b.WriteString("\n## Steps\n\n")
	b.WriteString("| Step | Status | Detail |\n")
	b.WriteString("| --- | --- | --- |\n")
	for _, step := range report.Steps {
		b.WriteString(fmt.Sprintf("| %s | %s | %s |\n", step.Name, step.Status, escapeTable(step.Detail)))
	}
	return os.WriteFile(filepath.Join(outputDir, "local_e2e_report.md"), []byte(b.String()), 0o644)
}

func timePtr(value time.Time) *time.Time {
	return &value
}

func strPtr(value string) *string {
	return &value
}

func escapeTable(value string) string {
	return strings.ReplaceAll(value, "|", "\\|")
}
