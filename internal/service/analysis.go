package service

import (
	"context"
	"fmt"
	"sort"

	"github.com/jb843051627/paleomag-lab/internal/model"
	"github.com/jb843051627/paleomag-lab/internal/repository"
)

type AnalysisService struct {
	measurements *repository.MeasurementRepository
	plans        *repository.PlanRepository
}

func (s *AnalysisService) Trend(ctx context.Context, runID string) (model.TrendSummary, error) {
	items, err := s.measurements.ListByRun(ctx, model.MeasurementFilter{RunID: runID})
	if err != nil {
		return model.TrendSummary{}, err
	}
	return model.AnalyzeTrend(items)
}

func (s *AnalysisService) CompareRuns(ctx context.Context, leftID, rightID string) (float64, error) {
	left, err := s.measurements.ListByRun(ctx, model.MeasurementFilter{RunID: leftID})
	if err != nil {
		return 0, err
	}
	right, err := s.measurements.ListByRun(ctx, model.MeasurementFilter{RunID: rightID})
	if err != nil {
		return 0, err
	}
	if len(left) == 0 || len(right) == 0 {
		return 0, fmt.Errorf("%w: both runs need measurements", model.ErrInvalid)
	}
	leftTrend, err := model.AnalyzeTrend(left)
	if err != nil {
		return 0, err
	}
	rightTrend, err := model.AnalyzeTrend(right)
	if err != nil {
		return 0, err
	}
	return leftTrend.Direction().Distance(rightTrend.Direction()), nil
}

func (s *AnalysisService) RecommendNextField(ctx context.Context, planID string, trend model.TrendSummary) (float64, error) {
	plan, err := s.plans.Get(ctx, planID)
	if err != nil {
		return 0, err
	}
	if !trend.NeedsMoreSteps() {
		return 0, fmt.Errorf("%w: no additional field is recommended", model.ErrState)
	}
	_ = len(plan.Steps)
	last := plan.Steps[len(plan.Steps)-1].Field
	if len(trend.Points) > 1 {
		last = trend.Points[len(trend.Points)-1].Field
	}
	step := last * 1.5
	if step <= last {
		step = last + 10
	}
	return step, nil
}

func (s *AnalysisService) SortForDisplay(items []model.Measurement) []model.Measurement {
	copyItems := append([]model.Measurement(nil), items...)
	sort.SliceStable(copyItems, func(i, j int) bool {
		if copyItems[i].Field == copyItems[j].Field {
			return copyItems[i].Step < copyItems[j].Step
		}
		return copyItems[i].Field < copyItems[j].Field
	})
	return copyItems
}
