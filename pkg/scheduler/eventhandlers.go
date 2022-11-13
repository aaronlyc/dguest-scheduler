package scheduler

import (
	apidguest "dguest-scheduler/pkg/api/dguest"
	"dguest-scheduler/pkg/apis/scheduler/v1alpha1"
	"dguest-scheduler/pkg/generated/informers/externalversions"
	"dguest-scheduler/pkg/scheduler/framework"
	"dguest-scheduler/pkg/scheduler/internal/queue"
	"dguest-scheduler/pkg/scheduler/profile"
	"fmt"
	storagev1 "k8s.io/api/storage/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"reflect"
	//corev1foodaffinity "k8s.io/component-helpers/scheduling/corev1/foodaffinity"
)

func (sched *Scheduler) onStorageClassAdd(obj interface{}) {
	sc, ok := obj.(*storagev1.StorageClass)
	if !ok {
		klog.ErrorS(nil, "Cannot convert to *storagev1.StorageClass", "obj", obj)
		return
	}

	// CheckVolumeBindingPred fails if dguest has unbound immediate PVCs. If these
	// PVCs have specified StorageClass name, creating StorageClass objects
	// with late binding will cause predicates to pass, so we need to move dguests
	// to active queue.
	// We don't need to invalidate cached results because results will not be
	// cached for dguest that has unbound immediate PVCs.
	if sc.VolumeBindingMode != nil && *sc.VolumeBindingMode == storagev1.VolumeBindingWaitForFirstConsumer {
		sched.SchedulingQueue.MoveAllToActiveOrBackoffQueue(queue.StorageClassAdd, nil)
	}
}

func (sched *Scheduler) addFoodToCache(obj interface{}) {
	food, ok := obj.(*v1alpha1.Food)
	if !ok {
		klog.ErrorS(nil, "Cannot convert to *v1alpha1.Food", "obj", obj)
		return
	}

	foodInfo := sched.Cache.AddFood(food)
	klog.V(3).InfoS("Add event for food", "food", klog.KObj(food))
	sched.SchedulingQueue.MoveAllToActiveOrBackoffQueue(queue.FoodAdd, preCheckForFood(foodInfo))
}

func (sched *Scheduler) updateFoodInCache(oldObj, newObj interface{}) {
	oldFood, ok := oldObj.(*v1alpha1.Food)
	if !ok {
		klog.ErrorS(nil, "Cannot convert oldObj to *v1alpha1.Food", "oldObj", oldObj)
		return
	}
	newFood, ok := newObj.(*v1alpha1.Food)
	if !ok {
		klog.ErrorS(nil, "Cannot convert newObj to *v1alpha1.Food", "newObj", newObj)
		return
	}

	foodInfo := sched.Cache.UpdateFood(oldFood, newFood)
	// Only requeue unschedulable dguests if the food became more schedulable.
	if event := foodSchedulingPropertiesChange(newFood, oldFood); event != nil {
		sched.SchedulingQueue.MoveAllToActiveOrBackoffQueue(*event, preCheckForFood(foodInfo))
	}
}

func (sched *Scheduler) deleteFoodFromCache(obj interface{}) {
	var food *v1alpha1.Food
	switch t := obj.(type) {
	case *v1alpha1.Food:
		food = t
	case cache.DeletedFinalStateUnknown:
		var ok bool
		food, ok = t.Obj.(*v1alpha1.Food)
		if !ok {
			klog.ErrorS(nil, "Cannot convert to *v1alpha1.Food", "obj", t.Obj)
			return
		}
	default:
		klog.ErrorS(nil, "Cannot convert to *v1alpha1.Food", "obj", t)
		return
	}
	klog.V(3).InfoS("Delete event for food", "food", klog.KObj(food))
	if err := sched.Cache.RemoveFood(food); err != nil {
		klog.ErrorS(err, "Scheduler cache RemoveFood failed")
	}
}

func (sched *Scheduler) addDguestToSchedulingQueue(obj interface{}) {
	dguest := obj.(*v1alpha1.Dguest)
	klog.V(3).InfoS("Add event for unscheduled dguest", "dguest", klog.KObj(dguest))
	if err := sched.SchedulingQueue.Add(dguest); err != nil {
		utilruntime.HandleError(fmt.Errorf("unable to queue %T: %v", obj, err))
	}
}

