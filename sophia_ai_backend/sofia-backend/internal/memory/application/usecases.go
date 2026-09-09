package application

import (
	"context"
	"errors"

	"github.com/armandoalvarado/sofia-backend/internal/memory/domain"
	usersdomain "github.com/armandoalvarado/sofia-backend/internal/users/domain"
	"github.com/google/uuid"
)

const DefaultListLimit = 50

type AISettingsReader interface {
	GetAISettings(userID string) (*usersdomain.AISettings, error)
}

type Service struct {
	repo       domain.MemoryRepository
	settings   AISettingsReader
	embeddings EmbeddingProvider
}

func NewService(repo domain.MemoryRepository, settings AISettingsReader, embeddings EmbeddingProvider) *Service {
	return &Service{repo: repo, settings: settings, embeddings: embeddings}
}

func (s *Service) CreateMemory(ctx context.Context, userID string, input domain.MemoryCreate) (*domain.Memory, error) {
	if err := s.ensureMemoryEnabled(userID); err != nil {
		return nil, err
	}
	memory, err := domain.NewMemory(uuid.New().String(), userID, input)
	if err != nil {
		return nil, err
	}
	if err := s.repo.Create(ctx, memory); err != nil {
		return nil, err
	}
	return memory, nil
}

func (s *Service) GetMemory(ctx context.Context, userID, memoryID string) (*domain.Memory, error) {
	if err := s.ensureMemoryEnabled(userID); err != nil {
		return nil, err
	}
	memory, err := s.repo.FindByID(ctx, userID, memoryID)
	if err != nil {
		return nil, err
	}
	if memory == nil {
		return nil, domain.ErrMemoryNotFound
	}
	if err := s.repo.TouchAccessed(ctx, userID, memoryID); err != nil {
		return nil, err
	}
	memory.TouchAccessed()
	return memory, nil
}

func (s *Service) ListMemories(ctx context.Context, filter domain.ListFilter) ([]*domain.Memory, error) {
	if err := s.ensureMemoryEnabled(filter.UserID); err != nil {
		return nil, err
	}
	if filter.Limit == 0 {
		filter.Limit = DefaultListLimit
	}
	if filter.Status == "" {
		filter.Status = domain.StatusActive
	}
	if err := domain.ValidateListFilter(filter); err != nil {
		return nil, err
	}
	return s.repo.List(ctx, filter)
}

func (s *Service) UpdateMemory(ctx context.Context, userID, memoryID string, update domain.MemoryUpdate) (*domain.Memory, error) {
	if err := s.ensureMemoryEnabled(userID); err != nil {
		return nil, err
	}
	memory, err := s.repo.FindByID(ctx, userID, memoryID)
	if err != nil {
		return nil, err
	}
	if memory == nil {
		return nil, domain.ErrMemoryNotFound
	}
	if err := memory.Apply(update); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, memory); err != nil {
		return nil, err
	}
	return memory, nil
}

func (s *Service) ArchiveMemory(ctx context.Context, userID, memoryID string) (*domain.Memory, error) {
	if err := s.ensureMemoryEnabled(userID); err != nil {
		return nil, err
	}
	return s.repo.Archive(ctx, userID, memoryID)
}

func (s *Service) DeleteMemory(ctx context.Context, userID, memoryID string) (*domain.Memory, error) {
	if err := s.ensureMemoryEnabled(userID); err != nil {
		return nil, err
	}
	return s.repo.DeleteSoft(ctx, userID, memoryID)
}

func (s *Service) SearchMemory(ctx context.Context, filter domain.SearchFilter) ([]*domain.Memory, error) {
	if err := s.ensureMemoryEnabled(filter.UserID); err != nil {
		return nil, err
	}
	if filter.Limit == 0 {
		filter.Limit = DefaultListLimit
	}
	if err := domain.ValidateSearchFilter(filter); err != nil {
		return nil, err
	}
	return s.repo.SearchBasic(ctx, filter)
}

func (s *Service) SearchByTerms(ctx context.Context, userID string, terms []string, limit int) ([]*domain.Memory, error) {
	if err := s.ensureMemoryEnabled(userID); err != nil {
		return nil, err
	}
	if len(terms) == 0 {
		return []*domain.Memory{}, nil
	}
	if limit <= 0 {
		limit = domain.DefaultSearchLimit()
	}
	return s.repo.SearchByTerms(ctx, userID, terms, limit)
}

func (s *Service) ensureMemoryEnabled(userID string) error {
	if s.settings == nil {
		return nil
	}
	settings, err := s.settings.GetAISettings(userID)
	if err != nil {
		if errors.Is(err, usersdomain.ErrAISettingsNotFound) {
			return nil
		}
		return err
	}
	if settings != nil && !settings.MemoryEnabled {
		return domain.ErrMemoryDisabled
	}
	return nil
}
