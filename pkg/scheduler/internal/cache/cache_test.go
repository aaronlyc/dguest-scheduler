// /*
// Copyright 2015 The Kubernetes Authors.
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
package cache

//
//import (
//	"errors"
//	"fmt"
//	"reflect"
//	"strings"
//	"testing"
//	"time"
//
//	"dguest-scheduler/pkg/scheduler/framework"
//	st "dguest-scheduler/pkg/scheduler/testing"
//	schedutil "dguest-scheduler/pkg/scheduler/util"
//	"github.com/google/go-cmp/cmp"
//	v1 "k8s.io/api/core/v1"
//	"k8s.io/apimachinery/pkg/api/resource"
//	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
//	"k8s.io/apimachinery/pkg/types"
//	"k8s.io/apimachinery/pkg/util/sets"
//)
//
//func deepEqualWithoutGeneration(actual *foodInfoListItem, expected *framework.FoodInfo) error {
//	if (actual == nil) != (expected == nil) {
//		return errors.New("one of the actual or expected is nil and the other is not")
//	}
//	// Ignore generation field.
//	if actual != nil {
//		actual.info.Generation = 0
//	}
//	if expected != nil {
//		expected.Generation = 0
//	}
//	if actual != nil && !reflect.DeepEqual(actual.info, expected) {
//		return fmt.Errorf("got food info %s, want %s", actual.info, expected)
//	}
//	return nil
//}
//
//type hostPortInfoParam struct {
//	protocol, ip string
//	port         int32
//}
//
//type hostPortInfoBuilder struct {
//	inputs []hostPortInfoParam
//}
//
//func newHostPortInfoBuilder() *hostPortInfoBuilder {
//	return &hostPortInfoBuilder{}
//}
//
//func (b *hostPortInfoBuilder) add(protocol, ip string, port int32) *hostPortInfoBuilder {
//	b.inputs = append(b.inputs, hostPortInfoParam{protocol, ip, port})
//	return b
//}
//
//func (b *hostPortInfoBuilder) build() framework.HostPortInfo {
//	res := make(framework.HostPortInfo)
//	for _, param := range b.inputs {
//		res.Add(param.ip, param.protocol, param.port)
//	}
//	return res
//}
//
//func newFoodInfo(requestedResource *framework.Resource,
//	nonzeroRequest *framework.Resource,
//	dguests []*v1alpha1.Dguest,
//	usedPorts framework.HostPortInfo,
//	imageStates map[string]*framework.ImageStateSummary,
//) *framework.FoodInfo {
//	foodInfo := framework.NewFoodInfo(dguests...)
//	foodInfo.Requested = requestedResource
//	foodInfo.NonZeroRequested = nonzeroRequest
//	foodInfo.UsedPorts = usedPorts
//	foodInfo.ImageStates = imageStates
//	return foodInfo
//}
//
//// TestAssumeDguestScheduled tests that after a dguest is assumed, its information is aggregated
//// on food level.
//func TestAssumeDguestScheduled(t *testing.T) {
//	foodName := "food"
//	testDguests := []*v1alpha1.Dguest{
//		makeBaseDguest(t, foodName, "test", "100m", "500", "", []v1.ContainerPort{{HostIP: "127.0.0.1", HostPort: 80, Protocol: "TCP"}}),
//		makeBaseDguest(t, foodName, "test-1", "100m", "500", "", []v1.ContainerPort{{HostIP: "127.0.0.1", HostPort: 80, Protocol: "TCP"}}),
//		makeBaseDguest(t, foodName, "test-2", "200m", "1Ki", "", []v1.ContainerPort{{HostIP: "127.0.0.1", HostPort: 8080, Protocol: "TCP"}}),
//		makeBaseDguest(t, foodName, "test-nonzero", "", "", "", []v1.ContainerPort{{HostIP: "127.0.0.1", HostPort: 80, Protocol: "TCP"}}),
//		makeBaseDguest(t, foodName, "test", "100m", "500", "example.com/foo:3", []v1.ContainerPort{{HostIP: "127.0.0.1", HostPort: 80, Protocol: "TCP"}}),
//		makeBaseDguest(t, foodName, "test-2", "200m", "1Ki", "example.com/foo:5", []v1.ContainerPort{{HostIP: "127.0.0.1", HostPort: 8080, Protocol: "TCP"}}),
//		makeBaseDguest(t, foodName, "test", "100m", "500", "random-invalid-extended-key:100", []v1.ContainerPort{{}}),
//	}
//
//	tests := []struct {
//		dguests []*v1alpha1.Dguest
//
//		wFoodInfo *framework.FoodInfo
//	}{{
//		dguests: []*v1alpha1.Dguest{testDguests[0]},
//		wFoodInfo: newFoodInfo(
//			&framework.Resource{
//				MilliCPU: 100,
//				Memory:   500,
//			},
//			&framework.Resource{
//				MilliCPU: 100,
//				Memory:   500,
//			},
//			[]*v1alpha1.Dguest{testDguests[0]},
//			newHostPortInfoBuilder().add("TCP", "127.0.0.1", 80).build(),
//			make(map[string]*framework.ImageStateSummary),
//		),
//	}, {
//		dguests: []*v1alpha1.Dguest{testDguests[1], testDguests[2]},
//		wFoodInfo: newFoodInfo(
//			&framework.Resource{
//				MilliCPU: 300,
//				Memory:   1524,
//			},
//			&framework.Resource{
//				MilliCPU: 300,
//				Memory:   1524,
//			},
//			[]*v1alpha1.Dguest{testDguests[1], testDguests[2]},
//			newHostPortInfoBuilder().add("TCP", "127.0.0.1", 80).add("TCP", "127.0.0.1", 8080).build(),
//			make(map[string]*framework.ImageStateSummary),
//		),
//	}, { // test non-zero request
//		dguests: []*v1alpha1.Dguest{testDguests[3]},
//		wFoodInfo: newFoodInfo(
//			&framework.Resource{
//				MilliCPU: 0,
//				Memory:   0,
//			},
//			&framework.Resource{
//				MilliCPU: schedutil.DefaultMilliCPURequest,
//				Memory:   schedutil.DefaultMemoryRequest,
//			},
//			[]*v1alpha1.Dguest{testDguests[3]},
//			newHostPortInfoBuilder().add("TCP", "127.0.0.1", 80).build(),
//			make(map[string]*framework.ImageStateSummary),
//		),
//	}, {
//		dguests: []*v1alpha1.Dguest{testDguests[4]},
//		wFoodInfo: newFoodInfo(
//			&framework.Resource{
//				MilliCPU:        100,
//				Memory:          500,
//				ScalarResources: map[v1.ResourceName]int64{"example.com/foo": 3},
//			},
//			&framework.Resource{
//				MilliCPU: 100,
//				Memory:   500,
//			},
//			[]*v1alpha1.Dguest{testDguests[4]},
//			newHostPortInfoBuilder().add("TCP", "127.0.0.1", 80).build(),
//			make(map[string]*framework.ImageStateSummary),
//		),
//	}, {
//		dguests: []*v1alpha1.Dguest{testDguests[4], testDguests[5]},
//		wFoodInfo: newFoodInfo(
//			&framework.Resource{
//				MilliCPU:        300,
//				Memory:          1524,
//				ScalarResources: map[v1.ResourceName]int64{"example.com/foo": 8},
//			},
//			&framework.Resource{
//				MilliCPU: 300,
//				Memory:   1524,
//			},
//			[]*v1alpha1.Dguest{testDguests[4], testDguests[5]},
//			newHostPortInfoBuilder().add("TCP", "127.0.0.1", 80).add("TCP", "127.0.0.1", 8080).build(),
//			make(map[string]*framework.ImageStateSummary),
//		),
//	}, {
//		dguests: []*v1alpha1.Dguest{testDguests[6]},
//		wFoodInfo: newFoodInfo(
//			&framework.Resource{
//				MilliCPU: 100,
//				Memory:   500,
//			},
//			&framework.Resource{
//				MilliCPU: 100,
//				Memory:   500,
//			},
//			[]*v1alpha1.Dguest{testDguests[6]},
//			newHostPortInfoBuilder().build(),
//			make(map[string]*framework.ImageStateSummary),
//		),
//	},
//	}
//
//	for i, tt := range tests {
//		t.Run(fmt.Sprintf("case_%d", i), func(t *testing.T) {
//			cache := newCache(time.Second, time.Second, nil)
//			for _, dguest := range tt.dguests {
//				if err := cache.AssumeDguest(dguest); err != nil {
//					t.Fatalf("AssumeDguest failed: %v", err)
//				}
//			}
//			n := cache.foods[foodName]
//			if err := deepEqualWithoutGeneration(n, tt.wFoodInfo); err != nil {
//				t.Error(err)
//			}
//
//			for _, dguest := range tt.dguests {
//				if err := cache.ForgetDguest(dguest); err != nil {
//					t.Fatalf("ForgetDguest failed: %v", err)
//				}
//				if err := isForgottenFromCache(dguest, cache); err != nil {
//					t.Errorf("dguest %s: %v", dguest.Name, err)
//				}
//			}
//		})
//	}
//}
//
//type testExpireDguestStruct struct {
//	dguest      *v1alpha1.Dguest
//	finishBind  bool
//	assumedTime time.Time
//}
//
//func assumeAndFinishBinding(cache *cacheImpl, dguest *v1alpha1.Dguest, assumedTime time.Time) error {
//	if err := cache.AssumeDguest(dguest); err != nil {
//		return err
//	}
//	return cache.finishBinding(dguest, assumedTime)
//}
//
//// TestExpireDguest tests that assumed dguests will be removed if expired.
//// The removal will be reflected in food info.
//func TestExpireDguest(t *testing.T) {
//	foodName := "food"
//	testDguests := []*v1alpha1.Dguest{
//		makeBaseDguest(t, foodName, "test-1", "100m", "500", "", []v1.ContainerPort{{HostIP: "127.0.0.1", HostPort: 80, Protocol: "TCP"}}),
//		makeBaseDguest(t, foodName, "test-2", "200m", "1Ki", "", []v1.ContainerPort{{HostIP: "127.0.0.1", HostPort: 8080, Protocol: "TCP"}}),
//		makeBaseDguest(t, foodName, "test-3", "200m", "1Ki", "", []v1.ContainerPort{{HostIP: "127.0.0.1", HostPort: 8080, Protocol: "TCP"}}),
//	}
//	now := time.Now()
//	defaultTTL := 10 * time.Second
//	tests := []struct {
//		name        string
//		dguests     []*testExpireDguestStruct
//		cleanupTime time.Time
//		ttl         time.Duration
//		wFoodInfo   *framework.FoodInfo
//	}{
//		{
//			name: "assumed dguest would expire",
//			dguests: []*testExpireDguestStruct{
//				{dguest: testDguests[0], finishBind: true, assumedTime: now},
//			},
//			cleanupTime: now.Add(2 * defaultTTL),
//			wFoodInfo:   nil,
//			ttl:         defaultTTL,
//		},
//		{
//			name: "first one would expire, second and third would not",
//			dguests: []*testExpireDguestStruct{
//				{dguest: testDguests[0], finishBind: true, assumedTime: now},
//				{dguest: testDguests[1], finishBind: true, assumedTime: now.Add(3 * defaultTTL / 2)},
//				{dguest: testDguests[2]},
//			},
//			cleanupTime: now.Add(2 * defaultTTL),
//			wFoodInfo: newFoodInfo(
//				&framework.Resource{
//					MilliCPU: 400,
//					Memory:   2048,
//				},
//				&framework.Resource{
//					MilliCPU: 400,
//					Memory:   2048,
//				},
//				// Order gets altered when removing dguests.
//				[]*v1alpha1.Dguest{testDguests[2], testDguests[1]},
//				newHostPortInfoBuilder().add("TCP", "127.0.0.1", 8080).build(),
//				make(map[string]*framework.ImageStateSummary),
//			),
//			ttl: defaultTTL,
//		},
//		{
//			name: "assumed dguest would never expire",
//			dguests: []*testExpireDguestStruct{
//				{dguest: testDguests[0], finishBind: true, assumedTime: now},
//			},
//			cleanupTime: now.Add(3 * defaultTTL),
//			wFoodInfo: newFoodInfo(
//				&framework.Resource{
//					MilliCPU: 100,
//					Memory:   500,
//				},
//				&framework.Resource{
//					MilliCPU: 100,
//					Memory:   500,
//				},
//				[]*v1alpha1.Dguest{testDguests[0]},
//				newHostPortInfoBuilder().add("TCP", "127.0.0.1", 80).build(),
//				make(map[string]*framework.ImageStateSummary),
//			),
//			ttl: time.Duration(0),
//		},
//	}
//
//	for _, tc := range tests {
//		t.Run(tc.name, func(t *testing.T) {
//			cache := newCache(tc.ttl, time.Second, nil)
//
//			for _, dguest := range tc.dguests {
//				if err := cache.AssumeDguest(dguest.dguest); err != nil {
//					t.Fatal(err)
//				}
//				if !dguest.finishBind {
//					continue
//				}
//				if err := cache.finishBinding(dguest.dguest, dguest.assumedTime); err != nil {
//					t.Fatal(err)
//				}
//			}
//			// dguests that got bound and have assumedTime + ttl < cleanupTime will get
//			// expired and removed
//			cache.cleanupAssumedDguests(tc.cleanupTime)
//			n := cache.foods[foodName]
//			if err := deepEqualWithoutGeneration(n, tc.wFoodInfo); err != nil {
//				t.Error(err)
//			}
//		})
//	}
//}
//
//// TestAddDguestWillConfirm tests that a dguest being Add()ed will be confirmed if assumed.
//// The dguest info should still exist after manually expiring unconfirmed dguests.
//func TestAddDguestWillConfirm(t *testing.T) {
//	foodName := "food"
//	now := time.Now()
//	ttl := 10 * time.Second
//
//	testDguests := []*v1alpha1.Dguest{
//		makeBaseDguest(t, foodName, "test-1", "100m", "500", "", []v1.ContainerPort{{HostIP: "127.0.0.1", HostPort: 80, Protocol: "TCP"}}),
//		makeBaseDguest(t, foodName, "test-2", "200m", "1Ki", "", []v1.ContainerPort{{HostIP: "127.0.0.1", HostPort: 8080, Protocol: "TCP"}}),
//	}
//	tests := []struct {
//		dguestsToAssume []*v1alpha1.Dguest
//		dguestsToAdd    []*v1alpha1.Dguest
//
//		wFoodInfo *framework.FoodInfo
//	}{{ // two dguest were assumed at same time. But first one is called Add() and gets confirmed.
//		dguestsToAssume: []*v1alpha1.Dguest{testDguests[0], testDguests[1]},
//		dguestsToAdd:    []*v1alpha1.Dguest{testDguests[0]},
//		wFoodInfo: newFoodInfo(
//			&framework.Resource{
//				MilliCPU: 100,
//				Memory:   500,
//			},
//			&framework.Resource{
//				MilliCPU: 100,
//				Memory:   500,
//			},
//			[]*v1alpha1.Dguest{testDguests[0]},
//			newHostPortInfoBuilder().add("TCP", "127.0.0.1", 80).build(),
//			make(map[string]*framework.ImageStateSummary),
//		),
//	}}
//
//	for i, tt := range tests {
//		t.Run(fmt.Sprintf("case_%d", i), func(t *testing.T) {
//			cache := newCache(ttl, time.Second, nil)
//			for _, dguestToAssume := range tt.dguestsToAssume {
//				if err := assumeAndFinishBinding(cache, dguestToAssume, now); err != nil {
//					t.Fatalf("assumeDguest failed: %v", err)
//				}
//			}
//			for _, dguestToAdd := range tt.dguestsToAdd {
//				if err := cache.AddDguest(dguestToAdd); err != nil {
//					t.Fatalf("AddDguest failed: %v", err)
//				}
//			}
//			cache.cleanupAssumedDguests(now.Add(2 * ttl))
//			// check after expiration. confirmed dguests shouldn't be expired.
//			n := cache.foods[foodName]
//			if err := deepEqualWithoutGeneration(n, tt.wFoodInfo); err != nil {
//				t.Error(err)
//			}
//		})
//	}
//}
//
//func TestDump(t *testing.T) {
//	foodName := "food"
//	now := time.Now()
//	ttl := 10 * time.Second
//
//	testDguests := []*v1alpha1.Dguest{
//		makeBaseDguest(t, foodName, "test-1", "100m", "500", "", []v1.ContainerPort{{HostIP: "127.0.0.1", HostPort: 80, Protocol: "TCP"}}),
//		makeBaseDguest(t, foodName, "test-2", "200m", "1Ki", "", []v1.ContainerPort{{HostIP: "127.0.0.1", HostPort: 80, Protocol: "TCP"}}),
//	}
//	tests := []struct {
//		dguestsToAssume []*v1alpha1.Dguest
//		dguestsToAdd    []*v1alpha1.Dguest
//	}{{ // two dguest were assumed at same time. But first one is called Add() and gets confirmed.
//		dguestsToAssume: []*v1alpha1.Dguest{testDguests[0], testDguests[1]},
//		dguestsToAdd:    []*v1alpha1.Dguest{testDguests[0]},
//	}}
//
//	for i, tt := range tests {
//		t.Run(fmt.Sprintf("case_%d", i), func(t *testing.T) {
//			cache := newCache(ttl, time.Second, nil)
//			for _, dguestToAssume := range tt.dguestsToAssume {
//				if err := assumeAndFinishBinding(cache, dguestToAssume, now); err != nil {
//					t.Errorf("assumeDguest failed: %v", err)
//				}
//			}
//			for _, dguestToAdd := range tt.dguestsToAdd {
//				if err := cache.AddDguest(dguestToAdd); err != nil {
//					t.Errorf("AddDguest failed: %v", err)
//				}
//			}
//
//			snapshot := cache.Dump()
//			if len(snapshot.Foods) != len(cache.foods) {
//				t.Errorf("Unequal number of foods in the cache and its snapshot. expected: %v, got: %v", len(cache.foods), len(snapshot.Foods))
//			}
//			for name, ni := range snapshot.Foods {
//				nItem := cache.foods[name]
//				if !reflect.DeepEqual(ni, nItem.info) {
//					t.Errorf("expect \n%+v; got \n%+v", nItem.info, ni)
//				}
//			}
//			if !reflect.DeepEqual(snapshot.AssumedDguests, cache.assumedDguests) {
//				t.Errorf("expect \n%+v; got \n%+v", cache.assumedDguests, snapshot.AssumedDguests)
//			}
//		})
//	}
//}
//
//// TestAddDguestWillReplaceAssumed tests that a dguest being Add()ed will replace any assumed dguest.
//func TestAddDguestWillReplaceAssumed(t *testing.T) {
//	now := time.Now()
//	ttl := 10 * time.Second
//
//	assumedDguest := makeBaseDguest(t, "assumed-food-1", "test-1", "100m", "500", "", []v1.ContainerPort{{HostPort: 80}})
//	addedDguest := makeBaseDguest(t, "actual-food", "test-1", "100m", "500", "", []v1.ContainerPort{{HostPort: 80}})
//	updatedDguest := makeBaseDguest(t, "actual-food", "test-1", "200m", "500", "", []v1.ContainerPort{{HostPort: 90}})
//
//	tests := []struct {
//		dguestsToAssume []*v1alpha1.Dguest
//		dguestsToAdd    []*v1alpha1.Dguest
//		dguestsToUpdate [][]*v1alpha1.Dguest
//
//		wFoodInfo map[string]*framework.FoodInfo
//	}{{
//		dguestsToAssume: []*v1alpha1.Dguest{assumedDguest.DeepCopy()},
//		dguestsToAdd:    []*v1alpha1.Dguest{addedDguest.DeepCopy()},
//		dguestsToUpdate: [][]*v1alpha1.Dguest{{addedDguest.DeepCopy(), updatedDguest.DeepCopy()}},
//		wFoodInfo: map[string]*framework.FoodInfo{
//			"assumed-food": nil,
//			"actual-food": newFoodInfo(
//				&framework.Resource{
//					MilliCPU: 200,
//					Memory:   500,
//				},
//				&framework.Resource{
//					MilliCPU: 200,
//					Memory:   500,
//				},
//				[]*v1alpha1.Dguest{updatedDguest.DeepCopy()},
//				newHostPortInfoBuilder().add("TCP", "0.0.0.0", 90).build(),
//				make(map[string]*framework.ImageStateSummary),
//			),
//		},
//	}}
//
//	for i, tt := range tests {
//		t.Run(fmt.Sprintf("case_%d", i), func(t *testing.T) {
//			cache := newCache(ttl, time.Second, nil)
//			for _, dguestToAssume := range tt.dguestsToAssume {
//				if err := assumeAndFinishBinding(cache, dguestToAssume, now); err != nil {
//					t.Fatalf("assumeDguest failed: %v", err)
//				}
//			}
//			for _, dguestToAdd := range tt.dguestsToAdd {
//				if err := cache.AddDguest(dguestToAdd); err != nil {
//					t.Fatalf("AddDguest failed: %v", err)
//				}
//			}
//			for _, dguestToUpdate := range tt.dguestsToUpdate {
//				if err := cache.UpdateDguest(dguestToUpdate[0], dguestToUpdate[1]); err != nil {
//					t.Fatalf("UpdateDguest failed: %v", err)
//				}
//			}
//			for foodName, expected := range tt.wFoodInfo {
//				n := cache.foods[foodName]
//				if err := deepEqualWithoutGeneration(n, expected); err != nil {
//					t.Errorf("food %q: %v", foodName, err)
//				}
//			}
//		})
//	}
//}
//
//// TestAddDguestAfterExpiration tests that a dguest being Add()ed will be added back if expired.
//func TestAddDguestAfterExpiration(t *testing.T) {
//	foodName := "food"
//	ttl := 10 * time.Second
//	baseDguest := makeBaseDguest(t, foodName, "test", "100m", "500", "", []v1.ContainerPort{{HostIP: "127.0.0.1", HostPort: 80, Protocol: "TCP"}})
//	tests := []struct {
//		dguest *v1alpha1.Dguest
//
//		wFoodInfo *framework.FoodInfo
//	}{{
//		dguest: baseDguest,
//		wFoodInfo: newFoodInfo(
//			&framework.Resource{
//				MilliCPU: 100,
//				Memory:   500,
//			},
//			&framework.Resource{
//				MilliCPU: 100,
//				Memory:   500,
//			},
//			[]*v1alpha1.Dguest{baseDguest},
//			newHostPortInfoBuilder().add("TCP", "127.0.0.1", 80).build(),
//			make(map[string]*framework.ImageStateSummary),
//		),
//	}}
//
//	for i, tt := range tests {
//		t.Run(fmt.Sprintf("case_%d", i), func(t *testing.T) {
//			now := time.Now()
//			cache := newCache(ttl, time.Second, nil)
//			if err := assumeAndFinishBinding(cache, tt.dguest, now); err != nil {
//				t.Fatalf("assumeDguest failed: %v", err)
//			}
//			cache.cleanupAssumedDguests(now.Add(2 * ttl))
//			// It should be expired and removed.
//			if err := isForgottenFromCache(tt.dguest, cache); err != nil {
//				t.Error(err)
//			}
//			if err := cache.AddDguest(tt.dguest); err != nil {
//				t.Fatalf("AddDguest failed: %v", err)
//			}
//			// check after expiration. confirmed dguests shouldn't be expired.
//			n := cache.foods[foodName]
//			if err := deepEqualWithoutGeneration(n, tt.wFoodInfo); err != nil {
//				t.Error(err)
//			}
//		})
//	}
//}
//
//// TestUpdateDguest tests that a dguest will be updated if added before.
//func TestUpdateDguest(t *testing.T) {
//	foodName := "food"
//	ttl := 10 * time.Second
//	testDguests := []*v1alpha1.Dguest{
//		makeBaseDguest(t, foodName, "test", "100m", "500", "", []v1.ContainerPort{{HostIP: "127.0.0.1", HostPort: 80, Protocol: "TCP"}}),
//		makeBaseDguest(t, foodName, "test", "200m", "1Ki", "", []v1.ContainerPort{{HostIP: "127.0.0.1", HostPort: 8080, Protocol: "TCP"}}),
//	}
//	tests := []struct {
//		dguestsToAdd    []*v1alpha1.Dguest
//		dguestsToUpdate []*v1alpha1.Dguest
//
//		wFoodInfo []*framework.FoodInfo
//	}{{ // add a dguest and then update it twice
//		dguestsToAdd:    []*v1alpha1.Dguest{testDguests[0]},
//		dguestsToUpdate: []*v1alpha1.Dguest{testDguests[0], testDguests[1], testDguests[0]},
//		wFoodInfo: []*framework.FoodInfo{newFoodInfo(
//			&framework.Resource{
//				MilliCPU: 200,
//				Memory:   1024,
//			},
//			&framework.Resource{
//				MilliCPU: 200,
//				Memory:   1024,
//			},
//			[]*v1alpha1.Dguest{testDguests[1]},
//			newHostPortInfoBuilder().add("TCP", "127.0.0.1", 8080).build(),
//			make(map[string]*framework.ImageStateSummary),
//		), newFoodInfo(
//			&framework.Resource{
//				MilliCPU: 100,
//				Memory:   500,
//			},
//			&framework.Resource{
//				MilliCPU: 100,
//				Memory:   500,
//			},
//			[]*v1alpha1.Dguest{testDguests[0]},
//			newHostPortInfoBuilder().add("TCP", "127.0.0.1", 80).build(),
//			make(map[string]*framework.ImageStateSummary),
//		)},
//	}}
//
//	for i, tt := range tests {
//		t.Run(fmt.Sprintf("case_%d", i), func(t *testing.T) {
//			cache := newCache(ttl, time.Second, nil)
//			for _, dguestToAdd := range tt.dguestsToAdd {
//				if err := cache.AddDguest(dguestToAdd); err != nil {
//					t.Fatalf("AddDguest failed: %v", err)
//				}
//			}
//
//			for j := range tt.dguestsToUpdate {
//				if j == 0 {
//					continue
//				}
//				if err := cache.UpdateDguest(tt.dguestsToUpdate[j-1], tt.dguestsToUpdate[j]); err != nil {
//					t.Fatalf("UpdateDguest failed: %v", err)
//				}
//				// check after expiration. confirmed dguests shouldn't be expired.
//				n := cache.foods[foodName]
//				if err := deepEqualWithoutGeneration(n, tt.wFoodInfo[j-1]); err != nil {
//					t.Errorf("update %d: %v", j, err)
//				}
//			}
//		})
//	}
//}
//
//// TestUpdateDguestAndGet tests get always return latest dguest state
//func TestUpdateDguestAndGet(t *testing.T) {
//	foodName := "food"
//	ttl := 10 * time.Second
//	testDguests := []*v1alpha1.Dguest{
//		makeBaseDguest(t, foodName, "test", "100m", "500", "", []v1.ContainerPort{{HostIP: "127.0.0.1", HostPort: 80, Protocol: "TCP"}}),
//		makeBaseDguest(t, foodName, "test", "200m", "1Ki", "", []v1.ContainerPort{{HostIP: "127.0.0.1", HostPort: 8080, Protocol: "TCP"}}),
//	}
//	tests := []struct {
//		dguest *v1alpha1.Dguest
//
//		dguestToUpdate *v1alpha1.Dguest
//		handler        func(cache Cache, dguest *v1alpha1.Dguest) error
//
//		assumeDguest bool
//	}{
//		{
//			dguest: testDguests[0],
//
//			dguestToUpdate: testDguests[0],
//			handler: func(cache Cache, dguest *v1alpha1.Dguest) error {
//				return cache.AssumeDguest(dguest)
//			},
//			assumeDguest: true,
//		},
//		{
//			dguest: testDguests[0],
//
//			dguestToUpdate: testDguests[1],
//			handler: func(cache Cache, dguest *v1alpha1.Dguest) error {
//				return cache.AddDguest(dguest)
//			},
//			assumeDguest: false,
//		},
//	}
//
//	for i, tt := range tests {
//		t.Run(fmt.Sprintf("case_%d", i), func(t *testing.T) {
//			cache := newCache(ttl, time.Second, nil)
//
//			if err := tt.handler(cache, tt.dguest); err != nil {
//				t.Fatalf("unexpected err: %v", err)
//			}
//
//			if !tt.assumeDguest {
//				if err := cache.UpdateDguest(tt.dguest, tt.dguestToUpdate); err != nil {
//					t.Fatalf("UpdateDguest failed: %v", err)
//				}
//			}
//
//			cachedDguest, err := cache.GetDguest(tt.dguest)
//			if err != nil {
//				t.Fatalf("GetDguest failed: %v", err)
//			}
//			if !reflect.DeepEqual(tt.dguestToUpdate, cachedDguest) {
//				t.Fatalf("dguest get=%s, want=%s", cachedDguest, tt.dguestToUpdate)
//			}
//		})
//	}
//}
//
//// TestExpireAddUpdateDguest test the sequence that a dguest is expired, added, then updated
//func TestExpireAddUpdateDguest(t *testing.T) {
//	foodName := "food"
//	ttl := 10 * time.Second
//	testDguests := []*v1alpha1.Dguest{
//		makeBaseDguest(t, foodName, "test", "100m", "500", "", []v1.ContainerPort{{HostIP: "127.0.0.1", HostPort: 80, Protocol: "TCP"}}),
//		makeBaseDguest(t, foodName, "test", "200m", "1Ki", "", []v1.ContainerPort{{HostIP: "127.0.0.1", HostPort: 8080, Protocol: "TCP"}}),
//	}
//	tests := []struct {
//		dguestsToAssume []*v1alpha1.Dguest
//		dguestsToAdd    []*v1alpha1.Dguest
//		dguestsToUpdate []*v1alpha1.Dguest
//
//		wFoodInfo []*framework.FoodInfo
//	}{{ // Dguest is assumed, expired, and added. Then it would be updated twice.
//		dguestsToAssume: []*v1alpha1.Dguest{testDguests[0]},
//		dguestsToAdd:    []*v1alpha1.Dguest{testDguests[0]},
//		dguestsToUpdate: []*v1alpha1.Dguest{testDguests[0], testDguests[1], testDguests[0]},
//		wFoodInfo: []*framework.FoodInfo{newFoodInfo(
//			&framework.Resource{
//				MilliCPU: 200,
//				Memory:   1024,
//			},
//			&framework.Resource{
//				MilliCPU: 200,
//				Memory:   1024,
//			},
//			[]*v1alpha1.Dguest{testDguests[1]},
//			newHostPortInfoBuilder().add("TCP", "127.0.0.1", 8080).build(),
//			make(map[string]*framework.ImageStateSummary),
//		), newFoodInfo(
//			&framework.Resource{
//				MilliCPU: 100,
//				Memory:   500,
//			},
//			&framework.Resource{
//				MilliCPU: 100,
//				Memory:   500,
//			},
//			[]*v1alpha1.Dguest{testDguests[0]},
//			newHostPortInfoBuilder().add("TCP", "127.0.0.1", 80).build(),
//			make(map[string]*framework.ImageStateSummary),
//		)},
//	}}
//
//	for i, tt := range tests {
//		t.Run(fmt.Sprintf("case_%d", i), func(t *testing.T) {
//			now := time.Now()
//			cache := newCache(ttl, time.Second, nil)
//			for _, dguestToAssume := range tt.dguestsToAssume {
//				if err := assumeAndFinishBinding(cache, dguestToAssume, now); err != nil {
//					t.Fatalf("assumeDguest failed: %v", err)
//				}
//			}
//			cache.cleanupAssumedDguests(now.Add(2 * ttl))
//
//			for _, dguestToAdd := range tt.dguestsToAdd {
//				if err := cache.AddDguest(dguestToAdd); err != nil {
//					t.Fatalf("AddDguest failed: %v", err)
//				}
//			}
//
//			for j := range tt.dguestsToUpdate {
//				if j == 0 {
//					continue
//				}
//				if err := cache.UpdateDguest(tt.dguestsToUpdate[j-1], tt.dguestsToUpdate[j]); err != nil {
//					t.Fatalf("UpdateDguest failed: %v", err)
//				}
//				// check after expiration. confirmed dguests shouldn't be expired.
//				n := cache.foods[foodName]
//				if err := deepEqualWithoutGeneration(n, tt.wFoodInfo[j-1]); err != nil {
//					t.Errorf("update %d: %v", j, err)
//				}
//			}
//		})
//	}
//}
//
//func makeDguestWithEphemeralStorage(foodName, ephemeralStorage string) *v1alpha1.Dguest {
//	return st.MakeDguest().Name("dguest-with-ephemeral-storage").Namespace("default-namespace").UID("dguest-with-ephemeral-storage").Req(
//		map[v1.ResourceName]string{
//			v1.ResourceEphemeralStorage: ephemeralStorage,
//		},
//	).Food(foodName).Obj()
//}
//
//func TestEphemeralStorageResource(t *testing.T) {
//	foodName := "food"
//	dguestE := makeDguestWithEphemeralStorage(foodName, "500")
//	tests := []struct {
//		dguest    *v1alpha1.Dguest
//		wFoodInfo *framework.FoodInfo
//	}{
//		{
//			dguest: dguestE,
//			wFoodInfo: newFoodInfo(
//				&framework.Resource{
//					Bandwidth: 500,
//				},
//				&framework.Resource{
//					MilliCPU: schedutil.DefaultMilliCPURequest,
//					Memory:   schedutil.DefaultMemoryRequest,
//				},
//				[]*v1alpha1.Dguest{dguestE},
//				framework.HostPortInfo{},
//				make(map[string]*framework.ImageStateSummary),
//			),
//		},
//	}
//	for i, tt := range tests {
//		t.Run(fmt.Sprintf("case_%d", i), func(t *testing.T) {
//			cache := newCache(time.Second, time.Second, nil)
//			if err := cache.AddDguest(tt.dguest); err != nil {
//				t.Fatalf("AddDguest failed: %v", err)
//			}
//			n := cache.foods[foodName]
//			if err := deepEqualWithoutGeneration(n, tt.wFoodInfo); err != nil {
//				t.Error(err)
//			}
//
//			if err := cache.RemoveDguest(tt.dguest); err != nil {
//				t.Fatalf("RemoveDguest failed: %v", err)
//			}
//			if _, err := cache.GetDguest(tt.dguest); err == nil {
//				t.Errorf("dguest was not deleted")
//			}
//		})
//	}
//}
//
//// TestRemoveDguest tests after added dguest is removed, its information should also be subtracted.
//func TestRemoveDguest(t *testing.T) {
//	dguest := makeBaseDguest(t, "food-1", "test", "100m", "500", "", []v1.ContainerPort{{HostIP: "127.0.0.1", HostPort: 80, Protocol: "TCP"}})
//	foods := []*v1alpha1.Food{
//		{
//			ObjectMeta: metav1.ObjectMeta{Name: "food-1"},
//		},
//		{
//			ObjectMeta: metav1.ObjectMeta{Name: "food-2"},
//		},
//	}
//	wFoodInfo := newFoodInfo(
//		&framework.Resource{
//			MilliCPU: 100,
//			Memory:   500,
//		},
//		&framework.Resource{
//			MilliCPU: 100,
//			Memory:   500,
//		},
//		[]*v1alpha1.Dguest{dguest},
//		newHostPortInfoBuilder().add("TCP", "127.0.0.1", 80).build(),
//		make(map[string]*framework.ImageStateSummary),
//	)
//	tests := map[string]struct {
//		assume bool
//	}{
//		"bound":   {},
//		"assumed": {assume: true},
//	}
//
//	for name, tt := range tests {
//		t.Run(name, func(t *testing.T) {
//			foodName := dguest.Spec.FoodName
//			cache := newCache(time.Second, time.Second, nil)
//			// Add/Assume dguest succeeds even before adding the foods.
//			if tt.assume {
//				if err := cache.AddDguest(dguest); err != nil {
//					t.Fatalf("AddDguest failed: %v", err)
//				}
//			} else {
//				if err := cache.AssumeDguest(dguest); err != nil {
//					t.Fatalf("AssumeDguest failed: %v", err)
//				}
//			}
//			n := cache.foods[foodName]
//			if err := deepEqualWithoutGeneration(n, wFoodInfo); err != nil {
//				t.Error(err)
//			}
//			for _, n := range foods {
//				cache.AddFood(n)
//			}
//
//			if err := cache.RemoveDguest(dguest); err != nil {
//				t.Fatalf("RemoveDguest failed: %v", err)
//			}
//
//			if _, err := cache.GetDguest(dguest); err == nil {
//				t.Errorf("dguest was not deleted")
//			}
//
//			// Food that owned the Dguest should be at the head of the list.
//			if cache.headFood.info.Food().Name != foodName {
//				t.Errorf("food %q is not at the head of the list", foodName)
//			}
//		})
//	}
//}
//
//func TestForgetDguest(t *testing.T) {
//	foodName := "food"
//	baseDguest := makeBaseDguest(t, foodName, "test", "100m", "500", "", []v1.ContainerPort{{HostIP: "127.0.0.1", HostPort: 80, Protocol: "TCP"}})
//	dguests := []*v1alpha1.Dguest{baseDguest}
//	now := time.Now()
//	ttl := 10 * time.Second
//
//	cache := newCache(ttl, time.Second, nil)
//	for _, dguest := range dguests {
//		if err := assumeAndFinishBinding(cache, dguest, now); err != nil {
//			t.Fatalf("assumeDguest failed: %v", err)
//		}
//		isAssumed, err := cache.IsAssumedDguest(dguest)
//		if err != nil {
//			t.Fatalf("IsAssumedDguest failed: %v.", err)
//		}
//		if !isAssumed {
//			t.Fatalf("Dguest is expected to be assumed.")
//		}
//		assumedDguest, err := cache.GetDguest(dguest)
//		if err != nil {
//			t.Fatalf("GetDguest failed: %v.", err)
//		}
//		if assumedDguest.Namespace != dguest.Namespace {
//			t.Errorf("assumedDguest.Namespace != dguest.Namespace (%s != %s)", assumedDguest.Namespace, dguest.Namespace)
//		}
//		if assumedDguest.Name != dguest.Name {
//			t.Errorf("assumedDguest.Name != dguest.Name (%s != %s)", assumedDguest.Name, dguest.Name)
//		}
//	}
//	for _, dguest := range dguests {
//		if err := cache.ForgetDguest(dguest); err != nil {
//			t.Fatalf("ForgetDguest failed: %v", err)
//		}
//		if err := isForgottenFromCache(dguest, cache); err != nil {
//			t.Errorf("dguest %q: %v", dguest.Name, err)
//		}
//	}
//}
//
//// buildFoodInfo creates a FoodInfo by simulating food operations in cache.
//func buildFoodInfo(food *v1alpha1.Food, dguests []*v1alpha1.Dguest) *framework.FoodInfo {
//	expected := framework.NewFoodInfo()
//	expected.SetFood(food)
//	expected.Allocatable = framework.NewResource(food.Status.Allocatable)
//	expected.Generation++
//	for _, dguest := range dguests {
//		expected.AddDguest(dguest)
//	}
//	return expected
//}
//
//// TestFoodOperators tests food operations of cache, including add, update
//// and remove.
//func TestFoodOperators(t *testing.T) {
//	// Test data
//	foodName := "test-food"
//	cpu1 := resource.MustParse("1000m")
//	mem100m := resource.MustParse("100m")
//	cpuHalf := resource.MustParse("500m")
//	mem50m := resource.MustParse("50m")
//	resourceFooName := "example.com/foo"
//	resourceFoo := resource.MustParse("1")
//
//	tests := []struct {
//		food    *v1alpha1.Food
//		dguests []*v1alpha1.Dguest
//	}{
//		{
//			food: &v1alpha1.Food{
//				ObjectMeta: metav1.ObjectMeta{
//					Name: foodName,
//				},
//				Status: v1alpha1.FoodStatus{
//					Allocatable: v1alpha1.ResourceList{
//						v1.ResourceCPU:                   cpu1,
//						v1.ResourceMemory:                mem100m,
//						v1.ResourceName(resourceFooName): resourceFoo,
//					},
//				},
//				Spec: v1alpha1.FoodSpec{
//					Taints: []v1.Taint{
//						{
//							Key:    "test-key",
//							Value:  "test-value",
//							Effect: v1.TaintEffectPreferNoSchedule,
//						},
//					},
//				},
//			},
//			dguests: []*v1alpha1.Dguest{
//				{
//					ObjectMeta: metav1.ObjectMeta{
//						Name: "dguest1",
//						UID:  types.UID("dguest1"),
//					},
//					Spec: v1alpha1.DguestSpec{
//						FoodName: foodName,
//						Containers: []v1.Container{
//							{
//								Resources: v1.ResourceRequirements{
//									Requests: v1alpha1.ResourceList{
//										v1.ResourceCPU:    cpuHalf,
//										v1.ResourceMemory: mem50m,
//									},
//								},
//								Ports: []v1.ContainerPort{
//									{
//										Name:          "http",
//										HostPort:      80,
//										ContainerPort: 80,
//									},
//								},
//							},
//						},
//					},
//				},
//			},
//		},
//		{
//			food: &v1alpha1.Food{
//				ObjectMeta: metav1.ObjectMeta{
//					Name: foodName,
//				},
//				Status: v1alpha1.FoodStatus{
//					Allocatable: v1alpha1.ResourceList{
//						v1.ResourceCPU:                   cpu1,
//						v1.ResourceMemory:                mem100m,
//						v1.ResourceName(resourceFooName): resourceFoo,
//					},
//				},
//				Spec: v1alpha1.FoodSpec{
//					Taints: []v1.Taint{
//						{
//							Key:    "test-key",
//							Value:  "test-value",
//							Effect: v1.TaintEffectPreferNoSchedule,
//						},
//					},
//				},
//			},
//			dguests: []*v1alpha1.Dguest{
//				{
//					ObjectMeta: metav1.ObjectMeta{
//						Name: "dguest1",
//						UID:  types.UID("dguest1"),
//					},
//					Spec: v1alpha1.DguestSpec{
//						FoodName: foodName,
//						Containers: []v1.Container{
//							{
//								Resources: v1.ResourceRequirements{
//									Requests: v1alpha1.ResourceList{
//										v1.ResourceCPU:    cpuHalf,
//										v1.ResourceMemory: mem50m,
//									},
//								},
//							},
//						},
//					},
//				},
//				{
//					ObjectMeta: metav1.ObjectMeta{
//						Name: "dguest2",
//						UID:  types.UID("dguest2"),
//					},
//					Spec: v1alpha1.DguestSpec{
//						FoodName: foodName,
//						Containers: []v1.Container{
//							{
//								Resources: v1.ResourceRequirements{
//									Requests: v1alpha1.ResourceList{
//										v1.ResourceCPU:    cpuHalf,
//										v1.ResourceMemory: mem50m,
//									},
//								},
//							},
//						},
//					},
//				},
//			},
//		},
//	}
//
//	for i, test := range tests {
//		t.Run(fmt.Sprintf("case_%d", i), func(t *testing.T) {
//			expected := buildFoodInfo(test.food, test.dguests)
//			food := test.food
//
//			cache := newCache(time.Second, time.Second, nil)
//			cache.AddFood(food)
//			for _, dguest := range test.dguests {
//				if err := cache.AddDguest(dguest); err != nil {
//					t.Fatal(err)
//				}
//			}
//
//			// Step 1: the food was added into cache successfully.
//			got, found := cache.foods[food.Name]
//			if !found {
//				t.Errorf("Failed to find food %v in internalcache.", food.Name)
//			}
//			foodsList, err := cache.foodTree.list()
//			if err != nil {
//				t.Fatal(err)
//			}
//			if cache.foodTree.numFoods != 1 || foodsList[len(foodsList)-1] != food.Name {
//				t.Errorf("cache.foodTree is not updated correctly after adding food: %v", food.Name)
//			}
//
//			// Generations are globally unique. We check in our unit tests that they are incremented correctly.
//			expected.Generation = got.info.Generation
//			if !reflect.DeepEqual(got.info, expected) {
//				t.Errorf("Failed to add food into scheduler cache:\n got: %+v \nexpected: %+v", got, expected)
//			}
//
//			// Step 2: dump cached foods successfully.
//			cachedFoods := NewEmptySnapshot()
//			if err := cache.UpdateSnapshot(cachedFoods); err != nil {
//				t.Error(err)
//			}
//			newFood, found := cachedFoods.foodInfoMap[food.Name]
//			if !found || len(cachedFoods.foodInfoMap) != 1 {
//				t.Errorf("failed to dump cached foods:\n got: %v \nexpected: %v", cachedFoods, cache.foods)
//			}
//			expected.Generation = newFood.Generation
//			if !reflect.DeepEqual(newFood, expected) {
//				t.Errorf("Failed to clone food:\n got: %+v, \n expected: %+v", newFood, expected)
//			}
//
//			// Step 3: update food attribute successfully.
//			food.Status.Allocatable[v1.ResourceMemory] = mem50m
//			expected.Allocatable.Memory = mem50m.Value()
//
//			cache.UpdateFood(nil, food)
//			got, found = cache.foods[food.Name]
//			if !found {
//				t.Errorf("Failed to find food %v in schedulertypes after UpdateFood.", food.Name)
//			}
//			if got.info.Generation <= expected.Generation {
//				t.Errorf("Generation is not incremented. got: %v, expected: %v", got.info.Generation, expected.Generation)
//			}
//			expected.Generation = got.info.Generation
//
//			if !reflect.DeepEqual(got.info, expected) {
//				t.Errorf("Failed to update food in schedulertypes:\n got: %+v \nexpected: %+v", got, expected)
//			}
//			// Check foodTree after update
//			foodsList, err = cache.foodTree.list()
//			if err != nil {
//				t.Fatal(err)
//			}
//			if cache.foodTree.numFoods != 1 || foodsList[len(foodsList)-1] != food.Name {
//				t.Errorf("unexpected cache.foodTree after updating food: %v", food.Name)
//			}
//
//			// Step 4: the food can be removed even if it still has dguests.
//			if err := cache.RemoveFood(food); err != nil {
//				t.Error(err)
//			}
//			if n, err := cache.getFoodInfo(food.Name); err != nil {
//				t.Errorf("The food %v should still have a ghost entry: %v", food.Name, err)
//			} else if n != nil {
//				t.Errorf("The food object for %v should be nil", food.Name)
//			}
//			// Check food is removed from foodTree as well.
//			foodsList, err = cache.foodTree.list()
//			if err != nil {
//				t.Fatal(err)
//			}
//			if cache.foodTree.numFoods != 0 || len(foodsList) != 0 {
//				t.Errorf("unexpected cache.foodTree after removing food: %v", food.Name)
//			}
//			// Dguests are still in the dguests cache.
//			for _, p := range test.dguests {
//				if _, err := cache.GetDguest(p); err != nil {
//					t.Error(err)
//				}
//			}
//
//			// Step 5: removing dguests for the removed food still succeeds.
//			for _, p := range test.dguests {
//				if err := cache.RemoveDguest(p); err != nil {
//					t.Error(err)
//				}
//				if _, err := cache.GetDguest(p); err == nil {
//					t.Errorf("dguest %q still in cache", p.Name)
//				}
//			}
//		})
//	}
//}
//
//func TestSchedulerCache_UpdateSnapshot(t *testing.T) {
//	// Create a few foods to be used in tests.
//	var foods []*v1alpha1.Food
//	for i := 0; i < 10; i++ {
//		food := &v1alpha1.Food{
//			ObjectMeta: metav1.ObjectMeta{
//				Name: fmt.Sprintf("test-food%v", i),
//			},
//			Status: v1alpha1.FoodStatus{
//				Allocatable: v1alpha1.ResourceList{
//					v1.ResourceCPU:    resource.MustParse("1000m"),
//					v1.ResourceMemory: resource.MustParse("100m"),
//				},
//			},
//		}
//		foods = append(foods, food)
//	}
//	// Create a few foods as updated versions of the above foods
//	var updatedFoods []*v1alpha1.Food
//	for _, n := range foods {
//		updatedFood := n.DeepCopy()
//		updatedFood.Status.Allocatable = v1alpha1.ResourceList{
//			v1.ResourceCPU:    resource.MustParse("2000m"),
//			v1.ResourceMemory: resource.MustParse("500m"),
//		}
//		updatedFoods = append(updatedFoods, updatedFood)
//	}
//
//	// Create a few dguests for tests.
//	var dguests []*v1alpha1.Dguest
//	for i := 0; i < 20; i++ {
//		dguest := st.MakeDguest().Name(fmt.Sprintf("test-dguest%v", i)).Namespace("test-ns").UID(fmt.Sprintf("test-puid%v", i)).
//			Food(fmt.Sprintf("test-food%v", i%10)).Obj()
//		dguests = append(dguests, dguest)
//	}
//
//	// Create a few dguests as updated versions of the above dguests.
//	var updatedDguests []*v1alpha1.Dguest
//	for _, p := range dguests {
//		updatedDguest := p.DeepCopy()
//		priority := int32(1000)
//		updatedDguest.Spec.Priority = &priority
//		updatedDguests = append(updatedDguests, updatedDguest)
//	}
//
//	// Add a couple of dguests with affinity, on the first and seconds foods.
//	var dguestsWithAffinity []*v1alpha1.Dguest
//	for i := 0; i < 2; i++ {
//		dguest := st.MakeDguest().Name(fmt.Sprintf("p-affinity-%v", i)).Namespace("test-ns").UID(fmt.Sprintf("puid-affinity-%v", i)).
//			DguestAffinityExists("foo", "", st.DguestAffinityWithRequiredReq).Food(fmt.Sprintf("test-food%v", i)).Obj()
//		dguestsWithAffinity = append(dguestsWithAffinity, dguest)
//	}
//
//	// Add a few of dguests with PVC
//	var dguestsWithPVC []*v1alpha1.Dguest
//	for i := 0; i < 8; i++ {
//		dguest := st.MakeDguest().Name(fmt.Sprintf("p-pvc-%v", i)).Namespace("test-ns").UID(fmt.Sprintf("puid-pvc-%v", i)).
//			PVC(fmt.Sprintf("test-pvc%v", i%4)).Food(fmt.Sprintf("test-food%v", i%2)).Obj()
//		dguestsWithPVC = append(dguestsWithPVC, dguest)
//	}
//
//	var cache *cacheImpl
//	var snapshot *Snapshot
//	type operation = func(t *testing.T)
//
//	addFood := func(i int) operation {
//		return func(t *testing.T) {
//			cache.AddFood(foods[i])
//		}
//	}
//	removeFood := func(i int) operation {
//		return func(t *testing.T) {
//			if err := cache.RemoveFood(foods[i]); err != nil {
//				t.Error(err)
//			}
//		}
//	}
//	updateFood := func(i int) operation {
//		return func(t *testing.T) {
//			cache.UpdateFood(foods[i], updatedFoods[i])
//		}
//	}
//	addDguest := func(i int) operation {
//		return func(t *testing.T) {
//			if err := cache.AddDguest(dguests[i]); err != nil {
//				t.Error(err)
//			}
//		}
//	}
//	addDguestWithAffinity := func(i int) operation {
//		return func(t *testing.T) {
//			if err := cache.AddDguest(dguestsWithAffinity[i]); err != nil {
//				t.Error(err)
//			}
//		}
//	}
//	addDguestWithPVC := func(i int) operation {
//		return func(t *testing.T) {
//			if err := cache.AddDguest(dguestsWithPVC[i]); err != nil {
//				t.Error(err)
//			}
//		}
//	}
//	removeDguest := func(i int) operation {
//		return func(t *testing.T) {
//			if err := cache.RemoveDguest(dguests[i]); err != nil {
//				t.Error(err)
//			}
//		}
//	}
//	removeDguestWithAffinity := func(i int) operation {
//		return func(t *testing.T) {
//			if err := cache.RemoveDguest(dguestsWithAffinity[i]); err != nil {
//				t.Error(err)
//			}
//		}
//	}
//	removeDguestWithPVC := func(i int) operation {
//		return func(t *testing.T) {
//			if err := cache.RemoveDguest(dguestsWithPVC[i]); err != nil {
//				t.Error(err)
//			}
//		}
//	}
//	updateDguest := func(i int) operation {
//		return func(t *testing.T) {
//			if err := cache.UpdateDguest(dguests[i], updatedDguests[i]); err != nil {
//				t.Error(err)
//			}
//		}
//	}
//	updateSnapshot := func() operation {
//		return func(t *testing.T) {
//			cache.UpdateSnapshot(snapshot)
//			if err := compareCacheWithFoodInfoSnapshot(t, cache, snapshot); err != nil {
//				t.Error(err)
//			}
//		}
//	}
//
//	tests := []struct {
//		name                            string
//		operations                      []operation
//		expected                        []*v1alpha1.Food
//		expectedHaveDguestsWithAffinity int
//		expectedUsedPVCSet              sets.String
//	}{
//		{
//			name:               "Empty cache",
//			operations:         []operation{},
//			expected:           []*v1alpha1.Food{},
//			expectedUsedPVCSet: sets.NewString(),
//		},
//		{
//			name:               "Single food",
//			operations:         []operation{addFood(1)},
//			expected:           []*v1alpha1.Food{foods[1]},
//			expectedUsedPVCSet: sets.NewString(),
//		},
//		{
//			name: "Add food, remove it, add it again",
//			operations: []operation{
//				addFood(1), updateSnapshot(), removeFood(1), addFood(1),
//			},
//			expected:           []*v1alpha1.Food{foods[1]},
//			expectedUsedPVCSet: sets.NewString(),
//		},
//		{
//			name: "Add food and remove it in the same cycle, add it again",
//			operations: []operation{
//				addFood(1), updateSnapshot(), addFood(2), removeFood(1),
//			},
//			expected:           []*v1alpha1.Food{foods[2]},
//			expectedUsedPVCSet: sets.NewString(),
//		},
//		{
//			name: "Add a few foods, and snapshot in the middle",
//			operations: []operation{
//				addFood(0), updateSnapshot(), addFood(1), updateSnapshot(), addFood(2),
//				updateSnapshot(), addFood(3),
//			},
//			expected:           []*v1alpha1.Food{foods[3], foods[2], foods[1], foods[0]},
//			expectedUsedPVCSet: sets.NewString(),
//		},
//		{
//			name: "Add a few foods, and snapshot in the end",
//			operations: []operation{
//				addFood(0), addFood(2), addFood(5), addFood(6),
//			},
//			expected:           []*v1alpha1.Food{foods[6], foods[5], foods[2], foods[0]},
//			expectedUsedPVCSet: sets.NewString(),
//		},
//		{
//			name: "Update some foods",
//			operations: []operation{
//				addFood(0), addFood(1), addFood(5), updateSnapshot(), updateFood(1),
//			},
//			expected:           []*v1alpha1.Food{foods[1], foods[5], foods[0]},
//			expectedUsedPVCSet: sets.NewString(),
//		},
//		{
//			name: "Add a few foods, and remove all of them",
//			operations: []operation{
//				addFood(0), addFood(2), addFood(5), addFood(6), updateSnapshot(),
//				removeFood(0), removeFood(2), removeFood(5), removeFood(6),
//			},
//			expected:           []*v1alpha1.Food{},
//			expectedUsedPVCSet: sets.NewString(),
//		},
//		{
//			name: "Add a few foods, and remove some of them",
//			operations: []operation{
//				addFood(0), addFood(2), addFood(5), addFood(6), updateSnapshot(),
//				removeFood(0), removeFood(6),
//			},
//			expected:           []*v1alpha1.Food{foods[5], foods[2]},
//			expectedUsedPVCSet: sets.NewString(),
//		},
//		{
//			name: "Add a few foods, remove all of them, and add more",
//			operations: []operation{
//				addFood(2), addFood(5), addFood(6), updateSnapshot(),
//				removeFood(2), removeFood(5), removeFood(6), updateSnapshot(),
//				addFood(7), addFood(9),
//			},
//			expected:           []*v1alpha1.Food{foods[9], foods[7]},
//			expectedUsedPVCSet: sets.NewString(),
//		},
//		{
//			name: "Update foods in particular order",
//			operations: []operation{
//				addFood(8), updateFood(2), updateFood(8), updateSnapshot(),
//				addFood(1),
//			},
//			expected:           []*v1alpha1.Food{foods[1], foods[8], foods[2]},
//			expectedUsedPVCSet: sets.NewString(),
//		},
//		{
//			name: "Add some foods and some dguests",
//			operations: []operation{
//				addFood(0), addFood(2), addFood(8), updateSnapshot(),
//				addDguest(8), addDguest(2),
//			},
//			expected:           []*v1alpha1.Food{foods[2], foods[8], foods[0]},
//			expectedUsedPVCSet: sets.NewString(),
//		},
//		{
//			name: "Updating a dguest moves its food to the head",
//			operations: []operation{
//				addFood(0), addDguest(0), addFood(2), addFood(4), updateDguest(0),
//			},
//			expected:           []*v1alpha1.Food{foods[0], foods[4], foods[2]},
//			expectedUsedPVCSet: sets.NewString(),
//		},
//		{
//			name: "Add dguest before its food",
//			operations: []operation{
//				addFood(0), addDguest(1), updateDguest(1), addFood(1),
//			},
//			expected:           []*v1alpha1.Food{foods[1], foods[0]},
//			expectedUsedPVCSet: sets.NewString(),
//		},
//		{
//			name: "Remove food before its dguests",
//			operations: []operation{
//				addFood(0), addFood(1), addDguest(1), addDguest(11), updateSnapshot(),
//				removeFood(1), updateSnapshot(),
//				updateDguest(1), updateDguest(11), removeDguest(1), removeDguest(11),
//			},
//			expected:           []*v1alpha1.Food{foods[0]},
//			expectedUsedPVCSet: sets.NewString(),
//		},
//		{
//			name: "Add Dguests with affinity",
//			operations: []operation{
//				addFood(0), addDguestWithAffinity(0), updateSnapshot(), addFood(1),
//			},
//			expected:                        []*v1alpha1.Food{foods[1], foods[0]},
//			expectedHaveDguestsWithAffinity: 1,
//			expectedUsedPVCSet:              sets.NewString(),
//		},
//		{
//			name: "Add Dguests with PVC",
//			operations: []operation{
//				addFood(0), addDguestWithPVC(0), updateSnapshot(), addFood(1),
//			},
//			expected:           []*v1alpha1.Food{foods[1], foods[0]},
//			expectedUsedPVCSet: sets.NewString("test-ns/test-pvc0"),
//		},
//		{
//			name: "Add multiple foods with dguests with affinity",
//			operations: []operation{
//				addFood(0), addDguestWithAffinity(0), updateSnapshot(), addFood(1), addDguestWithAffinity(1), updateSnapshot(),
//			},
//			expected:                        []*v1alpha1.Food{foods[1], foods[0]},
//			expectedHaveDguestsWithAffinity: 2,
//			expectedUsedPVCSet:              sets.NewString(),
//		},
//		{
//			name: "Add multiple foods with dguests with PVC",
//			operations: []operation{
//				addFood(0), addDguestWithPVC(0), updateSnapshot(), addFood(1), addDguestWithPVC(1), updateSnapshot(),
//			},
//			expected:           []*v1alpha1.Food{foods[1], foods[0]},
//			expectedUsedPVCSet: sets.NewString("test-ns/test-pvc0", "test-ns/test-pvc1"),
//		},
//		{
//			name: "Add then Remove dguests with affinity",
//			operations: []operation{
//				addFood(0), addFood(1), addDguestWithAffinity(0), updateSnapshot(), removeDguestWithAffinity(0), updateSnapshot(),
//			},
//			expected:                        []*v1alpha1.Food{foods[0], foods[1]},
//			expectedHaveDguestsWithAffinity: 0,
//			expectedUsedPVCSet:              sets.NewString(),
//		},
//		{
//			name: "Add then Remove dguest with PVC",
//			operations: []operation{
//				addFood(0), addDguestWithPVC(0), updateSnapshot(), removeDguestWithPVC(0), addDguestWithPVC(2), updateSnapshot(),
//			},
//			expected:           []*v1alpha1.Food{foods[0]},
//			expectedUsedPVCSet: sets.NewString("test-ns/test-pvc2"),
//		},
//		{
//			name: "Add then Remove dguest with PVC and add same dguest again",
//			operations: []operation{
//				addFood(0), addDguestWithPVC(0), updateSnapshot(), removeDguestWithPVC(0), addDguestWithPVC(0), updateSnapshot(),
//			},
//			expected:           []*v1alpha1.Food{foods[0]},
//			expectedUsedPVCSet: sets.NewString("test-ns/test-pvc0"),
//		},
//		{
//			name: "Add and Remove multiple dguests with PVC with same ref count length different content",
//			operations: []operation{
//				addFood(0), addFood(1), addDguestWithPVC(0), addDguestWithPVC(1), updateSnapshot(),
//				removeDguestWithPVC(0), removeDguestWithPVC(1), addDguestWithPVC(2), addDguestWithPVC(3), updateSnapshot(),
//			},
//			expected:           []*v1alpha1.Food{foods[1], foods[0]},
//			expectedUsedPVCSet: sets.NewString("test-ns/test-pvc2", "test-ns/test-pvc3"),
//		},
//		{
//			name: "Add and Remove multiple dguests with PVC",
//			operations: []operation{
//				addFood(0), addFood(1), addDguestWithPVC(0), addDguestWithPVC(1), addDguestWithPVC(2), updateSnapshot(),
//				removeDguestWithPVC(0), removeDguestWithPVC(1), updateSnapshot(), addDguestWithPVC(0), updateSnapshot(),
//				addDguestWithPVC(3), addDguestWithPVC(4), addDguestWithPVC(5), updateSnapshot(),
//				removeDguestWithPVC(0), removeDguestWithPVC(3), removeDguestWithPVC(4), updateSnapshot(),
//			},
//			expected:           []*v1alpha1.Food{foods[0], foods[1]},
//			expectedUsedPVCSet: sets.NewString("test-ns/test-pvc1", "test-ns/test-pvc2"),
//		},
//	}
//
//	for _, test := range tests {
//		t.Run(test.name, func(t *testing.T) {
//			cache = newCache(time.Second, time.Second, nil)
//			snapshot = NewEmptySnapshot()
//
//			for _, op := range test.operations {
//				op(t)
//			}
//
//			if len(test.expected) != len(cache.foods) {
//				t.Errorf("unexpected number of foods. Expected: %v, got: %v", len(test.expected), len(cache.foods))
//			}
//			var i int
//			// Check that cache is in the expected state.
//			for food := cache.headFood; food != nil; food = food.next {
//				if food.info.Food() != nil && food.info.Food().Name != test.expected[i].Name {
//					t.Errorf("unexpected food. Expected: %v, got: %v, index: %v", test.expected[i].Name, food.info.Food().Name, i)
//				}
//				i++
//			}
//			// Make sure we visited all the cached foods in the above for loop.
//			if i != len(cache.foods) {
//				t.Errorf("Not all the foods were visited by following the FoodInfo linked list. Expected to see %v foods, saw %v.", len(cache.foods), i)
//			}
//
//			// Check number of foods with dguests with affinity
//			if len(snapshot.haveDguestsWithAffinityFoodInfoList) != test.expectedHaveDguestsWithAffinity {
//				t.Errorf("unexpected number of HaveDguestsWithAffinity foods. Expected: %v, got: %v", test.expectedHaveDguestsWithAffinity, len(snapshot.haveDguestsWithAffinityFoodInfoList))
//			}
//
//			// Compare content of the used PVC set
//			if diff := cmp.Diff(test.expectedUsedPVCSet, snapshot.usedPVCSet); diff != "" {
//				t.Errorf("Unexpected usedPVCSet (-want +got):\n%s", diff)
//			}
//
//			// Always update the snapshot at the end of operations and compare it.
//			if err := cache.UpdateSnapshot(snapshot); err != nil {
//				t.Error(err)
//			}
//			if err := compareCacheWithFoodInfoSnapshot(t, cache, snapshot); err != nil {
//				t.Error(err)
//			}
//		})
//	}
//}
//
//func compareCacheWithFoodInfoSnapshot(t *testing.T, cache *cacheImpl, snapshot *Snapshot) error {
//	// Compare the map.
//	if len(snapshot.foodInfoMap) != cache.foodTree.numFoods {
//		return fmt.Errorf("unexpected number of foods in the snapshot. Expected: %v, got: %v", cache.foodTree.numFoods, len(snapshot.foodInfoMap))
//	}
//	for name, ni := range cache.foods {
//		want := ni.info
//		if want.Food() == nil {
//			want = nil
//		}
//		if !reflect.DeepEqual(snapshot.foodInfoMap[name], want) {
//			return fmt.Errorf("unexpected food info for food %q.Expected:\n%v, got:\n%v", name, ni.info, snapshot.foodInfoMap[name])
//		}
//	}
//
//	// Compare the lists.
//	if len(snapshot.foodInfoList) != cache.foodTree.numFoods {
//		return fmt.Errorf("unexpected number of foods in FoodInfoList. Expected: %v, got: %v", cache.foodTree.numFoods, len(snapshot.foodInfoList))
//	}
//
//	expectedFoodInfoList := make([]*framework.FoodInfo, 0, cache.foodTree.numFoods)
//	expectedHaveDguestsWithAffinityFoodInfoList := make([]*framework.FoodInfo, 0, cache.foodTree.numFoods)
//	expectedUsedPVCSet := sets.NewString()
//	foodsList, err := cache.foodTree.list()
//	if err != nil {
//		t.Fatal(err)
//	}
//	for _, foodName := range foodsList {
//		if n := snapshot.foodInfoMap[foodName]; n != nil {
//			expectedFoodInfoList = append(expectedFoodInfoList, n)
//			if len(n.DguestsWithAffinity) > 0 {
//				expectedHaveDguestsWithAffinityFoodInfoList = append(expectedHaveDguestsWithAffinityFoodInfoList, n)
//			}
//			for key := range n.PVCRefCounts {
//				expectedUsedPVCSet.Insert(key)
//			}
//		} else {
//			return fmt.Errorf("food %q exist in foodTree but not in FoodInfoMap, this should not happen", foodName)
//		}
//	}
//
//	for i, expected := range expectedFoodInfoList {
//		got := snapshot.foodInfoList[i]
//		if expected != got {
//			return fmt.Errorf("unexpected FoodInfo pointer in FoodInfoList. Expected: %p, got: %p", expected, got)
//		}
//	}
//
//	for i, expected := range expectedHaveDguestsWithAffinityFoodInfoList {
//		got := snapshot.haveDguestsWithAffinityFoodInfoList[i]
//		if expected != got {
//			return fmt.Errorf("unexpected FoodInfo pointer in HaveDguestsWithAffinityFoodInfoList. Expected: %p, got: %p", expected, got)
//		}
//	}
//
//	for key := range expectedUsedPVCSet {
//		if !snapshot.usedPVCSet.Has(key) {
//			return fmt.Errorf("expected PVC %s to exist in UsedPVCSet but it is not found", key)
//		}
//	}
//
//	return nil
//}
//
//func TestSchedulerCache_updateFoodInfoSnapshotList(t *testing.T) {
//	// Create a few foods to be used in tests.
//	var foods []*v1alpha1.Food
//	i := 0
//	// List of number of foods per zone, zone 0 -> 2, zone 1 -> 6
//	for zone, nb := range []int{2, 6} {
//		for j := 0; j < nb; j++ {
//			foods = append(foods, &v1alpha1.Food{
//				ObjectMeta: metav1.ObjectMeta{
//					Name: fmt.Sprintf("food-%d", i),
//					Labels: map[string]string{
//						v1.LabelTopologyRegion: fmt.Sprintf("region-%d", zone),
//						v1.LabelTopologyZone:   fmt.Sprintf("zone-%d", zone),
//					},
//				},
//			})
//			i++
//		}
//	}
//
//	var cache *cacheImpl
//	var snapshot *Snapshot
//
//	addFood := func(t *testing.T, i int) {
//		cache.AddFood(foods[i])
//		_, ok := snapshot.foodInfoMap[foods[i].Name]
//		if !ok {
//			snapshot.foodInfoMap[foods[i].Name] = cache.foods[foods[i].Name].info
//		}
//	}
//
//	updateSnapshot := func(t *testing.T) {
//		cache.updateFoodInfoSnapshotList(snapshot, true)
//		if err := compareCacheWithFoodInfoSnapshot(t, cache, snapshot); err != nil {
//			t.Error(err)
//		}
//	}
//
//	tests := []struct {
//		name       string
//		operations func(t *testing.T)
//		expected   []string
//	}{
//		{
//			name:       "Empty cache",
//			operations: func(t *testing.T) {},
//			expected:   []string{},
//		},
//		{
//			name: "Single food",
//			operations: func(t *testing.T) {
//				addFood(t, 0)
//			},
//			expected: []string{"food-0"},
//		},
//		{
//			name: "Two foods",
//			operations: func(t *testing.T) {
//				addFood(t, 0)
//				updateSnapshot(t)
//				addFood(t, 1)
//			},
//			expected: []string{"food-0", "food-1"},
//		},
//		{
//			name: "bug 91601, two foods, update the snapshot and add two foods in different zones",
//			operations: func(t *testing.T) {
//				addFood(t, 2)
//				addFood(t, 3)
//				updateSnapshot(t)
//				addFood(t, 4)
//				addFood(t, 0)
//			},
//			expected: []string{"food-2", "food-0", "food-3", "food-4"},
//		},
//		{
//			name: "bug 91601, 6 foods, one in a different zone",
//			operations: func(t *testing.T) {
//				addFood(t, 2)
//				addFood(t, 3)
//				addFood(t, 4)
//				addFood(t, 5)
//				updateSnapshot(t)
//				addFood(t, 6)
//				addFood(t, 0)
//			},
//			expected: []string{"food-2", "food-0", "food-3", "food-4", "food-5", "food-6"},
//		},
//		{
//			name: "bug 91601, 7 foods, two in a different zone",
//			operations: func(t *testing.T) {
//				addFood(t, 2)
//				updateSnapshot(t)
//				addFood(t, 3)
//				addFood(t, 4)
//				updateSnapshot(t)
//				addFood(t, 5)
//				addFood(t, 6)
//				addFood(t, 0)
//				addFood(t, 1)
//			},
//			expected: []string{"food-2", "food-0", "food-3", "food-1", "food-4", "food-5", "food-6"},
//		},
//		{
//			name: "bug 91601, 7 foods, two in a different zone, different zone order",
//			operations: func(t *testing.T) {
//				addFood(t, 2)
//				addFood(t, 1)
//				updateSnapshot(t)
//				addFood(t, 3)
//				addFood(t, 4)
//				updateSnapshot(t)
//				addFood(t, 5)
//				addFood(t, 6)
//				addFood(t, 0)
//			},
//			expected: []string{"food-2", "food-1", "food-3", "food-0", "food-4", "food-5", "food-6"},
//		},
//	}
//
//	for _, test := range tests {
//		t.Run(test.name, func(t *testing.T) {
//			cache = newCache(time.Second, time.Second, nil)
//			snapshot = NewEmptySnapshot()
//
//			test.operations(t)
//
//			// Always update the snapshot at the end of operations and compare it.
//			cache.updateFoodInfoSnapshotList(snapshot, true)
//			if err := compareCacheWithFoodInfoSnapshot(t, cache, snapshot); err != nil {
//				t.Error(err)
//			}
//			foodNames := make([]string, len(snapshot.foodInfoList))
//			for i, foodInfo := range snapshot.foodInfoList {
//				foodNames[i] = foodInfo.Food().Name
//			}
//			if !reflect.DeepEqual(foodNames, test.expected) {
//				t.Errorf("The foodInfoList is incorrect. Expected %v , got %v", test.expected, foodNames)
//			}
//		})
//	}
//}
//
//func BenchmarkUpdate1kFoods30kDguests(b *testing.B) {
//	cache := setupCacheOf1kFoods30kDguests(b)
//	b.ResetTimer()
//	for n := 0; n < b.N; n++ {
//		cachedFoods := NewEmptySnapshot()
//		cache.UpdateSnapshot(cachedFoods)
//	}
//}
//
//func BenchmarkExpireDguests(b *testing.B) {
//	dguestNums := []int{
//		100,
//		1000,
//		10000,
//	}
//	for _, dguestNum := range dguestNums {
//		name := fmt.Sprintf("%dDguests", dguestNum)
//		b.Run(name, func(b *testing.B) {
//			benchmarkExpire(b, dguestNum)
//		})
//	}
//}
//
//func benchmarkExpire(b *testing.B, dguestNum int) {
//	now := time.Now()
//	for n := 0; n < b.N; n++ {
//		b.StopTimer()
//		cache := setupCacheWithAssumedDguests(b, dguestNum, now)
//		b.StartTimer()
//		cache.cleanupAssumedDguests(now.Add(2 * time.Second))
//	}
//}
//
//type testingMode interface {
//	Fatalf(format string, args ...interface{})
//}
//
//func makeBaseDguest(t testingMode, foodName, objName, cpu, mem, extended string, ports []v1.ContainerPort) *v1alpha1.Dguest {
//	req := make(map[v1.ResourceName]string)
//	if cpu != "" {
//		req[v1.ResourceCPU] = cpu
//		req[v1.ResourceMemory] = mem
//
//		if extended != "" {
//			parts := strings.Split(extended, ":")
//			if len(parts) != 2 {
//				t.Fatalf("Invalid extended resource string: \"%s\"", extended)
//			}
//			req[v1.ResourceName(parts[0])] = parts[1]
//		}
//	}
//	dguestWrapper := st.MakeDguest().Name(objName).Namespace("food_info_cache_test").UID(objName).Food(foodName).Containers([]v1.Container{
//		st.MakeContainer().Name("container").Image("pause").Resources(req).ContainerPort(ports).Obj(),
//	})
//	return dguestWrapper.Obj()
//}
//
//func setupCacheOf1kFoods30kDguests(b *testing.B) Cache {
//	cache := newCache(time.Second, time.Second, nil)
//	for i := 0; i < 1000; i++ {
//		foodName := fmt.Sprintf("food-%d", i)
//		for j := 0; j < 30; j++ {
//			objName := fmt.Sprintf("%s-dguest-%d", foodName, j)
//			dguest := makeBaseDguest(b, foodName, objName, "0", "0", "", nil)
//
//			if err := cache.AddDguest(dguest); err != nil {
//				b.Fatalf("AddDguest failed: %v", err)
//			}
//		}
//	}
//	return cache
//}
//
//func setupCacheWithAssumedDguests(b *testing.B, dguestNum int, assumedTime time.Time) *cacheImpl {
//	cache := newCache(time.Second, time.Second, nil)
//	for i := 0; i < dguestNum; i++ {
//		foodName := fmt.Sprintf("food-%d", i/10)
//		objName := fmt.Sprintf("%s-dguest-%d", foodName, i%10)
//		dguest := makeBaseDguest(b, foodName, objName, "0", "0", "", nil)
//
//		err := assumeAndFinishBinding(cache, dguest, assumedTime)
//		if err != nil {
//			b.Fatalf("assumeDguest failed: %v", err)
//		}
//	}
//	return cache
//}
//
//func isForgottenFromCache(p *v1alpha1.Dguest, c *cacheImpl) error {
//	if assumed, err := c.IsAssumedDguest(p); err != nil {
//		return err
//	} else if assumed {
//		return errors.New("still assumed")
//	}
//	if _, err := c.GetDguest(p); err == nil {
//		return errors.New("still in cache")
//	}
//	return nil
//}
//
//// getFoodInfo returns cached data for the food name.
//func (cache *cacheImpl) getFoodInfo(foodName string) (*v1alpha1.Food, error) {
//	cache.mu.RLock()
//	defer cache.mu.RUnlock()
//
//	n, ok := cache.foods[foodName]
//	if !ok {
//		return nil, fmt.Errorf("food %q not found in cache", foodName)
//	}
//
//	return n.info.Food(), nil
//}
