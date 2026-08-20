package service

import (
	"context"

	"github.com/jb843051627/paleomag-lab/internal/clock"
	"github.com/jb843051627/paleomag-lab/internal/model"
	"github.com/jb843051627/paleomag-lab/internal/repository"
)

type AuditService struct {
	repo  *repository.AuditRepository
	clock clock.Clock
}

func (s *AuditService) List(ctx context.Context, filter model.AuditFilter) ([]model.AuditEvent, error) {
	return s.repo.List(ctx, filter)
}

func (s *AuditService) Latest(ctx context.Context, entityType, entityID string) (model.AuditEvent, error) {
	return s.repo.Latest(ctx, entityType, entityID)
}
