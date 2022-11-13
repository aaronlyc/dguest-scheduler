/*
Copyright The Aaron Project.

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

package v1alpha1

import (
	"context"
	v1alpha1 "dguest-scheduler/pkg/apis/scheduler/v1alpha1"
	scheme "dguest-scheduler/pkg/generated/clientset/versioned/scheme"
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// FoodsGetter has a method to return a FoodInterface.
// A group's client should implement this interface.
type FoodsGetter interface {
	Foods(namespace string) FoodInterface
}

// FoodInterface has methods to work with Food resources.
type FoodInterface interface {
	Create(ctx context.Context, food *v1alpha1.Food, opts v1.CreateOptions) (*v1alpha1.Food, error)
	Update(ctx context.Context, food *v1alpha1.Food, opts v1.UpdateOptions) (*v1alpha1.Food, error)
	UpdateStatus(ctx context.Context, food *v1alpha1.Food, opts v1.UpdateOptions) (*v1alpha1.Food, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha1.Food, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1alpha1.FoodList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.Food, err error)
	FoodExpansion
}

// foods implements FoodInterface
type foods struct {
	client rest.Interface
	ns     string
}

// newFoods returns a Foods
func newFoods(c *SchedulerV1alpha1Client, namespace string) *foods {
	return &foods{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the food, and returns the corresponding food object, and an error if there is any.
func (c *foods) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.Food, err error) {
	result = &v1alpha1.Food{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("foods").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of Foods that match those selectors.
func (c *foods) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.FoodList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha1.FoodList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("foods").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested foods.
func (c *foods) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("foods").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a food and creates it.  Returns the server's representation of the food, and an error, if there is any.
func (c *foods) Create(ctx context.Context, food *v1alpha1.Food, opts v1.CreateOptions) (result *v1alpha1.Food, err error) {
	result = &v1alpha1.Food{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("foods").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(food).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a food and updates it. Returns the server's representation of the food, and an error, if there is any.
func (c *foods) Update(ctx context.Context, food *v1alpha1.Food, opts v1.UpdateOptions) (result *v1alpha1.Food, err error) {
	result = &v1alpha1.Food{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("foods").
		Name(food.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(food).
		Do(ctx).
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *foods) UpdateStatus(ctx context.Context, food *v1alpha1.Food, opts v1.UpdateOptions) (result *v1alpha1.Food, err error) {
	result = &v1alpha1.Food{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("foods").
		Name(food.Name).
		SubResource("status").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(food).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the food and deletes it. Returns an error if one occurs.
func (c *foods) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("foods").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *foods) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("foods").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched food.
func (c *foods) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.Food, err error) {
	result = &v1alpha1.Food{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("foods").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
