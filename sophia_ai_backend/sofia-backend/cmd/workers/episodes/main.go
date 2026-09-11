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
	"github.com/armandoalvarado/sofia-backend/internal/learning/episodeworker"
)

func main() {
	logger := log.New(os.Stdout, "episode-worker ", log.LstdFlags|log.LUTC)
	cfg, err := config.Load()
	if err != nil {
		logger.Fatalf("load config: %v", err)
	}
	if !cfg.EpisodeWorkerEnabled {
		logger.Printf("disabled; set EPISODE_WORKER_ENABLED=true to run")
		return
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
	runner := episodeworker.New(repos.Users, repos.Conversations, repos.Messages, modules.Learning)
	lastRun := ""
	run := func() {
		now := time.Now().UTC()
		date := now.Format(time.DateOnly)
		if now.Hour() != cfg.EpisodeRunHourUTC || date == lastRun {
			return
		}
		count, err := runner.RunOnce(ctx)
		if err != nil {
			logger.Printf("run failed: %v", err)
			return
		}
		lastRun = date
		logger.Printf("run complete conversations=%d", count)
	}
	run()
	ticker := time.NewTicker(cfg.EpisodeWorkerInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			run()
		}
	}
}
