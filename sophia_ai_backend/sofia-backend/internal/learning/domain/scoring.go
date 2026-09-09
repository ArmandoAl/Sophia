package domain

import "math"

const (
	DefaultAutonomyThreshold   = 0.85
	MinAutoExecuteSamples      = 10
	ApprovalHalfLifeDays       = 60.0
	PredictionModelHeuristicV1 = "heuristic_v1"
)

type ToolHistory struct {
	Total               int     `json:"total"`
	ApprovedDirect      int     `json:"approved_direct"`
	ApprovedCorrected   int     `json:"approved_corrected"`
	Rejected            int     `json:"rejected"`
	DaysSinceLastSample float64 `json:"days_since_last_sample"`
}

type ScoredBelief struct {
	ID          string
	Confidence  float64
	Contradicts bool
}

func PredictApproval(h ToolHistory, beliefs []ScoredBelief) (float64, []string) {
	total := float64(h.Total + 2)
	base := float64(h.ApprovedDirect+1) / total
	base += 0.5 * float64(h.ApprovedCorrected) / total

	used := make([]string, 0, len(beliefs))
	for _, belief := range beliefs {
		if belief.ID != "" {
			used = append(used, belief.ID)
		}
		if belief.Contradicts {
			base -= 0.25 * belief.Confidence
		} else {
			base += 0.10 * belief.Confidence
		}
	}

	days := h.DaysSinceLastSample
	if days < 0 {
		days = 0
	}
	decay := math.Pow(0.5, days/ApprovalHalfLifeDays)
	base = 0.5 + (base-0.5)*decay
	if base < 0 {
		base = 0
	}
	if base > 1 {
		base = 1
	}
	return base, used
}

func ShouldAutoExecute(p float64, h ToolHistory, reversible bool, threshold float64) bool {
	return p >= threshold && h.Total >= MinAutoExecuteSamples && reversible
}
