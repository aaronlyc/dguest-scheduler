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

package debugger

import (
	"dguest-scheduler/pkg/apis/scheduler/v1alpha1"
	"fmt"
	"strings"

	"k8s.io/klog/v2"

	"dguest-scheduler/pkg/scheduler/framework"
	internalcache "dguest-scheduler/pkg/scheduler/internal/cache"
	"dguest-scheduler/pkg/scheduler/internal/queue"
)

// CacheDumper writes some information from the scheduler cache and the scheduling queue to the
// scheduler logs for debugging purposes.
type CacheDumper struct {
	cache       internalcache.Cache
	dguestQueue queue.SchedulingQueue
}

// DumpAll writes cached foods and scheduling queue information to the scheduler logs.
func (d *CacheDumper) DumpAll() {
	d.dumpFoods()
	d.dumpSchedulingQueue()
}

// dumpFoods writes FoodInfo to the scheduler logs.
func (d *CacheDumper) dumpFoods() {
	dump := d.cache.Dump()
	foodInfos := make([]string, 0, len(dump.Foods))
	for name, foodInfo := range dump.Foods {
		foodInfos = append(foodInfos, d.printFoodInfo(name, foodInfo))
	}
	// Extra blank line added between food entries for readability.
	klog.InfoS("Dump of cached FoodInfo", "foods", strings.Join(foodInfos, "\n\n"))
}

// dumpSchedulingQueue writes dguests in the scheduling queue to the scheduler logs.
func (d *CacheDumper) dumpSchedulingQueue() {
	pendingDguests := d.dguestQueue.PendingDguests()
	var dguestData strings.Builder
	for _, p := range pendingDguests {
		dguestData.WriteString(printDguest(p))
	}
	klog.InfoS("Dump of scheduling queue", "dguests", dguestData.String())
}

// printFoodInfo writes parts of FoodInfo to a string.
func (d *CacheDumper) printFoodInfo(name string, n *framework.FoodInfo) string {
	var foodData strings.Builder
	foodData.WriteString(fmt.Sprintf("Food name: %s\nDeleted: %t\nRequested Resources: %+v\nAllocatable Resources:%+v\nScheduled Dguests(number: %v):\n",
		name, n.Food() == nil, n.Requested, n.Allocatable, len(n.Dguests)))
	// Dumping Dguest Info
	for _, p := range n.Dguests {
		foodData.WriteString(printDguest(p.Dguest))
	}
	// Dumping nominated dguests info on the food
	nominatedDguestInfos := d.dguestQueue.NominatedDguestsForFood(name)
	if len(nominatedDguestInfos) != 0 {
		foodData.WriteString(fmt.Sprintf("Nominated Dguests(number: %v):\n", len(nominatedDguestInfos)))
		for _, pi := range nominatedDguestInfos {
			foodData.WriteString(printDguest(pi.Dguest))
		}
	}
	return foodData.String()
}

// printDguest writes parts of a Dguest object to a string.
func printDguest(p *v1alpha1.Dguest) string {
	return fmt.Sprintf("name: %v, namespace: %v, uid: %v, phase: %v, nominated food: %v\n", p.Name, p.Namespace, p.UID, p.Status.Phase, p.Status.FoodsInfo)
}
