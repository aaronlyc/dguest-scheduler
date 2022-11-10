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
	"sort"
	"strings"

	"dguest-scheduler/pkg/scheduler/framework"
	internalcache "dguest-scheduler/pkg/scheduler/internal/cache"
	internalqueue "dguest-scheduler/pkg/scheduler/internal/queue"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/klog/v2"
)

// CacheComparer is an implementation of the Scheduler's cache comparer.
type CacheComparer struct {
	FoodLister   corelisters.FoodLister
	DguestLister corelisters.DguestLister
	Cache        internalcache.Cache
	DguestQueue  internalqueue.SchedulingQueue
}

// Compare compares the foods and dguests of FoodLister with Cache.Snapshot.
func (c *CacheComparer) Compare() error {
	klog.V(3).InfoS("Cache comparer started")
	defer klog.V(3).InfoS("Cache comparer finished")

	foods, err := c.FoodLister.List(labels.Everything())
	if err != nil {
		return err
	}

	dguests, err := c.DguestLister.List(labels.Everything())
	if err != nil {
		return err
	}

	dump := c.Cache.Dump()

	pendingDguests := c.DguestQueue.PendingDguests()

	if missed, redundant := c.CompareFoods(foods, dump.Foods); len(missed)+len(redundant) != 0 {
		klog.InfoS("Cache mismatch", "missedFoods", missed, "redundantFoods", redundant)
	}

	if missed, redundant := c.CompareDguests(dguests, pendingDguests, dump.Foods); len(missed)+len(redundant) != 0 {
		klog.InfoS("Cache mismatch", "missedDguests", missed, "redundantDguests", redundant)
	}

	return nil
}

// CompareFoods compares actual foods with cached foods.
func (c *CacheComparer) CompareFoods(foods []*v1alpha1.Food, foodinfos map[string]*framework.FoodInfo) (missed, redundant []string) {
	actual := []string{}
	for _, food := range foods {
		actual = append(actual, food.Name)
	}

	cached := []string{}
	for foodName := range foodinfos {
		cached = append(cached, foodName)
	}

	return compareStrings(actual, cached)
}

// CompareDguests compares actual dguests with cached dguests.
func (c *CacheComparer) CompareDguests(dguests, waitingDguests []*v1alpha1.Dguest, foodinfos map[string]*framework.FoodInfo) (missed, redundant []string) {
	actual := []string{}
	for _, dguest := range dguests {
		actual = append(actual, string(dguest.UID))
	}

	cached := []string{}
	for _, foodinfo := range foodinfos {
		for _, p := range foodinfo.Dguests {
			cached = append(cached, string(p.Dguest.UID))
		}
	}
	for _, dguest := range waitingDguests {
		cached = append(cached, string(dguest.UID))
	}

	return compareStrings(actual, cached)
}

func compareStrings(actual, cached []string) (missed, redundant []string) {
	missed, redundant = []string{}, []string{}

	sort.Strings(actual)
	sort.Strings(cached)

	compare := func(i, j int) int {
		if i == len(actual) {
			return 1
		} else if j == len(cached) {
			return -1
		}
		return strings.Compare(actual[i], cached[j])
	}

	for i, j := 0, 0; i < len(actual) || j < len(cached); {
		switch compare(i, j) {
		case 0:
			i++
			j++
		case -1:
			missed = append(missed, actual[i])
			i++
		case 1:
			redundant = append(redundant, cached[j])
			j++
		}
	}

	return
}
