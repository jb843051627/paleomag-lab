package service

import (
	"context"
	"fmt"
	"time"

	"github.com/jb843051627/paleomag-lab/internal/clock"
	"github.com/jb843051627/paleomag-lab/internal/model"
	"github.com/jb843051627/paleomag-lab/internal/repository"
)

type ReportService struct {
	specimens       *repository.SpecimenRepository
	plans           *repository.PlanRepository
	runs            *repository.MeasurementRunRepository
	interpretations *repository.InterpretationRepository
	clock           clock.Clock
}

func (s *ReportService) Snapshot(ctx context.Context, specimenID string) (model.ReportSnapshot, error) {
	specimen, err := s.specimens.Get(ctx, specimenID)
	if err != nil {
		return model.ReportSnapshot{}, err
	}
	plans, err := s.plans.ListBySpecimen(ctx, specimenID)
	if err != nil {
		return model.ReportSnapshot{}, err
	}
	runs, err := s.runs.ListBySpecimen(ctx, specimenID)
	if err != nil {
		return model.ReportSnapshot{}, err
	}
	interpretations, err := s.interpretations.ListBySpecimen(ctx, specimenID, model.InterpretationApproved)
	if err != nil {
		return model.ReportSnapshot{}, err
	}
	if len(interpretations) == 0 {
		return model.ReportSnapshot{}, fmt.Errorf("%w: no approved interpretation", model.ErrState)
	}
	snapshot := model.ReportSnapshot{Specimen: specimen, Plans: plans, Runs: runs, Interpretation: interpretations[0], GeneratedAt: s.clock.Now(), Warnings: []string{}}
	if specimen.Status != model.SpecimenArchived {
		snapshot.AddWarning("specimen has not been archived")
	}
	if snapshot.AcceptedRunCount() == 0 {
		snapshot.AddWarning("no accepted measurement run")
	}
	if err := snapshot.Validate(); err != nil {
		return model.ReportSnapshot{}, err
	}
	return snapshot, nil
}

func (s *ReportService) RefreshGeneratedAt(snapshot *model.ReportSnapshot) {
	if snapshot != nil {
		snapshot.GeneratedAt = s.clock.Now()
	}
}

func (s *ReportService) IsStale(snapshot model.ReportSnapshot, maxAge time.Duration) bool {
	if maxAge <= 0 {
		maxAge = 24 * time.Hour
	}
	return s.clock.Now().Sub(snapshot.GeneratedAt) > maxAge
}

func (s *ReportService) ValidateWarnings(snapshot model.ReportSnapshot) error {
	if err := snapshot.Validate(); err != nil {
		return err
	}
	if snapshot.HasWarnings() {
		return fmt.Errorf("%w: report contains %d warnings", model.ErrState, len(snapshot.Warnings))
	}
	return nil
}
