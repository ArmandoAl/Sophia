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
	deepseekinfra "github.com/armandoalvarado/sofia-backend/internal/ai/runtime/infrastructure/deepseek"
	"github.com/armandoalvarado/sofia-backend/internal/config"
)

func main() {
	provider := flag.String("provider", "auto", "model provider: fake, deepseek, or auto")
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
	case "auto", "deepseek":
		if cfg.DeepSeekAPIKey == "" {
			return writeSkippedDeepSeekReports(outputDir, cfg.DeepSeekModels[""], "DEEPSEEK_API_KEY not set")
		}
		return runDeepSeekEvals(ctx, cfg, outputDir, dataset)
	default:
		return fmt.Errorf("unsupported provider %q", provider)
	}
}

func runDeepSeekEvals(ctx context.Context, cfg config.Config, outputDir string, dataset *evals.Dataset) error {
	options := []deepseekinfra.Option{}
	if cfg.DeepSeekBaseURL != "" {
		options = append(options, deepseekinfra.WithEndpoint(cfg.DeepSeekBaseURL))
	}
	client, err := deepseekinfra.NewClient(cfg.DeepSeekAPIKey, cfg.DeepSeekModels, options...)
	if err != nil {
		return err
	}
	results := runProvider(ctx, client, dataset)
	model := cfg.DeepSeekModels[""]
	if err := writeReport(filepath.Join(outputDir, "deepseek_eval_report.md"), "deepseek", model, results); err != nil {
		return err
	}
	if err := writeJSONReport(filepath.Join(outputDir, "deepseek_eval_report.json"), evals.NewReport("deepseek", model, "completed", "", results)); err != nil {
		return err
	}
	if !evals.AllPassed(results) {
		return errors.New("deepseek AI evals failed")
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

func writeSkippedDeepSeekReports(outputDir, model, reason string) error {
	if err := writeSkippedDeepSeekReport(filepath.Join(outputDir, "deepseek_eval_report.md"), model, reason); err != nil {
		return err
	}
	report := evals.NewReport("deepseek", model, "skipped", reason, nil)
	return writeJSONReport(filepath.Join(outputDir, "deepseek_eval_report.json"), report)
}

func writeSkippedDeepSeekReport(path, model, reason string) error {
	report := "# DeepSeek Eval Report\n\n"
	report += "Provider: `deepseek`\n\n"
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
