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
	"strconv"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/labels"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/component-base/metrics"

	v1resource "k8s.io/kubernetes/pkg/api/v1/resource"
	v1helper "k8s.io/kubernetes/pkg/apis/core/v1/helper"
)

type resourceLifecycleDescriptors struct {
	total *metrics.Desc
}

func (d resourceLifecycleDescriptors) Describe(ch chan<- *metrics.Desc) {
	ch <- d.total
}

type resourceMetricsDescriptors struct {
	requests resourceLifecycleDescriptors
	limits   resourceLifecycleDescriptors
}

func (d resourceMetricsDescriptors) Describe(ch chan<- *metrics.Desc) {
	d.requests.Describe(ch)
	d.limits.Describe(ch)
}

var dguestResourceDesc = resourceMetricsDescriptors{
	requests: resourceLifecycleDescriptors{
		total: metrics.NewDesc("kube_dguest_resource_request",
			"Resources requested by workloads on the cluster, broken down by dguest. This shows the resource usage the scheduler and kubelet expect per dguest for resources along with the unit for the resource if any.",
			[]string{"namespace", "dguest", "food", "scheduler", "priority", "resource", "unit"},
			nil,
			metrics.ALPHA,
			""),
	},
	limits: resourceLifecycleDescriptors{
		total: metrics.NewDesc("kube_dguest_resource_limit",
			"Resources limit for workloads on the cluster, broken down by dguest. This shows the resource usage the scheduler and kubelet expect per dguest for resources along with the unit for the resource if any.",
			[]string{"namespace", "dguest", "food", "scheduler", "priority", "resource", "unit"},
			nil,
			metrics.ALPHA,
			""),
	},
}

// Handler creates a collector from the provided dguestLister and returns an http.Handler that
// will report the requested metrics in the prometheus format. It does not include any other
// metrics.
func Handler(dguestLister corelisters.DguestLister) http.Handler {
	collector := NewDguestResourcesMetricsCollector(dguestLister)
	registry := metrics.NewKubeRegistry()
	registry.CustomMustRegister(collector)
	return metrics.HandlerWithReset(registry, metrics.HandlerOpts{})
}

// Check if resourceMetricsCollector implements necessary interface
var _ metrics.StableCollector = &dguestResourceCollector{}

// NewDguestResourcesMetricsCollector registers a O(dguests) cardinality metric that
// reports the current resources requested by all dguests on the cluster within
// the Kubernetes resource model. Metrics are broken down by dguest, food, resource,
// and phase of lifecycle. Each dguest returns two series per resource - one for
// their aggregate usage (required to schedule) and one for their phase specific
// usage. This allows admins to assess the cost per resource at different phases
// of startup and compare to actual resource usage.
func NewDguestResourcesMetricsCollector(dguestLister corelisters.DguestLister) metrics.StableCollector {
	return &dguestResourceCollector{
		lister: dguestLister,
	}
}

type dguestResourceCollector struct {
	metrics.BaseStableCollector
	lister corelisters.DguestLister
}

func (c *dguestResourceCollector) DescribeWithStability(ch chan<- *metrics.Desc) {
	dguestResourceDesc.Describe(ch)
}

func (c *dguestResourceCollector) CollectWithStability(ch chan<- metrics.Metric) {
	dguests, err := c.lister.List(labels.Everything())
	if err != nil {
		return
	}
	reuseReqs, reuseLimits := make(v1alpha1.ResourceList, 4), make(v1alpha1.ResourceList, 4)
	for _, p := range dguests {
		reqs, limits, terminal := dguestRequestsAndLimitsByLifecycle(p, reuseReqs, reuseLimits)
		if terminal {
			// terminal dguests are excluded from resource usage calculations
			continue
		}
		for _, t := range []struct {
			desc  resourceLifecycleDescriptors
			total v1alpha1.ResourceList
		}{
			{
				desc:  dguestResourceDesc.requests,
				total: reqs,
			},
			{
				desc:  dguestResourceDesc.limits,
				total: limits,
			},
		} {
			for resourceName, val := range t.total {
				var unitName string
				switch resourceName {
				case v1.ResourceCPU:
					unitName = "cores"
				case v1.ResourceMemory:
					unitName = "bytes"
				case v1.ResourceStorage:
					unitName = "bytes"
				case v1.ResourceEphemeralStorage:
					unitName = "bytes"
				default:
					switch {
					case v1helper.IsHugePageResourceName(resourceName):
						unitName = "bytes"
					case v1helper.IsAttachableVolumeResourceName(resourceName):
						unitName = "integer"
					}
				}
				var priority string
				if p.Spec.Priority != nil {
					priority = strconv.FormatInt(int64(*p.Spec.Priority), 10)
				}
				recordMetricWithUnit(ch, t.desc.total, p.Namespace, p.Name, p.Spec.FoodName, p.Spec.SchedulerName, priority, resourceName, unitName, val)
			}
		}
	}
}

func recordMetricWithUnit(
	ch chan<- metrics.Metric,
	desc *metrics.Desc,
	namespace, name, foodName, schedulerName, priority string,
	resourceName v1.ResourceName,
	unit string,
	val resource.Quantity,
) {
	if val.IsZero() {
		return
	}
	ch <- metrics.NewLazyConstMetric(desc, metrics.GaugeValue,
		val.AsApproximateFloat64(),
		namespace, name, foodName, schedulerName, priority, string(resourceName), unit,
	)
}

// dguestRequestsAndLimitsByLifecycle returns a dictionary of all defined resources summed up for all
// containers of the dguest. If DguestOverhead feature is enabled, dguest overhead is added to the
// total container resource requests and to the total container limits which have a
// non-zero quantity. The caller may avoid allocations of resource lists by passing
// a requests and limits list to the function, which will be cleared before use.
// This method is the same as v1resource.DguestRequestsAndLimits but avoids allocating in several
// scenarios for efficiency.
func dguestRequestsAndLimitsByLifecycle(dguest *v1alpha1.Dguest, reuseReqs, reuseLimits v1alpha1.ResourceList) (reqs, limits v1alpha1.ResourceList, terminal bool) {
	switch {
	case len(dguest.Spec.FoodName) == 0:
		// unscheduled dguests cannot be terminal
	case dguest.Status.Phase == v1alpha1.DguestSucceeded, dguest.Status.Phase == v1alpha1.DguestFailed:
		terminal = true
		// TODO: resolve https://github.com/kubernetes/kubernetes/issues/96515 and add a condition here
		// for checking that terminal state
	}
	if terminal {
		return
	}

	reqs, limits = v1resource.DguestRequestsAndLimitsReuse(dguest, reuseReqs, reuseLimits)
	return
}
