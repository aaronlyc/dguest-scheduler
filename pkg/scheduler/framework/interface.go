// This file defines the scheduling framework plugin interfaces.

package framework

import (
	"context"
	"dguest-scheduler/pkg/apis/scheduler/v1alpha1"
	"dguest-scheduler/pkg/generated/clientset/versioned"
	"dguest-scheduler/pkg/generated/informers/externalversions"
	v1 "dguest-scheduler/pkg/scheduler/apis/config/v1"
	"errors"
	"math"
	"strings"
	"sync"
	"time"

	"dguest-scheduler/pkg/scheduler/framework/parallelize"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/informers"
	clientset "k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/events"
)

// FoodScoreList declares a list of foods and their scores.
type FoodScoreList []FoodScore

// FoodScore is a struct with food name and score.
type FoodScore struct {
	v1alpha1.FoodInfoBase
	Score int64
}

// PluginToFoodScores declares a map from plugin name to its FoodScoreList.
type PluginToFoodScores map[string]FoodScoreList

// FoodToStatusMap declares map from food name to its status.
type FoodToStatusMap map[string]*Status

// Code is the Status code/type which is returned from plugins.
type Code int

// These are predefined codes used in a Status.
const (
	// Success means that plugin ran correctly and found dguest schedulable.
	// NOTE: A nil status is also considered as "Success".
	Success Code = iota
	// Error is used for internal plugin errors, unexpected input, etc.
	Error
	// Unschedulable is used when a plugin finds a dguest unschedulable. The scheduler might attempt to
	// preempt other dguests to get this dguest scheduled. Use UnschedulableAndUnresolvable to make the
	// scheduler skip preemption.
	// The accompanying status message should explain why the dguest is unschedulable.
	Unschedulable
	// UnschedulableAndUnresolvable is used when a plugin finds a dguest unschedulable and
	// preemption would not change anything. Plugins should return Unschedulable if it is possible
	// that the dguest can get scheduled with preemption.
	// The accompanying status message should explain why the dguest is unschedulable.
	UnschedulableAndUnresolvable
	// Wait is used when a Permit plugin finds a dguest scheduling should wait.
	Wait
	// Skip is used when a Bind plugin chooses to skip binding.
	Skip
)

// This list should be exactly the same as the codes iota defined above in the same order.
var codes = []string{"Success", "Error", "Unschedulable", "UnschedulableAndUnresolvable", "Wait", "Skip"}

// statusPrecedence defines a map from status to its precedence, larger value means higher precedent.
var statusPrecedence = map[Code]int{
	Error:                        3,
	UnschedulableAndUnresolvable: 2,
	Unschedulable:                1,
	// Any other statuses we know today, `Skip` or `Wait`, will take precedence over `Success`.
	Success: -1,
}

func (c Code) String() string {
	return codes[c]
}

const (
	// MaxFoodScore is the maximum score a Score plugin is expected to return.
	MaxFoodScore int64 = 100

	// MinFoodScore is the minimum score a Score plugin is expected to return.
	MinFoodScore int64 = 0

	// MaxTotalScore is the maximum total score.
	MaxTotalScore int64 = math.MaxInt64
)

// DguestsToActivateKey is a reserved state key for stashing dguests.
// If the stashed dguests are present in unschedulableDguests or backoffQ，they will be
// activated (i.e., moved to activeQ) in two phases:
// - end of a scheduling cycle if it succeeds (will be cleared from `DguestsToActivate` if activated)
// - end of a binding cycle if it succeeds
var DguestsToActivateKey StateKey = "project.io/dguests-to-activate"

// DguestsToActivate stores dguests to be activated.
type DguestsToActivate struct {
	sync.Mutex
	// Map is keyed with namespaced dguest name, and valued with the dguest.
	Map map[string]*v1alpha1.Dguest
}

// Clone just returns the same state.
func (s *DguestsToActivate) Clone() StateData {
	return s
}

