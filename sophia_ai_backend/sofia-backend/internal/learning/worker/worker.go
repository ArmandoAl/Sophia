package worker

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"sort"
	"strings"
	"sync"
	"time"

	actionsdomain "github.com/armandoalvarado/sofia-backend/internal/ai/actions/domain"
	runtimedomain "github.com/armandoalvarado/sofia-backend/internal/ai/runtime/domain"
	learningapp "github.com/armandoalvarado/sofia-backend/internal/learning/application"
	learningdomain "github.com/armandoalvarado/sofia-backend/internal/learning/domain"
	"github.com/armandoalvarado/sofia-backend/internal/platform/jsonschema"
	usersapp "github.com/armandoalvarado/sofia-backend/internal/users/application"
	usersdomain "github.com/armandoalvarado/sofia-backend/internal/users/domain"
)

const (
	defaultLeaseDuration  = 2 * time.Minute
	defaultWorkerID       = "synthesis-worker"
	defaultRunHourLocal   = 3
	approvalDropThreshold = 0.15
	quarantineWindowDays  = 7
	promptTokenCeiling    = learningdomain.PromptTokenCeiling
	synthesisListLimit    = 1000
)

var synthesisOutputSchema = json.RawMessage(`{
  "type": "object",
  "required": ["reinforced", "contradicted", "novel"],
  "additionalProperties": false,
  "properties": {
    "reinforced": {
      "type": "array",
      "items": {
        "type": "object",
        "required": ["belief_id", "evidence", "confidence_delta"],
        "additionalProperties": false,
        "properties": {
          "belief_id": {"type": "string"},
          "evidence": {"type": "string"},
          "confidence_delta": {"type": "number"}
        }
      }
    },
    "contradicted": {
      "type": "array",
      "items": {
        "type": "object",
        "required": ["belief_id", "evidence", "note"],
        "additionalProperties": false,
        "properties": {
          "belief_id": {"type": "string"},
          "evidence": {"type": "string"},
          "note": {"type": "string"}
        }
      }
    },
    "novel": {
      "type": "array",
      "items": {
        "type": "object",
        "required": ["statement", "category", "confidence"],
        "additionalProperties": false,
        "properties": {
          "statement": {"type": "string"},
          "category": {"type": "string"},
          "confidence": {"type": "number"},
          "scope": {"type": "string"},
          "scope_key": {"type": "string"}
        }
      }
    }
  }
}`)

type ProposalRepository interface {
	ListByDateRange(ctx context.Context, userID string, from, to time.Time) ([]*actionsdomain.AIActionProposal, error)
}

type UserIDLister interface {
	ListIDs() ([]string, error)
}

type Users interface {
	GetMe(userID string) (*usersapp.Me, error)
	UpdateAISettings(userID string, update usersdomain.AISettingsUpdate) (*usersdomain.AISettings, error)
}

type AuditRecorder interface {
	RecordAuditLog(ctx context.Context, userID, action, resourceType, resourceID string, metadata map[string]any) error
}

type Options struct {
	WorkerID      string
	LeaseDuration time.Duration
	RunHourLocal  int
	// IgnoreRunHour omite el gate de hora. Para ejecuciones disparadas por un
	// scheduler externo: la invocación ya es la señal de que es el momento.
	IgnoreRunHour bool
	Now           func() time.Time
	Logger        *log.Logger
}

type Deps struct {
	Proposals ProposalRepository
	Learning  *learningapp.Service
	Users     Users
	UserIDs   UserIDLister
	Model     runtimedomain.ModelClient
	Audit     AuditRecorder
}

type Worker struct {
	deps     Deps
	options  Options
	inFlight sync.Map
}

type RunResult struct {
	Scanned   int
	Processed int
	Skipped   int
	Failed    int
}

type synthesisOutput struct {
	Reinforced   []reinforcedBelief   `json:"reinforced"`
	Contradicted []contradictedBelief `json:"contradicted"`
	Novel        []novelBelief        `json:"novel"`
}

