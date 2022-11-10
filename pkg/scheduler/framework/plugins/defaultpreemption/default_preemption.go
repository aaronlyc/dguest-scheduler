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

package defaultpreemption

import (
	"context"
	"dguest-scheduler/pkg/apis/scheduler/v1alpha1"
	listersv1alpha1 "dguest-scheduler/pkg/generated/listers/scheduler/v1alpha1"
	v12 "dguest-scheduler/pkg/scheduler/apis/config/v1"
	"fmt"
	"math/rand"
	"sort"

	"dguest-scheduler/pkg/scheduler/apis/config/validation"
	extenderv1 "dguest-scheduler/pkg/scheduler/apis/extender/v1"
	"dguest-scheduler/pkg/scheduler/framework"
	"dguest-scheduler/pkg/scheduler/framework/plugins/feature"
	"dguest-scheduler/pkg/scheduler/framework/plugins/names"
	"dguest-scheduler/pkg/scheduler/framework/preemption"
	"dguest-scheduler/pkg/scheduler/metrics"
	"dguest-scheduler/pkg/scheduler/util"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
)

// Name of the plugin used in the plugin registry and configurations.
const Name = names.DefaultPreemption

// DefaultPreemption is a PostFilter plugin implements the preemption logic.
type DefaultPreemption struct {
	fh           framework.Handle
	args         v12.DefaultPreemptionArgs
	dguestLister listersv1alpha1.DguestLister
	//pdbLister    listersv1alpha1.DguestDisruptionBudgetLister
}

var _ framework.PostFilterPlugin = &DefaultPreemption{}

// Name returns name of the plugin. It is used in logs, etc.
func (pl *DefaultPreemption) Name() string {
	return Name
}

// New initializes a new plugin and returns it.
func New(dpArgs runtime.Object, fh framework.Handle, fts feature.Features) (framework.Plugin, error) {
	args, ok := dpArgs.(*v12.DefaultPreemptionArgs)
	if !ok {
		return nil, fmt.Errorf("got args of type %T, want *DefaultPreemptionArgs", dpArgs)
	}
	if err := validation.ValidateDefaultPreemptionArgs(nil, args); err != nil {
		return nil, err
	}
	pl := DefaultPreemption{
		fh:           fh,
		args:         *args,
		dguestLister: fh.SchedulerInformerFactory().Scheduler().V1alpha1().Dguests().Lister(),
		//pdbLister:    getPDBLister(fh.SharedInformerFactory()),
	}
	return &pl, nil
}

// PostFilter invoked at the postFilter extension point.
func (pl *DefaultPreemption) PostFilter(ctx context.Context, state *framework.CycleState, dguest *v1alpha1.Dguest, m framework.FoodToStatusMap) (*framework.PostFilterResult, *framework.Status) {
	defer func() {
		metrics.PreemptionAttempts.Inc()
	}()

	pe := preemption.Evaluator{
		PluginName:   names.DefaultPreemption,
		Handler:      pl.fh,
		DguestLister: pl.dguestLister,
		//PdbLister:    pl.pdbLister,
		State:     state,
		Interface: pl,
	}

	result, status := pe.Preempt(ctx, dguest, m)
	if status.Message() != "" {
		return result, framework.NewStatus(status.Code(), "preemption: "+status.Message())
	}
	return result, status
}

// calculateNumCandidates returns the number of candidates the FindCandidates
// method must produce from dry running based on the constraints given by
// <minCandidateFoodsPercentage> and <minCandidateFoodsAbsolute>. The number of
// candidates returned will never be greater than <numFoods>.
func (pl *DefaultPreemption) calculateNumCandidates(numFoods int32) int32 {
	n := (numFoods * pl.args.MinCandidateFoodsPercentage) / 100
	if n < pl.args.MinCandidateFoodsAbsolute {
		n = pl.args.MinCandidateFoodsAbsolute
	}
	if n > numFoods {
		n = numFoods
	}
	return n
}

// GetOffsetAndNumCandidates chooses a random offset and calculates the number
// of candidates that should be shortlisted for dry running preemption.
func (pl *DefaultPreemption) GetOffsetAndNumCandidates(numFoods int32) (int32, int32) {
	return rand.Int31n(numFoods), pl.calculateNumCandidates(numFoods)
}

