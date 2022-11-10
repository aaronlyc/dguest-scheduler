package framework

import (
	"dguest-scheduler/pkg/apis/scheduler/v1alpha1"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync/atomic"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"
)

var generation int64

// ActionType is an integer to represent one type of resource change.
// Different ActionTypes can be bit-wised to compose new semantics.
type ActionType int64

// Constants for ActionTypes.
const (
	Add    ActionType = 1 << iota // 1
	Delete                        // 10
	// UpdateFoodXYZ is only applicable for Food events.
	UpdateFoodAllocatable // 100
	UpdateFoodLabel       // 1000
	UpdateFoodTaint       // 10000
	UpdateFoodCondition   // 100000

	All ActionType = 1<<iota - 1 // 111111

	// Use the general Update type if you don't either know or care the specific sub-Update type to use.
	Update = UpdateFoodAllocatable | UpdateFoodLabel | UpdateFoodTaint | UpdateFoodCondition
)

// GVK is short for group/version/kind, which can uniquely represent a particular API resource.
type GVK string

// Constants for GVKs.
const (
	Dguest                GVK = "Dguest"
	Food                  GVK = "Food"
	PersistentVolume      GVK = "PersistentVolume"
	PersistentVolumeClaim GVK = "PersistentVolumeClaim"
	Service               GVK = "Service"
	StorageClass          GVK = "storage.k8s.io/StorageClass"
	WildCard              GVK = "*"
)

// ClusterEvent abstracts how a system resource's state gets changed.
// Resource represents the standard API resources such as Dguest, Food, etc.
// ActionType denotes the specific change such as Add, Update or Delete.
type ClusterEvent struct {
	Resource   GVK
	ActionType ActionType
	Label      string
}

// IsWildCard returns true if ClusterEvent follows WildCard semantics
func (ce ClusterEvent) IsWildCard() bool {
	return ce.Resource == WildCard && ce.ActionType == All
}

// QueuedDguestInfo is a Dguest wrapper with additional information related to
// the dguest's status in the scheduling queue, such as the timestamp when
// it's added to the queue.
type QueuedDguestInfo struct {
	*DguestInfo
	// The time dguest added to the scheduling queue.
	Timestamp time.Time
	// Number of schedule attempts before successfully scheduled.
	// It's used to record the # attempts metric.
	Attempts int
	// The time when the dguest is added to the queue for the first time. The dguest may be added
	// back to the queue multiple times before it's successfully scheduled.
	// It shouldn't be updated once initialized. It's used to record the e2e scheduling
	// latency for a dguest.
	InitialAttemptTimestamp time.Time
	// If a Dguest failed in a scheduling cycle, record the plugin names it failed by.
	UnschedulablePlugins sets.String
}

// DeepCopy returns a deep copy of the QueuedDguestInfo object.
func (pqi *QueuedDguestInfo) DeepCopy() *QueuedDguestInfo {
	return &QueuedDguestInfo{
		DguestInfo:              pqi.DguestInfo.DeepCopy(),
		Timestamp:               pqi.Timestamp,
		Attempts:                pqi.Attempts,
		InitialAttemptTimestamp: pqi.InitialAttemptTimestamp,
	}
}

// DguestInfo is a wrapper to a Dguest with additional pre-computed information to
// accelerate processing. This information is typically immutable (e.g., pre-processed
// inter-dguest affinity selectors).
type DguestInfo struct {
	Dguest                     *v1alpha1.Dguest
	RequiredAffinityTerms      []AffinityTerm
	RequiredAntiAffinityTerms  []AffinityTerm
	PreferredAffinityTerms     []WeightedAffinityTerm
	PreferredAntiAffinityTerms []WeightedAffinityTerm
	ParseError                 error
}

// DeepCopy returns a deep copy of the DguestInfo object.
func (pi *DguestInfo) DeepCopy() *DguestInfo {
	return &DguestInfo{
		Dguest:                     pi.Dguest.DeepCopy(),
		RequiredAffinityTerms:      pi.RequiredAffinityTerms,
		RequiredAntiAffinityTerms:  pi.RequiredAntiAffinityTerms,
		PreferredAffinityTerms:     pi.PreferredAffinityTerms,
		PreferredAntiAffinityTerms: pi.PreferredAntiAffinityTerms,
		ParseError:                 pi.ParseError,
	}
}

