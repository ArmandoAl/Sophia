package application

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"time"

	activitiesdomain "github.com/armandoalvarado/sofia-backend/internal/activities/domain"
	actionsdomain "github.com/armandoalvarado/sofia-backend/internal/ai/actions/domain"
	learningdomain "github.com/armandoalvarado/sofia-backend/internal/learning/domain"
	memorydomain "github.com/armandoalvarado/sofia-backend/internal/memory/domain"
	"github.com/armandoalvarado/sofia-backend/internal/platform/jsonschema"
	remindersdomain "github.com/armandoalvarado/sofia-backend/internal/reminders/domain"
	toolsdomain "github.com/armandoalvarado/sofia-backend/internal/tools/domain"
	usersdomain "github.com/armandoalvarado/sofia-backend/internal/users/domain"
	"github.com/google/uuid"
)

const DefaultListLimit = 50

type ToolReader interface {
	GetTool(ctx context.Context, name string) (*toolsdomain.ToolDefinition, error)
}

type AISettingsReader interface {
	GetAISettings(userID string) (*usersdomain.AISettings, error)
}

type ActivityCreator interface {
	CreateActivity(ctx context.Context, userID string, input activitiesdomain.ActivityCreate) (*activitiesdomain.Activity, error)
}

type ReminderCreator interface {
	CreateReminder(ctx context.Context, userID string, input remindersdomain.ReminderCreate) (*remindersdomain.Reminder, error)
}

type MemoryCreator interface {
	CreateMemory(ctx context.Context, userID string, input memorydomain.MemoryCreate) (*memorydomain.Memory, error)
}

type BeliefSearcher interface {
	SearchBeliefs(ctx context.Context, userID string, terms []string, limit int) ([]*learningdomain.Belief, error)
	GetBelief(ctx context.Context, userID, beliefID string) (*learningdomain.Belief, error)
}

type AuditRecorder interface {
	RecordAuditLog(ctx context.Context, userID, action, resourceType, resourceID string, metadata map[string]any) error
}

type ToolExecutor interface {
	Execute(ctx context.Context, userID string, proposal *actionsdomain.AIActionProposal) (json.RawMessage, error)
}

type ToolExecutorFunc func(ctx context.Context, userID string, proposal *actionsdomain.AIActionProposal) (json.RawMessage, error)

func (f ToolExecutorFunc) Execute(ctx context.Context, userID string, proposal *actionsdomain.AIActionProposal) (json.RawMessage, error) {
	return f(ctx, userID, proposal)
}

type ExecutorRegistry struct {
	executors map[string]ToolExecutor
}

func NewExecutorRegistry() *ExecutorRegistry {
	return &ExecutorRegistry{executors: map[string]ToolExecutor{}}
}

func (r *ExecutorRegistry) Register(toolName string, executor ToolExecutor) {
	if r == nil || toolName == "" || executor == nil {
		return
	}
	r.executors[toolName] = executor
}

func (r *ExecutorRegistry) Execute(ctx context.Context, userID string, proposal *actionsdomain.AIActionProposal) (json.RawMessage, error) {
	if r == nil || proposal == nil {
		return nil, errors.New("tool execution is not implemented")
	}
	executor, ok := r.executors[proposal.ToolName]
	if !ok {
		return nil, errors.New("tool execution is not implemented")
	}
	return executor.Execute(ctx, userID, proposal)
}

type Service struct {
	repo       actionsdomain.AIActionProposalRepository
	tools      ToolReader
	settings   AISettingsReader
	activities ActivityCreator
	reminders  ReminderCreator
	memories   MemoryCreator
	beliefs    BeliefSearcher
	executors  *ExecutorRegistry
	audit      AuditRecorder
	threshold  float64
	now        func() time.Time
}

func NewService(repo actionsdomain.AIActionProposalRepository, tools ToolReader, settings AISettingsReader, activities ActivityCreator, reminders ReminderCreator, memories MemoryCreator) *Service {
	service := &Service{
		repo:       repo,
		tools:      tools,
		settings:   settings,
		activities: activities,
		reminders:  reminders,
		memories:   memories,
		executors:  NewExecutorRegistry(),
		threshold:  learningdomain.DefaultAutonomyThreshold,
		now:        time.Now,
	}
	service.registerDefaultExecutors()
	return service
}

