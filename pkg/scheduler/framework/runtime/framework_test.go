// /*
// Copyright 2019 The Kubernetes Authors.

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// */

package runtime

// import (
// 	"context"
// 	"dguest-scheduler/pkg/apis/scheduler/v1alpha1"
// 	v1 "dguest-scheduler/pkg/scheduler/apis/config/v1"
// 	"errors"
// 	"fmt"
// 	"reflect"
// 	"strings"
// 	"testing"
// 	"time"

// 	"dguest-scheduler/pkg/scheduler/framework"
// 	internalqueue "dguest-scheduler/pkg/scheduler/internal/queue"
// 	"dguest-scheduler/pkg/scheduler/metrics"

// 	"github.com/google/go-cmp/cmp"
// 	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
// 	"k8s.io/apimachinery/pkg/runtime"
// 	"k8s.io/apimachinery/pkg/types"
// 	"k8s.io/apimachinery/pkg/util/sets"
// 	"k8s.io/apimachinery/pkg/util/wait"
// 	"k8s.io/component-base/metrics/testutil"
// )

// const (
// 	queueSortPlugin                   = "no-op-queue-sort-plugin"
// 	scoreWithNormalizePlugin1         = "score-with-normalize-plugin-1"
// 	scoreWithNormalizePlugin2         = "score-with-normalize-plugin-2"
// 	scorePlugin1                      = "score-plugin-1"
// 	pluginNotImplementingScore        = "plugin-not-implementing-score"
// 	preFilterPluginName               = "prefilter-plugin"
// 	preFilterWithExtensionsPluginName = "prefilter-with-extensions-plugin"
// 	duplicatePluginName               = "duplicate-plugin"
// 	testPlugin                        = "test-plugin"
// 	permitPlugin                      = "permit-plugin"
// 	bindPlugin                        = "bind-plugin"

// 	testProfileName = "test-profile"
// 	foodName        = "testFood"
// )

// // TestScoreWithNormalizePlugin implements ScoreWithNormalizePlugin interface.
// // TestScorePlugin only implements ScorePlugin interface.
// var _ framework.ScorePlugin = &TestScoreWithNormalizePlugin{}
// var _ framework.ScorePlugin = &TestScorePlugin{}

// func newScoreWithNormalizePlugin1(injArgs runtime.Object, f framework.Handle) (framework.Plugin, error) {
// 	var inj injectedResult
// 	if err := DecodeInto(injArgs, &inj); err != nil {
// 		return nil, err
// 	}
// 	return &TestScoreWithNormalizePlugin{scoreWithNormalizePlugin1, inj}, nil
// }

// func newScoreWithNormalizePlugin2(injArgs runtime.Object, f framework.Handle) (framework.Plugin, error) {
// 	var inj injectedResult
// 	if err := DecodeInto(injArgs, &inj); err != nil {
// 		return nil, err
// 	}
// 	return &TestScoreWithNormalizePlugin{scoreWithNormalizePlugin2, inj}, nil
// }

// func newScorePlugin1(injArgs runtime.Object, f framework.Handle) (framework.Plugin, error) {
// 	var inj injectedResult
// 	if err := DecodeInto(injArgs, &inj); err != nil {
// 		return nil, err
// 	}
// 	return &TestScorePlugin{scorePlugin1, inj}, nil
// }

// func newPluginNotImplementingScore(_ runtime.Object, _ framework.Handle) (framework.Plugin, error) {
// 	return &PluginNotImplementingScore{}, nil
// }

// type TestScoreWithNormalizePlugin struct {
// 	name string
// 	inj  injectedResult
// }

// func (pl *TestScoreWithNormalizePlugin) Name() string {
// 	return pl.name
// }

// func (pl *TestScoreWithNormalizePlugin) NormalizeScore(ctx context.Context, state *framework.CycleState, dguest *v1alpha1.Dguest, scores framework.FoodScoreList) *framework.Status {
// 	return injectNormalizeRes(pl.inj, scores)
// }

// func (pl *TestScoreWithNormalizePlugin) Score(ctx context.Context, state *framework.CycleState, p *v1alpha1.Dguest, selectedFood *v1alpha1.FoodInfoBase) (int64, *framework.Status) {
// 	return setScoreRes(pl.inj)
// }

// func (pl *TestScoreWithNormalizePlugin) ScoreExtensions() framework.ScoreExtensions {
// 	return pl
// }

// // TestScorePlugin only implements ScorePlugin interface.
// type TestScorePlugin struct {
// 	name string
// 	inj  injectedResult
// }

// func (pl *TestScorePlugin) Name() string {
// 	return pl.name
// }

// func (pl *TestScorePlugin) PreScore(ctx context.Context, state *framework.CycleState, dguest *v1alpha1.Dguest, foods []*v1alpha1.Food) *framework.Status {
// 	return framework.NewStatus(framework.Code(pl.inj.PreScoreStatus), "injected status")
// }

// func (pl *TestScorePlugin) Score(ctx context.Context, state *framework.CycleState, p *v1alpha1.Dguest, selectedFood *v1alpha1.FoodInfoBase) (int64, *framework.Status) {
// 	return setScoreRes(pl.inj)
// }

// func (pl *TestScorePlugin) ScoreExtensions() framework.ScoreExtensions {
// 	return nil
// }

// // PluginNotImplementingScore doesn't implement the ScorePlugin interface.
// type PluginNotImplementingScore struct{}

// func (pl *PluginNotImplementingScore) Name() string {
// 	return pluginNotImplementingScore
// }

// func newTestPlugin(injArgs runtime.Object, f framework.Handle) (framework.Plugin, error) {
// 	return &TestPlugin{name: testPlugin}, nil
// }

// // TestPlugin implements all Plugin interfaces.
// type TestPlugin struct {
// 	name string
// 	inj  injectedResult
// }

// func (pl *TestPlugin) AddDguest(ctx context.Context, state *framework.CycleState, dguestToSchedule *v1alpha1.Dguest, dguestInfoToAdd *framework.DguestInfo, foodInfo *framework.FoodInfo) *framework.Status {
// 	return framework.NewStatus(framework.Code(pl.inj.PreFilterAddDguestStatus), "injected status")
// }
// func (pl *TestPlugin) RemoveDguest(ctx context.Context, state *framework.CycleState, dguestToSchedule *v1alpha1.Dguest, dguestInfoToRemove *framework.DguestInfo, foodInfo *framework.FoodInfo) *framework.Status {
// 	return framework.NewStatus(framework.Code(pl.inj.PreFilterRemoveDguestStatus), "injected status")
// }

// func (pl *TestPlugin) Name() string {
// 	return pl.name
// }

// func (pl *TestPlugin) Less(*framework.QueuedDguestInfo, *framework.QueuedDguestInfo) bool {
// 	return false
// }

// func (pl *TestPlugin) Score(ctx context.Context, state *framework.CycleState, p *v1alpha1.Dguest, selectedFood *v1alpha1.FoodInfoBase) (int64, *framework.Status) {
// 	return 0, framework.NewStatus(framework.Code(pl.inj.ScoreStatus), "injected status")
// }

// func (pl *TestPlugin) ScoreExtensions() framework.ScoreExtensions {
// 	return nil
// }

// func (pl *TestPlugin) PreFilter(ctx context.Context, state *framework.CycleState, p *v1alpha1.Dguest) (*framework.PreFilterResult, *framework.Status) {
// 	return nil, framework.NewStatus(framework.Code(pl.inj.PreFilterStatus), "injected status")
// }

// func (pl *TestPlugin) PreFilterExtensions() framework.PreFilterExtensions {
// 	return pl
// }

// func (pl *TestPlugin) Filter(ctx context.Context, state *framework.CycleState, dguest *v1alpha1.Dguest, foodInfo *framework.FoodInfo) *framework.Status {
// 	return framework.NewStatus(framework.Code(pl.inj.FilterStatus), "injected filter status")
// }

// func (pl *TestPlugin) PostFilter(_ context.Context, _ *framework.CycleState, _ *v1alpha1.Dguest, _ framework.FoodToStatusMap) (*framework.PostFilterResult, *framework.Status) {
// 	return nil, framework.NewStatus(framework.Code(pl.inj.PostFilterStatus), "injected status")
// }

// func (pl *TestPlugin) PreScore(ctx context.Context, state *framework.CycleState, dguest *v1alpha1.Dguest, foods []*v1alpha1.Food) *framework.Status {
// 	return framework.NewStatus(framework.Code(pl.inj.PreScoreStatus), "injected status")
// }

// func (pl *TestPlugin) Reserve(ctx context.Context, state *framework.CycleState, p *v1alpha1.Dguest, selectedFood *v1alpha1.FoodInfoBase) *framework.Status {
// 	return framework.NewStatus(framework.Code(pl.inj.ReserveStatus), "injected status")
// }

// func (pl *TestPlugin) Unreserve(ctx context.Context, state *framework.CycleState, p *v1alpha1.Dguest, selectedFood *v1alpha1.FoodInfoBase) {
// }

// func (pl *TestPlugin) PreBind(ctx context.Context, state *framework.CycleState, p *v1alpha1.Dguest, selectedFood *v1alpha1.FoodInfoBase) *framework.Status {
// 	return framework.NewStatus(framework.Code(pl.inj.PreBindStatus), "injected status")
// }

// func (pl *TestPlugin) PostBind(ctx context.Context, state *framework.CycleState, p *v1alpha1.Dguest, selectedFood *v1alpha1.FoodInfoBase) {
// }

// func (pl *TestPlugin) Permit(ctx context.Context, state *framework.CycleState, p *v1alpha1.Dguest, selectedFood *v1alpha1.FoodInfoBase) (*framework.Status, time.Duration) {
// 	return framework.NewStatus(framework.Code(pl.inj.PermitStatus), "injected status"), time.Duration(0)
// }

// func (pl *TestPlugin) Bind(ctx context.Context, state *framework.CycleState, p *v1alpha1.Dguest, selectedFood *v1alpha1.FoodInfoBase) *framework.Status {
// 	return framework.NewStatus(framework.Code(pl.inj.BindStatus), "injected status")
// }

// // TestPreFilterPlugin only implements PreFilterPlugin interface.
// type TestPreFilterPlugin struct {
// 	PreFilterCalled int
// }

// func (pl *TestPreFilterPlugin) Name() string {
// 	return preFilterPluginName
// }

// func (pl *TestPreFilterPlugin) PreFilter(ctx context.Context, state *framework.CycleState, p *v1alpha1.Dguest) (*framework.PreFilterResult, *framework.Status) {
// 	pl.PreFilterCalled++
// 	return nil, nil
// }

// func (pl *TestPreFilterPlugin) PreFilterExtensions() framework.PreFilterExtensions {
// 	return nil
// }

// // TestPreFilterWithExtensionsPlugin implements Add/Remove interfaces.
// type TestPreFilterWithExtensionsPlugin struct {
// 	PreFilterCalled int
// 	AddCalled       int
// 	RemoveCalled    int
// }

// func (pl *TestPreFilterWithExtensionsPlugin) Name() string {
// 	return preFilterWithExtensionsPluginName
// }

// func (pl *TestPreFilterWithExtensionsPlugin) PreFilter(ctx context.Context, state *framework.CycleState, p *v1alpha1.Dguest) (*framework.PreFilterResult, *framework.Status) {
// 	pl.PreFilterCalled++
// 	return nil, nil
// }

// func (pl *TestPreFilterWithExtensionsPlugin) AddDguest(ctx context.Context, state *framework.CycleState, dguestToSchedule *v1alpha1.Dguest,
// 	dguestInfoToAdd *framework.DguestInfo, foodInfo *framework.FoodInfo) *framework.Status {
// 	pl.AddCalled++
// 	return nil
// }

// func (pl *TestPreFilterWithExtensionsPlugin) RemoveDguest(ctx context.Context, state *framework.CycleState, dguestToSchedule *v1alpha1.Dguest,
// 	dguestInfoToRemove *framework.DguestInfo, foodInfo *framework.FoodInfo) *framework.Status {
// 	pl.RemoveCalled++
// 	return nil
// }

// func (pl *TestPreFilterWithExtensionsPlugin) PreFilterExtensions() framework.PreFilterExtensions {
// 	return pl
// }

// type TestDuplicatePlugin struct {
// }

// func (dp *TestDuplicatePlugin) Name() string {
// 	return duplicatePluginName
// }

// func (dp *TestDuplicatePlugin) PreFilter(ctx context.Context, state *framework.CycleState, p *v1alpha1.Dguest) (*framework.PreFilterResult, *framework.Status) {
// 	return nil, nil
// }

// func (dp *TestDuplicatePlugin) PreFilterExtensions() framework.PreFilterExtensions {
// 	return nil
// }

// var _ framework.PreFilterPlugin = &TestDuplicatePlugin{}

// func newDuplicatePlugin(_ runtime.Object, _ framework.Handle) (framework.Plugin, error) {
// 	return &TestDuplicatePlugin{}, nil
// }

// // TestPermitPlugin only implements PermitPlugin interface.
// type TestPermitPlugin struct {
// 	PreFilterCalled int
// }

// func (pp *TestPermitPlugin) Name() string {
// 	return permitPlugin
// }
// func (pp *TestPermitPlugin) Permit(ctx context.Context, state *framework.CycleState, p *v1alpha1.Dguest, selectedFood *v1alpha1.FoodInfoBase) (*framework.Status, time.Duration) {
// 	return framework.NewStatus(framework.Wait), 10 * time.Second
// }

// var _ framework.QueueSortPlugin = &TestQueueSortPlugin{}

// func newQueueSortPlugin(_ runtime.Object, _ framework.Handle) (framework.Plugin, error) {
// 	return &TestQueueSortPlugin{}, nil
// }

// // TestQueueSortPlugin is a no-op implementation for QueueSort extension point.
// type TestQueueSortPlugin struct{}

// func (pl *TestQueueSortPlugin) Name() string {
// 	return queueSortPlugin
// }

// func (pl *TestQueueSortPlugin) Less(_, _ *framework.QueuedDguestInfo) bool {
// 	return false
// }

// var _ framework.BindPlugin = &TestBindPlugin{}

// func newBindPlugin(_ runtime.Object, _ framework.Handle) (framework.Plugin, error) {
// 	return &TestBindPlugin{}, nil
// }

// // TestBindPlugin is a no-op implementation for Bind extension point.
// type TestBindPlugin struct{}

