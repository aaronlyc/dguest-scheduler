//go:build !ignore_autogenerated
// +build !ignore_autogenerated

/*
Copyright The Aaron Project.

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

// Code generated by deepcopy-gen. DO NOT EDIT.

package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Dguest) DeepCopyInto(out *Dguest) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Dguest.
func (in *Dguest) DeepCopy() *Dguest {
	if in == nil {
		return nil
	}
	out := new(Dguest)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Dguest) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DguestCondition) DeepCopyInto(out *DguestCondition) {
	*out = *in
	in.LastProbeTime.DeepCopyInto(&out.LastProbeTime)
	in.LastTransitionTime.DeepCopyInto(&out.LastTransitionTime)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DguestCondition.
func (in *DguestCondition) DeepCopy() *DguestCondition {
	if in == nil {
		return nil
	}
	out := new(DguestCondition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DguestFoodInfo) DeepCopyInto(out *DguestFoodInfo) {
	*out = *in
	in.SchedulerdTime.DeepCopyInto(&out.SchedulerdTime)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DguestFoodInfo.
func (in *DguestFoodInfo) DeepCopy() *DguestFoodInfo {
	if in == nil {
		return nil
	}
	out := new(DguestFoodInfo)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DguestList) DeepCopyInto(out *DguestList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Dguest, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DguestList.
func (in *DguestList) DeepCopy() *DguestList {
	if in == nil {
		return nil
	}
	out := new(DguestList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *DguestList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DguestSpec) DeepCopyInto(out *DguestSpec) {
	*out = *in
	if in.ActiveDeadlineSeconds != nil {
		in, out := &in.ActiveDeadlineSeconds, &out.ActiveDeadlineSeconds
		*out = new(int64)
		**out = **in
	}
	if in.WantBill != nil {
		in, out := &in.WantBill, &out.WantBill
		*out = make([]Dish, len(*in))
		copy(*out, *in)
	}
	if in.Tolerations != nil {
		in, out := &in.Tolerations, &out.Tolerations
		*out = make([]v1.Toleration, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Overhead != nil {
		in, out := &in.Overhead, &out.Overhead
		*out = make(ResourceList, len(*in))
		for key, val := range *in {
			(*out)[key] = val.DeepCopy()
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DguestSpec.
func (in *DguestSpec) DeepCopy() *DguestSpec {
	if in == nil {
		return nil
	}
	out := new(DguestSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DguestStatus) DeepCopyInto(out *DguestStatus) {
	*out = *in
	if in.FoodsInfo != nil {
		in, out := &in.FoodsInfo, &out.FoodsInfo
		*out = make(map[string]FoodsInfoSlice, len(*in))
		for key, val := range *in {
			var outVal []DguestFoodInfo
			if val == nil {
				(*out)[key] = nil
			} else {
				in, out := &val, &outVal
				*out = make(FoodsInfoSlice, len(*in))
				for i := range *in {
					(*in)[i].DeepCopyInto(&(*out)[i])
				}
			}
			(*out)[key] = outVal
		}
	}
	if in.OriginalFoods != nil {
		in, out := &in.OriginalFoods, &out.OriginalFoods
		*out = make([]OriginalFoodInfo, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]DguestCondition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DguestStatus.
func (in *DguestStatus) DeepCopy() *DguestStatus {
	if in == nil {
		return nil
	}
	out := new(DguestStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Dish) DeepCopyInto(out *Dish) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Dish.
func (in *Dish) DeepCopy() *Dish {
	if in == nil {
		return nil
	}
	out := new(Dish)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Food) DeepCopyInto(out *Food) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Food.
func (in *Food) DeepCopy() *Food {
	if in == nil {
		return nil
	}
	out := new(Food)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Food) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FoodCondition) DeepCopyInto(out *FoodCondition) {
	*out = *in
	in.LastHeartbeatTime.DeepCopyInto(&out.LastHeartbeatTime)
	in.LastTransitionTime.DeepCopyInto(&out.LastTransitionTime)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FoodCondition.
func (in *FoodCondition) DeepCopy() *FoodCondition {
	if in == nil {
		return nil
	}
	out := new(FoodCondition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FoodInfo) DeepCopyInto(out *FoodInfo) {
	*out = *in
	in.ReleaseTime.DeepCopyInto(&out.ReleaseTime)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]IngredientStatus, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FoodInfo.
func (in *FoodInfo) DeepCopy() *FoodInfo {
	if in == nil {
		return nil
	}
	out := new(FoodInfo)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FoodInfoBase) DeepCopyInto(out *FoodInfoBase) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FoodInfoBase.
func (in *FoodInfoBase) DeepCopy() *FoodInfoBase {
	if in == nil {
		return nil
	}
	out := new(FoodInfoBase)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FoodList) DeepCopyInto(out *FoodList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Food, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FoodList.
func (in *FoodList) DeepCopy() *FoodList {
	if in == nil {
		return nil
	}
	out := new(FoodList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *FoodList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FoodSpec) DeepCopyInto(out *FoodSpec) {
	*out = *in
	if in.Taints != nil {
		in, out := &in.Taints, &out.Taints
		*out = make([]v1.Taint, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FoodSpec.
func (in *FoodSpec) DeepCopy() *FoodSpec {
	if in == nil {
		return nil
	}
	out := new(FoodSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FoodStatus) DeepCopyInto(out *FoodStatus) {
	*out = *in
	if in.Capacity != nil {
		in, out := &in.Capacity, &out.Capacity
		*out = make(ResourceList, len(*in))
		for key, val := range *in {
			(*out)[key] = val.DeepCopy()
		}
	}
	if in.Allocatable != nil {
		in, out := &in.Allocatable, &out.Allocatable
		*out = make(ResourceList, len(*in))
		for key, val := range *in {
			(*out)[key] = val.DeepCopy()
		}
	}
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]FoodCondition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.FoodInfo != nil {
		in, out := &in.FoodInfo, &out.FoodInfo
		*out = new(FoodInfo)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FoodStatus.
func (in *FoodStatus) DeepCopy() *FoodStatus {
	if in == nil {
		return nil
	}
	out := new(FoodStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in FoodsInfoSlice) DeepCopyInto(out *FoodsInfoSlice) {
	{
		in := &in
		*out = make(FoodsInfoSlice, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
		return
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FoodsInfoSlice.
func (in FoodsInfoSlice) DeepCopy() FoodsInfoSlice {
	if in == nil {
		return nil
	}
	out := new(FoodsInfoSlice)
	in.DeepCopyInto(out)
	return *out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IngredientStatus) DeepCopyInto(out *IngredientStatus) {
	*out = *in
	in.CreateTime.DeepCopyInto(&out.CreateTime)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IngredientStatus.
func (in *IngredientStatus) DeepCopy() *IngredientStatus {
	if in == nil {
		return nil
	}
	out := new(IngredientStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OriginalFoodInfo) DeepCopyInto(out *OriginalFoodInfo) {
	*out = *in
	out.FoodInfoBase = in.FoodInfoBase
	in.RemoveTime.DeepCopyInto(&out.RemoveTime)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OriginalFoodInfo.
func (in *OriginalFoodInfo) DeepCopy() *OriginalFoodInfo {
	if in == nil {
		return nil
	}
	out := new(OriginalFoodInfo)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in ResourceList) DeepCopyInto(out *ResourceList) {
	{
		in := &in
		*out = make(ResourceList, len(*in))
		for key, val := range *in {
			(*out)[key] = val.DeepCopy()
		}
		return
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourceList.
func (in ResourceList) DeepCopy() ResourceList {
	if in == nil {
		return nil
	}
	out := new(ResourceList)
	in.DeepCopyInto(out)
	return *out
}
