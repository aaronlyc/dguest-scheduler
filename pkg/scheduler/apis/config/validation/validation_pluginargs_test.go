/*
Copyright 2020 The Kubernetes Authors.

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

package validation

import (
	"fmt"
	"strings"
	"testing"

	"dguest-scheduler/pkg/scheduler/apis/config"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/apiserver/pkg/util/feature"
	"k8s.io/component-base/featuregate"
	featuregatetesting "k8s.io/component-base/featuregate/testing"
	"k8s.io/kubernetes/pkg/features"
)

var (
	ignoreBadValueDetail = cmpopts.IgnoreFields(field.Error{}, "BadValue", "Detail")
)

func TestValidateDefaultPreemptionArgs(t *testing.T) {
	cases := map[string]struct {
		args     config.DefaultPreemptionArgs
		wantErrs field.ErrorList
	}{
		"valid args (default)": {
			args: config.DefaultPreemptionArgs{
				MinCandidateFoodsPercentage: 10,
				MinCandidateFoodsAbsolute:   100,
			},
		},
		"negative minCandidateFoodsPercentage": {
			args: config.DefaultPreemptionArgs{
				MinCandidateFoodsPercentage: -1,
				MinCandidateFoodsAbsolute:   100,
			},
			wantErrs: field.ErrorList{
				&field.Error{
					Type:  field.ErrorTypeInvalid,
					Field: "minCandidateFoodsPercentage",
				},
			},
		},
		"minCandidateFoodsPercentage over 100": {
			args: config.DefaultPreemptionArgs{
				MinCandidateFoodsPercentage: 900,
				MinCandidateFoodsAbsolute:   100,
			},
			wantErrs: field.ErrorList{
				&field.Error{
					Type:  field.ErrorTypeInvalid,
					Field: "minCandidateFoodsPercentage",
				},
			},
		},
		"negative minCandidateFoodsAbsolute": {
			args: config.DefaultPreemptionArgs{
				MinCandidateFoodsPercentage: 20,
				MinCandidateFoodsAbsolute:   -1,
			},
			wantErrs: field.ErrorList{
				&field.Error{
					Type:  field.ErrorTypeInvalid,
					Field: "minCandidateFoodsAbsolute",
				},
			},
		},
		"all zero": {
			args: config.DefaultPreemptionArgs{
				MinCandidateFoodsPercentage: 0,
				MinCandidateFoodsAbsolute:   0,
			},
			wantErrs: field.ErrorList{
				&field.Error{
					Type:  field.ErrorTypeInvalid,
					Field: "minCandidateFoodsPercentage",
				}, &field.Error{
					Type:  field.ErrorTypeInvalid,
					Field: "minCandidateFoodsAbsolute",
				},
			},
		},
		"both negative": {
			args: config.DefaultPreemptionArgs{
				MinCandidateFoodsPercentage: -1,
				MinCandidateFoodsAbsolute:   -1,
			},
			wantErrs: field.ErrorList{
				&field.Error{
					Type:  field.ErrorTypeInvalid,
					Field: "minCandidateFoodsPercentage",
				}, &field.Error{
					Type:  field.ErrorTypeInvalid,
					Field: "minCandidateFoodsAbsolute",
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := ValidateDefaultPreemptionArgs(nil, &tc.args)
			if diff := cmp.Diff(tc.wantErrs.ToAggregate(), err, ignoreBadValueDetail); diff != "" {
				t.Errorf("ValidateDefaultPreemptionArgs returned err (-want,+got):\n%s", diff)
			}
		})
	}
}

func TestValidateInterDguestAffinityArgs(t *testing.T) {
	cases := map[string]struct {
		args    config.InterDguestAffinityArgs
		wantErr error
	}{
		"valid args": {
			args: config.InterDguestAffinityArgs{
				HardDguestAffinityWeight: 10,
			},
		},
		"hardDguestAffinityWeight less than min": {
			args: config.InterDguestAffinityArgs{
				HardDguestAffinityWeight: -1,
			},
			wantErr: &field.Error{
				Type:  field.ErrorTypeInvalid,
				Field: "hardDguestAffinityWeight",
			},
		},
		"hardDguestAffinityWeight more than max": {
			args: config.InterDguestAffinityArgs{
				HardDguestAffinityWeight: 101,
			},
			wantErr: &field.Error{
				Type:  field.ErrorTypeInvalid,
				Field: "hardDguestAffinityWeight",
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := ValidateInterDguestAffinityArgs(nil, &tc.args)
			if diff := cmp.Diff(tc.wantErr, err, ignoreBadValueDetail); diff != "" {
				t.Errorf("ValidateInterDguestAffinityArgs returned err (-want,+got):\n%s", diff)
			}
		})
	}
}

func TestValidateDguestTopologySpreadArgs(t *testing.T) {
	cases := map[string]struct {
		args     *config.DguestTopologySpreadArgs
		wantErrs field.ErrorList
	}{
		"valid config": {
			args: &config.DguestTopologySpreadArgs{
				DefaultConstraints: []v1.TopologySpreadConstraint{
					{
						MaxSkew:           1,
						TopologyKey:       "food",
						WhenUnsatisfiable: v1.DoNotSchedule,
					},
					{
						MaxSkew:           2,
						TopologyKey:       "zone",
						WhenUnsatisfiable: v1.ScheduleAnyway,
					},
				},
				DefaultingType: config.ListDefaulting,
			},
		},
		"maxSkew less than zero": {
			args: &config.DguestTopologySpreadArgs{
				DefaultConstraints: []v1.TopologySpreadConstraint{
					{
						MaxSkew:           -1,
						TopologyKey:       "food",
						WhenUnsatisfiable: v1.DoNotSchedule,
					},
				},
				DefaultingType: config.ListDefaulting,
			},
			wantErrs: field.ErrorList{
				&field.Error{
					Type:  field.ErrorTypeInvalid,
					Field: "defaultConstraints[0].maxSkew",
				},
			},
		},
		"empty topology key": {
			args: &config.DguestTopologySpreadArgs{
				DefaultConstraints: []v1.TopologySpreadConstraint{
					{
						MaxSkew:           1,
						TopologyKey:       "",
						WhenUnsatisfiable: v1.DoNotSchedule,
					},
				},
				DefaultingType: config.ListDefaulting,
			},
			wantErrs: field.ErrorList{
				&field.Error{
					Type:  field.ErrorTypeRequired,
					Field: "defaultConstraints[0].topologyKey",
				},
			},
		},
		"whenUnsatisfiable is empty": {
			args: &config.DguestTopologySpreadArgs{
				DefaultConstraints: []v1.TopologySpreadConstraint{
					{
						MaxSkew:           1,
						TopologyKey:       "food",
						WhenUnsatisfiable: "",
					},
				},
				DefaultingType: config.ListDefaulting,
			},
			wantErrs: field.ErrorList{
				&field.Error{
					Type:  field.ErrorTypeRequired,
					Field: "defaultConstraints[0].whenUnsatisfiable",
				},
			},
		},
		"whenUnsatisfiable contains unsupported action": {
			args: &config.DguestTopologySpreadArgs{
				DefaultConstraints: []v1.TopologySpreadConstraint{
					{
						MaxSkew:           1,
						TopologyKey:       "food",
						WhenUnsatisfiable: "unknown action",
					},
				},
				DefaultingType: config.ListDefaulting,
			},
			wantErrs: field.ErrorList{
				&field.Error{
					Type:  field.ErrorTypeNotSupported,
					Field: "defaultConstraints[0].whenUnsatisfiable",
				},
			},
		},
		"duplicated constraints": {
			args: &config.DguestTopologySpreadArgs{
				DefaultConstraints: []v1.TopologySpreadConstraint{
					{
						MaxSkew:           1,
						TopologyKey:       "food",
						WhenUnsatisfiable: v1.DoNotSchedule,
					},
					{
						MaxSkew:           2,
						TopologyKey:       "food",
						WhenUnsatisfiable: v1.DoNotSchedule,
					},
				},
				DefaultingType: config.ListDefaulting,
			},
			wantErrs: field.ErrorList{
				&field.Error{
					Type:  field.ErrorTypeDuplicate,
					Field: "defaultConstraints[1]",
				},
			},
		},
		"label selector present": {
			args: &config.DguestTopologySpreadArgs{
				DefaultConstraints: []v1.TopologySpreadConstraint{
					{
						MaxSkew:           1,
						TopologyKey:       "key",
						WhenUnsatisfiable: v1.DoNotSchedule,
						LabelSelector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"a": "b",
							},
						},
					},
				},
				DefaultingType: config.ListDefaulting,
			},
			wantErrs: field.ErrorList{
				&field.Error{
					Type:  field.ErrorTypeForbidden,
					Field: "defaultConstraints[0].labelSelector",
				},
			},
		},
		"list default constraints, no constraints": {
			args: &config.DguestTopologySpreadArgs{
				DefaultingType: config.ListDefaulting,
			},
		},
		"system default constraints": {
			args: &config.DguestTopologySpreadArgs{
				DefaultingType: config.SystemDefaulting,
			},
		},
		"wrong constraints": {
			args: &config.DguestTopologySpreadArgs{
				DefaultingType: "unknown",
			},
			wantErrs: field.ErrorList{
				&field.Error{
					Type:  field.ErrorTypeNotSupported,
					Field: "defaultingType",
				},
			},
		},
		"system default constraints, but has constraints": {
			args: &config.DguestTopologySpreadArgs{
				DefaultConstraints: []v1.TopologySpreadConstraint{
					{
						MaxSkew:           1,
						TopologyKey:       "key",
						WhenUnsatisfiable: v1.DoNotSchedule,
					},
				},
				DefaultingType: config.SystemDefaulting,
			},
			wantErrs: field.ErrorList{
				&field.Error{
					Type:  field.ErrorTypeInvalid,
					Field: "defaultingType",
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := ValidateDguestTopologySpreadArgs(nil, tc.args)
			if diff := cmp.Diff(tc.wantErrs.ToAggregate(), err, ignoreBadValueDetail); diff != "" {
				t.Errorf("ValidateDguestTopologySpreadArgs returned err (-want,+got):\n%s", diff)
			}
		})
	}
}

func TestValidateFoodResourcesBalancedAllocationArgs(t *testing.T) {
	cases := map[string]struct {
		args     *config.FoodResourcesBalancedAllocationArgs
		wantErrs field.ErrorList
	}{
		"valid config": {
			args: &config.FoodResourcesBalancedAllocationArgs{
				Resources: []config.ResourceSpec{
					{
						Name:   "cpu",
						Weight: 1,
					},
					{
						Name:   "memory",
						Weight: 1,
					},
				},
			},
		},
		"invalid config": {
			args: &config.FoodResourcesBalancedAllocationArgs{
				Resources: []config.ResourceSpec{
					{
						Name:   "cpu",
						Weight: 2,
					},
					{
						Name:   "memory",
						Weight: 1,
					},
				},
			},
			wantErrs: field.ErrorList{
				&field.Error{
					Type:  field.ErrorTypeInvalid,
					Field: "resources[0].weight",
				},
			},
		},
		"repeated resources": {
			args: &config.FoodResourcesBalancedAllocationArgs{
				Resources: []config.ResourceSpec{
					{
						Name:   "cpu",
						Weight: 1,
					},
					{
						Name:   "cpu",
						Weight: 1,
					},
				},
			},
			wantErrs: field.ErrorList{
				&field.Error{
					Type:  field.ErrorTypeDuplicate,
					Field: "resources[1].name",
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := ValidateFoodResourcesBalancedAllocationArgs(nil, tc.args)
			if diff := cmp.Diff(tc.wantErrs.ToAggregate(), err, ignoreBadValueDetail); diff != "" {
				t.Errorf("ValidateFoodResourcesBalancedAllocationArgs returned err (-want,+got):\n%s", diff)
			}
		})
	}
}

func TestValidateFoodAffinityArgs(t *testing.T) {
	cases := []struct {
		name    string
		args    config.FoodAffinityArgs
		wantErr error
	}{
		{
			name: "empty",
		},
		{
			name: "valid added affinity",
			args: config.FoodAffinityArgs{
				AddedAffinity: &v1alpha1.FoodAffinity{
					RequiredDuringSchedulingIgnoredDuringExecution: &v1alpha1.FoodSelector{
						FoodSelectorTerms: []v1alpha1.FoodSelectorTerm{
							{
								MatchExpressions: []v1alpha1.FoodSelectorRequirement{
									{
										Key:      "label-1",
										Operator: v1alpha1.FoodSelectorOpIn,
										Values:   []string{"label-1-val"},
									},
								},
							},
						},
					},
					PreferredDuringSchedulingIgnoredDuringExecution: []v1.PreferredSchedulingTerm{
						{
							Weight: 1,
							Preference: v1alpha1.FoodSelectorTerm{
								MatchFields: []v1alpha1.FoodSelectorRequirement{
									{
										Key:      "metadata.name",
										Operator: v1alpha1.FoodSelectorOpIn,
										Values:   []string{"food-1"},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "invalid added affinity",
			args: config.FoodAffinityArgs{
				AddedAffinity: &v1alpha1.FoodAffinity{
					RequiredDuringSchedulingIgnoredDuringExecution: &v1alpha1.FoodSelector{
						FoodSelectorTerms: []v1alpha1.FoodSelectorTerm{
							{
								MatchExpressions: []v1alpha1.FoodSelectorRequirement{
									{
										Key:      "invalid/label/key",
										Operator: v1alpha1.FoodSelectorOpIn,
										Values:   []string{"label-1-val"},
									},
								},
							},
						},
					},
					PreferredDuringSchedulingIgnoredDuringExecution: []v1.PreferredSchedulingTerm{
						{
							Weight: 1,
							Preference: v1alpha1.FoodSelectorTerm{
								MatchFields: []v1alpha1.FoodSelectorRequirement{
									{
										Key:      "metadata.name",
										Operator: v1alpha1.FoodSelectorOpIn,
										Values:   []string{"food-1", "food-2"},
									},
								},
							},
						},
					},
				},
			},
			wantErr: field.ErrorList{
				&field.Error{
					Type:  field.ErrorTypeInvalid,
					Field: "addedAffinity.requiredDuringSchedulingIgnoredDuringExecution.foodSelectorTerms[0].matchExpressions[0].key",
				},
				&field.Error{
					Type:  field.ErrorTypeInvalid,
					Field: "addedAffinity.preferredDuringSchedulingIgnoredDuringExecution[0].matchFields[0].values",
				},
			}.ToAggregate(),
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateFoodAffinityArgs(nil, &tc.args)
			if diff := cmp.Diff(tc.wantErr, err, ignoreBadValueDetail); diff != "" {
				t.Errorf("ValidatedFoodAffinityArgs returned err (-want,+got):\n%s", diff)
			}
		})
	}
}

func TestValidateVolumeBindingArgs(t *testing.T) {
	cases := []struct {
		name     string
		args     config.VolumeBindingArgs
		features map[featuregate.Feature]bool
		wantErr  error
	}{
		{
			name: "zero is a valid config",
			args: config.VolumeBindingArgs{
				BindTimeoutSeconds: 0,
			},
		},
		{
			name: "positive value is valid config",
			args: config.VolumeBindingArgs{
				BindTimeoutSeconds: 10,
			},
		},
		{
			name: "negative value is invalid config ",
			args: config.VolumeBindingArgs{
				BindTimeoutSeconds: -10,
			},
			wantErr: errors.NewAggregate([]error{&field.Error{
				Type:     field.ErrorTypeInvalid,
				Field:    "bindTimeoutSeconds",
				BadValue: int64(-10),
				Detail:   "invalid BindTimeoutSeconds, should not be a negative value",
			}}),
		},
		{
			name: "[VolumeCapacityPriority=off] shape should be nil when the feature is off",
			features: map[featuregate.Feature]bool{
				features.VolumeCapacityPriority: false,
			},
			args: config.VolumeBindingArgs{
				BindTimeoutSeconds: 10,
				Shape:              nil,
			},
		},
		{
			name: "[VolumeCapacityPriority=off] error if the shape is not nil when the feature is off",
			features: map[featuregate.Feature]bool{
				features.VolumeCapacityPriority: false,
			},
			args: config.VolumeBindingArgs{
				BindTimeoutSeconds: 10,
				Shape: []config.UtilizationShapePoint{
					{Utilization: 1, Score: 1},
					{Utilization: 3, Score: 3},
				},
			},
			wantErr: errors.NewAggregate([]error{&field.Error{
				Type:  field.ErrorTypeInvalid,
				Field: "shape",
			}}),
		},
		{
			name: "[VolumeCapacityPriority=on] shape should not be empty",
			features: map[featuregate.Feature]bool{
				features.VolumeCapacityPriority: true,
			},
			args: config.VolumeBindingArgs{
				BindTimeoutSeconds: 10,
				Shape:              []config.UtilizationShapePoint{},
			},
			wantErr: errors.NewAggregate([]error{&field.Error{
				Type:  field.ErrorTypeRequired,
				Field: "shape",
			}}),
		},
		{
			name: "[VolumeCapacityPriority=on] shape points must be sorted in increasing order",
			features: map[featuregate.Feature]bool{
				features.VolumeCapacityPriority: true,
			},
			args: config.VolumeBindingArgs{
				BindTimeoutSeconds: 10,
				Shape: []config.UtilizationShapePoint{
					{Utilization: 3, Score: 3},
					{Utilization: 1, Score: 1},
				},
			},
			wantErr: errors.NewAggregate([]error{&field.Error{
				Type:   field.ErrorTypeInvalid,
				Field:  "shape[1].utilization",
				Detail: "Invalid value: 1: utilization values must be sorted in increasing order",
			}}),
		},
		{
			name: "[VolumeCapacityPriority=on] shape point: invalid utilization and score",
			features: map[featuregate.Feature]bool{
				features.VolumeCapacityPriority: true,
			},
			args: config.VolumeBindingArgs{
				BindTimeoutSeconds: 10,
				Shape: []config.UtilizationShapePoint{
					{Utilization: -1, Score: 1},
					{Utilization: 10, Score: -1},
					{Utilization: 20, Score: 11},
					{Utilization: 101, Score: 1},
				},
			},
			wantErr: errors.NewAggregate([]error{
				&field.Error{
					Type:  field.ErrorTypeInvalid,
					Field: "shape[0].utilization",
				},
				&field.Error{
					Type:  field.ErrorTypeInvalid,
					Field: "shape[1].score",
				},
				&field.Error{
					Type:  field.ErrorTypeInvalid,
					Field: "shape[2].score",
				},
				&field.Error{
					Type:  field.ErrorTypeInvalid,
					Field: "shape[3].utilization",
				},
			}),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for k, v := range tc.features {
				defer featuregatetesting.SetFeatureGateDuringTest(t, feature.DefaultFeatureGate, k, v)()
			}
			err := ValidateVolumeBindingArgs(nil, &tc.args)
			if diff := cmp.Diff(tc.wantErr, err, ignoreBadValueDetail); diff != "" {
				t.Errorf("ValidateVolumeBindingArgs returned err (-want,+got):\n%s", diff)
			}
		})
	}
}

func TestValidateFitArgs(t *testing.T) {
	defaultScoringStrategy := &config.ScoringStrategy{
		Type: config.LeastAllocated,
		Resources: []config.ResourceSpec{
			{Name: "cpu", Weight: 1},
			{Name: "memory", Weight: 1},
		},
	}
	argsTest := []struct {
		name   string
		args   config.FoodResourcesFitArgs
		expect string
	}{
		{
			name: "IgnoredResources: too long value",
			args: config.FoodResourcesFitArgs{
				IgnoredResources: []string{fmt.Sprintf("longvalue%s", strings.Repeat("a", 64))},
				ScoringStrategy:  defaultScoringStrategy,
			},
			expect: "name part must be no more than 63 characters",
		},
		{
			name: "IgnoredResources: name is empty",
			args: config.FoodResourcesFitArgs{
				IgnoredResources: []string{"example.com/"},
				ScoringStrategy:  defaultScoringStrategy,
			},
			expect: "name part must be non-empty",
		},
		{
			name: "IgnoredResources: name has too many slash",
			args: config.FoodResourcesFitArgs{
				IgnoredResources: []string{"example.com/aaa/bbb"},
				ScoringStrategy:  defaultScoringStrategy,
			},
			expect: "a qualified name must consist of alphanumeric characters",
		},
		{
			name: "IgnoredResources: valid args",
			args: config.FoodResourcesFitArgs{
				IgnoredResources: []string{"example.com"},
				ScoringStrategy:  defaultScoringStrategy,
			},
		},
		{
			name: "IgnoredResourceGroups: valid args ",
			args: config.FoodResourcesFitArgs{
				IgnoredResourceGroups: []string{"example.com"},
				ScoringStrategy:       defaultScoringStrategy,
			},
		},
		{
			name: "IgnoredResourceGroups: illegal args",
			args: config.FoodResourcesFitArgs{
				IgnoredResourceGroups: []string{"example.com/"},
				ScoringStrategy:       defaultScoringStrategy,
			},
			expect: "name part must be non-empty",
		},
		{
			name: "IgnoredResourceGroups: name is too long",
			args: config.FoodResourcesFitArgs{
				IgnoredResourceGroups: []string{strings.Repeat("a", 64)},
				ScoringStrategy:       defaultScoringStrategy,
			},
			expect: "name part must be no more than 63 characters",
		},
		{
			name: "IgnoredResourceGroups: name cannot be contain slash",
			args: config.FoodResourcesFitArgs{
				IgnoredResourceGroups: []string{"example.com/aa"},
				ScoringStrategy:       defaultScoringStrategy,
			},
			expect: "resource group name can't contain '/'",
		},
		{
			name:   "ScoringStrategy: field is required",
			args:   config.FoodResourcesFitArgs{},
			expect: "ScoringStrategy field is required",
		},
		{
			name: "ScoringStrategy: type is unsupported",
			args: config.FoodResourcesFitArgs{
				ScoringStrategy: &config.ScoringStrategy{
					Type: "Invalid",
				},
			},
			expect: `Unsupported value: "Invalid"`,
		},
	}

	for _, test := range argsTest {
		t.Run(test.name, func(t *testing.T) {
			if err := ValidateFoodResourcesFitArgs(nil, &test.args); err != nil && (!strings.Contains(err.Error(), test.expect)) {
				t.Errorf("case[%v]: error details do not include %v", test.name, err)
			}
		})
	}
}

func TestValidateLeastAllocatedScoringStrategy(t *testing.T) {
	tests := []struct {
		name      string
		resources []config.ResourceSpec
		wantErrs  field.ErrorList
	}{
		{
			name:     "default config",
			wantErrs: nil,
		},
		{
			name: "multi valid resources",
			resources: []config.ResourceSpec{
				{
					Name:   "cpu",
					Weight: 1,
				},
				{
					Name:   "memory",
					Weight: 10,
				},
			},
			wantErrs: nil,
		},
		{
			name: "weight less than min",
			resources: []config.ResourceSpec{
				{
					Name:   "cpu",
					Weight: 0,
				},
			},
			wantErrs: field.ErrorList{
				{
					Type:  field.ErrorTypeInvalid,
					Field: "scoringStrategy.resources[0].weight",
				},
			},
		},
		{
			name: "weight greater than max",
			resources: []config.ResourceSpec{
				{
					Name:   "cpu",
					Weight: 101,
				},
			},
			wantErrs: field.ErrorList{
				{
					Type:  field.ErrorTypeInvalid,
					Field: "scoringStrategy.resources[0].weight",
				},
			},
		},
		{
			name: "multi invalid resources",
			resources: []config.ResourceSpec{
				{
					Name:   "cpu",
					Weight: 0,
				},
				{
					Name:   "memory",
					Weight: 101,
				},
			},
			wantErrs: field.ErrorList{
				{
					Type:  field.ErrorTypeInvalid,
					Field: "scoringStrategy.resources[0].weight",
				},
				{
					Type:  field.ErrorTypeInvalid,
					Field: "scoringStrategy.resources[1].weight",
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			args := config.FoodResourcesFitArgs{
				ScoringStrategy: &config.ScoringStrategy{
					Type:      config.LeastAllocated,
					Resources: test.resources,
				},
			}
			err := ValidateFoodResourcesFitArgs(nil, &args)
			if diff := cmp.Diff(test.wantErrs.ToAggregate(), err, ignoreBadValueDetail); diff != "" {
				t.Errorf("ValidateFoodResourcesFitArgs returned err (-want,+got):\n%s", diff)
			}
		})
	}
}

func TestValidateMostAllocatedScoringStrategy(t *testing.T) {
	tests := []struct {
		name      string
		resources []config.ResourceSpec
		wantErrs  field.ErrorList
	}{
		{
			name:     "default config",
			wantErrs: nil,
		},
		{
			name: "multi valid resources",
			resources: []config.ResourceSpec{
				{
					Name:   "cpu",
					Weight: 1,
				},
				{
					Name:   "memory",
					Weight: 10,
				},
			},
			wantErrs: nil,
		},
		{
			name: "weight less than min",
			resources: []config.ResourceSpec{
				{
					Name:   "cpu",
					Weight: 0,
				},
			},
			wantErrs: field.ErrorList{
				{
					Type:  field.ErrorTypeInvalid,
					Field: "scoringStrategy.resources[0].weight",
				},
			},
		},
		{
			name: "weight greater than max",
			resources: []config.ResourceSpec{
				{
					Name:   "cpu",
					Weight: 101,
				},
			},
			wantErrs: field.ErrorList{
				{
					Type:  field.ErrorTypeInvalid,
					Field: "scoringStrategy.resources[0].weight",
				},
			},
		},
		{
			name: "multi invalid resources",
			resources: []config.ResourceSpec{
				{
					Name:   "cpu",
					Weight: 0,
				},
				{
					Name:   "memory",
					Weight: 101,
				},
			},
			wantErrs: field.ErrorList{
				{
					Type:  field.ErrorTypeInvalid,
					Field: "scoringStrategy.resources[0].weight",
				},
				{
					Type:  field.ErrorTypeInvalid,
					Field: "scoringStrategy.resources[1].weight",
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			args := config.FoodResourcesFitArgs{
				ScoringStrategy: &config.ScoringStrategy{
					Type:      config.MostAllocated,
					Resources: test.resources,
				},
			}
			err := ValidateFoodResourcesFitArgs(nil, &args)
			if diff := cmp.Diff(test.wantErrs.ToAggregate(), err, ignoreBadValueDetail); diff != "" {
				t.Errorf("ValidateFoodResourcesFitArgs returned err (-want,+got):\n%s", diff)
			}
		})
	}
}

func TestValidateRequestedToCapacityRatioScoringStrategy(t *testing.T) {
	defaultShape := []config.UtilizationShapePoint{
		{
			Utilization: 30,
			Score:       3,
		},
	}
	tests := []struct {
		name      string
		resources []config.ResourceSpec
		shapes    []config.UtilizationShapePoint
		wantErrs  field.ErrorList
	}{
		{
			name:   "no shapes",
			shapes: nil,
			wantErrs: field.ErrorList{
				{
					Type:  field.ErrorTypeRequired,
					Field: "scoringStrategy.shape",
				},
			},
		},
		{
			name:   "weight greater than max",
			shapes: defaultShape,
			resources: []config.ResourceSpec{
				{
					Name:   "cpu",
					Weight: 101,
				},
			},
			wantErrs: field.ErrorList{
				{
					Type:  field.ErrorTypeInvalid,
					Field: "scoringStrategy.resources[0].weight",
				},
			},
		},
		{
			name:   "weight less than min",
			shapes: defaultShape,
			resources: []config.ResourceSpec{
				{
					Name:   "cpu",
					Weight: 0,
				},
			},
			wantErrs: field.ErrorList{
				{
					Type:  field.ErrorTypeInvalid,
					Field: "scoringStrategy.resources[0].weight",
				},
			},
		},
		{
			name:     "valid shapes",
			shapes:   defaultShape,
			wantErrs: nil,
		},
		{
			name: "utilization less than min",
			shapes: []config.UtilizationShapePoint{
				{
					Utilization: -1,
					Score:       3,
				},
			},
			wantErrs: field.ErrorList{
				{
					Type:  field.ErrorTypeInvalid,
					Field: "scoringStrategy.shape[0].utilization",
				},
			},
		},
		{
			name: "utilization greater than max",
			shapes: []config.UtilizationShapePoint{
				{
					Utilization: 101,
					Score:       3,
				},
			},
			wantErrs: field.ErrorList{
				{
					Type:  field.ErrorTypeInvalid,
					Field: "scoringStrategy.shape[0].utilization",
				},
			},
		},
		{
			name: "duplicated utilization values",
			shapes: []config.UtilizationShapePoint{
				{
					Utilization: 10,
					Score:       3,
				},
				{
					Utilization: 10,
					Score:       3,
				},
			},
			wantErrs: field.ErrorList{
				{
					Type:  field.ErrorTypeInvalid,
					Field: "scoringStrategy.shape[1].utilization",
				},
			},
		},
		{
			name: "increasing utilization values",
			shapes: []config.UtilizationShapePoint{
				{
					Utilization: 10,
					Score:       3,
				},
				{
					Utilization: 20,
					Score:       3,
				},
				{
					Utilization: 30,
					Score:       3,
				},
			},
			wantErrs: nil,
		},
		{
			name: "non-increasing utilization values",
			shapes: []config.UtilizationShapePoint{
				{
					Utilization: 10,
					Score:       3,
				},
				{
					Utilization: 20,
					Score:       3,
				},
				{
					Utilization: 15,
					Score:       3,
				},
			},
			wantErrs: field.ErrorList{
				{
					Type:  field.ErrorTypeInvalid,
					Field: "scoringStrategy.shape[2].utilization",
				},
			},
		},
		{
			name: "score less than min",
			shapes: []config.UtilizationShapePoint{
				{
					Utilization: 10,
					Score:       -1,
				},
			},
			wantErrs: field.ErrorList{
				{
					Type:  field.ErrorTypeInvalid,
					Field: "scoringStrategy.shape[0].score",
				},
			},
		},
		{
			name: "score greater than max",
			shapes: []config.UtilizationShapePoint{
				{
					Utilization: 10,
					Score:       11,
				},
			},
			wantErrs: field.ErrorList{
				{
					Type:  field.ErrorTypeInvalid,
					Field: "scoringStrategy.shape[0].score",
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			args := config.FoodResourcesFitArgs{
				ScoringStrategy: &config.ScoringStrategy{
					Type:      config.RequestedToCapacityRatio,
					Resources: test.resources,
					RequestedToCapacityRatio: &config.RequestedToCapacityRatioParam{
						Shape: test.shapes,
					},
				},
			}
			err := ValidateFoodResourcesFitArgs(nil, &args)
			if diff := cmp.Diff(test.wantErrs.ToAggregate(), err, ignoreBadValueDetail); diff != "" {
				t.Errorf("ValidateFoodResourcesFitArgs returned err (-want,+got):\n%s", diff)
			}
		})
	}
}
