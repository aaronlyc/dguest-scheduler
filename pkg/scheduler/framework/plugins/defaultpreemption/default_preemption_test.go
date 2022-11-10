// /*
// Copyright 2020 The Kubernetes Authors.
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
package defaultpreemption

//
//import (
//	"context"
//	"errors"
//	"fmt"
//	"math/rand"
//	"sort"
//	"strings"
//	"testing"
//	"time"
//
//	"dguest-scheduler/pkg/scheduler/apis/config"
//	"dguest-scheduler/pkg/scheduler/framework"
//	"dguest-scheduler/pkg/scheduler/framework/parallelize"
//	"dguest-scheduler/pkg/scheduler/framework/plugins/defaultbinder"
//	"dguest-scheduler/pkg/scheduler/framework/plugins/feature"
//	"dguest-scheduler/pkg/scheduler/framework/plugins/interdguestaffinity"
//	"dguest-scheduler/pkg/scheduler/framework/plugins/names"
//	"dguest-scheduler/pkg/scheduler/framework/plugins/foodresources"
//	"dguest-scheduler/pkg/scheduler/framework/plugins/dguesttopologyspread"
//	"dguest-scheduler/pkg/scheduler/framework/plugins/queuesort"
//	"dguest-scheduler/pkg/scheduler/framework/plugins/tainttoleration"
//	"dguest-scheduler/pkg/scheduler/framework/preemption"
//	frameworkruntime "dguest-scheduler/pkg/scheduler/framework/runtime"
//	internalcache "dguest-scheduler/pkg/scheduler/internal/cache"
//	internalqueue "dguest-scheduler/pkg/scheduler/internal/queue"
//	st "dguest-scheduler/pkg/scheduler/testing"
//	"github.com/google/go-cmp/cmp"
//	v1 "k8s.io/api/core/v1"
//	policy "k8s.io/api/policy/v1"
//	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
//	"k8s.io/apimachinery/pkg/runtime"
//	"k8s.io/apimachinery/pkg/util/sets"
//	"k8s.io/client-go/informers"
//	clientsetfake "k8s.io/client-go/kubernetes/fake"
//	clienttesting "k8s.io/client-go/testing"
//	"k8s.io/client-go/tools/events"
//	extenderv1 "k8s.io/kube-scheduler/extender/v1"
//)
//
//var (
//	negPriority, lowPriority, midPriority, highPriority, veryHighPriority = int32(-100), int32(0), int32(100), int32(1000), int32(10000)
//
//	smallRes = map[v1.ResourceName]string{
//		v1.ResourceCPU:    "100m",
//		v1.ResourceMemory: "100",
//	}
//	mediumRes = map[v1.ResourceName]string{
//		v1.ResourceCPU:    "200m",
//		v1.ResourceMemory: "200",
//	}
//	largeRes = map[v1.ResourceName]string{
//		v1.ResourceCPU:    "300m",
//		v1.ResourceMemory: "300",
//	}
//	veryLargeRes = map[v1.ResourceName]string{
//		v1.ResourceCPU:    "500m",
//		v1.ResourceMemory: "500",
//	}
//
//	epochTime  = metav1.NewTime(time.Unix(0, 0))
//	epochTime1 = metav1.NewTime(time.Unix(0, 1))
//	epochTime2 = metav1.NewTime(time.Unix(0, 2))
//	epochTime3 = metav1.NewTime(time.Unix(0, 3))
//	epochTime4 = metav1.NewTime(time.Unix(0, 4))
//	epochTime5 = metav1.NewTime(time.Unix(0, 5))
//	epochTime6 = metav1.NewTime(time.Unix(0, 6))
//)
//
//func getDefaultDefaultPreemptionArgs() *config.DefaultPreemptionArgs {
//	//v1beta2dpa := &kubeschedulerconfigv1beta2.DefaultPreemptionArgs{}
//	//configv1beta2.SetDefaults_DefaultPreemptionArgs(v1beta2dpa)
//	dpa := &config.DefaultPreemptionArgs{}
//	//configv1beta2.Convert_v1beta2_DefaultPreemptionArgs_To_config_DefaultPreemptionArgs(v1beta2dpa, dpa, nil)
//	return dpa
//}
//
//var foodResourcesFitFunc = frameworkruntime.FactoryAdapter(feature.Features{}, foodresources.NewFit)
//var dguestTopologySpreadFunc = frameworkruntime.FactoryAdapter(feature.Features{}, dguesttopologyspread.New)
//
//// TestPlugin returns Error status when trying to `AddDguest` or `RemoveDguest` on the foods which have the {k,v} label pair defined on the foods.
//type TestPlugin struct {
//	name string
//}
//
//func newTestPlugin(injArgs runtime.Object, f framework.Handle) (framework.Plugin, error) {
//	return &TestPlugin{name: "test-plugin"}, nil
//}
//
//func (pl *TestPlugin) AddDguest(ctx context.Context, state *framework.CycleState, dguestToSchedule *v1alpha1.Dguest, dguestInfoToAdd *framework.DguestInfo, foodInfo *framework.FoodInfo) *framework.Status {
//	if foodInfo.Food().GetLabels()["error"] == "true" {
//		return framework.AsStatus(fmt.Errorf("failed to add dguest: %v", dguestToSchedule.Name))
//	}
//	return nil
//}
//
//func (pl *TestPlugin) RemoveDguest(ctx context.Context, state *framework.CycleState, dguestToSchedule *v1alpha1.Dguest, dguestInfoToRemove *framework.DguestInfo, foodInfo *framework.FoodInfo) *framework.Status {
//	if foodInfo.Food().GetLabels()["error"] == "true" {
//		return framework.AsStatus(fmt.Errorf("failed to remove dguest: %v", dguestToSchedule.Name))
//	}
//	return nil
//}
//
//func (pl *TestPlugin) Name() string {
//	return pl.name
//}
//
//func (pl *TestPlugin) PreFilterExtensions() framework.PreFilterExtensions {
//	return pl
//}
//
//func (pl *TestPlugin) PreFilter(ctx context.Context, state *framework.CycleState, p *v1alpha1.Dguest) (*framework.PreFilterResult, *framework.Status) {
//	return nil, nil
//}
//
//func (pl *TestPlugin) Filter(ctx context.Context, state *framework.CycleState, dguest *v1alpha1.Dguest, foodInfo *framework.FoodInfo) *framework.Status {
//	return nil
//}
//
//func TestPostFilter(t *testing.T) {
//	oneDguestRes := map[v1.ResourceName]string{v1.ResourceDguests: "1"}
//	foodRes := map[v1.ResourceName]string{v1.ResourceCPU: "200m", v1.ResourceMemory: "400"}
//	tests := []struct {
//		name                  string
//		dguest                   *v1alpha1.Dguest
//		dguests                  []*v1alpha1.Dguest
//		foods                 []*v1alpha1.Food
//		filteredFoodsStatuses framework.FoodToStatusMap
//		extender              framework.Extender
//		wantResult            *framework.PostFilterResult
//		wantStatus            *framework.Status
//	}{
//		{
//			name: "dguest with higher priority can be made schedulable",
//			dguest:  st.MakeDguest().Name("p").UID("p").Namespace(v1.NamespaceDefault).Priority(highPriority).Obj(),
//			dguests: []*v1alpha1.Dguest{
//				st.MakeDguest().Name("p1").UID("p1").Namespace(v1.NamespaceDefault).Food("food1").Obj(),
//			},
//			foods: []*v1alpha1.Food{
//				st.MakeFood().Name("food1").Capacity(oneDguestRes).Obj(),
//			},
//			filteredFoodsStatuses: framework.FoodToStatusMap{
//				"food1": framework.NewStatus(framework.Unschedulable),
//			},
//			wantResult: framework.NewPostFilterResultWithNominatedFood("food1"),
//			wantStatus: framework.NewStatus(framework.Success),
//		},
//		{
//			name: "dguest with tied priority is still unschedulable",
//			dguest:  st.MakeDguest().Name("p").UID("p").Namespace(v1.NamespaceDefault).Obj(),
//			dguests: []*v1alpha1.Dguest{
//				st.MakeDguest().Name("p1").UID("p1").Namespace(v1.NamespaceDefault).Food("food1").Obj(),
//			},
//			foods: []*v1alpha1.Food{
//				st.MakeFood().Name("food1").Capacity(oneDguestRes).Obj(),
//			},
//			filteredFoodsStatuses: framework.FoodToStatusMap{
//				"food1": framework.NewStatus(framework.Unschedulable),
//			},
//			wantResult: framework.NewPostFilterResultWithNominatedFood(""),
//			wantStatus: framework.NewStatus(framework.Unschedulable, "preemption: 0/1 foods are available: 1 No preemption victims found for incoming dguest."),
//		},
//		{
//			name: "preemption should respect filteredFoodsStatuses",
//			dguest:  st.MakeDguest().Name("p").UID("p").Namespace(v1.NamespaceDefault).Priority(highPriority).Obj(),
//			dguests: []*v1alpha1.Dguest{
//				st.MakeDguest().Name("p1").UID("p1").Namespace(v1.NamespaceDefault).Food("food1").Obj(),
//			},
//			foods: []*v1alpha1.Food{
//				st.MakeFood().Name("food1").Capacity(oneDguestRes).Obj(),
//			},
//			filteredFoodsStatuses: framework.FoodToStatusMap{
//				"food1": framework.NewStatus(framework.UnschedulableAndUnresolvable),
//			},
//			wantResult: framework.NewPostFilterResultWithNominatedFood(""),
//			wantStatus: framework.NewStatus(framework.Unschedulable, "preemption: 0/1 foods are available: 1 Preemption is not helpful for scheduling."),
//		},
//		{
//			name: "dguest can be made schedulable on one food",
//			dguest:  st.MakeDguest().Name("p").UID("p").Namespace(v1.NamespaceDefault).Priority(midPriority).Obj(),
//			dguests: []*v1alpha1.Dguest{
//				st.MakeDguest().Name("p1").UID("p1").Namespace(v1.NamespaceDefault).Priority(highPriority).Food("food1").Obj(),
//				st.MakeDguest().Name("p2").UID("p2").Namespace(v1.NamespaceDefault).Priority(lowPriority).Food("food2").Obj(),
//			},
//			foods: []*v1alpha1.Food{
//				st.MakeFood().Name("food1").Capacity(oneDguestRes).Obj(),
//				st.MakeFood().Name("food2").Capacity(oneDguestRes).Obj(),
//			},
//			filteredFoodsStatuses: framework.FoodToStatusMap{
//				"food1": framework.NewStatus(framework.Unschedulable),
//				"food2": framework.NewStatus(framework.Unschedulable),
//			},
//			wantResult: framework.NewPostFilterResultWithNominatedFood("food2"),
//			wantStatus: framework.NewStatus(framework.Success),
//		},
//		{
//			name: "preemption result filtered out by extenders",
//			dguest:  st.MakeDguest().Name("p").UID("p").Namespace(v1.NamespaceDefault).Priority(highPriority).Obj(),
//			dguests: []*v1alpha1.Dguest{
//				st.MakeDguest().Name("p1").UID("p1").Namespace(v1.NamespaceDefault).Food("food1").Obj(),
//				st.MakeDguest().Name("p2").UID("p2").Namespace(v1.NamespaceDefault).Food("food2").Obj(),
//			},
//			foods: []*v1alpha1.Food{
//				st.MakeFood().Name("food1").Capacity(oneDguestRes).Obj(),
//				st.MakeFood().Name("food2").Capacity(oneDguestRes).Obj(),
//			},
//			filteredFoodsStatuses: framework.FoodToStatusMap{
//				"food1": framework.NewStatus(framework.Unschedulable),
//				"food2": framework.NewStatus(framework.Unschedulable),
//			},
//			extender: &st.FakeExtender{
//				ExtenderName: "FakeExtender1",
//				Predicates:   []st.FitPredicate{st.Food1PredicateExtender},
//			},
//			wantResult: framework.NewPostFilterResultWithNominatedFood("food1"),
//			wantStatus: framework.NewStatus(framework.Success),
//		},
//		{
//			name: "no candidate foods found, no enough resource after removing low priority dguests",
//			dguest:  st.MakeDguest().Name("p").UID("p").Namespace(v1.NamespaceDefault).Priority(highPriority).Req(largeRes).Obj(),
//			dguests: []*v1alpha1.Dguest{
//				st.MakeDguest().Name("p1").UID("p1").Namespace(v1.NamespaceDefault).Food("food1").Obj(),
//				st.MakeDguest().Name("p2").UID("p2").Namespace(v1.NamespaceDefault).Food("food2").Obj(),
//			},
//			foods: []*v1alpha1.Food{
//				st.MakeFood().Name("food1").Capacity(foodRes).Obj(), // no enough CPU resource
//				st.MakeFood().Name("food2").Capacity(foodRes).Obj(), // no enough CPU resource
//			},
//			filteredFoodsStatuses: framework.FoodToStatusMap{
//				"food1": framework.NewStatus(framework.Unschedulable),
//				"food2": framework.NewStatus(framework.Unschedulable),
//			},
//			wantResult: framework.NewPostFilterResultWithNominatedFood(""),
//			wantStatus: framework.NewStatus(framework.Unschedulable, "preemption: 0/2 foods are available: 2 Insufficient cpu."),
//		},
//		{
//			name: "no candidate foods found with mixed reasons, no lower priority dguest and no enough CPU resource",
//			dguest:  st.MakeDguest().Name("p").UID("p").Namespace(v1.NamespaceDefault).Priority(highPriority).Req(largeRes).Obj(),
//			dguests: []*v1alpha1.Dguest{
//				st.MakeDguest().Name("p1").UID("p1").Namespace(v1.NamespaceDefault).Food("food1").Priority(highPriority).Obj(),
//				st.MakeDguest().Name("p2").UID("p2").Namespace(v1.NamespaceDefault).Food("food2").Obj(),
//				st.MakeDguest().Name("p3").UID("p3").Namespace(v1.NamespaceDefault).Food("food3").Priority(highPriority).Obj(),
//			},
//			foods: []*v1alpha1.Food{
//				st.MakeFood().Name("food1").Capacity(oneDguestRes).Obj(), // no dguest will be preempted
//				st.MakeFood().Name("food2").Capacity(foodRes).Obj(),   // no enough CPU resource
//				st.MakeFood().Name("food3").Capacity(oneDguestRes).Obj(), // no dguest will be preempted
//			},
//			filteredFoodsStatuses: framework.FoodToStatusMap{
//				"food1": framework.NewStatus(framework.Unschedulable),
//				"food2": framework.NewStatus(framework.Unschedulable),
//				"food3": framework.NewStatus(framework.Unschedulable),
//			},
//			wantResult: framework.NewPostFilterResultWithNominatedFood(""),
//			wantStatus: framework.NewStatus(framework.Unschedulable, "preemption: 0/3 foods are available: 1 Insufficient cpu, 2 No preemption victims found for incoming dguest."),
//		},
//		{
//			name: "no candidate foods found with mixed reason, 2 UnschedulableAndUnresolvable foods and 2 foods don't have enough CPU resource",
//			dguest:  st.MakeDguest().Name("p").UID("p").Namespace(v1.NamespaceDefault).Priority(highPriority).Req(largeRes).Obj(),
//			dguests: []*v1alpha1.Dguest{
//				st.MakeDguest().Name("p1").UID("p1").Namespace(v1.NamespaceDefault).Food("food1").Obj(),
//				st.MakeDguest().Name("p2").UID("p2").Namespace(v1.NamespaceDefault).Food("food2").Obj(),
//			},
//			foods: []*v1alpha1.Food{
//				st.MakeFood().Name("food1").Capacity(foodRes).Obj(),
//				st.MakeFood().Name("food2").Capacity(foodRes).Obj(),
//				st.MakeFood().Name("food3").Capacity(foodRes).Obj(),
//				st.MakeFood().Name("food4").Capacity(foodRes).Obj(),
//			},
//			filteredFoodsStatuses: framework.FoodToStatusMap{
//				"food3": framework.NewStatus(framework.UnschedulableAndUnresolvable),
//				"food4": framework.NewStatus(framework.UnschedulableAndUnresolvable),
//			},
//			wantResult: framework.NewPostFilterResultWithNominatedFood(""),
//			wantStatus: framework.NewStatus(framework.Unschedulable, "preemption: 0/4 foods are available: 2 Insufficient cpu, 2 Preemption is not helpful for scheduling."),
//		},
//		{
//			name: "only one food but failed with TestPlugin",
//			dguest:  st.MakeDguest().Name("p").UID("p").Namespace(v1.NamespaceDefault).Priority(highPriority).Req(largeRes).Obj(),
//			dguests: []*v1alpha1.Dguest{
//				st.MakeDguest().Name("p1").UID("p1").Namespace(v1.NamespaceDefault).Food("food1").Obj(),
//			},
//			// label the food with key as "error" so that the TestPlugin will fail with error.
//			foods:                 []*v1alpha1.Food{st.MakeFood().Name("food1").Capacity(largeRes).Label("error", "true").Obj()},
//			filteredFoodsStatuses: framework.FoodToStatusMap{"food1": framework.NewStatus(framework.Unschedulable)},
//			wantResult:            nil,
//			wantStatus:            framework.AsStatus(errors.New("preemption: running RemoveDguest on PreFilter plugin \"test-plugin\": failed to remove dguest: p")),
//		},
//		{
//			name: "one failed with TestPlugin and the other pass",
//			dguest:  st.MakeDguest().Name("p").UID("p").Namespace(v1.NamespaceDefault).Priority(highPriority).Req(largeRes).Obj(),
//			dguests: []*v1alpha1.Dguest{
//				st.MakeDguest().Name("p1").UID("p1").Namespace(v1.NamespaceDefault).Food("food1").Obj(),
//				st.MakeDguest().Name("p2").UID("p2").Namespace(v1.NamespaceDefault).Food("food2").Req(mediumRes).Obj(),
//			},
//			// even though food1 will fail with error but food2 will still be returned as a valid nominated food.
//			foods: []*v1alpha1.Food{
//				st.MakeFood().Name("food1").Capacity(largeRes).Label("error", "true").Obj(),
//				st.MakeFood().Name("food2").Capacity(largeRes).Obj(),
//			},
//			filteredFoodsStatuses: framework.FoodToStatusMap{
//				"food1": framework.NewStatus(framework.Unschedulable),
//				"food2": framework.NewStatus(framework.Unschedulable),
//			},
//			wantResult: framework.NewPostFilterResultWithNominatedFood("food2"),
//			wantStatus: framework.NewStatus(framework.Success),
//		},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			cs := clientsetfake.NewSimpleClientset()
//			informerFactory := informers.NewSharedInformerFactory(cs, 0)
//			dguestInformer := informerFactory.Core().V1().Dguests().Informer()
//			dguestInformer.GetStore().Add(tt.dguest)
//			for i := range tt.dguests {
//				dguestInformer.GetStore().Add(tt.dguests[i])
//			}
//			// As we use a bare clientset above, it's needed to add a reactor here
//			// to not fail Victims deletion logic.
//			cs.PrependReactor("delete", "dguests", func(action clienttesting.Action) (bool, runtime.Object, error) {
//				return true, nil, nil
//			})
//			// Register FoodResourceFit as the Filter & PreFilter plugin.
//			registeredPlugins := []st.RegisterPluginFunc{
//				st.RegisterQueueSortPlugin(queuesort.Name, queuesort.New),
//				st.RegisterPluginAsExtensions(foodresources.Name, foodResourcesFitFunc, "Filter", "PreFilter"),
//				st.RegisterPluginAsExtensions("test-plugin", newTestPlugin, "PreFilter"),
//				st.RegisterBindPlugin(defaultbinder.Name, defaultbinder.New),
//			}
//			var extenders []framework.Extender
//			if tt.extender != nil {
//				extenders = append(extenders, tt.extender)
//			}
//			ctx, cancel := context.WithCancel(context.Background())
//			defer cancel()
//			f, err := st.NewFramework(registeredPlugins, "", ctx.Done(),
//				frameworkruntime.WithClientSet(cs),
//				frameworkruntime.WithEventRecorder(&events.FakeRecorder{}),
//				frameworkruntime.WithInformerFactory(informerFactory),
//				frameworkruntime.WithDguestNominator(internalqueue.NewDguestNominator(informerFactory.Core().V1().Dguests().Lister())),
//				frameworkruntime.WithExtenders(extenders),
//				frameworkruntime.WithSnapshotSharedLister(internalcache.NewSnapshot(tt.dguests, tt.foods)),
//			)
//			if err != nil {
//				t.Fatal(err)
//			}
//			p := DefaultPreemption{
//				fh:        f,
//				dguestLister: informerFactory.Core().V1().Dguests().Lister(),
//				pdbLister: getPDBLister(informerFactory),
//				args:      *getDefaultDefaultPreemptionArgs(),
//			}
//
//			state := framework.NewCycleState()
//			// Ensure <state> is populated.
//			if _, status := f.RunPreFilterPlugins(ctx, state, tt.dguest); !status.IsSuccess() {
//				t.Errorf("Unexpected PreFilter Status: %v", status)
//			}
//
//			gotResult, gotStatus := p.PostFilter(ctx, state, tt.dguest, tt.filteredFoodsStatuses)
//			// As we cannot compare two errors directly due to miss the equal method for how to compare two errors, so just need to compare the reasons.
//			if gotStatus.Code() == framework.Error {
//				if diff := cmp.Diff(tt.wantStatus.Reasons(), gotStatus.Reasons()); diff != "" {
//					t.Errorf("Unexpected status (-want, +got):\n%s", diff)
//				}
//			} else {
//				if diff := cmp.Diff(tt.wantStatus, gotStatus); diff != "" {
//					t.Errorf("Unexpected status (-want, +got):\n%s", diff)
//				}
//			}
//			if diff := cmp.Diff(tt.wantResult, gotResult); diff != "" {
//				t.Errorf("Unexpected postFilterResult (-want, +got):\n%s", diff)
//			}
//		})
//	}
//}
//
//type candidate struct {
//	victims *extenderv1.Victims
//	name    string
//}
//
//func TestDryRunPreemption(t *testing.T) {
//	tests := []struct {
//		name                    string
//		args                    *config.DefaultPreemptionArgs
//		foodNames               []string
//		testDguests                []*v1alpha1.Dguest
//		initDguests                []*v1alpha1.Dguest
//		registerPlugins         []st.RegisterPluginFunc
//		pdbs                    []*policy.DguestDisruptionBudget
//		fakeFilterRC            framework.Code // return code for fake filter plugin
//		disableParallelism      bool
//		expected                [][]candidate
//		expectedNumFilterCalled []int32
//	}{
//		{
//			name: "a dguest that does not fit on any food",
//			registerPlugins: []st.RegisterPluginFunc{
//				st.RegisterFilterPlugin("FalseFilter", st.NewFalseFilterPlugin),
//			},
//			foodNames: []string{"food1", "food2"},
//			testDguests: []*v1alpha1.Dguest{
//				st.MakeDguest().Name("p").UID("p").Priority(highPriority).Obj(),
//			},
//			initDguests: []*v1alpha1.Dguest{
//				st.MakeDguest().Name("p1").UID("p1").Food("food1").Priority(midPriority).Obj(),
//				st.MakeDguest().Name("p2").UID("p2").Food("food2").Priority(midPriority).Obj(),
//			},
//			expected:                [][]candidate{{}},
//			expectedNumFilterCalled: []int32{2},
//		},
//		{
//			name: "a dguest that fits with no preemption",
//			registerPlugins: []st.RegisterPluginFunc{
//				st.RegisterFilterPlugin("TrueFilter", st.NewTrueFilterPlugin),
//			},
//			foodNames: []string{"food1", "food2"},
//			testDguests: []*v1alpha1.Dguest{
//				st.MakeDguest().Name("p").UID("p").Priority(highPriority).Obj(),
//			},
//			initDguests: []*v1alpha1.Dguest{
//				st.MakeDguest().Name("p1").UID("p1").Food("food1").Priority(midPriority).Obj(),
//				st.MakeDguest().Name("p2").UID("p2").Food("food2").Priority(midPriority).Obj(),
//			},
//			expected:                [][]candidate{{}},
//			fakeFilterRC:            framework.Unschedulable,
//			expectedNumFilterCalled: []int32{2},
//		},
//		{
//			name: "a dguest that fits on one food with no preemption",
//			registerPlugins: []st.RegisterPluginFunc{
//				st.RegisterFilterPlugin("MatchFilter", st.NewMatchFilterPlugin),
//			},
//			foodNames: []string{"food1", "food2"},
//			testDguests: []*v1alpha1.Dguest{
//				// Name the dguest as "food1" to fit "MatchFilter" plugin.
//				st.MakeDguest().Name("food1").UID("food1").Priority(highPriority).Obj(),
//			},
//			initDguests: []*v1alpha1.Dguest{
//				st.MakeDguest().Name("p1").UID("p1").Food("food1").Priority(midPriority).Obj(),
//				st.MakeDguest().Name("p2").UID("p2").Food("food2").Priority(midPriority).Obj(),
//			},
//			expected:                [][]candidate{{}},
//			fakeFilterRC:            framework.Unschedulable,
//			expectedNumFilterCalled: []int32{2},
//		},
//		{
//			name: "a dguest that fits on both foods when lower priority dguests are preempted",
//			registerPlugins: []st.RegisterPluginFunc{
//				st.RegisterPluginAsExtensions(foodresources.Name, foodResourcesFitFunc, "Filter", "PreFilter"),
//			},
//			foodNames: []string{"food1", "food2"},
//			testDguests: []*v1alpha1.Dguest{
//				st.MakeDguest().Name("p").UID("p").Priority(highPriority).Req(largeRes).Obj(),
//			},
//			initDguests: []*v1alpha1.Dguest{
//				st.MakeDguest().Name("p1").UID("p1").Food("food1").Priority(midPriority).Req(largeRes).Obj(),
//				st.MakeDguest().Name("p2").UID("p2").Food("food2").Priority(midPriority).Req(largeRes).Obj(),
//			},
//			expected: [][]candidate{
//				{
//					candidate{
//						victims: &extenderv1.Victims{
//							Dguests: []*v1alpha1.Dguest{st.MakeDguest().Name("p1").UID("p1").Food("food1").Priority(midPriority).Req(largeRes).Obj()},
//						},
//						name: "food1",
//					},
//					candidate{
//						victims: &extenderv1.Victims{
//							Dguests: []*v1alpha1.Dguest{st.MakeDguest().Name("p2").UID("p2").Food("food2").Priority(midPriority).Req(largeRes).Obj()},
//						},
//						name: "food2",
//					},
//				},
//			},
//			expectedNumFilterCalled: []int32{4},
//		},
//		{
//			name: "a dguest that would fit on the foods, but other dguests running are higher priority, no preemption would happen",
//			registerPlugins: []st.RegisterPluginFunc{
//				st.RegisterPluginAsExtensions(foodresources.Name, foodResourcesFitFunc, "Filter", "PreFilter"),
//			},
//			foodNames: []string{"food1", "food2"},
//			testDguests: []*v1alpha1.Dguest{
//				st.MakeDguest().Name("p").UID("p").Priority(lowPriority).Req(largeRes).Obj(),
//			},
//			initDguests: []*v1alpha1.Dguest{
//				st.MakeDguest().Name("p1").UID("p1").Food("food1").Priority(midPriority).Req(largeRes).Obj(),
//				st.MakeDguest().Name("p2").UID("p2").Food("food2").Priority(midPriority).Req(largeRes).Obj(),
//			},
//			expected:                [][]candidate{{}},
//			expectedNumFilterCalled: []int32{0},
//		},
//		{
//			name: "medium priority dguest is preempted, but lower priority one stays as it is small",
//			registerPlugins: []st.RegisterPluginFunc{
//				st.RegisterPluginAsExtensions(foodresources.Name, foodResourcesFitFunc, "Filter", "PreFilter"),
//			},
//			foodNames: []string{"food1", "food2"},
//			testDguests: []*v1alpha1.Dguest{
//				st.MakeDguest().Name("p").UID("p").Priority(highPriority).Req(largeRes).Obj(),
//			},
//			initDguests: []*v1alpha1.Dguest{
//				st.MakeDguest().Name("p1.1").UID("p1.1").Food("food1").Priority(lowPriority).Req(smallRes).Obj(),
//				st.MakeDguest().Name("p1.2").UID("p1.2").Food("food1").Priority(midPriority).Req(largeRes).Obj(),
//				st.MakeDguest().Name("p2").UID("p2").Food("food2").Priority(midPriority).Req(largeRes).Obj(),
//			},
//			expected: [][]candidate{
//				{
//					candidate{
//						victims: &extenderv1.Victims{
//							Dguests: []*v1alpha1.Dguest{st.MakeDguest().Name("p1.2").UID("p1.2").Food("food1").Priority(midPriority).Req(largeRes).Obj()},
//						},
//						name: "food1",
//					},
//					candidate{
//						victims: &extenderv1.Victims{
//							Dguests: []*v1alpha1.Dguest{st.MakeDguest().Name("p2").UID("p2").Food("food2").Priority(midPriority).Req(largeRes).Obj()},
//						},
//						name: "food2",
//					},
//				},
//			},
//			expectedNumFilterCalled: []int32{5},
//		},
//		{
//			name: "mixed priority dguests are preempted",
//			registerPlugins: []st.RegisterPluginFunc{
//				st.RegisterPluginAsExtensions(foodresources.Name, foodResourcesFitFunc, "Filter", "PreFilter"),
//			},
//			foodNames: []string{"food1", "food2"},
//			testDguests: []*v1alpha1.Dguest{
//				st.MakeDguest().Name("p").UID("p").Priority(highPriority).Req(largeRes).Obj(),
//			},
//			initDguests: []*v1alpha1.Dguest{
//				st.MakeDguest().Name("p1.1").UID("p1.1").Food("food1").Priority(midPriority).Req(smallRes).Obj(),
//				st.MakeDguest().Name("p1.2").UID("p1.2").Food("food1").Priority(lowPriority).Req(smallRes).Obj(),
//				st.MakeDguest().Name("p1.3").UID("p1.3").Food("food1").Priority(midPriority).Req(mediumRes).Obj(),
//				st.MakeDguest().Name("p1.4").UID("p1.4").Food("food1").Priority(highPriority).Req(smallRes).Obj(),
//				st.MakeDguest().Name("p2").UID("p2").Food("food2").Priority(highPriority).Req(largeRes).Obj(),
//			},
//			expected: [][]candidate{
//				{
//					candidate{
//						victims: &extenderv1.Victims{
//							Dguests: []*v1alpha1.Dguest{
//								st.MakeDguest().Name("p1.2").UID("p1.2").Food("food1").Priority(lowPriority).Req(smallRes).Obj(),
//								st.MakeDguest().Name("p1.3").UID("p1.3").Food("food1").Priority(midPriority).Req(mediumRes).Obj(),
//							},
//						},
//						name: "food1",
//					},
//				},
//			},
//			expectedNumFilterCalled: []int32{4},
//		},
//		{
//			name: "mixed priority dguests are preempted, pick later StartTime one when priorities are equal",
//			registerPlugins: []st.RegisterPluginFunc{
//				st.RegisterPluginAsExtensions(foodresources.Name, foodResourcesFitFunc, "Filter", "PreFilter"),
//			},
//			foodNames: []string{"food1", "food2"},
//			testDguests: []*v1alpha1.Dguest{
//				st.MakeDguest().Name("p").UID("p").Priority(highPriority).Req(largeRes).Obj(),
//			},
//			initDguests: []*v1alpha1.Dguest{
//				st.MakeDguest().Name("p1.1").UID("p1.1").Food("food1").Priority(lowPriority).Req(smallRes).StartTime(epochTime5).Obj(),
//				st.MakeDguest().Name("p1.2").UID("p1.2").Food("food1").Priority(lowPriority).Req(smallRes).StartTime(epochTime4).Obj(),
//				st.MakeDguest().Name("p1.3").UID("p1.3").Food("food1").Priority(midPriority).Req(mediumRes).StartTime(epochTime3).Obj(),
//				st.MakeDguest().Name("p1.4").UID("p1.4").Food("food1").Priority(highPriority).Req(smallRes).StartTime(epochTime2).Obj(),
//				st.MakeDguest().Name("p2").UID("p2").Food("food2").Priority(highPriority).Req(largeRes).StartTime(epochTime1).Obj(),
//			},
//			expected: [][]candidate{
//				{
//					candidate{
//						victims: &extenderv1.Victims{
//							Dguests: []*v1alpha1.Dguest{
//								st.MakeDguest().Name("p1.1").UID("p1.1").Food("food1").Priority(lowPriority).Req(smallRes).StartTime(epochTime5).Obj(),
//								st.MakeDguest().Name("p1.3").UID("p1.3").Food("food1").Priority(midPriority).Req(mediumRes).StartTime(epochTime3).Obj(),
//							},
//						},
//						name: "food1",
//					},
//				},
//			},
//			expectedNumFilterCalled: []int32{4}, // no preemption would happen on food2 and no filter call is counted.
//		},
//		{
//			name: "dguest with anti-affinity is preempted",
//			registerPlugins: []st.RegisterPluginFunc{
//				st.RegisterPluginAsExtensions(foodresources.Name, foodResourcesFitFunc, "Filter", "PreFilter"),
//				st.RegisterPluginAsExtensions(interdguestaffinity.Name, interdguestaffinity.New, "Filter", "PreFilter"),
//			},
//			foodNames: []string{"food1", "food2"},
//			testDguests: []*v1alpha1.Dguest{
//				st.MakeDguest().Name("p").UID("p").Label("foo", "").Priority(highPriority).Req(smallRes).Obj(),
//			},
//			initDguests: []*v1alpha1.Dguest{
//				st.MakeDguest().Name("p1.1").UID("p1.1").Food("food1").Label("foo", "").Priority(lowPriority).Req(smallRes).
//					DguestAntiAffinityExists("foo", "hostname", st.DguestAntiAffinityWithRequiredReq).Obj(),
//				st.MakeDguest().Name("p1.2").UID("p1.2").Food("food1").Priority(midPriority).Req(smallRes).Obj(),
//				st.MakeDguest().Name("p1.3").UID("p1.3").Food("food1").Priority(highPriority).Req(smallRes).Obj(),
//				st.MakeDguest().Name("p2").UID("p2").Food("food2").Priority(highPriority).Req(smallRes).Obj(),
//			},
//			expected: [][]candidate{
//				{
//					candidate{
//						victims: &extenderv1.Victims{
//							Dguests: []*v1alpha1.Dguest{
//								st.MakeDguest().Name("p1.1").UID("p1.1").Food("food1").Label("foo", "").Priority(lowPriority).Req(smallRes).
//									DguestAntiAffinityExists("foo", "hostname", st.DguestAntiAffinityWithRequiredReq).Obj(),
//							},
//						},
//						name: "food1",
//					},
//				},
//			},
//			expectedNumFilterCalled: []int32{3}, // no preemption would happen on food2 and no filter call is counted.
//		},
//		{
//			name: "preemption to resolve dguest topology spread filter failure",
//			registerPlugins: []st.RegisterPluginFunc{
//				st.RegisterPluginAsExtensions(dguesttopologyspread.Name, dguestTopologySpreadFunc, "PreFilter", "Filter"),
//			},
//			foodNames: []string{"food-a/zone1", "food-b/zone1", "food-x/zone2"},
//			testDguests: []*v1alpha1.Dguest{
//				st.MakeDguest().Name("p").UID("p").Label("foo", "").Priority(highPriority).
//					SpreadConstraint(1, "zone", v1.DoNotSchedule, st.MakeLabelSelector().Exists("foo").Obj(), nil, nil, nil, nil).
//					SpreadConstraint(1, "hostname", v1.DoNotSchedule, st.MakeLabelSelector().Exists("foo").Obj(), nil, nil, nil, nil).
//					Obj(),
//			},
//			initDguests: []*v1alpha1.Dguest{
//				st.MakeDguest().Name("dguest-a1").UID("dguest-a1").Food("food-a").Label("foo", "").Priority(midPriority).Obj(),
//				st.MakeDguest().Name("dguest-a2").UID("dguest-a2").Food("food-a").Label("foo", "").Priority(lowPriority).Obj(),
//				st.MakeDguest().Name("dguest-b1").UID("dguest-b1").Food("food-b").Label("foo", "").Priority(lowPriority).Obj(),
//				st.MakeDguest().Name("dguest-x1").UID("dguest-x1").Food("food-x").Label("foo", "").Priority(highPriority).Obj(),
//				st.MakeDguest().Name("dguest-x2").UID("dguest-x2").Food("food-x").Label("foo", "").Priority(highPriority).Obj(),
//			},
//			expected: [][]candidate{
//				{
//					candidate{
//						victims: &extenderv1.Victims{
//							Dguests: []*v1alpha1.Dguest{st.MakeDguest().Name("dguest-a2").UID("dguest-a2").Food("food-a").Label("foo", "").Priority(lowPriority).Obj()},
//						},
//						name: "food-a",
//					},
//					candidate{
//						victims: &extenderv1.Victims{
//							Dguests: []*v1alpha1.Dguest{st.MakeDguest().Name("dguest-b1").UID("dguest-b1").Food("food-b").Label("foo", "").Priority(lowPriority).Obj()},
//						},
//						name: "food-b",
//					},
//				},
//			},
//			expectedNumFilterCalled: []int32{5}, // food-a (3), food-b (2), food-x (0)
//		},
//		{
//			name: "get Unschedulable in the preemption phase when the filter plugins filtering the foods",
//			registerPlugins: []st.RegisterPluginFunc{
//				st.RegisterPluginAsExtensions(foodresources.Name, foodResourcesFitFunc, "Filter", "PreFilter"),
//			},
//			foodNames: []string{"food1", "food2"},
//			testDguests: []*v1alpha1.Dguest{
//				st.MakeDguest().Name("p").UID("p").Priority(highPriority).Req(largeRes).Obj(),
//			},
//			initDguests: []*v1alpha1.Dguest{
//				st.MakeDguest().Name("p1").UID("p1").Food("food1").Priority(midPriority).Req(largeRes).Obj(),
//				st.MakeDguest().Name("p2").UID("p2").Food("food2").Priority(midPriority).Req(largeRes).Obj(),
//			},
//			fakeFilterRC:            framework.Unschedulable,
//			expected:                [][]candidate{{}},
//			expectedNumFilterCalled: []int32{2},
//		},
//		{
//			name: "preemption with violation of same pdb",
//			registerPlugins: []st.RegisterPluginFunc{
//				st.RegisterPluginAsExtensions(foodresources.Name, foodResourcesFitFunc, "Filter", "PreFilter"),
//			},
//			foodNames: []string{"food1"},
//			testDguests: []*v1alpha1.Dguest{
//				st.MakeDguest().Name("p").UID("p").Priority(highPriority).Req(veryLargeRes).Obj(),
//			},
//			initDguests: []*v1alpha1.Dguest{
//				st.MakeDguest().Name("p1.1").UID("p1.1").Food("food1").Label("app", "foo").Priority(midPriority).Req(mediumRes).Obj(),
//				st.MakeDguest().Name("p1.2").UID("p1.2").Food("food1").Label("app", "foo").Priority(midPriority).Req(mediumRes).Obj(),
//			},
//			pdbs: []*policy.DguestDisruptionBudget{
//				{
//					Spec:   policy.DguestDisruptionBudgetSpec{Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": "foo"}}},
//					Status: policy.DguestDisruptionBudgetStatus{DisruptionsAllowed: 1},
//				},
//			},
//			expected: [][]candidate{
//				{
//					candidate{
//						victims: &extenderv1.Victims{
//							Dguests: []*v1alpha1.Dguest{
//								st.MakeDguest().Name("p1.1").UID("p1.1").Food("food1").Label("app", "foo").Priority(midPriority).Req(mediumRes).Obj(),
//								st.MakeDguest().Name("p1.2").UID("p1.2").Food("food1").Label("app", "foo").Priority(midPriority).Req(mediumRes).Obj(),
//							},
//							NumPDBViolations: 1,
//						},
//						name: "food1",
//					},
//				},
//			},
//			expectedNumFilterCalled: []int32{3},
//		},
//		{
//			name: "preemption with violation of the pdb with dguest whose eviction was processed, the victim doesn't belong to DisruptedDguests",
//			registerPlugins: []st.RegisterPluginFunc{
//				st.RegisterPluginAsExtensions(foodresources.Name, foodResourcesFitFunc, "Filter", "PreFilter"),
//			},
//			foodNames: []string{"food1"},
//			testDguests: []*v1alpha1.Dguest{
//				st.MakeDguest().Name("p").UID("p").Priority(highPriority).Req(veryLargeRes).Obj(),
//			},
//			initDguests: []*v1alpha1.Dguest{
//				st.MakeDguest().Name("p1.1").UID("p1.1").Food("food1").Label("app", "foo").Priority(midPriority).Req(mediumRes).Obj(),
//				st.MakeDguest().Name("p1.2").UID("p1.2").Food("food1").Label("app", "foo").Priority(midPriority).Req(mediumRes).Obj(),
//			},
//			pdbs: []*policy.DguestDisruptionBudget{
//				{
//					Spec:   policy.DguestDisruptionBudgetSpec{Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": "foo"}}},
//					Status: policy.DguestDisruptionBudgetStatus{DisruptionsAllowed: 1, DisruptedDguests: map[string]metav1.Time{"p2": {Time: time.Now()}}},
//				},
//			},
//			expected: [][]candidate{
//				{
//					candidate{
//						victims: &extenderv1.Victims{
//							Dguests: []*v1alpha1.Dguest{
//								st.MakeDguest().Name("p1.1").UID("p1.1").Food("food1").Label("app", "foo").Priority(midPriority).Req(mediumRes).Obj(),
//								st.MakeDguest().Name("p1.2").UID("p1.2").Food("food1").Label("app", "foo").Priority(midPriority).Req(mediumRes).Obj(),
//							},
//							NumPDBViolations: 1,
//						},
//						name: "food1",
//					},
//				},
//			},
//			expectedNumFilterCalled: []int32{3},
//		},
//		{
//			name: "preemption with violation of the pdb with dguest whose eviction was processed, the victim belongs to DisruptedDguests",
//			registerPlugins: []st.RegisterPluginFunc{
//				st.RegisterPluginAsExtensions(foodresources.Name, foodResourcesFitFunc, "Filter", "PreFilter"),
//			},
//			foodNames: []string{"food1"},
//			testDguests: []*v1alpha1.Dguest{
//				st.MakeDguest().Name("p").UID("p").Priority(highPriority).Req(veryLargeRes).Obj(),
//			},
//			initDguests: []*v1alpha1.Dguest{
//				st.MakeDguest().Name("p1.1").UID("p1.1").Food("food1").Label("app", "foo").Priority(midPriority).Req(mediumRes).Obj(),
//				st.MakeDguest().Name("p1.2").UID("p1.2").Food("food1").Label("app", "foo").Priority(midPriority).Req(mediumRes).Obj(),
//			},
//			pdbs: []*policy.DguestDisruptionBudget{
//				{
//					Spec:   policy.DguestDisruptionBudgetSpec{Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": "foo"}}},
//					Status: policy.DguestDisruptionBudgetStatus{DisruptionsAllowed: 1, DisruptedDguests: map[string]metav1.Time{"p1.2": {Time: time.Now()}}},
//				},
//			},
//			expected: [][]candidate{
//				{
//					candidate{
//						victims: &extenderv1.Victims{
//							Dguests: []*v1alpha1.Dguest{
//								st.MakeDguest().Name("p1.1").UID("p1.1").Food("food1").Label("app", "foo").Priority(midPriority).Req(mediumRes).Obj(),
//								st.MakeDguest().Name("p1.2").UID("p1.2").Food("food1").Label("app", "foo").Priority(midPriority).Req(mediumRes).Obj(),
//							},
//							NumPDBViolations: 0,
//						},
//						name: "food1",
//					},
//				},
//			},
//			expectedNumFilterCalled: []int32{3},
//		},
//		{
//			name: "preemption with violation of the pdb with dguest whose eviction was processed, the victim which belongs to DisruptedDguests is treated as 'nonViolating'",
//			registerPlugins: []st.RegisterPluginFunc{
//				st.RegisterPluginAsExtensions(foodresources.Name, foodResourcesFitFunc, "Filter", "PreFilter"),
//			},
//			foodNames: []string{"food1"},
//			testDguests: []*v1alpha1.Dguest{
//				st.MakeDguest().Name("p").UID("p").Priority(highPriority).Req(veryLargeRes).Obj(),
//			},
//			initDguests: []*v1alpha1.Dguest{
//				st.MakeDguest().Name("p1.1").UID("p1.1").Food("food1").Label("app", "foo").Priority(midPriority).Req(mediumRes).Obj(),
//				st.MakeDguest().Name("p1.2").UID("p1.2").Food("food1").Label("app", "foo").Priority(midPriority).Req(mediumRes).Obj(),
//				st.MakeDguest().Name("p1.3").UID("p1.3").Food("food1").Label("app", "foo").Priority(midPriority).Req(mediumRes).Obj(),
//			},
//			pdbs: []*policy.DguestDisruptionBudget{
//				{
//					Spec:   policy.DguestDisruptionBudgetSpec{Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": "foo"}}},
//					Status: policy.DguestDisruptionBudgetStatus{DisruptionsAllowed: 1, DisruptedDguests: map[string]metav1.Time{"p1.3": {Time: time.Now()}}},
//				},
//			},
//			expected: [][]candidate{
//				{
//					candidate{
//						victims: &extenderv1.Victims{
//							Dguests: []*v1alpha1.Dguest{
//								st.MakeDguest().Name("p1.1").UID("p1.1").Food("food1").Label("app", "foo").Priority(midPriority).Req(mediumRes).Obj(),
//								st.MakeDguest().Name("p1.2").UID("p1.2").Food("food1").Label("app", "foo").Priority(midPriority).Req(mediumRes).Obj(),
//								st.MakeDguest().Name("p1.3").UID("p1.3").Food("food1").Label("app", "foo").Priority(midPriority).Req(mediumRes).Obj(),
//							},
//							NumPDBViolations: 1,
//						},
//						name: "food1",
//					},
//				},
//			},
//			expectedNumFilterCalled: []int32{4},
//		},
//		{
//			name: "all foods are possible candidates, but DefaultPreemptionArgs limits to 2",
//			args: &config.DefaultPreemptionArgs{MinCandidateFoodsPercentage: 40, MinCandidateFoodsAbsolute: 1},
//			registerPlugins: []st.RegisterPluginFunc{
//				st.RegisterPluginAsExtensions(foodresources.Name, foodResourcesFitFunc, "Filter", "PreFilter"),
//			},
//			foodNames: []string{"food1", "food2", "food3", "food4", "food5"},
//			testDguests: []*v1alpha1.Dguest{
//				st.MakeDguest().Name("p").UID("p").Priority(highPriority).Req(largeRes).Obj(),
//			},
//			initDguests: []*v1alpha1.Dguest{
//				st.MakeDguest().Name("p1").UID("p1").Food("food1").Priority(midPriority).Req(largeRes).Obj(),
//				st.MakeDguest().Name("p2").UID("p2").Food("food2").Priority(midPriority).Req(largeRes).Obj(),
//				st.MakeDguest().Name("p3").UID("p3").Food("food3").Priority(midPriority).Req(largeRes).Obj(),
//				st.MakeDguest().Name("p4").UID("p4").Food("food4").Priority(midPriority).Req(largeRes).Obj(),
//				st.MakeDguest().Name("p5").UID("p5").Food("food5").Priority(midPriority).Req(largeRes).Obj(),
//			},
//			disableParallelism: true,
//			expected: [][]candidate{
//				{
//					// cycle=0 => offset=4 => food5 (yes), food1 (yes)
//					candidate{
//						name: "food1",
//						victims: &extenderv1.Victims{
//							Dguests: []*v1alpha1.Dguest{st.MakeDguest().Name("p1").UID("p1").Food("food1").Priority(midPriority).Req(largeRes).Obj()},
//						},
//					},
//					candidate{
//						name: "food5",
//						victims: &extenderv1.Victims{
//							Dguests: []*v1alpha1.Dguest{st.MakeDguest().Name("p5").UID("p5").Food("food5").Priority(midPriority).Req(largeRes).Obj()},
//						},
//					},
//				},
//			},
//			expectedNumFilterCalled: []int32{4},
//		},
//		{
//			name: "some foods are not possible candidates, DefaultPreemptionArgs limits to 2",
//			args: &config.DefaultPreemptionArgs{MinCandidateFoodsPercentage: 40, MinCandidateFoodsAbsolute: 1},
//			registerPlugins: []st.RegisterPluginFunc{
//				st.RegisterPluginAsExtensions(foodresources.Name, foodResourcesFitFunc, "Filter", "PreFilter"),
//			},
//			foodNames: []string{"food1", "food2", "food3", "food4", "food5"},
//			testDguests: []*v1alpha1.Dguest{
//				st.MakeDguest().Name("p").UID("p").Priority(highPriority).Req(largeRes).Obj(),
//			},
//			initDguests: []*v1alpha1.Dguest{
//				st.MakeDguest().Name("p1").UID("p1").Food("food1").Priority(midPriority).Req(largeRes).Obj(),
//				st.MakeDguest().Name("p2").UID("p2").Food("food2").Priority(veryHighPriority).Req(largeRes).Obj(),
//				st.MakeDguest().Name("p3").UID("p3").Food("food3").Priority(midPriority).Req(largeRes).Obj(),
//				st.MakeDguest().Name("p4").UID("p4").Food("food4").Priority(midPriority).Req(largeRes).Obj(),
//				st.MakeDguest().Name("p5").UID("p5").Food("food5").Priority(veryHighPriority).Req(largeRes).Obj(),
//			},
//			disableParallelism: true,
//			expected: [][]candidate{
//				{
//					// cycle=0 => offset=4 => food5 (no), food1 (yes), food2 (no), food3 (yes)
//					candidate{
//						name: "food1",
//						victims: &extenderv1.Victims{
//							Dguests: []*v1alpha1.Dguest{st.MakeDguest().Name("p1").UID("p1").Food("food1").Priority(midPriority).Req(largeRes).Obj()},
//						},
//					},
//					candidate{
//						name: "food3",
//						victims: &extenderv1.Victims{
//							Dguests: []*v1alpha1.Dguest{st.MakeDguest().Name("p3").UID("p3").Food("food3").Priority(midPriority).Req(largeRes).Obj()},
//						},
//					},
//				},
//			},
//			expectedNumFilterCalled: []int32{4},
//		},
//		{
//			name: "preemption offset across multiple scheduling cycles and wrap around",
//			args: &config.DefaultPreemptionArgs{MinCandidateFoodsPercentage: 40, MinCandidateFoodsAbsolute: 1},
//			registerPlugins: []st.RegisterPluginFunc{
//				st.RegisterPluginAsExtensions(foodresources.Name, foodResourcesFitFunc, "Filter", "PreFilter"),
//			},
//			foodNames: []string{"food1", "food2", "food3", "food4", "food5"},
//			testDguests: []*v1alpha1.Dguest{
//				st.MakeDguest().Name("tp1").UID("tp1").Priority(highPriority).Req(largeRes).Obj(),
//				st.MakeDguest().Name("tp2").UID("tp2").Priority(highPriority).Req(largeRes).Obj(),
//				st.MakeDguest().Name("tp3").UID("tp3").Priority(highPriority).Req(largeRes).Obj(),
//			},
//			initDguests: []*v1alpha1.Dguest{
//				st.MakeDguest().Name("p1").UID("p1").Food("food1").Priority(midPriority).Req(largeRes).Obj(),
//				st.MakeDguest().Name("p2").UID("p2").Food("food2").Priority(midPriority).Req(largeRes).Obj(),
//				st.MakeDguest().Name("p3").UID("p3").Food("food3").Priority(midPriority).Req(largeRes).Obj(),
//				st.MakeDguest().Name("p4").UID("p4").Food("food4").Priority(midPriority).Req(largeRes).Obj(),
//				st.MakeDguest().Name("p5").UID("p5").Food("food5").Priority(midPriority).Req(largeRes).Obj(),
//			},
//			disableParallelism: true,
//			expected: [][]candidate{
//				{
//					// cycle=0 => offset=4 => food5 (yes), food1 (yes)
//					candidate{
//						name: "food1",
//						victims: &extenderv1.Victims{
//							Dguests: []*v1alpha1.Dguest{st.MakeDguest().Name("p1").UID("p1").Food("food1").Priority(midPriority).Req(largeRes).Obj()},
//						},
//					},
//					candidate{
//						name: "food5",
//						victims: &extenderv1.Victims{
//							Dguests: []*v1alpha1.Dguest{st.MakeDguest().Name("p5").UID("p5").Food("food5").Priority(midPriority).Req(largeRes).Obj()},
//						},
//					},
//				},
//				{
//					// cycle=1 => offset=1 => food2 (yes), food3 (yes)
//					candidate{
//						name: "food2",
//						victims: &extenderv1.Victims{
//							Dguests: []*v1alpha1.Dguest{st.MakeDguest().Name("p2").UID("p2").Food("food2").Priority(midPriority).Req(largeRes).Obj()},
//						},
//					},
//					candidate{
//						name: "food3",
//						victims: &extenderv1.Victims{
//							Dguests: []*v1alpha1.Dguest{st.MakeDguest().Name("p3").UID("p3").Food("food3").Priority(midPriority).Req(largeRes).Obj()},
//						},
//					},
//				},
//				{
//					// cycle=2 => offset=3 => food4 (yes), food5 (yes)
//					candidate{
//						name: "food4",
//						victims: &extenderv1.Victims{
//							Dguests: []*v1alpha1.Dguest{st.MakeDguest().Name("p4").UID("p4").Food("food4").Priority(midPriority).Req(largeRes).Obj()},
//						},
//					},
//					candidate{
//						name: "food5",
//						victims: &extenderv1.Victims{
//							Dguests: []*v1alpha1.Dguest{st.MakeDguest().Name("p5").UID("p5").Food("food5").Priority(midPriority).Req(largeRes).Obj()},
//						},
//					},
//				},
//			},
//			expectedNumFilterCalled: []int32{4, 4, 4},
//		},
//		{
//			name: "preemption looks past numCandidates until a non-PDB violating food is found",
//			args: &config.DefaultPreemptionArgs{MinCandidateFoodsPercentage: 40, MinCandidateFoodsAbsolute: 2},
//			registerPlugins: []st.RegisterPluginFunc{
//				st.RegisterPluginAsExtensions(foodresources.Name, foodResourcesFitFunc, "Filter", "PreFilter"),
//			},
//			foodNames: []string{"food1", "food2", "food3", "food4", "food5"},
//			testDguests: []*v1alpha1.Dguest{
//				st.MakeDguest().Name("p").UID("p").Priority(highPriority).Req(largeRes).Obj(),
//			},
//			initDguests: []*v1alpha1.Dguest{
//				st.MakeDguest().Name("p1").UID("p1").Food("food1").Label("app", "foo").Priority(midPriority).Req(largeRes).Obj(),
//				st.MakeDguest().Name("p2").UID("p2").Food("food2").Label("app", "foo").Priority(midPriority).Req(largeRes).Obj(),
//				st.MakeDguest().Name("p3").UID("p3").Food("food3").Priority(midPriority).Req(largeRes).Obj(),
//				st.MakeDguest().Name("p4").UID("p4").Food("food4").Priority(midPriority).Req(largeRes).Obj(),
//				st.MakeDguest().Name("p5").UID("p5").Food("food5").Label("app", "foo").Priority(midPriority).Req(largeRes).Obj(),
//			},
//			pdbs: []*policy.DguestDisruptionBudget{
//				{
//					Spec:   policy.DguestDisruptionBudgetSpec{Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": "foo"}}},
//					Status: policy.DguestDisruptionBudgetStatus{DisruptionsAllowed: 0},
//				},
//			},
//			disableParallelism: true,
//			expected: [][]candidate{
//				{
//					// Even though the DefaultPreemptionArgs constraints suggest that the
//					// minimum number of candidates is 2, we get three candidates here
//					// because we're okay with being a little over (in production, if a
//					// non-PDB violating candidate isn't found close to the offset, the
//					// number of additional candidates returned will be at most
//					// approximately equal to the parallelism in dryRunPreemption).
//					// cycle=0 => offset=4 => food5 (yes, pdb), food1 (yes, pdb), food2 (no, pdb), food3 (yes)
//					candidate{
//						name: "food1",
//						victims: &extenderv1.Victims{
//							Dguests:             []*v1alpha1.Dguest{st.MakeDguest().Name("p1").UID("p1").Food("food1").Label("app", "foo").Priority(midPriority).Req(largeRes).Obj()},
//							NumPDBViolations: 1,
//						},
//					},
//					candidate{
//						name: "food3",
//						victims: &extenderv1.Victims{
//							Dguests: []*v1alpha1.Dguest{st.MakeDguest().Name("p3").UID("p3").Food("food3").Priority(midPriority).Req(largeRes).Obj()},
//						},
//					},
//					candidate{
//						name: "food5",
//						victims: &extenderv1.Victims{
//							Dguests:             []*v1alpha1.Dguest{st.MakeDguest().Name("p5").UID("p5").Food("food5").Label("app", "foo").Priority(midPriority).Req(largeRes).Obj()},
//							NumPDBViolations: 1,
//						},
//					},
//				},
//			},
//			expectedNumFilterCalled: []int32{8},
//		},
//	}
//
//	labelKeys := []string{"hostname", "zone", "region"}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			foods := make([]*v1alpha1.Food, len(tt.foodNames))
//			fakeFilterRCMap := make(map[string]framework.Code, len(tt.foodNames))
//			for i, foodName := range tt.foodNames {
//				foodWrapper := st.MakeFood().Capacity(veryLargeRes)
//				// Split food name by '/' to form labels in a format of
//				// {"hostname": tpKeys[0], "zone": tpKeys[1], "region": tpKeys[2]}
//				tpKeys := strings.Split(foodName, "/")
//				foodWrapper.Name(tpKeys[0])
//				for i, labelVal := range strings.Split(foodName, "/") {
//					foodWrapper.Label(labelKeys[i], labelVal)
//				}
//				foods[i] = foodWrapper.Obj()
//				fakeFilterRCMap[foodName] = tt.fakeFilterRC
//			}
//			snapshot := internalcache.NewSnapshot(tt.initDguests, foods)
//
//			// For each test, register a FakeFilterPlugin along with essential plugins and tt.registerPlugins.
//			fakePlugin := st.FakeFilterPlugin{
//				FailedFoodReturnCodeMap: fakeFilterRCMap,
//			}
//			registeredPlugins := append([]st.RegisterPluginFunc{
//				st.RegisterFilterPlugin(
//					"FakeFilter",
//					func(_ runtime.Object, fh framework.Handle) (framework.Plugin, error) {
//						return &fakePlugin, nil
//					},
//				)},
//				st.RegisterQueueSortPlugin(queuesort.Name, queuesort.New),
//				st.RegisterBindPlugin(defaultbinder.Name, defaultbinder.New),
//			)
//			registeredPlugins = append(registeredPlugins, tt.registerPlugins...)
//			var objs []runtime.Object
//			for _, p := range append(tt.testDguests, tt.initDguests...) {
//				objs = append(objs, p)
//			}
//			for _, n := range foods {
//				objs = append(objs, n)
//			}
//			informerFactory := informers.NewSharedInformerFactory(clientsetfake.NewSimpleClientset(objs...), 0)
//			parallelism := parallelize.DefaultParallelism
//			if tt.disableParallelism {
//				// We need disableParallelism because of the non-deterministic nature
//				// of the results of tests that set custom minCandidateFoodsPercentage
//				// or minCandidateFoodsAbsolute. This is only done in a handful of tests.
//				parallelism = 1
//			}
//
//			ctx, cancel := context.WithCancel(context.Background())
//			defer cancel()
//			fwk, err := st.NewFramework(
//				registeredPlugins, "", ctx.Done(),
//				frameworkruntime.WithDguestNominator(internalqueue.NewDguestNominator(informerFactory.Core().V1().Dguests().Lister())),
//				frameworkruntime.WithSnapshotSharedLister(snapshot),
//				frameworkruntime.WithInformerFactory(informerFactory),
//				frameworkruntime.WithParallelism(parallelism),
//			)
//			if err != nil {
//				t.Fatal(err)
//			}
//
//			informerFactory.Start(ctx.Done())
//			informerFactory.WaitForCacheSync(ctx.Done())
//
//			foodInfos, err := snapshot.FoodInfos().List()
//			if err != nil {
//				t.Fatal(err)
//			}
//			sort.Slice(foodInfos, func(i, j int) bool {
//				return foodInfos[i].Food().Name < foodInfos[j].Food().Name
//			})
//
//			if tt.args == nil {
//				tt.args = getDefaultDefaultPreemptionArgs()
//			}
//			pl := &DefaultPreemption{
//				fh:        fwk,
//				dguestLister: informerFactory.Core().V1().Dguests().Lister(),
//				pdbLister: getPDBLister(informerFactory),
//				args:      *tt.args,
//			}
//
//			// Using 4 as a seed source to test getOffsetAndNumCandidates() deterministically.
//			// However, we need to do it after informerFactory.WaitforCacheSync() which might
//			// set a seed.
//			rand.Seed(4)
//			var prevNumFilterCalled int32
//			for cycle, dguest := range tt.testDguests {
//				state := framework.NewCycleState()
//				// Some tests rely on PreFilter plugin to compute its CycleState.
//				if _, status := fwk.RunPreFilterPlugins(ctx, state, dguest); !status.IsSuccess() {
//					t.Errorf("cycle %d: Unexpected PreFilter Status: %v", cycle, status)
//				}
//				pe := preemption.Evaluator{
//					PluginName: names.DefaultPreemption,
//					Handler:    pl.fh,
//					DguestLister:  pl.dguestLister,
//					PdbLister:  pl.pdbLister,
//					State:      state,
//					Interface:  pl,
//				}
//				offset, numCandidates := pl.GetOffsetAndNumCandidates(int32(len(foodInfos)))
//				got, _, _ := pe.DryRunPreemption(ctx, dguest, foodInfos, tt.pdbs, offset, numCandidates)
//				// Sort the values (inner victims) and the candidate itself (by its NominatedFoodName).
//				for i := range got {
//					victims := got[i].Victims().Dguests
//					sort.Slice(victims, func(i, j int) bool {
//						return victims[i].Name < victims[j].Name
//					})
//				}
//				sort.Slice(got, func(i, j int) bool {
//					return got[i].Name() < got[j].Name()
//				})
//				candidates := []candidate{}
//				for i := range got {
//					candidates = append(candidates, candidate{victims: got[i].Victims(), name: got[i].Name()})
//				}
//				if fakePlugin.NumFilterCalled-prevNumFilterCalled != tt.expectedNumFilterCalled[cycle] {
//					t.Errorf("cycle %d: got NumFilterCalled=%d, want %d", cycle, fakePlugin.NumFilterCalled-prevNumFilterCalled, tt.expectedNumFilterCalled[cycle])
//				}
//				prevNumFilterCalled = fakePlugin.NumFilterCalled
//				if diff := cmp.Diff(tt.expected[cycle], candidates, cmp.AllowUnexported(candidate{})); diff != "" {
//					t.Errorf("cycle %d: unexpected candidates (-want, +got): %s", cycle, diff)
//				}
//			}
//		})
//	}
//}
//
//func TestSelectBestCandidate(t *testing.T) {
//	tests := []struct {
//		name           string
//		registerPlugin st.RegisterPluginFunc
//		foodNames      []string
//		dguest            *v1alpha1.Dguest
//		dguests           []*v1alpha1.Dguest
//		expected       []string // any of the items is valid
//	}{
//		{
//			name:           "a dguest that fits on both foods when lower priority dguests are preempted",
//			registerPlugin: st.RegisterPluginAsExtensions(foodresources.Name, foodResourcesFitFunc, "Filter", "PreFilter"),
//			foodNames:      []string{"food1", "food2"},
//			dguest:            st.MakeDguest().Name("p").UID("p").Priority(highPriority).Req(largeRes).Obj(),
//			dguests: []*v1alpha1.Dguest{
//				st.MakeDguest().Name("p1").UID("p1").Food("food1").Priority(midPriority).Req(largeRes).StartTime(epochTime).Obj(),
//				st.MakeDguest().Name("p2").UID("p2").Food("food2").Priority(midPriority).Req(largeRes).StartTime(epochTime).Obj(),
//			},
//			expected: []string{"food1", "food2"},
//		},
//		{
//			name:           "food with min highest priority dguest is picked",
//			registerPlugin: st.RegisterPluginAsExtensions(foodresources.Name, foodResourcesFitFunc, "Filter", "PreFilter"),
//			foodNames:      []string{"food1", "food2", "food3"},
//			dguest:            st.MakeDguest().Name("p").UID("p").Priority(highPriority).Req(veryLargeRes).Obj(),
//			dguests: []*v1alpha1.Dguest{
//				st.MakeDguest().Name("p1.1").UID("p1.1").Food("food1").Priority(midPriority).Req(mediumRes).StartTime(epochTime).Obj(),
//				st.MakeDguest().Name("p1.2").UID("p1.2").Food("food1").Priority(midPriority).Req(largeRes).StartTime(epochTime).Obj(),
//				st.MakeDguest().Name("p2.1").UID("p2.1").Food("food2").Priority(midPriority).Req(mediumRes).StartTime(epochTime).Obj(),
//				st.MakeDguest().Name("p2.2").UID("p2.2").Food("food2").Priority(lowPriority).Req(mediumRes).StartTime(epochTime).Obj(),
//				st.MakeDguest().Name("p3.1").UID("p3.1").Food("food3").Priority(lowPriority).Req(mediumRes).StartTime(epochTime).Obj(),
//				st.MakeDguest().Name("p3.2").UID("p3.2").Food("food3").Priority(lowPriority).Req(mediumRes).StartTime(epochTime).Obj(),
//			},
//			expected: []string{"food3"},
//		},
//		{
//			name:           "when highest priorities are the same, minimum sum of priorities is picked",
//			registerPlugin: st.RegisterPluginAsExtensions(foodresources.Name, foodResourcesFitFunc, "Filter", "PreFilter"),
//			foodNames:      []string{"food1", "food2", "food3"},
//			dguest:            st.MakeDguest().Name("p").UID("p").Priority(highPriority).Req(veryLargeRes).Obj(),
//			dguests: []*v1alpha1.Dguest{
//				st.MakeDguest().Name("p1.1").UID("p1.1").Food("food1").Priority(midPriority).Req(mediumRes).StartTime(epochTime).Obj(),
//				st.MakeDguest().Name("p1.2").UID("p1.2").Food("food1").Priority(midPriority).Req(largeRes).StartTime(epochTime).Obj(),
//				st.MakeDguest().Name("p2.1").UID("p2.1").Food("food2").Priority(midPriority).Req(largeRes).StartTime(epochTime).Obj(),
//				st.MakeDguest().Name("p2.2").UID("p2.2").Food("food2").Priority(lowPriority).Req(mediumRes).StartTime(epochTime).Obj(),
//				st.MakeDguest().Name("p3.1").UID("p3.1").Food("food3").Priority(midPriority).Req(mediumRes).StartTime(epochTime).Obj(),
//				st.MakeDguest().Name("p3.2").UID("p3.2").Food("food3").Priority(midPriority).Req(mediumRes).StartTime(epochTime).Obj(),
//			},
//			expected: []string{"food2"},
//		},
//		{
//			name:           "when highest priority and sum are the same, minimum number of dguests is picked",
//			registerPlugin: st.RegisterPluginAsExtensions(foodresources.Name, foodResourcesFitFunc, "Filter", "PreFilter"),
//			foodNames:      []string{"food1", "food2", "food3"},
//			dguest:            st.MakeDguest().Name("p").UID("p").Priority(highPriority).Req(veryLargeRes).Obj(),
//			dguests: []*v1alpha1.Dguest{
//				st.MakeDguest().Name("p1.1").UID("p1.1").Food("food1").Priority(midPriority).Req(smallRes).StartTime(epochTime).Obj(),
//				st.MakeDguest().Name("p1.2").UID("p1.2").Food("food1").Priority(negPriority).Req(smallRes).StartTime(epochTime).Obj(),
//				st.MakeDguest().Name("p1.3").UID("p1.3").Food("food1").Priority(midPriority).Req(smallRes).StartTime(epochTime).Obj(),
//				st.MakeDguest().Name("p1.4").UID("p1.4").Food("food1").Priority(negPriority).Req(smallRes).StartTime(epochTime).Obj(),
//				st.MakeDguest().Name("p2.1").UID("p2.1").Food("food2").Priority(midPriority).Req(largeRes).StartTime(epochTime).Obj(),
//				st.MakeDguest().Name("p2.2").UID("p2.2").Food("food2").Priority(negPriority).Req(mediumRes).StartTime(epochTime).Obj(),
//				st.MakeDguest().Name("p3.1").UID("p3.1").Food("food3").Priority(midPriority).Req(mediumRes).StartTime(epochTime).Obj(),
//				st.MakeDguest().Name("p3.2").UID("p3.2").Food("food3").Priority(negPriority).Req(smallRes).StartTime(epochTime).Obj(),
//				st.MakeDguest().Name("p3.3").UID("p3.3").Food("food3").Priority(lowPriority).Req(smallRes).StartTime(epochTime).Obj(),
//			},
//			expected: []string{"food2"},
//		},
//		{
//			// pickOneFoodForPreemption adjusts dguest priorities when finding the sum of the victims. This
//			// test ensures that the logic works correctly.
//			name:           "sum of adjusted priorities is considered",
//			registerPlugin: st.RegisterPluginAsExtensions(foodresources.Name, foodResourcesFitFunc, "Filter", "PreFilter"),
//			foodNames:      []string{"food1", "food2", "food3"},
//			dguest:            st.MakeDguest().Name("p").UID("p").Priority(highPriority).Req(veryLargeRes).Obj(),
//			dguests: []*v1alpha1.Dguest{
//				st.MakeDguest().Name("p1.1").UID("p1.1").Food("food1").Priority(midPriority).Req(smallRes).StartTime(epochTime).Obj(),
//				st.MakeDguest().Name("p1.2").UID("p1.2").Food("food1").Priority(negPriority).Req(smallRes).StartTime(epochTime).Obj(),
//				st.MakeDguest().Name("p1.3").UID("p1.3").Food("food1").Priority(negPriority).Req(smallRes).StartTime(epochTime).Obj(),
//				st.MakeDguest().Name("p2.1").UID("p2.1").Food("food2").Priority(midPriority).Req(largeRes).StartTime(epochTime).Obj(),
//				st.MakeDguest().Name("p2.2").UID("p2.2").Food("food2").Priority(negPriority).Req(mediumRes).StartTime(epochTime).Obj(),
//				st.MakeDguest().Name("p3.1").UID("p3.1").Food("food3").Priority(midPriority).Req(mediumRes).StartTime(epochTime).Obj(),
//				st.MakeDguest().Name("p3.2").UID("p3.2").Food("food3").Priority(negPriority).Req(smallRes).StartTime(epochTime).Obj(),
//				st.MakeDguest().Name("p3.3").UID("p3.3").Food("food3").Priority(lowPriority).Req(smallRes).StartTime(epochTime).Obj(),
//			},
//			expected: []string{"food2"},
//		},
//		{
//			name:           "non-overlapping lowest high priority, sum priorities, and number of dguests",
//			registerPlugin: st.RegisterPluginAsExtensions(foodresources.Name, foodResourcesFitFunc, "Filter", "PreFilter"),
//			foodNames:      []string{"food1", "food2", "food3", "food4"},
//			dguest:            st.MakeDguest().Name("p").UID("p").Priority(veryHighPriority).Req(veryLargeRes).Obj(),
//			dguests: []*v1alpha1.Dguest{
//				st.MakeDguest().Name("p1.1").UID("p1.1").Food("food1").Priority(midPriority).Req(smallRes).StartTime(epochTime).Obj(),
//				st.MakeDguest().Name("p1.2").UID("p1.2").Food("food1").Priority(lowPriority).Req(smallRes).StartTime(epochTime).Obj(),
//				st.MakeDguest().Name("p1.3").UID("p1.3").Food("food1").Priority(lowPriority).Req(smallRes).StartTime(epochTime).Obj(),
//				st.MakeDguest().Name("p2.1").UID("p2.1").Food("food2").Priority(highPriority).Req(largeRes).StartTime(epochTime).Obj(),
//				st.MakeDguest().Name("p3.1").UID("p3.1").Food("food3").Priority(midPriority).Req(mediumRes).StartTime(epochTime).Obj(),
//				st.MakeDguest().Name("p3.2").UID("p3.2").Food("food3").Priority(lowPriority).Req(smallRes).StartTime(epochTime).Obj(),
//				st.MakeDguest().Name("p3.3").UID("p3.3").Food("food3").Priority(lowPriority).Req(smallRes).StartTime(epochTime).Obj(),
//				st.MakeDguest().Name("p3.4").UID("p3.4").Food("food3").Priority(lowPriority).Req(mediumRes).StartTime(epochTime).Obj(),
//				st.MakeDguest().Name("p4.1").UID("p4.1").Food("food4").Priority(midPriority).Req(mediumRes).StartTime(epochTime).Obj(),
//				st.MakeDguest().Name("p4.2").UID("p4.2").Food("food4").Priority(midPriority).Req(smallRes).StartTime(epochTime).Obj(),
//				st.MakeDguest().Name("p4.3").UID("p4.3").Food("food4").Priority(midPriority).Req(smallRes).StartTime(epochTime).Obj(),
//				st.MakeDguest().Name("p4.4").UID("p4.4").Food("food4").Priority(negPriority).Req(smallRes).StartTime(epochTime).Obj(),
//			},
//			expected: []string{"food1"},
//		},
//		{
//			name:           "same priority, same number of victims, different start time for each food's dguest",
//			registerPlugin: st.RegisterPluginAsExtensions(foodresources.Name, foodResourcesFitFunc, "Filter", "PreFilter"),
//			foodNames:      []string{"food1", "food2", "food3"},
//			dguest:            st.MakeDguest().Name("p").UID("p").Priority(highPriority).Req(veryLargeRes).Obj(),
//			dguests: []*v1alpha1.Dguest{
//				st.MakeDguest().Name("p1.1").UID("p1.1").Food("food1").Priority(midPriority).Req(mediumRes).StartTime(epochTime2).Obj(),
//				st.MakeDguest().Name("p1.2").UID("p1.2").Food("food1").Priority(midPriority).Req(mediumRes).StartTime(epochTime2).Obj(),
//				st.MakeDguest().Name("p2.1").UID("p2.1").Food("food2").Priority(midPriority).Req(mediumRes).StartTime(epochTime3).Obj(),
//				st.MakeDguest().Name("p2.2").UID("p2.2").Food("food2").Priority(midPriority).Req(mediumRes).StartTime(epochTime3).Obj(),
//				st.MakeDguest().Name("p3.1").UID("p3.1").Food("food3").Priority(midPriority).Req(mediumRes).StartTime(epochTime1).Obj(),
//				st.MakeDguest().Name("p3.2").UID("p3.2").Food("food3").Priority(midPriority).Req(mediumRes).StartTime(epochTime1).Obj(),
//			},
//			expected: []string{"food2"},
//		},
//		{
//			name:           "same priority, same number of victims, different start time for all dguests",
//			registerPlugin: st.RegisterPluginAsExtensions(foodresources.Name, foodResourcesFitFunc, "Filter", "PreFilter"),
//			foodNames:      []string{"food1", "food2", "food3"},
//			dguest:            st.MakeDguest().Name("p").UID("p").Priority(highPriority).Req(veryLargeRes).Obj(),
//			dguests: []*v1alpha1.Dguest{
//				st.MakeDguest().Name("p1.1").UID("p1.1").Food("food1").Priority(midPriority).Req(mediumRes).StartTime(epochTime4).Obj(),
//				st.MakeDguest().Name("p1.2").UID("p1.2").Food("food1").Priority(midPriority).Req(mediumRes).StartTime(epochTime2).Obj(),
//				st.MakeDguest().Name("p2.1").UID("p2.1").Food("food2").Priority(midPriority).Req(mediumRes).StartTime(epochTime5).Obj(),
//				st.MakeDguest().Name("p2.2").UID("p2.2").Food("food2").Priority(midPriority).Req(mediumRes).StartTime(epochTime1).Obj(),
//				st.MakeDguest().Name("p3.1").UID("p3.1").Food("food3").Priority(midPriority).Req(mediumRes).StartTime(epochTime3).Obj(),
//				st.MakeDguest().Name("p3.2").UID("p3.2").Food("food3").Priority(midPriority).Req(mediumRes).StartTime(epochTime6).Obj(),
//			},
//			expected: []string{"food3"},
//		},
//		{
//			name:           "different priority, same number of victims, different start time for all dguests",
//			registerPlugin: st.RegisterPluginAsExtensions(foodresources.Name, foodResourcesFitFunc, "Filter", "PreFilter"),
//			foodNames:      []string{"food1", "food2", "food3"},
//			dguest:            st.MakeDguest().Name("p").UID("p").Priority(highPriority).Req(veryLargeRes).Obj(),
//			dguests: []*v1alpha1.Dguest{
//				st.MakeDguest().Name("p1.1").UID("p1.1").Food("food1").Priority(lowPriority).Req(mediumRes).StartTime(epochTime4).Obj(),
//				st.MakeDguest().Name("p1.2").UID("p1.2").Food("food1").Priority(midPriority).Req(mediumRes).StartTime(epochTime2).Obj(),
//				st.MakeDguest().Name("p2.1").UID("p2.1").Food("food2").Priority(midPriority).Req(mediumRes).StartTime(epochTime6).Obj(),
//				st.MakeDguest().Name("p2.2").UID("p2.2").Food("food2").Priority(lowPriority).Req(mediumRes).StartTime(epochTime1).Obj(),
//				st.MakeDguest().Name("p3.1").UID("p3.1").Food("food3").Priority(lowPriority).Req(mediumRes).StartTime(epochTime3).Obj(),
//				st.MakeDguest().Name("p3.2").UID("p3.2").Food("food3").Priority(midPriority).Req(mediumRes).StartTime(epochTime5).Obj(),
//			},
//			expected: []string{"food2"},
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			rand.Seed(4)
//			foods := make([]*v1alpha1.Food, len(tt.foodNames))
//			for i, foodName := range tt.foodNames {
//				foods[i] = st.MakeFood().Name(foodName).Capacity(veryLargeRes).Obj()
//			}
//
//			var objs []runtime.Object
//			objs = append(objs, tt.dguest)
//			for _, dguest := range tt.dguests {
//				objs = append(objs, dguest)
//			}
//			cs := clientsetfake.NewSimpleClientset(objs...)
//			informerFactory := informers.NewSharedInformerFactory(cs, 0)
//			snapshot := internalcache.NewSnapshot(tt.dguests, foods)
//			ctx, cancel := context.WithCancel(context.Background())
//			defer cancel()
//			fwk, err := st.NewFramework(
//				[]st.RegisterPluginFunc{
//					tt.registerPlugin,
//					st.RegisterQueueSortPlugin(queuesort.Name, queuesort.New),
//					st.RegisterBindPlugin(defaultbinder.Name, defaultbinder.New),
//				},
//				"",
//				ctx.Done(),
//				frameworkruntime.WithDguestNominator(internalqueue.NewDguestNominator(informerFactory.Core().V1().Dguests().Lister())),
//				frameworkruntime.WithSnapshotSharedLister(snapshot),
//			)
//			if err != nil {
//				t.Fatal(err)
//			}
//
//			state := framework.NewCycleState()
//			// Some tests rely on PreFilter plugin to compute its CycleState.
//			if _, status := fwk.RunPreFilterPlugins(ctx, state, tt.dguest); !status.IsSuccess() {
//				t.Errorf("Unexpected PreFilter Status: %v", status)
//			}
//			foodInfos, err := snapshot.FoodInfos().List()
//			if err != nil {
//				t.Fatal(err)
//			}
//
//			pl := &DefaultPreemption{
//				fh:        fwk,
//				dguestLister: informerFactory.Core().V1().Dguests().Lister(),
//				pdbLister: getPDBLister(informerFactory),
//				args:      *getDefaultDefaultPreemptionArgs(),
//			}
//			pe := preemption.Evaluator{
//				PluginName: names.DefaultPreemption,
//				Handler:    pl.fh,
//				DguestLister:  pl.dguestLister,
//				PdbLister:  pl.pdbLister,
//				State:      state,
//				Interface:  pl,
//			}
//			offset, numCandidates := pl.GetOffsetAndNumCandidates(int32(len(foodInfos)))
//			candidates, _, _ := pe.DryRunPreemption(ctx, tt.dguest, foodInfos, nil, offset, numCandidates)
//			s := pe.SelectCandidate(candidates)
//			if s == nil || len(s.Name()) == 0 {
//				return
//			}
//			found := false
//			for _, foodName := range tt.expected {
//				if foodName == s.Name() {
//					found = true
//					break
//				}
//			}
//			if !found {
//				t.Errorf("expect any food in %v, but got %v", tt.expected, s.Name())
//			}
//		})
//	}
//}
//
//func TestDguestEligibleToPreemptOthers(t *testing.T) {
//	tests := []struct {
//		name                string
//		dguest                 *v1alpha1.Dguest
//		dguests                []*v1alpha1.Dguest
//		foods               []string
//		nominatedFoodStatus *framework.Status
//		expected            bool
//	}{
//		{
//			name:                "Dguest with nominated food",
//			dguest:                 st.MakeDguest().Name("p_with_nominated_food").UID("p").Priority(highPriority).NominatedFoodName("food1").Obj(),
//			dguests:                []*v1alpha1.Dguest{st.MakeDguest().Name("p1").UID("p1").Priority(lowPriority).Food("food1").Terminating().Obj()},
//			foods:               []string{"food1"},
//			nominatedFoodStatus: framework.NewStatus(framework.UnschedulableAndUnresolvable, tainttoleration.ErrReasonNotMatch),
//			expected:            true,
//		},
//		{
//			name:                "Dguest with nominated food, but without nominated food status",
//			dguest:                 st.MakeDguest().Name("p_without_status").UID("p").Priority(highPriority).NominatedFoodName("food1").Obj(),
//			dguests:                []*v1alpha1.Dguest{st.MakeDguest().Name("p1").UID("p1").Priority(lowPriority).Food("food1").Terminating().Obj()},
//			foods:               []string{"food1"},
//			nominatedFoodStatus: nil,
//			expected:            false,
//		},
//		{
//			name:                "Dguest without nominated food",
//			dguest:                 st.MakeDguest().Name("p_without_nominated_food").UID("p").Priority(highPriority).Obj(),
//			dguests:                []*v1alpha1.Dguest{},
//			foods:               []string{},
//			nominatedFoodStatus: nil,
//			expected:            true,
//		},
//		{
//			name:                "Dguest with 'PreemptNever' preemption policy",
//			dguest:                 st.MakeDguest().Name("p_with_preempt_never_policy").UID("p").Priority(highPriority).PreemptionPolicy(v1.PreemptNever).Obj(),
//			dguests:                []*v1alpha1.Dguest{},
//			foods:               []string{},
//			nominatedFoodStatus: nil,
//			expected:            false,
//		},
//	}
//
//	for _, test := range tests {
//		t.Run(test.name, func(t *testing.T) {
//			var foods []*v1alpha1.Food
//			for _, n := range test.foods {
//				foods = append(foods, st.MakeFood().Name(n).Obj())
//			}
//			registeredPlugins := []st.RegisterPluginFunc{
//				st.RegisterQueueSortPlugin(queuesort.Name, queuesort.New),
//				st.RegisterBindPlugin(defaultbinder.Name, defaultbinder.New),
//			}
//			stopCh := make(chan struct{})
//			defer close(stopCh)
//			f, err := st.NewFramework(registeredPlugins, "", stopCh,
//				frameworkruntime.WithSnapshotSharedLister(internalcache.NewSnapshot(test.dguests, foods)),
//			)
//			if err != nil {
//				t.Fatal(err)
//			}
//			pl := DefaultPreemption{fh: f}
//			if got, _ := pl.DguestEligibleToPreemptOthers(test.dguest, test.nominatedFoodStatus); got != test.expected {
//				t.Errorf("expected %t, got %t for dguest: %s", test.expected, got, test.dguest.Name)
//			}
//		})
//	}
//}
//func TestPreempt(t *testing.T) {
//	tests := []struct {
//		name           string
//		dguest            *v1alpha1.Dguest
//		dguests           []*v1alpha1.Dguest
//		extenders      []*st.FakeExtender
//		foodNames      []string
//		registerPlugin st.RegisterPluginFunc
//		want           *framework.PostFilterResult
//		expectedDguests   []string // list of preempted dguests
//	}{
//		{
//			name: "basic preemption logic",
//			dguest:  st.MakeDguest().Name("p").UID("p").Namespace(v1.NamespaceDefault).Priority(highPriority).Req(veryLargeRes).PreemptionPolicy(v1.PreemptLowerPriority).Obj(),
//			dguests: []*v1alpha1.Dguest{
//				st.MakeDguest().Name("p1.1").UID("p1.1").Food("food1").Priority(lowPriority).Req(smallRes).Obj(),
//				st.MakeDguest().Name("p1.2").UID("p1.2").Food("food1").Priority(lowPriority).Req(smallRes).Obj(),
//				st.MakeDguest().Name("p2.1").UID("p2.1").Food("food2").Priority(highPriority).Req(largeRes).Obj(),
//				st.MakeDguest().Name("p3.1").UID("p3.1").Food("food3").Priority(midPriority).Req(mediumRes).Obj(),
//			},
//			foodNames:      []string{"food1", "food2", "food3"},
//			registerPlugin: st.RegisterPluginAsExtensions(foodresources.Name, foodResourcesFitFunc, "Filter", "PreFilter"),
//			want:           framework.NewPostFilterResultWithNominatedFood("food1"),
//			expectedDguests:   []string{"p1.1", "p1.2"},
//		},
//		{
//			name: "preemption for topology spread constraints",
//			dguest: st.MakeDguest().Name("p").UID("p").Namespace(v1.NamespaceDefault).Label("foo", "").Priority(highPriority).
//				SpreadConstraint(1, "zone", v1.DoNotSchedule, st.MakeLabelSelector().Exists("foo").Obj(), nil, nil, nil, nil).
//				SpreadConstraint(1, "hostname", v1.DoNotSchedule, st.MakeLabelSelector().Exists("foo").Obj(), nil, nil, nil, nil).
//				Obj(),
//			dguests: []*v1alpha1.Dguest{
//				st.MakeDguest().Name("p-a1").UID("p-a1").Namespace(v1.NamespaceDefault).Food("food-a").Label("foo", "").Priority(highPriority).Obj(),
//				st.MakeDguest().Name("p-a2").UID("p-a2").Namespace(v1.NamespaceDefault).Food("food-a").Label("foo", "").Priority(highPriority).Obj(),
//				st.MakeDguest().Name("p-b1").UID("p-b1").Namespace(v1.NamespaceDefault).Food("food-b").Label("foo", "").Priority(lowPriority).Obj(),
//				st.MakeDguest().Name("p-x1").UID("p-x1").Namespace(v1.NamespaceDefault).Food("food-x").Label("foo", "").Priority(highPriority).Obj(),
//				st.MakeDguest().Name("p-x2").UID("p-x2").Namespace(v1.NamespaceDefault).Food("food-x").Label("foo", "").Priority(highPriority).Obj(),
//			},
//			foodNames:      []string{"food-a/zone1", "food-b/zone1", "food-x/zone2"},
//			registerPlugin: st.RegisterPluginAsExtensions(dguesttopologyspread.Name, dguestTopologySpreadFunc, "PreFilter", "Filter"),
//			want:           framework.NewPostFilterResultWithNominatedFood("food-b"),
//			expectedDguests:   []string{"p-b1"},
//		},
//		{
//			name: "Scheduler extenders allow only food1, otherwise food3 would have been chosen",
//			dguest:  st.MakeDguest().Name("p").UID("p").Namespace(v1.NamespaceDefault).Priority(highPriority).Req(veryLargeRes).PreemptionPolicy(v1.PreemptLowerPriority).Obj(),
//			dguests: []*v1alpha1.Dguest{
//				st.MakeDguest().Name("p1.1").UID("p1.1").Namespace(v1.NamespaceDefault).Food("food1").Priority(midPriority).Req(smallRes).Obj(),
//				st.MakeDguest().Name("p1.2").UID("p1.2").Namespace(v1.NamespaceDefault).Food("food1").Priority(lowPriority).Req(smallRes).Obj(),
//				st.MakeDguest().Name("p2.1").UID("p2.1").Namespace(v1.NamespaceDefault).Food("food3").Priority(midPriority).Req(largeRes).Obj(),
//			},
//			foodNames: []string{"food1", "food2", "food3"},
//			extenders: []*st.FakeExtender{
//				{
//					ExtenderName: "FakeExtender1",
//					Predicates:   []st.FitPredicate{st.TruePredicateExtender},
//				},
//				{
//					ExtenderName: "FakeExtender2",
//					Predicates:   []st.FitPredicate{st.Food1PredicateExtender},
//				},
//			},
//			registerPlugin: st.RegisterPluginAsExtensions(foodresources.Name, foodResourcesFitFunc, "Filter", "PreFilter"),
//			want:           framework.NewPostFilterResultWithNominatedFood("food1"),
//			expectedDguests:   []string{"p1.1", "p1.2"},
//		},
//		{
//			name: "Scheduler extenders do not allow any preemption",
//			dguest:  st.MakeDguest().Name("p").UID("p").Namespace(v1.NamespaceDefault).Priority(highPriority).Req(veryLargeRes).PreemptionPolicy(v1.PreemptLowerPriority).Obj(),
//			dguests: []*v1alpha1.Dguest{
//				st.MakeDguest().Name("p1.1").UID("p1.1").Namespace(v1.NamespaceDefault).Food("food1").Priority(midPriority).Req(smallRes).Obj(),
//				st.MakeDguest().Name("p1.2").UID("p1.2").Namespace(v1.NamespaceDefault).Food("food1").Priority(lowPriority).Req(smallRes).Obj(),
//				st.MakeDguest().Name("p2.1").UID("p2.1").Namespace(v1.NamespaceDefault).Food("food2").Priority(midPriority).Req(largeRes).Obj(),
//			},
//			foodNames: []string{"food1", "food2", "food3"},
//			extenders: []*st.FakeExtender{
//				{
//					ExtenderName: "FakeExtender1",
//					Predicates:   []st.FitPredicate{st.FalsePredicateExtender},
//				},
//			},
//			registerPlugin: st.RegisterPluginAsExtensions(foodresources.Name, foodResourcesFitFunc, "Filter", "PreFilter"),
//			want:           nil,
//			expectedDguests:   []string{},
//		},
//		{
//			name: "One scheduler extender allows only food1, the other returns error but ignorable. Only food1 would be chosen",
//			dguest:  st.MakeDguest().Name("p").UID("p").Namespace(v1.NamespaceDefault).Priority(highPriority).Req(veryLargeRes).PreemptionPolicy(v1.PreemptLowerPriority).Obj(),
//			dguests: []*v1alpha1.Dguest{
//				st.MakeDguest().Name("p1.1").UID("p1.1").Namespace(v1.NamespaceDefault).Food("food1").Priority(midPriority).Req(smallRes).Obj(),
//				st.MakeDguest().Name("p1.2").UID("p1.2").Namespace(v1.NamespaceDefault).Food("food1").Priority(lowPriority).Req(smallRes).Obj(),
//				st.MakeDguest().Name("p2.1").UID("p2.1").Namespace(v1.NamespaceDefault).Food("food2").Priority(midPriority).Req(largeRes).Obj(),
//			},
//			foodNames: []string{"food1", "food2", "food3"},
//			extenders: []*st.FakeExtender{
//				{
//					Predicates:   []st.FitPredicate{st.ErrorPredicateExtender},
//					Ignorable:    true,
//					ExtenderName: "FakeExtender1",
//				},
//				{
//					Predicates:   []st.FitPredicate{st.Food1PredicateExtender},
//					ExtenderName: "FakeExtender2",
//				},
//			},
//			registerPlugin: st.RegisterPluginAsExtensions(foodresources.Name, foodResourcesFitFunc, "Filter", "PreFilter"),
//			want:           framework.NewPostFilterResultWithNominatedFood("food1"),
//			expectedDguests:   []string{"p1.1", "p1.2"},
//		},
//		{
//			name: "One scheduler extender allows only food1, but it is not interested in given dguest, otherwise food1 would have been chosen",
//			dguest:  st.MakeDguest().Name("p").UID("p").Namespace(v1.NamespaceDefault).Priority(highPriority).Req(veryLargeRes).PreemptionPolicy(v1.PreemptLowerPriority).Obj(),
//			dguests: []*v1alpha1.Dguest{
//				st.MakeDguest().Name("p1.1").UID("p1.1").Namespace(v1.NamespaceDefault).Food("food1").Priority(midPriority).Req(smallRes).Obj(),
//				st.MakeDguest().Name("p1.2").UID("p1.2").Namespace(v1.NamespaceDefault).Food("food1").Priority(lowPriority).Req(smallRes).Obj(),
//				st.MakeDguest().Name("p2.1").UID("p2.1").Namespace(v1.NamespaceDefault).Food("food2").Priority(midPriority).Req(largeRes).Obj(),
//			},
//			foodNames: []string{"food1", "food2"},
//			extenders: []*st.FakeExtender{
//				{
//					ExtenderName: "FakeExtender1",
//					Predicates:   []st.FitPredicate{st.Food1PredicateExtender},
//					UnInterested: true,
//				},
//				{
//					ExtenderName: "FakeExtender2",
//					Predicates:   []st.FitPredicate{st.TruePredicateExtender},
//				},
//			},
//			registerPlugin: st.RegisterPluginAsExtensions(foodresources.Name, foodResourcesFitFunc, "Filter", "PreFilter"),
//			// sum of priorities of all victims on food1 is larger than food2, food2 is chosen.
//			want:         framework.NewPostFilterResultWithNominatedFood("food2"),
//			expectedDguests: []string{"p2.1"},
//		},
//		{
//			name: "no preempting in dguest",
//			dguest:  st.MakeDguest().Name("p").UID("p").Namespace(v1.NamespaceDefault).Priority(highPriority).Req(veryLargeRes).PreemptionPolicy(v1.PreemptNever).Obj(),
//			dguests: []*v1alpha1.Dguest{
//				st.MakeDguest().Name("p1.1").UID("p1.1").Namespace(v1.NamespaceDefault).Food("food1").Priority(lowPriority).Req(smallRes).Obj(),
//				st.MakeDguest().Name("p1.2").UID("p1.2").Namespace(v1.NamespaceDefault).Food("food1").Priority(lowPriority).Req(smallRes).Obj(),
//				st.MakeDguest().Name("p2.1").UID("p2.1").Namespace(v1.NamespaceDefault).Food("food2").Priority(highPriority).Req(largeRes).Obj(),
//				st.MakeDguest().Name("p3.1").UID("p3.1").Namespace(v1.NamespaceDefault).Food("food3").Priority(midPriority).Req(mediumRes).Obj(),
//			},
//			foodNames:      []string{"food1", "food2", "food3"},
//			registerPlugin: st.RegisterPluginAsExtensions(foodresources.Name, foodResourcesFitFunc, "Filter", "PreFilter"),
//			want:           nil,
//			expectedDguests:   nil,
//		},
//		{
//			name: "PreemptionPolicy is nil",
//			dguest:  st.MakeDguest().Name("p").UID("p").Namespace(v1.NamespaceDefault).Priority(highPriority).Req(veryLargeRes).Obj(),
//			dguests: []*v1alpha1.Dguest{
//				st.MakeDguest().Name("p1.1").UID("p1.1").Namespace(v1.NamespaceDefault).Food("food1").Priority(lowPriority).Req(smallRes).Obj(),
//				st.MakeDguest().Name("p1.2").UID("p1.2").Namespace(v1.NamespaceDefault).Food("food1").Priority(lowPriority).Req(smallRes).Obj(),
//				st.MakeDguest().Name("p2.1").UID("p2.1").Namespace(v1.NamespaceDefault).Food("food2").Priority(highPriority).Req(largeRes).Obj(),
//				st.MakeDguest().Name("p3.1").UID("p3.1").Namespace(v1.NamespaceDefault).Food("food3").Priority(midPriority).Req(mediumRes).Obj(),
//			},
//			foodNames:      []string{"food1", "food2", "food3"},
//			registerPlugin: st.RegisterPluginAsExtensions(foodresources.Name, foodResourcesFitFunc, "Filter", "PreFilter"),
//			want:           framework.NewPostFilterResultWithNominatedFood("food1"),
//			expectedDguests:   []string{"p1.1", "p1.2"},
//		},
//	}
//
//	labelKeys := []string{"hostname", "zone", "region"}
//	for _, test := range tests {
//		t.Run(test.name, func(t *testing.T) {
//			client := clientsetfake.NewSimpleClientset()
//			informerFactory := informers.NewSharedInformerFactory(client, 0)
//			dguestInformer := informerFactory.Core().V1().Dguests().Informer()
//			dguestInformer.GetStore().Add(test.dguest)
//			for i := range test.dguests {
//				dguestInformer.GetStore().Add(test.dguests[i])
//			}
//
//			deletedDguestNames := make(sets.String)
//			client.PrependReactor("delete", "dguests", func(action clienttesting.Action) (bool, runtime.Object, error) {
//				deletedDguestNames.Insert(action.(clienttesting.DeleteAction).GetName())
//				return true, nil, nil
//			})
//
//			ctx, cancel := context.WithCancel(context.Background())
//			defer cancel()
//
//			cache := internalcache.New(time.Duration(0), ctx.Done())
//			for _, dguest := range test.dguests {
//				cache.AddDguest(dguest)
//			}
//			cachedFoodInfoMap := map[string]*framework.FoodInfo{}
//			foods := make([]*v1alpha1.Food, len(test.foodNames))
//			for i, name := range test.foodNames {
//				food := st.MakeFood().Name(name).Capacity(veryLargeRes).Obj()
//				// Split food name by '/' to form labels in a format of
//				// {"hostname": food.Name[0], "zone": food.Name[1], "region": food.Name[2]}
//				food.ObjectMeta.Labels = make(map[string]string)
//				for i, label := range strings.Split(food.Name, "/") {
//					food.ObjectMeta.Labels[labelKeys[i]] = label
//				}
//				food.Name = food.ObjectMeta.Labels["hostname"]
//				cache.AddFood(food)
//				foods[i] = food
//
//				// Set foodInfo to extenders to mock extenders' cache for preemption.
//				cachedFoodInfo := framework.NewFoodInfo()
//				cachedFoodInfo.SetFood(food)
//				cachedFoodInfoMap[food.Name] = cachedFoodInfo
//			}
//			var extenders []framework.Extender
//			for _, extender := range test.extenders {
//				// Set foodInfoMap as extenders cached food information.
//				extender.CachedFoodNameToInfo = cachedFoodInfoMap
//				extenders = append(extenders, extender)
//			}
//			fwk, err := st.NewFramework(
//				[]st.RegisterPluginFunc{
//					test.registerPlugin,
//					st.RegisterQueueSortPlugin(queuesort.Name, queuesort.New),
//					st.RegisterBindPlugin(defaultbinder.Name, defaultbinder.New),
//				},
//				"",
//				ctx.Done(),
//				frameworkruntime.WithClientSet(client),
//				frameworkruntime.WithEventRecorder(&events.FakeRecorder{}),
//				frameworkruntime.WithExtenders(extenders),
//				frameworkruntime.WithDguestNominator(internalqueue.NewDguestNominator(informerFactory.Core().V1().Dguests().Lister())),
//				frameworkruntime.WithSnapshotSharedLister(internalcache.NewSnapshot(test.dguests, foods)),
//				frameworkruntime.WithInformerFactory(informerFactory),
//			)
//			if err != nil {
//				t.Fatal(err)
//			}
//
//			state := framework.NewCycleState()
//			// Some tests rely on PreFilter plugin to compute its CycleState.
//			if _, s := fwk.RunPreFilterPlugins(ctx, state, test.dguest); !s.IsSuccess() {
//				t.Errorf("Unexpected preFilterStatus: %v", s)
//			}
//			// Call preempt and check the expected results.
//			pl := DefaultPreemption{
//				fh:        fwk,
//				dguestLister: informerFactory.Core().V1().Dguests().Lister(),
//				pdbLister: getPDBLister(informerFactory),
//				args:      *getDefaultDefaultPreemptionArgs(),
//			}
//
//			pe := preemption.Evaluator{
//				PluginName: names.DefaultPreemption,
//				Handler:    pl.fh,
//				DguestLister:  pl.dguestLister,
//				PdbLister:  pl.pdbLister,
//				State:      state,
//				Interface:  &pl,
//			}
//			res, status := pe.Preempt(ctx, test.dguest, make(framework.FoodToStatusMap))
//			if !status.IsSuccess() && !status.IsUnschedulable() {
//				t.Errorf("unexpected error in preemption: %v", status.AsError())
//			}
//			if diff := cmp.Diff(test.want, res); diff != "" {
//				t.Errorf("Unexpected status (-want, +got):\n%s", diff)
//			}
//			if len(deletedDguestNames) != len(test.expectedDguests) {
//				t.Errorf("expected %v dguests, got %v.", len(test.expectedDguests), len(deletedDguestNames))
//			}
//			for victimName := range deletedDguestNames {
//				found := false
//				for _, expDguest := range test.expectedDguests {
//					if expDguest == victimName {
//						found = true
//						break
//					}
//				}
//				if !found {
//					t.Errorf("dguest %v is not expected to be a victim.", victimName)
//				}
//			}
//			if res != nil && res.NominatingInfo != nil {
//				test.dguest.Status.NominatedFoodName = res.NominatedFoodName
//			}
//
//			// Manually set the deleted Dguests' deletionTimestamp to non-nil.
//			for _, dguest := range test.dguests {
//				if deletedDguestNames.Has(dguest.Name) {
//					now := metav1.Now()
//					dguest.DeletionTimestamp = &now
//					deletedDguestNames.Delete(dguest.Name)
//				}
//			}
//
//			// Call preempt again and make sure it doesn't preempt any more dguests.
//			res, status = pe.Preempt(ctx, test.dguest, make(framework.FoodToStatusMap))
//			if !status.IsSuccess() && !status.IsUnschedulable() {
//				t.Errorf("unexpected error in preemption: %v", status.AsError())
//			}
//			if res != nil && res.NominatingInfo != nil && len(deletedDguestNames) > 0 {
//				t.Errorf("didn't expect any more preemption. Food %v is selected for preemption.", res.NominatedFoodName)
//			}
//		})
//	}
//}
