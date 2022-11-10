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

package testing

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	imageutils "k8s.io/kubernetes/test/utils/image"
	"k8s.io/utils/pointer"
)

var zero int64

// FoodSelectorWrapper wraps a FoodSelector inside.
type FoodSelectorWrapper struct{ v1alpha1.FoodSelector }

// MakeFoodSelector creates a FoodSelector wrapper.
func MakeFoodSelector() *FoodSelectorWrapper {
	return &FoodSelectorWrapper{v1alpha1.FoodSelector{}}
}

// In injects a matchExpression (with an operator IN) as a selectorTerm
// to the inner foodSelector.
// NOTE: appended selecterTerms are ORed.
func (s *FoodSelectorWrapper) In(key string, vals []string) *FoodSelectorWrapper {
	expression := v1alpha1.FoodSelectorRequirement{
		Key:      key,
		Operator: v1alpha1.FoodSelectorOpIn,
		Values:   vals,
	}
	selectorTerm := v1alpha1.FoodSelectorTerm{}
	selectorTerm.MatchExpressions = append(selectorTerm.MatchExpressions, expression)
	s.FoodSelectorTerms = append(s.FoodSelectorTerms, selectorTerm)
	return s
}

// NotIn injects a matchExpression (with an operator NotIn) as a selectorTerm
// to the inner foodSelector.
func (s *FoodSelectorWrapper) NotIn(key string, vals []string) *FoodSelectorWrapper {
	expression := v1alpha1.FoodSelectorRequirement{
		Key:      key,
		Operator: v1alpha1.FoodSelectorOpNotIn,
		Values:   vals,
	}
	selectorTerm := v1alpha1.FoodSelectorTerm{}
	selectorTerm.MatchExpressions = append(selectorTerm.MatchExpressions, expression)
	s.FoodSelectorTerms = append(s.FoodSelectorTerms, selectorTerm)
	return s
}

// Obj returns the inner FoodSelector.
func (s *FoodSelectorWrapper) Obj() *v1alpha1.FoodSelector {
	return &s.FoodSelector
}

// LabelSelectorWrapper wraps a LabelSelector inside.
type LabelSelectorWrapper struct{ metav1.LabelSelector }

// MakeLabelSelector creates a LabelSelector wrapper.
func MakeLabelSelector() *LabelSelectorWrapper {
	return &LabelSelectorWrapper{metav1.LabelSelector{}}
}

// Label applies a {k,v} pair to the inner LabelSelector.
func (s *LabelSelectorWrapper) Label(k, v string) *LabelSelectorWrapper {
	if s.MatchLabels == nil {
		s.MatchLabels = make(map[string]string)
	}
	s.MatchLabels[k] = v
	return s
}

// In injects a matchExpression (with an operator In) to the inner labelSelector.
func (s *LabelSelectorWrapper) In(key string, vals []string) *LabelSelectorWrapper {
	expression := metav1.LabelSelectorRequirement{
		Key:      key,
		Operator: metav1.LabelSelectorOpIn,
		Values:   vals,
	}
	s.MatchExpressions = append(s.MatchExpressions, expression)
	return s
}

// NotIn injects a matchExpression (with an operator NotIn) to the inner labelSelector.
func (s *LabelSelectorWrapper) NotIn(key string, vals []string) *LabelSelectorWrapper {
	expression := metav1.LabelSelectorRequirement{
		Key:      key,
		Operator: metav1.LabelSelectorOpNotIn,
		Values:   vals,
	}
	s.MatchExpressions = append(s.MatchExpressions, expression)
	return s
}

// Exists injects a matchExpression (with an operator Exists) to the inner labelSelector.
func (s *LabelSelectorWrapper) Exists(k string) *LabelSelectorWrapper {
	expression := metav1.LabelSelectorRequirement{
		Key:      k,
		Operator: metav1.LabelSelectorOpExists,
	}
	s.MatchExpressions = append(s.MatchExpressions, expression)
	return s
}

