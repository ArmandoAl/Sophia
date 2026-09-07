package application

import (
	"time"

	authdomain "github.com/armandoalvarado/sofia-backend/internal/auth/domain"
	"github.com/armandoalvarado/sofia-backend/internal/users/domain"
)

type Service struct {
	authUsers  authdomain.UserRepository
	profiles   domain.UserProfileRepository
	aiSettings domain.AISettingsRepository
}

type Me struct {
	User       *authdomain.User
	Profile    *domain.UserProfile
	AISettings *domain.AISettings
}

func NewService(
	authUsers authdomain.UserRepository,
	profiles domain.UserProfileRepository,
	aiSettings domain.AISettingsRepository,
) *Service {
	return &Service{
		authUsers:  authUsers,
		profiles:   profiles,
		aiSettings: aiSettings,
	}
}

func (s *Service) ProvisionDefaults(user *authdomain.User) error {
	profile := domain.NewDefaultUserProfile(user.ID, user.Name)
	if err := s.profiles.Save(profile); err != nil {
		return err
	}

	settings := domain.NewDefaultAISettings(user.ID)
	if err := s.aiSettings.Save(settings); err != nil {
		return err
	}

	return nil
}

func (s *Service) GetMe(userID string) (*Me, error) {
	user, err := s.authUsers.FindByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, authdomain.ErrUserNotFound
	}

	profile, err := s.profiles.FindByUserID(userID)
	if err != nil {
		return nil, err
	}
	if profile == nil {
		profile = domain.NewDefaultUserProfile(user.ID, user.Name)
		if err := s.profiles.Save(profile); err != nil {
			return nil, err
		}
	}

	settings, err := s.aiSettings.FindByUserID(userID)
	if err != nil {
		return nil, err
	}
	if settings == nil {
		settings = domain.NewDefaultAISettings(user.ID)
		if err := s.aiSettings.Save(settings); err != nil {
			return nil, err
		}
	}

	return &Me{User: user, Profile: profile, AISettings: settings}, nil
}

func (s *Service) UpdateProfile(userID string, update domain.ProfileUpdate) (*domain.UserProfile, error) {
	profile, err := s.profiles.FindByUserID(userID)
	if err != nil {
		return nil, err
	}
	if profile == nil {
		return nil, domain.ErrProfileNotFound
	}

	if err := profile.Apply(update); err != nil {
		return nil, err
	}
	profile.UpdatedAt = time.Now()

	if err := s.profiles.Save(profile); err != nil {
		return nil, err
	}
	return profile, nil
}

func (s *Service) GetAISettings(userID string) (*domain.AISettings, error) {
	settings, err := s.aiSettings.FindByUserID(userID)
	if err != nil {
		return nil, err
	}
	if settings == nil {
		return nil, domain.ErrAISettingsNotFound
	}
	return settings, nil
}

func (s *Service) UpdateAISettings(userID string, update domain.AISettingsUpdate) (*domain.AISettings, error) {
	settings, err := s.aiSettings.FindByUserID(userID)
	if err != nil {
		return nil, err
	}
	if settings == nil {
		return nil, domain.ErrAISettingsNotFound
	}

	if err := settings.Apply(update); err != nil {
		return nil, err
	}
	settings.UpdatedAt = time.Now()

	if err := s.aiSettings.Save(settings); err != nil {
		return nil, err
	}
	return settings, nil
}

func (s *Service) CompleteOnboarding(userID string) (*domain.UserProfile, error) {
	profile, err := s.profiles.FindByUserID(userID)
	if err != nil {
		return nil, err
	}
	if profile == nil {
		return nil, domain.ErrProfileNotFound
	}

	profile.CompleteOnboarding()
	if err := s.profiles.Save(profile); err != nil {
		return nil, err
	}
	return profile, nil
}
