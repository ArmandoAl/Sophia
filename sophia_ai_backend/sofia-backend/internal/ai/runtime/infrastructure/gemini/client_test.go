package gemini

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	actionsdomain "github.com/armandoalvarado/sofia-backend/internal/ai/actions/domain"
	runtimedomain "github.com/armandoalvarado/sofia-backend/internal/ai/runtime/domain"
	toolsdomain "github.com/armandoalvarado/sofia-backend/internal/tools/domain"
)

func TestNewClientRequiresAPIKey(t *testing.T) {
	if _, err := NewClient("", "gemini-test"); !errors.Is(err, ErrMissingAPIKey) {
		t.Fatalf("expected ErrMissingAPIKey, got %v", err)
	}
}

func TestGenerateParsesStructuredOutput(t *testing.T) {
	clientHTTP := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if !strings.Contains(r.URL.RawQuery, "key=test-key") {
			t.Fatalf("missing api key in query: %s", r.URL.String())
		}
		var req geminiRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.GenerationConfig.ResponseMimeType != "application/json" {
			t.Fatalf("expected JSON response mime type: %+v", req.GenerationConfig)
		}
		if req.SystemInstruction == nil || !strings.Contains(req.SystemInstruction.Parts[0].Text, "available_tools") {
			t.Fatalf("missing stable prompt prefix: %+v", req.SystemInstruction)
		}
		if len(req.Contents) == 0 || len(req.Contents[0].Parts) == 0 || !strings.Contains(req.Contents[0].Parts[0].Text, "conversation_history") {
			t.Fatalf("missing conversation_history in prompt: %+v", req.Contents)
		}
		output := `{"assistant_message":"Listo, puedo proponerte el recordatorio.","proposed_actions":[{"tool_name":"create_reminder","proposed_input":{"title":"Estudiar","scheduled_at":"2026-07-02T15:00:00Z","timezone":"America/Tijuana"},"reason":"El usuario pidio un recordatorio.","risk_level":"medium","requires_confirmation":true}]}`
		body, _ := json.Marshal(map[string]any{
			"candidates": []map[string]any{{
				"content": map[string]any{
					"parts": []map[string]string{{"text": output}},
				},
			}},
		})
		return jsonResponse(http.StatusOK, body), nil
	})}

	client, err := NewClient("test-key", "gemini-test", WithEndpoint("https://gemini.test"), WithHTTPClient(clientHTTP))
	if err != nil {
		t.Fatal(err)
	}
	response, err := client.Generate(context.Background(), runtimedomain.ModelRequest{
		UserID:  "user_001",
		Message: "Recuérdame estudiar mañana",
		Context: runtimedomain.ContextSummary{User: runtimedomain.UserSummary{ID: "user_001", Email: "[REDACTED]"}},
		Tools: []runtimedomain.ToolSummary{{
			Name:                 toolsdomain.ToolCreateReminder,
			Category:             toolsdomain.CategoryReminders,
			RequiresConfirmation: true,
			RiskLevel:            actionsdomain.RiskMedium,
			InputSchema:          json.RawMessage(`{"type":"object"}`),
		}},
	})
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}
	if response.AssistantMessage == "" || len(response.PlannedActions) != 1 {
		t.Fatalf("unexpected response: %+v", response)
	}
	if response.PlannedActions[0].ToolName != toolsdomain.ToolCreateReminder {
		t.Fatalf("unexpected planned action: %+v", response.PlannedActions[0])
	}
}

func TestGenerateSynthesisUsesDecisionsPromptAndUsage(t *testing.T) {
	clientHTTP := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		var req geminiRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if len(req.Contents) == 0 || strings.Contains(req.Contents[0].Parts[0].Text, "conversation_history") {
			t.Fatalf("synthesis prompt included conversation: %+v", req.Contents)
		}
		if _, ok := req.GenerationConfig.ResponseSchema["properties"].(map[string]any)["reinforced"]; !ok {
			t.Fatalf("expected synthesis schema, got %+v", req.GenerationConfig.ResponseSchema)
		}
		output := `{"reinforced":[],"contradicted":[],"novel":[]}`
		body, _ := json.Marshal(map[string]any{
			"candidates": []map[string]any{{
				"content": map[string]any{"parts": []map[string]string{{"text": output}}},
			}},
			"usageMetadata": map[string]any{"promptTokenCount": 21, "candidatesTokenCount": 9, "cachedContentTokenCount": 13},
		})
		return jsonResponse(http.StatusOK, body), nil
	})}

	client, err := NewClient("test-key", "gemini-test", WithEndpoint("https://gemini.test"), WithHTTPClient(clientHTTP))
	if err != nil {
		t.Fatal(err)
	}
	response, err := client.Generate(context.Background(), runtimedomain.ModelRequest{
		Message: `{"decisions":{"corrections":[]}}`,
		Task:    runtimedomain.TaskSynthesize,
	})
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}
	if response.AssistantMessage != `{"reinforced":[],"contradicted":[],"novel":[]}` {
		t.Fatalf("unexpected synthesis output: %s", response.AssistantMessage)
	}
	if response.Usage.InputTokens != 21 || response.Usage.OutputTokens != 9 || response.Usage.CachedInputTokens != 13 || response.Usage.Model != "gemini-test" {
		t.Fatalf("unexpected usage: %+v", response.Usage)
	}
}

