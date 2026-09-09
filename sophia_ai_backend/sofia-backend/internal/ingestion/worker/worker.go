package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	runtimedomain "github.com/armandoalvarado/sofia-backend/internal/ai/runtime/domain"
	ingestiondomain "github.com/armandoalvarado/sofia-backend/internal/ingestion/domain"
	learningapp "github.com/armandoalvarado/sofia-backend/internal/learning/application"
	learningdomain "github.com/armandoalvarado/sofia-backend/internal/learning/domain"
	"github.com/armandoalvarado/sofia-backend/internal/platform/jsonschema"
)

const WindowSize = 50

var outputSchema = json.RawMessage(`{
  "type":"object",
  "required":["beliefs"],
  "additionalProperties":false,
  "properties":{"beliefs":{"type":"array","items":{
    "type":"object",
    "required":["statement","category","scope","scope_key","confidence"],
    "additionalProperties":false,
    "properties":{
      "statement":{"type":"string"},
      "category":{"type":"string","enum":["schedule","communication","priorities","work_style","personal","constraint"]},
      "scope":{"type":"string","enum":["global","person","mode"]},
      "scope_key":{"type":"string"},
      "confidence":{"type":"number"}
    }
  }}}
}`)

type Worker struct {
	batches   ingestiondomain.BatchRepository
	learning  *learningapp.Service
	model     runtimedomain.ModelClient
	workerID  string
	maxTokens int
	now       func() time.Time
}

type Options struct {
	WorkerID  string
	MaxTokens int
	Now       func() time.Time
}

type extractedBelief struct {
	Statement  string  `json:"statement"`
	Category   string  `json:"category"`
	Scope      string  `json:"scope"`
	ScopeKey   string  `json:"scope_key"`
	Confidence float64 `json:"confidence"`
}

type extractionOutput struct {
	Beliefs []extractedBelief `json:"beliefs"`
}

type windowMessage struct {
	Source       string                     `json:"source"`
	ExternalID   string                     `json:"external_id"`
	Participants []string                   `json:"participants"`
	Message      ingestiondomain.RawMessage `json:"message"`
}

func New(batches ingestiondomain.BatchRepository, learning *learningapp.Service, model runtimedomain.ModelClient, options Options) *Worker {
	if options.WorkerID == "" {
		options.WorkerID = "ingestion-worker"
	}
	if options.Now == nil {
		options.Now = time.Now
	}
	return &Worker{batches: batches, learning: learning, model: model, workerID: options.WorkerID, maxTokens: options.MaxTokens, now: options.Now}
}

func (w *Worker) RunOnce(ctx context.Context) (*ingestiondomain.Batch, error) {
	now := w.now().UTC()
	batch, err := w.batches.ClaimNext(ctx, w.workerID, now, now.Add(2*time.Minute))
	if err != nil || batch == nil {
		return batch, err
	}
	windows := splitMessages(batch.Conversations)
	for i := batch.WindowsProcessed; i < len(windows); i++ {
		message, err := json.Marshal(map[string]any{"instructions": "Return only observable facts and preferences about the importing user.", "output_schema": outputSchema, "messages": windows[i]})
		if err != nil {
			return batch, w.fail(ctx, batch, err)
		}
		response, err := w.model.Generate(ctx, runtimedomain.ModelRequest{UserID: batch.UserID, Message: string(message), Task: runtimedomain.TaskExtract})
		if err != nil {
			return batch, w.fail(ctx, batch, err)
		}
		batch.InputTokens += response.Usage.InputTokens
		batch.OutputTokens += response.Usage.OutputTokens
		if w.maxTokens > 0 && batch.InputTokens+batch.OutputTokens > w.maxTokens {
			return batch, w.fail(ctx, batch, fmt.Errorf("ingestion token budget exceeded: %d > %d", batch.InputTokens+batch.OutputTokens, w.maxTokens))
		}
		if err := jsonschema.Validate(outputSchema, json.RawMessage(response.AssistantMessage)); err != nil {
			return batch, w.fail(ctx, batch, err)
		}
		var output extractionOutput
		if err := json.Unmarshal([]byte(response.AssistantMessage), &output); err != nil {
			return batch, w.fail(ctx, batch, err)
		}
		for _, value := range output.Beliefs {
			scope, scopeKey := normalizeScope(ctx, w.learning, batch.UserID, value.Scope, value.ScopeKey)
			belief, err := w.learning.UpsertBeliefWithTrust(ctx, batch.UserID, value.Statement, value.Category, scope, scopeKey, learningdomain.TrustInferred, value.Confidence, batch.ID)
			if err != nil {
				return batch, w.fail(ctx, batch, err)
			}
			if belief.BatchID == batch.ID && belief.EvidenceCount == 1 {
				batch.BeliefsCreated++
			}
		}
		batch.WindowsProcessed = i + 1
		batch.UpdatedAt = w.now().UTC()
		batch.ProcessingUntil = batch.UpdatedAt.Add(2 * time.Minute)
		if err := w.batches.Update(ctx, batch); err != nil {
			return batch, err
		}
	}
	batch.Status, batch.ProcessingBy, batch.ProcessingUntil, batch.UpdatedAt = ingestiondomain.StatusDone, "", time.Time{}, w.now().UTC()
	return batch, w.batches.Update(ctx, batch)
}

func (w *Worker) fail(ctx context.Context, batch *ingestiondomain.Batch, cause error) error {
	batch.Status, batch.FailureReason, batch.ProcessingBy, batch.ProcessingUntil, batch.UpdatedAt = ingestiondomain.StatusFailed, cause.Error(), "", time.Time{}, w.now().UTC()
	if err := w.batches.Update(ctx, batch); err != nil {
		return errors.Join(cause, err)
	}
	return cause
}

func splitMessages(conversations []ingestiondomain.RawConversation) [][]windowMessage {
	all := make([]windowMessage, 0)
	for _, conversation := range conversations {
		for _, message := range conversation.Messages {
			all = append(all, windowMessage{Source: conversation.Source, ExternalID: conversation.ExternalID, Participants: conversation.Participants, Message: message})
		}
	}
	windows := make([][]windowMessage, 0, (len(all)+WindowSize-1)/WindowSize)
	for len(all) > 0 {
		size := min(WindowSize, len(all))
		windows = append(windows, all[:size])
		all = all[size:]
	}
	return windows
}

func normalizeScope(ctx context.Context, learning *learningapp.Service, userID, scope, scopeKey string) (string, string) {
	scope = strings.ToLower(strings.TrimSpace(scope))
	scopeKey = strings.ToLower(strings.TrimSpace(scopeKey))
	if scope == "" || scope == learningdomain.ScopeGlobal {
		return learningdomain.ScopeGlobal, ""
	}
	if scope != learningdomain.ScopePerson && scope != learningdomain.ScopeMode {
		return learningdomain.ScopeGlobal, ""
	}
	value, err := learning.FindUserContextByScopeKey(ctx, userID, scopeKey)
	if err != nil || value == nil || value.Kind != scope {
		return learningdomain.ScopeGlobal, ""
	}
	return scope, scopeKey
}
