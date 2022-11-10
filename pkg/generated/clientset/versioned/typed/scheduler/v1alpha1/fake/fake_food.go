/*
Copyright The Kubernetes Authors.

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

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"
	v1alpha1 "dguest-scheduler/pkg/apis/scheduler/v1alpha1"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeFoods implements FoodInterface
type FakeFoods struct {
	Fake *FakeSchedulerV1alpha1
	ns   string
}

var foodsResource = schema.GroupVersionResource{Group: "scheduler.k8s.io", Version: "v1alpha1", Resource: "foods"}

var foodsKind = schema.GroupVersionKind{Group: "scheduler.k8s.io", Version: "v1alpha1", Kind: "Food"}

// Get takes name of the food, and returns the corresponding food object, and an error if there is any.
func (c *FakeFoods) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.Food, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(foodsResource, c.ns, name), &v1alpha1.Food{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Food), err
}

// List takes label and field selectors, and returns the list of Foods that match those selectors.
func (c *FakeFoods) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.FoodList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(foodsResource, foodsKind, c.ns, opts), &v1alpha1.FoodList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.FoodList{ListMeta: obj.(*v1alpha1.FoodList).ListMeta}
	for _, item := range obj.(*v1alpha1.FoodList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested foods.
func (c *FakeFoods) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(foodsResource, c.ns, opts))

}

// Create takes the representation of a food and creates it.  Returns the server's representation of the food, and an error, if there is any.
func (c *FakeFoods) Create(ctx context.Context, food *v1alpha1.Food, opts v1.CreateOptions) (result *v1alpha1.Food, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(foodsResource, c.ns, food), &v1alpha1.Food{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Food), err
}

// Update takes the representation of a food and updates it. Returns the server's representation of the food, and an error, if there is any.
func (c *FakeFoods) Update(ctx context.Context, food *v1alpha1.Food, opts v1.UpdateOptions) (result *v1alpha1.Food, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(foodsResource, c.ns, food), &v1alpha1.Food{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Food), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeFoods) UpdateStatus(ctx context.Context, food *v1alpha1.Food, opts v1.UpdateOptions) (*v1alpha1.Food, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(foodsResource, "status", c.ns, food), &v1alpha1.Food{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Food), err
}

// Delete takes name of the food and deletes it. Returns an error if one occurs.
func (c *FakeFoods) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(foodsResource, c.ns, name, opts), &v1alpha1.Food{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeFoods) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(foodsResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.FoodList{})
	return err
}

// Patch applies the patch and returns the patched food.
func (c *FakeFoods) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.Food, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(foodsResource, c.ns, name, pt, data, subresources...), &v1alpha1.Food{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Food), err
}
