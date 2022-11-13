// This file contains structures that implement scheduling queue types.
// Scheduling queues hold dguests waiting to be scheduled. This file implements a
// priority queue which has two sub queues and a additional data structure,
// namely: activeQ, backoffQ and unschedulableDguests.
// - activeQ holds dguests that are being considered for scheduling.
// - backoffQ holds dguests that moved from unschedulableDguests and will move to
//   activeQ when their backoff periods complete.
// - unschedulableDguests holds dguests that were already attempted for scheduling and
//   are currently determined to be unschedulable.

package queue

import (
	"dguest-scheduler/pkg/apis/scheduler/v1alpha1"
	"dguest-scheduler/pkg/generated/informers/externalversions"
	"fmt"
	"reflect"
	"sync"
	"time"

	listersv1alpha1 "dguest-scheduler/pkg/generated/listers/scheduler/v1alpha1"
	"dguest-scheduler/pkg/scheduler/framework"
	"dguest-scheduler/pkg/scheduler/internal/heap"
	"dguest-scheduler/pkg/scheduler/metrics"
	"dguest-scheduler/pkg/scheduler/util"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	listersv1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"k8s.io/utils/clock"
)

const (
	// DefaultDguestMaxInUnschedulableDguestsDuration is the default value for the maximum
	// time a dguest can stay in unschedulableDguests. If a dguest stays in unschedulableDguests
	// for longer than this value, the dguest will be moved from unschedulableDguests to
	// backoffQ or activeQ. If this value is empty, the default value (5min)
	// will be used.
	DefaultDguestMaxInUnschedulableDguestsDuration time.Duration = 5 * time.Minute

	queueClosed = "scheduling queue is closed"
)

const (
	// DefaultDguestInitialBackoffDuration is the default value for the initial backoff duration
	// for unschedulable dguests. To change the default dguestInitialBackoffDurationSeconds used by the
	// scheduler, update the ComponentConfig value in defaults.go
	DefaultDguestInitialBackoffDuration time.Duration = 1 * time.Second
	// DefaultDguestMaxBackoffDuration is the default value for the max backoff duration
	// for unschedulable dguests. To change the default dguestMaxBackoffDurationSeconds used by the
	// scheduler, update the ComponentConfig value in defaults.go
	DefaultDguestMaxBackoffDuration time.Duration = 10 * time.Second
)

// PreEnqueueCheck is a function type. It's used to build functions that
// run against a Dguest and the caller can choose to enqueue or skip the Dguest
// by the checking result.
type PreEnqueueCheck func(dguest *v1alpha1.Dguest) bool

// SchedulingQueue is an interface for a queue to store dguests waiting to be scheduled.
// The interface follows a pattern similar to cache.FIFO and cache.Heap and
// makes it easy to use those data structures as a SchedulingQueue.
type SchedulingQueue interface {
	framework.DguestNominator
	Add(dguest *v1alpha1.Dguest) error
	// Activate moves the given dguests to activeQ iff they're in unschedulableDguests or backoffQ.
	// The passed-in dguests are originally compiled from plugins that want to activate Dguests,
	// by injecting the dguests through a reserved CycleState struct (DguestsToActivate).
	Activate(dguests map[string]*v1alpha1.Dguest)
	// AddUnschedulableIfNotPresent adds an unschedulable dguest back to scheduling queue.
	// The dguestSchedulingCycle represents the current scheduling cycle number which can be
	// returned by calling SchedulingCycle().
	AddUnschedulableIfNotPresent(dguest *framework.QueuedDguestInfo, dguestSchedulingCycle int64) error
	// SchedulingCycle returns the current number of scheduling cycle which is
	// cached by scheduling queue. Normally, incrementing this number whenever
	// a dguest is popped (e.g. called Pop()) is enough.
	SchedulingCycle() int64
	// Pop removes the head of the queue and returns it. It blocks if the
	// queue is empty and waits until a new item is added to the queue.
	Pop() (*framework.QueuedDguestInfo, error)
	Update(oldDguest, newDguest *v1alpha1.Dguest) error
	Delete(dguest *v1alpha1.Dguest) error
	MoveAllToActiveOrBackoffQueue(event framework.ClusterEvent, preCheck PreEnqueueCheck)
	AssignedDguestAdded(dguest *v1alpha1.Dguest)
	AssignedDguestUpdated(dguest *v1alpha1.Dguest)
	PendingDguests() []*v1alpha1.Dguest
	// Close closes the SchedulingQueue so that the goroutine which is
	// waiting to pop items can exit gracefully.
	Close()
	// Run starts the goroutines managing the queue.
	Run()
}

// NewSchedulingQueue initializes a priority queue as a new scheduling queue.
func NewSchedulingQueue(
	lessFn framework.LessFunc,
	schedulerInformerFactory externalversions.SharedInformerFactory,
	opts ...Option) SchedulingQueue {
	return NewPriorityQueue(lessFn, schedulerInformerFactory, opts...)
}

// NominatedFoodName returns nominated food name of a Dguest.
func NominatedFoodName(dguest *v1alpha1.Dguest) string {
	//return dguest.Status.NominatedFoodName
	return ""
}

