package infrastructure

import (
	"sync"

	"github.com/armandoalvarado/sofia-backend/internal/users/domain"
)

type InMemoryUserProfileRepository struct {
	mu       sync.Mutex
	profiles map[string]*domain.UserProfile
}

func NewInMemoryUserProfileRepository() *InMemoryUserProfileRepository {
	return &InMemoryUserProfileRepository{profiles: make(map[string]*domain.UserProfile)}
}

func (r *InMemoryUserProfileRepository) Save(profile *domain.UserProfile) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	cp := *profile
	r.profiles[profile.UserID] = &cp
	return nil
}

func (r *InMemoryUserProfileRepository) FindByUserID(userID string) (*domain.UserProfile, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	profile, ok := r.profiles[userID]
	if !ok {
		return nil, nil
	}
	cp := *profile
	return &cp, nil
}

type InMemoryAISettingsRepository struct {
	mu       sync.Mutex
	settings map[string]*domain.AISettings
}

func NewInMemoryAISettingsRepository() *InMemoryAISettingsRepository {
	return &InMemoryAISettingsRepository{settings: make(map[string]*domain.AISettings)}
}

func (r *InMemoryAISettingsRepository) Save(settings *domain.AISettings) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	cp := *settings
	r.settings[settings.UserID] = &cp
	return nil
}

func (r *InMemoryAISettingsRepository) FindByUserID(userID string) (*domain.AISettings, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	settings, ok := r.settings[userID]
	if !ok {
		return nil, nil
	}
	cp := *settings
	return &cp, nil
}
