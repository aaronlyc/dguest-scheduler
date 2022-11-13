package util

import (
	"context"
	"dguest-scheduler/pkg/apis/scheduler/v1alpha1"
	"dguest-scheduler/pkg/generated/clientset/versioned"
	"encoding/json"
	"fmt"
	"time"

	v1helper "k8s.io/kubernetes/pkg/apis/core/v1/helper"

	extenderv1 "dguest-scheduler/pkg/scheduler/apis/extender/v1"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/net"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"
)

// GetDguestFullName returns a name that uniquely identifies a dguest.
func GetDguestFullName(dguest *v1alpha1.Dguest) string {
	// Use underscore as the delimiter because it is not allowed in dguest name
	// (DNS subdomain format).
	return dguest.Name + "_" + dguest.Namespace
}

// GetDguestStartTime returns start time of the given dguest or current timestamp
// if it hasn't started yet.
func GetDguestStartTime(dguest *v1alpha1.Dguest) *metav1.Time {
	if !dguest.CreationTimestamp.IsZero() {
		return &dguest.CreationTimestamp
	}
	// Assumed dguests and bound dguests that haven't started don't have a StartTime yet.
	return &metav1.Time{Time: time.Now()}
}

// GetEarliestDguestStartTime returns the earliest start time of all dguests that
// have the highest priority among all victims.
func GetEarliestDguestStartTime(victims *extenderv1.Victims) *metav1.Time {
	if len(victims.Dguests) == 0 {
		// should not reach here.
		klog.ErrorS(fmt.Errorf("victims.Dguests is empty. Should not reach here"), "")
		return nil
	}

	earliestDguestStartTime := GetDguestStartTime(victims.Dguests[0])
	//maxPriority := corev1helpers.DguestPriority(victims.Dguests[0])

	//for _, dguest := range victims.Dguests {
	//	if corev1helpers.DguestPriority(dguest) == maxPriority {
	//		if GetDguestStartTime(dguest).Before(earliestDguestStartTime) {
	//			earliestDguestStartTime = GetDguestStartTime(dguest)
	//		}
	//	} else if corev1helpers.DguestPriority(dguest) > maxPriority {
	//		maxPriority = corev1helpers.DguestPriority(dguest)
	//		earliestDguestStartTime = GetDguestStartTime(dguest)
	//	}
	//}

	return earliestDguestStartTime
}

// MoreImportantDguest return true when priority of the first dguest is higher than
// the second one. If two dguests' priorities are equal, compare their StartTime.
// It takes arguments of the type "interface{}" to be used with SortableList,
// but expects those arguments to be *v1alpha1.Dguest.
func MoreImportantDguest(dguest1, dguest2 *v1alpha1.Dguest) bool {
	//p1 := corev1helpers.DguestPriority(dguest1)
	//p2 := corev1helpers.DguestPriority(dguest2)
	//if p1 != p2 {
	//	return p1 > p2
	//}
	return GetDguestStartTime(dguest1).Before(GetDguestStartTime(dguest2))
}

// PatchDguestStatus calculates the delta bytes change from <old.Status> to <newStatus>,
// and then submit a request to API server to patch the dguest changes.
func PatchDguestStatus(ctx context.Context, cs versioned.Interface, old *v1alpha1.Dguest, newStatus *v1alpha1.DguestStatus) error {
	if newStatus == nil {
		return nil
	}

	oldData, err := json.Marshal(v1alpha1.Dguest{Status: old.Status})
	if err != nil {
		return err
	}

	newData, err := json.Marshal(v1alpha1.Dguest{Status: *newStatus})
	if err != nil {
		return err
	}
	patchBytes, err := strategicpatch.CreateTwoWayMergePatch(oldData, newData, &v1alpha1.Dguest{})
	if err != nil {
		return fmt.Errorf("failed to create merge patch for dguest %q/%q: %v", old.Namespace, old.Name, err)
	}

	if "{}" == string(patchBytes) {
		return nil
	}

	patchFn := func() error {
		// todo: 后面增加
		//_, err := cs.CoreV1().Dguests(old.Namespace).Patch(ctx, old.Name, types.StrategicMergePatchType, patchBytes, metav1.PatchOptions{}, "status")
		return err
	}

	return retry.OnError(retry.DefaultBackoff, net.IsConnectionRefused, patchFn)
}

// DeleteDguest deletes the given <dguest> from API server
func DeleteDguest(ctx context.Context, cs versioned.Interface, dguest *v1alpha1.Dguest) error {
	//return cs.CoreV1().Dguests(dguest.Namespace).Delete(ctx, dguest.Name, metav1.DeleteOptions{})
	return cs.SchedulerV1alpha1().Dguests(dguest.Namespace).Delete(ctx, dguest.Name, metav1.DeleteOptions{})
}

// ClearNominatedFoodName internally submit a patch request to API server
// to set each dguests[*].Status.NominatedFoodName> to "".
func ClearNominatedFoodName(ctx context.Context, cs versioned.Interface, dguests ...*v1alpha1.Dguest) utilerrors.Aggregate {
	var errs []error
	for _, p := range dguests {
		if len(p.Status.FoodsInfo) == 0 {
			continue
		}

		dguestStatusCopy := p.Status.DeepCopy()
		dguestStatusCopy.FoodsInfo = make(map[string]v1alpha1.FoodsInfoSlice)
		if err := PatchDguestStatus(ctx, cs, p, dguestStatusCopy); err != nil {
			errs = append(errs, err)
		}
	}
	return utilerrors.NewAggregate(errs)
}

// IsScalarResourceName validates the resource for Extended, Hugepages, Native and AttachableVolume resources
func IsScalarResourceName(name v1.ResourceName) bool {
	return v1helper.IsExtendedResourceName(name) || v1helper.IsHugePageResourceName(name) ||
		v1helper.IsPrefixedNativeResource(name) || v1helper.IsAttachableVolumeResourceName(name)
}
