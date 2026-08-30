package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/armandoalvarado/sofia-backend/internal/reminders/domain"
	"github.com/armandoalvarado/sofia-backend/internal/reminders/infrastructure"
	"github.com/armandoalvarado/sofia-backend/internal/reminders/worker"
)

type scenarioResult struct {
	Name   string
	Status string
	Detail string
}

type operationalReport struct {
	GeneratedAt time.Time
	Scenarios   []scenarioResult
}

func main() {
	outputDir := flag.String("output-dir", "docs", "directory where reminder_worker_operational_report.md is written")
	flag.Parse()

	report, err := runOperationalValidation(context.Background(), *outputDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "reminder worker lease validation failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("wrote %s with %d scenarios\n", filepath.Join(*outputDir, "reminder_worker_operational_report.md"), len(report.Scenarios))
}

func runOperationalValidation(ctx context.Context, outputDir string) (operationalReport, error) {
	report := operationalReport{GeneratedAt: time.Now().UTC()}
	scenarios := []func(context.Context) scenarioResult{
		validateActiveLeaseBlocksSecondWorker,
		validateExpiredLeaseAllowsRetry,
		validateRecurringReminderMovesNextRunAt,
		validateFailedReminderDoesNotDuplicate,
	}
	for _, scenario := range scenarios {
		report.Scenarios = append(report.Scenarios, scenario(ctx))
	}
	if err := writeOperationalReport(outputDir, report); err != nil {
		return report, err
	}
	for _, scenario := range report.Scenarios {
		if scenario.Status != "pass" {
			return report, fmt.Errorf("%s: %s", scenario.Name, scenario.Detail)
		}
	}
	return report, nil
}

func validateActiveLeaseBlocksSecondWorker(ctx context.Context) scenarioResult {
	now := fixedNow()
	repo := infrastructure.NewInMemoryReminderRepository()
	provider := newBlockingDeliveryProvider()
	reminder := syntheticReminder("lease-active", "user-1", now.Add(-time.Minute))
	if err := repo.Create(ctx, reminder); err != nil {
		return failScenario("active lease blocks second worker", err.Error())
	}

	workerA := worker.New(repo, provider, worker.Options{WorkerID: "worker-a", LeaseDuration: time.Minute, Now: func() time.Time { return now }})
	workerB := worker.New(repo, provider, worker.Options{WorkerID: "worker-b", LeaseDuration: time.Minute, Now: func() time.Time { return now }})

	done := make(chan worker.RunResult, 1)
	go func() {
		result, _ := workerA.RunOnce(ctx)
		done <- result
	}()

	if !provider.waitStarted(time.Second) {
		return failScenario("active lease blocks second worker", "worker-a did not enter delivery")
	}
	second, err := workerB.RunOnce(ctx)
	if err != nil {
		provider.release()
		return failScenario("active lease blocks second worker", err.Error())
	}
	provider.release()
	first := <-done

	stored, _ := repo.FindByID(ctx, "user-1", "lease-active")
	if provider.count() != 1 || first.Delivered != 1 || second.Delivered != 0 || second.Skipped != 1 || stored == nil || stored.Status != domain.StatusDelivered {
		return failScenario("active lease blocks second worker", fmt.Sprintf("first=%+v second=%+v deliveries=%d stored=%+v", first, second, provider.count(), stored))
	}
	return passScenario("active lease blocks second worker", "worker-b saw the due candidate but ClaimDue rejected the active lease; only worker-a delivered")
}

func validateExpiredLeaseAllowsRetry(ctx context.Context) scenarioResult {
	now := fixedNow()
	repo := infrastructure.NewInMemoryReminderRepository()
	reminder := syntheticReminder("lease-expired", "user-1", now.Add(-time.Minute))
	if err := repo.Create(ctx, reminder); err != nil {
		return failScenario("expired lease allows retry", err.Error())
	}
	if _, err := repo.ClaimDue(ctx, "user-1", "lease-expired", "worker-a", now, now.Add(time.Minute)); err != nil {
		return failScenario("expired lease allows retry", err.Error())
	}
	if _, err := repo.ClaimDue(ctx, "user-1", "lease-expired", "worker-b", now.Add(30*time.Second), now.Add(2*time.Minute)); !errors.Is(err, domain.ErrReminderClaimed) {
		return failScenario("expired lease allows retry", fmt.Sprintf("expected active lease rejection, got %v", err))
	}
	claimed, err := repo.ClaimDue(ctx, "user-1", "lease-expired", "worker-b", now.Add(2*time.Minute), now.Add(3*time.Minute))
	if err != nil {
		return failScenario("expired lease allows retry", err.Error())
	}
	if claimed.ProcessingBy != "worker-b" {
		return failScenario("expired lease allows retry", fmt.Sprintf("expected worker-b, got %+v", claimed))
	}
	return passScenario("expired lease allows retry", "active lease rejected retry before expiry; expired lease was claimed by worker-b")
}

func validateRecurringReminderMovesNextRunAt(ctx context.Context) scenarioResult {
	now := fixedNow()
	repo := infrastructure.NewInMemoryReminderRepository()
	provider := &countingDeliveryProvider{}
	reminder := syntheticReminder("recurring-daily", "user-1", now.Add(-time.Minute))
	reminder.RecurrenceRule = domain.RecurrenceDaily
	reminder.RecurrenceInterval = 1
	if err := repo.Create(ctx, reminder); err != nil {
		return failScenario("recurring reminder moves next_run_at", err.Error())
	}

	result, err := worker.New(repo, provider, worker.Options{WorkerID: "worker-a", Now: func() time.Time { return now }}).RunOnce(ctx)
	if err != nil {
		return failScenario("recurring reminder moves next_run_at", err.Error())
	}
	stored, _ := repo.FindByID(ctx, "user-1", "recurring-daily")
	expected := reminder.NextRunAt.AddDate(0, 0, 1)
	if result.Delivered != 1 || provider.count() != 1 || stored == nil || stored.Status != domain.StatusPending || !stored.NextRunAt.Equal(expected) {
		return failScenario("recurring reminder moves next_run_at", fmt.Sprintf("result=%+v deliveries=%d expected_next=%s stored=%+v", result, provider.count(), expected.Format(time.RFC3339), stored))
	}
	return passScenario("recurring reminder moves next_run_at", fmt.Sprintf("next_run_at advanced to %s and status remained pending", stored.NextRunAt.Format(time.RFC3339)))
}

func validateFailedReminderDoesNotDuplicate(ctx context.Context) scenarioResult {
	now := fixedNow()
	repo := infrastructure.NewInMemoryReminderRepository()
	provider := &countingDeliveryProvider{err: errors.New("synthetic provider failure")}
	reminder := syntheticReminder("failed-once", "user-1", now.Add(-time.Minute))
	if err := repo.Create(ctx, reminder); err != nil {
		return failScenario("failed reminder does not duplicate", err.Error())
	}
	w := worker.New(repo, provider, worker.Options{WorkerID: "worker-a", Now: func() time.Time { return now }})
	first, err := w.RunOnce(ctx)
	if err != nil {
		return failScenario("failed reminder does not duplicate", err.Error())
	}
	second, err := w.RunOnce(ctx)
	if err != nil {
		return failScenario("failed reminder does not duplicate", err.Error())
	}
	stored, _ := repo.FindByID(ctx, "user-1", "failed-once")
	if first.Failed != 1 || second.Scanned != 0 || provider.count() != 1 || stored == nil || stored.Status != domain.StatusFailed || stored.RetryCount != 1 {
		return failScenario("failed reminder does not duplicate", fmt.Sprintf("first=%+v second=%+v deliveries=%d stored=%+v", first, second, provider.count(), stored))
	}
	return passScenario("failed reminder does not duplicate", "failed reminder moved out of pending scan after one provider failure")
}

func writeOperationalReport(outputDir string, report operationalReport) error {
	if strings.TrimSpace(outputDir) == "" {
		outputDir = "docs"
	}
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return err
	}
	var b strings.Builder
	b.WriteString("# Reminder Worker Operational Report\n\n")
	b.WriteString(fmt.Sprintf("Generated at: `%s`\n\n", report.GeneratedAt.Format(time.RFC3339)))
	b.WriteString("## Scope\n\n")
	b.WriteString("This command runs deterministic synthetic scenarios against the reminder worker lease logic. It does not send notifications and does not require production Firestore credentials.\n\n")
	b.WriteString("## Results\n\n")
	b.WriteString("| Scenario | Status | Detail |\n")
	b.WriteString("| --- | --- | --- |\n")
	for _, scenario := range report.Scenarios {
		b.WriteString(fmt.Sprintf("| %s | %s | %s |\n", scenario.Name, scenario.Status, escapeMarkdownTable(scenario.Detail)))
	}
	b.WriteString("\n## Notes\n\n")
	b.WriteString("- The simulation uses the same worker and repository lease contract used by runtime code.\n")
	b.WriteString("- Firestore production correctness still depends on the transactional `ClaimDue` implementation and emulator coverage.\n")
	b.WriteString("- No visible notifications are sent.\n")
	return os.WriteFile(filepath.Join(outputDir, "reminder_worker_operational_report.md"), []byte(b.String()), 0o644)
}

