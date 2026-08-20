package service

import (
	"context"
	"fmt"
	"math"

	"github.com/jb843051627/paleomag-lab/internal/clock"
	"github.com/jb843051627/paleomag-lab/internal/model"
	"github.com/jb843051627/paleomag-lab/internal/repository"
)

type InterpretationService struct {
	repo         *repository.InterpretationRepository
	runs         *repository.MeasurementRunRepository
	measurements *repository.MeasurementRepository
	specimens    *repository.SpecimenRepository
	audits       *AuditService
	clock        clock.Clock
}

func (s *InterpretationService) Generate(ctx context.Context, runID, actor string) (model.Interpretation, error) {
	run, err := s.runs.Get(ctx, runID)
	if err != nil {
		return model.Interpretation{}, err
	}
	if run.Status != model.RunAccepted && run.Status != model.RunEvaluated {
		return model.Interpretation{}, fmt.Errorf("%w: run must be accepted or evaluated", model.ErrState)
	}
	items, err := s.measurements.ListByRun(ctx, model.MeasurementFilter{RunID: runID})
	if err != nil {
		return model.Interpretation{}, err
	}
	if len(items) < 2 {
		return model.Interpretation{}, fmt.Errorf("%w: at least two measurements are required", model.ErrInvalid)
	}
	var x, y, z, intensity float64
	for _, item := range items {
		x += item.X
		y += item.Y
		z += item.Z
		intensity += item.Intensity
	}
	count := float64(len(items))
	x, y, z, intensity = x/count, y/count, z/count, intensity/count
	declination := math.Mod(math.Atan2(y, x)*180/math.Pi+360, 360)
	inclination := math.Atan2(z, math.Hypot(x, y)) * 180 / math.Pi
	confidence := confidenceScore(items)
	item := model.Interpretation{ID: newID("interpretation"), SpecimenID: run.SpecimenID, RunID: runID, StartStep: items[0].Step, EndStep: items[len(items)-1].Step, Declination: declination, Inclination: inclination, Intensity: intensity, Confidence: confidence, Status: model.InterpretationDraft, CreatedAt: s.clock.Now(), UpdatedAt: s.clock.Now()}
	if err := item.Validate(); err != nil {
		return model.Interpretation{}, err
	}
	if err := s.repo.Create(ctx, item); err != nil {
		return model.Interpretation{}, err
	}
	if err := auditState(ctx, s.audits, "interpretation", item.ID, "generated", actor, nil, item); err != nil {
		return model.Interpretation{}, err
	}
	return item, nil
}

func confidenceScore(items []model.Measurement) float64 {
	if len(items) == 0 {
		return 0
	}
	mean := 0.0
	for _, item := range items {
		mean += item.VectorLength()
	}
	mean /= float64(len(items))
	if mean == 0 {
		return 0
	}
	variance := 0.0
	for _, item := range items {
		delta := item.VectorLength() - mean
		variance += delta * delta
	}
	spread := math.Sqrt(variance / float64(len(items)))
	confidence := 1 - spread/mean
	if confidence < 0 {
		return 0
	}
	if confidence > 1 {
		return 1
	}
	return confidence
}

func (s *InterpretationService) Get(ctx context.Context, id string) (model.Interpretation, error) {
	return s.repo.Get(ctx, id)
}

func (s *InterpretationService) ListBySpecimen(ctx context.Context, specimenID string, status model.InterpretationStatus) ([]model.Interpretation, error) {
	return s.repo.ListBySpecimen(ctx, specimenID, status)
}

func (s *InterpretationService) Submit(ctx context.Context, id, actor string) (model.Interpretation, error) {
	item, err := s.repo.Get(ctx, id)
	if err != nil {
		return model.Interpretation{}, err
	}
	before := item
	if err := item.Submit(s.clock.Now()); err != nil {
		return model.Interpretation{}, err
	}
	if err := s.repo.Update(ctx, item); err != nil {
		return model.Interpretation{}, err
	}
	if err := auditState(ctx, s.audits, "interpretation", id, "submitted", actor, before, item); err != nil {
		return model.Interpretation{}, err
	}
	return item, nil
}

func (s *InterpretationService) UpdateNotes(ctx context.Context, id, notes, actor string) (model.Interpretation, error) {
	item, err := s.repo.Get(ctx, id)
	if err != nil {
		return model.Interpretation{}, err
	}
	before := item
	item.Notes = notes
	item.UpdatedAt = s.clock.Now()
	if err := s.repo.Update(ctx, item); err != nil {
		return model.Interpretation{}, err
	}
	if err := auditState(ctx, s.audits, "interpretation", id, "notes_updated", actor, before, item); err != nil {
		return model.Interpretation{}, err
	}
	return item, nil
}
