package firestoremigration

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"
)

const (
	ActionCreate = "create"
	ActionMerge  = "merge"
	ActionReview = "manual_review"

	TargetModeStaging = "staging"
	TargetModeInPlace = "in-place"
	TargetModeModular = "modular-only"
)

type LegacyDocument struct {
	Collection string
	ID         string
	Data       map[string]any
}

type TargetDocument struct {
	Collection string         `json:"collection"`
	ID         string         `json:"id"`
	Action     string         `json:"action"`
	Data       map[string]any `json:"data,omitempty"`
	Warnings   []string       `json:"warnings,omitempty"`
}

type Conflict struct {
	Collection string `json:"collection"`
	ID         string `json:"id"`
	Severity   string `json:"severity"`
	Message    string `json:"message"`
}

type Preview struct {
	GeneratedAt string           `json:"generated_at"`
	DryRun      bool             `json:"dry_run"`
	Apply       bool             `json:"apply"`
	Limit       int              `json:"limit"`
	TargetMode  string           `json:"target_mode"`
	Targets     []TargetDocument `json:"targets"`
	Conflicts   []Conflict       `json:"conflicts,omitempty"`
	Warnings    []string         `json:"warnings,omitempty"`
}

func NewPreview(limit int, dryRun, apply bool, targetMode string, targets []TargetDocument, warnings []string, now time.Time) Preview {
	sort.Slice(targets, func(i, j int) bool {
		if targets[i].Collection == targets[j].Collection {
			return targets[i].ID < targets[j].ID
		}
		return targets[i].Collection < targets[j].Collection
	})
	sort.Strings(warnings)
	return Preview{
		GeneratedAt: now.UTC().Format(time.RFC3339),
		DryRun:      dryRun,
		Apply:       apply,
		Limit:       limit,
		TargetMode:  defaultTargetMode(targetMode),
		Targets:     targets,
		Warnings:    warnings,
	}
}

func MapLegacyUser(doc LegacyDocument, now time.Time) []TargetDocument {
	userID := fallbackID(doc)
	name := stringValue(doc.Data, "name", "Sofia User")
	email := strings.ToLower(strings.TrimSpace(stringValue(doc.Data, "email", "")))
	timezone := stringValue(doc.Data, "timezone", "America/Tijuana")
	createdAt := timeValue(doc.Data, "created_at", now)
	if raw, ok := doc.Data["createdAt"]; ok {
		createdAt = anyTime(raw, createdAt)
	}

	warnings := []string{"legacy user id will be reused as modular user_id"}
	if email == "" {
		warnings = append(warnings, "legacy user has no email; auth user requires manual review")
	}
	if timezone == "America/Tijuana" && strings.TrimSpace(stringValue(doc.Data, "timezone", "")) == "" {
		warnings = append(warnings, "timezone defaulted to America/Tijuana")
	}

	return []TargetDocument{
		{
			Collection: "users",
			ID:         userID,
			Action:     ActionReview,
			Data: map[string]any{
				"id":         userID,
				"name":       name,
				"email":      email,
				"role":       "user",
				"created_at": createdAt,
				"updated_at": now,
			},
			Warnings: append(warnings, "password_hash is not present in legacy user; do not overwrite auth users automatically"),
		},
		{
			Collection: "user_profiles",
			ID:         userID,
			Action:     ActionCreate,
			Data: map[string]any{
				"id":                   userID,
				"user_id":              userID,
				"display_name":         name,
				"preferred_name":       name,
				"timezone":             timezone,
				"locale":               "es-MX",
				"avatar_url":           "",
				"onboarding_completed": false,
				"created_at":           createdAt,
				"updated_at":           now,
			},
			Warnings: warnings,
		},
		{
			Collection: "ai_settings",
			ID:         userID,
			Action:     ActionCreate,
			Data: map[string]any{
				"user_id":             userID,
				"assistant_name":      "Sofía",
				"communication_style": defaultCommunicationStyle(doc.Data),
				"proactivity_level":   "medium",
				"autonomy_level":      legacyAutonomy(doc.Data),
				"memory_enabled":      true,
				"reminders_enabled":   true,
				"planning_enabled":    true,
				"preferred_language":  "es",
				"created_at":          createdAt,
				"updated_at":          now,
			},
			Warnings: []string{
				"ai settings are inferred from legacy user preferences",
				"memory/reminders/planning default to enabled",
				"preferred_language defaulted to es",
			},
		},
	}
}

