package model

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

type DemagPlan struct {
	ID         string      `json:"id"`
	SpecimenID string      `json:"specimen_id"`
	Name       string      `json:"name"`
	Method     string      `json:"method"`
	Status     PlanStatus  `json:"status"`
	Operator   string      `json:"operator"`
	Steps      []DemagStep `json:"steps"`
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
}

type DemagStep struct {
	ID          string     `json:"id"`
	PlanID      string     `json:"plan_id"`
	Sequence    int        `json:"sequence"`
	Field       float64    `json:"field"`
	Temperature float64    `json:"temperature"`
	Status      StepStatus `json:"status"`
	Note        string     `json:"note"`
}

func NewDemagPlan(id, specimenID, name, method, operator string, now time.Time) DemagPlan {
	return DemagPlan{ID: id, SpecimenID: specimenID, Name: strings.TrimSpace(name), Method: strings.TrimSpace(method), Status: PlanDraft, Operator: strings.TrimSpace(operator), CreatedAt: now, UpdatedAt: now, Steps: []DemagStep{}}
}

func (p DemagPlan) Validate() error {
	if p.ID == "" || p.SpecimenID == "" || p.Name == "" || p.Method == "" || p.Operator == "" {
		return fmt.Errorf("%w: plan identity and method are required", ErrInvalid)
	}
	if len(p.Steps) == 0 {
		return fmt.Errorf("%w: plan needs at least one step", ErrInvalid)
	}
	copySteps := append([]DemagStep(nil), p.Steps...)
	sort.Slice(copySteps, func(i, j int) bool { return copySteps[i].Sequence < copySteps[j].Sequence })
	for i, step := range copySteps {
		if step.Sequence != i+1 || step.Field < 0 || step.Temperature < -273.15 {
			return fmt.Errorf("%w: invalid demagnetization step %d", ErrInvalid, step.Sequence)
		}
		if i > 0 && step.Field <= copySteps[i-1].Field {
			return fmt.Errorf("%w: field values must increase", ErrInvalid)
		}
	}
	return nil
}

func (p *DemagPlan) AddStep(step DemagStep, now time.Time) error {
	if p.Status != PlanDraft {
		return fmt.Errorf("%w: steps can only be added to draft plan", ErrState)
	}
	step.PlanID = p.ID
	step.Sequence = len(p.Steps) + 1
	step.Status = StepPending
	p.Steps = append(p.Steps, step)
	p.UpdatedAt = now
	return nil
}

func (p *DemagPlan) Prepare(now time.Time) error {
	if p.Status != PlanDraft {
		return fmt.Errorf("%w: only draft plan can be prepared", ErrState)
	}
	if err := p.Validate(); err != nil {
		return err
	}
	p.Status = PlanReady
	p.UpdatedAt = now
	return nil
}

func (p *DemagPlan) Start(now time.Time) error {
	if p.Status != PlanReady {
		return fmt.Errorf("%w: plan must be ready before start", ErrState)
	}
	p.Status = PlanRunning
	p.UpdatedAt = now
	return nil
}

func (p *DemagPlan) Complete(now time.Time) error {
	if p.Status != PlanRunning {
		return fmt.Errorf("%w: plan is not running", ErrState)
	}
	for i := range p.Steps {
		if p.Steps[i].Status == StepPending {
			return fmt.Errorf("%w: pending step %d remains", ErrState, p.Steps[i].Sequence)
		}
	}
	p.Status = PlanCompleted
	p.UpdatedAt = now
	return nil
}

func (p *DemagPlan) Reject(reason string, now time.Time) error {
	if p.Status != PlanDraft && p.Status != PlanReady && p.Status != PlanRunning {
		return fmt.Errorf("%w: plan cannot be rejected from %s", ErrState, p.Status)
	}
	if strings.TrimSpace(reason) == "" {
		return fmt.Errorf("%w: rejection reason is required", ErrInvalid)
	}
	p.Status = PlanRejected
	p.UpdatedAt = now
	return nil
}

func (p DemagPlan) NextStep() *DemagStep {
	for i := range p.Steps {
		if p.Steps[i].Status == StepPending {
			step := p.Steps[i]
			return &step
		}
	}
	return nil
}