// func (t TestBindPlugin) Name() string {
// 	return bindPlugin
// }

// func (t TestBindPlugin) Bind(ctx context.Context, state *framework.CycleState, p *v1alpha1.Dguest, selectedFood *v1alpha1.FoodInfoBase) *framework.Status {
// 	return nil
// }

// var registry = func() Registry {
// 	r := make(Registry)
// 	r.Register(scoreWithNormalizePlugin1, newScoreWithNormalizePlugin1)
// 	r.Register(scoreWithNormalizePlugin2, newScoreWithNormalizePlugin2)
// 	r.Register(scorePlugin1, newScorePlugin1)
// 	r.Register(pluginNotImplementingScore, newPluginNotImplementingScore)
// 	r.Register(duplicatePluginName, newDuplicatePlugin)
// 	r.Register(testPlugin, newTestPlugin)
// 	r.Register(queueSortPlugin, newQueueSortPlugin)
// 	r.Register(bindPlugin, newBindPlugin)
// 	return r
// }()

// var defaultWeights = map[string]int32{
// 	scoreWithNormalizePlugin1: 1,
// 	scoreWithNormalizePlugin2: 2,
// 	scorePlugin1:              1,
// }

// var state = &framework.CycleState{}

// // Dguest is only used for logging errors.
// var dguest = &v1alpha1.Dguest{}
// var food = &v1alpha1.Food{
// 	ObjectMeta: metav1.ObjectMeta{
// 		Name: foodName,
// 	},
// }
// var lowPriority, highPriority = int32(0), int32(1000)
// var lowPriorityDguest = &v1alpha1.Dguest{
// 	ObjectMeta: metav1.ObjectMeta{UID: "low"},
// 	Spec:       v1alpha1.DguestSpec{},
// }
// var highPriorityDguest = &v1alpha1.Dguest{
// 	ObjectMeta: metav1.ObjectMeta{UID: "high"},
// 	Spec:       v1alpha1.DguestSpec{},
// }
// var foods = []*v1alpha1.Food{
// 	{ObjectMeta: metav1.ObjectMeta{Name: "food1"}},
// 	{ObjectMeta: metav1.ObjectMeta{Name: "food2"}},
// }

// var (
// 	errInjectedStatus       = errors.New("injected status")
// 	errInjectedFilterStatus = errors.New("injected filter status")
// )

// func newFrameworkWithQueueSortAndBind(r Registry, profile v1.SchedulerProfile, stopCh <-chan struct{}, opts ...Option) (framework.Framework, error) {
// 	if _, ok := r[queueSortPlugin]; !ok {
// 		r[queueSortPlugin] = newQueueSortPlugin
// 	}
// 	if _, ok := r[bindPlugin]; !ok {
// 		r[bindPlugin] = newBindPlugin
// 	}

// 	if len(profile.Plugins.QueueSort.Enabled) == 0 {
// 		profile.Plugins.QueueSort.Enabled = append(profile.Plugins.QueueSort.Enabled, v1.Plugin{Name: queueSortPlugin})
// 	}
// 	if len(profile.Plugins.Bind.Enabled) == 0 {
// 		profile.Plugins.Bind.Enabled = append(profile.Plugins.Bind.Enabled, v1.Plugin{Name: bindPlugin})
// 	}
// 	return NewFramework(r, &profile, stopCh, opts...)
// }

// func TestInitFrameworkWithScorePlugins(t *testing.T) {
// 	tests := []struct {
// 		name    string
// 		plugins *v1.Plugins
// 		// If initErr is true, we expect framework initialization to fail.
// 		initErr bool
// 	}{
// 		{
// 			name:    "enabled Score plugin doesn't exist in registry",
// 			plugins: buildScoreConfigDefaultWeights("notExist"),
// 			initErr: true,
// 		},
// 		{
// 			name:    "enabled Score plugin doesn't extend the ScorePlugin interface",
// 			plugins: buildScoreConfigDefaultWeights(pluginNotImplementingScore),
// 			initErr: true,
// 		},
// 		{
// 			name:    "Score plugins are nil",
// 			plugins: &v1.Plugins{},
// 		},
// 		{
// 			name:    "enabled Score plugin list is empty",
// 			plugins: buildScoreConfigDefaultWeights(),
// 		},
// 		{
// 			name:    "enabled plugin only implements ScorePlugin interface",
// 			plugins: buildScoreConfigDefaultWeights(scorePlugin1),
// 		},
// 		{
// 			name:    "enabled plugin implements ScoreWithNormalizePlugin interface",
// 			plugins: buildScoreConfigDefaultWeights(scoreWithNormalizePlugin1),
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			profile := v1.SchedulerProfile{Plugins: tt.plugins}
// 			stopCh := make(chan struct{})
// 			defer close(stopCh)
// 			_, err := newFrameworkWithQueueSortAndBind(registry, profile, stopCh)
// 			if tt.initErr && err == nil {
// 				t.Fatal("Framework initialization should fail")
// 			}
// 			if !tt.initErr && err != nil {
// 				t.Fatalf("Failed to create framework for testing: %v", err)
// 			}
// 		})
// 	}
// }

// func TestNewFrameworkErrors(t *testing.T) {
// 	tests := []struct {
// 		name      string
// 		plugins   *v1.Plugins
// 		pluginCfg []v1.PluginConfig
// 		wantErr   string
// 	}{
// 		{
// 			name: "duplicate plugin name",
// 			plugins: &v1.Plugins{
// 				PreFilter: v1.PluginSet{
// 					Enabled: []v1.Plugin{
// 						{Name: duplicatePluginName, Weight: 1},
// 						{Name: duplicatePluginName, Weight: 1},
// 					},
// 				},
// 			},
// 			pluginCfg: []v1.PluginConfig{
// 				{Name: duplicatePluginName},
// 			},
// 			wantErr: "already registered",
// 		},
// 		{
// 			name: "duplicate plugin config",
// 			plugins: &v1.Plugins{
// 				PreFilter: v1.PluginSet{
// 					Enabled: []v1.Plugin{
// 						{Name: duplicatePluginName, Weight: 1},
// 					},
// 				},
// 			},
// 			pluginCfg: []v1.PluginConfig{
// 				{Name: duplicatePluginName},
// 				{Name: duplicatePluginName},
// 			},
// 			wantErr: "repeated config for plugin",
// 		},
// 	}

// 	for _, tc := range tests {
// 		t.Run(tc.name, func(t *testing.T) {
// 			profile := &v1.SchedulerProfile{
// 				Plugins:      tc.plugins,
// 				PluginConfig: tc.pluginCfg,
// 			}
// 			_, err := NewFramework(registry, profile, wait.NeverStop)
// 			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
// 				t.Errorf("Unexpected error, got %v, expect: %s", err, tc.wantErr)
// 			}
// 		})
// 	}
// }

