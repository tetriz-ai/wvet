// Copyright 2026 Workloom Solutions Private Limited
// SPDX-License-Identifier: MIT

package waitgroup_test

import (
	"context"
	"errors"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/tetriz-ai/wvet/waitgroup"
)

// stuck returns a WaitGroup whose counter does not reach zero until the test
// ends, so the waiting goroutine is released at cleanup.
func stuck(t *testing.T) *sync.WaitGroup {
	wg := &sync.WaitGroup{}
	wg.Add(1)
	t.Cleanup(wg.Done)
	return wg
}

func TestSafeWait(t *testing.T) {
	tests := []struct {
		name    string
		wg      func(t *testing.T) *sync.WaitGroup
		timeout time.Duration
		want    error
	}{
		{"already zero", func(*testing.T) *sync.WaitGroup { return &sync.WaitGroup{} }, time.Second, nil},
		{"reaches zero before the timeout", func(*testing.T) *sync.WaitGroup {
			wg := &sync.WaitGroup{}
			wg.Add(1)
			time.AfterFunc(10*time.Millisecond, wg.Done)
			return wg
		}, 5 * time.Second, nil},
		{"never reaches zero", stuck, 20 * time.Millisecond, waitgroup.ErrTimeout},
		{"zero timeout, counter at zero", func(*testing.T) *sync.WaitGroup { return &sync.WaitGroup{} }, 0, waitgroup.ErrTimeout},
		{"zero timeout, counter above zero", stuck, 0, waitgroup.ErrTimeout},
		{"negative timeout, counter above zero", stuck, -time.Second, waitgroup.ErrTimeout},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := waitgroup.SafeWait(tt.wg(t), tt.timeout); !errors.Is(got, tt.want) || (got == nil) != (tt.want == nil) {
				t.Errorf("SafeWait() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSafeWaitContext(t *testing.T) {
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	tests := []struct {
		name string
		ctx  context.Context
		wg   func(t *testing.T) *sync.WaitGroup
		want error
	}{
		{"canceled, counter above zero", canceled, stuck, context.Canceled},
		{"canceled, counter at zero", canceled, func(*testing.T) *sync.WaitGroup { return &sync.WaitGroup{} }, context.Canceled},
		{"canceled while waiting", func() context.Context {
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
			t.Cleanup(cancel)
			return ctx
		}(), stuck, context.DeadlineExceeded},
		{"live context, counter at zero", context.Background(), func(*testing.T) *sync.WaitGroup { return &sync.WaitGroup{} }, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := waitgroup.SafeWaitContext(tt.ctx, tt.wg(t)); !errors.Is(got, tt.want) || (got == nil) != (tt.want == nil) {
				t.Errorf("SafeWaitContext() = %v, want %v", got, tt.want)
			}
		})
	}
}

// A wait that is over before it starts leaves no goroutine blocked on wg.
func TestNoWaiterWhenAlreadyOver(t *testing.T) {
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	tests := []struct {
		name string
		wait func(*sync.WaitGroup) error
		want error
	}{
		{"SafeWait, zero timeout", func(wg *sync.WaitGroup) error { return waitgroup.SafeWait(wg, 0) }, waitgroup.ErrTimeout},
		{"SafeWait, negative timeout", func(wg *sync.WaitGroup) error { return waitgroup.SafeWait(wg, -time.Second) }, waitgroup.ErrTimeout},
		{"SafeWaitContext, canceled", func(wg *sync.WaitGroup) error { return waitgroup.SafeWaitContext(canceled, wg) }, context.Canceled},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			awaitWaiters(t, 0, "left by earlier tests")
			wg := stuck(t)
			if err := tt.wait(wg); !errors.Is(err, tt.want) {
				t.Fatalf("wait returned %v, want %v", err, tt.want)
			}
			if n := waiters(); n != 0 {
				t.Errorf("waiters after the wait = %d, want 0 (no new goroutine)", n)
			}
		})
	}
}

// waiters counts the goroutines SafeWaitContext has started that are still
// running. It reads the stacks rather than runtime.NumGoroutine, because the
// waiters earlier tests released at cleanup may still be exiting.
func waiters() int {
	buf := make([]byte, 1<<20)
	for {
		n := runtime.Stack(buf, true)
		if n < len(buf) {
			return strings.Count(string(buf[:n]), "created by github.com/tetriz-ai/wvet/waitgroup.SafeWaitContext")
		}
		buf = make([]byte, 2*len(buf))
	}
}

// awaitWaiters polls until want waiters are running, or fails after 5s.
func awaitWaiters(t *testing.T, want int, what string) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for waiters() != want && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if n := waiters(); n != want {
		t.Fatalf("waiters %s = %d, want %d", what, n, want)
	}
}

// After a timeout, one goroutine stays blocked on wg until its counter reaches zero.
func TestTimeoutLeavesOneWaiter(t *testing.T) {
	awaitWaiters(t, 0, "left by earlier tests")
	wg := &sync.WaitGroup{}
	wg.Add(1)
	if err := waitgroup.SafeWait(wg, 10*time.Millisecond); !errors.Is(err, waitgroup.ErrTimeout) {
		t.Fatalf("SafeWait() = %v, want ErrTimeout", err)
	}
	if n := waiters(); n != 1 {
		t.Errorf("waiters after the timeout = %d, want 1", n)
	}
	wg.Done()
	awaitWaiters(t, 0, "after Done")
}

// A nil WaitGroup panics in the caller's goroutine, where recover sees it.
func TestNilWaitGroupPanicsInCaller(t *testing.T) {
	tests := []struct {
		name string
		wait func()
	}{
		{"SafeWait", func() { _ = waitgroup.SafeWait(nil, time.Second) }},
		{"SafeWaitContext", func() { _ = waitgroup.SafeWaitContext(context.Background(), nil) }},
		{"SafeWait, zero timeout", func() { _ = waitgroup.SafeWait(nil, 0) }},
		{"SafeWaitContext, canceled", func() {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			_ = waitgroup.SafeWaitContext(ctx, nil)
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recovered := func() (v any) {
				defer func() { v = recover() }()
				tt.wait()
				return nil
			}()
			if recovered == nil {
				t.Fatal("no panic reached the caller")
			}
			if s, _ := recovered.(string); !strings.Contains(s, "nil *sync.WaitGroup") {
				t.Errorf("panic = %v, want the nil WaitGroup message", recovered)
			}
		})
	}
}
