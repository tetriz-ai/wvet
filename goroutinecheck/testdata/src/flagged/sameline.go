// Copyright 2026 Workloom Solutions Private Limited
// SPDX-License-Identifier: MIT

package flagged

// This file is not gofmt-formatted on purpose: two statements share a line.

// secondOnTheLine: the directive names the first go statement, not the second.
func secondOnTheLine() {
	//goroutinecheck:ignore
	go work(); go work() // want "go statement"
}