// func TestNewFrameworkMultiPointExpansion(t *testing.T) {
// 	tests := []struct {
// 		name        string
// 		plugins     *v1.Plugins
// 		wantPlugins *v1.Plugins
// 		wantErr     string
// 	}{
// 		{
// 			name: "plugin expansion",
// 			plugins: &v1.Plugins{
// 				MultiPoint: v1.PluginSet{
// 					Enabled: []v1.Plugin{
// 						{Name: testPlugin, Weight: 5},
// 					},
// 				},
// 			},
// 			wantPlugins: &v1.Plugins{
// 				QueueSort:  v1.PluginSet{Enabled: []v1.Plugin{{Name: testPlugin}}},
// 				PreFilter:  v1.PluginSet{Enabled: []v1.Plugin{{Name: testPlugin}}},
// 				Filter:     v1.PluginSet{Enabled: []v1.Plugin{{Name: testPlugin}}},
// 				PostFilter: v1.PluginSet{Enabled: []v1.Plugin{{Name: testPlugin}}},
// 				PreScore:   v1.PluginSet{Enabled: []v1.Plugin{{Name: testPlugin}}},
// 				Score:      v1.PluginSet{Enabled: []v1.Plugin{{Name: testPlugin, Weight: 5}}},
// 				Reserve:    v1.PluginSet{Enabled: []v1.Plugin{{Name: testPlugin}}},
// 				Permit:     v1.PluginSet{Enabled: []v1.Plugin{{Name: testPlugin}}},
// 				PreBind:    v1.PluginSet{Enabled: []v1.Plugin{{Name: testPlugin}}},
// 				Bind:       v1.PluginSet{Enabled: []v1.Plugin{{Name: testPlugin}}},
// 				PostBind:   v1.PluginSet{Enabled: []v1.Plugin{{Name: testPlugin}}},
// 			},
// 		},
// 		{
// 			name: "disable MultiPoint plugin at some extension points",
// 			plugins: &v1.Plugins{
// 				MultiPoint: v1.PluginSet{
// 					Enabled: []v1.Plugin{
// 						{Name: testPlugin},
// 					},
// 				},
// 				PreScore: v1.PluginSet{
// 					Disabled: []v1.Plugin{
// 						{Name: testPlugin},
// 					},
// 				},
// 				Score: v1.PluginSet{
// 					Disabled: []v1.Plugin{
// 						{Name: testPlugin},
// 					},
// 				},
// 			},
// 			wantPlugins: &v1.Plugins{
// 				QueueSort:  v1.PluginSet{Enabled: []v1.Plugin{{Name: testPlugin}}},
// 				PreFilter:  v1.PluginSet{Enabled: []v1.Plugin{{Name: testPlugin}}},
// 				Filter:     v1.PluginSet{Enabled: []v1.Plugin{{Name: testPlugin}}},
// 				PostFilter: v1.PluginSet{Enabled: []v1.Plugin{{Name: testPlugin}}},
// 				Reserve:    v1.PluginSet{Enabled: []v1.Plugin{{Name: testPlugin}}},
// 				Permit:     v1.PluginSet{Enabled: []v1.Plugin{{Name: testPlugin}}},
// 				PreBind:    v1.PluginSet{Enabled: []v1.Plugin{{Name: testPlugin}}},
// 				Bind:       v1.PluginSet{Enabled: []v1.Plugin{{Name: testPlugin}}},
// 				PostBind:   v1.PluginSet{Enabled: []v1.Plugin{{Name: testPlugin}}},
// 			},
// 		},
// 		{
// 			name: "Multiple MultiPoint plugins",
// 			plugins: &v1.Plugins{
// 				MultiPoint: v1.PluginSet{
// 					Enabled: []v1.Plugin{
// 						{Name: testPlugin},
// 						{Name: scorePlugin1},
// 					},
// 				},
// 			},
// 			wantPlugins: &v1.Plugins{
// 				QueueSort:  v1.PluginSet{Enabled: []v1.Plugin{{Name: testPlugin}}},
// 				PreFilter:  v1.PluginSet{Enabled: []v1.Plugin{{Name: testPlugin}}},
// 				Filter:     v1.PluginSet{Enabled: []v1.Plugin{{Name: testPlugin}}},
// 				PostFilter: v1.PluginSet{Enabled: []v1.Plugin{{Name: testPlugin}}},
// 				PreScore: v1.PluginSet{Enabled: []v1.Plugin{
// 					{Name: testPlugin},
// 					{Name: scorePlugin1},
// 				}},
// 				Score: v1.PluginSet{Enabled: []v1.Plugin{
// 					{Name: testPlugin, Weight: 1},
// 					{Name: scorePlugin1, Weight: 1},
// 				}},
// 				Reserve:  v1.PluginSet{Enabled: []v1.Plugin{{Name: testPlugin}}},
// 				Permit:   v1.PluginSet{Enabled: []v1.Plugin{{Name: testPlugin}}},
// 				PreBind:  v1.PluginSet{Enabled: []v1.Plugin{{Name: testPlugin}}},
// 				Bind:     v1.PluginSet{Enabled: []v1.Plugin{{Name: testPlugin}}},
// 				PostBind: v1.PluginSet{Enabled: []v1.Plugin{{Name: testPlugin}}},
// 			},
// 		},
// 		{
// 			name: "disable MultiPoint extension",
// 			plugins: &v1.Plugins{
// 				MultiPoint: v1.PluginSet{
// 					Enabled: []v1.Plugin{
// 						{Name: testPlugin},
// 						{Name: scorePlugin1},
// 					},
// 				},
// 				PreScore: v1.PluginSet{
// 					Disabled: []v1.Plugin{
// 						{Name: "*"},
// 					},
// 				},
// 			},
// 			wantPlugins: &v1.Plugins{
// 				QueueSort:  v1.PluginSet{Enabled: []v1.Plugin{{Name: testPlugin}}},
// 				PreFilter:  v1.PluginSet{Enabled: []v1.Plugin{{Name: testPlugin}}},
// 				Filter:     v1.PluginSet{Enabled: []v1.Plugin{{Name: testPlugin}}},
// 				PostFilter: v1.PluginSet{Enabled: []v1.Plugin{{Name: testPlugin}}},
// 				Score: v1.PluginSet{Enabled: []v1.Plugin{
// 					{Name: testPlugin, Weight: 1},
// 					{Name: scorePlugin1, Weight: 1},
// 				}},
// 				Reserve:  v1.PluginSet{Enabled: []v1.Plugin{{Name: testPlugin}}},
// 				Permit:   v1.PluginSet{Enabled: []v1.Plugin{{Name: testPlugin}}},
// 				PreBind:  v1.PluginSet{Enabled: []v1.Plugin{{Name: testPlugin}}},
// 				Bind:     v1.PluginSet{Enabled: []v1.Plugin{{Name: testPlugin}}},
// 				PostBind: v1.PluginSet{Enabled: []v1.Plugin{{Name: testPlugin}}},
// 			},
// 		},
// 		{
// 			name: "Reorder MultiPoint plugins (specified extension takes precedence)",
// 			plugins: &v1.Plugins{
// 				MultiPoint: v1.PluginSet{
// 					Enabled: []v1.Plugin{
// 						{Name: scoreWithNormalizePlugin1},
// 						{Name: testPlugin},
// 						{Name: scorePlugin1},
// 					},
// 				},
// 				Score: v1.PluginSet{
// 					Enabled: []v1.Plugin{
// 						{Name: scorePlugin1},
// 						{Name: testPlugin},
// 					},
// 				},
// 			},
// 			wantPlugins: &v1.Plugins{
// 				QueueSort:  v1.PluginSet{Enabled: []v1.Plugin{{Name: testPlugin}}},
// 				PreFilter:  v1.PluginSet{Enabled: []v1.Plugin{{Name: testPlugin}}},
// 				Filter:     v1.PluginSet{Enabled: []v1.Plugin{{Name: testPlugin}}},
// 				PostFilter: v1.PluginSet{Enabled: []v1.Plugin{{Name: testPlugin}}},
// 				PreScore: v1.PluginSet{Enabled: []v1.Plugin{
// 					{Name: testPlugin},
// 					{Name: scorePlugin1},
// 				}},
// 				Score: v1.PluginSet{Enabled: []v1.Plugin{
// 					{Name: scorePlugin1, Weight: 1},
// 					{Name: testPlugin, Weight: 1},
// 					{Name: scoreWithNormalizePlugin1, Weight: 1},
// 				}},
// 				Reserve:  v1.PluginSet{Enabled: []v1.Plugin{{Name: testPlugin}}},
// 				Permit:   v1.PluginSet{Enabled: []v1.Plugin{{Name: testPlugin}}},
// 				PreBind:  v1.PluginSet{Enabled: []v1.Plugin{{Name: testPlugin}}},
// 				Bind:     v1.PluginSet{Enabled: []v1.Plugin{{Name: testPlugin}}},
// 				PostBind: v1.PluginSet{Enabled: []v1.Plugin{{Name: testPlugin}}},
// 			},
// 		},
// 		{
// 			name: "Reorder MultiPoint plugins (specified extension only takes precedence when it exists in MultiPoint)",
// 			plugins: &v1.Plugins{
// 				MultiPoint: v1.PluginSet{
// 					Enabled: []v1.Plugin{
// 						{Name: testPlugin},
// 						{Name: scorePlugin1},
// 					},
// 				},
// 				Score: v1.PluginSet{
// 					Enabled: []v1.Plugin{
// 						{Name: scoreWithNormalizePlugin1},
// 						{Name: scorePlugin1},
// 						{Name: testPlugin},
// 					},
// 				},
// 			},
// 			wantPlugins: &v1.Plugins{
// 				QueueSort:  v1.PluginSet{Enabled: []v1.Plugin{{Name: testPlugin}}},
// 				PreFilter:  v1.PluginSet{Enabled: []v1.Plugin{{Name: testPlugin}}},
// 				Filter:     v1.PluginSet{Enabled: []v1.Plugin{{Name: testPlugin}}},
// 				PostFilter: v1.PluginSet{Enabled: []v1.Plugin{{Name: testPlugin}}},
// 				PreScore: v1.PluginSet{Enabled: []v1.Plugin{
// 					{Name: testPlugin},
// 					{Name: scorePlugin1},
// 				}},
// 				Score: v1.PluginSet{Enabled: []v1.Plugin{
// 					{Name: scorePlugin1, Weight: 1},
// 					{Name: testPlugin, Weight: 1},
// 					{Name: scoreWithNormalizePlugin1, Weight: 1},
// 				}},
// 				Reserve:  v1.PluginSet{Enabled: []v1.Plugin{{Name: testPlugin}}},
// 				Permit:   v1.PluginSet{Enabled: []v1.Plugin{{Name: testPlugin}}},
// 				PreBind:  v1.PluginSet{Enabled: []v1.Plugin{{Name: testPlugin}}},
// 				Bind:     v1.PluginSet{Enabled: []v1.Plugin{{Name: testPlugin}}},
// 				PostBind: v1.PluginSet{Enabled: []v1.Plugin{{Name: testPlugin}}},
// 			},
// 		},
// 		{
// 			name: "Override MultiPoint plugins weights",
// 			plugins: &v1.Plugins{
// 				MultiPoint: v1.PluginSet{
// 					Enabled: []v1.Plugin{
// 						{Name: testPlugin},
// 						{Name: scorePlugin1},
// 					},
// 				},
// 				Score: v1.PluginSet{
// 					Enabled: []v1.Plugin{
// 						{Name: scorePlugin1, Weight: 5},
// 						{Name: testPlugin, Weight: 3},
// 					},
// 				},
// 			},
// 			wantPlugins: &v1.Plugins{
// 				QueueSort:  v1.PluginSet{Enabled: []v1.Plugin{{Name: testPlugin}}},
// 				PreFilter:  v1.PluginSet{Enabled: []v1.Plugin{{Name: testPlugin}}},
// 				Filter:     v1.PluginSet{Enabled: []v1.Plugin{{Name: testPlugin}}},
// 				PostFilter: v1.PluginSet{Enabled: []v1.Plugin{{Name: testPlugin}}},
// 				PreScore: v1.PluginSet{Enabled: []v1.Plugin{
// 					{Name: testPlugin},
// 					{Name: scorePlugin1},
// 				}},
// 				Score: v1.PluginSet{Enabled: []v1.Plugin{
// 					{Name: scorePlugin1, Weight: 5},
// 					{Name: testPlugin, Weight: 3},
// 				}},
// 				Reserve:  v1.PluginSet{Enabled: []v1.Plugin{{Name: testPlugin}}},
// 				Permit:   v1.PluginSet{Enabled: []v1.Plugin{{Name: testPlugin}}},
// 				PreBind:  v1.PluginSet{Enabled: []v1.Plugin{{Name: testPlugin}}},
// 				Bind:     v1.PluginSet{Enabled: []v1.Plugin{{Name: testPlugin}}},
// 				PostBind: v1.PluginSet{Enabled: []v1.Plugin{{Name: testPlugin}}},
// 			},
// 		},
// 		{
// 			name: "disable and enable MultiPoint plugins with '*'",
// 			plugins: &v1.Plugins{
// 				MultiPoint: v1.PluginSet{
// 					Enabled: []v1.Plugin{
// 						{Name: queueSortPlugin},
// 						{Name: bindPlugin},
// 						{Name: scorePlugin1},
// 					},
// 					Disabled: []v1.Plugin{
// 						{Name: "*"},
// 					},
// 				},
// 			},
// 			wantPlugins: &v1.Plugins{
// 				QueueSort: v1.PluginSet{Enabled: []v1.Plugin{{Name: queueSortPlugin}}},
// 				PreScore: v1.PluginSet{Enabled: []v1.Plugin{
// 					{Name: scorePlugin1},
// 				}},
// 				Score: v1.PluginSet{Enabled: []v1.Plugin{
// 					{Name: scorePlugin1, Weight: 1},
// 				}},
// 				Bind: v1.PluginSet{Enabled: []v1.Plugin{{Name: bindPlugin}}},
// 			},
// 		},
// 		{
// 			name: "disable and enable MultiPoint plugin by name",
// 			plugins: &v1.Plugins{
// 				MultiPoint: v1.PluginSet{
// 					Enabled: []v1.Plugin{
// 						{Name: bindPlugin},
// 						{Name: queueSortPlugin},
// 						{Name: scorePlugin1},
// 					},
// 					Disabled: []v1.Plugin{
// 						{Name: scorePlugin1},
// 					},
// 				},
// 			},
// 			wantPlugins: &v1.Plugins{
// 				QueueSort: v1.PluginSet{Enabled: []v1.Plugin{{Name: queueSortPlugin}}},
// 				PreScore: v1.PluginSet{Enabled: []v1.Plugin{
// 					{Name: scorePlugin1},
// 				}},
// 				Score: v1.PluginSet{Enabled: []v1.Plugin{
// 					{Name: scorePlugin1, Weight: 1},
// 				}},
// 				Bind: v1.PluginSet{Enabled: []v1.Plugin{{Name: bindPlugin}}},
// 			},
// 		},
// 		{
// 			name: "Expect 'already registered' error",
// 			plugins: &v1.Plugins{
// 				MultiPoint: v1.PluginSet{
// 					Enabled: []v1.Plugin{
// 						{Name: testPlugin},
// 						{Name: testPlugin},
// 					},
// 				},
// 			},
// 			wantErr: "already registered",
// 		},
// 	}

// 	for _, tc := range tests {
// 		t.Run(tc.name, func(t *testing.T) {
// 			stopCh := make(chan struct{})
// 			defer close(stopCh)
// 			fw, err := NewFramework(registry, &v1.SchedulerProfile{Plugins: tc.plugins}, stopCh)
// 			if err != nil {
// 				if tc.wantErr == "" || !strings.Contains(err.Error(), tc.wantErr) {
// 					t.Fatalf("Unexpected error, got %v, expect: %s", err, tc.wantErr)
// 				}
// 			} else {
// 				if tc.wantErr != "" {
// 					t.Fatalf("Unexpected error, got %v, expect: %s", err, tc.wantErr)
// 				}
// 			}

// 			if tc.wantErr == "" {
// 				if diff := cmp.Diff(tc.wantPlugins, fw.ListPlugins()); diff != "" {
// 					t.Fatalf("Unexpected eventToPlugin map (-want,+got):%s", diff)
// 				}
// 			}
// 		})
// 	}
// }

// // fakeNoopPlugin doesn't implement interface framework.EnqueueExtensions.
// type fakeNoopPlugin struct{}

// func (*fakeNoopPlugin) Name() string { return "fakeNoop" }

// func (*fakeNoopPlugin) Filter(_ context.Context, _ *framework.CycleState, _ *v1alpha1.Dguest, _ *framework.FoodInfo) *framework.Status {
// 	return nil
// }

// type fakeFoodPlugin struct{}

// func (*fakeFoodPlugin) Name() string { return "fakeFood" }

// func (*fakeFoodPlugin) Filter(_ context.Context, _ *framework.CycleState, _ *v1alpha1.Dguest, _ *framework.FoodInfo) *framework.Status {
// 	return nil
// }

// func (*fakeFoodPlugin) EventsToRegister() []framework.ClusterEvent {
// 	return []framework.ClusterEvent{
// 		{Resource: framework.Dguest, ActionType: framework.All},
// 		{Resource: framework.Food, ActionType: framework.Delete},
// 		//{Resource: framework.CSIFood, ActionType: framework.Update | framework.Delete},
// 	}
// }

// type fakeDguestPlugin struct{}

// func (*fakeDguestPlugin) Name() string { return "fakeDguest" }

// func (*fakeDguestPlugin) Filter(_ context.Context, _ *framework.CycleState, _ *v1alpha1.Dguest, _ *framework.FoodInfo) *framework.Status {
// 	return nil
// }

// func (*fakeDguestPlugin) EventsToRegister() []framework.ClusterEvent {
// 	return []framework.ClusterEvent{
// 		{Resource: framework.Dguest, ActionType: framework.All},
// 		{Resource: framework.Food, ActionType: framework.Add | framework.Delete},
// 		{Resource: framework.PersistentVolumeClaim, ActionType: framework.Delete},
// 	}
// }

// // fakeNoopRuntimePlugin implement interface framework.EnqueueExtensions, but returns nil
// // at runtime. This can simulate a plugin registered at scheduler setup, but does nothing
// // due to some disabled feature gate.
// type fakeNoopRuntimePlugin struct{}

// func (*fakeNoopRuntimePlugin) Name() string { return "fakeNoopRuntime" }

// func (*fakeNoopRuntimePlugin) Filter(_ context.Context, _ *framework.CycleState, _ *v1alpha1.Dguest, _ *framework.FoodInfo) *framework.Status {
// 	return nil
// }

// func (*fakeNoopRuntimePlugin) EventsToRegister() []framework.ClusterEvent { return nil }

