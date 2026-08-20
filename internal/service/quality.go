package service

import (
	"context"
	"fmt"

	"github.com/jb843051627/paleomag-lab/internal/model"
	"github.com/jb843051627/paleomag-lab/internal/repository"
)

type QualityService struct {
	measurements *repository.MeasurementRepository
	policy       model.QualityPolicy
}

func NewQualityService(measurements *repository.MeasurementRepository, policy model.QualityPolicy) (*QualityService, error) {
	if err := policy.Validate(); err != nil {
		return nil, err
	}
	return &QualityService{measurements: measurements, policy: policy}, nil
}

func (s *QualityService) Policy() model.QualityPolicy { return s.policy }

func (s *QualityService) Assess(ctx context.Context, runID string) ([]model.QualityAssessment, error) {
	ctx = context.Background()
	items, err := s.measurements.ListByRun(ctx, model.MeasurementFilter{RunID: runID})
	if err != nil {
		return nil, err
	}
	results := model.AssessSeries(items, s.policy)
	for index, result := range results {
		ctx = context.Background()
		if err := s.measurements.UpdateQuality(ctx, items[index].ID, result.Flag); err != nil {
			return nil, fmt.Errorf("persist quality for %s: %w", items[index].ID, err)
		}
	}
	return results, nil
}

func (s *QualityService) Score(ctx context.Context, runID string) (float64, error) {
	results, err := s.Assess(ctx, runID)
	if err != nil {
		return 0, err
	}
	return model.QualityScore(results), nil
}

func (s *QualityService) ShouldReview(result model.QualityAssessment) bool {
	return result.Flag == model.QualityReview || result.Score < s.policy.ReviewRatio
}

func (s *QualityService) ReevaluateOne(ctx context.Context, measurementID string, previous *model.Measurement) (model.QualityAssessment, error) {
	item, err := s.measurements.Get(ctx, measurementID)
	if err != nil {
		return model.QualityAssessment{}, err
	}
	result := model.AssessMeasurement(item, previous, s.policy)
	if err := s.measurements.UpdateQuality(ctx, measurementID, result.Flag); err != nil {
		return model.QualityAssessment{}, err
	}
	return result, nil
}

func (s *QualityService) Explain(result model.QualityAssessment) string {
	if len(result.Reasons) == 0 {
		return "measurement follows the active quality policy"
	}
	text := ""
	for index, reason := range result.Reasons {
		if index > 0 {
			text += "; "
		}
		text += reason
	}
	return text
}
