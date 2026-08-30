package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/armandoalvarado/sofia-backend/internal/ai/evals"
	runtimedomain "github.com/armandoalvarado/sofia-backend/internal/ai/runtime/domain"
	runtimeinfra "github.com/armandoalvarado/sofia-backend/internal/ai/runtime/infrastructure"
	geminiinfra "github.com/armandoalvarado/sofia-backend/internal/ai/runtime/infrastructure/gemini"
	"github.com/armandoalvarado/sofia-backend/internal/config"
)

func main() {
	provider := flag.String("provider", "auto", "model provider: fake, gemini, or auto")
	outputDir := flag.String("output-dir", "docs", "directory where ai_eval_report.md is written")
	datasetPath := flag.String("dataset", "docs/evals/ai_runtime_cases.json", "versioned eval dataset path")
	flag.Parse()

	if err := run(context.Background(), *provider, *outputDir, *datasetPath); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context, provider, outputDir, datasetPath string) error {
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return err
	}

	dataset, err := evals.LoadDataset(datasetPath)
	if err != nil {
		log.Printf("dataset unavailable at %s; using embedded eval cases: %v", datasetPath, err)
	}

	fakeResults := runProvider(ctx, runtimeinfra.NewFakeModelClient(), dataset)
	if err := writeReport(filepath.Join(outputDir, "ai_eval_report.md"), "fake", "", fakeResults); err != nil {
		return err
	}
	if !evals.AllPassed(fakeResults) {
		return errors.New("fake AI evals failed")
	}

	provider = strings.ToLower(strings.TrimSpace(provider))
	if provider == "" {
		provider = "auto"
	}
	if provider == "fake" {
		return nil
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	switch provider {
	case "auto":
		if cfg.GeminiAPIKey == "" {
			if err := writeSkippedGeminiReports(outputDir, cfg.GeminiModel, "GEMINI_API_KEY not set"); err != nil {
				return err
			}
			return nil
		}
	case "gemini":
		if cfg.GeminiAPIKey == "" {
			return writeSkippedGeminiReports(outputDir, cfg.GeminiModel, "GEMINI_API_KEY not set")
		}
	default:
		return fmt.Errorf("unsupported provider %q", provider)
	}

	client, err := geminiinfra.NewClient(cfg.GeminiAPIKey, cfg.GeminiModel)
	if err != nil {
		return err
	}
	geminiResults := runProvider(ctx, client, dataset)
	if err := writeReport(filepath.Join(outputDir, "gemini_eval_report.md"), "gemini", cfg.GeminiModel, geminiResults); err != nil {
		return err
	}
	if err := writeJSONReport(filepath.Join(outputDir, "gemini_eval_report.json"), evals.NewReport("gemini", cfg.GeminiModel, "completed", "", geminiResults)); err != nil {
		return err
	}
	if !evals.AllPassed(geminiResults) {
		return errors.New("gemini AI evals failed")
	}
	return nil
}

func runProvider(ctx context.Context, model runtimedomain.ModelClient, dataset *evals.Dataset) []evals.Result {
	runner := evals.NewRunner(model)
	runner.Dataset = dataset
	return runner.Run(ctx)
}

func writeReport(path, provider, model string, results []evals.Result) error {
	report := evals.RenderMarkdown(provider, results)
	if strings.TrimSpace(model) != "" {
		report += fmt.Sprintf("\nModel: `%s`\n", model)
	}
	if err := os.WriteFile(path, []byte(report), 0o644); err != nil {
		return err
	}
	log.Printf("wrote %s", path)
	return nil
}

func writeSkippedGeminiReports(outputDir, model, reason string) error {
	if err := writeSkippedGeminiReport(filepath.Join(outputDir, "gemini_eval_report.md"), model, reason); err != nil {
		return err
	}
	report := evals.NewReport("gemini", model, "skipped", reason, nil)
	return writeJSONReport(filepath.Join(outputDir, "gemini_eval_report.json"), report)
}

func writeSkippedGeminiReport(path, model, reason string) error {
	report := "# Gemini Eval Report\n\n"
	report += "Provider: `gemini`\n\n"
	if strings.TrimSpace(model) != "" {
		report += fmt.Sprintf("Model: `%s`\n\n", model)
	}
	report += "Status: `skipped`\n\n"
	report += fmt.Sprintf("Reason: `%s`\n", reason)
	if err := os.WriteFile(path, []byte(report), 0o644); err != nil {
		return err
	}
	log.Printf("wrote %s", path)
	return nil
}

func writeJSONReport(path string, report evals.Report) error {
	raw, err := evals.RenderJSON(report)
	if err != nil {
		return err
	}
	raw = append(raw, '\n')
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		return err
	}
	log.Printf("wrote %s", path)
	return nil
}