// func TestNewFrameworkFillEventToPluginMap(t *testing.T) {
// 	tests := []struct {
// 		name    string
// 		plugins []framework.Plugin
// 		want    map[framework.ClusterEvent]sets.String
// 	}{
// 		{
// 			name:    "no-op plugin",
// 			plugins: []framework.Plugin{&fakeNoopPlugin{}},
// 			want: map[framework.ClusterEvent]sets.String{
// 				{Resource: framework.Dguest, ActionType: framework.All}: sets.NewString("fakeNoop", bindPlugin, queueSortPlugin),
// 				{Resource: framework.Food, ActionType: framework.All}:   sets.NewString("fakeNoop", bindPlugin, queueSortPlugin),
// 				//{Resource: framework.CSIFood, ActionType: framework.All}:               sets.NewString("fakeNoop", bindPlugin, queueSortPlugin),
// 				{Resource: framework.PersistentVolume, ActionType: framework.All}:      sets.NewString("fakeNoop", bindPlugin, queueSortPlugin),
// 				{Resource: framework.PersistentVolumeClaim, ActionType: framework.All}: sets.NewString("fakeNoop", bindPlugin, queueSortPlugin),
// 				{Resource: framework.StorageClass, ActionType: framework.All}:          sets.NewString("fakeNoop", bindPlugin, queueSortPlugin),
// 			},
// 		},
// 		{
// 			name:    "food plugin",
// 			plugins: []framework.Plugin{&fakeFoodPlugin{}},
// 			want: map[framework.ClusterEvent]sets.String{
// 				{Resource: framework.Dguest, ActionType: framework.All}:  sets.NewString("fakeFood", bindPlugin, queueSortPlugin),
// 				{Resource: framework.Food, ActionType: framework.Delete}: sets.NewString("fakeFood"),
// 				{Resource: framework.Food, ActionType: framework.All}:    sets.NewString(bindPlugin, queueSortPlugin),
// 				//{Resource: framework.CSIFood, ActionType: framework.Update | framework.Delete}: sets.NewString("fakeFood"),
// 				//{Resource: framework.CSIFood, ActionType: framework.All}:                       sets.NewString(bindPlugin, queueSortPlugin),
// 				{Resource: framework.PersistentVolume, ActionType: framework.All}:      sets.NewString(bindPlugin, queueSortPlugin),
// 				{Resource: framework.PersistentVolumeClaim, ActionType: framework.All}: sets.NewString(bindPlugin, queueSortPlugin),
// 				{Resource: framework.StorageClass, ActionType: framework.All}:          sets.NewString(bindPlugin, queueSortPlugin),
// 			},
// 		},
// 		{
// 			name:    "dguest plugin",
// 			plugins: []framework.Plugin{&fakeDguestPlugin{}},
// 			want: map[framework.ClusterEvent]sets.String{
// 				{Resource: framework.Dguest, ActionType: framework.All}:                   sets.NewString("fakeDguest", bindPlugin, queueSortPlugin),
// 				{Resource: framework.Food, ActionType: framework.Add | framework.Delete}:  sets.NewString("fakeDguest"),
// 				{Resource: framework.Food, ActionType: framework.All}:                     sets.NewString(bindPlugin, queueSortPlugin),
// 				{Resource: framework.PersistentVolumeClaim, ActionType: framework.Delete}: sets.NewString("fakeDguest"),
// 				{Resource: framework.PersistentVolumeClaim, ActionType: framework.All}:    sets.NewString(bindPlugin, queueSortPlugin),
// 				//{Resource: framework.CSIFood, ActionType: framework.All}:                  sets.NewString(bindPlugin, queueSortPlugin),
// 				{Resource: framework.PersistentVolume, ActionType: framework.All}: sets.NewString(bindPlugin, queueSortPlugin),
// 				{Resource: framework.StorageClass, ActionType: framework.All}:     sets.NewString(bindPlugin, queueSortPlugin),
// 			},
// 		},
// 		{
// 			name:    "food and dguest plugin",
// 			plugins: []framework.Plugin{&fakeFoodPlugin{}, &fakeDguestPlugin{}},
// 			want: map[framework.ClusterEvent]sets.String{
// 				{Resource: framework.Food, ActionType: framework.Delete}:                 sets.NewString("fakeFood"),
// 				{Resource: framework.Food, ActionType: framework.Add | framework.Delete}: sets.NewString("fakeDguest"),
// 				{Resource: framework.Dguest, ActionType: framework.All}:                  sets.NewString("fakeFood", "fakeDguest", bindPlugin, queueSortPlugin),
// 				//{Resource: framework.CSIFood, ActionType: framework.Update | framework.Delete}: sets.NewString("fakeFood"),
// 				{Resource: framework.PersistentVolumeClaim, ActionType: framework.Delete}: sets.NewString("fakeDguest"),
// 				{Resource: framework.Food, ActionType: framework.All}:                     sets.NewString(bindPlugin, queueSortPlugin),
// 				//{Resource: framework.CSIFood, ActionType: framework.All}:                       sets.NewString(bindPlugin, queueSortPlugin),
// 				{Resource: framework.PersistentVolume, ActionType: framework.All}:      sets.NewString(bindPlugin, queueSortPlugin),
// 				{Resource: framework.PersistentVolumeClaim, ActionType: framework.All}: sets.NewString(bindPlugin, queueSortPlugin),
// 				{Resource: framework.StorageClass, ActionType: framework.All}:          sets.NewString(bindPlugin, queueSortPlugin),
// 			},
// 		},
// 		{
// 			name:    "no-op runtime plugin",
// 			plugins: []framework.Plugin{&fakeNoopRuntimePlugin{}},
// 			want: map[framework.ClusterEvent]sets.String{
// 				{Resource: framework.Dguest, ActionType: framework.All}: sets.NewString(bindPlugin, queueSortPlugin),
// 				{Resource: framework.Food, ActionType: framework.All}:   sets.NewString(bindPlugin, queueSortPlugin),
// 				//{Resource: framework.CSIFood, ActionType: framework.All}:               sets.NewString(bindPlugin, queueSortPlugin),
// 				{Resource: framework.PersistentVolume, ActionType: framework.All}:      sets.NewString(bindPlugin, queueSortPlugin),
// 				{Resource: framework.PersistentVolumeClaim, ActionType: framework.All}: sets.NewString(bindPlugin, queueSortPlugin),
// 				{Resource: framework.StorageClass, ActionType: framework.All}:          sets.NewString(bindPlugin, queueSortPlugin),
// 			},
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			registry := Registry{}
// 			cfgPls := &v1.Plugins{}
// 			for _, pl := range tt.plugins {
// 				tmpPl := pl
// 				if err := registry.Register(pl.Name(), func(_ runtime.Object, _ framework.Handle) (framework.Plugin, error) {
// 					return tmpPl, nil
// 				}); err != nil {
// 					t.Fatalf("fail to register filter plugin (%s)", pl.Name())
// 				}
// 				cfgPls.Filter.Enabled = append(cfgPls.Filter.Enabled, v1.Plugin{Name: pl.Name()})
// 			}

// 			got := make(map[framework.ClusterEvent]sets.String)
// 			profile := v1.SchedulerProfile{Plugins: cfgPls}
// 			stopCh := make(chan struct{})
// 			defer close(stopCh)
// 			_, err := newFrameworkWithQueueSortAndBind(registry, profile, stopCh, WithClusterEventMap(got))
// 			if err != nil {
// 				t.Fatal(err)
// 			}
// 			if diff := cmp.Diff(tt.want, got); diff != "" {
// 				t.Errorf("Unexpected eventToPlugin map (-want,+got):%s", diff)
// 			}
// 		})
// 	}
// }

// func TestRunScorePlugins(t *testing.T) {
// 	tests := []struct {
// 		name          string
// 		registry      Registry
// 		plugins       *v1.Plugins
// 		pluginConfigs []v1.PluginConfig
// 		want          framework.PluginToFoodScores
// 		// If err is true, we expect RunScorePlugin to fail.
// 		err bool
// 	}{
// 		{
// 			name:    "no Score plugins",
// 			plugins: buildScoreConfigDefaultWeights(),
// 			want:    framework.PluginToFoodScores{},
// 		},
// 		{
// 			name:    "single Score plugin",
// 			plugins: buildScoreConfigDefaultWeights(scorePlugin1),
// 			pluginConfigs: []v1.PluginConfig{
// 				{
// 					Name: scorePlugin1,
// 					Args: &runtime.Unknown{
// 						Raw: []byte(`{ "scoreRes": 1 }`),
// 					},
// 				},
// 			},
// 			// scorePlugin1 Score returns 1, weight=1, so want=1.
// 			want: framework.PluginToFoodScores{
// 				scorePlugin1: {{Name: "food1", Score: 1}, {Name: "food2", Score: 1}},
// 			},
// 		},
// 		{
// 			name: "single ScoreWithNormalize plugin",
// 			// registry: registry,
// 			plugins: buildScoreConfigDefaultWeights(scoreWithNormalizePlugin1),
// 			pluginConfigs: []v1.PluginConfig{
// 				{
// 					Name: scoreWithNormalizePlugin1,
// 					Args: &runtime.Unknown{
// 						Raw: []byte(`{ "scoreRes": 10, "normalizeRes": 5 }`),
// 					},
// 				},
// 			},
// 			// scoreWithNormalizePlugin1 Score returns 10, but NormalizeScore overrides to 5, weight=1, so want=5
// 			want: framework.PluginToFoodScores{
// 				scoreWithNormalizePlugin1: {{Name: "food1", Score: 5}, {Name: "food2", Score: 5}},
// 			},
// 		},
// 		{
// 			name:    "2 Score plugins, 2 NormalizeScore plugins",
// 			plugins: buildScoreConfigDefaultWeights(scorePlugin1, scoreWithNormalizePlugin1, scoreWithNormalizePlugin2),
// 			pluginConfigs: []v1.PluginConfig{
// 				{
// 					Name: scorePlugin1,
// 					Args: &runtime.Unknown{
// 						Raw: []byte(`{ "scoreRes": 1 }`),
// 					},
// 				},
// 				{
// 					Name: scoreWithNormalizePlugin1,
// 					Args: &runtime.Unknown{
// 						Raw: []byte(`{ "scoreRes": 3, "normalizeRes": 4}`),
// 					},
// 				},
// 				{
// 					Name: scoreWithNormalizePlugin2,
// 					Args: &runtime.Unknown{
// 						Raw: []byte(`{ "scoreRes": 4, "normalizeRes": 5}`),
// 					},
// 				},
// 			},
// 			// scorePlugin1 Score returns 1, weight =1, so want=1.
// 			// scoreWithNormalizePlugin1 Score returns 3, but NormalizeScore overrides to 4, weight=1, so want=4.
// 			// scoreWithNormalizePlugin2 Score returns 4, but NormalizeScore overrides to 5, weight=2, so want=10.
// 			want: framework.PluginToFoodScores{
// 				scorePlugin1:              {{Name: "food1", Score: 1}, {Name: "food2", Score: 1}},
// 				scoreWithNormalizePlugin1: {{Name: "food1", Score: 4}, {Name: "food2", Score: 4}},
// 				scoreWithNormalizePlugin2: {{Name: "food1", Score: 10}, {Name: "food2", Score: 10}},
// 			},
// 		},
// 		{
// 			name: "score fails",
// 			pluginConfigs: []v1.PluginConfig{
// 				{
// 					Name: scoreWithNormalizePlugin1,
// 					Args: &runtime.Unknown{
// 						Raw: []byte(`{ "scoreStatus": 1 }`),
// 					},
// 				},
// 			},
// 			plugins: buildScoreConfigDefaultWeights(scorePlugin1, scoreWithNormalizePlugin1),
// 			err:     true,
// 		},
// 		{
// 			name: "normalize fails",
// 			pluginConfigs: []v1.PluginConfig{
// 				{
// 					Name: scoreWithNormalizePlugin1,
// 					Args: &runtime.Unknown{
// 						Raw: []byte(`{ "normalizeStatus": 1 }`),
// 					},
// 				},
// 			},
// 			plugins: buildScoreConfigDefaultWeights(scorePlugin1, scoreWithNormalizePlugin1),
// 			err:     true,
// 		},
// 		{
// 			name:    "Score plugin return score greater than MaxFoodScore",
// 			plugins: buildScoreConfigDefaultWeights(scorePlugin1),
// 			pluginConfigs: []v1.PluginConfig{
// 				{
// 					Name: scorePlugin1,
// 					Args: &runtime.Unknown{
// 						Raw: []byte(fmt.Sprintf(`{ "scoreRes": %d }`, framework.MaxFoodScore+1)),
// 					},
// 				},
// 			},
// 			err: true,
// 		},
// 		{
// 			name:    "Score plugin return score less than MinFoodScore",
// 			plugins: buildScoreConfigDefaultWeights(scorePlugin1),
// 			pluginConfigs: []v1.PluginConfig{
// 				{
// 					Name: scorePlugin1,
// 					Args: &runtime.Unknown{
// 						Raw: []byte(fmt.Sprintf(`{ "scoreRes": %d }`, framework.MinFoodScore-1)),
// 					},
// 				},
// 			},
// 			err: true,
// 		},
// 		{
// 			name:    "ScoreWithNormalize plugin return score greater than MaxFoodScore",
// 			plugins: buildScoreConfigDefaultWeights(scoreWithNormalizePlugin1),
// 			pluginConfigs: []v1.PluginConfig{
// 				{
// 					Name: scoreWithNormalizePlugin1,
// 					Args: &runtime.Unknown{
// 						Raw: []byte(fmt.Sprintf(`{ "normalizeRes": %d }`, framework.MaxFoodScore+1)),
// 					},
// 				},
// 			},
// 			err: true,
// 		},
// 		{
// 			name:    "ScoreWithNormalize plugin return score less than MinFoodScore",
// 			plugins: buildScoreConfigDefaultWeights(scoreWithNormalizePlugin1),
// 			pluginConfigs: []v1.PluginConfig{
// 				{
// 					Name: scoreWithNormalizePlugin1,
// 					Args: &runtime.Unknown{
// 						Raw: []byte(fmt.Sprintf(`{ "normalizeRes": %d }`, framework.MinFoodScore-1)),
// 					},
// 				},
// 			},
// 			err: true,
// 		},
// 		{
// 			name: "single Score plugin with MultiPointExpansion",
// 			plugins: &v1.Plugins{
// 				MultiPoint: v1.PluginSet{
// 					Enabled: []v1.Plugin{
// 						{Name: scorePlugin1},
// 					},
// 				},
// 				Score: v1.PluginSet{
// 					Enabled: []v1.Plugin{
// 						{Name: scorePlugin1, Weight: 3},
// 					},
// 				},
// 			},
// 			pluginConfigs: []v1.PluginConfig{
// 				{
// 					Name: scorePlugin1,
// 					Args: &runtime.Unknown{
// 						Raw: []byte(`{ "scoreRes": 1 }`),
// 					},
// 				},
// 			},
// 			// scorePlugin1 Score returns 1, weight=3, so want=3.
// 			want: framework.PluginToFoodScores{
// 				scorePlugin1: {{Name: "food1", Score: 3}, {Name: "food2", Score: 3}},
// 			},
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			// Inject the results via Args in PluginConfig.
// 			profile := v1.SchedulerProfile{
// 				Plugins:      tt.plugins,
// 				PluginConfig: tt.pluginConfigs,
// 			}
// 			ctx, cancel := context.WithCancel(context.Background())
// 			defer cancel()
// 			f, err := newFrameworkWithQueueSortAndBind(registry, profile, ctx.Done())
// 			if err != nil {
// 				t.Fatalf("Failed to create framework for testing: %v", err)
// 			}

// 			res, status := f.RunScorePlugins(ctx, state, dguest, foods)

// 			if tt.err {
// 				if status.IsSuccess() {
// 					t.Errorf("Expected status to be non-success. got: %v", status.Code().String())
// 				}
// 				return
// 			}

// 			if !status.IsSuccess() {
// 				t.Errorf("Expected status to be success.")
// 			}
// 			if !reflect.DeepEqual(res, tt.want) {
// 				t.Errorf("Score map after RunScorePlugin. got: %+v, want: %+v.", res, tt.want)
// 			}
// 		})
// 	}
// }

// func TestPreFilterPlugins(t *testing.T) {
// 	preFilter1 := &TestPreFilterPlugin{}
// 	preFilter2 := &TestPreFilterWithExtensionsPlugin{}
// 	r := make(Registry)
// 	r.Register(preFilterPluginName,
// 		func(_ runtime.Object, fh framework.Handle) (framework.Plugin, error) {
// 			return preFilter1, nil
// 		})
// 	r.Register(preFilterWithExtensionsPluginName,
// 		func(_ runtime.Object, fh framework.Handle) (framework.Plugin, error) {
// 			return preFilter2, nil
// 		})
// 	plugins := &v1.Plugins{PreFilter: v1.PluginSet{Enabled: []v1.Plugin{{Name: preFilterWithExtensionsPluginName}, {Name: preFilterPluginName}}}}
// 	t.Run("TestPreFilterPlugin", func(t *testing.T) {
// 		profile := v1.SchedulerProfile{Plugins: plugins}
// 		ctx, cancel := context.WithCancel(context.Background())
// 		defer cancel()