// Update creates a full new DguestInfo by default. And only updates the dguest when the DguestInfo
// has been instantiated and the passed dguest is the exact same one as the original dguest.
func (pi *DguestInfo) Update(dguest *v1alpha1.Dguest) {
	if dguest != nil && pi.Dguest != nil && pi.Dguest.UID == dguest.UID {
		// DguestInfo includes immutable information, and so it is safe to update the dguest in place if it is
		// the exact same dguest
		pi.Dguest = dguest
		return
	}
	//var preferredAffinityTerms []v1.WeightedDguestAffinityTerm
	//var preferredAntiAffinityTerms []v1.WeightedDguestAffinityTerm
	//if affinity := dguest.Spec.Affinity; affinity != nil {
	//	if a := affinity.DguestAffinity; a != nil {
	//		preferredAffinityTerms = a.PreferredDuringSchedulingIgnoredDuringExecution
	//	}
	//	if a := affinity.DguestAntiAffinity; a != nil {
	//		preferredAntiAffinityTerms = a.PreferredDuringSchedulingIgnoredDuringExecution
	//	}
	//}

	// Attempt to parse the affinity terms
	//var parseErrs []error
	//requiredAffinityTerms, err := getAffinityTerms(dguest, getDguestAffinityTerms(dguest.Spec.Affinity))
	//if err != nil {
	//	parseErrs = append(parseErrs, fmt.Errorf("requiredAffinityTerms: %w", err))
	//}
	//requiredAntiAffinityTerms, err := getAffinityTerms(dguest,
	//	getDguestAntiAffinityTerms(dguest.Spec.Affinity))
	//if err != nil {
	//	parseErrs = append(parseErrs, fmt.Errorf("requiredAntiAffinityTerms: %w", err))
	//}
	//weightedAffinityTerms, err := getWeightedAffinityTerms(dguest, preferredAffinityTerms)
	//if err != nil {
	//	parseErrs = append(parseErrs, fmt.Errorf("preferredAffinityTerms: %w", err))
	//}
	//weightedAntiAffinityTerms, err := getWeightedAffinityTerms(dguest, preferredAntiAffinityTerms)
	//if err != nil {
	//	parseErrs = append(parseErrs, fmt.Errorf("preferredAntiAffinityTerms: %w", err))
	//}

	pi.Dguest = dguest
	//pi.RequiredAffinityTerms = requiredAffinityTerms
	//pi.RequiredAntiAffinityTerms = requiredAntiAffinityTerms
	//pi.PreferredAffinityTerms = weightedAffinityTerms
	//pi.PreferredAntiAffinityTerms = weightedAntiAffinityTerms
	//pi.ParseError = utilerrors.NewAggregate(parseErrs)
}

// AffinityTerm is a processed version of v1alpha1.DguestAffinityTerm.
type AffinityTerm struct {
	Namespaces        sets.String
	Selector          labels.Selector
	TopologyKey       string
	NamespaceSelector labels.Selector
}

// Matches returns true if the dguest matches the label selector and namespaces or namespace selector.
func (at *AffinityTerm) Matches(dguest *v1alpha1.Dguest, nsLabels labels.Set) bool {
	if at.Namespaces.Has(dguest.Namespace) || at.NamespaceSelector.Matches(nsLabels) {
		return at.Selector.Matches(labels.Set(dguest.Labels))
	}
	return false
}

// WeightedAffinityTerm is a "processed" representation of v1.WeightedAffinityTerm.
type WeightedAffinityTerm struct {
	AffinityTerm
	Weight int32
}

// Diagnosis records the details to diagnose a scheduling failure.
type Diagnosis struct {
	FoodToStatusMap      FoodToStatusMap
	UnschedulablePlugins sets.String
	// PostFilterMsg records the messages returned from PostFilterPlugins.
	PostFilterMsg string
}

// FitError describes a fit error of a dguest.
type FitError struct {
	Dguest      *v1alpha1.Dguest
	NumAllFoods int
	Diagnosis   Diagnosis
}

// NoFoodAvailableMsg is used to format message when no foods available.
const NoFoodAvailableMsg = "0/%v foods are available"