type reinforcedBelief struct {
	BeliefID        string  `json:"belief_id"`
	Evidence        string  `json:"evidence"`
	ConfidenceDelta float64 `json:"confidence_delta"`
}

type contradictedBelief struct {
	BeliefID string `json:"belief_id"`
	Evidence string `json:"evidence"`
	Note     string `json:"note"`
}

type novelBelief struct {
	Statement  string  `json:"statement"`
	Category   string  `json:"category"`
	Confidence float64 `json:"confidence"`
	Scope      string  `json:"scope"`
	ScopeKey   string  `json:"scope_key"`
}

type decisionSignals struct {
	Corrections     []map[string]any
	Rejections      []map[string]any
	ApprovalsByTool map[string]int
	Stats           learningdomain.DailyStats
}

func New(deps Deps, options Options) *Worker {
	if options.WorkerID == "" {
		options.WorkerID = defaultWorkerID
	}
	if options.LeaseDuration <= 0 {
		options.LeaseDuration = defaultLeaseDuration
	}
	if options.RunHourLocal < 0 || options.RunHourLocal > 23 {
		options.RunHourLocal = defaultRunHourLocal
	}
	if options.Now == nil {
		options.Now = time.Now
	}
	return &Worker{deps: deps, options: options}
}

func (w *Worker) RunOnce(ctx context.Context) (RunResult, error) {
	if w.deps.Learning != nil {
		if _, err := w.deps.Learning.ArchiveExpiredStates(ctx, w.options.Now()); err != nil {
			return RunResult{}, err
		}
		if _, err := w.deps.Learning.ArchiveStaleEpisodes(ctx, w.options.Now()); err != nil {
			return RunResult{}, err
		}
	}
	if w.deps.UserIDs == nil {
		return RunResult{}, nil
	}
	ids, err := w.deps.UserIDs.ListIDs()
	if err != nil {
		return RunResult{}, err
	}
	result := RunResult{Scanned: len(ids)}
	for _, userID := range ids {
		if err := ctx.Err(); err != nil {
			return result, err
		}
		loc := w.userLocation(userID)
		now := w.options.Now().In(loc)
		if !w.options.IgnoreRunHour && now.Hour() != w.options.RunHourLocal {
			result.Skipped++
			continue
		}
		if _, err := w.deps.Learning.PromoteEntityCandidates(ctx, userID, now); err != nil {
			w.logf("entity candidate maintenance failed user=%s err=%v", userID, err)
			result.Failed++
			continue
		}
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
		day := today.AddDate(0, 0, -1)
		if err := w.RunDaily(ctx, userID, day); err != nil {
			w.logf("run daily failed user=%s day=%s err=%v", userID, day.Format(learningdomain.DateLayout), err)
			result.Failed++
			continue
		}
		result.Processed++
	}
	return result, nil
}