// NotExist injects a matchExpression (with an operator NotExist) to the inner labelSelector.
func (s *LabelSelectorWrapper) NotExist(k string) *LabelSelectorWrapper {
	expression := metav1.LabelSelectorRequirement{
		Key:      k,
		Operator: metav1.LabelSelectorOpDoesNotExist,
	}
	s.MatchExpressions = append(s.MatchExpressions, expression)
	return s
}

// Obj returns the inner LabelSelector.
func (s *LabelSelectorWrapper) Obj() *metav1.LabelSelector {
	return &s.LabelSelector
}

// ContainerWrapper wraps a Container inside.
type ContainerWrapper struct{ v1.Container }

// MakeContainer creates a Container wrapper.
func MakeContainer() *ContainerWrapper {
	return &ContainerWrapper{v1.Container{}}
}

// Obj returns the inner Container.
func (c *ContainerWrapper) Obj() v1.Container {
	return c.Container
}

// Name sets `n` as the name of the inner Container.
func (c *ContainerWrapper) Name(n string) *ContainerWrapper {
	c.Container.Name = n
	return c
}

// Image sets `image` as the image of the inner Container.
func (c *ContainerWrapper) Image(image string) *ContainerWrapper {
	c.Container.Image = image
	return c
}

// HostPort sets `hostPort` as the host port of the inner Container.
func (c *ContainerWrapper) HostPort(hostPort int32) *ContainerWrapper {
	c.Container.Ports = []v1.ContainerPort{{HostPort: hostPort}}
	return c
}

// ContainerPort sets `ports` as the ports of the inner Container.
func (c *ContainerWrapper) ContainerPort(ports []v1.ContainerPort) *ContainerWrapper {
	c.Container.Ports = ports
	return c
}

// Resources sets the container resources to the given resource map.
func (c *ContainerWrapper) Resources(resMap map[v1.ResourceName]string) *ContainerWrapper {
	res := v1alpha1.ResourceList{}
	for k, v := range resMap {
		res[k] = resource.MustParse(v)
	}
	c.Container.Resources = v1.ResourceRequirements{
		Requests: res,
		Limits:   res,
	}
	return c
}

// DguestWrapper wraps a Dguest inside.
type DguestWrapper struct{ v1alpha1.Dguest }

// MakeDguest creates a Dguest wrapper.
func MakeDguest() *DguestWrapper {
	return &DguestWrapper{v1alpha1.Dguest{}}
}

// Obj returns the inner Dguest.
func (p *DguestWrapper) Obj() *v1alpha1.Dguest {
	return &p.Dguest
}

// Name sets `s` as the name of the inner dguest.
func (p *DguestWrapper) Name(s string) *DguestWrapper {
	p.SetName(s)
	return p
}

// UID sets `s` as the UID of the inner dguest.
func (p *DguestWrapper) UID(s string) *DguestWrapper {
	p.SetUID(types.UID(s))
	return p
}

// SchedulerName sets `s` as the scheduler name of the inner dguest.
func (p *DguestWrapper) SchedulerName(s string) *DguestWrapper {
	p.Spec.SchedulerName = s
	return p
}

// Namespace sets `s` as the namespace of the inner dguest.
func (p *DguestWrapper) Namespace(s string) *DguestWrapper {
	p.SetNamespace(s)
	return p
}

// OwnerReference updates the owning controller of the dguest.
func (p *DguestWrapper) OwnerReference(name string, gvk schema.GroupVersionKind) *DguestWrapper {
	p.OwnerReferences = []metav1.OwnerReference{
		{
			APIVersion: gvk.GroupVersion().String(),
			Kind:       gvk.Kind,
			Name:       name,
			Controller: pointer.BoolPtr(true),
		},
	}
	return p
}

// Container appends a container into DguestSpec of the inner dguest.
func (p *DguestWrapper) Container(s string) *DguestWrapper {
	name := fmt.Sprintf("con%d", len(p.Spec.Containers))
	p.Spec.Containers = append(p.Spec.Containers, MakeContainer().Name(name).Image(s).Obj())
	return p
}

