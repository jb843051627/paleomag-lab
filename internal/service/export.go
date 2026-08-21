package service

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jb843051627/paleomag-lab/internal/clock"
	"github.com/jb843051627/paleomag-lab/internal/model"
	"github.com/jb843051627/paleomag-lab/internal/repository"
)

type ExportService struct {
	exports         *repository.ExportRepository
	specimens       *repository.SpecimenRepository
	plans           *repository.PlanRepository
	runs            *repository.MeasurementRunRepository
	measurements    *repository.MeasurementRepository
	interpretations *repository.InterpretationRepository
	audits          *AuditService
	clock           clock.Clock
}

func (s *ExportService) Request(ctx context.Context, specimenID, format, requestedBy string) (model.ExportJob, error) {
	if _, err := s.specimens.Get(ctx, specimenID); err != nil {
		return model.ExportJob{}, err
	}
	job := model.ExportJob{ID: newID("export"), SpecimenID: specimenID, Format: strings.ToLower(format), RequestedBy: requestedBy, Status: model.ExportQueued, CreatedAt: s.clock.Now(), UpdatedAt: s.clock.Now()}
	if err := job.Validate(); err != nil {
		return model.ExportJob{}, err
	}
	if err := s.exports.Create(ctx, job); err != nil {
		return model.ExportJob{}, err
	}
	if err := auditState(ctx, s.audits, "export_job", job.ID, "queued", requestedBy, nil, job); err != nil {
		return model.ExportJob{}, err
	}
	return job, nil
}

func (s *ExportService) Get(ctx context.Context, id string) (model.ExportJob, error) {
	return s.exports.Get(ctx, id)
}

func (s *ExportService) List(ctx context.Context, specimenID string) ([]model.ExportJob, error) {
	return s.exports.ListBySpecimen(ctx, specimenID)
}

func (s *ExportService) Run(ctx context.Context, jobID string) (model.ExportJob, error) {
	job, err := s.exports.Get(ctx, jobID)
	if err != nil {
		return model.ExportJob{}, err
	}
	if err := job.Start(s.clock.Now()); err != nil {
		return model.ExportJob{}, err
	}
	if err := s.exports.Update(ctx, job); err != nil {
		return model.ExportJob{}, err
	}
	result, err := s.BuildReport(ctx, job.SpecimenID, job.Format)
	if err != nil {
		_ = job.Fail(err, s.clock.Now())
		_ = s.exports.Update(ctx, job)
		return model.ExportJob{}, err
	}
	if err := job.Complete(result, s.clock.Now()); err != nil {
		return model.ExportJob{}, err
	}
	if err := s.exports.Update(ctx, job); err != nil {
		return model.ExportJob{}, err
	}
	return job, nil
}

func (s *ExportService) BuildReport(ctx context.Context, specimenID, format string) (string, error) {
	specimen, err := s.specimens.Get(ctx, specimenID)
	if err != nil {
		return "", err
	}
	plans, err := s.plans.ListBySpecimen(ctx, specimenID)
	if err != nil {
		return "", err
	}
	runs, err := s.runs.ListBySpecimen(ctx, specimenID)
	if err != nil {
		return "", err
	}
	interpretations, err := s.interpretations.ListBySpecimen(ctx, specimenID, model.InterpretationApproved)
	if err != nil {
		return "", err
	}
	if len(interpretations) == 0 {
		return "", fmt.Errorf("%w: specimen %q has no approved interpretation", model.ErrState, specimenID)
	}
	if strings.EqualFold(format, "csv") {
		return buildCSV(specimen, plans, runs, interpretations)
	}
	report := struct {
		Specimen        model.Specimen         `json:"specimen"`
		Plans           []model.DemagPlan      `json:"plans"`
		Runs            []model.MeasurementRun `json:"runs"`
		Interpretations []model.Interpretation `json:"interpretations"`
	}{specimen, plans, runs, interpretations}
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal report: %w", err)
	}
	return string(data), nil
}

func buildCSV(specimen model.Specimen, plans []model.DemagPlan, runs []model.MeasurementRun, interpretations []model.Interpretation) (string, error) {
	var builder strings.Builder
	writer := csv.NewWriter(&builder)
	rows := [][]string{{"section", "id", "status", "value"}, {"specimen", specimen.ID, string(specimen.Status), specimen.Code}}
	for _, plan := range plans {
		rows = append(rows, []string{"plan", plan.ID, string(plan.Status), plan.Name})
	}
	for _, run := range runs {
		rows = append(rows, []string{"run", run.ID, string(run.Status), fmt.Sprintf("%.4f", run.QualityScore)})
	}
	for _, interpretation := range interpretations {
		rows = append(rows, []string{"interpretation", interpretation.ID, string(interpretation.Status), fmt.Sprintf("%.4f", interpretation.Confidence)})
	}
	if err := writer.WriteAll(rows); err != nil {
		return "", fmt.Errorf("write csv report: %w", err)
	}
	return builder.String(), nil
}
