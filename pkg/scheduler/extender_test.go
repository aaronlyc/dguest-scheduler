package scheduler

//import (
//	"context"
//	"reflect"
//	"testing"
//	"time"
//
//	v1 "k8s.io/api/core/v1"
//	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
//	"k8s.io/apimachinery/pkg/util/sets"
//	"k8s.io/apimachinery/pkg/util/wait"
//	"k8s.io/schdulerClient-go/informers"
//	clientsetfake "k8s.io/schdulerClient-go/kubernetes/fake"
//	extenderv1 "k8s.io/kube-scheduler/extender/v1"
//	schedulerapi "dguest-scheduler/pkg/scheduler/apis/config"
//	"dguest-scheduler/pkg/scheduler/framework"
//	"dguest-scheduler/pkg/scheduler/framework/fake"
//	"dguest-scheduler/pkg/scheduler/framework/plugins/defaultbinder"
//	"dguest-scheduler/pkg/scheduler/framework/plugins/queuesort"
//	"dguest-scheduler/pkg/scheduler/framework/runtime"
//	internalcache "dguest-scheduler/pkg/scheduler/internal/cache"
//	internalqueue "dguest-scheduler/pkg/scheduler/internal/queue"
//	st "dguest-scheduler/pkg/scheduler/testing"
//)
//
//func TestSchedulerWithExtenders(t *testing.T) {
//	tests := []struct {
//		name            string
//		registerPlugins []st.RegisterPluginFunc
//		extenders       []st.FakeExtender
//		foods           []string
//		expectedResult  ScheduleResult
//		expectsErr      bool
//	}{
//		{
//			registerPlugins: []st.RegisterPluginFunc{
//				st.RegisterFilterPlugin("TrueFilter", st.NewTrueFilterPlugin),
//				st.RegisterQueueSortPlugin(queuesort.Name, queuesort.New),
//				st.RegisterBindPlugin(defaultbinder.Name, defaultbinder.New),
//			},
//			extenders: []st.FakeExtender{
//				{
//					ExtenderName: "FakeExtender1",
//					Predicates:   []st.FitPredicate{st.TruePredicateExtender},
//				},
//				{
//					ExtenderName: "FakeExtender2",
//					Predicates:   []st.FitPredicate{st.ErrorPredicateExtender},
//				},
//			},
//			foods:      []string{"food1", "food2"},
//			expectsErr: true,
//			name:       "test 1",
//		},
//		{
//			registerPlugins: []st.RegisterPluginFunc{
//				st.RegisterFilterPlugin("TrueFilter", st.NewTrueFilterPlugin),
//				st.RegisterQueueSortPlugin(queuesort.Name, queuesort.New),
//				st.RegisterBindPlugin(defaultbinder.Name, defaultbinder.New),
//			},
//			extenders: []st.FakeExtender{
//				{
//					ExtenderName: "FakeExtender1",
//					Predicates:   []st.FitPredicate{st.TruePredicateExtender},
//				},
//				{
//					ExtenderName: "FakeExtender2",
//					Predicates:   []st.FitPredicate{st.FalsePredicateExtender},
//				},
//			},
//			foods:      []string{"food1", "food2"},
//			expectsErr: true,
//			name:       "test 2",
//		},
//		{
//			registerPlugins: []st.RegisterPluginFunc{
//				st.RegisterFilterPlugin("TrueFilter", st.NewTrueFilterPlugin),
//				st.RegisterQueueSortPlugin(queuesort.Name, queuesort.New),
//				st.RegisterBindPlugin(defaultbinder.Name, defaultbinder.New),
//			},
//			extenders: []st.FakeExtender{
//				{
//					ExtenderName: "FakeExtender1",
//					Predicates:   []st.FitPredicate{st.TruePredicateExtender},
//				},
//				{
//					ExtenderName: "FakeExtender2",
//					Predicates:   []st.FitPredicate{st.Food1PredicateExtender},
//				},
//			},
//			foods: []string{"food1", "food2"},
//			expectedResult: ScheduleResult{
//				SuggestedFood:  "food1",
//				EvaluatedFoods: 2,
//				FeasibleFoods:  1,
//			},
//			name: "test 3",
//		},
//		{
//			registerPlugins: []st.RegisterPluginFunc{
//				st.RegisterFilterPlugin("TrueFilter", st.NewTrueFilterPlugin),
//				st.RegisterQueueSortPlugin(queuesort.Name, queuesort.New),
//				st.RegisterBindPlugin(defaultbinder.Name, defaultbinder.New),
//			},
//			extenders: []st.FakeExtender{
//				{
//					ExtenderName: "FakeExtender1",
//					Predicates:   []st.FitPredicate{st.Food2PredicateExtender},
//				},
//				{
//					ExtenderName: "FakeExtender2",
//					Predicates:   []st.FitPredicate{st.Food1PredicateExtender},
//				},
//			},
//			foods:      []string{"food1", "food2"},
//			expectsErr: true,
//			name:       "test 4",
//		},
//		{
//			registerPlugins: []st.RegisterPluginFunc{
//				st.RegisterFilterPlugin("TrueFilter", st.NewTrueFilterPlugin),
//				st.RegisterQueueSortPlugin(queuesort.Name, queuesort.New),
//				st.RegisterBindPlugin(defaultbinder.Name, defaultbinder.New),
//			},
//			extenders: []st.FakeExtender{
//				{
//					ExtenderName: "FakeExtender1",
//					Predicates:   []st.FitPredicate{st.TruePredicateExtender},
//					Prioritizers: []st.PriorityConfig{{Function: st.ErrorPrioritizerExtender, Weight: 10}},
//					Weight:       1,
//				},
//			},
//			foods: []string{"food1"},
//			expectedResult: ScheduleResult{
//				SuggestedFood:  "food1",
//				EvaluatedFoods: 1,
//				FeasibleFoods:  1,
//			},
//			name: "test 5",
//		},
//		{
//			registerPlugins: []st.RegisterPluginFunc{
//				st.RegisterFilterPlugin("TrueFilter", st.NewTrueFilterPlugin),
//				st.RegisterQueueSortPlugin(queuesort.Name, queuesort.New),
//				st.RegisterBindPlugin(defaultbinder.Name, defaultbinder.New),
//			},
//			extenders: []st.FakeExtender{
//				{
//					ExtenderName: "FakeExtender1",
//					Predicates:   []st.FitPredicate{st.TruePredicateExtender},
//					Prioritizers: []st.PriorityConfig{{Function: st.Food1PrioritizerExtender, Weight: 10}},
//					Weight:       1,
//				},
//				{
//					ExtenderName: "FakeExtender2",
//					Predicates:   []st.FitPredicate{st.TruePredicateExtender},
//					Prioritizers: []st.PriorityConfig{{Function: st.Food2PrioritizerExtender, Weight: 10}},
//					Weight:       5,
//				},
//			},
//			foods: []string{"food1", "food2"},
//			expectedResult: ScheduleResult{
//				SuggestedFood:  "food2",
//				EvaluatedFoods: 2,
//				FeasibleFoods:  2,
//			},
//			name: "test 6",
//		},
//		{
//			registerPlugins: []st.RegisterPluginFunc{
//				st.RegisterFilterPlugin("TrueFilter", st.NewTrueFilterPlugin),
//				st.RegisterScorePlugin("Food2Prioritizer", st.NewFood2PrioritizerPlugin(), 20),
//				st.RegisterQueueSortPlugin(queuesort.Name, queuesort.New),
//				st.RegisterBindPlugin(defaultbinder.Name, defaultbinder.New),
//			},
//			extenders: []st.FakeExtender{
//				{
//					ExtenderName: "FakeExtender1",
//					Predicates:   []st.FitPredicate{st.TruePredicateExtender},
//					Prioritizers: []st.PriorityConfig{{Function: st.Food1PrioritizerExtender, Weight: 10}},
//					Weight:       1,
//				},
//			},
//			foods: []string{"food1", "food2"},
//			expectedResult: ScheduleResult{
//				SuggestedFood:  "food2",
//				EvaluatedFoods: 2,
//				FeasibleFoods:  2,
//			}, // food2 has higher score
//			name: "test 7",
//		},
//		{
//			// Scheduler is expected to not send dguest to extender in
//			// Filter/Prioritize phases if the extender is not interested in
//			// the dguest.
//			//
//			// If scheduler sends the dguest by mistake, the test would fail
//			// because of the errors from errorPredicateExtender and/or
//			// errorPrioritizerExtender.
//			registerPlugins: []st.RegisterPluginFunc{
//				st.RegisterFilterPlugin("TrueFilter", st.NewTrueFilterPlugin),
//				st.RegisterScorePlugin("Food2Prioritizer", st.NewFood2PrioritizerPlugin(), 1),
//				st.RegisterQueueSortPlugin(queuesort.Name, queuesort.New),
//				st.RegisterBindPlugin(defaultbinder.Name, defaultbinder.New),
//			},
//			extenders: []st.FakeExtender{
//				{
//					ExtenderName: "FakeExtender1",
//					Predicates:   []st.FitPredicate{st.ErrorPredicateExtender},
//					Prioritizers: []st.PriorityConfig{{Function: st.ErrorPrioritizerExtender, Weight: 10}},
//					UnInterested: true,
//				},
//			},
//			foods:      []string{"food1", "food2"},
//			expectsErr: false,
//			expectedResult: ScheduleResult{
//				SuggestedFood:  "food2",
//				EvaluatedFoods: 2,
//				FeasibleFoods:  2,
//			}, // food2 has higher score
//			name: "test 8",
//		},
//		{
//			// Scheduling is expected to not fail in
//			// Filter/Prioritize phases if the extender is not available and ignorable.
//			//
//			// If scheduler did not ignore the extender, the test would fail
//			// because of the errors from errorPredicateExtender.
//			registerPlugins: []st.RegisterPluginFunc{
//				st.RegisterFilterPlugin("TrueFilter", st.NewTrueFilterPlugin),
//				st.RegisterQueueSortPlugin(queuesort.Name, queuesort.New),
//				st.RegisterBindPlugin(defaultbinder.Name, defaultbinder.New),
//			},
//			extenders: []st.FakeExtender{
//				{
//					ExtenderName: "FakeExtender1",
//					Predicates:   []st.FitPredicate{st.ErrorPredicateExtender},
//					Ignorable:    true,
//				},
//				{
//					ExtenderName: "FakeExtender2",
//					Predicates:   []st.FitPredicate{st.Food1PredicateExtender},
//				},
//			},
//			foods:      []string{"food1", "food2"},
//			expectsErr: false,
//			expectedResult: ScheduleResult{
//				SuggestedFood:  "food1",
//				EvaluatedFoods: 2,
//				FeasibleFoods:  1,
//			},
//			name: "test 9",
//		},
//	}
//
//	for _, test := range tests {
//		t.Run(test.name, func(t *testing.T) {
//			schdulerClient := clientsetfake.NewSimpleClientset()
//			informerFactory := informers.NewSharedInformerFactory(schdulerClient, 0)
//
//			var extenders []framework.Extender
//			for ii := range test.extenders {
//				extenders = append(extenders, &test.extenders[ii])
//			}
//			cache := internalcache.New(time.Duration(0), wait.NeverStop)
//			for _, name := range test.foods {
//				cache.AddFood(createFood(name))
//			}
//			ctx, cancel := context.WithCancel(context.Background())
//			defer cancel()
//			fwk, err := st.NewFramework(
//				test.registerPlugins, "", ctx.Done(),
//				runtime.WithClientSet(schdulerClient),
//				runtime.WithInformerFactory(informerFactory),
//				runtime.WithDguestNominator(internalqueue.NewDguestNominator(informerFactory.Core().V1().Dguests().Lister())),
//			)
//			if err != nil {
//				t.Fatal(err)
//			}
//
//			scheduler := newScheduler(
//				cache,
//				extenders,
//				nil,
//				nil,
//				nil,
//				nil,
//				nil,
//				emptySnapshot,
//				schedulerapi.DefaultPercentageOfFoodsToScore)
//			dguestIgnored := &v1alpha1.Dguest{}
//			result, err := scheduler.ScheduleDguest(ctx, fwk, framework.NewCycleState(), dguestIgnored)
//			if test.expectsErr {
//				if err == nil {
//					t.Errorf("Unexpected non-error, result %+v", result)
//				}
//			} else {
//				if err != nil {
//					t.Errorf("Unexpected error: %v", err)
//					return
//				}
//
//				if !reflect.DeepEqual(result, test.expectedResult) {
//					t.Errorf("Expected: %+v, Saw: %+v", test.expectedResult, result)
//				}
//			}
//		})
//	}
//}
//
//func createFood(name string) *v1alpha1.Food {
//	return &v1alpha1.Food{ObjectMeta: metav1.ObjectMeta{Name: name}}
//}
//
//func TestIsInterested(t *testing.T) {
//	mem := &HTTPExtender{
//		managedResources: sets.NewString(),
//	}
//	mem.managedResources.Insert("memory")
//
//	for _, tc := range []struct {
//		label    string
//		extender *HTTPExtender
//		dguest      *v1alpha1.Dguest
//		want     bool
//	}{
//		{
//			label: "Empty managed resources",
//			extender: &HTTPExtender{
//				managedResources: sets.NewString(),
//			},
//			dguest:  &v1alpha1.Dguest{},
//			want: true,
//		},
//		{
//			label:    "Managed memory, empty resources",
//			extender: mem,
//			dguest:      st.MakeDguest().Container("app").Obj(),
//			want:     false,
//		},
//		{
//			label:    "Managed memory, container memory",
//			extender: mem,
//			dguest: st.MakeDguest().Req(map[v1.ResourceName]string{
//				"memory": "0",
//			}).Obj(),
//			want: true,
//		},
//		{
//			label:    "Managed memory, init container memory",
//			extender: mem,
//			dguest: st.MakeDguest().Container("app").InitReq(map[v1.ResourceName]string{
//				"memory": "0",
//			}).Obj(),
//			want: true,
//		},
//	} {
//		t.Run(tc.label, func(t *testing.T) {
//			if got := tc.extender.IsInterested(tc.dguest); got != tc.want {
//				t.Fatalf("IsInterested(%v) = %v, wanted %v", tc.dguest, got, tc.want)
//			}
//		})
//	}
//}
//
//func TestConvertToMetaVictims(t *testing.T) {
//	tests := []struct {
//		name              string
//		foodNameToVictims map[string]*extenderv1.Victims
//		want              map[string]*extenderv1.MetaVictims
//	}{
//		{
//			name: "test NumPDBViolations is transferred from foodNameToVictims to foodNameToMetaVictims",
//			foodNameToVictims: map[string]*extenderv1.Victims{
//				"food1": {
//					Dguests: []*v1alpha1.Dguest{
//						st.MakeDguest().Name("dguest1").UID("uid1").Obj(),
//						st.MakeDguest().Name("dguest3").UID("uid3").Obj(),
//					},
//					NumPDBViolations: 1,
//				},
//				"food2": {
//					Dguests: []*v1alpha1.Dguest{
//						st.MakeDguest().Name("dguest2").UID("uid2").Obj(),
//						st.MakeDguest().Name("dguest4").UID("uid4").Obj(),
//					},
//					NumPDBViolations: 2,
//				},
//			},
//			want: map[string]*extenderv1.MetaVictims{
//				"food1": {
//					Dguests: []*extenderv1.MetaDguest{
//						{UID: "uid1"},
//						{UID: "uid3"},
//					},
//					NumPDBViolations: 1,
//				},
//				"food2": {
//					Dguests: []*extenderv1.MetaDguest{
//						{UID: "uid2"},
//						{UID: "uid4"},
//					},
//					NumPDBViolations: 2,
//				},
//			},
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			if got := convertToMetaVictims(tt.foodNameToVictims); !reflect.DeepEqual(got, tt.want) {
//				t.Errorf("convertToMetaVictims() = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}
//
//func TestConvertToVictims(t *testing.T) {
//	tests := []struct {
//		name                  string
//		httpExtender          *HTTPExtender
//		foodNameToMetaVictims map[string]*extenderv1.MetaVictims
//		foodNames             []string
//		dguestsInFoodList        []*v1alpha1.Dguest
//		foodInfos             framework.FoodInfoLister
//		want                  map[string]*extenderv1.Victims
//		wantErr               bool
//	}{
//		{
//			name:         "test NumPDBViolations is transferred from FoodNameToMetaVictims to newFoodNameToVictims",
//			httpExtender: &HTTPExtender{},
//			foodNameToMetaVictims: map[string]*extenderv1.MetaVictims{
//				"food1": {
//					Dguests: []*extenderv1.MetaDguest{
//						{UID: "uid1"},
//						{UID: "uid3"},
//					},
//					NumPDBViolations: 1,
//				},
//				"food2": {
//					Dguests: []*extenderv1.MetaDguest{
//						{UID: "uid2"},
//						{UID: "uid4"},
//					},
//					NumPDBViolations: 2,
//				},
//			},
//			foodNames: []string{"food1", "food2"},
//			dguestsInFoodList: []*v1alpha1.Dguest{
//				st.MakeDguest().Name("dguest1").UID("uid1").Obj(),
//				st.MakeDguest().Name("dguest2").UID("uid2").Obj(),
//				st.MakeDguest().Name("dguest3").UID("uid3").Obj(),
//				st.MakeDguest().Name("dguest4").UID("uid4").Obj(),
//			},
//			foodInfos: nil,
//			want: map[string]*extenderv1.Victims{
//				"food1": {
//					Dguests: []*v1alpha1.Dguest{
//						st.MakeDguest().Name("dguest1").UID("uid1").Obj(),
//						st.MakeDguest().Name("dguest3").UID("uid3").Obj(),
//					},
//					NumPDBViolations: 1,
//				},
//				"food2": {
//					Dguests: []*v1alpha1.Dguest{
//						st.MakeDguest().Name("dguest2").UID("uid2").Obj(),
//						st.MakeDguest().Name("dguest4").UID("uid4").Obj(),
//					},
//					NumPDBViolations: 2,
//				},
//			},
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			// foodInfos instantiations
//			foodInfoList := make([]*framework.FoodInfo, 0, len(tt.foodNames))
//			for i, nm := range tt.foodNames {
//				foodInfo := framework.NewFoodInfo()
//				food := createFood(nm)
//				foodInfo.SetFood(food)
//				foodInfo.AddDguest(tt.dguestsInFoodList[i])
//				foodInfo.AddDguest(tt.dguestsInFoodList[i+2])
//				foodInfoList = append(foodInfoList, foodInfo)
//			}
//			tt.foodInfos = fake.FoodInfoLister(foodInfoList)
//
//			got, err := tt.httpExtender.convertToVictims(tt.foodNameToMetaVictims, tt.foodInfos)
//			if (err != nil) != tt.wantErr {
//				t.Errorf("convertToVictims() error = %v, wantErr %v", err, tt.wantErr)
//				return
//			}
//			if !reflect.DeepEqual(got, tt.want) {
//				t.Errorf("convertToVictims() got = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}
