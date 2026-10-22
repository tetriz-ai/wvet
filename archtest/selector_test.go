// Copyright 2026 Workloom Solutions Private Limited
// SPDX-License-Identifier: MIT

package archtest_test

import (
	"go/parser"
	"testing"

	"github.com/tetriz-ai/wvet/archtest"
)

type selectorResult struct {
	refers bool
	pkg    string
	name   string
}

// The three selector helpers agree on every type shape.
func TestSelectorHelpers(t *testing.T) {
	tests := []struct {
		expr string
		want selectorResult
	}{
		{"pkg.T", selectorResult{true, "pkg", "T"}},
		{"*pkg.T", selectorResult{true, "pkg", "T"}},
		{"[]*pkg.T", selectorResult{true, "pkg", "T"}},
		{"[4]pkg.T", selectorResult{true, "pkg", "T"}},
		{"other.T", selectorResult{false, "other", "T"}},
		{"T", selectorResult{false, "", ""}},
		{"string", selectorResult{false, "", ""}},
		{"func() pkg.T", selectorResult{false, "", ""}},
		{"map[string]*pkg.T", selectorResult{true, "pkg", "T"}},
		{"map[pkg.K]string", selectorResult{true, "pkg", "K"}},
		{"chan pkg.T", selectorResult{true, "pkg", "T"}},
		{"[]chan<- *pkg.T", selectorResult{true, "pkg", "T"}},
		{"(pkg.T)", selectorResult{true, "pkg", "T"}},
	}
	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			e, err := parser.ParseExpr(tt.expr)
			if err != nil {
				t.Fatal(err)
			}
			got := selectorResult{archtest.TypeRefersToPackage(e, "pkg"), archtest.SelectorPkg(e), archtest.SelectorTypeName(e)}
			if archtest.TypeRefersToPackage(e, "") {
				t.Errorf("TypeRefersToPackage(%s, \"\") = true, want false: no type is in an unnamed package", tt.expr)
			}
			if got != tt.want {
				t.Errorf("helpers = %+v, want %+v", got, tt.want)
			}
		})
	}
}
