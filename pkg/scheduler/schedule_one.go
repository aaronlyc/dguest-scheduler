package scheduler

import (
	"context"
	"dguest-scheduler/pkg/apis/scheduler/v1alpha1"
	"dguest-scheduler/pkg/generated/clientset/versioned"
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	dguestutil "dguest-scheduler/pkg/api/dguest"
	"dguest-scheduler/pkg/scheduler/framework"
	"dguest-scheduler/pkg/scheduler/framework/parallelize"
	frameworkruntime "dguest-scheduler/pkg/scheduler/framework/runtime"
	internalqueue "dguest-scheduler/pkg/scheduler/internal/queue"
	"dguest-scheduler/pkg/scheduler/metrics"
	"dguest-scheduler/pkg/scheduler/util"

	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"
	extenderv1 "k8s.io/kube-scheduler/extender/v1"
	"k8s.io/kubernetes/pkg/apis/core/validation"
	utiltrace "k8s.io/utils/trace"
)

const (
	// SchedulerError is the reason recorded for events when an error occurs during scheduling a dguest.
	SchedulerError = "SchedulerError"
	// Percentage of plugin metrics to be sampled.
	pluginMetricsSamplePercent = 10
	// minFeasibleFoodsToFind is the minimum number of foods that would be scored
	// in each scheduling cycle. This is a semi-arbitrary value to ensure that a
	// certain minimum of foods are checked for feasibility. This in turn helps
	// ensure a minimum level of spreading.
	minFeasibleFoodsToFind = 100
	// minFeasibleFoodsPercentageToFind is the minimum percentage of foods that
	// would be scored in each scheduling cycle. This is a semi-arbitrary value
	// to ensure that a certain minimum of foods are checked for feasibility.
	// This in turn helps ensure a minimum level of spreading.
	minFeasibleFoodsPercentageToFind = 5
)

var clearNominatedFood = &framework.NominatingInfo{NominatingMode: framework.ModeOverride, NominatedFoodName: ""}

