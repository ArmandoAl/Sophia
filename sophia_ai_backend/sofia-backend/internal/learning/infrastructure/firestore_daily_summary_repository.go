package infrastructure

import (
	"context"
	"errors"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/armandoalvarado/sofia-backend/internal/learning/domain"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ domain.DailySummaryRepository = (*FirestoreDailySummaryRepository)(nil)

const dailySummariesCollection = "daily_summaries"

type FirestoreDailySummaryRepository struct {
	client  *firestore.Client
	timeout time.Duration
}

type firestoreDailySummary struct {
	ID                  string                 `firestore:"id"`
	UserID              string                 `firestore:"user_id"`
	Date                string                 `firestore:"date"`
	GeneratedAt         time.Time              `firestore:"generated_at"`
	Observations        []string               `firestore:"observations"`
	DeltaVsPrevious     firestoreDailyDelta    `firestore:"delta_vs_previous"`
	Stats               firestoreDailyStats    `firestore:"stats"`
	SynthesisCost       firestoreSynthesisCost `firestore:"synthesis_cost"`
	PromptVersionBefore string                 `firestore:"prompt_version_before"`
	PromptVersionAfter  string                 `firestore:"prompt_version_after"`
}

type firestoreDailyDelta struct {
	Reinforced   []string `firestore:"reinforced"`
	Contradicted []string `firestore:"contradicted"`
	Novel        []string `firestore:"novel"`
}

type firestoreDailyStats struct {
	ProposalsTotal        int     `firestore:"proposals_total"`
	ApprovedDirect        int     `firestore:"approved_direct"`
	ApprovedCorrected     int     `firestore:"approved_corrected"`
	Rejected              int     `firestore:"rejected"`
	Ignored               int     `firestore:"ignored"`
	MedianDecisionSeconds float64 `firestore:"median_decision_seconds"`
}

type firestoreSynthesisCost struct {
	InputTokens  int    `firestore:"input_tokens"`
	OutputTokens int    `firestore:"output_tokens"`
	Model        string `firestore:"model"`
	LatencyMS    int64  `firestore:"latency_ms"`
}

func NewFirestoreDailySummaryRepository(client *firestore.Client) *FirestoreDailySummaryRepository {
	return &FirestoreDailySummaryRepository{client: client, timeout: 5 * time.Second}
}

func (r *FirestoreDailySummaryRepository) Create(ctx context.Context, summary *domain.DailySummary) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	_, err := r.client.Collection(dailySummariesCollection).Doc(summary.ID).Create(ctx, dailySummaryToDocument(summary))
	if status.Code(err) == codes.AlreadyExists {
		return domain.ErrDailySummaryAlreadyExists
	}
	return err
}

func (r *FirestoreDailySummaryRepository) Update(ctx context.Context, summary *domain.DailySummary) error {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	doc := dailySummaryToDocument(summary)
	_, err := r.client.Collection(dailySummariesCollection).Doc(summary.ID).Update(ctx, []firestore.Update{
		{Path: "delta_vs_previous", Value: doc.DeltaVsPrevious},
		{Path: "prompt_version_before", Value: doc.PromptVersionBefore},
		{Path: "prompt_version_after", Value: doc.PromptVersionAfter},
	})
	if status.Code(err) == codes.NotFound {
		return domain.ErrDailySummaryNotFound
	}
	return err
}

func (r *FirestoreDailySummaryRepository) FindByDate(ctx context.Context, userID, date string) (*domain.DailySummary, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	doc, err := r.client.Collection(dailySummariesCollection).Doc(domain.DailySummaryID(userID, date)).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return r.findByUserDate(ctx, userID, date)
		}
		return nil, err
	}
	summary, err := documentToDailySummary(doc)
	if err != nil {
		return nil, err
	}
	if summary.UserID != userID {
		return nil, nil
	}
	return summary, nil
}

func (r *FirestoreDailySummaryRepository) findByUserDate(ctx context.Context, userID, date string) (*domain.DailySummary, error) {
	iter := r.client.Collection(dailySummariesCollection).
		Where("user_id", "==", userID).
		Where("date", "==", date).
		Limit(1).
		Documents(ctx)
	defer iter.Stop()
	doc, err := iter.Next()
	if errors.Is(err, iterator.Done) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return documentToDailySummary(doc)
}

