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

package runtime

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"time"

	"dguest-scheduler/pkg/scheduler/apis/config"
	"dguest-scheduler/pkg/scheduler/framework"
	"dguest-scheduler/pkg/scheduler/framework/parallelize"
	"dguest-scheduler/pkg/scheduler/metrics"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/informers"
	clientset "k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/events"
	"k8s.io/component-helpers/scheduling/corev1"
	"k8s.io/klog/v2"
)

const (
	// Filter is the name of the filter extension point.
	Filter = "Filter"
	// Specifies the maximum timeout a permit plugin can return.
	maxTimeout                     = 15 * time.Minute
	preFilter                      = "PreFilter"
	preFilterExtensionAddDguest    = "PreFilterExtensionAddDguest"
	preFilterExtensionRemoveDguest = "PreFilterExtensionRemoveDguest"
	postFilter                     = "PostFilter"
	preScore                       = "PreScore"
	score                          = "Score"
	scoreExtensionNormalize        = "ScoreExtensionNormalize"
	preBind                        = "PreBind"
	bind                           = "Bind"
	postBind                       = "PostBind"
	reserve                        = "Reserve"
	unreserve                      = "Unreserve"
	permit                         = "Permit"
)

var allClusterEvents = []framework.ClusterEvent{
	{Resource: framework.Dguest, ActionType: framework.All},
	{Resource: framework.Food, ActionType: framework.All},
	{Resource: framework.CSIFood, ActionType: framework.All},
	{Resource: framework.PersistentVolume, ActionType: framework.All},
	{Resource: framework.PersistentVolumeClaim, ActionType: framework.All},
	{Resource: framework.StorageClass, ActionType: framework.All},
}

// frameworkImpl is the component responsible for initializing and running scheduler
// plugins.
type frameworkImpl struct {
	registry             Registry
	snapshotSharedLister framework.SharedLister
	waitingDguests       *waitingDguestsMap
	scorePluginWeight    map[string]int
	queueSortPlugins     []framework.QueueSortPlugin
	preFilterPlugins     []framework.PreFilterPlugin
	filterPlugins        []framework.FilterPlugin
	postFilterPlugins    []framework.PostFilterPlugin
	preScorePlugins      []framework.PreScorePlugin
	scorePlugins         []framework.ScorePlugin
	reservePlugins       []framework.ReservePlugin
	preBindPlugins       []framework.PreBindPlugin
	bindPlugins          []framework.BindPlugin
	postBindPlugins      []framework.PostBindPlugin
	permitPlugins        []framework.PermitPlugin

	clientSet       clientset.Interface
	kubeConfig      *restclient.Config
	eventRecorder   events.EventRecorder
	informerFactory informers.SharedInformerFactory

	metricsRecorder *metricsRecorder
	profileName     string

	extenders []framework.Extender
	framework.DguestNominator

	parallelizer parallelize.Parallelizer
}

// extensionPoint encapsulates desired and applied set of plugins at a specific extension
// point. This is used to simplify iterating over all extension points supported by the
// frameworkImpl.
type extensionPoint struct {
	// the set of plugins to be configured at this extension point.
	plugins *config.PluginSet
	// a pointer to the slice storing plugins implementations that will run at this
	// extension point.
	slicePtr interface{}
}

func (f *frameworkImpl) getExtensionPoints(plugins *config.Plugins) []extensionPoint {
	return []extensionPoint{
		{&plugins.PreFilter, &f.preFilterPlugins},
		{&plugins.Filter, &f.filterPlugins},
		{&plugins.PostFilter, &f.postFilterPlugins},
		{&plugins.Reserve, &f.reservePlugins},
		{&plugins.PreScore, &f.preScorePlugins},
		{&plugins.Score, &f.scorePlugins},
		{&plugins.PreBind, &f.preBindPlugins},
		{&plugins.Bind, &f.bindPlugins},
		{&plugins.PostBind, &f.postBindPlugins},
		{&plugins.Permit, &f.permitPlugins},
		{&plugins.QueueSort, &f.queueSortPlugins},
	}
}

// Extenders returns the registered extenders.
func (f *frameworkImpl) Extenders() []framework.Extender {
	return f.extenders
}

type frameworkOptions struct {
	componentConfigVersion string
	clientSet              clientset.Interface
	kubeConfig             *restclient.Config
	eventRecorder          events.EventRecorder
	informerFactory        informers.SharedInformerFactory
	snapshotSharedLister   framework.SharedLister
	metricsRecorder        *metricsRecorder
	dguestNominator        framework.DguestNominator
	extenders              []framework.Extender
	captureProfile         CaptureProfile
	clusterEventMap        map[framework.ClusterEvent]sets.String
	parallelizer           parallelize.Parallelizer
}

// Option for the frameworkImpl.
type Option func(*frameworkOptions)

// WithComponentConfigVersion sets the component config version to the
// SchedulerConfiguration version used. The string should be the full
// scheme group/version of the external type we converted from (for example
// "kubescheduler.config.k8s.io/v1beta2")
func WithComponentConfigVersion(componentConfigVersion string) Option {
	return func(o *frameworkOptions) {
		o.componentConfigVersion = componentConfigVersion
	}
}

// WithClientSet sets clientSet for the scheduling frameworkImpl.
func WithClientSet(clientSet clientset.Interface) Option {
	return func(o *frameworkOptions) {
		o.clientSet = clientSet
	}
}

// WithKubeConfig sets kubeConfig for the scheduling frameworkImpl.
func WithKubeConfig(kubeConfig *restclient.Config) Option {
	return func(o *frameworkOptions) {
		o.kubeConfig = kubeConfig
	}
}

// WithEventRecorder sets clientSet for the scheduling frameworkImpl.
func WithEventRecorder(recorder events.EventRecorder) Option {
	return func(o *frameworkOptions) {
		o.eventRecorder = recorder
	}
}

// WithInformerFactory sets informer factory for the scheduling frameworkImpl.
func WithInformerFactory(informerFactory informers.SharedInformerFactory) Option {
	return func(o *frameworkOptions) {
		o.informerFactory = informerFactory
	}
}

// WithSnapshotSharedLister sets the SharedLister of the snapshot.
func WithSnapshotSharedLister(snapshotSharedLister framework.SharedLister) Option {
	return func(o *frameworkOptions) {
		o.snapshotSharedLister = snapshotSharedLister
	}
}

// WithDguestNominator sets dguestNominator for the scheduling frameworkImpl.
func WithDguestNominator(nominator framework.DguestNominator) Option {
	return func(o *frameworkOptions) {
		o.dguestNominator = nominator
	}
}

// WithExtenders sets extenders for the scheduling frameworkImpl.
func WithExtenders(extenders []framework.Extender) Option {
	return func(o *frameworkOptions) {
		o.extenders = extenders
	}
}

// WithParallelism sets parallelism for the scheduling frameworkImpl.
func WithParallelism(parallelism int) Option {
	return func(o *frameworkOptions) {
		o.parallelizer = parallelize.NewParallelizer(parallelism)
	}
}

// CaptureProfile is a callback to capture a finalized profile.
type CaptureProfile func(config.KubeSchedulerProfile)

