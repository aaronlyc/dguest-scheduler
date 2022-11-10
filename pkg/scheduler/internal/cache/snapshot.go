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

package cache

import (
	"dguest-scheduler/pkg/apis/scheduler/v1alpha1"
	"fmt"

	"dguest-scheduler/pkg/scheduler/framework"
	"k8s.io/apimachinery/pkg/util/sets"
)

// Snapshot is a snapshot of cache FoodInfo and FoodTree order. The scheduler takes a
// snapshot at the beginning of each scheduling cycle and uses it for its operations in that cycle.
type Snapshot struct {
	// foodInfoMap a map of food name to a snapshot of its FoodInfo.
	foodInfoMap map[string]*framework.FoodInfo
	// foodInfoList is the list of foods as ordered in the cache's foodTree.
	foodInfoList []*framework.FoodInfo
	// haveDguestsWithAffinityFoodInfoList is the list of foods with at least one dguest declaring affinity terms.
	haveDguestsWithAffinityFoodInfoList []*framework.FoodInfo
	// haveDguestsWithRequiredAntiAffinityFoodInfoList is the list of foods with at least one dguest declaring
	// required anti-affinity terms.
	haveDguestsWithRequiredAntiAffinityFoodInfoList []*framework.FoodInfo
	// usedPVCSet contains a set of PVC names that have one or more scheduled dguests using them,
	// keyed in the format "namespace/name".
	usedPVCSet sets.String
	generation int64
}

var _ framework.SharedLister = &Snapshot{}

// NewEmptySnapshot initializes a Snapshot struct and returns it.
func NewEmptySnapshot() *Snapshot {
	return &Snapshot{
		foodInfoMap: make(map[string]*framework.FoodInfo),
		usedPVCSet:  sets.NewString(),
	}
}

// NewSnapshot initializes a Snapshot struct and returns it.
func NewSnapshot(dguests []*v1alpha1.Dguest, foods []*v1alpha1.Food) *Snapshot {
	foodInfoMap := createFoodInfoMap(dguests, foods)
	foodInfoList := make([]*framework.FoodInfo, 0, len(foodInfoMap))
	haveDguestsWithAffinityFoodInfoList := make([]*framework.FoodInfo, 0, len(foodInfoMap))
	haveDguestsWithRequiredAntiAffinityFoodInfoList := make([]*framework.FoodInfo, 0, len(foodInfoMap))
	for _, v := range foodInfoMap {
		foodInfoList = append(foodInfoList, v)
		if len(v.DguestsWithAffinity) > 0 {
			haveDguestsWithAffinityFoodInfoList = append(haveDguestsWithAffinityFoodInfoList, v)
		}
		if len(v.DguestsWithRequiredAntiAffinity) > 0 {
			haveDguestsWithRequiredAntiAffinityFoodInfoList = append(haveDguestsWithRequiredAntiAffinityFoodInfoList, v)
		}
	}

	s := NewEmptySnapshot()
	s.foodInfoMap = foodInfoMap
	s.foodInfoList = foodInfoList
	s.haveDguestsWithAffinityFoodInfoList = haveDguestsWithAffinityFoodInfoList
	s.haveDguestsWithRequiredAntiAffinityFoodInfoList = haveDguestsWithRequiredAntiAffinityFoodInfoList
	s.usedPVCSet = createUsedPVCSet(dguests)

	return s
}

// createFoodInfoMap obtains a list of dguests and pivots that list into a map
// where the keys are food names and the values are the aggregated information
// for that food.
func createFoodInfoMap(dguests []*v1alpha1.Dguest, foods []*v1alpha1.Food) map[string]*framework.FoodInfo {
	foodNameToInfo := make(map[string]*framework.FoodInfo)
	for _, dguest := range dguests {
		for _, info := range dguest.Status.FoodsInfo {
			foodName := info.Name
			if _, ok := foodNameToInfo[foodName]; !ok {
				foodNameToInfo[foodName] = framework.NewFoodInfo()
			}
			foodNameToInfo[foodName].AddDguest(dguest)
		}
	}
	imageExistenceMap := createImageExistenceMap(foods)

	for _, food := range foods {
		if _, ok := foodNameToInfo[food.Name]; !ok {
			foodNameToInfo[food.Name] = framework.NewFoodInfo()
		}
		foodInfo := foodNameToInfo[food.Name]
		foodInfo.SetFood(food)
		foodInfo.ImageStates = getFoodImageStates(food, imageExistenceMap)
	}
	return foodNameToInfo
}