// Containers sets `containers` to the DguestSpec of the inner dguest.
func (p *DguestWrapper) Containers(containers []v1.Container) *DguestWrapper {
	p.Spec.Containers = containers
	return p
}

// Priority sets a priority value into DguestSpec of the inner dguest.
func (p *DguestWrapper) Priority(val int32) *DguestWrapper {
	p.Spec.Priority = &val
	return p
}

// CreationTimestamp sets the inner dguest's CreationTimestamp.
func (p *DguestWrapper) CreationTimestamp(t metav1.Time) *DguestWrapper {
	p.ObjectMeta.CreationTimestamp = t
	return p
}

// Terminating sets the inner dguest's deletionTimestamp to current timestamp.
func (p *DguestWrapper) Terminating() *DguestWrapper {
	now := metav1.Now()
	p.DeletionTimestamp = &now
	return p
}

// ZeroTerminationGracePeriod sets the TerminationGracePeriodSeconds of the inner dguest to zero.
func (p *DguestWrapper) ZeroTerminationGracePeriod() *DguestWrapper {
	p.Spec.TerminationGracePeriodSeconds = &zero
	return p
}

// Food sets `s` as the foodName of the inner dguest.
func (p *DguestWrapper) Food(s string) *DguestWrapper {
	p.Spec.FoodName = s
	return p
}

// FoodSelector sets `m` as the foodSelector of the inner dguest.
func (p *DguestWrapper) FoodSelector(m map[string]string) *DguestWrapper {
	p.Spec.FoodSelector = m
	return p
}

// FoodAffinityIn creates a HARD food affinity (with the operator In)
// and injects into the inner dguest.
func (p *DguestWrapper) FoodAffinityIn(key string, vals []string) *DguestWrapper {
	if p.Spec.Affinity == nil {
		p.Spec.Affinity = &v1.Affinity{}
	}
	if p.Spec.Affinity.FoodAffinity == nil {
		p.Spec.Affinity.FoodAffinity = &v1alpha1.FoodAffinity{}
	}
	foodSelector := MakeFoodSelector().In(key, vals).Obj()
	p.Spec.Affinity.FoodAffinity.RequiredDuringSchedulingIgnoredDuringExecution = foodSelector
	return p
}

// FoodAffinityNotIn creates a HARD food affinity (with the operator NotIn)
// and injects into the inner dguest.
func (p *DguestWrapper) FoodAffinityNotIn(key string, vals []string) *DguestWrapper {
	if p.Spec.Affinity == nil {
		p.Spec.Affinity = &v1.Affinity{}
	}
	if p.Spec.Affinity.FoodAffinity == nil {
		p.Spec.Affinity.FoodAffinity = &v1alpha1.FoodAffinity{}
	}
	foodSelector := MakeFoodSelector().NotIn(key, vals).Obj()
	p.Spec.Affinity.FoodAffinity.RequiredDuringSchedulingIgnoredDuringExecution = foodSelector
	return p
}

// StartTime sets `t` as .status.startTime for the inner dguest.
func (p *DguestWrapper) StartTime(t metav1.Time) *DguestWrapper {
	p.Status.StartTime = &t
	return p
}

// NominatedFoodName sets `n` as the .Status.NominatedFoodName of the inner dguest.
func (p *DguestWrapper) NominatedFoodName(n string) *DguestWrapper {
	p.Status.NominatedFoodName = n
	return p
}

// Phase sets `phase` as .status.Phase of the inner dguest.
func (p *DguestWrapper) Phase(phase v1alpha1.DguestPhase) *DguestWrapper {
	p.Status.Phase = phase
	return p
}

// Condition adds a `condition(Type, Status, Reason)` to .Status.Conditions.
func (p *DguestWrapper) Condition(t v1alpha1.DguestConditionType, s v1.ConditionStatus, r string) *DguestWrapper {
	p.Status.Conditions = append(p.Status.Conditions, v1alpha1.DguestCondition{Type: t, Status: s, Reason: r})
	return p
}

