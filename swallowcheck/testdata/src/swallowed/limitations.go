// Copyright 2026 Workloom Solutions Private Limited
// SPDX-License-Identifier: MIT

package swallowed

import "context"

func cancellationProbe(ctx context.Context) {
	if ctx.Err() != nil { // want "swallowed"
		return
	}
}
