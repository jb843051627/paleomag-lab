package model

import (
	"fmt"
	"math"
	"sort"
)

type TrendPoint struct {
	Step        int     `json:"step"`
	Field       float64 `json:"field"`
	Intensity   float64 `json:"intensity"`
	Declination float64 `json:"declination"`
	Inclination float64 `json:"inclination"`
	Change      float64 `json:"change"`
}

type TrendSummary struct {
	Points               []TrendPoint `json:"points"`
	Stable               bool         `json:"stable"`
	Polarity             string       `json:"polarity"`
	IntensityLoss        float64      `json:"intensity_loss"`
	MaximumAngularChange float64      `json:"maximum_angular_change"`
	SuggestedEndStep     int          `json:"suggested_end_step"`
}

func AnalyzeTrend(items []Measurement) (TrendSummary, error) {
	if len(items) == 0 {
		return TrendSummary{}, fmt.Errorf("%w: no measurements for trend", ErrInvalid)
	}
	sorted := append([]Measurement(nil), items...)
	sort.SliceStable(sorted, func(i, j int) bool { return sorted[i].Step < sorted[j].Step })
	points := make([]TrendPoint, 0, len(sorted))
	var previous *Vector
	var maximumChange float64
	for _, item := range sorted {
		vector := VectorFromMeasurement(item)
		if err := vector.Validate(); err != nil {
			return TrendSummary{}, fmt.Errorf("%w: step %d: %v", ErrInvalid, item.Step, err)
		}
		change := 0.0
		if previous != nil {
			value, err := AngularDistance(*previous, vector)
			if err != nil {
				return TrendSummary{}, err
			}
			change = value
			if value > maximumChange {
				maximumChange = value
			}
		}
		points = append(points, TrendPoint{Step: item.Step, Field: item.Field, Intensity: item.Intensity, Declination: vector.Declination(), Inclination: vector.Inclination(), Change: change})
		copy := vector
		previous = &copy
	}
	first, last := points[0], points[len(points)-1]
	intensityLoss := 0.0
	if first.Intensity > 0 {
		intensityLoss = 1 - last.Intensity/first.Intensity
	}
	polarity := "normal"
	if last.Inclination < 0 {
		polarity = "reversed"
	}
	stable := maximumChange <= 25 && intensityLoss >= 0 && intensityLoss <= 0.8
	endStep := last.Step
	if len(points) > 2 {
		for index := 1; index < len(points)-1; index++ {
			if points[index].Change <= 15 {
				endStep = points[index].Step
				break
			}
		}
	}
	return TrendSummary{Points: points, Stable: stable, Polarity: polarity, IntensityLoss: intensityLoss, MaximumAngularChange: maximumChange, SuggestedEndStep: endStep}, nil
}

func (s TrendSummary) Validate() error {
	if len(s.Points) == 0 || s.SuggestedEndStep < 1 {
		return fmt.Errorf("%w: trend summary is empty", ErrInvalid)
	}
	if math.IsNaN(s.IntensityLoss) || math.IsInf(s.IntensityLoss, 0) {
		return fmt.Errorf("%w: trend intensity loss is not finite", ErrInvalid)
	}
	return nil
}

func (s TrendSummary) NeedsMoreSteps() bool {
	return !s.Stable && len(s.Points) < 12 && s.IntensityLoss < 0.95
}

func (s TrendSummary) Direction() Vector {
	if len(s.Points) == 0 {
		return Vector{}
	}
	point := s.Points[len(s.Points)-1]
	inclination := point.Inclination * math.Pi / 180
	declination := point.Declination * math.Pi / 180
	return Vector{X: math.Cos(inclination) * math.Cos(declination), Y: math.Cos(inclination) * math.Sin(declination), Z: math.Sin(inclination)}
}