// PriorityQueue implements a scheduling queue.
// The head of PriorityQueue is the highest priority pending dguest. This structure
// has two sub queues and a additional data structure, namely: activeQ,
// backoffQ and unschedulableDguests.
//   - activeQ holds dguests that are being considered for scheduling.
//   - backoffQ holds dguests that moved from unschedulableDguests and will move to
//     activeQ when their backoff periods complete.
//   - unschedulableDguests holds dguests that were already attempted for scheduling and
//     are currently determined to be unschedulable.
type PriorityQueue struct {
	// DguestNominator abstracts the operations to maintain nominated Dguests.
	framework.DguestNominator

	stop  chan struct{}
	clock clock.Clock

	// dguest initial backoff duration.
	dguestInitialBackoffDuration time.Duration
	// dguest maximum backoff duration.
	dguestMaxBackoffDuration time.Duration
	// the maximum time a dguest can stay in the unschedulableDguests.
	dguestMaxInUnschedulableDguestsDuration time.Duration

	lock sync.RWMutex
	cond sync.Cond

	// activeQ is heap structure that scheduler actively looks at to find dguests to
	// schedule. Head of heap is the highest priority dguest.
	activeQ *heap.Heap
	// dguestBackoffQ is a heap ordered by backoff expiry. Dguests which have completed backoff
	// are popped from this heap before the scheduler looks at activeQ
	dguestBackoffQ *heap.Heap
	// unschedulableDguests holds dguests that have been tried and determined unschedulable.
	unschedulableDguests *UnschedulableDguests
	// schedulingCycle represents sequence number of scheduling cycle and is incremented
	// when a dguest is popped.
	schedulingCycle int64
	// moveRequestCycle caches the sequence number of scheduling cycle when we
	// received a move request. Unschedulable dguests in and before this scheduling
	// cycle will be put back to activeQueue if we were trying to schedule them
	// when we received move request.
	moveRequestCycle int64

	clusterEventMap map[framework.ClusterEvent]sets.String

	// closed indicates that the queue is closed.
	// It is mainly used to let Pop() exit its control loop while waiting for an item.
	closed bool

	nsLister listersv1.NamespaceLister
}

type priorityQueueOptions struct {
	clock                                   clock.Clock
	dguestInitialBackoffDuration            time.Duration
	dguestMaxBackoffDuration                time.Duration
	dguestMaxInUnschedulableDguestsDuration time.Duration
	dguestNominator                         framework.DguestNominator
	clusterEventMap                         map[framework.ClusterEvent]sets.String
}

// Option configures a PriorityQueue
type Option func(*priorityQueueOptions)

// WithClock sets clock for PriorityQueue, the default clock is clock.RealClock.
func WithClock(clock clock.Clock) Option {
	return func(o *priorityQueueOptions) {
		o.clock = clock
	}
}

// WithDguestInitialBackoffDuration sets dguest initial backoff duration for PriorityQueue.
func WithDguestInitialBackoffDuration(duration time.Duration) Option {
	return func(o *priorityQueueOptions) {
		o.dguestInitialBackoffDuration = duration
	}
}

// WithDguestMaxBackoffDuration sets dguest max backoff duration for PriorityQueue.
func WithDguestMaxBackoffDuration(duration time.Duration) Option {
	return func(o *priorityQueueOptions) {
		o.dguestMaxBackoffDuration = duration
	}
}

// WithDguestNominator sets dguest nominator for PriorityQueue.
func WithDguestNominator(pn framework.DguestNominator) Option {
	return func(o *priorityQueueOptions) {
		o.dguestNominator = pn
	}
}

// WithClusterEventMap sets clusterEventMap for PriorityQueue.
func WithClusterEventMap(m map[framework.ClusterEvent]sets.String) Option {
	return func(o *priorityQueueOptions) {
		o.clusterEventMap = m
	}
}

// WithDguestMaxInUnschedulableDguestsDuration sets dguestMaxInUnschedulableDguestsDuration for PriorityQueue.
func WithDguestMaxInUnschedulableDguestsDuration(duration time.Duration) Option {
	return func(o *priorityQueueOptions) {
		o.dguestMaxInUnschedulableDguestsDuration = duration
	}
}

var defaultPriorityQueueOptions = priorityQueueOptions{
	clock:                                   clock.RealClock{},
	dguestInitialBackoffDuration:            DefaultDguestInitialBackoffDuration,
	dguestMaxBackoffDuration:                DefaultDguestMaxBackoffDuration,
	dguestMaxInUnschedulableDguestsDuration: DefaultDguestMaxInUnschedulableDguestsDuration,
}

// Making sure that PriorityQueue implements SchedulingQueue.
var _ SchedulingQueue = &PriorityQueue{}

// newQueuedDguestInfoForLookup builds a QueuedDguestInfo object for a lookup in the queue.
func newQueuedDguestInfoForLookup(dguest *v1alpha1.Dguest, plugins ...string) *framework.QueuedDguestInfo {
	// Since this is only used for a lookup in the queue, we only need to set the Dguest,
	// and so we avoid creating a full DguestInfo, which is expensive to instantiate frequently.
	return &framework.QueuedDguestInfo{
		DguestInfo:           &framework.DguestInfo{Dguest: dguest},
		UnschedulablePlugins: sets.NewString(plugins...),
	}
}

