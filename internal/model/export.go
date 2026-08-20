package model

import (
	"fmt"
	"strings"
	"time"
)

type ExportJob struct {
	ID          string       `json:"id"`
	SpecimenID  string       `json:"specimen_id"`
	Format      string       `json:"format"`
	RequestedBy string       `json:"requested_by"`
	Status      ExportStatus `json:"status"`
	Result      string       `json:"result"`
	Error       string       `json:"error,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

func (j ExportJob) Validate() error {
	if j.ID == "" || j.SpecimenID == "" || strings.TrimSpace(j.RequestedBy) == "" {
		return fmt.Errorf("%w: export identity is required", ErrInvalid)
	}
	if j.Format != "json" && j.Format != "csv" {
		return fmt.Errorf("%w: export format must be json or csv", ErrInvalid)
	}
	return nil
}

func (j *ExportJob) Start(now time.Time) error {
	if j.Status != ExportQueued {
		return fmt.Errorf("%w: export is not queued", ErrState)
	}
	j.Status = ExportRunning
	j.UpdatedAt = now
	return nil
}

func (j *ExportJob) Complete(result string, now time.Time) error {
	if j.Status != ExportRunning {
		return fmt.Errorf("%w: export is not running", ErrState)
	}
	j.Status = ExportComplete
	j.Result = result
	j.UpdatedAt = now
	return nil
}

func (j *ExportJob) Fail(err error, now time.Time) error {
	if j.Status != ExportRunning {
		return fmt.Errorf("%w: export is not running", ErrState)
	}
	j.Status = ExportComplete
	if err != nil {
		j.Error = err.Error()
	}
	j.UpdatedAt = now
	return nil
}
