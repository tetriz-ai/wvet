// Copyright 2026 Workloom Solutions Private Limited
// SPDX-License-Identifier: MIT

// Package constraints holds a constraint used from another package.
package constraints

import "sync"

// OnlyWaitGroup allows only *sync.WaitGroup.
type OnlyWaitGroup interface {
	*sync.WaitGroup
	Wait()
}
