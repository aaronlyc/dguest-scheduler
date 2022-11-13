package defaults

import (
	"dguest-scheduler/pkg/scheduler/apis/config/v1"
	"dguest-scheduler/pkg/scheduler/framework/plugins/names"
)

// PluginsV1beta2 default set of v1beta2 plugins.
var PluginsV1beta2 = &v1.Plugins{
	QueueSort: v1.PluginSet{
		Enabled: []v1.Plugin{
			{Name: names.PrioritySort},
		},
	},
	PreFilter: v1.PluginSet{
		Enabled: []v1.Plugin{
			{Name: names.FoodResourcesFit},
			{Name: names.FoodPorts},
			{Name: names.DguestTopologySpread},
			{Name: names.FoodAffinity},
		},
	},
	Filter: v1.PluginSet{
		Enabled: []v1.Plugin{
			{Name: names.FoodUnschedulable},
			{Name: names.FoodName},
			{Name: names.TaintToleration},
			{Name: names.FoodAffinity},
			{Name: names.FoodResourcesFit},
			{Name: names.DguestTopologySpread},
		},
	},
	PostFilter: v1.PluginSet{
		Enabled: []v1.Plugin{
			{Name: names.DefaultPreemption},
		},
	},
	PreScore: v1.PluginSet{
		Enabled: []v1.Plugin{
			{Name: names.DguestTopologySpread},
			{Name: names.TaintToleration},
			{Name: names.FoodAffinity},
		},
	},
	Score: v1.PluginSet{
		Enabled: []v1.Plugin{
			{Name: names.FoodResourcesBalancedAllocation, Weight: 1},
			{Name: names.FoodResourcesFit, Weight: 1},
			{Name: names.FoodAffinity, Weight: 1},
			// Weight is doubled because:
			// - This is a score coming from user preference.
			// - It makes its signal comparable to FoodResourcesLeastAllocated.
			{Name: names.DguestTopologySpread, Weight: 2},
			{Name: names.TaintToleration, Weight: 1},
		},
	},
	Reserve: v1.PluginSet{},
	PreBind: v1.PluginSet{},
	Bind: v1.PluginSet{
		Enabled: []v1.Plugin{
			{Name: names.DefaultBinder},
		},
	},
}

// PluginConfigsV1beta2 default plugin configurations. This could get versioned, but since
// all available versions produce the same defaults, we just have one for now.
var PluginConfigsV1beta2 = []v1.PluginConfig{
	{
		Name: "DefaultPreemption",
		Args: &v1.DefaultPreemptionArgs{
			MinCandidateFoodsPercentage: 10,
			MinCandidateFoodsAbsolute:   100,
		},
	},
	{
		Name: "InterDguestAffinity",
		Args: &v1.InterDguestAffinityArgs{
			HardDguestAffinityWeight: 1,
		},
	},
	{
		Name: "FoodAffinity",
		Args: &v1.FoodAffinityArgs{},
	},
	{
		Name: "FoodResourcesBalancedAllocation",
		Args: &v1.FoodResourcesBalancedAllocationArgs{
			Resources: []v1.ResourceSpec{{Name: "cpu", Weight: 1}, {Name: "memory", Weight: 1}},
		},
	},
	{
		Name: "FoodResourcesFit",
		Args: &v1.FoodResourcesFitArgs{
			ScoringStrategy: &v1.ScoringStrategy{
				Type:      v1.LeastAllocated,
				Resources: []v1.ResourceSpec{{Name: "cpu", Weight: 1}, {Name: "memory", Weight: 1}},
			},
		},
	},
	{
		Name: "DguestTopologySpread",
		Args: &v1.DguestTopologySpreadArgs{
			DefaultingType: v1.SystemDefaulting,
		},
	},
}

// PluginsV1beta3 is the set of default v1beta3 plugins (before MultiPoint expansion)
var PluginsV1beta3 = &v1.Plugins{
	MultiPoint: v1.PluginSet{
		Enabled: []v1.Plugin{
			{Name: names.PrioritySort},
			{Name: names.FoodUnschedulable},
			{Name: names.TaintToleration, Weight: 3},
			{Name: names.FoodAffinity, Weight: 2},
			{Name: names.FoodResourcesFit, Weight: 1},
			{Name: names.DguestTopologySpread, Weight: 2},
			{Name: names.DefaultPreemption},
			{Name: names.FoodResourcesBalancedAllocation, Weight: 1},
			{Name: names.DefaultBinder},
		},
	},
}