func MapLegacyUserModularOnly(doc LegacyDocument, now time.Time) []TargetDocument {
	targets := MapLegacyUser(doc, now)
	result := make([]TargetDocument, 0, 2)
	for _, target := range targets {
		if target.Collection == "user_profiles" || target.Collection == "ai_settings" {
			result = append(result, target)
		}
	}
	return result
}

func MapLegacyActivity(doc LegacyDocument, now time.Time, targetMode string) TargetDocument {
	id := fallbackID(doc)
	collection := "migration_activities_preview"
	action := ActionCreate
	startAt := timeValue(doc.Data, "start_at", timeValue(doc.Data, "start_time", now))
	dueAt := timeValue(doc.Data, "due_at", timeValue(doc.Data, "end_time", startAt))
	priority := mapPriority(doc.Data["priority"])
	status := "pending"
	if boolValue(doc.Data, "confirmed", false) {
		status = "active"
	}
	userID := refID(doc.Data["user_id"])
	warnings := []string{
		"legacy activity id will be reused",
		"target-mode=staging writes normalized preview outside production activities",
		"timezone defaulted to America/Tijuana",
		"created_at is migration time because legacy activity has no created_at",
	}
	if targetMode == TargetModeInPlace {
		collection = "activities"
		action = ActionMerge
		warnings[1] = "target-mode=in-place uses production activities; merge can overwrite normalized fields"
	}
	if userID == "" {
		warnings = append(warnings, "legacy activity user_id could not be resolved")
	} else {
		warnings = append(warnings, "user_id extracted from legacy DocumentRef")
	}
	if startAt.Before(now) {
		warnings = append(warnings, "activity start_at is in the past")
	}
	if dueAt.Before(now) {
		warnings = append(warnings, "activity due_at is in the past")
	}

	return TargetDocument{
		Collection: collection,
		ID:         id,
		Action:     action,
		Data: map[string]any{
			"id":              id,
			"user_id":         userID,
			"type":            normalizeActivityType(stringValue(doc.Data, "type", "task")),
			"title":           stringValue(doc.Data, "title", "Untitled activity"),
			"description":     stringValue(doc.Data, "description", ""),
			"status":          status,
			"priority":        priority,
			"start_at":        startAt,
			"due_at":          dueAt,
			"timezone":        "America/Tijuana",
			"tags":            []string{},
			"source":          "imported",
			"created_by":      "user",
			"created_at":      now,
			"updated_at":      now,
			"has_reminder":    false,
			"recurrence_rule": "",
			"ai_manageable":   boolValue(doc.Data, "reschedulable", false),
			"ai_notes":        "",
		},
		Warnings: warnings,
	}
}

func MapLegacyReminder(doc LegacyDocument, now time.Time, targetMode string) TargetDocument {
	id := fallbackID(doc)
	collection := "migration_reminders_preview"
	action := ActionCreate
	scheduledAt := timeValue(doc.Data, "scheduled_at", timeValue(doc.Data, "time", now))
	userID := refID(doc.Data["user_id"])
	warnings := []string{
		"legacy reminder id will be reused",
		"target-mode=staging writes normalized preview outside production reminders",
		"timezone defaulted to America/Tijuana",
		"created_at is migration time because legacy reminder has no created_at",
	}
	if targetMode == TargetModeInPlace {
		collection = "reminders"
		action = ActionMerge
		warnings[1] = "target-mode=in-place uses production reminders; merge can overwrite normalized fields"
	}
	if userID == "" {
		warnings = append(warnings, "legacy reminder user_id could not be resolved")
	} else {
		warnings = append(warnings, "user_id extracted from legacy DocumentRef")
	}
	if scheduledAt.Before(now) {
		warnings = append(warnings, "scheduled_at is in the past")
	}

	return TargetDocument{
		Collection: collection,
		ID:         id,
		Action:     action,
		Data: map[string]any{
			"id":              id,
			"user_id":         userID,
			"activity_id":     "",
			"title":           stringValue(doc.Data, "title", stringValue(doc.Data, "text", "Untitled reminder")),
			"description":     "",
			"status":          "pending",
			"scheduled_at":    scheduledAt,
			"timezone":        "America/Tijuana",
			"recurrence_rule": "",
			"next_run_at":     scheduledAt,
			"retry_count":     0,
			"max_retries":     3,
			"source":          defaultReminderSource(doc.Data),
			"created_by":      "user",
			"created_at":      now,
			"updated_at":      now,
		},
		Warnings: warnings,
	}
}

