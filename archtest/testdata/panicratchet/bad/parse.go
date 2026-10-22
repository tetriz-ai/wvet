// Copyright 2026 Workloom Solutions Private Limited
// SPDX-License-Identifier: MIT

package parse

// MustInt panics on bad input; one panic is grandfathered.
func MustInt(s string) int {
	if s == "" {
		panic("empty")
	}
	if s[0] == '-' {
		panic("negative") // a second panic, added after the allowlist was written
	}
	return len(s)
}
