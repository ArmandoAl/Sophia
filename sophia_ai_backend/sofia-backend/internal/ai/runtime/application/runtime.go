package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode"

	activitiesdomain "github.com/armandoalvarado/sofia-backend/internal/activities/domain"
	actionsapp "github.com/armandoalvarado/sofia-backend/internal/ai/actions/application"
	actionsdomain "github.com/armandoalvarado/sofia-backend/internal/ai/actions/domain"
	"github.com/armandoalvarado/sofia-backend/internal/ai/runtime/domain"
	insightsdomain "github.com/armandoalvarado/sofia-backend/internal/insights/domain"
	learningdomain "github.com/armandoalvarado/sofia-backend/internal/learning/domain"
	memorydomain "github.com/armandoalvarado/sofia-backend/internal/memory/domain"
	remindersdomain "github.com/armandoalvarado/sofia-backend/internal/reminders/domain"
	toolsdomain "github.com/armandoalvarado/sofia-backend/internal/tools/domain"
	usersapp "github.com/armandoalvarado/sofia-backend/internal/users/application"
	usersdomain "github.com/armandoalvarado/sofia-backend/internal/users/domain"
	"github.com/google/uuid"
)

const (
	DefaultContextLimit = 5

	tokenQuotaMemoriesPercent   = 40
	tokenQuotaHistoryPercent    = 25
	tokenQuotaActivitiesPercent = 25
	tokenQuotaRemindersPercent  = 10
)

var emailRedactPattern = regexp.MustCompile(`[\w.+-]+@[\w-]+\.[\w.-]+`)

type UserReader interface {
	GetMe(userID string) (*usersapp.Me, error)
}

type ActivityLister interface {
	ListActivities(ctx context.Context, filter activitiesdomain.ListFilter) ([]*activitiesdomain.Activity, error)
}

type ReminderDueLister interface {
	ListDueReminders(ctx context.Context, userID string, now time.Time, limit int) ([]*remindersdomain.Reminder, error)
}

type InsightsSummarizer interface {
	GetInsightsSummary(ctx context.Context, filter insightsdomain.ListFilter) (*insightsdomain.Summary, error)
}

type MemorySearcher interface {
	SearchByTerms(ctx context.Context, userID string, terms []string, limit int) ([]*memorydomain.Memory, error)
}

type ToolLister interface {
	ListTools(ctx context.Context) ([]*toolsdomain.ToolDefinition, error)
}

type ActionProposalCreator interface {
	CreateActionProposal(ctx context.Context, userID string, input actionsdomain.ProposalCreate) (*actionsdomain.AIActionProposal, error)
	AutoExecuteActionProposal(ctx context.Context, userID, proposalID string) (*actionsdomain.AIActionProposal, error)
}

type AutonomyEvaluator interface {
	EvaluateAutonomy(ctx context.Context, userID, toolName string) (actionsapp.AutonomyDecision, error)
}

type AuditRecorder interface {
	RecordAuditLog(ctx context.Context, userID, action, resourceType, resourceID string, metadata map[string]any) error
}

type PromptBaseReader interface {
	ActivePromptContent(ctx context.Context, userID string) (string, error)
}

type LearningContextReader interface {
	ListUserContexts(ctx context.Context, userID string) ([]*learningdomain.UserContext, error)
	FindUserContextByScopeKey(ctx context.Context, userID, scopeKey string) (*learningdomain.UserContext, error)
	ContextBeliefStatements(ctx context.Context, userID, scopeKey string) ([]string, error)
}

type ContextBuilder struct {
	users      UserReader
	activities ActivityLister
	reminders  ReminderDueLister
	insights   InsightsSummarizer
	memories   MemorySearcher
	prompts    PromptBaseReader
	contexts   LearningContextReader
	limit      int
	tokenLimit int
}

func NewContextBuilder(users UserReader, activities ActivityLister, reminders ReminderDueLister, insights InsightsSummarizer, memories MemorySearcher, limit, tokenLimit int) *ContextBuilder {
	if limit <= 0 {
		limit = DefaultContextLimit
	}
	return &ContextBuilder{users: users, activities: activities, reminders: reminders, insights: insights, memories: memories, limit: limit, tokenLimit: tokenLimit}
}

