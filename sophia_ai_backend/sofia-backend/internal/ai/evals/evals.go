package evals

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	activitiesapp "github.com/armandoalvarado/sofia-backend/internal/activities/application"
	activitiesinfra "github.com/armandoalvarado/sofia-backend/internal/activities/infrastructure"
	actionsapp "github.com/armandoalvarado/sofia-backend/internal/ai/actions/application"
	actionsinfra "github.com/armandoalvarado/sofia-backend/internal/ai/actions/infrastructure"
	runtimeapp "github.com/armandoalvarado/sofia-backend/internal/ai/runtime/application"
	runtimedomain "github.com/armandoalvarado/sofia-backend/internal/ai/runtime/domain"
	runtimeinfra "github.com/armandoalvarado/sofia-backend/internal/ai/runtime/infrastructure"
	authdomain "github.com/armandoalvarado/sofia-backend/internal/auth/domain"
	authinfra "github.com/armandoalvarado/sofia-backend/internal/auth/infrastructure"
	insightsapp "github.com/armandoalvarado/sofia-backend/internal/insights/application"
	insightsinfra "github.com/armandoalvarado/sofia-backend/internal/insights/infrastructure"
	memoryapp "github.com/armandoalvarado/sofia-backend/internal/memory/application"
	memoryinfra "github.com/armandoalvarado/sofia-backend/internal/memory/infrastructure"
	remindersapp "github.com/armandoalvarado/sofia-backend/internal/reminders/application"
	remindersinfra "github.com/armandoalvarado/sofia-backend/internal/reminders/infrastructure"
	toolsapp "github.com/armandoalvarado/sofia-backend/internal/tools/application"
	toolsinfra "github.com/armandoalvarado/sofia-backend/internal/tools/infrastructure"
	usersapp "github.com/armandoalvarado/sofia-backend/internal/users/application"
	usersinfra "github.com/armandoalvarado/sofia-backend/internal/users/infrastructure"
)

const evalUserID = "eval-user"

type Result struct {
	Case                 string `json:"case"`
	Profile              string `json:"profile"`
	Input                string `json:"input"`
	ExpectedBehavior     string `json:"expected_behavior"`
	ActualBehavior       string `json:"actual_behavior"`
	Passed               bool   `json:"passed"`
	Reason               string `json:"reason"`
	LatencyMS            int64  `json:"latency_ms"`
	RejectedActionsCount int    `json:"rejected_actions_count"`
	ProposedActionsCount int    `json:"proposed_actions_count"`
	SchemaFailuresCount  int    `json:"schema_failures_count"`
	PolicyFailuresCount  int    `json:"policy_failures_count"`
}

type Report struct {
	GeneratedAt string        `json:"generated_at"`
	Provider    ProviderMeta  `json:"provider"`
	Status      string        `json:"status"`
	Reason      string        `json:"reason,omitempty"`
	Summary     ReportSummary `json:"summary"`
	Results     []Result      `json:"results"`
	Privacy     PrivacyMeta   `json:"privacy"`
}

type ProviderMeta struct {
	Name      string     `json:"name"`
	Model     string     `json:"model,omitempty"`
	LatencyMS int64      `json:"latency_ms"`
	Status    string     `json:"status"`
	Usage     *UsageMeta `json:"usage,omitempty"`
	Cost      *CostMeta  `json:"cost,omitempty"`
}

type UsageMeta struct {
	PromptTokens     int `json:"prompt_tokens,omitempty"`
	CandidatesTokens int `json:"candidates_tokens,omitempty"`
	TotalTokens      int `json:"total_tokens,omitempty"`
}

type CostMeta struct {
	EstimatedUSD float64 `json:"estimated_usd"`
	Source       string  `json:"source"`
}

type ReportSummary struct {
	Passed               int   `json:"passed"`
	Failed               int   `json:"failed"`
	AvgLatencyMS         int64 `json:"avg_latency_ms"`
	ProposedActionsCount int   `json:"proposed_actions_count"`
	RejectedActionsCount int   `json:"rejected_actions_count"`
	SchemaFailuresCount  int   `json:"schema_failures_count"`
	PolicyFailuresCount  int   `json:"policy_failures_count"`
}

type PrivacyMeta struct {
	PromptsRedacted       bool   `json:"prompts_redacted"`
	FullPromptsStored     bool   `json:"full_prompts_stored"`
	SensitiveValuesStored bool   `json:"sensitive_values_stored"`
	Note                  string `json:"note"`
}

