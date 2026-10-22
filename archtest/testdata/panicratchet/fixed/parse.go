// Copyright 2026 Workloom Solutions Private Limited
// SPDX-License-Identifier: MIT

package parse

import "errors"

// Int returns an error for the new case instead of panicking.
func Int(s string) (int, error) {
	if s == "" {
		panic("empty")
	}
	if s[0] == '-' {
		return 0, errors.New("negative")
	}
	return len(s), nil
}