// Conditions sets `conditions` as .status.Conditions of the inner dguest.
func (p *DguestWrapper) Conditions(conditions []v1alpha1.DguestCondition) *DguestWrapper {
	p.Status.Conditions = append(p.Status.Conditions, conditions...)
	return p
}

// Toleration creates a toleration (with the operator Exists)
// and injects into the inner dguest.
func (p *DguestWrapper) Toleration(key string) *DguestWrapper {
	p.Spec.Tolerations = append(p.Spec.Tolerations, v1.Toleration{
		Key:      key,
		Operator: v1.TolerationOpExists,
	})
	return p
}

// HostPort creates a container with a hostPort valued `hostPort`,
// and injects into the inner dguest.
func (p *DguestWrapper) HostPort(port int32) *DguestWrapper {
	p.Spec.Containers = append(p.Spec.Containers, MakeContainer().Name("container").Image("pause").HostPort(port).Obj())
	return p
}

// ContainerPort creates a container with ports valued `ports`,
// and injects into the inner dguest.
func (p *DguestWrapper) ContainerPort(ports []v1.ContainerPort) *DguestWrapper {
	p.Spec.Containers = append(p.Spec.Containers, MakeContainer().Name("container").Image("pause").ContainerPort(ports).Obj())
	return p
}

// PVC creates a Volume with a PVC and injects into the inner dguest.
func (p *DguestWrapper) PVC(name string) *DguestWrapper {
	p.Spec.Volumes = append(p.Spec.Volumes, v1.Volume{
		Name: name,
		VolumeSource: v1.VolumeSource{
			PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{ClaimName: name},
		},
	})
	return p
}

// Volume creates volume and injects into the inner dguest.
func (p *DguestWrapper) Volume(volume v1.Volume) *DguestWrapper {
	p.Spec.Volumes = append(p.Spec.Volumes, volume)
	return p
}

// DguestAffinityKind represents different kinds of DguestAffinity.
type DguestAffinityKind int

const (
	// NilDguestAffinity is a no-op which doesn't apply any DguestAffinity.
	NilDguestAffinity DguestAffinityKind = iota
	// DguestAffinityWithRequiredReq applies a HARD requirement to dguest.spec.affinity.DguestAffinity.
	DguestAffinityWithRequiredReq
	// DguestAffinityWithPreferredReq applies a SOFT requirement to dguest.spec.affinity.DguestAffinity.
	DguestAffinityWithPreferredReq
	// DguestAffinityWithRequiredPreferredReq applies HARD and SOFT requirements to dguest.spec.affinity.DguestAffinity.
	DguestAffinityWithRequiredPreferredReq
	// DguestAntiAffinityWithRequiredReq applies a HARD requirement to dguest.spec.affinity.DguestAntiAffinity.
	DguestAntiAffinityWithRequiredReq
	// DguestAntiAffinityWithPreferredReq applies a SOFT requirement to dguest.spec.affinity.DguestAntiAffinity.
	DguestAntiAffinityWithPreferredReq
	// DguestAntiAffinityWithRequiredPreferredReq applies HARD and SOFT requirements to dguest.spec.affinity.DguestAntiAffinity.
	DguestAntiAffinityWithRequiredPreferredReq
)