func (s *Service) SetAuditRecorder(audit AuditRecorder) {
	s.audit = audit
}

func (s *Service) SetBeliefSearcher(beliefs BeliefSearcher) {
	s.beliefs = beliefs
}

func (s *Service) SetAutonomyThreshold(threshold float64) {
	if threshold < 0 || threshold > 1 {
		threshold = learningdomain.DefaultAutonomyThreshold
	}
	s.threshold = threshold
}

func (s *Service) CreateActionProposal(ctx context.Context, userID string, input actionsdomain.ProposalCreate) (*actionsdomain.AIActionProposal, error) {
	tool, err := s.getEnabledTool(ctx, input.ToolName)
	if err != nil {
		return nil, err
	}
	if err := jsonschema.Validate(tool.InputSchema, input.ProposedInput); err != nil {
		return nil, actionsdomain.ErrInvalidProposedInput
	}
	if err := s.ensureAutonomy(userID, tool, input.RiskLevel, false); err != nil {
		return nil, err
	}
	proposal, err := actionsdomain.NewAIActionProposal(uuid.New().String(), userID, input, tool.RequiresConfirmation)
	if err != nil {
		return nil, err
	}
	s.applyPrediction(ctx, userID, tool, proposal)
	if err := s.repo.Create(ctx, proposal); err != nil {
		return nil, err
	}
	s.recordAudit(ctx, userID, "ai_action_proposal_generated", "ai_action_proposal", proposal.ID, map[string]any{
		"tool_name":  proposal.ToolName,
		"risk_level": proposal.RiskLevel,
	})
	return proposal, nil
}

func (s *Service) ListActionProposals(ctx context.Context, filter actionsdomain.ListFilter) ([]*actionsdomain.AIActionProposal, error) {
	if filter.Limit == 0 {
		filter.Limit = DefaultListLimit
	}
	return s.repo.List(ctx, filter)
}

func (s *Service) GetActionProposal(ctx context.Context, userID, proposalID string) (*actionsdomain.AIActionProposal, error) {
	proposal, err := s.repo.FindByID(ctx, userID, proposalID)
	if err != nil {
		return nil, err
	}
	if proposal == nil {
		return nil, actionsdomain.ErrProposalNotFound
	}
	return proposal, nil
}

func (s *Service) ConfirmActionProposal(ctx context.Context, userID, proposalID string, correctedInput json.RawMessage, decisionLatencyMS *int64) (*actionsdomain.AIActionProposal, error) {
	proposal, err := s.GetActionProposal(ctx, userID, proposalID)
	if err != nil {
		return nil, err
	}
	if len(bytes.TrimSpace(correctedInput)) > 0 && !bytes.Equal(bytes.TrimSpace(correctedInput), []byte("null")) {
		tool, err := s.getEnabledTool(ctx, proposal.ToolName)
		if err != nil {
			return nil, err
		}
		if err := jsonschema.Validate(tool.InputSchema, correctedInput); err != nil {
			return nil, actionsdomain.ErrInvalidProposedInput
		}
	}
	if err := proposal.Confirm(correctedInput); err != nil {
		return nil, err
	}
	if decisionLatencyMS != nil {
		proposal.DecisionLatencyMS = *decisionLatencyMS
	}
	if err := s.repo.Update(ctx, proposal); err != nil {
		return nil, err
	}
	s.recordAudit(ctx, userID, "ai_action_proposal_confirmed", "ai_action_proposal", proposal.ID, map[string]any{"tool_name": proposal.ToolName})
	return proposal, nil
}

