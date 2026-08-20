package repository

import (
	"context"
	"database/sql"
	"fmt"
)

type TransactionRepository struct{ db DBLike }

func NewTransactionRepository(db DBLike) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (r *TransactionRepository) UpdateRunAndSpecimen(ctx context.Context, tx *sql.Tx, runID, runStatus, specimenID, specimenStatus, updatedAt string) error {
	result, err := tx.ExecContext(ctx, `UPDATE measurement_runs SET status=?, completed_at=? WHERE id=?`, runStatus, updatedAt, runID)
	if err != nil {
		return fmt.Errorf("update run in transaction: %w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("inspect run transaction update: %w", err)
	}
	if count != 1 {
		return fmt.Errorf("run transaction update affected %d rows", count)
	}
	result, err = tx.ExecContext(ctx, `UPDATE specimens SET status=?, updated_at=? WHERE id=?`, specimenStatus, updatedAt, specimenID)
	if err != nil {
		return fmt.Errorf("update specimen in transaction: %w", err)
	}
	count, err = result.RowsAffected()
	if err != nil {
		return fmt.Errorf("inspect specimen transaction update: %w", err)
	}
	if count != 1 {
		return fmt.Errorf("specimen transaction update affected %d rows", count)
	}
	return nil
}

func (r *TransactionRepository) LockRun(ctx context.Context, tx *sql.Tx, runID string) error {
	row := tx.QueryRowContext(ctx, `SELECT id FROM measurement_runs WHERE id=?`, runID)
	var id string
	if err := row.Scan(&id); err != nil {
		return fmt.Errorf("lock run: %w", err)
	}
	return nil
}
