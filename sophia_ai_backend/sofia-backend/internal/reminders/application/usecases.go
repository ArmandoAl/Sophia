package application

import (
	"context"
	"errors"
	"time"

	activitiesdomain "github.com/armandoalvarado/sofia-backend/internal/activities/domain"
	"github.com/armandoalvarado/sofia-backend/internal/reminders/domain"
	usersdomain "github.com/armandoalvarado/sofia-backend/internal/users/domain"
	"github.com/google/uuid"
)

const DefaultListLimit = 50

type AISettingsReader interface {
	GetAISettings(userID string) (*usersdomain.AISettings, error)
}

type Service struct {
	repo       domain.ReminderRepository
	activities activitiesdomain.ActivityRepository
	settings   AISettingsReader
}

func NewService(
	repo domain.ReminderRepository,
	activities activitiesdomain.ActivityRepository,
	settings AISettingsReader,
) *Service {
	return &Service{repo: repo, activities: activities, settings: settings}
}

func (s *Service) EnsureEnabled(userID string) error {
	return s.ensureRemindersEnabled(userID)
}

func (s *Service) CreateReminder(ctx context.Context, userID string, input domain.ReminderCreate) (*domain.Reminder, error) {
	if err := s.ensureRemindersEnabled(userID); err != nil {
		return nil, err
	}
	if err := s.ensureActivityBelongsToUser(ctx, userID, input.ActivityID); err != nil {
		return nil, err
	}

	reminder, err := domain.NewReminder(uuid.New().String(), userID, input)
	if err != nil {
		return nil, err
	}
	if err := s.repo.Create(ctx, reminder); err != nil {
		return nil, err
	}
	return reminder, nil
}

func (s *Service) CreateForActivity(ctx context.Context, activity *activitiesdomain.Activity) error {
	if activity == nil || !activity.HasReminder || activity.ReminderAt == nil {
		return nil
	}
	input := domain.ReminderCreate{
		ActivityID:     activity.ID,
		Title:          activity.Title,
		Description:    activity.Description,
		ScheduledAt:    *activity.ReminderAt,
		Timezone:       activity.Timezone,
		RecurrenceRule: activity.RecurrenceRule,
		Source:         domain.SourceActivity,
		CreatedBy:      domain.CreatedByUser,
	}
	_, err := s.CreateReminder(ctx, activity.UserID, input)
	return err
}

func (s *Service) GetReminder(ctx context.Context, userID, reminderID string) (*domain.Reminder, error) {
	reminder, err := s.repo.FindByID(ctx, userID, reminderID)
	if err != nil {
		return nil, err
	}
	if reminder == nil {
		return nil, domain.ErrReminderNotFound
	}
	return reminder, nil
}

func (s *Service) ListReminders(ctx context.Context, filter domain.ListFilter) ([]*domain.Reminder, error) {
	if filter.Limit == 0 {
		filter.Limit = DefaultListLimit
	}
	if err := domain.ValidateFilter(filter); err != nil {
		return nil, err
	}
	return s.repo.List(ctx, filter)
}

func (s *Service) UpdateReminder(ctx context.Context, userID, reminderID string, update domain.ReminderUpdate) (*domain.Reminder, error) {
	reminder, err := s.GetReminder(ctx, userID, reminderID)
	if err != nil {
		return nil, err
	}
	if err := reminder.Apply(update); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, reminder); err != nil {
		return nil, err
	}
	return reminder, nil
}

func (s *Service) CancelReminder(ctx context.Context, userID, reminderID string) (*domain.Reminder, error) {
	return s.repo.Cancel(ctx, userID, reminderID)
}

func (s *Service) ArchiveReminder(ctx context.Context, userID, reminderID string) (*domain.Reminder, error) {
	return s.repo.Archive(ctx, userID, reminderID)
}

func (s *Service) ListDueReminders(ctx context.Context, userID string, now time.Time, limit int) ([]*domain.Reminder, error) {
	if limit == 0 {
		limit = DefaultListLimit
	}
	if limit < 0 {
		return nil, domain.ErrInvalidLimit
	}
	return s.repo.ListDue(ctx, userID, now, limit)
}

func (s *Service) MarkReminderDelivered(ctx context.Context, userID, reminderID string) (*domain.Reminder, error) {
	return s.repo.MarkDelivered(ctx, userID, reminderID)
}

func (s *Service) MarkReminderFailed(ctx context.Context, userID, reminderID, reason string) (*domain.Reminder, error) {
	return s.repo.MarkFailed(ctx, userID, reminderID, reason)
}

func (s *Service) CancelPendingForActivity(ctx context.Context, userID, activityID string) error {
	reminders, err := s.repo.List(ctx, domain.ListFilter{
		UserID:     userID,
		ActivityID: activityID,
		Status:     domain.StatusPending,
		Limit:      DefaultListLimit,
	})
	if err != nil {
		return err
	}
	for _, reminder := range reminders {
		if _, err := s.repo.Cancel(ctx, userID, reminder.ID); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) ensureRemindersEnabled(userID string) error {
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
	if settings != nil && !settings.RemindersEnabled {
		return domain.ErrRemindersDisabled
	}
	return nil
}

func (s *Service) ensureActivityBelongsToUser(ctx context.Context, userID, activityID string) error {
	if activityID == "" || s.activities == nil {
		return nil
	}
	activity, err := s.activities.FindByID(ctx, userID, activityID)
	if err != nil {
		return err
	}
	if activity == nil {
		return domain.ErrActivityNotAllowed
	}
	return nil
}