func (s *Service) RejectActionProposal(ctx context.Context, userID, proposalID string, reason string, decisionLatencyMS *int64) (*actionsdomain.AIActionProposal, error) {
	proposal, err := s.GetActionProposal(ctx, userID, proposalID)
	if err != nil {
		return nil, err
	}
	if err := proposal.Reject(reason); err != nil {
		return nil, err
	}
	if decisionLatencyMS != nil {
		proposal.DecisionLatencyMS = *decisionLatencyMS
	}
	if err := s.repo.Update(ctx, proposal); err != nil {
		return nil, err
	}
	s.recordAudit(ctx, userID, "ai_action_proposal_rejected", "ai_action_proposal", proposal.ID, map[string]any{"tool_name": proposal.ToolName})
	return proposal, nil
}

func (s *Service) ExecuteConfirmedActionProposal(ctx context.Context, userID, proposalID string) (*actionsdomain.AIActionProposal, error) {
	proposal, err := s.GetActionProposal(ctx, userID, proposalID)
	if err != nil {
		return nil, err
	}
	tool, err := s.getEnabledTool(ctx, proposal.ToolName)
	if err != nil {
		return nil, err
	}
	if proposal.RequiresConfirmation && proposal.Status != actionsdomain.StatusConfirmed {
		return nil, actionsdomain.ErrConfirmationRequired
	}
	if proposal.Status != actionsdomain.StatusConfirmed && proposal.Status != actionsdomain.StatusProposed {
		return nil, actionsdomain.ErrInvalidTransition
	}
	if err := s.ensureAutonomy(userID, tool, proposal.RiskLevel, true); err != nil {
		return nil, err
	}

	result, execErr := s.executors.Execute(ctx, userID, proposal)
	if execErr != nil {
		_ = proposal.MarkFailed(execErr.Error())
		if err := s.repo.Update(ctx, proposal); err != nil {
			return nil, err
		}
		s.recordAudit(ctx, userID, "ai_action_proposal_failed", "ai_action_proposal", proposal.ID, map[string]any{
			"tool_name": proposal.ToolName,
			"error":     execErr.Error(),
		})
		return proposal, nil
	}
	if err := proposal.MarkExecuted(result); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, proposal); err != nil {
		return nil, err
	}
	s.recordAudit(ctx, userID, "ai_action_proposal_executed", "ai_action_proposal", proposal.ID, map[string]any{"tool_name": proposal.ToolName})
	return proposal, nil
}

func (s *Service) AutoExecuteActionProposal(ctx context.Context, userID, proposalID string) (*actionsdomain.AIActionProposal, error) {
	proposal, err := s.GetActionProposal(ctx, userID, proposalID)
	if err != nil {
		return nil, err
	}
	if proposal.Status != actionsdomain.StatusProposed {
		return nil, actionsdomain.ErrInvalidTransition
	}
	settings, err := s.settings.GetAISettings(userID)
	if err != nil && !errors.Is(err, usersdomain.ErrAISettingsNotFound) {
		return nil, err
	}
	level := usersdomain.AutonomySuggestive
	if settings != nil {
		level = settings.AutonomyLevel
	}
	if level != usersdomain.AutonomySemiAutonomous {
		return nil, actionsdomain.ErrAutonomyNotAllowed
	}
	decision, err := s.EvaluateAutonomy(ctx, userID, proposal.ToolName)
	if err != nil {
		return nil, err
	}
	if !decision.AutoExecute {
		return nil, actionsdomain.ErrAutonomyNotAllowed
	}
	if !s.hasExecutor(proposal.ToolName) {
		return proposal, nil
	}

	result, execErr := s.executors.Execute(ctx, userID, proposal)
	if execErr != nil {
		_ = proposal.MarkFailed(execErr.Error())
		proposal.AutonomyModeUsed = actionsdomain.AutonomyModeAutoExecuted
		if err := s.repo.Update(ctx, proposal); err != nil {
			return nil, err
		}
		s.recordAudit(ctx, userID, "ai_action_proposal_failed", "ai_action_proposal", proposal.ID, map[string]any{
			"tool_name": proposal.ToolName,
			"error":     execErr.Error(),
		})
		return proposal, nil
	}
	if err := proposal.MarkExecuted(result); err != nil {
		return nil, err
	}
	proposal.AutonomyModeUsed = actionsdomain.AutonomyModeAutoExecuted
	if err := s.repo.Update(ctx, proposal); err != nil {
		return nil, err
	}
	s.recordAudit(ctx, userID, "ai_action_proposal_auto_executed", "ai_action_proposal", proposal.ID, map[string]any{
		"tool_name":          proposal.ToolName,
		"predicted_approval": proposal.PredictedApproval,
		"prediction_basis":   proposal.PredictionBasis,
		"prediction_model":   proposal.PredictionModelVersion,
	})
	return proposal, nil
}

