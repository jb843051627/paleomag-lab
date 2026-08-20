package service

import (
	"context"
	"fmt"
	"time"

	"github.com/jb843051627/paleomag-lab/internal/clock"
	"github.com/jb843051627/paleomag-lab/internal/model"
	"github.com/jb843051627/paleomag-lab/internal/repository"
)

type InstrumentService struct {
	repo         *repository.InstrumentRepository
	calibrations *repository.CalibrationRepository
	audits       *AuditService
	clock        clock.Clock
}

type RegisterInstrumentInput struct {
	Name     string  `json:"name"`
	Serial   string  `json:"serial"`
	Kind     string  `json:"kind"`
	MinField float64 `json:"min_field"`
	MaxField float64 `json:"max_field"`
}

func (s *InstrumentService) Register(ctx context.Context, input RegisterInstrumentInput) (model.Instrument, error) {
	item := model.Instrument{ID: newID("instrument"), Name: input.Name, Serial: input.Serial, Kind: input.Kind, MinField: input.MinField, MaxField: input.MaxField, Status: model.InstrumentReady, CreatedAt: s.clock.Now(), UpdatedAt: s.clock.Now()}
	if err := item.Validate(); err != nil {
		return model.Instrument{}, err
	}
	if err := s.repo.Create(ctx, item); err != nil {
		return model.Instrument{}, err
	}
	if err := auditState(ctx, s.audits, "instrument", item.ID, "registered", input.Serial, nil, item); err != nil {
		return model.Instrument{}, err
	}
	return item, nil
}

func (s *InstrumentService) Get(ctx context.Context, id string) (model.Instrument, error) {
	return s.repo.Get(ctx, id)
}

func (s *InstrumentService) List(ctx context.Context, status model.InstrumentStatus) ([]model.Instrument, error) {
	return s.repo.List(ctx, status)
}

func (s *InstrumentService) BeginCalibration(ctx context.Context, instrumentID, reference string, actor string) (model.CalibrationRun, error) {
	instrument, err := s.repo.Get(ctx, instrumentID)
	if err != nil {
		return model.CalibrationRun{}, err
	}
	before := instrument
	if err := instrument.BeginCalibration(s.clock.Now()); err != nil {
		return model.CalibrationRun{}, err
	}
	if err := s.repo.Update(ctx, instrument); err != nil {
		return model.CalibrationRun{}, err
	}
	if reference == "" {
		return model.CalibrationRun{}, fmt.Errorf("%w: calibration reference is required", model.ErrInvalid)
	}
	if err := s.repo.Update(ctx, instrument); err != nil {
		return model.CalibrationRun{}, err
	}
	item := model.CalibrationRun{ID: newID("cal"), InstrumentID: instrumentID, Reference: reference, Status: model.CalibrationPending, StartedAt: s.clock.Now(), ValidUntil: s.clock.Now().Add(24 * time.Hour)}
	if err := s.calibrations.Create(ctx, item); err != nil {
		return model.CalibrationRun{}, err
	}
	if err := auditState(ctx, s.audits, "instrument", instrumentID, "calibration_started", actor, before, instrument); err != nil {
		return model.CalibrationRun{}, err
	}
	return item, nil
}

func (s *InstrumentService) CompleteCalibration(ctx context.Context, calibrationID string, valid bool, offsets [3]float64, actor string) (model.CalibrationRun, error) {
	item, err := s.calibrations.Get(ctx, calibrationID)
	if err != nil {
		return model.CalibrationRun{}, err
	}
	instrument, err := s.repo.Get(ctx, item.InstrumentID)
	if err != nil {
		return model.CalibrationRun{}, err
	}
	before := item
	item.OffsetX, item.OffsetY, item.OffsetZ = offsets[0], offsets[1], offsets[2]
	if err := item.Complete(valid, s.clock.Now(), s.clock.Now().Add(24*time.Hour)); err != nil {
		return model.CalibrationRun{}, err
	}
	if err := s.calibrations.Update(ctx, item); err != nil {
		return model.CalibrationRun{}, err
	}
	if err := instrument.FinishCalibration(valid, s.clock.Now()); err != nil {
		return model.CalibrationRun{}, err
	}
	if err := s.repo.Update(ctx, instrument); err != nil {
		return model.CalibrationRun{}, err
	}
	if err := auditState(ctx, s.audits, "calibration", calibrationID, "completed", actor, before, item); err != nil {
		return model.CalibrationRun{}, err
	}
	return item, nil
}

func (s *InstrumentService) LatestCalibration(ctx context.Context, instrumentID string) (model.CalibrationRun, error) {
	return s.calibrations.LatestUsable(ctx, instrumentID, s.clock.Now().Format(time.RFC3339Nano))
}

func (s *InstrumentService) CalibrationHistory(ctx context.Context, instrumentID string) ([]model.CalibrationRun, error) {
	return s.calibrations.ListByInstrument(ctx, instrumentID)
}
