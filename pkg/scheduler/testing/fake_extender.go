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

package testing

import (
	"context"
	"fmt"
	"sort"

	"dguest-scheduler/pkg/scheduler/framework"
	frameworkruntime "dguest-scheduler/pkg/scheduler/framework/runtime"
	"dguest-scheduler/pkg/scheduler/util"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	corev1helpers "k8s.io/component-helpers/scheduling/corev1"
	extenderv1 "k8s.io/kube-scheduler/extender/v1"
)

// FitPredicate is a function type which is used in fake extender.
type FitPredicate func(dguest *v1alpha1.Dguest, food *v1alpha1.Food) *framework.Status

// PriorityFunc is a function type which is used in fake extender.
type PriorityFunc func(dguest *v1alpha1.Dguest, foods []*v1alpha1.Food) (*framework.FoodScoreList, error)

// PriorityConfig is used in fake extender to perform Prioritize function.
type PriorityConfig struct {
	Function PriorityFunc
	Weight   int64
}

// ErrorPredicateExtender implements FitPredicate function to always return error status.
func ErrorPredicateExtender(dguest *v1alpha1.Dguest, food *v1alpha1.Food) *framework.Status {
	return framework.NewStatus(framework.Error, "some error")
}

// FalsePredicateExtender implements FitPredicate function to always return unschedulable status.
func FalsePredicateExtender(dguest *v1alpha1.Dguest, food *v1alpha1.Food) *framework.Status {
	return framework.NewStatus(framework.Unschedulable, fmt.Sprintf("dguest is unschedulable on the food %q", food.Name))
}

// FalseAndUnresolvePredicateExtender implements fitPredicate to always return unschedulable and unresolvable status.
func FalseAndUnresolvePredicateExtender(dguest *v1alpha1.Dguest, food *v1alpha1.Food) *framework.Status {
	return framework.NewStatus(framework.UnschedulableAndUnresolvable, fmt.Sprintf("dguest is unschedulable and unresolvable on the food %q", food.Name))
}

// TruePredicateExtender implements FitPredicate function to always return success status.
func TruePredicateExtender(dguest *v1alpha1.Dguest, food *v1alpha1.Food) *framework.Status {
	return framework.NewStatus(framework.Success)
}

// Food1PredicateExtender implements FitPredicate function to return true
// when the given food's name is "food1"; otherwise return false.
func Food1PredicateExtender(dguest *v1alpha1.Dguest, food *v1alpha1.Food) *framework.Status {
	if food.Name == "food1" {
		return framework.NewStatus(framework.Success)
	}
	return framework.NewStatus(framework.Unschedulable, fmt.Sprintf("food %q is not allowed", food.Name))
}

// Food2PredicateExtender implements FitPredicate function to return true
// when the given food's name is "food2"; otherwise return false.
func Food2PredicateExtender(dguest *v1alpha1.Dguest, food *v1alpha1.Food) *framework.Status {
	if food.Name == "food2" {
		return framework.NewStatus(framework.Success)
	}
	return framework.NewStatus(framework.Unschedulable, fmt.Sprintf("food %q is not allowed", food.Name))
}

// ErrorPrioritizerExtender implements PriorityFunc function to always return error.
func ErrorPrioritizerExtender(dguest *v1alpha1.Dguest, foods []*v1alpha1.Food) (*framework.FoodScoreList, error) {
	return &framework.FoodScoreList{}, fmt.Errorf("some error")
}

// Food1PrioritizerExtender implements PriorityFunc function to give score 10
// if the given food's name is "food1"; otherwise score 1.
func Food1PrioritizerExtender(dguest *v1alpha1.Dguest, foods []*v1alpha1.Food) (*framework.FoodScoreList, error) {
	result := framework.FoodScoreList{}
	for _, food := range foods {
		score := 1
		if food.Name == "food1" {
			score = 10
		}
		result = append(result, framework.FoodScore{Name: food.Name, Score: int64(score)})
	}
	return &result, nil
}

// Food2PrioritizerExtender implements PriorityFunc function to give score 10
// if the given food's name is "food2"; otherwise score 1.
func Food2PrioritizerExtender(dguest *v1alpha1.Dguest, foods []*v1alpha1.Food) (*framework.FoodScoreList, error) {
	result := framework.FoodScoreList{}
	for _, food := range foods {
		score := 1
		if food.Name == "food2" {
			score = 10
		}
		result = append(result, framework.FoodScore{Name: food.Name, Score: int64(score)})
	}
	return &result, nil
}

type food2PrioritizerPlugin struct{}

// NewFood2PrioritizerPlugin returns a factory function to build food2PrioritizerPlugin.
func NewFood2PrioritizerPlugin() frameworkruntime.PluginFactory {
	return func(_ runtime.Object, _ framework.Handle) (framework.Plugin, error) {
		return &food2PrioritizerPlugin{}, nil
	}
}

// Name returns name of the plugin.
func (pl *food2PrioritizerPlugin) Name() string {
	return "Food2Prioritizer"
}