func TestGenerateRejectsInvalidStructuredOutput(t *testing.T) {
	clientHTTP := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		body, _ := json.Marshal(map[string]any{
			"candidates": []map[string]any{{
				"content": map[string]any{
					"parts": []map[string]string{{"text": `{"assistant_message":"","proposed_actions":[{"tool_name":"create_memory"}]}`}},
				},
			}},
		})
		return jsonResponse(http.StatusOK, body), nil
	})}

	client, err := NewClient("test-key", "gemini-test", WithEndpoint("https://gemini.test"), WithHTTPClient(clientHTTP))
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.Generate(context.Background(), runtimedomain.ModelRequest{Message: "hello"})
	if !errors.Is(err, ErrInvalidGeminiOutput) {
		t.Fatalf("expected ErrInvalidGeminiOutput, got %v", err)
	}
}

func TestGenerateReturnsRequestErrorForHTTPFailure(t *testing.T) {
	clientHTTP := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return jsonResponse(http.StatusUnauthorized, []byte(`{"error":{"message":"bad key"}}`)), nil
	})}

	client, err := NewClient("test-key", "gemini-test", WithEndpoint("https://gemini.test"), WithHTTPClient(clientHTTP))
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.Generate(context.Background(), runtimedomain.ModelRequest{Message: "hello"})
	if !errors.Is(err, ErrGeminiRequestFailed) {
		t.Fatalf("expected ErrGeminiRequestFailed, got %v", err)
	}
	var geminiErr *GeminiError
	if !errors.As(err, &geminiErr) {
		t.Fatalf("expected GeminiError, got %T", err)
	}
	if geminiErr.StatusCode != http.StatusUnauthorized || geminiErr.ErrorType != ErrorTypeUnauthorized {
		t.Fatalf("unexpected GeminiError: %+v", geminiErr)
	}
}

func TestGenerateClassifiesPermissionDeniedAndRedactsAPIKey(t *testing.T) {
	clientHTTP := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return jsonResponse(http.StatusForbidden, []byte(`{"error":{"message":"Permission denied for API key test-key","status":"PERMISSION_DENIED"}}`)), nil
	})}

	client, err := NewClient("test-key", "gemini-test", WithEndpoint("https://gemini.test"), WithHTTPClient(clientHTTP))
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.Generate(context.Background(), runtimedomain.ModelRequest{Message: "hello"})
	var geminiErr *GeminiError
	if !errors.As(err, &geminiErr) {
		t.Fatalf("expected GeminiError, got %v", err)
	}
	if geminiErr.ErrorType != ErrorTypePermissionDenied {
		t.Fatalf("expected permission denied, got %+v", geminiErr)
	}
	if strings.Contains(err.Error(), "test-key") || strings.Contains(geminiErr.BodyRedacted, "test-key") {
		t.Fatalf("error leaked API key: %v", err)
	}
}

func TestGenerateClassifiesQuotaAndModelNotFound(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
		wantType   string
	}{
		{name: "quota", statusCode: http.StatusTooManyRequests, body: `{"error":{"message":"Quota exceeded"}}`, wantType: ErrorTypeQuota},
		{name: "model", statusCode: http.StatusNotFound, body: `{"error":{"message":"Model not found"}}`, wantType: ErrorTypeModelNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clientHTTP := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				return jsonResponse(tt.statusCode, []byte(tt.body)), nil
			})}
			client, err := NewClient("test-key", "gemini-test", WithEndpoint("https://gemini.test"), WithHTTPClient(clientHTTP))
			if err != nil {
				t.Fatal(err)
			}
			_, err = client.Generate(context.Background(), runtimedomain.ModelRequest{Message: "hello"})
			var geminiErr *GeminiError
			if !errors.As(err, &geminiErr) {
				t.Fatalf("expected GeminiError, got %v", err)
			}
			if geminiErr.ErrorType != tt.wantType {
				t.Fatalf("expected %s, got %+v", tt.wantType, geminiErr)
			}
		})
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func jsonResponse(status int, body []byte) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(bytes.NewReader(body)),
	}
}
