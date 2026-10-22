// Copyright 2026 Workloom Solutions Private Limited
// SPDX-License-Identifier: MIT

package archtest_test

import (
	"go/ast"
	"go/token"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tetriz-ai/wvet/archtest"
)

// Three toy rules, one per reporter. Each runs on a known-bad tree, where it
// must fail, and on the fixed tree, where it must pass.

// noPrintln: code outside package main does not call fmt.Println.
func noPrintln(t testing.TB, root string) {
	var violations []archtest.Violation
	archtest.WalkGoFilesFullParse(t, root, func(path string, fset *token.FileSet, f *ast.File) {
		if f.Name.Name == "main" {
			return
		}
		ast.Inspect(f, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if ok && sel.Sel.Name == "Println" && archtest.ImportPathForAlias(archtest.SelectorPkg(sel), f.Imports) == "fmt" {
				violations = append(violations, archtest.Violation{
					File:    archtest.RelPath(root, path),
					Line:    fset.Position(call.Pos()).Line,
					Message: "fmt.Println outside package main; return the value or log it",
				})
			}
			return true
		})
	})
	archtest.ReportViolations(t, "no-println", violations, false)
}

// layering: files under domain/ do not import net/http. One legacy file is
// allowlisted.
func layering(t testing.TB, root string) {
	var violations []archtest.Violation
	archtest.WalkGoFiles(t, root, func(path string, _ *token.FileSet, f *ast.File) {
		rel := archtest.RelPath(root, path)
		if !strings.HasPrefix(rel, "domain/") {
			return
		}
		for _, imp := range f.Imports {
			if imp.Path.Value == `"net/http"` {
				violations = append(violations, archtest.Violation{File: rel, Message: "domain imports net/http"})
			}
		}
	})
	archtest.ReportWithFileAllowlist(t, "layering", violations, map[string]bool{"domain/legacy.go": true})
}

// panicRatchet: the number of panic calls per file may shrink, never grow.
func panicRatchet(t testing.TB, root string) {
	var violations []archtest.Violation
	archtest.WalkGoFilesFullParse(t, root, func(path string, fset *token.FileSet, f *ast.File) {
		ast.Inspect(f, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			if ident, isIdent := call.Fun.(*ast.Ident); isIdent && ident.Name == "panic" {
				violations = append(violations, archtest.Violation{
					File: archtest.RelPath(root, path), Line: fset.Position(call.Pos()).Line, Message: "panic",
				})
			}
			return true
		})
	})
	archtest.ReportWithCountAllowlist(t, "panic-ratchet", violations, map[string]int{"parse.go": 1})
}

func TestExampleRules(t *testing.T) {
	rules := []struct {
		name string
		rule func(testing.TB, string)
		want string // a line the failure on the bad tree must name
	}{
		{"noprintln", noPrintln, "store/store.go:10"},
		{"layering", layering, "domain/order.go — domain imports net/http"},
		{"panicratchet", panicRatchet, "parse.go: 2 violation(s) but only 1 grandfathered"},
	}
	for _, r := range rules {
		t.Run(r.name+"/bad fails", func(t *testing.T) {
			rec := &recorder{}
			r.rule(rec, filepath.Join("testdata", r.name, "bad"))
			if !rec.failed {
				t.Fatalf("rule passed on its known-bad tree:\n%s", rec.out.String())
			}
			if !strings.Contains(rec.out.String(), r.want) {
				t.Errorf("failure does not name %q:\n%s", r.want, rec.out.String())
			}
		})
		t.Run(r.name+"/fixed passes", func(t *testing.T) {
			r.rule(t, filepath.Join("testdata", r.name, "fixed"))
		})
	}
}
