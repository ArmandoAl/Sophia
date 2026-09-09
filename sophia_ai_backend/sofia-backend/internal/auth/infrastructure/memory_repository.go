package infrastructure

import (
	"sync"

	"github.com/armandoalvarado/sofia-backend/internal/auth/domain"
)

type InMemoryUserRepository struct {
	mu           sync.Mutex
	usersByEmail map[string]*domain.User
	usersByID    map[string]*domain.User
}

func NewInMemoryUserRepository() *InMemoryUserRepository {
	return &InMemoryUserRepository{
		usersByEmail: make(map[string]*domain.User),
		usersByID:    make(map[string]*domain.User),
	}
}

func (r *InMemoryUserRepository) Save(user *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.usersByEmail[user.Email] = user
	r.usersByID[user.ID] = user
	return nil
}

func (r *InMemoryUserRepository) FindByEmail(email string) (*domain.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	user, ok := r.usersByEmail[email]
	if !ok {
		return nil, nil
	}
	return user, nil
}

func (r *InMemoryUserRepository) FindByID(id string) (*domain.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	user, ok := r.usersByID[id]
	if !ok {
		return nil, nil
	}
	return user, nil
}

func (r *InMemoryUserRepository) ListIDs() ([]string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	ids := make([]string, 0, len(r.usersByID))
	for id := range r.usersByID {
		ids = append(ids, id)
	}
	return ids, nil
}
