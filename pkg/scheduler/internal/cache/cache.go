package cache

import (
	"dguest-scheduler/pkg/apis/scheduler/v1alpha1"
	"fmt"
	"os"
	"sync"
	"time"

	apidguest "dguest-scheduler/pkg/api/dguest"
	"dguest-scheduler/pkg/scheduler/framework"
	"dguest-scheduler/pkg/scheduler/metrics"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
)

var (
	cleanAssumedPeriod = 1 * time.Second
)

// New returns a Cache implementation.
// It automatically starts a go routine that manages expiration of assumed dguests.
// "ttl" is how long the assumed dguest will get expired.
// "stop" is the channel that would close the background goroutine.
func New(ttl time.Duration, stop <-chan struct{}) Cache {
	cache := newCache(ttl, cleanAssumedPeriod, stop)
	cache.run()
	return cache
}

// foodInfoListItem holds a FoodInfo pointer and acts as an item in a doubly
// linked list. When a FoodInfo is updated, it goes to the head of the list.
// The items closer to the head are the most recently updated items.
type foodInfoListItem struct {
	info *framework.FoodInfo
	next *foodInfoListItem
	prev *foodInfoListItem
}

type cacheImpl struct {
	stop   <-chan struct{}
	ttl    time.Duration
	period time.Duration

	// This mutex guards all fields within this cache struct.
	mu sync.RWMutex
	// a set of assumed dguest keys.
	// The key could further be used to get an entry in dguestStates.
	assumedDguests sets.String
	// a map from dguest key to dguestState.
	dguestStates map[string]*dguestState
	foods        map[string]*foodInfoListItem
	// headFood points to the most recently updated FoodInfo in "foods". It is the
	// head of the linked list.
	headFood *foodInfoListItem
	foodTree *foodTree
	// A map from image name to its imageState.
	// imageStates map[string]*imageState
}

type dguestState struct {
	dguest *v1alpha1.Dguest
	// Used by assumedDguest to determinate expiration.
	// If deadline is nil, assumedDguest will never expire.
	deadline *time.Time
	// Used to block cache from expiring assumedDguest if binding still runs
	bindingFinished bool
}

// type imageState struct {
// 	// Size of the image
// 	size int64
// 	// A set of food names for foods having this image present
// 	foods sets.String
// }

// createImageStateSummary returns a summarizing snapshot of the given image's state.
// func (cache *cacheImpl) createImageStateSummary(state *imageState) *framework.ImageStateSummary {
// 	return &framework.ImageStateSummary{
// 		Size:     state.size,
// 		NumFoods: len(state.foods),
// 	}
// }

func newCache(ttl, period time.Duration, stop <-chan struct{}) *cacheImpl {
	return &cacheImpl{
		ttl:    ttl,
		period: period,
		stop:   stop,

		foods:          make(map[string]*foodInfoListItem),
		foodTree:       newFoodTree(nil),
		assumedDguests: make(sets.String),
		dguestStates:   make(map[string]*dguestState),
		// imageStates:    make(map[string]*imageState),
	}
}

// newFoodInfoListItem initializes a new foodInfoListItem.
func newFoodInfoListItem(ni *framework.FoodInfo) *foodInfoListItem {
	return &foodInfoListItem{
		info: ni,
	}
}

// moveFoodInfoToHead moves a FoodInfo to the head of "cache.foods" doubly
// linked list. The head is the most recently updated FoodInfo.
// We assume cache lock is already acquired.
func (cache *cacheImpl) moveFoodInfoToHead(name string) {
	ni, ok := cache.foods[name]
	if !ok {
		klog.ErrorS(nil, "No food info with given name found in the cache", "food", klog.KRef("", name))
		return
	}
	// if the food info list item is already at the head, we are done.
	if ni == cache.headFood {
		return
	}

	if ni.prev != nil {
		ni.prev.next = ni.next
	}
	if ni.next != nil {
		ni.next.prev = ni.prev
	}
	if cache.headFood != nil {
		cache.headFood.prev = ni
	}
	ni.next = cache.headFood
	ni.prev = nil
	cache.headFood = ni
}

