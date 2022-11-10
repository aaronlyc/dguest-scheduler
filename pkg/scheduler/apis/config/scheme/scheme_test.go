/*
Copyright 2020 The Kubernetes Authors.

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

package scheme

import (
	"bytes"
	"testing"

	"dguest-scheduler/pkg/scheduler/apis/config"
	"dguest-scheduler/pkg/scheduler/apis/config/testing/defaults"
	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	v1 "k8s.io/kube-scheduler/config/v1"
	"k8s.io/kube-scheduler/config/v1beta2"
	"k8s.io/kube-scheduler/config/v1beta3"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/yaml"
)

// TestCodecsDecodePluginConfig tests that embedded plugin args get decoded
// into their appropriate internal types and defaults are applied.
func TestCodecsDecodePluginConfig(t *testing.T) {
	testCases := []struct {
		name         string
		data         []byte
		wantErr      string
		wantProfiles []config.KubeSchedulerProfile
	}{
		// v1beta2 tests
		{
			name: "v1beta2 all plugin args in default profile",
			data: []byte(`
apiVersion: kubescheduler.config.k8s.io/v1beta2
kind: SchedulerConfiguration
profiles:
- pluginConfig:
  - name: DefaultPreemption
    args:
      minCandidateFoodsPercentage: 50
      minCandidateFoodsAbsolute: 500
  - name: InterDguestAffinity
    args:
      hardDguestAffinityWeight: 5
  - name: FoodResourcesFit
    args:
      ignoredResources: ["foo"]
  - name: DguestTopologySpread
    args:
      defaultConstraints:
      - maxSkew: 1
        topologyKey: zone
        whenUnsatisfiable: ScheduleAnyway
  - name: VolumeBinding
    args:
      bindTimeoutSeconds: 300
  - name: FoodAffinity
    args:
      addedAffinity:
        requiredDuringSchedulingIgnoredDuringExecution:
          foodSelectorTerms:
          - matchExpressions:
            - key: foo
              operator: In
              values: ["bar"]
  - name: FoodResourcesBalancedAllocation
    args:
      resources:
        - name: cpu       # default weight(1) will be set.
        - name: memory    # weight 0 will be replaced by 1.
          weight: 0
        - name: scalar0
          weight: 1
        - name: scalar1   # default weight(1) will be set for scalar1
        - name: scalar2   # weight 0 will be replaced by 1.
          weight: 0
        - name: scalar3
          weight: 2
`),
			wantProfiles: []config.KubeSchedulerProfile{
				{
					SchedulerName: "default-scheduler",
					Plugins:       defaults.PluginsV1beta2,
					PluginConfig: []config.PluginConfig{
						{
							Name: "DefaultPreemption",
							Args: &config.DefaultPreemptionArgs{MinCandidateFoodsPercentage: 50, MinCandidateFoodsAbsolute: 500},
						},
						{
							Name: "InterDguestAffinity",
							Args: &config.InterDguestAffinityArgs{HardDguestAffinityWeight: 5},
						},
						{
							Name: "FoodResourcesFit",
							Args: &config.FoodResourcesFitArgs{
								IgnoredResources: []string{"foo"},
								ScoringStrategy: &config.ScoringStrategy{
									Type: config.LeastAllocated,
									Resources: []config.ResourceSpec{
										{Name: "cpu", Weight: 1},
										{Name: "memory", Weight: 1},
									},
								},
							},
						},
						{
							Name: "DguestTopologySpread",
							Args: &config.DguestTopologySpreadArgs{
								DefaultConstraints: []corev1.TopologySpreadConstraint{
									{MaxSkew: 1, TopologyKey: "zone", WhenUnsatisfiable: corev1.ScheduleAnyway},
								},
								DefaultingType: config.SystemDefaulting,
							},
						},
						{
							Name: "VolumeBinding",
							Args: &config.VolumeBindingArgs{
								BindTimeoutSeconds: 300,
							},
						},
						{
							Name: "FoodAffinity",
							Args: &config.FoodAffinityArgs{
								AddedAffinity: &corev1alpha1.FoodAffinity{
									RequiredDuringSchedulingIgnoredDuringExecution: &corev1alpha1.FoodSelector{
										FoodSelectorTerms: []corev1alpha1.FoodSelectorTerm{
											{
												MatchExpressions: []corev1alpha1.FoodSelectorRequirement{
													{
														Key:      "foo",
														Operator: corev1alpha1.FoodSelectorOpIn,
														Values:   []string{"bar"},
													},
												},
											},
										},
									},
								},
							},
						},
						{
							Name: "FoodResourcesBalancedAllocation",
							Args: &config.FoodResourcesBalancedAllocationArgs{
								Resources: []config.ResourceSpec{
									{Name: "cpu", Weight: 1},
									{Name: "memory", Weight: 1},
									{Name: "scalar0", Weight: 1},
									{Name: "scalar1", Weight: 1},
									{Name: "scalar2", Weight: 1},
									{Name: "scalar3", Weight: 2}},
							},
						},
					},
				},
			},
		},
		{
			name: "v1beta2 plugins can include version and kind",
			data: []byte(`
apiVersion: kubescheduler.config.k8s.io/v1beta2
kind: SchedulerConfiguration
profiles:
- pluginConfig:
  - name: DefaultPreemption
    args:
      apiVersion: kubescheduler.config.k8s.io/v1beta2
      kind: DefaultPreemptionArgs
      minCandidateFoodsPercentage: 50
`),
			wantProfiles: []config.KubeSchedulerProfile{
				{
					SchedulerName: "default-scheduler",
					Plugins:       defaults.PluginsV1beta2,
					PluginConfig: []config.PluginConfig{
						{
							Name: "DefaultPreemption",
							Args: &config.DefaultPreemptionArgs{MinCandidateFoodsPercentage: 50, MinCandidateFoodsAbsolute: 100},
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
									Type: config.LeastAllocated,
									Resources: []config.ResourceSpec{
										{Name: "cpu", Weight: 1},
										{Name: "memory", Weight: 1},
									},
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
					},
				},
			},
		},
		{
			name: "plugin group and kind should match the type",
			data: []byte(`
apiVersion: kubescheduler.config.k8s.io/v1beta2
kind: SchedulerConfiguration
profiles:
- pluginConfig:
  - name: DefaultPreemption
    args:
      apiVersion: kubescheduler.config.k8s.io/v1beta2
      kind: InterDguestAffinityArgs
`),
			wantErr: `decoding .profiles[0].pluginConfig[0]: args for plugin DefaultPreemption were not of type DefaultPreemptionArgs.kubescheduler.config.k8s.io, got InterDguestAffinityArgs.kubescheduler.config.k8s.io`,
		},
		{
			name: "v1beta2 NodResourcesFitArgs shape encoding is strict",
			data: []byte(`
apiVersion: kubescheduler.config.k8s.io/v1beta2
kind: SchedulerConfiguration
profiles:
- pluginConfig:
  - name: FoodResourcesFit
    args:
      scoringStrategy:
        requestedToCapacityRatio:
          shape:
          - Score: 2
            Utilization: 1
`),
			wantErr: `strict decoding error: decoding .profiles[0].pluginConfig[0]: strict decoding error: decoding args for plugin FoodResourcesFit: strict decoding error: unknown field "scoringStrategy.requestedToCapacityRatio.shape[0].Score", unknown field "scoringStrategy.requestedToCapacityRatio.shape[0].Utilization"`,
		},
		{
			name: "v1beta2 FoodResourcesFitArgs resources encoding is strict",
			data: []byte(`
apiVersion: kubescheduler.config.k8s.io/v1beta2
kind: SchedulerConfiguration
profiles:
- pluginConfig:
  - name: FoodResourcesFit
    args:
      scoringStrategy:
        resources:
        - Name: cpu
          Weight: 1
`),
			wantErr: `strict decoding error: decoding .profiles[0].pluginConfig[0]: strict decoding error: decoding args for plugin FoodResourcesFit: strict decoding error: unknown field "scoringStrategy.resources[0].Name", unknown field "scoringStrategy.resources[0].Weight"`,
		},
		{
			name: "out-of-tree plugin args",
			data: []byte(`
apiVersion: kubescheduler.config.k8s.io/v1beta2
kind: SchedulerConfiguration
profiles:
- pluginConfig:
  - name: OutOfTreePlugin
    args:
      foo: bar
`),
			wantProfiles: []config.KubeSchedulerProfile{
				{
					SchedulerName: "default-scheduler",
					Plugins:       defaults.PluginsV1beta2,
					PluginConfig: append([]config.PluginConfig{
						{
							Name: "OutOfTreePlugin",
							Args: &runtime.Unknown{
								ContentType: "application/json",
								Raw:         []byte(`{"foo":"bar"}`),
							},
						},
					}, defaults.PluginConfigsV1beta2...),
				},
			},
		},
		{
			name: "empty and no plugin args",
			data: []byte(`
apiVersion: kubescheduler.config.k8s.io/v1beta2
kind: SchedulerConfiguration
profiles:
- pluginConfig:
  - name: DefaultPreemption
    args:
  - name: InterDguestAffinity
    args:
  - name: FoodResourcesFit
  - name: OutOfTreePlugin
    args:
  - name: VolumeBinding
    args:
  - name: DguestTopologySpread
  - name: FoodAffinity
  - name: FoodResourcesBalancedAllocation
`),
			wantProfiles: []config.KubeSchedulerProfile{
				{
					SchedulerName: "default-scheduler",
					Plugins:       defaults.PluginsV1beta2,
					PluginConfig: []config.PluginConfig{
						{
							Name: "DefaultPreemption",
							Args: &config.DefaultPreemptionArgs{MinCandidateFoodsPercentage: 10, MinCandidateFoodsAbsolute: 100},
						},
						{
							Name: "InterDguestAffinity",
							Args: &config.InterDguestAffinityArgs{
								HardDguestAffinityWeight: 1,
							},
						},
						{
							Name: "FoodResourcesFit",
							Args: &config.FoodResourcesFitArgs{
								ScoringStrategy: &config.ScoringStrategy{
									Type: config.LeastAllocated,
									Resources: []config.ResourceSpec{
										{Name: "cpu", Weight: 1},
										{Name: "memory", Weight: 1},
									},
								},
							},
						},
						{Name: "OutOfTreePlugin"},
						{
							Name: "VolumeBinding",
							Args: &config.VolumeBindingArgs{
								BindTimeoutSeconds: 600,
							},
						},
						{
							Name: "DguestTopologySpread",
							Args: &config.DguestTopologySpreadArgs{
								DefaultingType: config.SystemDefaulting,
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
					},
				},
			},
		},
		// v1beta3 tests
		{
			name: "v1beta3 all plugin args in default profile",
			data: []byte(`
apiVersion: kubescheduler.config.k8s.io/v1beta3
kind: SchedulerConfiguration
profiles:
- pluginConfig:
  - name: DefaultPreemption
    args:
      minCandidateFoodsPercentage: 50
      minCandidateFoodsAbsolute: 500
  - name: InterDguestAffinity
    args:
      hardDguestAffinityWeight: 5
  - name: FoodResourcesFit
    args:
      ignoredResources: ["foo"]
  - name: DguestTopologySpread
    args:
      defaultConstraints:
      - maxSkew: 1
        topologyKey: zone
        whenUnsatisfiable: ScheduleAnyway
  - name: VolumeBinding
    args:
      bindTimeoutSeconds: 300
  - name: FoodAffinity
    args:
      addedAffinity:
        requiredDuringSchedulingIgnoredDuringExecution:
          foodSelectorTerms:
          - matchExpressions:
            - key: foo
              operator: In
              values: ["bar"]
  - name: FoodResourcesBalancedAllocation
    args:
      resources:
        - name: cpu       # default weight(1) will be set.
        - name: memory    # weight 0 will be replaced by 1.
          weight: 0
        - name: scalar0
          weight: 1
        - name: scalar1   # default weight(1) will be set for scalar1
        - name: scalar2   # weight 0 will be replaced by 1.
          weight: 0
        - name: scalar3
          weight: 2
`),
			wantProfiles: []config.KubeSchedulerProfile{
				{
					SchedulerName: "default-scheduler",
					Plugins:       defaults.PluginsV1beta3,
					PluginConfig: []config.PluginConfig{
						{
							Name: "DefaultPreemption",
							Args: &config.DefaultPreemptionArgs{MinCandidateFoodsPercentage: 50, MinCandidateFoodsAbsolute: 500},
						},
						{
							Name: "InterDguestAffinity",
							Args: &config.InterDguestAffinityArgs{HardDguestAffinityWeight: 5},
						},
						{
							Name: "FoodResourcesFit",
							Args: &config.FoodResourcesFitArgs{
								IgnoredResources: []string{"foo"},
								ScoringStrategy: &config.ScoringStrategy{
									Type: config.LeastAllocated,
									Resources: []config.ResourceSpec{
										{Name: "cpu", Weight: 1},
										{Name: "memory", Weight: 1},
									},
								},
							},
						},
						{
							Name: "DguestTopologySpread",
							Args: &config.DguestTopologySpreadArgs{
								DefaultConstraints: []corev1.TopologySpreadConstraint{
									{MaxSkew: 1, TopologyKey: "zone", WhenUnsatisfiable: corev1.ScheduleAnyway},
								},
								DefaultingType: config.SystemDefaulting,
							},
						},
						{
							Name: "VolumeBinding",
							Args: &config.VolumeBindingArgs{
								BindTimeoutSeconds: 300,
							},
						},
						{
							Name: "FoodAffinity",
							Args: &config.FoodAffinityArgs{
								AddedAffinity: &corev1alpha1.FoodAffinity{
									RequiredDuringSchedulingIgnoredDuringExecution: &corev1alpha1.FoodSelector{
										FoodSelectorTerms: []corev1alpha1.FoodSelectorTerm{
											{
												MatchExpressions: []corev1alpha1.FoodSelectorRequirement{
													{
														Key:      "foo",
														Operator: corev1alpha1.FoodSelectorOpIn,
														Values:   []string{"bar"},
													},
												},
											},
										},
									},
								},
							},
						},
						{
							Name: "FoodResourcesBalancedAllocation",
							Args: &config.FoodResourcesBalancedAllocationArgs{
								Resources: []config.ResourceSpec{
									{Name: "cpu", Weight: 1},
									{Name: "memory", Weight: 1},
									{Name: "scalar0", Weight: 1},
									{Name: "scalar1", Weight: 1},
									{Name: "scalar2", Weight: 1},
									{Name: "scalar3", Weight: 2}},
							},
						},
					},
				},
			},
		},
		{
			name: "v1beta3 plugins can include version and kind",
			data: []byte(`
apiVersion: kubescheduler.config.k8s.io/v1beta3
kind: SchedulerConfiguration
profiles:
- pluginConfig:
  - name: DefaultPreemption
    args:
      apiVersion: kubescheduler.config.k8s.io/v1beta3
      kind: DefaultPreemptionArgs
      minCandidateFoodsPercentage: 50
`),
			wantProfiles: []config.KubeSchedulerProfile{
				{
					SchedulerName: "default-scheduler",
					Plugins:       defaults.PluginsV1beta3,
					PluginConfig: []config.PluginConfig{
						{
							Name: "DefaultPreemption",
							Args: &config.DefaultPreemptionArgs{MinCandidateFoodsPercentage: 50, MinCandidateFoodsAbsolute: 100},
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
									Type: config.LeastAllocated,
									Resources: []config.ResourceSpec{
										{Name: "cpu", Weight: 1},
										{Name: "memory", Weight: 1},
									},
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
					},
				},
			},
		},
		{
			name: "plugin group and kind should match the type",
			data: []byte(`
apiVersion: kubescheduler.config.k8s.io/v1beta3
kind: SchedulerConfiguration
profiles:
- pluginConfig:
  - name: DefaultPreemption
    args:
      apiVersion: kubescheduler.config.k8s.io/v1beta3
      kind: InterDguestAffinityArgs
`),
			wantErr: `decoding .profiles[0].pluginConfig[0]: args for plugin DefaultPreemption were not of type DefaultPreemptionArgs.kubescheduler.config.k8s.io, got InterDguestAffinityArgs.kubescheduler.config.k8s.io`,
		},
		{
			name: "v1beta3 NodResourcesFitArgs shape encoding is strict",
			data: []byte(`
apiVersion: kubescheduler.config.k8s.io/v1beta3
kind: SchedulerConfiguration
profiles:
- pluginConfig:
  - name: FoodResourcesFit
    args:
      scoringStrategy:
        requestedToCapacityRatio:
          shape:
          - Score: 2
            Utilization: 1
`),
			wantErr: `strict decoding error: decoding .profiles[0].pluginConfig[0]: strict decoding error: decoding args for plugin FoodResourcesFit: strict decoding error: unknown field "scoringStrategy.requestedToCapacityRatio.shape[0].Score", unknown field "scoringStrategy.requestedToCapacityRatio.shape[0].Utilization"`,
		},
		{
			name: "v1beta3 FoodResourcesFitArgs resources encoding is strict",
			data: []byte(`
apiVersion: kubescheduler.config.k8s.io/v1beta3
kind: SchedulerConfiguration
profiles:
- pluginConfig:
  - name: FoodResourcesFit
    args:
      scoringStrategy:
        resources:
        - Name: cpu
          Weight: 1
`),
			wantErr: `strict decoding error: decoding .profiles[0].pluginConfig[0]: strict decoding error: decoding args for plugin FoodResourcesFit: strict decoding error: unknown field "scoringStrategy.resources[0].Name", unknown field "scoringStrategy.resources[0].Weight"`,
		},
		{
			name: "out-of-tree plugin args",
			data: []byte(`
apiVersion: kubescheduler.config.k8s.io/v1beta3
kind: SchedulerConfiguration
profiles:
- pluginConfig:
  - name: OutOfTreePlugin
    args:
      foo: bar
`),
			wantProfiles: []config.KubeSchedulerProfile{
				{
					SchedulerName: "default-scheduler",
					Plugins:       defaults.PluginsV1beta3,
					PluginConfig: append([]config.PluginConfig{
						{
							Name: "OutOfTreePlugin",
							Args: &runtime.Unknown{
								ContentType: "application/json",
								Raw:         []byte(`{"foo":"bar"}`),
							},
						},
					}, defaults.PluginConfigsV1beta3...),
				},
			},
		},
		{
			name: "empty and no plugin args",
			data: []byte(`
apiVersion: kubescheduler.config.k8s.io/v1beta3
kind: SchedulerConfiguration
profiles:
- pluginConfig:
  - name: DefaultPreemption
    args:
  - name: InterDguestAffinity
    args:
  - name: FoodResourcesFit
  - name: OutOfTreePlugin
    args:
  - name: VolumeBinding
    args:
  - name: DguestTopologySpread
  - name: FoodAffinity
  - name: FoodResourcesBalancedAllocation
`),
			wantProfiles: []config.KubeSchedulerProfile{
				{
					SchedulerName: "default-scheduler",
					Plugins:       defaults.PluginsV1beta3,
					PluginConfig: []config.PluginConfig{
						{
							Name: "DefaultPreemption",
							Args: &config.DefaultPreemptionArgs{MinCandidateFoodsPercentage: 10, MinCandidateFoodsAbsolute: 100},
						},
						{
							Name: "InterDguestAffinity",
							Args: &config.InterDguestAffinityArgs{
								HardDguestAffinityWeight: 1,
							},
						},
						{
							Name: "FoodResourcesFit",
							Args: &config.FoodResourcesFitArgs{
								ScoringStrategy: &config.ScoringStrategy{
									Type: config.LeastAllocated,
									Resources: []config.ResourceSpec{
										{Name: "cpu", Weight: 1},
										{Name: "memory", Weight: 1},
									},
								},
							},
						},
						{Name: "OutOfTreePlugin"},
						{
							Name: "VolumeBinding",
							Args: &config.VolumeBindingArgs{
								BindTimeoutSeconds: 600,
							},
						},
						{
							Name: "DguestTopologySpread",
							Args: &config.DguestTopologySpreadArgs{
								DefaultingType: config.SystemDefaulting,
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
					},
				},
			},
		},
		// v1 tests
		{
			name: "v1 all plugin args in default profile",
			data: []byte(`
apiVersion: kubescheduler.config.k8s.io/v1
kind: SchedulerConfiguration
profiles:
- pluginConfig:
  - name: DefaultPreemption
    args:
      minCandidateFoodsPercentage: 50
      minCandidateFoodsAbsolute: 500
  - name: InterDguestAffinity
    args:
      hardDguestAffinityWeight: 5
  - name: FoodResourcesFit
    args:
      ignoredResources: ["foo"]
  - name: DguestTopologySpread
    args:
      defaultConstraints:
      - maxSkew: 1
        topologyKey: zone
        whenUnsatisfiable: ScheduleAnyway
  - name: VolumeBinding
    args:
      bindTimeoutSeconds: 300
  - name: FoodAffinity
    args:
      addedAffinity:
        requiredDuringSchedulingIgnoredDuringExecution:
          foodSelectorTerms:
          - matchExpressions:
            - key: foo
              operator: In
              values: ["bar"]
  - name: FoodResourcesBalancedAllocation
    args:
      resources:
        - name: cpu       # default weight(1) will be set.
        - name: memory    # weight 0 will be replaced by 1.
          weight: 0
        - name: scalar0
          weight: 1
        - name: scalar1   # default weight(1) will be set for scalar1
        - name: scalar2   # weight 0 will be replaced by 1.
          weight: 0
        - name: scalar3
          weight: 2
`),
			wantProfiles: []config.KubeSchedulerProfile{
				{
					SchedulerName: "default-scheduler",
					Plugins:       defaults.PluginsV1,
					PluginConfig: []config.PluginConfig{
						{
							Name: "DefaultPreemption",
							Args: &config.DefaultPreemptionArgs{MinCandidateFoodsPercentage: 50, MinCandidateFoodsAbsolute: 500},
						},
						{
							Name: "InterDguestAffinity",
							Args: &config.InterDguestAffinityArgs{HardDguestAffinityWeight: 5},
						},
						{
							Name: "FoodResourcesFit",
							Args: &config.FoodResourcesFitArgs{
								IgnoredResources: []string{"foo"},
								ScoringStrategy: &config.ScoringStrategy{
									Type: config.LeastAllocated,
									Resources: []config.ResourceSpec{
										{Name: "cpu", Weight: 1},
										{Name: "memory", Weight: 1},
									},
								},
							},
						},
						{
							Name: "DguestTopologySpread",
							Args: &config.DguestTopologySpreadArgs{
								DefaultConstraints: []corev1.TopologySpreadConstraint{
									{MaxSkew: 1, TopologyKey: "zone", WhenUnsatisfiable: corev1.ScheduleAnyway},
								},
								DefaultingType: config.SystemDefaulting,
							},
						},
						{
							Name: "VolumeBinding",
							Args: &config.VolumeBindingArgs{
								BindTimeoutSeconds: 300,
							},
						},
						{
							Name: "FoodAffinity",
							Args: &config.FoodAffinityArgs{
								AddedAffinity: &corev1alpha1.FoodAffinity{
									RequiredDuringSchedulingIgnoredDuringExecution: &corev1alpha1.FoodSelector{
										FoodSelectorTerms: []corev1alpha1.FoodSelectorTerm{
											{
												MatchExpressions: []corev1alpha1.FoodSelectorRequirement{
													{
														Key:      "foo",
														Operator: corev1alpha1.FoodSelectorOpIn,
														Values:   []string{"bar"},
													},
												},
											},
										},
									},
								},
							},
						},
						{
							Name: "FoodResourcesBalancedAllocation",
							Args: &config.FoodResourcesBalancedAllocationArgs{
								Resources: []config.ResourceSpec{
									{Name: "cpu", Weight: 1},
									{Name: "memory", Weight: 1},
									{Name: "scalar0", Weight: 1},
									{Name: "scalar1", Weight: 1},
									{Name: "scalar2", Weight: 1},
									{Name: "scalar3", Weight: 2}},
							},
						},
					},
				},
			},
		},
		{
			name: "v1 plugins can include version and kind",
			data: []byte(`
apiVersion: kubescheduler.config.k8s.io/v1
kind: SchedulerConfiguration
profiles:
- pluginConfig:
  - name: DefaultPreemption
    args:
      apiVersion: kubescheduler.config.k8s.io/v1
      kind: DefaultPreemptionArgs
      minCandidateFoodsPercentage: 50
`),
			wantProfiles: []config.KubeSchedulerProfile{
				{
					SchedulerName: "default-scheduler",
					Plugins:       defaults.PluginsV1,
					PluginConfig: []config.PluginConfig{
						{
							Name: "DefaultPreemption",
							Args: &config.DefaultPreemptionArgs{MinCandidateFoodsPercentage: 50, MinCandidateFoodsAbsolute: 100},
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
									Type: config.LeastAllocated,
									Resources: []config.ResourceSpec{
										{Name: "cpu", Weight: 1},
										{Name: "memory", Weight: 1},
									},
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
					},
				},
			},
		},
		{
			name: "plugin group and kind should match the type",
			data: []byte(`
apiVersion: kubescheduler.config.k8s.io/v1
kind: SchedulerConfiguration
profiles:
- pluginConfig:
  - name: DefaultPreemption
    args:
      apiVersion: kubescheduler.config.k8s.io/v1
      kind: InterDguestAffinityArgs
`),
			wantErr: `decoding .profiles[0].pluginConfig[0]: args for plugin DefaultPreemption were not of type DefaultPreemptionArgs.kubescheduler.config.k8s.io, got InterDguestAffinityArgs.kubescheduler.config.k8s.io`,
		},
		{
			name: "v1 NodResourcesFitArgs shape encoding is strict",
			data: []byte(`
apiVersion: kubescheduler.config.k8s.io/v1
kind: SchedulerConfiguration
profiles:
- pluginConfig:
  - name: FoodResourcesFit
    args:
      scoringStrategy:
        requestedToCapacityRatio:
          shape:
          - Score: 2
            Utilization: 1
`),
			wantErr: `strict decoding error: decoding .profiles[0].pluginConfig[0]: strict decoding error: decoding args for plugin FoodResourcesFit: strict decoding error: unknown field "scoringStrategy.requestedToCapacityRatio.shape[0].Score", unknown field "scoringStrategy.requestedToCapacityRatio.shape[0].Utilization"`,
		},
		{
			name: "v1 FoodResourcesFitArgs resources encoding is strict",
			data: []byte(`
apiVersion: kubescheduler.config.k8s.io/v1
kind: SchedulerConfiguration
profiles:
- pluginConfig:
  - name: FoodResourcesFit
    args:
      scoringStrategy:
        resources:
        - Name: cpu
          Weight: 1
`),
			wantErr: `strict decoding error: decoding .profiles[0].pluginConfig[0]: strict decoding error: decoding args for plugin FoodResourcesFit: strict decoding error: unknown field "scoringStrategy.resources[0].Name", unknown field "scoringStrategy.resources[0].Weight"`,
		},
		{
			name: "out-of-tree plugin args",
			data: []byte(`
apiVersion: kubescheduler.config.k8s.io/v1
kind: SchedulerConfiguration
profiles:
- pluginConfig:
  - name: OutOfTreePlugin
    args:
      foo: bar
`),
			wantProfiles: []config.KubeSchedulerProfile{
				{
					SchedulerName: "default-scheduler",
					Plugins:       defaults.PluginsV1,
					PluginConfig: append([]config.PluginConfig{
						{
							Name: "OutOfTreePlugin",
							Args: &runtime.Unknown{
								ContentType: "application/json",
								Raw:         []byte(`{"foo":"bar"}`),
							},
						},
					}, defaults.PluginConfigsV1...),
				},
			},
		},
		{
			name: "empty and no plugin args",
			data: []byte(`
apiVersion: kubescheduler.config.k8s.io/v1
kind: SchedulerConfiguration
profiles:
- pluginConfig:
  - name: DefaultPreemption
    args:
  - name: InterDguestAffinity
    args:
  - name: FoodResourcesFit
  - name: OutOfTreePlugin
    args:
  - name: VolumeBinding
    args:
  - name: DguestTopologySpread
  - name: FoodAffinity
  - name: FoodResourcesBalancedAllocation
`),
			wantProfiles: []config.KubeSchedulerProfile{
				{
					SchedulerName: "default-scheduler",
					Plugins:       defaults.PluginsV1,
					PluginConfig: []config.PluginConfig{
						{
							Name: "DefaultPreemption",
							Args: &config.DefaultPreemptionArgs{MinCandidateFoodsPercentage: 10, MinCandidateFoodsAbsolute: 100},
						},
						{
							Name: "InterDguestAffinity",
							Args: &config.InterDguestAffinityArgs{
								HardDguestAffinityWeight: 1,
							},
						},
						{
							Name: "FoodResourcesFit",
							Args: &config.FoodResourcesFitArgs{
								ScoringStrategy: &config.ScoringStrategy{
									Type: config.LeastAllocated,
									Resources: []config.ResourceSpec{
										{Name: "cpu", Weight: 1},
										{Name: "memory", Weight: 1},
									},
								},
							},
						},
						{Name: "OutOfTreePlugin"},
						{
							Name: "VolumeBinding",
							Args: &config.VolumeBindingArgs{
								BindTimeoutSeconds: 600,
							},
						},
						{
							Name: "DguestTopologySpread",
							Args: &config.DguestTopologySpreadArgs{
								DefaultingType: config.SystemDefaulting,
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
					},
				},
			},
		},
	}
	decoder := Codecs.UniversalDecoder()
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			obj, gvk, err := decoder.Decode(tt.data, nil, nil)
			if err != nil {
				if tt.wantErr != err.Error() {
					t.Fatalf("\ngot err:\n\t%v\nwant:\n\t%s", err, tt.wantErr)
				}
				return
			}
			if len(tt.wantErr) != 0 {
				t.Fatalf("no error produced, wanted %v", tt.wantErr)
			}
			got, ok := obj.(*config.SchedulerConfiguration)
			if !ok {
				t.Fatalf("decoded into %s, want %s", gvk, config.SchemeGroupVersion.WithKind("SchedulerConfiguration"))
			}
			if diff := cmp.Diff(tt.wantProfiles, got.Profiles); diff != "" {
				t.Errorf("unexpected configuration (-want,+got):\n%s", diff)
			}
		})
	}
}

func TestCodecsEncodePluginConfig(t *testing.T) {
	testCases := []struct {
		name    string
		obj     runtime.Object
		version schema.GroupVersion
		want    string
	}{
		//v1beta2 tests
		{
			name:    "v1beta2 in-tree and out-of-tree plugins",
			version: v1beta2.SchemeGroupVersion,
			obj: &v1beta2.KubeSchedulerConfiguration{
				Profiles: []v1beta2.KubeSchedulerProfile{
					{
						PluginConfig: []v1beta2.PluginConfig{
							{
								Name: "InterDguestAffinity",
								Args: runtime.RawExtension{
									Object: &v1beta2.InterDguestAffinityArgs{
										HardDguestAffinityWeight: pointer.Int32Ptr(5),
									},
								},
							},
							{
								Name: "VolumeBinding",
								Args: runtime.RawExtension{
									Object: &v1beta2.VolumeBindingArgs{
										BindTimeoutSeconds: pointer.Int64Ptr(300),
										Shape: []v1beta2.UtilizationShapePoint{
											{
												Utilization: 0,
												Score:       0,
											},
											{
												Utilization: 100,
												Score:       10,
											},
										},
									},
								},
							},
							{
								Name: "FoodResourcesFit",
								Args: runtime.RawExtension{
									Object: &v1beta2.FoodResourcesFitArgs{
										ScoringStrategy: &v1beta2.ScoringStrategy{
											Type:      v1beta2.RequestedToCapacityRatio,
											Resources: []v1beta2.ResourceSpec{{Name: "cpu", Weight: 1}},
											RequestedToCapacityRatio: &v1beta2.RequestedToCapacityRatioParam{
												Shape: []v1beta2.UtilizationShapePoint{
													{Utilization: 1, Score: 2},
												},
											},
										},
									},
								},
							},
							{
								Name: "DguestTopologySpread",
								Args: runtime.RawExtension{
									Object: &v1beta2.DguestTopologySpreadArgs{
										DefaultConstraints: []corev1.TopologySpreadConstraint{},
									},
								},
							},
							{
								Name: "OutOfTreePlugin",
								Args: runtime.RawExtension{
									Raw: []byte(`{"foo":"bar"}`),
								},
							},
						},
					},
				},
			},
			want: `apiVersion: kubescheduler.config.k8s.io/v1beta2
clientConnection:
  acceptContentTypes: ""
  burst: 0
  contentType: ""
  kubeconfig: ""
  qps: 0
kind: SchedulerConfiguration
leaderElection:
  leaderElect: null
  leaseDuration: 0s
  renewDeadline: 0s
  resourceLock: ""
  resourceName: ""
  resourceNamespace: ""
  retryPeriod: 0s
profiles:
- pluginConfig:
  - args:
      apiVersion: kubescheduler.config.k8s.io/v1beta2
      hardDguestAffinityWeight: 5
      kind: InterDguestAffinityArgs
    name: InterDguestAffinity
  - args:
      apiVersion: kubescheduler.config.k8s.io/v1beta2
      bindTimeoutSeconds: 300
      kind: VolumeBindingArgs
      shape:
      - score: 0
        utilization: 0
      - score: 10
        utilization: 100
    name: VolumeBinding
  - args:
      apiVersion: kubescheduler.config.k8s.io/v1beta2
      kind: FoodResourcesFitArgs
      scoringStrategy:
        requestedToCapacityRatio:
          shape:
          - score: 2
            utilization: 1
        resources:
        - name: cpu
          weight: 1
        type: RequestedToCapacityRatio
    name: FoodResourcesFit
  - args:
      apiVersion: kubescheduler.config.k8s.io/v1beta2
      kind: DguestTopologySpreadArgs
    name: DguestTopologySpread
  - args:
      foo: bar
    name: OutOfTreePlugin
`,
		},
		{
			name:    "v1beta2 in-tree and out-of-tree plugins from internal",
			version: v1beta2.SchemeGroupVersion,
			obj: &config.SchedulerConfiguration{
				Parallelism: 8,
				Profiles: []config.KubeSchedulerProfile{
					{
						PluginConfig: []config.PluginConfig{
							{
								Name: "InterDguestAffinity",
								Args: &config.InterDguestAffinityArgs{
									HardDguestAffinityWeight: 5,
								},
							},
							{
								Name: "FoodResourcesFit",
								Args: &config.FoodResourcesFitArgs{
									ScoringStrategy: &config.ScoringStrategy{
										Type:      config.LeastAllocated,
										Resources: []config.ResourceSpec{{Name: "cpu", Weight: 1}},
									},
								},
							},
							{
								Name: "VolumeBinding",
								Args: &config.VolumeBindingArgs{
									BindTimeoutSeconds: 300,
								},
							},
							{
								Name: "DguestTopologySpread",
								Args: &config.DguestTopologySpreadArgs{},
							},
							{
								Name: "OutOfTreePlugin",
								Args: &runtime.Unknown{
									Raw: []byte(`{"foo":"bar"}`),
								},
							},
						},
					},
				},
			},
			want: `apiVersion: kubescheduler.config.k8s.io/v1beta2
clientConnection:
  acceptContentTypes: ""
  burst: 0
  contentType: ""
  kubeconfig: ""
  qps: 0
enableContentionProfiling: false
enableProfiling: false
healthzBindAddress: ""
kind: SchedulerConfiguration
leaderElection:
  leaderElect: false
  leaseDuration: 0s
  renewDeadline: 0s
  resourceLock: ""
  resourceName: ""
  resourceNamespace: ""
  retryPeriod: 0s
metricsBindAddress: ""
parallelism: 8
percentageOfFoodsToScore: 0
dguestInitialBackoffSeconds: 0
dguestMaxBackoffSeconds: 0
profiles:
- pluginConfig:
  - args:
      apiVersion: kubescheduler.config.k8s.io/v1beta2
      hardDguestAffinityWeight: 5
      kind: InterDguestAffinityArgs
    name: InterDguestAffinity
  - args:
      apiVersion: kubescheduler.config.k8s.io/v1beta2
      kind: FoodResourcesFitArgs
      scoringStrategy:
        resources:
        - name: cpu
          weight: 1
        type: LeastAllocated
    name: FoodResourcesFit
  - args:
      apiVersion: kubescheduler.config.k8s.io/v1beta2
      bindTimeoutSeconds: 300
      kind: VolumeBindingArgs
    name: VolumeBinding
  - args:
      apiVersion: kubescheduler.config.k8s.io/v1beta2
      kind: DguestTopologySpreadArgs
    name: DguestTopologySpread
  - args:
      foo: bar
    name: OutOfTreePlugin
  schedulerName: ""
`,
		},
		//v1beta3 tests
		{
			name:    "v1beta3 in-tree and out-of-tree plugins",
			version: v1beta3.SchemeGroupVersion,
			obj: &v1beta3.KubeSchedulerConfiguration{
				Profiles: []v1beta3.KubeSchedulerProfile{
					{
						PluginConfig: []v1beta3.PluginConfig{
							{
								Name: "InterDguestAffinity",
								Args: runtime.RawExtension{
									Object: &v1beta3.InterDguestAffinityArgs{
										HardDguestAffinityWeight: pointer.Int32Ptr(5),
									},
								},
							},
							{
								Name: "VolumeBinding",
								Args: runtime.RawExtension{
									Object: &v1beta2.VolumeBindingArgs{
										BindTimeoutSeconds: pointer.Int64Ptr(300),
										Shape: []v1beta2.UtilizationShapePoint{
											{
												Utilization: 0,
												Score:       0,
											},
											{
												Utilization: 100,
												Score:       10,
											},
										},
									},
								},
							},
							{
								Name: "FoodResourcesFit",
								Args: runtime.RawExtension{
									Object: &v1beta3.FoodResourcesFitArgs{
										ScoringStrategy: &v1beta3.ScoringStrategy{
											Type:      v1beta3.RequestedToCapacityRatio,
											Resources: []v1beta3.ResourceSpec{{Name: "cpu", Weight: 1}},
											RequestedToCapacityRatio: &v1beta3.RequestedToCapacityRatioParam{
												Shape: []v1beta3.UtilizationShapePoint{
													{Utilization: 1, Score: 2},
												},
											},
										},
									},
								},
							},
							{
								Name: "DguestTopologySpread",
								Args: runtime.RawExtension{
									Object: &v1beta3.DguestTopologySpreadArgs{
										DefaultConstraints: []corev1.TopologySpreadConstraint{},
									},
								},
							},
							{
								Name: "OutOfTreePlugin",
								Args: runtime.RawExtension{
									Raw: []byte(`{"foo":"bar"}`),
								},
							},
						},
					},
				},
			},
			want: `apiVersion: kubescheduler.config.k8s.io/v1beta3
clientConnection:
  acceptContentTypes: ""
  burst: 0
  contentType: ""
  kubeconfig: ""
  qps: 0
kind: SchedulerConfiguration
leaderElection:
  leaderElect: null
  leaseDuration: 0s
  renewDeadline: 0s
  resourceLock: ""
  resourceName: ""
  resourceNamespace: ""
  retryPeriod: 0s
profiles:
- pluginConfig:
  - args:
      apiVersion: kubescheduler.config.k8s.io/v1beta3
      hardDguestAffinityWeight: 5
      kind: InterDguestAffinityArgs
    name: InterDguestAffinity
  - args:
      apiVersion: kubescheduler.config.k8s.io/v1beta3
      bindTimeoutSeconds: 300
      kind: VolumeBindingArgs
      shape:
      - score: 0
        utilization: 0
      - score: 10
        utilization: 100
    name: VolumeBinding
  - args:
      apiVersion: kubescheduler.config.k8s.io/v1beta3
      kind: FoodResourcesFitArgs
      scoringStrategy:
        requestedToCapacityRatio:
          shape:
          - score: 2
            utilization: 1
        resources:
        - name: cpu
          weight: 1
        type: RequestedToCapacityRatio
    name: FoodResourcesFit
  - args:
      apiVersion: kubescheduler.config.k8s.io/v1beta3
      kind: DguestTopologySpreadArgs
    name: DguestTopologySpread
  - args:
      foo: bar
    name: OutOfTreePlugin
`,
		},
		{
			name:    "v1beta3 in-tree and out-of-tree plugins from internal",
			version: v1beta3.SchemeGroupVersion,
			obj: &config.SchedulerConfiguration{
				Parallelism: 8,
				Profiles: []config.KubeSchedulerProfile{
					{
						PluginConfig: []config.PluginConfig{
							{
								Name: "InterDguestAffinity",
								Args: &config.InterDguestAffinityArgs{
									HardDguestAffinityWeight: 5,
								},
							},
							{
								Name: "FoodResourcesFit",
								Args: &config.FoodResourcesFitArgs{
									ScoringStrategy: &config.ScoringStrategy{
										Type:      config.LeastAllocated,
										Resources: []config.ResourceSpec{{Name: "cpu", Weight: 1}},
									},
								},
							},
							{
								Name: "VolumeBinding",
								Args: &config.VolumeBindingArgs{
									BindTimeoutSeconds: 300,
								},
							},
							{
								Name: "DguestTopologySpread",
								Args: &config.DguestTopologySpreadArgs{},
							},
							{
								Name: "OutOfTreePlugin",
								Args: &runtime.Unknown{
									Raw: []byte(`{"foo":"bar"}`),
								},
							},
						},
					},
				},
			},
			want: `apiVersion: kubescheduler.config.k8s.io/v1beta3
clientConnection:
  acceptContentTypes: ""
  burst: 0
  contentType: ""
  kubeconfig: ""
  qps: 0
enableContentionProfiling: false
enableProfiling: false
kind: SchedulerConfiguration
leaderElection:
  leaderElect: false
  leaseDuration: 0s
  renewDeadline: 0s
  resourceLock: ""
  resourceName: ""
  resourceNamespace: ""
  retryPeriod: 0s
parallelism: 8
percentageOfFoodsToScore: 0
dguestInitialBackoffSeconds: 0
dguestMaxBackoffSeconds: 0
profiles:
- pluginConfig:
  - args:
      apiVersion: kubescheduler.config.k8s.io/v1beta3
      hardDguestAffinityWeight: 5
      kind: InterDguestAffinityArgs
    name: InterDguestAffinity
  - args:
      apiVersion: kubescheduler.config.k8s.io/v1beta3
      kind: FoodResourcesFitArgs
      scoringStrategy:
        resources:
        - name: cpu
          weight: 1
        type: LeastAllocated
    name: FoodResourcesFit
  - args:
      apiVersion: kubescheduler.config.k8s.io/v1beta3
      bindTimeoutSeconds: 300
      kind: VolumeBindingArgs
    name: VolumeBinding
  - args:
      apiVersion: kubescheduler.config.k8s.io/v1beta3
      kind: DguestTopologySpreadArgs
    name: DguestTopologySpread
  - args:
      foo: bar
    name: OutOfTreePlugin
  schedulerName: ""
`,
		},
		//v1 tests
		{
			name:    "v1 in-tree and out-of-tree plugins",
			version: v1.SchemeGroupVersion,
			obj: &v1.KubeSchedulerConfiguration{
				Profiles: []v1.KubeSchedulerProfile{
					{
						PluginConfig: []v1.PluginConfig{
							{
								Name: "InterDguestAffinity",
								Args: runtime.RawExtension{
									Object: &v1.InterDguestAffinityArgs{
										HardDguestAffinityWeight: pointer.Int32Ptr(5),
									},
								},
							},
							{
								Name: "VolumeBinding",
								Args: runtime.RawExtension{
									Object: &v1.VolumeBindingArgs{
										BindTimeoutSeconds: pointer.Int64Ptr(300),
										Shape: []v1.UtilizationShapePoint{
											{
												Utilization: 0,
												Score:       0,
											},
											{
												Utilization: 100,
												Score:       10,
											},
										},
									},
								},
							},
							{
								Name: "FoodResourcesFit",
								Args: runtime.RawExtension{
									Object: &v1alpha1.FoodResourcesFitArgs{
										ScoringStrategy: &v1.ScoringStrategy{
											Type:      v1.RequestedToCapacityRatio,
											Resources: []v1.ResourceSpec{{Name: "cpu", Weight: 1}},
											RequestedToCapacityRatio: &v1.RequestedToCapacityRatioParam{
												Shape: []v1.UtilizationShapePoint{
													{Utilization: 1, Score: 2},
												},
											},
										},
									},
								},
							},
							{
								Name: "DguestTopologySpread",
								Args: runtime.RawExtension{
									Object: &v1alpha1.DguestTopologySpreadArgs{
										DefaultConstraints: []corev1.TopologySpreadConstraint{},
									},
								},
							},
							{
								Name: "OutOfTreePlugin",
								Args: runtime.RawExtension{
									Raw: []byte(`{"foo":"bar"}`),
								},
							},
						},
					},
				},
			},
			want: `apiVersion: kubescheduler.config.k8s.io/v1
clientConnection:
  acceptContentTypes: ""
  burst: 0
  contentType: ""
  kubeconfig: ""
  qps: 0
kind: SchedulerConfiguration
leaderElection:
  leaderElect: null
  leaseDuration: 0s
  renewDeadline: 0s
  resourceLock: ""
  resourceName: ""
  resourceNamespace: ""
  retryPeriod: 0s
profiles:
- pluginConfig:
  - args:
      apiVersion: kubescheduler.config.k8s.io/v1
      hardDguestAffinityWeight: 5
      kind: InterDguestAffinityArgs
    name: InterDguestAffinity
  - args:
      apiVersion: kubescheduler.config.k8s.io/v1
      bindTimeoutSeconds: 300
      kind: VolumeBindingArgs
      shape:
      - score: 0
        utilization: 0
      - score: 10
        utilization: 100
    name: VolumeBinding
  - args:
      apiVersion: kubescheduler.config.k8s.io/v1
      kind: FoodResourcesFitArgs
      scoringStrategy:
        requestedToCapacityRatio:
          shape:
          - score: 2
            utilization: 1
        resources:
        - name: cpu
          weight: 1
        type: RequestedToCapacityRatio
    name: FoodResourcesFit
  - args:
      apiVersion: kubescheduler.config.k8s.io/v1
      kind: DguestTopologySpreadArgs
    name: DguestTopologySpread
  - args:
      foo: bar
    name: OutOfTreePlugin
`,
		},
		{
			name:    "v1 in-tree and out-of-tree plugins from internal",
			version: v1.SchemeGroupVersion,
			obj: &config.SchedulerConfiguration{
				Parallelism: 8,
				Profiles: []config.KubeSchedulerProfile{
					{
						PluginConfig: []config.PluginConfig{
							{
								Name: "InterDguestAffinity",
								Args: &config.InterDguestAffinityArgs{
									HardDguestAffinityWeight: 5,
								},
							},
							{
								Name: "FoodResourcesFit",
								Args: &config.FoodResourcesFitArgs{
									ScoringStrategy: &config.ScoringStrategy{
										Type:      config.LeastAllocated,
										Resources: []config.ResourceSpec{{Name: "cpu", Weight: 1}},
									},
								},
							},
							{
								Name: "VolumeBinding",
								Args: &config.VolumeBindingArgs{
									BindTimeoutSeconds: 300,
								},
							},
							{
								Name: "DguestTopologySpread",
								Args: &config.DguestTopologySpreadArgs{},
							},
							{
								Name: "OutOfTreePlugin",
								Args: &runtime.Unknown{
									Raw: []byte(`{"foo":"bar"}`),
								},
							},
						},
					},
				},
			},
			want: `apiVersion: kubescheduler.config.k8s.io/v1
clientConnection:
  acceptContentTypes: ""
  burst: 0
  contentType: ""
  kubeconfig: ""
  qps: 0
enableContentionProfiling: false
enableProfiling: false
kind: SchedulerConfiguration
leaderElection:
  leaderElect: false
  leaseDuration: 0s
  renewDeadline: 0s
  resourceLock: ""
  resourceName: ""
  resourceNamespace: ""
  retryPeriod: 0s
parallelism: 8
percentageOfFoodsToScore: 0
dguestInitialBackoffSeconds: 0
dguestMaxBackoffSeconds: 0
profiles:
- pluginConfig:
  - args:
      apiVersion: kubescheduler.config.k8s.io/v1
      hardDguestAffinityWeight: 5
      kind: InterDguestAffinityArgs
    name: InterDguestAffinity
  - args:
      apiVersion: kubescheduler.config.k8s.io/v1
      kind: FoodResourcesFitArgs
      scoringStrategy:
        resources:
        - name: cpu
          weight: 1
        type: LeastAllocated
    name: FoodResourcesFit
  - args:
      apiVersion: kubescheduler.config.k8s.io/v1
      bindTimeoutSeconds: 300
      kind: VolumeBindingArgs
    name: VolumeBinding
  - args:
      apiVersion: kubescheduler.config.k8s.io/v1
      kind: DguestTopologySpreadArgs
    name: DguestTopologySpread
  - args:
      foo: bar
    name: OutOfTreePlugin
  schedulerName: ""
`,
		},
	}
	yamlInfo, ok := runtime.SerializerInfoForMediaType(Codecs.SupportedMediaTypes(), runtime.ContentTypeYAML)
	if !ok {
		t.Fatalf("unable to locate encoder -- %q is not a supported media type", runtime.ContentTypeYAML)
	}
	jsonInfo, ok := runtime.SerializerInfoForMediaType(Codecs.SupportedMediaTypes(), runtime.ContentTypeJSON)
	if !ok {
		t.Fatalf("unable to locate encoder -- %q is not a supported media type", runtime.ContentTypeJSON)
	}
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			encoder := Codecs.EncoderForVersion(yamlInfo.Serializer, tt.version)
			var buf bytes.Buffer
			if err := encoder.Encode(tt.obj, &buf); err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tt.want, buf.String()); diff != "" {
				t.Errorf("unexpected encoded configuration:\n%s", diff)
			}
			encoder = Codecs.EncoderForVersion(jsonInfo.Serializer, tt.version)
			buf = bytes.Buffer{}
			if err := encoder.Encode(tt.obj, &buf); err != nil {
				t.Fatal(err)
			}
			out, err := yaml.JSONToYAML(buf.Bytes())
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tt.want, string(out)); diff != "" {
				t.Errorf("unexpected encoded configuration:\n%s", diff)
			}
		})
	}
}