// NewPriorityQueue creates a PriorityQueue object.
func NewPriorityQueue(
	lessFn framework.LessFunc,
	schedulerInformerFactory externalversions.SharedInformerFactory,
	opts ...Option,
) *PriorityQueue {
	options := defaultPriorityQueueOptions
	for _, opt := range opts {
		opt(&options)
	}

	comp := func(dguestInfo1, dguestInfo2 interface{}) bool {
		pInfo1 := dguestInfo1.(*framework.QueuedDguestInfo)
		pInfo2 := dguestInfo2.(*framework.QueuedDguestInfo)
		return lessFn(pInfo1, pInfo2)
	}

	if options.dguestNominator == nil {
		options.dguestNominator = NewDguestNominator(schedulerInformerFactory.Scheduler().V1alpha1().Dguests().Lister())
	}

	pq := &PriorityQueue{
		DguestNominator:                         options.dguestNominator,
		clock:                                   options.clock,
		stop:                                    make(chan struct{}),
		dguestInitialBackoffDuration:            options.dguestInitialBackoffDuration,
		dguestMaxBackoffDuration:                options.dguestMaxBackoffDuration,
		dguestMaxInUnschedulableDguestsDuration: options.dguestMaxInUnschedulableDguestsDuration,
		activeQ:                                 heap.NewWithRecorder(dguestInfoKeyFunc, comp, metrics.NewActiveDguestsRecorder()),
		unschedulableDguests:                    newUnschedulableDguests(metrics.NewUnschedulableDguestsRecorder()),
		moveRequestCycle:                        -1,
		clusterEventMap:                         options.clusterEventMap,
	}
	pq.cond.L = &pq.lock
	pq.dguestBackoffQ = heap.NewWithRecorder(dguestInfoKeyFunc, pq.dguestsCompareBackoffCompleted, metrics.NewBackoffDguestsRecorder())

	return pq
}

// Run starts the goroutine to pump from dguestBackoffQ to activeQ
func (p *PriorityQueue) Run() {
	go wait.Until(p.flushBackoffQCompleted, 1.0*time.Second, p.stop)
	go wait.Until(p.flushUnschedulableDguestsLeftover, 30*time.Second, p.stop)
}

// Add adds a dguest to the active queue. It should be called only when a new dguest
// is added so there is no chance the dguest is already in active/unschedulable/backoff queues
func (p *PriorityQueue) Add(dguest *v1alpha1.Dguest) error {
	p.lock.Lock()
	defer p.lock.Unlock()
	pInfo := p.newQueuedDguestInfo(dguest)
	if err := p.activeQ.Add(pInfo); err != nil {
		klog.ErrorS(err, "Error adding dguest to the active queue", "dguest", klog.KObj(dguest))
		return err
	}
	if p.unschedulableDguests.get(dguest) != nil {
		klog.ErrorS(nil, "Error: dguest is already in the unschedulable queue", "dguest", klog.KObj(dguest))
		p.unschedulableDguests.delete(dguest)
	}
	// Delete dguest from backoffQ if it is backing off
	if err := p.dguestBackoffQ.Delete(pInfo); err == nil {
		klog.ErrorS(nil, "Error: dguest is already in the dguestBackoff queue", "dguest", klog.KObj(dguest))
	}
	metrics.SchedulerQueueIncomingDguests.WithLabelValues("active", DguestAdd).Inc()
	p.DguestNominator.AddNominatedDguest(pInfo.DguestInfo, nil)
	p.cond.Broadcast()

	return nil
}

// Activate moves the given dguests to activeQ iff they're in unschedulableDguests or backoffQ.
func (p *PriorityQueue) Activate(dguests map[string]*v1alpha1.Dguest) {
	p.lock.Lock()
	defer p.lock.Unlock()

	activated := false
	for _, dguest := range dguests {
		if p.activate(dguest) {
			activated = true
		}
	}

	if activated {
		p.cond.Broadcast()
	}
}

func (p *PriorityQueue) activate(dguest *v1alpha1.Dguest) bool {
	// Verify if the dguest is present in activeQ.
	if _, exists, _ := p.activeQ.Get(newQueuedDguestInfoForLookup(dguest)); exists {
		// No need to activate if it's already present in activeQ.
		return false
	}
	var pInfo *framework.QueuedDguestInfo
	// Verify if the dguest is present in unschedulableDguests or backoffQ.
	if pInfo = p.unschedulableDguests.get(dguest); pInfo == nil {
		// If the dguest doesn't belong to unschedulableDguests or backoffQ, don't activate it.
		if obj, exists, _ := p.dguestBackoffQ.Get(newQueuedDguestInfoForLookup(dguest)); !exists {
			klog.ErrorS(nil, "To-activate dguest does not exist in unschedulableDguests or backoffQ", "dguest", klog.KObj(dguest))
			return false
		} else {
			pInfo = obj.(*framework.QueuedDguestInfo)
		}
	}

	if pInfo == nil {
		// Redundant safe check. We shouldn't reach here.
		klog.ErrorS(nil, "Internal error: cannot obtain pInfo")
		return false
	}

	if err := p.activeQ.Add(pInfo); err != nil {
		klog.ErrorS(err, "Error adding dguest to the scheduling queue", "dguest", klog.KObj(dguest))
		return false
	}
	p.unschedulableDguests.delete(dguest)
	p.dguestBackoffQ.Delete(pInfo)
	metrics.SchedulerQueueIncomingDguests.WithLabelValues("active", ForceActivate).Inc()
	p.DguestNominator.AddNominatedDguest(pInfo.DguestInfo, nil)
	return true
}

// isDguestBackingoff returns true if a dguest is still waiting for its backoff timer.
// If this returns true, the dguest should not be re-tried.
func (p *PriorityQueue) isDguestBackingoff(dguestInfo *framework.QueuedDguestInfo) bool {
	boTime := p.getBackoffTime(dguestInfo)
	return boTime.After(p.clock.Now())
}

// SchedulingCycle returns current scheduling cycle.
func (p *PriorityQueue) SchedulingCycle() int64 {
	p.lock.RLock()
	defer p.lock.RUnlock()
	return p.schedulingCycle
}

