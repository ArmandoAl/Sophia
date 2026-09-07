package worker

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	notificationsinfra "github.com/armandoalvarado/sofia-backend/internal/notifications/infrastructure"
	"github.com/armandoalvarado/sofia-backend/internal/reminders/domain"
	"github.com/armandoalvarado/sofia-backend/internal/reminders/infrastructure"
)

type recordingDeliveryProvider struct {
	mu        sync.Mutex
	err       error
	delivered []string
}

func (p *recordingDeliveryProvider) Deliver(ctx context.Context, reminder *domain.Reminder) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.delivered = append(p.delivered, reminder.ID)
	return p.err
}

func (p *recordingDeliveryProvider) count() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.delivered)
}

func TestWorkerProcessesDueReminders(t *testing.T) {
	now := time.Date(2026, 1, 2, 15, 0, 0, 0, time.UTC)
	repo := infrastructure.NewInMemoryReminderRepository()
	provider := &recordingDeliveryProvider{}
	reminder := mustReminder(t, "reminder-1", "user-1", now.Add(-time.Minute))
	mustCreate(t, repo, reminder)

	result, err := New(repo, provider, Options{Now: func() time.Time { return now }}).RunOnce(context.Background())
	if err != nil {
		t.Fatalf("RunOnce returned error: %v", err)
	}

	if result.Delivered != 1 || provider.count() != 1 {
		t.Fatalf("expected one delivered reminder, result=%+v delivered=%v", result, provider.delivered)
	}
	stored, _ := repo.FindByID(context.Background(), "user-1", "reminder-1")
	if stored.Status != domain.StatusDelivered {
		t.Fatalf("expected delivered status, got %q", stored.Status)
	}
}

func TestWorkerDoesNotProcessNotDueReminders(t *testing.T) {
	now := time.Date(2026, 1, 2, 15, 0, 0, 0, time.UTC)
	repo := infrastructure.NewInMemoryReminderRepository()
	provider := &recordingDeliveryProvider{}
	mustCreate(t, repo, mustReminder(t, "reminder-1", "user-1", now.Add(time.Hour)))

	result, err := New(repo, provider, Options{Now: func() time.Time { return now }}).RunOnce(context.Background())
	if err != nil {
		t.Fatalf("RunOnce returned error: %v", err)
	}

	if result.Scanned != 0 || provider.count() != 0 {
		t.Fatalf("expected no processing, result=%+v delivered=%v", result, provider.delivered)
	}
}

func TestWorkerRespectsMaxRetries(t *testing.T) {
	now := time.Date(2026, 1, 2, 15, 0, 0, 0, time.UTC)
	repo := infrastructure.NewInMemoryReminderRepository()
	provider := &recordingDeliveryProvider{}
	reminder := mustReminder(t, "reminder-1", "user-1", now.Add(-time.Minute))
	reminder.RetryCount = 3
	reminder.MaxRetries = 3
	mustCreate(t, repo, reminder)

	result, err := New(repo, provider, Options{Now: func() time.Time { return now }}).RunOnce(context.Background())
	if err != nil {
		t.Fatalf("RunOnce returned error: %v", err)
	}

	if result.Failed != 1 || provider.count() != 0 {
		t.Fatalf("expected max retries failure without delivery, result=%+v delivered=%v", result, provider.delivered)
	}
	stored, _ := repo.FindByID(context.Background(), "user-1", "reminder-1")
	if stored.Status != domain.StatusFailed || stored.FailureReason != "max retries reached" {
		t.Fatalf("unexpected failed reminder: %+v", stored)
	}
}

func TestWorkerMarksFailedWhenProviderFails(t *testing.T) {
	now := time.Date(2026, 1, 2, 15, 0, 0, 0, time.UTC)
	repo := infrastructure.NewInMemoryReminderRepository()
	provider := &recordingDeliveryProvider{err: errors.New("provider unavailable")}
	mustCreate(t, repo, mustReminder(t, "reminder-1", "user-1", now.Add(-time.Minute)))

	result, err := New(repo, provider, Options{Now: func() time.Time { return now }}).RunOnce(context.Background())
	if err != nil {
		t.Fatalf("RunOnce returned error: %v", err)
	}

	if result.Failed != 1 {
		t.Fatalf("expected one failed reminder, got %+v", result)
	}
	stored, _ := repo.FindByID(context.Background(), "user-1", "reminder-1")
	if stored.Status != domain.StatusFailed || stored.RetryCount != 1 {
		t.Fatalf("expected failed reminder with retry increment, got %+v", stored)
	}
}

