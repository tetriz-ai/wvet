// Copyright 2026 Workloom Solutions Private Limited
// SPDX-License-Identifier: MIT

package archtest_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/tetriz-ai/wvet/archtest"
)

func write(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for name, body := range files {
		p := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(p), 0o750); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}
}

// walked returns the files fn saw, relative to root, sorted.
func walked(t *testing.T, tb testing.TB, root string, full bool) []string {
	t.Helper()
	var got []string
	fn := func(path string, _ *token.FileSet, _ *ast.File) { got = append(got, archtest.RelPath(root, path)) }
	if full {
		archtest.WalkGoFilesFullParse(tb, root, fn)
	} else {
		archtest.WalkGoFiles(tb, root, fn)
	}
	slices.Sort(got)
	return got
}

func TestWalkVisits(t *testing.T) {
	tree := map[string]string{
		"a.go":                   "package a\n",
		"a_test.go":              "package a\n",
		"notes.txt":              "not go\n",
		"sub/b.go":               "package sub\n",
		"testdata/c.go":          "package c\n",
		"vendor/d.go":            "package d\n",
		"node_modules/e.go":      "package e\n",
		".hidden/f.go":           "package f\n",
		"_scratch/g.go":          "package g\n",
		"sub/testdata/deep/h.go": "package h\n",
		"sub/.git/objects/i.go":  "package i\n",
		"sub/_build/j.go":        "package j\n",
		"sub/underscore_ok/k.go": "package k\n",
		"sub/dot.name.ok/l.go":   "package l\n",
	}
	root := t.TempDir()
	write(t, root, tree)
	want := []string{"a.go", "sub/b.go", "sub/dot.name.ok/l.go", "sub/underscore_ok/k.go"}
	for _, full := range []bool{false, true} {
		rec := &recorder{}
		if got := walked(t, rec, root, full); !slices.Equal(got, want) {
			t.Errorf("full=%v: walked %v, want %v", full, got, want)
		}
		if rec.failed {
			t.Errorf("full=%v: walk failed the test: %s", full, rec.out.String())
		}
	}
}

// A walk rooted at a directory whose own name the walkers skip still walks it.
func TestWalkRootIsNeverSkipped(t *testing.T) {
	for _, name := range []string{"testdata", "vendor", ".config", "_work"} {
		t.Run(name, func(t *testing.T) {
			root := filepath.Join(t.TempDir(), name)
			write(t, root, map[string]string{"a.go": "package a\n", "inner/testdata/b.go": "package b\n"})
			for _, full := range []bool{false, true} {
				if got := walked(t, &recorder{}, root, full); !slices.Equal(got, []string{"a.go"}) {
					t.Errorf("full=%v: walked %v under a root named %s, want [a.go]", full, got, name)
				}
			}
		})
	}
}

// A file that does not parse fails the test and names the file; the rest of
// the tree is still walked.
func TestWalkFailsOnUnparseableFile(t *testing.T) {
	tests := []struct {
		name string
		full bool
		src  string
	}{
		{"imports-only walk, broken import block", false, "package a\n\nimport (\n\t\"os\"\n"},
		{"full walk, broken body", true, "package a\n\nfunc broken( {\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			write(t, root, map[string]string{"broken.go": tt.src, "ok.go": "package a\n"})
			rec := &recorder{}
			got := walked(t, rec, root, tt.full)
			if !rec.failed {
				t.Fatalf("walk did not fail the test on broken.go")
			}
			if !strings.Contains(rec.out.String(), "broken.go") {
				t.Errorf("failure %q does not name broken.go", rec.out.String())
			}
			if !slices.Equal(got, []string{"ok.go"}) {
				t.Errorf("walked %v, want [ok.go]", got)
			}
		})
	}
}

func TestWalkMissingDirFails(t *testing.T) {
	rec := &recorder{}
	archtest.WalkGoFiles(rec, filepath.Join(t.TempDir(), "missing"), func(string, *token.FileSet, *ast.File) {})
	if !rec.failed {
		t.Error("walking a missing directory did not fail the test")
	}
}

func TestWalkParseModes(t *testing.T) {
	root := t.TempDir()
	write(t, root, map[string]string{"a.go": "package a\n\nimport \"os\"\n\nfunc F() { _ = os.Args }\n"})
	tests := []struct {
		full      bool
		wantDecls int
	}{{false, 1}, {true, 2}}
	for _, tt := range tests {
		var decls int
		fn := func(_ string, _ *token.FileSet, f *ast.File) { decls = len(f.Decls) }
		if tt.full {
			archtest.WalkGoFilesFullParse(t, root, fn)
		} else {
			archtest.WalkGoFiles(t, root, fn)
		}
		if decls != tt.wantDecls {
			t.Errorf("full=%v: %d declarations parsed, want %d", tt.full, decls, tt.wantDecls)
		}
	}
}