func (b *ContextBuilder) SetPromptBaseReader(reader PromptBaseReader) {
	b.prompts = reader
}

func (b *ContextBuilder) SetLearningContextReader(reader LearningContextReader) {
	b.contexts = reader
}

func (b *ContextBuilder) Build(ctx context.Context, userID, message string, activeContexts ...string) (domain.ContextSummary, error) {
	activeContext := ""
	if len(activeContexts) > 0 {
		activeContext = activeContexts[0]
	}
	me, err := b.users.GetMe(userID)
	if err != nil {
		return domain.ContextSummary{}, err
	}
	now := time.Now().UTC()
	if me.Profile != nil && strings.TrimSpace(me.Profile.Timezone) != "" {
		if location, err := time.LoadLocation(me.Profile.Timezone); err == nil {
			now = now.In(location)
		}
	}

	summary := domain.ContextSummary{
		User: domain.UserSummary{
			ID:    me.User.ID,
			Email: "[REDACTED]",
			Role:  me.User.Role,
		},
		Limits: map[string]int{
			"recent_activities": b.limit,
			"due_reminders":     b.limit,
			"relevant_memories": b.limit,
		},
		CurrentDateTime: now.Format(time.RFC3339),
		TokenBudget:     domain.TokenBudget{MaxApproxTokens: b.tokenLimit},
	}
	if me.Profile != nil {
		summary.Profile = domain.ProfileSummary{
			DisplayName:   me.Profile.DisplayName,
			PreferredName: me.Profile.PreferredName,
			Timezone:      me.Profile.Timezone,
			Locale:        me.Profile.Locale,
		}
	}
	if me.AISettings != nil {
		summary.AISettings = domain.AISettingsSummary{
			AssistantName:     me.AISettings.AssistantName,
			AutonomyLevel:     me.AISettings.AutonomyLevel,
			MemoryEnabled:     me.AISettings.MemoryEnabled,
			RemindersEnabled:  me.AISettings.RemindersEnabled,
			PlanningEnabled:   me.AISettings.PlanningEnabled,
			PreferredLanguage: me.AISettings.PreferredLanguage,
		}
	}

	activities, err := b.activities.ListActivities(ctx, activitiesdomain.ListFilter{UserID: userID, Limit: b.limit})
	if err != nil {
		return domain.ContextSummary{}, err
	}
	contextValue, err := b.resolveActiveContext(ctx, userID, activeContext)
	if err != nil {
		return domain.ContextSummary{}, err
	}
	if contextValue != nil {
		beliefs, err := b.contexts.ContextBeliefStatements(ctx, userID, contextValue.ScopeKey())
		if err != nil {
			return domain.ContextSummary{}, err
		}
		summary.ActiveContext = &domain.ActiveContext{ScopeKey: contextValue.ScopeKey(), Label: contextValue.Label, Beliefs: beliefs}
	}

	var due []*remindersdomain.Reminder
	if summary.AISettings.RemindersEnabled {
		due, err = b.reminders.ListDueReminders(ctx, userID, time.Now(), b.limit)
		if err != nil {
			return domain.ContextSummary{}, err
		}
		summary.RemindersIncluded = true
	}

	insights, err := b.insights.GetInsightsSummary(ctx, insightsdomain.ListFilter{UserID: userID, Limit: b.limit * 4})
	if err != nil {
		return domain.ContextSummary{}, err
	}
	summary.InsightsSummary = map[string]any{
		"average_energy":    insights.AverageEnergy,
		"average_stress":    insights.AverageStress,
		"average_focus":     insights.AverageFocus,
		"total_outcomes":    insights.TotalOutcomes,
		"total_reflections": insights.TotalReflections,
	}

	maxTokens := summary.TokenBudget.MaxApproxTokens
	memoriesQuota := quotaTokens(maxTokens, tokenQuotaMemoriesPercent)
	historyQuota := quotaTokens(maxTokens, tokenQuotaHistoryPercent)
	activitiesQuota := quotaTokens(maxTokens, tokenQuotaActivitiesPercent)
	remindersQuota := maxTokens - memoriesQuota - historyQuota - activitiesQuota

	leftover := chargeBudget(&summary.TokenBudget, approximateJSONTokens(summary.InsightsSummary), memoriesQuota)

	if summary.AISettings.MemoryEnabled {
		terms := memorydomain.ExtractTerms(message, 12)
		if len(terms) > 0 {
			memories, err := b.memories.SearchByTerms(ctx, userID, terms, b.limit)
			if err != nil {
				return domain.ContextSummary{}, err
			}
			var remaining int
			summary.RelevantMemories, remaining = fitItems(summarizeMemories(memories), &summary.TokenBudget, leftover)
			leftover = remaining
		}
		summary.MemoryIncluded = true
	}

	historyReserve := historyQuota + leftover
	activityCap := minInt(activitiesQuota, remainingAfterReserve(&summary.TokenBudget, historyReserve))
	summary.RecentActivities, leftover = fitItems(summarizeActivities(activities), &summary.TokenBudget, activityCap)
	if summary.RemindersIncluded {
		reminderCap := minInt(remindersQuota+leftover, remainingAfterReserve(&summary.TokenBudget, historyReserve))
		summary.DueReminders, _ = fitItems(summarizeReminders(due), &summary.TokenBudget, reminderCap)
	}

	if b.prompts != nil {
		content, err := b.prompts.ActivePromptContent(ctx, userID)
		if err != nil {
			return domain.ContextSummary{}, err
		}
		summary.PromptBase = content
	}

	return summary, nil
}