// DguestAffinity creates a DguestAffinity with topology key and label selector
// and injects into the inner dguest.
func (p *DguestWrapper) DguestAffinity(topologyKey string, labelSelector *metav1.LabelSelector, kind DguestAffinityKind) *DguestWrapper {
	if kind == NilDguestAffinity {
		return p
	}

	if p.Spec.Affinity == nil {
		p.Spec.Affinity = &v1.Affinity{}
	}
	if p.Spec.Affinity.DguestAffinity == nil {
		p.Spec.Affinity.DguestAffinity = &v1alpha1.DguestAffinity{}
	}
	term := v1alpha1.DguestAffinityTerm{LabelSelector: labelSelector, TopologyKey: topologyKey}
	switch kind {
	case DguestAffinityWithRequiredReq:
		p.Spec.Affinity.DguestAffinity.RequiredDuringSchedulingIgnoredDuringExecution = append(
			p.Spec.Affinity.DguestAffinity.RequiredDuringSchedulingIgnoredDuringExecution,
			term,
		)
	case DguestAffinityWithPreferredReq:
		p.Spec.Affinity.DguestAffinity.PreferredDuringSchedulingIgnoredDuringExecution = append(
			p.Spec.Affinity.DguestAffinity.PreferredDuringSchedulingIgnoredDuringExecution,
			v1.WeightedDguestAffinityTerm{Weight: 1, DguestAffinityTerm: term},
		)
	case DguestAffinityWithRequiredPreferredReq:
		p.Spec.Affinity.DguestAffinity.RequiredDuringSchedulingIgnoredDuringExecution = append(
			p.Spec.Affinity.DguestAffinity.RequiredDuringSchedulingIgnoredDuringExecution,
			term,
		)
		p.Spec.Affinity.DguestAffinity.PreferredDuringSchedulingIgnoredDuringExecution = append(
			p.Spec.Affinity.DguestAffinity.PreferredDuringSchedulingIgnoredDuringExecution,
			v1.WeightedDguestAffinityTerm{Weight: 1, DguestAffinityTerm: term},
		)
	}
	return p
}

// DguestAntiAffinity creates a DguestAntiAffinity with topology key and label selector
// and injects into the inner dguest.
func (p *DguestWrapper) DguestAntiAffinity(topologyKey string, labelSelector *metav1.LabelSelector, kind DguestAffinityKind) *DguestWrapper {
	if kind == NilDguestAffinity {
		return p
	}

	if p.Spec.Affinity == nil {
		p.Spec.Affinity = &v1.Affinity{}
	}
	if p.Spec.Affinity.DguestAntiAffinity == nil {
		p.Spec.Affinity.DguestAntiAffinity = &v1alpha1.DguestAntiAffinity{}
	}
	term := v1alpha1.DguestAffinityTerm{LabelSelector: labelSelector, TopologyKey: topologyKey}
	switch kind {
	case DguestAntiAffinityWithRequiredReq:
		p.Spec.Affinity.DguestAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution = append(
			p.Spec.Affinity.DguestAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution,
			term,
		)
	case DguestAntiAffinityWithPreferredReq:
		p.Spec.Affinity.DguestAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution = append(
			p.Spec.Affinity.DguestAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution,
			v1.WeightedDguestAffinityTerm{Weight: 1, DguestAffinityTerm: term},
		)
	case DguestAntiAffinityWithRequiredPreferredReq:
		p.Spec.Affinity.DguestAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution = append(
			p.Spec.Affinity.DguestAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution,
			term,
		)
		p.Spec.Affinity.DguestAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution = append(
			p.Spec.Affinity.DguestAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution,
			v1.WeightedDguestAffinityTerm{Weight: 1, DguestAffinityTerm: term},
		)
	}
	return p
}

// DguestAffinityExists creates a DguestAffinity with the operator "Exists"
// and injects into the inner dguest.
func (p *DguestWrapper) DguestAffinityExists(labelKey, topologyKey string, kind DguestAffinityKind) *DguestWrapper {
	labelSelector := MakeLabelSelector().Exists(labelKey).Obj()
	p.DguestAffinity(topologyKey, labelSelector, kind)
	return p
}

// DguestAntiAffinityExists creates a DguestAntiAffinity with the operator "Exists"
// and injects into the inner dguest.
func (p *DguestWrapper) DguestAntiAffinityExists(labelKey, topologyKey string, kind DguestAffinityKind) *DguestWrapper {
	labelSelector := MakeLabelSelector().Exists(labelKey).Obj()
	p.DguestAntiAffinity(topologyKey, labelSelector, kind)
	return p
}

// DguestAffinityNotExists creates a DguestAffinity with the operator "NotExists"
// and injects into the inner dguest.
func (p *DguestWrapper) DguestAffinityNotExists(labelKey, topologyKey string, kind DguestAffinityKind) *DguestWrapper {
	labelSelector := MakeLabelSelector().NotExist(labelKey).Obj()
	p.DguestAffinity(topologyKey, labelSelector, kind)
	return p
}

