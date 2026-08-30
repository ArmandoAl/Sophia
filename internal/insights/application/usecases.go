package application

import (
	"context"

	activitiesdomain "github.com/armandoalvarado/sofia-backend/internal/activities/domain"
	"github.com/armandoalvarado/sofia-backend/internal/insights/domain"
	"github.com/google/uuid"
)

const DefaultListLimit = 50

type Service struct {
	moods       domain.MoodEntryRepository
	outcomes    domain.ActivityOutcomeRepository
	reflections domain.ReflectionRepository
	activities  activitiesdomain.ActivityRepository
}

func NewService(
	moods domain.MoodEntryRepository,
	outcomes domain.ActivityOutcomeRepository,
	reflections domain.ReflectionRepository,
	activities activitiesdomain.ActivityRepository,
) *Service {
	return &Service{
		moods:       moods,
		outcomes:    outcomes,
		reflections: reflections,
		activities:  activities,
	}
}

func (s *Service) CreateMoodEntry(ctx context.Context, userID string, input domain.MoodEntryCreate) (*domain.MoodEntry, error) {
	if err := s.ensureActivityBelongsToUser(ctx, userID, input.ActivityID); err != nil {
		return nil, err
	}
	entry, err := domain.NewMoodEntry(uuid.New().String(), userID, input)
	if err != nil {
		return nil, err
	}
	if err := s.moods.Create(ctx, entry); err != nil {
		return nil, err
	}
	return entry, nil
}

func (s *Service) ListMoodEntries(ctx context.Context, filter domain.ListFilter) ([]*domain.MoodEntry, error) {
	filter = normalizeFilter(filter)
	if err := domain.ValidateFilter(filter); err != nil {
		return nil, err
	}
	return s.moods.List(ctx, filter)
}

func (s *Service) CreateActivityOutcome(ctx context.Context, userID string, input domain.ActivityOutcomeCreate) (*domain.ActivityOutcome, error) {
	if err := s.ensureActivityBelongsToUser(ctx, userID, input.ActivityID); err != nil {
		return nil, err
	}
	outcome, err := domain.NewActivityOutcome(uuid.New().String(), userID, input)
	if err != nil {
		return nil, err
	}
	if err := s.outcomes.Create(ctx, outcome); err != nil {
		return nil, err
	}
	return outcome, nil
}

func (s *Service) ListActivityOutcomes(ctx context.Context, filter domain.ListFilter) ([]*domain.ActivityOutcome, error) {
	filter = normalizeFilter(filter)
	if err := domain.ValidateFilter(filter); err != nil {
		return nil, err
	}
	return s.outcomes.List(ctx, filter)
}

func (s *Service) CreateReflection(ctx context.Context, userID string, input domain.ReflectionCreate) (*domain.Reflection, error) {
	if err := s.ensureActivityBelongsToUser(ctx, userID, input.ActivityID); err != nil {
		return nil, err
	}
	reflection, err := domain.NewReflection(uuid.New().String(), userID, input)
	if err != nil {
		return nil, err
	}
	if err := s.reflections.Create(ctx, reflection); err != nil {
		return nil, err
	}
	return reflection, nil
}

func (s *Service) ListReflections(ctx context.Context, filter domain.ListFilter) ([]*domain.Reflection, error) {
	filter = normalizeFilter(filter)
	if err := domain.ValidateFilter(filter); err != nil {
		return nil, err
	}
	return s.reflections.List(ctx, filter)
}

func (s *Service) GetInsightsSummary(ctx context.Context, filter domain.ListFilter) (*domain.Summary, error) {
	filter = normalizeFilter(filter)
	if err := domain.ValidateFilter(filter); err != nil {
		return nil, err
	}

	moods, err := s.moods.List(ctx, filter)
	if err != nil {
		return nil, err
	}
	outcomes, err := s.outcomes.List(ctx, filter)
	if err != nil {
		return nil, err
	}
	reflections, err := s.reflections.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	summary := &domain.Summary{
		TotalOutcomes:    len(outcomes),
		TotalReflections: len(reflections),
	}
	if len(moods) == 0 {
		return summary, nil
	}

	var energy, stress, focus int
	for _, mood := range moods {
		energy += mood.EnergyLevel
		stress += mood.StressLevel
		focus += mood.FocusLevel
	}
	count := float64(len(moods))
	summary.AverageEnergy = float64(energy) / count
	summary.AverageStress = float64(stress) / count
	summary.AverageFocus = float64(focus) / count
	return summary, nil
}

func (s *Service) ensureActivityBelongsToUser(ctx context.Context, userID, activityID string) error {
	if activityID == "" || s.activities == nil {
		return nil
	}
	activity, err := s.activities.FindByID(ctx, userID, activityID)
	if err != nil {
		return err
	}
	if activity == nil {
		return domain.ErrActivityNotAllowed
	}
	return nil
}

func normalizeFilter(filter domain.ListFilter) domain.ListFilter {
	if filter.Limit == 0 {
		filter.Limit = DefaultListLimit
	}
	return filter
}
