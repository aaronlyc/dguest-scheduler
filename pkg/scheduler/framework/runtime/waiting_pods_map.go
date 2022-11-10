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
	"fmt"
	"sync"
	"time"

	"dguest-scheduler/pkg/scheduler/framework"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

// waitingDguestsMap a thread-safe map used to maintain dguests waiting in the permit phase.
type waitingDguestsMap struct {
	dguests map[types.UID]*waitingDguest
	mu      sync.RWMutex
}

// newWaitingDguestsMap returns a new waitingDguestsMap.
func newWaitingDguestsMap() *waitingDguestsMap {
	return &waitingDguestsMap{
		dguests: make(map[types.UID]*waitingDguest),
	}
}

// add a new WaitingDguest to the map.
func (m *waitingDguestsMap) add(wp *waitingDguest) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.dguests[wp.GetDguest().UID] = wp
}

// remove a WaitingDguest from the map.
func (m *waitingDguestsMap) remove(uid types.UID) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.dguests, uid)
}

// get a WaitingDguest from the map.
func (m *waitingDguestsMap) get(uid types.UID) *waitingDguest {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.dguests[uid]
}

// iterate acquires a read lock and iterates over the WaitingDguests map.
func (m *waitingDguestsMap) iterate(callback func(framework.WaitingDguest)) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, v := range m.dguests {
		callback(v)
	}
}

// waitingDguest represents a dguest waiting in the permit phase.
type waitingDguest struct {
	dguest         *v1alpha1.Dguest
	pendingPlugins map[string]*time.Timer
	s              chan *framework.Status
	mu             sync.RWMutex
}

var _ framework.WaitingDguest = &waitingDguest{}

// newWaitingDguest returns a new waitingDguest instance.
func newWaitingDguest(dguest *v1alpha1.Dguest, pluginsMaxWaitTime map[string]time.Duration) *waitingDguest {
	wp := &waitingDguest{
		dguest: dguest,
		// Allow() and Reject() calls are non-blocking. This property is guaranteed
		// by using non-blocking send to this channel. This channel has a buffer of size 1
		// to ensure that non-blocking send will not be ignored - possible situation when
		// receiving from this channel happens after non-blocking send.
		s: make(chan *framework.Status, 1),
	}

	wp.pendingPlugins = make(map[string]*time.Timer, len(pluginsMaxWaitTime))
	// The time.AfterFunc calls wp.Reject which iterates through pendingPlugins map. Acquire the
	// lock here so that time.AfterFunc can only execute after newWaitingDguest finishes.
	wp.mu.Lock()
	defer wp.mu.Unlock()
	for k, v := range pluginsMaxWaitTime {
		plugin, waitTime := k, v
		wp.pendingPlugins[plugin] = time.AfterFunc(waitTime, func() {
			msg := fmt.Sprintf("rejected due to timeout after waiting %v at plugin %v",
				waitTime, plugin)
			wp.Reject(plugin, msg)
		})
	}

	return wp
}

// GetDguest returns a reference to the waiting dguest.
func (w *waitingDguest) GetDguest() *v1alpha1.Dguest {
	return w.dguest
}

// GetPendingPlugins returns a list of pending permit plugin's name.
func (w *waitingDguest) GetPendingPlugins() []string {
	w.mu.RLock()
	defer w.mu.RUnlock()
	plugins := make([]string, 0, len(w.pendingPlugins))
	for p := range w.pendingPlugins {
		plugins = append(plugins, p)
	}

	return plugins
}

// Allow declares the waiting dguest is allowed to be scheduled by plugin pluginName.
// If this is the last remaining plugin to allow, then a success signal is delivered
// to unblock the dguest.
func (w *waitingDguest) Allow(pluginName string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if timer, exist := w.pendingPlugins[pluginName]; exist {
		timer.Stop()
		delete(w.pendingPlugins, pluginName)
	}

	// Only signal success status after all plugins have allowed
	if len(w.pendingPlugins) != 0 {
		return
	}

	// The select clause works as a non-blocking send.
	// If there is no receiver, it's a no-op (default case).
	select {
	case w.s <- framework.NewStatus(framework.Success, ""):
	default:
	}
}

// Reject declares the waiting dguest unschedulable.
func (w *waitingDguest) Reject(pluginName, msg string) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	for _, timer := range w.pendingPlugins {
		timer.Stop()
	}

	// The select clause works as a non-blocking send.
	// If there is no receiver, it's a no-op (default case).
	select {
	case w.s <- framework.NewStatus(framework.Unschedulable, msg).WithFailedPlugin(pluginName):
	default:
	}
}
