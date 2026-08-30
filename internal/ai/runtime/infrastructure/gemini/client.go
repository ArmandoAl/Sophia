package gemini

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

const defaultEndpoint = "https://generativelanguage.googleapis.com/v1beta"

var (
	ErrMissingAPIKey        = errors.New("gemini api key is required")
	ErrMissingModel         = errors.New("gemini model is required")
	ErrInvalidGeminiOutput  = errors.New("gemini returned invalid structured output")
	ErrGeminiRequestFailed  = errors.New("gemini request failed")
	ErrGeminiEmptyCandidate = errors.New("gemini returned no candidates")
)

const (
	ErrorTypeUnauthorized     = "unauthorized"
	ErrorTypePermissionDenied = "permission_denied"
	ErrorTypeModelNotFound    = "model_not_found"
	ErrorTypeQuota            = "quota"
	ErrorTypeRequestFailed    = "request_failed"
)

type GeminiError struct {
	StatusCode   int
	ErrorType    string
	BodyRedacted string
}

func (e *GeminiError) Error() string {
	if e == nil {
		return ErrGeminiRequestFailed.Error()
	}
	if e.BodyRedacted == "" {
		return fmt.Sprintf("%s: status=%d type=%s", ErrGeminiRequestFailed, e.StatusCode, e.ErrorType)
	}
	return fmt.Sprintf("%s: status=%d type=%s body=%s", ErrGeminiRequestFailed, e.StatusCode, e.ErrorType, e.BodyRedacted)
}

func (e *GeminiError) Unwrap() error {
	return ErrGeminiRequestFailed
}

func (e *GeminiError) ProviderErrorType() string {
	if e == nil {
		return ""
	}
	return e.ErrorType
}

type Client struct {
	apiKey     string
	model      string
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

func NewClient(apiKey, model string, opts ...Option) (*Client, error) {
	client := &Client{
		apiKey:     strings.TrimSpace(apiKey),
		model:      strings.TrimSpace(model),
		endpoint:   defaultEndpoint,
		httpClient: &http.Client{Timeout: 20 * time.Second},
	}
	for _, opt := range opts {
		opt(client)
	}
	if client.apiKey == "" {
		return nil, ErrMissingAPIKey
	}
	if client.model == "" {
		return nil, ErrMissingModel
	}
	if client.endpoint == "" {
		client.endpoint = defaultEndpoint
	}
	return client, nil
}

func (c *Client) Generate(ctx context.Context, request runtimedomain.ModelRequest) (runtimedomain.ModelResponse, error) {
	payload, err := json.Marshal(geminiRequest{
		Contents: []content{{
			Role: "user",
			Parts: []part{{
				Text: buildPrompt(request),
			}},
		}},
		GenerationConfig: generationConfig{
			Temperature:      0.2,
			ResponseMimeType: "application/json",
			ResponseSchema:   responseSchema(),
		},
	})
	if err != nil {
		return runtimedomain.ModelResponse{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url(), bytes.NewReader(payload))
	if err != nil {
		return runtimedomain.ModelResponse{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return runtimedomain.ModelResponse{}, fmt.Errorf("%w: %v", ErrGeminiRequestFailed, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return runtimedomain.ModelResponse{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return runtimedomain.ModelResponse{}, newGeminiHTTPError(resp.StatusCode, body, c.apiKey)
	}
	return parseResponse(body)
}

func (c *Client) url() string {
	return fmt.Sprintf("%s/models/%s:generateContent?key=%s", c.endpoint, c.model, c.apiKey)
}

func buildPrompt(request runtimedomain.ModelRequest) string {
	payload := map[string]any{
		"instructions": []string{
			"You are Sofia's planning runtime. Return JSON only.",
			"Never execute actions. Only propose actions.",
			"Only use tools from available_tools.",
			"Never include user_id, email, token, secret, password, or owner fields in proposed_input.",
			"If uncertain, return no proposed_actions and a concise assistant_message.",
		},
		"output_shape": map[string]any{
			"assistant_message": "string",
			"proposed_actions": []map[string]any{{
				"tool_name":             "string",
				"proposed_input":        "object",
				"reason":                "string",
				"risk_level":            "low|medium|high",
				"requires_confirmation": true,
			}},
		},
		"user_message":    truncate(request.Message, 1000),
		"context_summary": request.Context,
		"available_tools": request.Tools,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return `{"instructions":["Return no proposed_actions."]}`
	}
	return string(raw)
}

func parseResponse(body []byte) (runtimedomain.ModelResponse, error) {
	var response geminiResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return runtimedomain.ModelResponse{}, fmt.Errorf("%w: %v", ErrInvalidGeminiOutput, err)
	}
	if len(response.Candidates) == 0 || len(response.Candidates[0].Content.Parts) == 0 {
		return runtimedomain.ModelResponse{}, ErrGeminiEmptyCandidate
	}
	text := strings.TrimSpace(response.Candidates[0].Content.Parts[0].Text)
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimPrefix(text, "```")
	text = strings.TrimSuffix(text, "```")
	text = strings.TrimSpace(text)

	var structured structuredOutput
	if err := json.Unmarshal([]byte(text), &structured); err != nil {
		return runtimedomain.ModelResponse{}, fmt.Errorf("%w: %v", ErrInvalidGeminiOutput, err)
	}
	if strings.TrimSpace(structured.AssistantMessage) == "" {
		return runtimedomain.ModelResponse{}, ErrInvalidGeminiOutput
	}
	modelResponse := runtimedomain.ModelResponse{AssistantMessage: truncate(structured.AssistantMessage, 500)}
	for _, action := range structured.ProposedActions {
		if strings.TrimSpace(action.ToolName) == "" || len(action.ProposedInput) == 0 || !json.Valid(action.ProposedInput) {
			return runtimedomain.ModelResponse{}, ErrInvalidGeminiOutput
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

func responseSchema() map[string]any {
	return map[string]any{
		"type":     "object",
		"required": []string{"assistant_message", "proposed_actions"},
		"properties": map[string]any{
			"assistant_message": map[string]any{"type": "string"},
			"proposed_actions": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type":     "object",
					"required": []string{"tool_name", "proposed_input", "reason", "risk_level", "requires_confirmation"},
					"properties": map[string]any{
						"tool_name":             map[string]any{"type": "string"},
						"proposed_input":        map[string]any{"type": "object"},
						"reason":                map[string]any{"type": "string"},
						"risk_level":            map[string]any{"type": "string", "enum": []string{"low", "medium", "high"}},
						"requires_confirmation": map[string]any{"type": "boolean"},
					},
				},
			},
		},
	}
}

type geminiRequest struct {
	Contents         []content        `json:"contents"`
	GenerationConfig generationConfig `json:"generationConfig"`
}

type generationConfig struct {
	Temperature      float64        `json:"temperature"`
	ResponseMimeType string         `json:"responseMimeType"`
	ResponseSchema   map[string]any `json:"responseSchema"`
}

type content struct {
	Role  string `json:"role,omitempty"`
	Parts []part `json:"parts"`
}

type part struct {
	Text string `json:"text"`
}

type geminiResponse struct {
	Candidates []struct {
		Content content `json:"content"`
	} `json:"candidates"`
}

func newGeminiHTTPError(statusCode int, body []byte, apiKey string) *GeminiError {
	bodyRedacted := redactErrorBody(body, apiKey)
	return &GeminiError{
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
