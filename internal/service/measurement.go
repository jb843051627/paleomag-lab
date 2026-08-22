package service

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/jb843051627/paleomag-lab/internal/clock"
	"github.com/jb843051627/paleomag-lab/internal/model"
	"github.com/jb843051627/paleomag-lab/internal/repository"
)

type MeasurementService struct {
	runs         *repository.MeasurementRunRepository
	measurements *repository.MeasurementRepository
	plans        *repository.PlanRepository
	specimens    *repository.SpecimenRepository
	instruments  *repository.InstrumentRepository
	calibrations *repository.CalibrationRepository
	audits       *AuditService
	quality      *QualityService
	clock        clock.Clock
}

type StartRunInput struct {
	SpecimenID    string `json:"specimen_id"`
	PlanID        string `json:"plan_id"`
	InstrumentID  string `json:"instrument_id"`
	CalibrationID string `json:"calibration_id"`
}

func (s *MeasurementService) StartRun(ctx context.Context, input StartRunInput, actor string) (model.MeasurementRun, error) {
	specimen, err := s.specimens.Get(ctx, input.SpecimenID)
	if err != nil {
		return model.MeasurementRun{}, fmt.Errorf("load specimen: %w", err)
	}
	plan, err := s.plans.Get(ctx, input.PlanID)
	if err != nil {
		return model.MeasurementRun{}, fmt.Errorf("load plan: %w", err)
	}
	if plan.SpecimenID != specimen.ID {
		return model.MeasurementRun{}, fmt.Errorf("%w: plan belongs to another specimen", model.ErrConflict)
	}
	if plan.Status != model.PlanReady {
		return model.MeasurementRun{}, fmt.Errorf("%w: plan already has an active or finished run", model.ErrState)
	}
	instrument, err := s.instruments.Get(ctx, input.InstrumentID)
	if err != nil {
		return model.MeasurementRun{}, fmt.Errorf("load instrument: %w", err)
	}
	var calibration model.CalibrationRun
	if input.CalibrationID != "" {
		calibration, err = s.calibrations.Get(ctx, input.CalibrationID)
	} else {
		calibration, err = s.calibrations.LatestUsable(ctx, instrument.ID, s.clock.Now().Format(time.RFC3339Nano))
	}
	if err != nil {
		return model.MeasurementRun{}, fmt.Errorf("load calibration: %w", err)
	}
	if calibration.InstrumentID != instrument.ID || !calibration.IsUsable(s.clock.Now()) {
		return model.MeasurementRun{}, fmt.Errorf("%w: calibration cannot be used", model.ErrCalibration)
	}
	step := plan.NextStep()
	if step == nil || !instrument.SupportsField(step.Field) {
		return model.MeasurementRun{}, fmt.Errorf("%w: plan step is outside instrument range", model.ErrInvalid)
	}
	plan.Status = model.PlanRunning
	plan.UpdatedAt = s.clock.Now()
	if err := s.plans.Update(ctx, plan); err != nil {
		return model.MeasurementRun{}, err
	}
	if specimen.Status == model.SpecimenOriented {
		specimen.Status = model.SpecimenInProgress
		specimen.UpdatedAt = s.clock.Now()
		if err := s.specimens.Update(ctx, specimen); err != nil {
			return model.MeasurementRun{}, err
		}
	}
	run := model.MeasurementRun{ID: newID("run"), SpecimenID: specimen.ID, PlanID: plan.ID, InstrumentID: instrument.ID, CalibrationID: calibration.ID, Status: model.RunCollecting, StartedAt: s.clock.Now()}
	if err := s.runs.Create(ctx, run); err != nil {
		return model.MeasurementRun{}, err
	}
	if err := auditState(ctx, s.audits, "measurement_run", run.ID, "started", actor, nil, run); err != nil {
		return model.MeasurementRun{}, err
	}
	return run, nil
}

func (s *MeasurementService) GetRun(ctx context.Context, id string) (model.MeasurementRun, error) {
	return s.runs.Get(ctx, id)
}

func (s *MeasurementService) Record(ctx context.Context, item model.Measurement, actor string) error {
	if err := item.Validate(); err != nil {
		return err
	}
	run, err := s.runs.Get(ctx, item.RunID)
	if err != nil {
		return err
	}
	if run.Status != model.RunCollecting {
		return fmt.Errorf("%w: run is not collecting", model.ErrState)
	}
	plan, err := s.plans.Get(ctx, run.PlanID)
	if err != nil {
		return err
	}
	if item.Step > len(plan.Steps) {
		return fmt.Errorf("%w: measurement step %d does not exist", model.ErrInvalid, item.Step)
	}
	expected := plan.Steps[item.Step-1]
	if math.Abs(expected.Field-item.Field) > 0.0001 {
		return fmt.Errorf("%w: measurement field does not match plan step", model.ErrInvalid)
	}
	if item.Quality == "" {
		item.Quality = model.QualityPending
	}
	if err := s.measurements.Create(ctx, item); err != nil {
		return err
	}
	return auditState(ctx, s.audits, "measurement_run", run.ID, "measurement_recorded", actor, nil, item)
}

