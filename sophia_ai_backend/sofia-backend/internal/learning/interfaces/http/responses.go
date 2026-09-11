package http

import (
	"time"

	"github.com/armandoalvarado/sofia-backend/internal/learning/domain"
)

type contextResponse struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	Kind         string    `json:"kind"`
	Slug         string    `json:"slug"`
	Label        string    `json:"label"`
	Aliases      []string  `json:"aliases"`
	Relationship string    `json:"relationship"`
	CreatedFrom  string    `json:"created_from"`
	Status       string    `json:"status"`
	MergedInto   string    `json:"merged_into,omitempty"`
	ThreadsMuted bool      `json:"threads_muted"`
	Active       bool      `json:"active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type contextsListResponse struct {
	Contexts []contextResponse `json:"contexts"`
}

type episodeResponse struct {
	ID         string    `json:"id"`
	OccurredAt time.Time `json:"occurred_at"`
	Summary    string    `json:"summary"`
	Topics     []string  `json:"topics,omitempty"`
	Salience   float64   `json:"salience"`
}

func contextResponseFromDomain(value *domain.UserContext) contextResponse {
	return contextResponse{
		ID: value.ID, UserID: value.UserID, Kind: value.Kind, Slug: value.Slug, Label: value.Label,
		Aliases: append([]string(nil), value.Aliases...), Relationship: value.Relationship, CreatedFrom: value.EffectiveCreatedFrom(),
		Status: value.EffectiveStatus(), MergedInto: value.MergedInto, ThreadsMuted: value.ThreadsMuted, Active: value.Active, CreatedAt: value.CreatedAt, UpdatedAt: value.UpdatedAt,
	}
}

func contextListResponseFromDomain(values []*domain.UserContext) contextsListResponse {
	result := contextsListResponse{Contexts: make([]contextResponse, 0, len(values))}
	for _, value := range values {
		result.Contexts = append(result.Contexts, contextResponseFromDomain(value))
	}
	return result
}

func episodeResponseFromDomain(value *domain.Episode) episodeResponse {
	return episodeResponse{ID: value.ID, OccurredAt: value.OccurredAt, Summary: value.Summary, Topics: append([]string(nil), value.Topics...), Salience: value.Salience}
}
