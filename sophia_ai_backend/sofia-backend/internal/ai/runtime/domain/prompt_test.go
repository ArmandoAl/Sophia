package domain

import (
	"crypto/sha256"
	"encoding/json"
	"testing"
)

func TestPromptPrefixIsByteIdenticalAcrossRequests(t *testing.T) {
	first := ModelRequest{
		UserID: "user-1", Message: "first", PromptBase: "daily prompt",
		Context: ContextSummary{CurrentDateTime: "2026-09-07T08:00:00-07:00"},
		History: []Turn{{Role: "user", Content: "old"}},
		Tools: []ToolSummary{
			{Name: "zeta", InputSchema: json.RawMessage(`{"type":"object","properties":{"b":{"type":"string"},"a":{"type":"number"}}}`)},
			{Name: "alpha", InputSchema: json.RawMessage(`{"required":["x"],"type":"object"}`)},
		},
	}
	second := ModelRequest{
		UserID: "user-1", Message: "second", PromptBase: "daily prompt",
		Context: ContextSummary{CurrentDateTime: "2026-09-07T09:00:00-07:00"},
		History: []Turn{{Role: "assistant", Content: "new"}},
		Tools: []ToolSummary{
			{Name: "alpha", InputSchema: json.RawMessage(`{"type":"object","required":["x"]}`)},
			{Name: "zeta", InputSchema: json.RawMessage(`{"properties":{"a":{"type":"number"},"b":{"type":"string"}},"type":"object"}`)},
		},
	}

	firstPrefix, err := BuildPromptPrefix(first)
	if err != nil {
		t.Fatal(err)
	}
	secondPrefix, err := BuildPromptPrefix(second)
	if err != nil {
		t.Fatal(err)
	}
	if sha256.Sum256(firstPrefix) != sha256.Sum256(secondPrefix) {
		t.Fatalf("prompt prefix changed between requests:\n%s\n%s", firstPrefix, secondPrefix)
	}
}
