// Copyright 2026 Workloom Solutions Private Limited
// SPDX-License-Identifier: MIT

// Package suppressed holds one suppressed goroutine.
package suppressed

func testFileWithGoroutine() {
	x := 1
	//goroutinecheck:ignore
	go func() {
		println(x)
	}()
}
