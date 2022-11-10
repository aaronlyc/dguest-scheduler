/*
Copyright 2019 The Kubernetes Authors.

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

package plugins

import (
	"dguest-scheduler/pkg/scheduler/framework/plugins/defaultbinder"
	"dguest-scheduler/pkg/scheduler/framework/plugins/foodname"
	"dguest-scheduler/pkg/scheduler/framework/plugins/queuesort"
	"dguest-scheduler/pkg/scheduler/framework/runtime"
)

// NewInTreeRegistry builds the registry with all the in-tree plugins.
// A scheduler that runs out of tree plugins can register additional plugins
// through the WithFrameworkOutOfTreeRegistry option.
func NewInTreeRegistry() runtime.Registry {
	//fts := plfeature.Features{
	//	EnableReadWriteOnceDguest:                       feature.DefaultFeatureGate.Enabled(features.ReadWriteOnceDguest),
	//	EnableVolumeCapacityPriority:                 feature.DefaultFeatureGate.Enabled(features.VolumeCapacityPriority),
	//	EnableMinDomainsInDguestTopologySpread:          feature.DefaultFeatureGate.Enabled(features.MinDomainsInDguestTopologySpread),
	//	EnableFoodInclusionPolicyInDguestTopologySpread: feature.DefaultFeatureGate.Enabled(features.FoodInclusionPolicyInDguestTopologySpread),
	//	EnableMatchLabelKeysInDguestTopologySpread:      feature.DefaultFeatureGate.Enabled(features.MatchLabelKeysInDguestTopologySpread),
	//}

	return runtime.Registry{
		//selectorspread.Name:                  selectorspread.New,
		//imagelocality.Name:                   imagelocality.New,
		//tainttoleration.Name:                 tainttoleration.New,
		foodname.Name: foodname.New,
		//foodports.Name:                       foodports.New,
		//foodaffinity.Name:                    foodaffinity.New,
		//dguesttopologyspread.Name:               runtime.FactoryAdapter(fts, dguesttopologyspread.New),
		//foodunschedulable.Name:               foodunschedulable.New,
		//foodresources.Name:                   runtime.FactoryAdapter(fts, foodresources.NewFit),
		//foodresources.BalancedAllocationName: runtime.FactoryAdapter(fts, foodresources.NewBalancedAllocation),
		//volumebinding.Name:                   runtime.FactoryAdapter(fts, volumebinding.New),
		//volumerestrictions.Name:              runtime.FactoryAdapter(fts, volumerestrictions.New),
		//volumezone.Name:                      volumezone.New,
		//foodvolumelimits.CSIName:             runtime.FactoryAdapter(fts, foodvolumelimits.NewCSI),
		//foodvolumelimits.EBSName:             runtime.FactoryAdapter(fts, foodvolumelimits.NewEBS),
		//foodvolumelimits.GCEPDName:           runtime.FactoryAdapter(fts, foodvolumelimits.NewGCEPD),
		//foodvolumelimits.AzureDiskName:       runtime.FactoryAdapter(fts, foodvolumelimits.NewAzureDisk),
		//foodvolumelimits.CinderName:          runtime.FactoryAdapter(fts, foodvolumelimits.NewCinder),
		//interdguestaffinity.Name:                interdguestaffinity.New,
		queuesort.Name:     queuesort.New,
		defaultbinder.Name: defaultbinder.New,
		//defaultpreemption.Name:               runtime.FactoryAdapter(fts, defaultpreemption.New),
	}
}
