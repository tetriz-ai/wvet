// Copyright 2026 Workloom Solutions Private Limited
// SPDX-License-Identifier: MIT

// Package waitgroup waits for a sync.WaitGroup with a bound, so a missed Done
// releases the caller instead of blocking it forever.
package waitgroup

import (
	"context"
	"errors"
	"sync"
	"time"
)

// ErrTimeout is returned by [SafeWait] when the timeout passes before the
// WaitGroup's counter reaches zero.
var ErrTimeout = errors.New("waitgroup: timed out waiting for the WaitGroup")

// SafeWait waits until wg's counter reaches zero or timeout passes. It returns
// nil in the first case and [ErrTimeout] in the second.
//
// A nil wg is checked first, and panics in the caller's goroutine, as
// wg.Wait() would. After that, a timeout of zero or less returns ErrTimeout at
// once, without starting a goroutine.
//
// After a timeout, one goroutine stays blocked on wg until its counter reaches
// zero. If that never happens, the goroutine is never released.
func SafeWait(wg *sync.WaitGroup, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	err := SafeWaitContext(ctx, wg)
	if errors.Is(err, context.DeadlineExceeded) {
		return ErrTimeout
	}
	return err
}

// testHookContextDone, when set by a test, runs after ctx is done and before
// the final look at done, so a test can finish the WaitGroup at the deadline.
var testHookContextDone func(done <-chan struct{})

// SafeWaitContext waits until wg's counter reaches zero or ctx is done. It
// returns nil in the first case and ctx.Err() in the second.
//
// A nil wg is checked first, and panics in the caller's goroutine, as
// wg.Wait() would. After that, a ctx that is already done returns ctx.Err() at
// once, without starting a goroutine.
//
// After ctx is done, one goroutine stays blocked on wg until its counter
// reaches zero. If that never happens, the goroutine is never released.
func SafeWaitContext(ctx context.Context, wg *sync.WaitGroup) error {
	if wg == nil {
		panic("waitgroup: SafeWaitContext called with a nil *sync.WaitGroup")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	done := make(chan struct{})
	// The goroutine waits on a non-nil wg and closes a channel. Wait panics only
	// when the WaitGroup is misused (reused before an earlier Wait returned), and
	// that panic, like any in a bare goroutine, ends the process.
	//goroutinecheck:ignore // the waiter is the bounded wait itself
	go func() {
		//waitgroupcheck:ignore // this is the bounded wait the check points at
		wg.Wait()
		close(done)
	}()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		if testHookContextDone != nil {
			testHookContextDone(done)
		}
		// The counter may have reached zero at the deadline too; that wins.
		select {
		case <-done:
			return nil
		default:
			return ctx.Err()
		}
	}
}
