package service

import (
	"context"
	"fmt"

	"github.com/jb843051627/paleomag-lab/internal/clock"
	"github.com/jb843051627/paleomag-lab/internal/model"
	"github.com/jb843051627/paleomag-lab/internal/repository"
)

type SpecimenService struct {
	repo   *repository.SpecimenRepository
	audits *AuditService
	clock  clock.Clock
}

type CreateSpecimenInput struct {
	Code      string `json:"code"`
	Lithology string `json:"lithology"`
	Site      string `json:"site"`
	Collector string `json:"collector"`
	Notes     string `json:"notes"`
}

func (s *SpecimenService) Create(ctx context.Context, input CreateSpecimenInput) (model.Specimen, error) {
	item := model.NewSpecimen(newID("sp"), input.Code, input.Lithology, input.Site, input.Collector, s.clock.Now())
	item.Notes = input.Notes
	if err := item.Validate(); err != nil {
		return model.Specimen{}, err
	}
	if err := s.repo.Create(ctx, item); err != nil {
		return model.Specimen{}, err
	}
	if err := auditState(ctx, s.audits, "specimen", item.ID, "created", input.Collector, nil, item); err != nil {
		return model.Specimen{}, err
	}
	return item, nil
}

func (s *SpecimenService) Get(ctx context.Context, id string) (model.Specimen, error) {
	return s.repo.Get(ctx, id)
}

func (s *SpecimenService) GetByCode(ctx context.Context, code string) (model.Specimen, error) {
	return s.repo.GetByCode(ctx, code)
}

func (s *SpecimenService) List(ctx context.Context, filter model.SpecimenFilter) ([]model.Specimen, error) {
	return s.repo.List(ctx, filter)
}

func (s *SpecimenService) Orient(ctx context.Context, id string, orientation model.Orientation, actor string) (model.Specimen, error) {
	item, err := s.repo.Get(ctx, id)
	if err != nil {
		return model.Specimen{}, err
	}
	before := item
	if err := item.SetOrientation(orientation, s.clock.Now()); err != nil {
		return model.Specimen{}, err
	}
	if err := s.repo.Update(ctx, item); err != nil {
		return model.Specimen{}, err
	}
	if err := auditState(ctx, s.audits, "specimen", id, "oriented", actor, before, item); err != nil {
		return model.Specimen{}, err
	}
	return item, nil
}

func (s *SpecimenService) StartWork(ctx context.Context, id, actor string) (model.Specimen, error) {
	item, err := s.repo.Get(ctx, id)
	if err != nil {
		return model.Specimen{}, err
	}
	before := item
	if err := item.StartWork(s.clock.Now()); err != nil {
		return model.Specimen{}, err
	}
	if err := s.repo.Update(ctx, item); err != nil {
		return model.Specimen{}, err
	}
	if err := auditState(ctx, s.audits, "specimen", id, "work_started", actor, before, item); err != nil {
		return model.Specimen{}, err
	}
	return item, nil
}

func (s *SpecimenService) MarkInterpreted(ctx context.Context, id, actor string) (model.Specimen, error) {
	item, err := s.repo.Get(ctx, id)
	if err != nil {
		return model.Specimen{}, err
	}
	before := item
	if err := item.MarkInterpreted(s.clock.Now()); err != nil {
		return model.Specimen{}, err
	}
	if err := s.repo.Update(ctx, item); err != nil {
		return model.Specimen{}, err
	}
	if err := auditState(ctx, s.audits, "specimen", id, "interpreted", actor, before, item); err != nil {
		return model.Specimen{}, err
	}
	return item, nil
}

func (s *SpecimenService) Archive(ctx context.Context, id, actor string) (model.Specimen, error) {
	item, err := s.repo.Get(ctx, id)
	if err != nil {
		return model.Specimen{}, err
	}
	before := item
	if err := item.Archive(s.clock.Now()); err != nil {
		return model.Specimen{}, err
	}
	if err := s.repo.Update(ctx, item); err != nil {
		return model.Specimen{}, err
	}
	if err := auditState(ctx, s.audits, "specimen", id, "archived", actor, before, item); err != nil {
		return model.Specimen{}, err
	}
	return item, nil
}

func (s *SpecimenService) UpdateNotes(ctx context.Context, id, notes, actor string) (model.Specimen, error) {
	item, err := s.repo.Get(ctx, id)
	if err != nil {
		return model.Specimen{}, err
	}
	if len(notes) > 2000 {
		return model.Specimen{}, fmt.Errorf("%w: notes exceed 2000 characters", model.ErrInvalid)
	}
	before := item
	item.Notes = notes
	item.UpdatedAt = s.clock.Now()
	if err := s.repo.Update(ctx, item); err != nil {
		return model.Specimen{}, err
	}
	if err := auditState(ctx, s.audits, "specimen", id, "notes_updated", actor, before, item); err != nil {
		return model.Specimen{}, err
	}
	return item, nil
}
