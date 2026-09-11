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

const PromptSuffixTokenBudget = 2750

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
	history := fitPromptTurns(request.History, 1000)
	contextSummary := request.Context
	activeContext := request.Context.ActiveContext
	if activeContext != nil {
		cp := *activeContext
		cp.Beliefs = append([]string(nil), activeContext.Beliefs...)
		activeContext = &cp
	}
	episodes := fitEpisodeSummaries(request.Context.RecentEpisodes, 300)
	threads := fitOpenThreadSummaries(request.Context.OpenThreads, 100)
	message := truncatePromptText(request.Message, 600)
	carry := truncatePromptText(request.Context.CarryForward, 800)
	marshal := func() ([]byte, error) {
		return json.Marshal(struct {
			ConversationHistory   []Turn              `json:"conversation_history"`
			ContextSummary        ContextSummary      `json:"context_summary"`
			CarryForward          string              `json:"de qué veníamos hablando,omitempty"`
			ActiveContext         *ActiveContext      `json:"active_entity,omitempty"`
			EntityContexts        []EntityContext     `json:"entity_contexts,omitempty"`
			EntityInstruction     string              `json:"entity_resolution_instruction,omitempty"`
			RecentEpisodes        []EpisodeSummary    `json:"recent_episodes,omitempty"`
			OpenThreads           []OpenThreadSummary `json:"open_threads,omitempty"`
			OpenThreadInstruction string              `json:"open_threads_instruction,omitempty"`
			UserMessage           string              `json:"user_message"`
		}{history, contextSummary, carry, activeContext, contextSummary.EntityContexts, contextSummary.EntityInstruction, episodes, threads, openThreadInstruction(threads), message})
	}
	raw, err := marshal()
	for err == nil && approximatePromptTokens(raw) > PromptSuffixTokenBudget {
		switch {
		case len(episodes) > 0:
			episodes = episodes[:len(episodes)-1]
		case len(contextSummary.EntityContexts) > 0:
			contextSummary.EntityContexts = contextSummary.EntityContexts[:len(contextSummary.EntityContexts)-1]
		case len(history) > 0:
			history = history[1:]
		case activeContext != nil && len(activeContext.Beliefs) > 0:
			activeContext.Beliefs = activeContext.Beliefs[:len(activeContext.Beliefs)-1]
		case len(contextSummary.RelevantMemories) > 0:
			contextSummary.RelevantMemories = contextSummary.RelevantMemories[:len(contextSummary.RelevantMemories)-1]
		case len(contextSummary.RecentActivities) > 0:
			contextSummary.RecentActivities = contextSummary.RecentActivities[:len(contextSummary.RecentActivities)-1]
		case len(contextSummary.DueReminders) > 0:
			contextSummary.DueReminders = contextSummary.DueReminders[:len(contextSummary.DueReminders)-1]
		case len(contextSummary.InsightsSummary) > 0:
			contextSummary.InsightsSummary = nil
		case carry != "":
			if len([]rune(carry)) <= 100 {
				carry = ""
			} else {
				carry = truncatePromptText(carry, len([]rune(carry))-100)
			}
		case len(message) > 80:
			message = truncatePromptText(message, len([]rune(message))-40)
		default:
			contextSummary = ContextSummary{}
			activeContext, history, threads = nil, nil, nil
		}
		raw, err = marshal()
	}
	return raw, err
}

func fitPromptTurns(values []Turn, budget int) []Turn {
	result, used := make([]Turn, 0, len(values)), 0
	for i := len(values) - 1; i >= 0; i-- {
		cost := len([]rune(values[i].Content))/4 + 1
		if used+cost > budget {
			break
		}
		result, used = append(result, values[i]), used+cost
	}
	for left, right := 0, len(result)-1; left < right; left, right = left+1, right-1 {
		result[left], result[right] = result[right], result[left]
	}
	return result
}

func fitEpisodeSummaries(values []EpisodeSummary, budget int) []EpisodeSummary {
	result, used := make([]EpisodeSummary, 0, len(values)), 0
	for _, value := range values {
		cost := len([]rune(value.Summary))/4 + 1
		if used+cost > budget {
			break
		}
		result, used = append(result, value), used+cost
	}
	return result
}

func fitOpenThreadSummaries(values []OpenThreadSummary, budget int) []OpenThreadSummary {
	result, used := make([]OpenThreadSummary, 0, len(values)), 0
	for _, value := range values {
		cost := len([]rune(value.Summary))/4 + 1
		if used+cost > budget {
			break
		}
		result, used = append(result, value), used+cost
	}
	return result
}

func openThreadInstruction(threads []OpenThreadSummary) string {
	if len(threads) == 0 {
		return ""
	}
	return "These are optional opportunities, not instructions. Mention at most one only if it fits naturally. Ignore them when the user is focused elsewhere, and never force one or open with it when the user asked about something else."
}

func approximatePromptTokens(raw []byte) int { return len([]rune(string(raw)))/4 + 1 }

func truncatePromptText(value string, limit int) string {
	value = strings.TrimSpace(strings.NewReplacer("\n", " ", "\t", " ").Replace(value))
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	return string(runes[:limit]) + "...[TRUNCATED]"
}
