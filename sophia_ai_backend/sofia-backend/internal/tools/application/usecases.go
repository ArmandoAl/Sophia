package application

import (
	"context"
	"encoding/json"

	"github.com/armandoalvarado/sofia-backend/internal/tools/domain"
)

type Service struct {
	repo domain.ToolDefinitionRepository
}

func NewService(repo domain.ToolDefinitionRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) SeedDefaultTools(ctx context.Context) error {
	for _, tool := range DefaultTools() {
		existing, err := s.repo.FindByName(ctx, tool.Name)
		if err != nil {
			return err
		}
		if existing != nil {
			continue
		}
		if err := s.repo.Save(ctx, tool); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) ListTools(ctx context.Context) ([]*domain.ToolDefinition, error) {
	return s.repo.List(ctx, false)
}

func (s *Service) GetTool(ctx context.Context, name string) (*domain.ToolDefinition, error) {
	tool, err := s.repo.FindByName(ctx, name)
	if err != nil {
		return nil, err
	}
	if tool == nil {
		return nil, domain.ErrToolNotFound
	}
	return tool, nil
}

func DefaultTools() []*domain.ToolDefinition {
	defs := []struct {
		name        string
		description string
		category    string
		confirm     bool
		minAutonomy string
		schema      string
	}{
		{domain.ToolCreateActivity, "Create an activity for the authenticated user.", domain.CategoryActivities, true, domain.AutonomyManual, `{"type":"object","required":["title","timezone"],"additionalProperties":false,"properties":{"title":{"type":"string"},"description":{"type":"string"},"type":{"type":"string","enum":["task","habit","routine","goal","event","checklist"]},"priority":{"type":"string","enum":["low","medium","high","urgent"]},"timezone":{"type":"string"}}}`},
		{domain.ToolUpdateActivity, "Update an existing activity.", domain.CategoryActivities, true, domain.AutonomySuggestive, `{"type":"object","required":["activity_id"],"additionalProperties":false,"properties":{"activity_id":{"type":"string"},"title":{"type":"string"},"status":{"type":"string","enum":["pending","active","completed","canceled","archived"]}}}`},
		{domain.ToolCompleteActivity, "Complete an activity.", domain.CategoryActivities, true, domain.AutonomySuggestive, `{"type":"object","required":["activity_id"],"additionalProperties":false,"properties":{"activity_id":{"type":"string"}}}`},
		{domain.ToolCreateReminder, "Create a reminder for the authenticated user.", domain.CategoryReminders, true, domain.AutonomyManual, `{"type":"object","required":["title","scheduled_at","timezone"],"additionalProperties":false,"properties":{"activity_id":{"type":"string"},"title":{"type":"string"},"description":{"type":"string"},"scheduled_at":{"type":"string"},"timezone":{"type":"string"}}}`},
		{domain.ToolCancelReminder, "Cancel a reminder.", domain.CategoryReminders, true, domain.AutonomySuggestive, `{"type":"object","required":["reminder_id"],"additionalProperties":false,"properties":{"reminder_id":{"type":"string"}}}`},
		{domain.ToolCreateMemory, "Create an explicit memory.", domain.CategoryMemory, true, domain.AutonomyManual, `{"type":"object","required":["title","content"],"additionalProperties":false,"properties":{"title":{"type":"string"},"content":{"type":"string"},"summary":{"type":"string"},"source":{"type":"string"},"type":{"type":"string","enum":["preference","fact","episodic","semantic","instruction","relationship","project","system_note"]},"confidence":{"type":"string","enum":["low","medium","high"]},"importance":{"type":"string","enum":["low","medium","high","critical"]},"visibility":{"type":"string","enum":["private","assistant_context","archived"]},"tags":{"type":"array","items":{"type":"string"}}}}`},
		{domain.ToolSearchMemory, "Search explicit memory.", domain.CategoryMemory, false, domain.AutonomySemiAutonomous, `{"type":"object","additionalProperties":false,"properties":{"query":{"type":"string"},"type":{"type":"string"},"tag":{"type":"string"},"importance":{"type":"string","enum":["low","medium","high","critical"]}}}`},
		{domain.ToolCreateReflection, "Create a reflection entry.", domain.CategoryInsights, true, domain.AutonomyManual, `{"type":"object","required":["content","occurred_at"],"additionalProperties":false,"properties":{"title":{"type":"string"},"content":{"type":"string"},"occurred_at":{"type":"string"},"tags":{"type":"array","items":{"type":"string"}}}}`},
	}
	result := make([]*domain.ToolDefinition, 0, len(defs))
	for _, def := range defs {
		tool, err := domain.NewToolDefinition(def.name, def.name, def.description, def.category, json.RawMessage(def.schema), nil, def.confirm, true, def.minAutonomy)
		if err != nil {
			panic(err)
		}
		result = append(result, tool)
	}
	return result
}