func createUsedPVCSet(dguests []*v1alpha1.Dguest) sets.String {
	usedPVCSet := sets.NewString()
	//for _, dguest := range dguests {
	//	if dguest.Spec.FoodName == "" {
	//		continue
	//	}
	//
	//	for _, v := range dguest.Spec.Volumes {
	//		if v.PersistentVolumeClaim == nil {
	//			continue
	//		}
	//
	//		key := framework.GetNamespacedName(dguest.Namespace, v.PersistentVolumeClaim.ClaimName)
	//		usedPVCSet.Insert(key)
	//	}
	//}
	return usedPVCSet
}

// getFoodImageStates returns the given food's image states based on the given imageExistence map.
func getFoodImageStates(food *v1alpha1.Food, imageExistenceMap map[string]sets.String) map[string]*framework.ImageStateSummary {
	imageStates := make(map[string]*framework.ImageStateSummary)

	//for _, image := range food.Status.Images {
	//	for _, name := range image.Names {
	//		imageStates[name] = &framework.ImageStateSummary{
	//			Size:     image.SizeBytes,
	//			NumFoods: len(imageExistenceMap[name]),
	//		}
	//	}
	//}
	return imageStates
}

// createImageExistenceMap returns a map recording on which foods the images exist, keyed by the images' names.
func createImageExistenceMap(foods []*v1alpha1.Food) map[string]sets.String {
	imageExistenceMap := make(map[string]sets.String)
	//for _, food := range foods {
	//	for _, image := range food.Status.Images {
	//		for _, name := range image.Names {
	//			if _, ok := imageExistenceMap[name]; !ok {
	//				imageExistenceMap[name] = sets.NewString(food.Name)
	//			} else {
	//				imageExistenceMap[name].Insert(food.Name)
	//			}
	//		}
	//	}
	//}
	return imageExistenceMap
}

// FoodInfos returns a FoodInfoLister.
func (s *Snapshot) FoodInfos() framework.FoodInfoLister {
	return s
}

// StorageInfos returns a StorageInfoLister.
func (s *Snapshot) StorageInfos() framework.StorageInfoLister {
	return s
}

// NumFoods returns the number of foods in the snapshot.
func (s *Snapshot) NumFoods() int {
	return len(s.foodInfoList)
}

// List returns the list of foods in the snapshot.
func (s *Snapshot) List() ([]*framework.FoodInfo, error) {
	return s.foodInfoList, nil
}

// HaveDguestsWithAffinityList returns the list of foods with at least one dguest with inter-dguest affinity
func (s *Snapshot) HaveDguestsWithAffinityList() ([]*framework.FoodInfo, error) {
	return s.haveDguestsWithAffinityFoodInfoList, nil
}

// HaveDguestsWithRequiredAntiAffinityList returns the list of foods with at least one dguest with
// required inter-dguest anti-affinity
func (s *Snapshot) HaveDguestsWithRequiredAntiAffinityList() ([]*framework.FoodInfo, error) {
	return s.haveDguestsWithRequiredAntiAffinityFoodInfoList, nil
}

// Get returns the FoodInfo of the given food name.
func (s *Snapshot) Get(foodName string) (*framework.FoodInfo, error) {
	if v, ok := s.foodInfoMap[foodName]; ok && v.Food() != nil {
		return v, nil
	}
	return nil, fmt.Errorf("foodinfo not found for food name %q", foodName)
}

func (s *Snapshot) IsPVCUsedByDguests(key string) bool {
	return s.usedPVCSet.Has(key)
}