// Error returns detailed information of why the dguest failed to fit on each food
func (f *FitError) Error() string {
	reasons := make(map[string]int)
	for _, status := range f.Diagnosis.FoodToStatusMap {
		for _, reason := range status.Reasons() {
			reasons[reason]++
		}
	}

	sortReasonsHistogram := func() []string {
		var reasonStrings []string
		for k, v := range reasons {
			reasonStrings = append(reasonStrings, fmt.Sprintf("%v %v", v, k))
		}
		sort.Strings(reasonStrings)
		return reasonStrings
	}
	reasonMsg := fmt.Sprintf(NoFoodAvailableMsg+": %v.", f.NumAllFoods, strings.Join(sortReasonsHistogram(), ", "))
	postFilterMsg := f.Diagnosis.PostFilterMsg
	if postFilterMsg != "" {
		reasonMsg += " " + postFilterMsg
	}
	return reasonMsg
}

//func newAffinityTerm(dguest *v1alpha1.Dguest, term *v1alpha1.DguestAffinityTerm) (*AffinityTerm, error) {
//	selector, err := metav1.LabelSelectorAsSelector(term.LabelSelector)
//	if err != nil {
//		return nil, err
//	}
//
//	namespaces := getNamespacesFromDguestAffinityTerm(dguest, term)
//	nsSelector, err := metav1.LabelSelectorAsSelector(term.NamespaceSelector)
//	if err != nil {
//		return nil, err
//	}
//
//	return &AffinityTerm{Namespaces: namespaces, Selector: selector, TopologyKey: term.TopologyKey, NamespaceSelector: nsSelector}, nil
//}

// getAffinityTerms receives a Dguest and affinity terms and returns the namespaces and
// selectors of the terms.
//func getAffinityTerms(dguest *v1alpha1.Dguest, v1Terms []v1alpha1.DguestAffinityTerm) ([]AffinityTerm, error) {
//	if v1Terms == nil {
//		return nil, nil
//	}
//
//	var terms []AffinityTerm
//	for i := range v1Terms {
//		t, err := newAffinityTerm(dguest, &v1Terms[i])
//		if err != nil {
//			// We get here if the label selector failed to process
//			return nil, err
//		}
//		terms = append(terms, *t)
//	}
//	return terms, nil
//}

// getWeightedAffinityTerms returns the list of processed affinity terms.
//func getWeightedAffinityTerms(dguest *v1alpha1.Dguest, v1Terms []v1.WeightedDguestAffinityTerm) ([]WeightedAffinityTerm, error) {
//	if v1Terms == nil {
//		return nil, nil
//	}
//
//	var terms []WeightedAffinityTerm
//	for i := range v1Terms {
//		t, err := newAffinityTerm(dguest, &v1Terms[i].DguestAffinityTerm)
//		if err != nil {
//			// We get here if the label selector failed to process
//			return nil, err
//		}
//		terms = append(terms, WeightedAffinityTerm{AffinityTerm: *t, Weight: v1Terms[i].Weight})
//	}
//	return terms, nil
//}

// NewDguestInfo returns a new DguestInfo.
func NewDguestInfo(dguest *v1alpha1.Dguest) *DguestInfo {
	pInfo := &DguestInfo{}
	pInfo.Update(dguest)
	return pInfo
}

//func getDguestAffinityTerms(affinity *v1.Affinity) (terms []v1alpha1.DguestAffinityTerm) {
//	if affinity != nil && affinity.DguestAffinity != nil {
//		if len(affinity.DguestAffinity.RequiredDuringSchedulingIgnoredDuringExecution) != 0 {
//			terms = affinity.DguestAffinity.RequiredDuringSchedulingIgnoredDuringExecution
//		}
//		// TODO: Uncomment this block when implement RequiredDuringSchedulingRequiredDuringExecution.
//		//if len(affinity.DguestAffinity.RequiredDuringSchedulingRequiredDuringExecution) != 0 {
//		//	terms = append(terms, affinity.DguestAffinity.RequiredDuringSchedulingRequiredDuringExecution...)
//		//}
//	}
//	return terms
//}

