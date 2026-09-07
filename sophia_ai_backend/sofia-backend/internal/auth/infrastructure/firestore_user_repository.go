package infrastructure

import (
	"context"
	"errors"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/armandoalvarado/sofia-backend/internal/auth/domain"
	usersdomain "github.com/armandoalvarado/sofia-backend/internal/users/domain"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const usersCollection = "users"
const userProfilesCollection = "user_profiles"
const aiSettingsCollection = "ai_settings"

type FirestoreUserRepository struct {
	client  *firestore.Client
	timeout time.Duration
}

type firestoreUser struct {
	ID           string    `firestore:"id"`
	Name         string    `firestore:"name"`
	Email        string    `firestore:"email"`
	PasswordHash string    `firestore:"password_hash"`
	Role         string    `firestore:"role"`
	CreatedAt    time.Time `firestore:"created_at"`
	UpdatedAt    time.Time `firestore:"updated_at"`
}

func NewFirestoreUserRepository(client *firestore.Client) *FirestoreUserRepository {
	return &FirestoreUserRepository{
		client:  client,
		timeout: 5 * time.Second,
	}
}

func (r *FirestoreUserRepository) Save(user *domain.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	user.Email = normalizeEmail(user.Email)

	existing, err := r.FindByEmail(user.Email)
	if err != nil {
		return err
	}
	if existing != nil && existing.ID != user.ID {
		return domain.ErrUserAlreadyExists
	}

	doc := domainToFirestoreUser(user)
	_, err = r.client.Collection(usersCollection).Doc(user.ID).Set(ctx, doc)
	return err
}

func (r *FirestoreUserRepository) SaveWithDefaults(user *domain.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	user.Email = normalizeEmail(user.Email)
	profile := usersdomain.NewDefaultUserProfile(user.ID, user.Name)
	settings := usersdomain.NewDefaultAISettings(user.ID)

	return r.client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		query := r.client.Collection(usersCollection).
			Where("email", "==", user.Email).
			Limit(1)
		iter := tx.Documents(query)
		defer iter.Stop()

		existing, err := iter.Next()
		if err == nil && existing.Ref.ID != user.ID {
			return domain.ErrUserAlreadyExists
		}
		if err != nil && !errors.Is(err, iterator.Done) {
			return err
		}

		userRef := r.client.Collection(usersCollection).Doc(user.ID)
		profileRef := r.client.Collection(userProfilesCollection).Doc(user.ID)
		settingsRef := r.client.Collection(aiSettingsCollection).Doc(user.ID)

		if err := tx.Set(userRef, domainToFirestoreUser(user)); err != nil {
			return err
		}
		if err := tx.Set(profileRef, firestoreUserProfileFromDomain(profile)); err != nil {
			return err
		}
		return tx.Set(settingsRef, firestoreAISettingsFromDomain(settings))
	})
}

func (r *FirestoreUserRepository) FindByEmail(email string) (*domain.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	iter := r.client.Collection(usersCollection).
		Where("email", "==", normalizeEmail(email)).
		Limit(1).
		Documents(ctx)
	defer iter.Stop()

	doc, err := iter.Next()
	if errors.Is(err, iterator.Done) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return documentToDomainUser(doc)
}

func (r *FirestoreUserRepository) FindByID(id string) (*domain.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	doc, err := r.client.Collection(usersCollection).Doc(id).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, nil
		}
		return nil, err
	}

	return documentToDomainUser(doc)
}

func domainToFirestoreUser(user *domain.User) firestoreUser {
	return firestoreUser{
		ID:           user.ID,
		Name:         user.Name,
		Email:        normalizeEmail(user.Email),
		PasswordHash: user.Password,
		Role:         user.Role,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}
}

func documentToDomainUser(doc *firestore.DocumentSnapshot) (*domain.User, error) {
	var stored firestoreUser
	if err := doc.DataTo(&stored); err != nil {
		return nil, err
	}
	return firestoreUserToDomain(stored), nil
}

func firestoreUserToDomain(stored firestoreUser) *domain.User {
	return &domain.User{
		ID:        stored.ID,
		Name:      stored.Name,
		Email:     normalizeEmail(stored.Email),
		Password:  stored.PasswordHash,
		Role:      stored.Role,
		CreatedAt: stored.CreatedAt,
		UpdatedAt: stored.UpdatedAt,
	}
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

func firestoreUserProfileFromDomain(profile *usersdomain.UserProfile) firestoreUserProfile {
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

func firestoreAISettingsFromDomain(settings *usersdomain.AISettings) firestoreAISettings {
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

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
