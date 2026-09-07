package http

import "github.com/armandoalvarado/sofia-backend/internal/users/domain"

type updateProfileRequest struct {
	DisplayName   *string `json:"display_name"`
	PreferredName *string `json:"preferred_name"`
	Timezone      *string `json:"timezone"`
	Locale        *string `json:"locale"`
	AvatarURL     *string `json:"avatar_url"`
}

func (r updateProfileRequest) toDomain() domain.ProfileUpdate {
	return domain.ProfileUpdate{
		DisplayName:   r.DisplayName,
		PreferredName: r.PreferredName,
		Timezone:      r.Timezone,
		Locale:        r.Locale,
		AvatarURL:     r.AvatarURL,
	}
}

type updateAISettingsRequest struct {
	AssistantName      *string `json:"assistant_name"`
	CommunicationStyle *string `json:"communication_style"`
	ProactivityLevel   *string `json:"proactivity_level"`
	AutonomyLevel      *string `json:"autonomy_level"`
	MemoryEnabled      *bool   `json:"memory_enabled"`
	RemindersEnabled   *bool   `json:"reminders_enabled"`
	PlanningEnabled    *bool   `json:"planning_enabled"`
	PreferredLanguage  *string `json:"preferred_language"`
}

func (r updateAISettingsRequest) toDomain() domain.AISettingsUpdate {
	return domain.AISettingsUpdate{
		AssistantName:      r.AssistantName,
		CommunicationStyle: r.CommunicationStyle,
		ProactivityLevel:   r.ProactivityLevel,
		AutonomyLevel:      r.AutonomyLevel,
		MemoryEnabled:      r.MemoryEnabled,
		RemindersEnabled:   r.RemindersEnabled,
		PlanningEnabled:    r.PlanningEnabled,
		PreferredLanguage:  r.PreferredLanguage,
	}
}
