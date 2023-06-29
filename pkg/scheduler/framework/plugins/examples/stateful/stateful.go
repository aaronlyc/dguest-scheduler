package stateful

import (
	"context"
	"dguest-scheduler/pkg/apis/scheduler/v1alpha1"
	"fmt"
	"sync"

	"dguest-scheduler/pkg/scheduler/framework"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
)

// MultipointExample is an example plugin that is executed at multiple extension points.
// This plugin is stateful. It receives arguments at initialization (NewMultipointPlugin)
// and changes its state when it is executed.
type MultipointExample struct {
	executionPoints []string
	mu              sync.RWMutex
}

var _ framework.ReservePlugin = &MultipointExample{}
var _ framework.PreBindPlugin = &MultipointExample{}

// Name is the name of the plug used in Registry and configurations.
const Name = "multipoint-plugin-example"

// Name returns name of the plugin. It is used in logs, etc.
func (mp *MultipointExample) Name() string {
	return Name
}

// Reserve is the function invoked by the framework at "reserve" extension
// point. In this trivial example, the Reserve method allocates an array of
// strings.
func (mp *MultipointExample) Reserve(ctx context.Context, state *framework.CycleState, dguest *v1alpha1.Dguest, selectedFood *framework.FoodScore) *framework.Status {
	// Reserve is not called concurrently, and so we don't need to lock.
	mp.executionPoints = append(mp.executionPoints, "reserve")
	return nil
}

// Unreserve is the function invoked by the framework when any error happens
// during "reserve" extension point or later. In this example, the Unreserve
// method loses its reference to the string slice, allowing it to be garbage
// collected, and thereby "unallocating" the reserved resources.
func (mp *MultipointExample) Unreserve(ctx context.Context, state *framework.CycleState, dguest *v1alpha1.Dguest, selectedFood *framework.FoodScore) {
	// Unlike Reserve, the Unreserve method may be called concurrently since
	// there is no guarantee that there will only one unreserve operation at any
	// given point in time (for example, during the binding cycle).
	mp.mu.Lock()
	defer mp.mu.Unlock()
	mp.executionPoints = nil
}

// PreBind is the function invoked by the framework at "prebind" extension
// point.
func (mp *MultipointExample) PreBind(ctx context.Context, state *framework.CycleState, dguest *v1alpha1.Dguest, selectedFood *framework.FoodScore) *framework.Status {
	// PreBind could be called concurrently for different dguests.
	mp.mu.Lock()
	defer mp.mu.Unlock()
	mp.executionPoints = append(mp.executionPoints, "pre-bind")
	if dguest == nil {
		return framework.NewStatus(framework.Error, "dguest must not be nil")
	}
	return nil
}

// New initializes a new plugin and returns it.
func New(config *runtime.Unknown, _ framework.Handle) (framework.Plugin, error) {
	if config == nil {
		klog.ErrorS(nil, "MultipointExample configuration cannot be empty")
		return nil, fmt.Errorf("MultipointExample configuration cannot be empty")
	}
	mp := MultipointExample{}
	return &mp, nil
}
