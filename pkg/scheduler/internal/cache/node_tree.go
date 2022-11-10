/*
Copyright 2018 The Kubernetes Authors.

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
	"errors"
	"fmt"
	"k8s.io/klog/v2"
)

// foodTree is a tree-like data structure that holds food names in each zone. Zone names are
// keys to "FoodTree.tree" and values of "FoodTree.tree" are arrays of food names.
// FoodTree is NOT thread-safe, any concurrent updates/reads from it must be synchronized by the caller.
// It is used only by schedulerCache, and should stay as such.
type foodTree struct {
	tree     map[string][]string // a map from zone (region-zone) to an array of foods in the zone.
	zones    []string            // a list of all the zones in the tree (keys)
	numFoods int
}

// newFoodTree creates a FoodTree from foods.
func newFoodTree(foods []*v1alpha1.Food) *foodTree {
	nt := &foodTree{
		tree: make(map[string][]string),
	}
	for _, n := range foods {
		nt.addFood(n)
	}
	return nt
}

// addFood adds a food and its corresponding zone to the tree. If the zone already exists, the food
// is added to the array of foods in that zone.
func (nt *foodTree) addFood(n *v1alpha1.Food) {
	zone := ""
	if na, ok := nt.tree[zone]; ok {
		for _, foodName := range na {
			if foodName == n.Name {
				klog.InfoS("Food already exists in the FoodTree", "food", klog.KObj(n))
				return
			}
		}
		nt.tree[zone] = append(na, n.Name)
	} else {
		nt.zones = append(nt.zones, zone)
		nt.tree[zone] = []string{n.Name}
	}
	klog.V(2).InfoS("Added food in listed group to FoodTree", "food", klog.KObj(n), "zone", zone)
	nt.numFoods++
}

// removeFood removes a food from the FoodTree.
func (nt *foodTree) removeFood(n *v1alpha1.Food) error {
	//zone := utilfood.GetZoneKey(n)
	//if na, ok := nt.tree[zone]; ok {
	//	for i, foodName := range na {
	//		if foodName == n.Name {
	//			nt.tree[zone] = append(na[:i], na[i+1:]...)
	//			if len(nt.tree[zone]) == 0 {
	//				nt.removeZone(zone)
	//			}
	//			klog.V(2).InfoS("Removed food in listed group from FoodTree", "food", klog.KObj(n), "zone", zone)
	//			nt.numFoods--
	//			return nil
	//		}
	//	}
	//}
	//klog.ErrorS(nil, "Food in listed group was not found", "food", klog.KObj(n), "zone", zone)
	return fmt.Errorf("food %q in group %q was not found", n.Name, "")
}

// removeZone removes a zone from tree.
// This function must be called while writer locks are hold.
func (nt *foodTree) removeZone(zone string) {
	delete(nt.tree, zone)
	for i, z := range nt.zones {
		if z == zone {
			nt.zones = append(nt.zones[:i], nt.zones[i+1:]...)
			return
		}
	}
}

// updateFood updates a food in the FoodTree.
func (nt *foodTree) updateFood(old, new *v1alpha1.Food) {
	//var oldZone string
	//if old != nil {
	//	oldZone = utilfood.GetZoneKey(old)
	//}
	//newZone := utilfood.GetZoneKey(new)
	//// If the zone ID of the food has not changed, we don't need to do anything. Name of the food
	//// cannot be changed in an update.
	//if oldZone == newZone {
	//	return
	//}
	nt.removeFood(old) // No error checking. We ignore whether the old food exists or not.
	nt.addFood(new)
}

// list returns the list of names of the food. FoodTree iterates over zones and in each zone iterates
// over foods in a round robin fashion.
func (nt *foodTree) list() ([]string, error) {
	if len(nt.zones) == 0 {
		return nil, nil
	}
	foodsList := make([]string, 0, nt.numFoods)
	numExhaustedZones := 0
	foodIndex := 0
	for len(foodsList) < nt.numFoods {
		if numExhaustedZones >= len(nt.zones) { // all zones are exhausted.
			return foodsList, errors.New("all zones exhausted before reaching count of foods expected")
		}
		for zoneIndex := 0; zoneIndex < len(nt.zones); zoneIndex++ {
			na := nt.tree[nt.zones[zoneIndex]]
			if foodIndex >= len(na) { // If the zone is exhausted, continue
				if foodIndex == len(na) { // If it is the first time the zone is exhausted
					numExhaustedZones++
				}
				continue
			}
			foodsList = append(foodsList, na[foodIndex])
		}
		foodIndex++
	}
	return foodsList, nil
}
