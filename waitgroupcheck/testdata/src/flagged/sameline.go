// Copyright 2026 Workloom Solutions Private Limited
// SPDX-License-Identifier: MIT

package flagged

import "sync"

// This file is not gofmt-formatted on purpose: two statements share a line.

// sameLine: the directive names the first statement on the next line, not the
// one that follows it there.
func sameLine(a, b *sync.WaitGroup) {
	//waitgroupcheck:ignore
	a.Wait(); b.Wait() // want "sync.WaitGroup.Wait blocks forever"
}
