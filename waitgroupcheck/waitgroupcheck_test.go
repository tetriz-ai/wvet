// Copyright 2026 Workloom Solutions Private Limited
// SPDX-License-Identifier: MIT

package waitgroupcheck_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/tetriz-ai/wvet/waitgroupcheck"
)

func TestWaitgroupcheck(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), waitgroupcheck.Analyzer,
		"flagged", "clean", "generated", "broken")
}
