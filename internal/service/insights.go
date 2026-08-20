package service

import (
	"context"
	"fmt"
	"sort"

	"github.com/jb843051627/paleomag-lab/internal/model"
	"github.com/jb843051627/paleomag-lab/internal/repository"
)

type InsightsService struct {
	queries   *repository.QueryRepository
	aggregate *repository.AggregateRepository
}

type Dashboard struct {
	Specimens              repository.SpecimenCounts `json:"specimens"`
	PendingInterpretations []model.Interpretation    `json:"pending_interpretations"`
	AcceptedSpecimens      int                       `json:"accepted_specimens"`
	GeneratedAt            string                    `json:"generated_at"`
}

func (s *InsightsService) Dashboard(ctx context.Context, limit int, now string) (Dashboard, error) {
	counts, err := s.queries.CountSpecimens(ctx)
	if err != nil {
		return Dashboard{}, err
	}
	pending, err := s.queries.PendingInterpretations(ctx, limit)
	if err != nil {
		return Dashboard{}, err
	}
	accepted, err := s.aggregate.AcceptedSpecimenCount(ctx)
	if err != nil {
		return Dashboard{}, err
	}
	return Dashboard{Specimens: counts, PendingInterpretations: pending, AcceptedSpecimens: accepted, GeneratedAt: now}, nil
}

func (s *InsightsService) Search(ctx context.Context, phrase string, limit int) ([]model.Specimen, error) {
	return s.queries.SearchByText(ctx, phrase, limit)
}

func (s *InsightsService) Quality(ctx context.Context, runID string) (repository.QualityDistribution, error) {
	return s.aggregate.QualityDistribution(ctx, runID)
}

func (s *InsightsService) Histogram(ctx context.Context, runID string, width float64) (map[int]int, error) {
	return s.aggregate.FieldHistogram(ctx, runID, width)
}

func (s *InsightsService) AverageIntensity(ctx context.Context, runID string) (float64, error) {
	return s.aggregate.AverageIntensity(ctx, runID)
}

func (s *InsightsService) LatestAcceptedRun(ctx context.Context, specimenID string) (model.MeasurementRun, error) {
	return s.queries.LatestAcceptedRun(ctx, specimenID)
}

type TimelineEntry struct {
	EntityID string `json:"entity_id"`
	Action   string `json:"action"`
	Actor    string `json:"actor"`
	At       string `json:"at"`
}

func (s *InsightsService) SortTimeline(entries []TimelineEntry) []TimelineEntry {
	copyEntries := entries
	sort.SliceStable(copyEntries, func(i, j int) bool { return copyEntries[i].At < copyEntries[j].At })
	return copyEntries
}

func (s *InsightsService) RequireRun(ctx context.Context, specimenID string) (model.MeasurementRun, error) {
	run, err := s.LatestAcceptedRun(ctx, specimenID)
	if err != nil {
		return model.MeasurementRun{}, fmt.Errorf("accepted run for %s: %w", specimenID, err)
	}
	if run.SpecimenID != specimenID {
		return model.MeasurementRun{}, fmt.Errorf("%w: run does not belong to specimen", model.ErrConflict)
	}
	return run, nil
}