// WithCaptureProfile sets a callback to capture the finalized profile.
func WithCaptureProfile(c CaptureProfile) Option {
	return func(o *frameworkOptions) {
		o.captureProfile = c
	}
}

func defaultFrameworkOptions(stopCh <-chan struct{}) frameworkOptions {
	return frameworkOptions{
		metricsRecorder: newMetricsRecorder(1000, time.Second, stopCh),
		clusterEventMap: make(map[framework.ClusterEvent]sets.String),
		parallelizer:    parallelize.NewParallelizer(parallelize.DefaultParallelism),
	}
}

// WithClusterEventMap sets clusterEventMap for the scheduling frameworkImpl.
func WithClusterEventMap(m map[framework.ClusterEvent]sets.String) Option {
	return func(o *frameworkOptions) {
		o.clusterEventMap = m
	}
}

var _ framework.Framework = &frameworkImpl{}

// NewFramework initializes plugins given the configuration and the registry.
func NewFramework(r Registry, profile *config.KubeSchedulerProfile, stopCh <-chan struct{}, opts ...Option) (framework.Framework, error) {
	options := defaultFrameworkOptions(stopCh)
	for _, opt := range opts {
		opt(&options)
	}

	f := &frameworkImpl{
		registry:             r,
		snapshotSharedLister: options.snapshotSharedLister,
		scorePluginWeight:    make(map[string]int),
		waitingDguests:       newWaitingDguestsMap(),
		clientSet:            options.clientSet,
		kubeConfig:           options.kubeConfig,
		eventRecorder:        options.eventRecorder,
		informerFactory:      options.informerFactory,
		metricsRecorder:      options.metricsRecorder,
		extenders:            options.extenders,
		DguestNominator:      options.dguestNominator,
		parallelizer:         options.parallelizer,
	}

	if profile == nil {
		return f, nil
	}

	f.profileName = profile.SchedulerName
	if profile.Plugins == nil {
		return f, nil
	}

	// get needed plugins from config
	pg := f.pluginsNeeded(profile.Plugins)

	pluginConfig := make(map[string]runtime.Object, len(profile.PluginConfig))
	for i := range profile.PluginConfig {
		name := profile.PluginConfig[i].Name
		if _, ok := pluginConfig[name]; ok {
			return nil, fmt.Errorf("repeated config for plugin %s", name)
		}
		pluginConfig[name] = profile.PluginConfig[i].Args
	}
	outputProfile := config.KubeSchedulerProfile{
		SchedulerName: f.profileName,
		Plugins:       profile.Plugins,
		PluginConfig:  make([]config.PluginConfig, 0, len(pg)),
	}

	pluginsMap := make(map[string]framework.Plugin)
	for name, factory := range r {
		// initialize only needed plugins.
		if !pg.Has(name) {
			continue
		}

		args := pluginConfig[name]
		if args != nil {
			outputProfile.PluginConfig = append(outputProfile.PluginConfig, config.PluginConfig{
				Name: name,
				Args: args,
			})
		}
		p, err := factory(args, f)
		if err != nil {
			return nil, fmt.Errorf("initializing plugin %q: %w", name, err)
		}
		pluginsMap[name] = p

		// Update ClusterEventMap in place.
		fillEventToPluginMap(p, options.clusterEventMap)
	}

	// initialize plugins per individual extension points
	for _, e := range f.getExtensionPoints(profile.Plugins) {
		if err := updatePluginList(e.slicePtr, *e.plugins, pluginsMap); err != nil {
			return nil, err
		}
	}

	// initialize multiPoint plugins to their expanded extension points
	if len(profile.Plugins.MultiPoint.Enabled) > 0 {
		if err := f.expandMultiPointPlugins(profile, pluginsMap); err != nil {
			return nil, err
		}
	}

	if len(f.queueSortPlugins) != 1 {
		return nil, fmt.Errorf("only one queue sort plugin required for profile with scheduler name %q, but got %d", profile.SchedulerName, len(f.queueSortPlugins))
	}
	if len(f.bindPlugins) == 0 {
		return nil, fmt.Errorf("at least one bind plugin is needed for profile with scheduler name %q", profile.SchedulerName)
	}

	if err := getScoreWeights(f, pluginsMap, append(profile.Plugins.Score.Enabled, profile.Plugins.MultiPoint.Enabled...)); err != nil {
		return nil, err
	}

	// Verifying the score weights again since Plugin.Name() could return a different
	// value from the one used in the configuration.
	for _, scorePlugin := range f.scorePlugins {
		if f.scorePluginWeight[scorePlugin.Name()] == 0 {
			return nil, fmt.Errorf("score plugin %q is not configured with weight", scorePlugin.Name())
		}
	}

	if options.captureProfile != nil {
		if len(outputProfile.PluginConfig) != 0 {
			sort.Slice(outputProfile.PluginConfig, func(i, j int) bool {
				return outputProfile.PluginConfig[i].Name < outputProfile.PluginConfig[j].Name
			})
		} else {
			outputProfile.PluginConfig = nil
		}
		options.captureProfile(outputProfile)
	}

	return f, nil
}

// getScoreWeights makes sure that, between MultiPoint-Score plugin weights and individual Score
// plugin weights there is not an overflow of MaxTotalScore.
func getScoreWeights(f *frameworkImpl, pluginsMap map[string]framework.Plugin, plugins []config.Plugin) error {
	var totalPriority int64
	scorePlugins := reflect.ValueOf(&f.scorePlugins).Elem()
	pluginType := scorePlugins.Type().Elem()
	for _, e := range plugins {
		pg := pluginsMap[e.Name]
		if !reflect.TypeOf(pg).Implements(pluginType) {
			continue
		}

		// We append MultiPoint plugins to the list of Score plugins. So if this plugin has already been
		// encountered, let the individual Score weight take precedence.
		if _, ok := f.scorePluginWeight[e.Name]; ok {
			continue
		}
		// a weight of zero is not permitted, plugins can be disabled explicitly
		// when configured.
		f.scorePluginWeight[e.Name] = int(e.Weight)
		if f.scorePluginWeight[e.Name] == 0 {
			f.scorePluginWeight[e.Name] = 1
		}

		// Checks totalPriority against MaxTotalScore to avoid overflow
		if int64(f.scorePluginWeight[e.Name])*framework.MaxFoodScore > framework.MaxTotalScore-totalPriority {
			return fmt.Errorf("total score of Score plugins could overflow")
		}
		totalPriority += int64(f.scorePluginWeight[e.Name]) * framework.MaxFoodScore
	}
	return nil
}

type orderedSet struct {
	set         map[string]int
	list        []string
	deletionCnt int
}

func newOrderedSet() *orderedSet {
	return &orderedSet{set: make(map[string]int)}
}

func (os *orderedSet) insert(s string) {
	if os.has(s) {
		return
	}
	os.set[s] = len(os.list)
	os.list = append(os.list, s)
}

func (os *orderedSet) has(s string) bool {
	_, found := os.set[s]
	return found
}

func (os *orderedSet) delete(s string) {
	if i, found := os.set[s]; found {
		delete(os.set, s)
		os.list = append(os.list[:i-os.deletionCnt], os.list[i+1-os.deletionCnt:]...)
		os.deletionCnt++
	}
}

