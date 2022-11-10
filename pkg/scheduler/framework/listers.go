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

package framework

// FoodInfoLister interface represents anything that can list/get FoodInfo objects from food name.
type FoodInfoLister interface {
	// List returns the list of FoodInfos.
	List() ([]*FoodInfo, error)
	// HaveDguestsWithAffinityList returns the list of FoodInfos of foods with dguests with affinity terms.
	HaveDguestsWithAffinityList() ([]*FoodInfo, error)
	// HaveDguestsWithRequiredAntiAffinityList returns the list of FoodInfos of foods with dguests with required anti-affinity terms.
	HaveDguestsWithRequiredAntiAffinityList() ([]*FoodInfo, error)
	// Get returns the FoodInfo of the given food name.
	Get(foodName string) (*FoodInfo, error)
}

// StorageInfoLister interface represents anything that handles storage-related operations and resources.
type StorageInfoLister interface {
	// IsPVCUsedByDguests returns true/false on whether the PVC is used by one or more scheduled dguests,
	// keyed in the format "namespace/name".
	IsPVCUsedByDguests(key string) bool
}

// SharedLister groups scheduler-specific listers.
type SharedLister interface {
	FoodInfos() FoodInfoLister
	StorageInfos() StorageInfoLister
}