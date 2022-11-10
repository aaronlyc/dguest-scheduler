package util

import (
	"dguest-scheduler/pkg/apis/scheduler/v1alpha1"
)

// For each of these resources, a dguest that doesn't request the resource explicitly
// will be treated as having requested the amount indicated below, for the purpose
// of computing priority only. This ensures that when scheduling zero-request dguests, such
// dguests will not all be scheduled to the food with the smallest in-use request,
// and that when scheduling regular dguests, such dguests will not see zero-request dguests as
// consuming no resources whatsoever. We chose these values to be similar to the
// resources that we give to cluster addon dguests (#10653). But they are pretty arbitrary.
// As described in #11713, we use request instead of limit to deal with resource requirements.
const (
	// DefaultMilliCPURequest defines default milli cpu request number.
	DefaultMilliCPURequest int64 = 100 // 0.1 core
	// DefaultMemoryRequest defines default memory request size.
	DefaultMemoryRequest int64 = 200 * 1024 * 1024 // 200 MB
)

// GetNonzeroRequests returns the default cpu and memory resource request if none is found or
// what is provided on the request.
func GetNonzeroRequests(requests *v1alpha1.ResourceList) (int64, int64) {
	return GetRequestForResource(v1alpha1.ResourceCPU, requests, true),
		GetRequestForResource(v1alpha1.ResourceMemory, requests, true)
}

// GetRequestForResource returns the requested values unless nonZero is true and there is no defined request
// for CPU and memory.
// If nonZero is true and the resource has no defined request for CPU or memory, it returns a default value.
func GetRequestForResource(resource v1alpha1.ResourceName, requests *v1alpha1.ResourceList, nonZero bool) int64 {
	if requests == nil {
		return 0
	}
	switch resource {
	case v1alpha1.ResourceCPU:
		// Override if un-set, but not if explicitly set to zero
		if _, found := (*requests)[v1alpha1.ResourceCPU]; !found && nonZero {
			return DefaultMilliCPURequest
		}
		return requests.Cpu().MilliValue()
	case v1alpha1.ResourceMemory:
		// Override if un-set, but not if explicitly set to zero
		if _, found := (*requests)[v1alpha1.ResourceMemory]; !found && nonZero {
			return DefaultMemoryRequest
		}
		return requests.Memory().Value()
	case v1alpha1.ResourceBandwidth:
		quantity, found := (*requests)[v1alpha1.ResourceBandwidth]
		if !found {
			return 0
		}
		return quantity.Value()
	default:
		quantity, found := (*requests)[resource]
		if !found {
			return 0
		}
		return quantity.Value()
	}
}