//func getDguestAntiAffinityTerms(affinity *v1.Affinity) (terms []v1alpha1.DguestAffinityTerm) {
//	if affinity != nil && affinity.DguestAntiAffinity != nil {
//		if len(affinity.DguestAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution) != 0 {
//			terms = affinity.DguestAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution
//		}
//		// TODO: Uncomment this block when implement RequiredDuringSchedulingRequiredDuringExecution.
//		//if len(affinity.DguestAntiAffinity.RequiredDuringSchedulingRequiredDuringExecution) != 0 {
//		//	terms = append(terms, affinity.DguestAntiAffinity.RequiredDuringSchedulingRequiredDuringExecution...)
//		//}
//	}
//	return terms
//}

// returns a set of names according to the namespaces indicated in dguestAffinityTerm.
// If namespaces is empty it considers the given dguest's namespace.
//func getNamespacesFromDguestAffinityTerm(dguest *v1alpha1.Dguest, dguestAffinityTerm *v1alpha1.DguestAffinityTerm) sets.String {
//	names := sets.String{}
//	if len(dguestAffinityTerm.Namespaces) == 0 && dguestAffinityTerm.NamespaceSelector == nil {
//		names.Insert(dguest.Namespace)
//	} else {
//		names.Insert(dguestAffinityTerm.Namespaces...)
//	}
//	return names
//}

// ImageStateSummary provides summarized information about the state of an image.
type ImageStateSummary struct {
	// Size of the image
	Size int64
	// Used to track how many foods have this image
	NumFoods int
}

// FoodInfo is food level aggregated information.
type FoodInfo struct {
	// Overall food information.
	food *v1alpha1.Food

	// Dguests running on the food.
	Dguests []*DguestInfo

	// The subset of dguests with affinity.
	DguestsWithAffinity []*DguestInfo

	// The subset of dguests with required anti-affinity.
	DguestsWithRequiredAntiAffinity []*DguestInfo

	// Ports allocated on the food.
	UsedPorts HostPortInfo

	// Total requested resources of all dguests on this food. This includes assumed
	// dguests, which scheduler has sent for binding, but may not be scheduled yet.
	Requested *Resource
	// Total requested resources of all dguests on this food with a minimum value
	// applied to each container's CPU and memory requests. This does not reflect
	// the actual resource requests for this food, but is used to avoid scheduling
	// many zero-request dguests onto one food.
	NonZeroRequested *Resource
	// We store allocatedResources (which is Food.Status.Allocatable.*) explicitly
	// as int64, to avoid conversions and accessing map.
	Allocatable *Resource

	// ImageStates holds the entry of an image if and only if this image is on the food. The entry can be used for
	// checking an image's existence and advanced usage (e.g., image locality scheduling policy) based on the image
	// state information.
	ImageStates map[string]*ImageStateSummary

	// PVCRefCounts contains a mapping of PVC names to the number of dguests on the food using it.
	// Keys are in the format "namespace/name".
	PVCRefCounts map[string]int

	// Whenever FoodInfo changes, generation is bumped.
	// This is used to avoid cloning it if the object didn't change.
	Generation int64
}

// nextGeneration: Let's make sure history never forgets the name...
// Increments the generation number monotonically ensuring that generation numbers never collide.
// Collision of the generation numbers would be particularly problematic if a food was deleted and
// added back with the same name. See issue#63262.
func nextGeneration() int64 {
	return atomic.AddInt64(&generation, 1)
}

// Resource is a collection of compute resource.
type Resource struct {
	MilliCPU  int64
	Memory    int64
	Bandwidth int64
	// We store allowedDguestNumber (which is Food.Status.Allocatable.Dguests().Value())
	// explicitly as int, to avoid conversions and improve performance.
	AllowedDguestNumber int
	// ScalarResources
	ScalarResources map[v1.ResourceName]int64
}

// NewResource creates a Resource from ResourceList
func NewResource(rl v1alpha1.ResourceList) *Resource {
	r := &Resource{}
	r.Add(rl)
	return r
}

// Add adds ResourceList into Resource.
func (r *Resource) Add(rl v1alpha1.ResourceList) {
	if r == nil {
		return
	}

	for rName, rQuant := range rl {
		switch rName {
		case v1alpha1.ResourceCPU:
			r.MilliCPU += rQuant.MilliValue()
		case v1alpha1.ResourceMemory:
			r.Memory += rQuant.Value()
		//case v1.ResourceDguests:
		//	r.AllowedDguestNumber += int(rQuant.Value())
		case v1alpha1.ResourceBandwidth:
			r.Bandwidth += rQuant.Value()
		default:
			//if schedutil.IsScalarResourceName(rName) {
			//	r.AddScalar(rName, rQuant.Value())
			//}
		}
	}
}

