package store

import (
	"context"
	"fmt"
)

func Migrate(ctx context.Context, d *DB) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS schema_migrations (version INTEGER PRIMARY KEY, applied_at TEXT NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS specimens (
			id TEXT PRIMARY KEY, code TEXT NOT NULL UNIQUE, lithology TEXT NOT NULL, site TEXT NOT NULL,
			collector TEXT NOT NULL, status TEXT NOT NULL, declination REAL, inclination REAL,
			coordinate TEXT, orientation_source TEXT, notes TEXT NOT NULL DEFAULT '', created_at TEXT NOT NULL, updated_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS demag_plans (
			id TEXT PRIMARY KEY, specimen_id TEXT NOT NULL REFERENCES specimens(id), name TEXT NOT NULL,
			method TEXT NOT NULL, status TEXT NOT NULL, operator TEXT NOT NULL, created_at TEXT NOT NULL, updated_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS demag_steps (
			id TEXT PRIMARY KEY, plan_id TEXT NOT NULL REFERENCES demag_plans(id) ON DELETE CASCADE,
			sequence_no INTEGER NOT NULL, field REAL NOT NULL, temperature REAL NOT NULL, status TEXT NOT NULL,
			note TEXT NOT NULL DEFAULT '', UNIQUE(plan_id, sequence_no)
		)`,
		`CREATE TABLE IF NOT EXISTS instruments (
			id TEXT PRIMARY KEY, name TEXT NOT NULL, serial TEXT NOT NULL UNIQUE, kind TEXT NOT NULL,
			min_field REAL NOT NULL, max_field REAL NOT NULL, status TEXT NOT NULL,
			last_calibrated_at TEXT, created_at TEXT NOT NULL, updated_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS calibration_runs (
			id TEXT PRIMARY KEY, instrument_id TEXT NOT NULL REFERENCES instruments(id), reference TEXT NOT NULL,
			offset_x REAL NOT NULL, offset_y REAL NOT NULL, offset_z REAL NOT NULL, status TEXT NOT NULL,
			valid_until TEXT NOT NULL, started_at TEXT NOT NULL, completed_at TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS measurement_runs (
			id TEXT PRIMARY KEY, specimen_id TEXT NOT NULL REFERENCES specimens(id), plan_id TEXT NOT NULL REFERENCES demag_plans(id),
			instrument_id TEXT NOT NULL REFERENCES instruments(id), calibration_id TEXT NOT NULL REFERENCES calibration_runs(id),
			status TEXT NOT NULL, started_at TEXT NOT NULL, completed_at TEXT, quality_score REAL NOT NULL DEFAULT 0
		)`,
		`CREATE TABLE IF NOT EXISTS measurements (
			id TEXT PRIMARY KEY, run_id TEXT NOT NULL REFERENCES measurement_runs(id) ON DELETE CASCADE,
			step_no INTEGER NOT NULL, field REAL NOT NULL, x REAL NOT NULL, y REAL NOT NULL, z REAL NOT NULL,
			intensity REAL NOT NULL, quality TEXT NOT NULL, measured_at TEXT NOT NULL, UNIQUE(run_id, step_no)
		)`,
		`CREATE TABLE IF NOT EXISTS interpretations (
			id TEXT PRIMARY KEY, specimen_id TEXT NOT NULL REFERENCES specimens(id), run_id TEXT NOT NULL REFERENCES measurement_runs(id),
			start_step INTEGER NOT NULL, end_step INTEGER NOT NULL, declination REAL NOT NULL, inclination REAL NOT NULL,
			intensity REAL NOT NULL, confidence REAL NOT NULL, status TEXT NOT NULL, notes TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL, updated_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS reviews (
			id TEXT PRIMARY KEY, interpretation_id TEXT NOT NULL REFERENCES interpretations(id), reviewer TEXT NOT NULL,
			decision TEXT NOT NULL, comment TEXT NOT NULL, created_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS export_jobs (
			id TEXT PRIMARY KEY, specimen_id TEXT NOT NULL REFERENCES specimens(id), format TEXT NOT NULL,
			requested_by TEXT NOT NULL, status TEXT NOT NULL, result TEXT NOT NULL DEFAULT '', error TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL, updated_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS audit_events (
			id TEXT PRIMARY KEY, entity_type TEXT NOT NULL, entity_id TEXT NOT NULL, action TEXT NOT NULL,
			actor TEXT NOT NULL, before_json TEXT NOT NULL, after_json TEXT NOT NULL, created_at TEXT NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_specimens_status ON specimens(status)`,
		`CREATE INDEX IF NOT EXISTS idx_plans_specimen ON demag_plans(specimen_id)`,
		`CREATE INDEX IF NOT EXISTS idx_runs_specimen ON measurement_runs(specimen_id)`,
		`CREATE INDEX IF NOT EXISTS idx_measurements_run ON measurements(run_id, step_no)`,
		`CREATE INDEX IF NOT EXISTS idx_interpretations_specimen ON interpretations(specimen_id, status)`,
		`CREATE INDEX IF NOT EXISTS idx_audits_entity ON audit_events(entity_type, entity_id, created_at)`,
	}

	tx, err := d.SQL.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin migration: %w", err)
	}
	for _, statement := range statements {
		if _, err := tx.ExecContext(ctx, statement); err != nil {
			tx.Rollback()
			return fmt.Errorf("migration statement: %w", err)
		}
	}
	if _, err := tx.ExecContext(ctx, `INSERT OR IGNORE INTO schema_migrations(version, applied_at) VALUES(1, datetime('now'))`); err != nil {
		tx.Rollback()
		return fmt.Errorf("record migration: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit migration: %w", err)
	}
	return nil
}