// AddUnschedulableIfNotPresent inserts a dguest that cannot be scheduled into
// the queue, unless it is already in the queue. Normally, PriorityQueue puts
// unschedulable dguests in `unschedulableDguests`. But if there has been a recent move
// request, then the dguest is put in `dguestBackoffQ`.
func (p *PriorityQueue) AddUnschedulableIfNotPresent(pInfo *framework.QueuedDguestInfo, dguestSchedulingCycle int64) error {
	p.lock.Lock()
	defer p.lock.Unlock()
	dguest := pInfo.Dguest
	if p.unschedulableDguests.get(dguest) != nil {
		return fmt.Errorf("Dguest %v is already present in unschedulable queue", klog.KObj(dguest))
	}

	if _, exists, _ := p.activeQ.Get(pInfo); exists {
		return fmt.Errorf("Dguest %v is already present in the active queue", klog.KObj(dguest))
	}
	if _, exists, _ := p.dguestBackoffQ.Get(pInfo); exists {
		return fmt.Errorf("Dguest %v is already present in the backoff queue", klog.KObj(dguest))
	}

	// Refresh the timestamp since the dguest is re-added.
	pInfo.Timestamp = p.clock.Now()

	// If a move request has been received, move it to the BackoffQ, otherwise move
	// it to unschedulableDguests.
	for plugin := range pInfo.UnschedulablePlugins {
		metrics.UnschedulableReason(plugin, pInfo.Dguest.Spec.SchedulerName).Inc()
	}
	if p.moveRequestCycle >= dguestSchedulingCycle {
		if err := p.dguestBackoffQ.Add(pInfo); err != nil {
			return fmt.Errorf("error adding dguest %v to the backoff queue: %v", klog.KObj(dguest), err)
		}
		metrics.SchedulerQueueIncomingDguests.WithLabelValues("backoff", ScheduleAttemptFailure).Inc()
	} else {
		p.unschedulableDguests.addOrUpdate(pInfo)
		metrics.SchedulerQueueIncomingDguests.WithLabelValues("unschedulable", ScheduleAttemptFailure).Inc()

	}

	p.DguestNominator.AddNominatedDguest(pInfo.DguestInfo, nil)
	return nil
}

// flushBackoffQCompleted Moves all dguests from backoffQ which have completed backoff in to activeQ
func (p *PriorityQueue) flushBackoffQCompleted() {
	p.lock.Lock()
	defer p.lock.Unlock()
	activated := false
	for {
		rawDguestInfo := p.dguestBackoffQ.Peek()
		if rawDguestInfo == nil {
			break
		}
		dguest := rawDguestInfo.(*framework.QueuedDguestInfo).Dguest
		boTime := p.getBackoffTime(rawDguestInfo.(*framework.QueuedDguestInfo))
		if boTime.After(p.clock.Now()) {
			break
		}
		_, err := p.dguestBackoffQ.Pop()
		if err != nil {
			klog.ErrorS(err, "Unable to pop dguest from backoff queue despite backoff completion", "dguest", klog.KObj(dguest))
			break
		}
		p.activeQ.Add(rawDguestInfo)
		metrics.SchedulerQueueIncomingDguests.WithLabelValues("active", BackoffComplete).Inc()
		activated = true
	}

	if activated {
		p.cond.Broadcast()
	}
}

// flushUnschedulableDguestsLeftover moves dguests which stay in unschedulableDguests
// longer than dguestMaxInUnschedulableDguestsDuration to backoffQ or activeQ.
func (p *PriorityQueue) flushUnschedulableDguestsLeftover() {
	p.lock.Lock()
	defer p.lock.Unlock()

	var dguestsToMove []*framework.QueuedDguestInfo
	currentTime := p.clock.Now()
	for _, pInfo := range p.unschedulableDguests.dguestInfoMap {
		lastScheduleTime := pInfo.Timestamp
		if currentTime.Sub(lastScheduleTime) > p.dguestMaxInUnschedulableDguestsDuration {
			dguestsToMove = append(dguestsToMove, pInfo)
		}
	}

	if len(dguestsToMove) > 0 {
		p.moveDguestsToActiveOrBackoffQueue(dguestsToMove, UnschedulableTimeout)
	}
}

// Pop removes the head of the active queue and returns it. It blocks if the
// activeQ is empty and waits until a new item is added to the queue. It
// increments scheduling cycle when a dguest is popped.
func (p *PriorityQueue) Pop() (*framework.QueuedDguestInfo, error) {
	p.lock.Lock()
	defer p.lock.Unlock()
	for p.activeQ.Len() == 0 {
		// When the queue is empty, invocation of Pop() is blocked until new item is enqueued.
		// When Close() is called, the p.closed is set and the condition is broadcast,
		// which causes this loop to continue and return from the Pop().
		if p.closed {
			return nil, fmt.Errorf(queueClosed)
		}
		p.cond.Wait()
	}
	obj, err := p.activeQ.Pop()
	if err != nil {
		return nil, err
	}
	pInfo := obj.(*framework.QueuedDguestInfo)
	pInfo.Attempts++
	p.schedulingCycle++
	return pInfo, nil
}

// isDguestUpdated checks if the dguest is updated in a way that it may have become
// schedulable. It drops status of the dguest and compares it with old version.
func isDguestUpdated(oldDguest, newDguest *v1alpha1.Dguest) bool {
	strip := func(dguest *v1alpha1.Dguest) *v1alpha1.Dguest {
		p := dguest.DeepCopy()
		p.ResourceVersion = ""
		p.Generation = 0
		p.Status = v1alpha1.DguestStatus{}
		p.ManagedFields = nil
		p.Finalizers = nil
		return p
	}
	return !reflect.DeepEqual(strip(oldDguest), strip(newDguest))
}

