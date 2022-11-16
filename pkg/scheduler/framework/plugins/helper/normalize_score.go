package helper

import (
	"dguest-scheduler/pkg/scheduler/framework"
)

// DefaultNormalizeScore generates a Normalize Score function that can normalize the
// scores from [0, max(scores)] to [0, maxPriority]. If reverse is set to true, it
// reverses the scores by subtracting it from maxPriority.
// Note: The input scores are always assumed to be non-negative integers.
func DefaultNormalizeScore(maxPriority int64, reverse bool, scores framework.FoodScoreList) *framework.Status {
	var maxCount int64
	for i := range scores {
		if scores[i].Score > maxCount {
			maxCount = scores[i].Score
		}
	}

	if maxCount == 0 {
		if reverse {
			for i := range scores {
				scores[i].Score = maxPriority
			}
		}
		return nil
	}

	for i := range scores {
		score := scores[i].Score

		score = maxPriority * score / maxCount
		if reverse {
			score = maxPriority - score
		}

		scores[i].Score = score
	}
	return nil
}