// 		f, err := newFrameworkWithQueueSortAndBind(r, profile, ctx.Done())
// 		if err != nil {
// 			t.Fatalf("Failed to create framework for testing: %v", err)
// 		}
// 		f.RunPreFilterPlugins(ctx, nil, nil)
// 		f.RunPreFilterExtensionAddDguest(ctx, nil, nil, nil, nil)
// 		f.RunPreFilterExtensionRemoveDguest(ctx, nil, nil, nil, nil)

// 		if preFilter1.PreFilterCalled != 1 {
// 			t.Errorf("preFilter1 called %v, expected: 1", preFilter1.PreFilterCalled)
// 		}
// 		if preFilter2.PreFilterCalled != 1 {
// 			t.Errorf("preFilter2 called %v, expected: 1", preFilter2.PreFilterCalled)
// 		}
// 		if preFilter2.AddCalled != 1 {
// 			t.Errorf("AddDguest called %v, expected: 1", preFilter2.AddCalled)
// 		}
// 		if preFilter2.RemoveCalled != 1 {
// 			t.Errorf("AddDguest called %v, expected: 1", preFilter2.RemoveCalled)
// 		}
// 	})
// }

// func TestFilterPlugins(t *testing.T) {
// 	tests := []struct {
// 		name          string
// 		plugins       []*TestPlugin
// 		wantStatus    *framework.Status
// 		wantStatusMap framework.PluginToStatus
// 	}{
// 		{
// 			name: "SuccessFilter",
// 			plugins: []*TestPlugin{
// 				{
// 					name: "TestPlugin",
// 					inj:  injectedResult{FilterStatus: int(framework.Success)},
// 				},
// 			},
// 			wantStatus:    nil,
// 			wantStatusMap: framework.PluginToStatus{},
// 		},
// 		{
// 			name: "ErrorFilter",
// 			plugins: []*TestPlugin{
// 				{
// 					name: "TestPlugin",
// 					inj:  injectedResult{FilterStatus: int(framework.Error)},
// 				},
// 			},
// 			wantStatus: framework.AsStatus(fmt.Errorf(`running "TestPlugin" filter plugin: %w`, errInjectedFilterStatus)).WithFailedPlugin("TestPlugin"),
// 			wantStatusMap: framework.PluginToStatus{
// 				"TestPlugin": framework.AsStatus(fmt.Errorf(`running "TestPlugin" filter plugin: %w`, errInjectedFilterStatus)).WithFailedPlugin("TestPlugin"),
// 			},
// 		},
// 		{
// 			name: "UnschedulableFilter",
// 			plugins: []*TestPlugin{
// 				{
// 					name: "TestPlugin",
// 					inj:  injectedResult{FilterStatus: int(framework.Unschedulable)},
// 				},
// 			},
// 			wantStatus: framework.NewStatus(framework.Unschedulable, "injected filter status").WithFailedPlugin("TestPlugin"),
// 			wantStatusMap: framework.PluginToStatus{
// 				"TestPlugin": framework.NewStatus(framework.Unschedulable, "injected filter status").WithFailedPlugin("TestPlugin"),
// 			},
// 		},
// 		{
// 			name: "UnschedulableAndUnresolvableFilter",
// 			plugins: []*TestPlugin{
// 				{
// 					name: "TestPlugin",
// 					inj: injectedResult{
// 						FilterStatus: int(framework.UnschedulableAndUnresolvable)},
// 				},
// 			},
// 			wantStatus: framework.NewStatus(framework.UnschedulableAndUnresolvable, "injected filter status").WithFailedPlugin("TestPlugin"),
// 			wantStatusMap: framework.PluginToStatus{
// 				"TestPlugin": framework.NewStatus(framework.UnschedulableAndUnresolvable, "injected filter status").WithFailedPlugin("TestPlugin"),
// 			},
// 		},
// 		// following tests cover multiple-plugins scenarios
// 		{
// 			name: "ErrorAndErrorFilters",
// 			plugins: []*TestPlugin{
// 				{
// 					name: "TestPlugin1",
// 					inj:  injectedResult{FilterStatus: int(framework.Error)},
// 				},

// 				{
// 					name: "TestPlugin2",
// 					inj:  injectedResult{FilterStatus: int(framework.Error)},
// 				},
// 			},
// 			wantStatus: framework.AsStatus(fmt.Errorf(`running "TestPlugin1" filter plugin: %w`, errInjectedFilterStatus)).WithFailedPlugin("TestPlugin1"),
// 			wantStatusMap: framework.PluginToStatus{
// 				"TestPlugin1": framework.AsStatus(fmt.Errorf(`running "TestPlugin1" filter plugin: %w`, errInjectedFilterStatus)).WithFailedPlugin("TestPlugin1"),
// 			},
// 		},
// 		{
// 			name: "SuccessAndSuccessFilters",
// 			plugins: []*TestPlugin{
// 				{
// 					name: "TestPlugin1",
// 					inj:  injectedResult{FilterStatus: int(framework.Success)},
// 				},

// 				{
// 					name: "TestPlugin2",
// 					inj:  injectedResult{FilterStatus: int(framework.Success)},
// 				},
// 			},
// 			wantStatus:    nil,
// 			wantStatusMap: framework.PluginToStatus{},
// 		},
// 		{
// 			name: "ErrorAndSuccessFilters",
// 			plugins: []*TestPlugin{
// 				{
// 					name: "TestPlugin1",
// 					inj:  injectedResult{FilterStatus: int(framework.Error)},
// 				},
// 				{
// 					name: "TestPlugin2",
// 					inj:  injectedResult{FilterStatus: int(framework.Success)},
// 				},
// 			},
// 			wantStatus: framework.AsStatus(fmt.Errorf(`running "TestPlugin1" filter plugin: %w`, errInjectedFilterStatus)).WithFailedPlugin("TestPlugin1"),
// 			wantStatusMap: framework.PluginToStatus{
// 				"TestPlugin1": framework.AsStatus(fmt.Errorf(`running "TestPlugin1" filter plugin: %w`, errInjectedFilterStatus)).WithFailedPlugin("TestPlugin1"),
// 			},
// 		},
// 		{
// 			name: "SuccessAndErrorFilters",
// 			plugins: []*TestPlugin{
// 				{

// 					name: "TestPlugin1",
// 					inj:  injectedResult{FilterStatus: int(framework.Success)},
// 				},
// 				{
// 					name: "TestPlugin2",
// 					inj:  injectedResult{FilterStatus: int(framework.Error)},
// 				},
// 			},
// 			wantStatus: framework.AsStatus(fmt.Errorf(`running "TestPlugin2" filter plugin: %w`, errInjectedFilterStatus)).WithFailedPlugin("TestPlugin2"),
// 			wantStatusMap: framework.PluginToStatus{
// 				"TestPlugin2": framework.AsStatus(fmt.Errorf(`running "TestPlugin2" filter plugin: %w`, errInjectedFilterStatus)).WithFailedPlugin("TestPlugin2"),
// 			},
// 		},
// 		{
// 			name: "SuccessAndUnschedulableFilters",
// 			plugins: []*TestPlugin{
// 				{
// 					name: "TestPlugin1",
// 					inj:  injectedResult{FilterStatus: int(framework.Success)},
// 				},

// 				{
// 					name: "TestPlugin2",
// 					inj:  injectedResult{FilterStatus: int(framework.Unschedulable)},
// 				},
// 			},
// 			wantStatus: framework.NewStatus(framework.Unschedulable, "injected filter status").WithFailedPlugin("TestPlugin2"),
// 			wantStatusMap: framework.PluginToStatus{
// 				"TestPlugin2": framework.NewStatus(framework.Unschedulable, "injected filter status").WithFailedPlugin("TestPlugin2"),
// 			},
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			registry := Registry{}
// 			cfgPls := &v1.Plugins{}
// 			for _, pl := range tt.plugins {
// 				// register all plugins
// 				tmpPl := pl
// 				if err := registry.Register(pl.name,
// 					func(_ runtime.Object, _ framework.Handle) (framework.Plugin, error) {
// 						return tmpPl, nil
// 					}); err != nil {
// 					t.Fatalf("fail to register filter plugin (%s)", pl.name)
// 				}
// 				// append plugins to filter pluginset
// 				cfgPls.Filter.Enabled = append(
// 					cfgPls.Filter.Enabled,
// 					v1.Plugin{Name: pl.name})
// 			}
// 			profile := v1.SchedulerProfile{Plugins: cfgPls}
// 			ctx, cancel := context.WithCancel(context.Background())
// 			defer cancel()
// 			f, err := newFrameworkWithQueueSortAndBind(registry, profile, ctx.Done())
// 			if err != nil {
// 				t.Fatalf("fail to create framework: %s", err)
// 			}
// 			gotStatusMap := f.RunFilterPlugins(ctx, nil, dguest, nil)
// 			gotStatus := gotStatusMap.Merge()
// 			if !reflect.DeepEqual(gotStatus, tt.wantStatus) {
// 				t.Errorf("wrong status code. got: %v, want:%v", gotStatus, tt.wantStatus)
// 			}
// 			if !reflect.DeepEqual(gotStatusMap, tt.wantStatusMap) {
// 				t.Errorf("wrong status map. got: %+v, want: %+v", gotStatusMap, tt.wantStatusMap)
// 			}
// 		})
// 	}
// }

// func TestPostFilterPlugins(t *testing.T) {
// 	tests := []struct {
// 		name       string
// 		plugins    []*TestPlugin
// 		wantStatus *framework.Status
// 	}{
// 		{
// 			name: "a single plugin makes a Dguest schedulable",
// 			plugins: []*TestPlugin{
// 				{
// 					name: "TestPlugin",
// 					inj:  injectedResult{PostFilterStatus: int(framework.Success)},
// 				},
// 			},
// 			wantStatus: framework.NewStatus(framework.Success, "injected status"),
// 		},
// 		{
// 			name: "plugin1 failed to make a Dguest schedulable, followed by plugin2 which makes the Dguest schedulable",
// 			plugins: []*TestPlugin{
// 				{
// 					name: "TestPlugin1",
// 					inj:  injectedResult{PostFilterStatus: int(framework.Unschedulable)},
// 				},
// 				{
// 					name: "TestPlugin2",
// 					inj:  injectedResult{PostFilterStatus: int(framework.Success)},
// 				},
// 			},
// 			wantStatus: framework.NewStatus(framework.Success, "injected status"),
// 		},
// 		{
// 			name: "plugin1 makes a Dguest schedulable, followed by plugin2 which cannot make the Dguest schedulable",
// 			plugins: []*TestPlugin{
// 				{
// 					name: "TestPlugin1",
// 					inj:  injectedResult{PostFilterStatus: int(framework.Success)},
// 				},
// 				{
// 					name: "TestPlugin2",
// 					inj:  injectedResult{PostFilterStatus: int(framework.Unschedulable)},
// 				},
// 			},
// 			wantStatus: framework.NewStatus(framework.Success, "injected status"),
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			registry := Registry{}
// 			cfgPls := &v1.Plugins{}
// 			for _, pl := range tt.plugins {
// 				// register all plugins
// 				tmpPl := pl
// 				if err := registry.Register(pl.name,
// 					func(_ runtime.Object, _ framework.Handle) (framework.Plugin, error) {
// 						return tmpPl, nil
// 					}); err != nil {
// 					t.Fatalf("fail to register postFilter plugin (%s)", pl.name)
// 				}
// 				// append plugins to filter pluginset
// 				cfgPls.PostFilter.Enabled = append(
// 					cfgPls.PostFilter.Enabled,
// 					v1.Plugin{Name: pl.name},
// 				)
// 			}
// 			profile := v1.SchedulerProfile{Plugins: cfgPls}
// 			ctx, cancel := context.WithCancel(context.Background())
// 			defer cancel()
// 			f, err := newFrameworkWithQueueSortAndBind(registry, profile, ctx.Done())
// 			if err != nil {
// 				t.Fatalf("fail to create framework: %s", err)
// 			}
// 			_, gotStatus := f.RunPostFilterPlugins(ctx, nil, dguest, nil)
// 			if !reflect.DeepEqual(gotStatus, tt.wantStatus) {
// 				t.Errorf("Unexpected status. got: %v, want: %v", gotStatus, tt.wantStatus)
// 			}
// 		})
// 	}
// }

// func TestFilterPluginsWithNominatedDguests(t *testing.T) {
// 	tests := []struct {
// 		name            string
// 		preFilterPlugin *TestPlugin
// 		filterPlugin    *TestPlugin
// 		dguest          *v1alpha1.Dguest
// 		nominatedDguest *v1alpha1.Dguest
// 		food            *v1alpha1.Food
// 		foodInfo        *framework.FoodInfo
// 		wantStatus      *framework.Status
// 	}{
// 		{
// 			name:            "food has no nominated dguest",
// 			preFilterPlugin: nil,
// 			filterPlugin:    nil,
// 			dguest:          lowPriorityDguest,
// 			nominatedDguest: nil,
// 			food:            food,
// 			foodInfo:        framework.NewFoodInfo(dguest),
// 			wantStatus:      nil,
// 		},
// 		{
// 			name: "food has a high-priority nominated dguest and all filters succeed",
// 			preFilterPlugin: &TestPlugin{
// 				name: "TestPlugin1",
// 				inj: injectedResult{
// 					PreFilterAddDguestStatus: int(framework.Success),
// 				},
// 			},
// 			filterPlugin: &TestPlugin{
// 				name: "TestPlugin2",
// 				inj: injectedResult{
// 					FilterStatus: int(framework.Success),
// 				},
// 			},
// 			dguest:          lowPriorityDguest,
// 			nominatedDguest: highPriorityDguest,
// 			food:            food,
// 			foodInfo:        framework.NewFoodInfo(dguest),
// 			wantStatus:      nil,
// 		},
// 		{
// 			name: "food has a high-priority nominated dguest and pre filters fail",
// 			preFilterPlugin: &TestPlugin{
// 				name: "TestPlugin1",
// 				inj: injectedResult{
// 					PreFilterAddDguestStatus: int(framework.Error),
// 				},
// 			},
// 			filterPlugin:    nil,
// 			dguest:          lowPriorityDguest,
// 			nominatedDguest: highPriorityDguest,
// 			food:            food,
// 			foodInfo:        framework.NewFoodInfo(dguest),
// 			wantStatus:      framework.AsStatus(fmt.Errorf(`running AddDguest on PreFilter plugin "TestPlugin1": %w`, errInjectedStatus)),
// 		},
// 		{
// 			name: "food has a high-priority nominated dguest and filters fail",
// 			preFilterPlugin: &TestPlugin{
// 				name: "TestPlugin1",
// 				inj: injectedResult{
// 					PreFilterAddDguestStatus: int(framework.Success),
// 				},
// 			},
// 			filterPlugin: &TestPlugin{
// 				name: "TestPlugin2",
// 				inj: injectedResult{
// 					FilterStatus: int(framework.Error),
// 				},
// 			},
// 			dguest:          lowPriorityDguest,
// 			nominatedDguest: highPriorityDguest,
// 			food:            food,
// 			foodInfo:        framework.NewFoodInfo(dguest),
// 			wantStatus:      framework.AsStatus(fmt.Errorf(`running "TestPlugin2" filter plugin: %w`, errInjectedFilterStatus)).WithFailedPlugin("TestPlugin2"),
// 		},
// 		{
// 			name: "food has a low-priority nominated dguest and pre filters return unschedulable",
// 			preFilterPlugin: &TestPlugin{
// 				name: "TestPlugin1",
// 				inj: injectedResult{
// 					PreFilterAddDguestStatus: int(framework.Unschedulable),
// 				},
// 			},
// 			filterPlugin: &TestPlugin{
// 				name: "TestPlugin2",
// 				inj: injectedResult{
// 					FilterStatus: int(framework.Success),
// 				},
// 			},
// 			dguest:          highPriorityDguest,
// 			nominatedDguest: lowPriorityDguest,
// 			food:            food,
// 			foodInfo:        framework.NewFoodInfo(dguest),
// 			wantStatus:      nil,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			registry := Registry{}
// 			cfgPls := &v1.Plugins{}

