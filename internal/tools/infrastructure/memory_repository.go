package infrastructure

import (
	"context"
	"sort"
	"sync"

	"github.com/armandoalvarado/sofia-backend/internal/tools/domain"
)

type InMemoryToolDefinitionRepository struct {
	mu    sync.Mutex
	tools map[string]*domain.ToolDefinition
}

func NewInMemoryToolDefinitionRepository() *InMemoryToolDefinitionRepository {
	return &InMemoryToolDefinitionRepository{tools: make(map[string]*domain.ToolDefinition)}
}

func (r *InMemoryToolDefinitionRepository) Save(ctx context.Context, tool *domain.ToolDefinition) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tools[tool.Name] = cloneTool(tool)
	return nil
}

func (r *InMemoryToolDefinitionRepository) FindByName(ctx context.Context, name string) (*domain.ToolDefinition, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	tool, ok := r.tools[name]
	if !ok {
		return nil, nil
	}
	return cloneTool(tool), nil
}

func (r *InMemoryToolDefinitionRepository) List(ctx context.Context, enabledOnly bool) ([]*domain.ToolDefinition, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	result := make([]*domain.ToolDefinition, 0, len(r.tools))
	for _, tool := range r.tools {
		if enabledOnly && !tool.Enabled {
			continue
		}
		result = append(result, cloneTool(tool))
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	return result, nil
}

func cloneTool(tool *domain.ToolDefinition) *domain.ToolDefinition {
	if tool == nil {
		return nil
	}
	cp := *tool
	cp.InputSchema = append([]byte(nil), tool.InputSchema...)
	cp.OutputSchema = append([]byte(nil), tool.OutputSchema...)
	return &cp
}
