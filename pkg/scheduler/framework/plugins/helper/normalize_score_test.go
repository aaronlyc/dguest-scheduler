package helper

import (
	"fmt"
	"testing"

	"dguest-scheduler/pkg/scheduler/framework"

	"github.com/google/go-cmp/cmp"
)

func TestDefaultNormalizeScore(t *testing.T) {
	tests := []struct {
		reverse        bool
		scores         []int64
		expectedScores []int64
	}{
		{
			scores:         []int64{1, 2, 3, 4},
			expectedScores: []int64{25, 50, 75, 100},
		},
		{
			reverse:        true,
			scores:         []int64{1, 2, 3, 4},
			expectedScores: []int64{75, 50, 25, 0},
		},
		{
			scores:         []int64{1000, 10, 20, 30},
			expectedScores: []int64{100, 1, 2, 3},
		},
		{
			reverse:        true,
			scores:         []int64{1000, 10, 20, 30},
			expectedScores: []int64{0, 99, 98, 97},
		},
		{
			scores:         []int64{1, 1, 1, 1},
			expectedScores: []int64{100, 100, 100, 100},
		},
		{
			scores:         []int64{1000, 1, 1, 1},
			expectedScores: []int64{100, 0, 0, 0},
		},
		{
			reverse:        true,
			scores:         []int64{0, 1, 1, 1},
			expectedScores: []int64{100, 0, 0, 0},
		},
		{
			scores:         []int64{0, 0, 0, 0},
			expectedScores: []int64{0, 0, 0, 0},
		},
		{
			reverse:        true,
			scores:         []int64{0, 0, 0, 0},
			expectedScores: []int64{100, 100, 100, 100},
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("case_%d", i), func(t *testing.T) {
			scores := framework.FoodScoreList{}
			for _, score := range test.scores {
				scores = append(scores, framework.FoodScore{Score: score})
			}

			expectedScores := framework.FoodScoreList{}
			for _, score := range test.expectedScores {
				expectedScores = append(expectedScores, framework.FoodScore{Score: score})
			}

			DefaultNormalizeScore(framework.MaxFoodScore, test.reverse, scores)
			if diff := cmp.Diff(expectedScores, scores); diff != "" {
				t.Errorf("Unexpected scores (-want, +got):\n%s", diff)
			}
		})
	}
}