func (f *frameworkImpl) expandMultiPointPlugins(profile *config.KubeSchedulerProfile, pluginsMap map[string]framework.Plugin) error {
	// initialize MultiPoint plugins
	for _, e := range f.getExtensionPoints(profile.Plugins) {
		plugins := reflect.ValueOf(e.slicePtr).Elem()
		pluginType := plugins.Type().Elem()
		// build enabledSet of plugins already registered via normal extension points
		// to check double registration
		enabledSet := newOrderedSet()
		for _, plugin := range e.plugins.Enabled {
			enabledSet.insert(plugin.Name)
		}

		disabledSet := sets.NewString()
		for _, disabledPlugin := range e.plugins.Disabled {
			disabledSet.Insert(disabledPlugin.Name)
		}
		if disabledSet.Has("*") {
			klog.V(4).InfoS("all plugins disabled for extension point, skipping MultiPoint expansion", "extension", pluginType)
			continue
		}

		// track plugins enabled via multipoint separately from those enabled by specific extensions,
		// so that we can distinguish between double-registration and explicit overrides
		multiPointEnabled := newOrderedSet()
		overridePlugins := newOrderedSet()
		for _, ep := range profile.Plugins.MultiPoint.Enabled {
			pg, ok := pluginsMap[ep.Name]
			if !ok {
				return fmt.Errorf("%s %q does not exist", pluginType.Name(), ep.Name)
			}

			// if this plugin doesn't implement the type for the current extension we're trying to expand, skip
			if !reflect.TypeOf(pg).Implements(pluginType) {
				continue
			}

			// a plugin that's enabled via MultiPoint can still be disabled for specific extension points
			if disabledSet.Has(ep.Name) {
				klog.V(4).InfoS("plugin disabled for extension point", "plugin", ep.Name, "extension", pluginType)
				continue
			}

			// if this plugin has already been enabled by the specific extension point,
			// the user intent is to override the default plugin or make some other explicit setting.
			// Either way, discard the MultiPoint value for this plugin.
			// This maintains expected behavior for overriding default plugins (see https://github.com/kubernetes/kubernetes/pull/99582)
			if enabledSet.has(ep.Name) {
				overridePlugins.insert(ep.Name)
				klog.InfoS("MultiPoint plugin is explicitly re-configured; overriding", "plugin", ep.Name)
				continue
			}

			// if this plugin is already registered via MultiPoint, then this is
			// a double registration and an error in the config.
			if multiPointEnabled.has(ep.Name) {
				return fmt.Errorf("plugin %q already registered as %q", ep.Name, pluginType.Name())
			}

			// we only need to update the multipoint set, since we already have the specific extension set from above
			multiPointEnabled.insert(ep.Name)
		}

		// Reorder plugins. Here is the expected order:
		// - part 1: overridePlugins. Their order stay intact as how they're specified in regular extension point.
		// - part 2: multiPointEnabled - i.e., plugin defined in multipoint but not in regular extension point.
		// - part 3: other plugins (excluded by part 1 & 2) in regular extension point.
		newPlugins := reflect.New(reflect.TypeOf(e.slicePtr).Elem()).Elem()
		// part 1
		for _, name := range enabledSet.list {
			if overridePlugins.has(name) {
				newPlugins = reflect.Append(newPlugins, reflect.ValueOf(pluginsMap[name]))
				enabledSet.delete(name)
			}
		}
		// part 2
		for _, name := range multiPointEnabled.list {
			newPlugins = reflect.Append(newPlugins, reflect.ValueOf(pluginsMap[name]))
		}
		// part 3
		for _, name := range enabledSet.list {
			newPlugins = reflect.Append(newPlugins, reflect.ValueOf(pluginsMap[name]))
		}
		plugins.Set(newPlugins)
	}
	return nil
}

func fillEventToPluginMap(p framework.Plugin, eventToPlugins map[framework.ClusterEvent]sets.String) {
	ext, ok := p.(framework.EnqueueExtensions)
	if !ok {
		// If interface EnqueueExtensions is not implemented, register the default events
		// to the plugin. This is to ensure backward compatibility.
		registerClusterEvents(p.Name(), eventToPlugins, allClusterEvents)
		return
	}

	events := ext.EventsToRegister()
	// It's rare that a plugin implements EnqueueExtensions but returns nil.
	// We treat it as: the plugin is not interested in any event, and hence dguest failed by that plugin
	// cannot be moved by any regular cluster event.
	if len(events) == 0 {
		klog.InfoS("Plugin's EventsToRegister() returned nil", "plugin", p.Name())
		return
	}
	// The most common case: a plugin implements EnqueueExtensions and returns non-nil result.
	registerClusterEvents(p.Name(), eventToPlugins, events)
}

func registerClusterEvents(name string, eventToPlugins map[framework.ClusterEvent]sets.String, evts []framework.ClusterEvent) {
	for _, evt := range evts {
		if eventToPlugins[evt] == nil {
			eventToPlugins[evt] = sets.NewString(name)
		} else {
			eventToPlugins[evt].Insert(name)
		}
	}
}

func updatePluginList(pluginList interface{}, pluginSet config.PluginSet, pluginsMap map[string]framework.Plugin) error {
	plugins := reflect.ValueOf(pluginList).Elem()
	pluginType := plugins.Type().Elem()
	set := sets.NewString()
	for _, ep := range pluginSet.Enabled {
		pg, ok := pluginsMap[ep.Name]
		if !ok {
			return fmt.Errorf("%s %q does not exist", pluginType.Name(), ep.Name)
		}

		if !reflect.TypeOf(pg).Implements(pluginType) {
			return fmt.Errorf("plugin %q does not extend %s plugin", ep.Name, pluginType.Name())
		}

		if set.Has(ep.Name) {
			return fmt.Errorf("plugin %q already registered as %q", ep.Name, pluginType.Name())
		}

		set.Insert(ep.Name)

		newPlugins := reflect.Append(plugins, reflect.ValueOf(pg))
		plugins.Set(newPlugins)
	}
	return nil
}

// QueueSortFunc returns the function to sort dguests in scheduling queue
func (f *frameworkImpl) QueueSortFunc() framework.LessFunc {
	if f == nil {
		// If frameworkImpl is nil, simply keep their order unchanged.
		// NOTE: this is primarily for tests.
		return func(_, _ *framework.QueuedDguestInfo) bool { return false }
	}

	if len(f.queueSortPlugins) == 0 {
		panic("No QueueSort plugin is registered in the frameworkImpl.")
	}

	// Only one QueueSort plugin can be enabled.
	return f.queueSortPlugins[0].Less
}

