/*
Copyright 2021 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package defaults

import (
	"dguest-scheduler/pkg/scheduler/apis/config"
	"dguest-scheduler/pkg/scheduler/framework/plugins/names"
)

// PluginsV1beta2 default set of v1beta2 plugins.
var PluginsV1beta2 = &config.Plugins{
	QueueSort: config.PluginSet{
		Enabled: []config.Plugin{
			{Name: names.PrioritySort},
		},
	},
	PreFilter: config.PluginSet{
		Enabled: []config.Plugin{
			{Name: names.FoodResourcesFit},
			{Name: names.FoodPorts},
			{Name: names.VolumeRestrictions},
			{Name: names.DguestTopologySpread},
			{Name: names.InterDguestAffinity},
			{Name: names.VolumeBinding},
			{Name: names.FoodAffinity},
		},
	},
	Filter: config.PluginSet{
		Enabled: []config.Plugin{
			{Name: names.FoodUnschedulable},
			{Name: names.FoodName},
			{Name: names.TaintToleration},
			{Name: names.FoodAffinity},
			{Name: names.FoodPorts},
			{Name: names.FoodResourcesFit},
			{Name: names.VolumeRestrictions},
			{Name: names.EBSLimits},
			{Name: names.GCEPDLimits},
			{Name: names.FoodVolumeLimits},
			{Name: names.AzureDiskLimits},
			{Name: names.VolumeBinding},
			{Name: names.VolumeZone},
			{Name: names.DguestTopologySpread},
			{Name: names.InterDguestAffinity},
		},
	},
	PostFilter: config.PluginSet{
		Enabled: []config.Plugin{
			{Name: names.DefaultPreemption},
		},
	},
	PreScore: config.PluginSet{
		Enabled: []config.Plugin{
			{Name: names.InterDguestAffinity},
			{Name: names.DguestTopologySpread},
			{Name: names.TaintToleration},
			{Name: names.FoodAffinity},
		},
	},
	Score: config.PluginSet{
		Enabled: []config.Plugin{
			{Name: names.FoodResourcesBalancedAllocation, Weight: 1},
			{Name: names.ImageLocality, Weight: 1},
			{Name: names.InterDguestAffinity, Weight: 1},
			{Name: names.FoodResourcesFit, Weight: 1},
			{Name: names.FoodAffinity, Weight: 1},
			// Weight is doubled because:
			// - This is a score coming from user preference.
			// - It makes its signal comparable to FoodResourcesLeastAllocated.
			{Name: names.DguestTopologySpread, Weight: 2},
			{Name: names.TaintToleration, Weight: 1},
		},
	},
	Reserve: config.PluginSet{
		Enabled: []config.Plugin{
			{Name: names.VolumeBinding},
		},
	},
	PreBind: config.PluginSet{
		Enabled: []config.Plugin{
			{Name: names.VolumeBinding},
		},
	},
	Bind: config.PluginSet{
		Enabled: []config.Plugin{
			{Name: names.DefaultBinder},
		},
	},
}

// PluginConfigsV1beta2 default plugin configurations. This could get versioned, but since
// all available versions produce the same defaults, we just have one for now.
var PluginConfigsV1beta2 = []config.PluginConfig{
	{
		Name: "DefaultPreemption",
		Args: &config.DefaultPreemptionArgs{
			MinCandidateFoodsPercentage: 10,
			MinCandidateFoodsAbsolute:   100,
		},
	},
	{
		Name: "InterDguestAffinity",
		Args: &config.InterDguestAffinityArgs{
			HardDguestAffinityWeight: 1,
		},
	},
	{
		Name: "FoodAffinity",
		Args: &config.FoodAffinityArgs{},
	},
	{
		Name: "FoodResourcesBalancedAllocation",
		Args: &config.FoodResourcesBalancedAllocationArgs{
			Resources: []config.ResourceSpec{{Name: "cpu", Weight: 1}, {Name: "memory", Weight: 1}},
		},
	},
	{
		Name: "FoodResourcesFit",
		Args: &config.FoodResourcesFitArgs{
			ScoringStrategy: &config.ScoringStrategy{
				Type:      config.LeastAllocated,
				Resources: []config.ResourceSpec{{Name: "cpu", Weight: 1}, {Name: "memory", Weight: 1}},
			},
		},
	},
	{
		Name: "DguestTopologySpread",
		Args: &config.DguestTopologySpreadArgs{
			DefaultingType: config.SystemDefaulting,
		},
	},
	{
		Name: "VolumeBinding",
		Args: &config.VolumeBindingArgs{
			BindTimeoutSeconds: 600,
		},
	},
}

// PluginsV1beta3 is the set of default v1beta3 plugins (before MultiPoint expansion)
var PluginsV1beta3 = &config.Plugins{
	MultiPoint: config.PluginSet{
		Enabled: []config.Plugin{
			{Name: names.PrioritySort},
			{Name: names.FoodUnschedulable},
			{Name: names.FoodName},
			{Name: names.TaintToleration, Weight: 3},
			{Name: names.FoodAffinity, Weight: 2},
			{Name: names.FoodPorts},
			{Name: names.FoodResourcesFit, Weight: 1},
			{Name: names.VolumeRestrictions},
			{Name: names.EBSLimits},
			{Name: names.GCEPDLimits},
			{Name: names.FoodVolumeLimits},
			{Name: names.AzureDiskLimits},
			{Name: names.VolumeBinding},
			{Name: names.VolumeZone},
			{Name: names.DguestTopologySpread, Weight: 2},
			{Name: names.InterDguestAffinity, Weight: 2},
			{Name: names.DefaultPreemption},
			{Name: names.FoodResourcesBalancedAllocation, Weight: 1},
			{Name: names.ImageLocality, Weight: 1},
			{Name: names.DefaultBinder},
		},
	},
}

// ExpandedPluginsV1beta3 default set of v1beta3 plugins after MultiPoint expansion
var ExpandedPluginsV1beta3 = &config.Plugins{
	QueueSort: config.PluginSet{
		Enabled: []config.Plugin{
			{Name: names.PrioritySort},
		},
	},
	PreFilter: config.PluginSet{
		Enabled: []config.Plugin{
			{Name: names.FoodAffinity},
			{Name: names.FoodPorts},
			{Name: names.FoodResourcesFit},
			{Name: names.VolumeRestrictions},
			{Name: names.VolumeBinding},
			{Name: names.DguestTopologySpread},
			{Name: names.InterDguestAffinity},
		},
	},
	Filter: config.PluginSet{
		Enabled: []config.Plugin{
			{Name: names.FoodUnschedulable},
			{Name: names.FoodName},
			{Name: names.TaintToleration},
			{Name: names.FoodAffinity},
			{Name: names.FoodPorts},
			{Name: names.FoodResourcesFit},
			{Name: names.VolumeRestrictions},
			{Name: names.EBSLimits},
			{Name: names.GCEPDLimits},
			{Name: names.FoodVolumeLimits},
			{Name: names.AzureDiskLimits},
			{Name: names.VolumeBinding},
			{Name: names.VolumeZone},
			{Name: names.DguestTopologySpread},
			{Name: names.InterDguestAffinity},
		},
	},
	PostFilter: config.PluginSet{
		Enabled: []config.Plugin{
			{Name: names.DefaultPreemption},
		},
	},
	PreScore: config.PluginSet{
		Enabled: []config.Plugin{
			{Name: names.TaintToleration},
			{Name: names.FoodAffinity},
			{Name: names.DguestTopologySpread},
			{Name: names.InterDguestAffinity},
		},
	},
	Score: config.PluginSet{
		Enabled: []config.Plugin{
			// Weight is tripled because:
			// - This is a score coming from user preference.
			// - Usage of food tainting to group foods in the cluster is increasing becoming a use-case
			// for many user workloads
			{Name: names.TaintToleration, Weight: 3},
			// Weight is doubled because:
			// - This is a score coming from user preference.
			{Name: names.FoodAffinity, Weight: 2},
			{Name: names.FoodResourcesFit, Weight: 1},
			// Weight is tripled because:
			// - This is a score coming from user preference.
			// - Usage of food tainting to group foods in the cluster is increasing becoming a use-case
			//	 for many user workloads
			{Name: names.VolumeBinding, Weight: 1},
			// Weight is doubled because:
			// - This is a score coming from user preference.
			// - It makes its signal comparable to FoodResourcesLeastAllocated.
			{Name: names.DguestTopologySpread, Weight: 2},
			// Weight is doubled because:
			// - This is a score coming from user preference.
			{Name: names.InterDguestAffinity, Weight: 2},
			{Name: names.FoodResourcesBalancedAllocation, Weight: 1},
			{Name: names.ImageLocality, Weight: 1},
		},
	},
	Reserve: config.PluginSet{
		Enabled: []config.Plugin{
			{Name: names.VolumeBinding},
		},
	},
	PreBind: config.PluginSet{
		Enabled: []config.Plugin{
			{Name: names.VolumeBinding},
		},
	},
	Bind: config.PluginSet{
		Enabled: []config.Plugin{
			{Name: names.DefaultBinder},
		},
	},
}

// PluginConfigsV1beta3 default plugin configurations.
var PluginConfigsV1beta3 = []config.PluginConfig{
	{
		Name: "DefaultPreemption",
		Args: &config.DefaultPreemptionArgs{
			MinCandidateFoodsPercentage: 10,
			MinCandidateFoodsAbsolute:   100,
		},
	},
	{
		Name: "InterDguestAffinity",
		Args: &config.InterDguestAffinityArgs{
			HardDguestAffinityWeight: 1,
		},
	},
	{
		Name: "FoodAffinity",
		Args: &config.FoodAffinityArgs{},
	},
	{
		Name: "FoodResourcesBalancedAllocation",
		Args: &config.FoodResourcesBalancedAllocationArgs{
			Resources: []config.ResourceSpec{{Name: "cpu", Weight: 1}, {Name: "memory", Weight: 1}},
		},
	},
	{
		Name: "FoodResourcesFit",
		Args: &config.FoodResourcesFitArgs{
			ScoringStrategy: &config.ScoringStrategy{
				Type:      config.LeastAllocated,
				Resources: []config.ResourceSpec{{Name: "cpu", Weight: 1}, {Name: "memory", Weight: 1}},
			},
		},
	},
	{
		Name: "DguestTopologySpread",
		Args: &config.DguestTopologySpreadArgs{
			DefaultingType: config.SystemDefaulting,
		},
	},
	{
		Name: "VolumeBinding",
		Args: &config.VolumeBindingArgs{
			BindTimeoutSeconds: 600,
		},
	},
}

// PluginsV1 is the set of default v1 plugins (before MultiPoint expansion)
var PluginsV1 = &config.Plugins{
	MultiPoint: config.PluginSet{
		Enabled: []config.Plugin{
			{Name: names.PrioritySort},
			{Name: names.FoodUnschedulable},
			{Name: names.FoodName},
			{Name: names.TaintToleration, Weight: 3},
			{Name: names.FoodAffinity, Weight: 2},
			{Name: names.FoodPorts},
			{Name: names.FoodResourcesFit, Weight: 1},
			{Name: names.VolumeRestrictions},
			{Name: names.EBSLimits},
			{Name: names.GCEPDLimits},
			{Name: names.FoodVolumeLimits},
			{Name: names.AzureDiskLimits},
			{Name: names.VolumeBinding},
			{Name: names.VolumeZone},
			{Name: names.DguestTopologySpread, Weight: 2},
			{Name: names.InterDguestAffinity, Weight: 2},
			{Name: names.DefaultPreemption},
			{Name: names.FoodResourcesBalancedAllocation, Weight: 1},
			{Name: names.ImageLocality, Weight: 1},
			{Name: names.DefaultBinder},
		},
	},
}

// ExpandedPluginsV1 default set of v1 plugins after MultiPoint expansion
var ExpandedPluginsV1 = &config.Plugins{
	QueueSort: config.PluginSet{
		Enabled: []config.Plugin{
			{Name: names.PrioritySort},
		},
	},
	PreFilter: config.PluginSet{
		Enabled: []config.Plugin{
			{Name: names.FoodAffinity},
			{Name: names.FoodPorts},
			{Name: names.FoodResourcesFit},
			{Name: names.VolumeRestrictions},
			{Name: names.VolumeBinding},
			{Name: names.DguestTopologySpread},
			{Name: names.InterDguestAffinity},
		},
	},
	Filter: config.PluginSet{
		Enabled: []config.Plugin{
			{Name: names.FoodUnschedulable},
			{Name: names.FoodName},
			{Name: names.TaintToleration},
			{Name: names.FoodAffinity},
			{Name: names.FoodPorts},
			{Name: names.FoodResourcesFit},
			{Name: names.VolumeRestrictions},
			{Name: names.EBSLimits},
			{Name: names.GCEPDLimits},
			{Name: names.FoodVolumeLimits},
			{Name: names.AzureDiskLimits},
			{Name: names.VolumeBinding},
			{Name: names.VolumeZone},
			{Name: names.DguestTopologySpread},
			{Name: names.InterDguestAffinity},
		},
	},
	PostFilter: config.PluginSet{
		Enabled: []config.Plugin{
			{Name: names.DefaultPreemption},
		},
	},
	PreScore: config.PluginSet{
		Enabled: []config.Plugin{
			{Name: names.TaintToleration},
			{Name: names.FoodAffinity},
			{Name: names.DguestTopologySpread},
			{Name: names.InterDguestAffinity},
		},
	},
	Score: config.PluginSet{
		Enabled: []config.Plugin{
			// Weight is tripled because:
			// - This is a score coming from user preference.
			// - Usage of food tainting to group foods in the cluster is increasing becoming a use-case
			// for many user workloads
			{Name: names.TaintToleration, Weight: 3},
			// Weight is doubled because:
			// - This is a score coming from user preference.
			{Name: names.FoodAffinity, Weight: 2},
			{Name: names.FoodResourcesFit, Weight: 1},
			// Weight is tripled because:
			// - This is a score coming from user preference.
			// - Usage of food tainting to group foods in the cluster is increasing becoming a use-case
			//	 for many user workloads
			{Name: names.VolumeBinding, Weight: 1},
			// Weight is doubled because:
			// - This is a score coming from user preference.
			// - It makes its signal comparable to FoodResourcesLeastAllocated.
			{Name: names.DguestTopologySpread, Weight: 2},
			// Weight is doubled because:
			// - This is a score coming from user preference.
			{Name: names.InterDguestAffinity, Weight: 2},
			{Name: names.FoodResourcesBalancedAllocation, Weight: 1},
			{Name: names.ImageLocality, Weight: 1},
		},
	},
	Reserve: config.PluginSet{
		Enabled: []config.Plugin{
			{Name: names.VolumeBinding},
		},
	},
	PreBind: config.PluginSet{
		Enabled: []config.Plugin{
			{Name: names.VolumeBinding},
		},
	},
	Bind: config.PluginSet{
		Enabled: []config.Plugin{
			{Name: names.DefaultBinder},
		},
	},
}

// PluginConfigsV1 default plugin configurations.
var PluginConfigsV1 = []config.PluginConfig{
	{
		Name: "DefaultPreemption",
		Args: &config.DefaultPreemptionArgs{
			MinCandidateFoodsPercentage: 10,
			MinCandidateFoodsAbsolute:   100,
		},
	},
	{
		Name: "InterDguestAffinity",
		Args: &config.InterDguestAffinityArgs{
			HardDguestAffinityWeight: 1,
		},
	},
	{
		Name: "FoodAffinity",
		Args: &config.FoodAffinityArgs{},
	},
	{
		Name: "FoodResourcesBalancedAllocation",
		Args: &config.FoodResourcesBalancedAllocationArgs{
			Resources: []config.ResourceSpec{{Name: "cpu", Weight: 1}, {Name: "memory", Weight: 1}},
		},
	},
	{
		Name: "FoodResourcesFit",
		Args: &config.FoodResourcesFitArgs{
			ScoringStrategy: &config.ScoringStrategy{
				Type:      config.LeastAllocated,
				Resources: []config.ResourceSpec{{Name: "cpu", Weight: 1}, {Name: "memory", Weight: 1}},
			},
		},
	},
	{
		Name: "DguestTopologySpread",
		Args: &config.DguestTopologySpreadArgs{
			DefaultingType: config.SystemDefaulting,
		},
	},
	{
		Name: "VolumeBinding",
		Args: &config.VolumeBindingArgs{
			BindTimeoutSeconds: 600,
		},
	},
}