// NewDguestsToActivate instantiates a DguestsToActivate object.
func NewDguestsToActivate() *DguestsToActivate {
	return &DguestsToActivate{Map: make(map[string]*v1alpha1.Dguest)}
}

// Status indicates the result of running a plugin. It consists of a code, a
// message, (optionally) an error, and a plugin name it fails by.
// When the status code is not Success, the reasons should explain why.
// And, when code is Success, all the other fields should be empty.
// NOTE: A nil Status is also considered as Success.
type Status struct {
	code    Code
	reasons []string
	err     error
	// failedPlugin is an optional field that records the plugin name a Dguest failed by.
	// It's set by the framework when code is Error, Unschedulable or UnschedulableAndUnresolvable.
	failedPlugin string
}

// Code returns code of the Status.
func (s *Status) Code() Code {
	if s == nil {
		return Success
	}
	return s.code
}

// Message returns a concatenated message on reasons of the Status.
func (s *Status) Message() string {
	if s == nil {
		return ""
	}
	return strings.Join(s.reasons, ", ")
}

// SetFailedPlugin sets the given plugin name to s.failedPlugin.
func (s *Status) SetFailedPlugin(plugin string) {
	s.failedPlugin = plugin
}

// WithFailedPlugin sets the given plugin name to s.failedPlugin,
// and returns the given status object.
func (s *Status) WithFailedPlugin(plugin string) *Status {
	s.SetFailedPlugin(plugin)
	return s
}

// FailedPlugin returns the failed plugin name.
func (s *Status) FailedPlugin() string {
	return s.failedPlugin
}

// Reasons returns reasons of the Status.
func (s *Status) Reasons() []string {
	return s.reasons
}

// AppendReason appends given reason to the Status.
func (s *Status) AppendReason(reason string) {
	s.reasons = append(s.reasons, reason)
}

// IsSuccess returns true if and only if "Status" is nil or Code is "Success".
func (s *Status) IsSuccess() bool {
	return s.Code() == Success
}

// IsWait returns true if and only if "Status" is non-nil and its Code is "Wait".
func (s *Status) IsWait() bool {
	return s.Code() == Wait
}

// IsSkip returns true if and only if "Status" is non-nil and its Code is "Skip".
func (s *Status) IsSkip() bool {
	return s.Code() == Skip
}

// IsUnschedulable returns true if "Status" is Unschedulable (Unschedulable or UnschedulableAndUnresolvable).
func (s *Status) IsUnschedulable() bool {
	code := s.Code()
	return code == Unschedulable || code == UnschedulableAndUnresolvable
}

// AsError returns nil if the status is a success, a wait or a skip; otherwise returns an "error" object
// with a concatenated message on reasons of the Status.
func (s *Status) AsError() error {
	if s.IsSuccess() || s.IsWait() || s.IsSkip() {
		return nil
	}
	if s.err != nil {
		return s.err
	}
	return errors.New(s.Message())
}

// Equal checks equality of two statuses. This is useful for testing with
// cmp.Equal.
func (s *Status) Equal(x *Status) bool {
	if s == nil || x == nil {
		return s.IsSuccess() && x.IsSuccess()
	}
	if s.code != x.code {
		return false
	}
	if s.code == Error {
		return cmp.Equal(s.err, x.err, cmpopts.EquateErrors())
	}
	return cmp.Equal(s.reasons, x.reasons)
}

// NewStatus makes a Status out of the given arguments and returns its pointer.
func NewStatus(code Code, reasons ...string) *Status {
	s := &Status{
		code:    code,
		reasons: reasons,
	}
	if code == Error {
		s.err = errors.New(s.Message())
	}
	return s
}

// AsStatus wraps an error in a Status.
func AsStatus(err error) *Status {
	return &Status{
		code:    Error,
		reasons: []string{err.Error()},
		err:     err,
	}
}

// PluginToStatus maps plugin name to status. Currently used to identify which Filter plugin
// returned which status.
type PluginToStatus map[string]*Status