// 			if tt.preFilterPlugin != nil {
// 				if err := registry.Register(tt.preFilterPlugin.name,
// 					func(_ runtime.Object, _ framework.Handle) (framework.Plugin, error) {
// 						return tt.preFilterPlugin, nil
// 					}); err != nil {
// 					t.Fatalf("fail to register preFilter plugin (%s)", tt.preFilterPlugin.name)
// 				}
// 				cfgPls.PreFilter.Enabled = append(
// 					cfgPls.PreFilter.Enabled,
// 					v1.Plugin{Name: tt.preFilterPlugin.name},
// 				)
// 			}
// 			if tt.filterPlugin != nil {
// 				if err := registry.Register(tt.filterPlugin.name,
// 					func(_ runtime.Object, _ framework.Handle) (framework.Plugin, error) {
// 						return tt.filterPlugin, nil
// 					}); err != nil {
// 					t.Fatalf("fail to register filter plugin (%s)", tt.filterPlugin.name)
// 				}
// 				cfgPls.Filter.Enabled = append(
// 					cfgPls.Filter.Enabled,
// 					v1.Plugin{Name: tt.filterPlugin.name},
// 				)
// 			}

// 			dguestNominator := internalqueue.NewDguestNominator(nil)
// 			if tt.nominatedDguest != nil {
// 				dguestNominator.AddNominatedDguest(
// 					framework.NewDguestInfo(tt.nominatedDguest),
// 					&framework.NominatingInfo{NominatingMode: framework.ModeOverride, NominatedFoodName: foodName})
// 			}
// 			profile := v1.SchedulerProfile{Plugins: cfgPls}
// 			ctx, cancel := context.WithCancel(context.Background())
// 			defer cancel()
// 			f, err := newFrameworkWithQueueSortAndBind(registry, profile, ctx.Done(), WithDguestNominator(dguestNominator))
// 			if err != nil {
// 				t.Fatalf("fail to create framework: %s", err)
// 			}
// 			tt.foodInfo.SetFood(tt.food)
// 			gotStatus := f.RunFilterPluginsWithNominatedDguests(ctx, nil, tt.dguest, tt.foodInfo)
// 			if !reflect.DeepEqual(gotStatus, tt.wantStatus) {
// 				t.Errorf("Unexpected status. got: %v, want: %v", gotStatus, tt.wantStatus)
// 			}
// 		})
// 	}
// }

// func TestPreBindPlugins(t *testing.T) {
// 	tests := []struct {
// 		name       string
// 		plugins    []*TestPlugin
// 		wantStatus *framework.Status
// 	}{
// 		{
// 			name:       "NoPreBindPlugin",
// 			plugins:    []*TestPlugin{},
// 			wantStatus: nil,
// 		},
// 		{
// 			name: "SuccessPreBindPlugins",
// 			plugins: []*TestPlugin{
// 				{
// 					name: "TestPlugin",
// 					inj:  injectedResult{PreBindStatus: int(framework.Success)},
// 				},
// 			},
// 			wantStatus: nil,
// 		},
// 		{
// 			name: "UnshedulablePreBindPlugin",
// 			plugins: []*TestPlugin{
// 				{
// 					name: "TestPlugin",
// 					inj:  injectedResult{PreBindStatus: int(framework.Unschedulable)},
// 				},
// 			},
// 			wantStatus: framework.AsStatus(fmt.Errorf(`running PreBind plugin "TestPlugin": %w`, errInjectedStatus)),
// 		},
// 		{
// 			name: "ErrorPreBindPlugin",
// 			plugins: []*TestPlugin{
// 				{
// 					name: "TestPlugin",
// 					inj:  injectedResult{PreBindStatus: int(framework.Error)},
// 				},
// 			},
// 			wantStatus: framework.AsStatus(fmt.Errorf(`running PreBind plugin "TestPlugin": %w`, errInjectedStatus)),
// 		},
// 		{
// 			name: "UnschedulablePreBindPlugin",
// 			plugins: []*TestPlugin{
// 				{
// 					name: "TestPlugin",
// 					inj:  injectedResult{PreBindStatus: int(framework.UnschedulableAndUnresolvable)},
// 				},
// 			},
// 			wantStatus: framework.AsStatus(fmt.Errorf(`running PreBind plugin "TestPlugin": %w`, errInjectedStatus)),
// 		},
// 		{
// 			name: "SuccessErrorPreBindPlugins",
// 			plugins: []*TestPlugin{
// 				{
// 					name: "TestPlugin",
// 					inj:  injectedResult{PreBindStatus: int(framework.Success)},
// 				},
// 				{
// 					name: "TestPlugin 1",
// 					inj:  injectedResult{PreBindStatus: int(framework.Error)},
// 				},
// 			},
// 			wantStatus: framework.AsStatus(fmt.Errorf(`running PreBind plugin "TestPlugin 1": %w`, errInjectedStatus)),
// 		},
// 		{
// 			name: "ErrorSuccessPreBindPlugin",
// 			plugins: []*TestPlugin{
// 				{
// 					name: "TestPlugin",
// 					inj:  injectedResult{PreBindStatus: int(framework.Error)},
// 				},
// 				{
// 					name: "TestPlugin 1",
// 					inj:  injectedResult{PreBindStatus: int(framework.Success)},
// 				},
// 			},
// 			wantStatus: framework.AsStatus(fmt.Errorf(`running PreBind plugin "TestPlugin": %w`, errInjectedStatus)),
// 		},
// 		{
// 			name: "SuccessSuccessPreBindPlugin",
// 			plugins: []*TestPlugin{
// 				{
// 					name: "TestPlugin",
// 					inj:  injectedResult{PreBindStatus: int(framework.Success)},
// 				},
// 				{
// 					name: "TestPlugin 1",
// 					inj:  injectedResult{PreBindStatus: int(framework.Success)},
// 				},
// 			},
// 			wantStatus: nil,
// 		},
// 		{
// 			name: "ErrorAndErrorPlugins",
// 			plugins: []*TestPlugin{
// 				{
// 					name: "TestPlugin",
// 					inj:  injectedResult{PreBindStatus: int(framework.Error)},
// 				},
// 				{
// 					name: "TestPlugin 1",
// 					inj:  injectedResult{PreBindStatus: int(framework.Error)},
// 				},
// 			},
// 			wantStatus: framework.AsStatus(fmt.Errorf(`running PreBind plugin "TestPlugin": %w`, errInjectedStatus)),
// 		},
// 		{
// 			name: "UnschedulableAndSuccessPreBindPlugin",
// 			plugins: []*TestPlugin{
// 				{
// 					name: "TestPlugin",
// 					inj:  injectedResult{PreBindStatus: int(framework.Unschedulable)},
// 				},
// 				{
// 					name: "TestPlugin 1",
// 					inj:  injectedResult{PreBindStatus: int(framework.Success)},
// 				},
// 			},
// 			wantStatus: framework.AsStatus(fmt.Errorf(`running PreBind plugin "TestPlugin": %w`, errInjectedStatus)),
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			registry := Registry{}
// 			configPlugins := &v1.Plugins{}

// 			for _, pl := range tt.plugins {
// 				tmpPl := pl
// 				if err := registry.Register(pl.name, func(_ runtime.Object, _ framework.Handle) (framework.Plugin, error) {
// 					return tmpPl, nil
// 				}); err != nil {
// 					t.Fatalf("Unable to register pre bind plugins: %s", pl.name)
// 				}

// 				configPlugins.PreBind.Enabled = append(
// 					configPlugins.PreBind.Enabled,
// 					v1.Plugin{Name: pl.name},
// 				)
// 			}
// 			profile := v1.SchedulerProfile{Plugins: configPlugins}
// 			ctx, cancel := context.WithCancel(context.Background())
// 			defer cancel()
// 			f, err := newFrameworkWithQueueSortAndBind(registry, profile, ctx.Done())
// 			if err != nil {
// 				t.Fatalf("fail to create framework: %s", err)
// 			}

// 			status := f.RunPreBindPlugins(ctx, nil, dguest, "")

// 			if !reflect.DeepEqual(status, tt.wantStatus) {
// 				t.Errorf("wrong status code. got %v, want %v", status, tt.wantStatus)
// 			}
// 		})
// 	}
// }

// func TestReservePlugins(t *testing.T) {
// 	tests := []struct {
// 		name       string
// 		plugins    []*TestPlugin
// 		wantStatus *framework.Status
// 	}{
// 		{
// 			name:       "NoReservePlugin",
// 			plugins:    []*TestPlugin{},
// 			wantStatus: nil,
// 		},
// 		{
// 			name: "SuccessReservePlugins",
// 			plugins: []*TestPlugin{
// 				{
// 					name: "TestPlugin",
// 					inj:  injectedResult{ReserveStatus: int(framework.Success)},
// 				},
// 			},
// 			wantStatus: nil,
// 		},
// 		{
// 			name: "UnshedulableReservePlugin",
// 			plugins: []*TestPlugin{
// 				{
// 					name: "TestPlugin",
// 					inj:  injectedResult{ReserveStatus: int(framework.Unschedulable)},
// 				},
// 			},
// 			wantStatus: framework.AsStatus(fmt.Errorf(`running Reserve plugin "TestPlugin": %w`, errInjectedStatus)),
// 		},
// 		{
// 			name: "ErrorReservePlugin",
// 			plugins: []*TestPlugin{
// 				{
// 					name: "TestPlugin",
// 					inj:  injectedResult{ReserveStatus: int(framework.Error)},
// 				},
// 			},
// 			wantStatus: framework.AsStatus(fmt.Errorf(`running Reserve plugin "TestPlugin": %w`, errInjectedStatus)),
// 		},
// 		{
// 			name: "UnschedulableReservePlugin",
// 			plugins: []*TestPlugin{
// 				{
// 					name: "TestPlugin",
// 					inj:  injectedResult{ReserveStatus: int(framework.UnschedulableAndUnresolvable)},
// 				},
// 			},
// 			wantStatus: framework.AsStatus(fmt.Errorf(`running Reserve plugin "TestPlugin": %w`, errInjectedStatus)),
// 		},
// 		{
// 			name: "SuccessSuccessReservePlugins",
// 			plugins: []*TestPlugin{
// 				{
// 					name: "TestPlugin",
// 					inj:  injectedResult{ReserveStatus: int(framework.Success)},
// 				},
// 				{
// 					name: "TestPlugin 1",
// 					inj:  injectedResult{ReserveStatus: int(framework.Success)},
// 				},
// 			},
// 			wantStatus: nil,
// 		},
// 		{
// 			name: "ErrorErrorReservePlugins",
// 			plugins: []*TestPlugin{
// 				{
// 					name: "TestPlugin",
// 					inj:  injectedResult{ReserveStatus: int(framework.Error)},
// 				},
// 				{
// 					name: "TestPlugin 1",
// 					inj:  injectedResult{ReserveStatus: int(framework.Error)},
// 				},
// 			},
// 			wantStatus: framework.AsStatus(fmt.Errorf(`running Reserve plugin "TestPlugin": %w`, errInjectedStatus)),
// 		},
// 		{
// 			name: "SuccessErrorReservePlugins",
// 			plugins: []*TestPlugin{
// 				{
// 					name: "TestPlugin",
// 					inj:  injectedResult{ReserveStatus: int(framework.Success)},
// 				},
// 				{
// 					name: "TestPlugin 1",
// 					inj:  injectedResult{ReserveStatus: int(framework.Error)},
// 				},
// 			},
// 			wantStatus: framework.AsStatus(fmt.Errorf(`running Reserve plugin "TestPlugin 1": %w`, errInjectedStatus)),
// 		},
// 		{
// 			name: "ErrorSuccessReservePlugin",
// 			plugins: []*TestPlugin{
// 				{
// 					name: "TestPlugin",
// 					inj:  injectedResult{ReserveStatus: int(framework.Error)},
// 				},
// 				{
// 					name: "TestPlugin 1",
// 					inj:  injectedResult{ReserveStatus: int(framework.Success)},
// 				},
// 			},
// 			wantStatus: framework.AsStatus(fmt.Errorf(`running Reserve plugin "TestPlugin": %w`, errInjectedStatus)),
// 		},
// 		{
// 			name: "UnschedulableAndSuccessReservePlugin",
// 			plugins: []*TestPlugin{
// 				{
// 					name: "TestPlugin",
// 					inj:  injectedResult{ReserveStatus: int(framework.Unschedulable)},
// 				},
// 				{
// 					name: "TestPlugin 1",
// 					inj:  injectedResult{ReserveStatus: int(framework.Success)},
// 				},
// 			},
// 			wantStatus: framework.AsStatus(fmt.Errorf(`running Reserve plugin "TestPlugin": %w`, errInjectedStatus)),
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			registry := Registry{}
// 			configPlugins := &v1.Plugins{}