type AutonomyDecision struct {
	Probability  float64
	Basis        []string
	History      learningdomain.ToolHistory
	Reversible   bool
	AutoExecute  bool
	ModelVersion string
	Beliefs      []learningdomain.ScoredBelief
}

type ProposalExplanation struct {
	PredictedApproval      float64                    `json:"predicted_approval"`
	PredictionBasis        []string                   `json:"prediction_basis"`
	PredictionModelVersion string                     `json:"prediction_model_version"`
	AutonomyModeUsed       string                     `json:"autonomy_mode_used"`
	ToolHistory            learningdomain.ToolHistory `json:"tool_history"`
	Beliefs                []explainedBelief          `json:"beliefs"`
}

type explainedBelief struct {
	ID          string  `json:"id"`
	Statement   string  `json:"statement,omitempty"`
	Confidence  float64 `json:"confidence"`
	Contradicts bool    `json:"contradicts"`
}

func (s *Service) EvaluateAutonomy(ctx context.Context, userID, toolName string) (AutonomyDecision, error) {
	tool, err := s.getEnabledTool(ctx, toolName)
	if err != nil {
		return AutonomyDecision{}, err
	}
	history, err := s.toolHistory(ctx, userID, toolName)
	if err != nil {
		return AutonomyDecision{}, err
	}
	beliefs, err := s.beliefsForTool(ctx, userID, tool)
	if err != nil {
		return AutonomyDecision{}, err
	}
	probability, basis := learningdomain.PredictApproval(history, beliefs)
	return AutonomyDecision{
		Probability:  probability,
		Basis:        basis,
		History:      history,
		Reversible:   tool.Reversible,
		AutoExecute:  learningdomain.ShouldAutoExecute(probability, history, tool.Reversible, s.threshold),
		ModelVersion: actionsdomain.PredictionModelHeuristicV1,
		Beliefs:      beliefs,
	}, nil
}

func (s *Service) ExplainActionProposal(ctx context.Context, userID, proposalID string) (*ProposalExplanation, error) {
	proposal, err := s.GetActionProposal(ctx, userID, proposalID)
	if err != nil {
		return nil, err
	}
	decision, err := s.EvaluateAutonomy(ctx, userID, proposal.ToolName)
	if err != nil {
		return nil, err
	}
	explained := make([]explainedBelief, 0, len(decision.Beliefs))
	for _, belief := range decision.Beliefs {
		item := explainedBelief{ID: belief.ID, Confidence: belief.Confidence, Contradicts: belief.Contradicts}
		if s.beliefs != nil {
			if stored, err := s.beliefs.GetBelief(ctx, userID, belief.ID); err == nil && stored != nil {
				item.Statement = stored.Statement
			}
		}
		explained = append(explained, item)
	}
	return &ProposalExplanation{
		PredictedApproval:      proposal.PredictedApproval,
		PredictionBasis:        append([]string(nil), proposal.PredictionBasis...),
		PredictionModelVersion: proposal.PredictionModelVersion,
		AutonomyModeUsed:       proposal.AutonomyModeUsed,
		ToolHistory:            decision.History,
		Beliefs:                explained,
	}, nil
}