func TestWorkerBasicIdempotency(t *testing.T) {
	now := time.Date(2026, 1, 2, 15, 0, 0, 0, time.UTC)
	repo := infrastructure.NewInMemoryReminderRepository()
	provider := &recordingDeliveryProvider{}
	mustCreate(t, repo, mustReminder(t, "reminder-1", "user-1", now.Add(-time.Minute)))
	worker := New(repo, provider, Options{Now: func() time.Time { return now }})

	if _, err := worker.RunOnce(context.Background()); err != nil {
		t.Fatalf("first RunOnce returned error: %v", err)
	}
	if _, err := worker.RunOnce(context.Background()); err != nil {
		t.Fatalf("second RunOnce returned error: %v", err)
	}

	if provider.count() != 1 {
		t.Fatalf("expected exactly one delivery, got %v", provider.delivered)
	}
}

func TestWorkerDoesNotDuplicateRecurringDeliveryOnRepeatedRun(t *testing.T) {
	now := time.Date(2026, 1, 2, 15, 0, 0, 0, time.UTC)
	repo := infrastructure.NewInMemoryReminderRepository()
	provider := &recordingDeliveryProvider{}
	reminder := mustReminder(t, "reminder-1", "user-1", now.Add(-time.Minute))
	reminder.RecurrenceRule = domain.RecurrenceDaily
	reminder.RecurrenceInterval = 1
	mustCreate(t, repo, reminder)
	worker := New(repo, provider, Options{Now: func() time.Time { return now }})

	if _, err := worker.RunOnce(context.Background()); err != nil {
		t.Fatalf("first RunOnce returned error: %v", err)
	}
	if _, err := worker.RunOnce(context.Background()); err != nil {
		t.Fatalf("second RunOnce returned error: %v", err)
	}

	if provider.count() != 1 {
		t.Fatalf("expected exactly one delivery for same due window, got %v", provider.delivered)
	}
	stored, _ := repo.FindByID(context.Background(), "user-1", "reminder-1")
	if stored.Status != domain.StatusPending || !stored.NextRunAt.After(now) {
		t.Fatalf("expected recurring reminder pending in future, got %+v", stored)
	}
}

func TestTwoWorkersDoNotProcessSameReminder(t *testing.T) {
	now := time.Date(2026, 1, 2, 15, 0, 0, 0, time.UTC)
	repo := infrastructure.NewInMemoryReminderRepository()
	provider := &recordingDeliveryProvider{}
	reminder := mustReminder(t, "reminder-1", "user-1", now.Add(-time.Minute))
	mustCreate(t, repo, reminder)

	workerA := New(repo, provider, Options{WorkerID: "worker-a", LeaseDuration: time.Minute, Now: func() time.Time { return now }})
	workerB := New(repo, provider, Options{WorkerID: "worker-b", LeaseDuration: time.Minute, Now: func() time.Time { return now }})

	var wg sync.WaitGroup
	wg.Add(2)
	outcomes := make(chan processOutcome, 2)
	go func() {
		defer wg.Done()
		outcomes <- workerA.processOne(context.Background(), reminder, now)
	}()
	go func() {
		defer wg.Done()
		outcomes <- workerB.processOne(context.Background(), reminder, now)
	}()
	wg.Wait()
	close(outcomes)

	delivered := 0
	for outcome := range outcomes {
		if outcome == outcomeDelivered {
			delivered++
		}
	}
	if delivered != 1 || provider.count() != 1 {
		t.Fatalf("expected one worker delivery, delivered=%d provider=%v", delivered, provider.delivered)
	}
}

func TestExpiredLeaseAllowsRetry(t *testing.T) {
	now := time.Date(2026, 1, 2, 15, 0, 0, 0, time.UTC)
	repo := infrastructure.NewInMemoryReminderRepository()
	reminder := mustReminder(t, "reminder-1", "user-1", now.Add(-time.Minute))
	mustCreate(t, repo, reminder)

	if _, err := repo.ClaimDue(context.Background(), "user-1", "reminder-1", "worker-a", now, now.Add(time.Minute)); err != nil {
		t.Fatalf("first ClaimDue returned error: %v", err)
	}
	if _, err := repo.ClaimDue(context.Background(), "user-1", "reminder-1", "worker-b", now.Add(30*time.Second), now.Add(2*time.Minute)); !errors.Is(err, domain.ErrReminderClaimed) {
		t.Fatalf("expected active lease claim error, got %v", err)
	}
	claimed, err := repo.ClaimDue(context.Background(), "user-1", "reminder-1", "worker-b", now.Add(2*time.Minute), now.Add(3*time.Minute))
	if err != nil {
		t.Fatalf("expired lease ClaimDue returned error: %v", err)
	}
	if claimed.ProcessingBy != "worker-b" {
		t.Fatalf("expected worker-b claim after expiry, got %+v", claimed)
	}
}

