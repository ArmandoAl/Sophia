package application

import (
	"context"

	"github.com/armandoalvarado/sofia-backend/internal/activities/domain"
	"github.com/google/uuid"
)

const DefaultListLimit = 50

type Service struct {
	repo      domain.ActivityRepository
	reminders ReminderBridge
}

func NewService(repo domain.ActivityRepository) *Service {
	return &Service{repo: repo}
}

type ReminderBridge interface {
	EnsureEnabled(userID string) error
	CreateForActivity(ctx context.Context, activity *domain.Activity) error
	CancelPendingForActivity(ctx context.Context, userID, activityID string) error
}

func (s *Service) SetReminderBridge(reminders ReminderBridge) {
	s.reminders = reminders
}

func (s *Service) CreateActivity(ctx context.Context, userID string, input domain.ActivityCreate) (*domain.Activity, error) {
	activity, err := domain.NewActivity(uuid.New().String(), userID, input)
	if err != nil {
		return nil, err
	}
	if s.reminders != nil && activity.HasReminder && activity.ReminderAt != nil {
		if err := s.reminders.EnsureEnabled(userID); err != nil {
			return nil, err
		}
	}
	if err := s.repo.Create(ctx, activity); err != nil {
		return nil, err
	}
	if s.reminders != nil {
		if err := s.reminders.CreateForActivity(ctx, activity); err != nil {
			return nil, err
		}
	}
	return activity, nil
}

func (s *Service) GetActivity(ctx context.Context, userID, activityID string) (*domain.Activity, error) {
	return s.repo.FindByID(ctx, userID, activityID)
}

func (s *Service) ListActivities(ctx context.Context, filter domain.ListFilter) ([]*domain.Activity, error) {
	if filter.Limit == 0 {
		filter.Limit = DefaultListLimit
	}
	if err := domain.ValidateFilter(filter); err != nil {
		return nil, err
	}
	return s.repo.List(ctx, filter)
}

func (s *Service) UpdateActivity(ctx context.Context, userID, activityID string, update domain.ActivityUpdate) (*domain.Activity, error) {
	activity, err := s.repo.FindByID(ctx, userID, activityID)
	if err != nil {
		return nil, err
	}
	if activity == nil {
		return nil, domain.ErrActivityNotFound
	}
	if err := activity.Apply(update); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, activity); err != nil {
		return nil, err
	}
	return activity, nil
}

func (s *Service) CompleteActivity(ctx context.Context, userID, activityID string) (*domain.Activity, error) {
	return s.repo.Complete(ctx, userID, activityID)
}

func (s *Service) CancelActivity(ctx context.Context, userID, activityID string) (*domain.Activity, error) {
	activity, err := s.repo.Cancel(ctx, userID, activityID)
	if err != nil {
		return nil, err
	}
	if s.reminders != nil {
		if err := s.reminders.CancelPendingForActivity(ctx, userID, activityID); err != nil {
			return nil, err
		}
	}
	return activity, nil
}

func (s *Service) ArchiveActivity(ctx context.Context, userID, activityID string) (*domain.Activity, error) {
	activity, err := s.repo.Archive(ctx, userID, activityID)
	if err != nil {
		return nil, err
	}
	if s.reminders != nil {
		if err := s.reminders.CancelPendingForActivity(ctx, userID, activityID); err != nil {
			return nil, err
		}
	}
	return activity, nil
}
