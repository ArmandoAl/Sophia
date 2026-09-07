package jsonschema

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestValidateAcceptsValidDocument(t *testing.T) {
	schema := json.RawMessage(`{"type":"object","required":["title"],"properties":{"title":{"type":"string"},"tags":{"type":"array","items":{"type":"string"}}},"additionalProperties":false}`)
	doc := json.RawMessage(`{"title":"Study","tags":["school"]}`)

	if err := Validate(schema, doc); err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}
}

func TestValidateRejectsMissingRequired(t *testing.T) {
	schema := json.RawMessage(`{"type":"object","required":["title"],"properties":{"title":{"type":"string"}}}`)
	doc := json.RawMessage(`{}`)

	if err := Validate(schema, doc); !errors.Is(err, ErrInvalidDocument) {
		t.Fatalf("expected ErrInvalidDocument, got %v", err)
	}
}

func TestValidateRejectsWrongType(t *testing.T) {
	schema := json.RawMessage(`{"type":"object","properties":{"title":{"type":"string"}}}`)
	doc := json.RawMessage(`{"title":123}`)

	if err := Validate(schema, doc); !errors.Is(err, ErrInvalidDocument) {
		t.Fatalf("expected ErrInvalidDocument, got %v", err)
	}
}

func TestValidateRejectsAdditionalProperties(t *testing.T) {
	schema := json.RawMessage(`{"type":"object","properties":{"title":{"type":"string"}},"additionalProperties":false}`)
	doc := json.RawMessage(`{"title":"Study","user_id":"other"}`)

	if err := Validate(schema, doc); !errors.Is(err, ErrInvalidDocument) {
		t.Fatalf("expected ErrInvalidDocument, got %v", err)
	}
}
