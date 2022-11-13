package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +enum
type FoodPhase string

// These are the valid phases of food.
const (
	// FoodPending means the food has been created/added by the system, but not configured.
	FoodPending FoodPhase = "Pending"
	// FoodRunning means the food has been configured and has Kubernetes components running.
	FoodRunning FoodPhase = "Running"
	// FoodTerminated means the food has been removed from the cluster.
	FoodTerminated FoodPhase = "Terminated"
)

type FoodConditionType string

// These are valid but not exhaustive conditions of food. A cloud provider may set a condition not listed here.
// The built-in set of conditions are:
// FoodReachable, FoodLive, FoodReady, FoodSchedulable, FoodRunnable.
const (
	// FoodReady means kubelet is healthy and ready to accept dguests.
	FoodReady FoodConditionType = "Ready"
	// FoodMemoryPressure means the kubelet is under pressure due to insufficient available memory.
	FoodMemoryPressure FoodConditionType = "MemoryPressure"
	// FoodDiskPressure means the kubelet is under pressure due to insufficient available disk.
	FoodDiskPressure FoodConditionType = "DiskPressure"
	// FoodBandwidthPressure means the kubelet is under pressure due to insufficient available PID.
	FoodBandwidthPressure FoodConditionType = "BandwidthPressure"
	// FoodNetworkUnavailable means that network for the food is not correctly configured.
	FoodNetworkUnavailable FoodConditionType = "NetworkUnavailable"
	FoodTIDPressure        FoodConditionType = "TIDPressure"
)

// FoodCondition contains condition information for a food.
type FoodCondition struct {
	// Type of food condition.
	Type FoodConditionType `json:"type,omitempty"`
	// Status of the condition, one of True, False, Unknown.
	Status ConditionStatus `json:"status,omitempty"`
	// Last time we got an update on a given condition.
	// +optional
	LastHeartbeatTime metav1.Time `json:"lastHeartbeatTime,omitempty"`
	// Last time the condition transit from one status to another.
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
	// (brief) reason for the condition's last transition.
	// +optional
	Reason string `json:"reason,omitempty"`
	// Human readable message indicating details about last transition.
	// +optional
	Message string `json:"message,omitempty"`
}

type FoodInfo struct {
	Version     string             `json:"version,omitempty"`
	ReleaseTime metav1.Time        `json:"releaseTime,omitempty"`
	CoreRunNode string             `json:"coreRunNode,omitempty"`
	Items       []IngredientStatus `json:"items,omitempty"`
}

type IngredientStatus struct {
	Namespace  string          `json:"namespace,omitempty"`
	Name       string          `json:"name,omitempty"`
	CreateTime metav1.Time     `json:"createTime,omitempty"`
	Phase      IngredientPhase `json:"phase,omitempty"`
}

type IngredientPhase string

const (
	IngredientGood IngredientPhase = "Good"
	IngredientBad  IngredientPhase = "Bad"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Food is a specification for a Food resource
type Food struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FoodSpec   `json:"spec"`
	Status FoodStatus `json:"status,omitempty"`
}

// FoodSpec is the spec for a Food resource
type FoodSpec struct {
	Unschedulable bool           `json:"unschedulable,omitempty"`
	Taints        []corev1.Taint `json:"taints,omitempty"`
}

// FoodStatus is the status for a Food resource
type FoodStatus struct {
	// Capacity represents the total resources of a food.
	Capacity ResourceList `json:"capacity,omitempty"`
	// Allocatable represents the resources of a food that are available for scheduling.
	Allocatable ResourceList    `json:"allocatable,omitempty"`
	Conditions  []FoodCondition `json:"conditions,omitempty"`
	FoodInfo    *FoodInfo       `json:"foodInfo,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// FoodList is a list of Food resources
type FoodList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Food `json:"items,omitempty"`
}
