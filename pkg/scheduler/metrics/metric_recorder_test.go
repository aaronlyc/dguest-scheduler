package metrics

import (
	"sync"
	"sync/atomic"
	"testing"
)

var _ MetricRecorder = &fakeDguestsRecorder{}

type fakeDguestsRecorder struct {
	counter int64
}

func (r *fakeDguestsRecorder) Inc() {
	atomic.AddInt64(&r.counter, 1)
}

func (r *fakeDguestsRecorder) Dec() {
	atomic.AddInt64(&r.counter, -1)
}

func (r *fakeDguestsRecorder) Clear() {
	atomic.StoreInt64(&r.counter, 0)
}

func TestInc(t *testing.T) {
	fakeRecorder := fakeDguestsRecorder{}
	var wg sync.WaitGroup
	loops := 100
	wg.Add(loops)
	for i := 0; i < loops; i++ {
		go func() {
			fakeRecorder.Inc()
			wg.Done()
		}()
	}
	wg.Wait()
	if fakeRecorder.counter != int64(loops) {
		t.Errorf("Expected %v, got %v", loops, fakeRecorder.counter)
	}
}

func TestDec(t *testing.T) {
	fakeRecorder := fakeDguestsRecorder{counter: 100}
	var wg sync.WaitGroup
	loops := 100
	wg.Add(loops)
	for i := 0; i < loops; i++ {
		go func() {
			fakeRecorder.Dec()
			wg.Done()
		}()
	}
	wg.Wait()
	if fakeRecorder.counter != int64(0) {
		t.Errorf("Expected %v, got %v", loops, fakeRecorder.counter)
	}
}

func TestClear(t *testing.T) {
	fakeRecorder := fakeDguestsRecorder{}
	var wg sync.WaitGroup
	incLoops, decLoops := 100, 80
	wg.Add(incLoops + decLoops)
	for i := 0; i < incLoops; i++ {
		go func() {
			fakeRecorder.Inc()
			wg.Done()
		}()
	}
	for i := 0; i < decLoops; i++ {
		go func() {
			fakeRecorder.Dec()
			wg.Done()
		}()
	}
	wg.Wait()
	if fakeRecorder.counter != int64(incLoops-decLoops) {
		t.Errorf("Expected %v, got %v", incLoops-decLoops, fakeRecorder.counter)
	}
	// verify Clear() works
	fakeRecorder.Clear()
	if fakeRecorder.counter != int64(0) {
		t.Errorf("Expected %v, got %v", 0, fakeRecorder.counter)
	}
}
