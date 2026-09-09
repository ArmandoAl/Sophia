package domain

import (
	"context"
	"encoding/json"
	"errors"
)

const (
	ModeDryRun = "dry_run"
	ModeReady  = "ready"
)

var (
	ErrInvalidMessage       = errors.New("message is required")
	ErrUnsafeActionProposal = errors.New("action proposal is not allowed by safety policy")
)

type RuntimeRequest struct {
	UserID         string
	Message        string
	DryRun         bool
	RequestID      string
	ConversationID string
	History        []Turn
}

type Turn struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type RuntimeResponse struct {
	Mode             string                 `json:"mode"`
	RequestID        string                 `json:"request_id"`
	ContextSummary   ContextSummary         `json:"context_summary"`
	AvailableTools   []ToolSummary          `json:"available_tools"`
	ProposedActions  []ActionProposalOutput `json:"proposed_actions"`
	AssistantMessage string                 `json:"assistant_message"`
	Observability    RuntimeObservability   `json:"observability"`
}

type RuntimeObservability struct {
	ProviderLatencyMS    int64  `json:"provider_latency_ms"`
	ProviderErrorType    string `json:"provider_error_type,omitempty"`
	PlannedActionsCount  int    `json:"planned_actions_count"`
	ProposalsCount       int    `json:"proposals_count"`
	RejectedActionsCount int    `json:"rejected_actions_count"`
}

type ContextSummary struct {
	User              UserSummary       `json:"user"`
	Profile           ProfileSummary    `json:"profile"`
	AISettings        AISettingsSummary `json:"ai_settings"`
	RecentActivities  []ItemSummary     `json:"recent_activities"`
	DueReminders      []ItemSummary     `json:"due_reminders"`
	InsightsSummary   map[string]any    `json:"insights_summary"`
	RelevantMemories  []ItemSummary     `json:"relevant_memories"`
	Limits            map[string]int    `json:"limits"`
	CurrentDateTime   string            `json:"current_datetime"`
	TokenBudget       TokenBudget       `json:"token_budget"`
	MemoryIncluded    bool              `json:"memory_included"`
	RemindersIncluded bool              `json:"reminders_included"`
	PromptBase        string            `json:"-"`
}

type TokenBudget struct {
	MaxApproxTokens  int `json:"max_approx_tokens"`
	UsedApproxTokens int `json:"used_approx_tokens"`
}

type UserSummary struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type ProfileSummary struct {
	DisplayName   string `json:"display_name"`
	PreferredName string `json:"preferred_name"`
	Timezone      string `json:"timezone"`
	Locale        string `json:"locale"`
}

type AISettingsSummary struct {
	AssistantName     string `json:"assistant_name"`
	AutonomyLevel     string `json:"autonomy_level"`
	MemoryEnabled     bool   `json:"memory_enabled"`
	RemindersEnabled  bool   `json:"reminders_enabled"`
	PlanningEnabled   bool   `json:"planning_enabled"`
	PreferredLanguage string `json:"preferred_language"`
}

type ItemSummary struct {
	ID         string   `json:"id"`
	Type       string   `json:"type,omitempty"`
	Title      string   `json:"title"`
	Status     string   `json:"status,omitempty"`
	Importance string   `json:"importance,omitempty"`
	Tags       []string `json:"tags,omitempty"`
}

type ToolSummary struct {
	Name                 string          `json:"name"`
	Category             string          `json:"category"`
	RequiresConfirmation bool            `json:"requires_confirmation"`
	RiskLevel            string          `json:"risk_level"`
	InputSchema          json.RawMessage `json:"input_schema,omitempty"`
}

type ActionProposalOutput struct {
	ID                   string          `json:"id,omitempty"`
	ToolName             string          `json:"tool_name"`
	Status               string          `json:"status"`
	ProposedInput        json.RawMessage `json:"proposed_input"`
	Reason               string          `json:"reason"`
	RiskLevel            string          `json:"risk_level"`
	RequiresConfirmation bool            `json:"requires_confirmation"`
}

type PlannedAction struct {
	ToolName      string
	ProposedInput json.RawMessage
	Reason        string
	RiskLevel     string
}

type ModelRequest struct {
	UserID     string
	Message    string
	Context    ContextSummary
	Tools      []ToolSummary
	History    []Turn
	PromptBase string
	Task       string
}

const TaskSynthesize = "synthesize"

type Usage struct {
	InputTokens       int    `json:"input_tokens,omitempty"`
	OutputTokens      int    `json:"output_tokens,omitempty"`
	CachedInputTokens int    `json:"cached_input_tokens,omitempty"`
	Model             string `json:"model,omitempty"`
}

type ModelResponse struct {
	AssistantMessage string
	PlannedActions   []PlannedAction
	Usage            Usage
}

type ContextBuilder interface {
	Build(ctx context.Context, userID, message string) (ContextSummary, error)
}

type Planner interface {
	Plan(ctx context.Context, request ModelRequest) (ModelResponse, error)
}

type ModelClient interface {
	Generate(ctx context.Context, request ModelRequest) (ModelResponse, error)
}

type ToolSelector interface {
	SelectTools(ctx context.Context, context ContextSummary, message string) ([]ToolSummary, error)
}

type SafetyPolicy interface {
	ValidateProposal(ctx context.Context, userID string, action PlannedAction, context ContextSummary, tools []ToolSummary) error
	CanExecute(ctx context.Context, userID string, action PlannedAction, context ContextSummary) error
}

type RuntimeService interface {
	HandleMessage(ctx context.Context, request RuntimeRequest) (*RuntimeResponse, error)
}