func (r *FirestoreDailySummaryRepository) ListRecent(ctx context.Context, userID string, limit int) ([]*domain.DailySummary, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	query := r.client.Collection(dailySummariesCollection).
		Where("user_id", "==", userID).
		OrderBy("date", firestore.Desc)
	if limit > 0 {
		query = query.Limit(limit)
	}
	iter := query.Documents(ctx)
	defer iter.Stop()
	result := make([]*domain.DailySummary, 0)
	for {
		doc, err := iter.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, err
		}
		summary, err := documentToDailySummary(doc)
		if err != nil {
			return nil, err
		}
		result = append(result, summary)
	}
	return result, nil
}

func dailySummaryToDocument(summary *domain.DailySummary) firestoreDailySummary {
	return firestoreDailySummary{
		ID:           summary.ID,
		UserID:       summary.UserID,
		Date:         summary.Date,
		GeneratedAt:  summary.GeneratedAt,
		Observations: cloneStrings(summary.Observations),
		DeltaVsPrevious: firestoreDailyDelta{
			Reinforced:   cloneStrings(summary.DeltaVsPrevious.Reinforced),
			Contradicted: cloneStrings(summary.DeltaVsPrevious.Contradicted),
			Novel:        cloneStrings(summary.DeltaVsPrevious.Novel),
		},
		Stats: firestoreDailyStats{
			ProposalsTotal:        summary.Stats.ProposalsTotal,
			ApprovedDirect:        summary.Stats.ApprovedDirect,
			ApprovedCorrected:     summary.Stats.ApprovedCorrected,
			Rejected:              summary.Stats.Rejected,
			Ignored:               summary.Stats.Ignored,
			MedianDecisionSeconds: summary.Stats.MedianDecisionSeconds,
		},
		SynthesisCost: firestoreSynthesisCost{
			InputTokens:  summary.SynthesisCost.InputTokens,
			OutputTokens: summary.SynthesisCost.OutputTokens,
			Model:        summary.SynthesisCost.Model,
			LatencyMS:    summary.SynthesisCost.LatencyMS,
		},
		PromptVersionBefore: summary.PromptVersionBefore,
		PromptVersionAfter:  summary.PromptVersionAfter,
	}
}

func documentToDailySummary(doc *firestore.DocumentSnapshot) (*domain.DailySummary, error) {
	var stored firestoreDailySummary
	if err := doc.DataTo(&stored); err != nil {
		return nil, err
	}
	return stored.toDomain(), nil
}

func (m firestoreDailySummary) toDomain() *domain.DailySummary {
	return &domain.DailySummary{
		ID:           m.ID,
		UserID:       m.UserID,
		Date:         m.Date,
		GeneratedAt:  m.GeneratedAt,
		Observations: cloneStrings(m.Observations),
		DeltaVsPrevious: domain.DailyDelta{
			Reinforced:   cloneStrings(m.DeltaVsPrevious.Reinforced),
			Contradicted: cloneStrings(m.DeltaVsPrevious.Contradicted),
			Novel:        cloneStrings(m.DeltaVsPrevious.Novel),
		},
		Stats: domain.DailyStats{
			ProposalsTotal:        m.Stats.ProposalsTotal,
			ApprovedDirect:        m.Stats.ApprovedDirect,
			ApprovedCorrected:     m.Stats.ApprovedCorrected,
			Rejected:              m.Stats.Rejected,
			Ignored:               m.Stats.Ignored,
			MedianDecisionSeconds: m.Stats.MedianDecisionSeconds,
		},
		SynthesisCost: domain.SynthesisCost{
			InputTokens:  m.SynthesisCost.InputTokens,
			OutputTokens: m.SynthesisCost.OutputTokens,
			Model:        m.SynthesisCost.Model,
			LatencyMS:    m.SynthesisCost.LatencyMS,
		},
		PromptVersionBefore: m.PromptVersionBefore,
		PromptVersionAfter:  m.PromptVersionAfter,
	}
}