type Dataset struct {
	Version     string        `json:"version"`
	Description string        `json:"description"`
	Cases       []DatasetCase `json:"cases"`
}

type DatasetCase struct {
	ID               string `json:"id"`
	Profile          string `json:"profile"`
	Name             string `json:"name"`
	Input            string `json:"input"`
	ExpectedBehavior string `json:"expected_behavior"`
}

type Runner struct {
	DefaultModel runtimedomain.ModelClient
	Dataset      *Dataset
}

func NewRunner(model runtimedomain.ModelClient) *Runner {
	if model == nil {
		model = runtimeinfra.NewFakeModelClient()
	}
	return &Runner{DefaultModel: model}
}

func LoadDataset(path string) (*Dataset, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var dataset Dataset
	if err := json.Unmarshal(raw, &dataset); err != nil {
		return nil, err
	}
	return &dataset, nil
}

func (r *Runner) Run(ctx context.Context) []Result {
	cases := r.cases()
	results := make([]Result, 0, len(cases))
	for _, tc := range cases {
		results = append(results, r.runCase(ctx, tc))
	}
	return results
}

func (r *Runner) cases() []evalCase {
	if r.Dataset == nil || len(r.Dataset.Cases) == 0 {
		return defaultCases()
	}
	cases := make([]evalCase, 0, len(r.Dataset.Cases))
	for _, item := range r.Dataset.Cases {
		tc, ok := caseByID(item.ID)
		if !ok {
			continue
		}
		if strings.TrimSpace(item.Name) != "" {
			tc.name = item.Name
		}
		if strings.TrimSpace(item.Profile) != "" {
			tc.profile = item.Profile
		}
		if strings.TrimSpace(item.Input) != "" && item.Input != "malicious model output" {
			tc.input = item.Input
		}
		if strings.TrimSpace(item.ExpectedBehavior) != "" {
			tc.expected = item.ExpectedBehavior
		}
		cases = append(cases, tc)
	}
	if len(cases) == 0 {
		return defaultCases()
	}
	return cases
}

func defaultCases() []evalCase {
	return []evalCase{
		createActivityCase(),
		createReminderCase(),
		createMemoryCase(),
		searchMemoryPrivacyCase(),
		rejectInjectedUserIDCase(),
		rejectUnknownToolCase(),
		rejectExtraFieldsCase(),
		blockMemoryDisabledCase(),
		blockReminderDisabledCase(),
		blockPlanningDisabledCase(),
		promptInjectionBasicCase(),
		noAutomaticExecutionCase(),
	}
}

func caseByID(id string) (evalCase, bool) {
	switch strings.TrimSpace(id) {
	case "create_activity":
		return createActivityCase(), true
	case "create_reminder":
		return createReminderCase(), true
	case "create_memory":
		return createMemoryCase(), true
	case "memory_privacy":
		return searchMemoryPrivacyCase(), true
	case "reject_user_id":
		return rejectInjectedUserIDCase(), true
	case "reject_unknown_tool":
		return rejectUnknownToolCase(), true
	case "reject_extra_fields":
		return rejectExtraFieldsCase(), true
	case "memory_disabled":
		return blockMemoryDisabledCase(), true
	case "reminders_disabled":
		return blockReminderDisabledCase(), true
	case "planning_disabled":
		return blockPlanningDisabledCase(), true
	case "prompt_injection_basic":
		return promptInjectionBasicCase(), true
	case "no_auto_execute":
		return noAutomaticExecutionCase(), true
	default:
		return evalCase{}, false
	}
}

