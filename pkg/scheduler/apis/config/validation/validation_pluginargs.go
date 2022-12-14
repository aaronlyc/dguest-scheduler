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
	v12 "dguest-scheduler/pkg/scheduler/apis/config/v1"
	"fmt"
	"strings"

	v1 "k8s.io/api/core/v1"
	metav1validation "k8s.io/apimachinery/pkg/apis/meta/v1/validation"
	"k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	"k8s.io/kubernetes/pkg/features"
)

var supportedScoringStrategyTypes = sets.NewString(
	string(v12.LeastAllocated),
	string(v12.MostAllocated),
	string(v12.RequestedToCapacityRatio),
)

// ValidateDefaultPreemptionArgs validates that DefaultPreemptionArgs are correct.
func ValidateDefaultPreemptionArgs(path *field.Path, args *v12.DefaultPreemptionArgs) error {
	var allErrs field.ErrorList
	percentagePath := path.Child("minCandidateFoodsPercentage")
	absolutePath := path.Child("minCandidateFoodsAbsolute")
	if err := validateMinCandidateFoodsPercentage(args.MinCandidateFoodsPercentage, percentagePath); err != nil {
		allErrs = append(allErrs, err)
	}
	if err := validateMinCandidateFoodsAbsolute(args.MinCandidateFoodsAbsolute, absolutePath); err != nil {
		allErrs = append(allErrs, err)
	}
	if args.MinCandidateFoodsPercentage == 0 && args.MinCandidateFoodsAbsolute == 0 {
		allErrs = append(allErrs,
			field.Invalid(percentagePath, args.MinCandidateFoodsPercentage, "cannot be zero at the same time as minCandidateFoodsAbsolute"),
			field.Invalid(absolutePath, args.MinCandidateFoodsAbsolute, "cannot be zero at the same time as minCandidateFoodsPercentage"))
	}
	return allErrs.ToAggregate()
}

// validateMinCandidateFoodsPercentage validates that
// minCandidateFoodsPercentage is within the allowed range.
func validateMinCandidateFoodsPercentage(minCandidateFoodsPercentage int32, p *field.Path) *field.Error {
	if minCandidateFoodsPercentage < 0 || minCandidateFoodsPercentage > 100 {
		return field.Invalid(p, minCandidateFoodsPercentage, "not in valid range [0, 100]")
	}
	return nil
}

// validateMinCandidateFoodsAbsolute validates that minCandidateFoodsAbsolute
// is within the allowed range.
func validateMinCandidateFoodsAbsolute(minCandidateFoodsAbsolute int32, p *field.Path) *field.Error {
	if minCandidateFoodsAbsolute < 0 {
		return field.Invalid(p, minCandidateFoodsAbsolute, "not in valid range [0, inf)")
	}
	return nil
}

// ValidateInterDguestAffinityArgs validates that InterDguestAffinityArgs are correct.
func ValidateInterDguestAffinityArgs(path *field.Path, args *v12.InterDguestAffinityArgs) error {
	return validateHardDguestAffinityWeight(path.Child("hardDguestAffinityWeight"), args.HardDguestAffinityWeight)
}

// validateHardDguestAffinityWeight validates that weight is within allowed range.
func validateHardDguestAffinityWeight(path *field.Path, w int32) error {
	const (
		minHardDguestAffinityWeight = 0
		maxHardDguestAffinityWeight = 100
	)

	if w < minHardDguestAffinityWeight || w > maxHardDguestAffinityWeight {
		msg := fmt.Sprintf("not in valid range [%d, %d]", minHardDguestAffinityWeight, maxHardDguestAffinityWeight)
		return field.Invalid(path, w, msg)
	}
	return nil
}

// ValidateDguestTopologySpreadArgs validates that DguestTopologySpreadArgs are correct.
// It replicates the validation from pkg/apis/core/validation.validateTopologySpreadConstraints
// with an additional check for .labelSelector to be nil.
func ValidateDguestTopologySpreadArgs(path *field.Path, args *v12.DguestTopologySpreadArgs) error {
	var allErrs field.ErrorList
	if err := validateDefaultingType(path.Child("defaultingType"), args.DefaultingType, args.DefaultConstraints); err != nil {
		allErrs = append(allErrs, err)
	}

	defaultConstraintsPath := path.Child("defaultConstraints")
	for i, c := range args.DefaultConstraints {
		p := defaultConstraintsPath.Index(i)
		if c.MaxSkew <= 0 {
			f := p.Child("maxSkew")
			allErrs = append(allErrs, field.Invalid(f, c.MaxSkew, "not in valid range (0, inf)"))
		}
		allErrs = append(allErrs, validateTopologyKey(p.Child("topologyKey"), c.TopologyKey)...)
		if err := validateWhenUnsatisfiable(p.Child("whenUnsatisfiable"), c.WhenUnsatisfiable); err != nil {
			allErrs = append(allErrs, err)
		}
		if c.LabelSelector != nil {
			f := field.Forbidden(p.Child("labelSelector"), "constraint must not define a selector, as they deduced for each dguest")
			allErrs = append(allErrs, f)
		}
		if err := validateConstraintNotRepeat(defaultConstraintsPath, args.DefaultConstraints, i); err != nil {
			allErrs = append(allErrs, err)
		}
	}
	if len(allErrs) == 0 {
		return nil
	}
	return allErrs.ToAggregate()
}