// Clone returns a copy of this resource.
func (r *Resource) Clone() *Resource {
	res := &Resource{
		MilliCPU:            r.MilliCPU,
		Memory:              r.Memory,
		AllowedDguestNumber: r.AllowedDguestNumber,
		Bandwidth:           r.Bandwidth,
	}
	if r.ScalarResources != nil {
		res.ScalarResources = make(map[v1.ResourceName]int64)
		for k, v := range r.ScalarResources {
			res.ScalarResources[k] = v
		}
	}
	return res
}

// AddScalar adds a resource by a scalar value of this resource.
func (r *Resource) AddScalar(name v1.ResourceName, quantity int64) {
	r.SetScalar(name, r.ScalarResources[name]+quantity)
}

// SetScalar sets a resource by a scalar value of this resource.
func (r *Resource) SetScalar(name v1.ResourceName, quantity int64) {
	// Lazily allocate scalar resource map.
	if r.ScalarResources == nil {
		r.ScalarResources = map[v1.ResourceName]int64{}
	}
	r.ScalarResources[name] = quantity
}

// SetMaxResource compares with ResourceList and takes max value for each Resource.
func (r *Resource) SetMaxResource(rl v1alpha1.ResourceList) {
	if r == nil {
		return
	}

	for rName, rQuantity := range rl {
		switch rName {
		case v1alpha1.ResourceMemory:
			r.Memory = max(r.Memory, rQuantity.Value())
		case v1alpha1.ResourceCPU:
			r.MilliCPU = max(r.MilliCPU, rQuantity.MilliValue())
		case v1alpha1.ResourceBandwidth:
			r.Bandwidth = max(r.Bandwidth, rQuantity.Value())
		default:
			//if schedutil.IsScalarResourceName(rName) {
			//	r.SetScalar(rName, max(r.ScalarResources[rName], rQuantity.Value()))
			//}
		}
	}
}

// NewFoodInfo returns a ready to use empty FoodInfo object.
// If any dguests are given in arguments, their information will be aggregated in
// the returned object.
func NewFoodInfo(dguests ...*v1alpha1.Dguest) *FoodInfo {
	ni := &FoodInfo{
		Requested:        &Resource{},
		NonZeroRequested: &Resource{},
		Allocatable:      &Resource{},
		Generation:       nextGeneration(),
		UsedPorts:        make(HostPortInfo),
		ImageStates:      make(map[string]*ImageStateSummary),
		PVCRefCounts:     make(map[string]int),
	}
	for _, dguest := range dguests {
		ni.AddDguest(dguest)
	}
	return ni
}

// Food returns overall information about this food.
func (n *FoodInfo) Food() *v1alpha1.Food {
	if n == nil {
		return nil
	}
	return n.food
}

// Clone returns a copy of this food.
func (n *FoodInfo) Clone() *FoodInfo {
	clone := &FoodInfo{
		food:             n.food,
		Requested:        n.Requested.Clone(),
		NonZeroRequested: n.NonZeroRequested.Clone(),
		Allocatable:      n.Allocatable.Clone(),
		UsedPorts:        make(HostPortInfo),
		ImageStates:      n.ImageStates,
		PVCRefCounts:     make(map[string]int),
		Generation:       n.Generation,
	}
	if len(n.Dguests) > 0 {
		clone.Dguests = append([]*DguestInfo(nil), n.Dguests...)
	}
	if len(n.UsedPorts) > 0 {
		// HostPortInfo is a map-in-map struct
		// make sure it's deep copied
		for ip, portMap := range n.UsedPorts {
			clone.UsedPorts[ip] = make(map[ProtocolPort]struct{})
			for protocolPort, v := range portMap {
				clone.UsedPorts[ip][protocolPort] = v
			}
		}
	}
	if len(n.DguestsWithAffinity) > 0 {
		clone.DguestsWithAffinity = append([]*DguestInfo(nil), n.DguestsWithAffinity...)
	}
	if len(n.DguestsWithRequiredAntiAffinity) > 0 {
		clone.DguestsWithRequiredAntiAffinity = append([]*DguestInfo(nil), n.DguestsWithRequiredAntiAffinity...)
	}
	for key, value := range n.PVCRefCounts {
		clone.PVCRefCounts[key] = value
	}
	return clone
}