// removeFoodInfoFromList removes a FoodInfo from the "cache.foods" doubly
// linked list.
// We assume cache lock is already acquired.
func (cache *cacheImpl) removeFoodInfoFromList(name string) {
	ni, ok := cache.foods[name]
	if !ok {
		klog.ErrorS(nil, "No food info with given name found in the cache", "food", klog.KRef("", name))
		return
	}

	if ni.prev != nil {
		ni.prev.next = ni.next
	}
	if ni.next != nil {
		ni.next.prev = ni.prev
	}
	// if the removed item was at the head, we must update the head.
	if ni == cache.headFood {
		cache.headFood = ni.next
	}
	delete(cache.foods, name)
}

// Dump produces a dump of the current scheduler cache. This is used for
// debugging purposes only and shouldn't be confused with UpdateSnapshot
// function.
// This method is expensive, and should be only used in non-critical path.
func (cache *cacheImpl) Dump() *Dump {
	cache.mu.RLock()
	defer cache.mu.RUnlock()

	foods := make(map[string]*framework.FoodInfo, len(cache.foods))
	for k, v := range cache.foods {
		foods[k] = v.info.Clone()
	}

	return &Dump{
		Foods:          foods,
		AssumedDguests: cache.assumedDguests.Union(nil),
	}
}

// UpdateSnapshot takes a snapshot of cached FoodInfo map. This is called at
// beginning of every scheduling cycle.
// The snapshot only includes Foods that are not deleted at the time this function is called.
// foodInfo.Food() is guaranteed to be not nil for all the foods in the snapshot.
// This function tracks generation number of FoodInfo and updates only the
// entries of an existing snapshot that have changed after the snapshot was taken.
func (cache *cacheImpl) UpdateSnapshot(foodSnapshot *Snapshot, cuisine string) error {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	// Get the last generation of the snapshot.
	snapshotGeneration := foodSnapshot.generation

	// Start from the head of the FoodInfo doubly linked list and update snapshot
	// of FoodInfos updated after the last snapshot.
	newSnapshotFoodInfo := make(map[string][]*framework.FoodInfo)
	for cachefood := cache.headFood; cachefood != nil; cachefood = cachefood.next {
		if cachefood.info.Generation <= snapshotGeneration {
			// all the foods are updated before the existing snapshot. We are done.
			break
		}
		if np := cachefood.info.Food(); np != nil {

			key := apidguest.FoodcuisineKey(np)
			_, ok := newSnapshotFoodInfo[key]
			if !ok {
				newSnapshotFoodInfo[key] = []*framework.FoodInfo{}
			}
			clone := cachefood.info.Clone()
			newSnapshotFoodInfo[key] = append(newSnapshotFoodInfo[key], clone)
		}
	}
	foodSnapshot.foodInfoMap = newSnapshotFoodInfo
	// Update the snapshot generation with the latest FoodInfo generation.
	if cache.headFood != nil {
		foodSnapshot.generation = cache.headFood.info.Generation
	}

	// Comparing to dguests in foodTree.
	// Deleted foods get removed from the tree, but they might remain in the foods map
	// if they still have non-deleted Dguests.
	if foodSnapshot.NumFoods(cuisine) > cache.foodTree.foodCount(cuisine) {
		cache.removeDeletedFoodsFromSnapshot(foodSnapshot, foodSnapshot.NumFoods(cuisine)-cache.foodTree.foodCount(cuisine))
	}

	// if len(foodSnapshot.foodInfoList) != cache.foodTree.numFoods {
	// 	errMsg := fmt.Sprintf("snapshot state is not consistent, length of FoodInfoList=%v not equal to length of foods in tree=%v "+
	// 		", length of FoodInfoMap=%v, length of foods in cache=%v"+
	// 		", trying to recover",
	// 		len(foodSnapshot.foodInfoList), cache.foodTree.numFoods,
	// 		len(foodSnapshot.foodInfoMap), len(cache.foods))
	// 	klog.ErrorS(nil, errMsg)
	// 	// We will try to recover by re-creating the lists for the next scheduling cycle, but still return an
	// 	// error to surface the problem, the error will likely cause a failure to the current scheduling cycle.
	// 	// cache.updateFoodInfoSnapshotList(foodSnapshot, true)
	// 	return fmt.Errorf(errMsg)
	// }

	return nil
}

