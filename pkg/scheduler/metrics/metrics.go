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

package metrics

import (
	"sync"
	"time"

	"k8s.io/component-base/metrics"
	"k8s.io/component-base/metrics/legacyregistry"
)

const (
	// SchedulerSubsystem - subsystem name used by scheduler
	SchedulerSubsystem = "scheduler"
	// Below are possible values for the operation label. Each represents a substep of e2e scheduling:

	// PrioritizingExtender - prioritizing extender operation label value
	PrioritizingExtender = "prioritizing_extender"
	// Binding - binding operation label value
	Binding = "binding"
	// E2eScheduling - e2e scheduling operation label value
)

// All the histogram based metrics have 1ms as size for the smallest bucket.
var (
	scheduleAttempts = metrics.NewCounterVec(
		&metrics.CounterOpts{
			Subsystem:      SchedulerSubsystem,
			Name:           "schedule_attempts_total",
			Help:           "Number of attempts to schedule dguests, by the result. 'unschedulable' means a dguest could not be scheduled, while 'error' means an internal scheduler problem.",
			StabilityLevel: metrics.STABLE,
		}, []string{"result", "profile"})

	e2eSchedulingLatency = metrics.NewHistogramVec(
		&metrics.HistogramOpts{
			Subsystem:         SchedulerSubsystem,
			Name:              "e2e_scheduling_duration_seconds",
			DeprecatedVersion: "1.23.0",
			Help:              "E2e scheduling latency in seconds (scheduling algorithm + binding). This metric is replaced by scheduling_attempt_duration_seconds.",
			Buckets:           metrics.ExponentialBuckets(0.001, 2, 15),
			StabilityLevel:    metrics.ALPHA,
		}, []string{"result", "profile"})
	schedulingLatency = metrics.NewHistogramVec(
		&metrics.HistogramOpts{
			Subsystem:      SchedulerSubsystem,
			Name:           "scheduling_attempt_duration_seconds",
			Help:           "Scheduling attempt latency in seconds (scheduling algorithm + binding)",
			Buckets:        metrics.ExponentialBuckets(0.001, 2, 15),
			StabilityLevel: metrics.STABLE,
		}, []string{"result", "profile"})
	SchedulingAlgorithmLatency = metrics.NewHistogram(
		&metrics.HistogramOpts{
			Subsystem:      SchedulerSubsystem,
			Name:           "scheduling_algorithm_duration_seconds",
			Help:           "Scheduling algorithm latency in seconds",
			Buckets:        metrics.ExponentialBuckets(0.001, 2, 15),
			StabilityLevel: metrics.ALPHA,
		},
	)
	PreemptionVictims = metrics.NewHistogram(
		&metrics.HistogramOpts{
			Subsystem: SchedulerSubsystem,
			Name:      "preemption_victims",
			Help:      "Number of selected preemption victims",
			// we think #victims>50 is pretty rare, therefore [50, +Inf) is considered a single bucket.
			Buckets:        metrics.LinearBuckets(5, 5, 10),
			StabilityLevel: metrics.STABLE,
		})
	PreemptionAttempts = metrics.NewCounter(
		&metrics.CounterOpts{
			Subsystem:      SchedulerSubsystem,
			Name:           "preemption_attempts_total",
			Help:           "Total preemption attempts in the cluster till now",
			StabilityLevel: metrics.STABLE,
		})
	pendingDguests = metrics.NewGaugeVec(
		&metrics.GaugeOpts{
			Subsystem:      SchedulerSubsystem,
			Name:           "pending_dguests",
			Help:           "Number of pending dguests, by the queue type. 'active' means number of dguests in activeQ; 'backoff' means number of dguests in backoffQ; 'unschedulable' means number of dguests in unschedulableDguests.",
			StabilityLevel: metrics.STABLE,
		}, []string{"queue"})
	SchedulerGoroutines = metrics.NewGaugeVec(
		&metrics.GaugeOpts{
			Subsystem:      SchedulerSubsystem,
			Name:           "scheduler_goroutines",
			Help:           "Number of running goroutines split by the work they do such as binding.",
			StabilityLevel: metrics.ALPHA,
		}, []string{"work"})

	DguestSchedulingDuration = metrics.NewHistogramVec(
		&metrics.HistogramOpts{
			Subsystem: SchedulerSubsystem,
			Name:      "dguest_scheduling_duration_seconds",
			Help:      "E2e latency for a dguest being scheduled which may include multiple scheduling attempts.",
			// Start with 10ms with the last bucket being [~88m, Inf).
			Buckets:        metrics.ExponentialBuckets(0.01, 2, 20),
			StabilityLevel: metrics.STABLE,
		},
		[]string{"attempts"})

	DguestSchedulingAttempts = metrics.NewHistogram(
		&metrics.HistogramOpts{
			Subsystem:      SchedulerSubsystem,
			Name:           "dguest_scheduling_attempts",
			Help:           "Number of attempts to successfully schedule a dguest.",
			Buckets:        metrics.ExponentialBuckets(1, 2, 5),
			StabilityLevel: metrics.STABLE,
		})

	FrameworkExtensionPointDuration = metrics.NewHistogramVec(
		&metrics.HistogramOpts{
			Subsystem: SchedulerSubsystem,
			Name:      "framework_extension_point_duration_seconds",
			Help:      "Latency for running all plugins of a specific extension point.",
			// Start with 0.1ms with the last bucket being [~200ms, Inf)
			Buckets:        metrics.ExponentialBuckets(0.0001, 2, 12),
			StabilityLevel: metrics.STABLE,
		},
		[]string{"extension_point", "status", "profile"})

	PluginExecutionDuration = metrics.NewHistogramVec(
		&metrics.HistogramOpts{
			Subsystem: SchedulerSubsystem,
			Name:      "plugin_execution_duration_seconds",
			Help:      "Duration for running a plugin at a specific extension point.",
			// Start with 0.01ms with the last bucket being [~22ms, Inf). We use a small factor (1.5)
			// so that we have better granularity since plugin latency is very sensitive.
			Buckets:        metrics.ExponentialBuckets(0.00001, 1.5, 20),
			StabilityLevel: metrics.ALPHA,
		},
		[]string{"plugin", "extension_point", "status"})

	SchedulerQueueIncomingDguests = metrics.NewCounterVec(
		&metrics.CounterOpts{
			Subsystem:      SchedulerSubsystem,
			Name:           "queue_incoming_dguests_total",
			Help:           "Number of dguests added to scheduling queues by event and queue type.",
			StabilityLevel: metrics.STABLE,
		}, []string{"queue", "event"})

	PermitWaitDuration = metrics.NewHistogramVec(
		&metrics.HistogramOpts{
			Subsystem:      SchedulerSubsystem,
			Name:           "permit_wait_duration_seconds",
			Help:           "Duration of waiting on permit.",
			Buckets:        metrics.ExponentialBuckets(0.001, 2, 15),
			StabilityLevel: metrics.ALPHA,
		},
		[]string{"result"})

	CacheSize = metrics.NewGaugeVec(
		&metrics.GaugeOpts{
			Subsystem:      SchedulerSubsystem,
			Name:           "scheduler_cache_size",
			Help:           "Number of foods, dguests, and assumed (bound) dguests in the scheduler cache.",
			StabilityLevel: metrics.ALPHA,
		}, []string{"type"})

	unschedulableReasons = metrics.NewGaugeVec(
		&metrics.GaugeOpts{
			Subsystem:      SchedulerSubsystem,
			Name:           "unschedulable_dguests",
			Help:           "The number of unschedulable dguests broken down by plugin name. A dguest will increment the gauge for all plugins that caused it to not schedule and so this metric have meaning only when broken down by plugin.",
			StabilityLevel: metrics.ALPHA,
		}, []string{"plugin", "profile"})

	metricsList = []metrics.Registerable{
		scheduleAttempts,
		e2eSchedulingLatency,
		schedulingLatency,
		SchedulingAlgorithmLatency,
		PreemptionVictims,
		PreemptionAttempts,
		pendingDguests,
		DguestSchedulingDuration,
		DguestSchedulingAttempts,
		FrameworkExtensionPointDuration,
		PluginExecutionDuration,
		SchedulerQueueIncomingDguests,
		SchedulerGoroutines,
		PermitWaitDuration,
		CacheSize,
		unschedulableReasons,
	}
)

