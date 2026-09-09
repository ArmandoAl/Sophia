package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunSkipsDeepSeekWithoutAPIKey(t *testing.T) {
	t.Setenv("ENV", "test")
	t.Setenv("DEEPSEEK_API_KEY", "")
	outputDir := t.TempDir()
	datasetPath := filepath.Join("..", "..", "..", "docs", "evals", "ai_runtime_cases.json")

	if err := run(context.Background(), "auto", outputDir, datasetPath); err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(outputDir, "ai_eval_report.md")); err != nil {
		t.Fatalf("expected fake eval report: %v", err)
	}
	deepSeekReport, err := os.ReadFile(filepath.Join(outputDir, "deepseek_eval_report.md"))
	if err != nil {
		t.Fatalf("expected DeepSeek skip report: %v", err)
	}
	if report := string(deepSeekReport); !strings.Contains(report, "Provider: `deepseek`") || !strings.Contains(report, "Status: `skipped`") || !strings.Contains(report, "DEEPSEEK_API_KEY not set") {
		t.Fatalf("unexpected DeepSeek report content: %s", deepSeekReport)
	}
	if _, err := os.Stat(filepath.Join(outputDir, "deepseek_eval_report.json")); err != nil {
		t.Fatalf("expected DeepSeek JSON skip report: %v", err)
	}
}
