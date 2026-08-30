package firestoreexporter

import (
	"strings"
	"testing"
	"time"
)

func TestInferType(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  string
	}{
		{name: "string", value: "sofia", want: "string"},
		{name: "bool", value: true, want: "bool"},
		{name: "number", value: 3, want: "number"},
		{name: "timestamp", value: time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC), want: "timestamp"},
		{name: "array", value: []any{"a"}, want: "array"},
		{name: "map", value: map[string]any{"a": "b"}, want: "map"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InferType(tt.value); got != tt.want {
				t.Fatalf("InferType() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRedactDocument(t *testing.T) {
	doc := map[string]any{
		"email":          "user@example.com",
		"password_hash":  "bcrypt-hash-value",
		"title":          "Buy milk",
		"proposed_input": map[string]any{"content": "private thought"},
	}

	got := RedactDocument(doc)

	if got["email"] == "user@example.com" {
		t.Fatalf("email was not redacted")
	}
	if got["password_hash"] == "bcrypt-hash-value" {
		t.Fatalf("password_hash was not redacted")
	}
	if got["title"] != "Buy milk" {
		t.Fatalf("non-sensitive title changed: %v", got["title"])
	}
	if got["proposed_input"] == doc["proposed_input"] {
		t.Fatalf("proposed_input was not redacted")
	}
}

func TestInferCollection(t *testing.T) {
	collection := InferCollection("activities", []map[string]any{
		{"id": "1", "title": "A", "priority": "medium", "tags": []any{"work"}},
		{"id": "2", "title": "B", "priority": "high"},
	}, true)

	if collection.SampledDocuments != 2 {
		t.Fatalf("SampledDocuments = %d, want 2", collection.SampledDocuments)
	}

	fields := map[string]FieldSnapshot{}
	for _, field := range collection.Fields {
		fields[field.Name] = field
	}
	if !fields["id"].RequiredInSample {
		t.Fatalf("id should be required in sample")
	}
	if !fields["tags"].Optional {
		t.Fatalf("tags should be optional")
	}
	if fields["tags"].Types[0] != "array" {
		t.Fatalf("tags type = %v, want array", fields["tags"].Types)
	}
}

func TestCompareExpectedActual(t *testing.T) {
	expected := ExpectedModel{Collections: map[string][]string{
		"users":      {"id", "email", "role"},
		"activities": {"id", "user_id", "title"},
		"memories":   {"id"},
	}}
	actual := []CollectionSnapshot{
		{Name: "users", Fields: []FieldSnapshot{{Name: "id"}, {Name: "email"}, {Name: "unexpected"}}},
		{Name: "activities", Fields: []FieldSnapshot{{Name: "id"}, {Name: "user_id"}, {Name: "title"}}},
		{Name: "extra", Fields: []FieldSnapshot{{Name: "id"}}},
	}

	comparison := CompareExpectedActual(expected, actual)

	if len(comparison.MissingCollections) != 1 || comparison.MissingCollections[0] != "memories" {
		t.Fatalf("MissingCollections = %v, want memories", comparison.MissingCollections)
	}
	if len(comparison.UnexpectedCollections) != 1 || comparison.UnexpectedCollections[0] != "extra" {
		t.Fatalf("UnexpectedCollections = %v, want extra", comparison.UnexpectedCollections)
	}
	diff := comparison.FieldDifferences["users"]
	if len(diff.MissingFields) != 1 || diff.MissingFields[0] != "role" {
		t.Fatalf("MissingFields = %v, want role", diff.MissingFields)
	}
	if len(diff.UnexpectedFields) != 1 || diff.UnexpectedFields[0] != "unexpected" {
		t.Fatalf("UnexpectedFields = %v, want unexpected", diff.UnexpectedFields)
	}
}

func TestRenderSchemaMarkdown(t *testing.T) {
	snapshot := BuildSnapshot("test-project", []CollectionSnapshot{
		InferCollection("users", []map[string]any{{"id": "1", "email": "a@example.com"}}, true),
	}, ExpectedModel{Collections: map[string][]string{"users": {"id", "email"}}}, time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC))

	got := RenderSchemaMarkdown(snapshot)
	if !strings.Contains(got, "# Database Schema Snapshot") {
		t.Fatalf("schema markdown missing title")
	}
	if !strings.Contains(got, "`users`") {
		t.Fatalf("schema markdown missing users collection")
	}
	if !strings.Contains(got, "Sensitive") {
		t.Fatalf("schema markdown missing sensitive column")
	}
}

func TestParseExpectedModel(t *testing.T) {
	doc := "### `users`\n\n| Aspecto | Detalle |\n| --- | --- |\n| Campos principales | `id`, `email`, `password_hash`, timestamps. |\n"

	got := ParseExpectedModel(doc)
	fields := got.Collections["users"]
	if len(fields) != 3 {
		t.Fatalf("fields = %v, want 3 fields", fields)
	}
}
