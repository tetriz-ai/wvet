// Copyright 2026 Workloom Solutions Private Limited
// SPDX-License-Identifier: MIT

// Package broken does not type-check; the well-typed wait is still reported.
package broken

import "sync"

type t struct{ wg sync.WaitGroup }

func (x *t) stop() { x.wg.Wait() } // want "sync.WaitGroup.Wait blocks forever"
