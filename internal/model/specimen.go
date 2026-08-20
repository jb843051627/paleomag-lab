package model

import (
	"fmt"
	"strings"
	"time"
)

type Specimen struct {
	ID          string         `json:"id"`
	Code        string         `json:"code"`
	Lithology   string         `json:"lithology"`
	Site        string         `json:"site"`
	Collector   string         `json:"collector"`
	Status      SpecimenStatus `json:"status"`
	Orientation *Orientation   `json:"orientation,omitempty"`
	Notes       string         `json:"notes"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

func NewSpecimen(id, code, lithology, site, collector string, now time.Time) Specimen {
	return Specimen{ID: id, Code: strings.TrimSpace(code), Lithology: strings.TrimSpace(lithology), Site: strings.TrimSpace(site), Collector: strings.TrimSpace(collector), Status: SpecimenRegistered, CreatedAt: now, UpdatedAt: now}
}

func (s Specimen) Validate() error {
	if strings.TrimSpace(s.ID) == "" || strings.TrimSpace(s.Code) == "" {
		return fmt.Errorf("%w: specimen id and code are required", ErrInvalid)
	}
	if s.Lithology == "" || s.Site == "" || s.Collector == "" {
		return fmt.Errorf("%w: lithology, site and collector are required", ErrInvalid)
	}
	if s.Orientation != nil {
		if err := s.Orientation.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (s *Specimen) SetOrientation(o Orientation, now time.Time) error {
	if s.Status != SpecimenRegistered {
		return fmt.Errorf("%w: orientation can only be set for registered specimen", ErrState)
	}
	if err := o.Validate(); err != nil {
		return err
	}
	s.Orientation = &o
	s.Status = SpecimenOriented
	s.UpdatedAt = now
	return nil
}

func (s *Specimen) StartWork(now time.Time) error {
	if s.Status != SpecimenOriented {
		return fmt.Errorf("%w: specimen must be oriented before work", ErrState)
	}
	s.Status = SpecimenInProgress
	s.UpdatedAt = now
	return nil
}

func (s *Specimen) MarkInterpreted(now time.Time) error {
	if s.Status != SpecimenInProgress {
		return fmt.Errorf("%w: specimen is not in progress", ErrState)
	}
	s.Status = SpecimenInterpreted
	s.UpdatedAt = now
	return nil
}

func (s *Specimen) Archive(now time.Time) error {
	if s.Status == SpecimenArchived {
		return fmt.Errorf("%w: specimen is already archived", ErrState)
	}
	s.Status = SpecimenArchived
	s.UpdatedAt = now
	return nil
}
