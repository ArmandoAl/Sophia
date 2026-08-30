package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunWithoutAPIKeyWritesSkippedDiagnostic(t *testing.T) {
	t.Setenv("ENV", "test")
	t.Setenv("AI_MODEL_PROVIDER", "gemini")
	t.Setenv("GEMINI_API_KEY", "")
	t.Setenv("GEMINI_MODEL", "gemini-2.5-flash")
	outputDir := t.TempDir()

	if err := run(context.Background(), outputDir, ""); err != nil {
		t.Fatalf("run returned error: %v", err)
	}
	raw, err := os.ReadFile(filepath.Join(outputDir, "gemini_provider_diagnostic.md"))
	if err != nil {
		t.Fatalf("expected diagnostic report: %v", err)
	}
	report := string(raw)
	if !strings.Contains(report, "status: `skipped`") || !strings.Contains(report, "model: `gemini-2.5-flash`") {
		t.Fatalf("unexpected report:\n%s", report)
	}
	if strings.Contains(report, "GEMINI_API_KEY=") {
		t.Fatalf("report leaked env shape:\n%s", report)
	}
}

func TestWriteDiagnosticDoesNotStoreAPIKey(t *testing.T) {
	path := filepath.Join(t.TempDir(), "gemini_provider_diagnostic.md")
	if err := writeDiagnostic(path, diagnosticResult{
		GeneratedAt: "2026-07-03T00:00:00Z",
		Provider:    "gemini",
		Model:       "gemini-1.5-flash",
		Status:      "failed",
		ErrorType:   "permission_denied",
		Reason:      "request failed with key [REDACTED_API_KEY]",
		ErrorBody:   `{"message":"Permission denied for [REDACTED_API_KEY]"}`,
	}); err != nil {
		t.Fatalf("writeDiagnostic returned error: %v", err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "test-key") {
		t.Fatalf("diagnostic leaked key:\n%s", raw)
	}
}
