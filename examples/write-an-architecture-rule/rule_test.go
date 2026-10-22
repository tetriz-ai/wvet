// Package rule holds three architecture rules written with archtest, one per
// reporter. Each walks the tree named by RULE_TREE under testdata: "fixed"
// when it is unset, where every rule passes.
package rule

import (
	"go/ast"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tetriz-ai/wvet/archtest"
)

func tree(t *testing.T) string {
	name := os.Getenv("RULE_TREE")
	if name == "" {
		name = "fixed"
	}
	return filepath.Join(archtest.ModuleRoot(t), "examples", "write-an-architecture-rule", "testdata", name)
}

// Code outside package main does not call fmt.Println.
func TestNoPrintln(t *testing.T) {
	root := tree(t)
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
					Message: "fmt.Println outside package main",
				})
			}
			return true
		})
	})
	archtest.ReportViolations(t, "no-println", violations, false)
}

// Files under domain/ do not import net/http. One legacy file is allowlisted.
func TestDomainLayering(t *testing.T) {
	root := tree(t)
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
	archtest.ReportWithFileAllowlist(t, "domain-layering", violations, map[string]bool{
		"domain/legacy.go": true,
	})
}

// The number of panic calls per file may shrink, never grow.
func TestPanicRatchet(t *testing.T) {
	root := tree(t)
	var violations []archtest.Violation
	archtest.WalkGoFilesFullParse(t, root, func(path string, fset *token.FileSet, f *ast.File) {
		ast.Inspect(f, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			if fn, ok := call.Fun.(*ast.Ident); ok && fn.Name == "panic" {
				violations = append(violations, archtest.Violation{
					File:    archtest.RelPath(root, path),
					Line:    fset.Position(call.Pos()).Line,
					Message: "panic",
				})
			}
			return true
		})
	})
	archtest.ReportWithCountAllowlist(t, "panic-ratchet", violations, map[string]int{
		"parse/parse.go": 1,
	})
}