// If certain foods were deleted after the last snapshot was taken, we should remove them from the snapshot.
func (cache *cacheImpl) removeDeletedFoodsFromSnapshot(snapshot *Snapshot, toDelete int) {
	for name := range snapshot.foodInfoMap {
		if toDelete <= 0 {
			break
		}
		if n, ok := cache.foods[name]; !ok || n.info.Food() == nil {
			delete(snapshot.foodInfoMap, name)
			toDelete--
		}
	}
}

// FoodCount returns the number of foods in the cache.
// DO NOT use outside of tests.
func (cache *cacheImpl) FoodCount() int {
	cache.mu.RLock()
	defer cache.mu.RUnlock()
	return len(cache.foods)
}

// DguestCount returns the number of dguests in the cache (including those from deleted foods).
// DO NOT use outside of tests.
func (cache *cacheImpl) DguestCount() (int, error) {
	cache.mu.RLock()
	defer cache.mu.RUnlock()
	// dguestFilter is expected to return true for most or all of the dguests. We
	// can avoid expensive array growth without wasting too much memory by
	// pre-allocating capacity.
	count := 0
	for _, n := range cache.foods {
		count += len(n.info.Dguests)
	}
	return count, nil
}

func (cache *cacheImpl) AssumeDguest(dguest *v1alpha1.Dguest) error {
	key, err := framework.GetDguestKey(dguest)
	if err != nil {
		return err
	}

	cache.mu.Lock()
	defer cache.mu.Unlock()
	if _, ok := cache.dguestStates[key]; ok {
		return fmt.Errorf("dguest %v is in the cache, so can't be assumed", key)
	}

	return cache.addDguest(dguest, true)
}

func (cache *cacheImpl) FinishBinding(dguest *v1alpha1.Dguest) error {
	return cache.finishBinding(dguest, time.Now())
}

// finishBinding exists to make tests deterministic by injecting now as an argument
func (cache *cacheImpl) finishBinding(dguest *v1alpha1.Dguest, now time.Time) error {
	key, err := framework.GetDguestKey(dguest)
	if err != nil {
		return err
	}

	cache.mu.RLock()
	defer cache.mu.RUnlock()

	klog.V(5).InfoS("Finished binding for dguest, can be expired", "dguest", klog.KObj(dguest))
	currState, ok := cache.dguestStates[key]
	if ok && cache.assumedDguests.Has(key) {
		if cache.ttl == time.Duration(0) {
			currState.deadline = nil
		} else {
			dl := now.Add(cache.ttl)
			currState.deadline = &dl
		}
		currState.bindingFinished = true
	}
	return nil
}

func (cache *cacheImpl) ForgetDguest(dguest *v1alpha1.Dguest) error {
	key, err := framework.GetDguestKey(dguest)
	if err != nil {
		return err
	}

	cache.mu.Lock()
	defer cache.mu.Unlock()

	currState, ok := cache.dguestStates[key]

	if ok && currState.dguest.Spec.FoodNamespacedName != dguest.Spec.FoodNamespacedName {
		return fmt.Errorf("dguest %v was assumed on %v but assigned to %v", key, dguest.Spec.FoodNamespacedName, currState.dguest.Spec.FoodNamespacedName)
	}

	// Only assumed dguest can be forgotten.
	if ok && cache.assumedDguests.Has(key) {
		return cache.removeDguest(dguest)
	}
	return fmt.Errorf("dguest %v wasn't assumed so cannot be forgotten", key)
}

