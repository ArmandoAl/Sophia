package firestoreexporter

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
)

type Snapshot struct {
	GeneratedAt string               `json:"generated_at"`
	ProjectID   string               `json:"project_id"`
	Collections []CollectionSnapshot `json:"collections"`
	Comparison  Comparison           `json:"comparison"`
}

type CollectionSnapshot struct {
	Name                    string           `json:"name"`
	SampledDocuments        int              `json:"sampled_documents"`
	Fields                  []FieldSnapshot  `json:"fields"`
	SampleDocumentsRedacted []map[string]any `json:"sample_documents_redacted"`
	Subcollections          []string         `json:"subcollections,omitempty"`
}

type FieldSnapshot struct {
	Name             string   `json:"name"`
	Types            []string `json:"types"`
	RequiredInSample bool     `json:"required_in_sample"`
	Sensitive        bool     `json:"sensitive"`
	Optional         bool     `json:"optional"`
}

type Comparison struct {
	ExpectedCollections   []string                  `json:"expected_collections"`
	ActualCollections     []string                  `json:"actual_collections"`
	MissingCollections    []string                  `json:"missing_collections"`
	UnexpectedCollections []string                  `json:"unexpected_collections"`
	FieldDifferences      map[string]CollectionDiff `json:"field_differences,omitempty"`
}

type CollectionDiff struct {
	MissingFields    []string `json:"missing_fields,omitempty"`
	UnexpectedFields []string `json:"unexpected_fields,omitempty"`
}

type ExpectedModel struct {
	Collections map[string][]string
}

var sensitiveNameFragments = []string{
	"password_hash",
	"password",
	"token",
	"secret",
	"email",
	"content",
	"proposed_input",
	"execution_result",
	"note",
	"reflection",
}

func BuildSnapshot(projectID string, collections []CollectionSnapshot, expected ExpectedModel, generatedAt time.Time) Snapshot {
	sort.Slice(collections, func(i, j int) bool {
		return collections[i].Name < collections[j].Name
	})
	for i := range collections {
		sortFields(collections[i].Fields)
		sort.Strings(collections[i].Subcollections)
	}
	return Snapshot{
		GeneratedAt: generatedAt.UTC().Format(time.RFC3339),
		ProjectID:   projectID,
		Collections: collections,
		Comparison:  CompareExpectedActual(expected, collections),
	}
}

func InferCollection(name string, docs []map[string]any, redact bool) CollectionSnapshot {
	fieldCounts := map[string]int{}
	fieldTypes := map[string]map[string]struct{}{}
	samples := make([]map[string]any, 0, len(docs))

	for _, doc := range docs {
		if redact {
			samples = append(samples, RedactDocument(doc))
		} else {
			samples = append(samples, cloneMap(doc))
		}
		for key, value := range doc {
			fieldCounts[key]++
			if _, ok := fieldTypes[key]; !ok {
				fieldTypes[key] = map[string]struct{}{}
			}
			fieldTypes[key][InferType(value)] = struct{}{}
		}
	}

	fields := make([]FieldSnapshot, 0, len(fieldCounts))
	for name, count := range fieldCounts {
		types := make([]string, 0, len(fieldTypes[name]))
		for typ := range fieldTypes[name] {
			types = append(types, typ)
		}
		sort.Strings(types)
		required := len(docs) > 0 && count == len(docs)
		fields = append(fields, FieldSnapshot{
			Name:             name,
			Types:            types,
			RequiredInSample: required,
			Sensitive:        IsSensitiveField(name),
			Optional:         !required,
		})
	}

	sortFields(fields)
	return CollectionSnapshot{
		Name:                    name,
		SampledDocuments:        len(docs),
		Fields:                  fields,
		SampleDocumentsRedacted: samples,
	}
}

func InferType(value any) string {
	switch typed := value.(type) {
	case nil:
		return "null"
	case string:
		return "string"
	case bool:
		return "bool"
	case int, int8, int16, int32, int64:
		return "number"
	case uint, uint8, uint16, uint32, uint64:
		return "number"
	case float32, float64:
		return "number"
	case time.Time:
		return "timestamp"
	case []any:
		return "array"
	case []string:
		return "array"
	case []int:
		return "array"
	case []float64:
		return "array"
	case map[string]any:
		return "map"
	case json.RawMessage:
		return "json"
	default:
		if strings.HasPrefix(fmt.Sprintf("%T", typed), "[]") {
			return "array"
		}
		if strings.HasPrefix(fmt.Sprintf("%T", typed), "map[") {
			return "map"
		}
		return fmt.Sprintf("%T", typed)
	}
}

func IsSensitiveField(name string) bool {
	normalized := strings.ToLower(strings.ReplaceAll(name, "-", "_"))
	for _, fragment := range sensitiveNameFragments {
		if strings.Contains(normalized, fragment) {
			return true
		}
	}
	return false
}

func RedactDocument(doc map[string]any) map[string]any {
	redacted := make(map[string]any, len(doc))
	for key, value := range doc {
		if IsSensitiveField(key) {
			redacted[key] = redactValue(value)
			continue
		}
		redacted[key] = redactNested(value)
	}
	return redacted
}

