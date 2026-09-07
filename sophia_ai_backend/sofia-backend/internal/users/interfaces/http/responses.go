package http

import (
	"time"

	"github.com/armandoalvarado/sofia-backend/internal/users/application"
	"github.com/armandoalvarado/sofia-backend/internal/users/domain"
)

type meResponse struct {
	User       userResponse       `json:"user"`
	Profile    profileResponse    `json:"profile"`
	AISettings aiSettingsResponse `json:"ai_settings"`
}

type userResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type profileResponse struct {
	ID                  string    `json:"id"`
	UserID              string    `json:"user_id"`
	DisplayName         string    `json:"display_name"`
	PreferredName       string    `json:"preferred_name"`
	Timezone            string    `json:"timezone"`
	Locale              string    `json:"locale"`
	AvatarURL           string    `json:"avatar_url,omitempty"`
	OnboardingCompleted bool      `json:"onboarding_completed"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

type aiSettingsResponse struct {
	UserID             string    `json:"user_id"`
	AssistantName      string    `json:"assistant_name"`
	CommunicationStyle string    `json:"communication_style"`
	ProactivityLevel   string    `json:"proactivity_level"`
	AutonomyLevel      string    `json:"autonomy_level"`
	MemoryEnabled      bool      `json:"memory_enabled"`
	RemindersEnabled   bool      `json:"reminders_enabled"`
	PlanningEnabled    bool      `json:"planning_enabled"`
	PreferredLanguage  string    `json:"preferred_language"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

func meResponseFromDomain(me *application.Me) meResponse {
	return meResponse{
		User: userResponse{
			ID:    me.User.ID,
			Name:  me.User.Name,
			Email: me.User.Email,
			Role:  me.User.Role,
		},
		Profile:    profileResponseFromDomain(me.Profile),
		AISettings: aiSettingsResponseFromDomain(me.AISettings),
	}
}

func profileResponseFromDomain(profile *domain.UserProfile) profileResponse {
	return profileResponse{
		ID:                  profile.ID,
		UserID:              profile.UserID,
		DisplayName:         profile.DisplayName,
		PreferredName:       profile.PreferredName,
		Timezone:            profile.Timezone,
		Locale:              profile.Locale,
		AvatarURL:           profile.AvatarURL,
		OnboardingCompleted: profile.OnboardingCompleted,
		CreatedAt:           profile.CreatedAt,
		UpdatedAt:           profile.UpdatedAt,
	}
}

func aiSettingsResponseFromDomain(settings *domain.AISettings) aiSettingsResponse {
	return aiSettingsResponse{
		UserID:             settings.UserID,
		AssistantName:      settings.AssistantName,
		CommunicationStyle: settings.CommunicationStyle,
		ProactivityLevel:   settings.ProactivityLevel,
		AutonomyLevel:      settings.AutonomyLevel,
		MemoryEnabled:      settings.MemoryEnabled,
		RemindersEnabled:   settings.RemindersEnabled,
		PlanningEnabled:    settings.PlanningEnabled,
		PreferredLanguage:  settings.PreferredLanguage,
		CreatedAt:          settings.CreatedAt,
		UpdatedAt:          settings.UpdatedAt,
	}
}
