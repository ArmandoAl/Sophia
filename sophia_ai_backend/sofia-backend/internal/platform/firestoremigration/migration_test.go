package firestoremigration

import (
	"testing"
	"time"
)

func TestMapLegacyUser(t *testing.T) {
	now := fixedTime()
	targets := MapLegacyUser(LegacyDocument{
		Collection: "users",
		ID:         "user_001",
		Data: map[string]any{
			"name":                "Armando",
			"email":               "ARMANDO@example.com",
			"timezone":            "America/Tijuana",
			"preferred_tone":      "direct",
			"ai_autonomy_enabled": true,
			"createdAt":           "2026-05-01T10:00:00Z",
		},
	}, now)

	if len(targets) != 3 {
		t.Fatalf("expected 3 targets, got %d", len(targets))
	}
	profile := findTarget(t, targets, "user_profiles")
	if profile.Data["user_id"] != "user_001" || profile.Data["timezone"] != "America/Tijuana" {
		t.Fatalf("unexpected profile target: %+v", profile)
	}
	settings := findTarget(t, targets, "ai_settings")
	if settings.Data["autonomy_level"] != "semi_autonomous" || settings.Data["communication_style"] != "direct" {
		t.Fatalf("unexpected settings target: %+v", settings)
	}
	user := findTarget(t, targets, "users")
	if user.Action != ActionReview || user.Data["email"] != "armando@example.com" {
		t.Fatalf("unexpected user target: %+v", user)
	}
}

func TestMapLegacyActivity(t *testing.T) {
	now := fixedTime()
	target := MapLegacyActivity(LegacyDocument{
		Collection: "activities",
		ID:         "activity_001",
		Data: map[string]any{
			"title":       "Junta LinkedIn",
			"description": "Reunión",
			"type":        "ai_proposed",
			"priority":    float64(7),
			"start_time":  "2026-05-03T07:00:00Z",
			"end_time":    "2026-05-03T08:00:00Z",
			"confirmed":   true,
			"user_id":     map[string]any{"ID": "user_001"},
		},
	}, now, TargetModeStaging)

	if target.Collection != "migration_activities_preview" || target.Action != ActionCreate {
		t.Fatalf("unexpected target metadata: %+v", target)
	}
	if target.Data["type"] != "task" || target.Data["priority"] != "high" || target.Data["status"] != "active" {
		t.Fatalf("unexpected activity mapping: %+v", target.Data)
	}
	if target.Data["user_id"] != "user_001" {
		t.Fatalf("expected user_001, got %v", target.Data["user_id"])
	}
	if len(target.Warnings) == 0 {
		t.Fatalf("expected activity warnings")
	}
}

func TestMapLegacyReminder(t *testing.T) {
	now := fixedTime()
	target := MapLegacyReminder(LegacyDocument{
		Collection: "reminders",
		ID:         "reminder_001",
		Data: map[string]any{
			"text":    "Tomar agua",
			"time":    "2026-05-03T09:00:00Z",
			"source":  "user",
			"user_id": map[string]any{"Path": "projects/x/databases/default/documents/users/user_001"},
		},
	}, now, TargetModeStaging)

	if target.Collection != "migration_reminders_preview" || target.Action != ActionCreate {
		t.Fatalf("unexpected target metadata: %+v", target)
	}
	if target.Data["title"] != "Tomar agua" || target.Data["source"] != "user" || target.Data["user_id"] != "user_001" {
		t.Fatalf("unexpected reminder mapping: %+v", target.Data)
	}
	if len(target.Warnings) == 0 {
		t.Fatalf("expected reminder warnings")
	}
}

func TestBuildPreviewDryRunDoesNotApply(t *testing.T) {
	preview := BuildPreview([]LegacyDocument{
		{Collection: "reminders", ID: "r1", Data: map[string]any{"text": "Reminder"}},
	}, 20, true, true, fixedTime())

	if !preview.DryRun || !preview.Apply {
		t.Fatalf("unexpected preview flags: %+v", preview)
	}
	if len(preview.Targets) != 1 {
		t.Fatalf("expected 1 target, got %d", len(preview.Targets))
	}
	if len(preview.Warnings) == 0 {
		t.Fatalf("expected warning that apply is ignored in dry-run")
	}
}

func TestMissingFieldsUseSafeDefaults(t *testing.T) {
	now := fixedTime()
	activity := MapLegacyActivity(LegacyDocument{Collection: "activities", ID: "a1", Data: map[string]any{}}, now, TargetModeStaging)
	reminder := MapLegacyReminder(LegacyDocument{Collection: "reminders", ID: "r1", Data: map[string]any{}}, now, TargetModeStaging)

	if activity.Data["title"] != "Untitled activity" || activity.Data["timezone"] != "America/Tijuana" {
		t.Fatalf("activity defaults not applied: %+v", activity.Data)
	}
	if reminder.Data["title"] != "Untitled reminder" || reminder.Data["max_retries"] != 3 {
		t.Fatalf("reminder defaults not applied: %+v", reminder.Data)
	}
}