func (b *ContextBuilder) resolveActiveContext(ctx context.Context, userID, explicit string) (*learningdomain.UserContext, error) {
	if b.contexts == nil {
		return nil, nil
	}
	if strings.TrimSpace(explicit) != "" {
		return b.contexts.FindUserContextByScopeKey(ctx, userID, explicit)
	}
	contexts, err := b.contexts.ListUserContexts(ctx, userID)
	if err != nil {
		return nil, err
	}
	activities, err := b.activities.ListActivities(ctx, activitiesdomain.ListFilter{UserID: userID, Status: activitiesdomain.StatusActive, Limit: b.limit})
	if err != nil {
		return nil, err
	}
	for _, activity := range activities {
		for _, contextValue := range contexts {
			if contextValue != nil && contextValue.Active && activityMentionsContext(activity, contextValue) {
				return contextValue, nil
			}
		}
	}
	// Model-inferred context is intentionally deferred until it can be resolved safely.
	return nil, nil
}

func activityMentionsContext(activity *activitiesdomain.Activity, contextValue *learningdomain.UserContext) bool {
	values := append([]string{contextValue.Slug}, contextValue.Aliases...)
	haystack := normalizeMention(strings.Join(append([]string{activity.Title}, activity.Tags...), " "))
	for _, value := range values {
		if strings.Contains(" "+haystack+" ", " "+normalizeMention(value)+" ") {
			return true
		}
	}
	return false
}

func normalizeMention(value string) string {
	return strings.Join(strings.FieldsFunc(strings.ToLower(value), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	}), " ")
}

type ToolSelector struct {
	tools  ToolLister
	policy *PolicyEngine
}

func NewToolSelector(tools ToolLister) *ToolSelector {
	return &ToolSelector{tools: tools, policy: NewPolicyEngine()}
}

func (s *ToolSelector) SelectTools(ctx context.Context, context domain.ContextSummary, _ string) ([]domain.ToolSummary, error) {
	tools, err := s.tools.ListTools(ctx)
	if err != nil {
		return nil, err
	}
	selected := make([]domain.ToolSummary, 0, len(tools))
	for _, tool := range tools {
		if tool == nil || !tool.Enabled {
			continue
		}
		risk := riskForTool(tool)
		if !s.policy.AllowTool(context, tool, risk) {
			continue
		}
		selected = append(selected, domain.ToolSummary{
			Name:                 tool.Name,
			Category:             tool.Category,
			RequiresConfirmation: tool.RequiresConfirmation,
			RiskLevel:            risk,
			InputSchema:          cloneJSON(tool.InputSchema),
		})
	}
	return selected, nil
}