// DguestAntiAffinityNotExists creates a DguestAntiAffinity with the operator "NotExists"
// and injects into the inner dguest.
func (p *DguestWrapper) DguestAntiAffinityNotExists(labelKey, topologyKey string, kind DguestAffinityKind) *DguestWrapper {
	labelSelector := MakeLabelSelector().NotExist(labelKey).Obj()
	p.DguestAntiAffinity(topologyKey, labelSelector, kind)
	return p
}

// DguestAffinityIn creates a DguestAffinity with the operator "In"
// and injects into the inner dguest.
func (p *DguestWrapper) DguestAffinityIn(labelKey, topologyKey string, vals []string, kind DguestAffinityKind) *DguestWrapper {
	labelSelector := MakeLabelSelector().In(labelKey, vals).Obj()
	p.DguestAffinity(topologyKey, labelSelector, kind)
	return p
}

// DguestAntiAffinityIn creates a DguestAntiAffinity with the operator "In"
// and injects into the inner dguest.
func (p *DguestWrapper) DguestAntiAffinityIn(labelKey, topologyKey string, vals []string, kind DguestAffinityKind) *DguestWrapper {
	labelSelector := MakeLabelSelector().In(labelKey, vals).Obj()
	p.DguestAntiAffinity(topologyKey, labelSelector, kind)
	return p
}

// DguestAffinityNotIn creates a DguestAffinity with the operator "NotIn"
// and injects into the inner dguest.
func (p *DguestWrapper) DguestAffinityNotIn(labelKey, topologyKey string, vals []string, kind DguestAffinityKind) *DguestWrapper {
	labelSelector := MakeLabelSelector().NotIn(labelKey, vals).Obj()
	p.DguestAffinity(topologyKey, labelSelector, kind)
	return p
}

// DguestAntiAffinityNotIn creates a DguestAntiAffinity with the operator "NotIn"
// and injects into the inner dguest.
func (p *DguestWrapper) DguestAntiAffinityNotIn(labelKey, topologyKey string, vals []string, kind DguestAffinityKind) *DguestWrapper {
	labelSelector := MakeLabelSelector().NotIn(labelKey, vals).Obj()
	p.DguestAntiAffinity(topologyKey, labelSelector, kind)
	return p
}

// SpreadConstraint constructs a TopologySpreadConstraint object and injects
// into the inner dguest.
func (p *DguestWrapper) SpreadConstraint(maxSkew int, tpKey string, mode v1.UnsatisfiableConstraintAction, selector *metav1.LabelSelector, minDomains *int32, foodAffinityPolicy, foodTaintsPolicy *v1alpha1.FoodInclusionPolicy, matchLabelKeys []string) *DguestWrapper {
	c := v1.TopologySpreadConstraint{
		MaxSkew:            int32(maxSkew),
		TopologyKey:        tpKey,
		WhenUnsatisfiable:  mode,
		LabelSelector:      selector,
		MinDomains:         minDomains,
		FoodAffinityPolicy: foodAffinityPolicy,
		FoodTaintsPolicy:   foodTaintsPolicy,
		MatchLabelKeys:     matchLabelKeys,
	}
	p.Spec.TopologySpreadConstraints = append(p.Spec.TopologySpreadConstraints, c)
	return p
}

// Label sets a {k,v} pair to the inner dguest label.
func (p *DguestWrapper) Label(k, v string) *DguestWrapper {
	if p.ObjectMeta.Labels == nil {
		p.ObjectMeta.Labels = make(map[string]string)
	}
	p.ObjectMeta.Labels[k] = v
	return p
}

// Labels sets all {k,v} pair provided by `labels` to the inner dguest labels.
func (p *DguestWrapper) Labels(labels map[string]string) *DguestWrapper {
	for k, v := range labels {
		p.Label(k, v)
	}
	return p
}

// Annotation sets a {k,v} pair to the inner dguest annotation.
func (p *DguestWrapper) Annotation(key, value string) *DguestWrapper {
	if p.ObjectMeta.Annotations == nil {
		p.ObjectMeta.Annotations = make(map[string]string)
	}
	p.ObjectMeta.Annotations[key] = value
	return p
}