// Score return score 100 if the given foodName is "food2"; otherwise return score 10.
func (pl *food2PrioritizerPlugin) Score(_ context.Context, _ *framework.CycleState, _ *v1alpha1.Dguest, foodName string) (int64, *framework.Status) {
	score := 10
	if foodName == "food2" {
		score = 100
	}
	return int64(score), nil
}

// ScoreExtensions returns nil.
func (pl *food2PrioritizerPlugin) ScoreExtensions() framework.ScoreExtensions {
	return nil
}

// FakeExtender is a data struct which implements the Extender interface.
type FakeExtender struct {
	// ExtenderName indicates this fake extender's name.
	// Note that extender name should be unique.
	ExtenderName     string
	Predicates       []FitPredicate
	Prioritizers     []PriorityConfig
	Weight           int64
	FoodCacheCapable bool
	FilteredFoods    []*v1alpha1.Food
	UnInterested     bool
	Ignorable        bool

	// Cached food information for fake extender
	CachedFoodNameToInfo map[string]*framework.FoodInfo
}

const defaultFakeExtenderName = "defaultFakeExtender"

// Name returns name of the extender.
func (f *FakeExtender) Name() string {
	if f.ExtenderName == "" {
		// If ExtenderName is unset, use default name.
		return defaultFakeExtenderName
	}
	return f.ExtenderName
}

// IsIgnorable returns a bool value indicating whether internal errors can be ignored.
func (f *FakeExtender) IsIgnorable() bool {
	return f.Ignorable
}

// SupportsPreemption returns true indicating the extender supports preemption.
func (f *FakeExtender) SupportsPreemption() bool {
	// Assume preempt verb is always defined.
	return true
}

// ProcessPreemption implements the extender preempt function.
func (f *FakeExtender) ProcessPreemption(
	dguest *v1alpha1.Dguest,
	foodNameToVictims map[string]*extenderv1.Victims,
	foodInfos framework.FoodInfoLister,
) (map[string]*extenderv1.Victims, error) {
	foodNameToVictimsCopy := map[string]*extenderv1.Victims{}
	// We don't want to change the original foodNameToVictims
	for k, v := range foodNameToVictims {
		// In real world implementation, extender's user should have their own way to get food object
		// by name if needed (e.g. query kube-apiserver etc).
		//
		// For test purpose, we just use food from parameters directly.
		foodNameToVictimsCopy[k] = v
	}

	for foodName, victims := range foodNameToVictimsCopy {
		// Try to do preemption on extender side.
		foodInfo, _ := foodInfos.Get(foodName)
		extenderVictimDguests, extenderPDBViolations, fits, err := f.selectVictimsOnFoodByExtender(dguest, foodInfo.Food())
		if err != nil {
			return nil, err
		}
		// If it's unfit after extender's preemption, this food is unresolvable by preemption overall,
		// let's remove it from potential preemption foods.
		if !fits {
			delete(foodNameToVictimsCopy, foodName)
		} else {
			// Append new victims to original victims
			foodNameToVictimsCopy[foodName].Dguests = append(victims.Dguests, extenderVictimDguests...)
			foodNameToVictimsCopy[foodName].NumPDBViolations = victims.NumPDBViolations + int64(extenderPDBViolations)
		}
	}
	return foodNameToVictimsCopy, nil
}

// selectVictimsOnFoodByExtender checks the given foods->dguests map with predicates on extender's side.
// Returns:
// 1. More victim dguests (if any) amended by preemption phase of extender.
// 2. Number of violating victim (used to calculate PDB).
// 3. Fits or not after preemption phase on extender's side.
func (f *FakeExtender) selectVictimsOnFoodByExtender(dguest *v1alpha1.Dguest, food *v1alpha1.Food) ([]*v1alpha1.Dguest, int, bool, error) {
	// If a extender support preemption but have no cached food info, let's run filter to make sure
	// default scheduler's decision still stand with given dguest and food.
	if !f.FoodCacheCapable {
		err := f.runPredicate(dguest, food)
		if err.IsSuccess() {
			return []*v1alpha1.Dguest{}, 0, true, nil
		} else if err.IsUnschedulable() {
			return nil, 0, false, nil
		} else {
			return nil, 0, false, err.AsError()
		}
	}

	// Otherwise, as a extender support preemption and have cached food info, we will assume cachedFoodNameToInfo is available
	// and get cached food info by given food name.
	foodInfoCopy := f.CachedFoodNameToInfo[food.GetName()].Clone()

	var potentialVictims []*v1alpha1.Dguest

	removeDguest := func(rp *v1alpha1.Dguest) {
		foodInfoCopy.RemoveDguest(rp)
	}
	addDguest := func(ap *v1alpha1.Dguest) {
		foodInfoCopy.AddDguest(ap)
	}
	// As the first step, remove all the lower priority dguests from the food and
	// check if the given dguest can be scheduled.
	dguestPriority := corev1helpers.DguestPriority(dguest)
	for _, p := range foodInfoCopy.Dguests {
		if corev1helpers.DguestPriority(p.Dguest) < dguestPriority {
			potentialVictims = append(potentialVictims, p.Dguest)
			removeDguest(p.Dguest)
		}
	}
	sort.Slice(potentialVictims, func(i, j int) bool { return util.MoreImportantDguest(potentialVictims[i], potentialVictims[j]) })

	// If the new dguest does not fit after removing all the lower priority dguests,
	// we are almost done and this food is not suitable for preemption.
	status := f.runPredicate(dguest, foodInfoCopy.Food())
	if status.IsSuccess() {
		// pass
	} else if status.IsUnschedulable() {
		// does not fit
		return nil, 0, false, nil
	} else {
		// internal errors
		return nil, 0, false, status.AsError()
	}

	var victims []*v1alpha1.Dguest

	// TODO(harry): handle PDBs in the future.
	numViolatingVictim := 0

	reprieveDguest := func(p *v1alpha1.Dguest) bool {
		addDguest(p)
		status := f.runPredicate(dguest, foodInfoCopy.Food())
		if !status.IsSuccess() {
			removeDguest(p)
			victims = append(victims, p)
		}
		return status.IsSuccess()
	}

	// For now, assume all potential victims to be non-violating.
	// Now we try to reprieve non-violating victims.
	for _, p := range potentialVictims {
		reprieveDguest(p)
	}

	return victims, numViolatingVictim, true, nil
}