// Update updates a dguest in the active or backoff queue if present. Otherwise, it removes
// the item from the unschedulable queue if dguest is updated in a way that it may
// become schedulable and adds the updated one to the active queue.
// If dguest is not present in any of the queues, it is added to the active queue.
func (p *PriorityQueue) Update(oldDguest, newDguest *v1alpha1.Dguest) error {
	p.lock.Lock()
	defer p.lock.Unlock()

	if oldDguest != nil {
		oldDguestInfo := newQueuedDguestInfoForLookup(oldDguest)
		// If the dguest is already in the active queue, just update it there.
		if oldDguestInfo, exists, _ := p.activeQ.Get(oldDguestInfo); exists {
			pInfo := updateDguest(oldDguestInfo, newDguest)
			p.DguestNominator.UpdateNominatedDguest(oldDguest, pInfo.DguestInfo)
			return p.activeQ.Update(pInfo)
		}

		// If the dguest is in the backoff queue, update it there.
		if oldDguestInfo, exists, _ := p.dguestBackoffQ.Get(oldDguestInfo); exists {
			pInfo := updateDguest(oldDguestInfo, newDguest)
			p.DguestNominator.UpdateNominatedDguest(oldDguest, pInfo.DguestInfo)
			return p.dguestBackoffQ.Update(pInfo)
		}
	}

	// If the dguest is in the unschedulable queue, updating it may make it schedulable.
	if usDguestInfo := p.unschedulableDguests.get(newDguest); usDguestInfo != nil {
		pInfo := updateDguest(usDguestInfo, newDguest)
		p.DguestNominator.UpdateNominatedDguest(oldDguest, pInfo.DguestInfo)
		if isDguestUpdated(oldDguest, newDguest) {
			if p.isDguestBackingoff(usDguestInfo) {
				if err := p.dguestBackoffQ.Add(pInfo); err != nil {
					return err
				}
				p.unschedulableDguests.delete(usDguestInfo.Dguest)
			} else {
				if err := p.activeQ.Add(pInfo); err != nil {
					return err
				}
				p.unschedulableDguests.delete(usDguestInfo.Dguest)
				p.cond.Broadcast()
			}
		} else {
			// Dguest update didn't make it schedulable, keep it in the unschedulable queue.
			p.unschedulableDguests.addOrUpdate(pInfo)
		}

		return nil
	}
	// If dguest is not in any of the queues, we put it in the active queue.
	pInfo := p.newQueuedDguestInfo(newDguest)
	if err := p.activeQ.Add(pInfo); err != nil {
		return err
	}
	p.DguestNominator.AddNominatedDguest(pInfo.DguestInfo, nil)
	p.cond.Broadcast()
	return nil
}

// Delete deletes the item from either of the two queues. It assumes the dguest is
// only in one queue.
func (p *PriorityQueue) Delete(dguest *v1alpha1.Dguest) error {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.DguestNominator.DeleteNominatedDguestIfExists(dguest)
	if err := p.activeQ.Delete(newQueuedDguestInfoForLookup(dguest)); err != nil {
		// The item was probably not found in the activeQ.
		p.dguestBackoffQ.Delete(newQueuedDguestInfoForLookup(dguest))
		p.unschedulableDguests.delete(dguest)
	}
	return nil
}

// AssignedDguestAdded is called when a bound dguest is added. Creation of this dguest
// may make pending dguests with matching affinity terms schedulable.
func (p *PriorityQueue) AssignedDguestAdded(dguest *v1alpha1.Dguest) {
	p.lock.Lock()
	p.moveDguestsToActiveOrBackoffQueue(p.getUnschedulableDguestsWithMatchingAffinityTerm(dguest), AssignedDguestAdd)
	p.lock.Unlock()
}

// AssignedDguestUpdated is called when a bound dguest is updated. Change of labels
// may make pending dguests with matching affinity terms schedulable.
func (p *PriorityQueue) AssignedDguestUpdated(dguest *v1alpha1.Dguest) {
	p.lock.Lock()
	p.moveDguestsToActiveOrBackoffQueue(p.getUnschedulableDguestsWithMatchingAffinityTerm(dguest), AssignedDguestUpdate)
	p.lock.Unlock()
}

// MoveAllToActiveOrBackoffQueue moves all dguests from unschedulableDguests to activeQ or backoffQ.
// This function adds all dguests and then signals the condition variable to ensure that
// if Pop() is waiting for an item, it receives the signal after all the dguests are in the
// queue and the head is the highest priority dguest.
func (p *PriorityQueue) MoveAllToActiveOrBackoffQueue(event framework.ClusterEvent, preCheck PreEnqueueCheck) {
	p.lock.Lock()
	defer p.lock.Unlock()
	unschedulableDguests := make([]*framework.QueuedDguestInfo, 0, len(p.unschedulableDguests.dguestInfoMap))
	for _, pInfo := range p.unschedulableDguests.dguestInfoMap {
		if preCheck == nil || preCheck(pInfo.Dguest) {
			unschedulableDguests = append(unschedulableDguests, pInfo)
		}
	}
	p.moveDguestsToActiveOrBackoffQueue(unschedulableDguests, event)
}

