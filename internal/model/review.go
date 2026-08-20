package model

import (
	"fmt"
	"strings"
	"time"
)

type Review struct {
	ID               string         `json:"id"`
	InterpretationID string         `json:"interpretation_id"`
	Reviewer         string         `json:"reviewer"`
	Decision         ReviewDecision `json:"decision"`
	Comment          string         `json:"comment"`
	CreatedAt        time.Time      `json:"created_at"`
}

func (r Review) Validate() error {
	if r.ID == "" || r.InterpretationID == "" || strings.TrimSpace(r.Reviewer) == "" {
		return fmt.Errorf("%w: review identity is required", ErrInvalid)
	}
	if r.Decision != ReviewApprove && r.Decision != ReviewReject && r.Decision != ReviewRework {
		return fmt.Errorf("%w: review decision is invalid", ErrInvalid)
	}
	if strings.TrimSpace(r.Comment) == "" {
		return fmt.Errorf("%w: review comment is required", ErrInvalid)
	}
	return nil
}
