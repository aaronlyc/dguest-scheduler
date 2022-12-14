package multipoint

import (
	"context"
	"dguest-scheduler/pkg/apis/scheduler/v1alpha1"

	"dguest-scheduler/pkg/scheduler/framework"

	"k8s.io/apimachinery/pkg/runtime"
)

// CommunicatingPlugin is an example of a plugin that implements two
// extension points. It communicates through state with another function.
type CommunicatingPlugin struct{}

var _ framework.ReservePlugin = CommunicatingPlugin{}
var _ framework.PreBindPlugin = CommunicatingPlugin{}

// Name is the name of the plugin used in Registry and configurations.
const Name = "multipoint-communicating-plugin"

// Name returns name of the plugin. It is used in logs, etc.
func (mc CommunicatingPlugin) Name() string {
	return Name
}

type stateData struct {
	data string
}

func (s *stateData) Clone() framework.StateData {
	copy := &stateData{
		data: s.data,
	}
	return copy
}

// Reserve is the function invoked by the framework at "reserve" extension point.
func (mc CommunicatingPlugin) Reserve(ctx context.Context, state *framework.CycleState, dguest *v1alpha1.Dguest, selectedFood *v1alpha1.FoodInfoBase) *framework.Status {
	if dguest == nil {
		return framework.NewStatus(framework.Error, "dguest cannot be nil")
	}
	if dguest.Name == "my-test-dguest" {
		state.Write(framework.StateKey(dguest.Name), &stateData{data: "never bind"})
	}
	return nil
}

// Unreserve is the function invoked by the framework when any error happens
// during "reserve" extension point or later.
func (mc CommunicatingPlugin) Unreserve(ctx context.Context, state *framework.CycleState, dguest *v1alpha1.Dguest, selectedFood *v1alpha1.FoodInfoBase) {
	if dguest.Name == "my-test-dguest" {
		// The dguest is at the end of its lifecycle -- let's clean up the allocated
		// resources. In this case, our clean up is simply deleting the key written
		// in the Reserve operation.
		state.Delete(framework.StateKey(dguest.Name))
	}
}

// PreBind is the function invoked by the framework at "prebind" extension point.
func (mc CommunicatingPlugin) PreBind(ctx context.Context, state *framework.CycleState, dguest *v1alpha1.Dguest, selectedFood *v1alpha1.FoodInfoBase) *framework.Status {
	if dguest == nil {
		return framework.NewStatus(framework.Error, "dguest cannot be nil")
	}
	if v, e := state.Read(framework.StateKey(dguest.Name)); e == nil {
		if value, ok := v.(*stateData); ok && value.data == "never bind" {
			return framework.NewStatus(framework.Unschedulable, "dguest is not permitted")
		}
	}
	return nil
}

// New initializes a new plugin and returns it.
func New(_ *runtime.Unknown, _ framework.Handle) (framework.Plugin, error) {
	return &CommunicatingPlugin{}, nil
}