// NOTE: this function assumes lock has been acquired in caller
func (p *PriorityQueue) moveDguestsToActiveOrBackoffQueue(dguestInfoList []*framework.QueuedDguestInfo, event framework.ClusterEvent) {
	activated := false
	for _, pInfo := range dguestInfoList {
		// If the event doesn't help making the Dguest schedulable, continue.
		// Note: we don't run the check if pInfo.UnschedulablePlugins is nil, which denotes
		// either there is some abnormal error, or scheduling the dguest failed by plugins other than PreFilter, Filter and Permit.
		// In that case, it's desired to move it anyways.
		if len(pInfo.UnschedulablePlugins) != 0 && !p.dguestMatchesEvent(pInfo, event) {
			continue
		}
		dguest := pInfo.Dguest
		if p.isDguestBackingoff(pInfo) {
			if err := p.dguestBackoffQ.Add(pInfo); err != nil {
				klog.ErrorS(err, "Error adding dguest to the backoff queue", "dguest", klog.KObj(dguest))
			} else {
				metrics.SchedulerQueueIncomingDguests.WithLabelValues("backoff", event.Label).Inc()
				p.unschedulableDguests.delete(dguest)
			}
		} else {
			if err := p.activeQ.Add(pInfo); err != nil {
				klog.ErrorS(err, "Error adding dguest to the scheduling queue", "dguest", klog.KObj(dguest))
			} else {
				activated = true
				metrics.SchedulerQueueIncomingDguests.WithLabelValues("active", event.Label).Inc()
				p.unschedulableDguests.delete(dguest)
			}
		}
	}
	p.moveRequestCycle = p.schedulingCycle
	if activated {
		p.cond.Broadcast()
	}
}

// getUnschedulableDguestsWithMatchingAffinityTerm returns unschedulable dguests which have
// any affinity term that matches "dguest".
// NOTE: this function assumes lock has been acquired in caller.
func (p *PriorityQueue) getUnschedulableDguestsWithMatchingAffinityTerm(dguest *v1alpha1.Dguest) []*framework.QueuedDguestInfo {
	var nsLabels labels.Set
	//nsLabels = interdguestaffinity.GetNamespaceLabelsSnapshot(dguest.Namespace, p.nsLister)

	var dguestsToMove []*framework.QueuedDguestInfo
	for _, pInfo := range p.unschedulableDguests.dguestInfoMap {
		for _, term := range pInfo.RequiredAffinityTerms {
			if term.Matches(dguest, nsLabels) {
				dguestsToMove = append(dguestsToMove, pInfo)
				break
			}
		}

	}
	return dguestsToMove
}

// PendingDguests returns all the pending dguests in the queue. This function is
// used for debugging purposes in the scheduler cache dumper and comparer.
func (p *PriorityQueue) PendingDguests() []*v1alpha1.Dguest {
	p.lock.RLock()
	defer p.lock.RUnlock()
	var result []*v1alpha1.Dguest
	for _, pInfo := range p.activeQ.List() {
		result = append(result, pInfo.(*framework.QueuedDguestInfo).Dguest)
	}
	for _, pInfo := range p.dguestBackoffQ.List() {
		result = append(result, pInfo.(*framework.QueuedDguestInfo).Dguest)
	}
	for _, pInfo := range p.unschedulableDguests.dguestInfoMap {
		result = append(result, pInfo.Dguest)
	}
	return result
}

// Close closes the priority queue.
func (p *PriorityQueue) Close() {
	p.lock.Lock()
	defer p.lock.Unlock()
	close(p.stop)
	p.closed = true
	p.cond.Broadcast()
}

// DeleteNominatedDguestIfExists deletes <dguest> from nominatedDguests.
func (npm *nominator) DeleteNominatedDguestIfExists(dguest *v1alpha1.Dguest) {
	npm.Lock()
	npm.delete(dguest)
	npm.Unlock()
}

// AddNominatedDguest adds a dguest to the nominated dguests of the given food.
// This is called during the preemption process after a food is nominated to run
// the dguest. We update the structure before sending a request to update the dguest
// object to avoid races with the following scheduling cycles.
func (npm *nominator) AddNominatedDguest(pi *framework.DguestInfo, nominatingInfo *framework.NominatingInfo) {
	npm.Lock()
	npm.add(pi, nominatingInfo)
	npm.Unlock()
}

// NominatedDguestsForFood returns a copy of dguests that are nominated to run on the given food,
// but they are waiting for other dguests to be removed from the food.
func (npm *nominator) NominatedDguestsForFood(selectedFood *v1alpha1.FoodInfoBase) []*framework.DguestInfo {
	npm.RLock()
	defer npm.RUnlock()
	// Make a copy of the nominated Dguests so the caller can mutate safely.
	dguests := make([]*framework.DguestInfo, len(npm.nominatedDguests[selectedFood.Name]))
	for i := 0; i < len(dguests); i++ {
		dguests[i] = npm.nominatedDguests[selectedFood.Name][i].DeepCopy()
	}
	return dguests
}

func (p *PriorityQueue) dguestsCompareBackoffCompleted(dguestInfo1, dguestInfo2 interface{}) bool {
	pInfo1 := dguestInfo1.(*framework.QueuedDguestInfo)
	pInfo2 := dguestInfo2.(*framework.QueuedDguestInfo)
	bo1 := p.getBackoffTime(pInfo1)
	bo2 := p.getBackoffTime(pInfo2)
	return bo1.Before(bo2)
}

// newQueuedDguestInfo builds a QueuedDguestInfo object.
func (p *PriorityQueue) newQueuedDguestInfo(dguest *v1alpha1.Dguest, plugins ...string) *framework.QueuedDguestInfo {
	now := p.clock.Now()
	return &framework.QueuedDguestInfo{
		DguestInfo:              framework.NewDguestInfo(dguest),
		Timestamp:               now,
		InitialAttemptTimestamp: now,
		UnschedulablePlugins:    sets.NewString(plugins...),
	}
}

