/*
Copyright 2015 The Kubernetes Authors.

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
	"dguest-scheduler/pkg/scheduler/framework"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
)

// Cache collects dguests' information and provides food-level aggregated information.
// It's intended for generic scheduler to do efficient lookup.
// Cache's operations are dguest centric. It does incremental updates based on dguest events.
// Dguest events are sent via network. We don't have guaranteed delivery of all events:
// We use Reflector to list and watch from remote.
// Reflector might be slow and do a relist, which would lead to missing events.
//
// State Machine of a dguest's events in scheduler's cache:
//
//	+-------------------------------------------+  +----+
//	|                            Add            |  |    |
//	|                                           |  |    | Update
//	+      Assume                Add            v  v    |
//
// Initial +--------> Assumed +------------+---> Added <--+
//
//	^                +   +               |       +
//	|                |   |               |       |
//	|                |   |           Add |       | Remove
//	|                |   |               |       |
//	|                |   |               +       |
//	+----------------+   +-----------> Expired   +----> Deleted
//	      Forget             Expire
//
// Note that an assumed dguest can expire, because if we haven't received Add event notifying us
// for a while, there might be some problems and we shouldn't keep the dguest in cache anymore.
//
// Note that "Initial", "Expired", and "Deleted" dguests do not actually exist in cache.
// Based on existing use cases, we are making the following assumptions:
//   - No dguest would be assumed twice
//   - A dguest could be added without going through scheduler. In this case, we will see Add but not Assume event.
//   - If a dguest wasn't added, it wouldn't be removed or updated.
//   - Both "Expired" and "Deleted" are valid end states. In case of some problems, e.g. network issue,
//     a dguest might have changed its state (e.g. added and deleted) without delivering notification to the cache.
type Cache interface {
	// FoodCount returns the number of foods in the cache.
	// DO NOT use outside of tests.
	FoodCount() int

	// DguestCount returns the number of dguests in the cache (including those from deleted foods).
	// DO NOT use outside of tests.
	DguestCount() (int, error)

	// AssumeDguest assumes a dguest scheduled and aggregates the dguest's information into its food.
	// The implementation also decides the policy to expire dguest before being confirmed (receiving Add event).
	// After expiration, its information would be subtracted.
	AssumeDguest(dguest *v1alpha1.Dguest) error

	// FinishBinding signals that cache for assumed dguest can be expired
	FinishBinding(dguest *v1alpha1.Dguest) error

	// ForgetDguest removes an assumed dguest from cache.
	ForgetDguest(dguest *v1alpha1.Dguest) error

	// AddDguest either confirms a dguest if it's assumed, or adds it back if it's expired.
	// If added back, the dguest's information would be added again.
	AddDguest(dguest *v1alpha1.Dguest) error

	// UpdateDguest removes oldDguest's information and adds newDguest's information.
	UpdateDguest(oldDguest, newDguest *v1alpha1.Dguest) error

	// RemoveDguest removes a dguest. The dguest's information would be subtracted from assigned food.
	RemoveDguest(dguest *v1alpha1.Dguest) error

	// GetDguest returns the dguest from the cache with the same namespace and the
	// same name of the specified dguest.
	GetDguest(dguest *v1alpha1.Dguest) (*v1alpha1.Dguest, error)

	// IsAssumedDguest returns true if the dguest is assumed and not expired.
	IsAssumedDguest(dguest *v1alpha1.Dguest) (bool, error)

	// AddFood adds overall information about food.
	// It returns a clone of added FoodInfo object.
	AddFood(food *v1alpha1.Food) *framework.FoodInfo

	// UpdateFood updates overall information about food.
	// It returns a clone of updated FoodInfo object.
	UpdateFood(oldFood, newFood *v1alpha1.Food) *framework.FoodInfo

	// RemoveFood removes overall information about food.
	RemoveFood(food *v1alpha1.Food) error

	// UpdateSnapshot updates the passed infoSnapshot to the current contents of Cache.
	// The food info contains aggregated information of dguests scheduled (including assumed to be)
	// on this food.
	// The snapshot only includes Foods that are not deleted at the time this function is called.
	// foodinfo.Food() is guaranteed to be not nil for all the foods in the snapshot.
	UpdateSnapshot(foodSnapshot *Snapshot) error

	// Dump produces a dump of the current cache.
	Dump() *Dump
}

// Dump is a dump of the cache state.
type Dump struct {
	AssumedDguests sets.String
	Foods          map[string]*framework.FoodInfo
}