func (s *Service) applyPrediction(ctx context.Context, userID string, tool *toolsdomain.ToolDefinition, proposal *actionsdomain.AIActionProposal) {
	decision, err := s.EvaluateAutonomy(ctx, userID, tool.Name)
	if err != nil {
		proposal.PredictedApproval = 0.5
		proposal.PredictionModelVersion = actionsdomain.PredictionModelHeuristicV1
		proposal.AutonomyModeUsed = actionsdomain.AutonomyModeProposed
		return
	}
	proposal.PredictedApproval = decision.Probability
	proposal.PredictionBasis = append([]string(nil), decision.Basis...)
	proposal.PredictionModelVersion = decision.ModelVersion
	proposal.AutonomyModeUsed = actionsdomain.AutonomyModeProposed
}

func (s *Service) toolHistory(ctx context.Context, userID, toolName string) (learningdomain.ToolHistory, error) {
	now := s.now()
	from := now.AddDate(-2, 0, 0)
	proposals, err := s.repo.ListByDateRange(ctx, userID, from, now.Add(time.Second))
	if err != nil {
		return learningdomain.ToolHistory{}, err
	}
	history := learningdomain.ToolHistory{}
	var last time.Time
	for _, proposal := range proposals {
		if proposal.ToolName != toolName {
			continue
		}
		switch proposal.Feedback {
		case actionsdomain.FeedbackApprovedDirect:
			history.Total++
			history.ApprovedDirect++
		case actionsdomain.FeedbackApprovedCorrected:
			history.Total++
			history.ApprovedCorrected++
		case actionsdomain.FeedbackRejected:
			history.Total++
			history.Rejected++
		default:
			continue
		}
		at := proposal.CreatedAt
		if proposal.DecidedAt != nil && !proposal.DecidedAt.IsZero() {
			at = *proposal.DecidedAt
		}
		if last.IsZero() || at.After(last) {
			last = at
		}
	}
	if !last.IsZero() {
		history.DaysSinceLastSample = now.Sub(last).Hours() / 24
		if history.DaysSinceLastSample < 0 {
			history.DaysSinceLastSample = 0
		}
	}
	return history, nil
}

func (s *Service) beliefsForTool(ctx context.Context, userID string, tool *toolsdomain.ToolDefinition) ([]learningdomain.ScoredBelief, error) {
	if s.beliefs == nil || tool == nil {
		return nil, nil
	}
	terms := memorydomain.ExtractTerms(tool.Name+" "+tool.Category+" "+tool.Description, 20)
	if len(terms) == 0 {
		return nil, nil
	}
	found, err := s.beliefs.SearchBeliefs(ctx, userID, terms, 10)
	if err != nil {
		return nil, err
	}
	now := s.now()
	result := make([]learningdomain.ScoredBelief, 0, len(found))
	for _, belief := range found {
		result = append(result, learningdomain.ScoredBelief{
			ID:          belief.ID,
			Confidence:  belief.DecayedConfidence(now),
			Contradicts: belief.ContradictionCount > belief.EvidenceCount,
		})
	}
	return result, nil
}

func (s *Service) hasExecutor(toolName string) bool {
	if s.executors == nil {
		return false
	}
	_, ok := s.executors.executors[toolName]
	return ok
}

func (s *Service) getEnabledTool(ctx context.Context, name string) (*toolsdomain.ToolDefinition, error) {
	tool, err := s.tools.GetTool(ctx, name)
	if err != nil {
		return nil, err
	}
	if !tool.Enabled {
		return nil, actionsdomain.ErrToolDisabled
	}
	return tool, nil
}

func (s *Service) ensureAutonomy(userID string, tool *toolsdomain.ToolDefinition, riskLevel string, executing bool) error {
	settings, err := s.settings.GetAISettings(userID)
	if err != nil {
		if errors.Is(err, usersdomain.ErrAISettingsNotFound) {
			return nil
		}
		return err
	}
	level := usersdomain.AutonomySuggestive
	if settings != nil {
		level = settings.AutonomyLevel
	}
	if level == usersdomain.AutonomyManual && executing && !tool.RequiresConfirmation {
		return actionsdomain.ErrAutonomyNotAllowed
	}
	if level != usersdomain.AutonomySemiAutonomous && executing && !tool.RequiresConfirmation && riskLevel != actionsdomain.RiskLow {
		return actionsdomain.ErrAutonomyNotAllowed
	}
	return nil
}