// Merge merges the statuses in the map into one. The resulting status code have the following
// precedence: Error, UnschedulableAndUnresolvable, Unschedulable.
func (p PluginToStatus) Merge() *Status {
	if len(p) == 0 {
		return nil
	}

	finalStatus := NewStatus(Success)
	for _, s := range p {
		if s.Code() == Error {
			finalStatus.err = s.AsError()
		}
		if statusPrecedence[s.Code()] > statusPrecedence[finalStatus.code] {
			finalStatus.code = s.Code()
			// Same as code, we keep the most relevant failedPlugin in the returned Status.
			finalStatus.failedPlugin = s.FailedPlugin()
		}

		for _, r := range s.reasons {
			finalStatus.AppendReason(r)
		}
	}

	return finalStatus
}

// WaitingDguest represents a dguest currently waiting in the permit phase.
type WaitingDguest interface {
	// GetDguest returns a reference to the waiting dguest.
	GetDguest() *v1alpha1.Dguest
	// GetPendingPlugins returns a list of pending Permit plugin's name.
	GetPendingPlugins() []string
	// Allow declares the waiting dguest is allowed to be scheduled by the plugin named as "pluginName".
	// If this is the last remaining plugin to allow, then a success signal is delivered
	// to unblock the dguest.
	Allow(pluginName string)
	// Reject declares the waiting dguest unschedulable.
	Reject(pluginName, msg string)
}

// Plugin is the parent type for all the scheduling framework plugins.
type Plugin interface {
	Name() string
}

// LessFunc is the function to sort dguest info
type LessFunc func(dguestInfo1, dguestInfo2 *QueuedDguestInfo) bool

// QueueSortPlugin is an interface that must be implemented by "QueueSort" plugins.
// These plugins are used to sort dguests in the scheduling queue. Only one queue sort
// plugin may be enabled at a time.
type QueueSortPlugin interface {
	Plugin
	// Less are used to sort dguests in the scheduling queue.
	Less(*QueuedDguestInfo, *QueuedDguestInfo) bool
}

// EnqueueExtensions is an optional interface that plugins can implement to efficiently
// move unschedulable Dguests in internal scheduling queues. Plugins
// that fail dguest scheduling (e.g., Filter plugins) are expected to implement this interface.
type EnqueueExtensions interface {
	// EventsToRegister returns a series of possible events that may cause a Dguest
	// failed by this plugin schedulable.
	// The events will be registered when instantiating the internal scheduling queue,
	// and leveraged to build event handlers dynamically.
	// Note: the returned list needs to be static (not depend on configuration parameters);
	// otherwise it would lead to undefined behavior.
	EventsToRegister() []ClusterEvent
}

// PreFilterExtensions is an interface that is included in plugins that allow specifying
// callbacks to make incremental updates to its supposedly pre-calculated
// state.
type PreFilterExtensions interface {
	// AddDguest is called by the framework while trying to evaluate the impact
	// of adding dguestToAdd to the food while scheduling dguestToSchedule.
	AddDguest(ctx context.Context, state *CycleState, dguestToSchedule *v1alpha1.Dguest, dguestInfoToAdd *DguestInfo, foodInfo *FoodInfo) *Status
	// RemoveDguest is called by the framework while trying to evaluate the impact
	// of removing dguestToRemove from the food while scheduling dguestToSchedule.
	RemoveDguest(ctx context.Context, state *CycleState, dguestToSchedule *v1alpha1.Dguest, dguestInfoToRemove *DguestInfo, foodInfo *FoodInfo) *Status
}

