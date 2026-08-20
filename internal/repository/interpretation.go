package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jb843051627/paleomag-lab/internal/model"
)

type InterpretationRepository struct{ db DBLike }

func (r *InterpretationRepository) Create(ctx context.Context, item model.Interpretation) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO interpretations(id, specimen_id, run_id, start_step, end_step, declination, inclination, intensity, confidence, status, notes, created_at, updated_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		item.ID, item.SpecimenID, item.RunID, item.StartStep, item.EndStep, item.Declination, item.Inclination, item.Intensity, item.Confidence, item.Status, item.Notes, timeText(item.CreatedAt), timeText(item.UpdatedAt))
	if err != nil {
		return fmt.Errorf("create interpretation: %w", err)
	}
	return nil
}

func (r *InterpretationRepository) Get(ctx context.Context, id string) (model.Interpretation, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, specimen_id, run_id, start_step, end_step, declination, inclination, intensity, confidence, status, notes, created_at, updated_at FROM interpretations WHERE id=?`, id)
	return scanInterpretation(row)
}

func (r *InterpretationRepository) ListBySpecimen(ctx context.Context, specimenID string, status model.InterpretationStatus) ([]model.Interpretation, error) {
	query := `SELECT id, specimen_id, run_id, start_step, end_step, declination, inclination, intensity, confidence, status, notes, created_at, updated_at FROM interpretations WHERE specimen_id=?`
	args := []any{specimenID}
	if status != "" {
		query += ` AND status=?`
		args = append(args, status)
	}
	query += ` ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list interpretations: %w", err)
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

func (r *InterpretationRepository) Update(ctx context.Context, item model.Interpretation) error {
	result, err := r.db.ExecContext(ctx, `UPDATE interpretations SET start_step=?, end_step=?, declination=?, inclination=?, intensity=?, confidence=?, status=?, notes=?, updated_at=? WHERE id=?`,
		item.StartStep, item.EndStep, item.Declination, item.Inclination, item.Intensity, item.Confidence, item.Status, item.Notes, timeText(item.UpdatedAt), item.ID)
	if err != nil {
		return fmt.Errorf("update interpretation: %w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("inspect interpretation update: %w", err)
	}
	if count == 0 {
		return fmt.Errorf("%w: interpretation %s", model.ErrNotFound, item.ID)
	}
	return nil
}

func scanInterpretation(s scanner) (model.Interpretation, error) {
	var item model.Interpretation
	var status, created, updated string
	if err := s.Scan(&item.ID, &item.SpecimenID, &item.RunID, &item.StartStep, &item.EndStep, &item.Declination, &item.Inclination, &item.Intensity, &item.Confidence, &status, &item.Notes, &created, &updated); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Interpretation{}, fmt.Errorf("%w: interpretation", model.ErrNotFound)
		}
		return model.Interpretation{}, fmt.Errorf("scan interpretation: %w", err)
	}
	item.Status = model.InterpretationStatus(status)
	var err error
	item.CreatedAt, err = parseTime(created)
	if err != nil {
		return model.Interpretation{}, err
	}
	item.UpdatedAt, err = parseTime(updated)
	if err != nil {
		return model.Interpretation{}, err
	}
	return item, nil
}
