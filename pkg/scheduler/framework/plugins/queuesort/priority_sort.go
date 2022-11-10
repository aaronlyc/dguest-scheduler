package queuesort

import (
	"dguest-scheduler/pkg/scheduler/framework"
	"dguest-scheduler/pkg/scheduler/framework/plugins/names"
	"k8s.io/apimachinery/pkg/runtime"
)

// Name is the name of the plugin used in the plugin registry and configurations.
const Name = names.PrioritySort

// PrioritySort is a plugin that implements Priority based sorting.
type PrioritySort struct{}

var _ framework.QueueSortPlugin = &PrioritySort{}

// Name returns name of the plugin.
func (pl *PrioritySort) Name() string {
	return Name
}

// Less is the function used by the activeQ heap algorithm to sort dguests.
// It sorts dguests based on their priority. When priorities are equal, it uses
// DguestQueueInfo.timestamp.
func (pl *PrioritySort) Less(pInfo1, pInfo2 *framework.QueuedDguestInfo) bool {
	//p1 := corev1helpers.DguestPriority(pInfo1.Dguest)
	//p2 := corev1helpers.DguestPriority(pInfo2.Dguest)
	//return (p1 > p2) || (p1 == p2 && pInfo1.Timestamp.Before(pInfo2.Timestamp))
	return pInfo1.Timestamp.Before(pInfo2.Timestamp)
}

// New initializes a new plugin and returns it.
func New(_ runtime.Object, handle framework.Handle) (framework.Plugin, error) {
	return &PrioritySort{}, nil
}
