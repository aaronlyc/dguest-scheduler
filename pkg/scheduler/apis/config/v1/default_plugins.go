package v1

import (
	"dguest-scheduler/pkg/scheduler/framework/plugins/names"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"
)

// getDefaultPlugins returns the default set of plugins.
func getDefaultPlugins() *Plugins {
	plugins := &Plugins{
		MultiPoint: PluginSet{
			Enabled: []Plugin{
				{Name: names.PrioritySort},
				//{Name: names.FoodUnschedulable},
				{Name: names.FoodName},
				//{Name: names.TaintToleration, Weight: 3},
				//{Name: names.FoodAffinity, Weight: 2},
				//{Name: names.FoodPorts},
				//{Name: names.FoodResourcesFit, Weight: 1},
				//{Name: names.VolumeRestrictions},
				//{Name: names.EBSLimits},
				//{Name: names.GCEPDLimits},
				//{Name: names.FoodVolumeLimits},
				//{Name: names.AzureDiskLimits},
				//{Name: names.VolumeBinding},
				//{Name: names.VolumeZone},
				//{Name: names.DguestTopologySpread, Weight: 2},
				//{Name: names.InterDguestAffinity, Weight: 2},
				//{Name: names.DefaultPreemption},
				//{Name: names.FoodResourcesBalancedAllocation, Weight: pointer.Int32(1)},
				//{Name: names.ImageLocality, Weight: pointer.Int32(1)},
				{Name: names.DefaultBinder},
			},
		},
	}

	return plugins
}

// mergePlugins merges the custom set into the given default one, handling disabled sets.
func mergePlugins(defaultPlugins, customPlugins *Plugins) *Plugins {
	if customPlugins == nil {
		return defaultPlugins
	}

	defaultPlugins.MultiPoint = mergePluginSet(defaultPlugins.MultiPoint, customPlugins.MultiPoint)
	defaultPlugins.QueueSort = mergePluginSet(defaultPlugins.QueueSort, customPlugins.QueueSort)
	defaultPlugins.PreFilter = mergePluginSet(defaultPlugins.PreFilter, customPlugins.PreFilter)
	defaultPlugins.Filter = mergePluginSet(defaultPlugins.Filter, customPlugins.Filter)
	defaultPlugins.PostFilter = mergePluginSet(defaultPlugins.PostFilter, customPlugins.PostFilter)
	defaultPlugins.PreScore = mergePluginSet(defaultPlugins.PreScore, customPlugins.PreScore)
	defaultPlugins.Score = mergePluginSet(defaultPlugins.Score, customPlugins.Score)
	defaultPlugins.Reserve = mergePluginSet(defaultPlugins.Reserve, customPlugins.Reserve)
	defaultPlugins.Permit = mergePluginSet(defaultPlugins.Permit, customPlugins.Permit)
	defaultPlugins.PreBind = mergePluginSet(defaultPlugins.PreBind, customPlugins.PreBind)
	defaultPlugins.Bind = mergePluginSet(defaultPlugins.Bind, customPlugins.Bind)
	defaultPlugins.PostBind = mergePluginSet(defaultPlugins.PostBind, customPlugins.PostBind)
	return defaultPlugins
}

type pluginIndex struct {
	index  int
	plugin Plugin
}

func mergePluginSet(defaultPluginSet, customPluginSet PluginSet) PluginSet {
	disabledPlugins := sets.NewString()
	enabledCustomPlugins := make(map[string]pluginIndex)
	// replacedPluginIndex is a set of index of plugins, which have replaced the default plugins.
	replacedPluginIndex := sets.NewInt()
	var disabled []Plugin
	for _, disabledPlugin := range customPluginSet.Disabled {
		// if the user is manually disabling any (or all, with "*") default plugins for an extension point,
		// we need to track that so that the MultiPoint extension logic in the framework can know to skip
		// inserting unspecified default plugins to this point.
		disabled = append(disabled, Plugin{Name: disabledPlugin.Name})
		disabledPlugins.Insert(disabledPlugin.Name)
	}

	// With MultiPoint, we may now have some disabledPlugins in the default registry
	// For example, we enable PluginX with Filter+Score through MultiPoint but disable its Score plugin by default.
	for _, disabledPlugin := range defaultPluginSet.Disabled {
		disabled = append(disabled, Plugin{Name: disabledPlugin.Name})
		disabledPlugins.Insert(disabledPlugin.Name)
	}

	for index, enabledPlugin := range customPluginSet.Enabled {
		enabledCustomPlugins[enabledPlugin.Name] = pluginIndex{index, enabledPlugin}
	}
	var enabledPlugins []Plugin
	if !disabledPlugins.Has("*") {
		for _, defaultEnabledPlugin := range defaultPluginSet.Enabled {
			if disabledPlugins.Has(defaultEnabledPlugin.Name) {
				continue
			}
			// The default plugin is explicitly re-configured, update the default plugin accordingly.
			if customPlugin, ok := enabledCustomPlugins[defaultEnabledPlugin.Name]; ok {
				klog.InfoS("Default plugin is explicitly re-configured; overriding", "plugin", defaultEnabledPlugin.Name)
				// Update the default plugin in place to preserve order.
				defaultEnabledPlugin = customPlugin.plugin
				replacedPluginIndex.Insert(customPlugin.index)
			}
			enabledPlugins = append(enabledPlugins, defaultEnabledPlugin)
		}
	}

	// Append all the custom plugins which haven't replaced any default plugins.
	// Note: duplicated custom plugins will still be appended here.
	// If so, the instantiation of scheduler framework will detect it and abort.
	for index, plugin := range customPluginSet.Enabled {
		if !replacedPluginIndex.Has(index) {
			enabledPlugins = append(enabledPlugins, plugin)
		}
	}
	return PluginSet{Enabled: enabledPlugins, Disabled: disabled}
}
