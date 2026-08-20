package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jb843051627/paleomag-lab/internal/model"
)

type AuditRepository struct{ db DBLike }

func (r *AuditRepository) Create(ctx context.Context, item model.AuditEvent) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO audit_events(id, entity_type, entity_id, action, actor, before_json, after_json, created_at) VALUES(?,?,?,?,?,?,?,?)`,
		item.ID, item.EntityType, item.EntityID, item.Action, item.Actor, item.Before, item.After, timeText(item.CreatedAt))
	if err != nil {
		return fmt.Errorf("create audit event: %w", err)
	}
	return nil
}

func (r *AuditRepository) List(ctx context.Context, filter model.AuditFilter) ([]model.AuditEvent, error) {
	limit := filter.Limit
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	query := `SELECT id, entity_type, entity_id, action, actor, before_json, after_json, created_at FROM audit_events WHERE 1=1`
	args := make([]any, 0, 4)
	if filter.EntityType != "" {
		query += ` AND entity_type=?`
		args = append(args, filter.EntityType)
	}
	if filter.EntityID != "" {
		query += ` AND entity_id=?`
		args = append(args, filter.EntityID)
	}
	if filter.Action != "" {
		query += ` AND action=?`
		args = append(args, filter.Action)
	}
	query += ` ORDER BY created_at DESC LIMIT ?`
	args = append(args, limit)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list audit events: %w", err)
	}
	defer rows.Close()
	items := make([]model.AuditEvent, 0)
	for rows.Next() {
		var item model.AuditEvent
		var created string
		if err := rows.Scan(&item.ID, &item.EntityType, &item.EntityID, &item.Action, &item.Actor, &item.Before, &item.After, &created); err != nil {
			return nil, fmt.Errorf("scan audit event: %w", err)
		}
		item.CreatedAt, err = parseTime(created)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate audit events: %w", err)
	}
	return items, nil
}

func (r *AuditRepository) Latest(ctx context.Context, entityType, entityID string) (model.AuditEvent, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, entity_type, entity_id, action, actor, before_json, after_json, created_at FROM audit_events WHERE entity_type=? AND entity_id=? ORDER BY created_at DESC LIMIT 1`, entityType, entityID)
	var item model.AuditEvent
	var created string
	if err := row.Scan(&item.ID, &item.EntityType, &item.EntityID, &item.Action, &item.Actor, &item.Before, &item.After, &created); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.AuditEvent{}, fmt.Errorf("%w: audit event", model.ErrNotFound)
		}
		return model.AuditEvent{}, fmt.Errorf("get latest audit event: %w", err)
	}
	var err error
	item.CreatedAt, err = parseTime(created)
	if err != nil {
		return model.AuditEvent{}, err
	}
	return item, nil
}