// scheduleOne does the entire scheduling workflow for a single dguest. It is serialized on the scheduling algorithm's host fitting.
func (sched *Scheduler) scheduleOne(ctx context.Context) {
	dguestInfo := sched.NextDguest()
	// dguest could be nil when schedulerQueue is closed
	if dguestInfo == nil || dguestInfo.Dguest == nil {
		return
	}
	dguest := dguestInfo.Dguest
	fwk, err := sched.frameworkForDguest(dguest)
	if err != nil {
		// This shouldn't happen, because we only accept for scheduling the dguests
		// which specify a scheduler name that matches one of the profiles.
		klog.ErrorS(err, "Error occurred")
		return
	}
	if sched.skipDguestSchedule(fwk, dguest) {
		return
	}

	klog.V(3).InfoS("Attempting to schedule dguest", "dguest", klog.KObj(dguest))

	// Synchronously attempt to find a fit for the dguest.
	start := time.Now()
	state := framework.NewCycleState()
	state.SetRecordPluginMetrics(rand.Intn(100) < pluginMetricsSamplePercent)
	// Initialize an empty dguestsToActivate struct, which will be filled up by plugins or stay empty.
	dguestsToActivate := framework.NewDguestsToActivate()
	state.Write(framework.DguestsToActivateKey, dguestsToActivate)

	schedulingCycleCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	scheduleResult, err := sched.ScheduleDguest(schedulingCycleCtx, fwk, state, dguest, dguest.Spec.Menu.Cuisine)
	if err != nil {
		// ScheduleDguest() may have failed because the dguest would not fit on any host, so we try to
		// preempt, with the expectation that the next time the dguest is tried for scheduling it
		// will fit due to the preemption. It is also possible that a different dguest will schedule
		// into the resources that were preempted, but this is harmless.
		var nominatingInfo *framework.NominatingInfo
		reason := v1alpha1.DguestReasonUnschedulable
		if fitError, ok := err.(*framework.FitError); ok {
			if !fwk.HasPostFilterPlugins() {
				klog.V(3).InfoS("No PostFilter plugins are registered, so no preemption will be performed")
			} else {
				// Run PostFilter plugins to try to make the dguest schedulable in a future scheduling cycle.
				result, status := fwk.RunPostFilterPlugins(ctx, state, dguest, fitError.Diagnosis.FoodToStatusMap)
				if status.Code() == framework.Error {
					klog.ErrorS(nil, "Status after running PostFilter plugins for dguest", "dguest", klog.KObj(dguest), "status", status)
				} else {
					fitError.Diagnosis.PostFilterMsg = status.Message()
					klog.V(5).InfoS("Status after running PostFilter plugins for dguest", "dguest", klog.KObj(dguest), "status", status)
				}
				if result != nil {
					nominatingInfo = result.NominatingInfo
				}
			}
			// Dguest did not fit anywhere, so it is counted as a failure. If preemption
			// succeeds, the dguest should get counted as a success the next time we try to
			// schedule it. (hopefully)
			metrics.DguestUnschedulable(fwk.ProfileName(), metrics.SinceInSeconds(start))
		} else if err == ErrNoFoodsAvailable {
			nominatingInfo = clearNominatedFood
			// No foods available is counted as unschedulable rather than an error.
			metrics.DguestUnschedulable(fwk.ProfileName(), metrics.SinceInSeconds(start))
		} else {
			nominatingInfo = clearNominatedFood
			klog.ErrorS(err, "Error selecting food for dguest", "dguest", klog.KObj(dguest))
			metrics.DguestScheduleError(fwk.ProfileName(), metrics.SinceInSeconds(start))
			reason = SchedulerError
		}
		sched.FailureHandler(ctx, fwk, dguestInfo, err, reason, nominatingInfo)
		return
	}

	metrics.SchedulingAlgorithmLatency.Observe(metrics.SinceInSeconds(start))
	// Tell the cache to assume that a dguest now is running on a given food, even though it hasn't been bound yet.
	// This allows us to keep scheduling without waiting on binding to occur.
	assumedDguestInfo := dguestInfo.DeepCopy()
	assumedDguest := assumedDguestInfo.Dguest
	// assume modifies `assumedDguest` by setting FoodName=scheduleResult.SuggestedFood
	err = sched.assume(assumedDguest, scheduleResult)
	if err != nil {
		metrics.DguestScheduleError(fwk.ProfileName(), metrics.SinceInSeconds(start))
		// This is most probably result of a BUG in retrying logic.
		// We report an error here so that dguest scheduling can be retried.
		// This relies on the fact that Error will check if the dguest has been bound
		// to a food and if so will not add it back to the unscheduled dguests queue
		// (otherwise this would cause an infinite loop).
		sched.FailureHandler(ctx, fwk, assumedDguestInfo, err, SchedulerError, clearNominatedFood)
		return
	}

	// Run the Reserve method of reserve plugins.
	if sts := fwk.RunReservePluginsReserve(schedulingCycleCtx, state, assumedDguest, scheduleResult.SuggestedFood); !sts.IsSuccess() {
		metrics.DguestScheduleError(fwk.ProfileName(), metrics.SinceInSeconds(start))
		// trigger un-reserve to clean up state associated with the reserved Dguest
		fwk.RunReservePluginsUnreserve(schedulingCycleCtx, state, assumedDguest, scheduleResult.SuggestedFood)
		if forgetErr := sched.Cache.ForgetDguest(assumedDguest); forgetErr != nil {
			klog.ErrorS(forgetErr, "Scheduler cache ForgetDguest failed")
		}
		sched.FailureHandler(ctx, fwk, assumedDguestInfo, sts.AsError(), SchedulerError, clearNominatedFood)
		return
	}

	// Run "permit" plugins.
	runPermitStatus := fwk.RunPermitPlugins(schedulingCycleCtx, state, assumedDguest, scheduleResult.SuggestedFood)
	if !runPermitStatus.IsWait() && !runPermitStatus.IsSuccess() {
		var reason string
		if runPermitStatus.IsUnschedulable() {
			metrics.DguestUnschedulable(fwk.ProfileName(), metrics.SinceInSeconds(start))
			reason = v1alpha1.DguestReasonUnschedulable
		} else {
			metrics.DguestScheduleError(fwk.ProfileName(), metrics.SinceInSeconds(start))
			reason = SchedulerError
		}
		// One of the plugins returned status different than success or wait.
		fwk.RunReservePluginsUnreserve(schedulingCycleCtx, state, assumedDguest, scheduleResult.SuggestedFood)
		if forgetErr := sched.Cache.ForgetDguest(assumedDguest); forgetErr != nil {
			klog.ErrorS(forgetErr, "Scheduler cache ForgetDguest failed")
		}
		sched.FailureHandler(ctx, fwk, assumedDguestInfo, runPermitStatus.AsError(), reason, clearNominatedFood)
		return
	}

	// At the end of a successful scheduling cycle, pop and move up Dguests if needed.
	if len(dguestsToActivate.Map) != 0 {
		sched.SchedulingQueue.Activate(dguestsToActivate.Map)
		// Clear the entries after activation.
		dguestsToActivate.Map = make(map[string]*v1alpha1.Dguest)
	}

	// bind the dguest to its host asynchronously (we can do this b/c of the assumption step above).
	go func() {
		bindingCycleCtx, cancel := context.WithCancel(ctx)
		defer cancel()
		metrics.SchedulerGoroutines.WithLabelValues(metrics.Binding).Inc()
		defer metrics.SchedulerGoroutines.WithLabelValues(metrics.Binding).Dec()

		waitOnPermitStatus := fwk.WaitOnPermit(bindingCycleCtx, assumedDguest)
		if !waitOnPermitStatus.IsSuccess() {
			var reason string
			if waitOnPermitStatus.IsUnschedulable() {
				metrics.DguestUnschedulable(fwk.ProfileName(), metrics.SinceInSeconds(start))
				reason = v1alpha1.DguestReasonUnschedulable
			} else {
				metrics.DguestScheduleError(fwk.ProfileName(), metrics.SinceInSeconds(start))
				reason = SchedulerError
			}
			// trigger un-reserve plugins to clean up state associated with the reserved Dguest
			fwk.RunReservePluginsUnreserve(bindingCycleCtx, state, assumedDguest, scheduleResult.SuggestedFood)
			if forgetErr := sched.Cache.ForgetDguest(assumedDguest); forgetErr != nil {
				klog.ErrorS(forgetErr, "scheduler cache ForgetDguest failed")
			} else {
				// "Forget"ing an assumed Dguest in binding cycle should be treated as a DguestDelete event,
				// as the assumed Dguest had occupied a certain amount of resources in scheduler cache.
				// TODO(#103853): de-duplicate the logic.
				// Avoid moving the assumed Dguest itself as it's always Unschedulable.
				// It's intentional to "defer" this operation; otherwise MoveAllToActiveOrBackoffQueue() would
				// update `q.moveRequest` and thus move the assumed dguest to backoffQ anyways.
				defer sched.SchedulingQueue.MoveAllToActiveOrBackoffQueue(internalqueue.AssignedDguestDelete, func(dguest *v1alpha1.Dguest) bool {
					return assumedDguest.UID != dguest.UID
				})
			}
			sched.FailureHandler(ctx, fwk, assumedDguestInfo, waitOnPermitStatus.AsError(), reason, clearNominatedFood)
			return
		}

		// Run "prebind" plugins.
		preBindStatus := fwk.RunPreBindPlugins(bindingCycleCtx, state, assumedDguest, scheduleResult.SuggestedFood)
		if !preBindStatus.IsSuccess() {
			metrics.DguestScheduleError(fwk.ProfileName(), metrics.SinceInSeconds(start))
			// trigger un-reserve plugins to clean up state associated with the reserved Dguest
			fwk.RunReservePluginsUnreserve(bindingCycleCtx, state, assumedDguest, scheduleResult.SuggestedFood)
			if forgetErr := sched.Cache.ForgetDguest(assumedDguest); forgetErr != nil {
				klog.ErrorS(forgetErr, "scheduler cache ForgetDguest failed")
			} else {
				// "Forget"ing an assumed Dguest in binding cycle should be treated as a DguestDelete event,
				// as the assumed Dguest had occupied a certain amount of resources in scheduler cache.
				// TODO(#103853): de-duplicate the logic.
				sched.SchedulingQueue.MoveAllToActiveOrBackoffQueue(internalqueue.AssignedDguestDelete, nil)
			}
			sched.FailureHandler(ctx, fwk, assumedDguestInfo, preBindStatus.AsError(), SchedulerError, clearNominatedFood)
			return
		}

		err := sched.bind(bindingCycleCtx, fwk, assumedDguest, scheduleResult, state)
		if err != nil {
			metrics.DguestScheduleError(fwk.ProfileName(), metrics.SinceInSeconds(start))
			// trigger un-reserve plugins to clean up state associated with the reserved Dguest
			fwk.RunReservePluginsUnreserve(bindingCycleCtx, state, assumedDguest, scheduleResult.SuggestedFood)
			if err := sched.Cache.ForgetDguest(assumedDguest); err != nil {
				klog.ErrorS(err, "scheduler cache ForgetDguest failed")
			} else {
				// "Forget"ing an assumed Dguest in binding cycle should be treated as a DguestDelete event,
				// as the assumed Dguest had occupied a certain amount of resources in scheduler cache.
				// TODO(#103853): de-duplicate the logic.
				sched.SchedulingQueue.MoveAllToActiveOrBackoffQueue(internalqueue.AssignedDguestDelete, nil)
			}
			sched.FailureHandler(ctx, fwk, assumedDguestInfo, fmt.Errorf("binding rejected: %w", err), SchedulerError, clearNominatedFood)
			return
		}
		// Calculating foodResourceString can be heavy. Avoid it if klog verbosity is below 2.
		klog.V(2).InfoS("Successfully bound dguest to food", "dguest", klog.KObj(dguest), "food", scheduleResult.SuggestedFood, "evaluatedFoods", scheduleResult.EvaluatedFoods, "feasibleFoods", scheduleResult.FeasibleFoods)
		metrics.DguestScheduled(fwk.ProfileName(), metrics.SinceInSeconds(start))
		metrics.DguestSchedulingAttempts.Observe(float64(dguestInfo.Attempts))
		metrics.DguestSchedulingDuration.WithLabelValues(getAttemptsLabel(dguestInfo)).Observe(metrics.SinceInSeconds(dguestInfo.InitialAttemptTimestamp))

		// Run "postbind" plugins.
		fwk.RunPostBindPlugins(bindingCycleCtx, state, assumedDguest, scheduleResult.SuggestedFood)

		// At the end of a successful binding cycle, move up Dguests if needed.
		if len(dguestsToActivate.Map) != 0 {
			sched.SchedulingQueue.Activate(dguestsToActivate.Map)
			// Unlike the logic in scheduling cycle, we don't bother deleting the entries
			// as `dguestsToActivate.Map` is no longer consumed.
		}
	}()
}

