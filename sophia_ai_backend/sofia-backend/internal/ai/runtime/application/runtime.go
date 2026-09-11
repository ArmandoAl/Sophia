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

	activitiesdomain "github.com/armandoalvarado/sofia-backend/internal/activities/domain"
	actionsapp "github.com/armandoalvarado/sofia-backend/internal/ai/actions/application"
	actionsdomain "github.com/armandoalvarado/sofia-backend/internal/ai/actions/domain"
	"github.com/armandoalvarado/sofia-backend/internal/ai/runtime/domain"
	insightsdomain "github.com/armandoalvarado/sofia-backend/internal/insights/domain"
	learningapp "github.com/armandoalvarado/sofia-backend/internal/learning/application"
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
	tokenQuotaActivitiesPercent = 40
	tokenQuotaRemindersPercent  = 20
	recentStateTokenCeiling     = 600
	conversationHistoryCeiling  = 1000
	carryForwardCharacterLimit  = 800
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
	FindUserContextByScopeKey(ctx context.Context, userID, scopeKey string) (*learningdomain.UserContext, error)
	FindUserContextByID(ctx context.Context, userID, entityID string) (*learningdomain.UserContext, error)
	ContextBeliefStatements(ctx context.Context, userID, scopeKey string) ([]string, error)
	ResolveEntities(ctx context.Context, userID, message string) ([]learningdomain.EntityMatch, error)
}

type EpisodicContextReader interface {
	SearchEpisodes(ctx context.Context, userID, entityID string, terms []string, limit int) ([]*learningdomain.Episode, error)
	GetOpenThreads(ctx context.Context, userID string, limit int) ([]learningapp.OpenThread, error)
	MarkOpenThreadRetaken(ctx context.Context, thread learningapp.OpenThread) error
}

