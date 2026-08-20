package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jb843051627/paleomag-lab/internal/model"
)

type CalibrationRepository struct{ db DBLike }

func (r *CalibrationRepository) Create(ctx context.Context, item model.CalibrationRun) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO calibration_runs(id, instrument_id, reference, offset_x, offset_y, offset_z, status, valid_until, started_at, completed_at) VALUES(?,?,?,?,?,?,?,?,?,?)`,
		item.ID, item.InstrumentID, item.Reference, item.OffsetX, item.OffsetY, item.OffsetZ, item.Status, timeText(item.ValidUntil), timeText(item.StartedAt), nullableTime(item.CompletedAt))
	if err != nil {
		return fmt.Errorf("create calibration: %w", err)
	}
	return nil
}

func (r *CalibrationRepository) Get(ctx context.Context, id string) (model.CalibrationRun, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, instrument_id, reference, offset_x, offset_y, offset_z, status, valid_until, started_at, completed_at FROM calibration_runs WHERE id=?`, id)
	return scanCalibration(row)
}

func (r *CalibrationRepository) LatestUsable(ctx context.Context, instrumentID string, now string) (model.CalibrationRun, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, instrument_id, reference, offset_x, offset_y, offset_z, status, valid_until, started_at, completed_at FROM calibration_runs WHERE instrument_id=? AND status=? AND valid_until>? ORDER BY valid_until DESC LIMIT 1`, instrumentID, model.CalibrationValid, now)
	return scanCalibration(row)
}

func (r *CalibrationRepository) ListByInstrument(ctx context.Context, instrumentID string) ([]model.CalibrationRun, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, instrument_id, reference, offset_x, offset_y, offset_z, status, valid_until, started_at, completed_at FROM calibration_runs WHERE instrument_id=? ORDER BY started_at DESC`, instrumentID)
	if err != nil {
		return nil, fmt.Errorf("list calibrations: %w", err)
	}
	defer rows.Close()
	items := make([]model.CalibrationRun, 0)
	for rows.Next() {
		item, err := scanCalibration(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *CalibrationRepository) Update(ctx context.Context, item model.CalibrationRun) error {
	result, err := r.db.ExecContext(ctx, `UPDATE calibration_runs SET reference=?, offset_x=?, offset_y=?, offset_z=?, status=?, valid_until=?, started_at=?, completed_at=? WHERE id=?`,
		item.Reference, item.OffsetX, item.OffsetY, item.OffsetZ, item.Status, timeText(item.ValidUntil), timeText(item.StartedAt), nullableTime(item.CompletedAt), item.ID)
	if err != nil {
		return fmt.Errorf("update calibration: %w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("inspect calibration update: %w", err)
	}
	if count == 0 {
		return fmt.Errorf("%w: calibration %s", model.ErrNotFound, item.ID)
	}
	return nil
}

func scanCalibration(s scanner) (model.CalibrationRun, error) {
	var item model.CalibrationRun
	var status string
	var validUntil, started, completed sql.NullString
	if err := s.Scan(&item.ID, &item.InstrumentID, &item.Reference, &item.OffsetX, &item.OffsetY, &item.OffsetZ, &status, &validUntil, &started, &completed); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.CalibrationRun{}, fmt.Errorf("%w: calibration", model.ErrNotFound)
		}
		return model.CalibrationRun{}, fmt.Errorf("scan calibration: %w", err)
	}
	item.Status = model.CalibrationStatus(status)
	var err error
	item.ValidUntil, err = parseTime(validUntil.String)
	if err != nil {
		return model.CalibrationRun{}, err
	}
	item.StartedAt, err = parseTime(started.String)
	if err != nil {
		return model.CalibrationRun{}, err
	}
	if completed.Valid {
		value, err := parseTime(completed.String)
		if err != nil {
			return model.CalibrationRun{}, err
		}
		item.CompletedAt = &value
	}
	return item, nil
}
