package http

import (
	"time"

	"github.com/armandoalvarado/sofia-backend/internal/insights/domain"
)

type moodEntryResponse struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	ActivityID  string    `json:"activity_id,omitempty"`
	Mood        string    `json:"mood"`
	EnergyLevel int       `json:"energy_level"`
	StressLevel int       `json:"stress_level"`
	FocusLevel  int       `json:"focus_level"`
	Note        string    `json:"note,omitempty"`
	OccurredAt  time.Time `json:"occurred_at"`
	Timezone    string    `json:"timezone"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type activityOutcomeResponse struct {
	ID                string    `json:"id"`
	UserID            string    `json:"user_id"`
	ActivityID        string    `json:"activity_id"`
	Result            string    `json:"result"`
	CompletionQuality int       `json:"completion_quality"`
	SatisfactionLevel int       `json:"satisfaction_level"`
	DifficultyLevel   int       `json:"difficulty_level"`
	Note              string    `json:"note,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type reflectionResponse struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	ActivityID string    `json:"activity_id,omitempty"`
	Title      string    `json:"title,omitempty"`
	Content    string    `json:"content"`
	Sentiment  string    `json:"sentiment,omitempty"`
	Tags       []string  `json:"tags"`
	OccurredAt time.Time `json:"occurred_at"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type moodEntriesListResponse struct {
	MoodEntries []moodEntryResponse `json:"mood_entries"`
}

type activityOutcomesListResponse struct {
	Outcomes []activityOutcomeResponse `json:"outcomes"`
}

type reflectionsListResponse struct {
	Reflections []reflectionResponse `json:"reflections"`
}

type summaryResponse struct {
	AverageEnergy    float64 `json:"average_energy"`
	AverageStress    float64 `json:"average_stress"`
	AverageFocus     float64 `json:"average_focus"`
	TotalOutcomes    int     `json:"total_outcomes"`
	TotalReflections int     `json:"total_reflections"`
}

func moodEntryResponseFromDomain(entry *domain.MoodEntry) moodEntryResponse {
	return moodEntryResponse{
		ID:          entry.ID,
		UserID:      entry.UserID,
		ActivityID:  entry.ActivityID,
		Mood:        entry.Mood,
		EnergyLevel: entry.EnergyLevel,
		StressLevel: entry.StressLevel,
		FocusLevel:  entry.FocusLevel,
		Note:        entry.Note,
		OccurredAt:  entry.OccurredAt,
		Timezone:    entry.Timezone,
		CreatedAt:   entry.CreatedAt,
		UpdatedAt:   entry.UpdatedAt,
	}
}

func activityOutcomeResponseFromDomain(outcome *domain.ActivityOutcome) activityOutcomeResponse {
	return activityOutcomeResponse{
		ID:                outcome.ID,
		UserID:            outcome.UserID,
		ActivityID:        outcome.ActivityID,
		Result:            outcome.Result,
		CompletionQuality: outcome.CompletionQuality,
		SatisfactionLevel: outcome.SatisfactionLevel,
		DifficultyLevel:   outcome.DifficultyLevel,
		Note:              outcome.Note,
		CreatedAt:         outcome.CreatedAt,
		UpdatedAt:         outcome.UpdatedAt,
	}
}

func reflectionResponseFromDomain(reflection *domain.Reflection) reflectionResponse {
	return reflectionResponse{
		ID:         reflection.ID,
		UserID:     reflection.UserID,
		ActivityID: reflection.ActivityID,
		Title:      reflection.Title,
		Content:    reflection.Content,
		Sentiment:  reflection.Sentiment,
		Tags:       append([]string(nil), reflection.Tags...),
		OccurredAt: reflection.OccurredAt,
		CreatedAt:  reflection.CreatedAt,
		UpdatedAt:  reflection.UpdatedAt,
	}
}

func moodEntriesListResponseFromDomain(entries []*domain.MoodEntry) moodEntriesListResponse {
	response := moodEntriesListResponse{MoodEntries: make([]moodEntryResponse, 0, len(entries))}
	for _, entry := range entries {
		response.MoodEntries = append(response.MoodEntries, moodEntryResponseFromDomain(entry))
	}
	return response
}

func activityOutcomesListResponseFromDomain(outcomes []*domain.ActivityOutcome) activityOutcomesListResponse {
	response := activityOutcomesListResponse{Outcomes: make([]activityOutcomeResponse, 0, len(outcomes))}
	for _, outcome := range outcomes {
		response.Outcomes = append(response.Outcomes, activityOutcomeResponseFromDomain(outcome))
	}
	return response
}

func reflectionsListResponseFromDomain(reflections []*domain.Reflection) reflectionsListResponse {
	response := reflectionsListResponse{Reflections: make([]reflectionResponse, 0, len(reflections))}
	for _, reflection := range reflections {
		response.Reflections = append(response.Reflections, reflectionResponseFromDomain(reflection))
	}
	return response
}

func summaryResponseFromDomain(summary *domain.Summary) summaryResponse {
	return summaryResponse{
		AverageEnergy:    summary.AverageEnergy,
		AverageStress:    summary.AverageStress,
		AverageFocus:     summary.AverageFocus,
		TotalOutcomes:    summary.TotalOutcomes,
		TotalReflections: summary.TotalReflections,
	}
}
