/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package imagelocality

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"dguest-scheduler/pkg/scheduler/framework"
	"dguest-scheduler/pkg/scheduler/framework/runtime"
	"dguest-scheduler/pkg/scheduler/internal/cache"
	"github.com/google/go-cmp/cmp"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestImageLocalityPriority(t *testing.T) {
	test40250 := v1alpha1.DguestSpec{
		Containers: []v1.Container{
			{

				Image: "gcr.io/40",
			},
			{
				Image: "gcr.io/250",
			},
		},
	}

	test40300 := v1alpha1.DguestSpec{
		Containers: []v1.Container{
			{
				Image: "gcr.io/40",
			},
			{
				Image: "gcr.io/300",
			},
		},
	}

	testMinMax := v1alpha1.DguestSpec{
		Containers: []v1.Container{
			{
				Image: "gcr.io/10",
			},
			{
				Image: "gcr.io/4000",
			},
		},
	}

	test300600900 := v1alpha1.DguestSpec{
		Containers: []v1.Container{
			{
				Image: "gcr.io/300",
			},
			{
				Image: "gcr.io/600",
			},
			{
				Image: "gcr.io/900",
			},
		},
	}

	test3040 := v1alpha1.DguestSpec{
		Containers: []v1.Container{
			{
				Image: "gcr.io/30",
			},
			{
				Image: "gcr.io/40",
			},
		},
	}

	food403002000 := v1alpha1.FoodStatus{
		Images: []v1.ContainerImage{
			{
				Names: []string{
					"gcr.io/40:latest",
					"gcr.io/40:v1",
					"gcr.io/40:v1",
				},
				SizeBytes: int64(40 * mb),
			},
			{
				Names: []string{
					"gcr.io/300:latest",
					"gcr.io/300:v1",
				},
				SizeBytes: int64(300 * mb),
			},
			{
				Names: []string{
					"gcr.io/2000:latest",
				},
				SizeBytes: int64(2000 * mb),
			},
		},
	}

	food25010 := v1alpha1.FoodStatus{
		Images: []v1.ContainerImage{
			{
				Names: []string{
					"gcr.io/250:latest",
				},
				SizeBytes: int64(250 * mb),
			},
			{
				Names: []string{
					"gcr.io/10:latest",
					"gcr.io/10:v1",
				},
				SizeBytes: int64(10 * mb),
			},
		},
	}

	food60040900 := v1alpha1.FoodStatus{
		Images: []v1.ContainerImage{
			{
				Names: []string{
					"gcr.io/600:latest",
				},
				SizeBytes: int64(600 * mb),
			},
			{
				Names: []string{
					"gcr.io/40:latest",
				},
				SizeBytes: int64(40 * mb),
			},
			{
				Names: []string{
					"gcr.io/900:latest",
				},
				SizeBytes: int64(900 * mb),
			},
		},
	}

	food300600900 := v1alpha1.FoodStatus{
		Images: []v1.ContainerImage{
			{
				Names: []string{
					"gcr.io/300:latest",
				},
				SizeBytes: int64(300 * mb),
			},
			{
				Names: []string{
					"gcr.io/600:latest",
				},
				SizeBytes: int64(600 * mb),
			},
			{
				Names: []string{
					"gcr.io/900:latest",
				},
				SizeBytes: int64(900 * mb),
			},
		},
	}

	food400030 := v1alpha1.FoodStatus{
		Images: []v1.ContainerImage{
			{
				Names: []string{
					"gcr.io/4000:latest",
				},
				SizeBytes: int64(4000 * mb),
			},
			{
				Names: []string{
					"gcr.io/30:latest",
				},
				SizeBytes: int64(30 * mb),
			},
		},
	}

	food203040 := v1alpha1.FoodStatus{
		Images: []v1.ContainerImage{
			{
				Names: []string{
					"gcr.io/20:latest",
				},
				SizeBytes: int64(20 * mb),
			},
			{
				Names: []string{
					"gcr.io/30:latest",
				},
				SizeBytes: int64(30 * mb),
			},
			{
				Names: []string{
					"gcr.io/40:latest",
				},
				SizeBytes: int64(40 * mb),
			},
		},
	}

	foodWithNoImages := v1alpha1.FoodStatus{}

	tests := []struct {
		dguest       *v1alpha1.Dguest
		dguests      []*v1alpha1.Dguest
		foods        []*v1alpha1.Food
		expectedList framework.FoodScoreList
		name         string
	}{
		{
			// Dguest: gcr.io/40 gcr.io/250

			// Food1
			// Image: gcr.io/40:latest 40MB
			// Score: 0 (40M/2 < 23M, min-threshold)

			// Food2
			// Image: gcr.io/250:latest 250MB
			// Score: 100 * (250M/2 - 23M)/(1000M * 2 - 23M) = 5
			dguest:       &v1alpha1.Dguest{Spec: test40250},
			foods:        []*v1alpha1.Food{makeImageFood("food1", food403002000), makeImageFood("food2", food25010)},
			expectedList: []framework.FoodScore{{Name: "food1", Score: 0}, {Name: "food2", Score: 5}},
			name:         "two images spread on two foods, prefer the larger image one",
		},
		{
			// Dguest: gcr.io/40 gcr.io/300

			// Food1
			// Image: gcr.io/40:latest 40MB, gcr.io/300:latest 300MB
			// Score: 100 * ((40M + 300M)/2 - 23M)/(1000M * 2 - 23M) = 7

			// Food2
			// Image: not present
			// Score: 0
			dguest:       &v1alpha1.Dguest{Spec: test40300},
			foods:        []*v1alpha1.Food{makeImageFood("food1", food403002000), makeImageFood("food2", food25010)},
			expectedList: []framework.FoodScore{{Name: "food1", Score: 7}, {Name: "food2", Score: 0}},
			name:         "two images on one food, prefer this food",
		},
		{
			// Dguest: gcr.io/4000 gcr.io/10

			// Food1
			// Image: gcr.io/4000:latest 2000MB
			// Score: 100 (4000 * 1/2 >= 1000M * 2, max-threshold)

			// Food2
			// Image: gcr.io/10:latest 10MB
			// Score: 0 (10M/2 < 23M, min-threshold)
			dguest:       &v1alpha1.Dguest{Spec: testMinMax},
			foods:        []*v1alpha1.Food{makeImageFood("food1", food400030), makeImageFood("food2", food25010)},
			expectedList: []framework.FoodScore{{Name: "food1", Score: framework.MaxFoodScore}, {Name: "food2", Score: 0}},
			name:         "if exceed limit, use limit",
		},
		{
			// Dguest: gcr.io/4000 gcr.io/10

			// Food1
			// Image: gcr.io/4000:latest 4000MB
			// Score: 100 * (4000M/3 - 23M)/(1000M * 2 - 23M) = 66

			// Food2
			// Image: gcr.io/10:latest 10MB
			// Score: 0 (10M*1/3 < 23M, min-threshold)

			// Food3
			// Image:
			// Score: 0
			dguest:       &v1alpha1.Dguest{Spec: testMinMax},
			foods:        []*v1alpha1.Food{makeImageFood("food1", food400030), makeImageFood("food2", food25010), makeImageFood("food3", foodWithNoImages)},
			expectedList: []framework.FoodScore{{Name: "food1", Score: 66}, {Name: "food2", Score: 0}, {Name: "food3", Score: 0}},
			name:         "if exceed limit, use limit (with food which has no images present)",
		},
		{
			// Dguest: gcr.io/300 gcr.io/600 gcr.io/900

			// Food1
			// Image: gcr.io/600:latest 600MB, gcr.io/900:latest 900MB
			// Score: 100 * (600M * 2/3 + 900M * 2/3 - 23M) / (1000M * 3 - 23M) = 32

			// Food2
			// Image: gcr.io/300:latest 300MB, gcr.io/600:latest 600MB, gcr.io/900:latest 900MB
			// Score: 100 * (300M * 1/3 + 600M * 2/3 + 900M * 2/3 - 23M) / (1000M *3 - 23M) = 36

			// Food3
			// Image:
			// Score: 0
			dguest:       &v1alpha1.Dguest{Spec: test300600900},
			foods:        []*v1alpha1.Food{makeImageFood("food1", food60040900), makeImageFood("food2", food300600900), makeImageFood("food3", foodWithNoImages)},
			expectedList: []framework.FoodScore{{Name: "food1", Score: 32}, {Name: "food2", Score: 36}, {Name: "food3", Score: 0}},
			name:         "dguest with multiple large images, food2 is preferred",
		},
		{
			// Dguest: gcr.io/30 gcr.io/40

			// Food1
			// Image: gcr.io/20:latest 20MB, gcr.io/30:latest 30MB gcr.io/40:latest 40MB
			// Score: 100 * (30M + 40M * 1/2 - 23M) / (1000M * 2 - 23M) = 1

			// Food2
			// Image: 100 * (30M - 23M) / (1000M * 2 - 23M) = 0
			// Score: 0
			dguest:       &v1alpha1.Dguest{Spec: test3040},
			foods:        []*v1alpha1.Food{makeImageFood("food1", food203040), makeImageFood("food2", food400030)},
			expectedList: []framework.FoodScore{{Name: "food1", Score: 1}, {Name: "food2", Score: 0}},
			name:         "dguest with multiple small images",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			snapshot := cache.NewSnapshot(nil, test.foods)
			state := framework.NewCycleState()
			fh, _ := runtime.NewFramework(nil, nil, ctx.Done(), runtime.WithSnapshotSharedLister(snapshot))

			p, _ := New(nil, fh)
			var gotList framework.FoodScoreList
			for _, n := range test.foods {
				foodName := n.ObjectMeta.Name
				score, status := p.(framework.ScorePlugin).Score(ctx, state, test.dguest, foodName)
				if !status.IsSuccess() {
					t.Errorf("unexpected error: %v", status)
				}
				gotList = append(gotList, framework.FoodScore{Name: foodName, Score: score})
			}

			if diff := cmp.Diff(test.expectedList, gotList); diff != "" {
				t.Errorf("Unexpected food score list (-want, +got):\n%s", diff)
			}
		})
	}
}

func TestNormalizedImageName(t *testing.T) {
	for _, testCase := range []struct {
		Name   string
		Input  string
		Output string
	}{
		{Name: "add :latest postfix 1", Input: "root", Output: "root:latest"},
		{Name: "add :latest postfix 2", Input: "gcr.io:5000/root", Output: "gcr.io:5000/root:latest"},
		{Name: "keep it as is 1", Input: "root:tag", Output: "root:tag"},
		{Name: "keep it as is 2", Input: "root@" + getImageFakeDigest("root"), Output: "root@" + getImageFakeDigest("root")},
	} {
		t.Run(testCase.Name, func(t *testing.T) {
			image := normalizedImageName(testCase.Input)
			if image != testCase.Output {
				t.Errorf("expected image reference: %q, got %q", testCase.Output, image)
			}
		})
	}
}

func makeImageFood(food string, status v1alpha1.FoodStatus) *v1alpha1.Food {
	return &v1alpha1.Food{
		ObjectMeta: metav1.ObjectMeta{Name: food},
		Status:     status,
	}
}

func getImageFakeDigest(fakeContent string) string {
	hash := sha256.Sum256([]byte(fakeContent))
	return "sha256:" + hex.EncodeToString(hash[:])
}