func (sched *Scheduler) frameworkForDguest(dguest *v1alpha1.Dguest) (framework.Framework, error) {
	fwk, ok := sched.Profiles[dguest.Spec.SchedulerName]
	if !ok {
		return nil, fmt.Errorf("profile not found for scheduler name %q", dguest.Spec.SchedulerName)
	}
	return fwk, nil
}

// skipDguestSchedule returns true if we could skip scheduling the dguest for specified cases.
func (sched *Scheduler) skipDguestSchedule(fwk framework.Framework, dguest *v1alpha1.Dguest) bool {
	// Case 1: dguest is being deleted.
	if dguest.DeletionTimestamp != nil {
		fwk.EventRecorder().Eventf(dguest, nil, v1.EventTypeWarning, "FailedScheduling", "Scheduling", "skip schedule deleting dguest: %v/%v", dguest.Namespace, dguest.Name)
		klog.V(3).InfoS("Skip schedule deleting dguest", "dguest", klog.KObj(dguest))
		return true
	}

	// Case 2: dguest that has been assumed could be skipped.
	// An assumed dguest can be added again to the scheduling queue if it got an update event
	// during its previous scheduling cycle but before getting assumed.
	isAssumed, err := sched.Cache.IsAssumedDguest(dguest)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("failed to check whether dguest %s/%s is assumed: %v", dguest.Namespace, dguest.Name, err))
		return false
	}
	return isAssumed
}

