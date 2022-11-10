package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ReSchedulerPolicy describes how the container should be restarted.
// Only one of the following restart policies may be specified.
// If none of the following policies is specified, the default one
// is RestartPolicyAlways.
// +enum
type ReSchedulerPolicy string

const (
	ReSchedulerPolicyAlways    ReSchedulerPolicy = "Always"
	ReSchedulerPolicyOnFailure ReSchedulerPolicy = "OnFailure"
	ReSchedulerPolicyNever     ReSchedulerPolicy = "Never"
)

// DguestFoodConditionType is a valid value for DguestFoodConditionType.Type
type DguestFoodConditionType string

// These are built-in conditions of dguest. An application may use a custom condition not listed here.
const (
	// DguestScheduled represents status of the scheduling process for this dguest.
	DguestScheduled DguestFoodConditionType = "Scheduled"
	// DguestReady means the dguest is able to service requests and should be added to the
	// load balancing pools of all matching services.
	DguestReady DguestFoodConditionType = "Ready"
)

// These are reasons for a dguest's transition to a condition.
const (
	// DguestReasonUnschedulable reason in DguestReasonUnschedulable means that the scheduler
	// can't schedule the dguest right now, for example due to insufficient resources in the cluster.
	DguestReasonUnschedulable = "Unschedulable"
)

type ConditionStatus string

// These are valid condition statuses. "ConditionTrue" means a resource is in the condition.
// "ConditionFalse" means a resource is not in the condition. "ConditionUnknown" means kubernetes
// can't decide if a resource is in the condition or not. In the future, we could add other
// intermediate conditions, e.g. ConditionDegraded.
const (
	ConditionTrue    ConditionStatus = "True"
	ConditionFalse   ConditionStatus = "False"
	ConditionUnknown ConditionStatus = "Unknown"
)

// DguestFoodCondition contains details for the current condition of this dguest.
type DguestFoodCondition struct {
	// Type is the type of the condition.
	Type DguestFoodConditionType `json:"type,omitempty"`
	// Status is the status of the condition.
	// Can be True, False, Unknown.
	Status ConditionStatus `json:"status,omitempty"`
	// Last time we probed the condition.
	// +optional
	LastProbeTime metav1.Time `json:"lastProbeTime,omitempty"	`
	// Last time the condition transitioned from one status to another.
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
	// Unique, one-word, CamelCase reason for the condition's last transition.
	// +optional
	Reason string `json:"reason,omitempty"`
	// Human-readable message indicating details about last transition.
	// +optional
	Message string `json:"message,omitempty"`
}

type DguestFoodInfo struct {
	Namespace      string              `json:"namespace,omitempty"`
	Name           string              `json:"name,omitempty"`
	SchedulerdTime metav1.Time         `json:"schedulerdTime,omitempty"`
	Condition      DguestFoodCondition `json:"condition,omitempty"`
}

// DguestPhase is a label for the condition of a dguest at the current time.
// +enum
type DguestPhase string

// These are the valid statuses of dguests.
const (
	// DguestPending means the dguest has been accepted by the system, but one or more of the containers
	// has not been started. This includes time before being bound to a food, as well as time spent
	// pulling images onto the host.
	DguestPending DguestPhase = "Pending"
	// DguestRunning means the dguest has been bound to a food and all of the containers have been started.
	// At least one container is still running or is in the process of being restarted.
	DguestRunning DguestPhase = "Running"
	// DguestSucceeded means that all containers in the dguest have voluntarily terminated
	// with a container exit code of 0, and the system is not going to restart any of these containers.
	DguestSucceeded DguestPhase = "Succeeded"
	// DguestFailed means that all containers in the dguest have terminated, and at least one container has
	// terminated in a failure (exited with a non-zero exit code or was stopped by the system).
	DguestFailed DguestPhase = "Failed"
	// DguestUnknown means that for some reason the state of the dguest could not be obtained, typically due
	// to an error in communicating with the host of the dguest.
	// Deprecated: It isn't being set since 2015 (74da3b14b0c0f658b3bb8d2def5094686d0e9095)
	DguestUnknown DguestPhase = "Unknown"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Dguest is a specification for a Food resource
type Dguest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DguestSpec   `json:"spec"`
	Status DguestStatus `json:"status"`
}

// DguestSpec is the spec for a Food resource
type DguestSpec struct {
	// Rescheduling policy, including the rescheduling policy when the replica on the tenant is abnormal
	ReSchedulerPolicy ReSchedulerPolicy `json:"reSchedulerPolicy,omitempty"`
	// The timeout period before becoming active
	ActiveDeadlineSeconds *int64 `json:"activeDeadlineSeconds,omitempty"`
	// The Bill the dguest want to scheduler.
	WantBill *string
	// Number of replicas
	Replicas *int32 `json:"replicas,omitempty"`
	// Name of the scheduler
	SchedulerName string `json:"schedulerName,omitempty"`
	// Tolerance: indicates that it can be scheduled to foods with tolerance
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`
	// Overhead such as bandwidth /cpu/ memory etc
	Overhead ResourceList `json:"overhead,omitempty"`
}

// DguestStatus is the status for a Food resource
type DguestStatus struct {
	FoodsInfo []DguestFoodInfo `json:"foodsInfo,omitempty"`
	Phase     DguestPhase      `json:"phase,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DguestList is a list of dguest resources
type DguestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Dguest `json:"items"`
}
