package model

import (
	"fmt"
	"math"
)

type QualityPolicy struct {
	MinimumIntensity   float64 `json:"minimum_intensity"`
	MaximumIntensity   float64 `json:"maximum_intensity"`
	MaximumAngularJump float64 `json:"maximum_angular_jump"`
	MaximumFieldGap    float64 `json:"maximum_field_gap"`
	ReviewRatio        float64 `json:"review_ratio"`
}

func DefaultQualityPolicy() QualityPolicy {
	return QualityPolicy{MinimumIntensity: 0.01, MaximumIntensity: 1e9, MaximumAngularJump: 35, MaximumFieldGap: 100, ReviewRatio: 0.25}
}

func (p QualityPolicy) Validate() error {
	if p.MinimumIntensity <= 0 || p.MaximumIntensity <= p.MinimumIntensity {
		return fmt.Errorf("%w: quality intensity range is invalid", ErrInvalid)
	}
	if p.MaximumAngularJump <= 0 || p.MaximumAngularJump > 180 {
		return fmt.Errorf("%w: angular jump threshold is invalid", ErrInvalid)
	}
	if p.MaximumFieldGap <= 0 || p.ReviewRatio < 0 || p.ReviewRatio > 1 {
		return fmt.Errorf("%w: quality thresholds are invalid", ErrInvalid)
	}
	return nil
}

type QualityAssessment struct {
	Flag       QualityFlag `json:"flag"`
	Score      float64     `json:"score"`
	Reasons    []string    `json:"reasons"`
	ComparedTo string      `json:"compared_to"`
}

func (a QualityAssessment) IsAcceptable() bool { return a.Flag == QualityGood && a.Score >= 0.6 }

func AssessMeasurement(item Measurement, previous *Measurement, policy QualityPolicy) QualityAssessment {
	assessment := QualityAssessment{Flag: QualityGood, Score: 1, Reasons: []string{}}
	if err := policy.Validate(); err != nil {
		assessment.Flag = QualityBad
		assessment.Score = 0
		assessment.Reasons = append(assessment.Reasons, err.Error())
		return assessment
	}
	if item.Intensity < policy.MinimumIntensity || item.Intensity > policy.MaximumIntensity {
		assessment.Flag = QualityReview
		assessment.Score -= 0.35
		assessment.Reasons = append(assessment.Reasons, "intensity outside policy")
	}
	if !VectorFromMeasurement(item).IsFinite() || item.VectorLength() == 0 {
		assessment.Flag = QualityBad
		assessment.Score = 0
		assessment.Reasons = append(assessment.Reasons, "vector is invalid")
	}
	if previous != nil {
		assessment.ComparedTo = previous.ID
		previousVector := VectorFromMeasurement(*previous)
		currentVector := VectorFromMeasurement(item)
		if angle, err := AngularDistance(previousVector, currentVector); err != nil {
			assessment.Flag = QualityReview
			assessment.Score -= 0.2
			assessment.Reasons = append(assessment.Reasons, "cannot compare vector")
		} else if angle > policy.MaximumAngularJump {
			assessment.Flag = QualityReview
			assessment.Score -= 0.25
			assessment.Reasons = append(assessment.Reasons, fmt.Sprintf("angular jump %.2f exceeds threshold", angle))
		}
		if math.Abs(item.Field-previous.Field) > policy.MaximumFieldGap {
			assessment.Flag = QualityReview
			assessment.Score -= 0.15
			assessment.Reasons = append(assessment.Reasons, "field gap exceeds threshold")
		}
	}
	if assessment.Score < 0 {
		assessment.Score = 0
	}
	if assessment.Score < 0.6 && assessment.Flag == QualityGood {
		assessment.Flag = QualityReview
	}
	return assessment
}

func AssessSeries(items []Measurement, policy QualityPolicy) []QualityAssessment {
	results := make([]QualityAssessment, len(items))
	var previous *Measurement
	for index := range items {
		results[index] = AssessMeasurement(items[index], previous, policy)
		if results[index].Flag != QualityBad {
			copy := items[index]
			previous = &copy
		}
	}
	return results
}

func QualityScore(results []QualityAssessment) float64 {
	if len(results) == 0 {
		return 0
	}
	total := 0.0
	for _, result := range results {
		total += result.Score
	}
	return total / float64(len(results))
}
