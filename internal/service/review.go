package service

import (
	"context"
	"fmt"

	"github.com/jb843051627/paleomag-lab/internal/clock"
	"github.com/jb843051627/paleomag-lab/internal/model"
	"github.com/jb843051627/paleomag-lab/internal/repository"
)

type ReviewService struct {
	reviews         *repository.ReviewRepository
	interpretations *repository.InterpretationRepository
	specimens       *repository.SpecimenRepository
	audits          *AuditService
	clock           clock.Clock
}

func (s *ReviewService) Decide(ctx context.Context, interpretationID, reviewer string, decision model.ReviewDecision, comment string) (model.Review, model.Interpretation, error) {
	interpretation, err := s.interpretations.Get(ctx, interpretationID)
	if err != nil {
		return model.Review{}, model.Interpretation{}, err
	}
	review := model.Review{ID: newID("review"), InterpretationID: interpretationID, Reviewer: reviewer, Decision: decision, Comment: comment, CreatedAt: s.clock.Now()}
	if err := review.Validate(); err != nil {
		return model.Review{}, model.Interpretation{}, err
	}
	before := interpretation
	if err := interpretation.ApplyDecision(decision, s.clock.Now()); err != nil {
		return model.Review{}, model.Interpretation{}, err
	}
	if err := s.reviews.Create(ctx, review); err != nil {
		return model.Review{}, model.Interpretation{}, err
	}
	if err := s.interpretations.Update(ctx, interpretation); err != nil {
		return model.Review{}, model.Interpretation{}, err
	}
	if decision == model.ReviewApprove {
		specimen, err := s.specimens.Get(ctx, interpretation.SpecimenID)
		if err != nil {
			return model.Review{}, model.Interpretation{}, fmt.Errorf("load reviewed specimen: %w", err)
		}
		if specimen.Status == model.SpecimenInProgress {
			if err := specimen.MarkInterpreted(s.clock.Now()); err != nil {
				return model.Review{}, model.Interpretation{}, err
			}
			if err := s.specimens.Update(ctx, specimen); err != nil {
				return model.Review{}, model.Interpretation{}, err
			}
		}
	}
	if err := auditState(ctx, s.audits, "interpretation", interpretationID, "reviewed", reviewer, before, interpretation); err != nil {
		return model.Review{}, model.Interpretation{}, err
	}
	return review, interpretation, nil
}

func (s *ReviewService) Get(ctx context.Context, id string) (model.Review, error) {
	return s.reviews.Get(ctx, id)
}

func (s *ReviewService) List(ctx context.Context, interpretationID string) ([]model.Review, error) {
	return s.reviews.ListByInterpretation(ctx, interpretationID)
}