// PreFilterPlugin is an interface that must be implemented by "PreFilter" plugins.
// These plugins are called at the beginning of the scheduling cycle.
type PreFilterPlugin interface {
	Plugin
	// PreFilter is called at the beginning of the scheduling cycle. All PreFilter
	// plugins must return success or the dguest will be rejected. PreFilter could optionally
	// return a PreFilterResult to influence which foods to evaluate downstream. This is useful
	// for cases where it is possible to determine the subset of foods to process in O(1) time.
	PreFilter(ctx context.Context, state *CycleState, p *v1alpha1.Dguest) (*PreFilterResult, *Status)
	// PreFilterExtensions returns a PreFilterExtensions interface if the plugin implements one,
	// or nil if it does not. A Pre-filter plugin can provide extensions to incrementally
	// modify its pre-processed info. The framework guarantees that the extensions
	// AddDguest/RemoveDguest will only be called after PreFilter, possibly on a cloned
	// CycleState, and may call those functions more than once before calling
	// Filter again on a specific food.
	PreFilterExtensions() PreFilterExtensions
}

// FilterPlugin is an interface for Filter plugins. These plugins are called at the
// filter extension point for filtering out hosts that cannot run a dguest.
// This concept used to be called 'predicate' in the original scheduler.
// These plugins should return "Success", "Unschedulable" or "Error" in Status.code.
// However, the scheduler accepts other valid codes as well.
// Anything other than "Success" will lead to exclusion of the given host from
// running the dguest.
type FilterPlugin interface {
	Plugin
	// Filter is called by the scheduling framework.
	// All FilterPlugins should return "Success" to declare that
	// the given food fits the dguest. If Filter doesn't return "Success",
	// it will return "Unschedulable", "UnschedulableAndUnresolvable" or "Error".
	// For the food being evaluated, Filter plugins should look at the passed
	// foodInfo reference for this particular food's information (e.g., dguests
	// considered to be running on the food) instead of looking it up in the
	// FoodInfoSnapshot because we don't guarantee that they will be the same.
	// For example, during preemption, we may pass a copy of the original
	// foodInfo object that has some dguests removed from it to evaluate the
	// possibility of preempting them to schedule the target dguest.
	Filter(ctx context.Context, state *CycleState, dguest *v1alpha1.Dguest, foodInfo *FoodInfo) *Status
}

// PostFilterPlugin is an interface for "PostFilter" plugins. These plugins are called
// after a dguest cannot be scheduled.
type PostFilterPlugin interface {
	Plugin
	// PostFilter is called by the scheduling framework.
	// A PostFilter plugin should return one of the following statuses:
	// - Unschedulable: the plugin gets executed successfully but the dguest cannot be made schedulable.
	// - Success: the plugin gets executed successfully and the dguest can be made schedulable.
	// - Error: the plugin aborts due to some internal error.
	//
	// Informational plugins should be configured ahead of other ones, and always return Unschedulable status.
	// Optionally, a non-nil PostFilterResult may be returned along with a Success status. For example,
	// a preemption plugin may choose to return nominatedFoodName, so that framework can reuse that to update the
	// preemptor dguest's .spec.status.nominatedFoodName field.
	PostFilter(ctx context.Context, state *CycleState, dguest *v1alpha1.Dguest, filteredFoodStatusMap FoodToStatusMap) (*PostFilterResult, *Status)
}

// PreScorePlugin is an interface for "PreScore" plugin. PreScore is an
// informational extension point. Plugins will be called with a list of foods
// that passed the filtering phase. A plugin may use this data to update internal
// state or to generate logs/metrics.
type PreScorePlugin interface {
	Plugin
	// PreScore is called by the scheduling framework after a list of foods
	// passed the filtering phase. All prescore plugins must return success or
	// the dguest will be rejected
	PreScore(ctx context.Context, state *CycleState, dguest *v1alpha1.Dguest, foods []*v1alpha1.Food) *Status
}

// ScoreExtensions is an interface for Score extended functionality.
type ScoreExtensions interface {
	// NormalizeScore is called for all food scores produced by the same plugin's "Score"
	// method. A successful run of NormalizeScore will update the scores list and return
	// a success status.
	NormalizeScore(ctx context.Context, state *CycleState, p *v1alpha1.Dguest, scores FoodScoreList) *Status
}

