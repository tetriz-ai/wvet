// Copyright 2026 Workloom Solutions Private Limited
// SPDX-License-Identifier: MIT

// Package broken does not type-check; the go statement is still reported.
package broken

func work() {}

func start() { go work() } // want "go statement"
