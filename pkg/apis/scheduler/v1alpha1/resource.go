package v1alpha1

import "k8s.io/apimachinery/pkg/api/resource"

// ResourceName is the name identifying various resources in a ResourceList.
type ResourceName string

// Resource names must be not more than 63 characters, consisting of upper- or lower-case alphanumeric characters,
// with the -, _, and . characters allowed anywhere, except the first or last character.
// The default convention, matching that for annotations, is to use lower-case names, with dashes, rather than
// camel case, separating compound words.
// Fully-qualified resource typenames are constructed from a DNS-style subdomain, followed by a slash `/` and a name.
const (
	// CPU, in cores. (500m = .5 cores)
	ResourceCPU ResourceName = "cpu"
	// Memory, in bytes. (500Gi = 500GiB = 500 * 1024 * 1024 * 1024)
	ResourceMemory ResourceName = "memory"
	// Volume size, in bytes (e,g. 5Gi = 5GiB = 5 * 1024 * 1024 * 1024)
	ResourceStorage ResourceName = "storage"
	// Bandwidth for the tenant.
	ResourceBandwidth ResourceName = "bandwidth"
	ResourceGate      ResourceName = "gate"
)

// ResourceList is a set of (resource name, quantity) pairs.
type ResourceList map[ResourceName]resource.Quantity

// Returns string version of ResourceName.
func (rn ResourceName) String() string {
	return string(rn)
}

// Cpu returns the Cpu limit if specified.
func (rl *ResourceList) Cpu() *resource.Quantity {
	return rl.Name(ResourceCPU, resource.DecimalSI)
}

// Memory returns the Memory limit if specified.
func (rl *ResourceList) Memory() *resource.Quantity {
	return rl.Name(ResourceMemory, resource.BinarySI)
}

// Storage returns the Storage limit if specified.
func (rl *ResourceList) Storage() *resource.Quantity {
	return rl.Name(ResourceStorage, resource.BinarySI)
}

// Bandwidth returns the Bandwidth limit if specified.
func (rl *ResourceList) Bandwidth() *resource.Quantity {
	return rl.Name(ResourceBandwidth, resource.DecimalSI)
}

// Name returns the resource with name if specified, otherwise it returns a nil quantity with default format.
func (rl *ResourceList) Name(name ResourceName, defaultFormat resource.Format) *resource.Quantity {
	if val, ok := (*rl)[name]; ok {
		return &val
	}
	return &resource.Quantity{Format: defaultFormat}
}
