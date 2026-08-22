package service

import (
	"context"
	"fmt"

	"github.com/jb843051627/paleomag-lab/internal/clock"
	"github.com/jb843051627/paleomag-lab/internal/model"
	"github.com/jb843051627/paleomag-lab/internal/repository"
)

type PlanService struct {
	repo      *repository.PlanRepository
	specimens *repository.SpecimenRepository
	audits    *AuditService
	clock     clock.Clock
}

type CreatePlanInput struct {
	SpecimenID string `json:"specimen_id"`
	Name       string `json:"name"`
	Method     string `json:"method"`
	Operator   string `json:"operator"`
}

func (s *PlanService) Create(ctx context.Context, input CreatePlanInput) (model.DemagPlan, error) {
	if _, err := s.specimens.Get(ctx, input.SpecimenID); err != nil {
		return model.DemagPlan{}, fmt.Errorf("load specimen for plan: %w", err)
	}
	plan := model.NewDemagPlan(newID("plan"), input.SpecimenID, input.Name, input.Method, input.Operator, s.clock.Now())
	if input.Name == "" || input.Method == "" || input.Operator == "" {
		return model.DemagPlan{}, fmt.Errorf("%w: plan metadata is required", model.ErrInvalid)
	}
	if err := s.repo.Create(ctx, plan); err != nil {
		return model.DemagPlan{}, err
	}
	if err := auditState(ctx, s.audits, "demag_plan", plan.ID, "created", input.Operator, nil, plan); err != nil {
		return model.DemagPlan{}, err
	}
	return plan, nil
}

func (s *PlanService) Get(ctx context.Context, id string) (model.DemagPlan, error) {
	return s.repo.Get(ctx, id)
}

func (s *PlanService) ListBySpecimen(ctx context.Context, specimenID string) ([]model.DemagPlan, error) {
	return s.repo.ListBySpecimen(ctx, specimenID)
}

func (s *PlanService) AddStep(ctx context.Context, planID string, field, temperature float64, note, actor string) (model.DemagPlan, error) {
	plan, err := s.repo.Get(ctx, planID)
	if err != nil {
		return model.DemagPlan{}, err
	}
	before := plan
	step := model.DemagStep{ID: newID("step"), Field: field, Temperature: temperature, Note: note}
	if err := plan.AddStep(step, s.clock.Now()); err != nil {
		return model.DemagPlan{}, err
	}
	step = plan.Steps[len(plan.Steps)-1]
	if err := s.repo.AddStep(ctx, step); err != nil {
		return model.DemagPlan{}, err
	}
	if err := s.repo.Update(ctx, plan); err != nil {
		return model.DemagPlan{}, err
	}
	if err := auditState(ctx, s.audits, "demag_plan", planID, "step_added", actor, before, plan); err != nil {
		return model.DemagPlan{}, err
	}
	return plan, nil
}

func (s *PlanService) Prepare(ctx context.Context, planID, actor string) (model.DemagPlan, error) {
	plan, err := s.repo.Get(ctx, planID)
	if err != nil {
		return model.DemagPlan{}, err
	}
	before := plan
	if err := plan.Prepare(s.clock.Now()); err != nil {
		return model.DemagPlan{}, err
	}
	if err := s.repo.Update(ctx, plan); err != nil {
		return model.DemagPlan{}, err
	}
	if err := auditState(ctx, s.audits, "demag_plan", planID, "prepared", actor, before, plan); err != nil {
		return model.DemagPlan{}, err
	}
	return plan, nil
}

func (s *PlanService) Start(ctx context.Context, planID, actor string) (model.DemagPlan, error) {
	plan, err := s.repo.Get(ctx, planID)
	if err != nil {
		return model.DemagPlan{}, err
	}
	before := plan
	if err := plan.Start(s.clock.Now()); err != nil {
		return model.DemagPlan{}, err
	}
	if err := s.repo.Update(ctx, plan); err != nil {
		return model.DemagPlan{}, err
	}
	if err := auditState(ctx, s.audits, "demag_plan", planID, "started", actor, before, plan); err != nil {
		return model.DemagPlan{}, err
	}
	return plan, nil
}

func (s *PlanService) Complete(ctx context.Context, planID, actor string) (model.DemagPlan, error) {
	plan, err := s.repo.Get(ctx, planID)
	if err != nil {
		return model.DemagPlan{}, err
	}
	before := plan
	if err := plan.Complete(s.clock.Now()); err != nil {
		return model.DemagPlan{}, err
	}
	if err := s.repo.Update(ctx, plan); err != nil {
		return model.DemagPlan{}, err
	}
	if err := auditState(ctx, s.audits, "demag_plan", planID, "completed", actor, before, plan); err != nil {
		return model.DemagPlan{}, err
	}
	return plan, nil
}

func (s *PlanService) Reject(ctx context.Context, planID, reason, actor string) (model.DemagPlan, error) {
	plan, err := s.repo.Get(ctx, planID)
	if err != nil {
		return model.DemagPlan{}, err
	}
	before := plan
	if err := plan.Reject(reason, s.clock.Now()); err != nil {
		return model.DemagPlan{}, err
	}
	if err := s.repo.Update(ctx, plan); err != nil {
		return model.DemagPlan{}, err
	}
	if err := auditState(ctx, s.audits, "demag_plan", planID, "rejected", actor, before, plan); err != nil {
		return model.DemagPlan{}, err
	}
	return plan, nil
}
