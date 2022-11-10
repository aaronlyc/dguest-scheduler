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
package helper

//
//import (
//	"fmt"
//	"testing"
//	"time"
//
//	st "dguest-scheduler/pkg/scheduler/testing"
//	"github.com/google/go-cmp/cmp"
//	v1 "k8s.io/api/core/v1"
//	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
//	"k8s.io/client-go/informers"
//	"k8s.io/client-go/kubernetes/fake"
//)
//
//func TestGetDguestServices(t *testing.T) {
//	fakeInformerFactory := informers.NewSharedInformerFactory(&fake.Clientset{}, 0*time.Second)
//	var services []*v1.Service
//	for i := 0; i < 3; i++ {
//		service := &v1.Service{
//			ObjectMeta: metav1.ObjectMeta{
//				Name:      fmt.Sprintf("service-%d", i),
//				Namespace: "test",
//			},
//			Spec: v1.ServiceSpec{
//				Selector: map[string]string{
//					"app": fmt.Sprintf("test-%d", i),
//				},
//			},
//		}
//		services = append(services, service)
//		fakeInformerFactory.Core().V1().Services().Informer().GetStore().Add(service)
//	}
//	var dguests []*v1alpha1.Dguest
//	for i := 0; i < 5; i++ {
//		dguest := st.MakeDguest().Name(fmt.Sprintf("test-dguest-%d", i)).
//			Namespace("test").
//			Label("app", fmt.Sprintf("test-%d", i)).
//			Label("label", fmt.Sprintf("label-%d", i)).
//			Obj()
//		dguests = append(dguests, dguest)
//	}
//
//	tests := []struct {
//		name   string
//		dguest *v1alpha1.Dguest
//		expect []*v1.Service
//	}{
//		{
//			name:   "GetDguestServices for dguest-0",
//			dguest: dguests[0],
//			expect: []*v1.Service{services[0]},
//		},
//		{
//			name:   "GetDguestServices for dguest-1",
//			dguest: dguests[1],
//			expect: []*v1.Service{services[1]},
//		},
//		{
//			name:   "GetDguestServices for dguest-2",
//			dguest: dguests[2],
//			expect: []*v1.Service{services[2]},
//		},
//		{
//			name:   "GetDguestServices for dguest-3",
//			dguest: dguests[3],
//			expect: nil,
//		},
//		{
//			name:   "GetDguestServices for dguest-4",
//			dguest: dguests[4],
//			expect: nil,
//		},
//	}
//	for _, test := range tests {
//		t.Run(test.name, func(t *testing.T) {
//			get, err := GetDguestServices(fakeInformerFactory.Core().V1().Services().Lister(), test.dguest)
//			if err != nil {
//				t.Errorf("Error from GetDguestServices: %v", err)
//			} else if diff := cmp.Diff(test.expect, get); diff != "" {
//				t.Errorf("Unexpected services (-want, +got):\n%s", diff)
//			}
//		})
//	}
//}
