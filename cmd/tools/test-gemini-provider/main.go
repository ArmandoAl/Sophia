package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	runtimedomain "github.com/armandoalvarado/sofia-backend/internal/ai/runtime/domain"
	geminiinfra "github.com/armandoalvarado/sofia-backend/internal/ai/runtime/infrastructure/gemini"
	"github.com/armandoalvarado/sofia-backend/internal/config"
)

type diagnosticResult struct {
	GeneratedAt string
	Provider    string
	Model       string
	Status      string
	ErrorType   string
	ErrorBody   string
	LatencyMS   int64
	Reason      string
}

func main() {
	outputDir := flag.String("output-dir", "docs", "directory where gemini_provider_diagnostic.md is written")
	model := flag.String("model", "", "Gemini model override, e.g. gemini-2.5-flash or gemini-1.5-flash")
	flag.Parse()

	if err := run(context.Background(), *outputDir, *model); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context, outputDir, modelOverride string) error {
	cfg, err := config.Load()
	if err != nil {
		if !strings.Contains(err.Error(), "GEMINI_API_KEY") {
			return err
		}
		cfg.GeminiAPIKey = strings.TrimSpace(os.Getenv("GEMINI_API_KEY"))
		cfg.GeminiModel = strings.TrimSpace(os.Getenv("GEMINI_MODEL"))
	}
	model := strings.TrimSpace(modelOverride)
	if model == "" {
		model = cfg.GeminiModel
	}
	if model == "" {
		model = "gemini-1.5-flash"
	}
	result := diagnosticResult{
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		Provider:    "gemini",
		Model:       model,
		Status:      "skipped",
		Reason:      "GEMINI_API_KEY not set",
	}
	if strings.TrimSpace(cfg.GeminiAPIKey) == "" {
		return writeDiagnostic(filepath.Join(outputDir, "gemini_provider_diagnostic.md"), result)
	}

	client, err := geminiinfra.NewClient(cfg.GeminiAPIKey, model)
	if err != nil {
		result.Status = "failed"
		result.ErrorType = classifyError(err)
		result.Reason = sanitize(err.Error())
		return writeDiagnostic(filepath.Join(outputDir, "gemini_provider_diagnostic.md"), result)
	}

	started := time.Now()
	_, err = client.Generate(ctx, runtimedomain.ModelRequest{
		UserID:  "diagnostic",
		Message: "Return a minimal readiness response. Do not propose actions.",
	})
	result.LatencyMS = time.Since(started).Milliseconds()
	if err != nil {
		result.Status = "failed"
		result.ErrorType = classifyError(err)
		result.Reason = sanitize(err.Error())
		var geminiErr *geminiinfra.GeminiError
		if errors.As(err, &geminiErr) {
			result.ErrorBody = sanitize(geminiErr.BodyRedacted)
		}
		return writeDiagnostic(filepath.Join(outputDir, "gemini_provider_diagnostic.md"), result)
	}

	result.Status = "ok"
	result.Reason = "minimal Gemini call succeeded"
	return writeDiagnostic(filepath.Join(outputDir, "gemini_provider_diagnostic.md"), result)
}

func classifyError(err error) string {
	var geminiErr *geminiinfra.GeminiError
	if errors.As(err, &geminiErr) {
		return geminiErr.ErrorType
	}
	switch {
	case errors.Is(err, geminiinfra.ErrMissingAPIKey):
		return "missing_api_key"
	case errors.Is(err, geminiinfra.ErrMissingModel):
		return "missing_model"
	case errors.Is(err, geminiinfra.ErrInvalidGeminiOutput):
		return "invalid_structured_output"
	default:
		return "provider_error"
	}
}

func writeDiagnostic(path string, result diagnosticResult) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	var b strings.Builder
	b.WriteString("# Gemini Provider Diagnostic\n\n")
	b.WriteString(fmt.Sprintf("Generated at: `%s`\n\n", result.GeneratedAt))
	b.WriteString(fmt.Sprintf("- provider: `%s`\n", result.Provider))
	b.WriteString(fmt.Sprintf("- model: `%s`\n", result.Model))
	b.WriteString(fmt.Sprintf("- status: `%s`\n", result.Status))
	b.WriteString(fmt.Sprintf("- latency_ms: `%d`\n", result.LatencyMS))
	if result.ErrorType != "" {
		b.WriteString(fmt.Sprintf("- error_type: `%s`\n", result.ErrorType))
	}
	if result.Reason != "" {
		b.WriteString(fmt.Sprintf("- reason: `%s`\n", sanitize(result.Reason)))
	}
	if result.ErrorBody != "" {
		b.WriteString(fmt.Sprintf("- error_body_redacted: `%s`\n", sanitize(result.ErrorBody)))
	}
	b.WriteString("\nSecurity: API keys, prompts and raw provider payloads are not stored in this report.\n")
	return os.WriteFile(path, []byte(b.String()), 0o644)
}

func sanitize(value string) string {
	value = strings.ReplaceAll(strings.TrimSpace(value), "\n", " ")
	if len([]rune(value)) > 500 {
		return string([]rune(value)[:500]) + "...[TRUNCATED]"
	}
	return value
}