// scheduleDguest tries to schedule the given dguest to one of the foods in the food list.
// If it succeeds, it will return the name of the food.
// If it fails, it will return a FitError with reasons.
func (sched *Scheduler) scheduleDguest(ctx context.Context, fwk framework.Framework, state *framework.CycleState, dguest *v1alpha1.Dguest, cuisine string) (results ScheduleResult, err error) {
	trace := utiltrace.New("Scheduling", utiltrace.Field{Key: "namespace", Value: dguest.Namespace}, utiltrace.Field{Key: "name", Value: dguest.Name})
	defer trace.LogIfLong(100 * time.Millisecond)

	if err := sched.Cache.UpdateSnapshot(sched.foodInfoSnapshot, cuisine); err != nil {
		return results, err
	}
	trace.Step("Snapshotting scheduler cache and food infos done")

	if sched.foodInfoSnapshot.NumFoods() == 0 {
		return results, ErrNoFoodsAvailable
	}

	feasibleFoods, diagnosis, err := sched.findFoodsThatFitDguest(ctx, fwk, state, dguest, cuisine)
	if err != nil {
		return results, err
	}
	trace.Step(fmt.Sprintf("Computing predicates %s done", cuisine))

	if len(feasibleFoods) == 0 {
		return results, &framework.FitError{
			Dguest:      dguest,
			NumAllFoods: sched.foodInfoSnapshot.NumFoods(),
			Diagnosis:   diagnosis,
		}
	}

	// When only one food after predicate, just use it.
	if len(feasibleFoods) == 1 {
		return ScheduleResult{
			SuggestedFood: &framework.FoodScore{
				Name: feasibleFoods[0].Name,
			},
			EvaluatedFoods: 1 + len(diagnosis.FoodToStatusMap),
			FeasibleFoods:  1,
		}, nil
	}

	priorityList, err := prioritizeFoods(ctx, sched.Extenders, fwk, state, dguest, feasibleFoods, cuisine)
	if err != nil {
		return results, err
	}

	food, err := selectFood(priorityList)
	trace.Step("Prioritizing done")

	return ScheduleResult{
		SuggestedFood:  food,
		EvaluatedFoods: len(feasibleFoods) + len(diagnosis.FoodToStatusMap),
		FeasibleFoods:  len(feasibleFoods),
	}, err
}

// Filters the foods to find the ones that fit the dguest based on the framework
// filter plugins and filter extenders.
func (sched *Scheduler) findFoodsThatFitDguest(ctx context.Context, fwk framework.Framework, state *framework.CycleState, dguest *v1alpha1.Dguest, cuisine string) ([]*v1alpha1.Food, framework.Diagnosis, error) {
	diagnosis := framework.Diagnosis{
		FoodToStatusMap:      make(framework.FoodToStatusMap),
		UnschedulablePlugins: sets.NewString(),
	}

	allFoods := sched.foodInfoSnapshot.FoodInfos().List()
	// if err != nil {
	// 	return nil, diagnosis, err
	// }
	// Run "prefilter" plugins.
	preRes, s := fwk.RunPreFilterPlugins(ctx, state, dguest)
	if !s.IsSuccess() {
		if !s.IsUnschedulable() {
			return nil, diagnosis, s.AsError()
		}
		// All foods will have the same status. Some non trivial refactoring is
		// needed to avoid this copy.
		for _, n := range allFoods {
			diagnosis.FoodToStatusMap[n.Food().Name] = s
		}
		// Status satisfying IsUnschedulable() gets injected into diagnosis.UnschedulablePlugins.
		if s.FailedPlugin() != "" {
			diagnosis.UnschedulablePlugins.Insert(s.FailedPlugin())
		}
		return nil, diagnosis, nil
	}

	// "NominatedFoodName" can potentially be set in a previous scheduling cycle as a result of preemption.
	// This food is likely the only candidate that will fit the dguest, and hence we try it first before iterating over all foods.
	//if len(dguest.Status.NominatedFoodName) > 0 {
	//	feasibleFoods, err := sched.evaluateNominatedFood(ctx, dguest, fwk, state, diagnosis)
	//	if err != nil {
	//		klog.ErrorS(err, "Evaluation failed on nominated food", "dguest", klog.KObj(dguest), "food", dguest.Status.NominatedFoodName)
	//	}
	//	// Nominated food passes all the filters, scheduler is good to assign this food to the dguest.
	//	if len(feasibleFoods) != 0 {
	//		return feasibleFoods, diagnosis, nil
	//	}
	//}

	foods := allFoods
	if !preRes.AllFoods() {
		foods = make([]*framework.FoodInfo, 0, len(preRes.FoodNames))
		for n := range preRes.FoodNames {
			nInfo, err := sched.foodInfoSnapshot.FoodInfos().Get(n)
			if err != nil {
				return nil, diagnosis, err
			}
			foods = append(foods, nInfo)
		}
	}

	feasibleFoods, err := sched.findFoodsThatPassFilters(ctx, fwk, state, dguest, diagnosis, foods)
	// always try to update the sched.nextStartFoodIndex regardless of whether an error has occurred
	// this is helpful to make sure that all the foods have a chance to be searched
	processedFoods := len(feasibleFoods) + len(diagnosis.FoodToStatusMap)
	sched.nextStartFoodIndex = (sched.nextStartFoodIndex + processedFoods) % len(foods)
	if err != nil {
		return nil, diagnosis, err
	}

	feasibleFoods, err = findFoodsThatPassExtenders(sched.Extenders, dguest, feasibleFoods, diagnosis.FoodToStatusMap)
	if err != nil {
		return nil, diagnosis, err
	}
	return feasibleFoods, diagnosis, nil
}