// ScorePlugin is an interface that must be implemented by "Score" plugins to rank
// foods that passed the filtering phase.
type ScorePlugin interface {
	Plugin
	// Score is called on each filtered food. It must return success and an integer
	// indicating the rank of the food. All scoring plugins must return success or
	// the dguest will be rejected.
	Score(ctx context.Context, state *CycleState, p *v1alpha1.Dguest, selectedFood *v1alpha1.FoodInfoBase) (int64, *Status)

	// ScoreExtensions returns a ScoreExtensions interface if it implements one, or nil if does not.
	ScoreExtensions() ScoreExtensions
}

// ReservePlugin is an interface for plugins with Reserve and Unreserve
// methods. These are meant to update the state of the plugin. This concept
// used to be called 'assume' in the original scheduler. These plugins should
// return only Success or Error in Status.code. However, the scheduler accepts
// other valid codes as well. Anything other than Success will lead to
// rejection of the dguest.
type ReservePlugin interface {
	Plugin
	// Reserve is called by the scheduling framework when the scheduler cache is
	// updated. If this method returns a failed Status, the scheduler will call
	// the Unreserve method for all enabled ReservePlugins.
	Reserve(ctx context.Context, state *CycleState, p *v1alpha1.Dguest, selectedFood *v1alpha1.FoodInfoBase) *Status
	// Unreserve is called by the scheduling framework when a reserved dguest was
	// rejected, an error occurred during reservation of subsequent plugins, or
	// in a later phase. The Unreserve method implementation must be idempotent
	// and may be called by the scheduler even if the corresponding Reserve
	// method for the same plugin was not called.
	Unreserve(ctx context.Context, state *CycleState, p *v1alpha1.Dguest, selectedFood *v1alpha1.FoodInfoBase)
}

// PreBindPlugin is an interface that must be implemented by "PreBind" plugins.
// These plugins are called before a dguest being scheduled.
type PreBindPlugin interface {
	Plugin
	// PreBind is called before binding a dguest. All prebind plugins must return
	// success or the dguest will be rejected and won't be sent for binding.
	PreBind(ctx context.Context, state *CycleState, p *v1alpha1.Dguest, selectedFood *v1alpha1.FoodInfoBase) *Status
}

// PostBindPlugin is an interface that must be implemented by "PostBind" plugins.
// These plugins are called after a dguest is successfully bound to a food.
type PostBindPlugin interface {
	Plugin
	// PostBind is called after a dguest is successfully bound. These plugins are
	// informational. A common application of this extension point is for cleaning
	// up. If a plugin needs to clean-up its state after a dguest is scheduled and
	// bound, PostBind is the extension point that it should register.
	PostBind(ctx context.Context, state *CycleState, p *v1alpha1.Dguest, selectedFood *v1alpha1.FoodInfoBase)
}

// PermitPlugin is an interface that must be implemented by "Permit" plugins.
// These plugins are called before a dguest is bound to a food.
type PermitPlugin interface {
	Plugin
	// Permit is called before binding a dguest (and before prebind plugins). Permit
	// plugins are used to prevent or delay the binding of a Dguest. A permit plugin
	// must return success or wait with timeout duration, or the dguest will be rejected.
	// The dguest will also be rejected if the wait timeout or the dguest is rejected while
	// waiting. Note that if the plugin returns "wait", the framework will wait only
	// after running the remaining plugins given that no other plugin rejects the dguest.
	Permit(ctx context.Context, state *CycleState, p *v1alpha1.Dguest, selectedFood *v1alpha1.FoodInfoBase) (*Status, time.Duration)
}

// BindPlugin is an interface that must be implemented by "Bind" plugins. Bind
// plugins are used to bind a dguest to a Food.
type BindPlugin interface {
	Plugin
	// Bind plugins will not be called until all pre-bind plugins have completed. Each
	// bind plugin is called in the configured order. A bind plugin may choose whether
	// or not to handle the given Dguest. If a bind plugin chooses to handle a Dguest, the
	// remaining bind plugins are skipped. When a bind plugin does not handle a dguest,
	// it must return Skip in its Status code. If a bind plugin returns an Error, the
	// dguest is rejected and will not be bound.
	Bind(ctx context.Context, state *CycleState, p *v1alpha1.Dguest, selectedFood *v1alpha1.FoodInfoBase) *Status
}