// RunPreFilterPlugins runs the set of configured PreFilter plugins. It returns
// *Status and its code is set to non-success if any of the plugins returns
// anything but Success. If a non-success status is returned, then the scheduling
// cycle is aborted.
func (f *frameworkImpl) RunPreFilterPlugins(ctx context.Context, state *framework.CycleState, dguest *v1alpha1.Dguest) (_ *framework.PreFilterResult, status *framework.Status) {
	startTime := time.Now()
	defer func() {
		metrics.FrameworkExtensionPointDuration.WithLabelValues(preFilter, status.Code().String(), f.profileName).Observe(metrics.SinceInSeconds(startTime))
	}()
	var result *framework.PreFilterResult
	var pluginsWithFoods []string
	for _, pl := range f.preFilterPlugins {
		r, s := f.runPreFilterPlugin(ctx, pl, state, dguest)
		if !s.IsSuccess() {
			s.SetFailedPlugin(pl.Name())
			if s.IsUnschedulable() {
				return nil, s
			}
			return nil, framework.AsStatus(fmt.Errorf("running PreFilter plugin %q: %w", pl.Name(), status.AsError())).WithFailedPlugin(pl.Name())
		}
		if !r.AllFoods() {
			pluginsWithFoods = append(pluginsWithFoods, pl.Name())
		}
		result = result.Merge(r)
		if !result.AllFoods() && len(result.FoodNames) == 0 {
			msg := fmt.Sprintf("food(s) didn't satisfy plugin(s) %v simultaneously", pluginsWithFoods)
			if len(pluginsWithFoods) == 1 {
				msg = fmt.Sprintf("food(s) didn't satisfy plugin %v", pluginsWithFoods[0])
			}
			return nil, framework.NewStatus(framework.Unschedulable, msg)
		}

	}
	return result, nil
}

func (f *frameworkImpl) runPreFilterPlugin(ctx context.Context, pl framework.PreFilterPlugin, state *framework.CycleState, dguest *v1alpha1.Dguest) (*framework.PreFilterResult, *framework.Status) {
	if !state.ShouldRecordPluginMetrics() {
		return pl.PreFilter(ctx, state, dguest)
	}
	startTime := time.Now()
	result, status := pl.PreFilter(ctx, state, dguest)
	f.metricsRecorder.observePluginDurationAsync(preFilter, pl.Name(), status, metrics.SinceInSeconds(startTime))
	return result, status
}

// RunPreFilterExtensionAddDguest calls the AddDguest interface for the set of configured
// PreFilter plugins. It returns directly if any of the plugins return any
// status other than Success.
func (f *frameworkImpl) RunPreFilterExtensionAddDguest(
	ctx context.Context,
	state *framework.CycleState,
	dguestToSchedule *v1alpha1.Dguest,
	dguestInfoToAdd *framework.DguestInfo,
	foodInfo *framework.FoodInfo,
) (status *framework.Status) {
	for _, pl := range f.preFilterPlugins {
		if pl.PreFilterExtensions() == nil {
			continue
		}
		status = f.runPreFilterExtensionAddDguest(ctx, pl, state, dguestToSchedule, dguestInfoToAdd, foodInfo)
		if !status.IsSuccess() {
			err := status.AsError()
			klog.ErrorS(err, "Failed running AddDguest on PreFilter plugin", "plugin", pl.Name(), "dguest", klog.KObj(dguestToSchedule))
			return framework.AsStatus(fmt.Errorf("running AddDguest on PreFilter plugin %q: %w", pl.Name(), err))
		}
	}

	return nil
}

func (f *frameworkImpl) runPreFilterExtensionAddDguest(ctx context.Context, pl framework.PreFilterPlugin, state *framework.CycleState, dguestToSchedule *v1alpha1.Dguest, dguestInfoToAdd *framework.DguestInfo, foodInfo *framework.FoodInfo) *framework.Status {
	if !state.ShouldRecordPluginMetrics() {
		return pl.PreFilterExtensions().AddDguest(ctx, state, dguestToSchedule, dguestInfoToAdd, foodInfo)
	}
	startTime := time.Now()
	status := pl.PreFilterExtensions().AddDguest(ctx, state, dguestToSchedule, dguestInfoToAdd, foodInfo)
	f.metricsRecorder.observePluginDurationAsync(preFilterExtensionAddDguest, pl.Name(), status, metrics.SinceInSeconds(startTime))
	return status
}

// RunPreFilterExtensionRemoveDguest calls the RemoveDguest interface for the set of configured
// PreFilter plugins. It returns directly if any of the plugins return any
// status other than Success.
func (f *frameworkImpl) RunPreFilterExtensionRemoveDguest(
	ctx context.Context,
	state *framework.CycleState,
	dguestToSchedule *v1alpha1.Dguest,
	dguestInfoToRemove *framework.DguestInfo,
	foodInfo *framework.FoodInfo,
) (status *framework.Status) {
	for _, pl := range f.preFilterPlugins {
		if pl.PreFilterExtensions() == nil {
			continue
		}
		status = f.runPreFilterExtensionRemoveDguest(ctx, pl, state, dguestToSchedule, dguestInfoToRemove, foodInfo)
		if !status.IsSuccess() {
			err := status.AsError()
			klog.ErrorS(err, "Failed running RemoveDguest on PreFilter plugin", "plugin", pl.Name(), "dguest", klog.KObj(dguestToSchedule))
			return framework.AsStatus(fmt.Errorf("running RemoveDguest on PreFilter plugin %q: %w", pl.Name(), err))
		}
	}

	return nil
}

func (f *frameworkImpl) runPreFilterExtensionRemoveDguest(ctx context.Context, pl framework.PreFilterPlugin, state *framework.CycleState, dguestToSchedule *v1alpha1.Dguest, dguestInfoToRemove *framework.DguestInfo, foodInfo *framework.FoodInfo) *framework.Status {
	if !state.ShouldRecordPluginMetrics() {
		return pl.PreFilterExtensions().RemoveDguest(ctx, state, dguestToSchedule, dguestInfoToRemove, foodInfo)
	}
	startTime := time.Now()
	status := pl.PreFilterExtensions().RemoveDguest(ctx, state, dguestToSchedule, dguestInfoToRemove, foodInfo)
	f.metricsRecorder.observePluginDurationAsync(preFilterExtensionRemoveDguest, pl.Name(), status, metrics.SinceInSeconds(startTime))
	return status
}

// RunFilterPlugins runs the set of configured Filter plugins for dguest on
// the given food. If any of these plugins doesn't return "Success", the
// given food is not suitable for running dguest.
// Meanwhile, the failure message and status are set for the given food.
func (f *frameworkImpl) RunFilterPlugins(
	ctx context.Context,
	state *framework.CycleState,
	dguest *v1alpha1.Dguest,
	foodInfo *framework.FoodInfo,
) framework.PluginToStatus {
	statuses := make(framework.PluginToStatus)
	for _, pl := range f.filterPlugins {
		pluginStatus := f.runFilterPlugin(ctx, pl, state, dguest, foodInfo)
		if !pluginStatus.IsSuccess() {
			if !pluginStatus.IsUnschedulable() {
				// Filter plugins are not supposed to return any status other than
				// Success or Unschedulable.
				errStatus := framework.AsStatus(fmt.Errorf("running %q filter plugin: %w", pl.Name(), pluginStatus.AsError())).WithFailedPlugin(pl.Name())
				return map[string]*framework.Status{pl.Name(): errStatus}
			}
			pluginStatus.SetFailedPlugin(pl.Name())
			statuses[pl.Name()] = pluginStatus
		}
	}

	return statuses
}

