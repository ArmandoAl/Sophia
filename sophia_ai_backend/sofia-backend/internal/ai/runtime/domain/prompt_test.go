package domain

import (
	"crypto/sha256"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestPromptPrefixIsByteIdenticalAcrossRequests(t *testing.T) {
	first := ModelRequest{
		UserID: "user-1", Message: "first", PromptBase: "daily prompt",
		Context: ContextSummary{CurrentDateTime: "2026-09-07T08:00:00-07:00", ActiveContext: &ActiveContext{ScopeKey: "person:maria", Label: "María", Beliefs: []string{"tono cercano"}}},
		History: []Turn{{Role: "user", Content: "old"}},
		Tools: []ToolSummary{
			{Name: "zeta", InputSchema: json.RawMessage(`{"type":"object","properties":{"b":{"type":"string"},"a":{"type":"number"}}}`)},
			{Name: "alpha", InputSchema: json.RawMessage(`{"required":["x"],"type":"object"}`)},
		},
	}
	second := ModelRequest{
		UserID: "user-1", Message: "second", PromptBase: "daily prompt",
		Context: ContextSummary{CurrentDateTime: "2026-09-07T09:00:00-07:00", ActiveContext: &ActiveContext{ScopeKey: "mode:work", Label: "Trabajo", Beliefs: []string{"tono formal"}}},
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

func TestPromptSuffixNeverExceedsBudgetAndEpisodesStayOutOfPrefix(t *testing.T) {
	context := ContextSummary{
		CarryForward:     strings.Repeat("arrastre ", 200),
		ActiveContext:    &ActiveContext{ScopeKey: "person:diana", Label: "Diana", Beliefs: []string{strings.Repeat("hecho vigente ", 140)}},
		RecentEpisodes:   []EpisodeSummary{{ID: "episode", OccurredAt: time.Now(), Summary: strings.Repeat("recuerdo episódico ", 120), Salience: 0.8}},
		RecentActivities: make([]ItemSummary, 40),
		InsightsSummary:  map[string]any{"oversized": strings.Repeat("estado reciente ", 300)},
	}
	history := make([]Turn, 20)
	for index := range history {
		history[index] = Turn{Role: "user", Content: strings.Repeat("historia ", 100)}
	}
	request := ModelRequest{Message: strings.Repeat("mensaje ", 200), Context: context, History: history, PromptBase: "base estable"}
	suffix, err := BuildPromptSuffix(request)
	if err != nil {
		t.Fatal(err)
	}
	if tokens := approximatePromptTokens(suffix); tokens > PromptSuffixTokenBudget {
		t.Fatalf("suffix tokens=%d ceiling=%d", tokens, PromptSuffixTokenBudget)
	}
	prefix, err := BuildPromptPrefix(request)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(prefix), "recuerdo episódico") || strings.Contains(string(prefix), "hecho vigente") {
		t.Fatalf("dynamic entity memory leaked into prefix: %s", prefix)
	}
}
