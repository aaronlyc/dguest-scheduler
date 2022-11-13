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

// Code generated by lister-gen. DO NOT EDIT.

package v1alpha1

import (
	v1alpha1 "dguest-scheduler/pkg/apis/scheduler/v1alpha1"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// DguestLister helps list Dguests.
// All objects returned here must be treated as read-only.
type DguestLister interface {
	// List lists all Dguests in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.Dguest, err error)
	// Dguests returns an object that can list and get Dguests.
	Dguests(namespace string) DguestNamespaceLister
	DguestListerExpansion
}

// dguestLister implements the DguestLister interface.
type dguestLister struct {
	indexer cache.Indexer
}

// NewDguestLister returns a new DguestLister.
func NewDguestLister(indexer cache.Indexer) DguestLister {
	return &dguestLister{indexer: indexer}
}

// List lists all Dguests in the indexer.
func (s *dguestLister) List(selector labels.Selector) (ret []*v1alpha1.Dguest, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.Dguest))
	})
	return ret, err
}

// Dguests returns an object that can list and get Dguests.
func (s *dguestLister) Dguests(namespace string) DguestNamespaceLister {
	return dguestNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// DguestNamespaceLister helps list and get Dguests.
// All objects returned here must be treated as read-only.
type DguestNamespaceLister interface {
	// List lists all Dguests in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.Dguest, err error)
	// Get retrieves the Dguest from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1alpha1.Dguest, error)
	DguestNamespaceListerExpansion
}

// dguestNamespaceLister implements the DguestNamespaceLister
// interface.
type dguestNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all Dguests in the indexer for a given namespace.
func (s dguestNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.Dguest, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.Dguest))
	})
	return ret, err
}

// Get retrieves the Dguest from the indexer for a given namespace and name.
func (s dguestNamespaceLister) Get(name string) (*v1alpha1.Dguest, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("dguest"), name)
	}
	return obj.(*v1alpha1.Dguest), nil
}
