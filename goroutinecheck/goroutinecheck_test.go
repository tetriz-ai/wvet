// Copyright 2026 Workloom Solutions Private Limited
// SPDX-License-Identifier: MIT

package goroutinecheck_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/tetriz-ai/wvet/goroutinecheck"
)

func TestGoroutinecheck(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), goroutinecheck.Analyzer,
		"flagged", "clean", "suppressed", "generated", "broken")
}
