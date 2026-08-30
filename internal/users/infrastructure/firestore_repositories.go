package infrastructure

import (
	"context"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/armandoalvarado/sofia-backend/internal/users/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	userProfilesCollection = "user_profiles"
	aiSettingsCollection   = "ai_settings"
)

type FirestoreUserProfileRepository struct {
	client  *firestore.Client
	timeout time.Duration
}

type FirestoreAISettingsRepository struct {
	client  *firestore.Client
	timeout time.Duration
}

func NewFirestoreUserProfileRepository(client *firestore.Client) *FirestoreUserProfileRepository {
	return &FirestoreUserProfileRepository{client: client, timeout: 5 * time.Second}
}

func NewFirestoreAISettingsRepository(client *firestore.Client) *FirestoreAISettingsRepository {
	return &FirestoreAISettingsRepository{client: client, timeout: 5 * time.Second}
}

func (r *FirestoreUserProfileRepository) Save(profile *domain.UserProfile) error {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	_, err := r.client.Collection(userProfilesCollection).Doc(profile.UserID).Set(ctx, profileToDocument(profile))
	return err
}

func (r *FirestoreUserProfileRepository) FindByUserID(userID string) (*domain.UserProfile, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	doc, err := r.client.Collection(userProfilesCollection).Doc(userID).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, nil
		}
		return nil, err
	}

	var stored firestoreUserProfile
	if err := doc.DataTo(&stored); err != nil {
		return nil, err
	}
	return stored.toDomain(), nil
}

func (r *FirestoreAISettingsRepository) Save(settings *domain.AISettings) error {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	_, err := r.client.Collection(aiSettingsCollection).Doc(settings.UserID).Set(ctx, aiSettingsToDocument(settings))
	return err
}

func (r *FirestoreAISettingsRepository) FindByUserID(userID string) (*domain.AISettings, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	doc, err := r.client.Collection(aiSettingsCollection).Doc(userID).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, nil
		}
		return nil, err
	}

	var stored firestoreAISettings
	if err := doc.DataTo(&stored); err != nil {
		return nil, err
	}
	return stored.toDomain(), nil
}

type firestoreUserProfile struct {
	ID                  string    `firestore:"id"`
	UserID              string    `firestore:"user_id"`
	DisplayName         string    `firestore:"display_name"`
	PreferredName       string    `firestore:"preferred_name"`
	Timezone            string    `firestore:"timezone"`
	Locale              string    `firestore:"locale"`
	AvatarURL           string    `firestore:"avatar_url"`
	OnboardingCompleted bool      `firestore:"onboarding_completed"`
	CreatedAt           time.Time `firestore:"created_at"`
	UpdatedAt           time.Time `firestore:"updated_at"`
}

type firestoreAISettings struct {
	UserID             string    `firestore:"user_id"`
	AssistantName      string    `firestore:"assistant_name"`
	CommunicationStyle string    `firestore:"communication_style"`
	ProactivityLevel   string    `firestore:"proactivity_level"`
	AutonomyLevel      string    `firestore:"autonomy_level"`
	MemoryEnabled      bool      `firestore:"memory_enabled"`
	RemindersEnabled   bool      `firestore:"reminders_enabled"`
	PlanningEnabled    bool      `firestore:"planning_enabled"`
	PreferredLanguage  string    `firestore:"preferred_language"`
	CreatedAt          time.Time `firestore:"created_at"`
	UpdatedAt          time.Time `firestore:"updated_at"`
}

func profileToDocument(profile *domain.UserProfile) firestoreUserProfile {
	return firestoreUserProfile{
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

func (p firestoreUserProfile) toDomain() *domain.UserProfile {
	return &domain.UserProfile{
		ID:                  p.ID,
		UserID:              p.UserID,
		DisplayName:         p.DisplayName,
		PreferredName:       p.PreferredName,
		Timezone:            p.Timezone,
		Locale:              p.Locale,
		AvatarURL:           p.AvatarURL,
		OnboardingCompleted: p.OnboardingCompleted,
		CreatedAt:           p.CreatedAt,
		UpdatedAt:           p.UpdatedAt,
	}
}

func aiSettingsToDocument(settings *domain.AISettings) firestoreAISettings {
	return firestoreAISettings{
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

func (s firestoreAISettings) toDomain() *domain.AISettings {
	return &domain.AISettings{
		UserID:             s.UserID,
		AssistantName:      s.AssistantName,
		CommunicationStyle: s.CommunicationStyle,
		ProactivityLevel:   s.ProactivityLevel,
		AutonomyLevel:      s.AutonomyLevel,
		MemoryEnabled:      s.MemoryEnabled,
		RemindersEnabled:   s.RemindersEnabled,
		PlanningEnabled:    s.PlanningEnabled,
		PreferredLanguage:  s.PreferredLanguage,
		CreatedAt:          s.CreatedAt,
		UpdatedAt:          s.UpdatedAt,
	}
}