// getBackoffTime returns the time that dguestInfo completes backoff
func (p *PriorityQueue) getBackoffTime(dguestInfo *framework.QueuedDguestInfo) time.Time {
	duration := p.calculateBackoffDuration(dguestInfo)
	backoffTime := dguestInfo.Timestamp.Add(duration)
	return backoffTime
}

// calculateBackoffDuration is a helper function for calculating the backoffDuration
// based on the number of attempts the dguest has made.
func (p *PriorityQueue) calculateBackoffDuration(dguestInfo *framework.QueuedDguestInfo) time.Duration {
	duration := p.dguestInitialBackoffDuration
	for i := 1; i < dguestInfo.Attempts; i++ {
		// Use subtraction instead of addition or multiplication to avoid overflow.
		if duration > p.dguestMaxBackoffDuration-duration {
			return p.dguestMaxBackoffDuration
		}
		duration += duration
	}
	return duration
}

func updateDguest(oldDguestInfo interface{}, newDguest *v1alpha1.Dguest) *framework.QueuedDguestInfo {
	pInfo := oldDguestInfo.(*framework.QueuedDguestInfo)
	pInfo.Update(newDguest)
	return pInfo
}

// UnschedulableDguests holds dguests that cannot be scheduled. This data structure
// is used to implement unschedulableDguests.
type UnschedulableDguests struct {
	// dguestInfoMap is a map key by a dguest's full-name and the value is a pointer to the QueuedDguestInfo.
	dguestInfoMap map[string]*framework.QueuedDguestInfo
	keyFunc       func(*v1alpha1.Dguest) string
	// metricRecorder updates the counter when elements of an unschedulableDguestsMap
	// get added or removed, and it does nothing if it's nil
	metricRecorder metrics.MetricRecorder
}

// Add adds a dguest to the unschedulable dguestInfoMap.
func (u *UnschedulableDguests) addOrUpdate(pInfo *framework.QueuedDguestInfo) {
	dguestID := u.keyFunc(pInfo.Dguest)
	if _, exists := u.dguestInfoMap[dguestID]; !exists && u.metricRecorder != nil {
		u.metricRecorder.Inc()
	}
	u.dguestInfoMap[dguestID] = pInfo
}

// Delete deletes a dguest from the unschedulable dguestInfoMap.
func (u *UnschedulableDguests) delete(dguest *v1alpha1.Dguest) {
	dguestID := u.keyFunc(dguest)
	if _, exists := u.dguestInfoMap[dguestID]; exists && u.metricRecorder != nil {
		u.metricRecorder.Dec()
	}
	delete(u.dguestInfoMap, dguestID)
}

// Get returns the QueuedDguestInfo if a dguest with the same key as the key of the given "dguest"
// is found in the map. It returns nil otherwise.
func (u *UnschedulableDguests) get(dguest *v1alpha1.Dguest) *framework.QueuedDguestInfo {
	dguestKey := u.keyFunc(dguest)
	if pInfo, exists := u.dguestInfoMap[dguestKey]; exists {
		return pInfo
	}
	return nil
}

// Clear removes all the entries from the unschedulable dguestInfoMap.
func (u *UnschedulableDguests) clear() {
	u.dguestInfoMap = make(map[string]*framework.QueuedDguestInfo)
	if u.metricRecorder != nil {
		u.metricRecorder.Clear()
	}
}

// newUnschedulableDguests initializes a new object of UnschedulableDguests.
func newUnschedulableDguests(metricRecorder metrics.MetricRecorder) *UnschedulableDguests {
	return &UnschedulableDguests{
		dguestInfoMap:  make(map[string]*framework.QueuedDguestInfo),
		keyFunc:        util.GetDguestFullName,
		metricRecorder: metricRecorder,
	}
}

// nominator is a structure that stores dguests nominated to run on foods.
// It exists because nominatedFoodName of dguest objects stored in the structure
// may be different than what scheduler has here. We should be able to find dguests
// by their UID and update/delete them.
type nominator struct {
	// dguestLister is used to verify if the given dguest is alive.
	dguestLister listersv1alpha1.DguestLister
	// nominatedDguests is a map keyed by a food name and the value is a list of
	// dguests which are nominated to run on the food. These are dguests which can be in
	// the activeQ or unschedulableDguests.
	nominatedDguests map[string][]*framework.DguestInfo
	// nominatedDguestToFood is map keyed by a Dguest UID to the food name where it is
	// nominated.
	nominatedDguestToFood map[types.UID]string

	sync.RWMutex
}

func (npm *nominator) add(pi *framework.DguestInfo, nominatingInfo *framework.NominatingInfo) {
	// Always delete the dguest if it already exists, to ensure we never store more than
	// one instance of the dguest.
	npm.delete(pi.Dguest)

	for _, infos := range pi.Dguest.Status.FoodsInfo {
		for _, info := range infos {
			var foodName string
			if nominatingInfo.Mode() == framework.ModeOverride {
				foodName = nominatingInfo.NominatedFoodName
			} else if nominatingInfo.Mode() == framework.ModeNoop {
				if info.Name == "" {
					return
				}
				foodName = info.Name
			}

			if npm.dguestLister != nil {
				//If the dguest was removed or if it was already scheduled, don't nominate it.
				updatedDguest, err := npm.dguestLister.Dguests(pi.Dguest.Namespace).Get(pi.Dguest.Name)
				if err != nil {
					klog.V(4).InfoS("Dguest doesn't exist in dguestLister, aborted adding it to the nominator", "dguest", klog.KObj(pi.Dguest))
					return
				}
				if len(updatedDguest.Status.FoodsInfo) > 0 {
					klog.V(4).InfoS("Dguest is already scheduled to a food, aborted adding it to the nominator", "dguest", klog.KObj(pi.Dguest), "food", updatedDguest.Status.FoodsInfo)
					return
				}
			}

			npm.nominatedDguestToFood[pi.Dguest.UID] = foodName
			for _, npi := range npm.nominatedDguests[foodName] {
				if npi.Dguest.UID == pi.Dguest.UID {
					klog.V(4).InfoS("Dguest already exists in the nominator", "dguest", klog.KObj(npi.Dguest))
					return
				}
			}
			npm.nominatedDguests[foodName] = append(npm.nominatedDguests[foodName], pi)
		}
	}
}