func (w *Worker) RunDaily(ctx context.Context, userID string, day time.Time) error {
	if w.deps.Learning == nil || w.deps.Proposals == nil {
		return errors.New("synthesis worker is missing required dependencies")
	}
	loc := w.userLocation(userID)
	from, to, date := localDayWindow(day, loc)
	leaseKey := userID + "_" + date
	if !w.tryLease(leaseKey) {
		return nil
	}
	defer w.inFlight.Delete(leaseKey)

	existing, err := w.deps.Learning.FindDailySummary(ctx, userID, date)
	if err != nil && !errors.Is(err, learningdomain.ErrDailySummaryNotFound) {
		return err
	}
	if existing != nil {
		return nil
	}

	proposals, err := w.deps.Proposals.ListByDateRange(ctx, userID, from, to)
	if err != nil {
		return err
	}
	if len(proposals) == 0 {
		return nil
	}

	signals := extractSignals(proposals)
	beliefs, err := w.deps.Learning.ListActiveBeliefs(ctx, userID, synthesisListLimit)
	if err != nil {
		return err
	}
	contexts, err := w.deps.Learning.ListUserContexts(ctx, userID)
	if err != nil {
		return err
	}

	message, err := buildSynthesisMessage(beliefs, contexts, signals)
	if err != nil {
		return err
	}

	started := w.options.Now()
	response, err := w.deps.Model.Generate(ctx, runtimedomain.ModelRequest{
		UserID:  userID,
		Message: message,
		Task:    runtimedomain.TaskSynthesize,
	})
	if err != nil {
		return err
	}
	latencyMS := w.options.Now().Sub(started).Milliseconds()

	if err := jsonschema.Validate(synthesisOutputSchema, json.RawMessage(response.AssistantMessage)); err != nil {
		return err
	}
	var output synthesisOutput
	if err := json.Unmarshal([]byte(response.AssistantMessage), &output); err != nil {
		return err
	}

	summary, err := w.deps.Learning.CreateDailySummary(ctx, userID, learningdomain.DailySummaryCreate{
		Date:         date,
		Observations: observationsFrom(signals, output),
		Stats:        signals.Stats,
		SynthesisCost: learningdomain.SynthesisCost{
			InputTokens:  response.Usage.InputTokens,
			OutputTokens: response.Usage.OutputTokens,
			Model:        response.Usage.Model,
			LatencyMS:    latencyMS,
		},
	})
	if errors.Is(err, learningdomain.ErrDailySummaryAlreadyExists) {
		return nil
	}
	if err != nil {
		return err
	}

	delta, err := w.applyDelta(ctx, userID, output)
	if err != nil {
		return err
	}

	beforeID := ""
	if previous, err := w.deps.Learning.GetActivePromptVersion(ctx, userID); err != nil {
		return err
	} else if previous != nil {
		beforeID = previous.ID
	}

	version, err := w.deps.Learning.RecompilePromptBaseFromSummary(ctx, userID, summary.ID)
	if err != nil {
		return err
	}
	afterID := ""
	if version != nil {
		afterID = version.ID
		if version.TokenCount > promptTokenCeiling {
			w.recordAudit(ctx, userID, "learning_prompt_token_ceiling_exceeded", "prompt_version", version.ID, map[string]any{
				"summary_id":    summary.ID,
				"token_count":   version.TokenCount,
				"token_ceiling": promptTokenCeiling,
			})
		}
	}

	summary.DeltaVsPrevious = delta
	summary.PromptVersionBefore = beforeID
	summary.PromptVersionAfter = afterID
	if err := w.deps.Learning.UpdateDailySummary(ctx, summary); err != nil {
		return err
	}

	return w.enforceQuarantine(ctx, userID, to)
}

