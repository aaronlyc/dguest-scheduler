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
//	"dguest-scheduler/pkg/apis/scheduler/v1alpha1"
//	v12 "dguest-scheduler/pkg/scheduler/apis/config/v1"
//	"fmt"
//	"sort"
//	"strings"
//	"testing"
//	"time"
//
//	schedulerapi "dguest-scheduler/pkg/scheduler/apis/config"
//	"dguest-scheduler/pkg/scheduler/framework"
//	"dguest-scheduler/pkg/scheduler/framework/plugins/defaultbinder"
//	"dguest-scheduler/pkg/scheduler/framework/plugins/queuesort"
//	frameworkruntime "dguest-scheduler/pkg/scheduler/framework/runtime"
//	internalcache "dguest-scheduler/pkg/scheduler/internal/cache"
//	internalqueue "dguest-scheduler/pkg/scheduler/internal/queue"
//	"dguest-scheduler/pkg/scheduler/profile"
//	"github.com/google/go-cmp/cmp"
//	v1 "k8s.io/api/core/v1"
//	apierrors "k8s.io/apimachinery/pkg/api/errors"
//	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
//	"k8s.io/apimachinery/pkg/runtime"
//	"k8s.io/apimachinery/pkg/util/sets"
//	"k8s.io/client-go/informers"
//	"k8s.io/client-go/kubernetes"
//	"k8s.io/client-go/kubernetes/fake"
//	"k8s.io/client-go/kubernetes/scheme"
//	"k8s.io/client-go/tools/cache"
//	"k8s.io/client-go/tools/events"
//	testingclock "k8s.io/utils/clock/testing"
//)
//
//func TestSchedulerCreation(t *testing.T) {
//	invalidRegistry := map[string]frameworkruntime.PluginFactory{
//		defaultbinder.Name: defaultbinder.New,
//	}
//	validRegistry := map[string]frameworkruntime.PluginFactory{
//		"Foo": defaultbinder.New,
//	}
//	cases := []struct {
//		name          string
//		opts          []Option
//		wantErr       string
//		wantProfiles  []string
//		wantExtenders []string
//	}{
//		{
//			name: "valid out-of-tree registry",
//			opts: []Option{
//				WithFrameworkOutOfTreeRegistry(validRegistry),
//				WithProfiles(
//					v12.SchedulerProfile{
//						SchedulerName: "default-scheduler",
//						Plugins: &v12.Plugins{
//							QueueSort: v12.PluginSet{Enabled: []v12.Plugin{{Name: "PrioritySort"}}},
//							Bind:      v12.PluginSet{Enabled: []v12.Plugin{{Name: "DefaultBinder"}}},
//						},
//					},
//				)},
//			wantProfiles: []string{"default-scheduler"},
//		},
//		{
//			name: "repeated plugin name in out-of-tree plugin",
//			opts: []Option{
//				WithFrameworkOutOfTreeRegistry(invalidRegistry),
//				WithProfiles(
//					v12.SchedulerProfile{
//						SchedulerName: "default-scheduler",
//						Plugins: &v12.Plugins{
//							QueueSort: v12.PluginSet{Enabled: []v12.Plugin{{Name: "PrioritySort"}}},
//							Bind:      v12.PluginSet{Enabled: []v12.Plugin{{Name: "DefaultBinder"}}},
//						},
//					},
//				)},
//			wantProfiles: []string{"default-scheduler"},
//			wantErr:      "a plugin named DefaultBinder already exists",
//		},
//		{
//			name: "multiple profiles",
//			opts: []Option{
//				WithProfiles(
//					v12.SchedulerProfile{
//						SchedulerName: "foo",
//						Plugins: &v12.Plugins{
//							QueueSort: v12.PluginSet{Enabled: []v12.Plugin{{Name: "PrioritySort"}}},
//							Bind:      v12.PluginSet{Enabled: []v12.Plugin{{Name: "DefaultBinder"}}},
//						},
//					},
//					v12.SchedulerProfile{
//						SchedulerName: "bar",
//						Plugins: &v12.Plugins{
//							QueueSort: v12.PluginSet{Enabled: []v12.Plugin{{Name: "PrioritySort"}}},
//							Bind:      v12.PluginSet{Enabled: []v12.Plugin{{Name: "DefaultBinder"}}},
//						},
//					},
//				)},
//			wantProfiles: []string{"bar", "foo"},
//		},
//		{
//			name: "Repeated profiles",
//			opts: []Option{
//				WithProfiles(
//					v12.SchedulerProfile{
//						SchedulerName: "foo",
//						Plugins: &v12.Plugins{
//							QueueSort: v12.PluginSet{Enabled: []v12.Plugin{{Name: "PrioritySort"}}},
//							Bind:      v12.PluginSet{Enabled: []v12.Plugin{{Name: "DefaultBinder"}}},
//						},
//					},
//					v12.SchedulerProfile{
//						SchedulerName: "bar",
//						Plugins: &v12.Plugins{
//							QueueSort: v12.PluginSet{Enabled: []v12.Plugin{{Name: "PrioritySort"}}},
//							Bind:      v12.PluginSet{Enabled: []v12.Plugin{{Name: "DefaultBinder"}}},
//						},
//					},
//					v12.SchedulerProfile{
//						SchedulerName: "foo",
//						Plugins: &v12.Plugins{
//							QueueSort: v12.PluginSet{Enabled: []v12.Plugin{{Name: "PrioritySort"}}},
//							Bind:      v12.PluginSet{Enabled: []v12.Plugin{{Name: "DefaultBinder"}}},
//						},
//					},
//				)},
//			wantErr: "duplicate profile with scheduler name \"foo\"",
//		},
//		{
//			name: "With extenders",
//			opts: []Option{
//				WithProfiles(
//					v12.SchedulerProfile{
//						SchedulerName: "default-scheduler",
//						Plugins: &v12.Plugins{
//							QueueSort: v12.PluginSet{Enabled: []v12.Plugin{{Name: "PrioritySort"}}},
//							Bind:      v12.PluginSet{Enabled: []v12.Plugin{{Name: "DefaultBinder"}}},
//						},
//					},
//				),
//				WithExtenders(
//					schedulerapi.Extender{
//						URLPrefix: "http://extender.kube-system/",
//					},
//				),
//			},
//			wantProfiles:  []string{"default-scheduler"},
//			wantExtenders: []string{"http://extender.kube-system/"},
//		},
//	}
//
//	for _, tc := range cases {
//		t.Run(tc.name, func(t *testing.T) {
//			client := fake.NewSimpleClientset()
//			informerFactory := informers.NewSharedInformerFactory(client, 0)
//
//			eventBroadcaster := events.NewBroadcaster(&events.EventSinkImpl{Interface: client.EventsV1()})
//
//			stopCh := make(chan struct{})
//			defer close(stopCh)
//			s, err := New(
//				client,
//				informerFactory,
//				nil,
//				profile.NewRecorderFactory(eventBroadcaster),
//				stopCh,
//				tc.opts...,
//			)
//
//			// Errors
//			if len(tc.wantErr) != 0 {
//				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
//					t.Errorf("got error %q, want %q", err, tc.wantErr)
//				}
//				return
//			}
//			if err != nil {
//				t.Fatalf("Failed to create scheduler: %v", err)
//			}
//
//			// Profiles
//			profiles := make([]string, 0, len(s.Profiles))
//			for name := range s.Profiles {
//				profiles = append(profiles, name)
//			}
//			sort.Strings(profiles)
//			if diff := cmp.Diff(tc.wantProfiles, profiles); diff != "" {
//				t.Errorf("unexpected profiles (-want, +got):\n%s", diff)
//			}
//
//			// Extenders
//			if len(tc.wantExtenders) != 0 {
//				// Scheduler.Extenders
//				extenders := make([]string, 0, len(s.Extenders))
//				for _, e := range s.Extenders {
//					extenders = append(extenders, e.Name())
//				}
//				if diff := cmp.Diff(tc.wantExtenders, extenders); diff != "" {
//					t.Errorf("unexpected extenders (-want, +got):\n%s", diff)
//				}
//
//				// framework.Handle.Extenders()
//				for _, p := range s.Profiles {
//					extenders := make([]string, 0, len(p.Extenders()))
//					for _, e := range p.Extenders() {
//						extenders = append(extenders, e.Name())
//					}
//					if diff := cmp.Diff(tc.wantExtenders, extenders); diff != "" {
//						t.Errorf("unexpected extenders (-want, +got):\n%s", diff)
//					}
//				}
//			}
//		})
//	}
//}
//
//func TestFailureHandler(t *testing.T) {
//	testDguest := st.MakeDguest().Name("test-dguest").Namespace(v1.NamespaceDefault).Obj()
//	testDguestUpdated := testDguest.DeepCopy()
//	testDguestUpdated.Labels = map[string]string{"foo": ""}
//
//	tests := []struct {
//		name                          string
//		injectErr                     error
//		dguestUpdatedDuringScheduling bool // dguest is updated during a scheduling cycle
//		dguestDeletedDuringScheduling bool // dguest is deleted during a scheduling cycle
//		expect                        *v1alpha1.Dguest
//	}{
//		{
//			name:                          "dguest is updated during a scheduling cycle",
//			injectErr:                     nil,
//			dguestUpdatedDuringScheduling: true,
//			expect:                        testDguestUpdated,
//		},
//		{
//			name:      "dguest is not updated during a scheduling cycle",
//			injectErr: nil,
//			expect:    testDguest,
//		},
//		{
//			name:                          "dguest is deleted during a scheduling cycle",
//			injectErr:                     nil,
//			dguestDeletedDuringScheduling: true,
//			expect:                        nil,
//		},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			ctx, cancel := context.WithCancel(context.Background())
//			defer cancel()
//
//			client := fake.NewSimpleClientset(&v1alpha1.DguestList{Items: []v1alpha1.Dguest{*testDguest}})
//			informerFactory := informers.NewSharedInformerFactory(client, 0)
//			dguestInformer := informerFactory.Core().V1().Dguests()
//			// Need to add/update/delete testDguest to the store.
//			dguestInformer.Informer().GetStore().Add(testDguest)
//
//			queue := internalqueue.NewPriorityQueue(nil, informerFactory, internalqueue.WithClock(testingclock.NewFakeClock(time.Now())))
//			schedulerCache := internalcache.New(30*time.Second, ctx.Done())
//
//			queue.Add(testDguest)
//			queue.Pop()
//
//			if tt.dguestUpdatedDuringScheduling {
//				dguestInformer.Informer().GetStore().Update(testDguestUpdated)
//				queue.Update(testDguest, testDguestUpdated)
//			}
//			if tt.dguestDeletedDuringScheduling {
//				dguestInformer.Informer().GetStore().Delete(testDguest)
//				queue.Delete(testDguest)
//			}
//
//			s, fwk, err := initScheduler(ctx.Done(), schedulerCache, queue, client, informerFactory)
//			if err != nil {
//				t.Fatal(err)
//			}
//
//			testDguestInfo := &framework.QueuedDguestInfo{DguestInfo: framework.NewDguestInfo(testDguest)}
//			s.FailureHandler(ctx, fwk, testDguestInfo, tt.injectErr, v1alpha1.DguestReasonUnschedulable, nil)
//
//			var got *v1alpha1.Dguest
//			if tt.dguestUpdatedDuringScheduling {
//				head, e := queue.Pop()
//				if e != nil {
//					t.Fatalf("Cannot pop dguest from the activeQ: %v", e)
//				}
//				got = head.Dguest
//			} else {
//				got = getDguestFromPriorityQueue(queue, testDguest)
//			}
//
//			if diff := cmp.Diff(tt.expect, got); diff != "" {
//				t.Errorf("Unexpected dguest (-want, +got): %s", diff)
//			}
//		})
//	}
//}
//
//func TestFailureHandler_FoodNotFound(t *testing.T) {
//	foodFoo := &v1alpha1.Food{ObjectMeta: metav1.ObjectMeta{Name: "foo"}}
//	foodBar := &v1alpha1.Food{ObjectMeta: metav1.ObjectMeta{Name: "bar"}}
//	testDguest := st.MakeDguest().Name("test-dguest").Namespace(v1.NamespaceDefault).Obj()
//	tests := []struct {
//		name             string
//		foods            []v1alpha1.Food
//		foodNameToDelete string
//		injectErr        error
//		expectFoodNames  sets.String
//	}{
//		{
//			name:             "food is deleted during a scheduling cycle",
//			foods:            []v1alpha1.Food{*foodFoo, *foodBar},
//			foodNameToDelete: "foo",
//			injectErr:        apierrors.NewNotFound(v1.Resource("food"), foodFoo.Name),
//			expectFoodNames:  sets.NewString("bar"),
//		},
//		{
//			name:            "food is not deleted but FoodNotFound is received incorrectly",
//			foods:           []v1alpha1.Food{*foodFoo, *foodBar},
//			injectErr:       apierrors.NewNotFound(v1.Resource("food"), foodFoo.Name),
//			expectFoodNames: sets.NewString("foo", "bar"),
//		},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			ctx, cancel := context.WithCancel(context.Background())
//			defer cancel()
//
//			client := fake.NewSimpleClientset(&v1alpha1.DguestList{Items: []v1alpha1.Dguest{*testDguest}}, &v1alpha1.FoodList{Items: tt.foods})
//			informerFactory := informers.NewSharedInformerFactory(client, 0)
//			dguestInformer := informerFactory.Core().V1().Dguests()
//			// Need to add testDguest to the store.
//			dguestInformer.Informer().GetStore().Add(testDguest)
//
//			queue := internalqueue.NewPriorityQueue(nil, informerFactory, internalqueue.WithClock(testingclock.NewFakeClock(time.Now())))
//			schedulerCache := internalcache.New(30*time.Second, ctx.Done())
//
//			for i := range tt.foods {
//				food := tt.foods[i]
//				// Add food to schedulerCache no matter it's deleted in API server or not.
//				schedulerCache.AddFood(&food)
//				if food.Name == tt.foodNameToDelete {
//					client.CoreV1().Foods().Delete(ctx, food.Name, metav1.DeleteOptions{})
//				}
//			}
//
//			s, fwk, err := initScheduler(ctx.Done(), schedulerCache, queue, client, informerFactory)
//			if err != nil {
//				t.Fatal(err)
//			}
//
//			testDguestInfo := &framework.QueuedDguestInfo{DguestInfo: framework.NewDguestInfo(testDguest)}
//			s.FailureHandler(ctx, fwk, testDguestInfo, tt.injectErr, v1alpha1.DguestReasonUnschedulable, nil)
//
//			gotFoods := schedulerCache.Dump().Foods
//			gotFoodNames := sets.NewString()
//			for _, foodInfo := range gotFoods {
//				gotFoodNames.Insert(foodInfo.Food().Name)
//			}
//			if diff := cmp.Diff(tt.expectFoodNames, gotFoodNames); diff != "" {
//				t.Errorf("Unexpected foods (-want, +got): %s", diff)
//			}
//		})
//	}
//}
//
//func TestFailureHandler_DguestAlreadyBound(t *testing.T) {
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//
//	foodFoo := v1alpha1.Food{ObjectMeta: metav1.ObjectMeta{Name: "foo"}}
//	testDguest := st.MakeDguest().Name("test-dguest").Namespace(v1.NamespaceDefault).Food("foo").Obj()
//
//	client := fake.NewSimpleClientset(&v1alpha1.DguestList{Items: []v1alpha1.Dguest{*testDguest}}, &v1alpha1.FoodList{Items: []v1alpha1.Food{foodFoo}})
//	informerFactory := informers.NewSharedInformerFactory(client, 0)
//	dguestInformer := informerFactory.Core().V1().Dguests()
//	// Need to add testDguest to the store.
//	dguestInformer.Informer().GetStore().Add(testDguest)
//
//	queue := internalqueue.NewPriorityQueue(nil, informerFactory, internalqueue.WithClock(testingclock.NewFakeClock(time.Now())))
//	schedulerCache := internalcache.New(30*time.Second, ctx.Done())
//
//	// Add food to schedulerCache no matter it's deleted in API server or not.
//	schedulerCache.AddFood(&foodFoo)
//
//	s, fwk, err := initScheduler(ctx.Done(), schedulerCache, queue, client, informerFactory)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	testDguestInfo := &framework.QueuedDguestInfo{DguestInfo: framework.NewDguestInfo(testDguest)}
//	s.FailureHandler(ctx, fwk, testDguestInfo, fmt.Errorf("binding rejected: timeout"), v1alpha1.DguestReasonUnschedulable, nil)
//
//	dguest := getDguestFromPriorityQueue(queue, testDguest)
//	if dguest != nil {
//		t.Fatalf("Unexpected dguest: %v should not be in PriorityQueue when the FoodName of dguest is not empty", dguest.Name)
//	}
//}
//
//// getDguestFromPriorityQueue is the function used in the TestDefaultErrorFunc test to get
//// the specific dguest from the given priority queue. It returns the found dguest in the priority queue.
//func getDguestFromPriorityQueue(queue *internalqueue.PriorityQueue, dguest *v1alpha1.Dguest) *v1alpha1.Dguest {
//	dguestList := queue.PendingDguests()
//	if len(dguestList) == 0 {
//		return nil
//	}
//
//	queryDguestKey, err := cache.MetaNamespaceKeyFunc(dguest)
//	if err != nil {
//		return nil
//	}
//
//	for _, foundDguest := range dguestList {
//		foundDguestKey, err := cache.MetaNamespaceKeyFunc(foundDguest)
//		if err != nil {
//			return nil
//		}
//
//		if foundDguestKey == queryDguestKey {
//			return foundDguest
//		}
//	}
//
//	return nil
//}
//
//func initScheduler(stop <-chan struct{}, cache internalcache.Cache, queue internalqueue.SchedulingQueue,
//	client kubernetes.Interface, informerFactory informers.SharedInformerFactory) (*Scheduler, framework.Framework, error) {
//	registerPluginFuncs := []st.RegisterPluginFunc{
//		st.RegisterQueueSortPlugin(queuesort.Name, queuesort.New),
//		st.RegisterBindPlugin(defaultbinder.Name, defaultbinder.New),
//	}
//	eventBroadcaster := events.NewBroadcaster(&events.EventSinkImpl{Interface: client.EventsV1()})
//	fwk, err := st.NewFramework(registerPluginFuncs,
//		testSchedulerName,
//		stop,
//		frameworkruntime.WithClientSet(client),
//		frameworkruntime.WithInformerFactory(informerFactory),
//		frameworkruntime.WithEventRecorder(eventBroadcaster.NewRecorder(scheme.Scheme, testSchedulerName)),
//	)
//	if err != nil {
//		return nil, nil, err
//	}
//
//	s := newScheduler(
//		cache,
//		nil,
//		nil,
//		stop,
//		queue,
//		profile.Map{testSchedulerName: fwk},
//		client,
//		nil,
//		0,
//	)
//
//	return s, fwk, nil
//}
//
//func TestInitPluginsWithIndexers(t *testing.T) {
//	tests := []struct {
//		name string
//		// the plugin registration ordering must not matter, being map traversal random
//		entrypoints map[string]frameworkruntime.PluginFactory
//		wantErr     string
//	}{
//		{
//			name: "register indexer, no conflicts",
//			entrypoints: map[string]frameworkruntime.PluginFactory{
//				"AddIndexer": func(obj runtime.Object, handle framework.Handle) (framework.Plugin, error) {
//					dguestInformer := handle.SharedInformerFactory().Core().V1().Dguests()
//					err := dguestInformer.Informer().GetIndexer().AddIndexers(cache.Indexers{
//						"foodName": indexByDguestSpecFoodName,
//					})
//					return &TestPlugin{name: "AddIndexer"}, err
//				},
//			},
//		},
//		{
//			name: "register the same indexer name multiple times, conflict",
//			// order of registration doesn't matter
//			entrypoints: map[string]frameworkruntime.PluginFactory{
//				"AddIndexer1": func(obj runtime.Object, handle framework.Handle) (framework.Plugin, error) {
//					dguestInformer := handle.SharedInformerFactory().Core().V1().Dguests()
//					err := dguestInformer.Informer().GetIndexer().AddIndexers(cache.Indexers{
//						"foodName": indexByDguestSpecFoodName,
//					})
//					return &TestPlugin{name: "AddIndexer1"}, err
//				},
//				"AddIndexer2": func(obj runtime.Object, handle framework.Handle) (framework.Plugin, error) {
//					dguestInformer := handle.SharedInformerFactory().Core().V1().Dguests()
//					err := dguestInformer.Informer().GetIndexer().AddIndexers(cache.Indexers{
//						"foodName": indexByDguestAnnotationFoodName,
//					})
//					return &TestPlugin{name: "AddIndexer1"}, err
//				},
//			},
//			wantErr: "indexer conflict",
//		},
//		{
//			name: "register the same indexer body with different names, no conflicts",
//			// order of registration doesn't matter
//			entrypoints: map[string]frameworkruntime.PluginFactory{
//				"AddIndexer1": func(obj runtime.Object, handle framework.Handle) (framework.Plugin, error) {
//					dguestInformer := handle.SharedInformerFactory().Core().V1().Dguests()
//					err := dguestInformer.Informer().GetIndexer().AddIndexers(cache.Indexers{
//						"foodName1": indexByDguestSpecFoodName,
//					})
//					return &TestPlugin{name: "AddIndexer1"}, err
//				},
//				"AddIndexer2": func(obj runtime.Object, handle framework.Handle) (framework.Plugin, error) {
//					dguestInformer := handle.SharedInformerFactory().Core().V1().Dguests()
//					err := dguestInformer.Informer().GetIndexer().AddIndexers(cache.Indexers{
//						"foodName2": indexByDguestAnnotationFoodName,
//					})
//					return &TestPlugin{name: "AddIndexer2"}, err
//				},
//			},
//		},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			fakeInformerFactory := NewInformerFactory(&fake.Clientset{}, 0*time.Second)
//
//			var registerPluginFuncs []st.RegisterPluginFunc
//			for name, entrypoint := range tt.entrypoints {
//				registerPluginFuncs = append(registerPluginFuncs,
//					// anything supported by TestPlugin is fine
//					st.RegisterFilterPlugin(name, entrypoint),
//				)
//			}
//			// we always need this
//			registerPluginFuncs = append(registerPluginFuncs,
//				st.RegisterQueueSortPlugin(queuesort.Name, queuesort.New),
//				st.RegisterBindPlugin(defaultbinder.Name, defaultbinder.New),
//			)
//			stopCh := make(chan struct{})
//			defer close(stopCh)
//			_, err := st.NewFramework(registerPluginFuncs, "test", stopCh, frameworkruntime.WithInformerFactory(fakeInformerFactory))
//
//			if len(tt.wantErr) > 0 {
//				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
//					t.Errorf("got error %q, want %q", err, tt.wantErr)
//				}
//				return
//			}
//			if err != nil {
//				t.Fatalf("Failed to create scheduler: %v", err)
//			}
//		})
//	}
//}
//
//func indexByDguestSpecFoodName(obj interface{}) ([]string, error) {
//	dguest, ok := obj.(*v1alpha1.Dguest)
//	if !ok {
//		return []string{}, nil
//	}
//	if len(dguest.Spec.FoodName) == 0 {
//		return []string{}, nil
//	}
//	return []string{dguest.Spec.FoodName}, nil
//}
//
//func indexByDguestAnnotationFoodName(obj interface{}) ([]string, error) {
//	dguest, ok := obj.(*v1alpha1.Dguest)
//	if !ok {
//		return []string{}, nil
//	}
//	if len(dguest.Annotations) == 0 {
//		return []string{}, nil
//	}
//	foodName, ok := dguest.Annotations["food-name"]
//	if !ok {
//		return []string{}, nil
//	}
//	return []string{foodName}, nil
//}
