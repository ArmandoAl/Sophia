package domain

import (
	"strings"
	"time"
)

const (
	DefaultAssistantName      = "Sofía"
	DefaultCommunicationStyle = "warm"
	DefaultPreferredLanguage  = "es"

	ProactivityLow    = "low"
	ProactivityMedium = "medium"
	ProactivityHigh   = "high"

	AutonomyManual         = "manual"
	AutonomySuggestive     = "suggestive"
	AutonomySemiAutonomous = "semi_autonomous"
)

type AISettings struct {
	UserID             string
	AssistantName      string
	CommunicationStyle string
	ProactivityLevel   string
	AutonomyLevel      string
	MemoryEnabled      bool
	RemindersEnabled   bool
	PlanningEnabled    bool
	PreferredLanguage  string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type AISettingsUpdate struct {
	AssistantName      *string
	CommunicationStyle *string
	ProactivityLevel   *string
	AutonomyLevel      *string
	MemoryEnabled      *bool
	RemindersEnabled   *bool
	PlanningEnabled    *bool
	PreferredLanguage  *string
}

type AISettingsRepository interface {
	Save(settings *AISettings) error
	FindByUserID(userID string) (*AISettings, error)
}

func NewDefaultAISettings(userID string) *AISettings {
	now := time.Now()
	return &AISettings{
		UserID:             userID,
		AssistantName:      DefaultAssistantName,
		CommunicationStyle: DefaultCommunicationStyle,
		ProactivityLevel:   ProactivityMedium,
		AutonomyLevel:      AutonomySuggestive,
		MemoryEnabled:      true,
		RemindersEnabled:   true,
		PlanningEnabled:    true,
		PreferredLanguage:  DefaultPreferredLanguage,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
}

func (s *AISettings) Apply(update AISettingsUpdate) error {
	if update.AssistantName != nil {
		s.AssistantName = strings.TrimSpace(*update.AssistantName)
	}
	if update.CommunicationStyle != nil {
		s.CommunicationStyle = strings.TrimSpace(*update.CommunicationStyle)
	}
	if update.ProactivityLevel != nil {
		s.ProactivityLevel = strings.TrimSpace(*update.ProactivityLevel)
	}
	if update.AutonomyLevel != nil {
		s.AutonomyLevel = strings.TrimSpace(*update.AutonomyLevel)
	}
	if update.MemoryEnabled != nil {
		s.MemoryEnabled = *update.MemoryEnabled
	}
	if update.RemindersEnabled != nil {
		s.RemindersEnabled = *update.RemindersEnabled
	}
	if update.PlanningEnabled != nil {
		s.PlanningEnabled = *update.PlanningEnabled
	}
	if update.PreferredLanguage != nil {
		s.PreferredLanguage = strings.TrimSpace(*update.PreferredLanguage)
	}
	return s.Validate()
}

func (s *AISettings) Validate() error {
	if strings.TrimSpace(s.AssistantName) == "" {
		return ErrInvalidAssistantName
	}
	if !IsValidProactivityLevel(s.ProactivityLevel) {
		return ErrInvalidProactivityLevel
	}
	if !IsValidAutonomyLevel(s.AutonomyLevel) {
		return ErrInvalidAutonomyLevel
	}
	if strings.TrimSpace(s.PreferredLanguage) == "" {
		return ErrInvalidPreferredLanguage
	}
	return nil
}

func IsValidProactivityLevel(value string) bool {
	switch value {
	case ProactivityLow, ProactivityMedium, ProactivityHigh:
		return true
	default:
		return false
	}
}

func IsValidAutonomyLevel(value string) bool {
	switch value {
	case AutonomyManual, AutonomySuggestive, AutonomySemiAutonomous:
		return true
	default:
		return false
	}
}