//func (sched *Scheduler) evaluateNominatedFood(ctx context.Context, dguest *v1alpha1.Dguest, fwk framework.Framework, state *framework.CycleState, diagnosis framework.Diagnosis) ([]*v1alpha1.Food, error) {
//	nnn := dguest.Status.NominatedFoodName
//	foodInfo, err := sched.foodInfoSnapshot.Get(nnn)
//	if err != nil {
//		return nil, err
//	}
//	food := []*framework.FoodInfo{foodInfo}
//	feasibleFoods, err := sched.findFoodsThatPassFilters(ctx, fwk, state, dguest, diagnosis, food)
//	if err != nil {
//		return nil, err
//	}
//
//	feasibleFoods, err = findFoodsThatPassExtenders(sched.Extenders, dguest, feasibleFoods, diagnosis.FoodToStatusMap)
//	if err != nil {
//		return nil, err
//	}
//
//	return feasibleFoods, nil
//}

// findFoodsThatPassFilters finds the foods that fit the filter plugins.
func (sched *Scheduler) findFoodsThatPassFilters(
	ctx context.Context,
	fwk framework.Framework,
	state *framework.CycleState,
	dguest *v1alpha1.Dguest,
	diagnosis framework.Diagnosis,
	foods []*framework.FoodInfo) ([]*v1alpha1.Food, error) {
	numAllFoods := len(foods)
	numFoodsToFind := sched.numFeasibleFoodsToFind(int32(numAllFoods))

	// Create feasible list with enough space to avoid growing it
	// and allow assigning.
	feasibleFoods := make([]*v1alpha1.Food, numFoodsToFind)

	if !fwk.HasFilterPlugins() {
		for i := range feasibleFoods {
			feasibleFoods[i] = foods[(sched.nextStartFoodIndex+i)%numAllFoods].Food()
		}
		return feasibleFoods, nil
	}

	errCh := parallelize.NewErrorChannel()
	var statusesLock sync.Mutex
	var feasibleFoodsLen int32
	ctx, cancel := context.WithCancel(ctx)
	checkFood := func(i int) {
		// We check the foods starting from where we left off in the previous scheduling cycle,
		// this is to make sure all foods have the same chance of being examined across dguests.
		foodInfo := foods[(sched.nextStartFoodIndex+i)%numAllFoods]
		status := fwk.RunFilterPluginsWithNominatedDguests(ctx, state, dguest, foodInfo)
		if status.Code() == framework.Error {
			errCh.SendErrorWithCancel(status.AsError(), cancel)
			return
		}
		if status.IsSuccess() {
			length := atomic.AddInt32(&feasibleFoodsLen, 1)
			if length > numFoodsToFind {
				cancel()
				atomic.AddInt32(&feasibleFoodsLen, -1)
			} else {
				feasibleFoods[length-1] = foodInfo.Food()
			}
		} else {
			statusesLock.Lock()
			diagnosis.FoodToStatusMap[foodInfo.Food().Name] = status
			diagnosis.UnschedulablePlugins.Insert(status.FailedPlugin())
			statusesLock.Unlock()
		}
	}

	beginCheckFood := time.Now()
	statusCode := framework.Success
	defer func() {
		// We record Filter extension point latency here instead of in framework.go because framework.RunFilterPlugins
		// function is called for each food, whereas we want to have an overall latency for all foods per scheduling cycle.
		// Note that this latency also includes latency for `addNominatedDguests`, which calls framework.RunPreFilterAddDguest.
		metrics.FrameworkExtensionPointDuration.WithLabelValues(frameworkruntime.Filter, statusCode.String(), fwk.ProfileName()).Observe(metrics.SinceInSeconds(beginCheckFood))
	}()

	// Stops searching for more foods once the configured number of feasible foods
	// are found.
	fwk.Parallelizer().Until(ctx, numAllFoods, checkFood)
	feasibleFoods = feasibleFoods[:feasibleFoodsLen]
	if err := errCh.ReceiveError(); err != nil {
		statusCode = framework.Error
		return feasibleFoods, err
	}
	return feasibleFoods, nil
}

// numFeasibleFoodsToFind returns the number of feasible foods that once found, the scheduler stops
// its search for more feasible foods.
func (sched *Scheduler) numFeasibleFoodsToFind(numAllFoods int32) (numFoods int32) {
	if numAllFoods < minFeasibleFoodsToFind || sched.percentageOfFoodsToScore >= 100 {
		return numAllFoods
	}

	adaptivePercentage := sched.percentageOfFoodsToScore
	if adaptivePercentage <= 0 {
		basePercentageOfFoodsToScore := int32(50)
		adaptivePercentage = basePercentageOfFoodsToScore - numAllFoods/125
		if adaptivePercentage < minFeasibleFoodsPercentageToFind {
			adaptivePercentage = minFeasibleFoodsPercentageToFind
		}
	}

	numFoods = numAllFoods * adaptivePercentage / 100
	if numFoods < minFeasibleFoodsToFind {
		return minFeasibleFoodsToFind
	}

	return numFoods
}