func (w *Worker) applyDelta(ctx context.Context, userID string, output synthesisOutput) (learningdomain.DailyDelta, error) {
	delta := learningdomain.DailyDelta{}
	contradicted := make([]*learningdomain.Belief, 0, len(output.Contradicted))
	for _, item := range output.Contradicted {
		beliefID := strings.TrimSpace(item.BeliefID)
		if beliefID == "" {
			continue
		}
		belief, err := w.deps.Learning.ContradictBelief(ctx, userID, beliefID)
		if err != nil {
			if errors.Is(err, learningdomain.ErrBeliefNotFound) {
				continue
			}
			return learningdomain.DailyDelta{}, err
		}
		delta.Contradicted = append(delta.Contradicted, belief.ID)
		contradicted = append(contradicted, belief)
	}

	for _, item := range output.Reinforced {
		beliefID := strings.TrimSpace(item.BeliefID)
		if beliefID == "" {
			continue
		}
		belief, err := w.deps.Learning.ReinforceBelief(ctx, userID, beliefID)
		if err != nil {
			if errors.Is(err, learningdomain.ErrBeliefNotFound) {
				continue
			}
			return learningdomain.DailyDelta{}, err
		}
		delta.Reinforced = append(delta.Reinforced, belief.ID)
	}

	novels := make([]*learningdomain.Belief, 0, len(output.Novel))
	for _, item := range output.Novel {
		if !learningdomain.IsValidCategory(item.Category) || strings.TrimSpace(item.Statement) == "" {
			continue
		}
		scope, scopeKey, err := w.validNovelScope(ctx, userID, item.Scope, item.ScopeKey)
		if err != nil {
			return learningdomain.DailyDelta{}, err
		}
		belief, err := w.deps.Learning.ApplyNovelBelief(ctx, userID, item.Statement, item.Category, item.Confidence, scope, scopeKey)
		if err != nil {
			return learningdomain.DailyDelta{}, err
		}
		delta.Novel = append(delta.Novel, belief.ID)
		novels = append(novels, belief)
	}

	for _, old := range contradicted {
		replacement := matchReplacement(old, novels)
		if replacement == nil || replacement.ID == old.ID {
			continue
		}
		if _, err := w.deps.Learning.SupersedeBelief(ctx, userID, old.ID, replacement.ID, "replaced by novel belief"); err != nil && !errors.Is(err, learningdomain.ErrBeliefNotActive) {
			return learningdomain.DailyDelta{}, err
		}
	}
	return delta, nil
}

func (w *Worker) validNovelScope(ctx context.Context, userID, scope, scopeKey string) (string, string, error) {
	scope = strings.ToLower(strings.TrimSpace(scope))
	scopeKey = strings.ToLower(strings.TrimSpace(scopeKey))
	if scope != learningdomain.ScopePerson && scope != learningdomain.ScopeMode {
		return learningdomain.ScopeGlobal, "", nil
	}
	contextValue, err := w.deps.Learning.FindUserContextByScopeKey(ctx, userID, scopeKey)
	if err != nil {
		return "", "", err
	}
	if contextValue == nil || contextValue.Kind != scope {
		return learningdomain.ScopeGlobal, "", nil
	}
	return scope, contextValue.ScopeKey(), nil
}

func matchReplacement(old *learningdomain.Belief, novels []*learningdomain.Belief) *learningdomain.Belief {
	for _, novel := range novels {
		if novel.Category != old.Category || novel.EffectiveScope() != old.EffectiveScope() || novel.ScopeKey != old.ScopeKey {
			continue
		}
		if learningdomain.HasHighTermOverlap(old.SearchTerms, novel.SearchTerms) {
			return novel
		}
	}
	return nil
}

func (w *Worker) enforceQuarantine(ctx context.Context, userID string, end time.Time) error {
	currentFrom := end.AddDate(0, 0, -quarantineWindowDays)
	previousFrom := end.AddDate(0, 0, -quarantineWindowDays*2)
	all, err := w.deps.Proposals.ListByDateRange(ctx, userID, previousFrom, end)
	if err != nil {
		return err
	}

	type windowStats struct {
		approved int
		decided  int
	}
	current := map[string]*windowStats{}
	previous := map[string]*windowStats{}
	for _, proposal := range all {
		approved, decided := decisionCounts(proposal)
		if decided == 0 {
			continue
		}
		bucket := previous
		if !proposal.CreatedAt.Before(currentFrom) {
			bucket = current
		}
		stats := bucket[proposal.ToolName]
		if stats == nil {
			stats = &windowStats{}
			bucket[proposal.ToolName] = stats
		}
		stats.approved += approved
		stats.decided += decided
	}

	dropped := false
	for toolName, nowStats := range current {
		beforeStats := previous[toolName]
		if beforeStats == nil || beforeStats.decided == 0 || nowStats.decided == 0 {
			continue
		}
		beforeRate := float64(beforeStats.approved) / float64(beforeStats.decided)
		nowRate := float64(nowStats.approved) / float64(nowStats.decided)
		if beforeRate-nowRate <= approvalDropThreshold {
			continue
		}
		dropped = true
		w.recordAudit(ctx, userID, "learning_tool_approval_drop", "tool", toolName, map[string]any{
			"tool_name":              toolName,
			"previous_approval_rate": beforeRate,
			"current_approval_rate":  nowRate,
			"drop":                   beforeRate - nowRate,
		})
	}
	if !dropped || w.deps.Users == nil {
		return nil
	}
	me, err := w.deps.Users.GetMe(userID)
	if err != nil || me == nil || me.AISettings == nil {
		return err
	}
	if me.AISettings.AutonomyLevel != usersdomain.AutonomySemiAutonomous {
		return nil
	}
	level := usersdomain.AutonomySuggestive
	_, err = w.deps.Users.UpdateAISettings(userID, usersdomain.AISettingsUpdate{AutonomyLevel: &level})
	return err
}

