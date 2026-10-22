// Copyright 2026 Workloom Solutions Private Limited
// SPDX-License-Identifier: MIT

// Package broken does not type-check.
package broken

var _ int = "not an int"

// Start starts a goroutine that does not recover.
func Start(work func()) {
	go work()
}
