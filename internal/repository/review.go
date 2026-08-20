package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jb843051627/paleomag-lab/internal/model"
)

type ReviewRepository struct{ db DBLike }

func (r *ReviewRepository) Create(ctx context.Context, item model.Review) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO reviews(id, interpretation_id, reviewer, decision, comment, created_at) VALUES(?,?,?,?,?,?)`, item.ID, item.InterpretationID, item.Reviewer, item.Decision, item.Comment, timeText(item.CreatedAt))
	if err != nil {
		return fmt.Errorf("create review: %w", err)
	}
	return nil
}

func (r *ReviewRepository) Get(ctx context.Context, id string) (model.Review, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, interpretation_id, reviewer, decision, comment, created_at FROM reviews WHERE id=?`, id)
	return scanReview(row)
}

func (r *ReviewRepository) ListByInterpretation(ctx context.Context, interpretationID string) ([]model.Review, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, interpretation_id, reviewer, decision, comment, created_at FROM reviews WHERE interpretation_id=? ORDER BY created_at DESC`, interpretationID)
	if err != nil {
		return nil, fmt.Errorf("list reviews: %w", err)
	}
	defer rows.Close()
	items := make([]model.Review, 0)
	for rows.Next() {
		item, err := scanReview(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func scanReview(s scanner) (model.Review, error) {
	var item model.Review
	var decision, created string
	if err := s.Scan(&item.ID, &item.InterpretationID, &item.Reviewer, &decision, &item.Comment, &created); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Review{}, fmt.Errorf("%w: review", model.ErrNotFound)
		}
		return model.Review{}, fmt.Errorf("scan review: %w", err)
	}
	item.Decision = model.ReviewDecision(decision)
	var err error
	item.CreatedAt, err = parseTime(created)
	if err != nil {
		return model.Review{}, err
	}
	return item, nil
}
