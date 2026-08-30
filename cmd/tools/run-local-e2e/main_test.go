package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunE2ECompletesAndWritesReport(t *testing.T) {
	outputDir := t.TempDir()
	report, err := runE2E(context.Background(), outputDir)
	if err != nil {
		t.Fatalf("runE2E returned error: %v", err)
	}
	if report.UserID == "" || report.ActivityID == "" || report.ReminderID == "" || report.MemoryID == "" || report.ProposalID == "" {
		t.Fatalf("expected important IDs to be populated: %+v", report)
	}
	if !report.ProposalOnlyPreserved {
		t.Fatalf("expected proposal-only to be preserved")
	}
	if !report.NoAutomaticAIExecution {
		t.Fatalf("expected no automatic AI execution")
	}
	if !report.ExplicitExecutionCreated {
		t.Fatalf("expected explicit execution to create activity")
	}
	if report.ProposalsBeforeDryRun != report.ProposalsAfterDryRun {
		t.Fatalf("dry_run persisted proposals: before=%d after=%d", report.ProposalsBeforeDryRun, report.ProposalsAfterDryRun)
	}
	content, err := os.ReadFile(filepath.Join(outputDir, "local_e2e_report.md"))
	if err != nil {
		t.Fatalf("expected report: %v", err)
	}
	text := string(content)
	for _, expected := range []string{
		"proposal_only_preserved: `true`",
		"no_automatic_ai_execution: `true`",
		"explicit_execution_created_activity: `true`",
		"execute proposal explicitly",
	} {
		if !strings.Contains(text, expected) {
			t.Fatalf("report missing %q: %s", expected, text)
		}
	}
}