func (r *Runner) runCase(ctx context.Context, tc evalCase) Result {
	started := time.Now()
	env := newEnv()
	model := r.DefaultModel
	if tc.model != nil {
		model = tc.model
	}
	env.runtime = env.newRuntime(model)
	if tc.setup != nil {
		if err := tc.setup(ctx, env); err != nil {
			return tc.fail(fmt.Sprintf("setup failed: %v", err), time.Since(started), nil)
		}
	}

	response, err := env.runtime.HandleMessage(ctx, runtimedomain.RuntimeRequest{
		UserID:  evalUserID,
		Message: tc.input,
		DryRun:  tc.dryRun,
	})
	if err != nil {
		return tc.fail(fmt.Sprintf("runtime error: %v", err), time.Since(started), nil)
	}
	passed, actual, reason := tc.assert(ctx, env, response)
	return Result{
		Case:                 tc.name,
		Profile:              defaultProfile(tc.profile),
		Input:                redact(tc.input),
		ExpectedBehavior:     tc.expected,
		ActualBehavior:       actual,
		Passed:               passed,
		Reason:               reason,
		LatencyMS:            time.Since(started).Milliseconds(),
		RejectedActionsCount: response.Observability.RejectedActionsCount,
		ProposedActionsCount: len(response.ProposedActions),
		SchemaFailuresCount:  schemaFailures(tc, response),
		PolicyFailuresCount:  policyFailures(tc, response),
	}
}

func RenderMarkdown(provider string, results []Result) string {
	var b strings.Builder
	b.WriteString("# AI Eval Report\n\n")
	b.WriteString("Generated by `cmd/tools/run-ai-evals`.\n\n")
	b.WriteString(fmt.Sprintf("Provider: `%s`\n\n", escape(provider)))
	b.WriteString(renderSummary(results))
	b.WriteString("\n")
	b.WriteString("| Profile | Case | Input | Expected behavior | Actual behavior | Result | Latency ms | Proposed | Rejected | Schema failures | Policy failures | Reason |\n")
	b.WriteString("| --- | --- | --- | --- | --- | --- | ---: | ---: | ---: | ---: | ---: | --- |\n")
	for _, result := range results {
		status := "PASS"
		if !result.Passed {
			status = "FAIL"
		}
		b.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s | %s | %d | %d | %d | %d | %d | %s |\n",
			escape(result.Profile),
			escape(result.Case),
			escape(result.Input),
			escape(result.ExpectedBehavior),
			escape(result.ActualBehavior),
			status,
			result.LatencyMS,
			result.ProposedActionsCount,
			result.RejectedActionsCount,
			result.SchemaFailuresCount,
			result.PolicyFailuresCount,
			escape(result.Reason),
		))
	}
	return b.String()
}

func NewReport(provider, model, status, reason string, results []Result) Report {
	summary := summarize(results)
	copiedResults := append([]Result(nil), results...)
	if copiedResults == nil {
		copiedResults = []Result{}
	}
	return Report{
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		Provider: ProviderMeta{
			Name:      strings.TrimSpace(provider),
			Model:     strings.TrimSpace(model),
			LatencyMS: totalLatency(results),
			Status:    strings.TrimSpace(status),
		},
		Status:  strings.TrimSpace(status),
		Reason:  strings.TrimSpace(reason),
		Summary: summary,
		Results: copiedResults,
		Privacy: PrivacyMeta{
			PromptsRedacted:       true,
			FullPromptsStored:     false,
			SensitiveValuesStored: false,
			Note:                  "Reports store redacted/truncated eval inputs only; full prompts, context payloads, raw model responses, API keys, tokens and secrets are excluded.",
		},
	}
}

func RenderJSON(report Report) ([]byte, error) {
	return json.MarshalIndent(report, "", "  ")
}

func renderSummary(results []Result) string {
	summary := summarize(results)
	return fmt.Sprintf("Summary: `%d passed`, `%d failed`, `avg_latency_ms=%d`, `proposed_actions=%d`, `rejected_actions=%d`, `schema_failures=%d`, `policy_failures=%d`\n",
		summary.Passed,
		summary.Failed,
		summary.AvgLatencyMS,
		summary.ProposedActionsCount,
		summary.RejectedActionsCount,
		summary.SchemaFailuresCount,
		summary.PolicyFailuresCount)
}

func summarize(results []Result) ReportSummary {
	var passed, failed, proposed, rejected, schemaFailures, policyFailures int
	var latency int64
	for _, result := range results {
		if result.Passed {
			passed++
		} else {
			failed++
		}
		latency += result.LatencyMS
		proposed += result.ProposedActionsCount
		rejected += result.RejectedActionsCount
		schemaFailures += result.SchemaFailuresCount
		policyFailures += result.PolicyFailuresCount
	}
	avgLatency := int64(0)
	if len(results) > 0 {
		avgLatency = latency / int64(len(results))
	}
	return ReportSummary{
		Passed:               passed,
		Failed:               failed,
		AvgLatencyMS:         avgLatency,
		ProposedActionsCount: proposed,
		RejectedActionsCount: rejected,
		SchemaFailuresCount:  schemaFailures,
		PolicyFailuresCount:  policyFailures,
	}
}