func (f *frameworkImpl) runFilterPlugin(ctx context.Context, pl framework.FilterPlugin, state *framework.CycleState, dguest *v1alpha1.Dguest, foodInfo *framework.FoodInfo) *framework.Status {
	if !state.ShouldRecordPluginMetrics() {
		return pl.Filter(ctx, state, dguest, foodInfo)
	}
	startTime := time.Now()
	status := pl.Filter(ctx, state, dguest, foodInfo)
	f.metricsRecorder.observePluginDurationAsync(Filter, pl.Name(), status, metrics.SinceInSeconds(startTime))
	return status
}

// RunPostFilterPlugins runs the set of configured PostFilter plugins until the first
// Success or Error is met, otherwise continues to execute all plugins.
func (f *frameworkImpl) RunPostFilterPlugins(ctx context.Context, state *framework.CycleState, dguest *v1alpha1.Dguest, filteredFoodStatusMap framework.FoodToStatusMap) (_ *framework.PostFilterResult, status *framework.Status) {
	startTime := time.Now()
	defer func() {
		metrics.FrameworkExtensionPointDuration.WithLabelValues(postFilter, status.Code().String(), f.profileName).Observe(metrics.SinceInSeconds(startTime))
	}()

	statuses := make(framework.PluginToStatus)
	// `result` records the last meaningful(non-noop) PostFilterResult.
	var result *framework.PostFilterResult
	for _, pl := range f.postFilterPlugins {
		r, s := f.runPostFilterPlugin(ctx, pl, state, dguest, filteredFoodStatusMap)
		if s.IsSuccess() {
			return r, s
		} else if !s.IsUnschedulable() {
			// Any status other than Success or Unschedulable is Error.
			return nil, framework.AsStatus(s.AsError())
		} else if r != nil && r.Mode() != framework.ModeNoop {
			result = r
		}
		statuses[pl.Name()] = s
	}

	return result, statuses.Merge()
}

func (f *frameworkImpl) runPostFilterPlugin(ctx context.Context, pl framework.PostFilterPlugin, state *framework.CycleState, dguest *v1alpha1.Dguest, filteredFoodStatusMap framework.FoodToStatusMap) (*framework.PostFilterResult, *framework.Status) {
	if !state.ShouldRecordPluginMetrics() {
		return pl.PostFilter(ctx, state, dguest, filteredFoodStatusMap)
	}
	startTime := time.Now()
	r, s := pl.PostFilter(ctx, state, dguest, filteredFoodStatusMap)
	f.metricsRecorder.observePluginDurationAsync(postFilter, pl.Name(), s, metrics.SinceInSeconds(startTime))
	return r, s
}

// RunFilterPluginsWithNominatedDguests runs the set of configured filter plugins
// for nominated dguest on the given food.
// This function is called from two different places: Schedule and Preempt.
// When it is called from Schedule, we want to test whether the dguest is
// schedulable on the food with all the existing dguests on the food plus higher
// and equal priority dguests nominated to run on the food.
// When it is called from Preempt, we should remove the victims of preemption
// and add the nominated dguests. Removal of the victims is done by
// SelectVictimsOnFood(). Preempt removes victims from PreFilter state and
// FoodInfo before calling this function.
func (f *frameworkImpl) RunFilterPluginsWithNominatedDguests(ctx context.Context, state *framework.CycleState, dguest *v1alpha1.Dguest, info *framework.FoodInfo) *framework.Status {
	var status *framework.Status

	dguestsAdded := false
	// We run filters twice in some cases. If the food has greater or equal priority
	// nominated dguests, we run them when those dguests are added to PreFilter state and foodInfo.
	// If all filters succeed in this pass, we run them again when these
	// nominated dguests are not added. This second pass is necessary because some
	// filters such as inter-dguest affinity may not pass without the nominated dguests.
	// If there are no nominated dguests for the food or if the first run of the
	// filters fail, we don't run the second pass.
	// We consider only equal or higher priority dguests in the first pass, because
	// those are the current "dguest" must yield to them and not take a space opened
	// for running them. It is ok if the current "dguest" take resources freed for
	// lower priority dguests.
	// Requiring that the new dguest is schedulable in both circumstances ensures that
	// we are making a conservative decision: filters like resources and inter-dguest
	// anti-affinity are more likely to fail when the nominated dguests are treated
	// as running, while filters like dguest affinity are more likely to fail when
	// the nominated dguests are treated as not running. We can't just assume the
	// nominated dguests are running because they are not running right now and in fact,
	// they may end up getting scheduled to a different food.
	for i := 0; i < 2; i++ {
		stateToUse := state
		foodInfoToUse := info
		if i == 0 {
			var err error
			dguestsAdded, stateToUse, foodInfoToUse, err = addNominatedDguests(ctx, f, dguest, state, info)
			if err != nil {
				return framework.AsStatus(err)
			}
		} else if !dguestsAdded || !status.IsSuccess() {
			break
		}

		statusMap := f.RunFilterPlugins(ctx, stateToUse, dguest, foodInfoToUse)
		status = statusMap.Merge()
		if !status.IsSuccess() && !status.IsUnschedulable() {
			return status
		}
	}

	return status
}

// addNominatedDguests adds dguests with equal or greater priority which are nominated
// to run on the food. It returns 1) whether any dguest was added, 2) augmented cycleState,
// 3) augmented foodInfo.
func addNominatedDguests(ctx context.Context, fh framework.Handle, dguest *v1alpha1.Dguest, state *framework.CycleState, foodInfo *framework.FoodInfo) (bool, *framework.CycleState, *framework.FoodInfo, error) {
	if fh == nil || foodInfo.Food() == nil {
		// This may happen only in tests.
		return false, state, foodInfo, nil
	}
	nominatedDguestInfos := fh.NominatedDguestsForFood(foodInfo.Food().Name)
	if len(nominatedDguestInfos) == 0 {
		return false, state, foodInfo, nil
	}
	foodInfoOut := foodInfo.Clone()
	stateOut := state.Clone()
	dguestsAdded := false
	for _, pi := range nominatedDguestInfos {
		if corev1alpha1.DguestPriority(pi.Dguest) >= corev1alpha1.DguestPriority(dguest) && pi.Dguest.UID != dguest.UID {
			foodInfoOut.AddDguestInfo(pi)
			status := fh.RunPreFilterExtensionAddDguest(ctx, stateOut, dguest, pi, foodInfoOut)
			if !status.IsSuccess() {
				return false, state, foodInfo, status.AsError()
			}
			dguestsAdded = true
		}
	}
	return dguestsAdded, stateOut, foodInfoOut, nil
}

// RunPreScorePlugins runs the set of configured pre-score plugins. If any
// of these plugins returns any status other than "Success", the given dguest is rejected.
func (f *frameworkImpl) RunPreScorePlugins(
	ctx context.Context,
	state *framework.CycleState,
	dguest *v1alpha1.Dguest,
	foods []*v1alpha1.Food,
) (status *framework.Status) {
	startTime := time.Now()
	defer func() {
		metrics.FrameworkExtensionPointDuration.WithLabelValues(preScore, status.Code().String(), f.profileName).Observe(metrics.SinceInSeconds(startTime))
	}()
	for _, pl := range f.preScorePlugins {
		status = f.runPreScorePlugin(ctx, pl, state, dguest, foods)
		if !status.IsSuccess() {
			return framework.AsStatus(fmt.Errorf("running PreScore plugin %q: %w", pl.Name(), status.AsError()))
		}
	}

	return nil
}

