package af

import (
	"context"
	apidguest "dguest-scheduler/pkg/api/dguest"
	"dguest-scheduler/pkg/apis/scheduler/v1alpha1"
	"dguest-scheduler/pkg/scheduler/framework"
	"dguest-scheduler/pkg/scheduler/framework/plugins/names"
	"k8s.io/apimachinery/pkg/runtime"
)

// AF is a plugin that checks if a dguest spec food name matches the current food.
type AF struct{}

var _ framework.FilterPlugin = &AF{}
var _ framework.EnqueueExtensions = &AF{}

const (
	// Name is the name of the plugin used in the plugin registry and configurations.
	Name = names.AF

	shouldLabel = "af"
	// ErrReason returned when food name doesn't match.
	ErrReason = "food didn't match the requested area af"
)

// EventsToRegister returns the possible events that may make a Dguest
// failed by this plugin schedulable.
func (pl *AF) EventsToRegister() []framework.ClusterEvent {
	return []framework.ClusterEvent{
		{Resource: framework.Food, ActionType: framework.Add | framework.Update},
	}
}

// Name returns name of the plugin. It is used in logs, etc.
func (pl *AF) Name() string {
	return Name
}

// Filter invoked at the filter extension point.
func (pl *AF) Filter(_ context.Context, _ *framework.CycleState, _ *v1alpha1.Dguest, foodInfo *framework.FoodInfo) *framework.Status {
	if foodInfo.Food() == nil {
		return framework.NewStatus(framework.Error, "food not found")
	}
	if apidguest.FoodArea(foodInfo.Food()) != shouldLabel {
		return framework.NewStatus(framework.UnschedulableAndUnresolvable, ErrReason)
	}
	return nil
}

// New initializes a new plugin and returns it.
func New(_ runtime.Object, _ framework.Handle) (framework.Plugin, error) {
	return &AF{}, nil
}