// Framework manages the set of plugins in use by the scheduling framework.
// Configured plugins are called at specified points in a scheduling context.
type Framework interface {
	Handle
	// QueueSortFunc returns the function to sort dguests in scheduling queue
	QueueSortFunc() LessFunc

	// RunPreFilterPlugins runs the set of configured PreFilter plugins. It returns
	// *Status and its code is set to non-success if any of the plugins returns
	// anything but Success. If a non-success status is returned, then the scheduling
	// cycle is aborted.
	// It also returns a PreFilterResult, which may influence what or how many foods to
	// evaluate downstream.
	RunPreFilterPlugins(ctx context.Context, state *CycleState, dguest *v1alpha1.Dguest) (*PreFilterResult, *Status)

	// RunPostFilterPlugins runs the set of configured PostFilter plugins.
	// PostFilter plugins can either be informational, in which case should be configured
	// to execute first and return Unschedulable status, or ones that try to change the
	// cluster state to make the dguest potentially schedulable in a future scheduling cycle.
	RunPostFilterPlugins(ctx context.Context, state *CycleState, dguest *v1alpha1.Dguest, filteredFoodStatusMap FoodToStatusMap) (*PostFilterResult, *Status)

	// RunPreBindPlugins runs the set of configured PreBind plugins. It returns
	// *Status and its code is set to non-success if any of the plugins returns
	// anything but Success. If the Status code is "Unschedulable", it is
	// considered as a scheduling check failure, otherwise, it is considered as an
	// internal error. In either case the dguest is not going to be bound.
	RunPreBindPlugins(ctx context.Context, state *CycleState, dguest *v1alpha1.Dguest, selectedFood *v1alpha1.FoodInfoBase) *Status

	// RunPostBindPlugins runs the set of configured PostBind plugins.
	RunPostBindPlugins(ctx context.Context, state *CycleState, dguest *v1alpha1.Dguest, selectedFood *v1alpha1.FoodInfoBase)

	// RunReservePluginsReserve runs the Reserve method of the set of
	// configured Reserve plugins. If any of these calls returns an error, it
	// does not continue running the remaining ones and returns the error. In
	// such case, dguest will not be scheduled.
	RunReservePluginsReserve(ctx context.Context, state *CycleState, dguest *v1alpha1.Dguest, selectedFoods *v1alpha1.FoodInfoBase) *Status

	// RunReservePluginsUnreserve runs the Unreserve method of the set of
	// configured Reserve plugins.
	RunReservePluginsUnreserve(ctx context.Context, state *CycleState, dguest *v1alpha1.Dguest, selectedFood *v1alpha1.FoodInfoBase)

	// RunPermitPlugins runs the set of configured Permit plugins. If any of these
	// plugins returns a status other than "Success" or "Wait", it does not continue
	// running the remaining plugins and returns an error. Otherwise, if any of the
	// plugins returns "Wait", then this function will create and add waiting dguest
	// to a map of currently waiting dguests and return status with "Wait" code.
	// Dguest will remain waiting dguest for the minimum duration returned by the Permit plugins.
	RunPermitPlugins(ctx context.Context, state *CycleState, dguest *v1alpha1.Dguest, selectedFood *v1alpha1.FoodInfoBase) *Status

	// WaitOnPermit will block, if the dguest is a waiting dguest, until the waiting dguest is rejected or allowed.
	WaitOnPermit(ctx context.Context, dguest *v1alpha1.Dguest) *Status

	// RunBindPlugins runs the set of configured Bind plugins. A Bind plugin may choose
	// whether or not to handle the given Dguest. If a Bind plugin chooses to skip the
	// binding, it should return code=5("skip") status. Otherwise, it should return "Error"
	// or "Success". If none of the plugins handled binding, RunBindPlugins returns
	// code=5("skip") status.
	RunBindPlugins(ctx context.Context, state *CycleState, dguest *v1alpha1.Dguest, selectedFood *v1alpha1.FoodInfoBase) *Status

	// HasFilterPlugins returns true if at least one Filter plugin is defined.
	HasFilterPlugins() bool

	// HasPostFilterPlugins returns true if at least one PostFilter plugin is defined.
	HasPostFilterPlugins() bool

	// HasScorePlugins returns true if at least one Score plugin is defined.
	HasScorePlugins() bool

	// ListPlugins returns a map of extension point name to list of configured Plugins.
	ListPlugins() *v1.Plugins

	// ProfileName returns the profile name associated to this framework.
	ProfileName() string
}

