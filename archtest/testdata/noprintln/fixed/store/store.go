// Copyright 2026 Workloom Solutions Private Limited
// SPDX-License-Identifier: MIT

package store

import "fmt"

// Save reports what happened to its caller.
func Save(key string) string {
	return fmt.Sprintf("saved %s", key)
}
