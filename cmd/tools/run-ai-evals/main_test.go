package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunSkipsGeminiWithoutAPIKey(t *testing.T) {
	t.Setenv("ENV", "test")
	t.Setenv("GEMINI_API_KEY", "")
	outputDir := t.TempDir()
	datasetPath := filepath.Join("..", "..", "..", "docs", "evals", "ai_runtime_cases.json")

	if err := run(context.Background(), "auto", outputDir, datasetPath); err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(outputDir, "ai_eval_report.md")); err != nil {
		t.Fatalf("expected fake eval report: %v", err)
	}
	geminiReport, err := os.ReadFile(filepath.Join(outputDir, "gemini_eval_report.md"))
	if err != nil {
		t.Fatalf("expected Gemini skip report: %v", err)
	}
	if string(geminiReport) == "" {
		t.Fatal("expected non-empty Gemini skip report")
	}
	report := string(geminiReport)
	if !strings.Contains(report, "Provider: `gemini`") || !strings.Contains(report, "Status: `skipped`") || !strings.Contains(report, "GEMINI_API_KEY not set") {
		t.Fatalf("unexpected Gemini report content: %s", geminiReport)
	}

	geminiJSON, err := os.ReadFile(filepath.Join(outputDir, "gemini_eval_report.json"))
	if err != nil {
		t.Fatalf("expected Gemini JSON skip report: %v", err)
	}
	var parsed struct {
		Provider struct {
			Name      string `json:"name"`
			Status    string `json:"status"`
			LatencyMS int64  `json:"latency_ms"`
		} `json:"provider"`
		Status  string `json:"status"`
		Reason  string `json:"reason"`
		Privacy struct {
			FullPromptsStored     bool `json:"full_prompts_stored"`
			SensitiveValuesStored bool `json:"sensitive_values_stored"`
		} `json:"privacy"`
	}
	if err := json.Unmarshal(geminiJSON, &parsed); err != nil {
		t.Fatalf("invalid Gemini JSON report: %v", err)
	}
	if parsed.Provider.Name != "gemini" || parsed.Provider.Status != "skipped" || parsed.Status != "skipped" {
		t.Fatalf("unexpected Gemini JSON metadata: %+v", parsed)
	}
	if parsed.Reason != "GEMINI_API_KEY not set" {
		t.Fatalf("unexpected Gemini JSON reason: %q", parsed.Reason)
	}
	if parsed.Privacy.FullPromptsStored || parsed.Privacy.SensitiveValuesStored {
		t.Fatalf("Gemini JSON privacy metadata should not allow prompt/sensitive storage: %+v", parsed.Privacy)
	}
}
