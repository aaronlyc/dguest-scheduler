package v1

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apiserver/pkg/util/feature"
	componentbaseconfigv1alpha1 "k8s.io/component-base/config/v1alpha1"
	"k8s.io/kubernetes/pkg/features"
)

var defaultResourceSpec = []ResourceSpec{
	{Name: string(v1.ResourceCPU), Weight: 1},
	{Name: string(v1.ResourceMemory), Weight: 1},
}

//func addDefaultingFuncs(scheme *runtime.Scheme) error {
//	return RegisterDefaults(scheme)
//}

func pluginsNames(p *Plugins) []string {
	if p == nil {
		return nil
	}
	extensions := []PluginSet{
		p.MultiPoint,
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

func setDefaults_KubeSchedulerProfile(prof *SchedulerProfile) {
	// Set default plugins.
	prof.Plugins = mergePlugins(getDefaultPlugins(), prof.Plugins)
	// Set default plugin configs.
	scheme := GetPluginArgConversionScheme()
	existingConfigs := sets.NewString()
	for j := range prof.PluginConfig {
		existingConfigs.Insert(prof.PluginConfig[j].Name)
		args := prof.PluginConfig[j].Args
		if _, isUnknown := args.(*runtime.Unknown); isUnknown {
			continue
		}
		scheme.Default(args)
	}

	// Append default configs for plugins that didn't have one explicitly set.
	for _, name := range pluginsNames(prof.Plugins) {
		if existingConfigs.Has(name) {
			continue
		}
		gvk := SchemeGroupVersion.WithKind(name + "Args")
		args, err := scheme.New(gvk)
		if err != nil {
			// This plugin is out-of-tree or doesn't require configuration.
			continue
		}
		scheme.Default(args)
		args.GetObjectKind().SetGroupVersionKind(gvk)
		prof.PluginConfig = append(prof.PluginConfig, PluginConfig{
			Name: name,
			Args: args,
		})
	}
}

// SetDefaults_KubeSchedulerConfiguration sets additional defaults
func SetDefaults_KubeSchedulerConfiguration(obj *SchedulerConfiguration) {
	if obj.Parallelism == 0 {
		obj.Parallelism = 16
	}

	if len(obj.Profiles) == 0 {
		obj.Profiles = append(obj.Profiles, SchedulerProfile{})
	}
	// Only apply a default scheduler name when there is a single profile.
	// Validation will ensure that every profile has a non-empty unique name.
	if len(obj.Profiles) == 1 && obj.Profiles[0].SchedulerName == "" {
		obj.Profiles[0].SchedulerName = DefaultSchedulerName
	}

	// Add the default set of plugins and apply the configuration.
	for i := range obj.Profiles {
		prof := &obj.Profiles[i]
		setDefaults_KubeSchedulerProfile(prof)
	}

	if obj.PercentageOfFoodsToScore == 0 {
		percentageOfFoodsToScore := int32(DefaultPercentageOfFoodsToScore)
		obj.PercentageOfFoodsToScore = percentageOfFoodsToScore
	}

	if len(obj.LeaderElection.ResourceLock) == 0 {
		// Use lease-based leader election to reduce cost.
		// We migrated for EndpointsLease lock in 1.17 and starting in 1.20 we
		// migrated to Lease lock.
		obj.LeaderElection.ResourceLock = "leases"
	}
	if len(obj.LeaderElection.ResourceNamespace) == 0 {
		obj.LeaderElection.ResourceNamespace = SchedulerDefaultLockObjectNamespace
	}
	if len(obj.LeaderElection.ResourceName) == 0 {
		obj.LeaderElection.ResourceName = SchedulerDefaultLockObjectName
	}

	if len(obj.ClientConnection.ContentType) == 0 {
		obj.ClientConnection.ContentType = ""
	}
	// Scheduler has an opinion about QPS/Burst, setting specific defaults for itself, instead of generic settings.
	if obj.ClientConnection.QPS == 0.0 {
		obj.ClientConnection.QPS = 50.0
	}
	if obj.ClientConnection.Burst == 0 {
		obj.ClientConnection.Burst = 100
	}

	// Use the default LeaderElectionConfiguration options
	componentbaseconfigv1alpha1.RecommendedDefaultLeaderElectionConfiguration(&componentbaseconfigv1alpha1.LeaderElectionConfiguration{
		LeaderElect:       &obj.LeaderElection.LeaderElect,
		LeaseDuration:     obj.LeaderElection.LeaseDuration,
		RenewDeadline:     obj.LeaderElection.RenewDeadline,
		RetryPeriod:       obj.LeaderElection.RetryPeriod,
		ResourceLock:      obj.LeaderElection.ResourceLock,
		ResourceName:      obj.LeaderElection.ResourceName,
		ResourceNamespace: obj.LeaderElection.ResourceNamespace,
	})

	if obj.DguestInitialBackoffSeconds == 0 {
		obj.DguestInitialBackoffSeconds = 1
	}

	if obj.DguestMaxBackoffSeconds == 0 {
		obj.DguestMaxBackoffSeconds = 10
	}

	// Enable profiling by default in the scheduler
	if obj.EnableProfiling == false {
		obj.EnableProfiling = true
	}

	// Enable contention profiling by default if profiling is enabled
	if obj.EnableProfiling && obj.EnableContentionProfiling == false {
		obj.EnableContentionProfiling = true
	}
}

func SetDefaults_DefaultPreemptionArgs(obj *DefaultPreemptionArgs) {
	if obj.MinCandidateFoodsPercentage == 0 {
		obj.MinCandidateFoodsPercentage = 10
	}
	if obj.MinCandidateFoodsAbsolute == 0 {
		obj.MinCandidateFoodsAbsolute = 100
	}
}

func SetDefaults_InterDguestAffinityArgs(obj *InterDguestAffinityArgs) {
	if obj.HardDguestAffinityWeight == 0 {
		obj.HardDguestAffinityWeight = 1
	}
}

func SetDefaults_VolumeBindingArgs(obj *VolumeBindingArgs) {
	if obj.BindTimeoutSeconds == 0 {
		obj.BindTimeoutSeconds = 600
	}
	if len(obj.Shape) == 0 && feature.DefaultFeatureGate.Enabled(features.VolumeCapacityPriority) {
		obj.Shape = []UtilizationShapePoint{
			{
				Utilization: 0,
				Score:       0,
			},
			{
				Utilization: 100,
				Score:       int32(MaxCustomPriorityScore),
			},
		}
	}
}

func SetDefaults_FoodResourcesBalancedAllocationArgs(obj *FoodResourcesBalancedAllocationArgs) {
	if len(obj.Resources) == 0 {
		obj.Resources = defaultResourceSpec
		return
	}
	// If the weight is not set or it is explicitly set to 0, then apply the default weight(1) instead.
	for i := range obj.Resources {
		if obj.Resources[i].Weight == 0 {
			obj.Resources[i].Weight = 1
		}
	}
}

func SetDefaults_DguestTopologySpreadArgs(obj *DguestTopologySpreadArgs) {
	if obj.DefaultingType == "" {
		obj.DefaultingType = SystemDefaulting
	}
}

func SetDefaults_FoodResourcesFitArgs(obj *FoodResourcesFitArgs) {
	if obj.ScoringStrategy == nil {
		obj.ScoringStrategy = &ScoringStrategy{
			Type:      ScoringStrategyType(LeastAllocated),
			Resources: defaultResourceSpec,
		}
	}
	if len(obj.ScoringStrategy.Resources) == 0 {
		// If no resources specified, use the default set.
		obj.ScoringStrategy.Resources = append(obj.ScoringStrategy.Resources, defaultResourceSpec...)
	}
	for i := range obj.ScoringStrategy.Resources {
		if obj.ScoringStrategy.Resources[i].Weight == 0 {
			obj.ScoringStrategy.Resources[i].Weight = 1
		}
	}
}
