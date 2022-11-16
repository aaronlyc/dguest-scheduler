package plugins

import (
	"dguest-scheduler/pkg/scheduler/framework/plugins/defaultbinder"
	"dguest-scheduler/pkg/scheduler/framework/plugins/nodeavailability"
	"dguest-scheduler/pkg/scheduler/framework/plugins/queuesort"
	"dguest-scheduler/pkg/scheduler/framework/runtime"
)

// NewInTreeRegistry builds the registry with all the in-tree plugins.
// A scheduler that runs out of tree plugins can register additional plugins
// through the WithFrameworkOutOfTreeRegistry option.
func NewInTreeRegistry() runtime.Registry {
	return runtime.Registry{
		nodeavailability.Name: nodeavailability.New,
		queuesort.Name:        queuesort.New,
		defaultbinder.Name:    defaultbinder.New,
	}
}
