package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jb843051627/paleomag-lab/internal/model"
)

type SpecimenRepository struct{ db DBLike }

func (r *SpecimenRepository) Create(ctx context.Context, specimen model.Specimen) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO specimens
		(id, code, lithology, site, collector, status, declination, inclination, coordinate, orientation_source, notes, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		specimen.ID, specimen.Code, specimen.Lithology, specimen.Site, specimen.Collector, specimen.Status,
		orientationValue(specimen.Orientation, "declination"), orientationValue(specimen.Orientation, "inclination"),
		orientationValue(specimen.Orientation, "coordinate"), orientationValue(specimen.Orientation, "source"), specimen.Notes,
		timeText(specimen.CreatedAt), timeText(specimen.UpdatedAt))
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			return fmt.Errorf("%w: specimen code already exists", model.ErrConflict)
		}
		return fmt.Errorf("create specimen: %w", err)
	}
	return nil
}

func (r *SpecimenRepository) Get(ctx context.Context, id string) (model.Specimen, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, code, lithology, site, collector, status, declination, inclination, coordinate, orientation_source, notes, created_at, updated_at FROM specimens WHERE id = ?`, id)
	return scanSpecimen(row)
}

func (r *SpecimenRepository) GetByCode(ctx context.Context, code string) (model.Specimen, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, code, lithology, site, collector, status, declination, inclination, coordinate, orientation_source, notes, created_at, updated_at FROM specimens WHERE code = ?`, code)
	return scanSpecimen(row)
}

func (r *SpecimenRepository) List(ctx context.Context, filter model.SpecimenFilter) ([]model.Specimen, error) {
	filter = filter.Normalize()
	query := `SELECT id, code, lithology, site, collector, status, declination, inclination, coordinate, orientation_source, notes, created_at, updated_at FROM specimens WHERE 1=1`
	args := make([]any, 0, 6)
	if filter.Status != "" {
		query += ` AND status = ?`
		args = append(args, filter.Status)
	}
	if filter.Site != "" {
		query += ` AND site = ?`
		args = append(args, filter.Site)
	}
	if filter.Lithology != "" {
		query += ` AND lithology = ?`
		args = append(args, filter.Lithology)
	}
	if filter.Search != "" {
		query += ` AND (code LIKE ? OR notes LIKE ?)`
		pattern := "%" + filter.Search + "%"
		args = append(args, pattern, pattern)
	}
	query += ` ORDER BY created_at DESC LIMIT ? OFFSET ?`
	args = append(args, filter.Limit, filter.Offset)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list specimens: %w", err)
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
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate specimens: %w", err)
	}
	return items, nil
}

func (r *SpecimenRepository) Update(ctx context.Context, specimen model.Specimen) error {
	result, err := r.db.ExecContext(ctx, `UPDATE specimens SET code=?, lithology=?, site=?, collector=?, status=?, declination=?, inclination=?, coordinate=?, orientation_source=?, notes=?, updated_at=? WHERE id=?`,
		specimen.Code, specimen.Lithology, specimen.Site, specimen.Collector, specimen.Status,
		orientationValue(specimen.Orientation, "declination"), orientationValue(specimen.Orientation, "inclination"),
		orientationValue(specimen.Orientation, "coordinate"), orientationValue(specimen.Orientation, "source"), specimen.Notes,
		timeText(specimen.UpdatedAt), specimen.ID)
	if err != nil {
		return fmt.Errorf("update specimen: %w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("inspect specimen update: %w", err)
	}
	if count == 0 {
		return fmt.Errorf("%w: specimen %s", model.ErrNotFound, specimen.ID)
	}
	return nil
}

func (r *SpecimenRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM specimens WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete specimen: %w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("inspect specimen delete: %w", err)
	}
	if count == 0 {
		return fmt.Errorf("%w: specimen %s", model.ErrNotFound, id)
	}
	return nil
}

func scanSpecimen(s scanner) (model.Specimen, error) {
	var item model.Specimen
	var status string
	var declination, inclination sql.NullFloat64
	var coordinate, source sql.NullString
	var created, updated string
	if err := s.Scan(&item.ID, &item.Code, &item.Lithology, &item.Site, &item.Collector, &status, &declination, &inclination, &coordinate, &source, &item.Notes, &created, &updated); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Specimen{}, fmt.Errorf("%w: specimen", model.ErrNotFound)
		}
		return model.Specimen{}, fmt.Errorf("scan specimen: %w", err)
	}
	item.Status = model.SpecimenStatus(status)
	var err error
	item.CreatedAt, err = parseTime(created)
	if err != nil {
		return model.Specimen{}, err
	}
	item.UpdatedAt, err = parseTime(updated)
	if err != nil {
		return model.Specimen{}, err
	}
	if declination.Valid || inclination.Valid || coordinate.Valid || source.Valid {
		item.Orientation = &model.Orientation{Declination: declination.Float64, Inclination: inclination.Float64, Coordinate: coordinate.String, Source: source.String}
	}
	return item, nil
}

func orientationValue(o *model.Orientation, field string) any {
	if o == nil {
		return nil
	}
	switch field {
	case "declination":
		return o.Declination
	case "inclination":
		return o.Inclination
	case "coordinate":
		return o.Coordinate
	case "source":
		return o.Source
	default:
		return nil
	}
}
