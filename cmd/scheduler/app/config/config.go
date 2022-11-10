package config

import (
	"dguest-scheduler/pkg/generated/clientset/versioned"
	"dguest-scheduler/pkg/generated/informers/externalversions"
	kubeschedulerconfig "dguest-scheduler/pkg/scheduler/apis/config/v1"
	"time"

	apiserver "k8s.io/apiserver/pkg/server"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/events"
	"k8s.io/client-go/tools/leaderelection"
)

// Config has all the context to run a Scheduler
type Config struct {
	// ComponentConfig is the scheduler server's configuration object.
	ComponentConfig kubeschedulerconfig.SchedulerConfiguration

	// LoopbackClientConfig is a config for a privileged loopback connection
	LoopbackClientConfig *restclient.Config

	Authentication apiserver.AuthenticationInfo
	Authorization  apiserver.AuthorizationInfo
	SecureServing  *apiserver.SecureServingInfo

	KubeConfig               *restclient.Config
	SchedulerClientSet       versioned.Interface
	SchedulerInformerFactory externalversions.SharedInformerFactory

	//nolint:staticcheck // SA1019 this deprecated field still needs to be used for now. It will be removed once the migration is done.
	EventBroadcaster events.EventBroadcasterAdapter

	// LeaderElection is optional.
	LeaderElection *leaderelection.LeaderElectionConfig

	// DguestMaxInUnschedulableDguestsDuration is the maximum time a dguest can stay in
	// unschedulableDguests. If a dguest stays in unschedulableDguests for longer than this
	// value, the dguest will be moved from unschedulableDguests to backoffQ or activeQ.
	// If this value is empty, the default value (5min) will be used.
	DguestMaxInUnschedulableDguestsDuration time.Duration
}

type completedConfig struct {
	*Config
}

// CompletedConfig same as Config, just to swap private object.
type CompletedConfig struct {
	// Embed a private pointer that cannot be instantiated outside of this package.
	*completedConfig
}

// Complete fills in any fields not set that are required to have valid data. It's mutating the receiver.
func (c *Config) Complete() CompletedConfig {
	cc := completedConfig{c}

	apiserver.AuthorizeClientBearerToken(c.LoopbackClientConfig, &c.Authentication, &c.Authorization)

	return CompletedConfig{&cc}
}