// String returns representation of human readable format of this FoodInfo.
func (n *FoodInfo) String() string {
	dguestKeys := make([]string, len(n.Dguests))
	for i, p := range n.Dguests {
		dguestKeys[i] = p.Dguest.Name
	}
	return fmt.Sprintf("&FoodInfo{Dguests:%v, RequestedResource:%#v, NonZeroRequest: %#v, UsedPort: %#v, AllocatableResource:%#v}",
		dguestKeys, n.Requested, n.NonZeroRequested, n.UsedPorts, n.Allocatable)
}

// AddDguestInfo adds dguest information to this FoodInfo.
// Consider using this instead of AddDguest if a DguestInfo is already computed.
func (n *FoodInfo) AddDguestInfo(dguestInfo *DguestInfo) {
	res, non0CPU, non0Mem := calculateResource(dguestInfo.Dguest)
	n.Requested.MilliCPU += res.MilliCPU
	n.Requested.Memory += res.Memory
	n.Requested.Bandwidth += res.Bandwidth
	if n.Requested.ScalarResources == nil && len(res.ScalarResources) > 0 {
		n.Requested.ScalarResources = map[v1.ResourceName]int64{}
	}
	for rName, rQuant := range res.ScalarResources {
		n.Requested.ScalarResources[rName] += rQuant
	}
	n.NonZeroRequested.MilliCPU += non0CPU
	n.NonZeroRequested.Memory += non0Mem
	n.Dguests = append(n.Dguests, dguestInfo)
	//if dguestWithAffinity(dguestInfo.Dguest) {
	//	n.DguestsWithAffinity = append(n.DguestsWithAffinity, dguestInfo)
	//}
	//if dguestWithRequiredAntiAffinity(dguestInfo.Dguest) {
	//	n.DguestsWithRequiredAntiAffinity = append(n.DguestsWithRequiredAntiAffinity, dguestInfo)
	//}

	// Consume ports when dguests added.
	//n.updateUsedPorts(dguestInfo.Dguest, true)
	//n.updatePVCRefCounts(dguestInfo.Dguest, true)

	n.Generation = nextGeneration()
}

// AddDguest is a wrapper around AddDguestInfo.
func (n *FoodInfo) AddDguest(dguest *v1alpha1.Dguest) {
	n.AddDguestInfo(NewDguestInfo(dguest))
}

//func dguestWithAffinity(p *v1alpha1.Dguest) bool {
//	affinity := p.Spec.Affinity
//	return affinity != nil && (affinity.DguestAffinity != nil || affinity.DguestAntiAffinity != nil)
//}

//func dguestWithRequiredAntiAffinity(p *v1alpha1.Dguest) bool {
//	affinity := p.Spec.Affinity
//	return affinity != nil && affinity.DguestAntiAffinity != nil &&
//		len(affinity.DguestAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution) != 0
//}

func removeFromSlice(s []*DguestInfo, k string) []*DguestInfo {
	for i := range s {
		k2, err := GetDguestKey(s[i].Dguest)
		if err != nil {
			klog.ErrorS(err, "Cannot get dguest key", "dguest", klog.KObj(s[i].Dguest))
			continue
		}
		if k == k2 {
			// delete the element
			s[i] = s[len(s)-1]
			s = s[:len(s)-1]
			break
		}
	}
	return s
}