func syntheticReminder(id, userID string, scheduledAt time.Time) *domain.Reminder {
	reminder, err := domain.NewReminder(id, userID, domain.ReminderCreate{
		Title:       "Synthetic reminder " + id,
		ScheduledAt: scheduledAt,
		Timezone:    "UTC",
	})
	if err != nil {
		panic(err)
	}
	return reminder
}

func fixedNow() time.Time {
	return time.Date(2026, 1, 2, 15, 0, 0, 0, time.UTC)
}

func passScenario(name, detail string) scenarioResult {
	return scenarioResult{Name: name, Status: "pass", Detail: detail}
}

func failScenario(name, detail string) scenarioResult {
	return scenarioResult{Name: name, Status: "fail", Detail: detail}
}

func escapeMarkdownTable(value string) string {
	return strings.ReplaceAll(value, "|", "\\|")
}

type countingDeliveryProvider struct {
	mu     sync.Mutex
	err    error
	counts map[string]int
}

func (p *countingDeliveryProvider) Deliver(ctx context.Context, reminder *domain.Reminder) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.counts == nil {
		p.counts = make(map[string]int)
	}
	p.counts[reminder.ID]++
	return p.err
}

func (p *countingDeliveryProvider) count() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	total := 0
	for _, count := range p.counts {
		total += count
	}
	return total
}

type blockingDeliveryProvider struct {
	mu       sync.Mutex
	counts   map[string]int
	started  chan struct{}
	releaseC chan struct{}
	once     sync.Once
}

func newBlockingDeliveryProvider() *blockingDeliveryProvider {
	return &blockingDeliveryProvider{
		counts:   make(map[string]int),
		started:  make(chan struct{}),
		releaseC: make(chan struct{}),
	}
}

func (p *blockingDeliveryProvider) Deliver(ctx context.Context, reminder *domain.Reminder) error {
	p.mu.Lock()
	p.counts[reminder.ID]++
	p.mu.Unlock()
	p.once.Do(func() { close(p.started) })
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-p.releaseC:
		return nil
	}
}

func (p *blockingDeliveryProvider) waitStarted(timeout time.Duration) bool {
	select {
	case <-p.started:
		return true
	case <-time.After(timeout):
		return false
	}
}

func (p *blockingDeliveryProvider) release() {
	select {
	case <-p.releaseC:
	default:
		close(p.releaseC)
	}
}

func (p *blockingDeliveryProvider) count() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	total := 0
	for _, count := range p.counts {
		total += count
	}
	return total
}
