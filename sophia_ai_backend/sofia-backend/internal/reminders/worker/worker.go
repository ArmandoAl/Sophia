package worker

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	notificationsdomain "github.com/armandoalvarado/sofia-backend/internal/notifications/domain"
	"github.com/armandoalvarado/sofia-backend/internal/reminders/domain"
	"github.com/google/uuid"
)

const (
	defaultBatchSize     = 50
	defaultLeaseDuration = 2 * time.Minute
	defaultWorkerID      = "reminder-worker"
)

type ReminderRepository interface {
	ListDueAll(ctx context.Context, now time.Time, limit int) ([]*domain.Reminder, error)
	FindByID(ctx context.Context, userID, reminderID string) (*domain.Reminder, error)
	ClaimDue(ctx context.Context, userID, reminderID, workerID string, now, leaseUntil time.Time) (*domain.Reminder, error)
	MarkDelivered(ctx context.Context, userID, reminderID string) (*domain.Reminder, error)
	MarkFailed(ctx context.Context, userID, reminderID, reason string) (*domain.Reminder, error)
}

type ReminderDeliveryProvider interface {
	Deliver(ctx context.Context, reminder *domain.Reminder) error
}

type DeliveryAttemptRecorder interface {
	Create(ctx context.Context, attempt *notificationsdomain.DeliveryAttempt) error
}

type NoopDeliveryProvider struct{}

func (NoopDeliveryProvider) Deliver(ctx context.Context, reminder *domain.Reminder) error {
	return nil
}

type Options struct {
	BatchSize     int
	WorkerID      string
	LeaseDuration time.Duration
	ProviderName  string
	DryRun        bool
	Attempts      DeliveryAttemptRecorder
	Now           func() time.Time
	Logger        *log.Logger
}

type Worker struct {
	repo     ReminderRepository
	delivery ReminderDeliveryProvider
	options  Options
	inFlight sync.Map
}

type RunResult struct {
	Scanned   int
	Delivered int
	Failed    int
	Skipped   int
}

func New(repo ReminderRepository, delivery ReminderDeliveryProvider, options Options) *Worker {
	if delivery == nil {
		delivery = NoopDeliveryProvider{}
	}
	if options.BatchSize <= 0 {
		options.BatchSize = defaultBatchSize
	}
	if options.WorkerID == "" {
		options.WorkerID = defaultWorkerID
	}
	if options.LeaseDuration <= 0 {
		options.LeaseDuration = defaultLeaseDuration
	}
	if options.ProviderName == "" {
		options.ProviderName = notificationsdomain.DeliveryProviderNoop
	}
	if options.Now == nil {
		options.Now = time.Now
	}
	return &Worker{repo: repo, delivery: delivery, options: options}
}

func (w *Worker) RunOnce(ctx context.Context) (RunResult, error) {
	now := w.options.Now()
	reminders, err := w.repo.ListDueAll(ctx, now, w.options.BatchSize)
	if err != nil {
		return RunResult{}, err
	}

	result := RunResult{Scanned: len(reminders)}
	for _, reminder := range reminders {
		outcome := w.processOne(ctx, reminder, now)
		switch outcome {
		case outcomeDelivered:
			result.Delivered++
		case outcomeFailed:
			result.Failed++
		case outcomeSkipped:
			result.Skipped++
		}
	}
	return result, nil
}

type processOutcome int

const (
	outcomeSkipped processOutcome = iota
	outcomeDelivered
	outcomeFailed
)

func (w *Worker) processOne(ctx context.Context, reminder *domain.Reminder, now time.Time) processOutcome {
	if reminder == nil {
		return outcomeSkipped
	}

	key := reminder.UserID + "/" + reminder.ID
	if _, loaded := w.inFlight.LoadOrStore(key, struct{}{}); loaded {
		return outcomeSkipped
	}
	defer w.inFlight.Delete(key)

	current, err := w.repo.ClaimDue(ctx, reminder.UserID, reminder.ID, w.options.WorkerID, now, now.Add(w.options.LeaseDuration))
	if errors.Is(err, domain.ErrReminderClaimed) {
		w.logf("skip reminder %s: already claimed", reminder.ID)
		return outcomeSkipped
	}
	if err != nil || current == nil {
		w.logf("skip reminder %s: stale or unavailable", reminder.ID)
		return outcomeSkipped
	}
	if current.MaxRetries > 0 && current.RetryCount >= current.MaxRetries {
		if _, err := w.repo.MarkFailed(ctx, current.UserID, current.ID, "max retries reached"); err != nil {
			w.logf("mark reminder %s failed: %v", current.ID, err)
		}
		w.recordAttempt(ctx, current, notificationsdomain.DeliveryStatusFailed, "max_retries_reached")
		return outcomeFailed
	}

	if err := w.delivery.Deliver(ctx, current); err != nil {
		if _, markErr := w.repo.MarkFailed(ctx, current.UserID, current.ID, safeFailureReason(err)); markErr != nil {
			w.logf("mark reminder %s failed: %v", current.ID, markErr)
		}
		w.recordAttempt(ctx, current, notificationsdomain.DeliveryStatusFailed, errorType(err))
		return outcomeFailed
	}

	if _, err := w.repo.MarkDelivered(ctx, current.UserID, current.ID); err != nil {
		w.logf("mark reminder %s delivered: %v", current.ID, err)
		w.recordAttempt(ctx, current, notificationsdomain.DeliveryStatusFailed, errorType(err))
		return outcomeFailed
	}
	w.recordAttempt(ctx, current, notificationsdomain.DeliveryStatusDelivered, "")
	return outcomeDelivered
}

func (w *Worker) recordAttempt(ctx context.Context, reminder *domain.Reminder, status, errorType string) {
	if w.options.Attempts == nil || reminder == nil {
		return
	}
	attempt := notificationsdomain.NewDeliveryAttempt(uuid.NewString(), reminder.UserID, reminder.ID, w.options.ProviderName, status, w.options.DryRun, errorType)
	if err := w.options.Attempts.Create(ctx, attempt); err != nil {
		w.logf("record delivery attempt for reminder %s: %v", reminder.ID, err)
	}
}

func (w *Worker) logf(format string, args ...any) {
	if w.options.Logger != nil {
		w.options.Logger.Printf(format, args...)
	}
}

func safeFailureReason(err error) string {
	if err == nil {
		return ""
	}
	reason := err.Error()
	if reason == "" {
		return "delivery failed"
	}
	const maxReasonLength = 500
	if len(reason) > maxReasonLength {
		return fmt.Sprintf("%s...", reason[:maxReasonLength])
	}
	return reason
}

func errorType(err error) string {
	if err == nil {
		return ""
	}
	var fcmErr *FCMError
	if errors.As(err, &fcmErr) && fcmErr.Type != "" {
		return fcmErr.Type
	}
	switch {
	case errors.Is(err, ErrFCMProjectID):
		return FCMErrorAuthError
	case errors.Is(err, ErrFCMNoDeviceToken):
		return FCMErrorInvalidToken
	case errors.Is(err, ErrFCMDisabled):
		return "fcm_disabled"
	default:
		return "delivery_error"
	}
}
