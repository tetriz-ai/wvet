// Copyright 2026 Workloom Solutions Private Limited
// SPDX-License-Identifier: MIT

// Package order has findings from two checks, in a file and in its test.
package order

import "sync"

// Run starts two goroutines with a Wait between them.
func Run(wg *sync.WaitGroup, work func()) {
	go work()
	wg.Wait()
	go work()
}
