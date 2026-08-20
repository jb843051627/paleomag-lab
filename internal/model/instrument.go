package model

import (
	"fmt"
	"strings"
	"time"
)

type Instrument struct {
	ID               string           `json:"id"`
	Name             string           `json:"name"`
	Serial           string           `json:"serial"`
	Kind             string           `json:"kind"`
	MinField         float64          `json:"min_field"`
	MaxField         float64          `json:"max_field"`
	Status           InstrumentStatus `json:"status"`
	LastCalibratedAt *time.Time       `json:"last_calibrated_at,omitempty"`
	CreatedAt        time.Time        `json:"created_at"`
	UpdatedAt        time.Time        `json:"updated_at"`
}

func (i Instrument) Validate() error {
	if i.ID == "" || strings.TrimSpace(i.Name) == "" || strings.TrimSpace(i.Serial) == "" || strings.TrimSpace(i.Kind) == "" {
		return fmt.Errorf("%w: instrument identity is incomplete", ErrInvalid)
	}
	if i.MinField >= i.MaxField || i.MinField < 0 {
		return fmt.Errorf("%w: instrument field range is invalid", ErrInvalid)
	}
	return nil
}

func (i *Instrument) BeginCalibration(now time.Time) error {
	if i.Status != InstrumentReady {
		return fmt.Errorf("%w: instrument is not ready", ErrState)
	}
	i.Status = InstrumentCalibrating
	i.UpdatedAt = now
	return nil
}

func (i *Instrument) FinishCalibration(valid bool, now time.Time) error {
	if i.Status != InstrumentCalibrating {
		return fmt.Errorf("%w: instrument is not calibrating", ErrState)
	}
	if valid {
		i.Status = InstrumentReady
		i.LastCalibratedAt = &now
	} else {
		i.Status = InstrumentReady
	}
	i.UpdatedAt = now
	return nil
}

func (i Instrument) SupportsField(field float64) bool {
	return field >= i.MinField && field <= i.MaxField
}
