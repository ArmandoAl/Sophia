package http

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/armandoalvarado/sofia-backend/internal/insights/application"
	"github.com/armandoalvarado/sofia-backend/internal/insights/domain"
)

type createMoodEntryRequest struct {
	ActivityID  string `json:"activity_id"`
	Mood        string `json:"mood"`
	EnergyLevel int    `json:"energy_level"`
	StressLevel int    `json:"stress_level"`
	FocusLevel  int    `json:"focus_level"`
	Note        string `json:"note"`
	OccurredAt  string `json:"occurred_at"`
	Timezone    string `json:"timezone"`
}

type createActivityOutcomeRequest struct {
	ActivityID        string `json:"activity_id"`
	Result            string `json:"result"`
	CompletionQuality int    `json:"completion_quality"`
	SatisfactionLevel int    `json:"satisfaction_level"`
	DifficultyLevel   int    `json:"difficulty_level"`
	Note              string `json:"note"`
}

type createReflectionRequest struct {
	ActivityID string   `json:"activity_id"`
	Title      string   `json:"title"`
	Content    string   `json:"content"`
	Sentiment  string   `json:"sentiment"`
	Tags       []string `json:"tags"`
	OccurredAt string   `json:"occurred_at"`
}

func (r createMoodEntryRequest) toDomain() (domain.MoodEntryCreate, error) {
	occurredAt, err := parseRequiredTime(r.OccurredAt)
	if err != nil {
		return domain.MoodEntryCreate{}, err
	}
	return domain.MoodEntryCreate{
		ActivityID:  r.ActivityID,
		Mood:        r.Mood,
		EnergyLevel: r.EnergyLevel,
		StressLevel: r.StressLevel,
		FocusLevel:  r.FocusLevel,
		Note:        r.Note,
		OccurredAt:  occurredAt,
		Timezone:    r.Timezone,
	}, nil
}

func (r createActivityOutcomeRequest) toDomain() domain.ActivityOutcomeCreate {
	return domain.ActivityOutcomeCreate{
		ActivityID:        r.ActivityID,
		Result:            r.Result,
		CompletionQuality: r.CompletionQuality,
		SatisfactionLevel: r.SatisfactionLevel,
		DifficultyLevel:   r.DifficultyLevel,
		Note:              r.Note,
	}
}

func (r createReflectionRequest) toDomain() (domain.ReflectionCreate, error) {
	occurredAt, err := parseRequiredTime(r.OccurredAt)
	if err != nil {
		return domain.ReflectionCreate{}, err
	}
	return domain.ReflectionCreate{
		ActivityID: r.ActivityID,
		Title:      r.Title,
		Content:    r.Content,
		Sentiment:  r.Sentiment,
		Tags:       r.Tags,
		OccurredAt: occurredAt,
	}, nil
}

func listFilterFromRequest(r *http.Request, userID string) (domain.ListFilter, error) {
	query := r.URL.Query()
	from, err := parseQueryTime(query.Get("from"))
	if err != nil {
		return domain.ListFilter{}, err
	}
	to, err := parseQueryTime(query.Get("to"))
	if err != nil {
		return domain.ListFilter{}, err
	}

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
		ActivityID: strings.TrimSpace(query.Get("activity_id")),
		From:       from,
		To:         to,
		Limit:      limit,
	}, nil
}

func parseRequiredTime(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, domain.ErrInvalidOccurredAt
	}
	return time.Parse(time.RFC3339, value)
}

func parseQueryTime(value string) (*time.Time, error) {
	if strings.TrimSpace(value) == "" {
		return nil, nil
	}
	parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(value))
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}