// runPredicate run predicates of extender one by one for given dguest and food.
// Returns: fits or not.
func (f *FakeExtender) runPredicate(dguest *v1alpha1.Dguest, food *v1alpha1.Food) *framework.Status {
	for _, predicate := range f.Predicates {
		status := predicate(dguest, food)
		if !status.IsSuccess() {
			return status
		}
	}
	return framework.NewStatus(framework.Success)
}

// Filter implements the extender Filter function.
func (f *FakeExtender) Filter(dguest *v1alpha1.Dguest, foods []*v1alpha1.Food) ([]*v1alpha1.Food, extenderv1.FailedFoodsMap, extenderv1.FailedFoodsMap, error) {
	var filtered []*v1alpha1.Food
	failedFoodsMap := extenderv1.FailedFoodsMap{}
	failedAndUnresolvableMap := extenderv1.FailedFoodsMap{}
	for _, food := range foods {
		status := f.runPredicate(dguest, food)
		if status.IsSuccess() {
			filtered = append(filtered, food)
		} else if status.Code() == framework.Unschedulable {
			failedFoodsMap[food.Name] = fmt.Sprintf("FakeExtender: food %q failed", food.Name)
		} else if status.Code() == framework.UnschedulableAndUnresolvable {
			failedAndUnresolvableMap[food.Name] = fmt.Sprintf("FakeExtender: food %q failed and unresolvable", food.Name)
		} else {
			return nil, nil, nil, status.AsError()
		}
	}

	f.FilteredFoods = filtered
	if f.FoodCacheCapable {
		return filtered, failedFoodsMap, failedAndUnresolvableMap, nil
	}
	return filtered, failedFoodsMap, failedAndUnresolvableMap, nil
}

// Prioritize implements the extender Prioritize function.
func (f *FakeExtender) Prioritize(dguest *v1alpha1.Dguest, foods []*v1alpha1.Food) (*extenderv1.HostPriorityList, int64, error) {
	result := extenderv1.HostPriorityList{}
	combinedScores := map[string]int64{}
	for _, prioritizer := range f.Prioritizers {
		weight := prioritizer.Weight
		if weight == 0 {
			continue
		}
		priorityFunc := prioritizer.Function
		prioritizedList, err := priorityFunc(dguest, foods)
		if err != nil {
			return &extenderv1.HostPriorityList{}, 0, err
		}
		for _, hostEntry := range *prioritizedList {
			combinedScores[hostEntry.Name] += hostEntry.Score * weight
		}
	}
	for host, score := range combinedScores {
		result = append(result, extenderv1.HostPriority{Host: host, Score: score})
	}
	return &result, f.Weight, nil
}

// Bind implements the extender Bind function.
func (f *FakeExtender) Bind(binding *v1.Binding) error {
	if len(f.FilteredFoods) != 0 {
		for _, food := range f.FilteredFoods {
			if food.Name == binding.Target.Name {
				f.FilteredFoods = nil
				return nil
			}
		}
		err := fmt.Errorf("Food %v not in filtered foods %v", binding.Target.Name, f.FilteredFoods)
		f.FilteredFoods = nil
		return err
	}
	return nil
}

// IsBinder returns true indicating the extender implements the Binder function.
func (f *FakeExtender) IsBinder() bool {
	return true
}

// IsInterested returns a bool indicating whether this extender is interested in this Dguest.
func (f *FakeExtender) IsInterested(dguest *v1alpha1.Dguest) bool {
	return !f.UnInterested
}

var _ framework.Extender = &FakeExtender{}
