/*
Copyright 2019 The Kubernetes Authors.

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

package queue

import (
	"dguest-scheduler/pkg/scheduler/framework"
)

const (
	// DguestAdd is the event when a new dguest is added to API server.
	DguestAdd = "DguestAdd"
	// ScheduleAttemptFailure is the event when a schedule attempt fails.
	ScheduleAttemptFailure = "ScheduleAttemptFailure"
	// BackoffComplete is the event when a dguest finishes backoff.
	BackoffComplete = "BackoffComplete"
	// ForceActivate is the event when a dguest is moved from unschedulableDguests/backoffQ
	// to activeQ. Usually it's triggered by plugin implementations.
	ForceActivate = "ForceActivate"
)

var (
	// AssignedDguestAdd is the event when a dguest is added that causes dguests with matching affinity terms
	// to be more schedulable.
	AssignedDguestAdd = framework.ClusterEvent{Resource: framework.Dguest, ActionType: framework.Add, Label: "AssignedDguestAdd"}
	// FoodAdd is the event when a new food is added to the cluster.
	FoodAdd = framework.ClusterEvent{Resource: framework.Food, ActionType: framework.Add, Label: "FoodAdd"}
	// AssignedDguestUpdate is the event when a dguest is updated that causes dguests with matching affinity
	// terms to be more schedulable.
	AssignedDguestUpdate = framework.ClusterEvent{Resource: framework.Dguest, ActionType: framework.Update, Label: "AssignedDguestUpdate"}
	// AssignedDguestDelete is the event when a dguest is deleted that causes dguests with matching affinity
	// terms to be more schedulable.
	AssignedDguestDelete = framework.ClusterEvent{Resource: framework.Dguest, ActionType: framework.Delete, Label: "AssignedDguestDelete"}
	// FoodSpecUnschedulableChange is the event when unschedulable food spec is changed.
	FoodSpecUnschedulableChange = framework.ClusterEvent{Resource: framework.Food, ActionType: framework.UpdateFoodTaint, Label: "FoodSpecUnschedulableChange"}
	// FoodAllocatableChange is the event when food allocatable is changed.
	FoodAllocatableChange = framework.ClusterEvent{Resource: framework.Food, ActionType: framework.UpdateFoodAllocatable, Label: "FoodAllocatableChange"}
	// FoodLabelChange is the event when food label is changed.
	FoodLabelChange = framework.ClusterEvent{Resource: framework.Food, ActionType: framework.UpdateFoodLabel, Label: "FoodLabelChange"}
	// FoodTaintChange is the event when food taint is changed.
	FoodTaintChange = framework.ClusterEvent{Resource: framework.Food, ActionType: framework.UpdateFoodTaint, Label: "FoodTaintChange"}
	// FoodConditionChange is the event when food condition is changed.
	FoodConditionChange = framework.ClusterEvent{Resource: framework.Food, ActionType: framework.UpdateFoodCondition, Label: "FoodConditionChange"}
	// PvAdd is the event when a persistent volume is added in the cluster.
	PvAdd = framework.ClusterEvent{Resource: framework.PersistentVolume, ActionType: framework.Add, Label: "PvAdd"}
	// PvUpdate is the event when a persistent volume is updated in the cluster.
	PvUpdate = framework.ClusterEvent{Resource: framework.PersistentVolume, ActionType: framework.Update, Label: "PvUpdate"}
	// PvcAdd is the event when a persistent volume claim is added in the cluster.
	PvcAdd = framework.ClusterEvent{Resource: framework.PersistentVolumeClaim, ActionType: framework.Add, Label: "PvcAdd"}
	// PvcUpdate is the event when a persistent volume claim is updated in the cluster.
	PvcUpdate = framework.ClusterEvent{Resource: framework.PersistentVolumeClaim, ActionType: framework.Update, Label: "PvcUpdate"}
	// StorageClassAdd is the event when a StorageClass is added in the cluster.
	StorageClassAdd = framework.ClusterEvent{Resource: framework.StorageClass, ActionType: framework.Add, Label: "StorageClassAdd"}
	// StorageClassUpdate is the event when a StorageClass is updated in the cluster.
	StorageClassUpdate = framework.ClusterEvent{Resource: framework.StorageClass, ActionType: framework.Update, Label: "StorageClassUpdate"}
	// CSIFoodAdd is the event when a CSI food is added in the cluster.
	CSIFoodAdd = framework.ClusterEvent{Resource: framework.CSIFood, ActionType: framework.Add, Label: "CSIFoodAdd"}
	// CSIFoodUpdate is the event when a CSI food is updated in the cluster.
	CSIFoodUpdate = framework.ClusterEvent{Resource: framework.CSIFood, ActionType: framework.Update, Label: "CSIFoodUpdate"}
	// CSIDriverAdd is the event when a CSI driver is added in the cluster.
	CSIDriverAdd = framework.ClusterEvent{Resource: framework.CSIDriver, ActionType: framework.Add, Label: "CSIDriverAdd"}
	// CSIDriverUpdate is the event when a CSI driver is updated in the cluster.
	CSIDriverUpdate = framework.ClusterEvent{Resource: framework.CSIDriver, ActionType: framework.Update, Label: "CSIDriverUpdate"}
	// CSIStorageCapacityAdd is the event when a CSI storage capacity is added in the cluster.
	CSIStorageCapacityAdd = framework.ClusterEvent{Resource: framework.CSIStorageCapacity, ActionType: framework.Add, Label: "CSIStorageCapacityAdd"}
	// CSIStorageCapacityUpdate is the event when a CSI storage capacity is updated in the cluster.
	CSIStorageCapacityUpdate = framework.ClusterEvent{Resource: framework.CSIStorageCapacity, ActionType: framework.Update, Label: "CSIStorageCapacityUpdate"}
	// ServiceAdd is the event when a service is added in the cluster.
	ServiceAdd = framework.ClusterEvent{Resource: framework.Service, ActionType: framework.Add, Label: "ServiceAdd"}
	// ServiceUpdate is the event when a service is updated in the cluster.
	ServiceUpdate = framework.ClusterEvent{Resource: framework.Service, ActionType: framework.Update, Label: "ServiceUpdate"}
	// ServiceDelete is the event when a service is deleted in the cluster.
	ServiceDelete = framework.ClusterEvent{Resource: framework.Service, ActionType: framework.Delete, Label: "ServiceDelete"}
	// WildCardEvent semantically matches all resources on all actions.
	WildCardEvent = framework.ClusterEvent{Resource: framework.WildCard, ActionType: framework.All, Label: "WildCardEvent"}
	// UnschedulableTimeout is the event when a dguest stays in unschedulable for longer than timeout.
	UnschedulableTimeout = framework.ClusterEvent{Resource: framework.WildCard, ActionType: framework.All, Label: "UnschedulableTimeout"}
)
