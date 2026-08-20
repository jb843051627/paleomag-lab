package model

import (
	"fmt"
	"sort"
)

type DemagProtocol struct {
	Name        string    `json:"name"`
	Method      string    `json:"method"`
	Temperature float64   `json:"temperature"`
	Fields      []float64 `json:"fields"`
	Rationale   string    `json:"rationale"`
}

func (p DemagProtocol) Validate() error {
	if p.Name == "" || p.Method == "" || p.Rationale == "" {
		return fmt.Errorf("%w: protocol metadata is required", ErrInvalid)
	}
	if len(p.Fields) < 2 {
		return fmt.Errorf("%w: protocol needs at least two fields", ErrInvalid)
	}
	copyFields := append([]float64(nil), p.Fields...)
	sort.Float64s(copyFields)
	for index, field := range p.Fields {
		if field < 0 {
			return fmt.Errorf("%w: field %d is negative", ErrInvalid, index+1)
		}
		if index > 0 && field <= p.Fields[index-1] {
			return fmt.Errorf("%w: protocol fields must be strictly increasing", ErrInvalid)
		}
	}
	if copyFields[len(copyFields)-1] > 10000 {
		return fmt.Errorf("%w: field exceeds instrument safety limit", ErrInvalid)
	}
	return nil
}

func (p DemagProtocol) StepCount() int { return len(p.Fields) }

func (p DemagProtocol) FieldAt(index int) (float64, error) {
	if index < 0 || index >= len(p.Fields) {
		return 0, fmt.Errorf("%w: protocol step %d", ErrNotFound, index)
	}
	return p.Fields[index], nil
}

func (p DemagProtocol) TotalField() float64 {
	total := 0.0
	for _, field := range p.Fields {
		total += field
	}
	return total
}