func decisionCounts(proposal *actionsdomain.AIActionProposal) (approved, decided int) {
	switch proposal.Feedback {
	case actionsdomain.FeedbackApprovedDirect, actionsdomain.FeedbackApprovedCorrected:
		return 1, 1
	case actionsdomain.FeedbackRejected:
		return 0, 1
	default:
		return 0, 0
	}
}

func extractSignals(proposals []*actionsdomain.AIActionProposal) decisionSignals {
	signals := decisionSignals{ApprovalsByTool: map[string]int{}}
	latencies := make([]int64, 0, len(proposals))
	for _, proposal := range proposals {
		signals.Stats.ProposalsTotal++
		if proposal.DecisionLatencyMS > 0 {
			latencies = append(latencies, proposal.DecisionLatencyMS)
		}
		switch proposal.Feedback {
		case actionsdomain.FeedbackApprovedDirect:
			signals.Stats.ApprovedDirect++
			signals.ApprovalsByTool[proposal.ToolName]++
		case actionsdomain.FeedbackApprovedCorrected:
			signals.Stats.ApprovedCorrected++
			signals.Corrections = append(signals.Corrections, map[string]any{
				"tool_name":        proposal.ToolName,
				"correction_delta": append([]string(nil), proposal.CorrectionDelta...),
			})
		case actionsdomain.FeedbackRejected:
			signals.Stats.Rejected++
			signals.Rejections = append(signals.Rejections, map[string]any{
				"tool_name": proposal.ToolName,
				"reason":    proposal.RejectionReason,
			})
		case actionsdomain.FeedbackIgnored, actionsdomain.FeedbackExpired:
			signals.Stats.Ignored++
		}
	}
	signals.Stats.MedianDecisionSeconds = medianSeconds(latencies)
	return signals
}

