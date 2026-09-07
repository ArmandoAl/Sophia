package deepseek

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	actionsdomain "github.com/armandoalvarado/sofia-backend/internal/ai/actions/domain"
	runtimedomain "github.com/armandoalvarado/sofia-backend/internal/ai/runtime/domain"
	toolsdomain "github.com/armandoalvarado/sofia-backend/internal/tools/domain"
)

func TestNewClientRequiresAPIKey(t *testing.T) {
	if _, err := NewClient("", "deepseek-test"); !errors.Is(err, ErrMissingAPIKey) {
		t.Fatalf("expected ErrMissingAPIKey, got %v", err)
	}
}

func TestGenerateParsesStructuredOutput(t *testing.T) {
	clientHTTP := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.String() != "https://deepseek.test/chat/completions" {
			t.Fatalf("unexpected URL: %s", r.URL.String())
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Fatalf("unexpected authorization header: %q", r.Header.Get("Authorization"))
		}
		var req deepSeekRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.Model != "deepseek-test" || len(req.Messages) != 2 || req.Messages[0].Role != "system" || req.Messages[1].Role != "user" {
			t.Fatalf("unexpected request: %+v", req)
		}
		if req.ResponseFormat.Type != "json_object" || req.Temperature != 0.2 {
			t.Fatalf("expected JSON response format: %+v", req)
		}
		if !strings.Contains(req.Messages[0].Content, "Never execute actions") || !strings.Contains(req.Messages[1].Content, "available_tools") {
			t.Fatalf("missing proposal-only prompt: %+v", req.Messages)
		}

		output := `{"assistant_message":"Listo, puedo proponerte el recordatorio.","proposed_actions":[{"tool_name":"create_reminder","proposed_input":{"title":"Estudiar","scheduled_at":"2026-07-02T15:00:00Z","timezone":"America/Tijuana"},"reason":"El usuario pidio un recordatorio.","risk_level":"medium","requires_confirmation":true}]}`
		body, _ := json.Marshal(map[string]any{
			"choices": []map[string]any{{
				"message": map[string]string{"role": "assistant", "content": output},
			}},
		})
		return jsonResponse(http.StatusOK, body), nil
	})}

	client, err := NewClient("test-key", "deepseek-test", WithEndpoint("https://deepseek.test/"), WithHTTPClient(clientHTTP))
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

func TestGenerateRejectsInvalidStructuredOutput(t *testing.T) {
	outputs := []string{
		`{"assistant_message":"","proposed_actions":[]}`,
		`{"assistant_message":"Listo","proposed_actions":[{"tool_name":"create_memory"}]}`,
	}
	for _, output := range outputs {
		clientHTTP := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			body, _ := json.Marshal(map[string]any{
				"choices": []map[string]any{{"message": map[string]string{"content": output}}},
			})
			return jsonResponse(http.StatusOK, body), nil
		})}
		client, err := NewClient("test-key", "deepseek-test", WithEndpoint("https://deepseek.test"), WithHTTPClient(clientHTTP))
		if err != nil {
			t.Fatal(err)
		}
		_, err = client.Generate(context.Background(), runtimedomain.ModelRequest{Message: "hello"})
		if !errors.Is(err, ErrInvalidDeepSeekOutput) {
			t.Fatalf("expected ErrInvalidDeepSeekOutput, got %v", err)
		}
	}
}

func TestGenerateReturnsTypedRedactedHTTPError(t *testing.T) {
	clientHTTP := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return jsonResponse(http.StatusForbidden, []byte(`{"error":{"message":"Permission denied for API key test-key","api_key":"test-key"}}`)), nil
	})}

	client, err := NewClient("test-key", "deepseek-test", WithEndpoint("https://deepseek.test"), WithHTTPClient(clientHTTP))
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.Generate(context.Background(), runtimedomain.ModelRequest{Message: "hello"})
	if !errors.Is(err, ErrDeepSeekRequestFailed) {
		t.Fatalf("expected ErrDeepSeekRequestFailed, got %v", err)
	}
	var deepSeekErr *DeepSeekError
	if !errors.As(err, &deepSeekErr) {
		t.Fatalf("expected DeepSeekError, got %T", err)
	}
	if deepSeekErr.ErrorType != ErrorTypePermissionDenied || strings.Contains(err.Error(), "test-key") {
		t.Fatalf("unexpected or leaking DeepSeekError: %v", err)
	}
}

func TestGenerateClassifiesHTTPErrors(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
		wantType   string
	}{
		{name: "unauthorized", statusCode: http.StatusUnauthorized, body: `{"error":{"message":"Bad key"}}`, wantType: ErrorTypeUnauthorized},
		{name: "quota", statusCode: http.StatusTooManyRequests, body: `{"error":{"message":"Quota exceeded"}}`, wantType: ErrorTypeQuota},
		{name: "model", statusCode: http.StatusNotFound, body: `{"error":{"message":"Model not found"}}`, wantType: ErrorTypeModelNotFound},
		{name: "request", statusCode: http.StatusInternalServerError, body: `{"error":{"message":"Internal error"}}`, wantType: ErrorTypeRequestFailed},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clientHTTP := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				return jsonResponse(tt.statusCode, []byte(tt.body)), nil
			})}
			client, err := NewClient("test-key", "deepseek-test", WithEndpoint("https://deepseek.test"), WithHTTPClient(clientHTTP))
			if err != nil {
				t.Fatal(err)
			}
			_, err = client.Generate(context.Background(), runtimedomain.ModelRequest{Message: "hello"})
			var deepSeekErr *DeepSeekError
			if !errors.As(err, &deepSeekErr) {
				t.Fatalf("expected DeepSeekError, got %v", err)
			}
			if deepSeekErr.ErrorType != tt.wantType {
				t.Fatalf("expected %s, got %+v", tt.wantType, deepSeekErr)
			}
		})
	}
}

func TestGenerateWrapsTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(50 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, err := NewClient(
		"test-key",
		"deepseek-test",
		WithEndpoint(server.URL),
		WithHTTPClient(&http.Client{Timeout: 5 * time.Millisecond}),
	)
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.Generate(context.Background(), runtimedomain.ModelRequest{Message: "hello"})
	if !errors.Is(err, ErrDeepSeekRequestFailed) {
		t.Fatalf("expected ErrDeepSeekRequestFailed, got %v", err)
	}
}

func TestGenerateRejectsTruncatedChoice(t *testing.T) {
	clientHTTP := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		body, _ := json.Marshal(map[string]any{
			"choices": []map[string]any{{
				"message": map[string]string{"content": `{"assistant_message":"incomplete`},
			}},
		})
		return jsonResponse(http.StatusOK, body), nil
	})}
	client, err := NewClient("test-key", "deepseek-test", WithEndpoint("https://deepseek.test"), WithHTTPClient(clientHTTP))
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.Generate(context.Background(), runtimedomain.ModelRequest{Message: "hello"})
	if !errors.Is(err, ErrInvalidDeepSeekOutput) {
		t.Fatalf("expected ErrInvalidDeepSeekOutput, got %v", err)
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
