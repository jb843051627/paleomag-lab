package store

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	_ "modernc.org/sqlite"
)

type DB struct {
	SQL  *sql.DB
	Path string
}

func Open(ctx context.Context, path string) (*DB, error) {
	if path == "" || path == ":memory:" {
		return nil, fmt.Errorf("database path must be a file")
	}
	if err := os.MkdirAll(filepathDir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create database directory: %w", err)
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	result := &DB{SQL: db, Path: path}
	if err := result.configure(ctx); err != nil {
		db.Close()
		return nil, err
	}
	if err := Migrate(ctx, result); err != nil {
		db.Close()
		return nil, err
	}
	return result, nil
}

func filepathDir(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' || path[i] == '\\' {
			if i == 0 {
				return path[:1]
			}
			return path[:i]
		}
	}
	return "."
}

func (d *DB) configure(ctx context.Context) error {
	statements := []string{
		"PRAGMA foreign_keys = ON",
		"PRAGMA busy_timeout = 5000",
		"PRAGMA journal_mode = WAL",
	}
	for _, statement := range statements {
		if _, err := d.SQL.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("configure sqlite: %w", err)
		}
	}
	return nil
}

func (d *DB) Close() error { return d.SQL.Close() }

func (d *DB) Ping(ctx context.Context) error { return d.SQL.PingContext(ctx) }
