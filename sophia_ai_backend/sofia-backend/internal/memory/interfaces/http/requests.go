package http

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/armandoalvarado/sofia-backend/internal/memory/application"
	"github.com/armandoalvarado/sofia-backend/internal/memory/domain"
)

type createMemoryRequest struct {
	Type       string   `json:"type"`
	Title      string   `json:"title"`
	Content    string   `json:"content"`
	Summary    string   `json:"summary"`
	Source     string   `json:"source"`
	SourceID   string   `json:"source_id"`
	Confidence string   `json:"confidence"`
	Importance string   `json:"importance"`
	Tags       []string `json:"tags"`
	Visibility string   `json:"visibility"`
}

type updateMemoryRequest struct {
	Type       *string   `json:"type"`
	Title      *string   `json:"title"`
	Content    *string   `json:"content"`
	Summary    *string   `json:"summary"`
	Source     *string   `json:"source"`
	SourceID   *string   `json:"source_id"`
	Confidence *string   `json:"confidence"`
	Importance *string   `json:"importance"`
	Tags       *[]string `json:"tags"`
	Visibility *string   `json:"visibility"`
	Status     *string   `json:"status"`
}

type searchMemoryRequest struct {
	Query      string `json:"query"`
	Type       string `json:"type"`
	Tag        string `json:"tag"`
	Importance string `json:"importance"`
	Limit      int    `json:"limit"`
}

func (r createMemoryRequest) toDomain() domain.MemoryCreate {
	return domain.MemoryCreate{
		Type:       r.Type,
		Title:      r.Title,
		Content:    r.Content,
		Summary:    r.Summary,
		Source:     r.Source,
		SourceID:   r.SourceID,
		Confidence: r.Confidence,
		Importance: r.Importance,
		Tags:       r.Tags,
		Visibility: r.Visibility,
	}
}

func (r updateMemoryRequest) toDomain() domain.MemoryUpdate {
	return domain.MemoryUpdate{
		Type:       r.Type,
		Title:      r.Title,
		Content:    r.Content,
		Summary:    r.Summary,
		Source:     r.Source,
		SourceID:   r.SourceID,
		Confidence: r.Confidence,
		Importance: r.Importance,
		Tags:       r.Tags,
		Visibility: r.Visibility,
		Status:     r.Status,
	}
}

func (r searchMemoryRequest) toDomain(userID string) domain.SearchFilter {
	return domain.SearchFilter{
		UserID:     userID,
		Query:      strings.TrimSpace(r.Query),
		Type:       strings.TrimSpace(r.Type),
		Tag:        strings.ToLower(strings.TrimSpace(r.Tag)),
		Importance: strings.TrimSpace(r.Importance),
		Limit:      r.Limit,
	}
}

func listFilterFromRequest(r *http.Request, userID string) (domain.ListFilter, error) {
	query := r.URL.Query()
	limit := application.DefaultListLimit
	if rawLimit := strings.TrimSpace(query.Get("limit")); rawLimit != "" {
		parsed, err := strconv.Atoi(rawLimit)
		if err != nil {
			return domain.ListFilter{}, domain.ErrInvalidLimit
		}
		limit = parsed
	}

	return domain.ListFilter{
		UserID:     userID,
		Type:       strings.TrimSpace(query.Get("type")),
		Tag:        strings.ToLower(strings.TrimSpace(query.Get("tag"))),
		Importance: strings.TrimSpace(query.Get("importance")),
		Status:     strings.TrimSpace(query.Get("status")),
		Limit:      limit,
		Cursor:     strings.TrimSpace(query.Get("cursor")),
	}, nil
}
