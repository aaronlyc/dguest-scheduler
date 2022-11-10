/*
Copyright 2020 The Kubernetes Authors.

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

package defaultbinder

import (
	"context"
	"dguest-scheduler/pkg/apis/scheduler/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"dguest-scheduler/pkg/scheduler/framework"
	"dguest-scheduler/pkg/scheduler/framework/plugins/names"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
)

// Name of the plugin used in the plugin registry and configurations.
const Name = names.DefaultBinder

// DefaultBinder binds dguests to foods using a k8s client.
type DefaultBinder struct {
	handle framework.Handle
}

var _ framework.BindPlugin = &DefaultBinder{}

// New creates a DefaultBinder.
func New(_ runtime.Object, handle framework.Handle) (framework.Plugin, error) {
	return &DefaultBinder{handle: handle}, nil
}

// Name returns the name of the plugin.
func (b DefaultBinder) Name() string {
	return Name
}

// Bind binds dguests to foods using the k8s client.
func (b DefaultBinder) Bind(ctx context.Context, state *framework.CycleState, p *v1alpha1.Dguest, foodName string) *framework.Status {
	klog.V(3).InfoS("Attempting to bind dguest to food", "dguest", klog.KObj(p), "food", klog.KRef("", foodName))
	//binding := &v1.Binding{
	//	ObjectMeta: metav1.ObjectMeta{Namespace: p.Namespace, Name: p.Name, UID: p.UID},
	//	Target:     v1.ObjectReference{Kind: "Food", Name: foodName},
	//}
	//err := b.handle.ClientSet().CoreV1().Dguests(binding.Namespace).Bind(ctx, binding, metav1.CreateOptions{})
	// todo:
	//b.handle.SchedulerClientSet().SchedulerV1alpha1().Dguests(binding.Namespace)
	//if err != nil {
	//	return framework.AsStatus(err)
	//}
	p.Status.FoodsInfo = append(p.Status.FoodsInfo, v1alpha1.DguestFoodInfo{
		Name:           foodName,
		SchedulerdTime: metav1.Now(),
	})
	return nil
}
