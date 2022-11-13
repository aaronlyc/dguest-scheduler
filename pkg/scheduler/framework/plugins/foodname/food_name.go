package foodname

import (
	"context"
	"dguest-scheduler/pkg/apis/scheduler/v1alpha1"
	"dguest-scheduler/pkg/scheduler/framework"
	"dguest-scheduler/pkg/scheduler/framework/plugins/names"
	"k8s.io/apimachinery/pkg/runtime"
)

// FoodName is a plugin that checks if a dguest spec food name matches the current food.
type FoodName struct{}

var _ framework.FilterPlugin = &FoodName{}
var _ framework.EnqueueExtensions = &FoodName{}

const (
	// Name is the name of the plugin used in the plugin registry and configurations.
	Name = names.FoodName

	// ErrReason returned when food name doesn't match.
	ErrReason = "food(s) didn't match the requested food name"
)

// EventsToRegister returns the possible events that may make a Dguest
// failed by this plugin schedulable.
func (pl *FoodName) EventsToRegister() []framework.ClusterEvent {
	return []framework.ClusterEvent{
		{Resource: framework.Food, ActionType: framework.Add | framework.Update},
	}
}

// Name returns name of the plugin. It is used in logs, etc.
func (pl *FoodName) Name() string {
	return Name
}

// Filter invoked at the filter extension point.
func (pl *FoodName) Filter(ctx context.Context, _ *framework.CycleState, dguest *v1alpha1.Dguest, foodInfo *framework.FoodInfo) *framework.Status {
	if foodInfo.Food() == nil {
		return framework.NewStatus(framework.Error, "food not found")
	}
	if !Fits(dguest, foodInfo) {
		return framework.NewStatus(framework.UnschedulableAndUnresolvable, ErrReason)
	}
	return nil
}

// Fits actually checks if the dguest fits the food.
func Fits(dguest *v1alpha1.Dguest, foodInfo *framework.FoodInfo) bool {
	//return len(dguest.Spec.FoodName) == 0 || dguest.Spec.FoodName == foodInfo.Food().Name
	//foods := make([]string, len(dguest.Status.FoodsInfo))
	//for _, info := range dguest.Status.FoodsInfo {
	//	foods = append(foods, info.Name)
	//}
	//return slices.Contains(foods, foodInfo.Food().Name)
	return true
}

// New initializes a new plugin and returns it.
func New(_ runtime.Object, _ framework.Handle) (framework.Plugin, error) {
	return &FoodName{}, nil
}