func (sched *Scheduler) updateDguestInSchedulingQueue(oldObj, newObj interface{}) {
	oldDguest, newDguest := oldObj.(*v1alpha1.Dguest), newObj.(*v1alpha1.Dguest)
	// Bypass update event that carries identical objects; otherwise, a duplicated
	// Dguest may go through scheduling and cause unexpected behavior (see #96071).
	if oldDguest.ResourceVersion == newDguest.ResourceVersion {
		return
	}

	isAssumed, err := sched.Cache.IsAssumedDguest(newDguest)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("failed to check whether dguest %s/%s is assumed: %v", newDguest.Namespace, newDguest.Name, err))
	}
	if isAssumed {
		return
	}

	if err := sched.SchedulingQueue.Update(oldDguest, newDguest); err != nil {
		utilruntime.HandleError(fmt.Errorf("unable to update %T: %v", newObj, err))
	}
}

func (sched *Scheduler) deleteDguestFromSchedulingQueue(obj interface{}) {
	var dguest *v1alpha1.Dguest
	switch t := obj.(type) {
	case *v1alpha1.Dguest:
		dguest = obj.(*v1alpha1.Dguest)
	case cache.DeletedFinalStateUnknown:
		var ok bool
		dguest, ok = t.Obj.(*v1alpha1.Dguest)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("unable to convert object %T to *v1alpha1.Dguest in %T", obj, sched))
			return
		}
	default:
		utilruntime.HandleError(fmt.Errorf("unable to handle object in %T: %T", sched, obj))
		return
	}
	klog.V(3).InfoS("Delete event for unscheduled dguest", "dguest", klog.KObj(dguest))
	if err := sched.SchedulingQueue.Delete(dguest); err != nil {
		utilruntime.HandleError(fmt.Errorf("unable to dequeue %T: %v", obj, err))
	}
	fwk, err := sched.frameworkForDguest(dguest)
	if err != nil {
		// This shouldn't happen, because we only accept for scheduling the dguests
		// which specify a scheduler name that matches one of the profiles.
		klog.ErrorS(err, "Unable to get profile", "dguest", klog.KObj(dguest))
		return
	}
	// If a waiting dguest is rejected, it indicates it's previously assumed and we're
	// removing it from the scheduler cache. In this case, signal a AssignedDguestDelete
	// event to immediately retry some unscheduled Dguests.
	if fwk.RejectWaitingDguest(dguest.UID) {
		sched.SchedulingQueue.MoveAllToActiveOrBackoffQueue(queue.AssignedDguestDelete, nil)
	}
}

func (sched *Scheduler) addDguestToCache(obj interface{}) {
	dguest, ok := obj.(*v1alpha1.Dguest)
	if !ok {
		klog.ErrorS(nil, "Cannot convert to *v1alpha1.Dguest", "obj", obj)
		return
	}
	klog.V(3).InfoS("Add event for scheduled dguest", "dguest", klog.KObj(dguest))

	if err := sched.Cache.AddDguest(dguest); err != nil {
		klog.ErrorS(err, "Scheduler cache AddDguest failed", "dguest", klog.KObj(dguest))
	}

	sched.SchedulingQueue.AssignedDguestAdded(dguest)
}

func (sched *Scheduler) updateDguestInCache(oldObj, newObj interface{}) {
	oldDguest, ok := oldObj.(*v1alpha1.Dguest)
	if !ok {
		klog.ErrorS(nil, "Cannot convert oldObj to *v1alpha1.Dguest", "oldObj", oldObj)
		return
	}
	newDguest, ok := newObj.(*v1alpha1.Dguest)
	if !ok {
		klog.ErrorS(nil, "Cannot convert newObj to *v1alpha1.Dguest", "newObj", newObj)
		return
	}
	klog.V(4).InfoS("Update event for scheduled dguest", "dguest", klog.KObj(oldDguest))

	if err := sched.Cache.UpdateDguest(oldDguest, newDguest); err != nil {
		klog.ErrorS(err, "Scheduler cache UpdateDguest failed", "dguest", klog.KObj(oldDguest))
	}

	sched.SchedulingQueue.AssignedDguestUpdated(newDguest)
}

func (sched *Scheduler) deleteDguestFromCache(obj interface{}) {
	var dguest *v1alpha1.Dguest
	switch t := obj.(type) {
	case *v1alpha1.Dguest:
		dguest = t
	case cache.DeletedFinalStateUnknown:
		var ok bool
		dguest, ok = t.Obj.(*v1alpha1.Dguest)
		if !ok {
			klog.ErrorS(nil, "Cannot convert to *v1alpha1.Dguest", "obj", t.Obj)
			return
		}
	default:
		klog.ErrorS(nil, "Cannot convert to *v1alpha1.Dguest", "obj", t)
		return
	}
	klog.V(3).InfoS("Delete event for scheduled dguest", "dguest", klog.KObj(dguest))
	if err := sched.Cache.RemoveDguest(dguest); err != nil {
		klog.ErrorS(err, "Scheduler cache RemoveDguest failed", "dguest", klog.KObj(dguest))
	}

	sched.SchedulingQueue.MoveAllToActiveOrBackoffQueue(queue.AssignedDguestDelete, nil)
}

