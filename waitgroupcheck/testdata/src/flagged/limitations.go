// Copyright 2026 Workloom Solutions Private Limited
// SPDX-License-Identifier: MIT

package flagged

import "sync"

func goManagedTasks() {
	var wg sync.WaitGroup
	wg.Go(func() {})
	wg.Wait() // want "sync.WaitGroup.Wait blocks forever"
}
