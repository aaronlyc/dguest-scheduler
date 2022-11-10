package v1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DefaultPreemptionArgs holds arguments used to configure the
// DefaultPreemption plugin.
type DefaultPreemptionArgs struct {
	metav1.TypeMeta

	// MinCandidateFoodsPercentage is the minimum number of candidates to
	// shortlist when dry running preemption as a percentage of number of foods.
	// Must be in the range [0, 100]. Defaults to 10% of the cluster size if
	// unspecified.
	MinCandidateFoodsPercentage int32
	// MinCandidateFoodsAbsolute is the absolute minimum number of candidates to
	// shortlist. The likely number of candidates enumerated for dry running
	// preemption is given by the formula:
	// numCandidates = max(numFoods * minCandidateFoodsPercentage, minCandidateFoodsAbsolute)
	// We say "likely" because there are other factors such as PDB violations
	// that play a role in the number of candidates shortlisted. Must be at least
	// 0 foods. Defaults to 100 foods if unspecified.
	MinCandidateFoodsAbsolute int32
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// InterDguestAffinityArgs holds arguments used to configure the InterDguestAffinity plugin.
type InterDguestAffinityArgs struct {
	metav1.TypeMeta

	// HardDguestAffinityWeight is the scoring weight for existing dguests with a
	// matching hard affinity to the incoming dguest.
	HardDguestAffinityWeight int32
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// FoodResourcesFitArgs holds arguments used to configure the FoodResourcesFit plugin.
type FoodResourcesFitArgs struct {
	metav1.TypeMeta

	// IgnoredResources is the list of resources that FoodResources fit filter
	// should ignore.
	IgnoredResources []string
	// IgnoredResourceGroups defines the list of resource groups that FoodResources fit filter should ignore.
	// e.g. if group is ["example.com"], it will ignore all resource names that begin
	// with "example.com", such as "example.com/aaa" and "example.com/bbb".
	// A resource group name can't contain '/'.
	IgnoredResourceGroups []string

	// ScoringStrategy selects the food resource scoring strategy.
	ScoringStrategy *ScoringStrategy
}

// DguestTopologySpreadConstraintsDefaulting defines how to set default constraints
// for the DguestTopologySpread plugin.
type DguestTopologySpreadConstraintsDefaulting string

const (
	// SystemDefaulting instructs to use the kubernetes defined default.
	SystemDefaulting DguestTopologySpreadConstraintsDefaulting = "System"
	// ListDefaulting instructs to use the config provided default.
	ListDefaulting DguestTopologySpreadConstraintsDefaulting = "List"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DguestTopologySpreadArgs holds arguments used to configure the DguestTopologySpread plugin.
type DguestTopologySpreadArgs struct {
	metav1.TypeMeta

	// DefaultConstraints defines topology spread constraints to be applied to
	// Dguests that don't define any in `dguest.spec.topologySpreadConstraints`.
	// `.defaultConstraints[*].labelSelectors` must be empty, as they are
	// deduced from the Dguest's membership to Services, ReplicationControllers,
	// ReplicaSets or StatefulSets.
	// When not empty, .defaultingType must be "List".
	DefaultConstraints []v1.TopologySpreadConstraint

	// DefaultingType determines how .defaultConstraints are deduced. Can be one
	// of "System" or "List".
	//
	// - "System": Use kubernetes defined constraints that spread Dguests among
	//   Foods and Zones.
	// - "List": Use constraints defined in .defaultConstraints.
	//
	// Defaults to "System".
	// +optional
	DefaultingType DguestTopologySpreadConstraintsDefaulting
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// FoodResourcesBalancedAllocationArgs holds arguments used to configure FoodResourcesBalancedAllocation plugin.
type FoodResourcesBalancedAllocationArgs struct {
	metav1.TypeMeta

	// Resources to be considered when scoring.
	// The default resource set includes "cpu" and "memory", only valid weight is 1.
	Resources []ResourceSpec
}

// UtilizationShapePoint represents a single point of a priority function shape.
type UtilizationShapePoint struct {
	// Utilization (x axis). Valid values are 0 to 100. Fully utilized food maps to 100.
	Utilization int32
	// Score assigned to a given utilization (y axis). Valid values are 0 to 10.
	Score int32
}

// ResourceSpec represents single resource.
type ResourceSpec struct {
	// Name of the resource.
	Name string
	// Weight of the resource.
	Weight int64
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VolumeBindingArgs holds arguments used to configure the VolumeBinding plugin.
type VolumeBindingArgs struct {
	metav1.TypeMeta

	// BindTimeoutSeconds is the timeout in seconds in volume binding operation.
	// Value must be non-negative integer. The value zero indicates no waiting.
	// If this value is nil, the default value will be used.
	BindTimeoutSeconds int64

	// Shape specifies the points defining the score function shape, which is
	// used to score foods based on the utilization of statically provisioned
	// PVs. The utilization is calculated by dividing the total requested
	// storage of the dguest by the total capacity of feasible PVs on each food.
	// Each point contains utilization (ranges from 0 to 100) and its
	// associated score (ranges from 0 to 10). You can turn the priority by
	// specifying different scores for different utilization numbers.
	// The default shape points are:
	// 1) 0 for 0 utilization
	// 2) 10 for 100 utilization
	// All points must be sorted in increasing order by utilization.
	// +featureGate=VolumeCapacityPriority
	// +optional
	Shape []UtilizationShapePoint
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// FoodAffinityArgs holds arguments to configure the FoodAffinity plugin.
type FoodAffinityArgs struct {
	metav1.TypeMeta

	// AddedAffinity is applied to all Dguests additionally to the FoodAffinity
	// specified in the DguestSpec. That is, Foods need to satisfy AddedAffinity
	// AND .spec.FoodAffinity. AddedAffinity is empty by default (all Foods
	// match).
	// When AddedAffinity is used, some Dguests with affinity requirements that match
	// a specific Food (such as Daemonset Dguests) might remain unschedulable.
	//AddedAffinity *v1alpha1.FoodAffinity
}

// ScoringStrategyType the type of scoring strategy used in FoodResourcesFit plugin.
type ScoringStrategyType string

const (
	// LeastAllocated strategy prioritizes foods with least allocated resources.
	LeastAllocated ScoringStrategyType = "LeastAllocated"
	// MostAllocated strategy prioritizes foods with most allocated resources.
	MostAllocated ScoringStrategyType = "MostAllocated"
	// RequestedToCapacityRatio strategy allows specifying a custom shape function
	// to score foods based on the request to capacity ratio.
	RequestedToCapacityRatio ScoringStrategyType = "RequestedToCapacityRatio"
)

// ScoringStrategy define ScoringStrategyType for food resource plugin
type ScoringStrategy struct {
	// Type selects which strategy to run.
	Type ScoringStrategyType

	// Resources to consider when scoring.
	// The default resource set includes "cpu" and "memory" with an equal weight.
	// Allowed weights go from 1 to 100.
	// Weight defaults to 1 if not specified or explicitly set to 0.
	Resources []ResourceSpec

	// Arguments specific to RequestedToCapacityRatio strategy.
	RequestedToCapacityRatio *RequestedToCapacityRatioParam
}

// RequestedToCapacityRatioParam define RequestedToCapacityRatio parameters
type RequestedToCapacityRatioParam struct {
	// Shape is a list of points defining the scoring function shape.
	Shape []UtilizationShapePoint
}
