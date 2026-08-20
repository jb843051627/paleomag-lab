package model

import (
	"fmt"
	"time"
)

type CalibrationRun struct {
	ID           string            `json:"id"`
	InstrumentID string            `json:"instrument_id"`
	Reference    string            `json:"reference"`
	OffsetX      float64           `json:"offset_x"`
	OffsetY      float64           `json:"offset_y"`
	OffsetZ      float64           `json:"offset_z"`
	Status       CalibrationStatus `json:"status"`
	ValidUntil   time.Time         `json:"valid_until"`
	StartedAt    time.Time         `json:"started_at"`
	CompletedAt  *time.Time        `json:"completed_at,omitempty"`
}

func (c CalibrationRun) Validate() error {
	if c.ID == "" || c.InstrumentID == "" || c.Reference == "" {
		return fmt.Errorf("%w: calibration identity is required", ErrInvalid)
	}
	if c.ValidUntil.Before(c.StartedAt) {
		return fmt.Errorf("%w: calibration expiry precedes start", ErrInvalid)
	}
	return nil
}

func (c CalibrationRun) IsUsable(now time.Time) bool {
	return c.Status == CalibrationValid && now.Before(c.ValidUntil)
}

func (c *CalibrationRun) Complete(valid bool, now time.Time, validUntil time.Time) error {
	if c.Status != CalibrationPending {
		return fmt.Errorf("%w: calibration already completed", ErrState)
	}
	c.CompletedAt = &now
	c.ValidUntil = validUntil
	if valid {
		c.Status = CalibrationValid
	} else {
		c.Status = CalibrationFailed
	}
	return nil
}
