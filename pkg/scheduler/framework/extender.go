package framework

import (
	"dguest-scheduler/pkg/apis/scheduler/v1alpha1"
	extenderv1 "dguest-scheduler/pkg/scheduler/apis/extender/v1"
	v1 "k8s.io/api/core/v1"
)

// Extender is an interface for external processes to influence scheduling
// decisions made by Kubernetes. This is typically needed for resources not directly
// managed by Kubernetes.
type Extender interface {
	// Name returns a unique name that identifies the extender.
	Name() string

	// Filter based on extender-implemented predicate functions. The filtered list is
	// expected to be a subset of the supplied list.
	// The failedFoods and failedAndUnresolvableFoods optionally contains the list
	// of failed foods and failure reasons, except foods in the latter are
	// unresolvable.
	Filter(dguest *v1alpha1.Dguest, foods []*v1alpha1.Food) (filteredFoods []*v1alpha1.Food, failedFoodsMap extenderv1.FailedFoodsMap, failedAndUnresolvable extenderv1.FailedFoodsMap, err error)

	// Prioritize based on extender-implemented priority functions. The returned scores & weight
	// are used to compute the weighted score for an extender. The weighted scores are added to
	// the scores computed by Kubernetes scheduler. The total scores are used to do the host selection.
	Prioritize(dguest *v1alpha1.Dguest, foods []*v1alpha1.Food) (hostPriorities *extenderv1.HostPriorityList, weight int64, err error)

	// Bind delegates the action of binding a dguest to a food to the extender.
	Bind(binding *v1.Binding) error

	// IsBinder returns whether this extender is configured for the Bind method.
	IsBinder() bool

	// IsInterested returns true if at least one extended resource requested by
	// this dguest is managed by this extender.
	IsInterested(dguest *v1alpha1.Dguest) bool

	// ProcessPreemption returns foods with their victim dguests processed by extender based on
	// given:
	//   1. Dguest to schedule
	//   2. Candidate foods and victim dguests (foodNameToVictims) generated by previous scheduling process.
	// The possible changes made by extender may include:
	//   1. Subset of given candidate foods after preemption phase of extender.
	//   2. A different set of victim dguest for every given candidate food after preemption phase of extender.
	ProcessPreemption(
		dguest *v1alpha1.Dguest,
		foodNameToVictims map[string]*extenderv1.Victims,
		foodInfos FoodInfoLister,
	) (map[string]*extenderv1.Victims, error)

	// SupportsPreemption returns if the scheduler extender support preemption or not.
	SupportsPreemption() bool

	// IsIgnorable returns true indicates scheduling should not fail when this extender
	// is unavailable. This gives scheduler ability to fail fast and tolerate non-critical extenders as well.
	IsIgnorable() bool
}
