package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/armandoalvarado/sofia-backend/internal/config"
	"github.com/armandoalvarado/sofia-backend/internal/database"
	notificationsinfra "github.com/armandoalvarado/sofia-backend/internal/notifications/infrastructure"
	remindersinfra "github.com/armandoalvarado/sofia-backend/internal/reminders/infrastructure"
	"github.com/armandoalvarado/sofia-backend/internal/reminders/worker"
)

func main() {
	logger := log.New(os.Stdout, "reminder-worker ", log.LstdFlags|log.LUTC)
	cfg, err := config.Load()
	if err != nil {
		logger.Fatalf("load config: %v", err)
	}
	if !cfg.ReminderWorkerEnabled {
		logger.Printf("disabled; set REMINDER_WORKER_ENABLED=true to run")
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	repo, devices, attempts, closeFn, err := buildRepositories(ctx, cfg)
	if err != nil {
		logger.Fatalf("build repository: %v", err)
	}
	defer closeFn()

	delivery, err := buildDeliveryProvider(cfg, devices)
	if err != nil {
		logger.Fatalf("build delivery provider: %v", err)
	}

	runner := worker.New(repo, delivery, worker.Options{
		BatchSize:     cfg.ReminderWorkerBatchSize,
		WorkerID:      cfg.ReminderWorkerID,
		LeaseDuration: cfg.ReminderWorkerLease,
		ProviderName:  cfg.ReminderDeliveryProvider,
		DryRun:        cfg.FCMDryRun,
		Attempts:      attempts,
		Logger:        logger,
	})

	logger.Printf("started env=%s persistence=%s worker_id=%s interval=%s batch_size=%d lease_duration=%s delivery_provider=%s fcm_enabled=%t fcm_dry_run=%t", cfg.Env, cfg.PersistenceDriver, cfg.ReminderWorkerID, cfg.ReminderWorkerInterval, cfg.ReminderWorkerBatchSize, cfg.ReminderWorkerLease, cfg.ReminderDeliveryProvider, cfg.FCMEnabled, cfg.FCMDryRun)
	runOnce(ctx, logger, runner)

	ticker := time.NewTicker(cfg.ReminderWorkerInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Printf("stopped")
			return
		case <-ticker.C:
			runOnce(ctx, logger, runner)
		}
	}
}

func buildRepositories(ctx context.Context, cfg config.Config) (worker.ReminderRepository, worker.DeviceTokenLister, worker.DeliveryAttemptRecorder, func(), error) {
	switch cfg.PersistenceDriver {
	case "memory":
		return remindersinfra.NewInMemoryReminderRepository(), notificationsinfra.NewInMemoryDeviceTokenRepository(), notificationsinfra.NewInMemoryDeliveryAttemptRepository(), func() {}, nil
	case "firestore":
		store, err := database.NewFirestore(ctx, cfg)
		if err != nil {
			return nil, nil, nil, nil, err
		}
		return remindersinfra.NewFirestoreReminderRepository(store.Client), notificationsinfra.NewFirestoreDeviceTokenRepository(store.Client), notificationsinfra.NewFirestoreDeliveryAttemptRepository(store.Client), func() {
			if err := store.Close(); err != nil {
				log.Printf("close firestore: %v", err)
			}
		}, nil
	default:
		return nil, nil, nil, nil, errors.New("unsupported persistence driver")
	}
}

func buildDeliveryProvider(cfg config.Config, devices worker.DeviceTokenLister) (worker.ReminderDeliveryProvider, error) {
	switch cfg.ReminderDeliveryProvider {
	case "noop", "":
		return worker.NoopDeliveryProvider{}, nil
	case "fcm":
		return worker.NewFCMDeliveryProvider(devices, worker.FCMDeliveryOptions{
			ProjectID: cfg.FCMProjectID,
			Enabled:   cfg.FCMEnabled,
			DryRun:    cfg.FCMDryRun,
		}), nil
	default:
		return nil, errors.New("unsupported reminder delivery provider")
	}
}

func runOnce(ctx context.Context, logger *log.Logger, runner *worker.Worker) {
	result, err := runner.RunOnce(ctx)
	if err != nil {
		logger.Printf("run failed: %v", err)
		return
	}
	logger.Printf("run complete scanned=%d delivered=%d failed=%d skipped=%d", result.Scanned, result.Delivered, result.Failed, result.Skipped)
}
