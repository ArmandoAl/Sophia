package application

import (
	"context"

	"github.com/armandoalvarado/sofia-backend/internal/notifications/domain"
	"github.com/google/uuid"
)

type Service struct {
	repo domain.DeviceTokenRepository
}

func NewService(repo domain.DeviceTokenRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) RegisterDeviceToken(ctx context.Context, userID string, input domain.DeviceTokenCreate) (*domain.DeviceToken, error) {
	token, err := domain.NewDeviceToken(uuid.NewString(), userID, input)
	if err != nil {
		return nil, err
	}
	if err := s.repo.Save(ctx, token); err != nil {
		return nil, err
	}
	return token, nil
}

func (s *Service) ListDeviceTokens(ctx context.Context, userID string) ([]*domain.DeviceToken, error) {
	return s.repo.ListByUser(ctx, userID)
}

func (s *Service) DeleteDeviceToken(ctx context.Context, userID, tokenID string) error {
	return s.repo.Delete(ctx, userID, tokenID)
}
