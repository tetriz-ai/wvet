//go:build e2e

// Copyright 2026 Workloom Solutions Private Limited
// SPDX-License-Identifier: MIT

package tagged

import "sync"

// Drain waits with no bound, in the e2e build only.
func Drain(wg *sync.WaitGroup) {
	wg.Wait()
}