func TestDeliveredReminderIsNotClaimedAgain(t *testing.T) {
	now := time.Date(2026, 1, 2, 15, 0, 0, 0, time.UTC)
	repo := infrastructure.NewInMemoryReminderRepository()
	reminder := mustReminder(t, "reminder-1", "user-1", now.Add(-time.Minute))
	mustCreate(t, repo, reminder)

	claimed, err := repo.ClaimDue(context.Background(), "user-1", "reminder-1", "worker-a", now, now.Add(time.Minute))
	if err != nil {
		t.Fatalf("ClaimDue returned error: %v", err)
	}
	if claimed.ProcessingBy == "" || claimed.ProcessingUntil == nil || claimed.LastAttemptAt == nil {
		t.Fatalf("expected claim metadata, got %+v", claimed)
	}
	delivered, err := repo.MarkDelivered(context.Background(), "user-1", "reminder-1")
	if err != nil {
		t.Fatalf("MarkDelivered returned error: %v", err)
	}
	if delivered.ProcessingBy != "" || delivered.ProcessingUntil != nil {
		t.Fatalf("expected delivered reminder to clear lease, got %+v", delivered)
	}
	if _, err := repo.ClaimDue(context.Background(), "user-1", "reminder-1", "worker-b", now.Add(2*time.Minute), now.Add(3*time.Minute)); !errors.Is(err, domain.ErrReminderClaimed) {
		t.Fatalf("expected delivered reminder not claimable, got %v", err)
	}
}

func TestProviderFailureClearsLeaseAndMarksFailed(t *testing.T) {
	now := time.Date(2026, 1, 2, 15, 0, 0, 0, time.UTC)
	repo := infrastructure.NewInMemoryReminderRepository()
	provider := &recordingDeliveryProvider{err: errors.New("provider unavailable")}
	mustCreate(t, repo, mustReminder(t, "reminder-1", "user-1", now.Add(-time.Minute)))

	result, err := New(repo, provider, Options{WorkerID: "worker-a", LeaseDuration: time.Minute, Now: func() time.Time { return now }}).RunOnce(context.Background())
	if err != nil {
		t.Fatalf("RunOnce returned error: %v", err)
	}
	if result.Failed != 1 {
		t.Fatalf("expected failed result, got %+v", result)
	}
	stored, _ := repo.FindByID(context.Background(), "user-1", "reminder-1")
	if stored.Status != domain.StatusFailed || stored.ProcessingBy != "" || stored.ProcessingUntil != nil || stored.LastAttemptAt == nil {
		t.Fatalf("expected failed reminder with cleared lease and attempt timestamp, got %+v", stored)
	}
}

func TestWorkerRecordsDeliveryAttempt(t *testing.T) {
	now := time.Date(2026, 1, 2, 15, 0, 0, 0, time.UTC)
	repo := infrastructure.NewInMemoryReminderRepository()
	attempts := notificationsinfra.NewInMemoryDeliveryAttemptRepository()
	provider := &recordingDeliveryProvider{}
	mustCreate(t, repo, mustReminder(t, "reminder-1", "user-1", now.Add(-time.Minute)))

	result, err := New(repo, provider, Options{
		WorkerID:     "worker-a",
		ProviderName: "fcm",
		DryRun:       true,
		Attempts:     attempts,
		Now:          func() time.Time { return now },
	}).RunOnce(context.Background())
	if err != nil {
		t.Fatalf("RunOnce returned error: %v", err)
	}
	if result.Delivered != 1 {
		t.Fatalf("expected delivered result, got %+v", result)
	}
	storedAttempts := attempts.ListAll()
	if len(storedAttempts) != 1 {
		t.Fatalf("expected one attempt, got %+v", storedAttempts)
	}
	if storedAttempts[0].UserID != "user-1" || storedAttempts[0].ReminderID != "reminder-1" || storedAttempts[0].Provider != "fcm" || !storedAttempts[0].DryRun {
		t.Fatalf("unexpected attempt: %+v", storedAttempts[0])
	}
}

func mustReminder(t *testing.T, id, userID string, scheduledAt time.Time) *domain.Reminder {
	t.Helper()
	reminder, err := domain.NewReminder(id, userID, domain.ReminderCreate{
		Title:       "Test reminder",
		ScheduledAt: scheduledAt,
		Timezone:    "UTC",
	})
	if err != nil {
		t.Fatalf("NewReminder returned error: %v", err)
	}
	return reminder
}

func mustCreate(t *testing.T, repo *infrastructure.InMemoryReminderRepository, reminder *domain.Reminder) {
	t.Helper()
	if err := repo.Create(context.Background(), reminder); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
}
