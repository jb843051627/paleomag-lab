package model

type SpecimenStatus string

const (
	SpecimenRegistered  SpecimenStatus = "registered"
	SpecimenOriented    SpecimenStatus = "oriented"
	SpecimenInProgress  SpecimenStatus = "in_progress"
	SpecimenInterpreted SpecimenStatus = "interpreted"
	SpecimenArchived    SpecimenStatus = "archived"
)

type PlanStatus string

const (
	PlanDraft     PlanStatus = "draft"
	PlanReady     PlanStatus = "ready"
	PlanRunning   PlanStatus = "running"
	PlanCompleted PlanStatus = "completed"
	PlanRejected  PlanStatus = "rejected"
)

type StepStatus string

const (
	StepPending StepStatus = "pending"
	StepDone    StepStatus = "done"
	StepOmitted StepStatus = "omitted"
)

type InstrumentStatus string

const (
	InstrumentReady       InstrumentStatus = "ready"
	InstrumentCalibrating InstrumentStatus = "calibrating"
	InstrumentRetired     InstrumentStatus = "retired"
)

type CalibrationStatus string

const (
	CalibrationPending CalibrationStatus = "pending"
	CalibrationValid   CalibrationStatus = "valid"
	CalibrationExpired CalibrationStatus = "expired"
	CalibrationFailed  CalibrationStatus = "failed"
)

type RunStatus string

const (
	RunCreated    RunStatus = "created"
	RunCollecting RunStatus = "collecting"
	RunEvaluated  RunStatus = "evaluated"
	RunAccepted   RunStatus = "accepted"
	RunRework     RunStatus = "rework"
)

type QualityFlag string

const (
	QualityPending QualityFlag = "pending"
	QualityGood    QualityFlag = "good"
	QualityReview  QualityFlag = "review"
	QualityBad     QualityFlag = "bad"
)

type InterpretationStatus string

const (
	InterpretationDraft    InterpretationStatus = "draft"
	InterpretationPending  InterpretationStatus = "pending_review"
	InterpretationApproved InterpretationStatus = "approved"
	InterpretationRejected InterpretationStatus = "rejected"
)

type ReviewDecision string

const (
	ReviewApprove ReviewDecision = "approve"
	ReviewReject  ReviewDecision = "reject"
	ReviewRework  ReviewDecision = "rework"
)

type ExportStatus string

const (
	ExportQueued   ExportStatus = "queued"
	ExportRunning  ExportStatus = "running"
	ExportComplete ExportStatus = "completed"
	ExportFailed   ExportStatus = "failed"
)
