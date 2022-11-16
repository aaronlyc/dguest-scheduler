package metrics

import (
	"k8s.io/component-base/metrics"
)

// MetricRecorder represents a metric recorder which takes action when the
// metric Inc(), Dec() and Clear()
type MetricRecorder interface {
	Inc()
	Dec()
	Clear()
}

var _ MetricRecorder = &PendingDguestsRecorder{}

// PendingDguestsRecorder is an implementation of MetricRecorder
type PendingDguestsRecorder struct {
	recorder metrics.GaugeMetric
}

// NewActiveDguestsRecorder returns ActiveDguests in a Prometheus metric fashion
func NewActiveDguestsRecorder() *PendingDguestsRecorder {
	return &PendingDguestsRecorder{
		recorder: ActiveDguests(),
	}
}

// NewUnschedulableDguestsRecorder returns UnschedulableDguests in a Prometheus metric fashion
func NewUnschedulableDguestsRecorder() *PendingDguestsRecorder {
	return &PendingDguestsRecorder{
		recorder: UnschedulableDguests(),
	}
}

// NewBackoffDguestsRecorder returns BackoffDguests in a Prometheus metric fashion
func NewBackoffDguestsRecorder() *PendingDguestsRecorder {
	return &PendingDguestsRecorder{
		recorder: BackoffDguests(),
	}
}

// Inc increases a metric counter by 1, in an atomic way
func (r *PendingDguestsRecorder) Inc() {
	r.recorder.Inc()
}

// Dec decreases a metric counter by 1, in an atomic way
func (r *PendingDguestsRecorder) Dec() {
	r.recorder.Dec()
}

// Clear set a metric counter to 0, in an atomic way
func (r *PendingDguestsRecorder) Clear() {
	r.recorder.Set(float64(0))
}
