package infrastructure

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/armandoalvarado/sofia-backend/internal/memory/domain"
)

type InMemoryMemoryRepository struct {
	mu       sync.Mutex
	memories map[string]*domain.Memory
}

func NewInMemoryMemoryRepository() *InMemoryMemoryRepository {
	return &InMemoryMemoryRepository{memories: make(map[string]*domain.Memory)}
}

func (r *InMemoryMemoryRepository) Create(ctx context.Context, memory *domain.Memory) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.memories[memory.ID] = cloneMemory(memory)
	return nil
}

func (r *InMemoryMemoryRepository) FindByID(ctx context.Context, userID, memoryID string) (*domain.Memory, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	memory, ok := r.memories[memoryID]
	if !ok || memory.UserID != userID || memory.Status == domain.StatusDeleted {
		return nil, nil
	}
	return cloneMemory(memory), nil
}

func (r *InMemoryMemoryRepository) List(ctx context.Context, filter domain.ListFilter) ([]*domain.Memory, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	result := make([]*domain.Memory, 0)
	for _, memory := range r.memories {
		if !matchesListFilter(memory, filter) {
			continue
		}
		result = append(result, cloneMemory(memory))
	}
	sortMemories(result)
	result = memoriesAfterCursor(result, filter.Cursor)
	if filter.Limit > 0 && len(result) > filter.Limit {
		result = result[:filter.Limit]
	}
	return result, nil
}

func memoriesAfterCursor(memories []*domain.Memory, cursor string) []*domain.Memory {
	if cursor == "" {
		return memories
	}
	for i, memory := range memories {
		if memory.ID == cursor {
			return memories[i+1:]
		}
	}
	return memories
}

func (r *InMemoryMemoryRepository) Update(ctx context.Context, memory *domain.Memory) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.memories[memory.ID] = cloneMemory(memory)
	return nil
}

func (r *InMemoryMemoryRepository) Archive(ctx context.Context, userID, memoryID string) (*domain.Memory, error) {
	return r.change(ctx, userID, memoryID, func(memory *domain.Memory) error {
		return memory.MarkArchived()
	})
}

func (r *InMemoryMemoryRepository) DeleteSoft(ctx context.Context, userID, memoryID string) (*domain.Memory, error) {
	return r.change(ctx, userID, memoryID, func(memory *domain.Memory) error {
		return memory.MarkDeleted()
	})
}

func (r *InMemoryMemoryRepository) SearchBasic(ctx context.Context, filter domain.SearchFilter) ([]*domain.Memory, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	result := make([]*domain.Memory, 0)
	for _, memory := range r.memories {
		if !matchesSearchFilter(memory, filter) {
			continue
		}
		result = append(result, cloneMemory(memory))
	}
	sortMemories(result)
	if filter.Limit > 0 && len(result) > filter.Limit {
		result = result[:filter.Limit]
	}
	return result, nil
}

func (r *InMemoryMemoryRepository) TouchAccessed(ctx context.Context, userID, memoryID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	memory, ok := r.memories[memoryID]
	if !ok || memory.UserID != userID || memory.Status == domain.StatusDeleted {
		return domain.ErrMemoryNotFound
	}
	cp := cloneMemory(memory)
	cp.TouchAccessed()
	r.memories[memoryID] = cp
	return nil
}

func (r *InMemoryMemoryRepository) change(ctx context.Context, userID, memoryID string, mutate func(*domain.Memory) error) (*domain.Memory, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	memory, ok := r.memories[memoryID]
	if !ok || memory.UserID != userID || memory.Status == domain.StatusDeleted {
		return nil, domain.ErrMemoryNotFound
	}
	cp := cloneMemory(memory)
	if err := mutate(cp); err != nil {
		return nil, err
	}
	r.memories[memoryID] = cloneMemory(cp)
	return cp, nil
}

func matchesListFilter(memory *domain.Memory, filter domain.ListFilter) bool {
	if memory.UserID != filter.UserID {
		return false
	}
	if filter.Status != "" && memory.Status != filter.Status {
		return false
	}
	if filter.Status == "" && memory.Status == domain.StatusDeleted {
		return false
	}
	if filter.Type != "" && memory.Type != filter.Type {
		return false
	}
	if filter.Importance != "" && memory.Importance != filter.Importance {
		return false
	}
	if filter.Tag != "" && !hasTag(memory.Tags, filter.Tag) {
		return false
	}
	return true
}

func matchesSearchFilter(memory *domain.Memory, filter domain.SearchFilter) bool {
	if memory.UserID != filter.UserID || memory.Status != domain.StatusActive {
		return false
	}
	if filter.Type != "" && memory.Type != filter.Type {
		return false
	}
	if filter.Importance != "" && memory.Importance != filter.Importance {
		return false
	}
	if filter.Tag != "" && !hasTag(memory.Tags, filter.Tag) {
		return false
	}
	query := strings.ToLower(strings.TrimSpace(filter.Query))
	if query != "" && !strings.Contains(strings.ToLower(memory.Title), query) &&
		!strings.Contains(strings.ToLower(memory.Content), query) &&
		!strings.Contains(strings.ToLower(memory.Summary), query) {
		return false
	}
	return true
}

func hasTag(tags []string, tag string) bool {
	tag = strings.ToLower(strings.TrimSpace(tag))
	for _, existing := range tags {
		if existing == tag {
			return true
		}
	}
	return false
}

func sortMemories(memories []*domain.Memory) {
	sort.Slice(memories, func(i, j int) bool {
		if memories[i].Importance == memories[j].Importance {
			return memories[i].UpdatedAt.After(memories[j].UpdatedAt)
		}
		return importanceRank(memories[i].Importance) > importanceRank(memories[j].Importance)
	})
}

func importanceRank(importance string) int {
	switch importance {
	case domain.ImportanceCritical:
		return 4
	case domain.ImportanceHigh:
		return 3
	case domain.ImportanceMedium:
		return 2
	default:
		return 1
	}
}

func cloneMemory(memory *domain.Memory) *domain.Memory {
	if memory == nil {
		return nil
	}
	cp := *memory
	cp.Tags = append([]string(nil), memory.Tags...)
	cp.LastAccessedAt = cloneTime(memory.LastAccessedAt)
	return &cp
}

func cloneTime(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	cp := *value
	return &cp
}
