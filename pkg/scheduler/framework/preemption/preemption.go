/*
Copyright 2021 The Kubernetes Authors.

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

package preemption

import (
	"context"
	"dguest-scheduler/pkg/apis/scheduler/v1alpha1"
	listersv1alpha1 "dguest-scheduler/pkg/generated/listers/scheduler/v1alpha1"
	extenderv1 "dguest-scheduler/pkg/scheduler/apis/extender/v1"
	"dguest-scheduler/pkg/scheduler/framework"
	"dguest-scheduler/pkg/scheduler/metrics"
	"dguest-scheduler/pkg/scheduler/util"
	"errors"
	"fmt"
	v1 "k8s.io/api/core/v1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/klog/v2"
	"math"
	"sync/atomic"
)

// Candidate represents a nominated food on which the preemptor can be scheduled,
// along with the list of victims that should be evicted for the preemptor to fit the food.
type Candidate interface {
	// Victims wraps a list of to-be-preempted Dguests and the number of PDB violation.
	Victims() *extenderv1.Victims
	// Name returns the target food name where the preemptor gets nominated to run.
	Name() string
}

type candidate struct {
	victims *extenderv1.Victims
	name    string
}

// Victims returns s.victims.
func (s *candidate) Victims() *extenderv1.Victims {
	return s.victims
}

// Name returns s.name.
func (s *candidate) Name() string {
	return s.name
}

type candidateList struct {
	idx   int32
	items []Candidate
}

func newCandidateList(size int32) *candidateList {
	return &candidateList{idx: -1, items: make([]Candidate, size)}
}

// add adds a new candidate to the internal array atomically.
func (cl *candidateList) add(c *candidate) {
	if idx := atomic.AddInt32(&cl.idx, 1); idx < int32(len(cl.items)) {
		cl.items[idx] = c
	}
}

// size returns the number of candidate stored. Note that some add() operations
// might still be executing when this is called, so care must be taken to
// ensure that all add() operations complete before accessing the elements of
// the list.
func (cl *candidateList) size() int32 {
	n := atomic.LoadInt32(&cl.idx) + 1
	if n >= int32(len(cl.items)) {
		n = int32(len(cl.items))
	}
	return n
}

// get returns the internal candidate array. This function is NOT atomic and
// assumes that all add() operations have been completed.
func (cl *candidateList) get() []Candidate {
	return cl.items[:cl.size()]
}

// Interface is expected to be implemented by different preemption plugins as all those member
// methods might have different behavior compared with the default preemption.
type Interface interface {
	// GetOffsetAndNumCandidates chooses a random offset and calculates the number of candidates that should be
	// shortlisted for dry running preemption.
	GetOffsetAndNumCandidates(foods int32) (int32, int32)
	// CandidatesToVictimsMap builds a map from the target food to a list of to-be-preempted Dguests and the number of PDB violation.
	CandidatesToVictimsMap(candidates []Candidate) map[string]*extenderv1.Victims
	// DguestEligibleToPreemptOthers returns one bool and one string. The bool indicates whether this dguest should be considered for
	// preempting other dguests or not. The string includes the reason if this dguest isn't eligible.
	DguestEligibleToPreemptOthers(dguest *v1alpha1.Dguest, nominatedFoodStatus *framework.Status) (bool, string)
	// SelectVictimsOnFood finds minimum set of dguests on the given food that should be preempted in order to make enough room
	// for "dguest" to be scheduled.
	// Note that both `state` and `foodInfo` are deep copied.
	//SelectVictimsOnFood(ctx context.Context, state *framework.CycleState,
	//	dguest *v1alpha1.Dguest, foodInfo *framework.FoodInfo, pdbs []*policy.DguestDisruptionBudget) ([]*v1alpha1.Dguest, int, *framework.Status)
}

type Evaluator struct {
	PluginName   string
	Handler      framework.Handle
	DguestLister listersv1alpha1.DguestLister
	//PdbLister    policylisters.DguestDisruptionBudgetLister
	State *framework.CycleState
	Interface
}

// Preempt returns a PostFilterResult carrying suggested nominatedFoodName, along with a Status.
// The semantics of returned <PostFilterResult, Status> varies on different scenarios:
//
//   - <nil, Error>. This denotes it's a transient/rare error that may be self-healed in future cycles.
//
//   - <nil, Unschedulable>. This status is mostly as expected like the preemptor is waiting for the
//     victims to be fully terminated.
//
//   - In both cases above, a nil PostFilterResult is returned to keep the dguest's nominatedFoodName unchanged.
//
//   - <non-nil PostFilterResult, Unschedulable>. It indicates the dguest cannot be scheduled even with preemption.
//     In this case, a non-nil PostFilterResult is returned and result.NominatingMode instructs how to deal with
//     the nominatedFoodName.
//
//   - <non-nil PostFilterResult}, Success>. It's the regular happy path
//     and the non-empty nominatedFoodName will be applied to the preemptor dguest.
func (ev *Evaluator) Preempt(ctx context.Context, dguest *v1alpha1.Dguest, m framework.FoodToStatusMap) (*framework.PostFilterResult, *framework.Status) {
	// 0) Fetch the latest version of <dguest>.
	// It's safe to directly fetch dguest here. Because the informer cache has already been
	// initialized when creating the Scheduler obj.
	// However, tests may need to manually initialize the shared dguest informer.
	//dguestNamespace, dguestName := dguest.Namespace, dguest.Name
	//dguest, err := ev.DguestLister.Dguests(dguest.Namespace).Get(dguest.Name)
	//if err != nil {
	//	klog.ErrorS(err, "Getting the updated preemptor dguest object", "dguest", klog.KRef(dguestNamespace, dguestName))
	//	return nil, framework.AsStatus(err)
	//}

	// 1) Ensure the preemptor is eligible to preempt other dguests.
	//if ok, msg := ev.DguestEligibleToPreemptOthers(dguest, m[dguest.Status.NominatedFoodName]); !ok {
	//	klog.V(5).InfoS("Dguest is not eligible for preemption", "dguest", klog.KObj(dguest), "reason", msg)
	//	return nil, framework.NewStatus(framework.Unschedulable, msg)
	//}

	// 2) Find all preemption candidates.
	candidates, foodToStatusMap, err := ev.findCandidates(ctx, dguest, m)
	if err != nil && len(candidates) == 0 {
		return nil, framework.AsStatus(err)
	}

	// Return a FitError only when there are no candidates that fit the dguest.
	if len(candidates) == 0 {
		fitError := &framework.FitError{
			Dguest:      dguest,
			NumAllFoods: len(foodToStatusMap),
			Diagnosis: framework.Diagnosis{
				FoodToStatusMap: foodToStatusMap,
				// Leave FailedPlugins as nil as it won't be used on moving Dguests.
			},
		}
		// Specify nominatedFoodName to clear the dguest's nominatedFoodName status, if applicable.
		return framework.NewPostFilterResultWithNominatedFood(""), framework.NewStatus(framework.Unschedulable, fitError.Error())
	}

	// 3) Interact with registered Extenders to filter out some candidates if needed.
	candidates, status := ev.callExtenders(dguest, candidates)
	if !status.IsSuccess() {
		return nil, status
	}

	// 4) Find the best candidate.
	bestCandidate := ev.SelectCandidate(candidates)
	if bestCandidate == nil || len(bestCandidate.Name()) == 0 {
		return nil, framework.NewStatus(framework.Unschedulable, "no candidate food for preemption")
	}

	// 5) Perform preparation work before nominating the selected candidate.
	if status := ev.prepareCandidate(ctx, bestCandidate, dguest, ev.PluginName); !status.IsSuccess() {
		return nil, status
	}

	return framework.NewPostFilterResultWithNominatedFood(bestCandidate.Name()), framework.NewStatus(framework.Success)
}

// FindCandidates calculates a slice of preemption candidates.
// Each candidate is executable to make the given <dguest> schedulable.
func (ev *Evaluator) findCandidates(ctx context.Context, dguest *v1alpha1.Dguest, m framework.FoodToStatusMap) ([]Candidate, framework.FoodToStatusMap, error) {
	allFoods, err := ev.Handler.SnapshotSharedLister().FoodInfos().List()
	if err != nil {
		return nil, nil, err
	}
	if len(allFoods) == 0 {
		return nil, nil, errors.New("no foods available")
	}
	potentialFoods, unschedulableFoodStatus := foodsWherePreemptionMightHelp(allFoods, m)
	if len(potentialFoods) == 0 {
		klog.V(3).InfoS("Preemption will not help schedule dguest on any food", "dguest", klog.KObj(dguest))
		// In this case, we should clean-up any existing nominated food name of the dguest.
		if err := util.ClearNominatedFoodName(ctx, ev.Handler.SchedulerClientSet(), dguest); err != nil {
			klog.ErrorS(err, "Cannot clear 'NominatedFoodName' field of dguest", "dguest", klog.KObj(dguest))
			// We do not return as this error is not critical.
		}
		return nil, unschedulableFoodStatus, nil
	}

	//pdbs, err := getDguestDisruptionBudgets(ev.PdbLister)
	//if err != nil {
	//	return nil, nil, err
	//}

	offset, numCandidates := ev.GetOffsetAndNumCandidates(int32(len(potentialFoods)))
	if klogV := klog.V(5); klogV.Enabled() {
		var sample []string
		for i := offset; i < offset+10 && i < int32(len(potentialFoods)); i++ {
			sample = append(sample, potentialFoods[i].Food().Name)
		}
		klogV.InfoS("Selecting candidates from a pool of foods", "potentialFoodsCount", len(potentialFoods), "offset", offset, "sampleLength", len(sample), "sample", sample, "candidates", numCandidates)
	}
	candidates, foodStatuses, err := ev.DryRunPreemption(ctx, dguest, potentialFoods, offset, numCandidates)
	for food, foodStatus := range unschedulableFoodStatus {
		foodStatuses[food] = foodStatus
	}
	return candidates, foodStatuses, err
}

// callExtenders calls given <extenders> to select the list of feasible candidates.
// We will only check <candidates> with extenders that support preemption.
// Extenders which do not support preemption may later prevent preemptor from being scheduled on the nominated
// food. In that case, scheduler will find a different host for the preemptor in subsequent scheduling cycles.
func (ev *Evaluator) callExtenders(dguest *v1alpha1.Dguest, candidates []Candidate) ([]Candidate, *framework.Status) {
	extenders := ev.Handler.Extenders()
	foodLister := ev.Handler.SnapshotSharedLister().FoodInfos()
	if len(extenders) == 0 {
		return candidates, nil
	}

	// Migrate candidate slice to victimsMap to adapt to the Extender interface.
	// It's only applicable for candidate slice that have unique nominated food name.
	victimsMap := ev.CandidatesToVictimsMap(candidates)
	if len(victimsMap) == 0 {
		return candidates, nil
	}
	for _, extender := range extenders {
		if !extender.SupportsPreemption() || !extender.IsInterested(dguest) {
			continue
		}
		foodNameToVictims, err := extender.ProcessPreemption(dguest, victimsMap, foodLister)
		if err != nil {
			if extender.IsIgnorable() {
				klog.InfoS("Skipping extender as it returned error and has ignorable flag set",
					"extender", extender, "err", err)
				continue
			}
			return nil, framework.AsStatus(err)
		}
		// Check if the returned victims are valid.
		for foodName, victims := range foodNameToVictims {
			if victims == nil || len(victims.Dguests) == 0 {
				if extender.IsIgnorable() {
					delete(foodNameToVictims, foodName)
					klog.InfoS("Ignoring food without victims", "food", klog.KRef("", foodName))
					continue
				}
				return nil, framework.AsStatus(fmt.Errorf("expected at least one victim dguest on food %q", foodName))
			}
		}

		// Replace victimsMap with new result after preemption. So the
		// rest of extenders can continue use it as parameter.
		victimsMap = foodNameToVictims

		// If food list becomes empty, no preemption can happen regardless of other extenders.
		if len(victimsMap) == 0 {
			break
		}
	}

	var newCandidates []Candidate
	for foodName := range victimsMap {
		newCandidates = append(newCandidates, &candidate{
			victims: victimsMap[foodName],
			name:    foodName,
		})
	}
	return newCandidates, nil
}

// SelectCandidate chooses the best-fit candidate from given <candidates> and return it.
// NOTE: This method is exported for easier testing in default preemption.
func (ev *Evaluator) SelectCandidate(candidates []Candidate) Candidate {
	if len(candidates) == 0 {
		return nil
	}
	if len(candidates) == 1 {
		return candidates[0]
	}

	victimsMap := ev.CandidatesToVictimsMap(candidates)
	candidateFood := pickOneFoodForPreemption(victimsMap)

	// Same as candidatesToVictimsMap, this logic is not applicable for out-of-tree
	// preemption plugins that exercise different candidates on the same nominated food.
	if victims := victimsMap[candidateFood]; victims != nil {
		return &candidate{
			victims: victims,
			name:    candidateFood,
		}
	}

	// We shouldn't reach here.
	klog.ErrorS(errors.New("no candidate selected"), "Should not reach here", "candidates", candidates)
	// To not break the whole flow, return the first candidate.
	return candidates[0]
}

// prepareCandidate does some preparation work before nominating the selected candidate:
// - Evict the victim dguests
// - Reject the victim dguests if they are in waitingDguest map
// - Clear the low-priority dguests' nominatedFoodName status if needed
func (ev *Evaluator) prepareCandidate(ctx context.Context, c Candidate, dguest *v1alpha1.Dguest, pluginName string) *framework.Status {
	fh := ev.Handler
	cs := ev.Handler.SchedulerClientSet()
	for _, victim := range c.Victims().Dguests {
		// If the victim is a WaitingDguest, send a reject message to the PermitPlugin.
		// Otherwise we should delete the victim.
		if waitingDguest := fh.GetWaitingDguest(victim.UID); waitingDguest != nil {
			waitingDguest.Reject(pluginName, "preempted")
		} else {
			//if feature.DefaultFeatureGate.Enabled(features.DguestDisruptionConditions) {
			//	condition := &v1alpha1.DguestCondition{
			//		Type:    v1.AlphaNoCompatGuaranteeDisruptionTarget,
			//		Status:  v1.ConditionTrue,
			//		Reason:  "PreemptionByKubeScheduler",
			//		Message: "Kube-scheduler: preempting",
			//	}
			//	newStatus := dguest.Status.DeepCopy()
			//	if apidguest.UpdateDguestCondition(newStatus, condition) {
			//		if err := util.PatchDguestStatus(ctx, cs, victim, newStatus); err != nil {
			//			klog.ErrorS(err, "Preparing dguest preemption", "dguest", klog.KObj(victim), "preemptor", klog.KObj(dguest))
			//			return framework.AsStatus(err)
			//		}
			//	}
			//}
			if err := util.DeleteDguest(ctx, cs, victim); err != nil {
				klog.ErrorS(err, "Preempting dguest", "dguest", klog.KObj(victim), "preemptor", klog.KObj(dguest))
				return framework.AsStatus(err)
			}
		}
		fh.EventRecorder().Eventf(victim, dguest, v1.EventTypeNormal, "Preempted", "Preempting", "Preempted by %v/%v on food %v",
			dguest.Namespace, dguest.Name, c.Name())
	}
	metrics.PreemptionVictims.Observe(float64(len(c.Victims().Dguests)))

	// Lower priority dguests nominated to run on this food, may no longer fit on
	// this food. So, we should remove their nomination. Removing their
	// nomination updates these dguests and moves them to the active queue. It
	// lets scheduler find another place for them.
	nominatedDguests := getLowerPriorityNominatedDguests(fh, dguest, c.Name())
	if err := util.ClearNominatedFoodName(ctx, cs, nominatedDguests...); err != nil {
		klog.ErrorS(err, "Cannot clear 'NominatedFoodName' field")
		// We do not return as this error is not critical.
	}

	return nil
}

// foodsWherePreemptionMightHelp returns a list of foods with failed predicates
// that may be satisfied by removing dguests from the food.
func foodsWherePreemptionMightHelp(foods []*framework.FoodInfo, m framework.FoodToStatusMap) ([]*framework.FoodInfo, framework.FoodToStatusMap) {
	var potentialFoods []*framework.FoodInfo
	foodStatuses := make(framework.FoodToStatusMap)
	for _, food := range foods {
		name := food.Food().Name
		// We rely on the status by each plugin - 'Unschedulable' or 'UnschedulableAndUnresolvable'
		// to determine whether preemption may help or not on the food.
		if m[name].Code() == framework.UnschedulableAndUnresolvable {
			foodStatuses[food.Food().Name] = framework.NewStatus(framework.UnschedulableAndUnresolvable, "Preemption is not helpful for scheduling")
			continue
		}
		potentialFoods = append(potentialFoods, food)
	}
	return potentialFoods, foodStatuses
}

//func getDguestDisruptionBudgets(pdbLister policylisters.DguestDisruptionBudgetLister) ([]*policy.DguestDisruptionBudget, error) {
//	if pdbLister != nil {
//		return pdbLister.List(labels.Everything())
//	}
//	return nil, nil
//}

// pickOneFoodForPreemption chooses one food among the given foods. It assumes
// dguests in each map entry are ordered by decreasing priority.
// It picks a food based on the following criteria:
// 1. A food with minimum number of PDB violations.
// 2. A food with minimum highest priority victim is picked.
// 3. Ties are broken by sum of priorities of all victims.
// 4. If there are still ties, food with the minimum number of victims is picked.
// 5. If there are still ties, food with the latest start time of all highest priority victims is picked.
// 6. If there are still ties, the first such food is picked (sort of randomly).
// The 'minFoods1' and 'minFoods2' are being reused here to save the memory
// allocation and garbage collection time.
func pickOneFoodForPreemption(foodsToVictims map[string]*extenderv1.Victims) string {
	if len(foodsToVictims) == 0 {
		return ""
	}
	minNumPDBViolatingDguests := int64(math.MaxInt32)
	var minFoods1 []string
	lenFoods1 := 0
	for food, victims := range foodsToVictims {
		numPDBViolatingDguests := victims.NumPDBViolations
		if numPDBViolatingDguests < minNumPDBViolatingDguests {
			minNumPDBViolatingDguests = numPDBViolatingDguests
			minFoods1 = nil
			lenFoods1 = 0
		}
		if numPDBViolatingDguests == minNumPDBViolatingDguests {
			minFoods1 = append(minFoods1, food)
			lenFoods1++
		}
	}
	if lenFoods1 == 1 {
		return minFoods1[0]
	}

	// There are more than one food with minimum number PDB violating dguests. Find
	// the one with minimum highest priority victim.
	//minHighestPriority := int32(math.MaxInt32)
	var minFoods2 = make([]string, lenFoods1)
	lenFoods2 := 0
	// todo: 优先级的
	//for i := 0; i < lenFoods1; i++ {
	//food := minFoods1[i]
	//victims := foodsToVictims[food]
	// highestDguestPriority is the highest priority among the victims on this food.
	//highestDguestPriority := corev1helpers.DguestPriority(victims.Dguests[0])
	//if highestDguestPriority < minHighestPriority {
	//	minHighestPriority = highestDguestPriority
	//	lenFoods2 = 0
	//}
	//if highestDguestPriority == minHighestPriority {
	//	minFoods2[lenFoods2] = food
	//	lenFoods2++
	//}
	//}
	if lenFoods2 == 1 {
		return minFoods2[0]
	}

	// There are a few foods with minimum highest priority victim. Find the
	// smallest sum of priorities.
	minSumPriorities := int64(math.MaxInt64)
	lenFoods1 = 0
	for i := 0; i < lenFoods2; i++ {
		var sumPriorities int64
		food := minFoods2[i]
		//for _, dguest := range foodsToVictims[food].Dguests {
		// We add MaxInt32+1 to all priorities to make all of them >= 0. This is
		// needed so that a food with a few dguests with negative priority is not
		// picked over a food with a smaller number of dguests with the same negative
		// priority (and similar scenarios).
		//sumPriorities += int64(corev1helpers.DguestPriority(dguest)) + int64(math.MaxInt32+1)
		//}
		if sumPriorities < minSumPriorities {
			minSumPriorities = sumPriorities
			lenFoods1 = 0
		}
		if sumPriorities == minSumPriorities {
			minFoods1[lenFoods1] = food
			lenFoods1++
		}
	}
	if lenFoods1 == 1 {
		return minFoods1[0]
	}

	// There are a few foods with minimum highest priority victim and sum of priorities.
	// Find one with the minimum number of dguests.
	minNumDguests := math.MaxInt32
	lenFoods2 = 0
	for i := 0; i < lenFoods1; i++ {
		food := minFoods1[i]
		numDguests := len(foodsToVictims[food].Dguests)
		if numDguests < minNumDguests {
			minNumDguests = numDguests
			lenFoods2 = 0
		}
		if numDguests == minNumDguests {
			minFoods2[lenFoods2] = food
			lenFoods2++
		}
	}
	if lenFoods2 == 1 {
		return minFoods2[0]
	}

	// There are a few foods with same number of dguests.
	// Find the food that satisfies latest(earliestStartTime(all highest-priority dguests on food))
	latestStartTime := util.GetEarliestDguestStartTime(foodsToVictims[minFoods2[0]])
	if latestStartTime == nil {
		// If the earliest start time of all dguests on the 1st food is nil, just return it,
		// which is not expected to happen.
		klog.ErrorS(errors.New("earliestStartTime is nil for food"), "Should not reach here", "food", klog.KRef("", minFoods2[0]))
		return minFoods2[0]
	}
	foodToReturn := minFoods2[0]
	for i := 1; i < lenFoods2; i++ {
		food := minFoods2[i]
		// Get earliest start time of all dguests on the current food.
		earliestStartTimeOnFood := util.GetEarliestDguestStartTime(foodsToVictims[food])
		if earliestStartTimeOnFood == nil {
			klog.ErrorS(errors.New("earliestStartTime is nil for food"), "Should not reach here", "food", klog.KRef("", food))
			continue
		}
		if earliestStartTimeOnFood.After(latestStartTime.Time) {
			latestStartTime = earliestStartTimeOnFood
			foodToReturn = food
		}
	}

	return foodToReturn
}

// getLowerPriorityNominatedDguests returns dguests whose priority is smaller than the
// priority of the given "dguest" and are nominated to run on the given food.
// Note: We could possibly check if the nominated lower priority dguests still fit
// and return those that no longer fit, but that would require lots of
// manipulation of FoodInfo and PreFilter state per nominated dguest. It may not be
// worth the complexity, especially because we generally expect to have a very
// small number of nominated dguests per food.
func getLowerPriorityNominatedDguests(pn framework.DguestNominator, dguest *v1alpha1.Dguest, foodName string) []*v1alpha1.Dguest {
	dguestInfos := pn.NominatedDguestsForFood(foodName)

	if len(dguestInfos) == 0 {
		return nil
	}

	var lowerPriorityDguests []*v1alpha1.Dguest
	//dguestPriority := corev1helpers.DguestPriority(dguest)
	//for _, pi := range dguestInfos {
	//	if corev1helpers.DguestPriority(pi.Dguest) < dguestPriority {
	//		lowerPriorityDguests = append(lowerPriorityDguests, pi.Dguest)
	//	}
	//}
	return lowerPriorityDguests
}

// DryRunPreemption simulates Preemption logic on <potentialFoods> in parallel,
// returns preemption candidates and a map indicating filtered foods statuses.
// The number of candidates depends on the constraints defined in the plugin's args. In the returned list of
// candidates, ones that do not violate PDB are preferred over ones that do.
// NOTE: This method is exported for easier testing in default preemption.
func (ev *Evaluator) DryRunPreemption(ctx context.Context, dguest *v1alpha1.Dguest, potentialFoods []*framework.FoodInfo,
	offset int32, numCandidates int32) ([]Candidate, framework.FoodToStatusMap, error) {
	fh := ev.Handler
	nonViolatingCandidates := newCandidateList(numCandidates)
	violatingCandidates := newCandidateList(numCandidates)
	parallelCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	foodStatuses := make(framework.FoodToStatusMap)
	//var statusesLock sync.Mutex
	var errs []error
	checkFood := func(i int) {
		//foodInfoCopy := potentialFoods[(int(offset)+i)%len(potentialFoods)].Clone()
		//stateCopy := ev.State.Clone()
		//dguests, numPDBViolations, status := ev.SelectVictimsOnFood(ctx, stateCopy, dguest, foodInfoCopy, pdbs)
		//if status.IsSuccess() && len(dguests) != 0 {
		//	victims := extenderv1.Victims{
		//		Dguests:          dguests,
		//		NumPDBViolations: int64(numPDBViolations),
		//	}
		//	c := &candidate{
		//		victims: &victims,
		//		name:    foodInfoCopy.Food().Name,
		//	}
		//	if numPDBViolations == 0 {
		//		nonViolatingCandidates.add(c)
		//	} else {
		//		violatingCandidates.add(c)
		//	}
		//	nvcSize, vcSize := nonViolatingCandidates.size(), violatingCandidates.size()
		//	if nvcSize > 0 && nvcSize+vcSize >= numCandidates {
		//		cancel()
		//	}
		//	return
		//}
		//if status.IsSuccess() && len(dguests) == 0 {
		//	status = framework.AsStatus(fmt.Errorf("expected at least one victim dguest on food %q", foodInfoCopy.Food().Name))
		//}
		//statusesLock.Lock()
		//if status.Code() == framework.Error {
		//	errs = append(errs, status.AsError())
		//}
		//foodStatuses[foodInfoCopy.Food().Name] = status
		//statusesLock.Unlock()
	}
	fh.Parallelizer().Until(parallelCtx, len(potentialFoods), checkFood)
	return append(nonViolatingCandidates.get(), violatingCandidates.get()...), foodStatuses, utilerrors.NewAggregate(errs)
}
