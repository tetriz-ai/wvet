// Copyright 2026 Workloom Solutions Private Limited
// SPDX-License-Identifier: MIT

package archtest_test

import (
	"fmt"
	"strings"
	"testing"
)

// recorder is a testing.TB that records instead of failing, so a test can
// assert what a reporter or walker did to the test it was given.
type recorder struct {
	testing.TB
	failed  bool
	skipped bool
	out     strings.Builder
}

func (r *recorder) Helper()                    {}
func (r *recorder) Log(args ...any)            { fmt.Fprintln(&r.out, args...) }
func (r *recorder) Logf(f string, args ...any) { fmt.Fprintf(&r.out, f+"\n", args...) }
func (r *recorder) Errorf(f string, args ...any) {
	r.failed = true
	fmt.Fprintf(&r.out, f+"\n", args...)
}
func (r *recorder) Fatal(args ...any) { r.failed = true; fmt.Fprintln(&r.out, args...) }
func (r *recorder) Fatalf(f string, args ...any) {
	r.failed = true
	fmt.Fprintf(&r.out, f+"\n", args...)
}
func (r *recorder) Skipf(f string, args ...any) {
	r.skipped = true
	fmt.Fprintf(&r.out, f+"\n", args...)
}