func (s *MeasurementService) RecordBatch(ctx context.Context, items []model.Measurement, actor string) (int, error) {
	count := 0
	for _, item := range items {
		if err := ctx.Err(); err != nil {
			return count, fmt.Errorf("%w: after %d measurements", model.ErrCancelled, count)
		}
		if err := s.Record(ctx, item, actor); err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}

func (s *MeasurementService) List(ctx context.Context, filter model.MeasurementFilter) ([]model.Measurement, error) {
	return s.measurements.ListByRun(ctx, filter)
}

func (s *MeasurementService) CompleteRun(ctx context.Context, runID, actor string) (model.MeasurementRun, model.MeasurementSummary, error) {
	run, err := s.runs.Get(ctx, runID)
	if err != nil {
		return model.MeasurementRun{}, model.MeasurementSummary{}, err
	}
	if run.Status != model.RunCollecting {
		return model.MeasurementRun{}, model.MeasurementSummary{}, fmt.Errorf("%w: run is not collecting", model.ErrState)
	}
	items, err := s.measurements.ListByRun(ctx, model.MeasurementFilter{RunID: runID})
	if err != nil {
		return model.MeasurementRun{}, model.MeasurementSummary{}, err
	}
	if len(items) == 0 {
		return model.MeasurementRun{}, model.MeasurementSummary{}, fmt.Errorf("%w: run has no measurements", model.ErrInvalid)
	}
	if _, err := s.EvaluateQuality(ctx, runID); err != nil {
		return model.MeasurementRun{}, model.MeasurementSummary{}, err
	}
	items, err = s.measurements.ListByRun(ctx, model.MeasurementFilter{RunID: runID})
	if err != nil {
		return model.MeasurementRun{}, model.MeasurementSummary{}, err
	}
	summary := model.SummarizeMeasurements(items)
	run.Status = model.RunEvaluated
	run.CompletedAt = timePtr(s.clock.Now())
	run.QualityScore = float64(summary.Good) / float64(summary.Count)
	if err := s.runs.Update(ctx, run); err != nil {
		return model.MeasurementRun{}, model.MeasurementSummary{}, err
	}
	if err := auditState(ctx, s.audits, "measurement_run", runID, "evaluated", actor, nil, run); err != nil {
		return model.MeasurementRun{}, model.MeasurementSummary{}, err
	}
	return run, summary, nil
}

func (s *MeasurementService) AcceptRun(ctx context.Context, runID, actor string) (model.MeasurementRun, error) {
	run, err := s.runs.Get(ctx, runID)
	if err != nil {
		return model.MeasurementRun{}, err
	}
	if run.Status != model.RunEvaluated || run.QualityScore < 0.6 {
		return model.MeasurementRun{}, fmt.Errorf("%w: run quality is insufficient", model.ErrState)
	}
	run.Status = model.RunAccepted
	if err := s.runs.Update(ctx, run); err != nil {
		return model.MeasurementRun{}, err
	}
	if err := auditState(ctx, s.audits, "measurement_run", runID, "accepted", actor, nil, run); err != nil {
		return model.MeasurementRun{}, err
	}
	return run, nil
}

func (s *MeasurementService) MarkRework(ctx context.Context, runID, actor string) (model.MeasurementRun, error) {
	run, err := s.runs.Get(ctx, runID)
	if err != nil {
		return model.MeasurementRun{}, err
	}
	if run.Status != model.RunEvaluated {
		return model.MeasurementRun{}, fmt.Errorf("%w: run must be evaluated before rework", model.ErrState)
	}
	run.Status = model.RunRework
	if err := s.runs.Update(ctx, run); err != nil {
		return model.MeasurementRun{}, err
	}
	if err := auditState(ctx, s.audits, "measurement_run", runID, "rework", actor, nil, run); err != nil {
		return model.MeasurementRun{}, err
	}
	return run, nil
}

func (s *MeasurementService) EvaluateQuality(ctx context.Context, runID string) (model.MeasurementSummary, error) {
	if s.quality == nil {
		return model.MeasurementSummary{}, fmt.Errorf("%w: quality service is not configured", model.ErrState)
	}
	if _, err := s.quality.Assess(ctx, runID); err != nil {
		return model.MeasurementSummary{}, err
	}
	items, err := s.measurements.ListByRun(ctx, model.MeasurementFilter{RunID: runID})
	if err != nil {
		return model.MeasurementSummary{}, err
	}
	return model.SummarizeMeasurements(items), nil
}

func (s *MeasurementService) Count(ctx context.Context, runID string) (int, error) {
	return s.measurements.CountByRun(ctx, runID)
}

func timePtr(value time.Time) *time.Time { return &value }
