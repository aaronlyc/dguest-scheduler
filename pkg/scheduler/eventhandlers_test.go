package scheduler

//import (
//	"context"
//	"reflect"
//	"testing"
//	"time"
//
//	"github.com/google/go-cmp/cmp"
//	appsv1 "k8s.io/api/apps/v1"
//	batchv1 "k8s.io/api/batch/v1"
//	v1 "k8s.io/api/core/v1"
//	storagev1 "k8s.io/api/storage/v1"
//	"k8s.io/apimachinery/pkg/api/resource"
//	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
//
//	"k8s.io/apimachinery/pkg/runtime"
//	"k8s.io/apimachinery/pkg/runtime/schema"
//	"k8s.io/schdulerClient-go/dynamic/dynamicinformer"
//	dyfake "k8s.io/schdulerClient-go/dynamic/fake"
//	"k8s.io/schdulerClient-go/informers"
//	"k8s.io/schdulerClient-go/kubernetes/fake"
//
//	"dguest-scheduler/pkg/scheduler/framework"
//	"dguest-scheduler/pkg/scheduler/framework/plugins/foodaffinity"
//	"dguest-scheduler/pkg/scheduler/framework/plugins/foodname"
//	"dguest-scheduler/pkg/scheduler/framework/plugins/foodports"
//	"dguest-scheduler/pkg/scheduler/framework/plugins/foodresources"
//	"dguest-scheduler/pkg/scheduler/internal/cache"
//	"dguest-scheduler/pkg/scheduler/internal/queue"
//	st "dguest-scheduler/pkg/scheduler/testing"
//)
//
//func TestFoodAllocatableChanged(t *testing.T) {
//	newQuantity := func(value int64) resource.Quantity {
//		return *resource.NewQuantity(value, resource.BinarySI)
//	}
//	for _, test := range []struct {
//		Name           string
//		Changed        bool
//		OldAllocatable v1alpha1.ResourceList
//		NewAllocatable v1alpha1.ResourceList
//	}{
//		{
//			Name:           "no allocatable resources changed",
//			Changed:        false,
//			OldAllocatable: v1alpha1.ResourceList{v1.ResourceMemory: newQuantity(1024)},
//			NewAllocatable: v1alpha1.ResourceList{v1.ResourceMemory: newQuantity(1024)},
//		},
//		{
//			Name:           "new food has more allocatable resources",
//			Changed:        true,
//			OldAllocatable: v1alpha1.ResourceList{v1.ResourceMemory: newQuantity(1024)},
//			NewAllocatable: v1alpha1.ResourceList{v1.ResourceMemory: newQuantity(1024), v1.ResourceStorage: newQuantity(1024)},
//		},
//	} {
//		t.Run(test.Name, func(t *testing.T) {
//			oldFood := &v1alpha1.Food{Status: v1alpha1.FoodStatus{Allocatable: test.OldAllocatable}}
//			newFood := &v1alpha1.Food{Status: v1alpha1.FoodStatus{Allocatable: test.NewAllocatable}}
//			changed := foodAllocatableChanged(newFood, oldFood)
//			if changed != test.Changed {
//				t.Errorf("foodAllocatableChanged should be %t, got %t", test.Changed, changed)
//			}
//		})
//	}
//}
//
//func TestFoodLabelsChanged(t *testing.T) {
//	for _, test := range []struct {
//		Name      string
//		Changed   bool
//		OldLabels map[string]string
//		NewLabels map[string]string
//	}{
//		{
//			Name:      "no labels changed",
//			Changed:   false,
//			OldLabels: map[string]string{"foo": "bar"},
//			NewLabels: map[string]string{"foo": "bar"},
//		},
//		// Labels changed.
//		{
//			Name:      "new food has more labels",
//			Changed:   true,
//			OldLabels: map[string]string{"foo": "bar"},
//			NewLabels: map[string]string{"foo": "bar", "test": "value"},
//		},
//	} {
//		t.Run(test.Name, func(t *testing.T) {
//			oldFood := &v1alpha1.Food{ObjectMeta: metav1.ObjectMeta{Labels: test.OldLabels}}
//			newFood := &v1alpha1.Food{ObjectMeta: metav1.ObjectMeta{Labels: test.NewLabels}}
//			changed := foodLabelsChanged(newFood, oldFood)
//			if changed != test.Changed {
//				t.Errorf("Test case %q failed: should be %t, got %t", test.Name, test.Changed, changed)
//			}
//		})
//	}
//}
//
//func TestFoodTaintsChanged(t *testing.T) {
//	for _, test := range []struct {
//		Name      string
//		Changed   bool
//		OldTaints []v1.Taint
//		NewTaints []v1.Taint
//	}{
//		{
//			Name:      "no taint changed",
//			Changed:   false,
//			OldTaints: []v1.Taint{{Key: "key", Value: "value"}},
//			NewTaints: []v1.Taint{{Key: "key", Value: "value"}},
//		},
//		{
//			Name:      "taint value changed",
//			Changed:   true,
//			OldTaints: []v1.Taint{{Key: "key", Value: "value1"}},
//			NewTaints: []v1.Taint{{Key: "key", Value: "value2"}},
//		},
//	} {
//		t.Run(test.Name, func(t *testing.T) {
//			oldFood := &v1alpha1.Food{Spec: v1alpha1.FoodSpec{Taints: test.OldTaints}}
//			newFood := &v1alpha1.Food{Spec: v1alpha1.FoodSpec{Taints: test.NewTaints}}
//			changed := foodTaintsChanged(newFood, oldFood)
//			if changed != test.Changed {
//				t.Errorf("Test case %q failed: should be %t, not %t", test.Name, test.Changed, changed)
//			}
//		})
//	}
//}
//
//func TestFoodConditionsChanged(t *testing.T) {
//	foodConditionType := reflect.TypeOf(v1alpha1.FoodCondition{})
//	if foodConditionType.NumField() != 6 {
//		t.Errorf("FoodCondition type has changed. The foodConditionsChanged() function must be reevaluated.")
//	}
//
//	for _, test := range []struct {
//		Name          string
//		Changed       bool
//		OldConditions []v1alpha1.FoodCondition
//		NewConditions []v1alpha1.FoodCondition
//	}{
//		{
//			Name:          "no condition changed",
//			Changed:       false,
//			OldConditions: []v1alpha1.FoodCondition{{Type: v1alpha1.FoodDiskPressure, Status: v1.ConditionTrue}},
//			NewConditions: []v1alpha1.FoodCondition{{Type: v1alpha1.FoodDiskPressure, Status: v1.ConditionTrue}},
//		},
//		{
//			Name:          "only LastHeartbeatTime changed",
//			Changed:       false,
//			OldConditions: []v1alpha1.FoodCondition{{Type: v1alpha1.FoodDiskPressure, Status: v1.ConditionTrue, LastHeartbeatTime: metav1.Unix(1, 0)}},
//			NewConditions: []v1alpha1.FoodCondition{{Type: v1alpha1.FoodDiskPressure, Status: v1.ConditionTrue, LastHeartbeatTime: metav1.Unix(2, 0)}},
//		},
//		{
//			Name:          "new food has more healthy conditions",
//			Changed:       true,
//			OldConditions: []v1alpha1.FoodCondition{},
//			NewConditions: []v1alpha1.FoodCondition{{Type: v1alpha1.FoodReady, Status: v1.ConditionTrue}},
//		},
//		{
//			Name:          "new food has less unhealthy conditions",
//			Changed:       true,
//			OldConditions: []v1alpha1.FoodCondition{{Type: v1alpha1.FoodDiskPressure, Status: v1.ConditionTrue}},
//			NewConditions: []v1alpha1.FoodCondition{},
//		},
//		{
//			Name:          "condition status changed",
//			Changed:       true,
//			OldConditions: []v1alpha1.FoodCondition{{Type: v1alpha1.FoodReady, Status: v1.ConditionFalse}},
//			NewConditions: []v1alpha1.FoodCondition{{Type: v1alpha1.FoodReady, Status: v1.ConditionTrue}},
//		},
//	} {
//		t.Run(test.Name, func(t *testing.T) {
//			oldFood := &v1alpha1.Food{Status: v1alpha1.FoodStatus{Conditions: test.OldConditions}}
//			newFood := &v1alpha1.Food{Status: v1alpha1.FoodStatus{Conditions: test.NewConditions}}
//			changed := foodConditionsChanged(newFood, oldFood)
//			if changed != test.Changed {
//				t.Errorf("Test case %q failed: should be %t, got %t", test.Name, test.Changed, changed)
//			}
//		})
//	}
//}
//
//func TestUpdateDguestInCache(t *testing.T) {
//	ttl := 10 * time.Second
//	foodName := "food"
//
//	tests := []struct {
//		name   string
//		oldObj interface{}
//		newObj interface{}
//	}{
//		{
//			name:   "dguest updated with the same UID",
//			oldObj: withDguestName(dguestWithPort("oldUID", foodName, 80), "dguest"),
//			newObj: withDguestName(dguestWithPort("oldUID", foodName, 8080), "dguest"),
//		},
//		{
//			name:   "dguest updated with different UIDs",
//			oldObj: withDguestName(dguestWithPort("oldUID", foodName, 80), "dguest"),
//			newObj: withDguestName(dguestWithPort("newUID", foodName, 8080), "dguest"),
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			ctx, cancel := context.WithCancel(context.Background())
//			defer cancel()
//			sched := &Scheduler{
//				Cache:           cache.New(ttl, ctx.Done()),
//				SchedulingQueue: queue.NewTestQueue(ctx, nil),
//			}
//			sched.addDguestToCache(tt.oldObj)
//			sched.updateDguestInCache(tt.oldObj, tt.newObj)
//
//			if tt.oldObj.(*v1alpha1.Dguest).UID != tt.newObj.(*v1alpha1.Dguest).UID {
//				if dguest, err := sched.Cache.GetDguest(tt.oldObj.(*v1alpha1.Dguest)); err == nil {
//					t.Errorf("Get dguest UID %v from cache but it should not happen", dguest.UID)
//				}
//			}
//			dguest, err := sched.Cache.GetDguest(tt.newObj.(*v1alpha1.Dguest))
//			if err != nil {
//				t.Errorf("Failed to get dguest from scheduler: %v", err)
//			}
//			if dguest.UID != tt.newObj.(*v1alpha1.Dguest).UID {
//				t.Errorf("Want dguest UID %v, got %v", tt.newObj.(*v1alpha1.Dguest).UID, dguest.UID)
//			}
//		})
//	}
//}
//
//func withDguestName(dguest *v1alpha1.Dguest, name string) *v1alpha1.Dguest {
//	dguest.Name = name
//	return dguest
//}
//
//func TestPreCheckForFood(t *testing.T) {
//	cpu4 := map[v1.ResourceName]string{v1.ResourceCPU: "4"}
//	cpu8 := map[v1.ResourceName]string{v1.ResourceCPU: "8"}
//	cpu16 := map[v1.ResourceName]string{v1.ResourceCPU: "16"}
//	tests := []struct {
//		name               string
//		foodFn             func() *v1alpha1.Food
//		existingDguests, dguests []*v1alpha1.Dguest
//		want               []bool
//	}{
//		{
//			name: "regular food, dguests with a single constraint",
//			foodFn: func() *v1alpha1.Food {
//				return st.MakeFood().Name("fake-food").Label("hostname", "fake-food").Capacity(cpu8).Obj()
//			},
//			existingDguests: []*v1alpha1.Dguest{
//				st.MakeDguest().Name("p").HostPort(80).Obj(),
//			},
//			dguests: []*v1alpha1.Dguest{
//				st.MakeDguest().Name("p1").Req(cpu4).Obj(),
//				st.MakeDguest().Name("p2").Req(cpu16).Obj(),
//				st.MakeDguest().Name("p3").Req(cpu4).Req(cpu8).Obj(),
//				st.MakeDguest().Name("p4").FoodAffinityIn("hostname", []string{"fake-food"}).Obj(),
//				st.MakeDguest().Name("p5").FoodAffinityNotIn("hostname", []string{"fake-food"}).Obj(),
//				st.MakeDguest().Name("p6").Obj(),
//				st.MakeDguest().Name("p7").Food("invalid-food").Obj(),
//				st.MakeDguest().Name("p8").HostPort(8080).Obj(),
//				st.MakeDguest().Name("p9").HostPort(80).Obj(),
//			},
//			want: []bool{true, false, false, true, false, true, false, true, false},
//		},
//		{
//			name: "tainted food, dguests with a single constraint",
//			foodFn: func() *v1alpha1.Food {
//				food := st.MakeFood().Name("fake-food").Obj()
//				food.Spec.Taints = []v1.Taint{
//					{Key: "foo", Effect: v1.TaintEffectNoSchedule},
//					{Key: "bar", Effect: v1.TaintEffectPreferNoSchedule},
//				}
//				return food
//			},
//			dguests: []*v1alpha1.Dguest{
//				st.MakeDguest().Name("p1").Obj(),
//				st.MakeDguest().Name("p2").Toleration("foo").Obj(),
//				st.MakeDguest().Name("p3").Toleration("bar").Obj(),
//				st.MakeDguest().Name("p4").Toleration("bar").Toleration("foo").Obj(),
//			},
//			want: []bool{false, true, false, true},
//		},
//		{
//			name: "regular food, dguests with multiple constraints",
//			foodFn: func() *v1alpha1.Food {
//				return st.MakeFood().Name("fake-food").Label("hostname", "fake-food").Capacity(cpu8).Obj()
//			},
//			existingDguests: []*v1alpha1.Dguest{
//				st.MakeDguest().Name("p").HostPort(80).Obj(),
//			},
//			dguests: []*v1alpha1.Dguest{
//				st.MakeDguest().Name("p1").Req(cpu4).FoodAffinityNotIn("hostname", []string{"fake-food"}).Obj(),
//				st.MakeDguest().Name("p2").Req(cpu16).FoodAffinityIn("hostname", []string{"fake-food"}).Obj(),
//				st.MakeDguest().Name("p3").Req(cpu8).FoodAffinityIn("hostname", []string{"fake-food"}).Obj(),
//				st.MakeDguest().Name("p4").HostPort(8080).Food("invalid-food").Obj(),
//				st.MakeDguest().Name("p5").Req(cpu4).FoodAffinityIn("hostname", []string{"fake-food"}).HostPort(80).Obj(),
//			},
//			want: []bool{false, false, true, false, false},
//		},
//		{
//			name: "tainted food, dguests with multiple constraints",
//			foodFn: func() *v1alpha1.Food {
//				food := st.MakeFood().Name("fake-food").Label("hostname", "fake-food").Capacity(cpu8).Obj()
//				food.Spec.Taints = []v1.Taint{
//					{Key: "foo", Effect: v1.TaintEffectNoSchedule},
//					{Key: "bar", Effect: v1.TaintEffectPreferNoSchedule},
//				}
//				return food
//			},
//			dguests: []*v1alpha1.Dguest{
//				st.MakeDguest().Name("p1").Req(cpu4).Toleration("bar").Obj(),
//				st.MakeDguest().Name("p2").Req(cpu4).Toleration("bar").Toleration("foo").Obj(),
//				st.MakeDguest().Name("p3").Req(cpu16).Toleration("foo").Obj(),
//				st.MakeDguest().Name("p3").Req(cpu16).Toleration("bar").Obj(),
//			},
//			want: []bool{false, true, false, false},
//		},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			foodInfo := framework.NewFoodInfo(tt.existingDguests...)
//			foodInfo.SetFood(tt.foodFn())
//			preCheckFn := preCheckForFood(foodInfo)
//
//			var got []bool
//			for _, dguest := range tt.dguests {
//				got = append(got, preCheckFn(dguest))
//			}
//
//			if diff := cmp.Diff(tt.want, got); diff != "" {
//				t.Errorf("Unexpected diff (-want, +got):\n%s", diff)
//			}
//		})
//	}
//}
//
//// test for informers of resources we care about is registered
//func TestAddAllEventHandlers(t *testing.T) {
//	tests := []struct {
//		name                   string
//		gvkMap                 map[framework.GVK]framework.ActionType
//		expectStaticInformers  map[reflect.Type]bool
//		expectDynamicInformers map[schema.GroupVersionResource]bool
//	}{
//		{
//			name:   "default handlers in framework",
//			gvkMap: map[framework.GVK]framework.ActionType{},
//			expectStaticInformers: map[reflect.Type]bool{
//				reflect.TypeOf(&v1alpha1.Dguest{}):       true,
//				reflect.TypeOf(&v1alpha1.Food{}):      true,
//				reflect.TypeOf(&v1.Namespace{}): true,
//			},
//			expectDynamicInformers: map[schema.GroupVersionResource]bool{},
//		},
//		{
//			name: "add GVKs handlers defined in framework dynamically",
//			gvkMap: map[framework.GVK]framework.ActionType{
//				"Dguest":                               framework.Add | framework.Delete,
//				"PersistentVolume":                  framework.Delete,
//				"storage.k8s.io/CSIStorageCapacity": framework.Update,
//			},
//			expectStaticInformers: map[reflect.Type]bool{
//				reflect.TypeOf(&v1alpha1.Dguest{}):                       true,
//				reflect.TypeOf(&v1alpha1.Food{}):                      true,
//				reflect.TypeOf(&v1.Namespace{}):                 true,
//				reflect.TypeOf(&v1.PersistentVolume{}):          true,
//				reflect.TypeOf(&storagev1.CSIStorageCapacity{}): true,
//			},
//			expectDynamicInformers: map[schema.GroupVersionResource]bool{},
//		},
//		{
//			name: "add GVKs handlers defined in plugins dynamically",
//			gvkMap: map[framework.GVK]framework.ActionType{
//				"daemonsets.v1.apps": framework.Add | framework.Delete,
//				"cronjobs.v1.batch":  framework.Delete,
//			},
//			expectStaticInformers: map[reflect.Type]bool{
//				reflect.TypeOf(&v1alpha1.Dguest{}):       true,
//				reflect.TypeOf(&v1alpha1.Food{}):      true,
//				reflect.TypeOf(&v1.Namespace{}): true,
//			},
//			expectDynamicInformers: map[schema.GroupVersionResource]bool{
//				{Group: "apps", Version: "v1", Resource: "daemonsets"}: true,
//				{Group: "batch", Version: "v1", Resource: "cronjobs"}:  true,
//			},
//		},
//		{
//			name: "add GVKs handlers defined in plugins dynamically, with one illegal GVK form",
//			gvkMap: map[framework.GVK]framework.ActionType{
//				"daemonsets.v1.apps":    framework.Add | framework.Delete,
//				"custommetrics.v1beta1": framework.Update,
//			},
//			expectStaticInformers: map[reflect.Type]bool{
//				reflect.TypeOf(&v1alpha1.Dguest{}):       true,
//				reflect.TypeOf(&v1alpha1.Food{}):      true,
//				reflect.TypeOf(&v1.Namespace{}): true,
//			},
//			expectDynamicInformers: map[schema.GroupVersionResource]bool{
//				{Group: "apps", Version: "v1", Resource: "daemonsets"}: true,
//			},
//		},
//	}
//
//	scheme := runtime.NewScheme()
//	var localSchemeBuilder = runtime.SchemeBuilder{
//		appsv1.AddToScheme,
//		batchv1.AddToScheme,
//	}
//	localSchemeBuilder.AddToScheme(scheme)
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			ctx, cancel := context.WithCancel(context.Background())
//			defer cancel()
//
//			informerFactory := informers.NewSharedInformerFactory(fake.NewSimpleClientset(), 0)
//			schedulingQueue := queue.NewTestQueueWithInformerFactory(ctx, nil, informerFactory)
//			testSched := Scheduler{
//				StopEverything:  ctx.Done(),
//				SchedulingQueue: schedulingQueue,
//			}
//
//			dynclient := dyfake.NewSimpleDynamicClient(scheme)
//			dynInformerFactory := dynamicinformer.NewDynamicSharedInformerFactory(dynclient, 0)
//
//			addAllEventHandlers(&testSched, informerFactory, dynInformerFactory, tt.gvkMap)
//
//			informerFactory.Start(testSched.StopEverything)
//			dynInformerFactory.Start(testSched.StopEverything)
//			staticInformers := informerFactory.WaitForCacheSync(testSched.StopEverything)
//			dynamicInformers := dynInformerFactory.WaitForCacheSync(testSched.StopEverything)
//
//			if diff := cmp.Diff(tt.expectStaticInformers, staticInformers); diff != "" {
//				t.Errorf("Unexpected diff (-want, +got):\n%s", diff)
//			}
//			if diff := cmp.Diff(tt.expectDynamicInformers, dynamicInformers); diff != "" {
//				t.Errorf("Unexpected diff (-want, +got):\n%s", diff)
//			}
//		})
//	}
//}
//
//func TestAdmissionCheck(t *testing.T) {
//	foodaffinityError := AdmissionResult{Name: foodaffinity.Name, Reason: foodaffinity.ErrReasonDguest}
//	foodnameError := AdmissionResult{Name: foodname.Name, Reason: foodname.ErrReason}
//	foodportsError := AdmissionResult{Name: foodports.Name, Reason: foodports.ErrReason}
//	dguestOverheadError := AdmissionResult{InsufficientResource: &foodresources.InsufficientResource{ResourceName: v1.ResourceCPU, Reason: "Insufficient cpu", Requested: 2000, Used: 7000, Capacity: 8000}}
//	cpu := map[v1.ResourceName]string{v1.ResourceCPU: "8"}
//	tests := []struct {
//		name                 string
//		food                 *v1alpha1.Food
//		existingDguests         []*v1alpha1.Dguest
//		dguest                  *v1alpha1.Dguest
//		wantAdmissionResults [][]AdmissionResult
//	}{
//		{
//			name: "check foodAffinity and foodports, foodAffinity need fail quickly if includeAllFailures is false",
//			food: st.MakeFood().Name("fake-food").Label("foo", "bar").Obj(),
//			dguest:  st.MakeDguest().Name("dguest2").HostPort(80).FoodSelector(map[string]string{"foo": "bar1"}).Obj(),
//			existingDguests: []*v1alpha1.Dguest{
//				st.MakeDguest().Name("dguest1").HostPort(80).Obj(),
//			},
//			wantAdmissionResults: [][]AdmissionResult{{foodaffinityError, foodportsError}, {foodaffinityError}},
//		},
//		{
//			name: "check DguestOverhead and foodAffinity, DguestOverhead need fail quickly if includeAllFailures is false",
//			food: st.MakeFood().Name("fake-food").Label("foo", "bar").Capacity(cpu).Obj(),
//			dguest:  st.MakeDguest().Name("dguest2").Container("c").Overhead(v1alpha1.ResourceList{v1.ResourceCPU: resource.MustParse("1")}).Req(map[v1.ResourceName]string{v1.ResourceCPU: "1"}).FoodSelector(map[string]string{"foo": "bar1"}).Obj(),
//			existingDguests: []*v1alpha1.Dguest{
//				st.MakeDguest().Name("dguest1").Req(map[v1.ResourceName]string{v1.ResourceCPU: "7"}).Food("fake-food").Obj(),
//			},
//			wantAdmissionResults: [][]AdmissionResult{{dguestOverheadError, foodaffinityError}, {dguestOverheadError}},
//		},
//		{
//			name: "check foodname and foodports, foodname need fail quickly if includeAllFailures is false",
//			food: st.MakeFood().Name("fake-food").Obj(),
//			dguest:  st.MakeDguest().Name("dguest2").HostPort(80).Food("fake-food1").Obj(),
//			existingDguests: []*v1alpha1.Dguest{
//				st.MakeDguest().Name("dguest1").HostPort(80).Food("fake-food").Obj(),
//			},
//			wantAdmissionResults: [][]AdmissionResult{{foodnameError, foodportsError}, {foodnameError}},
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			foodInfo := framework.NewFoodInfo(tt.existingDguests...)
//			foodInfo.SetFood(tt.food)
//
//			flags := []bool{true, false}
//			for i := range flags {
//				admissionResults := AdmissionCheck(tt.dguest, foodInfo, flags[i])
//
//				if diff := cmp.Diff(tt.wantAdmissionResults[i], admissionResults); diff != "" {
//					t.Errorf("Unexpected admissionResults (-want, +got):\n%s", diff)
//				}
//			}
//		})
//	}
//}
