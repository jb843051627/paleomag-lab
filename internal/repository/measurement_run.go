package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jb843051627/paleomag-lab/internal/model"
)

type MeasurementRunRepository struct{ db DBLike }

func (r *MeasurementRunRepository) Create(ctx context.Context, item model.MeasurementRun) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO measurement_runs(id, specimen_id, plan_id, instrument_id, calibration_id, status, started_at, completed_at, quality_score) VALUES(?,?,?,?,?,?,?,?,?)`,
		item.ID, item.SpecimenID, item.PlanID, item.InstrumentID, item.CalibrationID, item.Status, timeText(item.StartedAt), nullableTime(item.CompletedAt), item.QualityScore)
	if err != nil {
		return fmt.Errorf("create measurement run: %w", err)
	}
	return nil
}

func (r *MeasurementRunRepository) Get(ctx context.Context, id string) (model.MeasurementRun, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, specimen_id, plan_id, instrument_id, calibration_id, status, started_at, completed_at, quality_score FROM measurement_runs WHERE id=?`, id)
	return scanMeasurementRun(row)
}

func (r *MeasurementRunRepository) ListBySpecimen(ctx context.Context, specimenID string) ([]model.MeasurementRun, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, specimen_id, plan_id, instrument_id, calibration_id, status, started_at, completed_at, quality_score FROM measurement_runs WHERE specimen_id=? ORDER BY started_at DESC`, specimenID)
	if err != nil {
		return nil, fmt.Errorf("list measurement runs: %w", err)
	}
	defer rows.Close()
	items := make([]model.MeasurementRun, 0)
	for rows.Next() {
		item, err := scanMeasurementRun(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *MeasurementRunRepository) Update(ctx context.Context, item model.MeasurementRun) error {
	result, err := r.db.ExecContext(ctx, `UPDATE measurement_runs SET status=?, completed_at=?, quality_score=? WHERE id=?`, item.Status, nullableTime(item.CompletedAt), item.QualityScore, item.ID)
	if err != nil {
		return fmt.Errorf("update measurement run: %w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("inspect measurement run update: %w", err)
	}
	if count == 0 {
		return fmt.Errorf("%w: measurement run %s", model.ErrNotFound, item.ID)
	}
	return nil
}

func scanMeasurementRun(s scanner) (model.MeasurementRun, error) {
	var item model.MeasurementRun
	var status string
	var started, completed sql.NullString
	if err := s.Scan(&item.ID, &item.SpecimenID, &item.PlanID, &item.InstrumentID, &item.CalibrationID, &status, &started, &completed, &item.QualityScore); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.MeasurementRun{}, fmt.Errorf("%w: measurement run", model.ErrNotFound)
		}
		return model.MeasurementRun{}, fmt.Errorf("scan measurement run: %w", err)
	}
	item.Status = model.RunStatus(status)
	var err error
	item.StartedAt, err = parseTime(started.String)
	if err != nil {
		return model.MeasurementRun{}, err
	}
	if completed.Valid {
		value, err := parseTime(completed.String)
		if err != nil {
			return model.MeasurementRun{}, err
		}
		item.CompletedAt = &value
	}
	return item, nil
}