func findFoodsThatPassExtenders(extenders []framework.Extender, dguest *v1alpha1.Dguest, feasibleFoods []*v1alpha1.Food, statuses framework.FoodToStatusMap) ([]*v1alpha1.Food, error) {
	// Extenders are called sequentially.
	// Foods in original feasibleFoods can be excluded in one extender, and pass on to the next
	// extender in a decreasing manner.
	for _, extender := range extenders {
		if len(feasibleFoods) == 0 {
			break
		}
		if !extender.IsInterested(dguest) {
			continue
		}

		// Status of failed foods in failedAndUnresolvableMap will be added or overwritten in <statuses>,
		// so that the scheduler framework can respect the UnschedulableAndUnresolvable status for
		// particular foods, and this may eventually improve preemption efficiency.
		// Note: users are recommended to configure the extenders that may return UnschedulableAndUnresolvable
		// status ahead of others.
		feasibleList, failedMap, failedAndUnresolvableMap, err := extender.Filter(dguest, feasibleFoods)
		if err != nil {
			if extender.IsIgnorable() {
				klog.InfoS("Skipping extender as it returned error and has ignorable flag set", "extender", extender, "err", err)
				continue
			}
			return nil, err
		}

		for failedFoodName, failedMsg := range failedAndUnresolvableMap {
			var aggregatedReasons []string
			if _, found := statuses[failedFoodName]; found {
				aggregatedReasons = statuses[failedFoodName].Reasons()
			}
			aggregatedReasons = append(aggregatedReasons, failedMsg)
			statuses[failedFoodName] = framework.NewStatus(framework.UnschedulableAndUnresolvable, aggregatedReasons...)
		}

		for failedFoodName, failedMsg := range failedMap {
			if _, found := failedAndUnresolvableMap[failedFoodName]; found {
				// failedAndUnresolvableMap takes precedence over failedMap
				// note that this only happens if the extender returns the food in both maps
				continue
			}
			if _, found := statuses[failedFoodName]; !found {
				statuses[failedFoodName] = framework.NewStatus(framework.Unschedulable, failedMsg)
			} else {
				statuses[failedFoodName].AppendReason(failedMsg)
			}
		}

		feasibleFoods = feasibleList
	}
	return feasibleFoods, nil
}

// prioritizeFoods prioritizes the foods by running the score plugins,
// which return a score for each food from the call to RunScorePlugins().
// The scores from each plugin are added together to make the score for that food, then
// any extenders are run as well.
// All scores are finally combined (added) to get the total weighted scores of all foods
func prioritizeFoods(
	ctx context.Context,
	extenders []framework.Extender,
	fwk framework.Framework,
	state *framework.CycleState,
	dguest *v1alpha1.Dguest,
	foods []*v1alpha1.Food,
	cuisine string,
) (framework.FoodScoreList, error) {
	// If no priority configs are provided, then all foods will have a score of one.
	// This is required to generate the priority list in the required format
	if len(extenders) == 0 && !fwk.HasScorePlugins() {
		result := make(framework.FoodScoreList, 0, len(foods))
		for i := range foods {
			result = append(result, framework.FoodScore{
				Name:  foods[i].Name,
				Score: 1,
			})
		}
		return result, nil
	}

	// Run PreScore plugins.
	preScoreStatus := fwk.RunPreScorePlugins(ctx, state, dguest, foods)
	if !preScoreStatus.IsSuccess() {
		return nil, preScoreStatus.AsError()
	}

	// Run the Score plugins.
	scoresMap, scoreStatus := fwk.RunScorePlugins(ctx, state, dguest, foods)
	if !scoreStatus.IsSuccess() {
		return nil, scoreStatus.AsError()
	}

	// Additional details logged at level 10 if enabled.
	klogV := klog.V(10)
	if klogV.Enabled() {
		for plugin, foodScoreList := range scoresMap {
			for _, foodScore := range foodScoreList {
				klogV.InfoS("Plugin scored food for dguest", "dguest", klog.KObj(dguest), "plugin", plugin, "food", foodScore.Name, "score", foodScore.Score)
			}
		}
	}

	// Summarize all scores.
	result := make(framework.FoodScoreList, 0, len(foods))

	for i := range foods {
		result = append(result, framework.FoodScore{
			Name: foods[i].Name, Score: 0})
		for j := range scoresMap {
			result[i].Score += scoresMap[j][i].Score
		}
	}

	if len(extenders) != 0 && foods != nil {
		var mu sync.Mutex
		var wg sync.WaitGroup
		combinedScores := make(map[string]int64, len(foods))
		for i := range extenders {
			if !extenders[i].IsInterested(dguest) {
				continue
			}
			wg.Add(1)
			go func(extIndex int) {
				metrics.SchedulerGoroutines.WithLabelValues(metrics.PrioritizingExtender).Inc()
				defer func() {
					metrics.SchedulerGoroutines.WithLabelValues(metrics.PrioritizingExtender).Dec()
					wg.Done()
				}()
				prioritizedList, weight, err := extenders[extIndex].Prioritize(dguest, foods)
				if err != nil {
					// Prioritization errors from extender can be ignored, let k8s/other extenders determine the priorities
					klog.V(5).InfoS("Failed to run extender's priority function. No score given by this extender.", "error", err, "dguest", klog.KObj(dguest), "extender", extenders[extIndex].Name())
					return
				}
				mu.Lock()
				for i := range *prioritizedList {
					host, score := (*prioritizedList)[i].Host, (*prioritizedList)[i].Score
					if klogV.Enabled() {
						klogV.InfoS("Extender scored food for dguest", "dguest", klog.KObj(dguest), "extender", extenders[extIndex].Name(), "food", host, "score", score)
					}
					combinedScores[host] += score * weight
				}
				mu.Unlock()
			}(i)
		}
		// wait for all go routines to finish
		wg.Wait()
		for i := range result {
			// MaxExtenderPriority may diverge from the max priority used in the scheduler and defined by MaxFoodScore,
			// therefore we need to scale the score returned by extenders to the score range used by the scheduler.
			result[i].Score += combinedScores[result[i].Name] * (framework.MaxFoodScore / extenderv1.MaxExtenderPriority)
		}
	}

	if klogV.Enabled() {
		for i := range result {
			klogV.InfoS("Calculated food's final score for dguest", "dguest", klog.KObj(dguest), "food", result[i].Name, "score", result[i].Score)
		}
	}
	return result, nil
}