func TestRelPath(t *testing.T) {
	root := filepath.Join(string(filepath.Separator)+"repo", "mod")
	tests := []struct {
		path string
		want string
	}{
		{filepath.Join(root, "pkg", "file.go"), "pkg/file.go"},
		{root, "."},
		{"relative/file.go", "relative/file.go"},
	}
	for _, tt := range tests {
		if got := archtest.RelPath(root, tt.path); got != tt.want {
			t.Errorf("RelPath(%q, %q) = %q, want %q", root, tt.path, got, tt.want)
		}
	}
}

func TestModuleRoot(t *testing.T) {
	root := t.TempDir()
	write(t, root, map[string]string{"go.mod": "module example.com/m\n", "a/b/c/x.txt": ""})
	t.Chdir(filepath.Join(root, "a", "b", "c"))
	got, err := filepath.EvalSymlinks(archtest.ModuleRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	want, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Errorf("ModuleRoot() = %q, want %q", got, want)
	}
}

func TestModuleRootNoGoMod(t *testing.T) {
	t.Chdir(t.TempDir())
	if _, err := os.Stat("/go.mod"); err == nil { //swallowcheck:ignore // probe: only presence matters
		t.Skip("the filesystem root holds a go.mod")
	}
	rec := &recorder{}
	if got := archtest.ModuleRoot(rec); got != "" || !rec.failed {
		t.Errorf("ModuleRoot() = %q, failed = %v; want \"\" and a failed test", got, rec.failed)
	}
}

func TestViolationString(t *testing.T) {
	tests := []struct {
		v    archtest.Violation
		want string
	}{
		{archtest.Violation{File: "a.go", Line: 3, Message: "m"}, "  a.go:3 — m"},
		{archtest.Violation{File: "a.go", Message: "whole file"}, "  a.go — whole file"},
	}
	for _, tt := range tests {
		if got := tt.v.String(); got != tt.want {
			t.Errorf("String() = %q, want %q", got, tt.want)
		}
	}
}

var unordered = []archtest.Violation{
	{File: "a.go", Line: 9, Message: "z second on its line"},
	{File: "c.go", Line: 2, Message: "m"},
	{File: "a.go", Line: 9, Message: "m"},
	{File: "b.go", Line: 1, Message: "m"},
	{File: "a.go", Line: 1, Message: "m"},
}

func TestReportViolations(t *testing.T) {
	tests := []struct {
		name        string
		violations  []archtest.Violation
		known       bool
		wantFailed  bool
		wantSkipped bool
	}{
		{"none", nil, false, false, false},
		{"some", unordered, false, true, false},
		{"some, known", unordered, true, false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := &recorder{}
			archtest.ReportViolations(rec, "RULE", tt.violations, tt.known)
			if rec.failed != tt.wantFailed || rec.skipped != tt.wantSkipped {
				t.Errorf("failed=%v skipped=%v, want %v %v\n%s", rec.failed, rec.skipped, tt.wantFailed, tt.wantSkipped, rec.out.String())
			}
		})
	}
}

// Reporter output is sorted by file, then line, whatever the input order.
func TestReportersSortOutput(t *testing.T) {
	wantOrder := []string{"a.go:1", "a.go:9 — m", "a.go:9 — z second on its line", "b.go:1", "c.go:2"}
	reporters := map[string]func(testing.TB){
		"ReportViolations":         func(tb testing.TB) { archtest.ReportViolations(tb, "R", unordered, false) },
		"ReportWithFileAllowlist":  func(tb testing.TB) { archtest.ReportWithFileAllowlist(tb, "R", unordered, nil) },
		"ReportWithCountAllowlist": func(tb testing.TB) { archtest.ReportWithCountAllowlist(tb, "R", unordered, nil) },
	}
	for name, report := range reporters {
		t.Run(name, func(t *testing.T) {
			for range 20 {
				rec := &recorder{}
				report(rec)
				out := rec.out.String()
				last := -1
				for _, w := range wantOrder {
					i := strings.Index(out, w)
					if i <= last {
						t.Fatalf("output not sorted by file and line:\n%s", out)
					}
					last = i
				}
			}
		})
	}
}

