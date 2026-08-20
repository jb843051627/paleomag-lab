package repository

import (
	"database/sql"
	"fmt"
	"time"
)

type scanner interface {
	Scan(dest ...any) error
}

func timeText(value time.Time) string { return value.UTC().Format(time.RFC3339Nano) }

func parseTime(value string) (time.Time, error) {
	if value == "" {
		return time.Time{}, nil
	}
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse time %q: %w", value, err)
	}
	return parsed, nil
}

func nullableTime(value *time.Time) any {
	if value == nil {
		return nil
	}
	return timeText(*value)
}

func scanTime(value string) (time.Time, error) { return parseTime(value) }

func noRows(err error) error {
	if err == sql.ErrNoRows {
		return fmt.Errorf("%w: %v", modelNotFound(), err)
	}
	return err
}

func modelNotFound() error {
	return fmt.Errorf("paleomag: not found")
}
