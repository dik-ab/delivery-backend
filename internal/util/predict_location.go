package util

import (
	"encoding/json"
	"math"
)

// RouteStep represents a step in the route (from Google Directions API)
type RouteStep struct {
	DurationSec int     `json:"duration_sec"`
	StartLat    float64 `json:"start_lat"`
	StartLng    float64 `json:"start_lng"`
	EndLat      float64 `json:"end_lat"`
	EndLng      float64 `json:"end_lng"`
	Polyline    string  `json:"polyline"` // encoded polyline for this step
}

// PredictedLocation represents the predicted position at a given time
type PredictedLocation struct {
	Lat              float64 `json:"lat"`
	Lng              float64 `json:"lng"`
	ProgressPercent  float64 `json:"progress_percent"`   // 0.0 ~ 100.0
	ElapsedSeconds   int     `json:"elapsed_seconds"`
	RemainingSeconds int     `json:"remaining_seconds"`
	StepIndex        int     `json:"step_index"`
}

// PredictLocationFromSteps calculates predicted position based on elapsed time
// stepsJSON: JSON string of []RouteStep from Google Directions API
// elapsedSec: seconds elapsed since departure
// totalDurationSec: total route duration in seconds
func PredictLocationFromSteps(stepsJSON string, elapsedSec int, totalDurationSec int) (*PredictedLocation, error) {
	var steps []RouteStep
	if err := json.Unmarshal([]byte(stepsJSON), &steps); err != nil {
		return nil, err
	}

	if len(steps) == 0 {
		return nil, nil
	}

	// Clamp elapsed time
	if elapsedSec <= 0 {
		return &PredictedLocation{
			Lat:              steps[0].StartLat,
			Lng:              steps[0].StartLng,
			ProgressPercent:  0,
			ElapsedSeconds:   0,
			RemainingSeconds: totalDurationSec,
			StepIndex:        0,
		}, nil
	}

	if elapsedSec >= totalDurationSec {
		lastStep := steps[len(steps)-1]
		return &PredictedLocation{
			Lat:              lastStep.EndLat,
			Lng:              lastStep.EndLng,
			ProgressPercent:  100,
			ElapsedSeconds:   totalDurationSec,
			RemainingSeconds: 0,
			StepIndex:        len(steps) - 1,
		}, nil
	}

	// Find which step the elapsed time falls into
	accumulatedSec := 0
	for i, step := range steps {
		if accumulatedSec+step.DurationSec > elapsedSec {
			// This is the step we're currently in
			stepElapsed := elapsedSec - accumulatedSec
			stepProgress := float64(stepElapsed) / float64(step.DurationSec)

			// Linear interpolation between step start and end
			lat := step.StartLat + (step.EndLat-step.StartLat)*stepProgress
			lng := step.StartLng + (step.EndLng-step.StartLng)*stepProgress

			progressPercent := float64(elapsedSec) / float64(totalDurationSec) * 100

			return &PredictedLocation{
				Lat:              math.Round(lat*1000000) / 1000000,
				Lng:              math.Round(lng*1000000) / 1000000,
				ProgressPercent:  math.Round(progressPercent*10) / 10,
				ElapsedSeconds:   elapsedSec,
				RemainingSeconds: totalDurationSec - elapsedSec,
				StepIndex:        i,
			}, nil
		}
		accumulatedSec += step.DurationSec
	}

	// Fallback: return last step's end
	lastStep := steps[len(steps)-1]
	return &PredictedLocation{
		Lat:              lastStep.EndLat,
		Lng:              lastStep.EndLng,
		ProgressPercent:  100,
		ElapsedSeconds:   totalDurationSec,
		RemainingSeconds: 0,
		StepIndex:        len(steps) - 1,
	}, nil
}