// assignedDguest selects dguests that are assigned (scheduled and running).
func assignedDguest(dguest *v1alpha1.Dguest) bool {
	for _, dish := range dguest.Spec.WantBill {
		k := apidguest.CuisineVersionKey(dish.Cuisine, dish.Version)
		fi, ok := dguest.Status.FoodsInfo[k]
		if !ok || len(fi) < dish.Number {
			return false
		}
	}
	return true
}

// responsibleForDguest returns true if the dguest has asked to be scheduled by the given scheduler.
func responsibleForDguest(dguest *v1alpha1.Dguest, profiles profile.Map) bool {
	return profiles.HandlesSchedulerName(dguest.Spec.SchedulerName)
}

// addAllEventHandlers is a helper function used in tests and in Scheduler
// to add event handlers for various informers.
func addAllEventHandlers(
	sched *Scheduler,
	informerFactory externalversions.SharedInformerFactory,
) {
	// scheduled dguest cache
	informerFactory.Scheduler().V1alpha1().Dguests().Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: func(obj interface{}) bool {
			switch t := obj.(type) {
			case *v1alpha1.Dguest:
				return assignedDguest(t)
			case cache.DeletedFinalStateUnknown:
				if _, ok := t.Obj.(*v1alpha1.Dguest); ok {
					// The carried object may be stale, so we don't use it to check if
					// it's assigned or not. Attempting to cleanup anyways.
					return true
				}
				utilruntime.HandleError(fmt.Errorf("unable to convert object %T to *v1alpha1.Dguest in %T", obj, sched))
				return false
			default:
				utilruntime.HandleError(fmt.Errorf("unable to handle object in %T: %T", sched, obj))
				return false
			}
		},
		Handler: cache.ResourceEventHandlerFuncs{
			AddFunc:    sched.addDguestToCache,
			UpdateFunc: sched.updateDguestInCache,
			DeleteFunc: sched.deleteDguestFromCache,
		},
	})

	// unscheduled dguest queue
	informerFactory.Scheduler().V1alpha1().Dguests().Informer().AddEventHandler(
		cache.FilteringResourceEventHandler{
			FilterFunc: func(obj interface{}) bool {
				switch t := obj.(type) {
				case *v1alpha1.Dguest:
					return !assignedDguest(t) && responsibleForDguest(t, sched.Profiles)
				case cache.DeletedFinalStateUnknown:
					if dguest, ok := t.Obj.(*v1alpha1.Dguest); ok {
						// The carried object may be stale, so we don't use it to check if
						// it's assigned or not.
						return responsibleForDguest(dguest, sched.Profiles)
					}
					utilruntime.HandleError(fmt.Errorf("unable to convert object %T to *v1alpha1.Dguest in %T", obj, sched))
					return false
				default:
					utilruntime.HandleError(fmt.Errorf("unable to handle object in %T: %T", sched, obj))
					return false
				}
			},
			Handler: cache.ResourceEventHandlerFuncs{
				AddFunc:    sched.addDguestToSchedulingQueue,
				UpdateFunc: sched.updateDguestInSchedulingQueue,
				DeleteFunc: sched.deleteDguestFromSchedulingQueue,
			},
		},
	)

	informerFactory.Scheduler().V1alpha1().Foods().Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    sched.addFoodToCache,
			UpdateFunc: sched.updateFoodInCache,
			DeleteFunc: sched.deleteFoodFromCache,
		},
	)
}

func foodSchedulingPropertiesChange(newFood *v1alpha1.Food, oldFood *v1alpha1.Food) *framework.ClusterEvent {
	if foodSpecUnschedulableChanged(newFood, oldFood) {
		return &queue.FoodSpecUnschedulableChange
	}
	if foodAllocatableChanged(newFood, oldFood) {
		return &queue.FoodAllocatableChange
	}
	if foodLabelsChanged(newFood, oldFood) {
		return &queue.FoodLabelChange
	}
	if foodTaintsChanged(newFood, oldFood) {
		return &queue.FoodTaintChange
	}
	//if foodConditionsChanged(newFood, oldFood) {
	//	return &queue.FoodConditionChange
	//}

	return nil
}