var registerMetrics sync.Once

// Register all metrics.
func Register() {
	// Register the metrics.
	registerMetrics.Do(func() {
		RegisterMetrics(metricsList...)
		//volumebindingmetrics.RegisterVolumeSchedulingMetrics()
	})
}

// RegisterMetrics registers a list of metrics.
// This function is exported because it is intended to be used by out-of-tree plugins to register their custom metrics.
func RegisterMetrics(extraMetrics ...metrics.Registerable) {
	for _, metric := range extraMetrics {
		legacyregistry.MustRegister(metric)
	}
}

// GetGather returns the gatherer. It used by test case outside current package.
func GetGather() metrics.Gatherer {
	return legacyregistry.DefaultGatherer
}

// ActiveDguests returns the pending dguests metrics with the label active
func ActiveDguests() metrics.GaugeMetric {
	return pendingDguests.With(metrics.Labels{"queue": "active"})
}

// BackoffDguests returns the pending dguests metrics with the label backoff
func BackoffDguests() metrics.GaugeMetric {
	return pendingDguests.With(metrics.Labels{"queue": "backoff"})
}

// UnschedulableDguests returns the pending dguests metrics with the label unschedulable
func UnschedulableDguests() metrics.GaugeMetric {
	return pendingDguests.With(metrics.Labels{"queue": "unschedulable"})
}

// SinceInSeconds gets the time since the specified start in seconds.
func SinceInSeconds(start time.Time) float64 {
	return time.Since(start).Seconds()
}

func UnschedulableReason(plugin string, profile string) metrics.GaugeMetric {
	return unschedulableReasons.With(metrics.Labels{"plugin": plugin, "profile": profile})
}
