package http

import (
	"time"

	"github.com/armandoalvarado/sofia-backend/internal/memory/domain"
)

type memoryResponse struct {
	ID             string     `json:"id"`
	UserID         string     `json:"user_id"`
	Type           string     `json:"type"`
	Title          string     `json:"title"`
	Content        string     `json:"content"`
	Summary        string     `json:"summary,omitempty"`
	Source         string     `json:"source"`
	SourceID       string     `json:"source_id,omitempty"`
	Confidence     string     `json:"confidence"`
	Importance     string     `json:"importance"`
	Tags           []string   `json:"tags"`
	Visibility     string     `json:"visibility"`
	Status         string     `json:"status"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	LastAccessedAt *time.Time `json:"last_accessed_at,omitempty"`
}

type memoriesListResponse struct {
	Memories   []memoryResponse `json:"memories"`
	NextCursor string           `json:"next_cursor,omitempty"`
}

func memoryResponseFromDomain(memory *domain.Memory) memoryResponse {
	return memoryResponse{
		ID:             memory.ID,
		UserID:         memory.UserID,
		Type:           memory.Type,
		Title:          memory.Title,
		Content:        memory.Content,
		Summary:        memory.Summary,
		Source:         memory.Source,
		SourceID:       memory.SourceID,
		Confidence:     memory.Confidence,
		Importance:     memory.Importance,
		Tags:           append([]string(nil), memory.Tags...),
		Visibility:     memory.Visibility,
		Status:         memory.Status,
		CreatedAt:      memory.CreatedAt,
		UpdatedAt:      memory.UpdatedAt,
		LastAccessedAt: memory.LastAccessedAt,
	}
}

func memoriesListResponseFromDomain(memories []*domain.Memory) memoriesListResponse {
	response := memoriesListResponse{Memories: make([]memoryResponse, 0, len(memories))}
	for _, memory := range memories {
		response.Memories = append(response.Memories, memoryResponseFromDomain(memory))
	}
	if len(memories) > 0 {
		response.NextCursor = memories[len(memories)-1].ID
	}
	return response
}
