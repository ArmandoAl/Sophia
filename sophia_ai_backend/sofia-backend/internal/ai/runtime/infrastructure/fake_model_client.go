package infrastructure

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	actionsdomain "github.com/armandoalvarado/sofia-backend/internal/ai/actions/domain"
	"github.com/armandoalvarado/sofia-backend/internal/ai/runtime/domain"
	toolsdomain "github.com/armandoalvarado/sofia-backend/internal/tools/domain"
)

type FakeModelClient struct{}

func NewFakeModelClient() *FakeModelClient {
	return &FakeModelClient{}
}

func (c *FakeModelClient) Generate(_ context.Context, request domain.ModelRequest) (domain.ModelResponse, error) {
	if request.Task == domain.TaskSynthesize {
		return domain.ModelResponse{
			AssistantMessage: `{"reinforced":[],"contradicted":[],"novel":[]}`,
			Usage:            domain.Usage{Model: "fake"},
		}, nil
	}
	response := domain.ModelResponse{
		AssistantMessage: "AI Runtime is ready, but no model provider is configured yet.",
		Usage:            domain.Usage{Model: "fake"},
	}

	message := strings.ToLower(request.Message)
	switch {
	case strings.Contains(message, "record") || strings.Contains(message, "recuerd") || strings.Contains(message, "recuérd"):
		if hasTool(request.Tools, toolsdomain.ToolCreateReminder) {
			input, _ := json.Marshal(map[string]any{
				"title":        truncate(request.Message, 80),
				"scheduled_at": time.Now().Add(24 * time.Hour).UTC().Format(time.RFC3339),
				"timezone":     defaultString(request.Context.Profile.Timezone, "UTC"),
			})
			response.PlannedActions = append(response.PlannedActions, domain.PlannedAction{
				ToolName:      toolsdomain.ToolCreateReminder,
				ProposedInput: input,
				Reason:        "The message appears to ask for a reminder. This is a simulated proposal; no model provider is configured.",
				RiskLevel:     actionsdomain.RiskMedium,
			})
		}
	case strings.Contains(message, "memoria") || strings.Contains(message, "remember that") || strings.Contains(message, "recuerda que"):
		if hasTool(request.Tools, toolsdomain.ToolCreateMemory) {
			input, _ := json.Marshal(map[string]any{
				"title":   truncate(request.Message, 80),
				"content": request.Message,
			})
			response.PlannedActions = append(response.PlannedActions, domain.PlannedAction{
				ToolName:      toolsdomain.ToolCreateMemory,
				ProposedInput: input,
				Reason:        "The message appears to contain explicit memory material. This is a simulated proposal.",
				RiskLevel:     actionsdomain.RiskMedium,
			})
		}
	case strings.Contains(message, "tarea") || strings.Contains(message, "actividad") || strings.Contains(message, "task"):
		if hasTool(request.Tools, toolsdomain.ToolCreateActivity) {
			input, _ := json.Marshal(map[string]any{
				"title":    truncate(request.Message, 80),
				"timezone": defaultString(request.Context.Profile.Timezone, "UTC"),
			})
			response.PlannedActions = append(response.PlannedActions, domain.PlannedAction{
				ToolName:      toolsdomain.ToolCreateActivity,
				ProposedInput: input,
				Reason:        "The message appears to ask for planning support. This is a simulated proposal.",
				RiskLevel:     actionsdomain.RiskMedium,
			})
		}
	}

	return response, nil
}

func hasTool(tools []domain.ToolSummary, name string) bool {
	for _, tool := range tools {
		if tool.Name == name {
			return true
		}
	}
	return false
}

func truncate(value string, limit int) string {
	value = strings.TrimSpace(value)
	if len(value) <= limit {
		return value
	}
	return value[:limit]
}

func defaultString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
