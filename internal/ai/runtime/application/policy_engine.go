package application

import (
	"encoding/json"

	actionsdomain "github.com/armandoalvarado/sofia-backend/internal/ai/actions/domain"
	"github.com/armandoalvarado/sofia-backend/internal/ai/runtime/domain"
	"github.com/armandoalvarado/sofia-backend/internal/platform/jsonschema"
	toolsdomain "github.com/armandoalvarado/sofia-backend/internal/tools/domain"
	usersdomain "github.com/armandoalvarado/sofia-backend/internal/users/domain"
)

type PolicyEngine struct{}

func NewPolicyEngine() *PolicyEngine {
	return &PolicyEngine{}
}

func (p *PolicyEngine) AllowTool(context domain.ContextSummary, tool *toolsdomain.ToolDefinition, risk string) bool {
	if tool == nil || !tool.Enabled {
		return false
	}
	if !autonomyAllows(context.AISettings.AutonomyLevel, tool.MinAutonomyLevel) {
		return false
	}
	if tool.Category == toolsdomain.CategoryMemory && !context.AISettings.MemoryEnabled {
		return false
	}
	if tool.Category == toolsdomain.CategoryReminders && !context.AISettings.RemindersEnabled {
		return false
	}
	if tool.Category == toolsdomain.CategoryActivities && !context.AISettings.PlanningEnabled {
		return false
	}
	if !tool.RequiresConfirmation && risk != actionsdomain.RiskLow && context.AISettings.AutonomyLevel != usersdomain.AutonomySemiAutonomous {
		return false
	}
	return true
}

func (p *PolicyEngine) AllowProposal(userID string, action domain.PlannedAction, context domain.ContextSummary, tool domain.ToolSummary) bool {
	if userID == "" || userID != context.User.ID {
		return false
	}
	if tool.Name == "" {
		return false
	}
	if action.ToolName == toolsdomain.ToolCreateMemory && !context.AISettings.MemoryEnabled {
		return false
	}
	if action.ToolName == toolsdomain.ToolCreateReminder && !context.AISettings.RemindersEnabled {
		return false
	}
	if isPlanningTool(action.ToolName) && !context.AISettings.PlanningEnabled {
		return false
	}
	risk := defaultRisk(action.RiskLevel)
	if !tool.RequiresConfirmation && risk != actionsdomain.RiskLow {
		return false
	}
	if !json.Valid(action.ProposedInput) || len(action.ProposedInput) == 0 {
		return false
	}
	return jsonschema.Validate(tool.InputSchema, action.ProposedInput) == nil
}
