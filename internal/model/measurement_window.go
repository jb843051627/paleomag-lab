package model

import (
	"fmt"
	"sort"
)

type MeasurementWindow struct {
	StartStep int `json:"start_step"`
	EndStep   int `json:"end_step"`
}

func (w MeasurementWindow) Validate() error {
	if w.StartStep < 1 || w.EndStep < w.StartStep {
		return fmt.Errorf("%w: measurement window is invalid", ErrInvalid)
	}
	return nil
}

func (w MeasurementWindow) Contains(step int) bool { return step >= w.StartStep && step <= w.EndStep }

func (w MeasurementWindow) Size() int { return w.EndStep - w.StartStep + 1 }

func (w MeasurementWindow) IsSingleStep() bool { return w.StartStep == w.EndStep }

func (w MeasurementWindow) Expand(steps int) MeasurementWindow {
	if steps < 0 {
		steps = 0
	}
	return MeasurementWindow{StartStep: w.StartStep, EndStep: w.EndStep + steps}
}

func (w MeasurementWindow) Includes(other MeasurementWindow) bool {
	return w.StartStep <= other.StartStep && w.EndStep >= other.EndStep
}

func (w MeasurementWindow) Label() string {
	return fmt.Sprintf("steps %d-%d", w.StartStep, w.EndStep)
}

func SelectMeasurementWindow(items []Measurement, window MeasurementWindow) ([]Measurement, error) {
	if err := window.Validate(); err != nil {
		return nil, err
	}
	selected := make([]Measurement, 0)
	for _, item := range items {
		if window.Contains(item.Step) {
			selected = append(selected, item)
		}
	}
	sort.SliceStable(selected, func(i, j int) bool { return selected[i].Step < selected[j].Step })
	if len(selected) == 0 {
		return nil, fmt.Errorf("%w: no measurement is inside window", ErrNotFound)
	}
	return selected, nil
}

func MergeWindows(left, right MeasurementWindow) (MeasurementWindow, error) {
	if err := left.Validate(); err != nil {
		return MeasurementWindow{}, err
	}
	if err := right.Validate(); err != nil {
		return MeasurementWindow{}, err
	}
	if left.EndStep+1 < right.StartStep || right.EndStep+1 < left.StartStep {
		return MeasurementWindow{}, fmt.Errorf("%w: windows are disjoint", ErrConflict)
	}
	start, end := left.StartStep, left.EndStep
	if right.StartStep < start {
		start = right.StartStep
	}
	if right.EndStep > end {
		end = right.EndStep
	}
	return MeasurementWindow{StartStep: start, EndStep: end}, nil
}