func TestBuildPreviewDefaultsActivitiesAndRemindersToStaging(t *testing.T) {
	preview := BuildPreview([]LegacyDocument{
		{Collection: "activities", ID: "a1", Data: map[string]any{"title": "Activity"}},
		{Collection: "reminders", ID: "r1", Data: map[string]any{"text": "Reminder"}},
	}, 20, true, false, fixedTime())

	if preview.TargetMode != TargetModeStaging {
		t.Fatalf("expected staging mode, got %q", preview.TargetMode)
	}
	if findTarget(t, preview.Targets, "migration_activities_preview").Action != ActionCreate {
		t.Fatalf("expected activity staging create target: %+v", preview.Targets)
	}
	if findTarget(t, preview.Targets, "migration_reminders_preview").Action != ActionCreate {
		t.Fatalf("expected reminder staging create target: %+v", preview.Targets)
	}
}

func TestBuildPreviewInPlaceMustBeExplicit(t *testing.T) {
	preview := BuildPreviewWithTargetMode([]LegacyDocument{
		{Collection: "activities", ID: "a1", Data: map[string]any{"title": "Activity"}},
		{Collection: "reminders", ID: "r1", Data: map[string]any{"text": "Reminder"}},
	}, 20, true, false, TargetModeInPlace, fixedTime())

	if preview.TargetMode != TargetModeInPlace {
		t.Fatalf("expected in-place mode, got %q", preview.TargetMode)
	}
	if findTarget(t, preview.Targets, "activities").Action != ActionMerge {
		t.Fatalf("expected in-place activity merge target: %+v", preview.Targets)
	}
	if findTarget(t, preview.Targets, "reminders").Action != ActionMerge {
		t.Fatalf("expected in-place reminder merge target: %+v", preview.Targets)
	}
}

func TestBuildPreviewModularOnlyExcludesUsers(t *testing.T) {
	preview := BuildPreviewWithTargetMode([]LegacyDocument{
		{Collection: "users", ID: "user_001", Data: map[string]any{"name": "Armando", "email": "a@example.com"}},
	}, 20, true, false, TargetModeModular, fixedTime())

	if preview.TargetMode != TargetModeModular {
		t.Fatalf("expected modular-only mode, got %q", preview.TargetMode)
	}
	if hasTarget(preview.Targets, "users") {
		t.Fatalf("modular-only must exclude users target: %+v", preview.Targets)
	}
	if !hasTarget(preview.Targets, "user_profiles") || !hasTarget(preview.Targets, "ai_settings") {
		t.Fatalf("modular-only should include profile/settings: %+v", preview.Targets)
	}
}

func TestBuildPreviewModularOnlyExcludesActivitiesAndReminders(t *testing.T) {
	preview := BuildPreviewWithTargetMode([]LegacyDocument{
		{Collection: "activities", ID: "a1", Data: map[string]any{"title": "Activity"}},
		{Collection: "reminders", ID: "r1", Data: map[string]any{"text": "Reminder"}},
	}, 20, true, false, TargetModeModular, fixedTime())

	if hasTarget(preview.Targets, "activities") || hasTarget(preview.Targets, "reminders") || hasTarget(preview.Targets, "migration_activities_preview") || hasTarget(preview.Targets, "migration_reminders_preview") {
		t.Fatalf("modular-only must exclude legacy activity/reminder targets: %+v", preview.Targets)
	}
	if len(preview.Warnings) == 0 {
		t.Fatal("expected modular-only exclusion warnings")
	}
}

func TestValidateApplyBlocksManualReview(t *testing.T) {
	preview := Preview{
		Targets: []TargetDocument{
			{Collection: "users", ID: "user_001", Action: ActionReview},
		},
	}

	if err := ValidateApplyAllowed(preview, false); err == nil {
		t.Fatal("expected manual_review to block apply")
	}
	if err := ValidateApplyAllowed(preview, true); err != nil {
		t.Fatalf("expected allowManualReview to pass, got %v", err)
	}
}

func TestValidateApplyBlocksInPlaceActivitiesReminders(t *testing.T) {
	preview := Preview{
		TargetMode: TargetModeInPlace,
		Targets: []TargetDocument{
			{Collection: "activities", ID: "a1", Action: ActionMerge},
		},
	}

	if err := ValidateApplyAllowed(preview, true); err == nil {
		t.Fatal("expected in-place activities apply to remain blocked")
	}
}

func TestValidateApplyBlocksHighConflict(t *testing.T) {
	preview := Preview{
		Conflicts: []Conflict{
			{Collection: "user_profiles", ID: "user_001", Severity: "high", Message: "exists"},
		},
	}

	if err := ValidateApplyAllowed(preview, true); err == nil {
		t.Fatal("expected high conflict to block apply")
	}
}

func findTarget(t *testing.T, targets []TargetDocument, collection string) TargetDocument {
	t.Helper()
	for _, target := range targets {
		if target.Collection == collection {
			return target
		}
	}
	t.Fatalf("target collection %s not found", collection)
	return TargetDocument{}
}

func hasTarget(targets []TargetDocument, collection string) bool {
	for _, target := range targets {
		if target.Collection == collection {
			return true
		}
	}
	return false
}

func fixedTime() time.Time {
	return time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)
}
