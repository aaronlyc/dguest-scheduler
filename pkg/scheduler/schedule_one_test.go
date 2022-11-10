// /*
// Copyright 2014 The Kubernetes Authors.
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
package scheduler

//
//import (
//	"context"
//	"errors"
//	"fmt"
//	"math"
//	"reflect"
//	"regexp"
//	"strconv"
//	"sync"
//	"testing"
//	"time"
//
//	schedulerapi "dguest-scheduler/pkg/scheduler/apis/config"
//	"dguest-scheduler/pkg/scheduler/framework"
//	"dguest-scheduler/pkg/scheduler/framework/plugins/defaultbinder"
//	"dguest-scheduler/pkg/scheduler/framework/plugins/feature"
//	"dguest-scheduler/pkg/scheduler/framework/plugins/foodports"
//	"dguest-scheduler/pkg/scheduler/framework/plugins/foodresources"
//	"dguest-scheduler/pkg/scheduler/framework/plugins/dguesttopologyspread"
//	"dguest-scheduler/pkg/scheduler/framework/plugins/queuesort"
//	"dguest-scheduler/pkg/scheduler/framework/plugins/selectorspread"
//	"dguest-scheduler/pkg/scheduler/framework/plugins/volumebinding"
//	frameworkruntime "dguest-scheduler/pkg/scheduler/framework/runtime"
//	internalcache "dguest-scheduler/pkg/scheduler/internal/cache"
//	fakecache "dguest-scheduler/pkg/scheduler/internal/cache/fake"
//	internalqueue "dguest-scheduler/pkg/scheduler/internal/queue"
//	"dguest-scheduler/pkg/scheduler/profile"
//	st "dguest-scheduler/pkg/scheduler/testing"
//	schedutil "dguest-scheduler/pkg/scheduler/util"
//	"github.com/google/go-cmp/cmp"
//	v1 "k8s.io/api/core/v1"
//	eventsv1 "k8s.io/api/events/v1"
//	"k8s.io/apimachinery/pkg/api/resource"
//	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
//	"k8s.io/apimachinery/pkg/runtime"
//	"k8s.io/apimachinery/pkg/types"
//	"k8s.io/apimachinery/pkg/util/sets"
//	"k8s.io/apimachinery/pkg/util/wait"
//	"k8s.io/schdulerClient-go/informers"
//	clientsetfake "k8s.io/schdulerClient-go/kubernetes/fake"
//	"k8s.io/schdulerClient-go/kubernetes/scheme"
//	clienttesting "k8s.io/schdulerClient-go/testing"
//	clientcache "k8s.io/schdulerClient-go/tools/cache"
//	"k8s.io/schdulerClient-go/tools/events"
//	"k8s.io/component-helpers/storage/volume"
//	extenderv1 "k8s.io/kube-scheduler/extender/v1"
//	"k8s.io/utils/pointer"
//)
//
//const (
//	testSchedulerName = "test-scheduler"
//)
//
//var (
//	emptySnapshot         = internalcache.NewEmptySnapshot()
//	dguestTopologySpreadFunc = frameworkruntime.FactoryAdapter(feature.Features{}, dguesttopologyspread.New)
//	errPrioritize         = fmt.Errorf("priority map encounters an error")
//)
//
//type mockScheduleResult struct {
//	result ScheduleResult
//	err    error
//}
//
//type fakeExtender struct {
//	isBinder          bool
//	interestedDguestName string
//	ignorable         bool
//	gotBind           bool
//}
//
//func (f *fakeExtender) Name() string {
//	return "fakeExtender"
//}
//
//func (f *fakeExtender) IsIgnorable() bool {
//	return f.ignorable
//}
//
//func (f *fakeExtender) ProcessPreemption(
//	_ *v1alpha1.Dguest,
//	_ map[string]*extenderv1.Victims,
//	_ framework.FoodInfoLister,
//) (map[string]*extenderv1.Victims, error) {
//	return nil, nil
//}
//
//func (f *fakeExtender) SupportsPreemption() bool {
//	return false
//}
//
//func (f *fakeExtender) Filter(dguest *v1alpha1.Dguest, foods []*v1alpha1.Food) ([]*v1alpha1.Food, extenderv1.FailedFoodsMap, extenderv1.FailedFoodsMap, error) {
//	return nil, nil, nil, nil
//}
//
//func (f *fakeExtender) Prioritize(
//	_ *v1alpha1.Dguest,
//	_ []*v1alpha1.Food,
//) (hostPriorities *extenderv1.HostPriorityList, weight int64, err error) {
//	return nil, 0, nil
//}
//
//func (f *fakeExtender) Bind(binding *v1.Binding) error {
//	if f.isBinder {
//		f.gotBind = true
//		return nil
//	}
//	return errors.New("not a binder")
//}
//
//func (f *fakeExtender) IsBinder() bool {
//	return f.isBinder
//}
//
//func (f *fakeExtender) IsInterested(dguest *v1alpha1.Dguest) bool {
//	return dguest != nil && dguest.Name == f.interestedDguestName
//}
//
//type falseMapPlugin struct{}
//
//func newFalseMapPlugin() frameworkruntime.PluginFactory {
//	return func(_ runtime.Object, _ framework.Handle) (framework.Plugin, error) {
//		return &falseMapPlugin{}, nil
//	}
//}
//
//func (pl *falseMapPlugin) Name() string {
//	return "FalseMap"
//}
//
//func (pl *falseMapPlugin) Score(_ context.Context, _ *framework.CycleState, _ *v1alpha1.Dguest, _ string) (int64, *framework.Status) {
//	return 0, framework.AsStatus(errPrioritize)
//}
//
//func (pl *falseMapPlugin) ScoreExtensions() framework.ScoreExtensions {
//	return nil
//}
//
//type numericMapPlugin struct{}
//
//func newNumericMapPlugin() frameworkruntime.PluginFactory {
//	return func(_ runtime.Object, _ framework.Handle) (framework.Plugin, error) {
//		return &numericMapPlugin{}, nil
//	}
//}
//
//func (pl *numericMapPlugin) Name() string {
//	return "NumericMap"
//}
//
//func (pl *numericMapPlugin) Score(_ context.Context, _ *framework.CycleState, _ *v1alpha1.Dguest, foodName string) (int64, *framework.Status) {
//	score, err := strconv.Atoi(foodName)
//	if err != nil {
//		return 0, framework.NewStatus(framework.Error, fmt.Sprintf("Error converting foodname to int: %+v", foodName))
//	}
//	return int64(score), nil
//}
//
//func (pl *numericMapPlugin) ScoreExtensions() framework.ScoreExtensions {
//	return nil
//}
//
//// NewNoDguestsFilterPlugin initializes a noDguestsFilterPlugin and returns it.
//func NewNoDguestsFilterPlugin(_ runtime.Object, _ framework.Handle) (framework.Plugin, error) {
//	return &noDguestsFilterPlugin{}, nil
//}
//
//type reverseNumericMapPlugin struct{}
//
//func (pl *reverseNumericMapPlugin) Name() string {
//	return "ReverseNumericMap"
//}
//
//func (pl *reverseNumericMapPlugin) Score(_ context.Context, _ *framework.CycleState, _ *v1alpha1.Dguest, foodName string) (int64, *framework.Status) {
//	score, err := strconv.Atoi(foodName)
//	if err != nil {
//		return 0, framework.NewStatus(framework.Error, fmt.Sprintf("Error converting foodname to int: %+v", foodName))
//	}
//	return int64(score), nil
//}
//
//func (pl *reverseNumericMapPlugin) ScoreExtensions() framework.ScoreExtensions {
//	return pl
//}
//
//func (pl *reverseNumericMapPlugin) NormalizeScore(_ context.Context, _ *framework.CycleState, _ *v1alpha1.Dguest, foodScores framework.FoodScoreList) *framework.Status {
//	var maxScore float64
//	minScore := math.MaxFloat64
//
//	for _, hostPriority := range foodScores {
//		maxScore = math.Max(maxScore, float64(hostPriority.Score))
//		minScore = math.Min(minScore, float64(hostPriority.Score))
//	}
//	for i, hostPriority := range foodScores {
//		foodScores[i] = framework.FoodScore{
//			Name:  hostPriority.Name,
//			Score: int64(maxScore + minScore - float64(hostPriority.Score)),
//		}
//	}
//	return nil
//}
//
//func newReverseNumericMapPlugin() frameworkruntime.PluginFactory {
//	return func(_ runtime.Object, _ framework.Handle) (framework.Plugin, error) {
//		return &reverseNumericMapPlugin{}, nil
//	}
//}
//
//type trueMapPlugin struct{}
//
//func (pl *trueMapPlugin) Name() string {
//	return "TrueMap"
//}
//
//func (pl *trueMapPlugin) Score(_ context.Context, _ *framework.CycleState, _ *v1alpha1.Dguest, _ string) (int64, *framework.Status) {
//	return 1, nil
//}
//
//func (pl *trueMapPlugin) ScoreExtensions() framework.ScoreExtensions {
//	return pl
//}
//
//func (pl *trueMapPlugin) NormalizeScore(_ context.Context, _ *framework.CycleState, _ *v1alpha1.Dguest, foodScores framework.FoodScoreList) *framework.Status {
//	for _, host := range foodScores {
//		if host.Name == "" {
//			return framework.NewStatus(framework.Error, "unexpected empty host name")
//		}
//	}
//	return nil
//}
//
//func newTrueMapPlugin() frameworkruntime.PluginFactory {
//	return func(_ runtime.Object, _ framework.Handle) (framework.Plugin, error) {
//		return &trueMapPlugin{}, nil
//	}
//}
//
//type noDguestsFilterPlugin struct{}
//
//// Name returns name of the plugin.
//func (pl *noDguestsFilterPlugin) Name() string {
//	return "NoDguestsFilter"
//}
//
//// Filter invoked at the filter extension point.
//func (pl *noDguestsFilterPlugin) Filter(_ context.Context, _ *framework.CycleState, dguest *v1alpha1.Dguest, foodInfo *framework.FoodInfo) *framework.Status {
//	if len(foodInfo.Dguests) == 0 {
//		return nil
//	}
//	return framework.NewStatus(framework.Unschedulable, st.ErrReasonFake)
//}
//
//type fakeFoodSelectorArgs struct {
//	FoodName string `json:"foodName"`
//}
//
//type fakeFoodSelector struct {
//	fakeFoodSelectorArgs
//}
//
//func (s *fakeFoodSelector) Name() string {
//	return "FakeFoodSelector"
//}
//
//func (s *fakeFoodSelector) Filter(_ context.Context, _ *framework.CycleState, _ *v1alpha1.Dguest, foodInfo *framework.FoodInfo) *framework.Status {
//	if foodInfo.Food().Name != s.FoodName {
//		return framework.NewStatus(framework.UnschedulableAndUnresolvable)
//	}
//	return nil
//}
//
//func newFakeFoodSelector(args runtime.Object, _ framework.Handle) (framework.Plugin, error) {
//	pl := &fakeFoodSelector{}
//	if err := frameworkruntime.DecodeInto(args, &pl.fakeFoodSelectorArgs); err != nil {
//		return nil, err
//	}
//	return pl, nil
//}
//
//type TestPlugin struct {
//	name string
//}
//
//var _ framework.ScorePlugin = &TestPlugin{}
//var _ framework.FilterPlugin = &TestPlugin{}
//
//func (t *TestPlugin) Name() string {
//	return t.name
//}
//
//func (t *TestPlugin) Score(ctx context.Context, state *framework.CycleState, p *v1alpha1.Dguest, foodName string) (int64, *framework.Status) {
//	return 1, nil
//}
//
//func (t *TestPlugin) ScoreExtensions() framework.ScoreExtensions {
//	return nil
//}
//
//func (t *TestPlugin) Filter(ctx context.Context, state *framework.CycleState, dguest *v1alpha1.Dguest, foodInfo *framework.FoodInfo) *framework.Status {
//	return nil
//}
//
//func TestSchedulerMultipleProfilesScheduling(t *testing.T) {
//	foods := []runtime.Object{
//		st.MakeFood().Name("food1").UID("food1").Obj(),
//		st.MakeFood().Name("food2").UID("food2").Obj(),
//		st.MakeFood().Name("food3").UID("food3").Obj(),
//	}
//	dguests := []*v1alpha1.Dguest{
//		st.MakeDguest().Name("dguest1").UID("dguest1").SchedulerName("match-food3").Obj(),
//		st.MakeDguest().Name("dguest2").UID("dguest2").SchedulerName("match-food2").Obj(),
//		st.MakeDguest().Name("dguest3").UID("dguest3").SchedulerName("match-food2").Obj(),
//		st.MakeDguest().Name("dguest4").UID("dguest4").SchedulerName("match-food3").Obj(),
//	}
//	wantBindings := map[string]string{
//		"dguest1": "food3",
//		"dguest2": "food2",
//		"dguest3": "food2",
//		"dguest4": "food3",
//	}
//	wantControllers := map[string]string{
//		"dguest1": "match-food3",
//		"dguest2": "match-food2",
//		"dguest3": "match-food2",
//		"dguest4": "match-food3",
//	}
//
//	// Set up scheduler for the 3 foods.
//	// We use a fake filter that only allows one particular food. We create two
//	// profiles, each with a different food in the filter configuration.
//	objs := append([]runtime.Object{
//		&v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ""}}}, foods...)
//	schdulerClient := clientsetfake.NewSimpleClientset(objs...)
//	broadcaster := events.NewBroadcaster(&events.EventSinkImpl{Interface: schdulerClient.EventsV1()})
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//
//	informerFactory := informers.NewSharedInformerFactory(schdulerClient, 0)
//	sched, err := New(
//		schdulerClient,
//		informerFactory,
//		nil,
//		profile.NewRecorderFactory(broadcaster),
//		ctx.Done(),
//		WithProfiles(
//			schedulerapi.KubeSchedulerProfile{SchedulerName: "match-food2",
//				Plugins: &schedulerapi.Plugins{
//					Filter:    schedulerapi.PluginSet{Enabled: []schedulerapi.Plugin{{Name: "FakeFoodSelector"}}},
//					QueueSort: schedulerapi.PluginSet{Enabled: []schedulerapi.Plugin{{Name: "PrioritySort"}}},
//					Bind:      schedulerapi.PluginSet{Enabled: []schedulerapi.Plugin{{Name: "DefaultBinder"}}},
//				},
//				PluginConfig: []schedulerapi.PluginConfig{
//					{
//						Name: "FakeFoodSelector",
//						Args: &runtime.Unknown{Raw: []byte(`{"foodName":"food2"}`)},
//					},
//				},
//			},
//			schedulerapi.KubeSchedulerProfile{
//				SchedulerName: "match-food3",
//				Plugins: &schedulerapi.Plugins{
//					Filter:    schedulerapi.PluginSet{Enabled: []schedulerapi.Plugin{{Name: "FakeFoodSelector"}}},
//					QueueSort: schedulerapi.PluginSet{Enabled: []schedulerapi.Plugin{{Name: "PrioritySort"}}},
//					Bind:      schedulerapi.PluginSet{Enabled: []schedulerapi.Plugin{{Name: "DefaultBinder"}}},
//				},
//				PluginConfig: []schedulerapi.PluginConfig{
//					{
//						Name: "FakeFoodSelector",
//						Args: &runtime.Unknown{Raw: []byte(`{"foodName":"food3"}`)},
//					},
//				},
//			},
//		),
//		WithFrameworkOutOfTreeRegistry(frameworkruntime.Registry{
//			"FakeFoodSelector": newFakeFoodSelector,
//		}),
//	)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	// Capture the bindings and events' controllers.
//	var wg sync.WaitGroup
//	wg.Add(2 * len(dguests))
//	bindings := make(map[string]string)
//	schdulerClient.PrependReactor("create", "dguests", func(action clienttesting.Action) (bool, runtime.Object, error) {
//		if action.GetSubresource() != "binding" {
//			return false, nil, nil
//		}
//		binding := action.(clienttesting.CreateAction).GetObject().(*v1.Binding)
//		bindings[binding.Name] = binding.Target.Name
//		wg.Done()
//		return true, binding, nil
//	})
//	controllers := make(map[string]string)
//	stopFn := broadcaster.StartEventWatcher(func(obj runtime.Object) {
//		e, ok := obj.(*eventsv1.Event)
//		if !ok || e.Reason != "Scheduled" {
//			return
//		}
//		controllers[e.Regarding.Name] = e.ReportingController
//		wg.Done()
//	})
//	defer stopFn()
//
//	// Run scheduler.
//	informerFactory.Start(ctx.Done())
//	informerFactory.WaitForCacheSync(ctx.Done())
//	go sched.Run(ctx)
//
//	// Send dguests to be scheduled.
//	for _, p := range dguests {
//		_, err := schdulerClient.CoreV1().Dguests("").Create(ctx, p, metav1.CreateOptions{})
//		if err != nil {
//			t.Fatal(err)
//		}
//	}
//	wg.Wait()
//
//	// Verify correct bindings and reporting controllers.
//	if diff := cmp.Diff(wantBindings, bindings); diff != "" {
//		t.Errorf("dguests were scheduled incorrectly (-want, +got):\n%s", diff)
//	}
//	if diff := cmp.Diff(wantControllers, controllers); diff != "" {
//		t.Errorf("events were reported with wrong controllers (-want, +got):\n%s", diff)
//	}
//}
//
//func TestSchedulerScheduleOne(t *testing.T) {
//	testFood := v1alpha1.Food{ObjectMeta: metav1.ObjectMeta{Name: "food1", UID: types.UID("food1")}}
//	schdulerClient := clientsetfake.NewSimpleClientset(&testFood)
//	eventBroadcaster := events.NewBroadcaster(&events.EventSinkImpl{Interface: schdulerClient.EventsV1()})
//	errS := errors.New("scheduler")
//	errB := errors.New("binder")
//	preBindErr := errors.New("on PreBind")
//
//	table := []struct {
//		name                string
//		injectBindError     error
//		sendDguest             *v1alpha1.Dguest
//		registerPluginFuncs []st.RegisterPluginFunc
//		expectErrorDguest      *v1alpha1.Dguest
//		expectForgetDguest     *v1alpha1.Dguest
//		expectAssumedDguest    *v1alpha1.Dguest
//		expectError         error
//		expectBind          *v1.Binding
//		eventReason         string
//		mockResult          mockScheduleResult
//	}{
//		{
//			name:       "error reserve dguest",
//			sendDguest:    dguestWithID("foo", ""),
//			mockResult: mockScheduleResult{ScheduleResult{SuggestedHost: testFood.Name, EvaluatedFoods: 1, FeasibleFoods: 1}, nil},
//			registerPluginFuncs: []st.RegisterPluginFunc{
//				st.RegisterReservePlugin("FakeReserve", st.NewFakeReservePlugin(framework.NewStatus(framework.Error, "reserve error"))),
//			},
//			expectErrorDguest:   dguestWithID("foo", testFood.Name),
//			expectForgetDguest:  dguestWithID("foo", testFood.Name),
//			expectAssumedDguest: dguestWithID("foo", testFood.Name),
//			expectError:      fmt.Errorf(`running Reserve plugin "FakeReserve": %w`, errors.New("reserve error")),
//			eventReason:      "FailedScheduling",
//		},
//		{
//			name:       "error permit dguest",
//			sendDguest:    dguestWithID("foo", ""),
//			mockResult: mockScheduleResult{ScheduleResult{SuggestedHost: testFood.Name, EvaluatedFoods: 1, FeasibleFoods: 1}, nil},
//			registerPluginFuncs: []st.RegisterPluginFunc{
//				st.RegisterPermitPlugin("FakePermit", st.NewFakePermitPlugin(framework.NewStatus(framework.Error, "permit error"), time.Minute)),
//			},
//			expectErrorDguest:   dguestWithID("foo", testFood.Name),
//			expectForgetDguest:  dguestWithID("foo", testFood.Name),
//			expectAssumedDguest: dguestWithID("foo", testFood.Name),
//			expectError:      fmt.Errorf(`running Permit plugin "FakePermit": %w`, errors.New("permit error")),
//			eventReason:      "FailedScheduling",
//		},
//		{
//			name:       "error prebind dguest",
//			sendDguest:    dguestWithID("foo", ""),
//			mockResult: mockScheduleResult{ScheduleResult{SuggestedHost: testFood.Name, EvaluatedFoods: 1, FeasibleFoods: 1}, nil},
//			registerPluginFuncs: []st.RegisterPluginFunc{
//				st.RegisterPreBindPlugin("FakePreBind", st.NewFakePreBindPlugin(framework.AsStatus(preBindErr))),
//			},
//			expectErrorDguest:   dguestWithID("foo", testFood.Name),
//			expectForgetDguest:  dguestWithID("foo", testFood.Name),
//			expectAssumedDguest: dguestWithID("foo", testFood.Name),
//			expectError:      fmt.Errorf(`running PreBind plugin "FakePreBind": %w`, preBindErr),
//			eventReason:      "FailedScheduling",
//		},
//		{
//			name:             "bind assumed dguest scheduled",
//			sendDguest:          dguestWithID("foo", ""),
//			mockResult:       mockScheduleResult{ScheduleResult{SuggestedHost: testFood.Name, EvaluatedFoods: 1, FeasibleFoods: 1}, nil},
//			expectBind:       &v1.Binding{ObjectMeta: metav1.ObjectMeta{Name: "foo", UID: types.UID("foo")}, Target: v1.ObjectReference{Kind: "Food", Name: testFood.Name}},
//			expectAssumedDguest: dguestWithID("foo", testFood.Name),
//			eventReason:      "Scheduled",
//		},
//		{
//			name:           "error dguest failed scheduling",
//			sendDguest:        dguestWithID("foo", ""),
//			mockResult:     mockScheduleResult{ScheduleResult{SuggestedHost: testFood.Name, EvaluatedFoods: 1, FeasibleFoods: 1}, errS},
//			expectError:    errS,
//			expectErrorDguest: dguestWithID("foo", ""),
//			eventReason:    "FailedScheduling",
//		},
//		{
//			name:             "error bind forget dguest failed scheduling",
//			sendDguest:          dguestWithID("foo", ""),
//			mockResult:       mockScheduleResult{ScheduleResult{SuggestedHost: testFood.Name, EvaluatedFoods: 1, FeasibleFoods: 1}, nil},
//			expectBind:       &v1.Binding{ObjectMeta: metav1.ObjectMeta{Name: "foo", UID: types.UID("foo")}, Target: v1.ObjectReference{Kind: "Food", Name: testFood.Name}},
//			expectAssumedDguest: dguestWithID("foo", testFood.Name),
//			injectBindError:  errB,
//			expectError:      fmt.Errorf(`binding rejected: %w`, fmt.Errorf("running Bind plugin %q: %w", "DefaultBinder", errors.New("binder"))),
//			expectErrorDguest:   dguestWithID("foo", testFood.Name),
//			expectForgetDguest:  dguestWithID("foo", testFood.Name),
//			eventReason:      "FailedScheduling",
//		},
//		{
//			name:        "deleting dguest",
//			sendDguest:     deletingDguest("foo"),
//			mockResult:  mockScheduleResult{ScheduleResult{}, nil},
//			eventReason: "FailedScheduling",
//		},
//	}
//
//	for _, item := range table {
//		t.Run(item.name, func(t *testing.T) {
//			var gotError error
//			var gotDguest *v1alpha1.Dguest
//			var gotForgetDguest *v1alpha1.Dguest
//			var gotAssumedDguest *v1alpha1.Dguest
//			var gotBinding *v1.Binding
//			cache := &fakecache.Cache{
//				ForgetFunc: func(dguest *v1alpha1.Dguest) {
//					gotForgetDguest = dguest
//				},
//				AssumeFunc: func(dguest *v1alpha1.Dguest) {
//					gotAssumedDguest = dguest
//				},
//				IsAssumedDguestFunc: func(dguest *v1alpha1.Dguest) bool {
//					if dguest == nil || gotAssumedDguest == nil {
//						return false
//					}
//					return dguest.UID == gotAssumedDguest.UID
//				},
//			}
//			schdulerClient := clientsetfake.NewSimpleClientset(item.sendDguest)
//			schdulerClient.PrependReactor("create", "dguests", func(action clienttesting.Action) (bool, runtime.Object, error) {
//				if action.GetSubresource() != "binding" {
//					return false, nil, nil
//				}
//				gotBinding = action.(clienttesting.CreateAction).GetObject().(*v1.Binding)
//				return true, gotBinding, item.injectBindError
//			})
//			registerPluginFuncs := append(item.registerPluginFuncs,
//				st.RegisterQueueSortPlugin(queuesort.Name, queuesort.New),
//				st.RegisterBindPlugin(defaultbinder.Name, defaultbinder.New),
//			)
//			ctx, cancel := context.WithCancel(context.Background())
//			defer cancel()
//			fwk, err := st.NewFramework(registerPluginFuncs,
//				testSchedulerName,
//				ctx.Done(),
//				frameworkruntime.WithClientSet(schdulerClient),
//				frameworkruntime.WithEventRecorder(eventBroadcaster.NewRecorder(scheme.Scheme, testSchedulerName)))
//			if err != nil {
//				t.Fatal(err)
//			}
//
//			s := newScheduler(
//				cache,
//				nil,
//				func() *framework.QueuedDguestInfo {
//					return &framework.QueuedDguestInfo{DguestInfo: framework.NewDguestInfo(item.sendDguest)}
//				},
//				nil,
//				internalqueue.NewTestQueue(ctx, nil),
//				profile.Map{
//					testSchedulerName: fwk,
//				},
//				schdulerClient,
//				nil,
//				0)
//			s.ScheduleDguest = func(ctx context.Context, fwk framework.Framework, state *framework.CycleState, dguest *v1alpha1.Dguest) (ScheduleResult, error) {
//				return item.mockResult.result, item.mockResult.err
//			}
//			s.FailureHandler = func(_ context.Context, fwk framework.Framework, p *framework.QueuedDguestInfo, err error, _ string, _ *framework.NominatingInfo) {
//				gotDguest = p.Dguest
//				gotError = err
//
//				msg := truncateMessage(err.Error())
//				fwk.EventRecorder().Eventf(p.Dguest, nil, v1.EventTypeWarning, "FailedScheduling", "Scheduling", msg)
//			}
//			called := make(chan struct{})
//			stopFunc := eventBroadcaster.StartEventWatcher(func(obj runtime.Object) {
//				e, _ := obj.(*eventsv1.Event)
//				if e.Reason != item.eventReason {
//					t.Errorf("got event %v, want %v", e.Reason, item.eventReason)
//				}
//				close(called)
//			})
//			s.scheduleOne(ctx)
//			<-called
//			if e, a := item.expectAssumedDguest, gotAssumedDguest; !reflect.DeepEqual(e, a) {
//				t.Errorf("assumed dguest: wanted %v, got %v", e, a)
//			}
//			if e, a := item.expectErrorDguest, gotDguest; !reflect.DeepEqual(e, a) {
//				t.Errorf("error dguest: wanted %v, got %v", e, a)
//			}
//			if e, a := item.expectForgetDguest, gotForgetDguest; !reflect.DeepEqual(e, a) {
//				t.Errorf("forget dguest: wanted %v, got %v", e, a)
//			}
//			if e, a := item.expectError, gotError; !reflect.DeepEqual(e, a) {
//				t.Errorf("error: wanted %v, got %v", e, a)
//			}
//			if diff := cmp.Diff(item.expectBind, gotBinding); diff != "" {
//				t.Errorf("got binding diff (-want, +got): %s", diff)
//			}
//			stopFunc()
//		})
//	}
//}
//
//func TestSchedulerNoPhantomDguestAfterExpire(t *testing.T) {
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//	queuedDguestStore := clientcache.NewFIFO(clientcache.MetaNamespaceKeyFunc)
//	scache := internalcache.New(100*time.Millisecond, ctx.Done())
//	dguest := dguestWithPort("dguest.Name", "", 8080)
//	food := v1alpha1.Food{ObjectMeta: metav1.ObjectMeta{Name: "food1", UID: types.UID("food1")}}
//	scache.AddFood(&food)
//
//	fns := []st.RegisterPluginFunc{
//		st.RegisterQueueSortPlugin(queuesort.Name, queuesort.New),
//		st.RegisterBindPlugin(defaultbinder.Name, defaultbinder.New),
//		st.RegisterPluginAsExtensions(foodports.Name, foodports.New, "Filter", "PreFilter"),
//	}
//	scheduler, bindingChan, errChan := setupTestSchedulerWithOneDguestOnFood(ctx, t, queuedDguestStore, scache, dguest, &food, fns...)
//
//	waitDguestExpireChan := make(chan struct{})
//	timeout := make(chan struct{})
//	go func() {
//		for {
//			select {
//			case <-timeout:
//				return
//			default:
//			}
//			dguests, err := scache.DguestCount()
//			if err != nil {
//				errChan <- fmt.Errorf("cache.List failed: %v", err)
//				return
//			}
//			if dguests == 0 {
//				close(waitDguestExpireChan)
//				return
//			}
//			time.Sleep(100 * time.Millisecond)
//		}
//	}()
//	// waiting for the assumed dguest to expire
//	select {
//	case err := <-errChan:
//		t.Fatal(err)
//	case <-waitDguestExpireChan:
//	case <-time.After(wait.ForeverTestTimeout):
//		close(timeout)
//		t.Fatalf("timeout timeout in waiting dguest expire after %v", wait.ForeverTestTimeout)
//	}
//
//	// We use conflicted dguest ports to incur fit predicate failure if first dguest not removed.
//	secondDguest := dguestWithPort("bar", "", 8080)
//	queuedDguestStore.Add(secondDguest)
//	scheduler.scheduleOne(ctx)
//	select {
//	case b := <-bindingChan:
//		expectBinding := &v1.Binding{
//			ObjectMeta: metav1.ObjectMeta{Name: "bar", UID: types.UID("bar")},
//			Target:     v1.ObjectReference{Kind: "Food", Name: food.Name},
//		}
//		if !reflect.DeepEqual(expectBinding, b) {
//			t.Errorf("binding want=%v, get=%v", expectBinding, b)
//		}
//	case <-time.After(wait.ForeverTestTimeout):
//		t.Fatalf("timeout in binding after %v", wait.ForeverTestTimeout)
//	}
//}
//
//func TestSchedulerNoPhantomDguestAfterDelete(t *testing.T) {
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//	queuedDguestStore := clientcache.NewFIFO(clientcache.MetaNamespaceKeyFunc)
//	scache := internalcache.New(10*time.Minute, ctx.Done())
//	firstDguest := dguestWithPort("dguest.Name", "", 8080)
//	food := v1alpha1.Food{ObjectMeta: metav1.ObjectMeta{Name: "food1", UID: types.UID("food1")}}
//	scache.AddFood(&food)
//	fns := []st.RegisterPluginFunc{
//		st.RegisterQueueSortPlugin(queuesort.Name, queuesort.New),
//		st.RegisterBindPlugin(defaultbinder.Name, defaultbinder.New),
//		st.RegisterPluginAsExtensions(foodports.Name, foodports.New, "Filter", "PreFilter"),
//	}
//	scheduler, bindingChan, errChan := setupTestSchedulerWithOneDguestOnFood(ctx, t, queuedDguestStore, scache, firstDguest, &food, fns...)
//
//	// We use conflicted dguest ports to incur fit predicate failure.
//	secondDguest := dguestWithPort("bar", "", 8080)
//	queuedDguestStore.Add(secondDguest)
//	// queuedDguestStore: [bar:8080]
//	// cache: [(assumed)foo:8080]
//
//	scheduler.scheduleOne(ctx)
//	select {
//	case err := <-errChan:
//		expectErr := &framework.FitError{
//			Dguest:         secondDguest,
//			NumAllFoods: 1,
//			Diagnosis: framework.Diagnosis{
//				FoodToStatusMap: framework.FoodToStatusMap{
//					food.Name: framework.NewStatus(framework.Unschedulable, foodports.ErrReason).WithFailedPlugin(foodports.Name),
//				},
//				UnschedulablePlugins: sets.NewString(foodports.Name),
//			},
//		}
//		if !reflect.DeepEqual(expectErr, err) {
//			t.Errorf("err want=%v, get=%v", expectErr, err)
//		}
//	case <-time.After(wait.ForeverTestTimeout):
//		t.Fatalf("timeout in fitting after %v", wait.ForeverTestTimeout)
//	}
//
//	// We mimic the workflow of cache behavior when a dguest is removed by user.
//	// Note: if the schedulerfoodinfo timeout would be super short, the first dguest would expire
//	// and would be removed itself (without any explicit actions on schedulerfoodinfo). Even in that case,
//	// explicitly AddDguest will as well correct the behavior.
//	firstDguest.Spec.FoodName = food.Name
//	if err := scache.AddDguest(firstDguest); err != nil {
//		t.Fatalf("err: %v", err)
//	}
//	if err := scache.RemoveDguest(firstDguest); err != nil {
//		t.Fatalf("err: %v", err)
//	}
//
//	queuedDguestStore.Add(secondDguest)
//	scheduler.scheduleOne(ctx)
//	select {
//	case b := <-bindingChan:
//		expectBinding := &v1.Binding{
//			ObjectMeta: metav1.ObjectMeta{Name: "bar", UID: types.UID("bar")},
//			Target:     v1.ObjectReference{Kind: "Food", Name: food.Name},
//		}
//		if !reflect.DeepEqual(expectBinding, b) {
//			t.Errorf("binding want=%v, get=%v", expectBinding, b)
//		}
//	case <-time.After(wait.ForeverTestTimeout):
//		t.Fatalf("timeout in binding after %v", wait.ForeverTestTimeout)
//	}
//}
//
//func TestSchedulerFailedSchedulingReasons(t *testing.T) {
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//	queuedDguestStore := clientcache.NewFIFO(clientcache.MetaNamespaceKeyFunc)
//	scache := internalcache.New(10*time.Minute, ctx.Done())
//
//	// Design the baseline for the dguests, and we will make foods that don't fit it later.
//	var cpu = int64(4)
//	var mem = int64(500)
//	dguestWithTooBigResourceRequests := dguestWithResources("bar", "", v1alpha1.ResourceList{
//		v1.ResourceCPU:    *(resource.NewQuantity(cpu, resource.DecimalSI)),
//		v1.ResourceMemory: *(resource.NewQuantity(mem, resource.DecimalSI)),
//	}, v1alpha1.ResourceList{
//		v1.ResourceCPU:    *(resource.NewQuantity(cpu, resource.DecimalSI)),
//		v1.ResourceMemory: *(resource.NewQuantity(mem, resource.DecimalSI)),
//	})
//
//	// create several foods which cannot schedule the above dguest
//	var foods []*v1alpha1.Food
//	var objects []runtime.Object
//	for i := 0; i < 100; i++ {
//		uid := fmt.Sprintf("food%v", i)
//		food := v1alpha1.Food{
//			ObjectMeta: metav1.ObjectMeta{Name: uid, UID: types.UID(uid)},
//			Status: v1alpha1.FoodStatus{
//				Capacity: v1alpha1.ResourceList{
//					v1.ResourceCPU:    *(resource.NewQuantity(cpu/2, resource.DecimalSI)),
//					v1.ResourceMemory: *(resource.NewQuantity(mem/5, resource.DecimalSI)),
//					v1.ResourceDguests:   *(resource.NewQuantity(10, resource.DecimalSI)),
//				},
//				Allocatable: v1alpha1.ResourceList{
//					v1.ResourceCPU:    *(resource.NewQuantity(cpu/2, resource.DecimalSI)),
//					v1.ResourceMemory: *(resource.NewQuantity(mem/5, resource.DecimalSI)),
//					v1.ResourceDguests:   *(resource.NewQuantity(10, resource.DecimalSI)),
//				}},
//		}
//		scache.AddFood(&food)
//		foods = append(foods, &food)
//		objects = append(objects, &food)
//	}
//
//	// Create expected failure reasons for all the foods. Hopefully they will get rolled up into a non-spammy summary.
//	failedFoodStatues := framework.FoodToStatusMap{}
//	for _, food := range foods {
//		failedFoodStatues[food.Name] = framework.NewStatus(
//			framework.Unschedulable,
//			fmt.Sprintf("Insufficient %v", v1.ResourceCPU),
//			fmt.Sprintf("Insufficient %v", v1.ResourceMemory),
//		).WithFailedPlugin(foodresources.Name)
//	}
//	fns := []st.RegisterPluginFunc{
//		st.RegisterQueueSortPlugin(queuesort.Name, queuesort.New),
//		st.RegisterBindPlugin(defaultbinder.Name, defaultbinder.New),
//		st.RegisterPluginAsExtensions(foodresources.Name, frameworkruntime.FactoryAdapter(feature.Features{}, foodresources.NewFit), "Filter", "PreFilter"),
//	}
//
//	informerFactory := informers.NewSharedInformerFactory(clientsetfake.NewSimpleClientset(objects...), 0)
//	scheduler, _, errChan := setupTestScheduler(ctx, queuedDguestStore, scache, informerFactory, nil, fns...)
//
//	queuedDguestStore.Add(dguestWithTooBigResourceRequests)
//	scheduler.scheduleOne(ctx)
//	select {
//	case err := <-errChan:
//		expectErr := &framework.FitError{
//			Dguest:         dguestWithTooBigResourceRequests,
//			NumAllFoods: len(foods),
//			Diagnosis: framework.Diagnosis{
//				FoodToStatusMap:      failedFoodStatues,
//				UnschedulablePlugins: sets.NewString(foodresources.Name),
//			},
//		}
//		if len(fmt.Sprint(expectErr)) > 150 {
//			t.Errorf("message is too spammy ! %v ", len(fmt.Sprint(expectErr)))
//		}
//		if !reflect.DeepEqual(expectErr, err) {
//			t.Errorf("\n err \nWANT=%+v,\nGOT=%+v", expectErr, err)
//		}
//	case <-time.After(wait.ForeverTestTimeout):
//		t.Fatalf("timeout after %v", wait.ForeverTestTimeout)
//	}
//}
//
//func TestSchedulerWithVolumeBinding(t *testing.T) {
//	findErr := fmt.Errorf("find err")
//	assumeErr := fmt.Errorf("assume err")
//	bindErr := fmt.Errorf("bind err")
//	schdulerClient := clientsetfake.NewSimpleClientset()
//
//	eventBroadcaster := events.NewBroadcaster(&events.EventSinkImpl{Interface: schdulerClient.EventsV1()})
//
//	// This can be small because we wait for dguest to finish scheduling first
//	chanTimeout := 2 * time.Second
//
//	table := []struct {
//		name               string
//		expectError        error
//		expectDguestBind      *v1.Binding
//		expectAssumeCalled bool
//		expectBindCalled   bool
//		eventReason        string
//		volumeBinderConfig *volumebinding.FakeVolumeBinderConfig
//	}{
//		{
//			name: "all bound",
//			volumeBinderConfig: &volumebinding.FakeVolumeBinderConfig{
//				AllBound: true,
//			},
//			expectAssumeCalled: true,
//			expectDguestBind:      &v1.Binding{ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "foo-ns", UID: types.UID("foo")}, Target: v1.ObjectReference{Kind: "Food", Name: "food1"}},
//			eventReason:        "Scheduled",
//		},
//		{
//			name: "bound/invalid pv affinity",
//			volumeBinderConfig: &volumebinding.FakeVolumeBinderConfig{
//				AllBound:    true,
//				FindReasons: volumebinding.ConflictReasons{volumebinding.ErrReasonFoodConflict},
//			},
//			eventReason: "FailedScheduling",
//			expectError: makePredicateError("1 food(s) had volume food affinity conflict"),
//		},
//		{
//			name: "unbound/no matches",
//			volumeBinderConfig: &volumebinding.FakeVolumeBinderConfig{
//				FindReasons: volumebinding.ConflictReasons{volumebinding.ErrReasonBindConflict},
//			},
//			eventReason: "FailedScheduling",
//			expectError: makePredicateError("1 food(s) didn't find available persistent volumes to bind"),
//		},
//		{
//			name: "bound and unbound unsatisfied",
//			volumeBinderConfig: &volumebinding.FakeVolumeBinderConfig{
//				FindReasons: volumebinding.ConflictReasons{volumebinding.ErrReasonBindConflict, volumebinding.ErrReasonFoodConflict},
//			},
//			eventReason: "FailedScheduling",
//			expectError: makePredicateError("1 food(s) didn't find available persistent volumes to bind, 1 food(s) had volume food affinity conflict"),
//		},
//		{
//			name:               "unbound/found matches/bind succeeds",
//			volumeBinderConfig: &volumebinding.FakeVolumeBinderConfig{},
//			expectAssumeCalled: true,
//			expectBindCalled:   true,
//			expectDguestBind:      &v1.Binding{ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "foo-ns", UID: types.UID("foo")}, Target: v1.ObjectReference{Kind: "Food", Name: "food1"}},
//			eventReason:        "Scheduled",
//		},
//		{
//			name: "predicate error",
//			volumeBinderConfig: &volumebinding.FakeVolumeBinderConfig{
//				FindErr: findErr,
//			},
//			eventReason: "FailedScheduling",
//			expectError: fmt.Errorf("running %q filter plugin: %v", volumebinding.Name, findErr),
//		},
//		{
//			name: "assume error",
//			volumeBinderConfig: &volumebinding.FakeVolumeBinderConfig{
//				AssumeErr: assumeErr,
//			},
//			expectAssumeCalled: true,
//			eventReason:        "FailedScheduling",
//			expectError:        fmt.Errorf("running Reserve plugin %q: %w", volumebinding.Name, assumeErr),
//		},
//		{
//			name: "bind error",
//			volumeBinderConfig: &volumebinding.FakeVolumeBinderConfig{
//				BindErr: bindErr,
//			},
//			expectAssumeCalled: true,
//			expectBindCalled:   true,
//			eventReason:        "FailedScheduling",
//			expectError:        fmt.Errorf("running PreBind plugin %q: %w", volumebinding.Name, bindErr),
//		},
//	}
//
//	for _, item := range table {
//		t.Run(item.name, func(t *testing.T) {
//			ctx, cancel := context.WithCancel(context.Background())
//			defer cancel()
//			fakeVolumeBinder := volumebinding.NewFakeVolumeBinder(item.volumeBinderConfig)
//			s, bindingChan, errChan := setupTestSchedulerWithVolumeBinding(ctx, fakeVolumeBinder, eventBroadcaster)
//			eventChan := make(chan struct{})
//			stopFunc := eventBroadcaster.StartEventWatcher(func(obj runtime.Object) {
//				e, _ := obj.(*eventsv1.Event)
//				if e, a := item.eventReason, e.Reason; e != a {
//					t.Errorf("expected %v, got %v", e, a)
//				}
//				close(eventChan)
//			})
//			s.scheduleOne(ctx)
//			// Wait for dguest to succeed or fail scheduling
//			select {
//			case <-eventChan:
//			case <-time.After(wait.ForeverTestTimeout):
//				t.Fatalf("scheduling timeout after %v", wait.ForeverTestTimeout)
//			}
//			stopFunc()
//			// Wait for scheduling to return an error or succeed binding.
//			var (
//				gotErr  error
//				gotBind *v1.Binding
//			)
//			select {
//			case gotErr = <-errChan:
//			case gotBind = <-bindingChan:
//			case <-time.After(chanTimeout):
//				t.Fatalf("did not receive dguest binding or error after %v", chanTimeout)
//			}
//			if item.expectError != nil {
//				if gotErr == nil || item.expectError.Error() != gotErr.Error() {
//					t.Errorf("err \nWANT=%+v,\nGOT=%+v", item.expectError, gotErr)
//				}
//			} else if gotErr != nil {
//				t.Errorf("err \nWANT=%+v,\nGOT=%+v", item.expectError, gotErr)
//			}
//			if !cmp.Equal(item.expectDguestBind, gotBind) {
//				t.Errorf("err \nWANT=%+v,\nGOT=%+v", item.expectDguestBind, gotBind)
//			}
//
//			if item.expectAssumeCalled != fakeVolumeBinder.AssumeCalled {
//				t.Errorf("expectedAssumeCall %v", item.expectAssumeCalled)
//			}
//
//			if item.expectBindCalled != fakeVolumeBinder.BindCalled {
//				t.Errorf("expectedBindCall %v", item.expectBindCalled)
//			}
//		})
//	}
//}
//
//func TestSchedulerBinding(t *testing.T) {
//	table := []struct {
//		dguestName      string
//		extenders    []framework.Extender
//		wantBinderID int
//		name         string
//	}{
//		{
//			name:    "the extender is not a binder",
//			dguestName: "dguest0",
//			extenders: []framework.Extender{
//				&fakeExtender{isBinder: false, interestedDguestName: "dguest0"},
//			},
//			wantBinderID: -1, // default binding.
//		},
//		{
//			name:    "one of the extenders is a binder and interested in dguest",
//			dguestName: "dguest0",
//			extenders: []framework.Extender{
//				&fakeExtender{isBinder: false, interestedDguestName: "dguest0"},
//				&fakeExtender{isBinder: true, interestedDguestName: "dguest0"},
//			},
//			wantBinderID: 1,
//		},
//		{
//			name:    "one of the extenders is a binder, but not interested in dguest",
//			dguestName: "dguest1",
//			extenders: []framework.Extender{
//				&fakeExtender{isBinder: false, interestedDguestName: "dguest1"},
//				&fakeExtender{isBinder: true, interestedDguestName: "dguest0"},
//			},
//			wantBinderID: -1, // default binding.
//		},
//	}
//
//	for _, test := range table {
//		t.Run(test.name, func(t *testing.T) {
//			dguest := st.MakeDguest().Name(test.dguestName).Obj()
//			defaultBound := false
//			schdulerClient := clientsetfake.NewSimpleClientset(dguest)
//			schdulerClient.PrependReactor("create", "dguests", func(action clienttesting.Action) (bool, runtime.Object, error) {
//				if action.GetSubresource() == "binding" {
//					defaultBound = true
//				}
//				return false, nil, nil
//			})
//			ctx, cancel := context.WithCancel(context.Background())
//			defer cancel()
//			fwk, err := st.NewFramework([]st.RegisterPluginFunc{
//				st.RegisterQueueSortPlugin(queuesort.Name, queuesort.New),
//				st.RegisterBindPlugin(defaultbinder.Name, defaultbinder.New),
//			}, "", ctx.Done(), frameworkruntime.WithClientSet(schdulerClient), frameworkruntime.WithEventRecorder(&events.FakeRecorder{}))
//			if err != nil {
//				t.Fatal(err)
//			}
//			sched := &Scheduler{
//				Extenders:                test.extenders,
//				Cache:                    internalcache.New(100*time.Millisecond, ctx.Done()),
//				foodInfoSnapshot:         nil,
//				percentageOfFoodsToScore: 0,
//			}
//			err = sched.bind(ctx, fwk, dguest, "food", nil)
//			if err != nil {
//				t.Error(err)
//			}
//
//			// Checking default binding.
//			if wantBound := test.wantBinderID == -1; defaultBound != wantBound {
//				t.Errorf("got bound with default binding: %v, want %v", defaultBound, wantBound)
//			}
//
//			// Checking extenders binding.
//			for i, ext := range test.extenders {
//				wantBound := i == test.wantBinderID
//				if gotBound := ext.(*fakeExtender).gotBind; gotBound != wantBound {
//					t.Errorf("got bound with extender #%d: %v, want %v", i, gotBound, wantBound)
//				}
//			}
//
//		})
//	}
//}
//
//func TestUpdateDguest(t *testing.T) {
//	tests := []struct {
//		name                     string
//		currentDguestConditions     []v1alpha1.DguestCondition
//		newDguestCondition          *v1alpha1.DguestCondition
//		currentNominatedFoodName string
//		newNominatingInfo        *framework.NominatingInfo
//		expectedPatchRequests    int
//		expectedPatchDataPattern string
//	}{
//		{
//			name:                 "Should make patch request to add dguest condition when there are none currently",
//			currentDguestConditions: []v1alpha1.DguestCondition{},
//			newDguestCondition: &v1alpha1.DguestCondition{
//				Type:               "newType",
//				Status:             "newStatus",
//				LastProbeTime:      metav1.NewTime(time.Date(2020, 5, 13, 1, 1, 1, 1, time.UTC)),
//				LastTransitionTime: metav1.NewTime(time.Date(2020, 5, 12, 1, 1, 1, 1, time.UTC)),
//				Reason:             "newReason",
//				Message:            "newMessage",
//			},
//			expectedPatchRequests:    1,
//			expectedPatchDataPattern: `{"status":{"conditions":\[{"lastProbeTime":"2020-05-13T01:01:01Z","lastTransitionTime":".*","message":"newMessage","reason":"newReason","status":"newStatus","type":"newType"}]}}`,
//		},
//		{
//			name: "Should make patch request to add a new dguest condition when there is already one with another type",
//			currentDguestConditions: []v1alpha1.DguestCondition{
//				{
//					Type:               "someOtherType",
//					Status:             "someOtherTypeStatus",
//					LastProbeTime:      metav1.NewTime(time.Date(2020, 5, 11, 0, 0, 0, 0, time.UTC)),
//					LastTransitionTime: metav1.NewTime(time.Date(2020, 5, 10, 0, 0, 0, 0, time.UTC)),
//					Reason:             "someOtherTypeReason",
//					Message:            "someOtherTypeMessage",
//				},
//			},
//			newDguestCondition: &v1alpha1.DguestCondition{
//				Type:               "newType",
//				Status:             "newStatus",
//				LastProbeTime:      metav1.NewTime(time.Date(2020, 5, 13, 1, 1, 1, 1, time.UTC)),
//				LastTransitionTime: metav1.NewTime(time.Date(2020, 5, 12, 1, 1, 1, 1, time.UTC)),
//				Reason:             "newReason",
//				Message:            "newMessage",
//			},
//			expectedPatchRequests:    1,
//			expectedPatchDataPattern: `{"status":{"\$setElementOrder/conditions":\[{"type":"someOtherType"},{"type":"newType"}],"conditions":\[{"lastProbeTime":"2020-05-13T01:01:01Z","lastTransitionTime":".*","message":"newMessage","reason":"newReason","status":"newStatus","type":"newType"}]}}`,
//		},
//		{
//			name: "Should make patch request to update an existing dguest condition",
//			currentDguestConditions: []v1alpha1.DguestCondition{
//				{
//					Type:               "currentType",
//					Status:             "currentStatus",
//					LastProbeTime:      metav1.NewTime(time.Date(2020, 5, 13, 0, 0, 0, 0, time.UTC)),
//					LastTransitionTime: metav1.NewTime(time.Date(2020, 5, 12, 0, 0, 0, 0, time.UTC)),
//					Reason:             "currentReason",
//					Message:            "currentMessage",
//				},
//			},
//			newDguestCondition: &v1alpha1.DguestCondition{
//				Type:               "currentType",
//				Status:             "newStatus",
//				LastProbeTime:      metav1.NewTime(time.Date(2020, 5, 13, 1, 1, 1, 1, time.UTC)),
//				LastTransitionTime: metav1.NewTime(time.Date(2020, 5, 12, 1, 1, 1, 1, time.UTC)),
//				Reason:             "newReason",
//				Message:            "newMessage",
//			},
//			expectedPatchRequests:    1,
//			expectedPatchDataPattern: `{"status":{"\$setElementOrder/conditions":\[{"type":"currentType"}],"conditions":\[{"lastProbeTime":"2020-05-13T01:01:01Z","lastTransitionTime":".*","message":"newMessage","reason":"newReason","status":"newStatus","type":"currentType"}]}}`,
//		},
//		{
//			name: "Should make patch request to update an existing dguest condition, but the transition time should remain unchanged because the status is the same",
//			currentDguestConditions: []v1alpha1.DguestCondition{
//				{
//					Type:               "currentType",
//					Status:             "currentStatus",
//					LastProbeTime:      metav1.NewTime(time.Date(2020, 5, 13, 0, 0, 0, 0, time.UTC)),
//					LastTransitionTime: metav1.NewTime(time.Date(2020, 5, 12, 0, 0, 0, 0, time.UTC)),
//					Reason:             "currentReason",
//					Message:            "currentMessage",
//				},
//			},
//			newDguestCondition: &v1alpha1.DguestCondition{
//				Type:               "currentType",
//				Status:             "currentStatus",
//				LastProbeTime:      metav1.NewTime(time.Date(2020, 5, 13, 1, 1, 1, 1, time.UTC)),
//				LastTransitionTime: metav1.NewTime(time.Date(2020, 5, 12, 0, 0, 0, 0, time.UTC)),
//				Reason:             "newReason",
//				Message:            "newMessage",
//			},
//			expectedPatchRequests:    1,
//			expectedPatchDataPattern: `{"status":{"\$setElementOrder/conditions":\[{"type":"currentType"}],"conditions":\[{"lastProbeTime":"2020-05-13T01:01:01Z","message":"newMessage","reason":"newReason","type":"currentType"}]}}`,
//		},
//		{
//			name: "Should not make patch request if dguest condition already exists and is identical and nominated food name is not set",
//			currentDguestConditions: []v1alpha1.DguestCondition{
//				{
//					Type:               "currentType",
//					Status:             "currentStatus",
//					LastProbeTime:      metav1.NewTime(time.Date(2020, 5, 13, 0, 0, 0, 0, time.UTC)),
//					LastTransitionTime: metav1.NewTime(time.Date(2020, 5, 12, 0, 0, 0, 0, time.UTC)),
//					Reason:             "currentReason",
//					Message:            "currentMessage",
//				},
//			},
//			newDguestCondition: &v1alpha1.DguestCondition{
//				Type:               "currentType",
//				Status:             "currentStatus",
//				LastProbeTime:      metav1.NewTime(time.Date(2020, 5, 13, 0, 0, 0, 0, time.UTC)),
//				LastTransitionTime: metav1.NewTime(time.Date(2020, 5, 12, 0, 0, 0, 0, time.UTC)),
//				Reason:             "currentReason",
//				Message:            "currentMessage",
//			},
//			currentNominatedFoodName: "food1",
//			expectedPatchRequests:    0,
//		},
//		{
//			name: "Should make patch request if dguest condition already exists and is identical but nominated food name is set and different",
//			currentDguestConditions: []v1alpha1.DguestCondition{
//				{
//					Type:               "currentType",
//					Status:             "currentStatus",
//					LastProbeTime:      metav1.NewTime(time.Date(2020, 5, 13, 0, 0, 0, 0, time.UTC)),
//					LastTransitionTime: metav1.NewTime(time.Date(2020, 5, 12, 0, 0, 0, 0, time.UTC)),
//					Reason:             "currentReason",
//					Message:            "currentMessage",
//				},
//			},
//			newDguestCondition: &v1alpha1.DguestCondition{
//				Type:               "currentType",
//				Status:             "currentStatus",
//				LastProbeTime:      metav1.NewTime(time.Date(2020, 5, 13, 0, 0, 0, 0, time.UTC)),
//				LastTransitionTime: metav1.NewTime(time.Date(2020, 5, 12, 0, 0, 0, 0, time.UTC)),
//				Reason:             "currentReason",
//				Message:            "currentMessage",
//			},
//			newNominatingInfo:        &framework.NominatingInfo{NominatingMode: framework.ModeOverride, NominatedFoodName: "food1"},
//			expectedPatchRequests:    1,
//			expectedPatchDataPattern: `{"status":{"nominatedFoodName":"food1"}}`,
//		},
//	}
//	for _, test := range tests {
//		t.Run(test.name, func(t *testing.T) {
//			actualPatchRequests := 0
//			var actualPatchData string
//			cs := &clientsetfake.Clientset{}
//			cs.AddReactor("patch", "dguests", func(action clienttesting.Action) (bool, runtime.Object, error) {
//				actualPatchRequests++
//				patch := action.(clienttesting.PatchAction)
//				actualPatchData = string(patch.GetPatch())
//				// For this test, we don't care about the result of the patched dguest, just that we got the expected
//				// patch request, so just returning &v1alpha1.Dguest{} here is OK because scheduler doesn't use the response.
//				return true, &v1alpha1.Dguest{}, nil
//			})
//
//			dguest := st.MakeDguest().Name("foo").NominatedFoodName(test.currentNominatedFoodName).Conditions(test.currentDguestConditions).Obj()
//
//			ctx, cancel := context.WithCancel(context.Background())
//			defer cancel()
//			if err := updateDguest(ctx, cs, dguest, test.newDguestCondition, test.newNominatingInfo); err != nil {
//				t.Fatalf("Error calling update: %v", err)
//			}
//
//			if actualPatchRequests != test.expectedPatchRequests {
//				t.Fatalf("Actual patch requests (%d) does not equal expected patch requests (%d), actual patch data: %v", actualPatchRequests, test.expectedPatchRequests, actualPatchData)
//			}
//
//			regex, err := regexp.Compile(test.expectedPatchDataPattern)
//			if err != nil {
//				t.Fatalf("Error compiling regexp for %v: %v", test.expectedPatchDataPattern, err)
//			}
//
//			if test.expectedPatchRequests > 0 && !regex.MatchString(actualPatchData) {
//				t.Fatalf("Patch data mismatch: Actual was %v, but expected to match regexp %v", actualPatchData, test.expectedPatchDataPattern)
//			}
//		})
//	}
//}
//
//func TestSelectHost(t *testing.T) {
//	tests := []struct {
//		name          string
//		list          framework.FoodScoreList
//		possibleHosts sets.String
//		expectsErr    bool
//	}{
//		{
//			name: "unique properly ordered scores",
//			list: []framework.FoodScore{
//				{Name: "food1.1", Score: 1},
//				{Name: "food2.1", Score: 2},
//			},
//			possibleHosts: sets.NewString("food2.1"),
//			expectsErr:    false,
//		},
//		{
//			name: "equal scores",
//			list: []framework.FoodScore{
//				{Name: "food1.1", Score: 1},
//				{Name: "food1.2", Score: 2},
//				{Name: "food1.3", Score: 2},
//				{Name: "food2.1", Score: 2},
//			},
//			possibleHosts: sets.NewString("food1.2", "food1.3", "food2.1"),
//			expectsErr:    false,
//		},
//		{
//			name: "out of order scores",
//			list: []framework.FoodScore{
//				{Name: "food1.1", Score: 3},
//				{Name: "food1.2", Score: 3},
//				{Name: "food2.1", Score: 2},
//				{Name: "food3.1", Score: 1},
//				{Name: "food1.3", Score: 3},
//			},
//			possibleHosts: sets.NewString("food1.1", "food1.2", "food1.3"),
//			expectsErr:    false,
//		},
//		{
//			name:          "empty priority list",
//			list:          []framework.FoodScore{},
//			possibleHosts: sets.NewString(),
//			expectsErr:    true,
//		},
//	}
//
//	for _, test := range tests {
//		t.Run(test.name, func(t *testing.T) {
//			// increase the randomness
//			for i := 0; i < 10; i++ {
//				got, err := selectHost(test.list)
//				if test.expectsErr {
//					if err == nil {
//						t.Error("Unexpected non-error")
//					}
//				} else {
//					if err != nil {
//						t.Errorf("Unexpected error: %v", err)
//					}
//					if !test.possibleHosts.Has(got) {
//						t.Errorf("got %s is not in the possible map %v", got, test.possibleHosts)
//					}
//				}
//			}
//		})
//	}
//}
//
//func TestFindFoodsThatPassExtenders(t *testing.T) {
//	tests := []struct {
//		name                  string
//		extenders             []st.FakeExtender
//		foods                 []*v1alpha1.Food
//		filteredFoodsStatuses framework.FoodToStatusMap
//		expectsErr            bool
//		expectedFoods         []*v1alpha1.Food
//		expectedStatuses      framework.FoodToStatusMap
//	}{
//		{
//			name: "error",
//			extenders: []st.FakeExtender{
//				{
//					ExtenderName: "FakeExtender1",
//					Predicates:   []st.FitPredicate{st.ErrorPredicateExtender},
//				},
//			},
//			foods:                 makeFoodList([]string{"a"}),
//			filteredFoodsStatuses: make(framework.FoodToStatusMap),
//			expectsErr:            true,
//		},
//		{
//			name: "success",
//			extenders: []st.FakeExtender{
//				{
//					ExtenderName: "FakeExtender1",
//					Predicates:   []st.FitPredicate{st.TruePredicateExtender},
//				},
//			},
//			foods:                 makeFoodList([]string{"a"}),
//			filteredFoodsStatuses: make(framework.FoodToStatusMap),
//			expectsErr:            false,
//			expectedFoods:         makeFoodList([]string{"a"}),
//			expectedStatuses:      make(framework.FoodToStatusMap),
//		},
//		{
//			name: "unschedulable",
//			extenders: []st.FakeExtender{
//				{
//					ExtenderName: "FakeExtender1",
//					Predicates: []st.FitPredicate{func(dguest *v1alpha1.Dguest, food *v1alpha1.Food) *framework.Status {
//						if food.Name == "a" {
//							return framework.NewStatus(framework.Success)
//						}
//						return framework.NewStatus(framework.Unschedulable, fmt.Sprintf("food %q is not allowed", food.Name))
//					}},
//				},
//			},
//			foods:                 makeFoodList([]string{"a", "b"}),
//			filteredFoodsStatuses: make(framework.FoodToStatusMap),
//			expectsErr:            false,
//			expectedFoods:         makeFoodList([]string{"a"}),
//			expectedStatuses: framework.FoodToStatusMap{
//				"b": framework.NewStatus(framework.Unschedulable, fmt.Sprintf("FakeExtender: food %q failed", "b")),
//			},
//		},
//		{
//			name: "unschedulable and unresolvable",
//			extenders: []st.FakeExtender{
//				{
//					ExtenderName: "FakeExtender1",
//					Predicates: []st.FitPredicate{func(dguest *v1alpha1.Dguest, food *v1alpha1.Food) *framework.Status {
//						if food.Name == "a" {
//							return framework.NewStatus(framework.Success)
//						}
//						if food.Name == "b" {
//							return framework.NewStatus(framework.Unschedulable, fmt.Sprintf("food %q is not allowed", food.Name))
//						}
//						return framework.NewStatus(framework.UnschedulableAndUnresolvable, fmt.Sprintf("food %q is not allowed", food.Name))
//					}},
//				},
//			},
//			foods:                 makeFoodList([]string{"a", "b", "c"}),
//			filteredFoodsStatuses: make(framework.FoodToStatusMap),
//			expectsErr:            false,
//			expectedFoods:         makeFoodList([]string{"a"}),
//			expectedStatuses: framework.FoodToStatusMap{
//				"b": framework.NewStatus(framework.Unschedulable, fmt.Sprintf("FakeExtender: food %q failed", "b")),
//				"c": framework.NewStatus(framework.UnschedulableAndUnresolvable, fmt.Sprintf("FakeExtender: food %q failed and unresolvable", "c")),
//			},
//		},
//		{
//			name: "extender may overwrite the statuses",
//			extenders: []st.FakeExtender{
//				{
//					ExtenderName: "FakeExtender1",
//					Predicates: []st.FitPredicate{func(dguest *v1alpha1.Dguest, food *v1alpha1.Food) *framework.Status {
//						if food.Name == "a" {
//							return framework.NewStatus(framework.Success)
//						}
//						if food.Name == "b" {
//							return framework.NewStatus(framework.Unschedulable, fmt.Sprintf("food %q is not allowed", food.Name))
//						}
//						return framework.NewStatus(framework.UnschedulableAndUnresolvable, fmt.Sprintf("food %q is not allowed", food.Name))
//					}},
//				},
//			},
//			foods: makeFoodList([]string{"a", "b", "c"}),
//			filteredFoodsStatuses: framework.FoodToStatusMap{
//				"c": framework.NewStatus(framework.Unschedulable, fmt.Sprintf("FakeFilterPlugin: food %q failed", "c")),
//			},
//			expectsErr:    false,
//			expectedFoods: makeFoodList([]string{"a"}),
//			expectedStatuses: framework.FoodToStatusMap{
//				"b": framework.NewStatus(framework.Unschedulable, fmt.Sprintf("FakeExtender: food %q failed", "b")),
//				"c": framework.NewStatus(framework.UnschedulableAndUnresolvable, fmt.Sprintf("FakeFilterPlugin: food %q failed", "c"), fmt.Sprintf("FakeExtender: food %q failed and unresolvable", "c")),
//			},
//		},
//		{
//			name: "multiple extenders",
//			extenders: []st.FakeExtender{
//				{
//					ExtenderName: "FakeExtender1",
//					Predicates: []st.FitPredicate{func(dguest *v1alpha1.Dguest, food *v1alpha1.Food) *framework.Status {
//						if food.Name == "a" {
//							return framework.NewStatus(framework.Success)
//						}
//						if food.Name == "b" {
//							return framework.NewStatus(framework.Unschedulable, fmt.Sprintf("food %q is not allowed", food.Name))
//						}
//						return framework.NewStatus(framework.UnschedulableAndUnresolvable, fmt.Sprintf("food %q is not allowed", food.Name))
//					}},
//				},
//				{
//					ExtenderName: "FakeExtender1",
//					Predicates: []st.FitPredicate{func(dguest *v1alpha1.Dguest, food *v1alpha1.Food) *framework.Status {
//						if food.Name == "a" {
//							return framework.NewStatus(framework.Success)
//						}
//						return framework.NewStatus(framework.Unschedulable, fmt.Sprintf("food %q is not allowed", food.Name))
//					}},
//				},
//			},
//			foods:                 makeFoodList([]string{"a", "b", "c"}),
//			filteredFoodsStatuses: make(framework.FoodToStatusMap),
//			expectsErr:            false,
//			expectedFoods:         makeFoodList([]string{"a"}),
//			expectedStatuses: framework.FoodToStatusMap{
//				"b": framework.NewStatus(framework.Unschedulable, fmt.Sprintf("FakeExtender: food %q failed", "b")),
//				"c": framework.NewStatus(framework.UnschedulableAndUnresolvable, fmt.Sprintf("FakeExtender: food %q failed and unresolvable", "c")),
//			},
//		},
//	}
//
//	cmpOpts := []cmp.Option{
//		cmp.Comparer(func(s1 framework.Status, s2 framework.Status) bool {
//			return s1.Code() == s2.Code() && reflect.DeepEqual(s1.Reasons(), s2.Reasons())
//		}),
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			var extenders []framework.Extender
//			for ii := range tt.extenders {
//				extenders = append(extenders, &tt.extenders[ii])
//			}
//
//			dguest := st.MakeDguest().Name("1").UID("1").Obj()
//			got, err := findFoodsThatPassExtenders(extenders, dguest, tt.foods, tt.filteredFoodsStatuses)
//			if tt.expectsErr {
//				if err == nil {
//					t.Error("Unexpected non-error")
//				}
//			} else {
//				if err != nil {
//					t.Errorf("Unexpected error: %v", err)
//				}
//				if diff := cmp.Diff(tt.expectedFoods, got); diff != "" {
//					t.Errorf("filtered foods (-want,+got):\n%s", diff)
//				}
//				if diff := cmp.Diff(tt.expectedStatuses, tt.filteredFoodsStatuses, cmpOpts...); diff != "" {
//					t.Errorf("filtered statuses (-want,+got):\n%s", diff)
//				}
//			}
//		})
//	}
//}
//
//func TestSchedulerScheduleDguest(t *testing.T) {
//	fts := feature.Features{}
//	tests := []struct {
//		name               string
//		registerPlugins    []st.RegisterPluginFunc
//		foods              []string
//		pvcs               []v1.PersistentVolumeClaim
//		dguest                *v1alpha1.Dguest
//		dguests               []*v1alpha1.Dguest
//		wantFoods          sets.String
//		wantEvaluatedFoods *int32
//		wErr               error
//	}{
//		{
//			registerPlugins: []st.RegisterPluginFunc{
//				st.RegisterQueueSortPlugin(queuesort.Name, queuesort.New),
//				st.RegisterFilterPlugin("FalseFilter", st.NewFalseFilterPlugin),
//				st.RegisterBindPlugin(defaultbinder.Name, defaultbinder.New),
//			},
//			foods: []string{"food1", "food2"},
//			dguest:   st.MakeDguest().Name("2").UID("2").Obj(),
//			name:  "test 1",
//			wErr: &framework.FitError{
//				Dguest:         st.MakeDguest().Name("2").UID("2").Obj(),
//				NumAllFoods: 2,
//				Diagnosis: framework.Diagnosis{
//					FoodToStatusMap: framework.FoodToStatusMap{
//						"food1": framework.NewStatus(framework.Unschedulable, st.ErrReasonFake).WithFailedPlugin("FalseFilter"),
//						"food2": framework.NewStatus(framework.Unschedulable, st.ErrReasonFake).WithFailedPlugin("FalseFilter"),
//					},
//					UnschedulablePlugins: sets.NewString("FalseFilter"),
//				},
//			},
//		},
//		{
//			registerPlugins: []st.RegisterPluginFunc{
//				st.RegisterQueueSortPlugin(queuesort.Name, queuesort.New),
//				st.RegisterFilterPlugin("TrueFilter", st.NewTrueFilterPlugin),
//				st.RegisterBindPlugin(defaultbinder.Name, defaultbinder.New),
//			},
//			foods:     []string{"food1", "food2"},
//			dguest:       st.MakeDguest().Name("ignore").UID("ignore").Obj(),
//			wantFoods: sets.NewString("food1", "food2"),
//			name:      "test 2",
//			wErr:      nil,
//		},
//		{
//			// Fits on a food where the dguest ID matches the food name
//			registerPlugins: []st.RegisterPluginFunc{
//				st.RegisterQueueSortPlugin(queuesort.Name, queuesort.New),
//				st.RegisterFilterPlugin("MatchFilter", st.NewMatchFilterPlugin),
//				st.RegisterBindPlugin(defaultbinder.Name, defaultbinder.New),
//			},
//			foods:     []string{"food1", "food2"},
//			dguest:       st.MakeDguest().Name("food2").UID("food2").Obj(),
//			wantFoods: sets.NewString("food2"),
//			name:      "test 3",
//			wErr:      nil,
//		},
//		{
//			registerPlugins: []st.RegisterPluginFunc{
//				st.RegisterQueueSortPlugin(queuesort.Name, queuesort.New),
//				st.RegisterFilterPlugin("TrueFilter", st.NewTrueFilterPlugin),
//				st.RegisterScorePlugin("NumericMap", newNumericMapPlugin(), 1),
//				st.RegisterBindPlugin(defaultbinder.Name, defaultbinder.New),
//			},
//			foods:     []string{"3", "2", "1"},
//			dguest:       st.MakeDguest().Name("ignore").UID("ignore").Obj(),
//			wantFoods: sets.NewString("3"),
//			name:      "test 4",
//			wErr:      nil,
//		},
//		{
//			registerPlugins: []st.RegisterPluginFunc{
//				st.RegisterQueueSortPlugin(queuesort.Name, queuesort.New),
//				st.RegisterFilterPlugin("MatchFilter", st.NewMatchFilterPlugin),
//				st.RegisterScorePlugin("NumericMap", newNumericMapPlugin(), 1),
//				st.RegisterBindPlugin(defaultbinder.Name, defaultbinder.New),
//			},
//			foods:     []string{"3", "2", "1"},
//			dguest:       st.MakeDguest().Name("2").UID("2").Obj(),
//			wantFoods: sets.NewString("2"),
//			name:      "test 5",
//			wErr:      nil,
//		},
//		{
//			registerPlugins: []st.RegisterPluginFunc{
//				st.RegisterQueueSortPlugin(queuesort.Name, queuesort.New),
//				st.RegisterFilterPlugin("TrueFilter", st.NewTrueFilterPlugin),
//				st.RegisterScorePlugin("NumericMap", newNumericMapPlugin(), 1),
//				st.RegisterScorePlugin("ReverseNumericMap", newReverseNumericMapPlugin(), 2),
//				st.RegisterBindPlugin(defaultbinder.Name, defaultbinder.New),
//			},
//			foods:     []string{"3", "2", "1"},
//			dguest:       st.MakeDguest().Name("2").UID("2").Obj(),
//			wantFoods: sets.NewString("1"),
//			name:      "test 6",
//			wErr:      nil,
//		},
//		{
//			registerPlugins: []st.RegisterPluginFunc{
//				st.RegisterQueueSortPlugin(queuesort.Name, queuesort.New),
//				st.RegisterFilterPlugin("TrueFilter", st.NewTrueFilterPlugin),
//				st.RegisterFilterPlugin("FalseFilter", st.NewFalseFilterPlugin),
//				st.RegisterScorePlugin("NumericMap", newNumericMapPlugin(), 1),
//				st.RegisterBindPlugin(defaultbinder.Name, defaultbinder.New),
//			},
//			foods: []string{"3", "2", "1"},
//			dguest:   st.MakeDguest().Name("2").UID("2").Obj(),
//			name:  "test 7",
//			wErr: &framework.FitError{
//				Dguest:         st.MakeDguest().Name("2").UID("2").Obj(),
//				NumAllFoods: 3,
//				Diagnosis: framework.Diagnosis{
//					FoodToStatusMap: framework.FoodToStatusMap{
//						"3": framework.NewStatus(framework.Unschedulable, st.ErrReasonFake).WithFailedPlugin("FalseFilter"),
//						"2": framework.NewStatus(framework.Unschedulable, st.ErrReasonFake).WithFailedPlugin("FalseFilter"),
//						"1": framework.NewStatus(framework.Unschedulable, st.ErrReasonFake).WithFailedPlugin("FalseFilter"),
//					},
//					UnschedulablePlugins: sets.NewString("FalseFilter"),
//				},
//			},
//		},
//		{
//			registerPlugins: []st.RegisterPluginFunc{
//				st.RegisterQueueSortPlugin(queuesort.Name, queuesort.New),
//				st.RegisterFilterPlugin("NoDguestsFilter", NewNoDguestsFilterPlugin),
//				st.RegisterFilterPlugin("MatchFilter", st.NewMatchFilterPlugin),
//				st.RegisterScorePlugin("NumericMap", newNumericMapPlugin(), 1),
//				st.RegisterBindPlugin(defaultbinder.Name, defaultbinder.New),
//			},
//			dguests: []*v1alpha1.Dguest{
//				st.MakeDguest().Name("2").UID("2").Food("2").Phase(v1alpha1.DguestRunning).Obj(),
//			},
//			dguest:   st.MakeDguest().Name("2").UID("2").Obj(),
//			foods: []string{"1", "2"},
//			name:  "test 8",
//			wErr: &framework.FitError{
//				Dguest:         st.MakeDguest().Name("2").UID("2").Obj(),
//				NumAllFoods: 2,
//				Diagnosis: framework.Diagnosis{
//					FoodToStatusMap: framework.FoodToStatusMap{
//						"1": framework.NewStatus(framework.Unschedulable, st.ErrReasonFake).WithFailedPlugin("MatchFilter"),
//						"2": framework.NewStatus(framework.Unschedulable, st.ErrReasonFake).WithFailedPlugin("NoDguestsFilter"),
//					},
//					UnschedulablePlugins: sets.NewString("MatchFilter", "NoDguestsFilter"),
//				},
//			},
//		},
//		{
//			// Dguest with existing PVC
//			registerPlugins: []st.RegisterPluginFunc{
//				st.RegisterQueueSortPlugin(queuesort.Name, queuesort.New),
//				st.RegisterPreFilterPlugin(volumebinding.Name, frameworkruntime.FactoryAdapter(fts, volumebinding.New)),
//				st.RegisterFilterPlugin("TrueFilter", st.NewTrueFilterPlugin),
//				st.RegisterBindPlugin(defaultbinder.Name, defaultbinder.New),
//			},
//			foods: []string{"food1", "food2"},
//			pvcs: []v1.PersistentVolumeClaim{
//				{
//					ObjectMeta: metav1.ObjectMeta{Name: "existingPVC", UID: types.UID("existingPVC"), Namespace: v1.NamespaceDefault},
//					Spec:       v1.PersistentVolumeClaimSpec{VolumeName: "existingPV"},
//				},
//			},
//			dguest:       st.MakeDguest().Name("ignore").UID("ignore").Namespace(v1.NamespaceDefault).PVC("existingPVC").Obj(),
//			wantFoods: sets.NewString("food1", "food2"),
//			name:      "existing PVC",
//			wErr:      nil,
//		},
//		{
//			// Dguest with non existing PVC
//			registerPlugins: []st.RegisterPluginFunc{
//				st.RegisterQueueSortPlugin(queuesort.Name, queuesort.New),
//				st.RegisterPreFilterPlugin(volumebinding.Name, frameworkruntime.FactoryAdapter(fts, volumebinding.New)),
//				st.RegisterFilterPlugin("TrueFilter", st.NewTrueFilterPlugin),
//				st.RegisterBindPlugin(defaultbinder.Name, defaultbinder.New),
//			},
//			foods: []string{"food1", "food2"},
//			dguest:   st.MakeDguest().Name("ignore").UID("ignore").PVC("unknownPVC").Obj(),
//			name:  "unknown PVC",
//			wErr: &framework.FitError{
//				Dguest:         st.MakeDguest().Name("ignore").UID("ignore").PVC("unknownPVC").Obj(),
//				NumAllFoods: 2,
//				Diagnosis: framework.Diagnosis{
//					FoodToStatusMap: framework.FoodToStatusMap{
//						"food1": framework.NewStatus(framework.UnschedulableAndUnresolvable, `persistentvolumeclaim "unknownPVC" not found`).WithFailedPlugin(volumebinding.Name),
//						"food2": framework.NewStatus(framework.UnschedulableAndUnresolvable, `persistentvolumeclaim "unknownPVC" not found`).WithFailedPlugin(volumebinding.Name),
//					},
//					UnschedulablePlugins: sets.NewString(volumebinding.Name),
//				},
//			},
//		},
//		{
//			// Dguest with deleting PVC
//			registerPlugins: []st.RegisterPluginFunc{
//				st.RegisterQueueSortPlugin(queuesort.Name, queuesort.New),
//				st.RegisterPreFilterPlugin(volumebinding.Name, frameworkruntime.FactoryAdapter(fts, volumebinding.New)),
//				st.RegisterFilterPlugin("TrueFilter", st.NewTrueFilterPlugin),
//				st.RegisterBindPlugin(defaultbinder.Name, defaultbinder.New),
//			},
//			foods: []string{"food1", "food2"},
//			pvcs:  []v1.PersistentVolumeClaim{{ObjectMeta: metav1.ObjectMeta{Name: "existingPVC", UID: types.UID("existingPVC"), Namespace: v1.NamespaceDefault, DeletionTimestamp: &metav1.Time{}}}},
//			dguest:   st.MakeDguest().Name("ignore").UID("ignore").Namespace(v1.NamespaceDefault).PVC("existingPVC").Obj(),
//			name:  "deleted PVC",
//			wErr: &framework.FitError{
//				Dguest:         st.MakeDguest().Name("ignore").UID("ignore").Namespace(v1.NamespaceDefault).PVC("existingPVC").Obj(),
//				NumAllFoods: 2,
//				Diagnosis: framework.Diagnosis{
//					FoodToStatusMap: framework.FoodToStatusMap{
//						"food1": framework.NewStatus(framework.UnschedulableAndUnresolvable, `persistentvolumeclaim "existingPVC" is being deleted`).WithFailedPlugin(volumebinding.Name),
//						"food2": framework.NewStatus(framework.UnschedulableAndUnresolvable, `persistentvolumeclaim "existingPVC" is being deleted`).WithFailedPlugin(volumebinding.Name),
//					},
//					UnschedulablePlugins: sets.NewString(volumebinding.Name),
//				},
//			},
//		},
//		{
//			registerPlugins: []st.RegisterPluginFunc{
//				st.RegisterQueueSortPlugin(queuesort.Name, queuesort.New),
//				st.RegisterFilterPlugin("TrueFilter", st.NewTrueFilterPlugin),
//				st.RegisterScorePlugin("FalseMap", newFalseMapPlugin(), 1),
//				st.RegisterScorePlugin("TrueMap", newTrueMapPlugin(), 2),
//				st.RegisterBindPlugin(defaultbinder.Name, defaultbinder.New),
//			},
//			foods: []string{"2", "1"},
//			dguest:   st.MakeDguest().Name("2").Obj(),
//			name:  "test error with priority map",
//			wErr:  fmt.Errorf("running Score plugins: %w", fmt.Errorf(`plugin "FalseMap" failed with: %w`, errPrioritize)),
//		},
//		{
//			name: "test dguesttopologyspread plugin - 2 foods with maxskew=1",
//			registerPlugins: []st.RegisterPluginFunc{
//				st.RegisterQueueSortPlugin(queuesort.Name, queuesort.New),
//				st.RegisterPluginAsExtensions(
//					dguesttopologyspread.Name,
//					dguestTopologySpreadFunc,
//					"PreFilter",
//					"Filter",
//				),
//				st.RegisterBindPlugin(defaultbinder.Name, defaultbinder.New),
//			},
//			foods: []string{"food1", "food2"},
//			dguest: st.MakeDguest().Name("p").UID("p").Label("foo", "").SpreadConstraint(1, "hostname", v1.DoNotSchedule, &metav1.LabelSelector{
//				MatchExpressions: []metav1.LabelSelectorRequirement{
//					{
//						Key:      "foo",
//						Operator: metav1.LabelSelectorOpExists,
//					},
//				},
//			}, nil, nil, nil, nil).Obj(),
//			dguests: []*v1alpha1.Dguest{
//				st.MakeDguest().Name("dguest1").UID("dguest1").Label("foo", "").Food("food1").Phase(v1alpha1.DguestRunning).Obj(),
//			},
//			wantFoods: sets.NewString("food2"),
//			wErr:      nil,
//		},
//		{
//			name: "test dguesttopologyspread plugin - 3 foods with maxskew=2",
//			registerPlugins: []st.RegisterPluginFunc{
//				st.RegisterQueueSortPlugin(queuesort.Name, queuesort.New),
//				st.RegisterPluginAsExtensions(
//					dguesttopologyspread.Name,
//					dguestTopologySpreadFunc,
//					"PreFilter",
//					"Filter",
//				),
//				st.RegisterBindPlugin(defaultbinder.Name, defaultbinder.New),
//			},
//			foods: []string{"food1", "food2", "food3"},
//			dguest: st.MakeDguest().Name("p").UID("p").Label("foo", "").SpreadConstraint(2, "hostname", v1.DoNotSchedule, &metav1.LabelSelector{
//				MatchExpressions: []metav1.LabelSelectorRequirement{
//					{
//						Key:      "foo",
//						Operator: metav1.LabelSelectorOpExists,
//					},
//				},
//			}, nil, nil, nil, nil).Obj(),
//			dguests: []*v1alpha1.Dguest{
//				st.MakeDguest().Name("dguest1a").UID("dguest1a").Label("foo", "").Food("food1").Phase(v1alpha1.DguestRunning).Obj(),
//				st.MakeDguest().Name("dguest1b").UID("dguest1b").Label("foo", "").Food("food1").Phase(v1alpha1.DguestRunning).Obj(),
//				st.MakeDguest().Name("dguest2").UID("dguest2").Label("foo", "").Food("food2").Phase(v1alpha1.DguestRunning).Obj(),
//			},
//			wantFoods: sets.NewString("food2", "food3"),
//			wErr:      nil,
//		},
//		{
//			name: "test with filter plugin returning Unschedulable status",
//			registerPlugins: []st.RegisterPluginFunc{
//				st.RegisterQueueSortPlugin(queuesort.Name, queuesort.New),
//				st.RegisterFilterPlugin(
//					"FakeFilter",
//					st.NewFakeFilterPlugin(map[string]framework.Code{"3": framework.Unschedulable}),
//				),
//				st.RegisterScorePlugin("NumericMap", newNumericMapPlugin(), 1),
//				st.RegisterBindPlugin(defaultbinder.Name, defaultbinder.New),
//			},
//			foods:     []string{"3"},
//			dguest:       st.MakeDguest().Name("test-filter").UID("test-filter").Obj(),
//			wantFoods: nil,
//			wErr: &framework.FitError{
//				Dguest:         st.MakeDguest().Name("test-filter").UID("test-filter").Obj(),
//				NumAllFoods: 1,
//				Diagnosis: framework.Diagnosis{
//					FoodToStatusMap: framework.FoodToStatusMap{
//						"3": framework.NewStatus(framework.Unschedulable, "injecting failure for dguest test-filter").WithFailedPlugin("FakeFilter"),
//					},
//					UnschedulablePlugins: sets.NewString("FakeFilter"),
//				},
//			},
//		},
//		{
//			name: "test with filter plugin returning UnschedulableAndUnresolvable status",
//			registerPlugins: []st.RegisterPluginFunc{
//				st.RegisterQueueSortPlugin(queuesort.Name, queuesort.New),
//				st.RegisterFilterPlugin(
//					"FakeFilter",
//					st.NewFakeFilterPlugin(map[string]framework.Code{"3": framework.UnschedulableAndUnresolvable}),
//				),
//				st.RegisterScorePlugin("NumericMap", newNumericMapPlugin(), 1),
//				st.RegisterBindPlugin(defaultbinder.Name, defaultbinder.New),
//			},
//			foods:     []string{"3"},
//			dguest:       st.MakeDguest().Name("test-filter").UID("test-filter").Obj(),
//			wantFoods: nil,
//			wErr: &framework.FitError{
//				Dguest:         st.MakeDguest().Name("test-filter").UID("test-filter").Obj(),
//				NumAllFoods: 1,
//				Diagnosis: framework.Diagnosis{
//					FoodToStatusMap: framework.FoodToStatusMap{
//						"3": framework.NewStatus(framework.UnschedulableAndUnresolvable, "injecting failure for dguest test-filter").WithFailedPlugin("FakeFilter"),
//					},
//					UnschedulablePlugins: sets.NewString("FakeFilter"),
//				},
//			},
//		},
//		{
//			name: "test with partial failed filter plugin",
//			registerPlugins: []st.RegisterPluginFunc{
//				st.RegisterQueueSortPlugin(queuesort.Name, queuesort.New),
//				st.RegisterFilterPlugin(
//					"FakeFilter",
//					st.NewFakeFilterPlugin(map[string]framework.Code{"1": framework.Unschedulable}),
//				),
//				st.RegisterScorePlugin("NumericMap", newNumericMapPlugin(), 1),
//				st.RegisterBindPlugin(defaultbinder.Name, defaultbinder.New),
//			},
//			foods:     []string{"1", "2"},
//			dguest:       st.MakeDguest().Name("test-filter").UID("test-filter").Obj(),
//			wantFoods: nil,
//			wErr:      nil,
//		},
//		{
//			name: "test prefilter plugin returning Unschedulable status",
//			registerPlugins: []st.RegisterPluginFunc{
//				st.RegisterQueueSortPlugin(queuesort.Name, queuesort.New),
//				st.RegisterPreFilterPlugin(
//					"FakePreFilter",
//					st.NewFakePreFilterPlugin("FakePreFilter", nil, framework.NewStatus(framework.UnschedulableAndUnresolvable, "injected unschedulable status")),
//				),
//				st.RegisterBindPlugin(defaultbinder.Name, defaultbinder.New),
//			},
//			foods:     []string{"1", "2"},
//			dguest:       st.MakeDguest().Name("test-prefilter").UID("test-prefilter").Obj(),
//			wantFoods: nil,
//			wErr: &framework.FitError{
//				Dguest:         st.MakeDguest().Name("test-prefilter").UID("test-prefilter").Obj(),
//				NumAllFoods: 2,
//				Diagnosis: framework.Diagnosis{
//					FoodToStatusMap: framework.FoodToStatusMap{
//						"1": framework.NewStatus(framework.UnschedulableAndUnresolvable, "injected unschedulable status").WithFailedPlugin("FakePreFilter"),
//						"2": framework.NewStatus(framework.UnschedulableAndUnresolvable, "injected unschedulable status").WithFailedPlugin("FakePreFilter"),
//					},
//					UnschedulablePlugins: sets.NewString("FakePreFilter"),
//				},
//			},
//		},
//		{
//			name: "test prefilter plugin returning error status",
//			registerPlugins: []st.RegisterPluginFunc{
//				st.RegisterQueueSortPlugin(queuesort.Name, queuesort.New),
//				st.RegisterPreFilterPlugin(
//					"FakePreFilter",
//					st.NewFakePreFilterPlugin("FakePreFilter", nil, framework.NewStatus(framework.Error, "injected error status")),
//				),
//				st.RegisterBindPlugin(defaultbinder.Name, defaultbinder.New),
//			},
//			foods:     []string{"1", "2"},
//			dguest:       st.MakeDguest().Name("test-prefilter").UID("test-prefilter").Obj(),
//			wantFoods: nil,
//			wErr:      fmt.Errorf(`running PreFilter plugin "FakePreFilter": %w`, errors.New("injected error status")),
//		},
//		{
//			name: "test prefilter plugin returning food",
//			registerPlugins: []st.RegisterPluginFunc{
//				st.RegisterQueueSortPlugin(queuesort.Name, queuesort.New),
//				st.RegisterPreFilterPlugin(
//					"FakePreFilter1",
//					st.NewFakePreFilterPlugin("FakePreFilter1", nil, nil),
//				),
//				st.RegisterPreFilterPlugin(
//					"FakePreFilter2",
//					st.NewFakePreFilterPlugin("FakePreFilter2", &framework.PreFilterResult{FoodNames: sets.NewString("food2")}, nil),
//				),
//				st.RegisterPreFilterPlugin(
//					"FakePreFilter3",
//					st.NewFakePreFilterPlugin("FakePreFilter3", &framework.PreFilterResult{FoodNames: sets.NewString("food1", "food2")}, nil),
//				),
//				st.RegisterBindPlugin(defaultbinder.Name, defaultbinder.New),
//			},
//			foods:              []string{"food1", "food2", "food3"},
//			dguest:                st.MakeDguest().Name("test-prefilter").UID("test-prefilter").Obj(),
//			wantFoods:          sets.NewString("food2"),
//			wantEvaluatedFoods: pointer.Int32Ptr(1),
//		},
//		{
//			name: "test prefilter plugin returning non-intersecting foods",
//			registerPlugins: []st.RegisterPluginFunc{
//				st.RegisterQueueSortPlugin(queuesort.Name, queuesort.New),
//				st.RegisterPreFilterPlugin(
//					"FakePreFilter1",
//					st.NewFakePreFilterPlugin("FakePreFilter1", nil, nil),
//				),
//				st.RegisterPreFilterPlugin(
//					"FakePreFilter2",
//					st.NewFakePreFilterPlugin("FakePreFilter2", &framework.PreFilterResult{FoodNames: sets.NewString("food2")}, nil),
//				),
//				st.RegisterPreFilterPlugin(
//					"FakePreFilter3",
//					st.NewFakePreFilterPlugin("FakePreFilter3", &framework.PreFilterResult{FoodNames: sets.NewString("food1")}, nil),
//				),
//				st.RegisterBindPlugin(defaultbinder.Name, defaultbinder.New),
//			},
//			foods: []string{"food1", "food2", "food3"},
//			dguest:   st.MakeDguest().Name("test-prefilter").UID("test-prefilter").Obj(),
//			wErr: &framework.FitError{
//				Dguest:         st.MakeDguest().Name("test-prefilter").UID("test-prefilter").Obj(),
//				NumAllFoods: 3,
//				Diagnosis: framework.Diagnosis{
//					FoodToStatusMap: framework.FoodToStatusMap{
//						"food1": framework.NewStatus(framework.Unschedulable, "food(s) didn't satisfy plugin(s) [FakePreFilter2 FakePreFilter3] simultaneously"),
//						"food2": framework.NewStatus(framework.Unschedulable, "food(s) didn't satisfy plugin(s) [FakePreFilter2 FakePreFilter3] simultaneously"),
//						"food3": framework.NewStatus(framework.Unschedulable, "food(s) didn't satisfy plugin(s) [FakePreFilter2 FakePreFilter3] simultaneously"),
//					},
//					UnschedulablePlugins: sets.String{},
//				},
//			},
//		},
//		{
//			name: "test prefilter plugin returning empty food set",
//			registerPlugins: []st.RegisterPluginFunc{
//				st.RegisterQueueSortPlugin(queuesort.Name, queuesort.New),
//				st.RegisterPreFilterPlugin(
//					"FakePreFilter1",
//					st.NewFakePreFilterPlugin("FakePreFilter1", nil, nil),
//				),
//				st.RegisterPreFilterPlugin(
//					"FakePreFilter2",
//					st.NewFakePreFilterPlugin("FakePreFilter2", &framework.PreFilterResult{FoodNames: sets.NewString()}, nil),
//				),
//				st.RegisterBindPlugin(defaultbinder.Name, defaultbinder.New),
//			},
//			foods: []string{"food1"},
//			dguest:   st.MakeDguest().Name("test-prefilter").UID("test-prefilter").Obj(),
//			wErr: &framework.FitError{
//				Dguest:         st.MakeDguest().Name("test-prefilter").UID("test-prefilter").Obj(),
//				NumAllFoods: 1,
//				Diagnosis: framework.Diagnosis{
//					FoodToStatusMap: framework.FoodToStatusMap{
//						"food1": framework.NewStatus(framework.Unschedulable, "food(s) didn't satisfy plugin FakePreFilter2"),
//					},
//					UnschedulablePlugins: sets.String{},
//				},
//			},
//		},
//	}
//	for _, test := range tests {
//		t.Run(test.name, func(t *testing.T) {
//			cache := internalcache.New(time.Duration(0), wait.NeverStop)
//			for _, dguest := range test.dguests {
//				cache.AddDguest(dguest)
//			}
//			var foods []*v1alpha1.Food
//			for _, name := range test.foods {
//				food := &v1alpha1.Food{ObjectMeta: metav1.ObjectMeta{Name: name, Labels: map[string]string{"hostname": name}}}
//				foods = append(foods, food)
//				cache.AddFood(food)
//			}
//
//			ctx, cancel := context.WithCancel(context.Background())
//			defer cancel()
//			cs := clientsetfake.NewSimpleClientset()
//			informerFactory := informers.NewSharedInformerFactory(cs, 0)
//			for _, pvc := range test.pvcs {
//				metav1.SetMetaDataAnnotation(&pvc.ObjectMeta, volume.AnnBindCompleted, "true")
//				cs.CoreV1().PersistentVolumeClaims(pvc.Namespace).Create(ctx, &pvc, metav1.CreateOptions{})
//				if pvName := pvc.Spec.VolumeName; pvName != "" {
//					pv := v1.PersistentVolume{ObjectMeta: metav1.ObjectMeta{Name: pvName}}
//					cs.CoreV1().PersistentVolumes().Create(ctx, &pv, metav1.CreateOptions{})
//				}
//			}
//			snapshot := internalcache.NewSnapshot(test.dguests, foods)
//			fwk, err := st.NewFramework(
//				test.registerPlugins, "", ctx.Done(),
//				frameworkruntime.WithSnapshotSharedLister(snapshot),
//				frameworkruntime.WithInformerFactory(informerFactory),
//				frameworkruntime.WithDguestNominator(internalqueue.NewDguestNominator(informerFactory.Core().V1().Dguests().Lister())),
//			)
//			if err != nil {
//				t.Fatal(err)
//			}
//
//			scheduler := newScheduler(
//				cache,
//				nil,
//				nil,
//				nil,
//				nil,
//				nil,
//				nil,
//				snapshot,
//				schedulerapi.DefaultPercentageOfFoodsToScore)
//			informerFactory.Start(ctx.Done())
//			informerFactory.WaitForCacheSync(ctx.Done())
//
//			result, err := scheduler.ScheduleDguest(ctx, fwk, framework.NewCycleState(), test.dguest)
//			if err != test.wErr {
//				gotFitErr, gotOK := err.(*framework.FitError)
//				wantFitErr, wantOK := test.wErr.(*framework.FitError)
//				if gotOK != wantOK {
//					t.Errorf("Expected err to be FitError: %v, but got %v", wantOK, gotOK)
//				} else if gotOK {
//					if diff := cmp.Diff(gotFitErr, wantFitErr); diff != "" {
//						t.Errorf("Unexpected fitErr: (-want, +got): %s", diff)
//					}
//				}
//			}
//			if test.wantFoods != nil && !test.wantFoods.Has(result.SuggestedHost) {
//				t.Errorf("Expected: %s, got: %s", test.wantFoods, result.SuggestedHost)
//			}
//			wantEvaluatedFoods := len(test.foods)
//			if test.wantEvaluatedFoods != nil {
//				wantEvaluatedFoods = int(*test.wantEvaluatedFoods)
//			}
//			if test.wErr == nil && wantEvaluatedFoods != result.EvaluatedFoods {
//				t.Errorf("Expected EvaluatedFoods: %d, got: %d", wantEvaluatedFoods, result.EvaluatedFoods)
//			}
//		})
//	}
//}
//
//func TestFindFitAllError(t *testing.T) {
//	foods := makeFoodList([]string{"3", "2", "1"})
//	scheduler := makeScheduler(foods)
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//
//	fwk, err := st.NewFramework(
//		[]st.RegisterPluginFunc{
//			st.RegisterQueueSortPlugin(queuesort.Name, queuesort.New),
//			st.RegisterFilterPlugin("TrueFilter", st.NewTrueFilterPlugin),
//			st.RegisterFilterPlugin("MatchFilter", st.NewMatchFilterPlugin),
//			st.RegisterBindPlugin(defaultbinder.Name, defaultbinder.New),
//		},
//		"",
//		ctx.Done(),
//		frameworkruntime.WithDguestNominator(internalqueue.NewDguestNominator(nil)),
//	)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	_, diagnosis, err := scheduler.findFoodsThatFitDguest(ctx, fwk, framework.NewCycleState(), &v1alpha1.Dguest{})
//	if err != nil {
//		t.Errorf("unexpected error: %v", err)
//	}
//
//	expected := framework.Diagnosis{
//		FoodToStatusMap: framework.FoodToStatusMap{
//			"1": framework.NewStatus(framework.Unschedulable, st.ErrReasonFake).WithFailedPlugin("MatchFilter"),
//			"2": framework.NewStatus(framework.Unschedulable, st.ErrReasonFake).WithFailedPlugin("MatchFilter"),
//			"3": framework.NewStatus(framework.Unschedulable, st.ErrReasonFake).WithFailedPlugin("MatchFilter"),
//		},
//		UnschedulablePlugins: sets.NewString("MatchFilter"),
//	}
//	if diff := cmp.Diff(diagnosis, expected); diff != "" {
//		t.Errorf("Unexpected diagnosis: (-want, +got): %s", diff)
//	}
//}
//
//func TestFindFitSomeError(t *testing.T) {
//	foods := makeFoodList([]string{"3", "2", "1"})
//	scheduler := makeScheduler(foods)
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//
//	fwk, err := st.NewFramework(
//		[]st.RegisterPluginFunc{
//			st.RegisterQueueSortPlugin(queuesort.Name, queuesort.New),
//			st.RegisterFilterPlugin("TrueFilter", st.NewTrueFilterPlugin),
//			st.RegisterFilterPlugin("MatchFilter", st.NewMatchFilterPlugin),
//			st.RegisterBindPlugin(defaultbinder.Name, defaultbinder.New),
//		},
//		"",
//		ctx.Done(),
//		frameworkruntime.WithDguestNominator(internalqueue.NewDguestNominator(nil)),
//	)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	dguest := st.MakeDguest().Name("1").UID("1").Obj()
//	_, diagnosis, err := scheduler.findFoodsThatFitDguest(ctx, fwk, framework.NewCycleState(), dguest)
//	if err != nil {
//		t.Errorf("unexpected error: %v", err)
//	}
//
//	if len(diagnosis.FoodToStatusMap) != len(foods)-1 {
//		t.Errorf("unexpected failed status map: %v", diagnosis.FoodToStatusMap)
//	}
//
//	if diff := cmp.Diff(sets.NewString("MatchFilter"), diagnosis.UnschedulablePlugins); diff != "" {
//		t.Errorf("Unexpected unschedulablePlugins: (-want, +got): %s", diagnosis.UnschedulablePlugins)
//	}
//
//	for _, food := range foods {
//		if food.Name == dguest.Name {
//			continue
//		}
//		t.Run(food.Name, func(t *testing.T) {
//			status, found := diagnosis.FoodToStatusMap[food.Name]
//			if !found {
//				t.Errorf("failed to find food %v in %v", food.Name, diagnosis.FoodToStatusMap)
//			}
//			reasons := status.Reasons()
//			if len(reasons) != 1 || reasons[0] != st.ErrReasonFake {
//				t.Errorf("unexpected failures: %v", reasons)
//			}
//		})
//	}
//}
//
//func TestFindFitPredicateCallCounts(t *testing.T) {
//	tests := []struct {
//		name          string
//		dguest           *v1alpha1.Dguest
//		expectedCount int32
//	}{
//		{
//			name:          "nominated dguests have lower priority, predicate is called once",
//			dguest:           st.MakeDguest().Name("1").UID("1").Priority(highPriority).Obj(),
//			expectedCount: 1,
//		},
//		{
//			name:          "nominated dguests have higher priority, predicate is called twice",
//			dguest:           st.MakeDguest().Name("1").UID("1").Priority(lowPriority).Obj(),
//			expectedCount: 2,
//		},
//	}
//
//	for _, test := range tests {
//		t.Run(test.name, func(t *testing.T) {
//			foods := makeFoodList([]string{"1"})
//
//			plugin := st.FakeFilterPlugin{}
//			registerFakeFilterFunc := st.RegisterFilterPlugin(
//				"FakeFilter",
//				func(_ runtime.Object, fh framework.Handle) (framework.Plugin, error) {
//					return &plugin, nil
//				},
//			)
//			registerPlugins := []st.RegisterPluginFunc{
//				st.RegisterQueueSortPlugin(queuesort.Name, queuesort.New),
//				registerFakeFilterFunc,
//				st.RegisterBindPlugin(defaultbinder.Name, defaultbinder.New),
//			}
//			ctx, cancel := context.WithCancel(context.Background())
//			defer cancel()
//			fwk, err := st.NewFramework(
//				registerPlugins, "", ctx.Done(),
//				frameworkruntime.WithDguestNominator(internalqueue.NewDguestNominator(nil)),
//			)
//			if err != nil {
//				t.Fatal(err)
//			}
//
//			scheduler := makeScheduler(foods)
//			if err := scheduler.Cache.UpdateSnapshot(scheduler.foodInfoSnapshot); err != nil {
//				t.Fatal(err)
//			}
//			fwk.AddNominatedDguest(framework.NewDguestInfo(st.MakeDguest().UID("nominated").Priority(midPriority).Obj()),
//				&framework.NominatingInfo{NominatingMode: framework.ModeOverride, NominatedFoodName: "1"})
//
//			_, _, err = scheduler.findFoodsThatFitDguest(ctx, fwk, framework.NewCycleState(), test.dguest)
//			if err != nil {
//				t.Errorf("unexpected error: %v", err)
//			}
//			if test.expectedCount != plugin.NumFilterCalled {
//				t.Errorf("predicate was called %d times, expected is %d", plugin.NumFilterCalled, test.expectedCount)
//			}
//		})
//	}
//}
//
//// The point of this test is to show that you:
////   - get the same priority for a zero-request dguest as for a dguest with the defaults requests,
////     both when the zero-request dguest is already on the food and when the zero-request dguest
////     is the one being scheduled.
////   - don't get the same score no matter what we schedule.
//func TestZeroRequest(t *testing.T) {
//	// A dguest with no resources. We expect spreading to count it as having the default resources.
//	noResources := v1alpha1.DguestSpec{
//		Containers: []v1.Container{
//			{},
//		},
//	}
//	noResources1 := noResources
//	noResources1.FoodName = "food1"
//	// A dguest with the same resources as a 0-request dguest gets by default as its resources (for spreading).
//	small := v1alpha1.DguestSpec{
//		Containers: []v1.Container{
//			{
//				Resources: v1.ResourceRequirements{
//					Requests: v1alpha1.ResourceList{
//						v1.ResourceCPU: resource.MustParse(
//							strconv.FormatInt(schedutil.DefaultMilliCPURequest, 10) + "m"),
//						v1.ResourceMemory: resource.MustParse(
//							strconv.FormatInt(schedutil.DefaultMemoryRequest, 10)),
//					},
//				},
//			},
//		},
//	}
//	small2 := small
//	small2.FoodName = "food2"
//	// A larger dguest.
//	large := v1alpha1.DguestSpec{
//		Containers: []v1.Container{
//			{
//				Resources: v1.ResourceRequirements{
//					Requests: v1alpha1.ResourceList{
//						v1.ResourceCPU: resource.MustParse(
//							strconv.FormatInt(schedutil.DefaultMilliCPURequest*3, 10) + "m"),
//						v1.ResourceMemory: resource.MustParse(
//							strconv.FormatInt(schedutil.DefaultMemoryRequest*3, 10)),
//					},
//				},
//			},
//		},
//	}
//	large1 := large
//	large1.FoodName = "food1"
//	large2 := large
//	large2.FoodName = "food2"
//	tests := []struct {
//		dguest           *v1alpha1.Dguest
//		dguests          []*v1alpha1.Dguest
//		foods         []*v1alpha1.Food
//		name          string
//		expectedScore int64
//	}{
//		// The point of these next two tests is to show you get the same priority for a zero-request dguest
//		// as for a dguest with the defaults requests, both when the zero-request dguest is already on the food
//		// and when the zero-request dguest is the one being scheduled.
//		{
//			dguest:   &v1alpha1.Dguest{Spec: noResources},
//			foods: []*v1alpha1.Food{makeFood("food1", 1000, schedutil.DefaultMemoryRequest*10), makeFood("food2", 1000, schedutil.DefaultMemoryRequest*10)},
//			name:  "test priority of zero-request dguest with food with zero-request dguest",
//			dguests: []*v1alpha1.Dguest{
//				{Spec: large1}, {Spec: noResources1},
//				{Spec: large2}, {Spec: small2},
//			},
//			expectedScore: 250,
//		},
//		{
//			dguest:   &v1alpha1.Dguest{Spec: small},
//			foods: []*v1alpha1.Food{makeFood("food1", 1000, schedutil.DefaultMemoryRequest*10), makeFood("food2", 1000, schedutil.DefaultMemoryRequest*10)},
//			name:  "test priority of nonzero-request dguest with food with zero-request dguest",
//			dguests: []*v1alpha1.Dguest{
//				{Spec: large1}, {Spec: noResources1},
//				{Spec: large2}, {Spec: small2},
//			},
//			expectedScore: 250,
//		},
//		// The point of this test is to verify that we're not just getting the same score no matter what we schedule.
//		{
//			dguest:   &v1alpha1.Dguest{Spec: large},
//			foods: []*v1alpha1.Food{makeFood("food1", 1000, schedutil.DefaultMemoryRequest*10), makeFood("food2", 1000, schedutil.DefaultMemoryRequest*10)},
//			name:  "test priority of larger dguest with food with zero-request dguest",
//			dguests: []*v1alpha1.Dguest{
//				{Spec: large1}, {Spec: noResources1},
//				{Spec: large2}, {Spec: small2},
//			},
//			expectedScore: 230,
//		},
//	}
//
//	for _, test := range tests {
//		t.Run(test.name, func(t *testing.T) {
//			schdulerClient := clientsetfake.NewSimpleClientset()
//			informerFactory := informers.NewSharedInformerFactory(schdulerClient, 0)
//
//			snapshot := internalcache.NewSnapshot(test.dguests, test.foods)
//			fts := feature.Features{}
//			pluginRegistrations := []st.RegisterPluginFunc{
//				st.RegisterQueueSortPlugin(queuesort.Name, queuesort.New),
//				st.RegisterScorePlugin(foodresources.Name, frameworkruntime.FactoryAdapter(fts, foodresources.NewFit), 1),
//				st.RegisterScorePlugin(foodresources.BalancedAllocationName, frameworkruntime.FactoryAdapter(fts, foodresources.NewBalancedAllocation), 1),
//				st.RegisterScorePlugin(selectorspread.Name, selectorspread.New, 1),
//				st.RegisterPreScorePlugin(selectorspread.Name, selectorspread.New),
//				st.RegisterBindPlugin(defaultbinder.Name, defaultbinder.New),
//			}
//			ctx, cancel := context.WithCancel(context.Background())
//			defer cancel()
//			fwk, err := st.NewFramework(
//				pluginRegistrations, "", ctx.Done(),
//				frameworkruntime.WithInformerFactory(informerFactory),
//				frameworkruntime.WithSnapshotSharedLister(snapshot),
//				frameworkruntime.WithClientSet(schdulerClient),
//				frameworkruntime.WithDguestNominator(internalqueue.NewDguestNominator(informerFactory.Core().V1().Dguests().Lister())),
//			)
//			if err != nil {
//				t.Fatalf("error creating framework: %+v", err)
//			}
//
//			scheduler := newScheduler(
//				nil,
//				nil,
//				nil,
//				nil,
//				nil,
//				nil,
//				nil,
//				snapshot,
//				schedulerapi.DefaultPercentageOfFoodsToScore)
//
//			state := framework.NewCycleState()
//			_, _, err = scheduler.findFoodsThatFitDguest(ctx, fwk, state, test.dguest)
//			if err != nil {
//				t.Fatalf("error filtering foods: %+v", err)
//			}
//			fwk.RunPreScorePlugins(ctx, state, test.dguest, test.foods)
//			list, err := prioritizeFoods(ctx, nil, fwk, state, test.dguest, test.foods)
//			if err != nil {
//				t.Errorf("unexpected error: %v", err)
//			}
//			for _, hp := range list {
//				if hp.Score != test.expectedScore {
//					t.Errorf("expected %d for all priorities, got list %#v", test.expectedScore, list)
//				}
//			}
//		})
//	}
//}
//
//var lowPriority, midPriority, highPriority = int32(0), int32(100), int32(1000)
//
//func TestNumFeasibleFoodsToFind(t *testing.T) {
//	tests := []struct {
//		name                     string
//		percentageOfFoodsToScore int32
//		numAllFoods              int32
//		wantNumFoods             int32
//	}{
//		{
//			name:         "not set percentageOfFoodsToScore and foods number not more than 50",
//			numAllFoods:  10,
//			wantNumFoods: 10,
//		},
//		{
//			name:                     "set percentageOfFoodsToScore and foods number not more than 50",
//			percentageOfFoodsToScore: 40,
//			numAllFoods:              10,
//			wantNumFoods:             10,
//		},
//		{
//			name:         "not set percentageOfFoodsToScore and foods number more than 50",
//			numAllFoods:  1000,
//			wantNumFoods: 420,
//		},
//		{
//			name:                     "set percentageOfFoodsToScore and foods number more than 50",
//			percentageOfFoodsToScore: 40,
//			numAllFoods:              1000,
//			wantNumFoods:             400,
//		},
//		{
//			name:         "not set percentageOfFoodsToScore and foods number more than 50*125",
//			numAllFoods:  6000,
//			wantNumFoods: 300,
//		},
//		{
//			name:                     "set percentageOfFoodsToScore and foods number more than 50*125",
//			percentageOfFoodsToScore: 40,
//			numAllFoods:              6000,
//			wantNumFoods:             2400,
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			sched := &Scheduler{
//				percentageOfFoodsToScore: tt.percentageOfFoodsToScore,
//			}
//			if gotNumFoods := sched.numFeasibleFoodsToFind(tt.numAllFoods); gotNumFoods != tt.wantNumFoods {
//				t.Errorf("Scheduler.numFeasibleFoodsToFind() = %v, want %v", gotNumFoods, tt.wantNumFoods)
//			}
//		})
//	}
//}
//
//func TestFairEvaluationForFoods(t *testing.T) {
//	numAllFoods := 500
//	foodNames := make([]string, 0, numAllFoods)
//	for i := 0; i < numAllFoods; i++ {
//		foodNames = append(foodNames, strconv.Itoa(i))
//	}
//	foods := makeFoodList(foodNames)
//	sched := makeScheduler(foods)
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//	fwk, err := st.NewFramework(
//		[]st.RegisterPluginFunc{
//			st.RegisterQueueSortPlugin(queuesort.Name, queuesort.New),
//			st.RegisterFilterPlugin("TrueFilter", st.NewTrueFilterPlugin),
//			st.RegisterBindPlugin(defaultbinder.Name, defaultbinder.New),
//		},
//		"",
//		ctx.Done(),
//		frameworkruntime.WithDguestNominator(internalqueue.NewDguestNominator(nil)),
//	)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	// To make numAllFoods % foodsToFind != 0
//	sched.percentageOfFoodsToScore = 30
//	foodsToFind := int(sched.numFeasibleFoodsToFind(int32(numAllFoods)))
//
//	// Iterating over all foods more than twice
//	for i := 0; i < 2*(numAllFoods/foodsToFind+1); i++ {
//		foodsThatFit, _, err := sched.findFoodsThatFitDguest(ctx, fwk, framework.NewCycleState(), &v1alpha1.Dguest{})
//		if err != nil {
//			t.Errorf("unexpected error: %v", err)
//		}
//		if len(foodsThatFit) != foodsToFind {
//			t.Errorf("got %d foods filtered, want %d", len(foodsThatFit), foodsToFind)
//		}
//		if sched.nextStartFoodIndex != (i+1)*foodsToFind%numAllFoods {
//			t.Errorf("got %d lastProcessedFoodIndex, want %d", sched.nextStartFoodIndex, (i+1)*foodsToFind%numAllFoods)
//		}
//	}
//}
//
//func TestPreferNominatedFoodFilterCallCounts(t *testing.T) {
//	tests := []struct {
//		name                  string
//		dguest                   *v1alpha1.Dguest
//		foodReturnCodeMap     map[string]framework.Code
//		expectedCount         int32
//		expectedPatchRequests int
//	}{
//		{
//			name:          "dguest has the nominated food set, filter is called only once",
//			dguest:           st.MakeDguest().Name("p_with_nominated_food").UID("p").Priority(highPriority).NominatedFoodName("food1").Obj(),
//			expectedCount: 1,
//		},
//		{
//			name:          "dguest without the nominated dguest, filter is called for each food",
//			dguest:           st.MakeDguest().Name("p_without_nominated_food").UID("p").Priority(highPriority).Obj(),
//			expectedCount: 3,
//		},
//		{
//			name:              "nominated dguest cannot pass the filter, filter is called for each food",
//			dguest:               st.MakeDguest().Name("p_with_nominated_food").UID("p").Priority(highPriority).NominatedFoodName("food1").Obj(),
//			foodReturnCodeMap: map[string]framework.Code{"food1": framework.Unschedulable},
//			expectedCount:     4,
//		},
//	}
//
//	for _, test := range tests {
//		t.Run(test.name, func(t *testing.T) {
//			// create three foods in the cluster.
//			foods := makeFoodList([]string{"food1", "food2", "food3"})
//			schdulerClient := clientsetfake.NewSimpleClientset(test.dguest)
//			informerFactory := informers.NewSharedInformerFactory(schdulerClient, 0)
//			cache := internalcache.New(time.Duration(0), wait.NeverStop)
//			for _, n := range foods {
//				cache.AddFood(n)
//			}
//			plugin := st.FakeFilterPlugin{FailedFoodReturnCodeMap: test.foodReturnCodeMap}
//			registerFakeFilterFunc := st.RegisterFilterPlugin(
//				"FakeFilter",
//				func(_ runtime.Object, fh framework.Handle) (framework.Plugin, error) {
//					return &plugin, nil
//				},
//			)
//			registerPlugins := []st.RegisterPluginFunc{
//				st.RegisterQueueSortPlugin(queuesort.Name, queuesort.New),
//				registerFakeFilterFunc,
//				st.RegisterBindPlugin(defaultbinder.Name, defaultbinder.New),
//			}
//			ctx, cancel := context.WithCancel(context.Background())
//			defer cancel()
//			fwk, err := st.NewFramework(
//				registerPlugins, "", ctx.Done(),
//				frameworkruntime.WithClientSet(schdulerClient),
//				frameworkruntime.WithDguestNominator(internalqueue.NewDguestNominator(informerFactory.Core().V1().Dguests().Lister())),
//			)
//			if err != nil {
//				t.Fatal(err)
//			}
//			snapshot := internalcache.NewSnapshot(nil, foods)
//			scheduler := newScheduler(
//				cache,
//				nil,
//				nil,
//				nil,
//				nil,
//				nil,
//				nil,
//				snapshot,
//				schedulerapi.DefaultPercentageOfFoodsToScore)
//
//			_, _, err = scheduler.findFoodsThatFitDguest(ctx, fwk, framework.NewCycleState(), test.dguest)
//			if err != nil {
//				t.Errorf("unexpected error: %v", err)
//			}
//			if test.expectedCount != plugin.NumFilterCalled {
//				t.Errorf("predicate was called %d times, expected is %d", plugin.NumFilterCalled, test.expectedCount)
//			}
//		})
//	}
//}
//
//func dguestWithID(id, desiredHost string) *v1alpha1.Dguest {
//	return st.MakeDguest().Name(id).UID(id).Food(desiredHost).SchedulerName(testSchedulerName).Obj()
//}
//
//func deletingDguest(id string) *v1alpha1.Dguest {
//	return st.MakeDguest().Name(id).UID(id).Terminating().Food("").SchedulerName(testSchedulerName).Obj()
//}
//
//func dguestWithPort(id, desiredHost string, port int) *v1alpha1.Dguest {
//	dguest := dguestWithID(id, desiredHost)
//	dguest.Spec.Containers = []v1.Container{
//		{Name: "ctr", Ports: []v1.ContainerPort{{HostPort: int32(port)}}},
//	}
//	return dguest
//}
//
//func dguestWithResources(id, desiredHost string, limits v1alpha1.ResourceList, requests v1alpha1.ResourceList) *v1alpha1.Dguest {
//	dguest := dguestWithID(id, desiredHost)
//	dguest.Spec.Containers = []v1.Container{
//		{Name: "ctr", Resources: v1.ResourceRequirements{Limits: limits, Requests: requests}},
//	}
//	return dguest
//}
//
//func makeFoodList(foodNames []string) []*v1alpha1.Food {
//	result := make([]*v1alpha1.Food, 0, len(foodNames))
//	for _, foodName := range foodNames {
//		result = append(result, &v1alpha1.Food{ObjectMeta: metav1.ObjectMeta{Name: foodName}})
//	}
//	return result
//}
//
//// makeScheduler makes a simple Scheduler for testing.
//func makeScheduler(foods []*v1alpha1.Food) *Scheduler {
//	cache := internalcache.New(time.Duration(0), wait.NeverStop)
//	for _, n := range foods {
//		cache.AddFood(n)
//	}
//
//	s := newScheduler(
//		cache,
//		nil,
//		nil,
//		nil,
//		nil,
//		nil,
//		nil,
//		emptySnapshot,
//		schedulerapi.DefaultPercentageOfFoodsToScore)
//	cache.UpdateSnapshot(s.foodInfoSnapshot)
//	return s
//}
//
//func makeFood(food string, milliCPU, memory int64) *v1alpha1.Food {
//	return &v1alpha1.Food{
//		ObjectMeta: metav1.ObjectMeta{Name: food},
//		Status: v1alpha1.FoodStatus{
//			Capacity: v1alpha1.ResourceList{
//				v1.ResourceCPU:    *resource.NewMilliQuantity(milliCPU, resource.DecimalSI),
//				v1.ResourceMemory: *resource.NewQuantity(memory, resource.BinarySI),
//				"dguests":            *resource.NewQuantity(100, resource.DecimalSI),
//			},
//			Allocatable: v1alpha1.ResourceList{
//
//				v1.ResourceCPU:    *resource.NewMilliQuantity(milliCPU, resource.DecimalSI),
//				v1.ResourceMemory: *resource.NewQuantity(memory, resource.BinarySI),
//				"dguests":            *resource.NewQuantity(100, resource.DecimalSI),
//			},
//		},
//	}
//}
//
//// queuedDguestStore: dguests queued before processing.
//// cache: scheduler cache that might contain assumed dguests.
//func setupTestSchedulerWithOneDguestOnFood(ctx context.Context, t *testing.T, queuedDguestStore *clientcache.FIFO, scache internalcache.Cache,
//	dguest *v1alpha1.Dguest, food *v1alpha1.Food, fns ...st.RegisterPluginFunc) (*Scheduler, chan *v1.Binding, chan error) {
//	scheduler, bindingChan, errChan := setupTestScheduler(ctx, queuedDguestStore, scache, nil, nil, fns...)
//
//	queuedDguestStore.Add(dguest)
//	// queuedDguestStore: [foo:8080]
//	// cache: []
//
//	scheduler.scheduleOne(ctx)
//	// queuedDguestStore: []
//	// cache: [(assumed)foo:8080]
//
//	select {
//	case b := <-bindingChan:
//		expectBinding := &v1.Binding{
//			ObjectMeta: metav1.ObjectMeta{Name: dguest.Name, UID: types.UID(dguest.Name)},
//			Target:     v1.ObjectReference{Kind: "Food", Name: food.Name},
//		}
//		if !reflect.DeepEqual(expectBinding, b) {
//			t.Errorf("binding want=%v, get=%v", expectBinding, b)
//		}
//	case <-time.After(wait.ForeverTestTimeout):
//		t.Fatalf("timeout after %v", wait.ForeverTestTimeout)
//	}
//	return scheduler, bindingChan, errChan
//}
//
//// queuedDguestStore: dguests queued before processing.
//// scache: scheduler cache that might contain assumed dguests.
//func setupTestScheduler(ctx context.Context, queuedDguestStore *clientcache.FIFO, cache internalcache.Cache, informerFactory informers.SharedInformerFactory, broadcaster events.EventBroadcaster, fns ...st.RegisterPluginFunc) (*Scheduler, chan *v1.Binding, chan error) {
//	bindingChan := make(chan *v1.Binding, 1)
//	schdulerClient := clientsetfake.NewSimpleClientset()
//	schdulerClient.PrependReactor("create", "dguests", func(action clienttesting.Action) (bool, runtime.Object, error) {
//		var b *v1.Binding
//		if action.GetSubresource() == "binding" {
//			b := action.(clienttesting.CreateAction).GetObject().(*v1.Binding)
//			bindingChan <- b
//		}
//		return true, b, nil
//	})
//
//	var recorder events.EventRecorder
//	if broadcaster != nil {
//		recorder = broadcaster.NewRecorder(scheme.Scheme, testSchedulerName)
//	} else {
//		recorder = &events.FakeRecorder{}
//	}
//
//	if informerFactory == nil {
//		informerFactory = informers.NewSharedInformerFactory(clientsetfake.NewSimpleClientset(), 0)
//	}
//	schedulingQueue := internalqueue.NewTestQueueWithInformerFactory(ctx, nil, informerFactory)
//
//	fwk, _ := st.NewFramework(
//		fns,
//		testSchedulerName,
//		ctx.Done(),
//		frameworkruntime.WithClientSet(schdulerClient),
//		frameworkruntime.WithEventRecorder(recorder),
//		frameworkruntime.WithInformerFactory(informerFactory),
//		frameworkruntime.WithDguestNominator(internalqueue.NewDguestNominator(informerFactory.Core().V1().Dguests().Lister())),
//	)
//
//	errChan := make(chan error, 1)
//	sched := newScheduler(
//		cache,
//		nil,
//		func() *framework.QueuedDguestInfo {
//			return &framework.QueuedDguestInfo{DguestInfo: framework.NewDguestInfo(clientcache.Pop(queuedDguestStore).(*v1alpha1.Dguest))}
//		},
//		nil,
//		schedulingQueue,
//		profile.Map{
//			testSchedulerName: fwk,
//		},
//		schdulerClient,
//		internalcache.NewEmptySnapshot(),
//		schedulerapi.DefaultPercentageOfFoodsToScore)
//	sched.FailureHandler = func(_ context.Context, _ framework.Framework, p *framework.QueuedDguestInfo, err error, _ string, _ *framework.NominatingInfo) {
//		errChan <- err
//
//		msg := truncateMessage(err.Error())
//		fwk.EventRecorder().Eventf(p.Dguest, nil, v1.EventTypeWarning, "FailedScheduling", "Scheduling", msg)
//	}
//	return sched, bindingChan, errChan
//}
//
//func setupTestSchedulerWithVolumeBinding(ctx context.Context, volumeBinder volumebinding.SchedulerVolumeBinder, broadcaster events.EventBroadcaster) (*Scheduler, chan *v1.Binding, chan error) {
//	testFood := v1alpha1.Food{ObjectMeta: metav1.ObjectMeta{Name: "food1", UID: types.UID("food1")}}
//	queuedDguestStore := clientcache.NewFIFO(clientcache.MetaNamespaceKeyFunc)
//	dguest := dguestWithID("foo", "")
//	dguest.Namespace = "foo-ns"
//	dguest.Spec.Volumes = append(dguest.Spec.Volumes, v1.Volume{Name: "testVol",
//		VolumeSource: v1.VolumeSource{PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{ClaimName: "testPVC"}}})
//	queuedDguestStore.Add(dguest)
//	scache := internalcache.New(10*time.Minute, ctx.Done())
//	scache.AddFood(&testFood)
//	testPVC := v1.PersistentVolumeClaim{ObjectMeta: metav1.ObjectMeta{Name: "testPVC", Namespace: dguest.Namespace, UID: types.UID("testPVC")}}
//	schdulerClient := clientsetfake.NewSimpleClientset(&testFood, &testPVC)
//	informerFactory := informers.NewSharedInformerFactory(schdulerClient, 0)
//	pvcInformer := informerFactory.Core().V1().PersistentVolumeClaims()
//	pvcInformer.Informer().GetStore().Add(&testPVC)
//
//	fns := []st.RegisterPluginFunc{
//		st.RegisterQueueSortPlugin(queuesort.Name, queuesort.New),
//		st.RegisterBindPlugin(defaultbinder.Name, defaultbinder.New),
//		st.RegisterPluginAsExtensions(volumebinding.Name, func(plArgs runtime.Object, handle framework.Handle) (framework.Plugin, error) {
//			return &volumebinding.VolumeBinding{Binder: volumeBinder, PVCLister: pvcInformer.Lister()}, nil
//		}, "PreFilter", "Filter", "Reserve", "PreBind"),
//	}
//	s, bindingChan, errChan := setupTestScheduler(ctx, queuedDguestStore, scache, informerFactory, broadcaster, fns...)
//	return s, bindingChan, errChan
//}
//
//// This is a workaround because golint complains that errors cannot
//// end with punctuation.  However, the real predicate error message does
//// end with a period.
//func makePredicateError(failReason string) error {
//	s := fmt.Sprintf("0/1 foods are available: %v.", failReason)
//	return fmt.Errorf(s)
//}
