package nodeavailability

import (
	"context"
	apidguest "dguest-scheduler/pkg/api/dguest"
	"dguest-scheduler/pkg/apis/scheduler/v1alpha1"
	"dguest-scheduler/pkg/generated/clientset/versioned"
	"dguest-scheduler/pkg/scheduler/framework"
	"dguest-scheduler/pkg/scheduler/framework/plugins/names"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// NodeAvailability is a plugin that checks if a dguest spec food name matches the current food.
type NodeAvailability struct {
	schedulerClientSet versioned.Interface
}

var _ framework.FilterPlugin = &NodeAvailability{}

var _ framework.EnqueueExtensions = &NodeAvailability{}
var _ framework.ScorePlugin = &NodeAvailability{}

const (
	// Name is the name of the plugin used in the plugin registry and configurations.
	Name = names.NodeAvailability

	ErrReasonSameName = "The name %s has already been selected"
	ErrReasonSameNode = "The node %s has already been selected"
)

// EventsToRegister returns the possible events that may make a Dguest
// failed by this plugin schedulable.
func (pl *NodeAvailability) EventsToRegister() []framework.ClusterEvent {
	return []framework.ClusterEvent{
		{Resource: framework.Dguest, ActionType: framework.Add | framework.Update},
		{Resource: framework.Food, ActionType: framework.Add | framework.Update},
	}
}

// Name returns name of the plugin. It is used in logs, etc.
func (pl *NodeAvailability) Name() string {
	return Name
}

// Filter invoked at the filter extension point.
func (pl *NodeAvailability) Filter(_ context.Context, _ *framework.CycleState, dguest *v1alpha1.Dguest, foodInfo *framework.FoodInfo) *framework.Status {
	if foodInfo.Food() == nil {
		return framework.NewStatus(framework.Error, "food not found")
	}

	return pl.FitStatus(dguest, foodInfo)
}

// FitStatus actually checks if the dguest fits the food.
func (pl *NodeAvailability) FitStatus(dguest *v1alpha1.Dguest, foodInfo *framework.FoodInfo) *framework.Status {
	cuisineVersion := apidguest.FoodCuisineVersionKey(foodInfo.Food())
	if dguest.Status.FoodsInfo != nil {
		foods, ok := dguest.Status.FoodsInfo[cuisineVersion]
		if ok {
			for _, food := range foods {
				// the same node, should not select
				if food.Name == foodInfo.Food().Name {
					return framework.NewStatus(framework.UnschedulableAndUnresolvable, fmt.Sprintf(ErrReasonSameName, food.Name))
				}
			}
		}
	}
	return nil
}

func (pl *NodeAvailability) Score(ctx context.Context, _ *framework.CycleState, dguest *v1alpha1.Dguest, selectedFood *v1alpha1.FoodInfoBase) (int64, *framework.Status) {
	if dguest.Status.FoodsInfo != nil {
		foods, ok := dguest.Status.FoodsInfo[selectedFood.CuisineVersion]
		if ok {
			getselectFood, err := pl.schedulerClientSet.SchedulerV1alpha1().Foods(selectedFood.Namespace).Get(ctx, selectedFood.Name, metav1.GetOptions{})
			if err != nil {
				return 0, framework.NewStatus(framework.Error, fmt.Sprintf("The food %s not found", getselectFood.Name))
			}

			for _, food := range foods {
				getFood, err := pl.schedulerClientSet.SchedulerV1alpha1().Foods(food.Namespace).Get(ctx, food.Name, metav1.GetOptions{})
				if err != nil {
					return 0, framework.NewStatus(framework.Error, fmt.Sprintf("The food %s not found", food.Name))
				}
				if getFood.Status.FoodInfo == nil || getselectFood.Status.FoodInfo == nil {
					return 0, framework.NewStatus(framework.Error, fmt.Sprintf("The food %s stauts.FoodInfo should not nil", food.Name))
				}
				// the same node, should not select
				if getFood.Status.FoodInfo.CoreRunNode == getselectFood.Status.FoodInfo.CoreRunNode {
					return 0, framework.NewStatus(framework.UnschedulableAndUnresolvable, fmt.Sprintf(ErrReasonSameNode, getFood.Status.FoodInfo.CoreRunNode))
				}
			}
		}
	}
	return 10, nil
}

func (pl *NodeAvailability) ScoreExtensions() framework.ScoreExtensions {
	return nil
}

// New initializes a new plugin and returns it.
func New(_ runtime.Object, f framework.Handle) (framework.Plugin, error) {
	return &NodeAvailability{
		schedulerClientSet: f.SchedulerClientSet(),
	}, nil
}
