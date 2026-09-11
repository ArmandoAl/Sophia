package http

import "github.com/armandoalvarado/sofia-backend/internal/learning/domain"

type createContextRequest struct {
	Kind         string   `json:"kind"`
	Slug         string   `json:"slug"`
	Label        string   `json:"label"`
	Aliases      []string `json:"aliases"`
	Relationship string   `json:"relationship"`
	CreatedFrom  string   `json:"created_from"`
	Status       string   `json:"status"`
	ThreadsMuted bool     `json:"threads_muted"`
}

type updateContextRequest struct {
	Kind         *string   `json:"kind"`
	Label        *string   `json:"label"`
	Aliases      *[]string `json:"aliases"`
	Active       *bool     `json:"active"`
	Relationship *string   `json:"relationship"`
	Status       *string   `json:"status"`
	ThreadsMuted *bool     `json:"threads_muted"`
}

func (r createContextRequest) toDomain() domain.UserContextCreate {
	return domain.UserContextCreate{Kind: r.Kind, Slug: r.Slug, Label: r.Label, Aliases: r.Aliases, Relationship: r.Relationship, CreatedFrom: r.CreatedFrom, Status: r.Status, ThreadsMuted: r.ThreadsMuted}
}

func (r updateContextRequest) toDomain() domain.UserContextUpdate {
	return domain.UserContextUpdate{Kind: r.Kind, Label: r.Label, Aliases: r.Aliases, Active: r.Active, Relationship: r.Relationship, Status: r.Status, ThreadsMuted: r.ThreadsMuted}
}