// Assumes that lock is already acquired.
func (cache *cacheImpl) addDguest(dguest *v1alpha1.Dguest, assumeDguest bool) error {
	key, err := framework.GetDguestKey(dguest)
	if err != nil {
		return err
	}
	selectFood := dguest.Spec.FoodNamespacedName
	n, ok := cache.foods[selectFood]
	if !ok {
		n = newFoodInfoListItem(framework.NewFoodInfo())
		cache.foods[selectFood] = n
	}
	n.info.AddDguest(dguest)
	cache.moveFoodInfoToHead(selectFood)
	ps := &dguestState{
		dguest: dguest,
	}
	cache.dguestStates[key] = ps
	if assumeDguest {
		cache.assumedDguests.Insert(key)
	}
	return nil
}

// Assumes that lock is already acquired.
func (cache *cacheImpl) updateDguest(oldDguest, newDguest *v1alpha1.Dguest) error {
	if err := cache.removeDguest(oldDguest); err != nil {
		return err
	}
	return cache.addDguest(newDguest, false)
}

// Assumes that lock is already acquired.
// Removes a dguest from the cached food info. If the food information was already
// removed and there are no more dguests left in the food, cleans up the food from
// the cache.
func (cache *cacheImpl) removeDguest(dguest *v1alpha1.Dguest) error {
	key, err := framework.GetDguestKey(dguest)
	if err != nil {
		return err
	}

	selectFood := dguest.Spec.FoodNamespacedName

	n, ok := cache.foods[selectFood]
	if !ok {
		klog.ErrorS(nil, "Food not found when trying to remove dguest", "food", selectFood, "dguest", klog.KObj(dguest))
	} else {
		if err := n.info.RemoveDguest(dguest); err != nil {
			return err
		}
		if len(n.info.Dguests) == 0 && n.info.Food() == nil {
			cache.removeFoodInfoFromList(selectFood)
		} else {
			cache.moveFoodInfoToHead(selectFood)
		}
	}

	delete(cache.dguestStates, key)
	delete(cache.assumedDguests, key)
	return nil
}

func (cache *cacheImpl) AddDguest(dguest *v1alpha1.Dguest) error {
	key, err := framework.GetDguestKey(dguest)
	if err != nil {
		return err
	}

	cache.mu.Lock()
	defer cache.mu.Unlock()

	currState, ok := cache.dguestStates[key]
	switch {
	case ok && cache.assumedDguests.Has(key):
		if currState.dguest.Spec.FoodNamespacedName != dguest.Spec.FoodNamespacedName {
			// The dguest was added to a different food than it was assumed to.
			klog.InfoS("Dguest was added to a different food than it was assumed", "dguest", klog.KObj(dguest), "assumedFood", dguest.Spec.FoodNamespacedName, "currentFood", currState.dguest.Spec.FoodNamespacedName)
			if err = cache.updateDguest(currState.dguest, dguest); err != nil {
				klog.ErrorS(err, "Error occurred while updating dguest")
			}
		} else {
			delete(cache.assumedDguests, key)
			cache.dguestStates[key].deadline = nil
			cache.dguestStates[key].dguest = dguest
		}
	case !ok:
		// Dguest was expired. We should add it back.
		if err = cache.addDguest(dguest, false); err != nil {
			klog.ErrorS(err, "Error occurred while adding dguest")
		}
	default:
		return fmt.Errorf("dguest %v was already in added state", key)
	}
	return nil
}

func (cache *cacheImpl) UpdateDguest(oldDguest, newDguest *v1alpha1.Dguest) error {
	key, err := framework.GetDguestKey(oldDguest)
	if err != nil {
		return err
	}

	cache.mu.Lock()
	defer cache.mu.Unlock()

	currState, ok := cache.dguestStates[key]
	// An assumed dguest won't have Update/Remove event. It needs to have Add event
	// before Update event, in which case the state would change from Assumed to Added.
	if ok && !cache.assumedDguests.Has(key) {
		if currState.dguest.Spec.FoodNamespacedName != newDguest.Spec.FoodNamespacedName {
			klog.ErrorS(nil, "Dguest updated on a different food than previously added to", "dguest", klog.KObj(oldDguest))
			klog.ErrorS(nil, "scheduler cache is corrupted and can badly affect scheduling decisions")
			os.Exit(1)
		}
		return cache.updateDguest(oldDguest, newDguest)
	}
	return fmt.Errorf("dguest %v is not added to scheduler cache, so cannot be updated", key)
}