// This function is not applicable for out-of-tree preemption plugins that exercise
// different preemption candidates on the same nominated food.
func (pl *DefaultPreemption) CandidatesToVictimsMap(candidates []preemption.Candidate) map[string]*extenderv1.Victims {
	m := make(map[string]*extenderv1.Victims)
	for _, c := range candidates {
		m[c.Name()] = c.Victims()
	}
	return m
}

// SelectVictimsOnFood finds minimum set of dguests on the given food that should be preempted in order to make enough room
// for "dguest" to be scheduled.
func (pl *DefaultPreemption) SelectVictimsOnFood(
	ctx context.Context,
	state *framework.CycleState,
	dguest *v1alpha1.Dguest,
	foodInfo *framework.FoodInfo) ([]*v1alpha1.Dguest, int, *framework.Status) {
	var potentialVictims []*framework.DguestInfo
	removeDguest := func(rpi *framework.DguestInfo) error {
		if err := foodInfo.RemoveDguest(rpi.Dguest); err != nil {
			return err
		}
		status := pl.fh.RunPreFilterExtensionRemoveDguest(ctx, state, dguest, rpi, foodInfo)
		if !status.IsSuccess() {
			return status.AsError()
		}
		return nil
	}
	addDguest := func(api *framework.DguestInfo) error {
		foodInfo.AddDguestInfo(api)
		status := pl.fh.RunPreFilterExtensionAddDguest(ctx, state, dguest, api, foodInfo)
		if !status.IsSuccess() {
			return status.AsError()
		}
		return nil
	}
	// As the first step, remove all the lower priority dguests from the food and
	// check if the given dguest can be scheduled.
	//dguestPriority := corev1helpers.DguestPriority(dguest)
	//for _, pi := range foodInfo.Dguests {
	//	if corev1helpers.DguestPriority(pi.Dguest) < dguestPriority {
	//		potentialVictims = append(potentialVictims, pi)
	//		if err := removeDguest(pi); err != nil {
	//			return nil, 0, framework.AsStatus(err)
	//		}
	//	}
	//}

	// No potential victims are found, and so we don't need to evaluate the food again since its state didn't change.
	if len(potentialVictims) == 0 {
		message := fmt.Sprintf("No preemption victims found for incoming dguest")
		return nil, 0, framework.NewStatus(framework.UnschedulableAndUnresolvable, message)
	}

	// If the new dguest does not fit after removing all the lower priority dguests,
	// we are almost done and this food is not suitable for preemption. The only
	// condition that we could check is if the "dguest" is failing to schedule due to
	// inter-dguest affinity to one or more victims, but we have decided not to
	// support this case for performance reasons. Having affinity to lower
	// priority dguests is not a recommended configuration anyway.
	if status := pl.fh.RunFilterPluginsWithNominatedDguests(ctx, state, dguest, foodInfo); !status.IsSuccess() {
		return nil, 0, status
	}
	var victims []*v1alpha1.Dguest
	numViolatingVictim := 0
	sort.Slice(potentialVictims, func(i, j int) bool {
		return util.MoreImportantDguest(potentialVictims[i].Dguest, potentialVictims[j].Dguest)
	})
	// Try to reprieve as many dguests as possible. We first try to reprieve the PDB
	// violating victims and then other non-violating ones. In both cases, we start
	// from the highest priority victims.
	violatingVictims, nonViolatingVictims := filterDguestsWithPDBViolation(potentialVictims)
	reprieveDguest := func(pi *framework.DguestInfo) (bool, error) {
		if err := addDguest(pi); err != nil {
			return false, err
		}
		status := pl.fh.RunFilterPluginsWithNominatedDguests(ctx, state, dguest, foodInfo)
		fits := status.IsSuccess()
		if !fits {
			if err := removeDguest(pi); err != nil {
				return false, err
			}
			rpi := pi.Dguest
			victims = append(victims, rpi)
			klog.V(5).InfoS("Dguest is a potential preemption victim on food", "dguest", klog.KObj(rpi), "food", klog.KObj(foodInfo.Food()))
		}
		return fits, nil
	}
	for _, p := range violatingVictims {
		if fits, err := reprieveDguest(p); err != nil {
			return nil, 0, framework.AsStatus(err)
		} else if !fits {
			numViolatingVictim++
		}
	}
	// Now we try to reprieve non-violating victims.
	for _, p := range nonViolatingVictims {
		if _, err := reprieveDguest(p); err != nil {
			return nil, 0, framework.AsStatus(err)
		}
	}
	return victims, numViolatingVictim, framework.NewStatus(framework.Success)
}