// 			for _, pl := range tt.plugins {
// 				tmpPl := pl
// 				if err := registry.Register(pl.name, func(_ runtime.Object, _ framework.Handle) (framework.Plugin, error) {
// 					return tmpPl, nil
// 				}); err != nil {
// 					t.Fatalf("Unable to register pre bind plugins: %s", pl.name)
// 				}

// 				configPlugins.Reserve.Enabled = append(
// 					configPlugins.Reserve.Enabled,
// 					v1.Plugin{Name: pl.name},
// 				)
// 			}
// 			profile := v1.SchedulerProfile{Plugins: configPlugins}
// 			ctx, cancel := context.WithCancel(context.Background())
// 			defer cancel()
// 			f, err := newFrameworkWithQueueSortAndBind(registry, profile, ctx.Done())
// 			if err != nil {
// 				t.Fatalf("fail to create framework: %s", err)
// 			}

// 			status := f.RunReservePluginsReserve(ctx, nil, dguest, "")

// 			if !reflect.DeepEqual(status, tt.wantStatus) {
// 				t.Errorf("wrong status code. got %v, want %v", status, tt.wantStatus)
// 			}
// 		})
// 	}
// }

// func TestPermitPlugins(t *testing.T) {
// 	tests := []struct {
// 		name    string
// 		plugins []*TestPlugin
// 		want    *framework.Status
// 	}{
// 		{
// 			name:    "NilPermitPlugin",
// 			plugins: []*TestPlugin{},
// 			want:    nil,
// 		},
// 		{
// 			name: "SuccessPermitPlugin",
// 			plugins: []*TestPlugin{
// 				{
// 					name: "TestPlugin",
// 					inj:  injectedResult{PermitStatus: int(framework.Success)},
// 				},
// 			},
// 			want: nil,
// 		},
// 		{
// 			name: "UnschedulablePermitPlugin",
// 			plugins: []*TestPlugin{
// 				{
// 					name: "TestPlugin",
// 					inj:  injectedResult{PermitStatus: int(framework.Unschedulable)},
// 				},
// 			},
// 			want: framework.NewStatus(framework.Unschedulable, "injected status").WithFailedPlugin("TestPlugin"),
// 		},
// 		{
// 			name: "ErrorPermitPlugin",
// 			plugins: []*TestPlugin{
// 				{
// 					name: "TestPlugin",
// 					inj:  injectedResult{PermitStatus: int(framework.Error)},
// 				},
// 			},
// 			want: framework.AsStatus(fmt.Errorf(`running Permit plugin "TestPlugin": %w`, errInjectedStatus)).WithFailedPlugin("TestPlugin"),
// 		},
// 		{
// 			name: "UnschedulableAndUnresolvablePermitPlugin",
// 			plugins: []*TestPlugin{
// 				{
// 					name: "TestPlugin",
// 					inj:  injectedResult{PermitStatus: int(framework.UnschedulableAndUnresolvable)},
// 				},
// 			},
// 			want: framework.NewStatus(framework.UnschedulableAndUnresolvable, "injected status").WithFailedPlugin("TestPlugin"),
// 		},
// 		{
// 			name: "WaitPermitPlugin",
// 			plugins: []*TestPlugin{
// 				{
// 					name: "TestPlugin",
// 					inj:  injectedResult{PermitStatus: int(framework.Wait)},
// 				},
// 			},
// 			want: framework.NewStatus(framework.Wait, `one or more plugins asked to wait and no plugin rejected dguest ""`),
// 		},
// 		{
// 			name: "SuccessSuccessPermitPlugin",
// 			plugins: []*TestPlugin{
// 				{
// 					name: "TestPlugin",
// 					inj:  injectedResult{PermitStatus: int(framework.Success)},
// 				},
// 				{
// 					name: "TestPlugin 1",
// 					inj:  injectedResult{PermitStatus: int(framework.Success)},
// 				},
// 			},
// 			want: nil,
// 		},
// 		{
// 			name: "ErrorAndErrorPlugins",
// 			plugins: []*TestPlugin{
// 				{
// 					name: "TestPlugin",
// 					inj:  injectedResult{PermitStatus: int(framework.Error)},
// 				},
// 				{
// 					name: "TestPlugin 1",
// 					inj:  injectedResult{PermitStatus: int(framework.Error)},
// 				},
// 			},
// 			want: framework.AsStatus(fmt.Errorf(`running Permit plugin "TestPlugin": %w`, errInjectedStatus)).WithFailedPlugin("TestPlugin"),
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			registry := Registry{}
// 			configPlugins := &v1.Plugins{}

// 			for _, pl := range tt.plugins {
// 				tmpPl := pl
// 				if err := registry.Register(pl.name, func(_ runtime.Object, _ framework.Handle) (framework.Plugin, error) {
// 					return tmpPl, nil
// 				}); err != nil {
// 					t.Fatalf("Unable to register Permit plugin: %s", pl.name)
// 				}

// 				configPlugins.Permit.Enabled = append(
// 					configPlugins.Permit.Enabled,
// 					v1.Plugin{Name: pl.name},
// 				)
// 			}
// 			profile := v1.SchedulerProfile{Plugins: configPlugins}
// 			ctx, cancel := context.WithCancel(context.Background())
// 			defer cancel()
// 			f, err := newFrameworkWithQueueSortAndBind(registry, profile, ctx.Done())
// 			if err != nil {
// 				t.Fatalf("fail to create framework: %s", err)
// 			}

// 			status := f.RunPermitPlugins(ctx, nil, dguest, "")

// 			if !reflect.DeepEqual(status, tt.want) {
// 				t.Errorf("wrong status code. got %v, want %v", status, tt.want)
// 			}
// 		})
// 	}
// }

// // withMetricsRecorder set metricsRecorder for the scheduling frameworkImpl.
// func withMetricsRecorder(recorder *metricsRecorder) Option {
// 	return func(o *frameworkOptions) {
// 		o.metricsRecorder = recorder
// 	}
// }

// func TestRecordingMetrics(t *testing.T) {
// 	state := &framework.CycleState{}
// 	state.SetRecordPluginMetrics(true)

// 	tests := []struct {
// 		name               string
// 		action             func(f framework.Framework)
// 		inject             injectedResult
// 		wantExtensionPoint string
// 		wantStatus         framework.Code
// 	}{
// 		{
// 			name:               "PreFilter - Success",
// 			action:             func(f framework.Framework) { f.RunPreFilterPlugins(context.Background(), state, dguest) },
// 			wantExtensionPoint: "PreFilter",
// 			wantStatus:         framework.Success,
// 		},
// 		{
// 			name:               "PreScore - Success",
// 			action:             func(f framework.Framework) { f.RunPreScorePlugins(context.Background(), state, dguest, nil) },
// 			wantExtensionPoint: "PreScore",
// 			wantStatus:         framework.Success,
// 		},
// 		{
// 			name:               "Score - Success",
// 			action:             func(f framework.Framework) { f.RunScorePlugins(context.Background(), state, dguest, foods) },
// 			wantExtensionPoint: "Score",
// 			wantStatus:         framework.Success,
// 		},
// 		{
// 			name:               "Reserve - Success",
// 			action:             func(f framework.Framework) { f.RunReservePluginsReserve(context.Background(), state, dguest, "") },
// 			wantExtensionPoint: "Reserve",
// 			wantStatus:         framework.Success,
// 		},
// 		{
// 			name:               "Unreserve - Success",
// 			action:             func(f framework.Framework) { f.RunReservePluginsUnreserve(context.Background(), state, dguest, "") },
// 			wantExtensionPoint: "Unreserve",
// 			wantStatus:         framework.Success,
// 		},
// 		{
// 			name:               "PreBind - Success",
// 			action:             func(f framework.Framework) { f.RunPreBindPlugins(context.Background(), state, dguest, "") },
// 			wantExtensionPoint: "PreBind",
// 			wantStatus:         framework.Success,
// 		},
// 		{
// 			name:               "Bind - Success",
// 			action:             func(f framework.Framework) { f.RunBindPlugins(context.Background(), state, dguest, "") },
// 			wantExtensionPoint: "Bind",
// 			wantStatus:         framework.Success,
// 		},
// 		{
// 			name:               "PostBind - Success",
// 			action:             func(f framework.Framework) { f.RunPostBindPlugins(context.Background(), state, dguest, "") },
// 			wantExtensionPoint: "PostBind",
// 			wantStatus:         framework.Success,
// 		},
// 		{
// 			name:               "Permit - Success",
// 			action:             func(f framework.Framework) { f.RunPermitPlugins(context.Background(), state, dguest, "") },
// 			wantExtensionPoint: "Permit",
// 			wantStatus:         framework.Success,
// 		},

// 		{
// 			name:               "PreFilter - Error",
// 			action:             func(f framework.Framework) { f.RunPreFilterPlugins(context.Background(), state, dguest) },
// 			inject:             injectedResult{PreFilterStatus: int(framework.Error)},
// 			wantExtensionPoint: "PreFilter",
// 			wantStatus:         framework.Error,
// 		},
// 		{
// 			name:               "PreScore - Error",
// 			action:             func(f framework.Framework) { f.RunPreScorePlugins(context.Background(), state, dguest, nil) },
// 			inject:             injectedResult{PreScoreStatus: int(framework.Error)},
// 			wantExtensionPoint: "PreScore",
// 			wantStatus:         framework.Error,
// 		},
// 		{
// 			name:               "Score - Error",
// 			action:             func(f framework.Framework) { f.RunScorePlugins(context.Background(), state, dguest, foods) },
// 			inject:             injectedResult{ScoreStatus: int(framework.Error)},
// 			wantExtensionPoint: "Score",
// 			wantStatus:         framework.Error,
// 		},
// 		{
// 			name:               "Reserve - Error",
// 			action:             func(f framework.Framework) { f.RunReservePluginsReserve(context.Background(), state, dguest, "") },
// 			inject:             injectedResult{ReserveStatus: int(framework.Error)},
// 			wantExtensionPoint: "Reserve",
// 			wantStatus:         framework.Error,
// 		},
// 		{
// 			name:               "PreBind - Error",
// 			action:             func(f framework.Framework) { f.RunPreBindPlugins(context.Background(), state, dguest, "") },
// 			inject:             injectedResult{PreBindStatus: int(framework.Error)},
// 			wantExtensionPoint: "PreBind",
// 			wantStatus:         framework.Error,
// 		},
// 		{
// 			name:               "Bind - Error",
// 			action:             func(f framework.Framework) { f.RunBindPlugins(context.Background(), state, dguest, "") },
// 			inject:             injectedResult{BindStatus: int(framework.Error)},
// 			wantExtensionPoint: "Bind",
// 			wantStatus:         framework.Error,
// 		},
// 		{
// 			name:               "Permit - Error",
// 			action:             func(f framework.Framework) { f.RunPermitPlugins(context.Background(), state, dguest, "") },
// 			inject:             injectedResult{PermitStatus: int(framework.Error)},
// 			wantExtensionPoint: "Permit",
// 			wantStatus:         framework.Error,
// 		},
// 		{
// 			name:               "Permit - Wait",
// 			action:             func(f framework.Framework) { f.RunPermitPlugins(context.Background(), state, dguest, "") },
// 			inject:             injectedResult{PermitStatus: int(framework.Wait)},
// 			wantExtensionPoint: "Permit",
// 			wantStatus:         framework.Wait,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			metrics.Register()
// 			metrics.FrameworkExtensionPointDuration.Reset()
// 			metrics.PluginExecutionDuration.Reset()

// 			plugin := &TestPlugin{name: testPlugin, inj: tt.inject}
// 			r := make(Registry)
// 			r.Register(testPlugin,
// 				func(_ runtime.Object, fh framework.Handle) (framework.Plugin, error) {
// 					return plugin, nil
// 				})
// 			pluginSet := v1.PluginSet{Enabled: []v1.Plugin{{Name: testPlugin, Weight: 1}}}
// 			plugins := &v1.Plugins{
// 				Score:     pluginSet,
// 				PreFilter: pluginSet,
// 				Filter:    pluginSet,
// 				PreScore:  pluginSet,
// 				Reserve:   pluginSet,
// 				Permit:    pluginSet,
// 				PreBind:   pluginSet,
// 				Bind:      pluginSet,
// 				PostBind:  pluginSet,
// 			}

// 			stopCh := make(chan struct{})
// 			recorder := newMetricsRecorder(100, time.Nanosecond, stopCh)
// 			profile := v1.SchedulerProfile{
// 				SchedulerName: testProfileName,
// 				Plugins:       plugins,
// 			}
// 			f, err := newFrameworkWithQueueSortAndBind(r, profile, stopCh, withMetricsRecorder(recorder))
// 			if err != nil {
// 				close(stopCh)
// 				t.Fatalf("Failed to create framework for testing: %v", err)
// 			}

// 			tt.action(f)

// 			// Stop the goroutine which records metrics and ensure it's stopped.
// 			close(stopCh)
// 			<-recorder.isStoppedCh
// 			// Try to clean up the metrics buffer again in case it's not empty.
// 			recorder.flushMetrics()

// 			collectAndCompareFrameworkMetrics(t, tt.wantExtensionPoint, tt.wantStatus)
// 			collectAndComparePluginMetrics(t, tt.wantExtensionPoint, testPlugin, tt.wantStatus)
// 		})
// 	}
// }

