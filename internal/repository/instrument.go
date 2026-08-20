package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jb843051627/paleomag-lab/internal/model"
)

type InstrumentRepository struct{ db DBLike }

func (r *InstrumentRepository) Create(ctx context.Context, item model.Instrument) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO instruments(id, name, serial, kind, min_field, max_field, status, last_calibrated_at, created_at, updated_at) VALUES(?,?,?,?,?,?,?,?,?,?)`,
		item.ID, item.Name, item.Serial, item.Kind, item.MinField, item.MaxField, item.Status, nullableTime(item.LastCalibratedAt), timeText(item.CreatedAt), timeText(item.UpdatedAt))
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			return fmt.Errorf("%w: instrument serial already exists", model.ErrConflict)
		}
		return fmt.Errorf("create instrument: %w", err)
	}
	return nil
}

func (r *InstrumentRepository) Get(ctx context.Context, id string) (model.Instrument, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, name, serial, kind, min_field, max_field, status, last_calibrated_at, created_at, updated_at FROM instruments WHERE id=?`, id)
	return scanInstrument(row)
}

func (r *InstrumentRepository) List(ctx context.Context, status model.InstrumentStatus) ([]model.Instrument, error) {
	query := `SELECT id, name, serial, kind, min_field, max_field, status, last_calibrated_at, created_at, updated_at FROM instruments`
	args := []any{}
	if status != "" {
		query += ` WHERE status=?`
		args = append(args, status)
	}
	query += ` ORDER BY name`
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list instruments: %w", err)
	}
	defer rows.Close()
	items := make([]model.Instrument, 0)
	for rows.Next() {
		item, err := scanInstrument(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *InstrumentRepository) Update(ctx context.Context, item model.Instrument) error {
	result, err := r.db.ExecContext(ctx, `UPDATE instruments SET name=?, serial=?, kind=?, min_field=?, max_field=?, status=?, last_calibrated_at=?, updated_at=? WHERE id=?`,
		item.Name, item.Serial, item.Kind, item.MinField, item.MaxField, item.Status, nullableTime(item.LastCalibratedAt), timeText(item.UpdatedAt), item.ID)
	if err != nil {
		return fmt.Errorf("update instrument: %w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("inspect instrument update: %w", err)
	}
	if count == 0 {
		return fmt.Errorf("%w: instrument %s", model.ErrNotFound, item.ID)
	}
	return nil
}

func scanInstrument(s scanner) (model.Instrument, error) {
	var item model.Instrument
	var status string
	var calibrated sql.NullString
	var created, updated string
	if err := s.Scan(&item.ID, &item.Name, &item.Serial, &item.Kind, &item.MinField, &item.MaxField, &status, &calibrated, &created, &updated); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Instrument{}, fmt.Errorf("%w: instrument", model.ErrNotFound)
		}
		return model.Instrument{}, fmt.Errorf("scan instrument: %w", err)
	}
	item.Status = model.InstrumentStatus(status)
	var err error
	item.CreatedAt, err = parseTime(created)
	if err != nil {
		return model.Instrument{}, err
	}
	item.UpdatedAt, err = parseTime(updated)
	if err != nil {
		return model.Instrument{}, err
	}
	if calibrated.Valid {
		value, err := parseTime(calibrated.String)
		if err != nil {
			return model.Instrument{}, err
		}
		item.LastCalibratedAt = &value
	}
	return item, nil
}
