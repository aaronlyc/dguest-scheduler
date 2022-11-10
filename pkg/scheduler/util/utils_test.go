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
package util

//
//import (
//	"context"
//	"dguest-scheduler/pkg/apis/scheduler/v1alpha1"
//	"errors"
//	"fmt"
//	"syscall"
//	"testing"
//	"time"
//
//	"github.com/google/go-cmp/cmp"
//	v1 "k8s.io/api/core/v1"
//	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
//	"k8s.io/apimachinery/pkg/runtime"
//	"k8s.io/apimachinery/pkg/util/net"
//	clientsetfake "k8s.io/client-go/kubernetes/fake"
//	clienttesting "k8s.io/client-go/testing"
//	extenderv1 "k8s.io/kube-scheduler/extender/v1"
//)
//
//func TestGetDguestFullName(t *testing.T) {
//	dguest := &v1alpha1.Dguest{
//		ObjectMeta: metav1.ObjectMeta{
//			Namespace: "test",
//			Name:      "dguest",
//		},
//	}
//	got := GetDguestFullName(dguest)
//	expected := fmt.Sprintf("%s_%s", dguest.Name, dguest.Namespace)
//	if got != expected {
//		t.Errorf("Got wrong full name, got: %s, expected: %s", got, expected)
//	}
//}
//
//func newPriorityDguestWithStartTime(name string, priority int32, startTime time.Time) *v1alpha1.Dguest {
//	return &v1alpha1.Dguest{
//		ObjectMeta: metav1.ObjectMeta{
//			Name: name,
//		},
//		Spec: v1alpha1.DguestSpec{
//			//Priority: &priority,
//		},
//		Status: v1alpha1.DguestStatus{
//			//StartTime: &metav1.Time{Time: startTime},
//		},
//	}
//}
//
//func TestGetEarliestDguestStartTime(t *testing.T) {
//	var priority int32 = 1
//	currentTime := time.Now()
//	tests := []struct {
//		name              string
//		dguests           []*v1alpha1.Dguest
//		expectedStartTime *metav1.Time
//	}{
//		{
//			name:              "Dguests length is 0",
//			dguests:           []*v1alpha1.Dguest{},
//			expectedStartTime: nil,
//		},
//		{
//			name: "generate new startTime",
//			dguests: []*v1alpha1.Dguest{
//				newPriorityDguestWithStartTime("dguest1", 1, currentTime.Add(-time.Second)),
//				{
//					ObjectMeta: metav1.ObjectMeta{
//						Name: "dguest2",
//					},
//					Spec: v1alpha1.DguestSpec{
//						Priority: &priority,
//					},
//				},
//			},
//			expectedStartTime: &metav1.Time{Time: currentTime.Add(-time.Second)},
//		},
//		{
//			name: "Dguest with earliest start time last in the list",
//			dguests: []*v1alpha1.Dguest{
//				newPriorityDguestWithStartTime("dguest1", 1, currentTime.Add(time.Second)),
//				newPriorityDguestWithStartTime("dguest2", 2, currentTime.Add(time.Second)),
//				newPriorityDguestWithStartTime("dguest3", 2, currentTime),
//			},
//			expectedStartTime: &metav1.Time{Time: currentTime},
//		},
//		{
//			name: "Dguest with earliest start time first in the list",
//			dguests: []*v1alpha1.Dguest{
//				newPriorityDguestWithStartTime("dguest1", 2, currentTime),
//				newPriorityDguestWithStartTime("dguest2", 2, currentTime.Add(time.Second)),
//				newPriorityDguestWithStartTime("dguest3", 2, currentTime.Add(2*time.Second)),
//			},
//			expectedStartTime: &metav1.Time{Time: currentTime},
//		},
//	}
//	for _, test := range tests {
//		t.Run(test.name, func(t *testing.T) {
//			startTime := GetEarliestDguestStartTime(&extenderv1.Victims{Dguests: test.dguests})
//			if !startTime.Equal(test.expectedStartTime) {
//				t.Errorf("startTime is not the expected result,got %v, expected %v", startTime, test.expectedStartTime)
//			}
//		})
//	}
//}
//
//func TestMoreImportantDguest(t *testing.T) {
//	currentTime := time.Now()
//	dguest1 := newPriorityDguestWithStartTime("dguest1", 1, currentTime)
//	dguest2 := newPriorityDguestWithStartTime("dguest2", 2, currentTime.Add(time.Second))
//	dguest3 := newPriorityDguestWithStartTime("dguest3", 2, currentTime)
//
//	tests := map[string]struct {
//		p1       *v1alpha1.Dguest
//		p2       *v1alpha1.Dguest
//		expected bool
//	}{
//		"Dguest with higher priority": {
//			p1:       dguest1,
//			p2:       dguest2,
//			expected: false,
//		},
//		"Dguest with older created time": {
//			p1:       dguest2,
//			p2:       dguest3,
//			expected: false,
//		},
//		"Dguests with same start time": {
//			p1:       dguest3,
//			p2:       dguest1,
//			expected: true,
//		},
//	}
//
//	for k, v := range tests {
//		t.Run(k, func(t *testing.T) {
//			got := MoreImportantDguest(v.p1, v.p2)
//			if got != v.expected {
//				t.Errorf("expected %t but got %t", v.expected, got)
//			}
//		})
//	}
//}
//
//func TestRemoveNominatedFoodName(t *testing.T) {
//	tests := []struct {
//		name                     string
//		currentNominatedFoodName string
//		newNominatedFoodName     string
//		expectedPatchRequests    int
//		expectedPatchData        string
//	}{
//		{
//			name:                     "Should make patch request to clear food name",
//			currentNominatedFoodName: "food1",
//			expectedPatchRequests:    1,
//			expectedPatchData:        `{"status":{"nominatedFoodName":null}}`,
//		},
//		{
//			name:                     "Should not make patch request if nominated food is already cleared",
//			currentNominatedFoodName: "",
//			expectedPatchRequests:    0,
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
//			dguest := &v1alpha1.Dguest{
//				ObjectMeta: metav1.ObjectMeta{Name: "foo"},
//				Status:     v1alpha1.DguestStatus{NominatedFoodName: test.currentNominatedFoodName},
//			}
//
//			ctx, cancel := context.WithCancel(context.Background())
//			defer cancel()
//			if err := ClearNominatedFoodName(ctx, cs, dguest); err != nil {
//				t.Fatalf("Error calling removeNominatedFoodName: %v", err)
//			}
//
//			if actualPatchRequests != test.expectedPatchRequests {
//				t.Fatalf("Actual patch requests (%d) dos not equal expected patch requests (%d)", actualPatchRequests, test.expectedPatchRequests)
//			}
//
//			if test.expectedPatchRequests > 0 && actualPatchData != test.expectedPatchData {
//				t.Fatalf("Patch data mismatch: Actual was %v, but expected %v", actualPatchData, test.expectedPatchData)
//			}
//		})
//	}
//}
//
//func TestPatchDguestStatus(t *testing.T) {
//	tests := []struct {
//		name   string
//		dguest v1alpha1.Dguest
//		client *clientsetfake.Clientset
//		// validateErr checks if error returned from PatchDguestStatus is expected one or not.
//		// (true means error is expected one.)
//		validateErr    func(goterr error) bool
//		statusToUpdate v1alpha1.DguestStatus
//	}{
//		{
//			name:   "Should update dguest conditions successfully",
//			client: clientsetfake.NewSimpleClientset(),
//			dguest: v1alpha1.Dguest{
//				ObjectMeta: metav1.ObjectMeta{
//					Namespace: "ns",
//					Name:      "dguest1",
//				},
//				Spec: v1alpha1.DguestSpec{
//					ImagePullSecrets: []v1.LocalObjectReference{{Name: "foo"}},
//				},
//			},
//			statusToUpdate: v1alpha1.DguestStatus{
//				Conditions: []v1alpha1.DguestCondition{
//					{
//						Type:   v1alpha1.DguestScheduled,
//						Status: v1.ConditionFalse,
//					},
//				},
//			},
//		},
//		{
//			// ref: #101697, #94626 - ImagePullSecrets are allowed to have empty secret names
//			// which would fail the 2-way merge patch generation on Dguest patches
//			// due to the mergeKey being the name field
//			name:   "Should update dguest conditions successfully on a dguest Spec with secrets with empty name",
//			client: clientsetfake.NewSimpleClientset(),
//			dguest: v1alpha1.Dguest{
//				ObjectMeta: metav1.ObjectMeta{
//					Namespace: "ns",
//					Name:      "dguest1",
//				},
//				Spec: v1alpha1.DguestSpec{
//					// this will serialize to imagePullSecrets:[{}]
//					ImagePullSecrets: make([]v1.LocalObjectReference, 1),
//				},
//			},
//			statusToUpdate: v1alpha1.DguestStatus{
//				Conditions: []v1alpha1.DguestCondition{
//					{
//						Type:   v1alpha1.DguestScheduled,
//						Status: v1.ConditionFalse,
//					},
//				},
//			},
//		},
//		{
//			name: "retry patch request when an 'connection refused' error is returned",
//			client: func() *clientsetfake.Clientset {
//				client := clientsetfake.NewSimpleClientset()
//
//				reqcount := 0
//				client.PrependReactor("patch", "dguests", func(action clienttesting.Action) (bool, runtime.Object, error) {
//					defer func() { reqcount++ }()
//					if reqcount == 0 {
//						// return an connection refused error for the first patch request.
//						return true, &v1alpha1.Dguest{}, fmt.Errorf("connection refused: %w", syscall.ECONNREFUSED)
//					}
//					if reqcount == 1 {
//						// not return error for the second patch request.
//						return false, &v1alpha1.Dguest{}, nil
//					}
//
//					// return error if requests comes in more than three times.
//					return true, nil, errors.New("requests comes in more than three times.")
//				})
//
//				return client
//			}(),
//			dguest: v1alpha1.Dguest{
//				ObjectMeta: metav1.ObjectMeta{
//					Namespace: "ns",
//					Name:      "dguest1",
//				},
//				Spec: v1alpha1.DguestSpec{
//					ImagePullSecrets: []v1.LocalObjectReference{{Name: "foo"}},
//				},
//			},
//			statusToUpdate: v1alpha1.DguestStatus{
//				Conditions: []v1alpha1.DguestCondition{
//					{
//						Type:   v1alpha1.DguestScheduled,
//						Status: v1.ConditionFalse,
//					},
//				},
//			},
//		},
//		{
//			name: "only 4 retries at most",
//			client: func() *clientsetfake.Clientset {
//				client := clientsetfake.NewSimpleClientset()
//
//				reqcount := 0
//				client.PrependReactor("patch", "dguests", func(action clienttesting.Action) (bool, runtime.Object, error) {
//					defer func() { reqcount++ }()
//					if reqcount >= 4 {
//						// return error if requests comes in more than four times.
//						return true, nil, errors.New("requests comes in more than four times.")
//					}
//
//					// return an connection refused error for the first patch request.
//					return true, &v1alpha1.Dguest{}, fmt.Errorf("connection refused: %w", syscall.ECONNREFUSED)
//				})
//
//				return client
//			}(),
//			dguest: v1alpha1.Dguest{
//				ObjectMeta: metav1.ObjectMeta{
//					Namespace: "ns",
//					Name:      "dguest1",
//				},
//				Spec: v1alpha1.DguestSpec{
//					ImagePullSecrets: []v1.LocalObjectReference{{Name: "foo"}},
//				},
//			},
//			validateErr: net.IsConnectionRefused,
//			statusToUpdate: v1alpha1.DguestStatus{
//				Conditions: []v1alpha1.DguestCondition{
//					{
//						Type:   v1alpha1.DguestScheduled,
//						Status: v1.ConditionFalse,
//					},
//				},
//			},
//		},
//	}
//
//	for _, tc := range tests {
//		t.Run(tc.name, func(t *testing.T) {
//			client := tc.client
//			_, err := client.CoreV1().Dguests(tc.dguest.Namespace).Create(context.TODO(), &tc.dguest, metav1.CreateOptions{})
//			if err != nil {
//				t.Fatal(err)
//			}
//
//			ctx, cancel := context.WithCancel(context.Background())
//			defer cancel()
//			err = PatchDguestStatus(ctx, client, &tc.dguest, &tc.statusToUpdate)
//			if err != nil && tc.validateErr == nil {
//				// shouldn't be error
//				t.Fatal(err)
//			}
//			if tc.validateErr != nil {
//				if !tc.validateErr(err) {
//					t.Fatalf("Returned unexpected error: %v", err)
//				}
//				return
//			}
//
//			retrievedDguest, err := client.CoreV1().Dguests(tc.dguest.Namespace).Get(ctx, tc.dguest.Name, metav1.GetOptions{})
//			if err != nil {
//				t.Fatal(err)
//			}
//
//			if diff := cmp.Diff(tc.statusToUpdate, retrievedDguest.Status); diff != "" {
//				t.Errorf("unexpected dguest status (-want,+got):\n%s", diff)
//			}
//		})
//	}
//}
