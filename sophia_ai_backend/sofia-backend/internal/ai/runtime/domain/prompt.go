package domain

import (
	"bytes"
	"encoding/json"
	"sort"
	"strings"
)

const PlanningSystemPrompt = "You are Sofia's planning runtime. Return JSON only.\n" +
	"Never execute actions. Only propose actions.\n" +
	"Only use tools from available_tools.\n" +
	"Never include user_id, email, token, secret, password, or owner fields in proposed_input.\n" +
	"Use context_summary.current_datetime as the authoritative current date and time (with UTC offset) to resolve relative expressions like today, tomorrow, or in one hour. Never ask the user what the current date is.\n" +
	"If uncertain, return no proposed_actions and a concise assistant_message."

const outputShape = `{"assistant_message":"string","proposed_actions":[{"proposed_input":"object","reason":"string","requires_confirmation":true,"risk_level":"low|medium|high","tool_name":"string"}]}`

// BuildPromptPrefix returns the provider-independent, cacheable prompt prefix.
func BuildPromptPrefix(request ModelRequest) ([]byte, error) {
	tools := append([]ToolSummary(nil), request.Tools...)
	for i := range tools {
		if len(tools[i].InputSchema) == 0 {
			continue
		}
		var schema any
		if err := json.Unmarshal(tools[i].InputSchema, &schema); err != nil {
			return nil, err
		}
		canonical, err := json.Marshal(schema)
		if err != nil {
			return nil, err
		}
		tools[i].InputSchema = canonical
	}
	sort.Slice(tools, func(i, j int) bool {
		if tools[i].Name != tools[j].Name {
			return tools[i].Name < tools[j].Name
		}
		if tools[i].Category != tools[j].Category {
			return tools[i].Category < tools[j].Category
		}
		if tools[i].RiskLevel != tools[j].RiskLevel {
			return tools[i].RiskLevel < tools[j].RiskLevel
		}
		if tools[i].RequiresConfirmation != tools[j].RequiresConfirmation {
			return !tools[i].RequiresConfirmation
		}
		return bytes.Compare(tools[i].InputSchema, tools[j].InputSchema) < 0
	})

	return json.Marshal(struct {
		Instructions   string          `json:"instructions"`
		PromptBase     string          `json:"prompt_base"`
		AvailableTools []ToolSummary   `json:"available_tools"`
		OutputShape    json.RawMessage `json:"output_shape"`
	}{PlanningSystemPrompt, strings.TrimSpace(request.PromptBase), tools, json.RawMessage(outputShape)})
}

// BuildPromptSuffix returns the request-specific prompt suffix.
func BuildPromptSuffix(request ModelRequest) ([]byte, error) {
	history := request.History
	if history == nil {
		history = []Turn{}
	}
	return json.Marshal(struct {
		ConversationHistory []Turn         `json:"conversation_history"`
		ContextSummary      ContextSummary `json:"context_summary"`
		ActiveContext       *ActiveContext `json:"active_context,omitempty"`
		UserMessage         string         `json:"user_message"`
	}{history, request.Context, request.Context.ActiveContext, truncatePromptText(request.Message, 1000)})
}

func truncatePromptText(value string, limit int) string {
	value = strings.TrimSpace(strings.NewReplacer("\n", " ", "\t", " ").Replace(value))
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	return string(runes[:limit]) + "...[TRUNCATED]"
}
