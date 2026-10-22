// Copyright 2026 Workloom Solutions Private Limited
// SPDX-License-Identifier: MIT

package swallowcheck_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/tetriz-ai/wvet/swallowcheck"
)

func TestSwallowcheck(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), swallowcheck.Analyzer, "swallowed", "handled")
}
