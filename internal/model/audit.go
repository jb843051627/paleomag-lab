package model

import "time"

type AuditEvent struct {
	ID         string    `json:"id"`
	EntityType string    `json:"entity_type"`
	EntityID   string    `json:"entity_id"`
	Action     string    `json:"action"`
	Actor      string    `json:"actor"`
	Before     string    `json:"before"`
	After      string    `json:"after"`
	CreatedAt  time.Time `json:"created_at"`
}

type AuditFilter struct {
	EntityType string
	EntityID   string
	Action     string
	Limit      int
}
