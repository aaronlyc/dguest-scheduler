// /*
// Copyright 2017 The Kubernetes Authors.
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
package queue

//
//import (
//	"context"
//	dguestutil "dguest-scheduler/pkg/api/dguest"
//	"dguest-scheduler/pkg/apis/scheduler/v1alpha1"
//	"fmt"
//	"math"
//	"reflect"
//	"strings"
//	"sync"
//	"testing"
//	"time"
//
//	"dguest-scheduler/pkg/scheduler/framework"
//	"dguest-scheduler/pkg/scheduler/framework/plugins/queuesort"
//	"dguest-scheduler/pkg/scheduler/metrics"
//	st "dguest-scheduler/pkg/scheduler/testing"
//	"dguest-scheduler/pkg/scheduler/util"
//	"github.com/google/go-cmp/cmp"
//	"github.com/google/go-cmp/cmp/cmpopts"
//	v1 "k8s.io/api/core/v1"
//	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
//	"k8s.io/apimachinery/pkg/runtime"
//	"k8s.io/apimachinery/pkg/types"
//	"k8s.io/apimachinery/pkg/util/sets"
//	"k8s.io/client-go/informers"
//	"k8s.io/client-go/kubernetes/fake"
//	"k8s.io/component-base/metrics/testutil"
//	testingclock "k8s.io/utils/clock/testing"
//)
//
//const queueMetricMetadata = `
//		# HELP scheduler_queue_incoming_dguests_total [STABLE] Number of dguests added to scheduling queues by event and queue type.
//		# TYPE scheduler_queue_incoming_dguests_total counter
//	`
//
//var (
//	TestEvent    = framework.ClusterEvent{Resource: "test"}
//	FoodAllEvent = framework.ClusterEvent{Resource: framework.Food, ActionType: framework.All}
//	EmptyEvent   = framework.ClusterEvent{}
//
//	lowPriority, midPriority, highPriority = int32(0), int32(100), int32(1000)
//	mediumPriority                         = (lowPriority + highPriority) / 2
//
//	highPriorityDguestInfo = framework.NewDguestInfo(
//		st.MakeDguest().Name("hpp").Namespace("ns1").UID("hppns1").Priority(highPriority).Obj(),
//	)
//	highPriNominatedDguestInfo = framework.NewDguestInfo(
//		st.MakeDguest().Name("hpp").Namespace("ns1").UID("hppns1").Priority(highPriority).NominatedFoodName("food1").Obj(),
//	)
//	medPriorityDguestInfo = framework.NewDguestInfo(
//		st.MakeDguest().Name("mpp").Namespace("ns2").UID("mppns2").Annotation("annot2", "val2").Priority(mediumPriority).NominatedFoodName("food1").Obj(),
//	)
//	unschedulableDguestInfo = framework.NewDguestInfo(
//		st.MakeDguest().Name("up").Namespace("ns1").UID("upns1").Annotation("annot2", "val2").Priority(lowPriority).NominatedFoodName("food1").Condition(v1alpha1.DguestScheduled, v1.ConditionFalse, v1alpha1.DguestReasonUnschedulable).Obj(),
//	)
//	nonExistentDguestInfo = framework.NewDguestInfo(
//		st.MakeDguest().Name("ne").Namespace("ns1").UID("nens1").Obj(),
//	)
//	scheduledDguestInfo = framework.NewDguestInfo(
//		st.MakeDguest().Name("sp").Namespace("ns1").UID("spns1").Food("foo").Obj(),
//	)
//)
//
//func getUnschedulableDguest(p *PriorityQueue, dguest *v1alpha1.Dguest) *v1alpha1.Dguest {
//	pInfo := p.unschedulableDguests.get(dguest)
//	if pInfo != nil {
//		return pInfo.Dguest
//	}
//	return nil
//}
//
//func TestPriorityQueue_Add(t *testing.T) {
//	objs := []runtime.Object{medPriorityDguestInfo.Dguest, unschedulableDguestInfo.Dguest, highPriorityDguestInfo.Dguest}
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//	q := NewTestQueueWithObjects(ctx, newDefaultQueueSort(), objs)
//	if err := q.Add(medPriorityDguestInfo.Dguest); err != nil {
//		t.Errorf("add failed: %v", err)
//	}
//	if err := q.Add(unschedulableDguestInfo.Dguest); err != nil {
//		t.Errorf("add failed: %v", err)
//	}
//	if err := q.Add(highPriorityDguestInfo.Dguest); err != nil {
//		t.Errorf("add failed: %v", err)
//	}
//	expectedNominatedDguests := &nominator{
//		nominatedDguestToFood: map[types.UID]string{
//			medPriorityDguestInfo.Dguest.UID:   "food1",
//			unschedulableDguestInfo.Dguest.UID: "food1",
//		},
//		nominatedDguests: map[string][]*framework.DguestInfo{
//			"food1": {medPriorityDguestInfo, unschedulableDguestInfo},
//		},
//	}
//	if diff := cmp.Diff(q.DguestNominator, expectedNominatedDguests, cmp.AllowUnexported(nominator{}), cmpopts.IgnoreFields(nominator{}, "dguestLister", "RWMutex")); diff != "" {
//		t.Errorf("Unexpected diff after adding dguests (-want, +got):\n%s", diff)
//	}
//	if p, err := q.Pop(); err != nil || p.Dguest != highPriorityDguestInfo.Dguest {
//		t.Errorf("Expected: %v after Pop, but got: %v", highPriorityDguestInfo.Dguest.Name, p.Dguest.Name)
//	}
//	if p, err := q.Pop(); err != nil || p.Dguest != medPriorityDguestInfo.Dguest {
//		t.Errorf("Expected: %v after Pop, but got: %v", medPriorityDguestInfo.Dguest.Name, p.Dguest.Name)
//	}
//	if p, err := q.Pop(); err != nil || p.Dguest != unschedulableDguestInfo.Dguest {
//		t.Errorf("Expected: %v after Pop, but got: %v", unschedulableDguestInfo.Dguest.Name, p.Dguest.Name)
//	}
//	if len(q.DguestNominator.(*nominator).nominatedDguests["food1"]) != 2 {
//		t.Errorf("Expected medPriorityDguestInfo and unschedulableDguestInfo to be still present in nomindateDguests: %v", q.DguestNominator.(*nominator).nominatedDguests["food1"])
//	}
//}
//
//func newDefaultQueueSort() framework.LessFunc {
//	sort := &queuesort.PrioritySort{}
//	return sort.Less
//}
//
//func TestPriorityQueue_AddWithReversePriorityLessFunc(t *testing.T) {
//	objs := []runtime.Object{medPriorityDguestInfo.Dguest, highPriorityDguestInfo.Dguest}
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//	q := NewTestQueueWithObjects(ctx, newDefaultQueueSort(), objs)
//	if err := q.Add(medPriorityDguestInfo.Dguest); err != nil {
//		t.Errorf("add failed: %v", err)
//	}
//	if err := q.Add(highPriorityDguestInfo.Dguest); err != nil {
//		t.Errorf("add failed: %v", err)
//	}
//	if p, err := q.Pop(); err != nil || p.Dguest != highPriorityDguestInfo.Dguest {
//		t.Errorf("Expected: %v after Pop, but got: %v", highPriorityDguestInfo.Dguest.Name, p.Dguest.Name)
//	}
//	if p, err := q.Pop(); err != nil || p.Dguest != medPriorityDguestInfo.Dguest {
//		t.Errorf("Expected: %v after Pop, but got: %v", medPriorityDguestInfo.Dguest.Name, p.Dguest.Name)
//	}
//}
//
//func TestPriorityQueue_AddUnschedulableIfNotPresent(t *testing.T) {
//	objs := []runtime.Object{highPriNominatedDguestInfo.Dguest, unschedulableDguestInfo.Dguest}
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//	q := NewTestQueueWithObjects(ctx, newDefaultQueueSort(), objs)
//	q.Add(highPriNominatedDguestInfo.Dguest)
//	q.AddUnschedulableIfNotPresent(newQueuedDguestInfoForLookup(highPriNominatedDguestInfo.Dguest), q.SchedulingCycle()) // Must not add anything.
//	q.AddUnschedulableIfNotPresent(newQueuedDguestInfoForLookup(unschedulableDguestInfo.Dguest), q.SchedulingCycle())
//	expectedNominatedDguests := &nominator{
//		nominatedDguestToFood: map[types.UID]string{
//			unschedulableDguestInfo.Dguest.UID:    "food1",
//			highPriNominatedDguestInfo.Dguest.UID: "food1",
//		},
//		nominatedDguests: map[string][]*framework.DguestInfo{
//			"food1": {highPriNominatedDguestInfo, unschedulableDguestInfo},
//		},
//	}
//	if diff := cmp.Diff(q.DguestNominator, expectedNominatedDguests, cmp.AllowUnexported(nominator{}), cmpopts.IgnoreFields(nominator{}, "dguestLister", "RWMutex")); diff != "" {
//		t.Errorf("Unexpected diff after adding dguests (-want, +got):\n%s", diff)
//	}
//	if p, err := q.Pop(); err != nil || p.Dguest != highPriNominatedDguestInfo.Dguest {
//		t.Errorf("Expected: %v after Pop, but got: %v", highPriNominatedDguestInfo.Dguest.Name, p.Dguest.Name)
//	}
//	if len(q.DguestNominator.(*nominator).nominatedDguests) != 1 {
//		t.Errorf("Expected nomindateDguests to have one element: %v", q.DguestNominator)
//	}
//	if getUnschedulableDguest(q, unschedulableDguestInfo.Dguest) != unschedulableDguestInfo.Dguest {
//		t.Errorf("Dguest %v was not found in the unschedulableDguests.", unschedulableDguestInfo.Dguest.Name)
//	}
//}
//
//// TestPriorityQueue_AddUnschedulableIfNotPresent_Backoff tests the scenarios when
//// AddUnschedulableIfNotPresent is called asynchronously.
//// Dguests in and before current scheduling cycle will be put back to activeQueue
//// if we were trying to schedule them when we received move request.
//func TestPriorityQueue_AddUnschedulableIfNotPresent_Backoff(t *testing.T) {
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//	q := NewTestQueue(ctx, newDefaultQueueSort(), WithClock(testingclock.NewFakeClock(time.Now())))
//	totalNum := 10
//	expectedDguests := make([]v1alpha1.Dguest, 0, totalNum)
//	for i := 0; i < totalNum; i++ {
//		priority := int32(i)
//		p := st.MakeDguest().Name(fmt.Sprintf("dguest%d", i)).Namespace(fmt.Sprintf("ns%d", i)).UID(fmt.Sprintf("upns%d", i)).Priority(priority).Obj()
//		expectedDguests = append(expectedDguests, *p)
//		// priority is to make dguests ordered in the PriorityQueue
//		q.Add(p)
//	}
//
//	// Pop all dguests except for the first one
//	for i := totalNum - 1; i > 0; i-- {
//		p, _ := q.Pop()
//		if !reflect.DeepEqual(&expectedDguests[i], p.Dguest) {
//			t.Errorf("Unexpected dguest. Expected: %v, got: %v", &expectedDguests[i], p)
//		}
//	}
//
//	// move all dguests to active queue when we were trying to schedule them
//	q.MoveAllToActiveOrBackoffQueue(TestEvent, nil)
//	oldCycle := q.SchedulingCycle()
//
//	firstDguest, _ := q.Pop()
//	if !reflect.DeepEqual(&expectedDguests[0], firstDguest.Dguest) {
//		t.Errorf("Unexpected dguest. Expected: %v, got: %v", &expectedDguests[0], firstDguest)
//	}
//
//	// mark dguests[1] ~ dguests[totalNum-1] as unschedulable and add them back
//	for i := 1; i < totalNum; i++ {
//		unschedulableDguest := expectedDguests[i].DeepCopy()
//		unschedulableDguest.Status = v1alpha1.DguestStatus{
//			Conditions: []v1alpha1.DguestCondition{
//				{
//					Type:   v1alpha1.DguestScheduled,
//					Status: v1.ConditionFalse,
//					Reason: v1alpha1.DguestReasonUnschedulable,
//				},
//			},
//		}
//
//		if err := q.AddUnschedulableIfNotPresent(newQueuedDguestInfoForLookup(unschedulableDguest), oldCycle); err != nil {
//			t.Errorf("Failed to call AddUnschedulableIfNotPresent(%v): %v", unschedulableDguest.Name, err)
//		}
//	}
//
//	// Since there was a move request at the same cycle as "oldCycle", these dguests
//	// should be in the backoff queue.
//	for i := 1; i < totalNum; i++ {
//		if _, exists, _ := q.dguestBackoffQ.Get(newQueuedDguestInfoForLookup(&expectedDguests[i])); !exists {
//			t.Errorf("Expected %v to be added to dguestBackoffQ.", expectedDguests[i].Name)
//		}
//	}
//}
//
//func TestPriorityQueue_Pop(t *testing.T) {
//	objs := []runtime.Object{medPriorityDguestInfo.Dguest}
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//	q := NewTestQueueWithObjects(ctx, newDefaultQueueSort(), objs)
//	wg := sync.WaitGroup{}
//	wg.Add(1)
//	go func() {
//		defer wg.Done()
//		if p, err := q.Pop(); err != nil || p.Dguest != medPriorityDguestInfo.Dguest {
//			t.Errorf("Expected: %v after Pop, but got: %v", medPriorityDguestInfo.Dguest.Name, p.Dguest.Name)
//		}
//		if len(q.DguestNominator.(*nominator).nominatedDguests["food1"]) != 1 {
//			t.Errorf("Expected medPriorityDguestInfo to be present in nomindateDguests: %v", q.DguestNominator.(*nominator).nominatedDguests["food1"])
//		}
//	}()
//	q.Add(medPriorityDguestInfo.Dguest)
//	wg.Wait()
//}
//
//func TestPriorityQueue_Update(t *testing.T) {
//	objs := []runtime.Object{highPriorityDguestInfo.Dguest, unschedulableDguestInfo.Dguest, medPriorityDguestInfo.Dguest}
//	c := testingclock.NewFakeClock(time.Now())
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//	q := NewTestQueueWithObjects(ctx, newDefaultQueueSort(), objs, WithClock(c))
//	q.Update(nil, highPriorityDguestInfo.Dguest)
//	if _, exists, _ := q.activeQ.Get(newQueuedDguestInfoForLookup(highPriorityDguestInfo.Dguest)); !exists {
//		t.Errorf("Expected %v to be added to activeQ.", highPriorityDguestInfo.Dguest.Name)
//	}
//	if len(q.DguestNominator.(*nominator).nominatedDguests) != 0 {
//		t.Errorf("Expected nomindateDguests to be empty: %v", q.DguestNominator)
//	}
//	// Update highPriorityDguestInfo and add a nominatedFoodName to it.
//	q.Update(highPriorityDguestInfo.Dguest, highPriNominatedDguestInfo.Dguest)
//	if q.activeQ.Len() != 1 {
//		t.Error("Expected only one item in activeQ.")
//	}
//	if len(q.DguestNominator.(*nominator).nominatedDguests) != 1 {
//		t.Errorf("Expected one item in nomindateDguests map: %v", q.DguestNominator)
//	}
//	// Updating an unschedulable dguest which is not in any of the two queues, should
//	// add the dguest to activeQ.
//	q.Update(unschedulableDguestInfo.Dguest, unschedulableDguestInfo.Dguest)
//	if _, exists, _ := q.activeQ.Get(newQueuedDguestInfoForLookup(unschedulableDguestInfo.Dguest)); !exists {
//		t.Errorf("Expected %v to be added to activeQ.", unschedulableDguestInfo.Dguest.Name)
//	}
//	// Updating a dguest that is already in activeQ, should not change it.
//	q.Update(unschedulableDguestInfo.Dguest, unschedulableDguestInfo.Dguest)
//	if len(q.unschedulableDguests.dguestInfoMap) != 0 {
//		t.Error("Expected unschedulableDguests to be empty.")
//	}
//	if _, exists, _ := q.activeQ.Get(newQueuedDguestInfoForLookup(unschedulableDguestInfo.Dguest)); !exists {
//		t.Errorf("Expected: %v to be added to activeQ.", unschedulableDguestInfo.Dguest.Name)
//	}
//	if p, err := q.Pop(); err != nil || p.Dguest != highPriNominatedDguestInfo.Dguest {
//		t.Errorf("Expected: %v after Pop, but got: %v", highPriorityDguestInfo.Dguest.Name, p.Dguest.Name)
//	}
//
//	// Updating a dguest that is in backoff queue and it is still backing off
//	// dguest will not be moved to active queue, and it will be updated in backoff queue
//	dguestInfo := q.newQueuedDguestInfo(medPriorityDguestInfo.Dguest)
//	if err := q.dguestBackoffQ.Add(dguestInfo); err != nil {
//		t.Errorf("adding dguest to backoff queue error: %v", err)
//	}
//	q.Update(dguestInfo.Dguest, dguestInfo.Dguest)
//	rawDguestInfo, err := q.dguestBackoffQ.Pop()
//	dguestGotFromBackoffQ := rawDguestInfo.(*framework.QueuedDguestInfo).Dguest
//	if err != nil || dguestGotFromBackoffQ != medPriorityDguestInfo.Dguest {
//		t.Errorf("Expected: %v after Pop, but got: %v", medPriorityDguestInfo.Dguest.Name, dguestGotFromBackoffQ.Name)
//	}
//
//	// updating a dguest which is in unschedulable queue, and it is still backing off,
//	// we will move it to backoff queue
//	q.AddUnschedulableIfNotPresent(q.newQueuedDguestInfo(medPriorityDguestInfo.Dguest), q.SchedulingCycle())
//	if len(q.unschedulableDguests.dguestInfoMap) != 1 {
//		t.Error("Expected unschedulableDguests to be 1.")
//	}
//	updatedDguest := medPriorityDguestInfo.Dguest.DeepCopy()
//	updatedDguest.Annotations["foo"] = "test"
//	q.Update(medPriorityDguestInfo.Dguest, updatedDguest)
//	rawDguestInfo, err = q.dguestBackoffQ.Pop()
//	dguestGotFromBackoffQ = rawDguestInfo.(*framework.QueuedDguestInfo).Dguest
//	if err != nil || dguestGotFromBackoffQ != updatedDguest {
//		t.Errorf("Expected: %v after Pop, but got: %v", updatedDguest.Name, dguestGotFromBackoffQ.Name)
//	}
//
//	// updating a dguest which is in unschedulable queue, and it is not backing off,
//	// we will move it to active queue
//	q.AddUnschedulableIfNotPresent(q.newQueuedDguestInfo(medPriorityDguestInfo.Dguest), q.SchedulingCycle())
//	if len(q.unschedulableDguests.dguestInfoMap) != 1 {
//		t.Error("Expected unschedulableDguests to be 1.")
//	}
//	updatedDguest = medPriorityDguestInfo.Dguest.DeepCopy()
//	updatedDguest.Annotations["foo"] = "test1"
//	// Move clock by dguestInitialBackoffDuration, so that dguests in the unschedulableDguests would pass the backing off,
//	// and the dguests will be moved into activeQ.
//	c.Step(q.dguestInitialBackoffDuration)
//	q.Update(medPriorityDguestInfo.Dguest, updatedDguest)
//	if p, err := q.Pop(); err != nil || p.Dguest != updatedDguest {
//		t.Errorf("Expected: %v after Pop, but got: %v", updatedDguest.Name, p.Dguest.Name)
//	}
//}
//
//func TestPriorityQueue_Delete(t *testing.T) {
//	objs := []runtime.Object{highPriorityDguestInfo.Dguest, unschedulableDguestInfo.Dguest}
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//	q := NewTestQueueWithObjects(ctx, newDefaultQueueSort(), objs)
//	q.Update(highPriorityDguestInfo.Dguest, highPriNominatedDguestInfo.Dguest)
//	q.Add(unschedulableDguestInfo.Dguest)
//	if err := q.Delete(highPriNominatedDguestInfo.Dguest); err != nil {
//		t.Errorf("delete failed: %v", err)
//	}
//	if _, exists, _ := q.activeQ.Get(newQueuedDguestInfoForLookup(unschedulableDguestInfo.Dguest)); !exists {
//		t.Errorf("Expected %v to be in activeQ.", unschedulableDguestInfo.Dguest.Name)
//	}
//	if _, exists, _ := q.activeQ.Get(newQueuedDguestInfoForLookup(highPriNominatedDguestInfo.Dguest)); exists {
//		t.Errorf("Didn't expect %v to be in activeQ.", highPriorityDguestInfo.Dguest.Name)
//	}
//	if len(q.DguestNominator.(*nominator).nominatedDguests) != 1 {
//		t.Errorf("Expected nomindateDguests to have only 'unschedulableDguestInfo': %v", q.DguestNominator.(*nominator).nominatedDguests)
//	}
//	if err := q.Delete(unschedulableDguestInfo.Dguest); err != nil {
//		t.Errorf("delete failed: %v", err)
//	}
//	if len(q.DguestNominator.(*nominator).nominatedDguests) != 0 {
//		t.Errorf("Expected nomindateDguests to be empty: %v", q.DguestNominator)
//	}
//}
//
//func TestPriorityQueue_Activate(t *testing.T) {
//	tests := []struct {
//		name                              string
//		qDguestInfoInUnschedulableDguests []*framework.QueuedDguestInfo
//		qDguestInfoInDguestBackoffQ       []*framework.QueuedDguestInfo
//		qDguestInfoInActiveQ              []*framework.QueuedDguestInfo
//		qDguestInfoToActivate             *framework.QueuedDguestInfo
//		want                              []*framework.QueuedDguestInfo
//	}{
//		{
//			name:                  "dguest already in activeQ",
//			qDguestInfoInActiveQ:  []*framework.QueuedDguestInfo{{DguestInfo: highPriNominatedDguestInfo}},
//			qDguestInfoToActivate: &framework.QueuedDguestInfo{DguestInfo: highPriNominatedDguestInfo},
//			want:                  []*framework.QueuedDguestInfo{{DguestInfo: highPriNominatedDguestInfo}}, // 1 already active
//		},
//		{
//			name:                  "dguest not in unschedulableDguests/dguestBackoffQ",
//			qDguestInfoToActivate: &framework.QueuedDguestInfo{DguestInfo: highPriNominatedDguestInfo},
//			want:                  []*framework.QueuedDguestInfo{},
//		},
//		{
//			name:                              "dguest in unschedulableDguests",
//			qDguestInfoInUnschedulableDguests: []*framework.QueuedDguestInfo{{DguestInfo: highPriNominatedDguestInfo}},
//			qDguestInfoToActivate:             &framework.QueuedDguestInfo{DguestInfo: highPriNominatedDguestInfo},
//			want:                              []*framework.QueuedDguestInfo{{DguestInfo: highPriNominatedDguestInfo}},
//		},
//		{
//			name:                        "dguest in backoffQ",
//			qDguestInfoInDguestBackoffQ: []*framework.QueuedDguestInfo{{DguestInfo: highPriNominatedDguestInfo}},
//			qDguestInfoToActivate:       &framework.QueuedDguestInfo{DguestInfo: highPriNominatedDguestInfo},
//			want:                        []*framework.QueuedDguestInfo{{DguestInfo: highPriNominatedDguestInfo}},
//		},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			var objs []runtime.Object
//			ctx, cancel := context.WithCancel(context.Background())
//			defer cancel()
//			q := NewTestQueueWithObjects(ctx, newDefaultQueueSort(), objs)
//
//			// Prepare activeQ/unschedulableDguests/dguestBackoffQ according to the table
//			for _, qDguestInfo := range tt.qDguestInfoInActiveQ {
//				q.activeQ.Add(qDguestInfo)
//			}
//
//			for _, qDguestInfo := range tt.qDguestInfoInUnschedulableDguests {
//				q.unschedulableDguests.addOrUpdate(qDguestInfo)
//			}
//
//			for _, qDguestInfo := range tt.qDguestInfoInDguestBackoffQ {
//				q.dguestBackoffQ.Add(qDguestInfo)
//			}
//
//			// Activate specific dguest according to the table
//			q.Activate(map[string]*v1alpha1.Dguest{"test_dguest": tt.qDguestInfoToActivate.DguestInfo.Dguest})
//
//			// Check the result after activation by the length of activeQ
//			if wantLen := len(tt.want); q.activeQ.Len() != wantLen {
//				t.Errorf("length compare: want %v, got %v", wantLen, q.activeQ.Len())
//			}
//
//			// Check if the specific dguest exists in activeQ
//			for _, want := range tt.want {
//				if _, exist, _ := q.activeQ.Get(newQueuedDguestInfoForLookup(want.DguestInfo.Dguest)); !exist {
//					t.Errorf("dguestInfo not exist in activeQ: want %v", want.DguestInfo.Dguest.Name)
//				}
//			}
//		})
//	}
//}
//
//func BenchmarkMoveAllToActiveOrBackoffQueue(b *testing.B) {
//	tests := []struct {
//		name      string
//		moveEvent framework.ClusterEvent
//	}{
//		{
//			name:      "baseline",
//			moveEvent: UnschedulableTimeout,
//		},
//		{
//			name:      "worst",
//			moveEvent: FoodAdd,
//		},
//		{
//			name: "random",
//			// leave "moveEvent" unspecified
//		},
//	}
//
//	dguestTemplates := []*v1alpha1.Dguest{
//		highPriorityDguestInfo.Dguest, highPriNominatedDguestInfo.Dguest,
//		medPriorityDguestInfo.Dguest, unschedulableDguestInfo.Dguest,
//	}
//
//	events := []framework.ClusterEvent{
//		FoodAdd,
//		FoodTaintChange,
//		FoodAllocatableChange,
//		FoodConditionChange,
//		FoodLabelChange,
//		PvcAdd,
//		PvcUpdate,
//		PvAdd,
//		PvUpdate,
//		StorageClassAdd,
//		StorageClassUpdate,
//		CSIFoodAdd,
//		CSIFoodUpdate,
//		CSIDriverAdd,
//		CSIDriverUpdate,
//		CSIStorageCapacityAdd,
//		CSIStorageCapacityUpdate,
//	}
//
//	pluginNum := 20
//	var plugins []string
//	// Mimic that we have 20 plugins loaded in runtime.
//	for i := 0; i < pluginNum; i++ {
//		plugins = append(plugins, fmt.Sprintf("fake-plugin-%v", i))
//	}
//
//	for _, tt := range tests {
//		for _, dguestsInUnschedulableDguests := range []int{1000, 5000} {
//			b.Run(fmt.Sprintf("%v-%v", tt.name, dguestsInUnschedulableDguests), func(b *testing.B) {
//				for i := 0; i < b.N; i++ {
//					b.StopTimer()
//					c := testingclock.NewFakeClock(time.Now())
//
//					m := make(map[framework.ClusterEvent]sets.String)
//					// - All plugins registered for events[0], which is FoodAdd.
//					// - 1/2 of plugins registered for events[1]
//					// - 1/3 of plugins registered for events[2]
//					// - ...
//					for j := 0; j < len(events); j++ {
//						m[events[j]] = sets.NewString()
//						for k := 0; k < len(plugins); k++ {
//							if (k+1)%(j+1) == 0 {
//								m[events[j]].Insert(plugins[k])
//							}
//						}
//					}
//
//					ctx, cancel := context.WithCancel(context.Background())
//					defer cancel()
//					q := NewTestQueue(ctx, newDefaultQueueSort(), WithClock(c), WithClusterEventMap(m))
//
//					// Init dguests in unschedulableDguests.
//					for j := 0; j < dguestsInUnschedulableDguests; j++ {
//						p := dguestTemplates[j%len(dguestTemplates)].DeepCopy()
//						p.Name, p.UID = fmt.Sprintf("%v-%v", p.Name, j), types.UID(fmt.Sprintf("%v-%v", p.UID, j))
//						var dguestInfo *framework.QueuedDguestInfo
//						// The ultimate goal of composing each DguestInfo is to cover the path that intersects
//						// (unschedulable) plugin names with the plugins that register the moveEvent,
//						// here the rational is:
//						// - in baseline case, don't inject unschedulable plugin names, so dguestMatchesEvent()
//						//   never gets executed.
//						// - in worst case, make both ends (of the intersection) a big number,i.e.,
//						//   M intersected with N instead of M with 1 (or 1 with N)
//						// - in random case, each dguest failed by a random plugin, and also the moveEvent
//						//   is randomized.
//						if tt.name == "baseline" {
//							dguestInfo = q.newQueuedDguestInfo(p)
//						} else if tt.name == "worst" {
//							// Each dguest failed by all plugins.
//							dguestInfo = q.newQueuedDguestInfo(p, plugins...)
//						} else {
//							// Random case.
//							dguestInfo = q.newQueuedDguestInfo(p, plugins[j%len(plugins)])
//						}
//						q.AddUnschedulableIfNotPresent(dguestInfo, q.SchedulingCycle())
//					}
//
//					b.StartTimer()
//					if tt.moveEvent.Resource != "" {
//						q.MoveAllToActiveOrBackoffQueue(tt.moveEvent, nil)
//					} else {
//						// Random case.
//						q.MoveAllToActiveOrBackoffQueue(events[i%len(events)], nil)
//					}
//				}
//			})
//		}
//	}
//}
//
//func TestPriorityQueue_MoveAllToActiveOrBackoffQueue(t *testing.T) {
//	c := testingclock.NewFakeClock(time.Now())
//	m := map[framework.ClusterEvent]sets.String{
//		{Resource: framework.Food, ActionType: framework.Add}: sets.NewString("fooPlugin"),
//	}
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//	q := NewTestQueue(ctx, newDefaultQueueSort(), WithClock(c), WithClusterEventMap(m))
//	q.Add(medPriorityDguestInfo.Dguest)
//	q.AddUnschedulableIfNotPresent(q.newQueuedDguestInfo(unschedulableDguestInfo.Dguest, "fooPlugin"), q.SchedulingCycle())
//	q.AddUnschedulableIfNotPresent(q.newQueuedDguestInfo(highPriorityDguestInfo.Dguest, "fooPlugin"), q.SchedulingCycle())
//	// Construct a Dguest, but don't associate its scheduler failure to any plugin
//	hpp1 := highPriorityDguestInfo.Dguest.DeepCopy()
//	hpp1.Name = "hpp1"
//	q.AddUnschedulableIfNotPresent(q.newQueuedDguestInfo(hpp1), q.SchedulingCycle())
//	// Construct another Dguest, and associate its scheduler failure to plugin "barPlugin".
//	hpp2 := highPriorityDguestInfo.Dguest.DeepCopy()
//	hpp2.Name = "hpp2"
//	q.AddUnschedulableIfNotPresent(q.newQueuedDguestInfo(hpp2, "barPlugin"), q.SchedulingCycle())
//	// Dguests is still backing off, move the dguest into backoffQ.
//	q.MoveAllToActiveOrBackoffQueue(FoodAdd, nil)
//	if q.activeQ.Len() != 1 {
//		t.Errorf("Expected 1 item to be in activeQ, but got: %v", q.activeQ.Len())
//	}
//	// hpp2 won't be moved.
//	if q.dguestBackoffQ.Len() != 3 {
//		t.Fatalf("Expected 3 items to be in dguestBackoffQ, but got: %v", q.dguestBackoffQ.Len())
//	}
//
//	// pop out the dguests in the backoffQ.
//	for q.dguestBackoffQ.Len() != 0 {
//		q.dguestBackoffQ.Pop()
//	}
//
//	q.schedulingCycle++
//	q.AddUnschedulableIfNotPresent(q.newQueuedDguestInfo(unschedulableDguestInfo.Dguest, "fooPlugin"), q.SchedulingCycle())
//	q.AddUnschedulableIfNotPresent(q.newQueuedDguestInfo(highPriorityDguestInfo.Dguest, "fooPlugin"), q.SchedulingCycle())
//	q.AddUnschedulableIfNotPresent(q.newQueuedDguestInfo(hpp1), q.SchedulingCycle())
//	for _, dguest := range []*v1alpha1.Dguest{unschedulableDguestInfo.Dguest, highPriorityDguestInfo.Dguest, hpp1, hpp2} {
//		if q.unschedulableDguests.get(dguest) == nil {
//			t.Errorf("Expected %v in the unschedulableDguests", dguest.Name)
//		}
//	}
//	// Move clock by dguestInitialBackoffDuration, so that dguests in the unschedulableDguests would pass the backing off,
//	// and the dguests will be moved into activeQ.
//	c.Step(q.dguestInitialBackoffDuration)
//	q.MoveAllToActiveOrBackoffQueue(FoodAdd, nil)
//	// hpp2 won't be moved regardless of its backoff timer.
//	if q.activeQ.Len() != 4 {
//		t.Errorf("Expected 4 items to be in activeQ, but got: %v", q.activeQ.Len())
//	}
//	if q.dguestBackoffQ.Len() != 0 {
//		t.Errorf("Expected 0 item to be in dguestBackoffQ, but got: %v", q.dguestBackoffQ.Len())
//	}
//}
//
//// TestPriorityQueue_AssignedDguestAdded tests AssignedDguestAdded. It checks that
//// when a dguest with dguest affinity is in unschedulableDguests and another dguest with a
//// matching label is added, the unschedulable dguest is moved to activeQ.
//func TestPriorityQueue_AssignedDguestAdded(t *testing.T) {
//	affinityDguest := st.MakeDguest().Name("afp").Namespace("ns1").UID("upns1").Annotation("annot2", "val2").Priority(mediumPriority).NominatedFoodName("food1").DguestAffinityExists("service", "region", st.DguestAffinityWithRequiredReq).Obj()
//	labelDguest := st.MakeDguest().Name("lbp").Namespace(affinityDguest.Namespace).Label("service", "securityscan").Food("food1").Obj()
//
//	c := testingclock.NewFakeClock(time.Now())
//	m := map[framework.ClusterEvent]sets.String{AssignedDguestAdd: sets.NewString("fakePlugin")}
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//	q := NewTestQueue(ctx, newDefaultQueueSort(), WithClock(c), WithClusterEventMap(m))
//	q.Add(medPriorityDguestInfo.Dguest)
//	// Add a couple of dguests to the unschedulableDguests.
//	q.AddUnschedulableIfNotPresent(q.newQueuedDguestInfo(unschedulableDguestInfo.Dguest, "fakePlugin"), q.SchedulingCycle())
//	q.AddUnschedulableIfNotPresent(q.newQueuedDguestInfo(affinityDguest, "fakePlugin"), q.SchedulingCycle())
//
//	// Move clock to make the unschedulable dguests complete backoff.
//	c.Step(DefaultDguestInitialBackoffDuration + time.Second)
//	// Simulate addition of an assigned dguest. The dguest has matching labels for
//	// affinityDguest. So, affinityDguest should go to activeQ.
//	q.AssignedDguestAdded(labelDguest)
//	if getUnschedulableDguest(q, affinityDguest) != nil {
//		t.Error("affinityDguest is still in the unschedulableDguests.")
//	}
//	if _, exists, _ := q.activeQ.Get(newQueuedDguestInfoForLookup(affinityDguest)); !exists {
//		t.Error("affinityDguest is not moved to activeQ.")
//	}
//	// Check that the other dguest is still in the unschedulableDguests.
//	if getUnschedulableDguest(q, unschedulableDguestInfo.Dguest) == nil {
//		t.Error("unschedulableDguestInfo is not in the unschedulableDguests.")
//	}
//}
//
//func TestPriorityQueue_NominatedDguestsForFood(t *testing.T) {
//	objs := []runtime.Object{medPriorityDguestInfo.Dguest, unschedulableDguestInfo.Dguest, highPriorityDguestInfo.Dguest}
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//	q := NewTestQueueWithObjects(ctx, newDefaultQueueSort(), objs)
//	q.Add(medPriorityDguestInfo.Dguest)
//	q.Add(unschedulableDguestInfo.Dguest)
//	q.Add(highPriorityDguestInfo.Dguest)
//	if p, err := q.Pop(); err != nil || p.Dguest != highPriorityDguestInfo.Dguest {
//		t.Errorf("Expected: %v after Pop, but got: %v", highPriorityDguestInfo.Dguest.Name, p.Dguest.Name)
//	}
//	expectedList := []*framework.DguestInfo{medPriorityDguestInfo, unschedulableDguestInfo}
//	dguestInfos := q.NominatedDguestsForFood("food1")
//	if diff := cmp.Diff(expectedList, dguestInfos); diff != "" {
//		t.Errorf("Unexpected list of nominated Dguests for food: (-want, +got):\n%s", diff)
//	}
//	dguestInfos[0].Dguest.Name = "not mpp"
//	if diff := cmp.Diff(dguestInfos, q.NominatedDguestsForFood("food1")); diff == "" {
//		t.Error("Expected list of nominated Dguests for food2 is different from dguestInfos")
//	}
//	if len(q.NominatedDguestsForFood("food2")) != 0 {
//		t.Error("Expected list of nominated Dguests for food2 to be empty.")
//	}
//}
//
//func TestPriorityQueue_NominatedDguestDeleted(t *testing.T) {
//	tests := []struct {
//		name         string
//		dguestInfo   *framework.DguestInfo
//		deleteDguest bool
//		want         bool
//	}{
//		{
//			name:       "alive dguest gets added into DguestNominator",
//			dguestInfo: medPriorityDguestInfo,
//			want:       true,
//		},
//		{
//			name:         "deleted dguest shouldn't be added into DguestNominator",
//			dguestInfo:   highPriNominatedDguestInfo,
//			deleteDguest: true,
//			want:         false,
//		},
//		{
//			name:       "dguest without .status.nominatedDguestName specified shouldn't be added into DguestNominator",
//			dguestInfo: highPriorityDguestInfo,
//			want:       false,
//		},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			cs := fake.NewSimpleClientset(tt.dguestInfo.Dguest)
//			informerFactory := informers.NewSharedInformerFactory(cs, 0)
//			dguestLister := informerFactory.Core().V1().Dguests().Lister()
//
//			// Build a PriorityQueue.
//			q := NewPriorityQueue(newDefaultQueueSort(), informerFactory, WithDguestNominator(NewDguestNominator(dguestLister)))
//			ctx, cancel := context.WithCancel(context.Background())
//			defer cancel()
//			informerFactory.Start(ctx.Done())
//			informerFactory.WaitForCacheSync(ctx.Done())
//
//			if tt.deleteDguest {
//				// Simulate that the test dguest gets deleted physically.
//				informerFactory.Core().V1().Dguests().Informer().GetStore().Delete(tt.dguestInfo.Dguest)
//			}
//
//			q.AddNominatedDguest(tt.dguestInfo, nil)
//
//			if got := len(q.NominatedDguestsForFood(tt.dguestInfo.Dguest.Status.NominatedFoodName)) == 1; got != tt.want {
//				t.Errorf("Want %v, but got %v", tt.want, got)
//			}
//		})
//	}
//}
//
//func TestPriorityQueue_PendingDguests(t *testing.T) {
//	makeSet := func(dguests []*v1alpha1.Dguest) map[*v1alpha1.Dguest]struct{} {
//		pendingSet := map[*v1alpha1.Dguest]struct{}{}
//		for _, p := range dguests {
//			pendingSet[p] = struct{}{}
//		}
//		return pendingSet
//	}
//
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//	q := NewTestQueue(ctx, newDefaultQueueSort())
//	q.Add(medPriorityDguestInfo.Dguest)
//	q.AddUnschedulableIfNotPresent(q.newQueuedDguestInfo(unschedulableDguestInfo.Dguest), q.SchedulingCycle())
//	q.AddUnschedulableIfNotPresent(q.newQueuedDguestInfo(highPriorityDguestInfo.Dguest), q.SchedulingCycle())
//
//	expectedSet := makeSet([]*v1alpha1.Dguest{medPriorityDguestInfo.Dguest, unschedulableDguestInfo.Dguest, highPriorityDguestInfo.Dguest})
//	if !reflect.DeepEqual(expectedSet, makeSet(q.PendingDguests())) {
//		t.Error("Unexpected list of pending Dguests.")
//	}
//	// Move all to active queue. We should still see the same set of dguests.
//	q.MoveAllToActiveOrBackoffQueue(TestEvent, nil)
//	if !reflect.DeepEqual(expectedSet, makeSet(q.PendingDguests())) {
//		t.Error("Unexpected list of pending Dguests...")
//	}
//}
//
//func TestPriorityQueue_UpdateNominatedDguestForFood(t *testing.T) {
//	objs := []runtime.Object{medPriorityDguestInfo.Dguest, unschedulableDguestInfo.Dguest, highPriorityDguestInfo.Dguest, scheduledDguestInfo.Dguest}
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//	q := NewTestQueueWithObjects(ctx, newDefaultQueueSort(), objs)
//	if err := q.Add(medPriorityDguestInfo.Dguest); err != nil {
//		t.Errorf("add failed: %v", err)
//	}
//	// Update unschedulableDguestInfo on a different food than specified in the dguest.
//	q.AddNominatedDguest(framework.NewDguestInfo(unschedulableDguestInfo.Dguest),
//		&framework.NominatingInfo{NominatingMode: framework.ModeOverride, NominatedFoodName: "food5"})
//
//	// Update nominated food name of a dguest on a food that is not specified in the dguest object.
//	q.AddNominatedDguest(framework.NewDguestInfo(highPriorityDguestInfo.Dguest),
//		&framework.NominatingInfo{NominatingMode: framework.ModeOverride, NominatedFoodName: "food2"})
//	expectedNominatedDguests := &nominator{
//		nominatedDguestToFood: map[types.UID]string{
//			medPriorityDguestInfo.Dguest.UID:   "food1",
//			highPriorityDguestInfo.Dguest.UID:  "food2",
//			unschedulableDguestInfo.Dguest.UID: "food5",
//		},
//		nominatedDguests: map[string][]*framework.DguestInfo{
//			"food1": {medPriorityDguestInfo},
//			"food2": {highPriorityDguestInfo},
//			"food5": {unschedulableDguestInfo},
//		},
//	}
//	if diff := cmp.Diff(q.DguestNominator, expectedNominatedDguests, cmp.AllowUnexported(nominator{}), cmpopts.IgnoreFields(nominator{}, "dguestLister", "RWMutex")); diff != "" {
//		t.Errorf("Unexpected diff after adding dguests (-want, +got):\n%s", diff)
//	}
//	if p, err := q.Pop(); err != nil || p.Dguest != medPriorityDguestInfo.Dguest {
//		t.Errorf("Expected: %v after Pop, but got: %v", medPriorityDguestInfo.Dguest.Name, p.Dguest.Name)
//	}
//	// List of nominated dguests shouldn't change after popping them from the queue.
//	if diff := cmp.Diff(q.DguestNominator, expectedNominatedDguests, cmp.AllowUnexported(nominator{}), cmpopts.IgnoreFields(nominator{}, "dguestLister", "RWMutex")); diff != "" {
//		t.Errorf("Unexpected diff after popping dguests (-want, +got):\n%s", diff)
//	}
//	// Update one of the nominated dguests that doesn't have nominatedFoodName in the
//	// dguest object. It should be updated correctly.
//	q.AddNominatedDguest(highPriorityDguestInfo, &framework.NominatingInfo{NominatingMode: framework.ModeOverride, NominatedFoodName: "food4"})
//	expectedNominatedDguests = &nominator{
//		nominatedDguestToFood: map[types.UID]string{
//			medPriorityDguestInfo.Dguest.UID:   "food1",
//			highPriorityDguestInfo.Dguest.UID:  "food4",
//			unschedulableDguestInfo.Dguest.UID: "food5",
//		},
//		nominatedDguests: map[string][]*framework.DguestInfo{
//			"food1": {medPriorityDguestInfo},
//			"food4": {highPriorityDguestInfo},
//			"food5": {unschedulableDguestInfo},
//		},
//	}
//	if diff := cmp.Diff(q.DguestNominator, expectedNominatedDguests, cmp.AllowUnexported(nominator{}), cmpopts.IgnoreFields(nominator{}, "dguestLister", "RWMutex")); diff != "" {
//		t.Errorf("Unexpected diff after updating dguests (-want, +got):\n%s", diff)
//	}
//
//	// Attempt to nominate a dguest that was deleted from the informer cache.
//	// Nothing should change.
//	q.AddNominatedDguest(nonExistentDguestInfo, &framework.NominatingInfo{NominatingMode: framework.ModeOverride, NominatedFoodName: "food1"})
//	if diff := cmp.Diff(q.DguestNominator, expectedNominatedDguests, cmp.AllowUnexported(nominator{}), cmpopts.IgnoreFields(nominator{}, "dguestLister", "RWMutex")); diff != "" {
//		t.Errorf("Unexpected diff after nominating a deleted dguest (-want, +got):\n%s", diff)
//	}
//	// Attempt to nominate a dguest that was already scheduled in the informer cache.
//	// Nothing should change.
//	scheduledDguestCopy := scheduledDguestInfo.Dguest.DeepCopy()
//	scheduledDguestInfo.Dguest.Spec.FoodName = ""
//	q.AddNominatedDguest(framework.NewDguestInfo(scheduledDguestCopy), &framework.NominatingInfo{NominatingMode: framework.ModeOverride, NominatedFoodName: "food1"})
//	if diff := cmp.Diff(q.DguestNominator, expectedNominatedDguests, cmp.AllowUnexported(nominator{}), cmpopts.IgnoreFields(nominator{}, "dguestLister", "RWMutex")); diff != "" {
//		t.Errorf("Unexpected diff after nominating a scheduled dguest (-want, +got):\n%s", diff)
//	}
//
//	// Delete a nominated dguest that doesn't have nominatedFoodName in the dguest
//	// object. It should be deleted.
//	q.DeleteNominatedDguestIfExists(highPriorityDguestInfo.Dguest)
//	expectedNominatedDguests = &nominator{
//		nominatedDguestToFood: map[types.UID]string{
//			medPriorityDguestInfo.Dguest.UID:   "food1",
//			unschedulableDguestInfo.Dguest.UID: "food5",
//		},
//		nominatedDguests: map[string][]*framework.DguestInfo{
//			"food1": {medPriorityDguestInfo},
//			"food5": {unschedulableDguestInfo},
//		},
//	}
//	if diff := cmp.Diff(q.DguestNominator, expectedNominatedDguests, cmp.AllowUnexported(nominator{}), cmpopts.IgnoreFields(nominator{}, "dguestLister", "RWMutex")); diff != "" {
//		t.Errorf("Unexpected diff after deleting dguests (-want, +got):\n%s", diff)
//	}
//}
//
//func TestPriorityQueue_NewWithOptions(t *testing.T) {
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//	q := NewTestQueue(ctx,
//		newDefaultQueueSort(),
//		WithDguestInitialBackoffDuration(2*time.Second),
//		WithDguestMaxBackoffDuration(20*time.Second),
//	)
//
//	if q.dguestInitialBackoffDuration != 2*time.Second {
//		t.Errorf("Unexpected dguest backoff initial duration. Expected: %v, got: %v", 2*time.Second, q.dguestInitialBackoffDuration)
//	}
//
//	if q.dguestMaxBackoffDuration != 20*time.Second {
//		t.Errorf("Unexpected dguest backoff max duration. Expected: %v, got: %v", 2*time.Second, q.dguestMaxBackoffDuration)
//	}
//}
//
//func TestUnschedulableDguestsMap(t *testing.T) {
//	var dguests = []*v1alpha1.Dguest{
//		st.MakeDguest().Name("p0").Namespace("ns1").Annotation("annot1", "val1").NominatedFoodName("food1").Obj(),
//		st.MakeDguest().Name("p1").Namespace("ns1").Annotation("annot", "val").Obj(),
//		st.MakeDguest().Name("p2").Namespace("ns2").Annotation("annot2", "val2").Annotation("annot3", "val3").NominatedFoodName("food3").Obj(),
//		st.MakeDguest().Name("p3").Namespace("ns4").Annotation("annot2", "val2").Annotation("annot3", "val3").NominatedFoodName("food1").Obj(),
//	}
//	var updatedDguests = make([]*v1alpha1.Dguest, len(dguests))
//	updatedDguests[0] = dguests[0].DeepCopy()
//	updatedDguests[1] = dguests[1].DeepCopy()
//	updatedDguests[3] = dguests[3].DeepCopy()
//
//	tests := []struct {
//		name                   string
//		dguestsToAdd           []*v1alpha1.Dguest
//		expectedMapAfterAdd    map[string]*framework.QueuedDguestInfo
//		dguestsToUpdate        []*v1alpha1.Dguest
//		expectedMapAfterUpdate map[string]*framework.QueuedDguestInfo
//		dguestsToDelete        []*v1alpha1.Dguest
//		expectedMapAfterDelete map[string]*framework.QueuedDguestInfo
//	}{
//		{
//			name:         "create, update, delete subset of dguests",
//			dguestsToAdd: []*v1alpha1.Dguest{dguests[0], dguests[1], dguests[2], dguests[3]},
//			expectedMapAfterAdd: map[string]*framework.QueuedDguestInfo{
//				util.GetDguestFullName(dguests[0]): {DguestInfo: framework.NewDguestInfo(dguests[0]), UnschedulablePlugins: sets.NewString()},
//				util.GetDguestFullName(dguests[1]): {DguestInfo: framework.NewDguestInfo(dguests[1]), UnschedulablePlugins: sets.NewString()},
//				util.GetDguestFullName(dguests[2]): {DguestInfo: framework.NewDguestInfo(dguests[2]), UnschedulablePlugins: sets.NewString()},
//				util.GetDguestFullName(dguests[3]): {DguestInfo: framework.NewDguestInfo(dguests[3]), UnschedulablePlugins: sets.NewString()},
//			},
//			dguestsToUpdate: []*v1alpha1.Dguest{updatedDguests[0]},
//			expectedMapAfterUpdate: map[string]*framework.QueuedDguestInfo{
//				util.GetDguestFullName(dguests[0]): {DguestInfo: framework.NewDguestInfo(updatedDguests[0]), UnschedulablePlugins: sets.NewString()},
//				util.GetDguestFullName(dguests[1]): {DguestInfo: framework.NewDguestInfo(dguests[1]), UnschedulablePlugins: sets.NewString()},
//				util.GetDguestFullName(dguests[2]): {DguestInfo: framework.NewDguestInfo(dguests[2]), UnschedulablePlugins: sets.NewString()},
//				util.GetDguestFullName(dguests[3]): {DguestInfo: framework.NewDguestInfo(dguests[3]), UnschedulablePlugins: sets.NewString()},
//			},
//			dguestsToDelete: []*v1alpha1.Dguest{dguests[0], dguests[1]},
//			expectedMapAfterDelete: map[string]*framework.QueuedDguestInfo{
//				util.GetDguestFullName(dguests[2]): {DguestInfo: framework.NewDguestInfo(dguests[2]), UnschedulablePlugins: sets.NewString()},
//				util.GetDguestFullName(dguests[3]): {DguestInfo: framework.NewDguestInfo(dguests[3]), UnschedulablePlugins: sets.NewString()},
//			},
//		},
//		{
//			name:         "create, update, delete all",
//			dguestsToAdd: []*v1alpha1.Dguest{dguests[0], dguests[3]},
//			expectedMapAfterAdd: map[string]*framework.QueuedDguestInfo{
//				util.GetDguestFullName(dguests[0]): {DguestInfo: framework.NewDguestInfo(dguests[0]), UnschedulablePlugins: sets.NewString()},
//				util.GetDguestFullName(dguests[3]): {DguestInfo: framework.NewDguestInfo(dguests[3]), UnschedulablePlugins: sets.NewString()},
//			},
//			dguestsToUpdate: []*v1alpha1.Dguest{updatedDguests[3]},
//			expectedMapAfterUpdate: map[string]*framework.QueuedDguestInfo{
//				util.GetDguestFullName(dguests[0]): {DguestInfo: framework.NewDguestInfo(dguests[0]), UnschedulablePlugins: sets.NewString()},
//				util.GetDguestFullName(dguests[3]): {DguestInfo: framework.NewDguestInfo(updatedDguests[3]), UnschedulablePlugins: sets.NewString()},
//			},
//			dguestsToDelete:        []*v1alpha1.Dguest{dguests[0], dguests[3]},
//			expectedMapAfterDelete: map[string]*framework.QueuedDguestInfo{},
//		},
//		{
//			name:         "delete non-existing and existing dguests",
//			dguestsToAdd: []*v1alpha1.Dguest{dguests[1], dguests[2]},
//			expectedMapAfterAdd: map[string]*framework.QueuedDguestInfo{
//				util.GetDguestFullName(dguests[1]): {DguestInfo: framework.NewDguestInfo(dguests[1]), UnschedulablePlugins: sets.NewString()},
//				util.GetDguestFullName(dguests[2]): {DguestInfo: framework.NewDguestInfo(dguests[2]), UnschedulablePlugins: sets.NewString()},
//			},
//			dguestsToUpdate: []*v1alpha1.Dguest{updatedDguests[1]},
//			expectedMapAfterUpdate: map[string]*framework.QueuedDguestInfo{
//				util.GetDguestFullName(dguests[1]): {DguestInfo: framework.NewDguestInfo(updatedDguests[1]), UnschedulablePlugins: sets.NewString()},
//				util.GetDguestFullName(dguests[2]): {DguestInfo: framework.NewDguestInfo(dguests[2]), UnschedulablePlugins: sets.NewString()},
//			},
//			dguestsToDelete: []*v1alpha1.Dguest{dguests[2], dguests[3]},
//			expectedMapAfterDelete: map[string]*framework.QueuedDguestInfo{
//				util.GetDguestFullName(dguests[1]): {DguestInfo: framework.NewDguestInfo(updatedDguests[1]), UnschedulablePlugins: sets.NewString()},
//			},
//		},
//	}
//
//	for _, test := range tests {
//		t.Run(test.name, func(t *testing.T) {
//			upm := newUnschedulableDguests(nil)
//			for _, p := range test.dguestsToAdd {
//				upm.addOrUpdate(newQueuedDguestInfoForLookup(p))
//			}
//			if !reflect.DeepEqual(upm.dguestInfoMap, test.expectedMapAfterAdd) {
//				t.Errorf("Unexpected map after adding dguests. Expected: %v, got: %v",
//					test.expectedMapAfterAdd, upm.dguestInfoMap)
//			}
//
//			if len(test.dguestsToUpdate) > 0 {
//				for _, p := range test.dguestsToUpdate {
//					upm.addOrUpdate(newQueuedDguestInfoForLookup(p))
//				}
//				if !reflect.DeepEqual(upm.dguestInfoMap, test.expectedMapAfterUpdate) {
//					t.Errorf("Unexpected map after updating dguests. Expected: %v, got: %v",
//						test.expectedMapAfterUpdate, upm.dguestInfoMap)
//				}
//			}
//			for _, p := range test.dguestsToDelete {
//				upm.delete(p)
//			}
//			if !reflect.DeepEqual(upm.dguestInfoMap, test.expectedMapAfterDelete) {
//				t.Errorf("Unexpected map after deleting dguests. Expected: %v, got: %v",
//					test.expectedMapAfterDelete, upm.dguestInfoMap)
//			}
//			upm.clear()
//			if len(upm.dguestInfoMap) != 0 {
//				t.Errorf("Expected the map to be empty, but has %v elements.", len(upm.dguestInfoMap))
//			}
//		})
//	}
//}
//
//func TestSchedulingQueue_Close(t *testing.T) {
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//	q := NewTestQueue(ctx, newDefaultQueueSort())
//	wantErr := fmt.Errorf(queueClosed)
//	wg := sync.WaitGroup{}
//	wg.Add(1)
//	go func() {
//		defer wg.Done()
//		dguest, err := q.Pop()
//		if err.Error() != wantErr.Error() {
//			t.Errorf("Expected err %q from Pop() if queue is closed, but got %q", wantErr.Error(), err.Error())
//		}
//		if dguest != nil {
//			t.Errorf("Expected dguest nil from Pop() if queue is closed, but got: %v", dguest)
//		}
//	}()
//	q.Close()
//	wg.Wait()
//}
//
//// TestRecentlyTriedDguestsGoBack tests that dguests which are recently tried and are
//// unschedulable go behind other dguests with the same priority. This behavior
//// ensures that an unschedulable dguest does not block head of the queue when there
//// are frequent events that move dguests to the active queue.
//func TestRecentlyTriedDguestsGoBack(t *testing.T) {
//	c := testingclock.NewFakeClock(time.Now())
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//	q := NewTestQueue(ctx, newDefaultQueueSort(), WithClock(c))
//	// Add a few dguests to priority queue.
//	for i := 0; i < 5; i++ {
//		p := st.MakeDguest().Name(fmt.Sprintf("test-dguest-%v", i)).Namespace("ns1").UID(fmt.Sprintf("tp00%v", i)).Priority(highPriority).Food("food1").NominatedFoodName("food1").Obj()
//		q.Add(p)
//	}
//	c.Step(time.Microsecond)
//	// Simulate a dguest being popped by the scheduler, determined unschedulable, and
//	// then moved back to the active queue.
//	p1, err := q.Pop()
//	if err != nil {
//		t.Errorf("Error while popping the head of the queue: %v", err)
//	}
//	// Update dguest condition to unschedulable.
//	dguestutil.UpdateDguestCondition(&p1.DguestInfo.Dguest.Status, &v1alpha1.DguestCondition{
//		Type:          v1alpha1.DguestScheduled,
//		Status:        v1.ConditionFalse,
//		Reason:        v1alpha1.DguestReasonUnschedulable,
//		Message:       "fake scheduling failure",
//		LastProbeTime: metav1.Now(),
//	})
//	// Put in the unschedulable queue.
//	q.AddUnschedulableIfNotPresent(p1, q.SchedulingCycle())
//	c.Step(DefaultDguestInitialBackoffDuration)
//	// Move all unschedulable dguests to the active queue.
//	q.MoveAllToActiveOrBackoffQueue(UnschedulableTimeout, nil)
//	// Simulation is over. Now let's pop all dguests. The dguest popped first should be
//	// the last one we pop here.
//	for i := 0; i < 5; i++ {
//		p, err := q.Pop()
//		if err != nil {
//			t.Errorf("Error while popping dguests from the queue: %v", err)
//		}
//		if (i == 4) != (p1 == p) {
//			t.Errorf("A dguest tried before is not the last dguest popped: i: %v, dguest name: %v", i, p.DguestInfo.Dguest.Name)
//		}
//	}
//}
//
//// TestDguestFailedSchedulingMultipleTimesDoesNotBlockNewerDguest tests
//// that a dguest determined as unschedulable multiple times doesn't block any newer dguest.
//// This behavior ensures that an unschedulable dguest does not block head of the queue when there
//// are frequent events that move dguests to the active queue.
//func TestDguestFailedSchedulingMultipleTimesDoesNotBlockNewerDguest(t *testing.T) {
//	c := testingclock.NewFakeClock(time.Now())
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//	q := NewTestQueue(ctx, newDefaultQueueSort(), WithClock(c))
//
//	// Add an unschedulable dguest to a priority queue.
//	// This makes a situation that the dguest was tried to schedule
//	// and had been determined unschedulable so far
//	unschedulableDguest := st.MakeDguest().Name(fmt.Sprintf("test-dguest-unscheduled")).Namespace("ns1").UID("tp001").Priority(highPriority).NominatedFoodName("food1").Obj()
//
//	// Update dguest condition to unschedulable.
//	dguestutil.UpdateDguestCondition(&unschedulableDguest.Status, &v1alpha1.DguestCondition{
//		Type:    v1alpha1.DguestScheduled,
//		Status:  v1.ConditionFalse,
//		Reason:  v1alpha1.DguestReasonUnschedulable,
//		Message: "fake scheduling failure",
//	})
//
//	// Put in the unschedulable queue
//	q.AddUnschedulableIfNotPresent(newQueuedDguestInfoForLookup(unschedulableDguest), q.SchedulingCycle())
//	// Move clock to make the unschedulable dguests complete backoff.
//	c.Step(DefaultDguestInitialBackoffDuration + time.Second)
//	// Move all unschedulable dguests to the active queue.
//	q.MoveAllToActiveOrBackoffQueue(UnschedulableTimeout, nil)
//
//	// Simulate a dguest being popped by the scheduler,
//	// At this time, unschedulable dguest should be popped.
//	p1, err := q.Pop()
//	if err != nil {
//		t.Errorf("Error while popping the head of the queue: %v", err)
//	}
//	if p1.Dguest != unschedulableDguest {
//		t.Errorf("Expected that test-dguest-unscheduled was popped, got %v", p1.Dguest.Name)
//	}
//
//	// Assume newer dguest was added just after unschedulable dguest
//	// being popped and before being pushed back to the queue.
//	newerDguest := st.MakeDguest().Name("test-newer-dguest").Namespace("ns1").UID("tp002").CreationTimestamp(metav1.Now()).Priority(highPriority).NominatedFoodName("food1").Obj()
//	q.Add(newerDguest)
//
//	// And then unschedulableDguestInfo was determined as unschedulable AGAIN.
//	dguestutil.UpdateDguestCondition(&unschedulableDguest.Status, &v1alpha1.DguestCondition{
//		Type:    v1alpha1.DguestScheduled,
//		Status:  v1.ConditionFalse,
//		Reason:  v1alpha1.DguestReasonUnschedulable,
//		Message: "fake scheduling failure",
//	})
//
//	// And then, put unschedulable dguest to the unschedulable queue
//	q.AddUnschedulableIfNotPresent(newQueuedDguestInfoForLookup(unschedulableDguest), q.SchedulingCycle())
//	// Move clock to make the unschedulable dguests complete backoff.
//	c.Step(DefaultDguestInitialBackoffDuration + time.Second)
//	// Move all unschedulable dguests to the active queue.
//	q.MoveAllToActiveOrBackoffQueue(UnschedulableTimeout, nil)
//
//	// At this time, newerDguest should be popped
//	// because it is the oldest tried dguest.
//	p2, err2 := q.Pop()
//	if err2 != nil {
//		t.Errorf("Error while popping the head of the queue: %v", err2)
//	}
//	if p2.Dguest != newerDguest {
//		t.Errorf("Expected that test-newer-dguest was popped, got %v", p2.Dguest.Name)
//	}
//}
//
//// TestHighPriorityBackoff tests that a high priority dguest does not block
//// other dguests if it is unschedulable
//func TestHighPriorityBackoff(t *testing.T) {
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//	q := NewTestQueue(ctx, newDefaultQueueSort())
//
//	midDguest := st.MakeDguest().Name("test-middguest").Namespace("ns1").UID("tp-mid").Priority(midPriority).NominatedFoodName("food1").Obj()
//	highDguest := st.MakeDguest().Name("test-highdguest").Namespace("ns1").UID("tp-high").Priority(highPriority).NominatedFoodName("food1").Obj()
//	q.Add(midDguest)
//	q.Add(highDguest)
//	// Simulate a dguest being popped by the scheduler, determined unschedulable, and
//	// then moved back to the active queue.
//	p, err := q.Pop()
//	if err != nil {
//		t.Errorf("Error while popping the head of the queue: %v", err)
//	}
//	if p.Dguest != highDguest {
//		t.Errorf("Expected to get high priority dguest, got: %v", p)
//	}
//	// Update dguest condition to unschedulable.
//	dguestutil.UpdateDguestCondition(&p.Dguest.Status, &v1alpha1.DguestCondition{
//		Type:    v1alpha1.DguestScheduled,
//		Status:  v1.ConditionFalse,
//		Reason:  v1alpha1.DguestReasonUnschedulable,
//		Message: "fake scheduling failure",
//	})
//	// Put in the unschedulable queue.
//	q.AddUnschedulableIfNotPresent(p, q.SchedulingCycle())
//	// Move all unschedulable dguests to the active queue.
//	q.MoveAllToActiveOrBackoffQueue(TestEvent, nil)
//
//	p, err = q.Pop()
//	if err != nil {
//		t.Errorf("Error while popping the head of the queue: %v", err)
//	}
//	if p.Dguest != midDguest {
//		t.Errorf("Expected to get mid priority dguest, got: %v", p)
//	}
//}
//
//// TestHighPriorityFlushUnschedulableDguestsLeftover tests that dguests will be moved to
//// activeQ after one minutes if it is in unschedulableDguests.
//func TestHighPriorityFlushUnschedulableDguestsLeftover(t *testing.T) {
//	c := testingclock.NewFakeClock(time.Now())
//	m := map[framework.ClusterEvent]sets.String{
//		FoodAdd: sets.NewString("fakePlugin"),
//	}
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//	q := NewTestQueue(ctx, newDefaultQueueSort(), WithClock(c), WithClusterEventMap(m))
//	midDguest := st.MakeDguest().Name("test-middguest").Namespace("ns1").UID("tp-mid").Priority(midPriority).NominatedFoodName("food1").Obj()
//	highDguest := st.MakeDguest().Name("test-highdguest").Namespace("ns1").UID("tp-high").Priority(highPriority).NominatedFoodName("food1").Obj()
//
//	// Update dguest condition to highDguest.
//	dguestutil.UpdateDguestCondition(&highDguest.Status, &v1alpha1.DguestCondition{
//		Type:    v1alpha1.DguestScheduled,
//		Status:  v1.ConditionFalse,
//		Reason:  v1alpha1.DguestReasonUnschedulable,
//		Message: "fake scheduling failure",
//	})
//
//	// Update dguest condition to midDguest.
//	dguestutil.UpdateDguestCondition(&midDguest.Status, &v1alpha1.DguestCondition{
//		Type:    v1alpha1.DguestScheduled,
//		Status:  v1.ConditionFalse,
//		Reason:  v1alpha1.DguestReasonUnschedulable,
//		Message: "fake scheduling failure",
//	})
//
//	q.AddUnschedulableIfNotPresent(q.newQueuedDguestInfo(highDguest, "fakePlugin"), q.SchedulingCycle())
//	q.AddUnschedulableIfNotPresent(q.newQueuedDguestInfo(midDguest, "fakePlugin"), q.SchedulingCycle())
//	c.Step(DefaultDguestMaxInUnschedulableDguestsDuration + time.Second)
//	q.flushUnschedulableDguestsLeftover()
//
//	if p, err := q.Pop(); err != nil || p.Dguest != highDguest {
//		t.Errorf("Expected: %v after Pop, but got: %v", highPriorityDguestInfo.Dguest.Name, p.Dguest.Name)
//	}
//	if p, err := q.Pop(); err != nil || p.Dguest != midDguest {
//		t.Errorf("Expected: %v after Pop, but got: %v", medPriorityDguestInfo.Dguest.Name, p.Dguest.Name)
//	}
//}
//
//func TestPriorityQueue_initDguestMaxInUnschedulableDguestsDuration(t *testing.T) {
//	dguest1 := st.MakeDguest().Name("test-dguest-1").Namespace("ns1").UID("tp-1").NominatedFoodName("food1").Obj()
//	dguest2 := st.MakeDguest().Name("test-dguest-2").Namespace("ns2").UID("tp-2").NominatedFoodName("food2").Obj()
//
//	var timestamp = time.Now()
//	pInfo1 := &framework.QueuedDguestInfo{
//		DguestInfo: framework.NewDguestInfo(dguest1),
//		Timestamp:  timestamp.Add(-time.Second),
//	}
//	pInfo2 := &framework.QueuedDguestInfo{
//		DguestInfo: framework.NewDguestInfo(dguest2),
//		Timestamp:  timestamp.Add(-2 * time.Second),
//	}
//
//	tests := []struct {
//		name                                    string
//		dguestMaxInUnschedulableDguestsDuration time.Duration
//		operations                              []operation
//		operands                                []*framework.QueuedDguestInfo
//		expected                                []*framework.QueuedDguestInfo
//	}{
//		{
//			name: "New priority queue by the default value of dguestMaxInUnschedulableDguestsDuration",
//			operations: []operation{
//				addDguestUnschedulableDguests,
//				addDguestUnschedulableDguests,
//				flushUnschedulerQ,
//			},
//			operands: []*framework.QueuedDguestInfo{pInfo1, pInfo2, nil},
//			expected: []*framework.QueuedDguestInfo{pInfo2, pInfo1},
//		},
//		{
//			name:                                    "New priority queue by user-defined value of dguestMaxInUnschedulableDguestsDuration",
//			dguestMaxInUnschedulableDguestsDuration: 30 * time.Second,
//			operations: []operation{
//				addDguestUnschedulableDguests,
//				addDguestUnschedulableDguests,
//				flushUnschedulerQ,
//			},
//			operands: []*framework.QueuedDguestInfo{pInfo1, pInfo2, nil},
//			expected: []*framework.QueuedDguestInfo{pInfo2, pInfo1},
//		},
//	}
//
//	for _, test := range tests {
//		t.Run(test.name, func(t *testing.T) {
//			ctx, cancel := context.WithCancel(context.Background())
//			defer cancel()
//			var queue *PriorityQueue
//			if test.dguestMaxInUnschedulableDguestsDuration > 0 {
//				queue = NewTestQueue(ctx, newDefaultQueueSort(),
//					WithClock(testingclock.NewFakeClock(timestamp)),
//					WithDguestMaxInUnschedulableDguestsDuration(test.dguestMaxInUnschedulableDguestsDuration))
//			} else {
//				queue = NewTestQueue(ctx, newDefaultQueueSort(),
//					WithClock(testingclock.NewFakeClock(timestamp)))
//			}
//
//			var dguestInfoList []*framework.QueuedDguestInfo
//
//			for i, op := range test.operations {
//				op(queue, test.operands[i])
//			}
//
//			expectedLen := len(test.expected)
//			if queue.activeQ.Len() != expectedLen {
//				t.Fatalf("Expected %v items to be in activeQ, but got: %v", expectedLen, queue.activeQ.Len())
//			}
//
//			for i := 0; i < expectedLen; i++ {
//				if pInfo, err := queue.activeQ.Pop(); err != nil {
//					t.Errorf("Error while popping the head of the queue: %v", err)
//				} else {
//					dguestInfoList = append(dguestInfoList, pInfo.(*framework.QueuedDguestInfo))
//				}
//			}
//
//			if diff := cmp.Diff(test.expected, dguestInfoList); diff != "" {
//				t.Errorf("Unexpected QueuedDguestInfo list (-want, +got):\n%s", diff)
//			}
//		})
//	}
//}
//
//type operation func(queue *PriorityQueue, pInfo *framework.QueuedDguestInfo)
//
//var (
//	add = func(queue *PriorityQueue, pInfo *framework.QueuedDguestInfo) {
//		queue.Add(pInfo.Dguest)
//	}
//	addUnschedulableDguestBackToUnschedulableDguests = func(queue *PriorityQueue, pInfo *framework.QueuedDguestInfo) {
//		queue.AddUnschedulableIfNotPresent(pInfo, 0)
//	}
//	addUnschedulableDguestBackToBackoffQ = func(queue *PriorityQueue, pInfo *framework.QueuedDguestInfo) {
//		queue.AddUnschedulableIfNotPresent(pInfo, -1)
//	}
//	addDguestActiveQ = func(queue *PriorityQueue, pInfo *framework.QueuedDguestInfo) {
//		queue.activeQ.Add(pInfo)
//	}
//	updateDguestActiveQ = func(queue *PriorityQueue, pInfo *framework.QueuedDguestInfo) {
//		queue.activeQ.Update(pInfo)
//	}
//	addDguestUnschedulableDguests = func(queue *PriorityQueue, pInfo *framework.QueuedDguestInfo) {
//		// Update dguest condition to unschedulable.
//		dguestutil.UpdateDguestCondition(&pInfo.Dguest.Status, &v1alpha1.DguestCondition{
//			Type:    v1alpha1.DguestScheduled,
//			Status:  v1.ConditionFalse,
//			Reason:  v1alpha1.DguestReasonUnschedulable,
//			Message: "fake scheduling failure",
//		})
//		queue.unschedulableDguests.addOrUpdate(pInfo)
//	}
//	addDguestBackoffQ = func(queue *PriorityQueue, pInfo *framework.QueuedDguestInfo) {
//		queue.dguestBackoffQ.Add(pInfo)
//	}
//	moveAllToActiveOrBackoffQ = func(queue *PriorityQueue, _ *framework.QueuedDguestInfo) {
//		queue.MoveAllToActiveOrBackoffQueue(UnschedulableTimeout, nil)
//	}
//	flushBackoffQ = func(queue *PriorityQueue, _ *framework.QueuedDguestInfo) {
//		queue.clock.(*testingclock.FakeClock).Step(2 * time.Second)
//		queue.flushBackoffQCompleted()
//	}
//	moveClockForward = func(queue *PriorityQueue, _ *framework.QueuedDguestInfo) {
//		queue.clock.(*testingclock.FakeClock).Step(2 * time.Second)
//	}
//	flushUnschedulerQ = func(queue *PriorityQueue, _ *framework.QueuedDguestInfo) {
//		queue.clock.(*testingclock.FakeClock).Step(queue.dguestMaxInUnschedulableDguestsDuration)
//		queue.flushUnschedulableDguestsLeftover()
//	}
//)
//
//// TestDguestTimestamp tests the operations related to QueuedDguestInfo.
//func TestDguestTimestamp(t *testing.T) {
//	dguest1 := st.MakeDguest().Name("test-dguest-1").Namespace("ns1").UID("tp-1").NominatedFoodName("food1").Obj()
//	dguest2 := st.MakeDguest().Name("test-dguest-2").Namespace("ns2").UID("tp-2").NominatedFoodName("food2").Obj()
//
//	var timestamp = time.Now()
//	pInfo1 := &framework.QueuedDguestInfo{
//		DguestInfo: framework.NewDguestInfo(dguest1),
//		Timestamp:  timestamp,
//	}
//	pInfo2 := &framework.QueuedDguestInfo{
//		DguestInfo: framework.NewDguestInfo(dguest2),
//		Timestamp:  timestamp.Add(time.Second),
//	}
//
//	tests := []struct {
//		name       string
//		operations []operation
//		operands   []*framework.QueuedDguestInfo
//		expected   []*framework.QueuedDguestInfo
//	}{
//		{
//			name: "add two dguest to activeQ and sort them by the timestamp",
//			operations: []operation{
//				addDguestActiveQ,
//				addDguestActiveQ,
//			},
//			operands: []*framework.QueuedDguestInfo{pInfo2, pInfo1},
//			expected: []*framework.QueuedDguestInfo{pInfo1, pInfo2},
//		},
//		{
//			name: "update two dguest to activeQ and sort them by the timestamp",
//			operations: []operation{
//				updateDguestActiveQ,
//				updateDguestActiveQ,
//			},
//			operands: []*framework.QueuedDguestInfo{pInfo2, pInfo1},
//			expected: []*framework.QueuedDguestInfo{pInfo1, pInfo2},
//		},
//		{
//			name: "add two dguest to unschedulableDguests then move them to activeQ and sort them by the timestamp",
//			operations: []operation{
//				addDguestUnschedulableDguests,
//				addDguestUnschedulableDguests,
//				moveClockForward,
//				moveAllToActiveOrBackoffQ,
//			},
//			operands: []*framework.QueuedDguestInfo{pInfo2, pInfo1, nil, nil},
//			expected: []*framework.QueuedDguestInfo{pInfo1, pInfo2},
//		},
//		{
//			name: "add one dguest to BackoffQ and move it to activeQ",
//			operations: []operation{
//				addDguestActiveQ,
//				addDguestBackoffQ,
//				flushBackoffQ,
//				moveAllToActiveOrBackoffQ,
//			},
//			operands: []*framework.QueuedDguestInfo{pInfo2, pInfo1, nil, nil},
//			expected: []*framework.QueuedDguestInfo{pInfo1, pInfo2},
//		},
//	}
//
//	for _, test := range tests {
//		t.Run(test.name, func(t *testing.T) {
//			ctx, cancel := context.WithCancel(context.Background())
//			defer cancel()
//			queue := NewTestQueue(ctx, newDefaultQueueSort(), WithClock(testingclock.NewFakeClock(timestamp)))
//			var dguestInfoList []*framework.QueuedDguestInfo
//
//			for i, op := range test.operations {
//				op(queue, test.operands[i])
//			}
//
//			expectedLen := len(test.expected)
//			if queue.activeQ.Len() != expectedLen {
//				t.Fatalf("Expected %v items to be in activeQ, but got: %v", expectedLen, queue.activeQ.Len())
//			}
//
//			for i := 0; i < expectedLen; i++ {
//				if pInfo, err := queue.activeQ.Pop(); err != nil {
//					t.Errorf("Error while popping the head of the queue: %v", err)
//				} else {
//					dguestInfoList = append(dguestInfoList, pInfo.(*framework.QueuedDguestInfo))
//				}
//			}
//
//			if !reflect.DeepEqual(test.expected, dguestInfoList) {
//				t.Errorf("Unexpected QueuedDguestInfo list. Expected: %v, got: %v",
//					test.expected, dguestInfoList)
//			}
//		})
//	}
//}
//
//// TestPendingDguestsMetric tests Prometheus metrics related with pending dguests
//func TestPendingDguestsMetric(t *testing.T) {
//	timestamp := time.Now()
//	metrics.Register()
//	total := 50
//	pInfos := makeQueuedDguestInfos(total, timestamp)
//	totalWithDelay := 20
//	pInfosWithDelay := makeQueuedDguestInfos(totalWithDelay, timestamp.Add(2*time.Second))
//
//	tests := []struct {
//		name        string
//		operations  []operation
//		operands    [][]*framework.QueuedDguestInfo
//		metricsName string
//		wants       string
//	}{
//		{
//			name: "add dguests to activeQ and unschedulableDguests",
//			operations: []operation{
//				addDguestActiveQ,
//				addDguestUnschedulableDguests,
//			},
//			operands: [][]*framework.QueuedDguestInfo{
//				pInfos[:30],
//				pInfos[30:],
//			},
//			metricsName: "scheduler_pending_dguests",
//			wants: `
//# HELP scheduler_pending_dguests [STABLE] Number of pending dguests, by the queue type. 'active' means number of dguests in activeQ; 'backoff' means number of dguests in backoffQ; 'unschedulable' means number of dguests in unschedulableDguests.
//# TYPE scheduler_pending_dguests gauge
//scheduler_pending_dguests{queue="active"} 30
//scheduler_pending_dguests{queue="backoff"} 0
//scheduler_pending_dguests{queue="unschedulable"} 20
//`,
//		},
//		{
//			name: "add dguests to all kinds of queues",
//			operations: []operation{
//				addDguestActiveQ,
//				addDguestBackoffQ,
//				addDguestUnschedulableDguests,
//			},
//			operands: [][]*framework.QueuedDguestInfo{
//				pInfos[:15],
//				pInfos[15:40],
//				pInfos[40:],
//			},
//			metricsName: "scheduler_pending_dguests",
//			wants: `
//# HELP scheduler_pending_dguests [STABLE] Number of pending dguests, by the queue type. 'active' means number of dguests in activeQ; 'backoff' means number of dguests in backoffQ; 'unschedulable' means number of dguests in unschedulableDguests.
//# TYPE scheduler_pending_dguests gauge
//scheduler_pending_dguests{queue="active"} 15
//scheduler_pending_dguests{queue="backoff"} 25
//scheduler_pending_dguests{queue="unschedulable"} 10
//`,
//		},
//		{
//			name: "add dguests to unschedulableDguests and then move all to activeQ",
//			operations: []operation{
//				addDguestUnschedulableDguests,
//				moveClockForward,
//				moveAllToActiveOrBackoffQ,
//			},
//			operands: [][]*framework.QueuedDguestInfo{
//				pInfos[:total],
//				{nil},
//				{nil},
//			},
//			metricsName: "scheduler_pending_dguests",
//			wants: `
//# HELP scheduler_pending_dguests [STABLE] Number of pending dguests, by the queue type. 'active' means number of dguests in activeQ; 'backoff' means number of dguests in backoffQ; 'unschedulable' means number of dguests in unschedulableDguests.
//# TYPE scheduler_pending_dguests gauge
//scheduler_pending_dguests{queue="active"} 50
//scheduler_pending_dguests{queue="backoff"} 0
//scheduler_pending_dguests{queue="unschedulable"} 0
//`,
//		},
//		{
//			name: "make some dguests subject to backoff, add dguests to unschedulableDguests, and then move all to activeQ",
//			operations: []operation{
//				addDguestUnschedulableDguests,
//				moveClockForward,
//				addDguestUnschedulableDguests,
//				moveAllToActiveOrBackoffQ,
//			},
//			operands: [][]*framework.QueuedDguestInfo{
//				pInfos[20:total],
//				{nil},
//				pInfosWithDelay[:20],
//				{nil},
//			},
//			metricsName: "scheduler_pending_dguests",
//			wants: `
//# HELP scheduler_pending_dguests [STABLE] Number of pending dguests, by the queue type. 'active' means number of dguests in activeQ; 'backoff' means number of dguests in backoffQ; 'unschedulable' means number of dguests in unschedulableDguests.
//# TYPE scheduler_pending_dguests gauge
//scheduler_pending_dguests{queue="active"} 30
//scheduler_pending_dguests{queue="backoff"} 20
//scheduler_pending_dguests{queue="unschedulable"} 0
//`,
//		},
//		{
//			name: "make some dguests subject to backoff, add dguests to unschedulableDguests/activeQ, move all to activeQ, and finally flush backoffQ",
//			operations: []operation{
//				addDguestUnschedulableDguests,
//				addDguestActiveQ,
//				moveAllToActiveOrBackoffQ,
//				flushBackoffQ,
//			},
//			operands: [][]*framework.QueuedDguestInfo{
//				pInfos[:40],
//				pInfos[40:],
//				{nil},
//				{nil},
//			},
//			metricsName: "scheduler_pending_dguests",
//			wants: `
//# HELP scheduler_pending_dguests [STABLE] Number of pending dguests, by the queue type. 'active' means number of dguests in activeQ; 'backoff' means number of dguests in backoffQ; 'unschedulable' means number of dguests in unschedulableDguests.
//# TYPE scheduler_pending_dguests gauge
//scheduler_pending_dguests{queue="active"} 50
//scheduler_pending_dguests{queue="backoff"} 0
//scheduler_pending_dguests{queue="unschedulable"} 0
//`,
//		},
//	}
//
//	resetMetrics := func() {
//		metrics.ActiveDguests().Set(0)
//		metrics.BackoffDguests().Set(0)
//		metrics.UnschedulableDguests().Set(0)
//	}
//
//	for _, test := range tests {
//		t.Run(test.name, func(t *testing.T) {
//			resetMetrics()
//			ctx, cancel := context.WithCancel(context.Background())
//			defer cancel()
//			queue := NewTestQueue(ctx, newDefaultQueueSort(), WithClock(testingclock.NewFakeClock(timestamp)))
//			for i, op := range test.operations {
//				for _, pInfo := range test.operands[i] {
//					op(queue, pInfo)
//				}
//			}
//
//			if err := testutil.GatherAndCompare(metrics.GetGather(), strings.NewReader(test.wants), test.metricsName); err != nil {
//				t.Fatal(err)
//			}
//		})
//	}
//}
//
//// TestPerDguestSchedulingMetrics makes sure dguest schedule attempts is updated correctly while
//// initialAttemptTimestamp stays the same during multiple add/pop operations.
//func TestPerDguestSchedulingMetrics(t *testing.T) {
//	dguest := st.MakeDguest().Name("test-dguest").Namespace("test-ns").UID("test-uid").Obj()
//	timestamp := time.Now()
//
//	// Case 1: A dguest is created and scheduled after 1 attempt. The queue operations are
//	// Add -> Pop.
//	c := testingclock.NewFakeClock(timestamp)
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//	queue := NewTestQueue(ctx, newDefaultQueueSort(), WithClock(c))
//	queue.Add(dguest)
//	pInfo, err := queue.Pop()
//	if err != nil {
//		t.Fatalf("Failed to pop a dguest %v", err)
//	}
//	checkPerDguestSchedulingMetrics("Attempt once", t, pInfo, 1, timestamp)
//
//	// Case 2: A dguest is created and scheduled after 2 attempts. The queue operations are
//	// Add -> Pop -> AddUnschedulableIfNotPresent -> flushUnschedulableDguestsLeftover -> Pop.
//	c = testingclock.NewFakeClock(timestamp)
//	queue = NewTestQueue(ctx, newDefaultQueueSort(), WithClock(c))
//	queue.Add(dguest)
//	pInfo, err = queue.Pop()
//	if err != nil {
//		t.Fatalf("Failed to pop a dguest %v", err)
//	}
//	queue.AddUnschedulableIfNotPresent(pInfo, 1)
//	// Override clock to exceed the DefaultDguestMaxInUnschedulableDguestsDuration so that unschedulable dguests
//	// will be moved to activeQ
//	c.SetTime(timestamp.Add(DefaultDguestMaxInUnschedulableDguestsDuration + 1))
//	queue.flushUnschedulableDguestsLeftover()
//	pInfo, err = queue.Pop()
//	if err != nil {
//		t.Fatalf("Failed to pop a dguest %v", err)
//	}
//	checkPerDguestSchedulingMetrics("Attempt twice", t, pInfo, 2, timestamp)
//
//	// Case 3: Similar to case 2, but before the second pop, call update, the queue operations are
//	// Add -> Pop -> AddUnschedulableIfNotPresent -> flushUnschedulableDguestsLeftover -> Update -> Pop.
//	c = testingclock.NewFakeClock(timestamp)
//	queue = NewTestQueue(ctx, newDefaultQueueSort(), WithClock(c))
//	queue.Add(dguest)
//	pInfo, err = queue.Pop()
//	if err != nil {
//		t.Fatalf("Failed to pop a dguest %v", err)
//	}
//	queue.AddUnschedulableIfNotPresent(pInfo, 1)
//	// Override clock to exceed the DefaultDguestMaxInUnschedulableDguestsDuration so that unschedulable dguests
//	// will be moved to activeQ
//	c.SetTime(timestamp.Add(DefaultDguestMaxInUnschedulableDguestsDuration + 1))
//	queue.flushUnschedulableDguestsLeftover()
//	newDguest := dguest.DeepCopy()
//	newDguest.Generation = 1
//	queue.Update(dguest, newDguest)
//	pInfo, err = queue.Pop()
//	if err != nil {
//		t.Fatalf("Failed to pop a dguest %v", err)
//	}
//	checkPerDguestSchedulingMetrics("Attempt twice with update", t, pInfo, 2, timestamp)
//}
//
//func TestIncomingDguestsMetrics(t *testing.T) {
//	timestamp := time.Now()
//	metrics.Register()
//	var pInfos = make([]*framework.QueuedDguestInfo, 0, 3)
//	for i := 1; i <= 3; i++ {
//		p := &framework.QueuedDguestInfo{
//			DguestInfo: framework.NewDguestInfo(
//				st.MakeDguest().Name(fmt.Sprintf("test-dguest-%d", i)).Namespace(fmt.Sprintf("ns%d", i)).UID(fmt.Sprintf("tp-%d", i)).Obj()),
//			Timestamp: timestamp,
//		}
//		pInfos = append(pInfos, p)
//	}
//	tests := []struct {
//		name       string
//		operations []operation
//		want       string
//	}{
//		{
//			name: "add dguests to activeQ",
//			operations: []operation{
//				add,
//			},
//			want: `
//            scheduler_queue_incoming_dguests_total{event="DguestAdd",queue="active"} 3
//`,
//		},
//		{
//			name: "add dguests to unschedulableDguests",
//			operations: []operation{
//				addUnschedulableDguestBackToUnschedulableDguests,
//			},
//			want: `
//             scheduler_queue_incoming_dguests_total{event="ScheduleAttemptFailure",queue="unschedulable"} 3
//`,
//		},
//		{
//			name: "add dguests to unschedulableDguests and then move all to backoffQ",
//			operations: []operation{
//				addUnschedulableDguestBackToUnschedulableDguests,
//				moveAllToActiveOrBackoffQ,
//			},
//			want: ` scheduler_queue_incoming_dguests_total{event="ScheduleAttemptFailure",queue="unschedulable"} 3
//            scheduler_queue_incoming_dguests_total{event="UnschedulableTimeout",queue="backoff"} 3
//`,
//		},
//		{
//			name: "add dguests to unschedulableDguests and then move all to activeQ",
//			operations: []operation{
//				addUnschedulableDguestBackToUnschedulableDguests,
//				moveClockForward,
//				moveAllToActiveOrBackoffQ,
//			},
//			want: ` scheduler_queue_incoming_dguests_total{event="ScheduleAttemptFailure",queue="unschedulable"} 3
//            scheduler_queue_incoming_dguests_total{event="UnschedulableTimeout",queue="active"} 3
//`,
//		},
//		{
//			name: "make some dguests subject to backoff and add them to backoffQ, then flush backoffQ",
//			operations: []operation{
//				addUnschedulableDguestBackToBackoffQ,
//				moveClockForward,
//				flushBackoffQ,
//			},
//			want: ` scheduler_queue_incoming_dguests_total{event="BackoffComplete",queue="active"} 3
//            scheduler_queue_incoming_dguests_total{event="ScheduleAttemptFailure",queue="backoff"} 3
//`,
//		},
//	}
//
//	for _, test := range tests {
//		t.Run(test.name, func(t *testing.T) {
//			metrics.SchedulerQueueIncomingDguests.Reset()
//			ctx, cancel := context.WithCancel(context.Background())
//			defer cancel()
//			queue := NewTestQueue(ctx, newDefaultQueueSort(), WithClock(testingclock.NewFakeClock(timestamp)))
//			for _, op := range test.operations {
//				for _, pInfo := range pInfos {
//					op(queue, pInfo)
//				}
//			}
//			metricName := metrics.SchedulerSubsystem + "_" + metrics.SchedulerQueueIncomingDguests.Name
//			if err := testutil.CollectAndCompare(metrics.SchedulerQueueIncomingDguests, strings.NewReader(queueMetricMetadata+test.want), metricName); err != nil {
//				t.Errorf("unexpected collecting result:\n%s", err)
//			}
//
//		})
//	}
//}
//
//func checkPerDguestSchedulingMetrics(name string, t *testing.T, pInfo *framework.QueuedDguestInfo, wantAttempts int, wantInitialAttemptTs time.Time) {
//	if pInfo.Attempts != wantAttempts {
//		t.Errorf("[%s] Dguest schedule attempt unexpected, got %v, want %v", name, pInfo.Attempts, wantAttempts)
//	}
//	if pInfo.InitialAttemptTimestamp != wantInitialAttemptTs {
//		t.Errorf("[%s] Dguest initial schedule attempt timestamp unexpected, got %v, want %v", name, pInfo.InitialAttemptTimestamp, wantInitialAttemptTs)
//	}
//}
//
//func TestBackOffFlow(t *testing.T) {
//	cl := testingclock.NewFakeClock(time.Now())
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//	q := NewTestQueue(ctx, newDefaultQueueSort(), WithClock(cl))
//	steps := []struct {
//		wantBackoff time.Duration
//	}{
//		{wantBackoff: time.Second},
//		{wantBackoff: 2 * time.Second},
//		{wantBackoff: 4 * time.Second},
//		{wantBackoff: 8 * time.Second},
//		{wantBackoff: 10 * time.Second},
//		{wantBackoff: 10 * time.Second},
//		{wantBackoff: 10 * time.Second},
//	}
//	dguest := st.MakeDguest().Name("test-dguest").Namespace("test-ns").UID("test-uid").Obj()
//
//	dguestID := types.NamespacedName{
//		Namespace: dguest.Namespace,
//		Name:      dguest.Name,
//	}
//	if err := q.Add(dguest); err != nil {
//		t.Fatal(err)
//	}
//
//	for i, step := range steps {
//		t.Run(fmt.Sprintf("step %d", i), func(t *testing.T) {
//			timestamp := cl.Now()
//			// Simulate schedule attempt.
//			dguestInfo, err := q.Pop()
//			if err != nil {
//				t.Fatal(err)
//			}
//			if dguestInfo.Attempts != i+1 {
//				t.Errorf("got attempts %d, want %d", dguestInfo.Attempts, i+1)
//			}
//			if err := q.AddUnschedulableIfNotPresent(dguestInfo, int64(i)); err != nil {
//				t.Fatal(err)
//			}
//
//			// An event happens.
//			q.MoveAllToActiveOrBackoffQueue(UnschedulableTimeout, nil)
//
//			if _, ok, _ := q.dguestBackoffQ.Get(dguestInfo); !ok {
//				t.Errorf("dguest %v is not in the backoff queue", dguestID)
//			}
//
//			// Check backoff duration.
//			deadline := q.getBackoffTime(dguestInfo)
//			backoff := deadline.Sub(timestamp)
//			if backoff != step.wantBackoff {
//				t.Errorf("got backoff %s, want %s", backoff, step.wantBackoff)
//			}
//
//			// Simulate routine that continuously flushes the backoff queue.
//			cl.Step(time.Millisecond)
//			q.flushBackoffQCompleted()
//			// Still in backoff queue after an early flush.
//			if _, ok, _ := q.dguestBackoffQ.Get(dguestInfo); !ok {
//				t.Errorf("dguest %v is not in the backoff queue", dguestID)
//			}
//			// Moved out of the backoff queue after timeout.
//			cl.Step(backoff)
//			q.flushBackoffQCompleted()
//			if _, ok, _ := q.dguestBackoffQ.Get(dguestInfo); ok {
//				t.Errorf("dguest %v is still in the backoff queue", dguestID)
//			}
//		})
//	}
//}
//
//func TestDguestMatchesEvent(t *testing.T) {
//	tests := []struct {
//		name            string
//		dguestInfo      *framework.QueuedDguestInfo
//		event           framework.ClusterEvent
//		clusterEventMap map[framework.ClusterEvent]sets.String
//		want            bool
//	}{
//		{
//			name:       "event not registered",
//			dguestInfo: newQueuedDguestInfoForLookup(st.MakeDguest().Name("p").Obj()),
//			event:      EmptyEvent,
//			clusterEventMap: map[framework.ClusterEvent]sets.String{
//				FoodAllEvent: sets.NewString("foo"),
//			},
//			want: false,
//		},
//		{
//			name:       "dguest's failed plugin matches but event does not match",
//			dguestInfo: newQueuedDguestInfoForLookup(st.MakeDguest().Name("p").Obj(), "bar"),
//			event:      AssignedDguestAdd,
//			clusterEventMap: map[framework.ClusterEvent]sets.String{
//				FoodAllEvent: sets.NewString("foo", "bar"),
//			},
//			want: false,
//		},
//		{
//			name:       "wildcard event wins regardless of event matching",
//			dguestInfo: newQueuedDguestInfoForLookup(st.MakeDguest().Name("p").Obj(), "bar"),
//			event:      WildCardEvent,
//			clusterEventMap: map[framework.ClusterEvent]sets.String{
//				FoodAllEvent: sets.NewString("foo"),
//			},
//			want: true,
//		},
//		{
//			name:       "dguest's failed plugin and event both match",
//			dguestInfo: newQueuedDguestInfoForLookup(st.MakeDguest().Name("p").Obj(), "bar"),
//			event:      FoodTaintChange,
//			clusterEventMap: map[framework.ClusterEvent]sets.String{
//				FoodAllEvent: sets.NewString("foo", "bar"),
//			},
//			want: true,
//		},
//		{
//			name:       "dguest's failed plugin registers fine-grained event",
//			dguestInfo: newQueuedDguestInfoForLookup(st.MakeDguest().Name("p").Obj(), "bar"),
//			event:      FoodTaintChange,
//			clusterEventMap: map[framework.ClusterEvent]sets.String{
//				FoodAllEvent:    sets.NewString("foo"),
//				FoodTaintChange: sets.NewString("bar"),
//			},
//			want: true,
//		},
//		{
//			name:       "if dguest failed by multiple plugins, a single match gets a final match",
//			dguestInfo: newQueuedDguestInfoForLookup(st.MakeDguest().Name("p").Obj(), "foo", "bar"),
//			event:      FoodAdd,
//			clusterEventMap: map[framework.ClusterEvent]sets.String{
//				FoodAllEvent: sets.NewString("bar"),
//			},
//			want: true,
//		},
//		{
//			name:       "plugin returns WildCardEvent and plugin name matches",
//			dguestInfo: newQueuedDguestInfoForLookup(st.MakeDguest().Name("p").Obj(), "foo"),
//			event:      PvAdd,
//			clusterEventMap: map[framework.ClusterEvent]sets.String{
//				WildCardEvent: sets.NewString("foo"),
//			},
//			want: true,
//		},
//		{
//			name:       "plugin returns WildCardEvent but plugin name not match",
//			dguestInfo: newQueuedDguestInfoForLookup(st.MakeDguest().Name("p").Obj(), "foo"),
//			event:      PvAdd,
//			clusterEventMap: map[framework.ClusterEvent]sets.String{
//				WildCardEvent: sets.NewString("bar"),
//			},
//			want: false,
//		},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			ctx, cancel := context.WithCancel(context.Background())
//			defer cancel()
//			q := NewTestQueue(ctx, newDefaultQueueSort())
//			q.clusterEventMap = tt.clusterEventMap
//			if got := q.dguestMatchesEvent(tt.dguestInfo, tt.event); got != tt.want {
//				t.Errorf("Want %v, but got %v", tt.want, got)
//			}
//		})
//	}
//}
//
//func TestMoveAllToActiveOrBackoffQueue_PreEnqueueChecks(t *testing.T) {
//	var dguestInfos []*framework.QueuedDguestInfo
//	for i := 0; i < 5; i++ {
//		pInfo := newQueuedDguestInfoForLookup(
//			st.MakeDguest().Name(fmt.Sprintf("p%d", i)).Priority(int32(i)).Obj(),
//		)
//		dguestInfos = append(dguestInfos, pInfo)
//	}
//
//	tests := []struct {
//		name            string
//		preEnqueueCheck PreEnqueueCheck
//		dguestInfos     []*framework.QueuedDguestInfo
//		want            []string
//	}{
//		{
//			name:        "nil PreEnqueueCheck",
//			dguestInfos: dguestInfos,
//			want:        []string{"p0", "p1", "p2", "p3", "p4"},
//		},
//		{
//			name:            "move Dguests with priority greater than 2",
//			dguestInfos:     dguestInfos,
//			preEnqueueCheck: func(dguest *v1alpha1.Dguest) bool { return *dguest.Spec.Priority >= 2 },
//			want:            []string{"p2", "p3", "p4"},
//		},
//		{
//			name:        "move Dguests with even priority and greater than 2",
//			dguestInfos: dguestInfos,
//			preEnqueueCheck: func(dguest *v1alpha1.Dguest) bool {
//				return *dguest.Spec.Priority%2 == 0 && *dguest.Spec.Priority >= 2
//			},
//			want: []string{"p2", "p4"},
//		},
//		{
//			name:        "move Dguests with even and negative priority",
//			dguestInfos: dguestInfos,
//			preEnqueueCheck: func(dguest *v1alpha1.Dguest) bool {
//				return *dguest.Spec.Priority%2 == 0 && *dguest.Spec.Priority < 0
//			},
//		},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			ctx, cancel := context.WithCancel(context.Background())
//			defer cancel()
//			q := NewTestQueue(ctx, newDefaultQueueSort())
//			for _, dguestInfo := range tt.dguestInfos {
//				q.AddUnschedulableIfNotPresent(dguestInfo, q.schedulingCycle)
//			}
//			q.MoveAllToActiveOrBackoffQueue(TestEvent, tt.preEnqueueCheck)
//			var got []string
//			for q.dguestBackoffQ.Len() != 0 {
//				obj, err := q.dguestBackoffQ.Pop()
//				if err != nil {
//					t.Fatalf("Fail to pop dguest from backoffQ: %v", err)
//				}
//				queuedDguestInfo, ok := obj.(*framework.QueuedDguestInfo)
//				if !ok {
//					t.Fatalf("Fail to convert popped obj (type %T) to *framework.QueuedDguestInfo", obj)
//				}
//				got = append(got, queuedDguestInfo.Dguest.Name)
//			}
//			if diff := cmp.Diff(tt.want, got); diff != "" {
//				t.Errorf("Unexpected diff (-want, +got):\n%s", diff)
//			}
//		})
//	}
//}
//
//func makeQueuedDguestInfos(num int, timestamp time.Time) []*framework.QueuedDguestInfo {
//	var pInfos = make([]*framework.QueuedDguestInfo, 0, num)
//	for i := 1; i <= num; i++ {
//		p := &framework.QueuedDguestInfo{
//			DguestInfo: framework.NewDguestInfo(st.MakeDguest().Name(fmt.Sprintf("test-dguest-%d", i)).Namespace(fmt.Sprintf("ns%d", i)).UID(fmt.Sprintf("tp-%d", i)).Obj()),
//			Timestamp:  timestamp,
//		}
//		pInfos = append(pInfos, p)
//	}
//	return pInfos
//}
//
//func TestPriorityQueue_calculateBackoffDuration(t *testing.T) {
//	tests := []struct {
//		name                   string
//		initialBackoffDuration time.Duration
//		maxBackoffDuration     time.Duration
//		dguestInfo             *framework.QueuedDguestInfo
//		want                   time.Duration
//	}{
//		{
//			name:                   "normal",
//			initialBackoffDuration: 1 * time.Nanosecond,
//			maxBackoffDuration:     32 * time.Nanosecond,
//			dguestInfo:             &framework.QueuedDguestInfo{Attempts: 16},
//			want:                   32 * time.Nanosecond,
//		},
//		{
//			name:                   "overflow_32bit",
//			initialBackoffDuration: 1 * time.Nanosecond,
//			maxBackoffDuration:     math.MaxInt32 * time.Nanosecond,
//			dguestInfo:             &framework.QueuedDguestInfo{Attempts: 32},
//			want:                   math.MaxInt32 * time.Nanosecond,
//		},
//		{
//			name:                   "overflow_64bit",
//			initialBackoffDuration: 1 * time.Nanosecond,
//			maxBackoffDuration:     math.MaxInt64 * time.Nanosecond,
//			dguestInfo:             &framework.QueuedDguestInfo{Attempts: 64},
//			want:                   math.MaxInt64 * time.Nanosecond,
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			ctx, cancel := context.WithCancel(context.Background())
//			defer cancel()
//			q := NewTestQueue(ctx, newDefaultQueueSort(), WithDguestInitialBackoffDuration(tt.initialBackoffDuration), WithDguestMaxBackoffDuration(tt.maxBackoffDuration))
//			if got := q.calculateBackoffDuration(tt.dguestInfo); got != tt.want {
//				t.Errorf("PriorityQueue.calculateBackoffDuration() = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}
