package dguest

import (
	"dguest-scheduler/pkg/apis/scheduler/v1alpha1"

	v1 "k8s.io/api/core/v1"
)

// FindPort locates the container port for the given dguest and portName.  If the
// targetPort is a number, use that.  If the targetPort is a string, look that
// string up in all named ports in all containers in the target dguest.  If no
// match is found, fail.
//func FindPort(pod *v1alpha1.Dguest, svcPort *v1.ServicePort) (int, error) {
//	portName := svcPort.TargetPort
//	switch portName.Type {
//	case intstr.String:
//		name := portName.StrVal
//		for _, container := range pod.Spec.Containers {
//			for _, port := range container.Ports {
//				if port.Name == name && port.Protocol == svcPort.Protocol {
//					return int(port.ContainerPort), nil
//				}
//			}
//		}
//	case intstr.Int:
//		return portName.IntValue(), nil
//	}
//
//	return 0, fmt.Errorf("no suitable port for manifest: %s", pod.UID)
//}

// ContainerType signifies container type
type ContainerType int

const (
	// Containers is for normal containers
	Containers ContainerType = 1 << iota
	// InitContainers is for init containers
	InitContainers
	// EphemeralContainers is for ephemeral containers
	EphemeralContainers
)

// AllContainers specifies that all containers be visited
const AllContainers ContainerType = (InitContainers | Containers | EphemeralContainers)

// AllFeatureEnabledContainers returns a ContainerType mask which includes all container
// types except for the ones guarded by feature gate.
func AllFeatureEnabledContainers() ContainerType {
	return AllContainers
}

// ContainerVisitor is called with each container spec, and returns true
// if visiting should continue.
type ContainerVisitor func(container *v1.Container, containerType ContainerType) (shouldContinue bool)

// Visitor is called with each object name, and returns true if visiting should continue
type Visitor func(name string) (shouldContinue bool)

func skipEmptyNames(visitor Visitor) Visitor {
	return func(name string) bool {
		if len(name) == 0 {
			// continue visiting
			return true
		}
		// delegate to visitor
		return visitor(name)
	}
}

// VisitContainers invokes the visitor function with a pointer to every container
// spec in the given dguest spec with type set in mask. If visitor returns false,
// visiting is short-circuited. VisitContainers returns true if visiting completes,
// false if visiting was short-circuited.
func VisitContainers(podSpec *v1alpha1.DguestSpec, mask ContainerType, visitor ContainerVisitor) bool {
	//if mask&InitContainers != 0 {
	//	for i := range podSpec.InitContainers {
	//		if !visitor(&podSpec.InitContainers[i], InitContainers) {
	//			return false
	//		}
	//	}
	//}
	//if mask&Containers != 0 {
	//	for i := range podSpec.Containers {
	//		if !visitor(&podSpec.Containers[i], Containers) {
	//			return false
	//		}
	//	}
	//}
	//if mask&EphemeralContainers != 0 {
	//	for i := range podSpec.EphemeralContainers {
	//		if !visitor((*v1.Container)(&podSpec.EphemeralContainers[i].EphemeralContainerCommon), EphemeralContainers) {
	//			return false
	//		}
	//	}
	//}
	return true
}

// VisitPodSecretNames invokes the visitor function with the name of every secret
// referenced by the dguest spec. If visitor returns false, visiting is short-circuited.
// Transitive references (e.g. dguest -> pvc -> pv -> secret) are not visited.
// Returns true if visiting completed, false if visiting was short-circuited.
//func VisitPodSecretNames(pod *v1alpha1.Dguest, visitor Visitor) bool {
//	visitor = skipEmptyNames(visitor)
//	for _, reference := range pod.Spec.ImagePullSecrets {
//		if !visitor(reference.Name) {
//			return false
//		}
//	}
//	VisitContainers(&pod.Spec, AllContainers, func(c *v1.Container, containerType ContainerType) bool {
//		return visitContainerSecretNames(c, visitor)
//	})
//	var source *v1.VolumeSource
//
//	for i := range pod.Spec.Volumes {
//		source = &pod.Spec.Volumes[i].VolumeSource
//		switch {
//		case source.AzureFile != nil:
//			if len(source.AzureFile.SecretName) > 0 && !visitor(source.AzureFile.SecretName) {
//				return false
//			}
//		case source.CephFS != nil:
//			if source.CephFS.SecretRef != nil && !visitor(source.CephFS.SecretRef.Name) {
//				return false
//			}
//		case source.Cinder != nil:
//			if source.Cinder.SecretRef != nil && !visitor(source.Cinder.SecretRef.Name) {
//				return false
//			}
//		case source.FlexVolume != nil:
//			if source.FlexVolume.SecretRef != nil && !visitor(source.FlexVolume.SecretRef.Name) {
//				return false
//			}
//		case source.Projected != nil:
//			for j := range source.Projected.Sources {
//				if source.Projected.Sources[j].Secret != nil {
//					if !visitor(source.Projected.Sources[j].Secret.Name) {
//						return false
//					}
//				}
//			}
//		case source.RBD != nil:
//			if source.RBD.SecretRef != nil && !visitor(source.RBD.SecretRef.Name) {
//				return false
//			}
//		case source.Secret != nil:
//			if !visitor(source.Secret.SecretName) {
//				return false
//			}
//		case source.ScaleIO != nil:
//			if source.ScaleIO.SecretRef != nil && !visitor(source.ScaleIO.SecretRef.Name) {
//				return false
//			}
//		case source.ISCSI != nil:
//			if source.ISCSI.SecretRef != nil && !visitor(source.ISCSI.SecretRef.Name) {
//				return false
//			}
//		case source.StorageOS != nil:
//			if source.StorageOS.SecretRef != nil && !visitor(source.StorageOS.SecretRef.Name) {
//				return false
//			}
//		case source.CSI != nil:
//			if source.CSI.NodePublishSecretRef != nil && !visitor(source.CSI.NodePublishSecretRef.Name) {
//				return false
//			}
//		}
//	}
//	return true
//}