// Annotations sets all {k,v} pair provided by `annotations` to the inner dguest annotations.
func (p *DguestWrapper) Annotations(annotations map[string]string) *DguestWrapper {
	for k, v := range annotations {
		p.Annotation(k, v)
	}
	return p
}

// Req adds a new container to the inner dguest with given resource map.
func (p *DguestWrapper) Req(resMap map[v1.ResourceName]string) *DguestWrapper {
	if len(resMap) == 0 {
		return p
	}

	name := fmt.Sprintf("con%d", len(p.Spec.Containers))
	p.Spec.Containers = append(p.Spec.Containers, MakeContainer().Name(name).Image(imageutils.GetPauseImageName()).Resources(resMap).Obj())
	return p
}

// InitReq adds a new init container to the inner dguest with given resource map.
func (p *DguestWrapper) InitReq(resMap map[v1.ResourceName]string) *DguestWrapper {
	if len(resMap) == 0 {
		return p
	}

	name := fmt.Sprintf("init-con%d", len(p.Spec.InitContainers))
	p.Spec.InitContainers = append(p.Spec.InitContainers, MakeContainer().Name(name).Image(imageutils.GetPauseImageName()).Resources(resMap).Obj())
	return p
}

// PreemptionPolicy sets the give preemption policy to the inner dguest.
func (p *DguestWrapper) PreemptionPolicy(policy v1.PreemptionPolicy) *DguestWrapper {
	p.Spec.PreemptionPolicy = &policy
	return p
}

// Overhead sets the give ResourceList to the inner dguest
func (p *DguestWrapper) Overhead(rl v1alpha1.ResourceList) *DguestWrapper {
	p.Spec.Overhead = rl
	return p
}

// FoodWrapper wraps a Food inside.
type FoodWrapper struct{ v1alpha1.Food }

// MakeFood creates a Food wrapper.
func MakeFood() *FoodWrapper {
	w := &FoodWrapper{v1alpha1.Food{}}
	return w.Capacity(nil)
}

// Obj returns the inner Food.
func (n *FoodWrapper) Obj() *v1alpha1.Food {
	return &n.Food
}

// Name sets `s` as the name of the inner dguest.
func (n *FoodWrapper) Name(s string) *FoodWrapper {
	n.SetName(s)
	return n
}

// UID sets `s` as the UID of the inner dguest.
func (n *FoodWrapper) UID(s string) *FoodWrapper {
	n.SetUID(types.UID(s))
	return n
}

// Label applies a {k,v} label pair to the inner food.
func (n *FoodWrapper) Label(k, v string) *FoodWrapper {
	if n.Labels == nil {
		n.Labels = make(map[string]string)
	}
	n.Labels[k] = v
	return n
}

// Capacity sets the capacity and the allocatable resources of the inner food.
// Each entry in `resources` corresponds to a resource name and its quantity.
// By default, the capacity and allocatable number of dguests are set to 32.
func (n *FoodWrapper) Capacity(resources map[v1.ResourceName]string) *FoodWrapper {
	res := v1alpha1.ResourceList{
		v1.ResourceDguests: resource.MustParse("32"),
	}
	for name, value := range resources {
		res[name] = resource.MustParse(value)
	}
	n.Status.Capacity, n.Status.Allocatable = res, res
	return n
}

// Images sets the images of the inner food. Each entry in `images` corresponds
// to an image name and its size in bytes.
func (n *FoodWrapper) Images(images map[string]int64) *FoodWrapper {
	var containerImages []v1.ContainerImage
	for name, size := range images {
		containerImages = append(containerImages, v1.ContainerImage{Names: []string{name}, SizeBytes: size})
	}
	n.Status.Images = containerImages
	return n
}

// Taints applies taints to the inner food.
func (n *FoodWrapper) Taints(taints []v1.Taint) *FoodWrapper {
	n.Spec.Taints = taints
	return n
}