func BuildPreview(docs []LegacyDocument, limit int, dryRun, apply bool, now time.Time) Preview {
	return BuildPreviewWithTargetMode(docs, limit, dryRun, apply, TargetModeStaging, now)
}

func BuildPreviewWithTargetMode(docs []LegacyDocument, limit int, dryRun, apply bool, targetMode string, now time.Time) Preview {
	targets := []TargetDocument{}
	warnings := []string{}
	targetMode = defaultTargetMode(targetMode)
	for _, doc := range docs {
		switch doc.Collection {
		case "users":
			if targetMode == TargetModeModular {
				targets = append(targets, MapLegacyUserModularOnly(doc, now)...)
				continue
			}
			targets = append(targets, MapLegacyUser(doc, now)...)
		case "activities":
			if targetMode == TargetModeModular {
				warnings = append(warnings, "target-mode=modular-only excludes legacy activities")
				continue
			}
			targets = append(targets, MapLegacyActivity(doc, now, targetMode))
		case "reminders":
			if targetMode == TargetModeModular {
				warnings = append(warnings, "target-mode=modular-only excludes legacy reminders")
				continue
			}
			targets = append(targets, MapLegacyReminder(doc, now, targetMode))
		default:
			warnings = append(warnings, fmt.Sprintf("collection %s is not migrated by this command", doc.Collection))
		}
	}
	if apply && dryRun {
		warnings = append(warnings, "dry-run is true; apply was ignored")
	}
	if targetMode == TargetModeStaging {
		warnings = append(warnings, "target-mode=staging; activities/reminders are written to migration preview collections")
	}
	if targetMode == TargetModeModular {
		warnings = append(warnings, "target-mode=modular-only; only user_profiles and ai_settings are generated")
	}
	return NewPreview(limit, dryRun, apply, targetMode, targets, warnings, now)
}

func RenderPreviewMarkdown(preview Preview) string {
	var b strings.Builder
	b.WriteString("# Migration Preview\n\n")
	b.WriteString("Generated at: `" + preview.GeneratedAt + "`\n\n")
	b.WriteString(fmt.Sprintf("- Dry run: `%t`\n", preview.DryRun))
	b.WriteString(fmt.Sprintf("- Apply requested: `%t`\n", preview.Apply))
	b.WriteString(fmt.Sprintf("- Limit: `%d`\n\n", preview.Limit))
	b.WriteString(fmt.Sprintf("- Target mode: `%s`\n\n", preview.TargetMode))
	b.WriteString("| Target collection | Document ID | Action | Warnings |\n")
	b.WriteString("| --- | --- | --- | --- |\n")
	for _, target := range preview.Targets {
		b.WriteString(fmt.Sprintf("| `%s` | `%s` | `%s` | %s |\n", target.Collection, target.ID, target.Action, strings.Join(target.Warnings, "; ")))
	}
	if len(preview.Conflicts) > 0 {
		b.WriteString("\n## Conflicts\n\n")
		b.WriteString("| Collection | Document ID | Severity | Message |\n")
		b.WriteString("| --- | --- | --- | --- |\n")
		for _, conflict := range preview.Conflicts {
			b.WriteString(fmt.Sprintf("| `%s` | `%s` | `%s` | %s |\n", conflict.Collection, conflict.ID, conflict.Severity, conflict.Message))
		}
	}
	if len(preview.Warnings) > 0 {
		b.WriteString("\n## Warnings\n\n")
		for _, warning := range preview.Warnings {
			b.WriteString("- " + warning + "\n")
		}
	}
	return b.String()
}