type ContextBuilder struct {
	users      UserReader
	activities ActivityLister
	reminders  ReminderDueLister
	insights   InsightsSummarizer
	memories   MemorySearcher
	prompts    PromptBaseReader
	contexts   LearningContextReader
	extractor  domain.ModelClient
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

func (b *ContextBuilder) SetExtractor(model domain.ModelClient) {
	b.extractor = model
}

func (b *ContextBuilder) Build(ctx context.Context, userID, message string, activeContexts ...string) (domain.ContextSummary, error) {
	state := domain.ContextBuildState{}
	if len(activeContexts) > 0 {
		state.ExplicitActiveContext = activeContexts[0]
	}
	return b.BuildStateful(ctx, userID, message, state)
}

func (b *ContextBuilder) BuildStateful(ctx context.Context, userID, message string, state domain.ContextBuildState) (domain.ContextSummary, error) {
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
	var contextValue *learningdomain.UserContext
	if b.contexts != nil {
		explicit := strings.TrimSpace(state.ExplicitActiveContext)
		if explicit != "" {
			contextValue, err = b.contexts.FindUserContextByScopeKey(ctx, userID, explicit)
		} else if strings.TrimSpace(state.CurrentEntityID) != "" {
			contextValue, err = b.contexts.FindUserContextByID(ctx, userID, state.CurrentEntityID)
		} else {
			contextValue, err = b.resolveActiveContext(ctx, userID, "")
		}
		if err != nil {
			return domain.ContextSummary{}, err
		}
		matches, err := b.contexts.ResolveEntities(ctx, userID, message)
		if err != nil {
			return domain.ContextSummary{}, err
		}
		if explicit == "" {
			if target := automaticContextTarget(matches, message, state.CurrentEntityID); target != nil {
				contextValue = target
			}
		}
		if contextValue != nil {
			summary.ContextChanged = state.CurrentEntityID != "" && state.CurrentEntityID != contextValue.ID
			beliefs, err := b.contexts.ContextBeliefStatements(ctx, userID, contextValue.ScopeKey())
			if err != nil {
				return domain.ContextSummary{}, err
			}
			summary.ActiveContext = &domain.ActiveContext{EntityID: contextValue.ID, ScopeKey: contextValue.ScopeKey(), Label: contextValue.Label, Relationship: contextValue.Relationship, Beliefs: beliefs}
			summary.ActiveEntity = &domain.EntityReference{ID: contextValue.ID, ScopeKey: contextValue.ScopeKey(), Label: contextValue.Label, Relationship: contextValue.Relationship}
			summary.CarryForwardEntity = contextValue.ID
			if summary.ContextChanged {
				summary.CarryForward, err = b.generateCarryForward(ctx, userID, state.History)
				if err != nil {
					return domain.ContextSummary{}, err
				}
			} else if state.CarryForwardEntity == contextValue.ID {
				summary.CarryForward = state.CarryForward
			}
			if episodic, ok := b.contexts.(EpisodicContextReader); ok {
				episodes, err := episodic.SearchEpisodes(ctx, userID, contextValue.ID, memorydomain.ExtractTerms(message, 12), learningdomain.DefaultEpisodeSearchLimit)
				if err != nil {
					return domain.ContextSummary{}, err
				}
				for _, episode := range episodes {
					summary.RecentEpisodes = append(summary.RecentEpisodes, domain.EpisodeSummary{ID: episode.ID, OccurredAt: episode.OccurredAt, Summary: episode.Summary, Topics: episode.Topics, Salience: episode.Salience})
				}
				if !state.OpenThreadRetaken {
					threads, err := episodic.GetOpenThreads(ctx, userID, 1)
					if err != nil {
						return domain.ContextSummary{}, err
					}
					if len(threads) > 0 {
						thread := threads[0]
						summary.OpenThreads = []domain.OpenThreadSummary{{ID: thread.ID, Source: thread.Source, EntityIDs: thread.EntityIDs, Summary: thread.Summary}}
						if err := episodic.MarkOpenThreadRetaken(ctx, thread); err != nil {
							return domain.ContextSummary{}, err
						}
						summary.OpenThreadRetaken = true
					}
				}
			}
		}
		beliefsByEntity := make(map[string][]string)
		for _, match := range matches {
			if contextValue != nil && match.Entity.ID == contextValue.ID && !match.Ambiguous {
				continue
			}
			beliefs, ok := beliefsByEntity[match.Entity.ID]
			if !ok {
				beliefs, err = b.contexts.ContextBeliefStatements(ctx, userID, match.Entity.ScopeKey())
				if err != nil {
					return domain.ContextSummary{}, err
				}
				beliefsByEntity[match.Entity.ID] = beliefs
			}
			summary.EntityContexts = append(summary.EntityContexts, domain.EntityContext{
				EntityID: match.Entity.ID, ScopeKey: match.Entity.ScopeKey(), Label: match.Entity.Label,
				Relationship: match.Entity.Relationship, Mention: match.Mention, Start: match.Start, End: match.End,
				Ambiguous: match.Ambiguous, Beliefs: beliefs,
			})
			if match.Ambiguous {
				summary.EntityInstruction = "Two or more entities match the same mention. Ask the user which person they mean before using candidate facts or making assumptions."
			}
		}
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

	maxTokens := minInt(recentStateTokenCeiling, summary.TokenBudget.MaxApproxTokens)
	memoriesQuota := quotaTokens(maxTokens, tokenQuotaMemoriesPercent)
	activitiesQuota := quotaTokens(maxTokens, tokenQuotaActivitiesPercent)
	remindersQuota := maxTokens - memoriesQuota - activitiesQuota

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

	activityCap := minInt(activitiesQuota, remainingAfterReserve(&summary.TokenBudget, 0))
	summary.RecentActivities, leftover = fitItems(summarizeActivities(activities), &summary.TokenBudget, activityCap)
	if summary.RemindersIncluded {
		reminderCap := minInt(remindersQuota+leftover, remainingAfterReserve(&summary.TokenBudget, 0))
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

func automaticContextTarget(matches []learningdomain.EntityMatch, message, currentEntityID string) *learningdomain.UserContext {
	counts := make(map[string]int)
	entities := make(map[string]*learningdomain.UserContext)
	for _, match := range matches {
		if match.Ambiguous || match.Entity == nil || match.Entity.ID == currentEntityID {
			continue
		}
		counts[match.Entity.ID]++
		entities[match.Entity.ID] = match.Entity
	}
	var target *learningdomain.UserContext
	for entityID, count := range counts {
		if count < 2 && !directedEntityMessage(message) {
			continue
		}
		if target != nil && target.ID != entityID {
			return nil
		}
		target = entities[entityID]
	}
	return target
}

func directedEntityMessage(message string) bool {
	if strings.Contains(message, "?") || strings.Contains(message, "¿") {
		return true
	}
	message = strings.ToLower(strings.TrimSpace(message))
	for _, prefix := range []string{"cuéntame", "cuentame", "dime", "ayúdame", "ayudame", "hablemos", "quiero", "necesito", "explícame", "explicame", "recuérdame", "recuerdame", "qué sabes", "que sabes", "cómo", "como", "por qué", "por que"} {
		if strings.HasPrefix(message, prefix) {
			return true
		}
	}
	return false
}

func (b *ContextBuilder) generateCarryForward(ctx context.Context, userID string, history []domain.Turn) (string, error) {
	if b.extractor == nil || len(history) == 0 {
		return "", nil
	}
	payload, err := json.Marshal(map[string]any{
		"task":                 "carry_forward",
		"instructions":         "Summarize what the conversation was discussing before the context change in 2 or 3 concise sentences.",
		"output_schema":        map[string]any{"type": "object", "required": []string{"carry_forward"}, "properties": map[string]any{"carry_forward": map[string]any{"type": "string"}}},
		"conversation_history": history,
	})
	if err != nil {
		return "", err
	}
	response, err := b.extractor.Generate(ctx, domain.ModelRequest{UserID: userID, Message: string(payload), Task: domain.TaskExtract})
	if err != nil {
		return "", err
	}
	text := strings.TrimSpace(response.AssistantMessage)
	var output struct {
		CarryForward string `json:"carry_forward"`
	}
	if json.Unmarshal([]byte(text), &output) == nil && strings.TrimSpace(output.CarryForward) != "" {
		text = strings.TrimSpace(output.CarryForward)
	}
	if len([]rune(text)) > carryForwardCharacterLimit {
		text = string([]rune(text)[:carryForwardCharacterLimit])
	}
	return text, nil
}

func (b *ContextBuilder) resolveActiveContext(ctx context.Context, userID, explicit string) (*learningdomain.UserContext, error) {
	if b.contexts == nil {
		return nil, nil
	}
	if strings.TrimSpace(explicit) != "" {
		return b.contexts.FindUserContextByScopeKey(ctx, userID, explicit)
	}
	activities, err := b.activities.ListActivities(ctx, activitiesdomain.ListFilter{UserID: userID, Status: activitiesdomain.StatusActive, Limit: b.limit})
	if err != nil {
		return nil, err
	}
	for _, activity := range activities {
		text := strings.ToLower(strings.Join(append([]string{activity.Title}, activity.Tags...), " "))
		matches, err := b.contexts.ResolveEntities(ctx, userID, text)
		if err != nil {
			return nil, err
		}
		if len(matches) == 1 && !matches[0].Ambiguous {
			return matches[0].Entity, nil
		}
	}
	// Model-inferred context is intentionally deferred until it can be resolved safely.
	return nil, nil
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

type statefulContextBuilder interface {
	BuildStateful(ctx context.Context, userID, message string, state domain.ContextBuildState) (domain.ContextSummary, error)
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

	var contextSummary domain.ContextSummary
	var err error
	if builder, ok := s.contextBuilder.(statefulContextBuilder); ok {
		contextSummary, err = builder.BuildStateful(ctx, request.UserID, request.Message, domain.ContextBuildState{
			ExplicitActiveContext: request.ActiveContext, CurrentEntityID: request.CurrentEntityID,
			CarryForward: request.CarryForward, CarryForwardEntity: request.CarryForwardEntity,
			ConversationID: request.ConversationID, History: request.History, OpenThreadRetaken: request.OpenThreadRetaken,
		})
	} else {
		contextSummary, err = s.contextBuilder.Build(ctx, request.UserID, request.Message, request.ActiveContext)
	}
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
	history := fitHistory(request.History, &contextSummary.TokenBudget, conversationHistoryCeiling)
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
		Mode:              domain.ModeDryRun,
		RequestID:         request.RequestID,
		ContextSummary:    contextSummary,
		AvailableTools:    tools,
		AssistantMessage:  modelResponse.AssistantMessage,
		ActiveEntity:      contextSummary.ActiveEntity,
		ContextChanged:    contextSummary.ContextChanged,
		CarryForward:      contextSummary.CarryForward,
		OpenThreadRetaken: contextSummary.OpenThreadRetaken,
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

func fitHistory(turns []domain.Turn, budget *domain.TokenBudget, ceiling int) []domain.Turn {
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
	if ceiling > 0 && available > ceiling {
		available = ceiling
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
