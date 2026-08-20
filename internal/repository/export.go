package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jb843051627/paleomag-lab/internal/model"
)

type ExportRepository struct{ db DBLike }

func (r *ExportRepository) Create(ctx context.Context, item model.ExportJob) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO export_jobs(id, specimen_id, format, requested_by, status, result, error, created_at, updated_at) VALUES(?,?,?,?,?,?,?,?,?)`,
		item.ID, item.SpecimenID, item.Format, item.RequestedBy, item.Status, item.Result, item.Error, timeText(item.CreatedAt), timeText(item.UpdatedAt))
	if err != nil {
		return fmt.Errorf("create export job: %w", err)
	}
	return nil
}

func (r *ExportRepository) Get(ctx context.Context, id string) (model.ExportJob, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, specimen_id, format, requested_by, status, result, error, created_at, updated_at FROM export_jobs WHERE id=?`, id)
	return scanExport(row)
}

func (r *ExportRepository) ListBySpecimen(ctx context.Context, specimenID string) ([]model.ExportJob, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, specimen_id, format, requested_by, status, result, error, created_at, updated_at FROM export_jobs WHERE specimen_id=? ORDER BY created_at DESC`, specimenID)
	if err != nil {
		return nil, fmt.Errorf("list export jobs: %w", err)
	}
	defer rows.Close()
	items := make([]model.ExportJob, 0)
	for rows.Next() {
		item, err := scanExport(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *ExportRepository) Update(ctx context.Context, item model.ExportJob) error {
	result, err := r.db.ExecContext(ctx, `UPDATE export_jobs SET status=?, result=?, error=?, updated_at=? WHERE id=?`, item.Status, item.Result, item.Error, timeText(item.UpdatedAt), item.ID)
	if err != nil {
		return fmt.Errorf("update export job: %w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("inspect export update: %w", err)
	}
	if count == 0 {
		return fmt.Errorf("%w: export job %s", model.ErrNotFound, item.ID)
	}
	return nil
}

func scanExport(s scanner) (model.ExportJob, error) {
	var item model.ExportJob
	var status, created, updated string
	if err := s.Scan(&item.ID, &item.SpecimenID, &item.Format, &item.RequestedBy, &status, &item.Result, &item.Error, &created, &updated); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.ExportJob{}, fmt.Errorf("%w: export job", model.ErrNotFound)
		}
		return model.ExportJob{}, fmt.Errorf("scan export job: %w", err)
	}
	item.Status = model.ExportStatus(status)
	var err error
	item.CreatedAt, err = parseTime(created)
	if err != nil {
		return model.ExportJob{}, err
	}
	item.UpdatedAt, err = parseTime(updated)
	if err != nil {
		return model.ExportJob{}, err
	}
	return item, nil
}