// func TestRunBindPlugins(t *testing.T) {
// 	tests := []struct {
// 		name       string
// 		injects    []framework.Code
// 		wantStatus framework.Code
// 	}{
// 		{
// 			name:       "simple success",
// 			injects:    []framework.Code{framework.Success},
// 			wantStatus: framework.Success,
// 		},
// 		{
// 			name:       "error on second",
// 			injects:    []framework.Code{framework.Skip, framework.Error, framework.Success},
// 			wantStatus: framework.Error,
// 		},
// 		{
// 			name:       "all skip",
// 			injects:    []framework.Code{framework.Skip, framework.Skip, framework.Skip},
// 			wantStatus: framework.Skip,
// 		},
// 		{
// 			name:       "error on third, but not reached",
// 			injects:    []framework.Code{framework.Skip, framework.Success, framework.Error},
// 			wantStatus: framework.Success,
// 		},
// 		{
// 			name:       "no bind plugin, returns default binder",
// 			injects:    []framework.Code{},
// 			wantStatus: framework.Success,
// 		},
// 		{
// 			name:       "invalid status",
// 			injects:    []framework.Code{framework.Unschedulable},
// 			wantStatus: framework.Error,
// 		},
// 		{
// 			name:       "simple error",
// 			injects:    []framework.Code{framework.Error},
// 			wantStatus: framework.Error,
// 		},
// 		{
// 			name:       "success on second, returns success",
// 			injects:    []framework.Code{framework.Skip, framework.Success},
// 			wantStatus: framework.Success,
// 		},
// 		{
// 			name:       "invalid status, returns error",
// 			injects:    []framework.Code{framework.Skip, framework.UnschedulableAndUnresolvable},
// 			wantStatus: framework.Error,
// 		},
// 		{
// 			name:       "error after success status, returns success",
// 			injects:    []framework.Code{framework.Success, framework.Error},
// 			wantStatus: framework.Success,
// 		},
// 		{
// 			name:       "success before invalid status, returns success",
// 			injects:    []framework.Code{framework.Success, framework.Error},
// 			wantStatus: framework.Success,
// 		},
// 		{
// 			name:       "success after error status, returns error",
// 			injects:    []framework.Code{framework.Error, framework.Success},
// 			wantStatus: framework.Error,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			metrics.Register()
// 			metrics.FrameworkExtensionPointDuration.Reset()
// 			metrics.PluginExecutionDuration.Reset()

// 			pluginSet := v1.PluginSet{}
// 			r := make(Registry)
// 			for i, inj := range tt.injects {
// 				name := fmt.Sprintf("bind-%d", i)
// 				plugin := &TestPlugin{name: name, inj: injectedResult{BindStatus: int(inj)}}
// 				r.Register(name,
// 					func(_ runtime.Object, fh framework.Handle) (framework.Plugin, error) {
// 						return plugin, nil
// 					})
// 				pluginSet.Enabled = append(pluginSet.Enabled, v1.Plugin{Name: name})
// 			}
// 			plugins := &v1.Plugins{Bind: pluginSet}
// 			stopCh := make(chan struct{})
// 			recorder := newMetricsRecorder(100, time.Nanosecond, stopCh)
// 			profile := v1.SchedulerProfile{
// 				SchedulerName: testProfileName,
// 				Plugins:       plugins,
// 			}
// 			fwk, err := newFrameworkWithQueueSortAndBind(r, profile, stopCh, withMetricsRecorder(recorder))
// 			if err != nil {
// 				close(stopCh)
// 				t.Fatal(err)
// 			}

// 			st := fwk.RunBindPlugins(context.Background(), state, dguest, "")
// 			if st.Code() != tt.wantStatus {
// 				t.Errorf("got status code %s, want %s", st.Code(), tt.wantStatus)
// 			}

// 			// Stop the goroutine which records metrics and ensure it's stopped.
// 			close(stopCh)
// 			<-recorder.isStoppedCh
// 			// Try to clean up the metrics buffer again in case it's not empty.
// 			recorder.flushMetrics()
// 			collectAndCompareFrameworkMetrics(t, "Bind", tt.wantStatus)
// 		})
// 	}
// }

// func TestPermitWaitDurationMetric(t *testing.T) {
// 	tests := []struct {
// 		name    string
// 		inject  injectedResult
// 		wantRes string
// 	}{
// 		{
// 			name: "WaitOnPermit - No Wait",
// 		},
// 		{
// 			name:    "WaitOnPermit - Wait Timeout",
// 			inject:  injectedResult{PermitStatus: int(framework.Wait)},
// 			wantRes: "Unschedulable",
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			metrics.Register()
// 			metrics.PermitWaitDuration.Reset()

// 			plugin := &TestPlugin{name: testPlugin, inj: tt.inject}
// 			r := make(Registry)
// 			err := r.Register(testPlugin,
// 				func(_ runtime.Object, fh framework.Handle) (framework.Plugin, error) {
// 					return plugin, nil
// 				})
// 			if err != nil {
// 				t.Fatal(err)
// 			}
// 			plugins := &v1.Plugins{
// 				Permit: v1.PluginSet{Enabled: []v1.Plugin{{Name: testPlugin, Weight: 1}}},
// 			}
// 			profile := v1.SchedulerProfile{Plugins: plugins}
// 			ctx, cancel := context.WithCancel(context.Background())
// 			defer cancel()
// 			f, err := newFrameworkWithQueueSortAndBind(r, profile, ctx.Done())
// 			if err != nil {
// 				t.Fatalf("Failed to create framework for testing: %v", err)
// 			}

// 			f.RunPermitPlugins(ctx, nil, dguest, "")
// 			f.WaitOnPermit(ctx, dguest)

// 			collectAndComparePermitWaitDuration(t, tt.wantRes)
// 		})
// 	}
// }

// func TestWaitOnPermit(t *testing.T) {
// 	dguest := &v1alpha1.Dguest{
// 		ObjectMeta: metav1.ObjectMeta{
// 			Name: "dguest",
// 			UID:  types.UID("dguest"),
// 		},
// 	}

// 	tests := []struct {
// 		name   string
// 		action func(f framework.Framework)
// 		want   *framework.Status
// 	}{
// 		{
// 			name: "Reject Waiting Dguest",
// 			action: func(f framework.Framework) {
// 				f.GetWaitingDguest(dguest.UID).Reject(permitPlugin, "reject message")
// 			},
// 			want: framework.NewStatus(framework.Unschedulable, "reject message").WithFailedPlugin(permitPlugin),
// 		},
// 		{
// 			name: "Allow Waiting Dguest",
// 			action: func(f framework.Framework) {
// 				f.GetWaitingDguest(dguest.UID).Allow(permitPlugin)
// 			},
// 			want: nil,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			testPermitPlugin := &TestPermitPlugin{}
// 			r := make(Registry)
// 			r.Register(permitPlugin,
// 				func(_ runtime.Object, fh framework.Handle) (framework.Plugin, error) {
// 					return testPermitPlugin, nil
// 				})
// 			plugins := &v1.Plugins{
// 				Permit: v1.PluginSet{Enabled: []v1.Plugin{{Name: permitPlugin, Weight: 1}}},
// 			}
// 			profile := v1.SchedulerProfile{Plugins: plugins}
// 			ctx, cancel := context.WithCancel(context.Background())
// 			defer cancel()
// 			f, err := newFrameworkWithQueueSortAndBind(r, profile, ctx.Done())
// 			if err != nil {
// 				t.Fatalf("Failed to create framework for testing: %v", err)
// 			}

// 			runPermitPluginsStatus := f.RunPermitPlugins(ctx, nil, dguest, "")
// 			if runPermitPluginsStatus.Code() != framework.Wait {
// 				t.Fatalf("Expected RunPermitPlugins to return status %v, but got %v",
// 					framework.Wait, runPermitPluginsStatus.Code())
// 			}

// 			go tt.action(f)

// 			got := f.WaitOnPermit(ctx, dguest)
// 			if !reflect.DeepEqual(tt.want, got) {
// 				t.Errorf("Unexpected status: want %v, but got %v", tt.want, got)
// 			}
// 		})
// 	}
// }

// func TestListPlugins(t *testing.T) {
// 	tests := []struct {
// 		name    string
// 		plugins *v1.Plugins
// 		want    *v1.Plugins
// 	}{
// 		{
// 			name:    "Add empty plugin",
// 			plugins: &v1.Plugins{},
// 			want: &v1.Plugins{
// 				QueueSort: v1.PluginSet{Enabled: []v1.Plugin{{Name: queueSortPlugin}}},
// 				Bind:      v1.PluginSet{Enabled: []v1.Plugin{{Name: bindPlugin}}},
// 			},
// 		},
// 		{
// 			name: "Add multiple plugins",
// 			plugins: &v1.Plugins{
// 				Score: v1.PluginSet{Enabled: []v1.Plugin{{Name: scorePlugin1, Weight: 3}, {Name: scoreWithNormalizePlugin1}}},
// 			},
// 			want: &v1.Plugins{
// 				QueueSort: v1.PluginSet{Enabled: []v1.Plugin{{Name: queueSortPlugin}}},
// 				Bind:      v1.PluginSet{Enabled: []v1.Plugin{{Name: bindPlugin}}},
// 				Score:     v1.PluginSet{Enabled: []v1.Plugin{{Name: scorePlugin1, Weight: 3}, {Name: scoreWithNormalizePlugin1, Weight: 1}}},
// 			},
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			profile := v1.SchedulerProfile{Plugins: tt.plugins}
// 			stopCh := make(chan struct{})
// 			defer close(stopCh)
// 			f, err := newFrameworkWithQueueSortAndBind(registry, profile, stopCh)
// 			if err != nil {
// 				t.Fatalf("Failed to create framework for testing: %v", err)
// 			}
// 			got := f.ListPlugins()
// 			if diff := cmp.Diff(tt.want, got); diff != "" {
// 				t.Errorf("unexpected plugins (-want,+got):\n%s", diff)
// 			}
// 		})
// 	}
// }

// func buildScoreConfigDefaultWeights(ps ...string) *v1.Plugins {
// 	return buildScoreConfigWithWeights(defaultWeights, ps...)
// }

// func buildScoreConfigWithWeights(weights map[string]int32, ps ...string) *v1.Plugins {
// 	var plugins []v1.Plugin
// 	for _, p := range ps {
// 		plugins = append(plugins, v1.Plugin{Name: p, Weight: weights[p]})
// 	}
// 	return &v1.Plugins{Score: v1.PluginSet{Enabled: plugins}}
// }

// type injectedResult struct {
// 	ScoreRes                    int64 `json:"scoreRes,omitempty"`
// 	NormalizeRes                int64 `json:"normalizeRes,omitempty"`
// 	ScoreStatus                 int   `json:"scoreStatus,omitempty"`
// 	NormalizeStatus             int   `json:"normalizeStatus,omitempty"`
// 	PreFilterStatus             int   `json:"preFilterStatus,omitempty"`
// 	PreFilterAddDguestStatus    int   `json:"preFilterAddDguestStatus,omitempty"`
// 	PreFilterRemoveDguestStatus int   `json:"preFilterRemoveDguestStatus,omitempty"`
// 	FilterStatus                int   `json:"filterStatus,omitempty"`
// 	PostFilterStatus            int   `json:"postFilterStatus,omitempty"`
// 	PreScoreStatus              int   `json:"preScoreStatus,omitempty"`
// 	ReserveStatus               int   `json:"reserveStatus,omitempty"`
// 	PreBindStatus               int   `json:"preBindStatus,omitempty"`
// 	BindStatus                  int   `json:"bindStatus,omitempty"`
// 	PermitStatus                int   `json:"permitStatus,omitempty"`
// }

// func setScoreRes(inj injectedResult) (int64, *framework.Status) {
// 	if framework.Code(inj.ScoreStatus) != framework.Success {
// 		return 0, framework.NewStatus(framework.Code(inj.ScoreStatus), "injecting failure.")
// 	}
// 	return inj.ScoreRes, nil
// }

// func injectNormalizeRes(inj injectedResult, scores framework.FoodScoreList) *framework.Status {
// 	if framework.Code(inj.NormalizeStatus) != framework.Success {
// 		return framework.NewStatus(framework.Code(inj.NormalizeStatus), "injecting failure.")
// 	}
// 	for i := range scores {
// 		scores[i].Score = inj.NormalizeRes
// 	}
// 	return nil
// }

// func collectAndComparePluginMetrics(t *testing.T, wantExtensionPoint, wantPlugin string, wantStatus framework.Code) {
// 	t.Helper()
// 	m := metrics.PluginExecutionDuration.WithLabelValues(wantPlugin, wantExtensionPoint, wantStatus.String())

// 	count, err := testutil.GetHistogramMetricCount(m)
// 	if err != nil {
// 		t.Errorf("Failed to get %s sampleCount, err: %v", metrics.PluginExecutionDuration.Name, err)
// 	}
// 	if count == 0 {
// 		t.Error("Expect at least 1 sample")
// 	}
// 	value, err := testutil.GetHistogramMetricValue(m)
// 	if err != nil {
// 		t.Errorf("Failed to get %s value, err: %v", metrics.PluginExecutionDuration.Name, err)
// 	}
// 	if value <= 0 {
// 		t.Errorf("Expect latency to be greater than 0, got: %v", value)
// 	}
// }

// func collectAndCompareFrameworkMetrics(t *testing.T, wantExtensionPoint string, wantStatus framework.Code) {
// 	t.Helper()
// 	m := metrics.FrameworkExtensionPointDuration.WithLabelValues(wantExtensionPoint, wantStatus.String(), testProfileName)

// 	count, err := testutil.GetHistogramMetricCount(m)
// 	if err != nil {
// 		t.Errorf("Failed to get %s sampleCount, err: %v", metrics.FrameworkExtensionPointDuration.Name, err)
// 	}
// 	if count != 1 {
// 		t.Errorf("Expect 1 sample, got: %v", count)
// 	}
// 	value, err := testutil.GetHistogramMetricValue(m)
// 	if err != nil {
// 		t.Errorf("Failed to get %s value, err: %v", metrics.FrameworkExtensionPointDuration.Name, err)
// 	}
// 	if value <= 0 {
// 		t.Errorf("Expect latency to be greater than 0, got: %v", value)
// 	}
// }

// func collectAndComparePermitWaitDuration(t *testing.T, wantRes string) {
// 	m := metrics.PermitWaitDuration.WithLabelValues(wantRes)
// 	count, err := testutil.GetHistogramMetricCount(m)
// 	if err != nil {
// 		t.Errorf("Failed to get %s sampleCount, err: %v", metrics.PermitWaitDuration.Name, err)
// 	}
// 	if wantRes == "" {
// 		if count != 0 {
// 			t.Errorf("Expect 0 sample, got: %v", count)
// 		}
// 	} else {
// 		if count != 1 {
// 			t.Errorf("Expect 1 sample, got: %v", count)
// 		}
// 		value, err := testutil.GetHistogramMetricValue(m)
// 		if err != nil {
// 			t.Errorf("Failed to get %s value, err: %v", metrics.PermitWaitDuration.Name, err)
// 		}
// 		if value <= 0 {
// 			t.Errorf("Expect latency to be greater than 0, got: %v", value)
// 		}
// 	}
// }