func (cache *cacheImpl) RemoveDguest(dguest *v1alpha1.Dguest) error {
	key, err := framework.GetDguestKey(dguest)
	if err != nil {
		return err
	}

	cache.mu.Lock()
	defer cache.mu.Unlock()

	currState, ok := cache.dguestStates[key]
	if !ok {
		return fmt.Errorf("dguest %v is not found in scheduler cache, so cannot be removed from it", key)
	}

	if currState.dguest.Spec.FoodNamespacedName != dguest.Spec.FoodNamespacedName {
		klog.ErrorS(nil, "Dguest was added to a different food than it was assumed", "dguest", klog.KObj(dguest), "assumedFood", dguest.Spec.FoodNamespacedName, "currentFood", currState.dguest.Spec.FoodNamespacedName)
		if len(dguest.Spec.FoodNamespacedName) > 0 {
			// An empty FoodName is possible when the scheduler misses a Delete
			// event and it gets the last known state from the informer cache.
			klog.ErrorS(nil, "scheduler cache is corrupted and can badly affect scheduling decisions")
			os.Exit(1)
		}
	}
	return cache.removeDguest(currState.dguest)
}

func (cache *cacheImpl) IsAssumedDguest(dguest *v1alpha1.Dguest) (bool, error) {
	key, err := framework.GetDguestKey(dguest)
	if err != nil {
		return false, err
	}

	cache.mu.RLock()
	defer cache.mu.RUnlock()

	return cache.assumedDguests.Has(key), nil
}

// GetDguest might return a dguest for which its food has already been deleted from
// the main cache. This is useful to properly process dguest update events.
func (cache *cacheImpl) GetDguest(dguest *v1alpha1.Dguest) (*v1alpha1.Dguest, error) {
	key, err := framework.GetDguestKey(dguest)
	if err != nil {
		return nil, err
	}

	cache.mu.RLock()
	defer cache.mu.RUnlock()

	dguestState, ok := cache.dguestStates[key]
	if !ok {
		return nil, fmt.Errorf("dguest %v does not exist in scheduler cache", key)
	}

	return dguestState.dguest, nil
}

func (cache *cacheImpl) AddFood(food *v1alpha1.Food) *framework.FoodInfo {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	n, ok := cache.foods[food.Name]
	if !ok {
		n = newFoodInfoListItem(framework.NewFoodInfo())
		cache.foods[food.Name] = n
	} else {
		cache.removeFoodImageStates(n.info.Food())
	}
	cache.moveFoodInfoToHead(food.Name)

	cache.foodTree.addFood(food)
	cache.addFoodImageStates(food, n.info)
	n.info.SetFood(food)
	return n.info.Clone()
}

func (cache *cacheImpl) UpdateFood(oldFood, newFood *v1alpha1.Food) *framework.FoodInfo {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	n, ok := cache.foods[newFood.Name]
	if !ok {
		n = newFoodInfoListItem(framework.NewFoodInfo())
		cache.foods[newFood.Name] = n
		cache.foodTree.addFood(newFood)
	} else {
		cache.removeFoodImageStates(n.info.Food())
	}
	cache.moveFoodInfoToHead(newFood.Name)

	cache.foodTree.updateFood(oldFood, newFood)
	cache.addFoodImageStates(newFood, n.info)
	n.info.SetFood(newFood)
	return n.info.Clone()
}

// RemoveFood removes a food from the cache's tree.
// The food might still have dguests because their deletion events didn't arrive
// yet. Those dguests are considered removed from the cache, being the food tree
// the source of truth.
// However, we keep a ghost food with the list of dguests until all dguest deletion
// events have arrived. A ghost food is skipped from snapshots.
func (cache *cacheImpl) RemoveFood(food *v1alpha1.Food) error {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	n, ok := cache.foods[food.Name]
	if !ok {
		return fmt.Errorf("food %v is not found", food.Name)
	}
	n.info.RemoveFood()
	// We remove FoodInfo for this food only if there aren't any dguests on this food.
	// We can't do it unconditionally, because notifications about dguests are delivered
	// in a different watch, and thus can potentially be observed later, even though
	// they happened before food removal.
	if len(n.info.Dguests) == 0 {
		cache.removeFoodInfoFromList(food.Name)
	} else {
		cache.moveFoodInfoToHead(food.Name)
	}
	if err := cache.foodTree.removeFood(food); err != nil {
		return err
	}
	cache.removeFoodImageStates(food)
	return nil
}

