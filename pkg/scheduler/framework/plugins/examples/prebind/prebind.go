package prebind

import (
	"context"
	"dguest-scheduler/pkg/apis/scheduler/v1alpha1"
	"dguest-scheduler/pkg/scheduler/framework"

	"k8s.io/apimachinery/pkg/runtime"
)

// StatelessPreBindExample is an example of a simple plugin that has no state
// and implements only one hook for prebind.
type StatelessPreBindExample struct{}

var _ framework.PreBindPlugin = StatelessPreBindExample{}

// Name is the name of the plugin used in Registry and configurations.
const Name = "stateless-prebind-plugin-example"

// Name returns name of the plugin. It is used in logs, etc.
func (sr StatelessPreBindExample) Name() string {
	return Name
}

// PreBind is the functions invoked by the framework at "prebind" extension point.
func (sr StatelessPreBindExample) PreBind(ctx context.Context, state *framework.CycleState, dguest *v1alpha1.Dguest, selectedFood *v1alpha1.FoodInfoBase) *framework.Status {
	if dguest == nil {
		return framework.NewStatus(framework.Error, "dguest cannot be nil")
	}
	if dguest.Namespace != "foo" {
		return framework.NewStatus(framework.Unschedulable, "only dguests from 'foo' namespace are allowed")
	}
	return nil
}

// New initializes a new plugin and returns it.
func New(_ *runtime.Unknown, _ framework.Handle) (framework.Plugin, error) {
	return &StatelessPreBindExample{}, nil
}
