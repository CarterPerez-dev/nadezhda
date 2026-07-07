// ©AngelaMos | 2026
// watch_test.go

package watch

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

type fakeTicker struct{ ch chan time.Time }

func (f *fakeTicker) C() <-chan time.Time { return f.ch }
func (f *fakeTicker) Stop()               {}

type fakeNotifier struct {
	mu    sync.Mutex
	calls int
	last  Report
}

func (f *fakeNotifier) Notify(_ context.Context, r Report) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls++
	f.last = r
	return nil
}

func (f *fakeNotifier) count() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.calls
}

func recvWithin(t *testing.T, ch <-chan int, d time.Duration) int {
	t.Helper()
	select {
	case v := <-ch:
		return v
	case <-time.After(d):
		t.Fatal("timed out waiting for a watch cycle")
		return 0
	}
}

func TestRunAtStartThenTicksThenGracefulShutdown(t *testing.T) {
	ticker := &fakeTicker{ch: make(chan time.Time)}
	ran := make(chan int, 4)
	count := 0
	ctx, cancel := context.WithCancel(context.Background())

	opts := Options{
		Interval:   time.Hour,
		RunAtStart: true,
		NewTicker:  func(time.Duration) Ticker { return ticker },
		Cycle: func(context.Context) (Report, error) {
			count++
			ran <- count
			return Report{}, nil
		},
	}

	done := make(chan error, 1)
	go func() { done <- Run(ctx, opts) }()

	if n := recvWithin(t, ran, 2*time.Second); n != 1 {
		t.Fatalf("RunAtStart cycle = %d, want 1", n)
	}
	ticker.ch <- time.Unix(0, 0)
	if n := recvWithin(t, ran, 2*time.Second); n != 2 {
		t.Fatalf("post-tick cycle = %d, want 2", n)
	}

	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run returned %v, want nil on graceful shutdown", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not return after ctx cancel")
	}
}

func TestRunReturnsNilAndSkipsCycleOnImmediateCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	called := false
	err := Run(ctx, Options{
		Interval:   time.Hour,
		RunAtStart: true,
		NewTicker:  func(time.Duration) Ticker { return &fakeTicker{ch: make(chan time.Time)} },
		Cycle: func(context.Context) (Report, error) {
			called = true
			return Report{}, nil
		},
	})
	if err != nil {
		t.Fatalf("Run = %v, want nil", err)
	}
	if called {
		t.Error("cycle should not run when the context is already cancelled")
	}
}

func TestRunContinuesAfterCycleError(t *testing.T) {
	ticker := &fakeTicker{ch: make(chan time.Time)}
	ran := make(chan int, 4)
	count := 0
	ctx, cancel := context.WithCancel(context.Background())

	err := make(chan error, 1)
	go func() {
		err <- Run(ctx, Options{
			Interval:   time.Hour,
			RunAtStart: true,
			NewTicker:  func(time.Duration) Ticker { return ticker },
			Cycle: func(context.Context) (Report, error) {
				count++
				if count == 1 {
					return Report{}, errors.New("boom")
				}
				ran <- count
				return Report{}, nil
			},
		})
	}()

	ticker.ch <- time.Unix(0, 0)
	if n := recvWithin(t, ran, 2*time.Second); n != 2 {
		t.Fatalf("second cycle = %d, want 2 (daemon survived the first cycle error)", n)
	}
	cancel()
	if e := <-err; e != nil {
		t.Fatalf("Run = %v, want nil", e)
	}
}

func TestOnceNotifiesWhenNotable(t *testing.T) {
	fn := &fakeNotifier{}
	err := Once(context.Background(), Options{
		Interval: time.Hour,
		Cycle: func(context.Context) (Report, error) {
			return Report{Notable: []NotableItem{{Title: "x"}}}, nil
		},
		Notifier: fn,
	})
	if err != nil {
		t.Fatal(err)
	}
	if fn.count() != 1 {
		t.Fatalf("notify calls = %d, want 1", fn.count())
	}
}

func TestOnceSkipsNotifyWhenNoNotable(t *testing.T) {
	fn := &fakeNotifier{}
	err := Once(context.Background(), Options{
		Interval: time.Hour,
		Cycle: func(context.Context) (Report, error) {
			return Report{NewArticles: 3}, nil
		},
		Notifier: fn,
	})
	if err != nil {
		t.Fatal(err)
	}
	if fn.count() != 0 {
		t.Fatalf("notify calls = %d, want 0 (nothing notable)", fn.count())
	}
}

func TestOnceReturnsCycleError(t *testing.T) {
	want := errors.New("cycle failed")
	err := Once(context.Background(), Options{
		Interval: time.Hour,
		Cycle: func(context.Context) (Report, error) {
			return Report{}, want
		},
	})
	if !errors.Is(err, want) {
		t.Fatalf("Once error = %v, want %v", err, want)
	}
}

func TestValidateRejectsBadOptions(t *testing.T) {
	if err := (Options{Interval: time.Hour}).validate(); err == nil {
		t.Error("validate should reject a nil Cycle")
	}
	if err := (Options{Cycle: func(context.Context) (Report, error) { return Report{}, nil }}).validate(); err == nil {
		t.Error("validate should reject a non-positive Interval")
	}
}
