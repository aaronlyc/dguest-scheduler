/*
Copyright 2018 The Kubernetes Authors.

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

package debugger

import (
	"os"
	"os/signal"

	listersv1alpha1 "dguest-scheduler/pkg/generated/listers/scheduler/v1alpha1"
	internalcache "dguest-scheduler/pkg/scheduler/internal/cache"
	internalqueue "dguest-scheduler/pkg/scheduler/internal/queue"
)

// CacheDebugger provides ways to check and write cache information for debugging.
type CacheDebugger struct {
	Comparer CacheComparer
	Dumper   CacheDumper
}

// New creates a CacheDebugger.
func New(
	foodLister listersv1alpha1.FoodLister,
	dguestLister listersv1alpha1.DguestLister,
	cache internalcache.Cache,
	dguestQueue internalqueue.SchedulingQueue,
) *CacheDebugger {
	return &CacheDebugger{
		Comparer: CacheComparer{
			FoodLister:   foodLister,
			DguestLister: dguestLister,
			Cache:        cache,
			DguestQueue:  dguestQueue,
		},
		Dumper: CacheDumper{
			cache:       cache,
			dguestQueue: dguestQueue,
		},
	}
}

// ListenForSignal starts a goroutine that will trigger the CacheDebugger's
// behavior when the process receives SIGINT (Windows) or SIGUSER2 (non-Windows).
func (d *CacheDebugger) ListenForSignal(stopCh <-chan struct{}) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, compareSignal)

	go func() {
		for {
			select {
			case <-stopCh:
				return
			case <-ch:
				d.Comparer.Compare()
				d.Dumper.DumpAll()
			}
		}
	}()
}
