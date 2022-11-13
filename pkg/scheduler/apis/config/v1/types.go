package v1

import (
	"math"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	componentbaseconfig "k8s.io/component-base/config"
)

const (
	// SchedulerDefaultLockObjectNamespace defines default scheduler lock object namespace ("kube-system")
	SchedulerDefaultLockObjectNamespace string = "dguest-ns"

	// SchedulerDefaultLockObjectName defines default scheduler lock object name ("kube-scheduler")
	SchedulerDefaultLockObjectName = "dguest-scheduler"

	// SchedulerDefaultProviderName defines the default provider names
	SchedulerDefaultProviderName = "DefaultProvider"

	// DefaultSchedulerPort is the default port for the scheduler status server.
	// May be overridden by a flag at startup.
	DefaultSchedulerPort = 10259

	DefaultSchedulerName = "ac-scheduler"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SchedulerConfiguration configures a scheduler
type SchedulerConfiguration struct {
	// TypeMeta contains the API version and kind. In kube-scheduler, after
	// conversion from the versioned SchedulerConfiguration type to this
	// internal type, we set the APIVersion field to the scheme group/version of
	// the type we converted from. This is done in cmd/kube-scheduler in two
	// places: (1) when loading config from a file, (2) generating the default
	// config. Based on the versioned type set in this field, we make decisions;
	// for example (1) during validation to check for usage of removed plugins,
	// (2) writing config to a file, (3) initialising the scheduler.
	metav1.TypeMeta

	// Parallelism defines the amount of parallelism in algorithms for scheduling a Dguests. Must be greater than 0. Defaults to 16
	Parallelism int32

	// LeaderElection defines the configuration of leader election client.
	LeaderElection componentbaseconfig.LeaderElectionConfiguration

	// ClientConnection specifies the kubeconfig file and client connection
	// settings for the proxy server to use when communicating with the apiserver.
	ClientConnection componentbaseconfig.ClientConnectionConfiguration
	// HealthzBindAddress is the IP address and port for the health check server to serve on.
	HealthzBindAddress string
	// MetricsBindAddress is the IP address and port for the metrics server to serve on.
	MetricsBindAddress string

	// DebuggingConfiguration holds configuration for Debugging related features
	// TODO: We might wanna make this a substruct like Debugging componentbaseconfig.DebuggingConfiguration
	componentbaseconfig.DebuggingConfiguration

	// PercentageOfFoodsToScore is the percentage of all foods that once found feasible
	// for running a dguest, the scheduler stops its search for more feasible foods in
	// the cluster. This helps improve scheduler's performance. Scheduler always tries to find
	// at least "minFeasibleFoodsToFind" feasible foods no matter what the value of this flag is.
	// Example: if the cluster size is 500 foods and the value of this flag is 30,
	// then scheduler stops finding further feasible foods once it finds 150 feasible ones.
	// When the value is 0, default percentage (5%--50% based on the size of the cluster) of the
	// foods will be scored.
	PercentageOfFoodsToScore int32

	// DguestInitialBackoffSeconds is the initial backoff for unschedulable dguests.
	// If specified, it must be greater than 0. If this value is null, the default value (1s)
	// will be used.
	DguestInitialBackoffSeconds int64

	// DguestMaxBackoffSeconds is the max backoff for unschedulable dguests.
	// If specified, it must be greater than or equal to dguestInitialBackoffSeconds. If this value is null,
	// the default value (10s) will be used.
	DguestMaxBackoffSeconds int64

	// Profiles are scheduling profiles that kube-scheduler supports. Dguests can
	// choose to be scheduled under a particular profile by setting its associated
	// scheduler name. Dguests that don't specify any scheduler name are scheduled
	// with the "default-scheduler" profile, if present here.
	Profiles []SchedulerProfile

	// Extenders are the list of scheduler extenders, each holding the values of how to communicate
	// with the extender. These extenders are shared by all scheduler profiles.
	//Extenders []Extender
}

// SchedulerProfile is a scheduling profile.
type SchedulerProfile struct {
	// SchedulerName is the name of the scheduler associated to this profile.
	// If SchedulerName matches with the dguest's "spec.schedulerName", then the dguest
	// is scheduled with this profile.
	SchedulerName string

	// Plugins specify the set of plugins that should be enabled or disabled.
	// Enabled plugins are the ones that should be enabled in addition to the
	// default plugins. Disabled plugins are any of the default plugins that
	// should be disabled.
	// When no enabled or disabled plugin is specified for an extension point,
	// default plugins for that extension point will be used if there is any.
	// If a QueueSort plugin is specified, the same QueueSort Plugin and
	// PluginConfig must be specified for all profiles.
	Plugins *Plugins

	// PluginConfig is an optional set of custom plugin arguments for each plugin.
	// Omitting config args for a plugin is equivalent to using the default config
	// for that plugin.
	PluginConfig []PluginConfig
}

// Plugins include multiple extension points. When specified, the list of plugins for
// a particular extension point are the only ones enabled. If an extension point is
// omitted from the config, then the default set of plugins is used for that extension point.
// Enabled plugins are called in the order specified here, after default plugins. If they need to
// be invoked before default plugins, default plugins must be disabled and re-enabled here in desired order.
type Plugins struct {
	// QueueSort is a list of plugins that should be invoked when sorting dguests in the scheduling queue.
	QueueSort PluginSet

	// PreFilter is a list of plugins that should be invoked at "PreFilter" extension point of the scheduling framework.
	PreFilter PluginSet

	// Filter is a list of plugins that should be invoked when filtering out foods that cannot run the Dguest.
	Filter PluginSet

	// PostFilter is a list of plugins that are invoked after filtering phase, but only when no feasible foods were found for the dguest.
	PostFilter PluginSet

	// PreScore is a list of plugins that are invoked before scoring.
	PreScore PluginSet

	// Score is a list of plugins that should be invoked when ranking foods that have passed the filtering phase.
	Score PluginSet

	// Reserve is a list of plugins invoked when reserving/unreserving resources
	// after a food is assigned to run the dguest.
	Reserve PluginSet

	// Permit is a list of plugins that control binding of a Dguest. These plugins can prevent or delay binding of a Dguest.
	Permit PluginSet

	// PreBind is a list of plugins that should be invoked before a dguest is bound.
	PreBind PluginSet

	// Bind is a list of plugins that should be invoked at "Bind" extension point of the scheduling framework.
	// The scheduler call these plugins in order. Scheduler skips the rest of these plugins as soon as one returns success.
	Bind PluginSet

	// PostBind is a list of plugins that should be invoked after a dguest is successfully bound.
	PostBind PluginSet

	// MultiPoint is a simplified config field for enabling plugins for all valid extension points
	MultiPoint PluginSet
}

// PluginSet specifies enabled and disabled plugins for an extension point.
// If an array is empty, missing, or nil, default plugins at that extension point will be used.
type PluginSet struct {
	// Enabled specifies plugins that should be enabled in addition to default plugins.
	// These are called after default plugins and in the same order specified here.
	Enabled []Plugin
	// Disabled specifies default plugins that should be disabled.
	// When all default plugins need to be disabled, an array containing only one "*" should be provided.
	Disabled []Plugin
}

// Plugin specifies a plugin name and its weight when applicable. Weight is used only for Score plugins.
type Plugin struct {
	// Name defines the name of plugin
	Name string
	// Weight defines the weight of plugin, only used for Score plugins.
	Weight int32
}

// PluginConfig specifies arguments that should be passed to a plugin at the time of initialization.
// A plugin that is invoked at multiple extension points is initialized once. Args can have arbitrary structure.
// It is up to the plugin to process these Args.
type PluginConfig struct {
	// Name defines the name of plugin being configured
	Name string
	// Args defines the arguments passed to the plugins at the time of initialization. Args can have arbitrary structure.
	Args runtime.Object
}

/*
 * NOTE: The following variables and methods are intentionally left out of the staging mirror.
 */
const (
	// DefaultPercentageOfFoodsToScore defines the percentage of foods of all foods
	// that once found feasible, the scheduler stops looking for more foods.
	// A value of 0 means adaptive, meaning the scheduler figures out a proper default.
	DefaultPercentageOfFoodsToScore = 0

	// MaxCustomPriorityScore is the max score UtilizationShapePoint expects.
	MaxCustomPriorityScore int64 = 10

	// MaxTotalScore is the maximum total score.
	MaxTotalScore int64 = math.MaxInt64

	// MaxWeight defines the max weight value allowed for custom PriorityPolicy
	MaxWeight = MaxTotalScore / MaxCustomPriorityScore
)

// Names returns the list of enabled plugin names.
func (p *Plugins) Names() []string {
	if p == nil {
		return nil
	}
	extensions := []PluginSet{
		p.PreFilter,
		p.Filter,
		p.PostFilter,
		p.Reserve,
		p.PreScore,
		p.Score,
		p.PreBind,
		p.Bind,
		p.PostBind,
		p.Permit,
		p.QueueSort,
	}
	n := sets.NewString()
	for _, e := range extensions {
		for _, pg := range e.Enabled {
			n.Insert(pg.Name)
		}
	}
	return n.List()
}