type SafetyPolicy struct {
	engine    *PolicyEngine
	autonomy  AutonomyEvaluator
	threshold float64
}

func NewSafetyPolicy() *SafetyPolicy {
	return &SafetyPolicy{engine: NewPolicyEngine(), threshold: learningdomain.DefaultAutonomyThreshold}
}

func (p *SafetyPolicy) SetAutonomy(eval AutonomyEvaluator, threshold float64) {
	p.autonomy = eval
	if threshold > 0 && threshold <= 1 {
		p.threshold = threshold
	}
}

func (p *SafetyPolicy) ValidateProposal(_ context.Context, userID string, action domain.PlannedAction, context domain.ContextSummary, tools []domain.ToolSummary) error {
	tool, ok := findTool(tools, action.ToolName)
	if !ok {
		return domain.ErrUnsafeActionProposal
	}
	if !p.engine.AllowProposal(strings.TrimSpace(userID), action, context, tool) {
		return domain.ErrUnsafeActionProposal
	}
	return nil
}

func (p *SafetyPolicy) CanExecute(ctx context.Context, userID string, action domain.PlannedAction, context domain.ContextSummary) error {
	if p.autonomy == nil {
		return domain.ErrUnsafeActionProposal
	}
	if context.AISettings.AutonomyLevel != usersdomain.AutonomySemiAutonomous {
		return domain.ErrUnsafeActionProposal
	}
	decision, err := p.autonomy.EvaluateAutonomy(ctx, userID, action.ToolName)
	if err != nil {
		return domain.ErrUnsafeActionProposal
	}
	if !learningdomain.ShouldAutoExecute(decision.Probability, decision.History, decision.Reversible, p.threshold) {
		return domain.ErrUnsafeActionProposal
	}
	return nil
}

type Planner struct {
	model domain.ModelClient
}

func NewPlanner(model domain.ModelClient) *Planner {
	return &Planner{model: model}
}

func (p *Planner) Plan(ctx context.Context, request domain.ModelRequest) (domain.ModelResponse, error) {
	return p.model.Generate(ctx, request)
}

type RuntimeService struct {
	contextBuilder domain.ContextBuilder
	toolSelector   domain.ToolSelector
	planner        domain.Planner
	safety         domain.SafetyPolicy
	actions        ActionProposalCreator
	audit          AuditRecorder
}

func NewRuntimeService(contextBuilder domain.ContextBuilder, toolSelector domain.ToolSelector, planner domain.Planner, safety domain.SafetyPolicy, actions ActionProposalCreator) *RuntimeService {
	return &RuntimeService{contextBuilder: contextBuilder, toolSelector: toolSelector, planner: planner, safety: safety, actions: actions}
}

func (s *RuntimeService) SetAuditRecorder(audit AuditRecorder) {
	s.audit = audit
}