func totalLatency(results []Result) int64 {
	var latency int64
	for _, result := range results {
		latency += result.LatencyMS
	}
	return latency
}

func AllPassed(results []Result) bool {
	for _, result := range results {
		if !result.Passed {
			return false
		}
	}
	return true
}

func escape(value string) string {
	value = strings.ReplaceAll(value, "\n", " ")
	value = strings.ReplaceAll(value, "|", "\\|")
	return strings.TrimSpace(value)
}

func redact(value string) string {
	value = strings.TrimSpace(value)
	if strings.Contains(value, "@") {
		return "[REDACTED]"
	}
	if len([]rune(value)) > 140 {
		return string([]rune(value)[:140]) + "...[TRUNCATED]"
	}
	return value
}

type evalCase struct {
	name     string
	profile  string
	input    string
	expected string
	dryRun   bool
	model    runtimedomain.ModelClient
	setup    func(context.Context, *env) error
	assert   func(context.Context, *env, *runtimedomain.RuntimeResponse) (bool, string, string)
}

func (tc evalCase) fail(reason string, elapsed time.Duration, response *runtimedomain.RuntimeResponse) Result {
	result := Result{Case: tc.name, Profile: defaultProfile(tc.profile), Input: redact(tc.input), ExpectedBehavior: tc.expected, ActualBehavior: "error", Passed: false, Reason: reason, LatencyMS: elapsed.Milliseconds()}
	if response != nil {
		result.RejectedActionsCount = response.Observability.RejectedActionsCount
		result.ProposedActionsCount = len(response.ProposedActions)
		result.SchemaFailuresCount = schemaFailures(tc, response)
		result.PolicyFailuresCount = policyFailures(tc, response)
	}
	return result
}

func defaultProfile(profile string) string {
	if strings.TrimSpace(profile) == "" {
		return "safety"
	}
	return strings.TrimSpace(profile)
}

func schemaFailures(tc evalCase, response *runtimedomain.RuntimeResponse) int {
	if strings.Contains(strings.ToLower(tc.name), "schema") || strings.Contains(strings.ToLower(tc.name), "campos extra") {
		return response.Observability.RejectedActionsCount
	}
	return 0
}

func policyFailures(tc evalCase, response *runtimedomain.RuntimeResponse) int {
	name := strings.ToLower(tc.name)
	if strings.Contains(name, "user_id") || strings.Contains(name, "tool inexistente") || strings.Contains(name, "injection") ||
		strings.Contains(name, "disabled") || strings.Contains(name, "bloquear") {
		return response.Observability.RejectedActionsCount
	}
	return 0
}

type env struct {
	users      *usersapp.Service
	activities *activitiesapp.Service
	reminders  *remindersapp.Service
	insights   *insightsapp.Service
	memories   *memoryapp.Service
	actions    *actionsapp.Service
	tools      *toolsapp.Service
	runtime    *runtimeapp.RuntimeService
}

func newEnv() *env {
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
	if err := tools.SeedDefaultTools(context.Background()); err != nil {
		panic(err)
	}
	actions := actionsapp.NewService(actionRepo, tools, users, activities, reminders, memories)
	activities.SetReminderBridge(reminders)

	user := authdomain.NewUser(evalUserID, "Eval User", "eval@example.com", "hash")
	if err := authRepo.Save(user); err != nil {
		panic(err)
	}
	if err := users.ProvisionDefaults(user); err != nil {
		panic(err)
	}

	e := &env{users: users, activities: activities, reminders: reminders, insights: insights, memories: memories, actions: actions, tools: tools}
	e.runtime = e.newRuntime(runtimeinfra.NewFakeModelClient())
	return e
}

func (e *env) newRuntime(model runtimedomain.ModelClient) *runtimeapp.RuntimeService {
	contextBuilder := runtimeapp.NewContextBuilder(e.users, e.activities, e.reminders, e.insights, e.memories, 5)
	toolSelector := runtimeapp.NewToolSelector(e.tools)
	planner := runtimeapp.NewPlanner(model)
	return runtimeapp.NewRuntimeService(contextBuilder, toolSelector, planner, runtimeapp.NewSafetyPolicy(), e.actions)
}
