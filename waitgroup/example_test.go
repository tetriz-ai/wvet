// Copyright 2026 Workloom Solutions Private Limited
// SPDX-License-Identifier: MIT

package waitgroup_test

import (
	"fmt"
	"sync"
	"time"

	"github.com/tetriz-ai/wvet/waitgroup"
)

func ExampleSafeWait() {
	var wg sync.WaitGroup
	wg.Add(1) // the matching Done is never called

	err := waitgroup.SafeWait(&wg, 50*time.Millisecond)
	fmt.Println(err)
	// Output: waitgroup: timed out waiting for the WaitGroup
}