func (f *frameworkImpl) runPreScorePlugin(ctx context.Context, pl framework.PreScorePlugin, state *framework.CycleState, dguest *v1alpha1.Dguest, foods []*v1alpha1.Food) *framework.Status {
	if !state.ShouldRecordPluginMetrics() {
		return pl.PreScore(ctx, state, dguest, foods)
	}
	startTime := time.Now()
	status := pl.PreScore(ctx, state, dguest, foods)
	f.metricsRecorder.observePluginDurationAsync(preScore, pl.Name(), status, metrics.SinceInSeconds(startTime))
	return status
}

// RunScorePlugins runs the set of configured scoring plugins. It returns a list that
// stores for each scoring plugin name the corresponding FoodScoreList(s).
// It also returns *Status, which is set to non-success if any of the plugins returns
// a non-success status.
func (f *frameworkImpl) RunScorePlugins(ctx context.Context, state *framework.CycleState, dguest *v1alpha1.Dguest, foods []*v1alpha1.Food) (ps framework.PluginToFoodScores, status *framework.Status) {
	startTime := time.Now()
	defer func() {
		metrics.FrameworkExtensionPointDuration.WithLabelValues(score, status.Code().String(), f.profileName).Observe(metrics.SinceInSeconds(startTime))
	}()
	pluginToFoodScores := make(framework.PluginToFoodScores, len(f.scorePlugins))
	for _, pl := range f.scorePlugins {
		pluginToFoodScores[pl.Name()] = make(framework.FoodScoreList, len(foods))
	}
	ctx, cancel := context.WithCancel(ctx)
	errCh := parallelize.NewErrorChannel()

	// Run Score method for each food in parallel.
	f.Parallelizer().Until(ctx, len(foods), func(index int) {
		for _, pl := range f.scorePlugins {
			foodName := foods[index].Name
			s, status := f.runScorePlugin(ctx, pl, state, dguest, foodName)
			if !status.IsSuccess() {
				err := fmt.Errorf("plugin %q failed with: %w", pl.Name(), status.AsError())
				errCh.SendErrorWithCancel(err, cancel)
				return
			}
			pluginToFoodScores[pl.Name()][index] = framework.FoodScore{
				Name:  foodName,
				Score: s,
			}
		}
	})
	if err := errCh.ReceiveError(); err != nil {
		return nil, framework.AsStatus(fmt.Errorf("running Score plugins: %w", err))
	}

	// Run NormalizeScore method for each ScorePlugin in parallel.
	f.Parallelizer().Until(ctx, len(f.scorePlugins), func(index int) {
		pl := f.scorePlugins[index]
		foodScoreList := pluginToFoodScores[pl.Name()]
		if pl.ScoreExtensions() == nil {
			return
		}
		status := f.runScoreExtension(ctx, pl, state, dguest, foodScoreList)
		if !status.IsSuccess() {
			err := fmt.Errorf("plugin %q failed with: %w", pl.Name(), status.AsError())
			errCh.SendErrorWithCancel(err, cancel)
			return
		}
	})
	if err := errCh.ReceiveError(); err != nil {
		return nil, framework.AsStatus(fmt.Errorf("running Normalize on Score plugins: %w", err))
	}

	// Apply score defaultWeights for each ScorePlugin in parallel.
	f.Parallelizer().Until(ctx, len(f.scorePlugins), func(index int) {
		pl := f.scorePlugins[index]
		// Score plugins' weight has been checked when they are initialized.
		weight := f.scorePluginWeight[pl.Name()]
		foodScoreList := pluginToFoodScores[pl.Name()]

		for i, foodScore := range foodScoreList {
			// return error if score plugin returns invalid score.
			if foodScore.Score > framework.MaxFoodScore || foodScore.Score < framework.MinFoodScore {
				err := fmt.Errorf("plugin %q returns an invalid score %v, it should in the range of [%v, %v] after normalizing", pl.Name(), foodScore.Score, framework.MinFoodScore, framework.MaxFoodScore)
				errCh.SendErrorWithCancel(err, cancel)
				return
			}
			foodScoreList[i].Score = foodScore.Score * int64(weight)
		}
	})
	if err := errCh.ReceiveError(); err != nil {
		return nil, framework.AsStatus(fmt.Errorf("applying score defaultWeights on Score plugins: %w", err))
	}

	return pluginToFoodScores, nil
}

func (f *frameworkImpl) runScorePlugin(ctx context.Context, pl framework.ScorePlugin, state *framework.CycleState, dguest *v1alpha1.Dguest, foodName string) (int64, *framework.Status) {
	if !state.ShouldRecordPluginMetrics() {
		return pl.Score(ctx, state, dguest, foodName)
	}
	startTime := time.Now()
	s, status := pl.Score(ctx, state, dguest, foodName)
	f.metricsRecorder.observePluginDurationAsync(score, pl.Name(), status, metrics.SinceInSeconds(startTime))
	return s, status
}

func (f *frameworkImpl) runScoreExtension(ctx context.Context, pl framework.ScorePlugin, state *framework.CycleState, dguest *v1alpha1.Dguest, foodScoreList framework.FoodScoreList) *framework.Status {
	if !state.ShouldRecordPluginMetrics() {
		return pl.ScoreExtensions().NormalizeScore(ctx, state, dguest, foodScoreList)
	}
	startTime := time.Now()
	status := pl.ScoreExtensions().NormalizeScore(ctx, state, dguest, foodScoreList)
	f.metricsRecorder.observePluginDurationAsync(scoreExtensionNormalize, pl.Name(), status, metrics.SinceInSeconds(startTime))
	return status
}

// RunPreBindPlugins runs the set of configured prebind plugins. It returns a
// failure (bool) if any of the plugins returns an error. It also returns an
// error containing the rejection message or the error occurred in the plugin.
func (f *frameworkImpl) RunPreBindPlugins(ctx context.Context, state *framework.CycleState, dguest *v1alpha1.Dguest, foodName string) (status *framework.Status) {
	startTime := time.Now()
	defer func() {
		metrics.FrameworkExtensionPointDuration.WithLabelValues(preBind, status.Code().String(), f.profileName).Observe(metrics.SinceInSeconds(startTime))
	}()
	for _, pl := range f.preBindPlugins {
		status = f.runPreBindPlugin(ctx, pl, state, dguest, foodName)
		if !status.IsSuccess() {
			err := status.AsError()
			klog.ErrorS(err, "Failed running PreBind plugin", "plugin", pl.Name(), "dguest", klog.KObj(dguest))
			return framework.AsStatus(fmt.Errorf("running PreBind plugin %q: %w", pl.Name(), err))
		}
	}
	return nil
}

func (f *frameworkImpl) runPreBindPlugin(ctx context.Context, pl framework.PreBindPlugin, state *framework.CycleState, dguest *v1alpha1.Dguest, foodName string) *framework.Status {
	if !state.ShouldRecordPluginMetrics() {
		return pl.PreBind(ctx, state, dguest, foodName)
	}
	startTime := time.Now()
	status := pl.PreBind(ctx, state, dguest, foodName)
	f.metricsRecorder.observePluginDurationAsync(preBind, pl.Name(), status, metrics.SinceInSeconds(startTime))
	return status
}

