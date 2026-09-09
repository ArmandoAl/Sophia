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
	"github.com/armandoalvarado/sofia-backend/internal/ingestion/worker"
)

func main() {
	logger := log.New(os.Stdout, "ingestion-worker ", log.LstdFlags|log.LUTC)
	cfg, err := config.Load()
	if err != nil {
		logger.Fatalf("load config: %v", err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	repos, err := app.BuildRepositories(cfg)
	if err != nil {
		logger.Fatalf("build repositories: %v", err)
	}
	defer repos.Close()
	modules, err := app.BuildModules(cfg, repos)
	if err != nil {
		logger.Fatalf("build modules: %v", err)
	}
	model, err := app.BuildModelClient(cfg)
	if err != nil {
		logger.Fatalf("build model client: %v", err)
	}
	host, _ := os.Hostname()
	runner := worker.New(repos.IngestionBatches, modules.Learning, model, worker.Options{WorkerID: "ingestion-" + host, MaxTokens: cfg.IngestionMaxTokensPerBatch})
	run(ctx, logger, runner)
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			run(ctx, logger, runner)
		}
	}
}

func run(ctx context.Context, logger *log.Logger, runner *worker.Worker) {
	batch, err := runner.RunOnce(ctx)
	if err != nil {
		logger.Printf("run failed: %v", err)
		return
	}
	if batch != nil {
		logger.Printf("batch complete id=%s status=%s windows=%d tokens=%d", batch.ID, batch.Status, batch.WindowsProcessed, batch.InputTokens+batch.OutputTokens)
	}
}