func (s *Service) registerDefaultExecutors() {
	s.executors.Register(toolsdomain.ToolCreateActivity, ToolExecutorFunc(func(ctx context.Context, userID string, proposal *actionsdomain.AIActionProposal) (json.RawMessage, error) {
		var input createActivityInput
		if err := json.Unmarshal(proposal.ExecutionInput(), &input); err != nil {
			return nil, err
		}
		activity, err := s.activities.CreateActivity(ctx, userID, input.toDomain())
		if err != nil {
			return nil, err
		}
		return json.Marshal(map[string]any{"activity_id": activity.ID, "status": activity.Status})
	}))
	s.executors.Register(toolsdomain.ToolCreateReminder, ToolExecutorFunc(func(ctx context.Context, userID string, proposal *actionsdomain.AIActionProposal) (json.RawMessage, error) {
		var input createReminderInput
		if err := json.Unmarshal(proposal.ExecutionInput(), &input); err != nil {
			return nil, err
		}
		reminderInput, err := input.toDomain()
		if err != nil {
			return nil, err
		}
		reminder, err := s.reminders.CreateReminder(ctx, userID, reminderInput)
		if err != nil {
			return nil, err
		}
		return json.Marshal(map[string]any{"reminder_id": reminder.ID, "status": reminder.Status})
	}))
	s.executors.Register(toolsdomain.ToolCreateMemory, ToolExecutorFunc(func(ctx context.Context, userID string, proposal *actionsdomain.AIActionProposal) (json.RawMessage, error) {
		var input createMemoryInput
		if err := json.Unmarshal(proposal.ExecutionInput(), &input); err != nil {
			return nil, err
		}
		memory, err := s.memories.CreateMemory(ctx, userID, input.toDomain())
		if err != nil {
			return nil, err
		}
		return json.Marshal(map[string]any{"memory_id": memory.ID, "status": memory.Status})
	}))
}

func (s *Service) recordAudit(ctx context.Context, userID, action, resourceType, resourceID string, metadata map[string]any) {
	if s.audit == nil {
		return
	}
	_ = s.audit.RecordAuditLog(ctx, userID, action, resourceType, resourceID, metadata)
}

type createActivityInput struct {
	Type        string `json:"type"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Priority    string `json:"priority"`
	Timezone    string `json:"timezone"`
}

func (i createActivityInput) toDomain() activitiesdomain.ActivityCreate {
	return activitiesdomain.ActivityCreate{Type: i.Type, Title: i.Title, Description: i.Description, Priority: i.Priority, Timezone: i.Timezone}
}

type createReminderInput struct {
	ActivityID  string `json:"activity_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	ScheduledAt string `json:"scheduled_at"`
	Timezone    string `json:"timezone"`
}

func (i createReminderInput) toDomain() (remindersdomain.ReminderCreate, error) {
	scheduledAt, err := time.Parse(time.RFC3339, i.ScheduledAt)
	if err != nil {
		return remindersdomain.ReminderCreate{}, err
	}
	return remindersdomain.ReminderCreate{ActivityID: i.ActivityID, Title: i.Title, Description: i.Description, ScheduledAt: scheduledAt, Timezone: i.Timezone}, nil
}

type createMemoryInput struct {
	Type       string   `json:"type"`
	Title      string   `json:"title"`
	Content    string   `json:"content"`
	Summary    string   `json:"summary"`
	Source     string   `json:"source"`
	Confidence string   `json:"confidence"`
	Importance string   `json:"importance"`
	Tags       []string `json:"tags"`
	Visibility string   `json:"visibility"`
}

func (i createMemoryInput) toDomain() memorydomain.MemoryCreate {
	return memorydomain.MemoryCreate{Type: i.Type, Title: i.Title, Content: i.Content, Summary: i.Summary, Source: i.Source, Confidence: i.Confidence, Importance: i.Importance, Tags: i.Tags, Visibility: i.Visibility}
}