func buildSynthesisMessage(beliefs []*learningdomain.Belief, contexts []*learningdomain.UserContext, signals decisionSignals) (string, error) {
	prior := make([]map[string]any, 0, len(beliefs))
	for _, belief := range beliefs {
		prior = append(prior, map[string]any{
			"belief_id":           belief.ID,
			"statement":           belief.Statement,
			"category":            belief.Category,
			"confidence":          belief.DecayedConfidence(time.Now().UTC()),
			"evidence_count":      belief.EvidenceCount,
			"contradiction_count": belief.ContradictionCount,
			"prompt_slot":         belief.PromptSlot,
			"scope":               belief.EffectiveScope(),
			"scope_key":           belief.ScopeKey,
			"subject_type":        belief.EffectiveSubjectType(),
			"subject_id":          belief.SubjectID,
			"fact_kind":           belief.FactKind,
			"valid_until":         belief.ValidUntil,
			"follow_up_at":        belief.FollowUpAt,
		})
	}
	availableContexts := make([]map[string]string, 0, len(contexts))
	for _, contextValue := range contexts {
		if contextValue != nil && contextValue.CanLoadContext() {
			availableContexts = append(availableContexts, map[string]string{"entity_id": contextValue.ID, "scope_key": contextValue.ScopeKey(), "label": contextValue.Label})
		}
	}
	payload := map[string]any{
		"instructions": []string{
			"Synthesize belief updates from DECISIONS only.",
			"Do not read or infer from conversations, messages, or free-form chat.",
			"A day with chatter and no proposals is zero signal.",
			"Return JSON with reinforced, contradicted, and novel. No extra keys.",
			"Novel beliefs may use only a scope_key from available_contexts; otherwise use global scope with an empty scope_key.",
		},
		"active_beliefs":     prior,
		"available_contexts": availableContexts,
		"decisions": map[string]any{
			"corrections":       signals.Corrections,
			"rejections":        signals.Rejections,
			"approvals_by_tool": signals.ApprovalsByTool,
		},
	}
	if payload["decisions"].(map[string]any)["corrections"] == nil {
		payload["decisions"].(map[string]any)["corrections"] = []map[string]any{}
	}
	if payload["decisions"].(map[string]any)["rejections"] == nil {
		payload["decisions"].(map[string]any)["rejections"] = []map[string]any{}
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func observationsFrom(signals decisionSignals, output synthesisOutput) []string {
	obs := make([]string, 0, 8)
	if signals.Stats.ApprovedCorrected > 0 {
		obs = append(obs, "corrections present")
	}
	if signals.Stats.Rejected > 0 {
		obs = append(obs, "rejections present")
	}
	if signals.Stats.ApprovedDirect > 0 {
		obs = append(obs, "direct approvals present")
	}
	if len(output.Novel) > 0 {
		obs = append(obs, "novel beliefs proposed")
	}
	if len(output.Contradicted) > 0 {
		obs = append(obs, "beliefs contradicted")
	}
	if len(output.Reinforced) > 0 {
		obs = append(obs, "beliefs reinforced")
	}
	return obs
}

func localDayWindow(day time.Time, loc *time.Location) (from, to time.Time, date string) {
	if loc == nil {
		loc = time.UTC
	}
	local := day.In(loc)
	from = time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, loc)
	to = from.AddDate(0, 0, 1)
	return from, to, from.Format(learningdomain.DateLayout)
}

func (w *Worker) userLocation(userID string) *time.Location {
	if w.deps.Users == nil {
		return time.UTC
	}
	me, err := w.deps.Users.GetMe(userID)
	if err != nil || me == nil || me.Profile == nil {
		return time.UTC
	}
	loc, err := time.LoadLocation(strings.TrimSpace(me.Profile.Timezone))
	if err != nil || loc == nil {
		return time.UTC
	}
	return loc
}

func (w *Worker) tryLease(key string) bool {
	now := w.options.Now()
	until := now.Add(w.options.LeaseDuration)
	if existing, ok := w.inFlight.Load(key); ok {
		if expiry, ok := existing.(time.Time); ok && expiry.After(now) {
			return false
		}
	}
	w.inFlight.Store(key, until)
	return true
}

func (w *Worker) recordAudit(ctx context.Context, userID, action, resourceType, resourceID string, metadata map[string]any) {
	if w.deps.Audit == nil {
		return
	}
	_ = w.deps.Audit.RecordAuditLog(ctx, userID, action, resourceType, resourceID, metadata)
}

func (w *Worker) logf(format string, args ...any) {
	if w.options.Logger == nil {
		return
	}
	w.options.Logger.Printf(format, args...)
}

func medianSeconds(values []int64) float64 {
	if len(values) == 0 {
		return 0
	}
	sorted := append([]int64(nil), values...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
	mid := len(sorted) / 2
	if len(sorted)%2 == 1 {
		return float64(sorted[mid]) / 1000
	}
	return float64(sorted[mid-1]+sorted[mid]) / 2 / 1000
}
