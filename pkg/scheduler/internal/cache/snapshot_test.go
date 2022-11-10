/*
Copyright 2018 The Kubernetes Authors.

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

package cache

import (
	"fmt"
	"reflect"
	"testing"

	"dguest-scheduler/pkg/scheduler/framework"
	st "dguest-scheduler/pkg/scheduler/testing"
	"github.com/google/go-cmp/cmp"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
)

const mb int64 = 1024 * 1024

func TestGetFoodImageStates(t *testing.T) {
	tests := []struct {
		food              *v1alpha1.Food
		imageExistenceMap map[string]sets.String
		expected          map[string]*framework.ImageStateSummary
	}{
		{
			food: &v1alpha1.Food{
				ObjectMeta: metav1.ObjectMeta{Name: "food-0"},
				Status: v1alpha1.FoodStatus{
					Images: []v1.ContainerImage{
						{
							Names: []string{
								"gcr.io/10:v1",
							},
							SizeBytes: int64(10 * mb),
						},
						{
							Names: []string{
								"gcr.io/200:v1",
							},
							SizeBytes: int64(200 * mb),
						},
					},
				},
			},
			imageExistenceMap: map[string]sets.String{
				"gcr.io/10:v1":  sets.NewString("food-0", "food-1"),
				"gcr.io/200:v1": sets.NewString("food-0"),
			},
			expected: map[string]*framework.ImageStateSummary{
				"gcr.io/10:v1": {
					Size:     int64(10 * mb),
					NumFoods: 2,
				},
				"gcr.io/200:v1": {
					Size:     int64(200 * mb),
					NumFoods: 1,
				},
			},
		},
		{
			food: &v1alpha1.Food{
				ObjectMeta: metav1.ObjectMeta{Name: "food-0"},
				Status:     v1alpha1.FoodStatus{},
			},
			imageExistenceMap: map[string]sets.String{
				"gcr.io/10:v1":  sets.NewString("food-1"),
				"gcr.io/200:v1": sets.NewString(),
			},
			expected: map[string]*framework.ImageStateSummary{},
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("case_%d", i), func(t *testing.T) {
			imageStates := getFoodImageStates(test.food, test.imageExistenceMap)
			if !reflect.DeepEqual(test.expected, imageStates) {
				t.Errorf("expected: %#v, got: %#v", test.expected, imageStates)
			}
		})
	}
}

func TestCreateImageExistenceMap(t *testing.T) {
	tests := []struct {
		foods    []*v1alpha1.Food
		expected map[string]sets.String
	}{
		{
			foods: []*v1alpha1.Food{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "food-0"},
					Status: v1alpha1.FoodStatus{
						Images: []v1.ContainerImage{
							{
								Names: []string{
									"gcr.io/10:v1",
								},
								SizeBytes: int64(10 * mb),
							},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "food-1"},
					Status: v1alpha1.FoodStatus{
						Images: []v1.ContainerImage{
							{
								Names: []string{
									"gcr.io/10:v1",
								},
								SizeBytes: int64(10 * mb),
							},
							{
								Names: []string{
									"gcr.io/200:v1",
								},
								SizeBytes: int64(200 * mb),
							},
						},
					},
				},
			},
			expected: map[string]sets.String{
				"gcr.io/10:v1":  sets.NewString("food-0", "food-1"),
				"gcr.io/200:v1": sets.NewString("food-1"),
			},
		},
		{
			foods: []*v1alpha1.Food{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "food-0"},
					Status:     v1alpha1.FoodStatus{},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "food-1"},
					Status: v1alpha1.FoodStatus{
						Images: []v1.ContainerImage{
							{
								Names: []string{
									"gcr.io/10:v1",
								},
								SizeBytes: int64(10 * mb),
							},
							{
								Names: []string{
									"gcr.io/200:v1",
								},
								SizeBytes: int64(200 * mb),
							},
						},
					},
				},
			},
			expected: map[string]sets.String{
				"gcr.io/10:v1":  sets.NewString("food-1"),
				"gcr.io/200:v1": sets.NewString("food-1"),
			},
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("case_%d", i), func(t *testing.T) {
			imageMap := createImageExistenceMap(test.foods)
			if !reflect.DeepEqual(test.expected, imageMap) {
				t.Errorf("expected: %#v, got: %#v", test.expected, imageMap)
			}
		})
	}
}

func TestCreateUsedPVCSet(t *testing.T) {
	tests := []struct {
		name     string
		dguests  []*v1alpha1.Dguest
		expected sets.String
	}{
		{
			name:     "empty dguests list",
			dguests:  []*v1alpha1.Dguest{},
			expected: sets.NewString(),
		},
		{
			name: "dguests not scheduled",
			dguests: []*v1alpha1.Dguest{
				st.MakeDguest().Name("foo").Namespace("foo").Obj(),
				st.MakeDguest().Name("bar").Namespace("bar").Obj(),
			},
			expected: sets.NewString(),
		},
		{
			name: "scheduled dguests that do not use any PVC",
			dguests: []*v1alpha1.Dguest{
				st.MakeDguest().Name("foo").Namespace("foo").Food("food-1").Obj(),
				st.MakeDguest().Name("bar").Namespace("bar").Food("food-2").Obj(),
			},
			expected: sets.NewString(),
		},
		{
			name: "scheduled dguests that use PVC",
			dguests: []*v1alpha1.Dguest{
				st.MakeDguest().Name("foo").Namespace("foo").Food("food-1").PVC("pvc1").Obj(),
				st.MakeDguest().Name("bar").Namespace("bar").Food("food-2").PVC("pvc2").Obj(),
			},
			expected: sets.NewString("foo/pvc1", "bar/pvc2"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			usedPVCs := createUsedPVCSet(test.dguests)
			if diff := cmp.Diff(test.expected, usedPVCs); diff != "" {
				t.Errorf("Unexpected usedPVCs (-want +got):\n%s", diff)
			}
		})
	}
}
