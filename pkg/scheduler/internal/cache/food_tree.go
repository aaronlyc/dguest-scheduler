package cache

import (
	apidguest "dguest-scheduler/pkg/api/dguest"
	"dguest-scheduler/pkg/apis/scheduler/v1alpha1"
	"fmt"

	"k8s.io/klog/v2"
)

// foodTree is a tree-like data structure that holds food names in each zone. Zone names are
// keys to "FoodTree.tree" and values of "FoodTree.tree" are arrays of food names.
// FoodTree is NOT thread-safe, any concurrent updates/reads from it must be synchronized by the caller.
// It is used only by schedulerCache, and should stay as such.
type foodTree struct {
	tree  map[string][]*v1alpha1.FoodInfoBase // a map from zone (region-zone) to an array of foods in the zone.
	zones []string                            // a list of all the zones in the tree (keys)
}

// newFoodTree creates a FoodTree from foods.
func newFoodTree(foods []*v1alpha1.Food) *foodTree {
	nt := &foodTree{
		tree: make(map[string][]*v1alpha1.FoodInfoBase),
	}
	for _, n := range foods {
		nt.addFood(n)
	}
	return nt
}

// addFood adds a food and its corresponding zone to the tree. If the zone already exists, the food
// is added to the array of foods in that zone.
func (nt *foodTree) addFood(f *v1alpha1.Food) {
	zone := apidguest.FoodCuisineVersionKey(f)
	if na, ok := nt.tree[zone]; ok {
		for _, foodBase := range na {
			if foodBase.Name == f.Name {
				klog.InfoS("Food already exists in the FoodTree", "food", klog.KObj(f))
				return
			}
		}
		nt.tree[zone] = append(na, &v1alpha1.FoodInfoBase{
			Namespace:      f.Namespace,
			Name:           f.Name,
			CuisineVersion: zone,
		})
	} else {
		nt.zones = append(nt.zones, zone)
		nt.tree[zone] = []*v1alpha1.FoodInfoBase{{
			Namespace:      f.Namespace,
			Name:           f.Name,
			CuisineVersion: zone,
		}}
	}
	klog.V(2).InfoS("Added food in listed group to FoodTree", "food", klog.KObj(f), "zone", zone)
}

// removeFood removes a food from the FoodTree.
func (nt *foodTree) removeFood(n *v1alpha1.Food) error {
	zone := apidguest.FoodCuisineVersionKey(n)
	if na, ok := nt.tree[zone]; ok {
		for i, foodInfo := range na {
			if foodInfo.Name == n.Name {
				nt.tree[zone] = append(na[:i], na[i+1:]...)
				if len(nt.tree[zone]) == 0 {
					nt.removeZone(zone)
				}
				klog.V(2).InfoS("Removed food in listed group from FoodTree", "food", klog.KObj(n), "zone", zone)
				return nil
			}
		}
	}
	klog.ErrorS(nil, "Food in listed group was not found", "food", klog.KObj(n), "zone", zone)
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
	var oldZone string
	if old != nil {
		oldZone = apidguest.FoodCuisineVersionKey(old)
	}
	newZone := apidguest.FoodCuisineVersionKey(new)
	// If the zone ID of the food has not changed, we don't need to do anything. Name of the food
	// cannot be changed in an update.
	if oldZone == newZone {
		return
	}
	nt.removeFood(old) // No error checking. We ignore whether the old food exists or not.
	nt.addFood(new)
}

func (nt *foodTree) ZoneMap() map[string][]*v1alpha1.FoodInfoBase {
	return nt.tree
}

// list returns the list of names of the food. FoodTree iterates over zones and in each zone iterates
// over foods in a round robin fashion.
func (nt *foodTree) list(cuisineVersion string) []*v1alpha1.FoodInfoBase {
	if len(nt.zones) == 0 {
		return nil
	}
	return nt.tree[cuisineVersion]
}

func (nt *foodTree) foodCount(cuisineVersion string) int {
	return len(nt.list(cuisineVersion))
}
