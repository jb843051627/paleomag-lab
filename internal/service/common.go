package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/jb843051627/paleomag-lab/internal/clock"
	"github.com/jb843051627/paleomag-lab/internal/model"
	"github.com/jb843051627/paleomag-lab/internal/repository"
)

type Services struct {
	Specimens       *SpecimenService
	Plans           *PlanService
	Instruments     *InstrumentService
	Measurements    *MeasurementService
	Interpretations *InterpretationService
	Reviews         *ReviewService
	Exports         *ExportService
	Audits          *AuditService
	Insights        *InsightsService
	Quality         *QualityService
	Protocols       *ProtocolService
	Reports         *ReportService
	Analysis        *AnalysisService
}

func NewAll(repos *repository.Repositories, clk clock.Clock) *Services {
	audits := &AuditService{repo: repos.Audits, clock: clk}
	quality, err := NewQualityService(repos.Measurements, model.DefaultQualityPolicy())
	if err != nil {
		panic(err)
	}
	return &Services{
		Audits:          audits,
		Specimens:       &SpecimenService{repo: repos.Specimens, audits: audits, clock: clk},
		Plans:           &PlanService{repo: repos.Plans, specimens: repos.Specimens, audits: audits, clock: clk},
		Instruments:     &InstrumentService{repo: repos.Instruments, calibrations: repos.Calibrations, audits: audits, clock: clk},
		Measurements:    &MeasurementService{runs: repos.MeasurementRuns, measurements: repos.Measurements, plans: repos.Plans, specimens: repos.Specimens, instruments: repos.Instruments, calibrations: repos.Calibrations, audits: audits, quality: quality, clock: clk},
		Interpretations: &InterpretationService{repo: repos.Interpretations, runs: repos.MeasurementRuns, measurements: repos.Measurements, specimens: repos.Specimens, audits: audits, clock: clk},
		Reviews:         &ReviewService{reviews: repos.Reviews, interpretations: repos.Interpretations, specimens: repos.Specimens, audits: audits, clock: clk},
		Exports:         &ExportService{exports: repos.Exports, specimens: repos.Specimens, plans: repos.Plans, runs: repos.MeasurementRuns, measurements: repos.Measurements, interpretations: repos.Interpretations, audits: audits, clock: clk},
		Insights:        &InsightsService{queries: repos.Queries, aggregate: repos.Aggregates},
		Quality:         quality,
		Protocols:       &ProtocolService{plans: repos.Plans, instruments: repos.Instruments},
		Reports:         &ReportService{specimens: repos.Specimens, plans: repos.Plans, runs: repos.MeasurementRuns, interpretations: repos.Interpretations, clock: clk},
		Analysis:        &AnalysisService{measurements: repos.Measurements, plans: repos.Plans},
	}
}

var idSequence atomic.Uint64

func newID(prefix string) string {
	sequence := idSequence.Add(1)
	return fmt.Sprintf("%s-%d-%d", prefix, time.Now().UnixNano(), sequence)
}

func auditState(ctx context.Context, audits *AuditService, entityType, entityID, action, actor string, before, after any) error {
	beforeBytes, err := json.Marshal(before)
	if err != nil {
		return fmt.Errorf("marshal audit before state: %w", err)
	}
	afterBytes, err := json.Marshal(after)
	if err != nil {
		return fmt.Errorf("marshal audit after state: %w", err)
	}
	event := model.AuditEvent{ID: newID("audit"), EntityType: entityType, EntityID: entityID, Action: action, Actor: actor, Before: string(beforeBytes), After: string(afterBytes), CreatedAt: audits.clock.Now()}
	return audits.repo.Create(ctx, event)
}