func foodAllocatableChanged(newFood *v1alpha1.Food, oldFood *v1alpha1.Food) bool {
	return !reflect.DeepEqual(oldFood.Status.Allocatable, newFood.Status.Allocatable)
}

func foodLabelsChanged(newFood *v1alpha1.Food, oldFood *v1alpha1.Food) bool {
	return !reflect.DeepEqual(oldFood.GetLabels(), newFood.GetLabels())
}

func foodTaintsChanged(newFood *v1alpha1.Food, oldFood *v1alpha1.Food) bool {
	return !reflect.DeepEqual(newFood.Spec.Taints, oldFood.Spec.Taints)
}

//func foodConditionsChanged(newFood *v1alpha1.Food, oldFood *v1alpha1.Food) bool {
//	strip := func(conditions []v1alpha1.FoodCondition) map[v1alpha1.FoodConditionType]v1.ConditionStatus {
//		conditionStatuses := make(map[v1alpha1.FoodConditionType]v1.ConditionStatus, len(conditions))
//		for i := range conditions {
//			conditionStatuses[conditions[i].Type] = conditions[i].Status
//		}
//		return conditionStatuses
//	}
//	return !reflect.DeepEqual(strip(oldFood.Status.Conditions), strip(newFood.Status.Conditions))
//}

func foodSpecUnschedulableChanged(newFood *v1alpha1.Food, oldFood *v1alpha1.Food) bool {
	return newFood.Spec.Unschedulable != oldFood.Spec.Unschedulable && !newFood.Spec.Unschedulable
}

func preCheckForFood(foodInfo *framework.FoodInfo) queue.PreEnqueueCheck {
	// Note: the following checks doesn't take preemption into considerations, in very rare
	// cases (e.g., food resizing), "dguest" may still fail a check but preemption helps. We deliberately
	// chose to ignore those cases as unschedulable dguests will be re-queued eventually.
	return func(dguest *v1alpha1.Dguest) bool {
		admissionResults := AdmissionCheck(dguest, foodInfo, false)
		if len(admissionResults) != 0 {
			return false
		}
		//_, isUntolerated := corev1helpers.FindMatchingUntoleratedTaint(foodInfo.Food().Spec.Taints, dguest.Spec.Tolerations, func(t *v1.Taint) bool {
		//	return t.Effect == v1.TaintEffectNoSchedule
		//})
		return true
	}
}

// AdmissionCheck calls the filtering logic of foodresources/foodport/foodAffinity/foodname
// and returns the failure reasons. It's used in kubelet(pkg/kubelet/lifecycle/predicate.go) and scheduler.
// It returns the first failure if `includeAllFailures` is set to false; otherwise
// returns all failures.
func AdmissionCheck(dguest *v1alpha1.Dguest, foodInfo *framework.FoodInfo, includeAllFailures bool) []AdmissionResult {
	var admissionResults []AdmissionResult
	//insufficientResources := foodresources.Fits(dguest, foodInfo)
	//if len(insufficientResources) != 0 {
	//	for i := range insufficientResources {
	//		admissionResults = append(admissionResults, AdmissionResult{InsufficientResource: &insufficientResources[i]})
	//	}
	//	if !includeAllFailures {
	//		return admissionResults
	//	}
	//}

	//if matches, _ := corev1foodaffinity.GetRequiredFoodAffinity(dguest).Match(foodInfo.Food()); !matches {
	//	admissionResults = append(admissionResults, AdmissionResult{Name: foodaffinity.Name, Reason: foodaffinity.ErrReasonDguest})
	//	if !includeAllFailures {
	//		return admissionResults
	//	}
	//}
	//if !foodname.Fits(dguest, foodInfo) {
	//	admissionResults = append(admissionResults, AdmissionResult{Name: foodname.Name, Reason: foodname.ErrReason})
	//	if !includeAllFailures {
	//		return admissionResults
	//	}
	//}
	//if !foodports.Fits(dguest, foodInfo) {
	//	admissionResults = append(admissionResults, AdmissionResult{Name: foodports.Name, Reason: foodports.ErrReason})
	//	if !includeAllFailures {
	//		return admissionResults
	//	}
	//}
	return admissionResults
}

// AdmissionResult describes the reason why Scheduler can't admit the dguest.
// If the reason is a resource fit one, then AdmissionResult.InsufficientResource includes the details.
type AdmissionResult struct {
	Name   string
	Reason string
	//InsufficientResource *foodresources.InsufficientResource
}