// RunBindPlugins runs the set of configured bind plugins until one returns a non `Skip` status.
func (f *frameworkImpl) RunBindPlugins(ctx context.Context, state *framework.CycleState, dguest *v1alpha1.Dguest, foodName string) (status *framework.Status) {
	startTime := time.Now()
	defer func() {
		metrics.FrameworkExtensionPointDuration.WithLabelValues(bind, status.Code().String(), f.profileName).Observe(metrics.SinceInSeconds(startTime))
	}()
	if len(f.bindPlugins) == 0 {
		return framework.NewStatus(framework.Skip, "")
	}
	for _, bp := range f.bindPlugins {
		status = f.runBindPlugin(ctx, bp, state, dguest, foodName)
		if status.IsSkip() {
			continue
		}
		if !status.IsSuccess() {
			err := status.AsError()
			klog.ErrorS(err, "Failed running Bind plugin", "plugin", bp.Name(), "dguest", klog.KObj(dguest))
			return framework.AsStatus(fmt.Errorf("running Bind plugin %q: %w", bp.Name(), err))
		}
		return status
	}
	return status
}

func (f *frameworkImpl) runBindPlugin(ctx context.Context, bp framework.BindPlugin, state *framework.CycleState, dguest *v1alpha1.Dguest, foodName string) *framework.Status {
	if !state.ShouldRecordPluginMetrics() {
		return bp.Bind(ctx, state, dguest, foodName)
	}
	startTime := time.Now()
	status := bp.Bind(ctx, state, dguest, foodName)
	f.metricsRecorder.observePluginDurationAsync(bind, bp.Name(), status, metrics.SinceInSeconds(startTime))
	return status
}

// RunPostBindPlugins runs the set of configured postbind plugins.
func (f *frameworkImpl) RunPostBindPlugins(ctx context.Context, state *framework.CycleState, dguest *v1alpha1.Dguest, foodName string) {
	startTime := time.Now()
	defer func() {
		metrics.FrameworkExtensionPointDuration.WithLabelValues(postBind, framework.Success.String(), f.profileName).Observe(metrics.SinceInSeconds(startTime))
	}()
	for _, pl := range f.postBindPlugins {
		f.runPostBindPlugin(ctx, pl, state, dguest, foodName)
	}
}

func (f *frameworkImpl) runPostBindPlugin(ctx context.Context, pl framework.PostBindPlugin, state *framework.CycleState, dguest *v1alpha1.Dguest, foodName string) {
	if !state.ShouldRecordPluginMetrics() {
		pl.PostBind(ctx, state, dguest, foodName)
		return
	}
	startTime := time.Now()
	pl.PostBind(ctx, state, dguest, foodName)
	f.metricsRecorder.observePluginDurationAsync(postBind, pl.Name(), nil, metrics.SinceInSeconds(startTime))
}

// RunReservePluginsReserve runs the Reserve method in the set of configured
// reserve plugins. If any of these plugins returns an error, it does not
// continue running the remaining ones and returns the error. In such a case,
// the dguest will not be scheduled and the caller will be expected to call
// RunReservePluginsUnreserve.
func (f *frameworkImpl) RunReservePluginsReserve(ctx context.Context, state *framework.CycleState, dguest *v1alpha1.Dguest, foodName string) (status *framework.Status) {
	startTime := time.Now()
	defer func() {
		metrics.FrameworkExtensionPointDuration.WithLabelValues(reserve, status.Code().String(), f.profileName).Observe(metrics.SinceInSeconds(startTime))
	}()
	for _, pl := range f.reservePlugins {
		status = f.runReservePluginReserve(ctx, pl, state, dguest, foodName)
		if !status.IsSuccess() {
			err := status.AsError()
			klog.ErrorS(err, "Failed running Reserve plugin", "plugin", pl.Name(), "dguest", klog.KObj(dguest))
			return framework.AsStatus(fmt.Errorf("running Reserve plugin %q: %w", pl.Name(), err))
		}
	}
	return nil
}

func (f *frameworkImpl) runReservePluginReserve(ctx context.Context, pl framework.ReservePlugin, state *framework.CycleState, dguest *v1alpha1.Dguest, foodName string) *framework.Status {
	if !state.ShouldRecordPluginMetrics() {
		return pl.Reserve(ctx, state, dguest, foodName)
	}
	startTime := time.Now()
	status := pl.Reserve(ctx, state, dguest, foodName)
	f.metricsRecorder.observePluginDurationAsync(reserve, pl.Name(), status, metrics.SinceInSeconds(startTime))
	return status
}

// RunReservePluginsUnreserve runs the Unreserve method in the set of
// configured reserve plugins.
func (f *frameworkImpl) RunReservePluginsUnreserve(ctx context.Context, state *framework.CycleState, dguest *v1alpha1.Dguest, foodName string) {
	startTime := time.Now()
	defer func() {
		metrics.FrameworkExtensionPointDuration.WithLabelValues(unreserve, framework.Success.String(), f.profileName).Observe(metrics.SinceInSeconds(startTime))
	}()
	// Execute the Unreserve operation of each reserve plugin in the
	// *reverse* order in which the Reserve operation was executed.
	for i := len(f.reservePlugins) - 1; i >= 0; i-- {
		f.runReservePluginUnreserve(ctx, f.reservePlugins[i], state, dguest, foodName)
	}
}

func (f *frameworkImpl) runReservePluginUnreserve(ctx context.Context, pl framework.ReservePlugin, state *framework.CycleState, dguest *v1alpha1.Dguest, foodName string) {
	if !state.ShouldRecordPluginMetrics() {
		pl.Unreserve(ctx, state, dguest, foodName)
		return
	}
	startTime := time.Now()
	pl.Unreserve(ctx, state, dguest, foodName)
	f.metricsRecorder.observePluginDurationAsync(unreserve, pl.Name(), nil, metrics.SinceInSeconds(startTime))
}

// RunPermitPlugins runs the set of configured permit plugins. If any of these
// plugins returns a status other than "Success" or "Wait", it does not continue
// running the remaining plugins and returns an error. Otherwise, if any of the
// plugins returns "Wait", then this function will create and add waiting dguest
// to a map of currently waiting dguests and return status with "Wait" code.
// Dguest will remain waiting dguest for the minimum duration returned by the permit plugins.
func (f *frameworkImpl) RunPermitPlugins(ctx context.Context, state *framework.CycleState, dguest *v1alpha1.Dguest, foodName string) (status *framework.Status) {
	startTime := time.Now()
	defer func() {
		metrics.FrameworkExtensionPointDuration.WithLabelValues(permit, status.Code().String(), f.profileName).Observe(metrics.SinceInSeconds(startTime))
	}()
	pluginsWaitTime := make(map[string]time.Duration)
	statusCode := framework.Success
	for _, pl := range f.permitPlugins {
		status, timeout := f.runPermitPlugin(ctx, pl, state, dguest, foodName)
		if !status.IsSuccess() {
			if status.IsUnschedulable() {
				klog.V(4).InfoS("Dguest rejected by permit plugin", "dguest", klog.KObj(dguest), "plugin", pl.Name(), "status", status.Message())
				status.SetFailedPlugin(pl.Name())
				return status
			}
			if status.IsWait() {
				// Not allowed to be greater than maxTimeout.
				if timeout > maxTimeout {
					timeout = maxTimeout
				}
				pluginsWaitTime[pl.Name()] = timeout
				statusCode = framework.Wait
			} else {
				err := status.AsError()
				klog.ErrorS(err, "Failed running Permit plugin", "plugin", pl.Name(), "dguest", klog.KObj(dguest))
				return framework.AsStatus(fmt.Errorf("running Permit plugin %q: %w", pl.Name(), err)).WithFailedPlugin(pl.Name())
			}
		}
	}
	if statusCode == framework.Wait {
		waitingDguest := newWaitingDguest(dguest, pluginsWaitTime)
		f.waitingDguests.add(waitingDguest)
		msg := fmt.Sprintf("one or more plugins asked to wait and no plugin rejected dguest %q", dguest.Name)
		klog.V(4).InfoS("One or more plugins asked to wait and no plugin rejected dguest", "dguest", klog.KObj(dguest))
		return framework.NewStatus(framework.Wait, msg)
	}
	return nil
}