func validateDefaultingType(p *field.Path, v v12.DguestTopologySpreadConstraintsDefaulting, constraints []v1.TopologySpreadConstraint) *field.Error {
	if v != v12.SystemDefaulting && v != v12.ListDefaulting {
		return field.NotSupported(p, v, []string{string(v12.SystemDefaulting), string(v12.ListDefaulting)})
	}
	if v == v12.SystemDefaulting && len(constraints) > 0 {
		return field.Invalid(p, v, "when .defaultConstraints are not empty")
	}
	return nil
}

func validateTopologyKey(p *field.Path, v string) field.ErrorList {
	var allErrs field.ErrorList
	if len(v) == 0 {
		allErrs = append(allErrs, field.Required(p, "can not be empty"))
	} else {
		allErrs = append(allErrs, metav1validation.ValidateLabelName(v, p)...)
	}
	return allErrs
}

func validateWhenUnsatisfiable(p *field.Path, v v1.UnsatisfiableConstraintAction) *field.Error {
	supportedScheduleActions := sets.NewString(string(v1.DoNotSchedule), string(v1.ScheduleAnyway))

	if len(v) == 0 {
		return field.Required(p, "can not be empty")
	}
	if !supportedScheduleActions.Has(string(v)) {
		return field.NotSupported(p, v, supportedScheduleActions.List())
	}
	return nil
}

func validateConstraintNotRepeat(path *field.Path, constraints []v1.TopologySpreadConstraint, idx int) *field.Error {
	c := &constraints[idx]
	for i := range constraints[:idx] {
		other := &constraints[i]
		if c.TopologyKey == other.TopologyKey && c.WhenUnsatisfiable == other.WhenUnsatisfiable {
			return field.Duplicate(path.Index(idx), fmt.Sprintf("{%v, %v}", c.TopologyKey, c.WhenUnsatisfiable))
		}
	}
	return nil
}

func validateFunctionShape(shape []v12.UtilizationShapePoint, path *field.Path) field.ErrorList {
	const (
		minUtilization = 0
		maxUtilization = 100
		minScore       = 0
		maxScore       = int32(v12.MaxCustomPriorityScore)
	)

	var allErrs field.ErrorList

	if len(shape) == 0 {
		allErrs = append(allErrs, field.Required(path, "at least one point must be specified"))
		return allErrs
	}

	for i := 1; i < len(shape); i++ {
		if shape[i-1].Utilization >= shape[i].Utilization {
			allErrs = append(allErrs, field.Invalid(path.Index(i).Child("utilization"), shape[i].Utilization, "utilization values must be sorted in increasing order"))
			break
		}
	}

	for i, point := range shape {
		if point.Utilization < minUtilization || point.Utilization > maxUtilization {
			msg := fmt.Sprintf("not in valid range [%d, %d]", minUtilization, maxUtilization)
			allErrs = append(allErrs, field.Invalid(path.Index(i).Child("utilization"), point.Utilization, msg))
		}

		if point.Score < minScore || point.Score > maxScore {
			msg := fmt.Sprintf("not in valid range [%d, %d]", minScore, maxScore)
			allErrs = append(allErrs, field.Invalid(path.Index(i).Child("score"), point.Score, msg))
		}
	}

	return allErrs
}

func validateResources(resources []v12.ResourceSpec, p *field.Path) field.ErrorList {
	var allErrs field.ErrorList
	for i, resource := range resources {
		if resource.Weight <= 0 || resource.Weight > 100 {
			msg := fmt.Sprintf("resource weight of %v not in valid range (0, 100]", resource.Name)
			allErrs = append(allErrs, field.Invalid(p.Index(i).Child("weight"), resource.Weight, msg))
		}
	}
	return allErrs
}

// ValidateFoodResourcesBalancedAllocationArgs validates that FoodResourcesBalancedAllocationArgs are set correctly.
func ValidateFoodResourcesBalancedAllocationArgs(path *field.Path, args *v12.FoodResourcesBalancedAllocationArgs) error {
	var allErrs field.ErrorList
	seenResources := sets.NewString()
	for i, resource := range args.Resources {
		if seenResources.Has(resource.Name) {
			allErrs = append(allErrs, field.Duplicate(path.Child("resources").Index(i).Child("name"), resource.Name))
		} else {
			seenResources.Insert(resource.Name)
		}
		if resource.Weight != 1 {
			allErrs = append(allErrs, field.Invalid(path.Child("resources").Index(i).Child("weight"), resource.Weight, "must be 1"))
		}
	}
	return allErrs.ToAggregate()
}