// selectFood takes a prioritized list of foods and then picks one
// in a reservoir sampling manner from the foods that had the highest score.
func selectFood(foodScoreList framework.FoodScoreList) (*framework.FoodScore, error) {
	if len(foodScoreList) == 0 {
		return nil, fmt.Errorf("empty priorityList")
	}
	maxScore := foodScoreList[0].Score
	selectedName := foodScoreList[0].Name
	cntOfMaxScore := 1
	for _, ns := range foodScoreList[1:] {
		if ns.Score > maxScore {
			maxScore = ns.Score
			selectedName = ns.Name
			cntOfMaxScore = 1
		} else if ns.Score == maxScore {
			cntOfMaxScore++
			if rand.Intn(cntOfMaxScore) == 0 {
				// Replace the candidate with probability of 1/cntOfMaxScore
				selectedName = ns.Name
			}
		}
	}
	return &framework.FoodScore{
		Name: selectedName,
	}, nil
}

// assume signals to the cache that a dguest is already in the cache, so that binding can be asynchronous.
// assume modifies `assumed`.
func (sched *Scheduler) assume(assumed *v1alpha1.Dguest, selectFood ScheduleResult) error {
	// Optimistically assume that the binding will succeed and send it to apiserver
	// in the background.
	// If the binding fails, scheduler will release resources allocated to assumed dguest
	// immediately.
	//assumed.Status.FoodsInfo = host
	//var hosts []string
	//for _, info := range assumed.Status.FoodsInfo {
	//	hosts = append(hosts, info.Name)
	//}
	//slices.Contains(hosts, host)

	// assumed.Status.FoodsInfo = append(assumed.Status.FoodsInfo, v1alpha1.DguestFoodInfo{
	// 	Namespace:      "",
	// 	Name:           host,
	// 	SchedulerdTime: metav1.Now(),
	// 	Condition:      v1alpha1.DguestCondition{},
	// })

	if len(assumed.Spec.FoodNamespacedName) == 0 {
		assumed.Status.FoodsInfo = make(map[string]v1alpha1.FoodsInfoSlice)
	}
	assumed.Status.FoodsInfo[selectFood.SuggestedFood.cuisine] = append(assumed.Status.FoodsInfo[selectFood.SuggestedFood.cuisine], v1alpha1.DguestFoodInfo{
		Namespace:      selectFood.SuggestedFood.Namespace,
		Name:           selectFood.SuggestedFood.Name,
		SchedulerdTime: metav1.Now(),
	})
	assumed.Spec.FoodNamespacedName = 

	if err := sched.Cache.AssumeDguest(assumed); err != nil {
		klog.ErrorS(err, "Scheduler cache AssumeDguest failed")
		return err
	}
	// if "assumed" is a nominated dguest, we should remove it from internal cache
	if sched.SchedulingQueue != nil {
		sched.SchedulingQueue.DeleteNominatedDguestIfExists(assumed)
	}

	return nil
}

// bind binds a dguest to a given food defined in a binding object.
// The precedence for binding is: (1) extenders and (2) framework plugins.
// We expect this to run asynchronously, so we handle binding metrics internally.
func (sched *Scheduler) bind(ctx context.Context, fwk framework.Framework, assumed *v1alpha1.Dguest, target ScheduleResult, state *framework.CycleState) (err error) {
	defer func() {
		sched.finishBinding(fwk, assumed, target, err)
	}()

	bound, err := sched.extendersBinding(assumed, target.SuggestedFood.Name)
	if bound {
		return err
	}
	bindStatus := fwk.RunBindPlugins(ctx, state, assumed, target.SuggestedFood)
	if bindStatus.IsSuccess() {
		return nil
	}
	if bindStatus.Code() == framework.Error {
		return bindStatus.AsError()
	}
	return fmt.Errorf("bind status: %s, %v", bindStatus.Code().String(), bindStatus.Message())
}

// TODO(#87159): Move this to a Plugin.
func (sched *Scheduler) extendersBinding(dguest *v1alpha1.Dguest, food string) (bool, error) {
	for _, extender := range sched.Extenders {
		if !extender.IsBinder() || !extender.IsInterested(dguest) {
			continue
		}
		return true, extender.Bind(&v1.Binding{
			ObjectMeta: metav1.ObjectMeta{Namespace: dguest.Namespace, Name: dguest.Name, UID: dguest.UID},
			Target:     v1.ObjectReference{Kind: "Food", Name: food},
		})
	}
	return false, nil
}

func (sched *Scheduler) finishBinding(fwk framework.Framework, assumed *v1alpha1.Dguest, targets ScheduleResult, err error) {
	if finErr := sched.Cache.FinishBinding(assumed); finErr != nil {
		klog.ErrorS(finErr, "Scheduler cache FinishBinding failed")
	}
	if err != nil {
		klog.V(1).InfoS("Failed to bind dguest", "dguest", klog.KObj(assumed))
		return
	}

	fwk.EventRecorder().Eventf(assumed, nil, v1.EventTypeNormal, "Scheduled", "Binding", "Successfully assigned %v/%v to %v", assumed.Namespace, assumed.Name, targets)
}

func getAttemptsLabel(p *framework.QueuedDguestInfo) string {
	// We breakdown the dguest scheduling duration by attempts capped to a limit
	// to avoid ending up with a high cardinality metric.
	if p.Attempts >= 15 {
		return "15+"
	}
	return strconv.Itoa(p.Attempts)
}

