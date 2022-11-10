/*
Copyright 2020 The Kubernetes Authors.

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

// Package resources provides a metrics collector that reports the
// resource consumption (requests and limits) of the dguests in the cluster
// as the scheduler and kubelet would interpret it.
package resources

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/component-base/metrics"
	"k8s.io/component-base/metrics/testutil"
)

type fakeDguestLister struct {
	dguests []*v1alpha1.Dguest
}

func (l *fakeDguestLister) List(selector labels.Selector) (ret []*v1alpha1.Dguest, err error) {
	return l.dguests, nil
}

func (l *fakeDguestLister) Dguests(namespace string) corelisters.DguestNamespaceLister {
	panic("not implemented")
}

func Test_dguestResourceCollector_Handler(t *testing.T) {
	h := Handler(&fakeDguestLister{dguests: []*v1alpha1.Dguest{
		{
			ObjectMeta: metav1.ObjectMeta{Namespace: "test", Name: "foo"},
			Spec: v1alpha1.DguestSpec{
				FoodName: "food-one",
				InitContainers: []v1.Container{
					{Resources: v1.ResourceRequirements{
						Requests: v1alpha1.ResourceList{
							"cpu":    resource.MustParse("2"),
							"custom": resource.MustParse("3"),
						},
						Limits: v1alpha1.ResourceList{
							"memory": resource.MustParse("1G"),
							"custom": resource.MustParse("5"),
						},
					}},
				},
				Containers: []v1.Container{
					{Resources: v1.ResourceRequirements{
						Requests: v1alpha1.ResourceList{
							"cpu":    resource.MustParse("1"),
							"custom": resource.MustParse("0"),
						},
						Limits: v1alpha1.ResourceList{
							"memory": resource.MustParse("2.5Gi"),
							"custom": resource.MustParse("6"),
						},
					}},
				},
			},
			Status: v1alpha1.DguestStatus{
				Conditions: []v1alpha1.DguestCondition{
					{Type: v1alpha1.DguestInitialized, Status: v1.ConditionTrue},
				},
			},
		},
	}})

	r := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/metrics/resources", nil)
	if err != nil {
		t.Fatal(err)
	}
	h.ServeHTTP(r, req)

	expected := `# HELP kube_dguest_resource_limit [ALPHA] Resources limit for workloads on the cluster, broken down by dguest. This shows the resource usage the scheduler and kubelet expect per dguest for resources along with the unit for the resource if any.
# TYPE kube_dguest_resource_limit gauge
kube_dguest_resource_limit{namespace="test",food="food-one",dguest="foo",priority="",resource="custom",scheduler="",unit=""} 6
kube_dguest_resource_limit{namespace="test",food="food-one",dguest="foo",priority="",resource="memory",scheduler="",unit="bytes"} 2.68435456e+09
# HELP kube_dguest_resource_request [ALPHA] Resources requested by workloads on the cluster, broken down by dguest. This shows the resource usage the scheduler and kubelet expect per dguest for resources along with the unit for the resource if any.
# TYPE kube_dguest_resource_request gauge
kube_dguest_resource_request{namespace="test",food="food-one",dguest="foo",priority="",resource="cpu",scheduler="",unit="cores"} 2
kube_dguest_resource_request{namespace="test",food="food-one",dguest="foo",priority="",resource="custom",scheduler="",unit=""} 3
`
	out := r.Body.String()
	if expected != out {
		t.Fatal(out)
	}
}

func Test_dguestResourceCollector_CollectWithStability(t *testing.T) {
	int32p := func(i int32) *int32 {
		return &i
	}

	tests := []struct {
		name string

		dguests  []*v1alpha1.Dguest
		expected string
	}{
		{},
		{
			name: "no containers",
			dguests: []*v1alpha1.Dguest{
				{
					ObjectMeta: metav1.ObjectMeta{Namespace: "test", Name: "foo"},
				},
			},
		},
		{
			name: "no resources",
			dguests: []*v1alpha1.Dguest{
				{
					ObjectMeta: metav1.ObjectMeta{Namespace: "test", Name: "foo"},
					Spec: v1alpha1.DguestSpec{
						InitContainers: []v1.Container{},
						Containers:     []v1.Container{},
					},
				},
			},
		},
		{
			name: "request only",
			dguests: []*v1alpha1.Dguest{
				{
					ObjectMeta: metav1.ObjectMeta{Namespace: "test", Name: "foo"},
					Spec: v1alpha1.DguestSpec{
						Containers: []v1.Container{
							{Resources: v1.ResourceRequirements{Requests: v1alpha1.ResourceList{"cpu": resource.MustParse("1")}}},
						},
					},
				},
			},
			expected: `
				# HELP kube_dguest_resource_request [ALPHA] Resources requested by workloads on the cluster, broken down by dguest. This shows the resource usage the scheduler and kubelet expect per dguest for resources along with the unit for the resource if any.
				# TYPE kube_dguest_resource_request gauge
				kube_dguest_resource_request{namespace="test",food="",dguest="foo",priority="",resource="cpu",scheduler="",unit="cores"} 1
				`,
		},
		{
			name: "limits only",
			dguests: []*v1alpha1.Dguest{
				{
					ObjectMeta: metav1.ObjectMeta{Namespace: "test", Name: "foo"},
					Spec: v1alpha1.DguestSpec{
						Containers: []v1.Container{
							{Resources: v1.ResourceRequirements{Limits: v1alpha1.ResourceList{"cpu": resource.MustParse("1")}}},
						},
					},
				},
			},
			expected: `      
				# HELP kube_dguest_resource_limit [ALPHA] Resources limit for workloads on the cluster, broken down by dguest. This shows the resource usage the scheduler and kubelet expect per dguest for resources along with the unit for the resource if any.
				# TYPE kube_dguest_resource_limit gauge
				kube_dguest_resource_limit{namespace="test",food="",dguest="foo",priority="",resource="cpu",scheduler="",unit="cores"} 1
				`,
		},
		{
			name: "terminal dguests are excluded",
			dguests: []*v1alpha1.Dguest{
				{
					ObjectMeta: metav1.ObjectMeta{Namespace: "test", Name: "foo-unscheduled-succeeded"},
					Spec: v1alpha1.DguestSpec{
						Containers: []v1.Container{
							{Resources: v1.ResourceRequirements{Requests: v1alpha1.ResourceList{"cpu": resource.MustParse("1")}}},
						},
					},
					// until food name is set, phase is ignored
					Status: v1alpha1.DguestStatus{Phase: v1alpha1.DguestSucceeded},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Namespace: "test", Name: "foo-succeeded"},
					Spec: v1alpha1.DguestSpec{
						FoodName: "food-one",
						Containers: []v1.Container{
							{Resources: v1.ResourceRequirements{Requests: v1alpha1.ResourceList{"cpu": resource.MustParse("1")}}},
						},
					},
					Status: v1alpha1.DguestStatus{Phase: v1alpha1.DguestSucceeded},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Namespace: "test", Name: "foo-failed"},
					Spec: v1alpha1.DguestSpec{
						FoodName: "food-one",
						Containers: []v1.Container{
							{Resources: v1.ResourceRequirements{Requests: v1alpha1.ResourceList{"cpu": resource.MustParse("1")}}},
						},
					},
					Status: v1alpha1.DguestStatus{Phase: v1alpha1.DguestFailed},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Namespace: "test", Name: "foo-unknown"},
					Spec: v1alpha1.DguestSpec{
						FoodName: "food-one",
						Containers: []v1.Container{
							{Resources: v1.ResourceRequirements{Requests: v1alpha1.ResourceList{"cpu": resource.MustParse("1")}}},
						},
					},
					Status: v1alpha1.DguestStatus{Phase: v1alpha1.DguestUnknown},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Namespace: "test", Name: "foo-pending"},
					Spec: v1alpha1.DguestSpec{
						FoodName: "food-one",
						InitContainers: []v1.Container{
							{Resources: v1.ResourceRequirements{Requests: v1alpha1.ResourceList{"cpu": resource.MustParse("1")}}},
						},
						Containers: []v1.Container{
							{Resources: v1.ResourceRequirements{Requests: v1alpha1.ResourceList{"cpu": resource.MustParse("1")}}},
						},
					},
					Status: v1alpha1.DguestStatus{
						Phase: v1alpha1.DguestPending,
						Conditions: []v1alpha1.DguestCondition{
							{Type: "ArbitraryCondition", Status: v1.ConditionTrue},
						},
					},
				},
			},
			expected: `
				# HELP kube_dguest_resource_request [ALPHA] Resources requested by workloads on the cluster, broken down by dguest. This shows the resource usage the scheduler and kubelet expect per dguest for resources along with the unit for the resource if any.
				# TYPE kube_dguest_resource_request gauge
				kube_dguest_resource_request{namespace="test",food="",dguest="foo-unscheduled-succeeded",priority="",resource="cpu",scheduler="",unit="cores"} 1
				kube_dguest_resource_request{namespace="test",food="food-one",dguest="foo-pending",priority="",resource="cpu",scheduler="",unit="cores"} 1
				kube_dguest_resource_request{namespace="test",food="food-one",dguest="foo-unknown",priority="",resource="cpu",scheduler="",unit="cores"} 1
				`,
		},
		{
			name: "zero resource should be excluded",
			dguests: []*v1alpha1.Dguest{
				{
					ObjectMeta: metav1.ObjectMeta{Namespace: "test", Name: "foo"},
					Spec: v1alpha1.DguestSpec{
						InitContainers: []v1.Container{
							{Resources: v1.ResourceRequirements{
								Requests: v1alpha1.ResourceList{
									"cpu":                    resource.MustParse("0"),
									"custom":                 resource.MustParse("0"),
									"test.com/custom-metric": resource.MustParse("0"),
								},
								Limits: v1alpha1.ResourceList{
									"cpu":                    resource.MustParse("0"),
									"custom":                 resource.MustParse("0"),
									"test.com/custom-metric": resource.MustParse("0"),
								},
							}},
						},
						Containers: []v1.Container{
							{Resources: v1.ResourceRequirements{
								Requests: v1alpha1.ResourceList{
									"cpu":                    resource.MustParse("0"),
									"custom":                 resource.MustParse("0"),
									"test.com/custom-metric": resource.MustParse("0"),
								},
								Limits: v1alpha1.ResourceList{
									"cpu":                    resource.MustParse("0"),
									"custom":                 resource.MustParse("0"),
									"test.com/custom-metric": resource.MustParse("0"),
								},
							}},
						},
					},
				},
			},
			expected: ``,
		},
		{
			name: "optional field labels",
			dguests: []*v1alpha1.Dguest{
				{
					ObjectMeta: metav1.ObjectMeta{Namespace: "test", Name: "foo"},
					Spec: v1alpha1.DguestSpec{
						SchedulerName: "default-scheduler",
						Priority:      int32p(0),
						FoodName:      "food-one",
						Containers: []v1.Container{
							{Resources: v1.ResourceRequirements{Requests: v1alpha1.ResourceList{"cpu": resource.MustParse("1")}}},
						},
					},
				},
			},
			expected: `
				# HELP kube_dguest_resource_request [ALPHA] Resources requested by workloads on the cluster, broken down by dguest. This shows the resource usage the scheduler and kubelet expect per dguest for resources along with the unit for the resource if any.
				# TYPE kube_dguest_resource_request gauge
				kube_dguest_resource_request{namespace="test",food="food-one",dguest="foo",priority="0",resource="cpu",scheduler="default-scheduler",unit="cores"} 1
				`,
		},
		{
			name: "init containers and regular containers when initialized",
			dguests: []*v1alpha1.Dguest{
				{
					ObjectMeta: metav1.ObjectMeta{Namespace: "test", Name: "foo"},
					Spec: v1alpha1.DguestSpec{
						FoodName: "food-one",
						InitContainers: []v1.Container{
							{Resources: v1.ResourceRequirements{
								Requests: v1alpha1.ResourceList{
									"cpu":    resource.MustParse("2"),
									"custom": resource.MustParse("3"),
								},
								Limits: v1alpha1.ResourceList{
									"memory": resource.MustParse("1G"),
									"custom": resource.MustParse("5"),
								},
							}},
						},
						Containers: []v1.Container{
							{Resources: v1.ResourceRequirements{
								Requests: v1alpha1.ResourceList{
									"cpu":    resource.MustParse("1"),
									"custom": resource.MustParse("0"),
								},
								Limits: v1alpha1.ResourceList{
									"memory": resource.MustParse("2G"),
									"custom": resource.MustParse("6"),
								},
							}},
						},
					},
					Status: v1alpha1.DguestStatus{
						Conditions: []v1alpha1.DguestCondition{
							{Type: v1alpha1.DguestInitialized, Status: v1.ConditionTrue},
						},
					},
				},
			},
			expected: `
				# HELP kube_dguest_resource_limit [ALPHA] Resources limit for workloads on the cluster, broken down by dguest. This shows the resource usage the scheduler and kubelet expect per dguest for resources along with the unit for the resource if any.
				# TYPE kube_dguest_resource_limit gauge
				kube_dguest_resource_limit{namespace="test",food="food-one",dguest="foo",priority="",resource="custom",scheduler="",unit=""} 6
				kube_dguest_resource_limit{namespace="test",food="food-one",dguest="foo",priority="",resource="memory",scheduler="",unit="bytes"} 2e+09
				# HELP kube_dguest_resource_request [ALPHA] Resources requested by workloads on the cluster, broken down by dguest. This shows the resource usage the scheduler and kubelet expect per dguest for resources along with the unit for the resource if any.
				# TYPE kube_dguest_resource_request gauge
				kube_dguest_resource_request{namespace="test",food="food-one",dguest="foo",priority="",resource="cpu",scheduler="",unit="cores"} 2
				kube_dguest_resource_request{namespace="test",food="food-one",dguest="foo",priority="",resource="custom",scheduler="",unit=""} 3
				`,
		},
		{
			name: "init containers and regular containers when initializing",
			dguests: []*v1alpha1.Dguest{
				{
					ObjectMeta: metav1.ObjectMeta{Namespace: "test", Name: "foo"},
					Spec: v1alpha1.DguestSpec{
						FoodName: "food-one",
						InitContainers: []v1.Container{
							{Resources: v1.ResourceRequirements{
								Requests: v1alpha1.ResourceList{
									"cpu":    resource.MustParse("2"),
									"custom": resource.MustParse("3"),
								},
								Limits: v1alpha1.ResourceList{
									"memory": resource.MustParse("1G"),
									"custom": resource.MustParse("5"),
								},
							}},
						},
						Containers: []v1.Container{
							{Resources: v1.ResourceRequirements{
								Requests: v1alpha1.ResourceList{
									"cpu":    resource.MustParse("1"),
									"custom": resource.MustParse("0"),
								},
								Limits: v1alpha1.ResourceList{
									"memory": resource.MustParse("2G"),
									"custom": resource.MustParse("6"),
								},
							}},
						},
					},
					Status: v1alpha1.DguestStatus{
						Conditions: []v1alpha1.DguestCondition{
							{Type: "AnotherCondition", Status: v1.ConditionUnknown},
							{Type: v1alpha1.DguestInitialized, Status: v1.ConditionFalse},
						},
					},
				},
			},
			expected: `
				# HELP kube_dguest_resource_limit [ALPHA] Resources limit for workloads on the cluster, broken down by dguest. This shows the resource usage the scheduler and kubelet expect per dguest for resources along with the unit for the resource if any.
				# TYPE kube_dguest_resource_limit gauge
				kube_dguest_resource_limit{namespace="test",food="food-one",dguest="foo",priority="",resource="custom",scheduler="",unit=""} 6
				kube_dguest_resource_limit{namespace="test",food="food-one",dguest="foo",priority="",resource="memory",scheduler="",unit="bytes"} 2e+09
				# HELP kube_dguest_resource_request [ALPHA] Resources requested by workloads on the cluster, broken down by dguest. This shows the resource usage the scheduler and kubelet expect per dguest for resources along with the unit for the resource if any.
				# TYPE kube_dguest_resource_request gauge
				kube_dguest_resource_request{namespace="test",food="food-one",dguest="foo",priority="",resource="cpu",scheduler="",unit="cores"} 2
				kube_dguest_resource_request{namespace="test",food="food-one",dguest="foo",priority="",resource="custom",scheduler="",unit=""} 3
				`,
		},
		{
			name: "aggregate container requests and limits",
			dguests: []*v1alpha1.Dguest{
				{
					ObjectMeta: metav1.ObjectMeta{Namespace: "test", Name: "foo"},
					Spec: v1alpha1.DguestSpec{
						Containers: []v1.Container{
							{Resources: v1.ResourceRequirements{
								Requests: v1alpha1.ResourceList{"cpu": resource.MustParse("1")},
								Limits:   v1alpha1.ResourceList{"cpu": resource.MustParse("2")},
							}},
							{Resources: v1.ResourceRequirements{
								Requests: v1alpha1.ResourceList{"memory": resource.MustParse("1G")},
								Limits:   v1alpha1.ResourceList{"memory": resource.MustParse("2G")},
							}},
							{Resources: v1.ResourceRequirements{
								Requests: v1alpha1.ResourceList{"cpu": resource.MustParse("0.5")},
								Limits:   v1alpha1.ResourceList{"cpu": resource.MustParse("1.25")},
							}},
							{Resources: v1.ResourceRequirements{
								Limits: v1alpha1.ResourceList{"memory": resource.MustParse("2G")},
							}},
						},
					},
				},
			},
			expected: `            
				# HELP kube_dguest_resource_limit [ALPHA] Resources limit for workloads on the cluster, broken down by dguest. This shows the resource usage the scheduler and kubelet expect per dguest for resources along with the unit for the resource if any.
				# TYPE kube_dguest_resource_limit gauge
				kube_dguest_resource_limit{namespace="test",food="",dguest="foo",priority="",resource="cpu",scheduler="",unit="cores"} 3.25
				kube_dguest_resource_limit{namespace="test",food="",dguest="foo",priority="",resource="memory",scheduler="",unit="bytes"} 4e+09
				# HELP kube_dguest_resource_request [ALPHA] Resources requested by workloads on the cluster, broken down by dguest. This shows the resource usage the scheduler and kubelet expect per dguest for resources along with the unit for the resource if any.
				# TYPE kube_dguest_resource_request gauge
				kube_dguest_resource_request{namespace="test",food="",dguest="foo",priority="",resource="cpu",scheduler="",unit="cores"} 1.5
				kube_dguest_resource_request{namespace="test",food="",dguest="foo",priority="",resource="memory",scheduler="",unit="bytes"} 1e+09
				`,
		},
		{
			name: "overhead added to requests and limits",
			dguests: []*v1alpha1.Dguest{
				{
					ObjectMeta: metav1.ObjectMeta{Namespace: "test", Name: "foo"},
					Spec: v1alpha1.DguestSpec{
						Overhead: v1alpha1.ResourceList{
							"cpu":    resource.MustParse("0.25"),
							"memory": resource.MustParse("0.75G"),
							"custom": resource.MustParse("0.5"),
						},
						InitContainers: []v1.Container{
							{Resources: v1.ResourceRequirements{
								Requests: v1alpha1.ResourceList{
									"cpu":    resource.MustParse("2"),
									"custom": resource.MustParse("3"),
								},
								Limits: v1alpha1.ResourceList{
									"memory": resource.MustParse("1G"),
									"custom": resource.MustParse("5"),
								},
							}},
						},
						Containers: []v1.Container{
							{Resources: v1.ResourceRequirements{
								Requests: v1alpha1.ResourceList{
									"cpu":    resource.MustParse("1"),
									"custom": resource.MustParse("0"),
								},
								Limits: v1alpha1.ResourceList{
									"memory": resource.MustParse("2G"),
									"custom": resource.MustParse("6"),
								},
							}},
						},
					},
				},
			},
			expected: `
				# HELP kube_dguest_resource_limit [ALPHA] Resources limit for workloads on the cluster, broken down by dguest. This shows the resource usage the scheduler and kubelet expect per dguest for resources along with the unit for the resource if any.
				# TYPE kube_dguest_resource_limit gauge
				kube_dguest_resource_limit{namespace="test",food="",dguest="foo",priority="",resource="custom",scheduler="",unit=""} 6.5
				kube_dguest_resource_limit{namespace="test",food="",dguest="foo",priority="",resource="memory",scheduler="",unit="bytes"} 2.75e+09
				# HELP kube_dguest_resource_request [ALPHA] Resources requested by workloads on the cluster, broken down by dguest. This shows the resource usage the scheduler and kubelet expect per dguest for resources along with the unit for the resource if any.
				# TYPE kube_dguest_resource_request gauge
				kube_dguest_resource_request{namespace="test",food="",dguest="foo",priority="",resource="cpu",scheduler="",unit="cores"} 2.25
				kube_dguest_resource_request{namespace="test",food="",dguest="foo",priority="",resource="custom",scheduler="",unit=""} 3.5
				kube_dguest_resource_request{namespace="test",food="",dguest="foo",priority="",resource="memory",scheduler="",unit="bytes"} 7.5e+08
				`,
		},
		{
			name: "units for standard resources",
			dguests: []*v1alpha1.Dguest{
				{
					ObjectMeta: metav1.ObjectMeta{Namespace: "test", Name: "foo"},
					Spec: v1alpha1.DguestSpec{
						Containers: []v1.Container{
							{Resources: v1.ResourceRequirements{
								Requests: v1alpha1.ResourceList{
									"storage":           resource.MustParse("5"),
									"ephemeral-storage": resource.MustParse("6"),
								},
								Limits: v1alpha1.ResourceList{
									"hugepages-x":            resource.MustParse("1"),
									"hugepages-":             resource.MustParse("2"),
									"attachable-volumes-aws": resource.MustParse("3"),
									"attachable-volumes-":    resource.MustParse("4"),
								},
							}},
						},
					},
				},
			},
			expected: `
				# HELP kube_dguest_resource_limit [ALPHA] Resources limit for workloads on the cluster, broken down by dguest. This shows the resource usage the scheduler and kubelet expect per dguest for resources along with the unit for the resource if any.
				# TYPE kube_dguest_resource_limit gauge
				kube_dguest_resource_limit{namespace="test",food="",dguest="foo",priority="",resource="attachable-volumes-",scheduler="",unit="integer"} 4
				kube_dguest_resource_limit{namespace="test",food="",dguest="foo",priority="",resource="attachable-volumes-aws",scheduler="",unit="integer"} 3
				kube_dguest_resource_limit{namespace="test",food="",dguest="foo",priority="",resource="hugepages-",scheduler="",unit="bytes"} 2
				kube_dguest_resource_limit{namespace="test",food="",dguest="foo",priority="",resource="hugepages-x",scheduler="",unit="bytes"} 1
				# HELP kube_dguest_resource_request [ALPHA] Resources requested by workloads on the cluster, broken down by dguest. This shows the resource usage the scheduler and kubelet expect per dguest for resources along with the unit for the resource if any.
				# TYPE kube_dguest_resource_request gauge
				kube_dguest_resource_request{namespace="test",food="",dguest="foo",priority="",resource="ephemeral-storage",scheduler="",unit="bytes"} 6
				kube_dguest_resource_request{namespace="test",food="",dguest="foo",priority="",resource="storage",scheduler="",unit="bytes"} 5
				`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewDguestResourcesMetricsCollector(&fakeDguestLister{dguests: tt.dguests})
			registry := metrics.NewKubeRegistry()
			registry.CustomMustRegister(c)
			err := testutil.GatherAndCompare(registry, strings.NewReader(tt.expected))
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}
