package framework

// FoodInfoLister interface represents anything that can list/get FoodInfo objects from food name.
type FoodInfoLister interface {
	// List returns the list of FoodInfos.
	List() []*FoodInfo
	// HaveDguestsWithAffinityList returns the list of FoodInfos of foods with dguests with affinity terms.
	HaveDguestsWithAffinityList() ([]*FoodInfo, error)
	// HaveDguestsWithRequiredAntiAffinityList returns the list of FoodInfos of foods with dguests with required anti-affinity terms.
	HaveDguestsWithRequiredAntiAffinityList() ([]*FoodInfo, error)
	// Get returns the FoodInfo of the given food name.
	Get(foodName string) (*FoodInfo, error)
}

// StorageInfoLister interface represents anything that handles storage-related operations and resources.
type StorageInfoLister interface {
	// IsPVCUsedByDguests returns true/false on whether the PVC is used by one or more scheduled dguests,
	// keyed in the format "namespace/name".
	IsPVCUsedByDguests(key string) bool
}

// SharedLister groups scheduler-specific listers.
type SharedLister interface {
	FoodInfos() FoodInfoLister
	// StorageInfos() StorageInfoLister
}
