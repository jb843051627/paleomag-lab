package store

import (
	"context"
	"time"
)

type Health struct {
	Path      string    `json:"path"`
	Connected bool      `json:"connected"`
	CheckedAt time.Time `json:"checked_at"`
}

func (d *DB) HealthCheck(ctx context.Context) Health {
	result := Health{Path: d.Path, CheckedAt: time.Now().UTC()}
	result.Connected = d.Ping(ctx) == nil
	return result
}