func ParseExpectedModel(markdown string) ExpectedModel {
	collections := map[string][]string{}
	lines := strings.Split(markdown, "\n")
	var current string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "### `") && strings.HasSuffix(trimmed, "`") {
			current = strings.TrimSuffix(strings.TrimPrefix(trimmed, "### `"), "`")
			collections[current] = nil
			continue
		}
		if current == "" || !strings.Contains(trimmed, "| Campos principales |") {
			continue
		}
		parts := strings.Split(trimmed, "|")
		if len(parts) < 4 {
			continue
		}
		collections[current] = extractBackticked(parts[2])
	}
	return ExpectedModel{Collections: collections}
}

func CompareExpectedActual(expected ExpectedModel, actual []CollectionSnapshot) Comparison {
	actualFields := map[string]map[string]struct{}{}
	actualNames := make([]string, 0, len(actual))
	for _, collection := range actual {
		actualNames = append(actualNames, collection.Name)
		fields := map[string]struct{}{}
		for _, field := range collection.Fields {
			fields[field.Name] = struct{}{}
		}
		actualFields[collection.Name] = fields
	}
	sort.Strings(actualNames)

	expectedNames := make([]string, 0, len(expected.Collections))
	for name := range expected.Collections {
		expectedNames = append(expectedNames, name)
	}
	sort.Strings(expectedNames)

	comparison := Comparison{
		ExpectedCollections: expectedNames,
		ActualCollections:   actualNames,
		FieldDifferences:    map[string]CollectionDiff{},
	}

	expectedSet := stringSet(expectedNames)
	actualSet := stringSet(actualNames)
	comparison.MissingCollections = missingFrom(expectedNames, actualSet)
	comparison.UnexpectedCollections = missingFrom(actualNames, expectedSet)

	for collection, expectedFields := range expected.Collections {
		fields, ok := actualFields[collection]
		if !ok || len(expectedFields) == 0 {
			continue
		}
		expectedFieldSet := stringSet(expectedFields)
		actualFieldNames := keys(fields)
		diff := CollectionDiff{
			MissingFields:    missingFrom(expectedFields, fields),
			UnexpectedFields: missingFrom(actualFieldNames, expectedFieldSet),
		}
		if len(diff.MissingFields) > 0 || len(diff.UnexpectedFields) > 0 {
			comparison.FieldDifferences[collection] = diff
		}
	}
	if len(comparison.FieldDifferences) == 0 {
		comparison.FieldDifferences = nil
	}
	return comparison
}

func RenderSchemaMarkdown(snapshot Snapshot) string {
	var b strings.Builder
	b.WriteString("# Database Schema Snapshot\n\n")
	b.WriteString("Generated at: `" + snapshot.GeneratedAt + "`\n\n")
	b.WriteString("Project ID: `" + snapshot.ProjectID + "`\n\n")
	for _, collection := range snapshot.Collections {
		b.WriteString("## `" + collection.Name + "`\n\n")
		b.WriteString("- Purpose inferred: " + InferPurpose(collection.Name) + "\n")
		b.WriteString(fmt.Sprintf("- Sampled documents: `%d`\n", collection.SampledDocuments))
		if len(collection.Subcollections) > 0 {
			b.WriteString("- Subcollections detected: `" + strings.Join(collection.Subcollections, "`, `") + "`\n")
		}
		b.WriteString("\n| Field | Types | Required in sample | Optional | Sensitive |\n")
		b.WriteString("| --- | --- | --- | --- | --- |\n")
		for _, field := range collection.Fields {
			b.WriteString(fmt.Sprintf("| `%s` | `%s` | %t | %t | %t |\n", field.Name, strings.Join(field.Types, ", "), field.RequiredInSample, field.Optional, field.Sensitive))
		}
		b.WriteString("\n")
		b.WriteString("Recommended indexes: " + RecommendIndexes(collection.Name) + "\n\n")
		if diff, ok := snapshot.Comparison.FieldDifferences[collection.Name]; ok {
			b.WriteString("Differences vs expected model:\n\n")
			if len(diff.MissingFields) > 0 {
				b.WriteString("- Missing fields: `" + strings.Join(diff.MissingFields, "`, `") + "`\n")
			}
			if len(diff.UnexpectedFields) > 0 {
				b.WriteString("- Unexpected fields: `" + strings.Join(diff.UnexpectedFields, "`, `") + "`\n")
			}
			b.WriteString("\n")
		}
	}

	if len(snapshot.Comparison.MissingCollections) > 0 || len(snapshot.Comparison.UnexpectedCollections) > 0 {
		b.WriteString("## Collection Differences\n\n")
		if len(snapshot.Comparison.MissingCollections) > 0 {
			b.WriteString("- Missing expected collections: `" + strings.Join(snapshot.Comparison.MissingCollections, "`, `") + "`\n")
		}
		if len(snapshot.Comparison.UnexpectedCollections) > 0 {
			b.WriteString("- Unexpected collections: `" + strings.Join(snapshot.Comparison.UnexpectedCollections, "`, `") + "`\n")
		}
	}
	return b.String()
}

