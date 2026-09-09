package deepseek

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	runtimedomain "github.com/armandoalvarado/sofia-backend/internal/ai/runtime/domain"
)

const defaultEndpoint = "https://api.deepseek.com"

var (
	ErrMissingAPIKey         = errors.New("deepseek api key is required")
	ErrMissingModel          = errors.New("deepseek model is required")
	ErrInvalidDeepSeekOutput = errors.New("deepseek returned invalid structured output")
	ErrDeepSeekRequestFailed = errors.New("deepseek request failed")
	ErrDeepSeekEmptyChoice   = errors.New("deepseek returned no choices")
)

const (
	ErrorTypeUnauthorized     = "unauthorized"
	ErrorTypePermissionDenied = "permission_denied"
	ErrorTypeModelNotFound    = "model_not_found"
	ErrorTypeQuota            = "quota"
	ErrorTypeRequestFailed    = "request_failed"
)

type DeepSeekError struct {
	StatusCode   int
	ErrorType    string
	BodyRedacted string
}

func (e *DeepSeekError) Error() string {
	if e == nil {
		return ErrDeepSeekRequestFailed.Error()
	}
	if e.BodyRedacted == "" {
		return fmt.Sprintf("%s: status=%d type=%s", ErrDeepSeekRequestFailed, e.StatusCode, e.ErrorType)
	}
	return fmt.Sprintf("%s: status=%d type=%s body=%s", ErrDeepSeekRequestFailed, e.StatusCode, e.ErrorType, e.BodyRedacted)
}

func (e *DeepSeekError) Unwrap() error {
	return ErrDeepSeekRequestFailed
}

func (e *DeepSeekError) ProviderErrorType() string {
	if e == nil {
		return ""
	}
	return e.ErrorType
}

type Client struct {
	apiKey     string
	models     map[string]string
	endpoint   string
	httpClient *http.Client
}

type Option func(*Client)

func WithEndpoint(endpoint string) Option {
	return func(c *Client) {
		c.endpoint = strings.TrimRight(strings.TrimSpace(endpoint), "/")
	}
}

func WithHTTPClient(client *http.Client) Option {
	return func(c *Client) {
		if client != nil {
			c.httpClient = client
		}
	}
}

func NewClient(apiKey string, models map[string]string, opts ...Option) (*Client, error) {
	client := &Client{
		apiKey:     strings.TrimSpace(apiKey),
		models:     make(map[string]string, len(models)),
		endpoint:   defaultEndpoint,
		httpClient: &http.Client{Timeout: 45 * time.Second},
	}
	for task, model := range models {
		client.models[strings.TrimSpace(task)] = strings.TrimSpace(model)
	}
	for _, opt := range opts {
		opt(client)
	}
	if client.apiKey == "" {
		return nil, ErrMissingAPIKey
	}
	if client.models[""] == "" {
		return nil, ErrMissingModel
	}
	if client.endpoint == "" {
		client.endpoint = defaultEndpoint
	}
	return client, nil
}

