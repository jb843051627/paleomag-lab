package model

import (
	"fmt"
	"math"
)

type CalibrationProfile struct {
	InstrumentID string  `json:"instrument_id"`
	Offset       Vector  `json:"offset"`
	Scale        Vector  `json:"scale"`
	NoiseFloor   float64 `json:"noise_floor"`
	Tolerance    float64 `json:"tolerance"`
}

func DefaultCalibrationProfile(instrumentID string) CalibrationProfile {
	return CalibrationProfile{InstrumentID: instrumentID, Scale: Vector{X: 1, Y: 1, Z: 1}, NoiseFloor: 0.001, Tolerance: 0.05}
}

func (p CalibrationProfile) Validate() error {
	if p.InstrumentID == "" {
		return fmt.Errorf("%w: calibration profile instrument is required", ErrInvalid)
	}
	if !p.Offset.IsFinite() || !p.Scale.IsFinite() || p.Scale.X <= 0 || p.Scale.Y <= 0 || p.Scale.Z <= 0 {
		return fmt.Errorf("%w: calibration profile scale is invalid", ErrInvalid)
	}
	if p.NoiseFloor < 0 || p.Tolerance <= 0 {
		return fmt.Errorf("%w: calibration thresholds are invalid", ErrInvalid)
	}
	return nil
}

func (p CalibrationProfile) Correct(raw Vector) (Vector, error) {
	if err := p.Validate(); err != nil {
		return Vector{}, err
	}
	if !raw.IsFinite() {
		return Vector{}, fmt.Errorf("%w: raw vector is not finite", ErrInvalid)
	}
	corrected := Vector{X: (raw.X - p.Offset.X) * p.Scale.X, Y: (raw.Y - p.Offset.Y) * p.Scale.Y, Z: (raw.Z - p.Offset.Z) * p.Scale.Z}
	if corrected.Length() < p.NoiseFloor {
		return Vector{}, fmt.Errorf("%w: vector is below noise floor", ErrInvalid)
	}
	return corrected, nil
}

func (p CalibrationProfile) Residual(expected, observed Vector) (Vector, error) {
	corrected, err := p.Correct(observed)
	if err != nil {
		return Vector{}, err
	}
	return corrected.Sub(expected), nil
}

type CalibrationAssessment struct {
	Profile         CalibrationProfile `json:"profile"`
	Samples         int                `json:"samples"`
	MeanResidual    float64            `json:"mean_residual"`
	MaximumResidual float64            `json:"maximum_residual"`
	Valid           bool               `json:"valid"`
	Reasons         []string           `json:"reasons"`
}

func AssessCalibration(profile CalibrationProfile, expected, observed []Vector) (CalibrationAssessment, error) {
	if err := profile.Validate(); err != nil {
		return CalibrationAssessment{}, err
	}
	if len(expected) == 0 || len(expected) != len(observed) {
		return CalibrationAssessment{}, fmt.Errorf("%w: calibration sample lengths differ", ErrInvalid)
	}
	assessment := CalibrationAssessment{Profile: profile, Samples: len(expected), Reasons: []string{}, Valid: true}
	for index := range expected {
		residual, err := profile.Residual(expected[index], observed[index])
		if err != nil {
			assessment.Valid = false
			assessment.Reasons = append(assessment.Reasons, fmt.Sprintf("sample %d: %v", index+1, err))
			continue
		}
		length := residual.Length()
		assessment.MeanResidual += length
		if length > assessment.MaximumResidual {
			assessment.MaximumResidual = length
		}
		if length > profile.Tolerance {
			assessment.Valid = false
			assessment.Reasons = append(assessment.Reasons, fmt.Sprintf("sample %d residual %.5f exceeds tolerance", index+1, length))
		}
	}
	assessment.MeanResidual /= float64(assessment.Samples)
	return assessment, nil
}

func (a CalibrationAssessment) Score() float64 {
	if a.Samples == 0 {
		return 0
	}
	if a.MeanResidual == 0 {
		return 1
	}
	score := 1 - a.MeanResidual/a.Profile.Tolerance
	if math.IsNaN(score) || score < 0 {
		return 0
	}
	if score > 1 {
		return 1
	}
	return score
}

func (a CalibrationAssessment) Explain() string {
	if a.Valid {
		return fmt.Sprintf("%d calibration samples are inside tolerance", a.Samples)
	}
	if len(a.Reasons) == 0 {
		return "calibration failed without a recorded reason"
	}
	text := ""
	for index, reason := range a.Reasons {
		if index > 0 {
			text += "; "
		}
		text += reason
	}
	return text
}