func RenderCollectionsSummary(snapshot Snapshot) string {
	stateByCollection := map[string]string{}
	for _, name := range snapshot.Comparison.MissingCollections {
		stateByCollection[name] = "missing"
	}
	for _, name := range snapshot.Comparison.UnexpectedCollections {
		stateByCollection[name] = "unexpected"
	}
	for _, collection := range snapshot.Collections {
		if _, ok := stateByCollection[collection.Name]; ok {
			continue
		}
		if diff, ok := snapshot.Comparison.FieldDifferences[collection.Name]; ok && (len(diff.MissingFields) > 0 || len(diff.UnexpectedFields) > 0) {
			stateByCollection[collection.Name] = "partial"
		} else {
			stateByCollection[collection.Name] = "expected"
		}
	}

	var b strings.Builder
	b.WriteString("# Database Collections Summary\n\n")
	b.WriteString("| Collection | Sampled documents | Detected fields | Status |\n")
	b.WriteString("| --- | ---: | ---: | --- |\n")
	for _, collection := range snapshot.Collections {
		b.WriteString(fmt.Sprintf("| `%s` | %d | %d | %s |\n", collection.Name, collection.SampledDocuments, len(collection.Fields), stateByCollection[collection.Name]))
	}
	for _, name := range snapshot.Comparison.MissingCollections {
		b.WriteString(fmt.Sprintf("| `%s` | 0 | 0 | missing |\n", name))
	}
	return b.String()
}

func InferPurpose(collection string) string {
	purposes := map[string]string{
		"users":               "auth credentials and identity",
		"user_profiles":       "user profile and onboarding state",
		"ai_settings":         "per-user Sofia behavior settings",
		"activities":          "planning units: tasks, habits, routines, goals, events and checklists",
		"reminders":           "scheduled reminder records",
		"mood_entries":        "mood, energy, stress and focus tracking",
		"activity_outcomes":   "activity results and subjective outcome metrics",
		"reflections":         "personal reflections",
		"memories":            "explicit assistant memory",
		"tool_definitions":    "internal AI tool registry",
		"ai_action_proposals": "controlled AI action proposals",
	}
	if purpose, ok := purposes[collection]; ok {
		return purpose
	}
	return "unknown collection; review ownership before relying on it"
}

func RecommendIndexes(collection string) string {
	indexes := map[string]string{
		"users":               "`email`",
		"user_profiles":       "`user_id`",
		"ai_settings":         "`user_id`",
		"activities":          "`user_id+status`, `user_id+type`, `user_id+priority`, range filters by date",
		"reminders":           "`user_id+status`, `user_id+activity_id`, `user_id+status+next_run_at`",
		"mood_entries":        "`user_id+occurred_at`, `user_id+activity_id`",
		"activity_outcomes":   "`user_id+activity_id`, `user_id+created_at`",
		"reflections":         "`user_id+occurred_at`, `user_id+activity_id`, tags query support",
		"memories":            "`user_id+status`, `user_id+type`, `user_id+importance`, tags query support",
		"tool_definitions":    "`name`, `enabled`, `category`",
		"ai_action_proposals": "`user_id+status`, `user_id+created_at`, `tool_name`",
	}
	if index, ok := indexes[collection]; ok {
		return index
	}
	return "review query patterns before creating indexes"
}

func redactValue(value any) any {
	if value == nil {
		return nil
	}
	text := fmt.Sprintf("%v", value)
	if len(text) <= 8 {
		return "[REDACTED]"
	}
	return text[:4] + "...[REDACTED]"
}

func redactNested(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return RedactDocument(typed)
	case []any:
		items := make([]any, 0, len(typed))
		for _, item := range typed {
			items = append(items, redactNested(item))
		}
		return items
	default:
		return value
	}
}

func cloneMap(doc map[string]any) map[string]any {
	clone := make(map[string]any, len(doc))
	for key, value := range doc {
		clone[key] = value
	}
	return clone
}

func extractBackticked(value string) []string {
	var fields []string
	parts := strings.Split(value, "`")
	for i := 1; i < len(parts); i += 2 {
		field := strings.TrimSpace(parts[i])
		if field != "" && field != "timestamps" {
			fields = append(fields, field)
		}
	}
	sort.Strings(fields)
	return fields
}

func sortFields(fields []FieldSnapshot) {
	sort.Slice(fields, func(i, j int) bool {
		return fields[i].Name < fields[j].Name
	})
}

func stringSet(items []string) map[string]struct{} {
	set := make(map[string]struct{}, len(items))
	for _, item := range items {
		set[item] = struct{}{}
	}
	return set
}

func missingFrom(items []string, existing map[string]struct{}) []string {
	var missing []string
	for _, item := range items {
		if _, ok := existing[item]; !ok {
			missing = append(missing, item)
		}
	}
	sort.Strings(missing)
	return missing
}

func keys(set map[string]struct{}) []string {
	items := make([]string, 0, len(set))
	for key := range set {
		items = append(items, key)
	}
	sort.Strings(items)
	return items
}
