// Copyright 2026 Workloom Solutions Private Limited
// SPDX-License-Identifier: MIT

package swallowed

import "testing"

func TestExpectedError(t *testing.T) {
	if _, err := ext(); err == nil { // want "swallowed"
		t.Fatal("expected error")
	}
}
