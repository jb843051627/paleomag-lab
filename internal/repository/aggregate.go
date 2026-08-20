package repository

import (
	"context"
	"fmt"

	"github.com/jb843051627/paleomag-lab/internal/model"
)

type AggregateRepository struct{ db DBLike }

func NewAggregateRepository(db DBLike) *AggregateRepository { return &AggregateRepository{db: db} }

type QualityDistribution struct {
	Good    int `json:"good"`
	Review  int `json:"review"`
	Bad     int `json:"bad"`
	Pending int `json:"pending"`
}

func (r *AggregateRepository) QualityDistribution(ctx context.Context, runID string) (QualityDistribution, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT quality, COUNT(*) FROM measurements WHERE run_id=? GROUP BY quality`, runID)
	if err != nil {
		return QualityDistribution{}, fmt.Errorf("quality distribution: %w", err)
	}
	defer rows.Close()
	var result QualityDistribution
	for rows.Next() {
		var quality string
		var count int
		if err := rows.Scan(&quality, &count); err != nil {
			return QualityDistribution{}, fmt.Errorf("scan quality distribution: %w", err)
		}
		switch model.QualityFlag(quality) {
		case model.QualityGood:
			result.Good = count
		case model.QualityReview:
			result.Review = count
		case model.QualityBad:
			result.Bad = count
		case model.QualityPending:
			result.Pending = count
		}
	}
	return result, rows.Err()
}

func (r *AggregateRepository) FieldHistogram(ctx context.Context, runID string, width float64) (map[int]int, error) {
	if width <= 0 {
		return nil, fmt.Errorf("%w: histogram width must be positive", model.ErrInvalid)
	}
	rows, err := r.db.QueryContext(ctx, `SELECT field FROM measurements WHERE run_id=? ORDER BY field`, runID)
	if err != nil {
		return nil, fmt.Errorf("field histogram: %w", err)
	}
	defer rows.Close()
	result := make(map[int]int)
	for rows.Next() {
		var field float64
		if err := rows.Scan(&field); err != nil {
			return nil, fmt.Errorf("scan field histogram: %w", err)
		}
		bucket := int(field / width)
		result[bucket]++
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate field histogram: %w", err)
	}
	return result, nil
}

func (r *AggregateRepository) AverageIntensity(ctx context.Context, runID string) (float64, error) {
	row := r.db.QueryRowContext(ctx, `SELECT COALESCE(AVG(intensity), 0) FROM measurements WHERE run_id=?`, runID)
	var value float64
	if err := row.Scan(&value); err != nil {
		return 0, fmt.Errorf("average intensity: %w", err)
	}
	return value, nil
}

func (r *AggregateRepository) AcceptedSpecimenCount(ctx context.Context) (int, error) {
	row := r.db.QueryRowContext(ctx, `SELECT COUNT(DISTINCT specimen_id) FROM measurement_runs WHERE status=?`, model.RunAccepted)
	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("accepted specimen count: %w", err)
	}
	return count, nil
}
