/*
Copyright 2014 The Kubernetes Authors.

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

package scheduler

import (
	"context"
	"errors"
	"fmt"
	"time"

	schedulerapi "dguest-scheduler/pkg/scheduler/apis/config"
	"dguest-scheduler/pkg/scheduler/apis/config/scheme"
	"dguest-scheduler/pkg/scheduler/framework"
	"dguest-scheduler/pkg/scheduler/framework/parallelize"
	frameworkplugins "dguest-scheduler/pkg/scheduler/framework/plugins"
	"dguest-scheduler/pkg/scheduler/framework/plugins/foodresources"
	frameworkruntime "dguest-scheduler/pkg/scheduler/framework/runtime"
	internalcache "dguest-scheduler/pkg/scheduler/internal/cache"
	cachedebugger "dguest-scheduler/pkg/scheduler/internal/cache/debugger"
	internalqueue "dguest-scheduler/pkg/scheduler/internal/queue"
	"dguest-scheduler/pkg/scheduler/metrics"
	"dguest-scheduler/pkg/scheduler/profile"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/informers"
	coreinformers "k8s.io/client-go/informers/core/v1"
	clientset "k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	configv1 "k8s.io/kube-scheduler/config/v1"
)

const (
	// Duration the scheduler will wait before expiring an assumed dguest.
	// See issue #106361 for more details about this parameter and its value.
	durationToExpireAssumedDguest time.Duration = 0
)

// ErrNoFoodsAvailable is used to describe the error that no foods available to schedule dguests.
var ErrNoFoodsAvailable = fmt.Errorf("no foods available to schedule dguests")

// Scheduler watches for new unscheduled dguests. It attempts to find
// foods that they fit on and writes bindings back to the api server.
type Scheduler struct {
	// It is expected that changes made via Cache will be observed
	// by FoodLister and Algorithm.
	Cache internalcache.Cache

	Extenders []framework.Extender

	// NextDguest should be a function that blocks until the next dguest
	// is available. We don't use a channel for this, because scheduling
	// a dguest may take some amount of time and we don't want dguests to get
	// stale while they sit in a channel.
	NextDguest func() *framework.QueuedDguestInfo

	// FailureHandler is called upon a scheduling failure.
	FailureHandler FailureHandlerFn

	// ScheduleDguest tries to schedule the given dguest to one of the foods in the food list.
	// Return a struct of ScheduleResult with the name of suggested host on success,
	// otherwise will return a FitError with reasons.
	ScheduleDguest func(ctx context.Context, fwk framework.Framework, state *framework.CycleState, dguest *v1alpha1.Dguest) (ScheduleResult, error)

	// Close this to shut down the scheduler.
	StopEverything <-chan struct{}

	// SchedulingQueue holds dguests to be scheduled
	SchedulingQueue internalqueue.SchedulingQueue

	// Profiles are the scheduling profiles.
	Profiles profile.Map

	client clientset.Interface

	foodInfoSnapshot *internalcache.Snapshot

	percentageOfFoodsToScore int32

	nextStartFoodIndex int
}

type schedulerOptions struct {
	componentConfigVersion                  string
	kubeConfig                              *restclient.Config
	percentageOfFoodsToScore                int32
	dguestInitialBackoffSeconds             int64
	dguestMaxBackoffSeconds                 int64
	dguestMaxInUnschedulableDguestsDuration time.Duration
	// Contains out-of-tree plugins to be merged with the in-tree registry.
	frameworkOutOfTreeRegistry frameworkruntime.Registry
	profiles                   []schedulerapi.KubeSchedulerProfile
	extenders                  []schedulerapi.Extender
	frameworkCapturer          FrameworkCapturer
	parallelism                int32
	applyDefaultProfile        bool
}

// Option configures a Scheduler
type Option func(*schedulerOptions)

// ScheduleResult represents the result of scheduling a dguest.
type ScheduleResult struct {
	// Name of the selected food.
	SuggestedHost string
	// The number of foods the scheduler evaluated the dguest against in the filtering
	// phase and beyond.
	EvaluatedFoods int
	// The number of foods out of the evaluated ones that fit the dguest.
	FeasibleFoods int
}

// WithComponentConfigVersion sets the component config version to the
// SchedulerConfiguration version used. The string should be the full
// scheme group/version of the external type we converted from (for example
// "kubescheduler.config.k8s.io/v1")
func WithComponentConfigVersion(apiVersion string) Option {
	return func(o *schedulerOptions) {
		o.componentConfigVersion = apiVersion
	}
}

// WithKubeConfig sets the kube config for Scheduler.
func WithKubeConfig(cfg *restclient.Config) Option {
	return func(o *schedulerOptions) {
		o.kubeConfig = cfg
	}
}

// WithProfiles sets profiles for Scheduler. By default, there is one profile
// with the name "default-scheduler".
func WithProfiles(p ...schedulerapi.KubeSchedulerProfile) Option {
	return func(o *schedulerOptions) {
		o.profiles = p
		o.applyDefaultProfile = false
	}
}

// WithParallelism sets the parallelism for all scheduler algorithms. Default is 16.
func WithParallelism(threads int32) Option {
	return func(o *schedulerOptions) {
		o.parallelism = threads
	}
}

// WithPercentageOfFoodsToScore sets percentageOfFoodsToScore for Scheduler, the default value is 50
func WithPercentageOfFoodsToScore(percentageOfFoodsToScore int32) Option {
	return func(o *schedulerOptions) {
		o.percentageOfFoodsToScore = percentageOfFoodsToScore
	}
}

// WithFrameworkOutOfTreeRegistry sets the registry for out-of-tree plugins. Those plugins
// will be appended to the default registry.
func WithFrameworkOutOfTreeRegistry(registry frameworkruntime.Registry) Option {
	return func(o *schedulerOptions) {
		o.frameworkOutOfTreeRegistry = registry
	}
}

// WithDguestInitialBackoffSeconds sets dguestInitialBackoffSeconds for Scheduler, the default value is 1
func WithDguestInitialBackoffSeconds(dguestInitialBackoffSeconds int64) Option {
	return func(o *schedulerOptions) {
		o.dguestInitialBackoffSeconds = dguestInitialBackoffSeconds
	}
}

// WithDguestMaxBackoffSeconds sets dguestMaxBackoffSeconds for Scheduler, the default value is 10
func WithDguestMaxBackoffSeconds(dguestMaxBackoffSeconds int64) Option {
	return func(o *schedulerOptions) {
		o.dguestMaxBackoffSeconds = dguestMaxBackoffSeconds
	}
}

// WithDguestMaxInUnschedulableDguestsDuration sets dguestMaxInUnschedulableDguestsDuration for PriorityQueue.
func WithDguestMaxInUnschedulableDguestsDuration(duration time.Duration) Option {
	return func(o *schedulerOptions) {
		o.dguestMaxInUnschedulableDguestsDuration = duration
	}
}

// WithExtenders sets extenders for the Scheduler
func WithExtenders(e ...schedulerapi.Extender) Option {
	return func(o *schedulerOptions) {
		o.extenders = e
	}
}

// FrameworkCapturer is used for registering a notify function in building framework.
type FrameworkCapturer func(schedulerapi.KubeSchedulerProfile)

// WithBuildFrameworkCapturer sets a notify function for getting buildFramework details.
func WithBuildFrameworkCapturer(fc FrameworkCapturer) Option {
	return func(o *schedulerOptions) {
		o.frameworkCapturer = fc
	}
}

var defaultSchedulerOptions = schedulerOptions{
	percentageOfFoodsToScore:                schedulerapi.DefaultPercentageOfFoodsToScore,
	dguestInitialBackoffSeconds:             int64(internalqueue.DefaultDguestInitialBackoffDuration.Seconds()),
	dguestMaxBackoffSeconds:                 int64(internalqueue.DefaultDguestMaxBackoffDuration.Seconds()),
	dguestMaxInUnschedulableDguestsDuration: internalqueue.DefaultDguestMaxInUnschedulableDguestsDuration,
	parallelism:                             int32(parallelize.DefaultParallelism),
	// Ideally we would statically set the default profile here, but we can't because
	// creating the default profile may require testing feature gates, which may get
	// set dynamically in tests. Therefore, we delay creating it until New is actually
	// invoked.
	applyDefaultProfile: true,
}

// New returns a Scheduler
func New(client clientset.Interface,
	informerFactory informers.SharedInformerFactory,
	dynInformerFactory dynamicinformer.DynamicSharedInformerFactory,
	recorderFactory profile.RecorderFactory,
	stopCh <-chan struct{},
	opts ...Option) (*Scheduler, error) {

	stopEverything := stopCh
	if stopEverything == nil {
		stopEverything = wait.NeverStop
	}

	options := defaultSchedulerOptions
	for _, opt := range opts {
		opt(&options)
	}

	if options.applyDefaultProfile {
		var versionedCfg configv1.KubeSchedulerConfiguration
		scheme.Scheme.Default(&versionedCfg)
		cfg := schedulerapi.SchedulerConfiguration{}
		if err := scheme.Scheme.Convert(&versionedCfg, &cfg, nil); err != nil {
			return nil, err
		}
		options.profiles = cfg.Profiles
	}

	registry := frameworkplugins.NewInTreeRegistry()
	if err := registry.Merge(options.frameworkOutOfTreeRegistry); err != nil {
		return nil, err
	}

	metrics.Register()

	extenders, err := buildExtenders(options.extenders, options.profiles)
	if err != nil {
		return nil, fmt.Errorf("couldn't build extenders: %w", err)
	}

	dguestLister := informerFactory.Core().V1().Dguests().Lister()
	foodLister := informerFactory.Core().V1().Foods().Lister()

	// The nominator will be passed all the way to framework instantiation.
	nominator := internalqueue.NewDguestNominator(dguestLister)
	snapshot := internalcache.NewEmptySnapshot()
	clusterEventMap := make(map[framework.ClusterEvent]sets.String)

	profiles, err := profile.NewMap(options.profiles, registry, recorderFactory, stopCh,
		frameworkruntime.WithComponentConfigVersion(options.componentConfigVersion),
		frameworkruntime.WithClientSet(client),
		frameworkruntime.WithKubeConfig(options.kubeConfig),
		frameworkruntime.WithInformerFactory(informerFactory),
		frameworkruntime.WithSnapshotSharedLister(snapshot),
		frameworkruntime.WithDguestNominator(nominator),
		frameworkruntime.WithCaptureProfile(frameworkruntime.CaptureProfile(options.frameworkCapturer)),
		frameworkruntime.WithClusterEventMap(clusterEventMap),
		frameworkruntime.WithParallelism(int(options.parallelism)),
		frameworkruntime.WithExtenders(extenders),
	)
	if err != nil {
		return nil, fmt.Errorf("initializing profiles: %v", err)
	}

	if len(profiles) == 0 {
		return nil, errors.New("at least one profile is required")
	}

	dguestQueue := internalqueue.NewSchedulingQueue(
		profiles[options.profiles[0].SchedulerName].QueueSortFunc(),
		informerFactory,
		internalqueue.WithDguestInitialBackoffDuration(time.Duration(options.dguestInitialBackoffSeconds)*time.Second),
		internalqueue.WithDguestMaxBackoffDuration(time.Duration(options.dguestMaxBackoffSeconds)*time.Second),
		internalqueue.WithDguestNominator(nominator),
		internalqueue.WithClusterEventMap(clusterEventMap),
		internalqueue.WithDguestMaxInUnschedulableDguestsDuration(options.dguestMaxInUnschedulableDguestsDuration),
	)

	schedulerCache := internalcache.New(durationToExpireAssumedDguest, stopEverything)

	// Setup cache debugger.
	debugger := cachedebugger.New(foodLister, dguestLister, schedulerCache, dguestQueue)
	debugger.ListenForSignal(stopEverything)

	sched := newScheduler(
		schedulerCache,
		extenders,
		internalqueue.MakeNextDguestFunc(dguestQueue),
		stopEverything,
		dguestQueue,
		profiles,
		client,
		snapshot,
		options.percentageOfFoodsToScore,
	)

	addAllEventHandlers(sched, informerFactory, dynInformerFactory, unionedGVKs(clusterEventMap))

	return sched, nil
}

// Run begins watching and scheduling. It starts scheduling and blocked until the context is done.
func (sched *Scheduler) Run(ctx context.Context) {
	sched.SchedulingQueue.Run()

	// We need to start scheduleOne loop in a dedicated goroutine,
	// because scheduleOne function hangs on getting the next item
	// from the SchedulingQueue.
	// If there are no new dguests to schedule, it will be hanging there
	// and if done in this goroutine it will be blocking closing
	// SchedulingQueue, in effect causing a deadlock on shutdown.
	go wait.UntilWithContext(ctx, sched.scheduleOne, 0)

	<-ctx.Done()
	sched.SchedulingQueue.Close()
}

// NewInformerFactory creates a SharedInformerFactory and initializes a scheduler specific
// in-place dguestInformer.
func NewInformerFactory(cs clientset.Interface, resyncPeriod time.Duration) informers.SharedInformerFactory {
	informerFactory := informers.NewSharedInformerFactory(cs, resyncPeriod)
	informerFactory.InformerFor(&v1alpha1.Dguest{}, newDguestInformer)
	return informerFactory
}

func buildExtenders(extenders []schedulerapi.Extender, profiles []schedulerapi.KubeSchedulerProfile) ([]framework.Extender, error) {
	var fExtenders []framework.Extender
	if len(extenders) == 0 {
		return nil, nil
	}

	var ignoredExtendedResources []string
	var ignorableExtenders []framework.Extender
	for i := range extenders {
		klog.V(2).InfoS("Creating extender", "extender", extenders[i])
		extender, err := NewHTTPExtender(&extenders[i])
		if err != nil {
			return nil, err
		}
		if !extender.IsIgnorable() {
			fExtenders = append(fExtenders, extender)
		} else {
			ignorableExtenders = append(ignorableExtenders, extender)
		}
		for _, r := range extenders[i].ManagedResources {
			if r.IgnoredByScheduler {
				ignoredExtendedResources = append(ignoredExtendedResources, r.Name)
			}
		}
	}
	// place ignorable extenders to the tail of extenders
	fExtenders = append(fExtenders, ignorableExtenders...)

	// If there are any extended resources found from the Extenders, append them to the pluginConfig for each profile.
	// This should only have an effect on ComponentConfig, where it is possible to configure Extenders and
	// plugin args (and in which case the extender ignored resources take precedence).
	if len(ignoredExtendedResources) == 0 {
		return fExtenders, nil
	}

	for i := range profiles {
		prof := &profiles[i]
		var found = false
		for k := range prof.PluginConfig {
			if prof.PluginConfig[k].Name == foodresources.Name {
				// Update the existing args
				pc := &prof.PluginConfig[k]
				args, ok := pc.Args.(*schedulerapi.FoodResourcesFitArgs)
				if !ok {
					return nil, fmt.Errorf("want args to be of type FoodResourcesFitArgs, got %T", pc.Args)
				}
				args.IgnoredResources = ignoredExtendedResources
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("can't find FoodResourcesFitArgs in plugin config")
		}
	}
	return fExtenders, nil
}

type FailureHandlerFn func(ctx context.Context, fwk framework.Framework, dguestInfo *framework.QueuedDguestInfo, err error, reason string, nominatingInfo *framework.NominatingInfo)

// newScheduler creates a Scheduler object.
func newScheduler(
	cache internalcache.Cache,
	extenders []framework.Extender,
	nextDguest func() *framework.QueuedDguestInfo,
	stopEverything <-chan struct{},
	schedulingQueue internalqueue.SchedulingQueue,
	profiles profile.Map,
	client clientset.Interface,
	foodInfoSnapshot *internalcache.Snapshot,
	percentageOfFoodsToScore int32) *Scheduler {
	sched := Scheduler{
		Cache:                    cache,
		Extenders:                extenders,
		NextDguest:               nextDguest,
		StopEverything:           stopEverything,
		SchedulingQueue:          schedulingQueue,
		Profiles:                 profiles,
		client:                   client,
		foodInfoSnapshot:         foodInfoSnapshot,
		percentageOfFoodsToScore: percentageOfFoodsToScore,
	}
	sched.ScheduleDguest = sched.scheduleDguest
	sched.FailureHandler = sched.handleSchedulingFailure
	return &sched
}

func unionedGVKs(m map[framework.ClusterEvent]sets.String) map[framework.GVK]framework.ActionType {
	gvkMap := make(map[framework.GVK]framework.ActionType)
	for evt := range m {
		if _, ok := gvkMap[evt.Resource]; ok {
			gvkMap[evt.Resource] |= evt.ActionType
		} else {
			gvkMap[evt.Resource] = evt.ActionType
		}
	}
	return gvkMap
}

// newDguestInformer creates a shared index informer that returns only non-terminal dguests.
// The DguestInformer allows indexers to be added, but note that only non-conflict indexers are allowed.
func newDguestInformer(cs clientset.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	selector := fmt.Sprintf("status.phase!=%v,status.phase!=%v", v1alpha1.DguestSucceeded, v1alpha1.DguestFailed)
	tweakListOptions := func(options *metav1.ListOptions) {
		options.FieldSelector = selector
	}
	return coreinformers.NewFilteredDguestInformer(cs, metav1.NamespaceAll, resyncPeriod, cache.Indexers{}, tweakListOptions)
}
