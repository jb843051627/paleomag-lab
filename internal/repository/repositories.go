package repository

import (
	"github.com/jb843051627/paleomag-lab/internal/store"
)

type Repositories struct {
	Specimens       *SpecimenRepository
	Plans           *PlanRepository
	Instruments     *InstrumentRepository
	Calibrations    *CalibrationRepository
	MeasurementRuns *MeasurementRunRepository
	Measurements    *MeasurementRepository
	Interpretations *InterpretationRepository
	Reviews         *ReviewRepository
	Exports         *ExportRepository
	Audits          *AuditRepository
	Queries         *QueryRepository
	Aggregates      *AggregateRepository
	Transactions    *TransactionRepository
}

func NewAll(db *store.DB) *Repositories {
	return &Repositories{
		Specimens:       &SpecimenRepository{db: db.SQL},
		Plans:           &PlanRepository{db: db.SQL},
		Instruments:     &InstrumentRepository{db: db.SQL},
		Calibrations:    &CalibrationRepository{db: db.SQL},
		MeasurementRuns: &MeasurementRunRepository{db: db.SQL},
		Measurements:    &MeasurementRepository{db: db.SQL},
		Interpretations: &InterpretationRepository{db: db.SQL},
		Reviews:         &ReviewRepository{db: db.SQL},
		Exports:         &ExportRepository{db: db.SQL},
		Audits:          &AuditRepository{db: db.SQL},
		Queries:         NewQueryRepository(db.SQL),
		Aggregates:      NewAggregateRepository(db.SQL),
		Transactions:    NewTransactionRepository(db.SQL),
	}
}