// ExpandedPluginsV1beta3 default set of v1beta3 plugins after MultiPoint expansion
var ExpandedPluginsV1beta3 = &v1.Plugins{
	QueueSort: v1.PluginSet{
		Enabled: []v1.Plugin{
			{Name: names.PrioritySort},
		},
	},
	PreFilter: v1.PluginSet{
		Enabled: []v1.Plugin{
			{Name: names.FoodAffinity},
			{Name: names.FoodPorts},
			{Name: names.FoodResourcesFit},
			{Name: names.DguestTopologySpread},
			{Name: names.InterDguestAffinity},
		},
	},
	Filter: v1.PluginSet{
		Enabled: []v1.Plugin{
			{Name: names.FoodUnschedulable},
			{Name: names.FoodName},
			{Name: names.TaintToleration},
			{Name: names.FoodAffinity},
			{Name: names.FoodResourcesFit},
			{Name: names.DguestTopologySpread},
			{Name: names.InterDguestAffinity},
		},
	},
	PostFilter: v1.PluginSet{
		Enabled: []v1.Plugin{
			{Name: names.DefaultPreemption},
		},
	},
	PreScore: v1.PluginSet{
		Enabled: []v1.Plugin{
			{Name: names.TaintToleration},
			{Name: names.FoodAffinity},
			{Name: names.DguestTopologySpread},
			{Name: names.InterDguestAffinity},
		},
	},
	Score: v1.PluginSet{
		Enabled: []v1.Plugin{
			// Weight is tripled because:
			// - This is a score coming from user preference.
			// - Usage of food tainting to group foods in the cluster is increasing becoming a use-case
			// for many user workloads
			{Name: names.TaintToleration, Weight: 3},
			// Weight is doubled because:
			// - This is a score coming from user preference.
			{Name: names.FoodAffinity, Weight: 2},
			{Name: names.FoodResourcesFit, Weight: 1},
			// Weight is doubled because:
			// - This is a score coming from user preference.
			// - It makes its signal comparable to FoodResourcesLeastAllocated.
			{Name: names.DguestTopologySpread, Weight: 2},
			// Weight is doubled because:
			// - This is a score coming from user preference.
			{Name: names.InterDguestAffinity, Weight: 2},
			{Name: names.FoodResourcesBalancedAllocation, Weight: 1},
		},
	},
	Reserve: v1.PluginSet{},
	PreBind: v1.PluginSet{},
	Bind: v1.PluginSet{
		Enabled: []v1.Plugin{
			{Name: names.DefaultBinder},
		},
	},
}

// PluginConfigsV1beta3 default plugin configurations.
var PluginConfigsV1beta3 = []v1.PluginConfig{
	{
		Name: "DefaultPreemption",
		Args: &v1.DefaultPreemptionArgs{
			MinCandidateFoodsPercentage: 10,
			MinCandidateFoodsAbsolute:   100,
		},
	},
	{
		Name: "InterDguestAffinity",
		Args: &v1.InterDguestAffinityArgs{
			HardDguestAffinityWeight: 1,
		},
	},
	{
		Name: "FoodAffinity",
		Args: &v1.FoodAffinityArgs{},
	},
	{
		Name: "FoodResourcesBalancedAllocation",
		Args: &v1.FoodResourcesBalancedAllocationArgs{
			Resources: []v1.ResourceSpec{{Name: "cpu", Weight: 1}, {Name: "memory", Weight: 1}},
		},
	},
	{
		Name: "FoodResourcesFit",
		Args: &v1.FoodResourcesFitArgs{
			ScoringStrategy: &v1.ScoringStrategy{
				Type:      v1.LeastAllocated,
				Resources: []v1.ResourceSpec{{Name: "cpu", Weight: 1}, {Name: "memory", Weight: 1}},
			},
		},
	},
	{
		Name: "DguestTopologySpread",
		Args: &v1.DguestTopologySpreadArgs{
			DefaultingType: v1.SystemDefaulting,
		},
	},
}