// ValidateFoodAffinityArgs validates that FoodAffinityArgs are correct.
func ValidateFoodAffinityArgs(path *field.Path, args *v12.FoodAffinityArgs) error {
	//if args.AddedAffinity == nil {
	//	return nil
	//}
	//affinity := args.AddedAffinity
	//var errs []error
	//if ns := affinity.RequiredDuringSchedulingIgnoredDuringExecution; ns != nil {
	//	_, err := foodaffinity.NewFoodSelector(ns, field.WithPath(path.Child("addedAffinity", "requiredDuringSchedulingIgnoredDuringExecution")))
	//	if err != nil {
	//		errs = append(errs, err)
	//	}
	//}
	//// TODO: Add validation for requiredDuringSchedulingRequiredDuringExecution when it gets added to the API.
	//if terms := affinity.PreferredDuringSchedulingIgnoredDuringExecution; len(terms) != 0 {
	//	_, err := foodaffinity.NewPreferredSchedulingTerms(terms, field.WithPath(path.Child("addedAffinity", "preferredDuringSchedulingIgnoredDuringExecution")))
	//	if err != nil {
	//		errs = append(errs, err)
	//	}
	//}
	return errors.Flatten(errors.NewAggregate(nil))
}

// VolumeBindingArgsValidationOptions contains the different settings for validation.
type VolumeBindingArgsValidationOptions struct {
	AllowVolumeCapacityPriority bool
}

// ValidateVolumeBindingArgs validates that VolumeBindingArgs are set correctly.
func ValidateVolumeBindingArgs(path *field.Path, args *v12.VolumeBindingArgs) error {
	return ValidateVolumeBindingArgsWithOptions(path, args, VolumeBindingArgsValidationOptions{
		AllowVolumeCapacityPriority: utilfeature.DefaultFeatureGate.Enabled(features.VolumeCapacityPriority),
	})
}

// ValidateVolumeBindingArgs validates that VolumeBindingArgs with scheduler features.
func ValidateVolumeBindingArgsWithOptions(path *field.Path, args *v12.VolumeBindingArgs, opts VolumeBindingArgsValidationOptions) error {
	var allErrs field.ErrorList

	if args.BindTimeoutSeconds < 0 {
		allErrs = append(allErrs, field.Invalid(path.Child("bindTimeoutSeconds"), args.BindTimeoutSeconds, "invalid BindTimeoutSeconds, should not be a negative value"))
	}

	if opts.AllowVolumeCapacityPriority {
		allErrs = append(allErrs, validateFunctionShape(args.Shape, path.Child("shape"))...)
	} else if args.Shape != nil {
		// When the feature is off, return an error if the config is not nil.
		// This prevents unexpected configuration from taking effect when the
		// feature turns on in the future.
		allErrs = append(allErrs, field.Invalid(path.Child("shape"), args.Shape, "unexpected field `shape`, remove it or turn on the feature gate VolumeCapacityPriority"))
	}
	return allErrs.ToAggregate()
}

func ValidateFoodResourcesFitArgs(path *field.Path, args *v12.FoodResourcesFitArgs) error {
	var allErrs field.ErrorList
	resPath := path.Child("ignoredResources")
	for i, res := range args.IgnoredResources {
		path := resPath.Index(i)
		if errs := metav1validation.ValidateLabelName(res, path); len(errs) != 0 {
			allErrs = append(allErrs, errs...)
		}
	}

	groupPath := path.Child("ignoredResourceGroups")
	for i, group := range args.IgnoredResourceGroups {
		path := groupPath.Index(i)
		if strings.Contains(group, "/") {
			allErrs = append(allErrs, field.Invalid(path, group, "resource group name can't contain '/'"))
		}
		if errs := metav1validation.ValidateLabelName(group, path); len(errs) != 0 {
			allErrs = append(allErrs, errs...)
		}
	}

	strategyPath := path.Child("scoringStrategy")
	if args.ScoringStrategy != nil {
		if !supportedScoringStrategyTypes.Has(string(args.ScoringStrategy.Type)) {
			allErrs = append(allErrs, field.NotSupported(strategyPath.Child("type"), args.ScoringStrategy.Type, supportedScoringStrategyTypes.List()))
		}
		allErrs = append(allErrs, validateResources(args.ScoringStrategy.Resources, strategyPath.Child("resources"))...)
		if args.ScoringStrategy.RequestedToCapacityRatio != nil {
			allErrs = append(allErrs, validateFunctionShape(args.ScoringStrategy.RequestedToCapacityRatio.Shape, strategyPath.Child("shape"))...)
		}
	}

	if len(allErrs) == 0 {
		return nil
	}
	return allErrs.ToAggregate()
}