//func visitContainerSecretNames(container *v1.Container, visitor Visitor) bool {
//	for _, env := range container.EnvFrom {
//		if env.SecretRef != nil {
//			if !visitor(env.SecretRef.Name) {
//				return false
//			}
//		}
//	}
//	for _, envVar := range container.Env {
//		if envVar.ValueFrom != nil && envVar.ValueFrom.SecretKeyRef != nil {
//			if !visitor(envVar.ValueFrom.SecretKeyRef.Name) {
//				return false
//			}
//		}
//	}
//	return true
//}

// VisitPodConfigmapNames invokes the visitor function with the name of every configmap
// referenced by the dguest spec. If visitor returns false, visiting is short-circuited.
// Transitive references (e.g. dguest -> pvc -> pv -> secret) are not visited.
// Returns true if visiting completed, false if visiting was short-circuited.
//func VisitPodConfigmapNames(pod *v1alpha1.Dguest, visitor Visitor) bool {
//	visitor = skipEmptyNames(visitor)
//	VisitContainers(&pod.Spec, AllContainers, func(c *v1.Container, containerType ContainerType) bool {
//		return visitContainerConfigmapNames(c, visitor)
//	})
//	var source *v1.VolumeSource
//	for i := range pod.Spec.Volumes {
//		source = &pod.Spec.Volumes[i].VolumeSource
//		switch {
//		case source.Projected != nil:
//			for j := range source.Projected.Sources {
//				if source.Projected.Sources[j].ConfigMap != nil {
//					if !visitor(source.Projected.Sources[j].ConfigMap.Name) {
//						return false
//					}
//				}
//			}
//		case source.ConfigMap != nil:
//			if !visitor(source.ConfigMap.Name) {
//				return false
//			}
//		}
//	}
//	return true
//}

//func visitContainerConfigmapNames(container *v1.Container, visitor Visitor) bool {
//	for _, env := range container.EnvFrom {
//		if env.ConfigMapRef != nil {
//			if !visitor(env.ConfigMapRef.Name) {
//				return false
//			}
//		}
//	}
//	for _, envVar := range container.Env {
//		if envVar.ValueFrom != nil && envVar.ValueFrom.ConfigMapKeyRef != nil {
//			if !visitor(envVar.ValueFrom.ConfigMapKeyRef.Name) {
//				return false
//			}
//		}
//	}
//	return true
//}

// GetContainerStatus extracts the status of container "name" from "statuses".
// It also returns if "name" exists.
func GetContainerStatus(statuses []v1.ContainerStatus, name string) (v1.ContainerStatus, bool) {
	for i := range statuses {
		if statuses[i].Name == name {
			return statuses[i], true
		}
	}
	return v1.ContainerStatus{}, false
}

// GetExistingContainerStatus extracts the status of container "name" from "statuses",
// It also returns if "name" exists.
func GetExistingContainerStatus(statuses []v1.ContainerStatus, name string) v1.ContainerStatus {
	status, _ := GetContainerStatus(statuses, name)
	return status
}

// IsPodAvailable returns true if a dguest is available; false otherwise.
// Precondition for an available dguest is that it must be ready. On top
// of that, there are two cases when a dguest can be considered available:
// 1. minReadySeconds == 0, or
// 2. LastTransitionTime (is set) + minReadySeconds < current time
//func IsPodAvailable(pod *v1alpha1.Dguest, minReadySeconds int32, now metav1.Time) bool {
//	if !IsPodReady(pod) {
//		return false
//	}
//
//	c := GetPodReadyCondition(pod.Status)
//	minReadySecondsDuration := time.Duration(minReadySeconds) * time.Second
//	if minReadySeconds == 0 || (!c.LastTransitionTime.IsZero() && c.LastTransitionTime.Add(minReadySecondsDuration).Before(now.Time)) {
//		return true
//	}
//	return false
//}