func (f *frameworkImpl) runPermitPlugin(ctx context.Context, pl framework.PermitPlugin, state *framework.CycleState, dguest *v1alpha1.Dguest, foodName string) (*framework.Status, time.Duration) {
	if !state.ShouldRecordPluginMetrics() {
		return pl.Permit(ctx, state, dguest, foodName)
	}
	startTime := time.Now()
	status, timeout := pl.Permit(ctx, state, dguest, foodName)
	f.metricsRecorder.observePluginDurationAsync(permit, pl.Name(), status, metrics.SinceInSeconds(startTime))
	return status, timeout
}

// WaitOnPermit will block, if the dguest is a waiting dguest, until the waiting dguest is rejected or allowed.
func (f *frameworkImpl) WaitOnPermit(ctx context.Context, dguest *v1alpha1.Dguest) *framework.Status {
	waitingDguest := f.waitingDguests.get(dguest.UID)
	if waitingDguest == nil {
		return nil
	}
	defer f.waitingDguests.remove(dguest.UID)
	klog.V(4).InfoS("Dguest waiting on permit", "dguest", klog.KObj(dguest))

	startTime := time.Now()
	s := <-waitingDguest.s
	metrics.PermitWaitDuration.WithLabelValues(s.Code().String()).Observe(metrics.SinceInSeconds(startTime))

	if !s.IsSuccess() {
		if s.IsUnschedulable() {
			klog.V(4).InfoS("Dguest rejected while waiting on permit", "dguest", klog.KObj(dguest), "status", s.Message())
			s.SetFailedPlugin(s.FailedPlugin())
			return s
		}
		err := s.AsError()
		klog.ErrorS(err, "Failed waiting on permit for dguest", "dguest", klog.KObj(dguest))
		return framework.AsStatus(fmt.Errorf("waiting on permit for dguest: %w", err)).WithFailedPlugin(s.FailedPlugin())
	}
	return nil
}

// SnapshotSharedLister returns the scheduler's SharedLister of the latest FoodInfo
// snapshot. The snapshot is taken at the beginning of a scheduling cycle and remains
// unchanged until a dguest finishes "Reserve". There is no guarantee that the information
// remains unchanged after "Reserve".
func (f *frameworkImpl) SnapshotSharedLister() framework.SharedLister {
	return f.snapshotSharedLister
}

// IterateOverWaitingDguests acquires a read lock and iterates over the WaitingDguests map.
func (f *frameworkImpl) IterateOverWaitingDguests(callback func(framework.WaitingDguest)) {
	f.waitingDguests.iterate(callback)
}

// GetWaitingDguest returns a reference to a WaitingDguest given its UID.
func (f *frameworkImpl) GetWaitingDguest(uid types.UID) framework.WaitingDguest {
	if wp := f.waitingDguests.get(uid); wp != nil {
		return wp
	}
	return nil // Returning nil instead of *waitingDguest(nil).
}

// RejectWaitingDguest rejects a WaitingDguest given its UID.
// The returned value indicates if the given dguest is waiting or not.
func (f *frameworkImpl) RejectWaitingDguest(uid types.UID) bool {
	if waitingDguest := f.waitingDguests.get(uid); waitingDguest != nil {
		waitingDguest.Reject("", "removed")
		return true
	}
	return false
}

// HasFilterPlugins returns true if at least one filter plugin is defined.
func (f *frameworkImpl) HasFilterPlugins() bool {
	return len(f.filterPlugins) > 0
}

// HasPostFilterPlugins returns true if at least one postFilter plugin is defined.
func (f *frameworkImpl) HasPostFilterPlugins() bool {
	return len(f.postFilterPlugins) > 0
}

// HasScorePlugins returns true if at least one score plugin is defined.
func (f *frameworkImpl) HasScorePlugins() bool {
	return len(f.scorePlugins) > 0
}

// ListPlugins returns a map of extension point name to plugin names configured at each extension
// point. Returns nil if no plugins where configured.
func (f *frameworkImpl) ListPlugins() *config.Plugins {
	m := config.Plugins{}

	for _, e := range f.getExtensionPoints(&m) {
		plugins := reflect.ValueOf(e.slicePtr).Elem()
		extName := plugins.Type().Elem().Name()
		var cfgs []config.Plugin
		for i := 0; i < plugins.Len(); i++ {
			name := plugins.Index(i).Interface().(framework.Plugin).Name()
			p := config.Plugin{Name: name}
			if extName == "ScorePlugin" {
				// Weights apply only to score plugins.
				p.Weight = int32(f.scorePluginWeight[name])
			}
			cfgs = append(cfgs, p)
		}
		if len(cfgs) > 0 {
			e.plugins.Enabled = cfgs
		}
	}
	return &m
}

// ClientSet returns a kubernetes clientset.
func (f *frameworkImpl) ClientSet() clientset.Interface {
	return f.clientSet
}

// KubeConfig returns a kubernetes config.
func (f *frameworkImpl) KubeConfig() *restclient.Config {
	return f.kubeConfig
}

// EventRecorder returns an event recorder.
func (f *frameworkImpl) EventRecorder() events.EventRecorder {
	return f.eventRecorder
}

// SharedInformerFactory returns a shared informer factory.
func (f *frameworkImpl) SharedInformerFactory() informers.SharedInformerFactory {
	return f.informerFactory
}

func (f *frameworkImpl) pluginsNeeded(plugins *config.Plugins) sets.String {
	pgSet := sets.String{}

	if plugins == nil {
		return pgSet
	}

	find := func(pgs *config.PluginSet) {
		for _, pg := range pgs.Enabled {
			pgSet.Insert(pg.Name)
		}
	}

	for _, e := range f.getExtensionPoints(plugins) {
		find(e.plugins)
	}
	// Parse MultiPoint separately since they are not returned by f.getExtensionPoints()
	find(&plugins.MultiPoint)

	return pgSet
}

// ProfileName returns the profile name associated to this framework.
func (f *frameworkImpl) ProfileName() string {
	return f.profileName
}

// Parallelizer returns a parallelizer holding parallelism for scheduler.
func (f *frameworkImpl) Parallelizer() parallelize.Parallelizer {
	return f.parallelizer
}
