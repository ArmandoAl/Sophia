package http

import (
	"time"

	"github.com/armandoalvarado/sofia-backend/internal/learning/domain"
)

type contextResponse struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Kind      string    `json:"kind"`
	Slug      string    `json:"slug"`
	Label     string    `json:"label"`
	Aliases   []string  `json:"aliases"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type contextsListResponse struct {
	Contexts []contextResponse `json:"contexts"`
}

func contextResponseFromDomain(value *domain.UserContext) contextResponse {
	return contextResponse{
		ID: value.ID, UserID: value.UserID, Kind: value.Kind, Slug: value.Slug, Label: value.Label,
		Aliases: append([]string(nil), value.Aliases...), Active: value.Active, CreatedAt: value.CreatedAt, UpdatedAt: value.UpdatedAt,
	}
}

func contextListResponseFromDomain(values []*domain.UserContext) contextsListResponse {
	result := contextsListResponse{Contexts: make([]contextResponse, 0, len(values))}
	for _, value := range values {
		result.Contexts = append(result.Contexts, contextResponseFromDomain(value))
	}
	return result
}
