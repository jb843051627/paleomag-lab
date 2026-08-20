package model

import "time"

type SpecimenFilter struct {
	Status    SpecimenStatus
	Site      string
	Lithology string
	Search    string
	Limit     int
	Offset    int
}

type MeasurementFilter struct {
	RunID    string
	Quality  QualityFlag
	MinField *float64
	MaxField *float64
	Since    *time.Time
	Until    *time.Time
	Limit    int
}

type ReviewFilter struct {
	Decision *ReviewDecision
	Reviewer string
	Limit    int
}

func (f SpecimenFilter) Normalize() SpecimenFilter {
	if f.Limit <= 0 || f.Limit > 200 {
		f.Limit = 50
	}
	if f.Offset < 0 {
		f.Offset = 0
	}
	return f
}

func (f MeasurementFilter) Normalize() MeasurementFilter {
	if f.Limit <= 0 || f.Limit > 500 {
		f.Limit = 100
	}
	return f
}