func TestReportWithFileAllowlist(t *testing.T) {
	tests := []struct {
		name       string
		violations []archtest.Violation
		allowlist  map[string]bool
		wantFailed bool
		wantLog    string
	}{
		{"none", nil, nil, false, ""},
		{"new file fails", []archtest.Violation{{File: "new.go", Line: 1, Message: "m"}}, map[string]bool{"old.go": true}, true, "NEW violation(s)"},
		{"allowlisted file warns", []archtest.Violation{{File: "old.go", Line: 1, Message: "m"}}, map[string]bool{"old.go": true}, false, "(WARN)"},
		{"stale entry is a hint", nil, map[string]bool{"fixed.go": true}, false, "fixed.go has no violations"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := &recorder{}
			archtest.ReportWithFileAllowlist(rec, "RULE", tt.violations, tt.allowlist)
			if rec.failed != tt.wantFailed {
				t.Errorf("failed = %v, want %v\n%s", rec.failed, tt.wantFailed, rec.out.String())
			}
			if !strings.Contains(rec.out.String(), tt.wantLog) {
				t.Errorf("output %q does not contain %q", rec.out.String(), tt.wantLog)
			}
		})
	}
}

func TestReportWithCountAllowlist(t *testing.T) {
	two := []archtest.Violation{{File: "hot.go", Line: 1, Message: "m"}, {File: "hot.go", Line: 5, Message: "m"}}
	tests := []struct {
		name       string
		violations []archtest.Violation
		allowlist  map[string]int
		wantFailed bool
		wantLog    string
	}{
		{"none", nil, nil, false, ""},
		{"unlisted file fails", two, map[string]int{"other.go": 2}, true, "hot.go:1"},
		{"more than the count fails", two, map[string]int{"hot.go": 1}, true, "only 1 grandfathered"},
		{"listed with count zero fails", two, map[string]int{"hot.go": 0}, true, "only 0 grandfathered"},
		{"exactly the count warns", two, map[string]int{"hot.go": 2}, false, "(WARN)"},
		{"fewer than the count asks to lower it", two, map[string]int{"hot.go": 3}, false, "lower the count"},
		{"stale entry is a hint", nil, map[string]int{"fixed.go": 1}, false, "fixed.go has no violations"},
		// Known limit: the ratchet counts, so fixing one and adding one in the same file passes.
		{"fix one, add one nets zero", []archtest.Violation{{File: "hot.go", Line: 1, Message: "m"}, {File: "hot.go", Line: 40, Message: "new"}}, map[string]int{"hot.go": 2}, false, "(WARN)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := &recorder{}
			archtest.ReportWithCountAllowlist(rec, "RULE", tt.violations, tt.allowlist)
			if rec.failed != tt.wantFailed {
				t.Errorf("failed = %v, want %v\n%s", rec.failed, tt.wantFailed, rec.out.String())
			}
			if !strings.Contains(rec.out.String(), tt.wantLog) {
				t.Errorf("output %q does not contain %q", rec.out.String(), tt.wantLog)
			}
		})
	}
}

func TestDefaultImportName(t *testing.T) {
	tests := []struct{ path, want string }{
		{"os", "os"},
		{"path/filepath", "filepath"},
		{"gopkg.in/yaml.v3", "yaml"},
		{"example.com/lib/v2", "lib"},
		{"example.com/go-cmp", "cmp"},
		{"example.com/go-lib/v3", "lib"},
		{"v2", "v2"},
		{"example.com/v1x", "v1x"},
		{"example.com/name.vx", "name.vx"},
		{"example.com/lib/v", "v"},
		{"gopkg.in/yaml.v", "yaml.v"},
	}
	for _, tt := range tests {
		if got := archtest.DefaultImportName(tt.path); got != tt.want {
			t.Errorf("DefaultImportName(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

func TestImportPathForAlias(t *testing.T) {
	f, err := parser.ParseFile(token.NewFileSet(), "x.go", `package p

import (
	"os"
	"gopkg.in/yaml.v3"
	"example.com/lib/v2"
	str "strings"
	`+"`flag`"+`
)
`, parser.ImportsOnly)
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct{ alias, want string }{
		{"os", "os"},
		{"yaml", "gopkg.in/yaml.v3"},
		{"lib", "example.com/lib/v2"},
		{"str", "strings"},
		{"flag", "flag"},
		{"strings", ""},
		{"v2", ""},
		{"missing", ""},
	}
	for _, tt := range tests {
		if got := archtest.ImportPathForAlias(tt.alias, f.Imports); got != tt.want {
			t.Errorf("ImportPathForAlias(%q) = %q, want %q", tt.alias, got, tt.want)
		}
	}
}

func TestIsEmptyInterfaceType(t *testing.T) {
	tests := []struct {
		expr string
		want bool
	}{
		{"any", true}, {"interface{}", true}, {"interface{ M() }", false}, {"error", false}, {"map[string]any", false},
	}
	for _, tt := range tests {
		e, err := parser.ParseExpr(tt.expr)
		if err != nil {
			t.Fatal(err)
		}
		if got := archtest.IsEmptyInterfaceType(e); got != tt.want {
			t.Errorf("IsEmptyInterfaceType(%s) = %v, want %v", tt.expr, got, tt.want)
		}
	}
}
