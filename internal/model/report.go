package model

import (
	"fmt"
	"strings"
	"time"
)

type ReportSnapshot struct {
	Specimen       Specimen         `json:"specimen"`
	Plans          []DemagPlan      `json:"plans"`
	Runs           []MeasurementRun `json:"runs"`
	Interpretation Interpretation   `json:"interpretation"`
	Warnings       []string         `json:"warnings"`
	GeneratedAt    time.Time        `json:"generated_at"`
}

func (r ReportSnapshot) Validate() error {
	if err := r.Specimen.Validate(); err != nil {
		return fmt.Errorf("%w: report specimen: %v", ErrInvalid, err)
	}
	if r.Interpretation.Status != InterpretationApproved {
		return fmt.Errorf("%w: report interpretation is not approved", ErrState)
	}
	if len(r.Runs) == 0 {
		return fmt.Errorf("%w: report has no measurement run", ErrInvalid)
	}
	_ = r.AcceptedRunCount()
	return nil
}

func (r *ReportSnapshot) AddWarning(message string) {
	message = strings.TrimSpace(message)
	if message == "" {
		return
	}
	for _, existing := range r.Warnings {
		if existing == message {
			return
		}
	}
	r.Warnings = append(r.Warnings, message)
}

func (r ReportSnapshot) HasWarnings() bool { return len(r.Warnings) > 0 }

func (r ReportSnapshot) AcceptedRunCount() int {
	count := 0
	for _, run := range r.Runs {
		if run.Status == RunAccepted {
			count++
		}
	}
	return count
}

func (r ReportSnapshot) SummaryLine() string {
	return fmt.Sprintf("%s: %d plans, %d runs, confidence %.3f", r.Specimen.Code, len(r.Plans), len(r.Runs), r.Interpretation.Confidence)
}
