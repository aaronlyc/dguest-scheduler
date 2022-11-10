/*
Copyright 2019 The Kubernetes Authors.

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

package foodname

import (
	"context"
	"reflect"
	"testing"

	"dguest-scheduler/pkg/scheduler/framework"
	st "dguest-scheduler/pkg/scheduler/testing"
	v1 "k8s.io/api/core/v1"
)

func TestFoodName(t *testing.T) {
	tests := []struct {
		dguest     *v1alpha1.Dguest
		food       *v1alpha1.Food
		name       string
		wantStatus *framework.Status
	}{
		{
			dguest: &v1alpha1.Dguest{},
			food:   &v1alpha1.Food{},
			name:   "no host specified",
		},
		{
			dguest: st.MakeDguest().Food("foo").Obj(),
			food:   st.MakeFood().Name("foo").Obj(),
			name:   "host matches",
		},
		{
			dguest:     st.MakeDguest().Food("bar").Obj(),
			food:       st.MakeFood().Name("foo").Obj(),
			name:       "host doesn't match",
			wantStatus: framework.NewStatus(framework.UnschedulableAndUnresolvable, ErrReason),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			foodInfo := framework.NewFoodInfo()
			foodInfo.SetFood(test.food)

			p, _ := New(nil, nil)
			gotStatus := p.(framework.FilterPlugin).Filter(context.Background(), nil, test.dguest, foodInfo)
			if !reflect.DeepEqual(gotStatus, test.wantStatus) {
				t.Errorf("status does not match: %v, want: %v", gotStatus, test.wantStatus)
			}
		})
	}
}