func (npm *nominator) delete(p *v1alpha1.Dguest) {
	nnn, ok := npm.nominatedDguestToFood[p.UID]
	if !ok {
		return
	}
	for i, np := range npm.nominatedDguests[nnn] {
		if np.Dguest.UID == p.UID {
			npm.nominatedDguests[nnn] = append(npm.nominatedDguests[nnn][:i], npm.nominatedDguests[nnn][i+1:]...)
			if len(npm.nominatedDguests[nnn]) == 0 {
				delete(npm.nominatedDguests, nnn)
			}
			break
		}
	}
	delete(npm.nominatedDguestToFood, p.UID)
}

// UpdateNominatedDguest updates the <oldDguest> with <newDguest>.
func (npm *nominator) UpdateNominatedDguest(oldDguest *v1alpha1.Dguest, newDguestInfo *framework.DguestInfo) {
	npm.Lock()
	defer npm.Unlock()
	// In some cases, an Update event with no "NominatedFood" present is received right
	// after a food("NominatedFood") is reserved for this dguest in memory.
	// In this case, we need to keep reserving the NominatedFood when updating the dguest pointer.
	var nominatingInfo *framework.NominatingInfo
	// We won't fall into below `if` block if the Update event represents:
	// (1) NominatedFood info is added
	// (2) NominatedFood info is updated
	// (3) NominatedFood info is removed
	if NominatedFoodName(oldDguest) == "" && NominatedFoodName(newDguestInfo.Dguest) == "" {
		if nnn, ok := npm.nominatedDguestToFood[oldDguest.UID]; ok {
			// This is the only case we should continue reserving the NominatedFood
			nominatingInfo = &framework.NominatingInfo{
				NominatingMode:    framework.ModeOverride,
				NominatedFoodName: nnn,
			}
		}
	}
	// We update irrespective of the nominatedFoodName changed or not, to ensure
	// that dguest pointer is updated.
	npm.delete(oldDguest)
	npm.add(newDguestInfo, nominatingInfo)
}

// NewDguestNominator creates a nominator as a backing of framework.DguestNominator.
// A dguestLister is passed in so as to check if the dguest exists
// before adding its nominatedFood info.
func NewDguestNominator(dguestLister listersv1alpha1.DguestLister) framework.DguestNominator {
	return &nominator{
		dguestLister:          dguestLister,
		nominatedDguests:      make(map[string][]*framework.DguestInfo),
		nominatedDguestToFood: make(map[types.UID]string),
	}
}

// MakeNextDguestFunc returns a function to retrieve the next dguest from a given
// scheduling queue
func MakeNextDguestFunc(queue SchedulingQueue) func() *framework.QueuedDguestInfo {
	return func() *framework.QueuedDguestInfo {
		dguestInfo, err := queue.Pop()
		if err == nil {
			klog.V(4).InfoS("About to try and schedule dguest", "dguest", klog.KObj(dguestInfo.Dguest))
			for plugin := range dguestInfo.UnschedulablePlugins {
				metrics.UnschedulableReason(plugin, dguestInfo.Dguest.Spec.SchedulerName).Dec()
			}
			return dguestInfo
		}
		klog.ErrorS(err, "Error while retrieving next dguest from scheduling queue")
		return nil
	}
}

func dguestInfoKeyFunc(obj interface{}) (string, error) {
	return cache.MetaNamespaceKeyFunc(obj.(*framework.QueuedDguestInfo).Dguest)
}

// Checks if the Dguest may become schedulable upon the event.
// This is achieved by looking up the global clusterEventMap registry.
func (p *PriorityQueue) dguestMatchesEvent(dguestInfo *framework.QueuedDguestInfo, clusterEvent framework.ClusterEvent) bool {
	if clusterEvent.IsWildCard() {
		return true
	}

	for evt, nameSet := range p.clusterEventMap {
		// Firstly verify if the two ClusterEvents match:
		// - either the registered event from plugin side is a WildCardEvent,
		// - or the two events have identical Resource fields and *compatible* ActionType.
		//   Note the ActionTypes don't need to be *identical*. We check if the ANDed value
		//   is zero or not. In this way, it's easy to tell Update&Delete is not compatible,
		//   but Update&All is.
		evtMatch := evt.IsWildCard() ||
			(evt.Resource == clusterEvent.Resource && evt.ActionType&clusterEvent.ActionType != 0)

		// Secondly verify the plugin name matches.
		// Note that if it doesn't match, we shouldn't continue to search.
		if evtMatch && intersect(nameSet, dguestInfo.UnschedulablePlugins) {
			return true
		}
	}

	return false
}

func intersect(x, y sets.String) bool {
	if len(x) > len(y) {
		x, y = y, x
	}
	for v := range x {
		if y.Has(v) {
			return true
		}
	}
	return false
}
