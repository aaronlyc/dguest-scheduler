// /*
// Copyright 2018 The Kubernetes Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// */
package framework

//
//import (
//	"fmt"
//	"reflect"
//	"strings"
//	"testing"
//
//	"github.com/google/go-cmp/cmp"
//	v1 "k8s.io/api/core/v1"
//	"k8s.io/apimachinery/pkg/api/resource"
//	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
//	"k8s.io/apimachinery/pkg/types"
//	"k8s.io/apimachinery/pkg/util/sets"
//)
//
//func TestNewResource(t *testing.T) {
//	tests := []struct {
//		name         string
//		resourceList v1alpha1.ResourceList
//		expected     *Resource
//	}{
//		{
//			name:         "empty resource",
//			resourceList: map[v1.ResourceName]resource.Quantity{},
//			expected:     &Resource{},
//		},
//		{
//			name: "complex resource",
//			resourceList: map[v1.ResourceName]resource.Quantity{
//				v1.ResourceCPU:                      *resource.NewScaledQuantity(4, -3),
//				v1.ResourceMemory:                   *resource.NewQuantity(2000, resource.BinarySI),
//				v1.ResourceDguests:                  *resource.NewQuantity(80, resource.BinarySI),
//				v1.ResourceEphemeralStorage:         *resource.NewQuantity(5000, resource.BinarySI),
//				"scalar.test/" + "scalar1":          *resource.NewQuantity(1, resource.DecimalSI),
//				v1.ResourceHugePagesPrefix + "test": *resource.NewQuantity(2, resource.BinarySI),
//			},
//			expected: &Resource{
//				MilliCPU:            4,
//				Memory:              2000,
//				Bandwidth:    5000,
//				AllowedDguestNumber: 80,
//				ScalarResources:     map[v1.ResourceName]int64{"scalar.test/scalar1": 1, "hugepages-test": 2},
//			},
//		},
//	}
//
//	for _, test := range tests {
//		t.Run(test.name, func(t *testing.T) {
//			r := NewResource(test.resourceList)
//			if !reflect.DeepEqual(test.expected, r) {
//				t.Errorf("expected: %#v, got: %#v", test.expected, r)
//			}
//		})
//	}
//}
//
//func TestResourceClone(t *testing.T) {
//	tests := []struct {
//		resource *Resource
//		expected *Resource
//	}{
//		{
//			resource: &Resource{},
//			expected: &Resource{},
//		},
//		{
//			resource: &Resource{
//				MilliCPU:            4,
//				Memory:              2000,
//				Bandwidth:    5000,
//				AllowedDguestNumber: 80,
//				ScalarResources:     map[v1.ResourceName]int64{"scalar.test/scalar1": 1, "hugepages-test": 2},
//			},
//			expected: &Resource{
//				MilliCPU:            4,
//				Memory:              2000,
//				Bandwidth:    5000,
//				AllowedDguestNumber: 80,
//				ScalarResources:     map[v1.ResourceName]int64{"scalar.test/scalar1": 1, "hugepages-test": 2},
//			},
//		},
//	}
//
//	for i, test := range tests {
//		t.Run(fmt.Sprintf("case_%d", i), func(t *testing.T) {
//			r := test.resource.Clone()
//			// Modify the field to check if the result is a clone of the origin one.
//			test.resource.MilliCPU += 1000
//			if !reflect.DeepEqual(test.expected, r) {
//				t.Errorf("expected: %#v, got: %#v", test.expected, r)
//			}
//		})
//	}
//}
//
//func TestResourceAddScalar(t *testing.T) {
//	tests := []struct {
//		resource       *Resource
//		scalarName     v1.ResourceName
//		scalarQuantity int64
//		expected       *Resource
//	}{
//		{
//			resource:       &Resource{},
//			scalarName:     "scalar1",
//			scalarQuantity: 100,
//			expected: &Resource{
//				ScalarResources: map[v1.ResourceName]int64{"scalar1": 100},
//			},
//		},
//		{
//			resource: &Resource{
//				MilliCPU:            4,
//				Memory:              2000,
//				Bandwidth:    5000,
//				AllowedDguestNumber: 80,
//				ScalarResources:     map[v1.ResourceName]int64{"hugepages-test": 2},
//			},
//			scalarName:     "scalar2",
//			scalarQuantity: 200,
//			expected: &Resource{
//				MilliCPU:            4,
//				Memory:              2000,
//				Bandwidth:    5000,
//				AllowedDguestNumber: 80,
//				ScalarResources:     map[v1.ResourceName]int64{"hugepages-test": 2, "scalar2": 200},
//			},
//		},
//	}
//
//	for _, test := range tests {
//		t.Run(string(test.scalarName), func(t *testing.T) {
//			test.resource.AddScalar(test.scalarName, test.scalarQuantity)
//			if !reflect.DeepEqual(test.expected, test.resource) {
//				t.Errorf("expected: %#v, got: %#v", test.expected, test.resource)
//			}
//		})
//	}
//}
//
//func TestSetMaxResource(t *testing.T) {
//	tests := []struct {
//		resource     *Resource
//		resourceList v1alpha1.ResourceList
//		expected     *Resource
//	}{
//		{
//			resource: &Resource{},
//			resourceList: map[v1.ResourceName]resource.Quantity{
//				v1.ResourceCPU:              *resource.NewScaledQuantity(4, -3),
//				v1.ResourceMemory:           *resource.NewQuantity(2000, resource.BinarySI),
//				v1.ResourceEphemeralStorage: *resource.NewQuantity(5000, resource.BinarySI),
//			},
//			expected: &Resource{
//				MilliCPU:         4,
//				Memory:           2000,
//				Bandwidth: 5000,
//			},
//		},
//		{
//			resource: &Resource{
//				MilliCPU:         4,
//				Memory:           4000,
//				Bandwidth: 5000,
//				ScalarResources:  map[v1.ResourceName]int64{"scalar.test/scalar1": 1, "hugepages-test": 2},
//			},
//			resourceList: map[v1.ResourceName]resource.Quantity{
//				v1.ResourceCPU:                      *resource.NewScaledQuantity(4, -3),
//				v1.ResourceMemory:                   *resource.NewQuantity(2000, resource.BinarySI),
//				v1.ResourceEphemeralStorage:         *resource.NewQuantity(7000, resource.BinarySI),
//				"scalar.test/scalar1":               *resource.NewQuantity(4, resource.DecimalSI),
//				v1.ResourceHugePagesPrefix + "test": *resource.NewQuantity(5, resource.BinarySI),
//			},
//			expected: &Resource{
//				MilliCPU:         4,
//				Memory:           4000,
//				Bandwidth: 7000,
//				ScalarResources:  map[v1.ResourceName]int64{"scalar.test/scalar1": 4, "hugepages-test": 5},
//			},
//		},
//	}
//
//	for i, test := range tests {
//		t.Run(fmt.Sprintf("case_%d", i), func(t *testing.T) {
//			test.resource.SetMaxResource(test.resourceList)
//			if !reflect.DeepEqual(test.expected, test.resource) {
//				t.Errorf("expected: %#v, got: %#v", test.expected, test.resource)
//			}
//		})
//	}
//}
//
//type testingMode interface {
//	Fatalf(format string, args ...interface{})
//}
//
//func makeBaseDguest(t testingMode, foodName, objName, cpu, mem, extended string, ports []v1.ContainerPort, volumes []v1.Volume) *v1alpha1.Dguest {
//	req := v1alpha1.ResourceList{}
//	if cpu != "" {
//		req = v1alpha1.ResourceList{
//			v1.ResourceCPU:    resource.MustParse(cpu),
//			v1.ResourceMemory: resource.MustParse(mem),
//		}
//		if extended != "" {
//			parts := strings.Split(extended, ":")
//			if len(parts) != 2 {
//				t.Fatalf("Invalid extended resource string: \"%s\"", extended)
//			}
//			req[v1.ResourceName(parts[0])] = resource.MustParse(parts[1])
//		}
//	}
//	return &v1alpha1.Dguest{
//		ObjectMeta: metav1.ObjectMeta{
//			UID:       types.UID(objName),
//			Namespace: "food_info_cache_test",
//			Name:      objName,
//		},
//		Spec: v1alpha1.DguestSpec{
//			Containers: []v1.Container{{
//				Resources: v1.ResourceRequirements{
//					Requests: req,
//				},
//				Ports: ports,
//			}},
//			FoodName: foodName,
//			Volumes:  volumes,
//		},
//	}
//}
//
//func TestNewFoodInfo(t *testing.T) {
//	foodName := "test-food"
//	dguests := []*v1alpha1.Dguest{
//		makeBaseDguest(t, foodName, "test-1", "100m", "500", "", []v1.ContainerPort{{HostIP: "127.0.0.1", HostPort: 80, Protocol: "TCP"}}, nil),
//		makeBaseDguest(t, foodName, "test-2", "200m", "1Ki", "", []v1.ContainerPort{{HostIP: "127.0.0.1", HostPort: 8080, Protocol: "TCP"}}, nil),
//	}
//
//	expected := &FoodInfo{
//		Requested: &Resource{
//			MilliCPU:            300,
//			Memory:              1524,
//			Bandwidth:    0,
//			AllowedDguestNumber: 0,
//			ScalarResources:     map[v1.ResourceName]int64(nil),
//		},
//		NonZeroRequested: &Resource{
//			MilliCPU:            300,
//			Memory:              1524,
//			Bandwidth:    0,
//			AllowedDguestNumber: 0,
//			ScalarResources:     map[v1.ResourceName]int64(nil),
//		},
//		Allocatable: &Resource{},
//		Generation:  2,
//		UsedPorts: HostPortInfo{
//			"127.0.0.1": map[ProtocolPort]struct{}{
//				{Protocol: "TCP", Port: 80}:   {},
//				{Protocol: "TCP", Port: 8080}: {},
//			},
//		},
//		ImageStates:  map[string]*ImageStateSummary{},
//		PVCRefCounts: map[string]int{},
//		Dguests: []*DguestInfo{
//			{
//				Dguest: &v1alpha1.Dguest{
//					ObjectMeta: metav1.ObjectMeta{
//						Namespace: "food_info_cache_test",
//						Name:      "test-1",
//						UID:       types.UID("test-1"),
//					},
//					Spec: v1alpha1.DguestSpec{
//						Containers: []v1.Container{
//							{
//								Resources: v1.ResourceRequirements{
//									Requests: v1alpha1.ResourceList{
//										v1.ResourceCPU:    resource.MustParse("100m"),
//										v1.ResourceMemory: resource.MustParse("500"),
//									},
//								},
//								Ports: []v1.ContainerPort{
//									{
//										HostIP:   "127.0.0.1",
//										HostPort: 80,
//										Protocol: "TCP",
//									},
//								},
//							},
//						},
//						FoodName: foodName,
//					},
//				},
//			},
//			{
//				Dguest: &v1alpha1.Dguest{
//					ObjectMeta: metav1.ObjectMeta{
//						Namespace: "food_info_cache_test",
//						Name:      "test-2",
//						UID:       types.UID("test-2"),
//					},
//					Spec: v1alpha1.DguestSpec{
//						Containers: []v1.Container{
//							{
//								Resources: v1.ResourceRequirements{
//									Requests: v1alpha1.ResourceList{
//										v1.ResourceCPU:    resource.MustParse("200m"),
//										v1.ResourceMemory: resource.MustParse("1Ki"),
//									},
//								},
//								Ports: []v1.ContainerPort{
//									{
//										HostIP:   "127.0.0.1",
//										HostPort: 8080,
//										Protocol: "TCP",
//									},
//								},
//							},
//						},
//						FoodName: foodName,
//					},
//				},
//			},
//		},
//	}
//
//	gen := generation
//	ni := NewFoodInfo(dguests...)
//	if ni.Generation <= gen {
//		t.Errorf("Generation is not incremented. previous: %v, current: %v", gen, ni.Generation)
//	}
//	expected.Generation = ni.Generation
//	if !reflect.DeepEqual(expected, ni) {
//		t.Errorf("expected: %#v, got: %#v", expected, ni)
//	}
//}
//
//func TestFoodInfoClone(t *testing.T) {
//	foodName := "test-food"
//	tests := []struct {
//		foodInfo *FoodInfo
//		expected *FoodInfo
//	}{
//		{
//			foodInfo: &FoodInfo{
//				Requested:        &Resource{},
//				NonZeroRequested: &Resource{},
//				Allocatable:      &Resource{},
//				Generation:       2,
//				UsedPorts: HostPortInfo{
//					"127.0.0.1": map[ProtocolPort]struct{}{
//						{Protocol: "TCP", Port: 80}:   {},
//						{Protocol: "TCP", Port: 8080}: {},
//					},
//				},
//				ImageStates:  map[string]*ImageStateSummary{},
//				PVCRefCounts: map[string]int{},
//				Dguests: []*DguestInfo{
//					{
//						Dguest: &v1alpha1.Dguest{
//							ObjectMeta: metav1.ObjectMeta{
//								Namespace: "food_info_cache_test",
//								Name:      "test-1",
//								UID:       types.UID("test-1"),
//							},
//							Spec: v1alpha1.DguestSpec{
//								Containers: []v1.Container{
//									{
//										Resources: v1.ResourceRequirements{
//											Requests: v1alpha1.ResourceList{
//												v1.ResourceCPU:    resource.MustParse("100m"),
//												v1.ResourceMemory: resource.MustParse("500"),
//											},
//										},
//										Ports: []v1.ContainerPort{
//											{
//												HostIP:   "127.0.0.1",
//												HostPort: 80,
//												Protocol: "TCP",
//											},
//										},
//									},
//								},
//								FoodName: foodName,
//							},
//						},
//					},
//					{
//						Dguest: &v1alpha1.Dguest{
//							ObjectMeta: metav1.ObjectMeta{
//								Namespace: "food_info_cache_test",
//								Name:      "test-2",
//								UID:       types.UID("test-2"),
//							},
//							Spec: v1alpha1.DguestSpec{
//								Containers: []v1.Container{
//									{
//										Resources: v1.ResourceRequirements{
//											Requests: v1alpha1.ResourceList{
//												v1.ResourceCPU:    resource.MustParse("200m"),
//												v1.ResourceMemory: resource.MustParse("1Ki"),
//											},
//										},
//										Ports: []v1.ContainerPort{
//											{
//												HostIP:   "127.0.0.1",
//												HostPort: 8080,
//												Protocol: "TCP",
//											},
//										},
//									},
//								},
//								FoodName: foodName,
//							},
//						},
//					},
//				},
//			},
//			expected: &FoodInfo{
//				Requested:        &Resource{},
//				NonZeroRequested: &Resource{},
//				Allocatable:      &Resource{},
//				Generation:       2,
//				UsedPorts: HostPortInfo{
//					"127.0.0.1": map[ProtocolPort]struct{}{
//						{Protocol: "TCP", Port: 80}:   {},
//						{Protocol: "TCP", Port: 8080}: {},
//					},
//				},
//				ImageStates:  map[string]*ImageStateSummary{},
//				PVCRefCounts: map[string]int{},
//				Dguests: []*DguestInfo{
//					{
//						Dguest: &v1alpha1.Dguest{
//							ObjectMeta: metav1.ObjectMeta{
//								Namespace: "food_info_cache_test",
//								Name:      "test-1",
//								UID:       types.UID("test-1"),
//							},
//							Spec: v1alpha1.DguestSpec{
//								Containers: []v1.Container{
//									{
//										Resources: v1.ResourceRequirements{
//											Requests: v1alpha1.ResourceList{
//												v1.ResourceCPU:    resource.MustParse("100m"),
//												v1.ResourceMemory: resource.MustParse("500"),
//											},
//										},
//										Ports: []v1.ContainerPort{
//											{
//												HostIP:   "127.0.0.1",
//												HostPort: 80,
//												Protocol: "TCP",
//											},
//										},
//									},
//								},
//								FoodName: foodName,
//							},
//						},
//					},
//					{
//						Dguest: &v1alpha1.Dguest{
//							ObjectMeta: metav1.ObjectMeta{
//								Namespace: "food_info_cache_test",
//								Name:      "test-2",
//								UID:       types.UID("test-2"),
//							},
//							Spec: v1alpha1.DguestSpec{
//								Containers: []v1.Container{
//									{
//										Resources: v1.ResourceRequirements{
//											Requests: v1alpha1.ResourceList{
//												v1.ResourceCPU:    resource.MustParse("200m"),
//												v1.ResourceMemory: resource.MustParse("1Ki"),
//											},
//										},
//										Ports: []v1.ContainerPort{
//											{
//												HostIP:   "127.0.0.1",
//												HostPort: 8080,
//												Protocol: "TCP",
//											},
//										},
//									},
//								},
//								FoodName: foodName,
//							},
//						},
//					},
//				},
//			},
//		},
//	}
//
//	for i, test := range tests {
//		t.Run(fmt.Sprintf("case_%d", i), func(t *testing.T) {
//			ni := test.foodInfo.Clone()
//			// Modify the field to check if the result is a clone of the origin one.
//			test.foodInfo.Generation += 10
//			test.foodInfo.UsedPorts.Remove("127.0.0.1", "TCP", 80)
//			if !reflect.DeepEqual(test.expected, ni) {
//				t.Errorf("expected: %#v, got: %#v", test.expected, ni)
//			}
//		})
//	}
//}
//
//func TestFoodInfoAddDguest(t *testing.T) {
//	foodName := "test-food"
//	dguests := []*v1alpha1.Dguest{
//		{
//			ObjectMeta: metav1.ObjectMeta{
//				Namespace: "food_info_cache_test",
//				Name:      "test-1",
//				UID:       types.UID("test-1"),
//			},
//			Spec: v1alpha1.DguestSpec{
//				Containers: []v1.Container{
//					{
//						Resources: v1.ResourceRequirements{
//							Requests: v1alpha1.ResourceList{
//								v1.ResourceCPU:    resource.MustParse("100m"),
//								v1.ResourceMemory: resource.MustParse("500"),
//							},
//						},
//						Ports: []v1.ContainerPort{
//							{
//								HostIP:   "127.0.0.1",
//								HostPort: 80,
//								Protocol: "TCP",
//							},
//						},
//					},
//				},
//				FoodName: foodName,
//				Overhead: v1alpha1.ResourceList{
//					v1.ResourceCPU: resource.MustParse("500m"),
//				},
//				Volumes: []v1.Volume{
//					{
//						VolumeSource: v1.VolumeSource{
//							PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
//								ClaimName: "pvc-1",
//							},
//						},
//					},
//				},
//			},
//		},
//		{
//			ObjectMeta: metav1.ObjectMeta{
//				Namespace: "food_info_cache_test",
//				Name:      "test-2",
//				UID:       types.UID("test-2"),
//			},
//			Spec: v1alpha1.DguestSpec{
//				Containers: []v1.Container{
//					{
//						Resources: v1.ResourceRequirements{
//							Requests: v1alpha1.ResourceList{
//								v1.ResourceCPU: resource.MustParse("200m"),
//							},
//						},
//						Ports: []v1.ContainerPort{
//							{
//								HostIP:   "127.0.0.1",
//								HostPort: 8080,
//								Protocol: "TCP",
//							},
//						},
//					},
//				},
//				FoodName: foodName,
//				Overhead: v1alpha1.ResourceList{
//					v1.ResourceCPU:    resource.MustParse("500m"),
//					v1.ResourceMemory: resource.MustParse("500"),
//				},
//				Volumes: []v1.Volume{
//					{
//						VolumeSource: v1.VolumeSource{
//							PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
//								ClaimName: "pvc-1",
//							},
//						},
//					},
//				},
//			},
//		},
//		{
//			ObjectMeta: metav1.ObjectMeta{
//				Namespace: "food_info_cache_test",
//				Name:      "test-3",
//				UID:       types.UID("test-3"),
//			},
//			Spec: v1alpha1.DguestSpec{
//				Containers: []v1.Container{
//					{
//						Resources: v1.ResourceRequirements{
//							Requests: v1alpha1.ResourceList{
//								v1.ResourceCPU: resource.MustParse("200m"),
//							},
//						},
//						Ports: []v1.ContainerPort{
//							{
//								HostIP:   "127.0.0.1",
//								HostPort: 8080,
//								Protocol: "TCP",
//							},
//						},
//					},
//				},
//				InitContainers: []v1.Container{
//					{
//						Resources: v1.ResourceRequirements{
//							Requests: v1alpha1.ResourceList{
//								v1.ResourceCPU:    resource.MustParse("500m"),
//								v1.ResourceMemory: resource.MustParse("200Mi"),
//							},
//						},
//					},
//				},
//				FoodName: foodName,
//				Overhead: v1alpha1.ResourceList{
//					v1.ResourceCPU:    resource.MustParse("500m"),
//					v1.ResourceMemory: resource.MustParse("500"),
//				},
//				Volumes: []v1.Volume{
//					{
//						VolumeSource: v1.VolumeSource{
//							PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
//								ClaimName: "pvc-2",
//							},
//						},
//					},
//				},
//			},
//		},
//	}
//	expected := &FoodInfo{
//		food: &v1alpha1.Food{
//			ObjectMeta: metav1.ObjectMeta{
//				Name: "test-food",
//			},
//		},
//		Requested: &Resource{
//			MilliCPU:            2300,
//			Memory:              209716700, //1500 + 200MB in initContainers
//			Bandwidth:    0,
//			AllowedDguestNumber: 0,
//			ScalarResources:     map[v1.ResourceName]int64(nil),
//		},
//		NonZeroRequested: &Resource{
//			MilliCPU:            2300,
//			Memory:              419431900, //200MB(initContainers) + 200MB(default memory value) + 1500 specified in requests/overhead
//			Bandwidth:    0,
//			AllowedDguestNumber: 0,
//			ScalarResources:     map[v1.ResourceName]int64(nil),
//		},
//		Allocatable: &Resource{},
//		Generation:  2,
//		UsedPorts: HostPortInfo{
//			"127.0.0.1": map[ProtocolPort]struct{}{
//				{Protocol: "TCP", Port: 80}:   {},
//				{Protocol: "TCP", Port: 8080}: {},
//			},
//		},
//		ImageStates:  map[string]*ImageStateSummary{},
//		PVCRefCounts: map[string]int{"food_info_cache_test/pvc-1": 2, "food_info_cache_test/pvc-2": 1},
//		Dguests: []*DguestInfo{
//			{
//				Dguest: &v1alpha1.Dguest{
//					ObjectMeta: metav1.ObjectMeta{
//						Namespace: "food_info_cache_test",
//						Name:      "test-1",
//						UID:       types.UID("test-1"),
//					},
//					Spec: v1alpha1.DguestSpec{
//						Containers: []v1.Container{
//							{
//								Resources: v1.ResourceRequirements{
//									Requests: v1alpha1.ResourceList{
//										v1.ResourceCPU:    resource.MustParse("100m"),
//										v1.ResourceMemory: resource.MustParse("500"),
//									},
//								},
//								Ports: []v1.ContainerPort{
//									{
//										HostIP:   "127.0.0.1",
//										HostPort: 80,
//										Protocol: "TCP",
//									},
//								},
//							},
//						},
//						FoodName: foodName,
//						Overhead: v1alpha1.ResourceList{
//							v1.ResourceCPU: resource.MustParse("500m"),
//						},
//						Volumes: []v1.Volume{
//							{
//								VolumeSource: v1.VolumeSource{
//									PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
//										ClaimName: "pvc-1",
//									},
//								},
//							},
//						},
//					},
//				},
//			},
//			{
//				Dguest: &v1alpha1.Dguest{
//					ObjectMeta: metav1.ObjectMeta{
//						Namespace: "food_info_cache_test",
//						Name:      "test-2",
//						UID:       types.UID("test-2"),
//					},
//					Spec: v1alpha1.DguestSpec{
//						Containers: []v1.Container{
//							{
//								Resources: v1.ResourceRequirements{
//									Requests: v1alpha1.ResourceList{
//										v1.ResourceCPU: resource.MustParse("200m"),
//									},
//								},
//								Ports: []v1.ContainerPort{
//									{
//										HostIP:   "127.0.0.1",
//										HostPort: 8080,
//										Protocol: "TCP",
//									},
//								},
//							},
//						},
//						FoodName: foodName,
//						Overhead: v1alpha1.ResourceList{
//							v1.ResourceCPU:    resource.MustParse("500m"),
//							v1.ResourceMemory: resource.MustParse("500"),
//						},
//						Volumes: []v1.Volume{
//							{
//								VolumeSource: v1.VolumeSource{
//									PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
//										ClaimName: "pvc-1",
//									},
//								},
//							},
//						},
//					},
//				},
//			},
//			{
//				Dguest: &v1alpha1.Dguest{
//					ObjectMeta: metav1.ObjectMeta{
//						Namespace: "food_info_cache_test",
//						Name:      "test-3",
//						UID:       types.UID("test-3"),
//					},
//					Spec: v1alpha1.DguestSpec{
//						Containers: []v1.Container{
//							{
//								Resources: v1.ResourceRequirements{
//									Requests: v1alpha1.ResourceList{
//										v1.ResourceCPU: resource.MustParse("200m"),
//									},
//								},
//								Ports: []v1.ContainerPort{
//									{
//										HostIP:   "127.0.0.1",
//										HostPort: 8080,
//										Protocol: "TCP",
//									},
//								},
//							},
//						},
//						InitContainers: []v1.Container{
//							{
//								Resources: v1.ResourceRequirements{
//									Requests: v1alpha1.ResourceList{
//										v1.ResourceCPU:    resource.MustParse("500m"),
//										v1.ResourceMemory: resource.MustParse("200Mi"),
//									},
//								},
//							},
//						},
//						FoodName: foodName,
//						Overhead: v1alpha1.ResourceList{
//							v1.ResourceCPU:    resource.MustParse("500m"),
//							v1.ResourceMemory: resource.MustParse("500"),
//						},
//						Volumes: []v1.Volume{
//							{
//								VolumeSource: v1.VolumeSource{
//									PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
//										ClaimName: "pvc-2",
//									},
//								},
//							},
//						},
//					},
//				},
//			},
//		},
//	}
//
//	ni := fakeFoodInfo()
//	gen := ni.Generation
//	for _, dguest := range dguests {
//		ni.AddDguest(dguest)
//		if ni.Generation <= gen {
//			t.Errorf("Generation is not incremented. Prev: %v, current: %v", gen, ni.Generation)
//		}
//		gen = ni.Generation
//	}
//
//	expected.Generation = ni.Generation
//	if !reflect.DeepEqual(expected, ni) {
//		t.Errorf("expected: %#v, got: %#v", expected, ni)
//	}
//}
//
//func TestFoodInfoRemoveDguest(t *testing.T) {
//	foodName := "test-food"
//	dguests := []*v1alpha1.Dguest{
//		makeBaseDguest(t, foodName, "test-1", "100m", "500", "",
//			[]v1.ContainerPort{{HostIP: "127.0.0.1", HostPort: 80, Protocol: "TCP"}},
//			[]v1.Volume{{VolumeSource: v1.VolumeSource{PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{ClaimName: "pvc-1"}}}}),
//		makeBaseDguest(t, foodName, "test-2", "200m", "1Ki", "", []v1.ContainerPort{{HostIP: "127.0.0.1", HostPort: 8080, Protocol: "TCP"}}, nil),
//	}
//
//	// add dguest Overhead
//	for _, dguest := range dguests {
//		dguest.Spec.Overhead = v1alpha1.ResourceList{
//			v1.ResourceCPU:    resource.MustParse("500m"),
//			v1.ResourceMemory: resource.MustParse("500"),
//		}
//	}
//
//	tests := []struct {
//		dguest           *v1alpha1.Dguest
//		errExpected      bool
//		expectedFoodInfo *FoodInfo
//	}{
//		{
//			dguest:      makeBaseDguest(t, foodName, "non-exist", "0", "0", "", []v1.ContainerPort{{}}, []v1.Volume{}),
//			errExpected: true,
//			expectedFoodInfo: &FoodInfo{
//				food: &v1alpha1.Food{
//					ObjectMeta: metav1.ObjectMeta{
//						Name: "test-food",
//					},
//				},
//				Requested: &Resource{
//					MilliCPU:            1300,
//					Memory:              2524,
//					Bandwidth:    0,
//					AllowedDguestNumber: 0,
//					ScalarResources:     map[v1.ResourceName]int64(nil),
//				},
//				NonZeroRequested: &Resource{
//					MilliCPU:            1300,
//					Memory:              2524,
//					Bandwidth:    0,
//					AllowedDguestNumber: 0,
//					ScalarResources:     map[v1.ResourceName]int64(nil),
//				},
//				Allocatable: &Resource{},
//				Generation:  2,
//				UsedPorts: HostPortInfo{
//					"127.0.0.1": map[ProtocolPort]struct{}{
//						{Protocol: "TCP", Port: 80}:   {},
//						{Protocol: "TCP", Port: 8080}: {},
//					},
//				},
//				ImageStates:  map[string]*ImageStateSummary{},
//				PVCRefCounts: map[string]int{"food_info_cache_test/pvc-1": 1},
//				Dguests: []*DguestInfo{
//					{
//						Dguest: &v1alpha1.Dguest{
//							ObjectMeta: metav1.ObjectMeta{
//								Namespace: "food_info_cache_test",
//								Name:      "test-1",
//								UID:       types.UID("test-1"),
//							},
//							Spec: v1alpha1.DguestSpec{
//								Containers: []v1.Container{
//									{
//										Resources: v1.ResourceRequirements{
//											Requests: v1alpha1.ResourceList{
//												v1.ResourceCPU:    resource.MustParse("100m"),
//												v1.ResourceMemory: resource.MustParse("500"),
//											},
//										},
//										Ports: []v1.ContainerPort{
//											{
//												HostIP:   "127.0.0.1",
//												HostPort: 80,
//												Protocol: "TCP",
//											},
//										},
//									},
//								},
//								FoodName: foodName,
//								Overhead: v1alpha1.ResourceList{
//									v1.ResourceCPU:    resource.MustParse("500m"),
//									v1.ResourceMemory: resource.MustParse("500"),
//								},
//								Volumes: []v1.Volume{
//									{
//										VolumeSource: v1.VolumeSource{
//											PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
//												ClaimName: "pvc-1",
//											},
//										},
//									},
//								},
//							},
//						},
//					},
//					{
//						Dguest: &v1alpha1.Dguest{
//							ObjectMeta: metav1.ObjectMeta{
//								Namespace: "food_info_cache_test",
//								Name:      "test-2",
//								UID:       types.UID("test-2"),
//							},
//							Spec: v1alpha1.DguestSpec{
//								Containers: []v1.Container{
//									{
//										Resources: v1.ResourceRequirements{
//											Requests: v1alpha1.ResourceList{
//												v1.ResourceCPU:    resource.MustParse("200m"),
//												v1.ResourceMemory: resource.MustParse("1Ki"),
//											},
//										},
//										Ports: []v1.ContainerPort{
//											{
//												HostIP:   "127.0.0.1",
//												HostPort: 8080,
//												Protocol: "TCP",
//											},
//										},
//									},
//								},
//								FoodName: foodName,
//								Overhead: v1alpha1.ResourceList{
//									v1.ResourceCPU:    resource.MustParse("500m"),
//									v1.ResourceMemory: resource.MustParse("500"),
//								},
//							},
//						},
//					},
//				},
//			},
//		},
//		{
//			dguest: &v1alpha1.Dguest{
//				ObjectMeta: metav1.ObjectMeta{
//					Namespace: "food_info_cache_test",
//					Name:      "test-1",
//					UID:       types.UID("test-1"),
//				},
//				Spec: v1alpha1.DguestSpec{
//					Containers: []v1.Container{
//						{
//							Resources: v1.ResourceRequirements{
//								Requests: v1alpha1.ResourceList{
//									v1.ResourceCPU:    resource.MustParse("100m"),
//									v1.ResourceMemory: resource.MustParse("500"),
//								},
//							},
//							Ports: []v1.ContainerPort{
//								{
//									HostIP:   "127.0.0.1",
//									HostPort: 80,
//									Protocol: "TCP",
//								},
//							},
//						},
//					},
//					FoodName: foodName,
//					Overhead: v1alpha1.ResourceList{
//						v1.ResourceCPU:    resource.MustParse("500m"),
//						v1.ResourceMemory: resource.MustParse("500"),
//					},
//					Volumes: []v1.Volume{
//						{
//							VolumeSource: v1.VolumeSource{
//								PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
//									ClaimName: "pvc-1",
//								},
//							},
//						},
//					},
//				},
//			},
//			errExpected: false,
//			expectedFoodInfo: &FoodInfo{
//				food: &v1alpha1.Food{
//					ObjectMeta: metav1.ObjectMeta{
//						Name: "test-food",
//					},
//				},
//				Requested: &Resource{
//					MilliCPU:            700,
//					Memory:              1524,
//					Bandwidth:    0,
//					AllowedDguestNumber: 0,
//					ScalarResources:     map[v1.ResourceName]int64(nil),
//				},
//				NonZeroRequested: &Resource{
//					MilliCPU:            700,
//					Memory:              1524,
//					Bandwidth:    0,
//					AllowedDguestNumber: 0,
//					ScalarResources:     map[v1.ResourceName]int64(nil),
//				},
//				Allocatable: &Resource{},
//				Generation:  3,
//				UsedPorts: HostPortInfo{
//					"127.0.0.1": map[ProtocolPort]struct{}{
//						{Protocol: "TCP", Port: 8080}: {},
//					},
//				},
//				ImageStates:  map[string]*ImageStateSummary{},
//				PVCRefCounts: map[string]int{},
//				Dguests: []*DguestInfo{
//					{
//						Dguest: &v1alpha1.Dguest{
//							ObjectMeta: metav1.ObjectMeta{
//								Namespace: "food_info_cache_test",
//								Name:      "test-2",
//								UID:       types.UID("test-2"),
//							},
//							Spec: v1alpha1.DguestSpec{
//								Containers: []v1.Container{
//									{
//										Resources: v1.ResourceRequirements{
//											Requests: v1alpha1.ResourceList{
//												v1.ResourceCPU:    resource.MustParse("200m"),
//												v1.ResourceMemory: resource.MustParse("1Ki"),
//											},
//										},
//										Ports: []v1.ContainerPort{
//											{
//												HostIP:   "127.0.0.1",
//												HostPort: 8080,
//												Protocol: "TCP",
//											},
//										},
//									},
//								},
//								FoodName: foodName,
//								Overhead: v1alpha1.ResourceList{
//									v1.ResourceCPU:    resource.MustParse("500m"),
//									v1.ResourceMemory: resource.MustParse("500"),
//								},
//							},
//						},
//					},
//				},
//			},
//		},
//	}
//
//	for i, test := range tests {
//		t.Run(fmt.Sprintf("case_%d", i), func(t *testing.T) {
//			ni := fakeFoodInfo(dguests...)
//
//			gen := ni.Generation
//			err := ni.RemoveDguest(test.dguest)
//			if err != nil {
//				if test.errExpected {
//					expectedErrorMsg := fmt.Errorf("no corresponding dguest %s in dguests of food %s", test.dguest.Name, ni.Food().Name)
//					if expectedErrorMsg == err {
//						t.Errorf("expected error: %v, got: %v", expectedErrorMsg, err)
//					}
//				} else {
//					t.Errorf("expected no error, got: %v", err)
//				}
//			} else {
//				if ni.Generation <= gen {
//					t.Errorf("Generation is not incremented. Prev: %v, current: %v", gen, ni.Generation)
//				}
//			}
//
//			test.expectedFoodInfo.Generation = ni.Generation
//			if !reflect.DeepEqual(test.expectedFoodInfo, ni) {
//				t.Errorf("expected: %#v, got: %#v", test.expectedFoodInfo, ni)
//			}
//		})
//	}
//}
//
//func fakeFoodInfo(dguests ...*v1alpha1.Dguest) *FoodInfo {
//	ni := NewFoodInfo(dguests...)
//	ni.SetFood(&v1alpha1.Food{
//		ObjectMeta: metav1.ObjectMeta{
//			Name: "test-food",
//		},
//	})
//	return ni
//}
//
//type hostPortInfoParam struct {
//	protocol, ip string
//	port         int32
//}
//
//func TestHostPortInfo_AddRemove(t *testing.T) {
//	tests := []struct {
//		desc    string
//		added   []hostPortInfoParam
//		removed []hostPortInfoParam
//		length  int
//	}{
//		{
//			desc: "normal add case",
//			added: []hostPortInfoParam{
//				{"TCP", "127.0.0.1", 79},
//				{"UDP", "127.0.0.1", 80},
//				{"TCP", "127.0.0.1", 81},
//				{"TCP", "127.0.0.1", 82},
//				// this might not make sense in real case, but the struct doesn't forbid it.
//				{"TCP", "0.0.0.0", 79},
//				{"UDP", "0.0.0.0", 80},
//				{"TCP", "0.0.0.0", 81},
//				{"TCP", "0.0.0.0", 82},
//				{"TCP", "0.0.0.0", 0},
//				{"TCP", "0.0.0.0", -1},
//			},
//			length: 8,
//		},
//		{
//			desc: "empty ip and protocol add should work",
//			added: []hostPortInfoParam{
//				{"", "127.0.0.1", 79},
//				{"UDP", "127.0.0.1", 80},
//				{"", "127.0.0.1", 81},
//				{"", "127.0.0.1", 82},
//				{"", "", 79},
//				{"UDP", "", 80},
//				{"", "", 81},
//				{"", "", 82},
//				{"", "", 0},
//				{"", "", -1},
//			},
//			length: 8,
//		},
//		{
//			desc: "normal remove case",
//			added: []hostPortInfoParam{
//				{"TCP", "127.0.0.1", 79},
//				{"UDP", "127.0.0.1", 80},
//				{"TCP", "127.0.0.1", 81},
//				{"TCP", "127.0.0.1", 82},
//				{"TCP", "0.0.0.0", 79},
//				{"UDP", "0.0.0.0", 80},
//				{"TCP", "0.0.0.0", 81},
//				{"TCP", "0.0.0.0", 82},
//			},
//			removed: []hostPortInfoParam{
//				{"TCP", "127.0.0.1", 79},
//				{"UDP", "127.0.0.1", 80},
//				{"TCP", "127.0.0.1", 81},
//				{"TCP", "127.0.0.1", 82},
//				{"TCP", "0.0.0.0", 79},
//				{"UDP", "0.0.0.0", 80},
//				{"TCP", "0.0.0.0", 81},
//				{"TCP", "0.0.0.0", 82},
//			},
//			length: 0,
//		},
//		{
//			desc: "empty ip and protocol remove should work",
//			added: []hostPortInfoParam{
//				{"TCP", "127.0.0.1", 79},
//				{"UDP", "127.0.0.1", 80},
//				{"TCP", "127.0.0.1", 81},
//				{"TCP", "127.0.0.1", 82},
//				{"TCP", "0.0.0.0", 79},
//				{"UDP", "0.0.0.0", 80},
//				{"TCP", "0.0.0.0", 81},
//				{"TCP", "0.0.0.0", 82},
//			},
//			removed: []hostPortInfoParam{
//				{"", "127.0.0.1", 79},
//				{"", "127.0.0.1", 81},
//				{"", "127.0.0.1", 82},
//				{"UDP", "127.0.0.1", 80},
//				{"", "", 79},
//				{"", "", 81},
//				{"", "", 82},
//				{"UDP", "", 80},
//			},
//			length: 0,
//		},
//	}
//
//	for _, test := range tests {
//		t.Run(test.desc, func(t *testing.T) {
//			hp := make(HostPortInfo)
//			for _, param := range test.added {
//				hp.Add(param.ip, param.protocol, param.port)
//			}
//			for _, param := range test.removed {
//				hp.Remove(param.ip, param.protocol, param.port)
//			}
//			if hp.Len() != test.length {
//				t.Errorf("%v failed: expect length %d; got %d", test.desc, test.length, hp.Len())
//				t.Error(hp)
//			}
//		})
//	}
//}
//
//func TestHostPortInfo_Check(t *testing.T) {
//	tests := []struct {
//		desc   string
//		added  []hostPortInfoParam
//		check  hostPortInfoParam
//		expect bool
//	}{
//		{
//			desc: "empty check should check 0.0.0.0 and TCP",
//			added: []hostPortInfoParam{
//				{"TCP", "127.0.0.1", 80},
//			},
//			check:  hostPortInfoParam{"", "", 81},
//			expect: false,
//		},
//		{
//			desc: "empty check should check 0.0.0.0 and TCP (conflicted)",
//			added: []hostPortInfoParam{
//				{"TCP", "127.0.0.1", 80},
//			},
//			check:  hostPortInfoParam{"", "", 80},
//			expect: true,
//		},
//		{
//			desc: "empty port check should pass",
//			added: []hostPortInfoParam{
//				{"TCP", "127.0.0.1", 80},
//			},
//			check:  hostPortInfoParam{"", "", 0},
//			expect: false,
//		},
//		{
//			desc: "0.0.0.0 should check all registered IPs",
//			added: []hostPortInfoParam{
//				{"TCP", "127.0.0.1", 80},
//			},
//			check:  hostPortInfoParam{"TCP", "0.0.0.0", 80},
//			expect: true,
//		},
//		{
//			desc: "0.0.0.0 with different protocol should be allowed",
//			added: []hostPortInfoParam{
//				{"UDP", "127.0.0.1", 80},
//			},
//			check:  hostPortInfoParam{"TCP", "0.0.0.0", 80},
//			expect: false,
//		},
//		{
//			desc: "0.0.0.0 with different port should be allowed",
//			added: []hostPortInfoParam{
//				{"TCP", "127.0.0.1", 79},
//				{"TCP", "127.0.0.1", 81},
//				{"TCP", "127.0.0.1", 82},
//			},
//			check:  hostPortInfoParam{"TCP", "0.0.0.0", 80},
//			expect: false,
//		},
//		{
//			desc: "normal ip should check all registered 0.0.0.0",
//			added: []hostPortInfoParam{
//				{"TCP", "0.0.0.0", 80},
//			},
//			check:  hostPortInfoParam{"TCP", "127.0.0.1", 80},
//			expect: true,
//		},
//		{
//			desc: "normal ip with different port/protocol should be allowed (0.0.0.0)",
//			added: []hostPortInfoParam{
//				{"TCP", "0.0.0.0", 79},
//				{"UDP", "0.0.0.0", 80},
//				{"TCP", "0.0.0.0", 81},
//				{"TCP", "0.0.0.0", 82},
//			},
//			check:  hostPortInfoParam{"TCP", "127.0.0.1", 80},
//			expect: false,
//		},
//		{
//			desc: "normal ip with different port/protocol should be allowed",
//			added: []hostPortInfoParam{
//				{"TCP", "127.0.0.1", 79},
//				{"UDP", "127.0.0.1", 80},
//				{"TCP", "127.0.0.1", 81},
//				{"TCP", "127.0.0.1", 82},
//			},
//			check:  hostPortInfoParam{"TCP", "127.0.0.1", 80},
//			expect: false,
//		},
//	}
//
//	for _, test := range tests {
//		t.Run(test.desc, func(t *testing.T) {
//			hp := make(HostPortInfo)
//			for _, param := range test.added {
//				hp.Add(param.ip, param.protocol, param.port)
//			}
//			if hp.CheckConflict(test.check.ip, test.check.protocol, test.check.port) != test.expect {
//				t.Errorf("expected %t; got %t", test.expect, !test.expect)
//			}
//		})
//	}
//}
//
//func TestGetNamespacesFromDguestAffinityTerm(t *testing.T) {
//	tests := []struct {
//		name string
//		term *v1alpha1.DguestAffinityTerm
//		want sets.String
//	}{
//		{
//			name: "dguestAffinityTerm_namespace_empty",
//			term: &v1alpha1.DguestAffinityTerm{},
//			want: sets.String{metav1.NamespaceDefault: sets.Empty{}},
//		},
//		{
//			name: "dguestAffinityTerm_namespace_not_empty",
//			term: &v1alpha1.DguestAffinityTerm{
//				Namespaces: []string{metav1.NamespacePublic, metav1.NamespaceSystem},
//			},
//			want: sets.NewString(metav1.NamespacePublic, metav1.NamespaceSystem),
//		},
//		{
//			name: "dguestAffinityTerm_namespace_selector_not_nil",
//			term: &v1alpha1.DguestAffinityTerm{
//				NamespaceSelector: &metav1.LabelSelector{},
//			},
//			want: sets.String{},
//		},
//	}
//
//	for _, test := range tests {
//		t.Run(test.name, func(t *testing.T) {
//			got := getNamespacesFromDguestAffinityTerm(&v1alpha1.Dguest{
//				ObjectMeta: metav1.ObjectMeta{
//					Name:      "topologies_dguest",
//					Namespace: metav1.NamespaceDefault,
//				},
//			}, test.term)
//			if diff := cmp.Diff(test.want, got); diff != "" {
//				t.Errorf("unexpected diff (-want, +got):\n%s", diff)
//			}
//		})
//	}
//}