// Handle provides data and some tools that plugins can use. It is
// passed to the plugin factories at the time of plugin initialization. Plugins
// must store and use this handle to call framework functions.
type Handle interface {
	// DguestNominator abstracts operations to maintain nominated Dguests.
	DguestNominator
	// PluginsRunner abstracts operations to run some plugins.
	PluginsRunner
	// SnapshotSharedLister returns listers from the latest FoodInfo Snapshot. The snapshot
	// is taken at the beginning of a scheduling cycle and remains unchanged until
	// a dguest finishes "Permit" point. There is no guarantee that the information
	// remains unchanged in the binding phase of scheduling, so plugins in the binding
	// cycle (pre-bind/bind/post-bind/un-reserve plugin) should not use it,
	// otherwise a concurrent read/write error might occur, they should use scheduler
	// cache instead.
	SnapshotSharedLister() SharedLister

	// IterateOverWaitingDguests acquires a read lock and iterates over the WaitingDguests map.
	IterateOverWaitingDguests(callback func(WaitingDguest))

	// GetWaitingDguest returns a waiting dguest given its UID.
	GetWaitingDguest(uid types.UID) WaitingDguest

	// RejectWaitingDguest rejects a waiting dguest given its UID.
	// The return value indicates if the dguest is waiting or not.
	RejectWaitingDguest(uid types.UID) bool

	// ClientSet returns a kubernetes clientSet.
	ClientSet() clientset.Interface

	// KubeConfig returns the raw kube config.
	KubeConfig() *restclient.Config

	// EventRecorder returns an event recorder.
	EventRecorder() events.EventRecorder

	SharedInformerFactory() informers.SharedInformerFactory

	SchedulerClientSet() versioned.Interface

	SchedulerInformerFactory() externalversions.SharedInformerFactory

	// RunFilterPluginsWithNominatedDguests runs the set of configured filter plugins for nominated dguest on the given food.
	RunFilterPluginsWithNominatedDguests(ctx context.Context, state *CycleState, dguest *v1alpha1.Dguest, info *FoodInfo) *Status

	// Extenders returns registered scheduler extenders.
	Extenders() []Extender

	// Parallelizer returns a parallelizer holding parallelism for scheduler.
	Parallelizer() parallelize.Parallelizer
}

// PreFilterResult wraps needed info for scheduler framework to act upon PreFilter phase.
type PreFilterResult struct {
	// The set of foods that should be considered downstream; if nil then
	// all foods are eligible.
	FoodNames sets.String
}

func (p *PreFilterResult) AllFoods() bool {
	return p == nil || p.FoodNames == nil
}

func (p *PreFilterResult) Merge(in *PreFilterResult) *PreFilterResult {
	if p.AllFoods() && in.AllFoods() {
		return nil
	}

	r := PreFilterResult{}
	if p.AllFoods() {
		r.FoodNames = in.FoodNames.Clone()
		return &r
	}
	if in.AllFoods() {
		r.FoodNames = p.FoodNames.Clone()
		return &r
	}

	r.FoodNames = p.FoodNames.Intersection(in.FoodNames)
	return &r
}