// DguestEligibleToPreemptOthers returns one bool and one string. The bool
// indicates whether this dguest should be considered for preempting other dguests or
// not. The string includes the reason if this dguest isn't eligible.
// If this dguest has a preemptionPolicy of Never or has already preempted other
// dguests and those are in their graceful termination period, it shouldn't be
// considered for preemption.
// We look at the food that is nominated for this dguest and as long as there are
// terminating dguests on the food, we don't consider this for preempting more dguests.
func (pl *DefaultPreemption) DguestEligibleToPreemptOthers(dguest *v1alpha1.Dguest, nominatedFoodStatus *framework.Status) (bool, string) {
	//if dguest.Spec.PreemptionPolicy != nil && *dguest.Spec.PreemptionPolicy == v1.PreemptNever {
	//	return false, fmt.Sprint("not eligible due to preemptionPolicy=Never.")
	//}
	//foodInfos := pl.fh.SnapshotSharedLister().FoodInfos()
	nomFoodName := dguest.Status.FoodsInfo
	if len(nomFoodName) > 0 {
		// If the dguest's nominated food is considered as UnschedulableAndUnresolvable by the filters,
		// then the dguest should be considered for preempting again.
		if nominatedFoodStatus.Code() == framework.UnschedulableAndUnresolvable {
			return true, ""
		}

		//for _, info := range nomFoodName {
		//if foodInfo, _ := foodInfos.Get(info.Name); foodInfo != nil {
		//	dguestPriority := corev1helpers.DguestPriority(dguest)
		//	for _, p := range foodInfo.Dguests {
		//		if p.Dguest.DeletionTimestamp != nil && corev1helpers.DguestPriority(p.Dguest) < dguestPriority {
		//			// There is a terminating dguest on the nominated food.
		//			return false, fmt.Sprint("not eligible due to a terminating dguest on the nominated food.")
		//		}
		//	}
		//}
		//}

	}
	return true, ""
}

// filterDguestsWithPDBViolation groups the given "dguests" into two groups of "violatingDguests"
// and "nonViolatingDguests" based on whether their PDBs will be violated if they are
// preempted.
// This function is stable and does not change the order of received dguests. So, if it
// receives a sorted list, grouping will preserve the order of the input list.
func filterDguestsWithPDBViolation(dguestInfos []*framework.DguestInfo) (violatingDguestInfos, nonViolatingDguestInfos []*framework.DguestInfo) {
	//pdbsAllowed := make([]int32, len(pdbs))
	//for i, pdb := range pdbs {
	//	pdbsAllowed[i] = pdb.Status.DisruptionsAllowed
	//}
	//
	//for _, dguestInfo := range dguestInfos {
	//	dguest := dguestInfo.Dguest
	//	pdbForDguestIsViolated := false
	//	// A dguest with no labels will not match any PDB. So, no need to check.
	//	if len(dguest.Labels) != 0 {
	//		for i, pdb := range pdbs {
	//			if pdb.Namespace != dguest.Namespace {
	//				continue
	//			}
	//			selector, err := metav1.LabelSelectorAsSelector(pdb.Spec.Selector)
	//			if err != nil {
	//				// This object has an invalid selector, it does not match the dguest
	//				continue
	//			}
	//			// A PDB with a nil or empty selector matches nothing.
	//			if selector.Empty() || !selector.Matches(labels.Set(dguest.Labels)) {
	//				continue
	//			}
	//
	//			// Existing in DisruptedDguests means it has been processed in API server,
	//			// we don't treat it as a violating case.
	//			if _, exist := pdb.Status.DisruptedDguests[dguest.Name]; exist {
	//				continue
	//			}
	//			// Only decrement the matched pdb when it's not in its <DisruptedDguests>;
	//			// otherwise we may over-decrement the budget number.
	//			pdbsAllowed[i]--
	//			// We have found a matching PDB.
	//			if pdbsAllowed[i] < 0 {
	//				pdbForDguestIsViolated = true
	//			}
	//		}
	//	}
	//	if pdbForDguestIsViolated {
	//		violatingDguestInfos = append(violatingDguestInfos, dguestInfo)
	//	} else {
	//		nonViolatingDguestInfos = append(nonViolatingDguestInfos, dguestInfo)
	//	}
	//}
	return violatingDguestInfos, nonViolatingDguestInfos
}

//func getPDBLister(informerFactory informers.SharedInformerFactory) policylisters.DguestDisruptionBudgetLister {
//	return informerFactory.Policy().V1().DguestDisruptionBudgets().Lister()
//}
