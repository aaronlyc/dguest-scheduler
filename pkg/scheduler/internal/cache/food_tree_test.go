package cache

//
//import (
//	"reflect"
//	"testing"
//
//	v1 "k8s.io/api/core/v1"
//	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
//)
//
//var allFoods = []*v1alpha1.Food{
//	// Food 0: a food without any region-zone label
//	{
//		ObjectMeta: metav1.ObjectMeta{
//			Name: "food-0",
//		},
//	},
//	// Food 1: a food with region label only
//	{
//		ObjectMeta: metav1.ObjectMeta{
//			Name: "food-1",
//			Labels: map[string]string{
//				v1.LabelTopologyRegion: "region-1",
//			},
//		},
//	},
//	// Food 2: a food with zone label only
//	{
//		ObjectMeta: metav1.ObjectMeta{
//			Name: "food-2",
//			Labels: map[string]string{
//				v1.LabelTopologyZone: "zone-2",
//			},
//		},
//	},
//	// Food 3: a food with proper region and zone labels
//	{
//		ObjectMeta: metav1.ObjectMeta{
//			Name: "food-3",
//			Labels: map[string]string{
//				v1.LabelTopologyRegion: "region-1",
//				v1.LabelTopologyZone:   "zone-2",
//			},
//		},
//	},
//	// Food 4: a food with proper region and zone labels
//	{
//		ObjectMeta: metav1.ObjectMeta{
//			Name: "food-4",
//			Labels: map[string]string{
//				v1.LabelTopologyRegion: "region-1",
//				v1.LabelTopologyZone:   "zone-2",
//			},
//		},
//	},
//	// Food 5: a food with proper region and zone labels in a different zone, same region as above
//	{
//		ObjectMeta: metav1.ObjectMeta{
//			Name: "food-5",
//			Labels: map[string]string{
//				v1.LabelTopologyRegion: "region-1",
//				v1.LabelTopologyZone:   "zone-3",
//			},
//		},
//	},
//	// Food 6: a food with proper region and zone labels in a new region and zone
//	{
//		ObjectMeta: metav1.ObjectMeta{
//			Name: "food-6",
//			Labels: map[string]string{
//				v1.LabelTopologyRegion: "region-2",
//				v1.LabelTopologyZone:   "zone-2",
//			},
//		},
//	},
//	// Food 7: a food with proper region and zone labels in a region and zone as food-6
//	{
//		ObjectMeta: metav1.ObjectMeta{
//			Name: "food-7",
//			Labels: map[string]string{
//				v1.LabelTopologyRegion: "region-2",
//				v1.LabelTopologyZone:   "zone-2",
//			},
//		},
//	},
//	// Food 8: a food with proper region and zone labels in a region and zone as food-6
//	{
//		ObjectMeta: metav1.ObjectMeta{
//			Name: "food-8",
//			Labels: map[string]string{
//				v1.LabelTopologyRegion: "region-2",
//				v1.LabelTopologyZone:   "zone-2",
//			},
//		},
//	},
//	// Food 9: a food with zone + region label and the deprecated zone + region label
//	{
//		ObjectMeta: metav1.ObjectMeta{
//			Name: "food-9",
//			Labels: map[string]string{
//				v1.LabelTopologyRegion:          "region-2",
//				v1.LabelTopologyZone:            "zone-2",
//				v1.LabelFailureDomainBetaRegion: "region-2",
//				v1.LabelFailureDomainBetaZone:   "zone-2",
//			},
//		},
//	},
//	// Food 10: a food with only the deprecated zone + region labels
//	{
//		ObjectMeta: metav1.ObjectMeta{
//			Name: "food-10",
//			Labels: map[string]string{
//				v1.LabelFailureDomainBetaRegion: "region-2",
//				v1.LabelFailureDomainBetaZone:   "zone-3",
//			},
//		},
//	},
//}
//
//func verifyFoodTree(t *testing.T, nt *foodTree, expectedTree map[string][]string) {
//	expectedNumFoods := int(0)
//	for _, na := range expectedTree {
//		expectedNumFoods += len(na)
//	}
//	if numFoods := nt.numFoods; numFoods != expectedNumFoods {
//		t.Errorf("unexpected foodTree.numFoods. Expected: %v, Got: %v", expectedNumFoods, numFoods)
//	}
//	if !reflect.DeepEqual(nt.tree, expectedTree) {
//		t.Errorf("The food tree is not the same as expected. Expected: %v, Got: %v", expectedTree, nt.tree)
//	}
//	if len(nt.zones) != len(expectedTree) {
//		t.Errorf("Number of zones in foodTree.zones is not expected. Expected: %v, Got: %v", len(expectedTree), len(nt.zones))
//	}
//	for _, z := range nt.zones {
//		if _, ok := expectedTree[z]; !ok {
//			t.Errorf("zone %v is not expected to exist in foodTree.zones", z)
//		}
//	}
//}
//
//func TestFoodTree_AddFood(t *testing.T) {
//	tests := []struct {
//		name         string
//		foodsToAdd   []*v1alpha1.Food
//		expectedTree map[string][]string
//	}{
//		{
//			name:         "single food no labels",
//			foodsToAdd:   allFoods[:1],
//			expectedTree: map[string][]string{"": {"food-0"}},
//		},
//		{
//			name:       "mix of foods with and without proper labels",
//			foodsToAdd: allFoods[:4],
//			expectedTree: map[string][]string{
//				"":                     {"food-0"},
//				"region-1:\x00:":       {"food-1"},
//				":\x00:zone-2":         {"food-2"},
//				"region-1:\x00:zone-2": {"food-3"},
//			},
//		},
//		{
//			name:       "mix of foods with and without proper labels and some zones with multiple foods",
//			foodsToAdd: allFoods[:7],
//			expectedTree: map[string][]string{
//				"":                     {"food-0"},
//				"region-1:\x00:":       {"food-1"},
//				":\x00:zone-2":         {"food-2"},
//				"region-1:\x00:zone-2": {"food-3", "food-4"},
//				"region-1:\x00:zone-3": {"food-5"},
//				"region-2:\x00:zone-2": {"food-6"},
//			},
//		},
//		{
//			name:       "foods also using deprecated zone/region label",
//			foodsToAdd: allFoods[9:],
//			expectedTree: map[string][]string{
//				"region-2:\x00:zone-2": {"food-9"},
//				"region-2:\x00:zone-3": {"food-10"},
//			},
//		},
//	}
//
//	for _, test := range tests {
//		t.Run(test.name, func(t *testing.T) {
//			nt := newFoodTree(nil)
//			for _, n := range test.foodsToAdd {
//				nt.addFood(n)
//			}
//			verifyFoodTree(t, nt, test.expectedTree)
//		})
//	}
//}
//
//func TestFoodTree_RemoveFood(t *testing.T) {
//	tests := []struct {
//		name          string
//		existingFoods []*v1alpha1.Food
//		foodsToRemove []*v1alpha1.Food
//		expectedTree  map[string][]string
//		expectError   bool
//	}{
//		{
//			name:          "remove a single food with no labels",
//			existingFoods: allFoods[:7],
//			foodsToRemove: allFoods[:1],
//			expectedTree: map[string][]string{
//				"region-1:\x00:":       {"food-1"},
//				":\x00:zone-2":         {"food-2"},
//				"region-1:\x00:zone-2": {"food-3", "food-4"},
//				"region-1:\x00:zone-3": {"food-5"},
//				"region-2:\x00:zone-2": {"food-6"},
//			},
//		},
//		{
//			name:          "remove a few foods including one from a zone with multiple foods",
//			existingFoods: allFoods[:7],
//			foodsToRemove: allFoods[1:4],
//			expectedTree: map[string][]string{
//				"":                     {"food-0"},
//				"region-1:\x00:zone-2": {"food-4"},
//				"region-1:\x00:zone-3": {"food-5"},
//				"region-2:\x00:zone-2": {"food-6"},
//			},
//		},
//		{
//			name:          "remove all foods",
//			existingFoods: allFoods[:7],
//			foodsToRemove: allFoods[:7],
//			expectedTree:  map[string][]string{},
//		},
//		{
//			name:          "remove non-existing food",
//			existingFoods: nil,
//			foodsToRemove: allFoods[:5],
//			expectedTree:  map[string][]string{},
//			expectError:   true,
//		},
//	}
//
//	for _, test := range tests {
//		t.Run(test.name, func(t *testing.T) {
//			nt := newFoodTree(test.existingFoods)
//			for _, n := range test.foodsToRemove {
//				err := nt.removeFood(n)
//				if test.expectError == (err == nil) {
//					t.Errorf("unexpected returned error value: %v", err)
//				}
//			}
//			verifyFoodTree(t, nt, test.expectedTree)
//		})
//	}
//}
//
//func TestFoodTree_UpdateFood(t *testing.T) {
//	tests := []struct {
//		name          string
//		existingFoods []*v1alpha1.Food
//		foodToUpdate  *v1alpha1.Food
//		expectedTree  map[string][]string
//	}{
//		{
//			name:          "update a food without label",
//			existingFoods: allFoods[:7],
//			foodToUpdate: &v1alpha1.Food{
//				ObjectMeta: metav1.ObjectMeta{
//					Name: "food-0",
//					Labels: map[string]string{
//						v1.LabelTopologyRegion: "region-1",
//						v1.LabelTopologyZone:   "zone-2",
//					},
//				},
//			},
//			expectedTree: map[string][]string{
//				"region-1:\x00:":       {"food-1"},
//				":\x00:zone-2":         {"food-2"},
//				"region-1:\x00:zone-2": {"food-3", "food-4", "food-0"},
//				"region-1:\x00:zone-3": {"food-5"},
//				"region-2:\x00:zone-2": {"food-6"},
//			},
//		},
//		{
//			name:          "update the only existing food",
//			existingFoods: allFoods[:1],
//			foodToUpdate: &v1alpha1.Food{
//				ObjectMeta: metav1.ObjectMeta{
//					Name: "food-0",
//					Labels: map[string]string{
//						v1.LabelTopologyRegion: "region-1",
//						v1.LabelTopologyZone:   "zone-2",
//					},
//				},
//			},
//			expectedTree: map[string][]string{
//				"region-1:\x00:zone-2": {"food-0"},
//			},
//		},
//		{
//			name:          "update non-existing food",
//			existingFoods: allFoods[:1],
//			foodToUpdate: &v1alpha1.Food{
//				ObjectMeta: metav1.ObjectMeta{
//					Name: "food-new",
//					Labels: map[string]string{
//						v1.LabelTopologyRegion: "region-1",
//						v1.LabelTopologyZone:   "zone-2",
//					},
//				},
//			},
//			expectedTree: map[string][]string{
//				"":                     {"food-0"},
//				"region-1:\x00:zone-2": {"food-new"},
//			},
//		},
//	}
//
//	for _, test := range tests {
//		t.Run(test.name, func(t *testing.T) {
//			nt := newFoodTree(test.existingFoods)
//			var oldFood *v1alpha1.Food
//			for _, n := range allFoods {
//				if n.Name == test.foodToUpdate.Name {
//					oldFood = n
//					break
//				}
//			}
//			if oldFood == nil {
//				oldFood = &v1alpha1.Food{ObjectMeta: metav1.ObjectMeta{Name: "nonexisting-food"}}
//			}
//			nt.updateFood(oldFood, test.foodToUpdate)
//			verifyFoodTree(t, nt, test.expectedTree)
//		})
//	}
//}
//
//func TestFoodTree_List(t *testing.T) {
//	tests := []struct {
//		name           string
//		foodsToAdd     []*v1alpha1.Food
//		expectedOutput []string
//	}{
//		{
//			name:           "empty tree",
//			foodsToAdd:     nil,
//			expectedOutput: nil,
//		},
//		{
//			name:           "one food",
//			foodsToAdd:     allFoods[:1],
//			expectedOutput: []string{"food-0"},
//		},
//		{
//			name:           "four foods",
//			foodsToAdd:     allFoods[:4],
//			expectedOutput: []string{"food-0", "food-1", "food-2", "food-3"},
//		},
//		{
//			name:           "all foods",
//			foodsToAdd:     allFoods[:9],
//			expectedOutput: []string{"food-0", "food-1", "food-2", "food-3", "food-5", "food-6", "food-4", "food-7", "food-8"},
//		},
//	}
//
//	for _, test := range tests {
//		t.Run(test.name, func(t *testing.T) {
//			nt := newFoodTree(test.foodsToAdd)
//
//			output, err := nt.list()
//			if err != nil {
//				t.Fatal(err)
//			}
//			if !reflect.DeepEqual(output, test.expectedOutput) {
//				t.Errorf("unexpected output. Expected: %v, Got: %v", test.expectedOutput, output)
//			}
//		})
//	}
//}
//
//func TestFoodTree_List_Exhausted(t *testing.T) {
//	nt := newFoodTree(allFoods[:9])
//	nt.numFoods++
//	_, err := nt.list()
//	if err == nil {
//		t.Fatal("Expected an error from zone exhaustion")
//	}
//}
//
//func TestFoodTreeMultiOperations(t *testing.T) {
//	tests := []struct {
//		name           string
//		foodsToAdd     []*v1alpha1.Food
//		foodsToRemove  []*v1alpha1.Food
//		operations     []string
//		expectedOutput []string
//	}{
//		{
//			name:           "add and remove all foods",
//			foodsToAdd:     allFoods[2:9],
//			foodsToRemove:  allFoods[2:9],
//			operations:     []string{"add", "add", "add", "remove", "remove", "remove"},
//			expectedOutput: nil,
//		},
//		{
//			name:           "add and remove some foods",
//			foodsToAdd:     allFoods[2:9],
//			foodsToRemove:  allFoods[2:9],
//			operations:     []string{"add", "add", "add", "remove"},
//			expectedOutput: []string{"food-3", "food-4"},
//		},
//		{
//			name:           "remove three foods",
//			foodsToAdd:     allFoods[2:9],
//			foodsToRemove:  allFoods[2:9],
//			operations:     []string{"add", "add", "add", "remove", "remove", "remove", "add"},
//			expectedOutput: []string{"food-5"},
//		},
//		{
//			name:           "add more foods to an exhausted zone",
//			foodsToAdd:     append(allFoods[4:9:9], allFoods[3]),
//			foodsToRemove:  nil,
//			operations:     []string{"add", "add", "add", "add", "add", "add"},
//			expectedOutput: []string{"food-4", "food-5", "food-6", "food-3", "food-7", "food-8"},
//		},
//		{
//			name:           "remove zone and add new",
//			foodsToAdd:     append(allFoods[3:5:5], allFoods[6:8]...),
//			foodsToRemove:  allFoods[3:5],
//			operations:     []string{"add", "add", "remove", "add", "add", "remove"},
//			expectedOutput: []string{"food-6", "food-7"},
//		},
//	}
//
//	for _, test := range tests {
//		t.Run(test.name, func(t *testing.T) {
//			nt := newFoodTree(nil)
//			addIndex := 0
//			removeIndex := 0
//			for _, op := range test.operations {
//				switch op {
//				case "add":
//					if addIndex >= len(test.foodsToAdd) {
//						t.Error("more add operations than foodsToAdd")
//					} else {
//						nt.addFood(test.foodsToAdd[addIndex])
//						addIndex++
//					}
//				case "remove":
//					if removeIndex >= len(test.foodsToRemove) {
//						t.Error("more remove operations than foodsToRemove")
//					} else {
//						nt.removeFood(test.foodsToRemove[removeIndex])
//						removeIndex++
//					}
//				default:
//					t.Errorf("unknown operation: %v", op)
//				}
//			}
//			output, err := nt.list()
//			if err != nil {
//				t.Fatal(err)
//			}
//			if !reflect.DeepEqual(output, test.expectedOutput) {
//				t.Errorf("unexpected output. Expected: %v, Got: %v", test.expectedOutput, output)
//			}
//		})
//	}
//}