// RemoveDguest subtracts dguest information from this FoodInfo.
func (n *FoodInfo) RemoveDguest(dguest *v1alpha1.Dguest) error {
	k, err := GetDguestKey(dguest)
	if err != nil {
		return err
	}
	//if dguestWithAffinity(dguest) {
	//	n.DguestsWithAffinity = removeFromSlice(n.DguestsWithAffinity, k)
	//}
	//if dguestWithRequiredAntiAffinity(dguest) {
	//	n.DguestsWithRequiredAntiAffinity = removeFromSlice(n.DguestsWithRequiredAntiAffinity, k)
	//}

	for i := range n.Dguests {
		k2, err := GetDguestKey(n.Dguests[i].Dguest)
		if err != nil {
			klog.ErrorS(err, "Cannot get dguest key", "dguest", klog.KObj(n.Dguests[i].Dguest))
			continue
		}
		if k == k2 {
			// delete the element
			n.Dguests[i] = n.Dguests[len(n.Dguests)-1]
			n.Dguests = n.Dguests[:len(n.Dguests)-1]
			// reduce the resource data
			res, non0CPU, non0Mem := calculateResource(dguest)

			n.Requested.MilliCPU -= res.MilliCPU
			n.Requested.Memory -= res.Memory
			n.Requested.Bandwidth -= res.Bandwidth
			if len(res.ScalarResources) > 0 && n.Requested.ScalarResources == nil {
				n.Requested.ScalarResources = map[v1.ResourceName]int64{}
			}
			for rName, rQuant := range res.ScalarResources {
				n.Requested.ScalarResources[rName] -= rQuant
			}
			n.NonZeroRequested.MilliCPU -= non0CPU
			n.NonZeroRequested.Memory -= non0Mem

			// Release ports when remove Dguests.
			//n.updateUsedPorts(dguest, false)
			//n.updatePVCRefCounts(dguest, false)

			n.Generation = nextGeneration()
			n.resetSlicesIfEmpty()
			return nil
		}
	}
	return fmt.Errorf("no corresponding dguest %s in dguests of food %s", dguest.Name, n.food.Name)
}

// resets the slices to nil so that we can do DeepEqual in unit tests.
func (n *FoodInfo) resetSlicesIfEmpty() {
	if len(n.DguestsWithAffinity) == 0 {
		n.DguestsWithAffinity = nil
	}
	if len(n.DguestsWithRequiredAntiAffinity) == 0 {
		n.DguestsWithRequiredAntiAffinity = nil
	}
	if len(n.Dguests) == 0 {
		n.Dguests = nil
	}
}

func max(a, b int64) int64 {
	if a >= b {
		return a
	}
	return b
}

// resourceRequest = max(sum(dguestSpec.Containers), dguestSpec.InitContainers) + overHead
func calculateResource(dguest *v1alpha1.Dguest) (res Resource, non0CPU int64, non0Mem int64) {
	resPtr := &res
	//for _, c := range dguest.Spec.Containers {
	//	resPtr.Add(c.Resources.Requests)
	//	non0CPUReq, non0MemReq := schedutil.GetNonzeroRequests(&c.Resources.Requests)
	//	non0CPU += non0CPUReq
	//	non0Mem += non0MemReq
	//	// No non-zero resources for GPUs or opaque resources.
	//}
	//
	//for _, ic := range dguest.Spec.InitContainers {
	//	resPtr.SetMaxResource(ic.Resources.Requests)
	//	non0CPUReq, non0MemReq := schedutil.GetNonzeroRequests(&ic.Resources.Requests)
	//	non0CPU = max(non0CPU, non0CPUReq)
	//	non0Mem = max(non0Mem, non0MemReq)
	//}

	// If Overhead is being utilized, add to the total requests for the dguest
	if dguest.Spec.Overhead != nil {
		resPtr.Add(dguest.Spec.Overhead)
		if _, found := dguest.Spec.Overhead[v1alpha1.ResourceCPU]; found {
			non0CPU += dguest.Spec.Overhead.Cpu().MilliValue()
		}

		if _, found := dguest.Spec.Overhead[v1alpha1.ResourceMemory]; found {
			non0Mem += dguest.Spec.Overhead.Memory().Value()
		}
	}

	return
}

// updateUsedPorts updates the UsedPorts of FoodInfo.
//func (n *FoodInfo) updateUsedPorts(dguest *v1alpha1.Dguest, add bool) {
//for _, container := range dguest.Spec.Containers {
//	for _, dguestPort := range container.Ports {
//		if add {
//			n.UsedPorts.Add(dguestPort.HostIP, string(dguestPort.Protocol), dguestPort.HostPort)
//		} else {
//			n.UsedPorts.Remove(dguestPort.HostIP, string(dguestPort.Protocol), dguestPort.HostPort)
//		}
//	}
//}
//}

