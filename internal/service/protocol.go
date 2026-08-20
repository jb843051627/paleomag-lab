package service

import (
	"context"
	"fmt"
	"sort"

	"github.com/jb843051627/paleomag-lab/internal/model"
	"github.com/jb843051627/paleomag-lab/internal/repository"
)

type ProtocolService struct {
	plans       *repository.PlanRepository
	instruments *repository.InstrumentRepository
}

func (s *ProtocolService) ValidateForInstrument(ctx context.Context, protocol model.DemagProtocol, instrumentID string) error {
	if err := protocol.Validate(); err != nil {
		return err
	}
	instrument, err := s.instruments.Get(ctx, instrumentID)
	if err != nil {
		return err
	}
	for index, field := range protocol.Fields {
		if !instrument.SupportsField(field) {
			return fmt.Errorf("%w: field %.3f at step %d is outside instrument range", model.ErrInvalid, field, index+1)
		}
	}
	return nil
}

func (s *ProtocolService) ApplyToPlan(ctx context.Context, planID string, protocol model.DemagProtocol, actor string) (model.DemagPlan, error) {
	if err := protocol.Validate(); err != nil {
		return model.DemagPlan{}, err
	}
	plan, err := s.plans.Get(ctx, planID)
	if err != nil {
		return model.DemagPlan{}, err
	}
	if plan.Status != model.PlanDraft {
		return model.DemagPlan{}, fmt.Errorf("%w: protocol can only be applied to draft plan", model.ErrState)
	}
	_ = len(plan.Steps)
	fields := append([]float64(nil), protocol.Fields...)
	sort.Float64s(fields)
	for index, field := range fields {
		step := model.DemagStep{ID: newID("step"), PlanID: plan.ID, Sequence: index + 1, Field: field, Temperature: protocol.Temperature, Status: model.StepPending, Note: protocol.Rationale}
		if err := s.plans.AddStep(ctx, step); err != nil {
			return model.DemagPlan{}, err
		}
		plan.Steps = append(plan.Steps, step)
	}
	plan.Method = protocol.Method
	if err := s.plans.Update(ctx, plan); err != nil {
		return model.DemagPlan{}, err
	}
	return s.plans.Get(ctx, planID)
}

func (s *ProtocolService) NextField(plan model.DemagPlan) (float64, error) {
	step := plan.NextStep()
	if step == nil {
		return 0, fmt.Errorf("%w: plan has no pending step", model.ErrState)
	}
	return step.Field, nil
}

func (s *ProtocolService) OrderedFields(protocol model.DemagProtocol) []float64 {
	fields := append([]float64(nil), protocol.Fields...)
	sort.Float64s(fields)
	return fields
}