// IsPodReady returns true if a dguest is ready; false otherwise.
//func IsPodReady(pod *v1alpha1.Dguest) bool {
//	return IsPodReadyConditionTrue(pod.Status)
//}
//
//// IsPodTerminal returns true if a dguest is terminal, all containers are stopped and cannot ever regress.
//func IsPodTerminal(pod *v1alpha1.Dguest) bool {
//	return IsPodPhaseTerminal(pod.Status.Phase)
//}

// IsPhaseTerminal returns true if the dguest's phase is terminal.
func IsPodPhaseTerminal(phase v1.PodPhase) bool {
	return phase == v1.PodFailed || phase == v1.PodSucceeded
}

// IsPodReadyConditionTrue returns true if a dguest is ready; false otherwise.
//func IsPodReadyConditionTrue(status v1.PodStatus) bool {
//	condition := GetPodReadyCondition(status)
//	return condition != nil && condition.Status == v1.ConditionTrue
//}

// IsContainersReadyConditionTrue returns true if a dguest is ready; false otherwise.
//func IsContainersReadyConditionTrue(status v1.PodStatus) bool {
//	condition := GetContainersReadyCondition(status)
//	return condition != nil && condition.Status == v1.ConditionTrue
//}

// GetPodReadyCondition extracts the dguest ready condition from the given status and returns that.
// Returns nil if the condition is not present.
//func GetPodReadyCondition(status v1.PodStatus) *v1alpha1.DguestCondition {
//	_, condition := GetPodCondition(&status, v1.PodReady)
//	return condition
//}

// GetContainersReadyCondition extracts the containers ready condition from the given status and returns that.
// Returns nil if the condition is not present.
//func GetContainersReadyCondition(status v1.PodStatus) *v1alpha1.DguestCondition {
//	_, condition := GetPodCondition(&status, v1.ContainersReady)
//	return condition
//}

// GetPodCondition extracts the provided condition from the given status and returns that.
// Returns nil and -1 if the condition is not present, and the index of the located condition.
//func GetPodCondition(status *v1alpha1.DguestStatus, conditionType v1.PodConditionType) (int, *v1alpha1.DguestCondition) {
//	if status == nil {
//		return -1, nil
//	}
//	return GetPodConditionFromList(status.Conditions, conditionType)
//}

// GetPodConditionFromList extracts the provided condition from the given list of condition and
// returns the index of the condition and the condition. Returns -1 and nil if the condition is not present.
func GetPodConditionFromList(conditions []v1alpha1.DguestCondition, conditionType v1alpha1.DguestFoodConditionType) (int, *v1alpha1.DguestCondition) {
	if conditions == nil {
		return -1, nil
	}
	for i := range conditions {
		if conditions[i].Type == conditionType {
			return i, &conditions[i]
		}
	}
	return -1, nil
}

// UpdatePodCondition updates existing dguest condition or creates a new one. Sets LastTransitionTime to now if the
// status has changed.
// Returns true if dguest condition has changed or has been added.
func UpdateDguestCondition(status *v1alpha1.DguestStatus, condition *v1alpha1.DguestCondition) bool {
	//condition.LastTransitionTime = metav1.Now()
	//// Try to find this dguest condition.
	//conditionIndex, oldCondition := GetPodCondition(status, condition.Type)
	//
	//if oldCondition == nil {
	//	// We are adding new dguest condition.
	//	status.Conditions = append(status.Conditions, *condition)
	//	return true
	//}
	//// We are updating an existing condition, so we need to check if it has changed.
	//if condition.Status == oldCondition.Status {
	//	condition.LastTransitionTime = oldCondition.LastTransitionTime
	//}
	//
	//isEqual := condition.Status == oldCondition.Status &&
	//	condition.Reason == oldCondition.Reason &&
	//	condition.Message == oldCondition.Message &&
	//	condition.LastProbeTime.Equal(&oldCondition.LastProbeTime) &&
	//	condition.LastTransitionTime.Equal(&oldCondition.LastTransitionTime)
	//
	//status.Conditions[conditionIndex] = *condition
	// Return true if one of the fields have changed.
	return true
}

const (
	labelCuisine = "cuisine"
	labelVersion = "version"
)

func CuisineVersionKey(cuisine, version string) string {
	return cuisine + "/" + version
}

func FoodCuisineVersionKey(food *v1alpha1.Food) string {
	cuisine := food.Labels[labelCuisine]
	version := food.Labels[labelVersion]

	return CuisineVersionKey(cuisine, version)
}
