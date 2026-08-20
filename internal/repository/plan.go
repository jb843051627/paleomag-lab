package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"

	"github.com/jb843051627/paleomag-lab/internal/model"
)

type PlanRepository struct{ db DBLike }

func (r *PlanRepository) Create(ctx context.Context, plan model.DemagPlan) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO demag_plans(id, specimen_id, name, method, status, operator, created_at, updated_at) VALUES(?,?,?,?,?,?,?,?)`,
		plan.ID, plan.SpecimenID, plan.Name, plan.Method, plan.Status, plan.Operator, timeText(plan.CreatedAt), timeText(plan.UpdatedAt))
	if err != nil {
		return fmt.Errorf("create plan: %w", err)
	}
	return nil
}

func (r *PlanRepository) AddStep(ctx context.Context, step model.DemagStep) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO demag_steps(id, plan_id, sequence_no, field, temperature, status, note) VALUES(?,?,?,?,?,?,?)`,
		step.ID, step.PlanID, step.Sequence, step.Field, step.Temperature, step.Status, step.Note)
	if err != nil {
		return fmt.Errorf("create plan step: %w", err)
	}
	return nil
}

func (r *PlanRepository) Get(ctx context.Context, id string) (model.DemagPlan, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, specimen_id, name, method, status, operator, created_at, updated_at FROM demag_plans WHERE id=?`, id)
	var p model.DemagPlan
	var status, created, updated string
	if err := row.Scan(&p.ID, &p.SpecimenID, &p.Name, &p.Method, &status, &p.Operator, &created, &updated); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.DemagPlan{}, fmt.Errorf("%w: plan %s", model.ErrNotFound, id)
		}
		return model.DemagPlan{}, fmt.Errorf("get plan: %w", err)
	}
	p.Status = model.PlanStatus(status)
	var err error
	p.CreatedAt, err = parseTime(created)
	if err != nil {
		return model.DemagPlan{}, err
	}
	p.UpdatedAt, err = parseTime(updated)
	if err != nil {
		return model.DemagPlan{}, err
	}
	p.Steps, err = r.listSteps(ctx, p.ID)
	if err != nil {
		return model.DemagPlan{}, err
	}
	return p, nil
}

func (r *PlanRepository) ListBySpecimen(ctx context.Context, specimenID string) ([]model.DemagPlan, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id FROM demag_plans WHERE specimen_id=? ORDER BY created_at DESC`, specimenID)
	if err != nil {
		return nil, fmt.Errorf("list specimen plans: %w", err)
	}
	ids := make([]string, 0)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return nil, fmt.Errorf("scan plan id: %w", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Close(); err != nil {
		return nil, fmt.Errorf("close plan list: %w", err)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate plan ids: %w", err)
	}
	plans := make([]model.DemagPlan, 0, len(ids))
	for _, id := range ids {
		plan, err := r.Get(ctx, id)
		if err != nil {
			return nil, err
		}
		plans = append(plans, plan)
	}
	return plans, nil
}

func (r *PlanRepository) Update(ctx context.Context, plan model.DemagPlan) error {
	result, err := r.db.ExecContext(ctx, `UPDATE demag_plans SET name=?, method=?, status=?, operator=?, updated_at=? WHERE id=?`, plan.Name, plan.Method, plan.Status, plan.Operator, timeText(plan.UpdatedAt), plan.ID)
	if err != nil {
		return fmt.Errorf("update plan: %w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("inspect plan update: %w", err)
	}
	if count == 0 {
		return fmt.Errorf("%w: plan %s", model.ErrNotFound, plan.ID)
	}
	return nil
}

func (r *PlanRepository) UpdateStep(ctx context.Context, step model.DemagStep) error {
	result, err := r.db.ExecContext(ctx, `UPDATE demag_steps SET field=?, temperature=?, status=?, note=? WHERE id=?`, step.Field, step.Temperature, step.Status, step.Note, step.ID)
	if err != nil {
		return fmt.Errorf("update plan step: %w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("inspect step update: %w", err)
	}
	if count == 0 {
		return fmt.Errorf("%w: step %s", model.ErrNotFound, step.ID)
	}
	return nil
}

func (r *PlanRepository) listSteps(ctx context.Context, planID string) ([]model.DemagStep, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, plan_id, sequence_no, field, temperature, status, note FROM demag_steps WHERE plan_id=?`, planID)
	if err != nil {
		return nil, fmt.Errorf("list plan steps: %w", err)
	}
	defer rows.Close()
	steps := make([]model.DemagStep, 0)
	for rows.Next() {
		var step model.DemagStep
		var status string
		if err := rows.Scan(&step.ID, &step.PlanID, &step.Sequence, &step.Field, &step.Temperature, &status, &step.Note); err != nil {
			return nil, fmt.Errorf("scan plan step: %w", err)
		}
		step.Status = model.StepStatus(status)
		steps = append(steps, step)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate plan steps: %w", err)
	}
	sort.Slice(steps, func(i, j int) bool { return steps[i].Sequence < steps[j].Sequence })
	return steps, nil
}