func ValidateApplyAllowed(preview Preview, allowManualReview bool) error {
	for _, target := range preview.Targets {
		if target.Action == ActionReview && !allowManualReview {
			return fmt.Errorf("manual_review target %s/%s blocks apply; pass explicit allow flag after review", target.Collection, target.ID)
		}
	}
	for _, conflict := range preview.Conflicts {
		if conflict.Severity == "critical" || conflict.Severity == "high" {
			return fmt.Errorf("conflict %s/%s blocks apply: %s", conflict.Collection, conflict.ID, conflict.Message)
		}
	}
	if preview.TargetMode == TargetModeInPlace && !allowManualReview {
		return fmt.Errorf("target-mode=in-place blocks apply without explicit manual review flag")
	}
	if preview.TargetMode == TargetModeInPlace {
		for _, target := range preview.Targets {
			if target.Collection == "activities" || target.Collection == "reminders" {
				return fmt.Errorf("target-mode=in-place legacy activity/reminder apply is blocked; use staging or modular-only")
			}
		}
	}
	return nil
}

func defaultTargetMode(value string) string {
	switch strings.TrimSpace(value) {
	case TargetModeInPlace:
		return TargetModeInPlace
	case TargetModeModular:
		return TargetModeModular
	default:
		return TargetModeStaging
	}
}

func MarshalPreview(preview Preview) ([]byte, error) {
	return json.MarshalIndent(preview, "", "  ")
}

func fallbackID(doc LegacyDocument) string {
	if doc.ID != "" {
		return doc.ID
	}
	return stringValue(doc.Data, "_document_id", "")
}

func stringValue(data map[string]any, key, fallback string) string {
	raw, ok := data[key]
	if !ok || raw == nil {
		return fallback
	}
	if value, ok := raw.(string); ok && strings.TrimSpace(value) != "" {
		return strings.TrimSpace(value)
	}
	return fallback
}

func boolValue(data map[string]any, key string, fallback bool) bool {
	raw, ok := data[key]
	if !ok {
		return fallback
	}
	if value, ok := raw.(bool); ok {
		return value
	}
	return fallback
}

func timeValue(data map[string]any, key string, fallback time.Time) time.Time {
	raw, ok := data[key]
	if !ok {
		return fallback
	}
	return anyTime(raw, fallback)
}

func anyTime(raw any, fallback time.Time) time.Time {
	switch value := raw.(type) {
	case time.Time:
		return value
	case string:
		parsed, err := time.Parse(time.RFC3339, value)
		if err == nil {
			return parsed
		}
	}
	return fallback
}

func refID(raw any) string {
	if raw == nil {
		return ""
	}
	if value, ok := raw.(interface{ GetID() string }); ok {
		return value.GetID()
	}
	if data, ok := raw.(map[string]any); ok {
		if id := stringValue(data, "ID", ""); id != "" {
			return id
		}
		if path := stringValue(data, "Path", ""); path != "" {
			parts := strings.Split(strings.Trim(path, "/"), "/")
			return parts[len(parts)-1]
		}
	}
	value := reflect.ValueOf(raw)
	if value.Kind() == reflect.Pointer && !value.IsNil() {
		value = value.Elem()
	}
	if value.Kind() == reflect.Struct {
		if field := value.FieldByName("ID"); field.IsValid() && field.Kind() == reflect.String {
			return field.String()
		}
		if field := value.FieldByName("Path"); field.IsValid() && field.Kind() == reflect.String && field.String() != "" {
			parts := strings.Split(strings.Trim(field.String(), "/"), "/")
			return parts[len(parts)-1]
		}
	}
	return ""
}

func mapPriority(raw any) string {
	var score float64
	switch value := raw.(type) {
	case int:
		score = float64(value)
	case int64:
		score = float64(value)
	case float64:
		score = value
	}
	switch {
	case score >= 8:
		return "urgent"
	case score >= 6:
		return "high"
	case score > 0 && score <= 3:
		return "low"
	default:
		return "medium"
	}
}

func normalizeActivityType(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "habit", "routine", "goal", "event", "checklist":
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return "task"
	}
}

func legacyAutonomy(data map[string]any) string {
	if boolValue(data, "ai_autonomy_enabled", false) {
		return "semi_autonomous"
	}
	return "suggestive"
}

func defaultCommunicationStyle(data map[string]any) string {
	if style := stringValue(data, "preferred_tone", ""); style != "" {
		return style
	}
	return "warm"
}

func defaultReminderSource(data map[string]any) string {
	if source := stringValue(data, "source", ""); source != "" {
		return source
	}
	return "imported"
}
