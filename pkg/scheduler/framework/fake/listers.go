/*
Copyright 2017 The Kubernetes Authors.

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
	"fmt"

	"dguest-scheduler/pkg/scheduler/framework"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	appslisters "k8s.io/client-go/listers/apps/v1"
	corelisters "k8s.io/client-go/listers/core/v1"
	storagelisters "k8s.io/client-go/listers/storage/v1"
)

var _ corelisters.ServiceLister = &ServiceLister{}

// ServiceLister implements ServiceLister on []v1.Service for test purposes.
type ServiceLister []*v1.Service

// Services returns nil.
func (f ServiceLister) Services(namespace string) corelisters.ServiceNamespaceLister {
	var services []*v1.Service
	for i := range f {
		if f[i].Namespace == namespace {
			services = append(services, f[i])
		}
	}
	return &serviceNamespaceLister{
		services:  services,
		namespace: namespace,
	}
}

// List returns v1.ServiceList, the list of all services.
func (f ServiceLister) List(labels.Selector) ([]*v1.Service, error) {
	return f, nil
}

// serviceNamespaceLister is implementation of ServiceNamespaceLister returned by Services() above.
type serviceNamespaceLister struct {
	services  []*v1.Service
	namespace string
}

func (f *serviceNamespaceLister) Get(name string) (*v1.Service, error) {
	return nil, fmt.Errorf("not implemented")
}

func (f *serviceNamespaceLister) List(selector labels.Selector) ([]*v1.Service, error) {
	return f.services, nil
}

var _ corelisters.ReplicationControllerLister = &ControllerLister{}

// ControllerLister implements ControllerLister on []v1.ReplicationController for test purposes.
type ControllerLister []*v1.ReplicationController

// List returns []v1.ReplicationController, the list of all ReplicationControllers.
func (f ControllerLister) List(labels.Selector) ([]*v1.ReplicationController, error) {
	return f, nil
}

// GetDguestControllers gets the ReplicationControllers that have the selector that match the labels on the given dguest
func (f ControllerLister) GetDguestControllers(dguest *v1alpha1.Dguest) (controllers []*v1.ReplicationController, err error) {
	var selector labels.Selector

	for i := range f {
		controller := f[i]
		if controller.Namespace != dguest.Namespace {
			continue
		}
		selector = labels.Set(controller.Spec.Selector).AsSelectorPreValidated()
		if selector.Matches(labels.Set(dguest.Labels)) {
			controllers = append(controllers, controller)
		}
	}
	if len(controllers) == 0 {
		err = fmt.Errorf("could not find Replication Controller for dguest %s in namespace %s with labels: %v", dguest.Name, dguest.Namespace, dguest.Labels)
	}

	return
}

// ReplicationControllers returns nil
func (f ControllerLister) ReplicationControllers(namespace string) corelisters.ReplicationControllerNamespaceLister {
	return nil
}

var _ appslisters.ReplicaSetLister = &ReplicaSetLister{}

// ReplicaSetLister implements ControllerLister on []extensions.ReplicaSet for test purposes.
type ReplicaSetLister []*appsv1.ReplicaSet

// List returns replica sets.
func (f ReplicaSetLister) List(labels.Selector) ([]*appsv1.ReplicaSet, error) {
	return f, nil
}

// GetDguestReplicaSets gets the ReplicaSets that have the selector that match the labels on the given dguest
func (f ReplicaSetLister) GetDguestReplicaSets(dguest *v1alpha1.Dguest) (rss []*appsv1.ReplicaSet, err error) {
	var selector labels.Selector

	for _, rs := range f {
		if rs.Namespace != dguest.Namespace {
			continue
		}
		selector, err = metav1.LabelSelectorAsSelector(rs.Spec.Selector)
		if err != nil {
			// This object has an invalid selector, it does not match the dguest
			continue
		}

		if selector.Matches(labels.Set(dguest.Labels)) {
			rss = append(rss, rs)
		}
	}
	if len(rss) == 0 {
		err = fmt.Errorf("could not find ReplicaSet for dguest %s in namespace %s with labels: %v", dguest.Name, dguest.Namespace, dguest.Labels)
	}

	return
}

// ReplicaSets returns nil
func (f ReplicaSetLister) ReplicaSets(namespace string) appslisters.ReplicaSetNamespaceLister {
	return nil
}

var _ appslisters.StatefulSetLister = &StatefulSetLister{}

// StatefulSetLister implements ControllerLister on []appsv1.StatefulSet for testing purposes.
type StatefulSetLister []*appsv1.StatefulSet

// List returns stateful sets.
func (f StatefulSetLister) List(labels.Selector) ([]*appsv1.StatefulSet, error) {
	return f, nil
}

// GetDguestStatefulSets gets the StatefulSets that have the selector that match the labels on the given dguest.
func (f StatefulSetLister) GetDguestStatefulSets(dguest *v1alpha1.Dguest) (sss []*appsv1.StatefulSet, err error) {
	var selector labels.Selector

	for _, ss := range f {
		if ss.Namespace != dguest.Namespace {
			continue
		}
		selector, err = metav1.LabelSelectorAsSelector(ss.Spec.Selector)
		if err != nil {
			// This object has an invalid selector, it does not match the dguest
			continue
		}
		if selector.Matches(labels.Set(dguest.Labels)) {
			sss = append(sss, ss)
		}
	}
	if len(sss) == 0 {
		err = fmt.Errorf("could not find StatefulSet for dguest %s in namespace %s with labels: %v", dguest.Name, dguest.Namespace, dguest.Labels)
	}
	return
}

// StatefulSets returns nil
func (f StatefulSetLister) StatefulSets(namespace string) appslisters.StatefulSetNamespaceLister {
	return nil
}

// persistentVolumeClaimNamespaceLister is implementation of PersistentVolumeClaimNamespaceLister returned by List() above.
type persistentVolumeClaimNamespaceLister struct {
	pvcs      []*v1.PersistentVolumeClaim
	namespace string
}

func (f *persistentVolumeClaimNamespaceLister) Get(name string) (*v1.PersistentVolumeClaim, error) {
	for _, pvc := range f.pvcs {
		if pvc.Name == name && pvc.Namespace == f.namespace {
			return pvc, nil
		}
	}
	return nil, fmt.Errorf("persistentvolumeclaim %q not found", name)
}

func (f persistentVolumeClaimNamespaceLister) List(selector labels.Selector) (ret []*v1.PersistentVolumeClaim, err error) {
	return nil, fmt.Errorf("not implemented")
}

// PersistentVolumeClaimLister declares a []v1.PersistentVolumeClaim type for testing.
type PersistentVolumeClaimLister []v1.PersistentVolumeClaim

var _ corelisters.PersistentVolumeClaimLister = PersistentVolumeClaimLister{}

// List gets PVC matching the namespace and PVC ID.
func (pvcs PersistentVolumeClaimLister) List(selector labels.Selector) (ret []*v1.PersistentVolumeClaim, err error) {
	return nil, fmt.Errorf("not implemented")
}

// PersistentVolumeClaims returns a fake PersistentVolumeClaimLister object.
func (pvcs PersistentVolumeClaimLister) PersistentVolumeClaims(namespace string) corelisters.PersistentVolumeClaimNamespaceLister {
	ps := make([]*v1.PersistentVolumeClaim, len(pvcs))
	for i := range pvcs {
		ps[i] = &pvcs[i]
	}
	return &persistentVolumeClaimNamespaceLister{
		pvcs:      ps,
		namespace: namespace,
	}
}

// FoodInfoLister declares a framework.FoodInfo type for testing.
type FoodInfoLister []*framework.FoodInfo

// Get returns a fake food object in the fake foods.
func (foods FoodInfoLister) Get(foodName string) (*framework.FoodInfo, error) {
	for _, food := range foods {
		if food != nil && food.Food().Name == foodName {
			return food, nil
		}
	}
	return nil, fmt.Errorf("unable to find food: %s", foodName)
}

// List lists all foods.
func (foods FoodInfoLister) List() ([]*framework.FoodInfo, error) {
	return foods, nil
}

// HaveDguestsWithAffinityList is supposed to list foods with at least one dguest with affinity. For the fake lister
// we just return everything.
func (foods FoodInfoLister) HaveDguestsWithAffinityList() ([]*framework.FoodInfo, error) {
	return foods, nil
}

// HaveDguestsWithRequiredAntiAffinityList is supposed to list foods with at least one dguest with
// required anti-affinity. For the fake lister we just return everything.
func (foods FoodInfoLister) HaveDguestsWithRequiredAntiAffinityList() ([]*framework.FoodInfo, error) {
	return foods, nil
}

var _ storagelisters.CSIFoodLister = CSIFoodLister{}

// CSIFoodLister declares a storagev1.CSIFood type for testing.
type CSIFoodLister storagev1.CSIFood

// Get returns a fake CSIFood object.
func (n CSIFoodLister) Get(name string) (*storagev1.CSIFood, error) {
	csiFood := storagev1.CSIFood(n)
	return &csiFood, nil
}

// List lists all CSIFoods in the indexer.
func (n CSIFoodLister) List(selector labels.Selector) (ret []*storagev1.CSIFood, err error) {
	return nil, fmt.Errorf("not implemented")
}

// PersistentVolumeLister declares a []v1.PersistentVolume type for testing.
type PersistentVolumeLister []v1.PersistentVolume

var _ corelisters.PersistentVolumeLister = PersistentVolumeLister{}

// Get returns a fake PV object in the fake PVs by PV ID.
func (pvs PersistentVolumeLister) Get(pvID string) (*v1.PersistentVolume, error) {
	for _, pv := range pvs {
		if pv.Name == pvID {
			return &pv, nil
		}
	}
	return nil, fmt.Errorf("unable to find persistent volume: %s", pvID)
}

// List lists all PersistentVolumes in the indexer.
func (pvs PersistentVolumeLister) List(selector labels.Selector) ([]*v1.PersistentVolume, error) {
	return nil, fmt.Errorf("not implemented")
}

// StorageClassLister declares a []storagev1.StorageClass type for testing.
type StorageClassLister []storagev1.StorageClass

var _ storagelisters.StorageClassLister = StorageClassLister{}

// Get returns a fake storage class object in the fake storage classes by name.
func (classes StorageClassLister) Get(name string) (*storagev1.StorageClass, error) {
	for _, sc := range classes {
		if sc.Name == name {
			return &sc, nil
		}
	}
	return nil, &errors.StatusError{
		ErrStatus: metav1.Status{
			Reason:  metav1.StatusReasonNotFound,
			Message: fmt.Sprintf("unable to find storage class: %s", name),
		},
	}
}

// List lists all StorageClass in the indexer.
func (classes StorageClassLister) List(selector labels.Selector) ([]*storagev1.StorageClass, error) {
	return nil, fmt.Errorf("not implemented")
}
