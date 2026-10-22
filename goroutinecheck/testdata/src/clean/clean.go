// Copyright 2026 Workloom Solutions Private Limited
// SPDX-License-Identifier: MIT

// Package clean holds code the analyzer must not report.
package clean

import (
	"context"
	"sync"
	"time"
)

// startedByWaitGroupGo: sync.WaitGroup.Go (Go 1.25) starts the goroutine itself.
func startedByWaitGroupGo() {
	var wg sync.WaitGroup
	wg.Go(work)
}

// startedByALibrary: a goroutine started inside another function is not seen.
func startedByALibrary() {
	time.AfterFunc(time.Second, work)
}

// run stands in for a helper that starts and recovers the goroutine itself.
func run(ctx context.Context, timeout time.Duration, fn func(context.Context)) {}

func throughAHelper() {
	run(context.Background(), 5*time.Second, func(ctx context.Context) {})
}

func work() {}

func ignored() {
	//goroutinecheck:ignore
	go work()
}

func ignoredWithReason() {
	//goroutinecheck:ignore // process-lifetime worker, recovers on its own
	go work()
}

func ignoredWithSpace() {
	// goroutinecheck:ignore
	go work()
}

func reasonOnTheLineAbove() {
	// The worker lives as long as the process.
	//goroutinecheck:ignore
	go work()
}

func bareNolint() {
	//nolint
	go work()
}

func nolintList() {
	//nolint:gosec,goroutinecheck // reason
	go work()
}
