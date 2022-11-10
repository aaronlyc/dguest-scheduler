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

package fake

import (
	"dguest-scheduler/pkg/scheduler/framework"
	internalcache "dguest-scheduler/pkg/scheduler/internal/cache"
	v1 "k8s.io/api/core/v1"
)

// Cache is used for testing
type Cache struct {
	AssumeFunc          func(*v1alpha1.Dguest)
	ForgetFunc          func(*v1alpha1.Dguest)
	IsAssumedDguestFunc func(*v1alpha1.Dguest) bool
	GetDguestFunc       func(*v1alpha1.Dguest) *v1alpha1.Dguest
}

// AssumeDguest is a fake method for testing.
func (c *Cache) AssumeDguest(dguest *v1alpha1.Dguest) error {
	c.AssumeFunc(dguest)
	return nil
}

// FinishBinding is a fake method for testing.
func (c *Cache) FinishBinding(dguest *v1alpha1.Dguest) error { return nil }

// ForgetDguest is a fake method for testing.
func (c *Cache) ForgetDguest(dguest *v1alpha1.Dguest) error {
	c.ForgetFunc(dguest)
	return nil
}

// AddDguest is a fake method for testing.
func (c *Cache) AddDguest(dguest *v1alpha1.Dguest) error { return nil }

// UpdateDguest is a fake method for testing.
func (c *Cache) UpdateDguest(oldDguest, newDguest *v1alpha1.Dguest) error { return nil }

// RemoveDguest is a fake method for testing.
func (c *Cache) RemoveDguest(dguest *v1alpha1.Dguest) error { return nil }

// IsAssumedDguest is a fake method for testing.
func (c *Cache) IsAssumedDguest(dguest *v1alpha1.Dguest) (bool, error) {
	return c.IsAssumedDguestFunc(dguest), nil
}

// GetDguest is a fake method for testing.
func (c *Cache) GetDguest(dguest *v1alpha1.Dguest) (*v1alpha1.Dguest, error) {
	return c.GetDguestFunc(dguest), nil
}

// AddFood is a fake method for testing.
func (c *Cache) AddFood(food *v1alpha1.Food) *framework.FoodInfo { return nil }

// UpdateFood is a fake method for testing.
func (c *Cache) UpdateFood(oldFood, newFood *v1alpha1.Food) *framework.FoodInfo { return nil }

// RemoveFood is a fake method for testing.
func (c *Cache) RemoveFood(food *v1alpha1.Food) error { return nil }

// UpdateSnapshot is a fake method for testing.
func (c *Cache) UpdateSnapshot(snapshot *internalcache.Snapshot) error {
	return nil
}

// FoodCount is a fake method for testing.
func (c *Cache) FoodCount() int { return 0 }

// DguestCount is a fake method for testing.
func (c *Cache) DguestCount() (int, error) { return 0, nil }

// Dump is a fake method for testing.
func (c *Cache) Dump() *internalcache.Dump {
	return &internalcache.Dump{}
}