// addFoodImageStates adds states of the images on given food to the given foodInfo and update the imageStates in
// scheduler cache. This function assumes the lock to scheduler cache has been acquired.
func (cache *cacheImpl) addFoodImageStates(food *v1alpha1.Food, foodInfo *framework.FoodInfo) {
	//newSum := make(map[string]*framework.ImageStateSummary)
	//
	//for _, image := range food.Status.Images {
	//	for _, name := range image.Names {
	//		// update the entry in imageStates
	//		state, ok := cache.imageStates[name]
	//		if !ok {
	//			state = &imageState{
	//				size:  image.SizeBytes,
	//				foods: sets.NewString(food.Name),
	//			}
	//			cache.imageStates[name] = state
	//		} else {
	//			state.foods.Insert(food.Name)
	//		}
	//		// create the imageStateSummary for this image
	//		if _, ok := newSum[name]; !ok {
	//			newSum[name] = cache.createImageStateSummary(state)
	//		}
	//	}
	//}
	//foodInfo.ImageStates = newSum
}

// removeFoodImageStates removes the given food record from image entries having the food
// in imageStates cache. After the removal, if any image becomes free, i.e., the image
// is no longer available on any food, the image entry will be removed from imageStates.
func (cache *cacheImpl) removeFoodImageStates(food *v1alpha1.Food) {
	if food == nil {
		return
	}

	//for _, image := range food.Status.Images {
	//	for _, name := range image.Names {
	//		state, ok := cache.imageStates[name]
	//		if ok {
	//			state.foods.Delete(food.Name)
	//			if len(state.foods) == 0 {
	//				// Remove the unused image to make sure the length of
	//				// imageStates represents the total number of different
	//				// images on all foods
	//				delete(cache.imageStates, name)
	//			}
	//		}
	//	}
	//}
}

func (cache *cacheImpl) run() {
	go wait.Until(cache.cleanupExpiredAssumedDguests, cache.period, cache.stop)
}

func (cache *cacheImpl) cleanupExpiredAssumedDguests() {
	cache.cleanupAssumedDguests(time.Now())
}

// cleanupAssumedDguests exists for making test deterministic by taking time as input argument.
// It also reports metrics on the cache size for foods, dguests, and assumed dguests.
func (cache *cacheImpl) cleanupAssumedDguests(now time.Time) {
	cache.mu.Lock()
	defer cache.mu.Unlock()
	defer cache.updateMetrics()

	// The size of assumedDguests should be small
	for key := range cache.assumedDguests {
		ps, ok := cache.dguestStates[key]
		if !ok {
			klog.ErrorS(nil, "Key found in assumed set but not in dguestStates, potentially a logical error")
			os.Exit(1)
		}
		if !ps.bindingFinished {
			klog.V(5).InfoS("Could not expire cache for dguest as binding is still in progress",
				"dguest", klog.KObj(ps.dguest))
			continue
		}
		if cache.ttl != 0 && now.After(*ps.deadline) {
			klog.InfoS("Dguest expired", "dguest", klog.KObj(ps.dguest))
			if err := cache.removeDguest(ps.dguest); err != nil {
				klog.ErrorS(err, "ExpireDguest failed", "dguest", klog.KObj(ps.dguest))
			}
		}
	}
}

// updateMetrics updates cache size metric values for dguests, assumed dguests, and foods
func (cache *cacheImpl) updateMetrics() {
	metrics.CacheSize.WithLabelValues("assumed_dguests").Set(float64(len(cache.assumedDguests)))
	metrics.CacheSize.WithLabelValues("dguests").Set(float64(len(cache.dguestStates)))
	metrics.CacheSize.WithLabelValues("foods").Set(float64(len(cache.foods)))
}
