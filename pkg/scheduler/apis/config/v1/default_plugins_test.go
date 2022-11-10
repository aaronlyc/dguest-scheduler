/*
Copyright 2022 The Kubernetes Authors.

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

package v1

import (
	"dguest-scheduler/pkg/scheduler/apis/config"
	"testing"

	"dguest-scheduler/pkg/scheduler/framework/plugins/names"
	"github.com/google/go-cmp/cmp"
	"k8s.io/apiserver/pkg/util/feature"
	"k8s.io/component-base/featuregate"
	featuregatetesting "k8s.io/component-base/featuregate/testing"
)

func TestApplyFeatureGates(t *testing.T) {
	tests := []struct {
		name       string
		features   map[featuregate.Feature]bool
		wantConfig *config.Plugins
	}{
		{
			name: "Feature gates disabled",
			wantConfig: &config.Plugins{
				MultiPoint: config.PluginSet{
					Enabled: []config.Plugin{
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
						//{Name: names.DguestTopologySpread, Weight: pointer.Int32(2)},
						//{Name: names.InterDguestAffinity, Weight: pointer.Int32(2)},
						//{Name: names.DefaultPreemption},
						//{Name: names.FoodResourcesBalancedAllocation, Weight: pointer.Int32(1)},
						//{Name: names.ImageLocality, Weight: pointer.Int32(1)},
						{Name: names.DefaultBinder},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for k, v := range test.features {
				defer featuregatetesting.SetFeatureGateDuringTest(t, feature.DefaultFeatureGate, k, v)()
			}

			gotConfig := getDefaultPlugins()
			if diff := cmp.Diff(test.wantConfig, gotConfig); diff != "" {
				t.Errorf("unexpected config diff (-want, +got): %s", diff)
			}
		})
	}
}

func TestMergePlugins(t *testing.T) {
	tests := []struct {
		name            string
		customPlugins   *config.Plugins
		defaultPlugins  *config.Plugins
		expectedPlugins *config.Plugins
	}{
		{
			name: "AppendCustomPlugin",
			customPlugins: &config.Plugins{
				Filter: config.PluginSet{
					Enabled: []config.Plugin{
						{Name: "CustomPlugin"},
					},
				},
			},
			defaultPlugins: &config.Plugins{
				Filter: config.PluginSet{
					Enabled: []config.Plugin{
						{Name: "DefaultPlugin1"},
						{Name: "DefaultPlugin2"},
					},
				},
			},
			expectedPlugins: &config.Plugins{
				Filter: config.PluginSet{
					Enabled: []config.Plugin{
						{Name: "DefaultPlugin1"},
						{Name: "DefaultPlugin2"},
						{Name: "CustomPlugin"},
					},
				},
			},
		},
		{
			name: "InsertAfterDefaultPlugins2",
			customPlugins: &config.Plugins{
				Filter: config.PluginSet{
					Enabled: []config.Plugin{
						{Name: "CustomPlugin"},
						{Name: "DefaultPlugin2"},
					},
					Disabled: []config.Plugin{
						{Name: "DefaultPlugin2"},
					},
				},
			},
			defaultPlugins: &config.Plugins{
				Filter: config.PluginSet{
					Enabled: []config.Plugin{
						{Name: "DefaultPlugin1"},
						{Name: "DefaultPlugin2"},
					},
				},
			},
			expectedPlugins: &config.Plugins{
				Filter: config.PluginSet{
					Enabled: []config.Plugin{
						{Name: "DefaultPlugin1"},
						{Name: "CustomPlugin"},
						{Name: "DefaultPlugin2"},
					},
					Disabled: []config.Plugin{
						{Name: "DefaultPlugin2"},
					},
				},
			},
		},
		{
			name: "InsertBeforeAllPlugins",
			customPlugins: &config.Plugins{
				Filter: config.PluginSet{
					Enabled: []config.Plugin{
						{Name: "CustomPlugin"},
						{Name: "DefaultPlugin1"},
						{Name: "DefaultPlugin2"},
					},
					Disabled: []config.Plugin{
						{Name: "*"},
					},
				},
			},
			defaultPlugins: &config.Plugins{
				Filter: config.PluginSet{
					Enabled: []config.Plugin{
						{Name: "DefaultPlugin1"},
						{Name: "DefaultPlugin2"},
					},
				},
			},
			expectedPlugins: &config.Plugins{
				Filter: config.PluginSet{
					Enabled: []config.Plugin{
						{Name: "CustomPlugin"},
						{Name: "DefaultPlugin1"},
						{Name: "DefaultPlugin2"},
					},
					Disabled: []config.Plugin{
						{Name: "*"},
					},
				},
			},
		},
		{
			name: "ReorderDefaultPlugins",
			customPlugins: &config.Plugins{
				Filter: config.PluginSet{
					Enabled: []config.Plugin{
						{Name: "DefaultPlugin2"},
						{Name: "DefaultPlugin1"},
					},
					Disabled: []config.Plugin{
						{Name: "*"},
					},
				},
			},
			defaultPlugins: &config.Plugins{
				Filter: config.PluginSet{
					Enabled: []config.Plugin{
						{Name: "DefaultPlugin1"},
						{Name: "DefaultPlugin2"},
					},
				},
			},
			expectedPlugins: &config.Plugins{
				Filter: config.PluginSet{
					Enabled: []config.Plugin{
						{Name: "DefaultPlugin2"},
						{Name: "DefaultPlugin1"},
					},
					Disabled: []config.Plugin{
						{Name: "*"},
					},
				},
			},
		},
		{
			name:          "ApplyNilCustomPlugin",
			customPlugins: nil,
			defaultPlugins: &config.Plugins{
				Filter: config.PluginSet{
					Enabled: []config.Plugin{
						{Name: "DefaultPlugin1"},
						{Name: "DefaultPlugin2"},
					},
				},
			},
			expectedPlugins: &config.Plugins{
				Filter: config.PluginSet{
					Enabled: []config.Plugin{
						{Name: "DefaultPlugin1"},
						{Name: "DefaultPlugin2"},
					},
				},
			},
		},
		{
			name: "CustomPluginOverrideDefaultPlugin",
			customPlugins: &config.Plugins{
				Filter: config.PluginSet{
					Enabled: []config.Plugin{
						{Name: "Plugin1", Weight: 2},
						{Name: "Plugin3", Weight: 3},
					},
				},
			},
			defaultPlugins: &config.Plugins{
				Filter: config.PluginSet{
					Enabled: []config.Plugin{
						{Name: "Plugin1"},
						{Name: "Plugin2"},
						{Name: "Plugin3"},
					},
				},
			},
			expectedPlugins: &config.Plugins{
				Filter: config.PluginSet{
					Enabled: []config.Plugin{
						{Name: "Plugin1", Weight: 2},
						{Name: "Plugin2"},
						{Name: "Plugin3", Weight: 3},
					},
				},
			},
		},
		{
			name: "OrderPreserveAfterOverride",
			customPlugins: &config.Plugins{
				Filter: config.PluginSet{
					Enabled: []config.Plugin{
						{Name: "Plugin2", Weight: 2},
						{Name: "Plugin1", Weight: 1},
					},
				},
			},
			defaultPlugins: &config.Plugins{
				Filter: config.PluginSet{
					Enabled: []config.Plugin{
						{Name: "Plugin1"},
						{Name: "Plugin2"},
						{Name: "Plugin3"},
					},
				},
			},
			expectedPlugins: &config.Plugins{
				Filter: config.PluginSet{
					Enabled: []config.Plugin{
						{Name: "Plugin1", Weight: 1},
						{Name: "Plugin2", Weight: 2},
						{Name: "Plugin3"},
					},
				},
			},
		},
		{
			name: "RepeatedCustomPlugin",
			customPlugins: &config.Plugins{
				Filter: config.PluginSet{
					Enabled: []config.Plugin{
						{Name: "Plugin1"},
						{Name: "Plugin2", Weight: 2},
						{Name: "Plugin3"},
						{Name: "Plugin2", Weight: 4},
					},
				},
			},
			defaultPlugins: &config.Plugins{
				Filter: config.PluginSet{
					Enabled: []config.Plugin{
						{Name: "Plugin1"},
						{Name: "Plugin2"},
						{Name: "Plugin3"},
					},
				},
			},
			expectedPlugins: &config.Plugins{
				Filter: config.PluginSet{
					Enabled: []config.Plugin{
						{Name: "Plugin1"},
						{Name: "Plugin2", Weight: 4},
						{Name: "Plugin3"},
						{Name: "Plugin2", Weight: 2},
					},
				},
			},
		},
		{
			name: "Append custom MultiPoint plugin",
			customPlugins: &config.Plugins{
				MultiPoint: config.PluginSet{
					Enabled: []config.Plugin{
						{Name: "CustomPlugin"},
					},
				},
			},
			defaultPlugins: &config.Plugins{
				MultiPoint: config.PluginSet{
					Enabled: []config.Plugin{
						{Name: "DefaultPlugin1"},
						{Name: "DefaultPlugin2"},
					},
				},
			},
			expectedPlugins: &config.Plugins{
				MultiPoint: config.PluginSet{
					Enabled: []config.Plugin{
						{Name: "DefaultPlugin1"},
						{Name: "DefaultPlugin2"},
						{Name: "CustomPlugin"},
					},
				},
			},
		},
		{
			name: "Append disabled Multipoint plugins",
			customPlugins: &config.Plugins{
				MultiPoint: config.PluginSet{
					Enabled: []config.Plugin{
						{Name: "CustomPlugin"},
						{Name: "CustomPlugin2"},
					},
					Disabled: []config.Plugin{
						{Name: "DefaultPlugin2"},
					},
				},
				Score: config.PluginSet{
					Disabled: []config.Plugin{
						{Name: "CustomPlugin2"},
					},
				},
			},
			defaultPlugins: &config.Plugins{
				MultiPoint: config.PluginSet{
					Enabled: []config.Plugin{
						{Name: "DefaultPlugin1"},
						{Name: "DefaultPlugin2"},
					},
				},
			},
			expectedPlugins: &config.Plugins{
				MultiPoint: config.PluginSet{
					Enabled: []config.Plugin{
						{Name: "DefaultPlugin1"},
						{Name: "CustomPlugin"},
						{Name: "CustomPlugin2"},
					},
					Disabled: []config.Plugin{
						{Name: "DefaultPlugin2"},
					},
				},
				Score: config.PluginSet{
					Disabled: []config.Plugin{
						{Name: "CustomPlugin2"},
					},
				},
			},
		},
		{
			name: "override default MultiPoint plugins with custom value",
			customPlugins: &config.Plugins{
				MultiPoint: config.PluginSet{
					Enabled: []config.Plugin{
						{Name: "DefaultPlugin", Weight: 5},
					},
				},
			},
			defaultPlugins: &config.Plugins{
				MultiPoint: config.PluginSet{
					Enabled: []config.Plugin{
						{Name: "DefaultPlugin"},
					},
				},
			},
			expectedPlugins: &config.Plugins{
				MultiPoint: config.PluginSet{
					Enabled: []config.Plugin{
						{Name: "DefaultPlugin", Weight: 5},
					},
				},
			},
		},
		{
			name: "disabled MultiPoint plugin in default set",
			defaultPlugins: &config.Plugins{
				MultiPoint: config.PluginSet{
					Enabled: []config.Plugin{
						{Name: "DefaultPlugin"},
					},
					Disabled: []config.Plugin{
						{Name: "DefaultPlugin2"},
					},
				},
			},
			customPlugins: &config.Plugins{
				MultiPoint: config.PluginSet{
					Enabled: []config.Plugin{
						{Name: "CustomPlugin"},
					},
				},
			},
			expectedPlugins: &config.Plugins{
				MultiPoint: config.PluginSet{
					Enabled: []config.Plugin{
						{Name: "DefaultPlugin"},
						{Name: "CustomPlugin"},
					},
					Disabled: []config.Plugin{
						{Name: "DefaultPlugin2"},
					},
				},
			},
		},
		{
			name: "disabled MultiPoint plugin in default set for specific extension point",
			defaultPlugins: &config.Plugins{
				MultiPoint: config.PluginSet{
					Enabled: []config.Plugin{
						{Name: "DefaultPlugin"},
					},
				},
				Score: config.PluginSet{
					Disabled: []config.Plugin{
						{Name: "DefaultPlugin2"},
					},
				},
			},
			customPlugins: &config.Plugins{
				MultiPoint: config.PluginSet{
					Enabled: []config.Plugin{
						{Name: "CustomPlugin"},
					},
				},
			},
			expectedPlugins: &config.Plugins{
				MultiPoint: config.PluginSet{
					Enabled: []config.Plugin{
						{Name: "DefaultPlugin"},
						{Name: "CustomPlugin"},
					},
				},
				Score: config.PluginSet{
					Disabled: []config.Plugin{
						{Name: "DefaultPlugin2"},
					},
				},
			},
		},
		{
			name: "multipoint with only disabled gets merged",
			defaultPlugins: &config.Plugins{
				MultiPoint: config.PluginSet{
					Enabled: []config.Plugin{
						{Name: "DefaultPlugin"},
					},
				},
			},
			customPlugins: &config.Plugins{
				MultiPoint: config.PluginSet{
					Disabled: []config.Plugin{
						{Name: "DefaultPlugin"},
					},
				},
			},
			expectedPlugins: &config.Plugins{
				MultiPoint: config.PluginSet{
					Disabled: []config.Plugin{
						{Name: "DefaultPlugin"},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gotPlugins := mergePlugins(test.defaultPlugins, test.customPlugins)
			if d := cmp.Diff(test.expectedPlugins, gotPlugins); d != "" {
				t.Fatalf("plugins mismatch (-want +got):\n%s", d)
			}
		})
	}
}
