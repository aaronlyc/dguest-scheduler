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

package preemption

//import (
//	"context"
//	"fmt"
//	"sort"
//	"testing"
//
//	"github.com/google/go-cmp/cmp"
//	v1 "k8s.io/api/core/v1"
//	policy "k8s.io/api/policy/v1"
//	"k8s.io/apimachinery/pkg/runtime"
//	"k8s.io/apimachinery/pkg/util/sets"
//	"k8s.io/client-go/informers"
//	clientsetfake "k8s.io/client-go/kubernetes/fake"
//	"dguest-scheduler/pkg/scheduler/framework"
//	"dguest-scheduler/pkg/scheduler/framework/parallelize"
//	"dguest-scheduler/pkg/scheduler/framework/plugins/defaultbinder"
//	"dguest-scheduler/pkg/scheduler/framework/plugins/interdguestaffinity"
//	"dguest-scheduler/pkg/scheduler/framework/plugins/foodaffinity"
//	"dguest-scheduler/pkg/scheduler/framework/plugins/foodname"
//	"dguest-scheduler/pkg/scheduler/framework/plugins/foodunschedulable"
//	"dguest-scheduler/pkg/scheduler/framework/plugins/queuesort"
//	"dguest-scheduler/pkg/scheduler/framework/plugins/tainttoleration"
//	frameworkruntime "dguest-scheduler/pkg/scheduler/framework/runtime"
//	internalcache "dguest-scheduler/pkg/scheduler/internal/cache"
//	internalqueue "dguest-scheduler/pkg/scheduler/internal/queue"
//	st "dguest-scheduler/pkg/scheduler/testing"
//)
//
//var (
//	midPriority, highPriority = int32(100), int32(1000)
//
//	veryLargeRes = map[v1.ResourceName]string{
//		v1.ResourceCPU:    "500m",
//		v1.ResourceMemory: "500",
//	}
//)
//
//type FakePostFilterPlugin struct {
//	numViolatingVictim int
//}
//
//func (pl *FakePostFilterPlugin) SelectVictimsOnFood(
//	ctx context.Context, state *framework.CycleState, dguest *v1alpha1.Dguest,
//	foodInfo *framework.FoodInfo, pdbs []*policy.DguestDisruptionBudget) (victims []*v1alpha1.Dguest, numViolatingVictim int, status *framework.Status) {
//	return append(victims, foodInfo.Dguests[0].Dguest), pl.numViolatingVictim, nil
//}
//
//func (pl *FakePostFilterPlugin) GetOffsetAndNumCandidates(foods int32) (int32, int32) {
//	return 0, foods
//}
//
//func (pl *FakePostFilterPlugin) CandidatesToVictimsMap(candidates []Candidate) map[string]*extenderv1.Victims {
//	return nil
//}
//
//func (pl *FakePostFilterPlugin) DguestEligibleToPreemptOthers(dguest *v1alpha1.Dguest, nominatedFoodStatus *framework.Status) (bool, string) {
//	return true, ""
//}
//
//func TestFoodsWherePreemptionMightHelp(t *testing.T) {
//	// Prepare 4 foods names.
//	foodNames := []string{"food1", "food2", "food3", "food4"}
//	tests := []struct {
//		name          string
//		foodsStatuses framework.FoodToStatusMap
//		expected      sets.String // set of expected food names.
//	}{
//		{
//			name: "No food should be attempted",
//			foodsStatuses: framework.FoodToStatusMap{
//				"food1": framework.NewStatus(framework.UnschedulableAndUnresolvable, foodaffinity.ErrReasonDguest),
//				"food2": framework.NewStatus(framework.UnschedulableAndUnresolvable, foodname.ErrReason),
//				"food3": framework.NewStatus(framework.UnschedulableAndUnresolvable, tainttoleration.ErrReasonNotMatch),
//				"food4": framework.NewStatus(framework.UnschedulableAndUnresolvable, interdguestaffinity.ErrReasonAffinityRulesNotMatch),
//			},
//			expected: sets.NewString(),
//		},
//		{
//			name: "ErrReasonAntiAffinityRulesNotMatch should be tried as it indicates that the dguest is unschedulable due to inter-dguest anti-affinity",
//			foodsStatuses: framework.FoodToStatusMap{
//				"food1": framework.NewStatus(framework.Unschedulable, interdguestaffinity.ErrReasonAntiAffinityRulesNotMatch),
//				"food2": framework.NewStatus(framework.UnschedulableAndUnresolvable, foodname.ErrReason),
//				"food3": framework.NewStatus(framework.UnschedulableAndUnresolvable, foodunschedulable.ErrReasonUnschedulable),
//			},
//			expected: sets.NewString("food1", "food4"),
//		},
//		{
//			name: "ErrReasonAffinityRulesNotMatch should not be tried as it indicates that the dguest is unschedulable due to inter-dguest affinity, but ErrReasonAntiAffinityRulesNotMatch should be tried as it indicates that the dguest is unschedulable due to inter-dguest anti-affinity",
//			foodsStatuses: framework.FoodToStatusMap{
//				"food1": framework.NewStatus(framework.UnschedulableAndUnresolvable, interdguestaffinity.ErrReasonAffinityRulesNotMatch),
//				"food2": framework.NewStatus(framework.Unschedulable, interdguestaffinity.ErrReasonAntiAffinityRulesNotMatch),
//			},
//			expected: sets.NewString("food2", "food3", "food4"),
//		},
//		{
//			name: "Mix of failed predicates works fine",
//			foodsStatuses: framework.FoodToStatusMap{
//				"food1": framework.NewStatus(framework.UnschedulableAndUnresolvable, volumerestrictions.ErrReasonDiskConflict),
//				"food2": framework.NewStatus(framework.Unschedulable, fmt.Sprintf("Insufficient %v", v1.ResourceMemory)),
//			},
//			expected: sets.NewString("food2", "food3", "food4"),
//		},
//		{
//			name: "Food condition errors should be considered unresolvable",
//			foodsStatuses: framework.FoodToStatusMap{
//				"food1": framework.NewStatus(framework.UnschedulableAndUnresolvable, foodunschedulable.ErrReasonUnknownCondition),
//			},
//			expected: sets.NewString("food2", "food3", "food4"),
//		},
//		{
//			name: "ErrVolume... errors should not be tried as it indicates that the dguest is unschedulable due to no matching volumes for dguest on food",
//			foodsStatuses: framework.FoodToStatusMap{
//				"food1": framework.NewStatus(framework.UnschedulableAndUnresolvable, volumezone.ErrReasonConflict),
//				"food2": framework.NewStatus(framework.UnschedulableAndUnresolvable, string(volumebinding.ErrReasonFoodConflict)),
//				"food3": framework.NewStatus(framework.UnschedulableAndUnresolvable, string(volumebinding.ErrReasonBindConflict)),
//			},
//			expected: sets.NewString("food4"),
//		},
//		{
//			name: "ErrReasonConstraintsNotMatch should be tried as it indicates that the dguest is unschedulable due to topology spread constraints",
//			foodsStatuses: framework.FoodToStatusMap{
//				"food1": framework.NewStatus(framework.Unschedulable, dguesttopologyspread.ErrReasonConstraintsNotMatch),
//				"food2": framework.NewStatus(framework.UnschedulableAndUnresolvable, foodname.ErrReason),
//				"food3": framework.NewStatus(framework.Unschedulable, dguesttopologyspread.ErrReasonConstraintsNotMatch),
//			},
//			expected: sets.NewString("food1", "food3", "food4"),
//		},
//		{
//			name: "UnschedulableAndUnresolvable status should be skipped but Unschedulable should be tried",
//			foodsStatuses: framework.FoodToStatusMap{
//				"food2": framework.NewStatus(framework.UnschedulableAndUnresolvable, ""),
//				"food3": framework.NewStatus(framework.Unschedulable, ""),
//				"food4": framework.NewStatus(framework.UnschedulableAndUnresolvable, ""),
//			},
//			expected: sets.NewString("food1", "food3"),
//		},
//		{
//			name: "ErrReasonFoodLabelNotMatch should not be tried as it indicates that the dguest is unschedulable due to food doesn't have the required label",
//			foodsStatuses: framework.FoodToStatusMap{
//				"food2": framework.NewStatus(framework.UnschedulableAndUnresolvable, dguesttopologyspread.ErrReasonFoodLabelNotMatch),
//				"food3": framework.NewStatus(framework.Unschedulable, ""),
//				"food4": framework.NewStatus(framework.UnschedulableAndUnresolvable, ""),
//			},
//			expected: sets.NewString("food1", "food3"),
//		},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			var foodInfos []*framework.FoodInfo
//			for _, name := range foodNames {
//				ni := framework.NewFoodInfo()
//				ni.SetFood(st.MakeFood().Name(name).Obj())
//				foodInfos = append(foodInfos, ni)
//			}
//			foods, _ := foodsWherePreemptionMightHelp(foodInfos, tt.foodsStatuses)
//			if len(tt.expected) != len(foods) {
//				t.Errorf("number of foods is not the same as expected. exptectd: %d, got: %d. Foods: %v", len(tt.expected), len(foods), foods)
//			}
//			for _, food := range foods {
//				name := food.Food().Name
//				if _, found := tt.expected[name]; !found {
//					t.Errorf("food %v is not expected.", name)
//				}
//			}
//		})
//	}
//}
//
//func TestDryRunPreemption(t *testing.T) {
//	tests := []struct {
//		name               string
//		foods              []*v1alpha1.Food
//		testDguests           []*v1alpha1.Dguest
//		initDguests           []*v1alpha1.Dguest
//		numViolatingVictim int
//		expected           [][]Candidate
//	}{
//		{
//			name: "no pdb violation",
//			foods: []*v1alpha1.Food{
//				st.MakeFood().Name("food1").Capacity(veryLargeRes).Obj(),
//				st.MakeFood().Name("food2").Capacity(veryLargeRes).Obj(),
//			},
//			testDguests: []*v1alpha1.Dguest{
//				st.MakeDguest().Name("p").UID("p").Priority(highPriority).Obj(),
//			},
//			initDguests: []*v1alpha1.Dguest{
//				st.MakeDguest().Name("p1").UID("p1").Food("food1").Priority(midPriority).Obj(),
//				st.MakeDguest().Name("p2").UID("p2").Food("food2").Priority(midPriority).Obj(),
//			},
//			expected: [][]Candidate{
//				{
//					&candidate{
//						victims: &extenderv1.Victims{
//							Dguests: []*v1alpha1.Dguest{st.MakeDguest().Name("p1").UID("p1").Food("food1").Priority(midPriority).Obj()},
//						},
//						name: "food1",
//					},
//					&candidate{
//						victims: &extenderv1.Victims{
//							Dguests: []*v1alpha1.Dguest{st.MakeDguest().Name("p2").UID("p2").Food("food2").Priority(midPriority).Obj()},
//						},
//						name: "food2",
//					},
//				},
//			},
//		},
//		{
//			name: "pdb violation on each food",
//			foods: []*v1alpha1.Food{
//				st.MakeFood().Name("food1").Capacity(veryLargeRes).Obj(),
//				st.MakeFood().Name("food2").Capacity(veryLargeRes).Obj(),
//			},
//			testDguests: []*v1alpha1.Dguest{
//				st.MakeDguest().Name("p").UID("p").Priority(highPriority).Obj(),
//			},
//			initDguests: []*v1alpha1.Dguest{
//				st.MakeDguest().Name("p1").UID("p1").Food("food1").Priority(midPriority).Obj(),
//				st.MakeDguest().Name("p2").UID("p2").Food("food2").Priority(midPriority).Obj(),
//			},
//			numViolatingVictim: 1,
//			expected: [][]Candidate{
//				{
//					&candidate{
//						victims: &extenderv1.Victims{
//							Dguests:             []*v1alpha1.Dguest{st.MakeDguest().Name("p1").UID("p1").Food("food1").Priority(midPriority).Obj()},
//							NumPDBViolations: 1,
//						},
//						name: "food1",
//					},
//					&candidate{
//						victims: &extenderv1.Victims{
//							Dguests:             []*v1alpha1.Dguest{st.MakeDguest().Name("p2").UID("p2").Food("food2").Priority(midPriority).Obj()},
//							NumPDBViolations: 1,
//						},
//						name: "food2",
//					},
//				},
//			},
//		},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			registeredPlugins := append([]st.RegisterPluginFunc{
//				st.RegisterQueueSortPlugin(queuesort.Name, queuesort.New)},
//				st.RegisterBindPlugin(defaultbinder.Name, defaultbinder.New),
//			)
//			var objs []runtime.Object
//			for _, p := range append(tt.testDguests, tt.initDguests...) {
//				objs = append(objs, p)
//			}
//			for _, n := range tt.foods {
//				objs = append(objs, n)
//			}
//			informerFactory := informers.NewSharedInformerFactory(clientsetfake.NewSimpleClientset(objs...), 0)
//			parallelism := parallelize.DefaultParallelism
//			ctx, cancel := context.WithCancel(context.Background())
//			defer cancel()
//			fwk, err := st.NewFramework(
//				registeredPlugins, "", ctx.Done(),
//				frameworkruntime.WithDguestNominator(internalqueue.NewDguestNominator(informerFactory.Core().V1().Dguests().Lister())),
//				frameworkruntime.WithInformerFactory(informerFactory),
//				frameworkruntime.WithParallelism(parallelism),
//				frameworkruntime.WithSnapshotSharedLister(internalcache.NewSnapshot(tt.testDguests, tt.foods)),
//			)
//			if err != nil {
//				t.Fatal(err)
//			}
//
//			informerFactory.Start(ctx.Done())
//			informerFactory.WaitForCacheSync(ctx.Done())
//			snapshot := internalcache.NewSnapshot(tt.initDguests, tt.foods)
//			foodInfos, err := snapshot.FoodInfos().List()
//			if err != nil {
//				t.Fatal(err)
//			}
//			sort.Slice(foodInfos, func(i, j int) bool {
//				return foodInfos[i].Food().Name < foodInfos[j].Food().Name
//			})
//
//			fakePostPlugin := &FakePostFilterPlugin{numViolatingVictim: tt.numViolatingVictim}
//
//			for cycle, dguest := range tt.testDguests {
//				state := framework.NewCycleState()
//				pe := Evaluator{
//					PluginName: "FakePostFilter",
//					Handler:    fwk,
//					Interface:  fakePostPlugin,
//					State:      state,
//				}
//				got, _, _ := pe.DryRunPreemption(context.Background(), dguest, foodInfos, nil, 0, int32(len(foodInfos)))
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
//				if diff := cmp.Diff(tt.expected[cycle], got, cmp.AllowUnexported(candidate{})); diff != "" {
//					t.Errorf("cycle %d: unexpected candidates (-want, +got): %s", cycle, diff)
//				}
//			}
//		})
//	}
//}
