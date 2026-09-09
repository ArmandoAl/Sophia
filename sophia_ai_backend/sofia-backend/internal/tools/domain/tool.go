package domain

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

const (
	ToolCreateActivity   = "create_activity"
	ToolUpdateActivity   = "update_activity"
	ToolCompleteActivity = "complete_activity"
	ToolCreateReminder   = "create_reminder"
	ToolCancelReminder   = "cancel_reminder"
	ToolCreateMemory     = "create_memory"
	ToolSearchMemory     = "search_memory"
	ToolCreateReflection = "create_reflection"

	CategoryActivities = "activities"
	CategoryReminders  = "reminders"
	CategoryMemory     = "memory"
	CategoryInsights   = "insights"

	AutonomyManual         = "manual"
	AutonomySuggestive     = "suggestive"
	AutonomySemiAutonomous = "semi_autonomous"
)

var (
	ErrToolNotFound       = errors.New("tool not found")
	ErrInvalidToolName    = errors.New("tool name is required")
	ErrInvalidCategory    = errors.New("tool category is required")
	ErrInvalidInputSchema = errors.New("input_schema is required")
	ErrInvalidAutonomy    = errors.New("invalid min_autonomy_level")
)

type ToolDefinition struct {
	ID                   string          `json:"id"`
	Name                 string          `json:"name"`
	Description          string          `json:"description"`
	Category             string          `json:"category"`
	InputSchema          json.RawMessage `json:"input_schema"`
	OutputSchema         json.RawMessage `json:"output_schema,omitempty"`
	RequiresConfirmation bool            `json:"requires_confirmation"`
	Enabled              bool            `json:"enabled"`
	Reversible           bool            `json:"reversible"`
	MinAutonomyLevel     string          `json:"min_autonomy_level"`
	CreatedAt            time.Time       `json:"created_at"`
	UpdatedAt            time.Time       `json:"updated_at"`
}

type ToolDefinitionRepository interface {
	Save(ctx context.Context, tool *ToolDefinition) error
	FindByName(ctx context.Context, name string) (*ToolDefinition, error)
	List(ctx context.Context, enabledOnly bool) ([]*ToolDefinition, error)
}

func NewToolDefinition(id, name, description, category string, inputSchema, outputSchema json.RawMessage, requiresConfirmation, enabled bool, minAutonomyLevel string) (*ToolDefinition, error) {
	now := time.Now()
	tool := &ToolDefinition{
		ID:                   id,
		Name:                 strings.TrimSpace(name),
		Description:          strings.TrimSpace(description),
		Category:             strings.TrimSpace(category),
		InputSchema:          cloneJSON(inputSchema),
		OutputSchema:         cloneJSON(outputSchema),
		RequiresConfirmation: requiresConfirmation,
		Enabled:              enabled,
		MinAutonomyLevel:     defaultString(minAutonomyLevel, AutonomyManual),
		CreatedAt:            now,
		UpdatedAt:            now,
	}
	if err := tool.Validate(); err != nil {
		return nil, err
	}
	return tool, nil
}

func (t *ToolDefinition) Validate() error {
	if strings.TrimSpace(t.Name) == "" {
		return ErrInvalidToolName
	}
	if strings.TrimSpace(t.Category) == "" {
		return ErrInvalidCategory
	}
	if len(t.InputSchema) == 0 || !json.Valid(t.InputSchema) {
		return ErrInvalidInputSchema
	}
	if len(t.OutputSchema) > 0 && !json.Valid(t.OutputSchema) {
		return ErrInvalidInputSchema
	}
	if !IsValidAutonomyLevel(t.MinAutonomyLevel) {
		return ErrInvalidAutonomy
	}
	return nil
}

func IsValidAutonomyLevel(value string) bool {
	switch value {
	case AutonomyManual, AutonomySuggestive, AutonomySemiAutonomous:
		return true
	default:
		return false
	}
}

func cloneJSON(value json.RawMessage) json.RawMessage {
	if value == nil {
		return nil
	}
	return append(json.RawMessage(nil), value...)
}

func defaultString(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}