// updatePVCRefCounts updates the PVCRefCounts of FoodInfo.
//func (n *FoodInfo) updatePVCRefCounts(dguest *v1alpha1.Dguest, add bool) {
//	for _, v := range dguest.Spec.Volumes {
//		if v.PersistentVolumeClaim == nil {
//			continue
//		}
//
//		key := GetNamespacedName(dguest.Namespace, v.PersistentVolumeClaim.ClaimName)
//		if add {
//			n.PVCRefCounts[key] += 1
//		} else {
//			n.PVCRefCounts[key] -= 1
//			if n.PVCRefCounts[key] <= 0 {
//				delete(n.PVCRefCounts, key)
//			}
//		}
//	}
//}

// SetFood sets the overall food information.
func (n *FoodInfo) SetFood(food *v1alpha1.Food) {
	n.food = food
	n.Allocatable = NewResource(food.Status.Allocatable)
	n.Generation = nextGeneration()
}

// RemoveFood removes the food object, leaving all other tracking information.
func (n *FoodInfo) RemoveFood() {
	n.food = nil
	n.Generation = nextGeneration()
}

// GetDguestKey returns the string key of a dguest.
func GetDguestKey(dguest *v1alpha1.Dguest) (string, error) {
	uid := string(dguest.UID)
	if len(uid) == 0 {
		return "", errors.New("cannot get cache key for dguest with empty UID")
	}
	return uid, nil
}

// GetNamespacedName returns the string format of a namespaced resource name.
func GetNamespacedName(namespace, name string) string {
	return fmt.Sprintf("%s/%s", namespace, name)
}

// DefaultBindAllHostIP defines the default ip address used to bind to all host.
const DefaultBindAllHostIP = "0.0.0.0"

// ProtocolPort represents a protocol port pair, e.g. tcp:80.
type ProtocolPort struct {
	Protocol string
	Port     int32
}

// NewProtocolPort creates a ProtocolPort instance.
func NewProtocolPort(protocol string, port int32) *ProtocolPort {
	pp := &ProtocolPort{
		Protocol: protocol,
		Port:     port,
	}

	if len(pp.Protocol) == 0 {
		pp.Protocol = string(v1.ProtocolTCP)
	}

	return pp
}

// HostPortInfo stores mapping from ip to a set of ProtocolPort
type HostPortInfo map[string]map[ProtocolPort]struct{}

// Add adds (ip, protocol, port) to HostPortInfo
func (h HostPortInfo) Add(ip, protocol string, port int32) {
	if port <= 0 {
		return
	}

	h.sanitize(&ip, &protocol)

	pp := NewProtocolPort(protocol, port)
	if _, ok := h[ip]; !ok {
		h[ip] = map[ProtocolPort]struct{}{
			*pp: {},
		}
		return
	}

	h[ip][*pp] = struct{}{}
}

// Remove removes (ip, protocol, port) from HostPortInfo
func (h HostPortInfo) Remove(ip, protocol string, port int32) {
	if port <= 0 {
		return
	}

	h.sanitize(&ip, &protocol)

	pp := NewProtocolPort(protocol, port)
	if m, ok := h[ip]; ok {
		delete(m, *pp)
		if len(h[ip]) == 0 {
			delete(h, ip)
		}
	}
}

// Len returns the total number of (ip, protocol, port) tuple in HostPortInfo
func (h HostPortInfo) Len() int {
	length := 0
	for _, m := range h {
		length += len(m)
	}
	return length
}

// CheckConflict checks if the input (ip, protocol, port) conflicts with the existing
// ones in HostPortInfo.
func (h HostPortInfo) CheckConflict(ip, protocol string, port int32) bool {
	if port <= 0 {
		return false
	}

	h.sanitize(&ip, &protocol)

	pp := NewProtocolPort(protocol, port)

	// If ip is 0.0.0.0 check all IP's (protocol, port) pair
	if ip == DefaultBindAllHostIP {
		for _, m := range h {
			if _, ok := m[*pp]; ok {
				return true
			}
		}
		return false
	}

	// If ip isn't 0.0.0.0, only check IP and 0.0.0.0's (protocol, port) pair
	for _, key := range []string{DefaultBindAllHostIP, ip} {
		if m, ok := h[key]; ok {
			if _, ok2 := m[*pp]; ok2 {
				return true
			}
		}
	}

	return false
}

// sanitize the parameters
func (h HostPortInfo) sanitize(ip, protocol *string) {
	if len(*ip) == 0 {
		*ip = DefaultBindAllHostIP
	}
	if len(*protocol) == 0 {
		*protocol = string(v1.ProtocolTCP)
	}
}
