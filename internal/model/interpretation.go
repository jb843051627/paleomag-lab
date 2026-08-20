package model

import (
	"fmt"
	"math"
	"time"
)

type Interpretation struct {
	ID          string               `json:"id"`
	SpecimenID  string               `json:"specimen_id"`
	RunID       string               `json:"run_id"`
	StartStep   int                  `json:"start_step"`
	EndStep     int                  `json:"end_step"`
	Declination float64              `json:"declination"`
	Inclination float64              `json:"inclination"`
	Intensity   float64              `json:"intensity"`
	Confidence  float64              `json:"confidence"`
	Status      InterpretationStatus `json:"status"`
	Notes       string               `json:"notes"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
}

func (i Interpretation) Validate() error {
	if i.ID == "" || i.SpecimenID == "" || i.RunID == "" {
		return fmt.Errorf("%w: interpretation identity is required", ErrInvalid)
	}
	if i.StartStep < 1 || i.EndStep < i.StartStep {
		return fmt.Errorf("%w: interpretation step range is invalid", ErrInvalid)
	}
	if i.Declination < 0 || i.Declination >= 360 || i.Inclination < -90 || i.Inclination > 90 {
		return fmt.Errorf("%w: interpretation direction is invalid", ErrInvalid)
	}
	if i.Intensity <= 0 || math.IsNaN(i.Confidence) || i.Confidence < 0 || i.Confidence > 1 {
		return fmt.Errorf("%w: interpretation score is invalid", ErrInvalid)
	}
	return nil
}

func (i *Interpretation) Submit(now time.Time) error {
	if i.Status != InterpretationDraft && i.Status != InterpretationRejected {
		return fmt.Errorf("%w: interpretation cannot be submitted from %s", ErrState, i.Status)
	}
	if err := i.Validate(); err != nil {
		return err
	}
	i.Status = InterpretationPending
	i.UpdatedAt = now
	return nil
}

func (i *Interpretation) ApplyDecision(decision ReviewDecision, now time.Time) error {
	if i.Status != InterpretationPending {
		return fmt.Errorf("%w: interpretation is not awaiting review", ErrState)
	}
	switch decision {
	case ReviewApprove:
		i.Status = InterpretationApproved
	case ReviewReject, ReviewRework:
		i.Status = InterpretationRejected
	default:
		return fmt.Errorf("%w: unsupported review decision", ErrInvalid)
	}
	i.UpdatedAt = now
	return nil
}
