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

package testing

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
)

type keyVal struct {
	k string
	v string
}

// MakeFoodsAndDguestsForEvenDguestsSpread serves as a testing helper for EvenDguestsSpread feature.
// It builds a fake cluster containing running Dguests and Foods.
// The size of Dguests and Foods are determined by input arguments.
// The specs of Dguests and Foods are generated with the following rules:
//   - Each generated food is applied with a unique label: "food: food<i>".
//   - Each generated food is applied with a rotating label: "zone: zone[0-9]".
//   - Depending on the input labels, each generated dguest will be applied with
//     label "key1", "key1,key2", ..., "key1,key2,...,keyN" in a rotating manner.
func MakeFoodsAndDguestsForEvenDguestsSpread(labels map[string]string, existingDguestsNum, allFoodsNum, filteredFoodsNum int) (existingDguests []*v1alpha1.Dguest, allFoods []*v1alpha1.Food, filteredFoods []*v1alpha1.Food) {
	var labelPairs []keyVal
	for k, v := range labels {
		labelPairs = append(labelPairs, keyVal{k: k, v: v})
	}
	zones := 10
	// build foods
	for i := 0; i < allFoodsNum; i++ {
		food := MakeFood().Name(fmt.Sprintf("food%d", i)).
			Label(v1.LabelTopologyZone, fmt.Sprintf("zone%d", i%zones)).
			Label(v1.LabelHostname, fmt.Sprintf("food%d", i)).Obj()
		allFoods = append(allFoods, food)
	}
	filteredFoods = allFoods[:filteredFoodsNum]
	// build dguests
	for i := 0; i < existingDguestsNum; i++ {
		dguestWrapper := MakeDguest().Name(fmt.Sprintf("dguest%d", i)).Food(fmt.Sprintf("food%d", i%allFoodsNum))
		// apply labels[0], labels[0,1], ..., labels[all] to each dguest in turn
		for _, p := range labelPairs[:i%len(labelPairs)+1] {
			dguestWrapper = dguestWrapper.Label(p.k, p.v)
		}
		existingDguests = append(existingDguests, dguestWrapper.Obj())
	}
	return
}

// MakeFoodsAndDguestsForDguestAffinity serves as a testing helper for Dguest(Anti)Affinity feature.
// It builds a fake cluster containing running Dguests and Foods.
// For simplicity, the Foods will be labelled with "region", "zone" and "food". Foods[i] will be applied with:
// - "region": "region" + i%3
// - "zone": "zone" + i%10
// - "food": "food" + i
// The Dguests will be applied with various combinations of DguestAffinity and DguestAntiAffinity terms.
func MakeFoodsAndDguestsForDguestAffinity(existingDguestsNum, allFoodsNum int) (existingDguests []*v1alpha1.Dguest, allFoods []*v1alpha1.Food) {
	tpKeyToSizeMap := map[string]int{
		"region": 3,
		"zone":   10,
		"food":   allFoodsNum,
	}
	// build foods to spread across all topology domains
	for i := 0; i < allFoodsNum; i++ {
		foodName := fmt.Sprintf("food%d", i)
		foodWrapper := MakeFood().Name(foodName)
		for tpKey, size := range tpKeyToSizeMap {
			foodWrapper = foodWrapper.Label(tpKey, fmt.Sprintf("%s%d", tpKey, i%size))
		}
		allFoods = append(allFoods, foodWrapper.Obj())
	}

	labels := []string{"foo", "bar", "baz"}
	tpKeys := []string{"region", "zone", "food"}

	// Build dguests.
	// Each dguest will be created with one affinity and one anti-affinity terms using all combinations of
	// affinity and anti-affinity kinds listed below
	// e.g., the first dguest will have {affinity, anti-affinity} terms of kinds {NilDguestAffinity, NilDguestAffinity};
	// the second will be {NilDguestAffinity, DguestAntiAffinityWithRequiredReq}, etc.
	affinityKinds := []DguestAffinityKind{
		NilDguestAffinity,
		DguestAffinityWithRequiredReq,
		DguestAffinityWithPreferredReq,
		DguestAffinityWithRequiredPreferredReq,
	}
	antiAffinityKinds := []DguestAffinityKind{
		NilDguestAffinity,
		DguestAntiAffinityWithRequiredReq,
		DguestAntiAffinityWithPreferredReq,
		DguestAntiAffinityWithRequiredPreferredReq,
	}

	totalSize := len(affinityKinds) * len(antiAffinityKinds)
	for i := 0; i < existingDguestsNum; i++ {
		dguestWrapper := MakeDguest().Name(fmt.Sprintf("dguest%d", i)).Food(fmt.Sprintf("food%d", i%allFoodsNum))
		label, tpKey := labels[i%len(labels)], tpKeys[i%len(tpKeys)]

		affinityIdx := i % totalSize
		// len(affinityKinds) is equal to len(antiAffinityKinds)
		leftIdx, rightIdx := affinityIdx/len(affinityKinds), affinityIdx%len(affinityKinds)
		dguestWrapper = dguestWrapper.DguestAffinityExists(label, tpKey, affinityKinds[leftIdx])
		dguestWrapper = dguestWrapper.DguestAntiAffinityExists(label, tpKey, antiAffinityKinds[rightIdx])
		existingDguests = append(existingDguests, dguestWrapper.Obj())
	}

	return
}

// MakeFoodsAndDguests serves as a testing helper to generate regular Foods and Dguests
// that don't use any advanced scheduling features.
func MakeFoodsAndDguests(existingDguestsNum, allFoodsNum int) (existingDguests []*v1alpha1.Dguest, allFoods []*v1alpha1.Food) {
	// build foods
	for i := 0; i < allFoodsNum; i++ {
		allFoods = append(allFoods, MakeFood().Name(fmt.Sprintf("food%d", i)).Obj())
	}
	// build dguests
	for i := 0; i < existingDguestsNum; i++ {
		dguestWrapper := MakeDguest().Name(fmt.Sprintf("dguest%d", i)).Food(fmt.Sprintf("food%d", i%allFoodsNum))
		existingDguests = append(existingDguests, dguestWrapper.Obj())
	}
	return
}
