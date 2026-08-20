package model

import (
	"fmt"
	"math"
	"time"
)

type MeasurementRun struct {
	ID            string     `json:"id"`
	SpecimenID    string     `json:"specimen_id"`
	PlanID        string     `json:"plan_id"`
	InstrumentID  string     `json:"instrument_id"`
	CalibrationID string     `json:"calibration_id"`
	Status        RunStatus  `json:"status"`
	StartedAt     time.Time  `json:"started_at"`
	CompletedAt   *time.Time `json:"completed_at,omitempty"`
	QualityScore  float64    `json:"quality_score"`
}

type Measurement struct {
	ID         string      `json:"id"`
	RunID      string      `json:"run_id"`
	Step       int         `json:"step"`
	Field      float64     `json:"field"`
	X          float64     `json:"x"`
	Y          float64     `json:"y"`
	Z          float64     `json:"z"`
	Intensity  float64     `json:"intensity"`
	Quality    QualityFlag `json:"quality"`
	MeasuredAt time.Time   `json:"measured_at"`
}

func (m Measurement) Validate() error {
	if m.ID == "" || m.RunID == "" || m.Step < 1 {
		return fmt.Errorf("%w: measurement identity and step are required", ErrInvalid)
	}
	if m.Field < 0 || math.IsNaN(m.X) || math.IsNaN(m.Y) || math.IsNaN(m.Z) {
		return fmt.Errorf("%w: measurement values are invalid", ErrInvalid)
	}
	if m.Intensity <= 0 {
		return fmt.Errorf("%w: intensity must be positive", ErrInvalid)
	}
	return nil
}

func (m Measurement) VectorLength() float64 {
	return math.Sqrt(m.X*m.X + m.Y*m.Y + m.Z*m.Z)
}

func (m Measurement) IsOutlier(reference float64) bool {
	if reference <= 0 {
		return false
	}
	ratio := m.VectorLength() / reference
	return ratio < 0.25 || ratio > 4
}