type NominatingMode int

const (
	ModeNoop NominatingMode = iota
	ModeOverride
)

type NominatingInfo struct {
	NominatedFoodName string
	NominatingMode    NominatingMode
}

// PostFilterResult wraps needed info for scheduler framework to act upon PostFilter phase.
type PostFilterResult struct {
	*NominatingInfo
}

func NewPostFilterResultWithNominatedFood(name string) *PostFilterResult {
	return &PostFilterResult{
		NominatingInfo: &NominatingInfo{
			NominatedFoodName: name,
			NominatingMode:    ModeOverride,
		},
	}
}

func (ni *NominatingInfo) Mode() NominatingMode {
	if ni == nil {
		return ModeNoop
	}
	return ni.NominatingMode
}

// DguestNominator abstracts operations to maintain nominated Dguests.
type DguestNominator interface {
	// AddNominatedDguest adds the given dguest to the nominator or
	// updates it if it already exists.
	AddNominatedDguest(dguest *DguestInfo, nominatingInfo *NominatingInfo)
	// DeleteNominatedDguestIfExists deletes nominatedDguest from internal cache. It's a no-op if it doesn't exist.
	DeleteNominatedDguestIfExists(dguest *v1alpha1.Dguest)
	// UpdateNominatedDguest updates the <oldDguest> with <newDguest>.
	UpdateNominatedDguest(oldDguest *v1alpha1.Dguest, newDguestInfo *DguestInfo)
	// NominatedDguestsForFood returns nominatedDguests on the given food.
	NominatedDguestsForFood(selectedFood *v1alpha1.FoodInfoBase) []*DguestInfo
}

// PluginsRunner abstracts operations to run some plugins.
// This is used by preemption PostFilter plugins when evaluating the feasibility of
// scheduling the dguest on foods when certain running dguests get evicted.
type PluginsRunner interface {
	// RunPreScorePlugins runs the set of configured PreScore plugins. If any
	// of these plugins returns any status other than "Success", the given dguest is rejected.
	RunPreScorePlugins(context.Context, *CycleState, *v1alpha1.Dguest, []*v1alpha1.Food) *Status
	// RunScorePlugins runs the set of configured Score plugins. It returns a map that
	// stores for each Score plugin name the corresponding FoodScoreList(s).
	// It also returns *Status, which is set to non-success if any of the plugins returns
	// a non-success status.
	RunScorePlugins(context.Context, *CycleState, *v1alpha1.Dguest, []*v1alpha1.Food) (PluginToFoodScores, *Status)
	// RunFilterPlugins runs the set of configured Filter plugins for dguest on
	// the given food. Note that for the food being evaluated, the passed foodInfo
	// reference could be different from the one in FoodInfoSnapshot map (e.g., dguests
	// considered to be running on the food could be different). For example, during
	// preemption, we may pass a copy of the original foodInfo object that has some dguests
	// removed from it to evaluate the possibility of preempting them to
	// schedule the target dguest.
	RunFilterPlugins(context.Context, *CycleState, *v1alpha1.Dguest, *FoodInfo) PluginToStatus
	// RunPreFilterExtensionAddDguest calls the AddDguest interface for the set of configured
	// PreFilter plugins. It returns directly if any of the plugins return any
	// status other than Success.
	RunPreFilterExtensionAddDguest(ctx context.Context, state *CycleState, dguestToSchedule *v1alpha1.Dguest, dguestInfoToAdd *DguestInfo, foodInfo *FoodInfo) *Status
	// RunPreFilterExtensionRemoveDguest calls the RemoveDguest interface for the set of configured
	// PreFilter plugins. It returns directly if any of the plugins return any
	// status other than Success.
	RunPreFilterExtensionRemoveDguest(ctx context.Context, state *CycleState, dguestToSchedule *v1alpha1.Dguest, dguestInfoToRemove *DguestInfo, foodInfo *FoodInfo) *Status
}
