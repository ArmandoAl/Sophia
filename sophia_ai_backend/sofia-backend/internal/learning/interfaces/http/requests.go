package http

import "github.com/armandoalvarado/sofia-backend/internal/learning/domain"

type createContextRequest struct {
	Kind    string   `json:"kind"`
	Slug    string   `json:"slug"`
	Label   string   `json:"label"`
	Aliases []string `json:"aliases"`
}

type updateContextRequest struct {
	Label   *string   `json:"label"`
	Aliases *[]string `json:"aliases"`
	Active  *bool     `json:"active"`
}

func (r createContextRequest) toDomain() domain.UserContextCreate {
	return domain.UserContextCreate{Kind: r.Kind, Slug: r.Slug, Label: r.Label, Aliases: r.Aliases}
}

func (r updateContextRequest) toDomain() domain.UserContextUpdate {
	return domain.UserContextUpdate{Label: r.Label, Aliases: r.Aliases, Active: r.Active}
}
