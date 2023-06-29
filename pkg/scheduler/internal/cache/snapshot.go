package cache

import (
	"fmt"

	"dguest-scheduler/pkg/scheduler/framework"
)

// Snapshot is a snapshot of cache FoodInfo and FoodTree order. The scheduler takes a
// snapshot at the beginning of each scheduling cycle and uses it for its operations in that cycle.
type Snapshot struct {
	// foodInfoMap a map of food name to a snapshot of its FoodInfo.
	foodInfoMap map[string]*framework.FoodInfo
	// foodInfoList is the list of foods as ordered in the cache's foodTree.
	// foodInfoList []*framework.FoodInfo
	// haveDguestsWithAffinityFoodInfoList is the list of foods with at least one dguest declaring affinity terms.
	haveDguestsWithAffinityFoodInfoList []*framework.FoodInfo
	// haveDguestsWithRequiredAntiAffinityFoodInfoList is the list of foods with at least one dguest declaring
	// required anti-affinity terms.
	haveDguestsWithRequiredAntiAffinityFoodInfoList []*framework.FoodInfo
	// usedPVCSet contains a set of PVC names that have one or more scheduled dguests using them,
	// keyed in the format "namespace/name".
	// usedPVCSet sets.String
	generation int64
}

var _ framework.SharedLister = &Snapshot{}

// NewEmptySnapshot initializes a Snapshot struct and returns it.
func NewEmptySnapshot() *Snapshot {
	return &Snapshot{
		foodInfoMap: make(map[string]*framework.FoodInfo),
		// usedPVCSet:  sets.NewString(),
	}
}

// NewSnapshot initializes a Snapshot struct and returns it.
// func NewSnapshot(dguests []*v1alpha1.Dguest, foods []*v1alpha1.Food) *Snapshot {
// 	foodInfoMap := createFoodInfoMap(dguests, foods)
// 	foodInfoList := make([]*framework.FoodInfo, 0, len(foodInfoMap))
// 	haveDguestsWithAffinityFoodInfoList := make([]*framework.FoodInfo, 0, len(foodInfoMap))
// 	haveDguestsWithRequiredAntiAffinityFoodInfoList := make([]*framework.FoodInfo, 0, len(foodInfoMap))
// 	for _, v := range foodInfoMap {
// 		foodInfoList = append(foodInfoList, v)
// 		if len(v.DguestsWithAffinity) > 0 {
// 			haveDguestsWithAffinityFoodInfoList = append(haveDguestsWithAffinityFoodInfoList, v)
// 		}
// 		if len(v.DguestsWithRequiredAntiAffinity) > 0 {
// 			haveDguestsWithRequiredAntiAffinityFoodInfoList = append(haveDguestsWithRequiredAntiAffinityFoodInfoList, v)
// 		}
// 	}

// 	s := NewEmptySnapshot()
// 	s.foodInfoMap = foodInfoMap
// 	s.foodInfoList = foodInfoList
// 	s.haveDguestsWithAffinityFoodInfoList = haveDguestsWithAffinityFoodInfoList
// 	s.haveDguestsWithRequiredAntiAffinityFoodInfoList = haveDguestsWithRequiredAntiAffinityFoodInfoList
// 	s.usedPVCSet = createUsedPVCSet(dguests)

// 	return s
// }

// createFoodInfoMap obtains a list of dguests and pivots that list into a map
// where the keys are food names and the values are the aggregated information
// for that food.
// func createFoodInfoMap(dguests []*v1alpha1.Dguest, foods []*v1alpha1.Food) map[string][]*framework.FoodInfo {
// 	foodNameToInfo := make(map[string][]*framework.FoodInfo)
// 	for _, dguest := range dguests {
// 		for key, info := range dguest.Status.FoodsInfo {
// 			foodName := info.Name
// 			if _, ok := foodNameToInfo[key]; !ok {
// 				foodNameToInfo[key] = []*framework.FoodInfo{&framework.FoodInfo{
// 					Food: []*framework.DguestInfo{},
// 				}}
// 			}
// 			foodNameToInfo[foodName].AddDguest(dguest)
// 		}
// 	}
// 	imageExistenceMap := createImageExistenceMap(foods)

// 	for _, food := range foods {
// 		key := apidguest.FoodcuisineKey(food)
// 		if _, ok := foodNameToInfo[key]; !ok {
// 			foodNameToInfo[key] = framework.NewFoodInfo()
// 		}
// 		foodInfo := foodNameToInfo[food.Name]
// 		foodInfo.SetFood(food)
// 		foodInfo.ImageStates = getFoodImageStates(food, imageExistenceMap)
// 	}
// 	return foodNameToInfo
// }

// FoodInfos returns a FoodInfoLister.
func (s *Snapshot) FoodInfos() framework.FoodInfoLister {
	return s
}

// StorageInfos returns a StorageInfoLister.
// func (s *Snapshot) StorageInfos() framework.StorageInfoLister {
// 	return s
// }

// NumFoods returns the number of foods in the snapshot.
func (s *Snapshot) NumFoods() int {
	return len(s.foodInfoMap)
}

// List returns the list of foods in the snapshot.
func (s *Snapshot) List() map[string]*framework.FoodInfo {
	return s.foodInfoMap
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
	if theFood, ok := s.foodInfoMap[foodName]; ok {
		return theFood, nil
	}
	return nil, fmt.Errorf("foodinfo not found for food key %+v", foodName)
}
