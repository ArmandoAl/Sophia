package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/armandoalvarado/sofia-backend/internal/config"
	"github.com/armandoalvarado/sofia-backend/internal/database"
	privacyapp "github.com/armandoalvarado/sofia-backend/internal/privacy/application"
	privacyinfra "github.com/armandoalvarado/sofia-backend/internal/privacy/infrastructure"
)

func main() {
	limit := flag.Int("limit", 20, "maximum approved delete requests to process")
	outputDir := flag.String("output-dir", "docs", "directory for the JSON execution report")
	flag.Parse()

	logger := log.New(os.Stdout, "delete-executor ", log.LstdFlags|log.LUTC)
	cfg, err := config.Load()
	if err != nil {
		logger.Fatalf("load config: %v", err)
	}
	if cfg.PersistenceDriver != "firestore" {
		logger.Fatalf("process-delete-requests requires PERSISTENCE_DRIVER=firestore")
	}

	ctx := context.Background()
	store, err := database.NewFirestore(ctx, cfg)
	if err != nil {
		logger.Fatalf("create firestore: %v", err)
	}
	defer func() {
		if err := store.Close(); err != nil {
			logger.Printf("close firestore: %v", err)
		}
	}()

	requests := privacyinfra.NewFirestoreDeleteRequestRepository(store.Client)
	auditLogs := privacyinfra.NewFirestoreAuditLogRepository(store.Client)
	deleter := privacyinfra.NewFirestoreUserDataDeleter(store.Client)
	executor := privacyapp.NewDeleteExecutor(requests, auditLogs, deleter, privacyapp.DeleteExecutorOptions{})

	report, err := executor.ProcessApproved(ctx, *limit)
	if err != nil {
		logger.Fatalf("process approved delete requests: %v", err)
	}
	if err := writeReport(*outputDir, report); err != nil {
		logger.Fatalf("write report: %v", err)
	}
	logger.Printf("completed processed=%d completed=%d failed=%d", report.Processed, report.Completed, report.Failed)
}

func writeReport(outputDir string, report *privacyapp.DeleteExecutionReport) error {
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return err
	}
	payload, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	path := filepath.Join(outputDir, "delete_requests_report.json")
	if err := os.WriteFile(path, payload, 0o644); err != nil {
		return err
	}
	fmt.Printf("wrote %s\n", path)
	return nil
}
