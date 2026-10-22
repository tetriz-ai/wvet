// Copyright 2026 Workloom Solutions Private Limited
// SPDX-License-Identifier: MIT

package flagged

import syncalias "sync"

func aliasedImport() {
	var wg syncalias.WaitGroup
	wg.Wait() // want "sync.WaitGroup.Wait blocks forever"
}
