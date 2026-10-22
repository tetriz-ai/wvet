// Copyright 2026 Workloom Solutions Private Limited
// SPDX-License-Identifier: MIT

package waitgroup

import (
	"context"
	"sync"
	"testing"
	"time"
)

// A counter that reaches zero at the deadline wins: the hook runs after ctx is
// done, finishes the WaitGroup and waits for the waiter to see it, so the
// final look at done is deterministic.
func TestFinishAtTheDeadlineWins(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	testHookContextDone = func(done <-chan struct{}) {
		wg.Done()
		<-done
	}
	t.Cleanup(func() { testHookContextDone = nil })
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	if err := SafeWaitContext(ctx, &wg); err != nil {
		t.Errorf("SafeWaitContext() = %v, want nil when the counter reaches zero at the deadline", err)
	}
}