func (s *RuntimeService) HandleMessage(ctx context.Context, request domain.RuntimeRequest) (*domain.RuntimeResponse, error) {
	if strings.TrimSpace(request.RequestID) == "" {
		request.RequestID = uuid.New().String()
	}
	request.Message = strings.TrimSpace(request.Message)
	if request.Message == "" {
		return nil, domain.ErrInvalidMessage
	}

	contextSummary, err := s.contextBuilder.Build(ctx, request.UserID, request.Message, request.ActiveContext)
	if err != nil {
		return nil, err
	}
	s.recordAudit(ctx, request.UserID, "ai_runtime_context_built", "ai_runtime", request.UserID, map[string]any{
		"request_id":         request.RequestID,
		"approx_tokens":      contextSummary.TokenBudget.UsedApproxTokens,
		"recent_activities":  len(contextSummary.RecentActivities),
		"due_reminders":      len(contextSummary.DueReminders),
		"relevant_memories":  len(contextSummary.RelevantMemories),
		"memory_included":    contextSummary.MemoryIncluded,
		"reminders_included": contextSummary.RemindersIncluded,
	})
	tools, err := s.toolSelector.SelectTools(ctx, contextSummary, request.Message)
	if err != nil {
		return nil, err
	}
	s.recordAudit(ctx, request.UserID, "ai_runtime_model_request_planned", "ai_runtime", request.UserID, map[string]any{
		"request_id":   request.RequestID,
		"tool_count":   len(tools),
		"dry_run":      request.DryRun,
		"message_hash": messageFingerprint(request.Message),
	})
	providerStarted := time.Now()
	history := fitHistory(request.History, &contextSummary.TokenBudget)
	modelResponse, err := s.planner.Plan(ctx, domain.ModelRequest{
		UserID:     request.UserID,
		Message:    request.Message,
		Context:    contextSummary,
		Tools:      tools,
		History:    history,
		PromptBase: contextSummary.PromptBase,
		Task:       domain.TaskPlan,
	})
	providerLatency := time.Since(providerStarted)
	if err != nil {
		s.recordAudit(ctx, request.UserID, "ai_runtime_provider_failed", "ai_runtime", request.UserID, map[string]any{
			"request_id":          request.RequestID,
			"provider_latency_ms": providerLatency.Milliseconds(),
			"provider_error_type": providerErrorType(err),
			"planned_actions":     0,
			"rejected_actions":    0,
			"persisted_proposals": 0,
			"dry_run":             request.DryRun,
		})
		return nil, err
	}
	s.recordAudit(ctx, request.UserID, "ai_runtime_proposals_generated", "ai_runtime", request.UserID, map[string]any{
		"request_id":          request.RequestID,
		"provider_latency_ms": modelLatencyMillis(providerLatency),
		"input_tokens":        modelResponse.Usage.InputTokens,
		"output_tokens":       modelResponse.Usage.OutputTokens,
		"cached_input_tokens": modelResponse.Usage.CachedInputTokens,
		"model":               modelResponse.Usage.Model,
		"planned_actions":     len(modelResponse.PlannedActions),
		"dry_run":             request.DryRun,
	})

	response := &domain.RuntimeResponse{
		Mode:             domain.ModeDryRun,
		RequestID:        request.RequestID,
		ContextSummary:   contextSummary,
		AvailableTools:   tools,
		AssistantMessage: modelResponse.AssistantMessage,
		Observability: domain.RuntimeObservability{
			ProviderLatencyMS:   modelLatencyMillis(providerLatency),
			PlannedActionsCount: len(modelResponse.PlannedActions),
		},
	}
	if !request.DryRun {
		response.Mode = domain.ModeReady
	}

	for _, action := range modelResponse.PlannedActions {
		if err := s.safety.ValidateProposal(ctx, request.UserID, action, contextSummary, tools); err != nil {
			response.Observability.RejectedActionsCount++
			continue
		}
		output := domain.ActionProposalOutput{
			ToolName:      action.ToolName,
			Status:        actionsdomain.StatusProposed,
			ProposedInput: cloneJSON(action.ProposedInput),
			Reason:        action.Reason,
			RiskLevel:     defaultRisk(action.RiskLevel),
		}
		if tool, ok := findTool(tools, action.ToolName); ok {
			output.RequiresConfirmation = tool.RequiresConfirmation
		}
		if !request.DryRun {
			proposal, err := s.actions.CreateActionProposal(ctx, request.UserID, actionsdomain.ProposalCreate{
				ToolName:       action.ToolName,
				ProposedInput:  action.ProposedInput,
				Reason:         action.Reason,
				RiskLevel:      output.RiskLevel,
				ConversationID: request.ConversationID,
			})
			if err != nil {
				response.Observability.RejectedActionsCount++
				continue
			}
			output.ID = proposal.ID
			output.Status = proposal.Status
			output.RequiresConfirmation = proposal.RequiresConfirmation
			if err := s.safety.CanExecute(ctx, request.UserID, action, contextSummary); err == nil {
				executed, execErr := s.actions.AutoExecuteActionProposal(ctx, request.UserID, proposal.ID)
				if execErr == nil && executed != nil {
					output.ID = executed.ID
					output.Status = executed.Status
					output.RequiresConfirmation = executed.RequiresConfirmation
				}
			}
		}
		response.ProposedActions = append(response.ProposedActions, output)
	}
	response.Observability.ProposalsCount = len(response.ProposedActions)
	s.recordAudit(ctx, request.UserID, "ai_runtime_proposals_evaluated", "ai_runtime", request.UserID, map[string]any{
		"request_id":          request.RequestID,
		"planned_actions":     response.Observability.PlannedActionsCount,
		"persisted_proposals": response.Observability.ProposalsCount,
		"rejected_actions":    response.Observability.RejectedActionsCount,
		"dry_run":             request.DryRun,
	})

	return response, nil
}

