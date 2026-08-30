package evals

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	runtimeinfra "github.com/armandoalvarado/sofia-backend/internal/ai/runtime/infrastructure"
)

func TestFakeProviderEvalsPass(t *testing.T) {
	results := NewRunner(runtimeinfra.NewFakeModelClient()).Run(context.Background())
	if !AllPassed(results) {
		t.Fatalf("expected all evals to pass:\n%s", RenderMarkdown("fake", results))
	}
	if len(results) < 11 {
		t.Fatalf("expected critical eval coverage, got %d results", len(results))
	}
}

func TestVersionedDatasetEvalsPass(t *testing.T) {
	dataset, err := LoadDataset(filepath.Join("..", "..", "..", "docs", "evals", "ai_runtime_cases.json"))
	if err != nil {
		t.Fatalf("LoadDataset returned error: %v", err)
	}
	runner := NewRunner(runtimeinfra.NewFakeModelClient())
	runner.Dataset = dataset
	results := runner.Run(context.Background())
	if !AllPassed(results) {
		t.Fatalf("expected dataset evals to pass:\n%s", RenderMarkdown("fake", results))
	}
	if len(results) != len(dataset.Cases) {
		t.Fatalf("expected %d dataset evals, got %d", len(dataset.Cases), len(results))
	}
}

func TestDatasetIncludesPromptInjectionCase(t *testing.T) {
	dataset, err := LoadDataset(filepath.Join("..", "..", "..", "docs", "evals", "ai_runtime_cases.json"))
	if err != nil {
		t.Fatalf("LoadDataset returned error: %v", err)
	}
	for _, item := range dataset.Cases {
		if item.ID == "prompt_injection_basic" {
			return
		}
	}
	t.Fatal("expected prompt_injection_basic eval case")
}

func TestEvalReportDoesNotExposeSensitiveMemoryContent(t *testing.T) {
	results := NewRunner(runtimeinfra.NewFakeModelClient()).Run(context.Background())
	report := RenderMarkdown("fake", results)
	if strings.Contains(report, "secret-token-12345") {
		t.Fatal("eval report leaked sensitive memory content")
	}
}

func TestEvalReportIncludesMetrics(t *testing.T) {
	results := NewRunner(runtimeinfra.NewFakeModelClient()).Run(context.Background())
	report := RenderMarkdown("fake", results)
	if !strings.Contains(report, "Latency ms") || !strings.Contains(report, "Schema failures") || !strings.Contains(report, "Policy failures") {
		t.Fatalf("eval report missing metrics columns:\n%s", report)
	}
}

func TestJSONReportDoesNotExposeSensitiveMemoryContent(t *testing.T) {
	results := NewRunner(runtimeinfra.NewFakeModelClient()).Run(context.Background())
	report := NewReport("fake", "", "completed", "", results)
	raw, err := RenderJSON(report)
	if err != nil {
		t.Fatalf("RenderJSON returned error: %v", err)
	}
	if strings.Contains(string(raw), "secret-token-12345") {
		t.Fatal("JSON eval report leaked sensitive memory content")
	}
	if !strings.Contains(string(raw), `"full_prompts_stored": false`) {
		t.Fatalf("JSON eval report missing privacy metadata:\n%s", raw)
	}
}