func (c *Client) Generate(ctx context.Context, request runtimedomain.ModelRequest) (runtimedomain.ModelResponse, error) {
	system, err := runtimedomain.BuildPromptPrefix(request)
	if err != nil {
		return runtimedomain.ModelResponse{}, err
	}
	userContent, err := runtimedomain.BuildPromptSuffix(request)
	if err != nil {
		return runtimedomain.ModelResponse{}, err
	}
	if request.Task == runtimedomain.TaskSynthesize || request.Task == runtimedomain.TaskExtract {
		system = []byte(synthesisSystemPrompt())
		if request.Task == runtimedomain.TaskExtract {
			system = []byte("Extract only observable facts and preferences from the supplied conversations. Return JSON only with beliefs; do not infer sensitive traits.")
		}
		userContent = []byte(request.Message)
	}
	model := c.models[request.Task]
	if model == "" {
		model = c.models[""]
	}
	payload, err := json.Marshal(deepSeekRequest{
		Model: model,
		Messages: []message{
			{Role: "system", Content: string(system)},
			{Role: "user", Content: string(userContent)},
		},
		Temperature:    0.2,
		ResponseFormat: responseFormat{Type: "json_object"},
	})
	if err != nil {
		return runtimedomain.ModelResponse{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url(), bytes.NewReader(payload))
	if err != nil {
		return runtimedomain.ModelResponse{}, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return runtimedomain.ModelResponse{}, fmt.Errorf("%w: %v", ErrDeepSeekRequestFailed, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return runtimedomain.ModelResponse{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return runtimedomain.ModelResponse{}, newDeepSeekHTTPError(resp.StatusCode, body, c.apiKey)
	}
	return parseResponse(body, request.Task, model)
}

func (c *Client) url() string {
	return c.endpoint + "/chat/completions"
}

func synthesisSystemPrompt() string {
	return strings.Join([]string{
		"You synthesize Sofia's user beliefs from DECISIONS only.",
		"Never read conversations, messages, or chat transcripts.",
		"Return JSON only with reinforced, contradicted, and novel.",
	}, "\n")
}

func parseResponse(body []byte, task, model string) (runtimedomain.ModelResponse, error) {
	var response deepSeekResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return runtimedomain.ModelResponse{}, fmt.Errorf("%w: %v", ErrInvalidDeepSeekOutput, err)
	}
	if len(response.Choices) == 0 {
		return runtimedomain.ModelResponse{}, ErrDeepSeekEmptyChoice
	}
	text := strings.TrimSpace(response.Choices[0].Message.Content)
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimPrefix(text, "```")
	text = strings.TrimSuffix(text, "```")
	text = strings.TrimSpace(text)
	usage := runtimedomain.Usage{
		InputTokens:       response.Usage.PromptTokens,
		OutputTokens:      response.Usage.CompletionTokens,
		CachedInputTokens: response.Usage.PromptCacheHitTokens,
		Model:             model,
	}
	if task == runtimedomain.TaskSynthesize || task == runtimedomain.TaskExtract {
		if !json.Valid([]byte(text)) {
			return runtimedomain.ModelResponse{}, ErrInvalidDeepSeekOutput
		}
		return runtimedomain.ModelResponse{AssistantMessage: text, Usage: usage}, nil
	}

	var structured structuredOutput
	if err := json.Unmarshal([]byte(text), &structured); err != nil {
		return runtimedomain.ModelResponse{}, fmt.Errorf("%w: %v", ErrInvalidDeepSeekOutput, err)
	}
	if strings.TrimSpace(structured.AssistantMessage) == "" {
		return runtimedomain.ModelResponse{}, ErrInvalidDeepSeekOutput
	}
	modelResponse := runtimedomain.ModelResponse{AssistantMessage: truncate(structured.AssistantMessage, 500), Usage: usage}
	for _, action := range structured.ProposedActions {
		if strings.TrimSpace(action.ToolName) == "" || len(action.ProposedInput) == 0 || !json.Valid(action.ProposedInput) {
			return runtimedomain.ModelResponse{}, ErrInvalidDeepSeekOutput
		}
		modelResponse.PlannedActions = append(modelResponse.PlannedActions, runtimedomain.PlannedAction{
			ToolName:      strings.TrimSpace(action.ToolName),
			ProposedInput: cloneJSON(action.ProposedInput),
			Reason:        truncate(action.Reason, 500),
			RiskLevel:     strings.TrimSpace(action.RiskLevel),
		})
	}
	return modelResponse, nil
}

type deepSeekRequest struct {
	Model          string         `json:"model"`
	Messages       []message      `json:"messages"`
	Temperature    float64        `json:"temperature"`
	ResponseFormat responseFormat `json:"response_format"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type responseFormat struct {
	Type string `json:"type"`
}

type deepSeekResponse struct {
	Choices []struct {
		Message message `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens         int `json:"prompt_tokens"`
		CompletionTokens     int `json:"completion_tokens"`
		PromptCacheHitTokens int `json:"prompt_cache_hit_tokens"`
	} `json:"usage"`
}

func newDeepSeekHTTPError(statusCode int, body []byte, apiKey string) *DeepSeekError {
	bodyRedacted := redactErrorBody(body, apiKey)
	return &DeepSeekError{
		StatusCode:   statusCode,
		ErrorType:    classifyHTTPError(statusCode, bodyRedacted),
		BodyRedacted: bodyRedacted,
	}
}

func classifyHTTPError(statusCode int, body string) string {
	lowerBody := strings.ToLower(body)
	switch statusCode {
	case http.StatusUnauthorized:
		return ErrorTypeUnauthorized
	case http.StatusForbidden:
		if strings.Contains(lowerBody, "quota") || strings.Contains(lowerBody, "rate limit") || strings.Contains(lowerBody, "exceeded") {
			return ErrorTypeQuota
		}
		return ErrorTypePermissionDenied
	case http.StatusNotFound:
		return ErrorTypeModelNotFound
	case http.StatusTooManyRequests:
		return ErrorTypeQuota
	}
	if strings.Contains(lowerBody, "quota") || strings.Contains(lowerBody, "rate limit") {
		return ErrorTypeQuota
	}
	if strings.Contains(lowerBody, "permission") || strings.Contains(lowerBody, "forbidden") {
		return ErrorTypePermissionDenied
	}
	if strings.Contains(lowerBody, "model") && strings.Contains(lowerBody, "not found") {
		return ErrorTypeModelNotFound
	}
	return ErrorTypeRequestFailed
}

func redactErrorBody(body []byte, apiKey string) string {
	value := strings.TrimSpace(string(body))
	if value == "" {
		return ""
	}
	if apiKey = strings.TrimSpace(apiKey); apiKey != "" {
		value = strings.ReplaceAll(value, apiKey, "[REDACTED_API_KEY]")
	}
	var payload any
	if json.Unmarshal([]byte(value), &payload) == nil {
		value = redactJSONValue(payload)
	}
	return truncate(value, 500)
}

func redactJSONValue(value any) string {
	redacted := redactAny(value)
	raw, err := json.Marshal(redacted)
	if err != nil {
		return "[REDACTED]"
	}
	return string(raw)
}

func redactAny(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		result := make(map[string]any, len(typed))
		for key, item := range typed {
			if isSensitiveKey(key) {
				result[key] = "[REDACTED]"
				continue
			}
			result[key] = redactAny(item)
		}
		return result
	case []any:
		result := make([]any, 0, len(typed))
		for _, item := range typed {
			result = append(result, redactAny(item))
		}
		return result
	case string:
		return redactSensitiveString(typed)
	default:
		return value
	}
}