// PluginsV1 is the set of default v1 plugins (before MultiPoint expansion)
var PluginsV1 = &v1.Plugins{
	MultiPoint: v1.PluginSet{
		Enabled: []v1.Plugin{
			{Name: names.PrioritySort},
			{Name: names.FoodUnschedulable},
			{Name: names.FoodName},
			{Name: names.TaintToleration, Weight: 3},
			{Name: names.FoodAffinity, Weight: 2},
			{Name: names.FoodPorts},
			{Name: names.FoodResourcesFit, Weight: 1},
			{Name: names.DguestTopologySpread, Weight: 2},
			{Name: names.InterDguestAffinity, Weight: 2},
			{Name: names.DefaultPreemption},
			{Name: names.FoodResourcesBalancedAllocation, Weight: 1},
			{Name: names.DefaultBinder},
		},
	},
}

// ExpandedPluginsV1 default set of v1 plugins after MultiPoint expansion
var ExpandedPluginsV1 = &v1.Plugins{
	QueueSort: v1.PluginSet{
		Enabled: []v1.Plugin{
			{Name: names.PrioritySort},
		},
	},
	PreFilter: v1.PluginSet{
		Enabled: []v1.Plugin{
			{Name: names.FoodAffinity},
			{Name: names.FoodPorts},
			{Name: names.FoodResourcesFit},
			{Name: names.DguestTopologySpread},
			{Name: names.InterDguestAffinity},
		},
	},
	Filter: v1.PluginSet{
		Enabled: []v1.Plugin{
			{Name: names.FoodUnschedulable},
			{Name: names.FoodName},
			{Name: names.TaintToleration},
			{Name: names.FoodAffinity},
			{Name: names.FoodResourcesFit},
			{Name: names.DguestTopologySpread},
			{Name: names.InterDguestAffinity},
		},
	},
	PostFilter: v1.PluginSet{
		Enabled: []v1.Plugin{
			{Name: names.DefaultPreemption},
		},
	},
	PreScore: v1.PluginSet{
		Enabled: []v1.Plugin{
			{Name: names.TaintToleration},
			{Name: names.FoodAffinity},
			{Name: names.DguestTopologySpread},
			{Name: names.InterDguestAffinity},
		},
	},
	Score: v1.PluginSet{
		Enabled: []v1.Plugin{
			// Weight is tripled because:
			// - This is a score coming from user preference.
			// - Usage of food tainting to group foods in the cluster is increasing becoming a use-case
			// for many user workloads
			{Name: names.TaintToleration, Weight: 3},
			// Weight is doubled because:
			// - This is a score coming from user preference.
			{Name: names.FoodAffinity, Weight: 2},
			{Name: names.FoodResourcesFit, Weight: 1},
			// Weight is doubled because:
			// - This is a score coming from user preference.
			// - It makes its signal comparable to FoodResourcesLeastAllocated.
			{Name: names.DguestTopologySpread, Weight: 2},
			// Weight is doubled because:
			// - This is a score coming from user preference.
			{Name: names.InterDguestAffinity, Weight: 2},
			{Name: names.FoodResourcesBalancedAllocation, Weight: 1},
		},
	},
	Reserve: v1.PluginSet{},
	PreBind: v1.PluginSet{},
	Bind: v1.PluginSet{
		Enabled: []v1.Plugin{
			{Name: names.DefaultBinder},
		},
	},
}

// PluginConfigsV1 default plugin configurations.
var PluginConfigsV1 = []v1.PluginConfig{
	{
		Name: "DefaultPreemption",
		Args: &v1.DefaultPreemptionArgs{
			MinCandidateFoodsPercentage: 10,
			MinCandidateFoodsAbsolute:   100,
		},
	},
	{
		Name: "InterDguestAffinity",
		Args: &v1.InterDguestAffinityArgs{
			HardDguestAffinityWeight: 1,
		},
	},
	{
		Name: "FoodAffinity",
		Args: &v1.FoodAffinityArgs{},
	},
	{
		Name: "FoodResourcesBalancedAllocation",
		Args: &v1.FoodResourcesBalancedAllocationArgs{
			Resources: []v1.ResourceSpec{{Name: "cpu", Weight: 1}, {Name: "memory", Weight: 1}},
		},
	},
	{
		Name: "FoodResourcesFit",
		Args: &v1.FoodResourcesFitArgs{
			ScoringStrategy: &v1.ScoringStrategy{
				Type:      v1.LeastAllocated,
				Resources: []v1.ResourceSpec{{Name: "cpu", Weight: 1}, {Name: "memory", Weight: 1}},
			},
		},
	},
	{
		Name: "DguestTopologySpread",
		Args: &v1.DguestTopologySpreadArgs{
			DefaultingType: v1.SystemDefaulting,
		},
	},
}
