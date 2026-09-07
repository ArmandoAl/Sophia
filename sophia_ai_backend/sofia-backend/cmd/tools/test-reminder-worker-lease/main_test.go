package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunOperationalValidationPassesAndWritesReport(t *testing.T) {
	outputDir := t.TempDir()
	report, err := runOperationalValidation(context.Background(), outputDir)
	if err != nil {
		t.Fatalf("runOperationalValidation returned error: %v", err)
	}
	if len(report.Scenarios) != 4 {
		t.Fatalf("expected 4 scenarios, got %d", len(report.Scenarios))
	}
	for _, scenario := range report.Scenarios {
		if scenario.Status != "pass" {
			t.Fatalf("expected pass scenario, got %+v", scenario)
		}
	}
	content, err := os.ReadFile(filepath.Join(outputDir, "reminder_worker_operational_report.md"))
	if err != nil {
		t.Fatalf("expected report file: %v", err)
	}
	reportText := string(content)
	for _, expected := range []string{
		"active lease blocks second worker",
		"expired lease allows retry",
		"recurring reminder moves next_run_at",
		"failed reminder does not duplicate",
	} {
		if !strings.Contains(reportText, expected) {
			t.Fatalf("report missing %q: %s", expected, reportText)
		}
	}
}