func isSensitiveKey(key string) bool {
	key = strings.ToLower(strings.TrimSpace(key))
	return strings.Contains(key, "key") ||
		strings.Contains(key, "token") ||
		strings.Contains(key, "secret") ||
		strings.Contains(key, "credential")
}

func redactSensitiveString(value string) string {
	if strings.Contains(strings.ToLower(value), "api key") {
		return strings.ReplaceAll(value, "API key", "API key [REDACTED]")
	}
	return value
}

type structuredOutput struct {
	AssistantMessage string `json:"assistant_message"`
	ProposedActions  []struct {
		ToolName             string          `json:"tool_name"`
		ProposedInput        json.RawMessage `json:"proposed_input"`
		Reason               string          `json:"reason"`
		RiskLevel            string          `json:"risk_level"`
		RequiresConfirmation bool            `json:"requires_confirmation"`
	} `json:"proposed_actions"`
}

func truncate(value string, limit int) string {
	value = strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(value, "\n", " "), "\t", " "))
	if len([]rune(value)) <= limit {
		return value
	}
	return string([]rune(value)[:limit]) + "...[TRUNCATED]"
}

func cloneJSON(value json.RawMessage) json.RawMessage {
	if value == nil {
		return nil
	}
	return append(json.RawMessage(nil), value...)
}
