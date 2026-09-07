package domain

import (
	"errors"
	"strings"
	"time"
)

const (
	DefaultTimezone = "America/Tijuana"
	DefaultLocale   = "es-MX"
)

var (
	ErrProfileNotFound          = errors.New("user profile not found")
	ErrAISettingsNotFound       = errors.New("ai settings not found")
	ErrInvalidTimezone          = errors.New("timezone is required")
	ErrInvalidLocale            = errors.New("locale is required")
	ErrInvalidAssistantName     = errors.New("assistant_name is required")
	ErrInvalidPreferredLanguage = errors.New("preferred_language is required")
	ErrInvalidProactivityLevel  = errors.New("invalid proactivity_level")
	ErrInvalidAutonomyLevel     = errors.New("invalid autonomy_level")
)

type UserProfile struct {
	ID                  string
	UserID              string
	DisplayName         string
	PreferredName       string
	Timezone            string
	Locale              string
	AvatarURL           string
	OnboardingCompleted bool
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type ProfileUpdate struct {
	DisplayName   *string
	PreferredName *string
	Timezone      *string
	Locale        *string
	AvatarURL     *string
}

type UserProfileRepository interface {
	Save(profile *UserProfile) error
	FindByUserID(userID string) (*UserProfile, error)
}

func NewDefaultUserProfile(userID, displayName string) *UserProfile {
	now := time.Now()
	displayName = strings.TrimSpace(displayName)
	return &UserProfile{
		ID:                  userID,
		UserID:              userID,
		DisplayName:         displayName,
		PreferredName:       displayName,
		Timezone:            DefaultTimezone,
		Locale:              DefaultLocale,
		OnboardingCompleted: false,
		CreatedAt:           now,
		UpdatedAt:           now,
	}
}

func (p *UserProfile) Apply(update ProfileUpdate) error {
	if update.DisplayName != nil {
		p.DisplayName = strings.TrimSpace(*update.DisplayName)
	}
	if update.PreferredName != nil {
		p.PreferredName = strings.TrimSpace(*update.PreferredName)
	}
	if update.Timezone != nil {
		p.Timezone = strings.TrimSpace(*update.Timezone)
	}
	if update.Locale != nil {
		p.Locale = strings.TrimSpace(*update.Locale)
	}
	if update.AvatarURL != nil {
		p.AvatarURL = strings.TrimSpace(*update.AvatarURL)
	}
	return p.Validate()
}

func (p *UserProfile) CompleteOnboarding() {
	p.OnboardingCompleted = true
	p.UpdatedAt = time.Now()
}

func (p *UserProfile) Validate() error {
	if strings.TrimSpace(p.Timezone) == "" {
		return ErrInvalidTimezone
	}
	if strings.TrimSpace(p.Locale) == "" {
		return ErrInvalidLocale
	}
	return nil
}
