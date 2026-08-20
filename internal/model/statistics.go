package model

import "math"

type MeasurementSummary struct {
	Count         int     `json:"count"`
	Good          int     `json:"good"`
	Review        int     `json:"review"`
	Bad           int     `json:"bad"`
	MeanField     float64 `json:"mean_field"`
	MeanIntensity float64 `json:"mean_intensity"`
	Spread        float64 `json:"spread"`
}

func SummarizeMeasurements(items []Measurement) MeasurementSummary {
	var out MeasurementSummary
	if len(items) == 0 {
		return out
	}
	var fieldSum, intensitySum float64
	minIntensity, maxIntensity := math.Inf(1), math.Inf(-1)
	for _, item := range items {
		out.Count++
		fieldSum += item.Field
		intensitySum += item.Intensity
		if item.Intensity < minIntensity {
			minIntensity = item.Intensity
		}
		if item.Intensity > maxIntensity {
			maxIntensity = item.Intensity
		}
		switch item.Quality {
		case QualityGood:
			out.Good++
		case QualityReview:
			out.Review++
		case QualityBad:
			out.Bad++
		}
	}
	out.MeanField = fieldSum / float64(out.Count)
	out.MeanIntensity = intensitySum / float64(out.Count)
	out.Spread = maxIntensity - minIntensity
	return out
}
