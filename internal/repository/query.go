package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/jb843051627/paleomag-lab/internal/model"
)

type QueryRepository struct{ db DBLike }

type SpecimenCounts struct {
	Registered  int `json:"registered"`
	Oriented    int `json:"oriented"`
	InProgress  int `json:"in_progress"`
	Interpreted int `json:"interpreted"`
	Archived    int `json:"archived"`
}

func NewQueryRepository(db DBLike) *QueryRepository { return &QueryRepository{db: db} }

func (r *QueryRepository) CountSpecimens(ctx context.Context) (SpecimenCounts, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT status, COUNT(*) FROM specimens GROUP BY status`)
	if err != nil {
		return SpecimenCounts{}, fmt.Errorf("count specimens: %w", err)
	}
	defer rows.Close()
	var counts SpecimenCounts
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return SpecimenCounts{}, fmt.Errorf("scan specimen counts: %w", err)
		}
		switch model.SpecimenStatus(status) {
		case model.SpecimenRegistered:
			counts.Registered = count
		case model.SpecimenOriented:
			counts.Oriented = count
		case model.SpecimenInProgress:
			counts.InProgress = count
		case model.SpecimenInterpreted:
			counts.Interpreted = count
		case model.SpecimenArchived:
			counts.Archived = count
		}
	}
	if err := rows.Err(); err != nil {
		return SpecimenCounts{}, fmt.Errorf("iterate specimen counts: %w", err)
	}
	return counts, nil
}

func (r *QueryRepository) PendingInterpretations(ctx context.Context, limit int) ([]model.Interpretation, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows, err := r.db.QueryContext(ctx, `SELECT id, specimen_id, run_id, start_step, end_step, declination, inclination, intensity, confidence, status, notes, created_at, updated_at FROM interpretations WHERE status=? ORDER BY created_at LIMIT ?`, model.InterpretationPending, limit)
	if err != nil {
		return nil, fmt.Errorf("pending interpretations: %w", err)
	}
	defer rows.Close()
	items := make([]model.Interpretation, 0)
	for rows.Next() {
		item, err := scanInterpretation(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *QueryRepository) SearchByText(ctx context.Context, phrase string, limit int) ([]model.Specimen, error) {
	phrase = strings.TrimSpace(phrase)
	if phrase == "" {
		return []model.Specimen{}, nil
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	pattern := "%" + phrase + "%"
	rows, err := r.db.QueryContext(ctx, `SELECT id, code, lithology, site, collector, status, declination, inclination, coordinate, orientation_source, notes, created_at, updated_at FROM specimens WHERE code LIKE ? OR site LIKE ? OR notes LIKE ? ORDER BY updated_at DESC LIMIT ?`, pattern, pattern, pattern, limit)
	if err != nil {
		return nil, fmt.Errorf("search specimens: %w", err)
	}
	defer rows.Close()
	items := make([]model.Specimen, 0)
	for rows.Next() {
		item, err := scanSpecimen(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *QueryRepository) LatestAcceptedRun(ctx context.Context, specimenID string) (model.MeasurementRun, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, specimen_id, plan_id, instrument_id, calibration_id, status, started_at, completed_at, quality_score FROM measurement_runs WHERE specimen_id=? AND status=? ORDER BY completed_at DESC LIMIT 1`, specimenID, model.RunAccepted)
	return scanMeasurementRun(row)
}