func (s *RuntimeService) recordAudit(ctx context.Context, userID, action, resourceType, resourceID string, metadata map[string]any) {
	if s.audit == nil {
		return
	}
	_ = s.audit.RecordAuditLog(ctx, userID, action, resourceType, resourceID, metadata)
}

func modelLatencyMillis(duration time.Duration) int64 {
	if duration < 0 {
		return 0
	}
	return duration.Milliseconds()
}

func providerErrorType(err error) string {
	if err == nil {
		return ""
	}
	var typed interface{ ProviderErrorType() string }
	if errors.As(err, &typed) && typed.ProviderErrorType() != "" {
		return typed.ProviderErrorType()
	}
	return fmt.Sprintf("%T", err)
}

func messageFingerprint(message string) string {
	message = strings.TrimSpace(message)
	if message == "" {
		return ""
	}
	return fmt.Sprintf("len:%d", len([]rune(message)))
}

func summarizeActivities(activities []*activitiesdomain.Activity) []domain.ItemSummary {
	items := make([]domain.ItemSummary, 0, len(activities))
	for _, activity := range activities {
		items = append(items, domain.ItemSummary{ID: activity.ID, Type: activity.Type, Title: activity.Title, Status: activity.Status, Tags: activity.Tags})
	}
	return items
}

func summarizeReminders(reminders []*remindersdomain.Reminder) []domain.ItemSummary {
	items := make([]domain.ItemSummary, 0, len(reminders))
	for _, reminder := range reminders {
		items = append(items, domain.ItemSummary{ID: reminder.ID, Title: reminder.Title, Status: reminder.Status})
	}
	return items
}

func summarizeMemories(memories []*memorydomain.Memory) []domain.ItemSummary {
	sort.SliceStable(memories, func(i, j int) bool {
		if importanceRank(memories[i].Importance) == importanceRank(memories[j].Importance) {
			return memories[i].UpdatedAt.After(memories[j].UpdatedAt)
		}
		return importanceRank(memories[i].Importance) > importanceRank(memories[j].Importance)
	})
	items := make([]domain.ItemSummary, 0, len(memories))
	for _, memory := range memories {
		if memory.Visibility != memorydomain.VisibilityAssistantContext || memory.Status != memorydomain.StatusActive {
			continue
		}
		title := memory.Title
		if memory.Summary != "" {
			title = memory.Summary
		}
		items = append(items, domain.ItemSummary{ID: memory.ID, Type: memory.Type, Title: redactText(title), Status: memory.Status, Importance: memory.Importance, Tags: memory.Tags})
	}
	return items
}

func importanceRank(value string) int {
	switch value {
	case memorydomain.ImportanceCritical:
		return 4
	case memorydomain.ImportanceHigh:
		return 3
	case memorydomain.ImportanceMedium:
		return 2
	case memorydomain.ImportanceLow:
		return 1
	default:
		return 0
	}
}

func fitItems(items []domain.ItemSummary, budget *domain.TokenBudget, cap int) ([]domain.ItemSummary, int) {
	if budget == nil || budget.MaxApproxTokens <= 0 || cap <= 0 {
		return []domain.ItemSummary{}, 0
	}
	remaining := cap
	if maxLeft := budget.MaxApproxTokens - budget.UsedApproxTokens; remaining > maxLeft {
		remaining = maxLeft
	}
	if remaining <= 0 {
		return []domain.ItemSummary{}, 0
	}
	result := make([]domain.ItemSummary, 0, len(items))
	for _, item := range items {
		item.Title = redactText(item.Title)
		cost := approximateTokens(item.Title) + approximateTokens(item.Type) + approximateTokens(strings.Join(item.Tags, " ")) + 4
		if cost > remaining {
			break
		}
		remaining -= cost
		budget.UsedApproxTokens += cost
		result = append(result, item)
	}
	return result, remaining
}

