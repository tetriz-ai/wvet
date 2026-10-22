// Copyright 2026 Workloom Solutions Private Limited
// SPDX-License-Identifier: MIT

package store

import "fmt"

// Save prints instead of returning what happened.
func Save(key string) {
	fmt.Println("saved", key)
}
