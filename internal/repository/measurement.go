package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jb843051627/paleomag-lab/internal/model"
)

type MeasurementRepository struct{ db DBLike }

func (r *MeasurementRepository) Create(ctx context.Context, item model.Measurement) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO measurements(id, run_id, step_no, field, x, y, z, intensity, quality, measured_at) VALUES(?,?,?,?,?,?,?,?,?,?)`,
		item.ID, item.RunID, item.Step, item.Field, item.X, item.Y, item.Z, item.Intensity, item.Quality, timeText(item.MeasuredAt))
	if err != nil {
		return fmt.Errorf("create measurement: %w", err)
	}
	return nil
}

func (r *MeasurementRepository) CreateTx(ctx context.Context, tx *sql.Tx, item model.Measurement) error {
	_, err := tx.ExecContext(ctx, `INSERT INTO measurements(id, run_id, step_no, field, x, y, z, intensity, quality, measured_at) VALUES(?,?,?,?,?,?,?,?,?,?)`,
		item.ID, item.RunID, item.Step, item.Field, item.X, item.Y, item.Z, item.Intensity, item.Quality, timeText(item.MeasuredAt))
	if err != nil {
		return fmt.Errorf("create measurement: %w", err)
	}
	return nil
}

func (r *MeasurementRepository) Get(ctx context.Context, id string) (model.Measurement, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, run_id, step_no, field, x, y, z, intensity, quality, measured_at FROM measurements WHERE id=?`, id)
	return scanMeasurement(row)
}

func (r *MeasurementRepository) ListByRun(ctx context.Context, filter model.MeasurementFilter) ([]model.Measurement, error) {
	filter = filter.Normalize()
	query := `SELECT id, run_id, step_no, field, x, y, z, intensity, quality, measured_at FROM measurements WHERE run_id=?`
	args := []any{filter.RunID}
	if filter.Quality != "" {
		query += ` AND quality=?`
		args = append(args, filter.Quality)
	}
	if filter.MinField != nil {
		query += ` AND field>=?`
		args = append(args, *filter.MinField)
	}
	if filter.MaxField != nil {
		query += ` AND field<=?`
		args = append(args, *filter.MaxField)
	}
	if filter.Since != nil {
		query += ` AND measured_at>=?`
		args = append(args, timeText(*filter.Since))
	}
	if filter.Until != nil {
		query += ` AND measured_at<?`
		args = append(args, timeText(*filter.Until))
	}
	query += ` ORDER BY step_no LIMIT ?`
	args = append(args, filter.Limit)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list measurements: %w", err)
	}
	defer rows.Close()
	items := make([]model.Measurement, 0)
	for rows.Next() {
		item, err := scanMeasurement(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate measurements: %w", err)
	}
	return items, nil
}

func (r *MeasurementRepository) UpdateQuality(ctx context.Context, id string, quality model.QualityFlag) error {
	result, err := r.db.ExecContext(ctx, `UPDATE measurements SET quality=? WHERE id=?`, quality, id)
	if err != nil {
		return fmt.Errorf("update measurement quality: %w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("inspect quality update: %w", err)
	}
	if count == 0 {
		return fmt.Errorf("%w: measurement %s", model.ErrNotFound, id)
	}
	return nil
}

func (r *MeasurementRepository) CountByRun(ctx context.Context, runID string) (int, error) {
	row := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM measurements WHERE run_id=?`, runID)
	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("count measurements: %w", err)
	}
	return count, nil
}

func scanMeasurement(s scanner) (model.Measurement, error) {
	var item model.Measurement
	var quality, measured string
	if err := s.Scan(&item.ID, &item.RunID, &item.Step, &item.Field, &item.X, &item.Y, &item.Z, &item.Intensity, &quality, &measured); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Measurement{}, fmt.Errorf("%w: measurement", model.ErrNotFound)
		}
		return model.Measurement{}, fmt.Errorf("scan measurement: %w", err)
	}
	item.Quality = model.QualityFlag(quality)
	var err error
	item.MeasuredAt, err = parseTime(measured)
	if err != nil {
		return model.Measurement{}, err
	}
	return item, nil
}