func fitHistory(turns []domain.Turn, budget *domain.TokenBudget) []domain.Turn {
	if len(turns) == 0 {
		return []domain.Turn{}
	}
	available := 0
	if budget != nil {
		available = budget.MaxApproxTokens - budget.UsedApproxTokens
		if available < 0 {
			available = 0
		}
	}
	start := len(turns)
	used := 0
	for i := len(turns) - 1; i >= 0; i-- {
		cost := approximateTokens(turns[i].Role) + approximateTokens(turns[i].Content) + 4
		if used+cost > available {
			break
		}
		used += cost
		start = i
	}
	if budget != nil {
		budget.UsedApproxTokens += used
	}
	result := make([]domain.Turn, len(turns)-start)
	copy(result, turns[start:])
	return result
}

func quotaTokens(max, percent int) int {
	if max <= 0 || percent <= 0 {
		return 0
	}
	return max * percent / 100
}

func chargeBudget(budget *domain.TokenBudget, cost, cap int) int {
	if budget == nil {
		return 0
	}
	if cost < 0 {
		cost = 0
	}
	budget.UsedApproxTokens += cost
	leftover := cap - cost
	if leftover < 0 {
		return 0
	}
	return leftover
}

func remainingAfterReserve(budget *domain.TokenBudget, reserve int) int {
	if budget == nil {
		return 0
	}
	left := budget.MaxApproxTokens - budget.UsedApproxTokens - reserve
	if left < 0 {
		return 0
	}
	return left
}

func approximateJSONTokens(value any) int {
	raw, err := json.Marshal(value)
	if err != nil {
		return 0
	}
	return approximateTokens(string(raw))
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func approximateTokens(text string) int {
	text = strings.TrimSpace(text)
	if text == "" {
		return 0
	}
	return len([]rune(text))/4 + 1
}

func redactText(text string) string {
	text = strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(text, "\n", " "), "\t", " "))
	text = emailRedactPattern.ReplaceAllString(text, "[email]")
	if len([]rune(text)) > 160 {
		return string([]rune(text)[:160]) + "...[TRUNCATED]"
	}
	return text
}

func riskForTool(tool *toolsdomain.ToolDefinition) string {
	if tool == nil {
		return actionsdomain.RiskMedium
	}
	switch tool.Name {
	case toolsdomain.ToolSearchMemory:
		return actionsdomain.RiskLow
	case toolsdomain.ToolCreateMemory, toolsdomain.ToolCreateReminder, toolsdomain.ToolCreateActivity, toolsdomain.ToolCreateReflection:
		return actionsdomain.RiskMedium
	default:
		return actionsdomain.RiskMedium
	}
}

func autonomyAllows(userLevel, required string) bool {
	return autonomyRank(userLevel) >= autonomyRank(required)
}

func autonomyRank(level string) int {
	switch level {
	case usersdomain.AutonomyManual:
		return 1
	case usersdomain.AutonomySuggestive, "":
		return 2
	case usersdomain.AutonomySemiAutonomous:
		return 3
	default:
		return 0
	}
}

func findTool(tools []domain.ToolSummary, name string) (domain.ToolSummary, bool) {
	for _, tool := range tools {
		if tool.Name == name {
			return tool, true
		}
	}
	return domain.ToolSummary{}, false
}

func isPlanningTool(tool string) bool {
	switch tool {
	case toolsdomain.ToolCreateActivity, toolsdomain.ToolUpdateActivity, toolsdomain.ToolCompleteActivity:
		return true
	default:
		return false
	}
}

func defaultRisk(value string) string {
	if actionsdomain.IsValidRiskLevel(value) {
		return value
	}
	return actionsdomain.RiskMedium
}

func cloneJSON(value json.RawMessage) json.RawMessage {
	if value == nil {
		return nil
	}
	return append(json.RawMessage(nil), value...)
}
