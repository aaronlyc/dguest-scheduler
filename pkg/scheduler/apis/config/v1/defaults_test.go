// /*
// Copyright 2022 The Kubernetes Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// */
package v1

//
//import (
//	"dguest-scheduler/pkg/scheduler/apis/config"
//	"testing"
//	"time"
//
//	"github.com/google/go-cmp/cmp"
//
//	"dguest-scheduler/pkg/scheduler/framework/plugins/names"
//	v1 "k8s.io/api/core/v1"
//	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
//	"k8s.io/apimachinery/pkg/runtime"
//	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
//	"k8s.io/apiserver/pkg/util/feature"
//	componentbaseconfig "k8s.io/component-base/config"
//	"k8s.io/component-base/featuregate"
//	featuregatetesting "k8s.io/component-base/featuregate/testing"
//	"k8s.io/kubernetes/pkg/features"
//)
//
//var pluginConfigs = []config.PluginConfig{
//	{
//		Name: "DefaultPreemption",
//		Args: &config.DefaultPreemptionArgs{
//			TypeMeta: metav1.TypeMeta{
//				Kind:       "DefaultPreemptionArgs",
//				APIVersion: "kubescheduler.config.k8s.io/v1",
//			},
//			MinCandidateFoodsPercentage: 10,
//			MinCandidateFoodsAbsolute:   100,
//		},
//	},
//	{
//		Name: "InterDguestAffinity",
//		Args: &config.InterDguestAffinityArgs{
//			TypeMeta: metav1.TypeMeta{
//				Kind:       "InterDguestAffinityArgs",
//				APIVersion: "kubescheduler.config.k8s.io/v1",
//			},
//			HardDguestAffinityWeight: 1,
//		},
//	},
//	{
//		Name: "FoodAffinity",
//		Args: &config.FoodAffinityArgs{
//			TypeMeta: metav1.TypeMeta{
//				Kind:       "FoodAffinityArgs",
//				APIVersion: "kubescheduler.config.k8s.io/v1",
//			}},
//	},
//	{
//		Name: "FoodResourcesBalancedAllocation",
//		Args: &config.FoodResourcesBalancedAllocationArgs{
//			TypeMeta: metav1.TypeMeta{
//				Kind:       "FoodResourcesBalancedAllocationArgs",
//				APIVersion: "kubescheduler.config.k8s.io/v1",
//			},
//			Resources: []config.ResourceSpec{{Name: "cpu", Weight: 1}, {Name: "memory", Weight: 1}},
//		},
//	},
//	{
//		Name: "FoodResourcesFit",
//		Args: &config.FoodResourcesFitArgs{
//			TypeMeta: metav1.TypeMeta{
//				Kind:       "FoodResourcesFitArgs",
//				APIVersion: "kubescheduler.config.k8s.io/v1",
//			},
//			ScoringStrategy: &config.ScoringStrategy{
//				Type:      config.LeastAllocated,
//				Resources: []config.ResourceSpec{{Name: "cpu", Weight: 1}, {Name: "memory", Weight: 1}},
//			},
//		},
//	},
//	{
//		Name: "DguestTopologySpread",
//		Args: &config.DguestTopologySpreadArgs{
//			TypeMeta: metav1.TypeMeta{
//				Kind:       "DguestTopologySpreadArgs",
//				APIVersion: "kubescheduler.config.k8s.io/v1",
//			},
//			DefaultingType: config.SystemDefaulting,
//		},
//	},
//	{
//		Name: "VolumeBinding",
//		Args: &config.VolumeBindingArgs{
//			TypeMeta: metav1.TypeMeta{
//				Kind:       "VolumeBindingArgs",
//				APIVersion: "kubescheduler.config.k8s.io/v1",
//			},
//			BindTimeoutSeconds: 600,
//		},
//	},
//}
//
//func TestSchedulerDefaults(t *testing.T) {
//	enable := true
//	tests := []struct {
//		name     string
//		config   *config.SchedulerConfiguration
//		expected *config.SchedulerConfiguration
//	}{
//		{
//			name:   "empty config",
//			config: &config.SchedulerConfiguration{},
//			expected: &config.SchedulerConfiguration{
//				Parallelism: 16,
//				DebuggingConfiguration: componentbaseconfig.DebuggingConfiguration{
//					EnableProfiling:           enable,
//					EnableContentionProfiling: enable,
//				},
//				LeaderElection: componentbaseconfig.LeaderElectionConfiguration{
//					LeaderElect:       true,
//					LeaseDuration:     metav1.Duration{Duration: 15 * time.Second},
//					RenewDeadline:     metav1.Duration{Duration: 10 * time.Second},
//					RetryPeriod:       metav1.Duration{Duration: 2 * time.Second},
//					ResourceLock:      "leases",
//					ResourceNamespace: "kube-system",
//					ResourceName:      "kube-scheduler",
//				},
//				ClientConnection: componentbaseconfig.ClientConnectionConfiguration{
//					QPS:         50,
//					Burst:       100,
//					ContentType: "application/vnd.scheduler.protobuf",
//				},
//				PercentageOfFoodsToScore:    0,
//				DguestInitialBackoffSeconds: 1,
//				DguestMaxBackoffSeconds:     10,
//				Profiles: []config.SchedulerProfile{
//					{
//						Plugins:       getDefaultPlugins(),
//						PluginConfig:  pluginConfigs,
//						SchedulerName: "default-scheduler",
//					},
//				},
//			},
//		},
//		{
//			name: "no scheduler name",
//			config: &config.SchedulerConfiguration{
//				Profiles: []config.SchedulerProfile{{}},
//			},
//			expected: &config.SchedulerConfiguration{
//				Parallelism: 16,
//				DebuggingConfiguration: componentbaseconfig.DebuggingConfiguration{
//					EnableProfiling:           enable,
//					EnableContentionProfiling: enable,
//				},
//				LeaderElection: componentbaseconfig.LeaderElectionConfiguration{
//					LeaderElect:       true,
//					LeaseDuration:     metav1.Duration{Duration: 15 * time.Second},
//					RenewDeadline:     metav1.Duration{Duration: 10 * time.Second},
//					RetryPeriod:       metav1.Duration{Duration: 2 * time.Second},
//					ResourceLock:      "leases",
//					ResourceNamespace: "kube-system",
//					ResourceName:      "kube-scheduler",
//				},
//				ClientConnection: componentbaseconfig.ClientConnectionConfiguration{
//					QPS:         50,
//					Burst:       100,
//					ContentType: "application/vnd.scheduler.protobuf",
//				},
//				PercentageOfFoodsToScore:    10,
//				DguestInitialBackoffSeconds: 1,
//				DguestMaxBackoffSeconds:     10,
//				Profiles: []config.SchedulerProfile{
//					{
//						SchedulerName: "default-scheduler",
//						Plugins:       getDefaultPlugins(),
//						PluginConfig:  pluginConfigs},
//				},
//			},
//		},
//		{
//			name: "two profiles",
//			config: &config.SchedulerConfiguration{
//				Parallelism: 16,
//				Profiles: []config.SchedulerProfile{
//					{
//						PluginConfig: []config.PluginConfig{
//							{Name: "FooPlugin"},
//						},
//					},
//					{
//						SchedulerName: "custom-scheduler",
//						Plugins: &config.Plugins{
//							Bind: config.PluginSet{
//								Enabled: []config.Plugin{
//									{Name: "BarPlugin"},
//								},
//								Disabled: []config.Plugin{
//									{Name: names.DefaultBinder},
//								},
//							},
//						},
//					},
//				},
//			},
//			expected: &config.SchedulerConfiguration{
//				Parallelism: 16,
//				DebuggingConfiguration: componentbaseconfig.DebuggingConfiguration{
//					EnableProfiling:           enable,
//					EnableContentionProfiling: enable,
//				},
//				LeaderElection: componentbaseconfig.LeaderElectionConfiguration{
//					LeaderElect:       true,
//					LeaseDuration:     metav1.Duration{Duration: 15 * time.Second},
//					RenewDeadline:     metav1.Duration{Duration: 10 * time.Second},
//					RetryPeriod:       metav1.Duration{Duration: 2 * time.Second},
//					ResourceLock:      "leases",
//					ResourceNamespace: "kube-system",
//					ResourceName:      "kube-scheduler",
//				},
//				ClientConnection: componentbaseconfig.ClientConnectionConfiguration{
//					QPS:         50,
//					Burst:       100,
//					ContentType: "application/vnd.scheduler.protobuf",
//				},
//				PercentageOfFoodsToScore:    0,
//				DguestInitialBackoffSeconds: 1,
//				DguestMaxBackoffSeconds:     10,
//				Profiles: []config.SchedulerProfile{
//					{
//						Plugins: getDefaultPlugins(),
//						PluginConfig: []config.PluginConfig{
//							{Name: "FooPlugin"},
//							{
//								Name: "DefaultPreemption",
//								Args: &config.DefaultPreemptionArgs{
//									TypeMeta: metav1.TypeMeta{
//										Kind:       "DefaultPreemptionArgs",
//										APIVersion: "kubescheduler.config.k8s.io/v1",
//									},
//									MinCandidateFoodsPercentage: 10,
//									MinCandidateFoodsAbsolute:   100,
//								},
//							},
//							{
//								Name: "InterDguestAffinity",
//								Args: &config.InterDguestAffinityArgs{
//									TypeMeta: metav1.TypeMeta{
//										Kind:       "InterDguestAffinityArgs",
//										APIVersion: "kubescheduler.config.k8s.io/v1",
//									},
//									HardDguestAffinityWeight: 1,
//								},
//							},
//							{
//								Name: "FoodAffinity",
//								Args: &config.FoodAffinityArgs{
//									TypeMeta: metav1.TypeMeta{
//										Kind:       "FoodAffinityArgs",
//										APIVersion: "kubescheduler.config.k8s.io/v1",
//									},
//								},
//							},
//							{
//								Name: "FoodResourcesBalancedAllocation",
//								Args: &config.FoodResourcesBalancedAllocationArgs{
//									TypeMeta: metav1.TypeMeta{
//										Kind:       "FoodResourcesBalancedAllocationArgs",
//										APIVersion: "kubescheduler.config.k8s.io/v1",
//									},
//									Resources: []config.ResourceSpec{{Name: "cpu", Weight: 1}, {Name: "memory", Weight: 1}},
//								},
//							},
//							{
//								Name: "FoodResourcesFit",
//								Args: &config.FoodResourcesFitArgs{
//									TypeMeta: metav1.TypeMeta{
//										Kind:       "FoodResourcesFitArgs",
//										APIVersion: "kubescheduler.config.k8s.io/v1",
//									},
//									ScoringStrategy: &config.ScoringStrategy{
//										Type:      config.LeastAllocated,
//										Resources: []config.ResourceSpec{{Name: "cpu", Weight: 1}, {Name: "memory", Weight: 1}},
//									},
//								},
//							},
//							{
//								Name: "DguestTopologySpread",
//								Args: &config.DguestTopologySpreadArgs{
//									TypeMeta: metav1.TypeMeta{
//										Kind:       "DguestTopologySpreadArgs",
//										APIVersion: "kubescheduler.config.k8s.io/v1",
//									},
//									DefaultingType: config.SystemDefaulting,
//								},
//							},
//							{
//								Name: "VolumeBinding",
//								Args: &config.VolumeBindingArgs{
//									TypeMeta: metav1.TypeMeta{
//										Kind:       "VolumeBindingArgs",
//										APIVersion: "kubescheduler.config.k8s.io/v1",
//									},
//									BindTimeoutSeconds: 600,
//								},
//							},
//						},
//					},
//					{
//						SchedulerName: "custom-scheduler",
//						Plugins: &config.Plugins{
//							MultiPoint: config.PluginSet{
//								Enabled: []config.Plugin{
//									{Name: names.PrioritySort},
//									//{Name: names.FoodUnschedulable},
//									{Name: names.FoodName},
//									//{Name: names.TaintToleration, Weight: pointer.Int32(3)},
//									//{Name: names.FoodAffinity, Weight: pointer.Int32(2)},
//									//{Name: names.FoodPorts},
//									//{Name: names.FoodResourcesFit, Weight: pointer.Int32(1)},
//									//{Name: names.VolumeRestrictions},
//									//{Name: names.EBSLimits},
//									//{Name: names.GCEPDLimits},
//									//{Name: names.FoodVolumeLimits},
//									//{Name: names.AzureDiskLimits},
//									//{Name: names.VolumeBinding},
//									//{Name: names.VolumeZone},
//									//{Name: names.DguestTopologySpread, Weight: pointer.Int32(2)},
//									//{Name: names.InterDguestAffinity, Weight: pointer.Int32(2)},
//									//{Name: names.DefaultPreemption},
//									//{Name: names.FoodResourcesBalancedAllocation, Weight: pointer.Int32(1)},
//									//{Name: names.ImageLocality, Weight: pointer.Int32(1)},
//									{Name: names.DefaultBinder},
//								},
//							},
//							Bind: config.PluginSet{
//								Enabled: []config.Plugin{
//									{Name: "BarPlugin"},
//								},
//								Disabled: []config.Plugin{
//									{Name: names.DefaultBinder},
//								},
//							},
//						},
//						PluginConfig: pluginConfigs,
//					},
//				},
//			},
//		},
//		{
//			name: "Prallelism with no port",
//			config: &config.SchedulerConfiguration{
//				Parallelism: 16,
//			},
//			expected: &config.SchedulerConfiguration{
//				Parallelism: 16,
//				DebuggingConfiguration: componentbaseconfig.DebuggingConfiguration{
//					EnableProfiling:           enable,
//					EnableContentionProfiling: enable,
//				},
//				LeaderElection: componentbaseconfig.LeaderElectionConfiguration{
//					LeaderElect:       true,
//					LeaseDuration:     metav1.Duration{Duration: 15 * time.Second},
//					RenewDeadline:     metav1.Duration{Duration: 10 * time.Second},
//					RetryPeriod:       metav1.Duration{Duration: 2 * time.Second},
//					ResourceLock:      "leases",
//					ResourceNamespace: "kube-system",
//					ResourceName:      "kube-scheduler",
//				},
//				ClientConnection: componentbaseconfig.ClientConnectionConfiguration{
//					QPS:         50,
//					Burst:       100,
//					ContentType: "application/vnd.scheduler.protobuf",
//				},
//				PercentageOfFoodsToScore:    0,
//				DguestInitialBackoffSeconds: 1,
//				DguestMaxBackoffSeconds:     10,
//				Profiles: []config.SchedulerProfile{
//					{
//						Plugins:       getDefaultPlugins(),
//						PluginConfig:  pluginConfigs,
//						SchedulerName: "default-scheduler",
//					},
//				},
//			},
//		},
//		{
//			name: "set non default parallelism",
//			config: &config.SchedulerConfiguration{
//				Parallelism: 8,
//			},
//			expected: &config.SchedulerConfiguration{
//				Parallelism: 8,
//				DebuggingConfiguration: componentbaseconfig.DebuggingConfiguration{
//					EnableProfiling:           enable,
//					EnableContentionProfiling: enable,
//				},
//				LeaderElection: componentbaseconfig.LeaderElectionConfiguration{
//					LeaderElect:       true,
//					LeaseDuration:     metav1.Duration{Duration: 15 * time.Second},
//					RenewDeadline:     metav1.Duration{Duration: 10 * time.Second},
//					RetryPeriod:       metav1.Duration{Duration: 2 * time.Second},
//					ResourceLock:      "leases",
//					ResourceNamespace: "kube-system",
//					ResourceName:      "kube-scheduler",
//				},
//				ClientConnection: componentbaseconfig.ClientConnectionConfiguration{
//					QPS:         50,
//					Burst:       100,
//					ContentType: "application/vnd.scheduler.protobuf",
//				},
//				PercentageOfFoodsToScore:    0,
//				DguestInitialBackoffSeconds: 1,
//				DguestMaxBackoffSeconds:     10,
//				Profiles: []config.SchedulerProfile{
//					{
//						Plugins:       getDefaultPlugins(),
//						PluginConfig:  pluginConfigs,
//						SchedulerName: "default-scheduler",
//					},
//				},
//			},
//		},
//	}
//	for _, tc := range tests {
//		t.Run(tc.name, func(t *testing.T) {
//			SetDefaults_KubeSchedulerConfiguration(tc.config)
//			if diff := cmp.Diff(tc.expected, tc.config); diff != "" {
//				t.Errorf("Got unexpected defaults (-want, +got):\n%s", diff)
//			}
//		})
//	}
//}
//
//func TestPluginArgsDefaults(t *testing.T) {
//	tests := []struct {
//		name     string
//		features map[featuregate.Feature]bool
//		in       runtime.Object
//		want     runtime.Object
//	}{
//		{
//			name: "DefaultPreemptionArgs empty",
//			in:   &config.DefaultPreemptionArgs{},
//			want: &config.DefaultPreemptionArgs{
//				MinCandidateFoodsPercentage: 10,
//				MinCandidateFoodsAbsolute:   100,
//			},
//		},
//		{
//			name: "DefaultPreemptionArgs with value",
//			in: &config.DefaultPreemptionArgs{
//				MinCandidateFoodsPercentage: 50,
//			},
//			want: &config.DefaultPreemptionArgs{
//				MinCandidateFoodsPercentage: 50,
//				MinCandidateFoodsAbsolute:   100,
//			},
//		},
//		{
//			name: "InterDguestAffinityArgs empty",
//			in:   &config.InterDguestAffinityArgs{},
//			want: &config.InterDguestAffinityArgs{
//				HardDguestAffinityWeight: 1,
//			},
//		},
//		{
//			name: "InterDguestAffinityArgs explicit 0",
//			in: &config.InterDguestAffinityArgs{
//				HardDguestAffinityWeight: 0,
//			},
//			want: &config.InterDguestAffinityArgs{
//				HardDguestAffinityWeight: 0,
//			},
//		},
//		{
//			name: "InterDguestAffinityArgs with value",
//			in: &config.InterDguestAffinityArgs{
//				HardDguestAffinityWeight: 5,
//			},
//			want: &config.InterDguestAffinityArgs{
//				HardDguestAffinityWeight: 5,
//			},
//		},
//		{
//			name: "FoodResourcesBalancedAllocationArgs resources empty",
//			in:   &config.FoodResourcesBalancedAllocationArgs{},
//			want: &config.FoodResourcesBalancedAllocationArgs{
//				Resources: []config.ResourceSpec{
//					{Name: "cpu", Weight: 1}, {Name: "memory", Weight: 1},
//				},
//			},
//		},
//		{
//			name: "FoodResourcesBalancedAllocationArgs with scalar resource",
//			in: &config.FoodResourcesBalancedAllocationArgs{
//				Resources: []config.ResourceSpec{
//					{Name: "scalar.io/scalar1", Weight: 1},
//				},
//			},
//			want: &config.FoodResourcesBalancedAllocationArgs{
//				Resources: []config.ResourceSpec{
//					{Name: "scalar.io/scalar1", Weight: 1},
//				},
//			},
//		},
//		{
//			name: "FoodResourcesBalancedAllocationArgs with mixed resources",
//			in: &config.FoodResourcesBalancedAllocationArgs{
//				Resources: []config.ResourceSpec{
//					{Name: string(v1.ResourceCPU), Weight: 1},
//					{Name: "scalar.io/scalar1", Weight: 1},
//				},
//			},
//			want: &config.FoodResourcesBalancedAllocationArgs{
//				Resources: []config.ResourceSpec{
//					{Name: string(v1.ResourceCPU), Weight: 1},
//					{Name: "scalar.io/scalar1", Weight: 1},
//				},
//			},
//		},
//		{
//			name: "FoodResourcesBalancedAllocationArgs have resource no weight",
//			in: &config.FoodResourcesBalancedAllocationArgs{
//				Resources: []config.ResourceSpec{
//					{Name: string(v1.ResourceCPU)},
//					{Name: "scalar.io/scalar0"},
//					{Name: "scalar.io/scalar1", Weight: 1},
//				},
//			},
//			want: &config.FoodResourcesBalancedAllocationArgs{
//				Resources: []config.ResourceSpec{
//					{Name: string(v1.ResourceCPU), Weight: 1},
//					{Name: "scalar.io/scalar0", Weight: 1},
//					{Name: "scalar.io/scalar1", Weight: 1},
//				},
//			},
//		},
//		{
//			name: "DguestTopologySpreadArgs resources empty",
//			in:   &config.DguestTopologySpreadArgs{},
//			want: &config.DguestTopologySpreadArgs{
//				DefaultingType: config.SystemDefaulting,
//			},
//		},
//		{
//			name: "DguestTopologySpreadArgs resources with value",
//			in: &config.DguestTopologySpreadArgs{
//				DefaultConstraints: []v1.TopologySpreadConstraint{
//					{
//						TopologyKey:       "planet",
//						WhenUnsatisfiable: v1.DoNotSchedule,
//						MaxSkew:           2,
//					},
//				},
//			},
//			want: &config.DguestTopologySpreadArgs{
//				DefaultConstraints: []v1.TopologySpreadConstraint{
//					{
//						TopologyKey:       "planet",
//						WhenUnsatisfiable: v1.DoNotSchedule,
//						MaxSkew:           2,
//					},
//				},
//				DefaultingType: config.SystemDefaulting,
//			},
//		},
//		{
//			name: "FoodResourcesFitArgs not set",
//			in:   &config.FoodResourcesFitArgs{},
//			want: &config.FoodResourcesFitArgs{
//				ScoringStrategy: &config.ScoringStrategy{
//					Type:      config.LeastAllocated,
//					Resources: defaultResourceSpec,
//				},
//			},
//		},
//		{
//			name: "FoodResourcesFitArgs Resources empty",
//			in: &config.FoodResourcesFitArgs{
//				ScoringStrategy: &config.ScoringStrategy{
//					Type: config.MostAllocated,
//				},
//			},
//			want: &config.FoodResourcesFitArgs{
//				ScoringStrategy: &config.ScoringStrategy{
//					Type:      config.MostAllocated,
//					Resources: defaultResourceSpec,
//				},
//			},
//		},
//		{
//			name: "VolumeBindingArgs empty, VolumeCapacityPriority disabled",
//			features: map[featuregate.Feature]bool{
//				features.VolumeCapacityPriority: false,
//			},
//			in: &config.VolumeBindingArgs{},
//			want: &config.VolumeBindingArgs{
//				BindTimeoutSeconds: 600,
//			},
//		},
//		{
//			name: "VolumeBindingArgs empty, VolumeCapacityPriority enabled",
//			features: map[featuregate.Feature]bool{
//				features.VolumeCapacityPriority: true,
//			},
//			in: &config.VolumeBindingArgs{},
//			want: &config.VolumeBindingArgs{
//				BindTimeoutSeconds: 600,
//				Shape: []config.UtilizationShapePoint{
//					{Utilization: 0, Score: 0},
//					{Utilization: 100, Score: 10},
//				},
//			},
//		},
//	}
//	for _, tc := range tests {
//		scheme := runtime.NewScheme()
//		utilruntime.Must(AddToScheme(scheme))
//		t.Run(tc.name, func(t *testing.T) {
//			for k, v := range tc.features {
//				defer featuregatetesting.SetFeatureGateDuringTest(t, feature.DefaultFeatureGate, k, v)()
//			}
//			scheme.Default(tc.in)
//			if diff := cmp.Diff(tc.in, tc.want); diff != "" {
//				t.Errorf("Got unexpected defaults (-want, +got):\n%s", diff)
//			}
//		})
//	}
//}
