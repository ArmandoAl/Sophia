package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/armandoalvarado/sofia-backend/internal/app"
	"github.com/armandoalvarado/sofia-backend/internal/config"
	learningapp "github.com/armandoalvarado/sofia-backend/internal/learning/application"
	"github.com/armandoalvarado/sofia-backend/internal/learning/worker"
	privacydomain "github.com/armandoalvarado/sofia-backend/internal/privacy/domain"
	usersapp "github.com/armandoalvarado/sofia-backend/internal/users/application"
	"github.com/google/uuid"
)

func main() {
	logger := log.New(os.Stdout, "synthesis-worker ", log.LstdFlags|log.LUTC)
	cfg, err := config.Load()
	if err != nil {
		logger.Fatalf("load config: %v", err)
	}
	if !cfg.SynthesisWorkerEnabled {
		logger.Printf("disabled; set SYNTHESIS_WORKER_ENABLED=true to run")
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	repos, err := app.BuildRepositories(cfg)
	if err != nil {
		logger.Fatalf("build repository: %v", err)
	}
	defer repos.Close()

	model, err := app.BuildModelClient(cfg)
	if err != nil {
		logger.Fatalf("build model client: %v", err)
	}

	learning := learningapp.NewService(repos.Beliefs, repos.PromptVersions, repos.DailySummaries)
	learning.SetContextRepository(repos.UserContexts)
	learning.SetEntityCandidateRepository(repos.EntityCandidates, cfg.EntityPromotionThreshold)
	learning.SetEpisodeStore(repos.Episodes, model, cfg.EpisodeMinSalience, cfg.EpisodeMaxPerDay)
	runner := worker.New(worker.Deps{
		Proposals: repos.ActionProposals,
		Learning:  learning,
		Users:     usersapp.NewService(repos.Users, repos.Profiles, repos.AISettings),
		UserIDs:   repos.Users,
		Model:     model,
		Audit:     auditLogRecorder{logs: repos.AuditLogs},
	}, worker.Options{
		WorkerID:      cfg.SynthesisWorkerID,
		LeaseDuration: cfg.SynthesisWorkerLease,
		RunHourLocal:  cfg.SynthesisRunHourLocal,
		Logger:        logger,
	})

	logger.Printf("started env=%s persistence=%s worker_id=%s interval=%s lease_duration=%s run_hour_local=%d provider=%s", cfg.Env, cfg.PersistenceDriver, cfg.SynthesisWorkerID, cfg.SynthesisWorkerInterval, cfg.SynthesisWorkerLease, cfg.SynthesisRunHourLocal, cfg.AIModelProvider)
	runOnce(ctx, logger, runner)

	ticker := time.NewTicker(cfg.SynthesisWorkerInterval)
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

type auditLogRecorder struct {
	logs privacydomain.AuditLogRepository
}

func (r auditLogRecorder) RecordAuditLog(ctx context.Context, userID, action, resourceType, resourceID string, metadata map[string]any) error {
	if r.logs == nil {
		return nil
	}
	return r.logs.Create(ctx, &privacydomain.AuditLog{
		ID:           uuid.New().String(),
		UserID:       userID,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		CreatedAt:    time.Now().UTC(),
		Metadata:     metadata,
	})
}

func runOnce(ctx context.Context, logger *log.Logger, runner *worker.Worker) {
	result, err := runner.RunOnce(ctx)
	if err != nil {
		logger.Printf("run failed: %v", err)
		return
	}
	logger.Printf("run complete scanned=%d processed=%d skipped=%d failed=%d", result.Scanned, result.Processed, result.Skipped, result.Failed)
}
