package application

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	activitiesdomain "github.com/armandoalvarado/sofia-backend/internal/activities/domain"
	actionsdomain "github.com/armandoalvarado/sofia-backend/internal/ai/actions/domain"
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
	executors  *ExecutorRegistry
	audit      AuditRecorder
}

func NewService(repo actionsdomain.AIActionProposalRepository, tools ToolReader, settings AISettingsReader, activities ActivityCreator, reminders ReminderCreator, memories MemoryCreator) *Service {
	service := &Service{repo: repo, tools: tools, settings: settings, activities: activities, reminders: reminders, memories: memories, executors: NewExecutorRegistry()}
	service.registerDefaultExecutors()
	return service
}

func (s *Service) SetAuditRecorder(audit AuditRecorder) {
	s.audit = audit
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

func (s *Service) ConfirmActionProposal(ctx context.Context, userID, proposalID string) (*actionsdomain.AIActionProposal, error) {
	proposal, err := s.GetActionProposal(ctx, userID, proposalID)
	if err != nil {
		return nil, err
	}
	if err := proposal.Confirm(); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, proposal); err != nil {
		return nil, err
	}
	s.recordAudit(ctx, userID, "ai_action_proposal_confirmed", "ai_action_proposal", proposal.ID, map[string]any{"tool_name": proposal.ToolName})
	return proposal, nil
}

func (s *Service) RejectActionProposal(ctx context.Context, userID, proposalID string) (*actionsdomain.AIActionProposal, error) {
	proposal, err := s.GetActionProposal(ctx, userID, proposalID)
	if err != nil {
		return nil, err
	}
	if err := proposal.Reject(); err != nil {
		return nil, err
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
		if err := json.Unmarshal(proposal.ProposedInput, &input); err != nil {
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
		if err := json.Unmarshal(proposal.ProposedInput, &input); err != nil {
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
		if err := json.Unmarshal(proposal.ProposedInput, &input); err != nil {
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