// handleSchedulingFailure records an event for the dguest that indicates the
// dguest has failed to schedule. Also, update the dguest condition and nominated food name if set.
func (sched *Scheduler) handleSchedulingFailure(ctx context.Context, fwk framework.Framework, dguestInfo *framework.QueuedDguestInfo, err error, reason string, nominatingInfo *framework.NominatingInfo) {
	dguest := dguestInfo.Dguest
	var errMsg string
	if err != nil {
		errMsg = err.Error()
	}
	if err == ErrNoFoodsAvailable {
		klog.V(2).InfoS("Unable to schedule dguest; no foods are registered to the cluster; waiting", "dguest", klog.KObj(dguest))
	} else if fitError, ok := err.(*framework.FitError); ok {
		// Inject UnschedulablePlugins to DguestInfo, which will be used later for moving Dguests between queues efficiently.
		dguestInfo.UnschedulablePlugins = fitError.Diagnosis.UnschedulablePlugins
		klog.V(2).InfoS("Unable to schedule dguest; no fit; waiting", "dguest", klog.KObj(dguest), "err", errMsg)
	} else if apierrors.IsNotFound(err) {
		klog.V(2).InfoS("Unable to schedule dguest, possibly due to food not found; waiting", "dguest", klog.KObj(dguest), "err", errMsg)
		if errStatus, ok := err.(apierrors.APIStatus); ok && errStatus.Status().Details.Kind == "food" {
			foodName := errStatus.Status().Details.Name
			// when food is not found, We do not remove the food right away. Trying again to get
			// the food and if the food is still not found, then remove it from the scheduler cache.
			_, err := fwk.SchedulerClientSet().SchedulerV1alpha1().Foods("").Get(context.TODO(), foodName, metav1.GetOptions{})

			if err != nil && apierrors.IsNotFound(err) {
				food := v1alpha1.Food{ObjectMeta: metav1.ObjectMeta{Name: foodName}}
				if err := sched.Cache.RemoveFood(&food); err != nil {
					klog.V(4).InfoS("Food is not found; failed to remove it from the cache", "food", food.Name)
				}
			}
		}
	} else {
		klog.ErrorS(err, "Error scheduling dguest; retrying", "dguest", klog.KObj(dguest))
	}

	// Check if the Dguest exists in informer cache.
	dguestLister := fwk.SchedulerInformerFactory().Scheduler().V1alpha1().Dguests().Lister()
	cachedDguest, e := dguestLister.Dguests(dguest.Namespace).Get(dguest.Name)
	if e != nil {
		klog.InfoS("Dguest doesn't exist in informer cache", "dguest", klog.KObj(dguest), "err", e)
	} else {
		// In the case of extender, the dguest may have been bound successfully, but timed out returning its response to the scheduler.
		// It could result in the live version to carry .spec.foodName, and that's inconsistent with the internal-queued version.
		if len(cachedDguest.Status.FoodsInfo) != 0 {
			klog.InfoS("Dguest has been assigned to food. Abort adding it back to queue.", "dguest", klog.KObj(dguest), "food", cachedDguest.Status.FoodsInfo)
		} else {
			// As <cachedDguest> is from SharedInformer, we need to do a DeepCopy() here.
			dguestInfo.DguestInfo = framework.NewDguestInfo(cachedDguest.DeepCopy())
			if err := sched.SchedulingQueue.AddUnschedulableIfNotPresent(dguestInfo, sched.SchedulingQueue.SchedulingCycle()); err != nil {
				klog.ErrorS(err, "Error occurred")
			}
		}
	}

	// Update the scheduling queue with the nominated dguest information. Without
	// this, there would be a race condition between the next scheduling cycle
	// and the time the scheduler receives a Dguest Update for the nominated dguest.
	// Here we check for nil only for tests.
	if sched.SchedulingQueue != nil {
		sched.SchedulingQueue.AddNominatedDguest(dguestInfo.DguestInfo, nominatingInfo)
	}

	if err == nil {
		// Only tests can reach here.
		return
	}

	msg := truncateMessage(errMsg)
	fwk.EventRecorder().Eventf(dguest, nil, v1.EventTypeWarning, "FailedScheduling", "Scheduling", msg)
	if err := updateDguest(ctx, sched.schdulerClient, dguest, &v1alpha1.DguestCondition{
		Type:    v1alpha1.DguestScheduled,
		Status:  v1alpha1.ConditionFalse,
		Reason:  reason,
		Message: errMsg,
	}, nominatingInfo); err != nil {
		klog.ErrorS(err, "Error updating dguest", "dguest", klog.KObj(dguest))
	}
}

// truncateMessage truncates a message if it hits the NoteLengthLimit.
func truncateMessage(message string) string {
	max := validation.NoteLengthLimit
	if len(message) <= max {
		return message
	}
	suffix := " ..."
	return message[:max-len(suffix)] + suffix
}

func updateDguest(ctx context.Context, client versioned.Interface, dguest *v1alpha1.Dguest, condition *v1alpha1.DguestCondition, nominatingInfo *framework.NominatingInfo) error {
	klog.V(3).InfoS("Updating dguest condition", "dguest", klog.KObj(dguest), "conditionType", condition.Type, "conditionStatus", condition.Status, "conditionReason", condition.Reason)
	dguestStatusCopy := dguest.Status.DeepCopy()
	// NominatedFoodName is updated only if we are trying to set it, and the value is
	// different from the existing one.
	nnnNeedsUpdate := nominatingInfo.Mode() == framework.ModeOverride
	if !dguestutil.UpdateDguestCondition(dguestStatusCopy, condition) && !nnnNeedsUpdate {
		return nil
	}
	if nnnNeedsUpdate {
		//dguestStatusCopy.NominatedFoodName = nominatingInfo.NominatedFoodName
	}
	return util.PatchDguestStatus(ctx, client, dguest, dguestStatusCopy)
}
