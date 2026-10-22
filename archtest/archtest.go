// Copyright 2026 Workloom Solutions Private Limited
// SPDX-License-Identifier: MIT

// Package archtest holds helpers for architecture tests: rules about a code
// base's structure, written as ordinary Go tests. A rule walks a tree of Go
// files, collects a [Violation] for each place that breaks it, and reports them
// with one of three reporters:
//
//   - [ReportViolations] fails on any violation, or logs and skips while the
//     rule has known violations.
//   - [ReportWithFileAllowlist] fails on a violation in a file the allowlist
//     does not name, and logs the rest.
//   - [ReportWithCountAllowlist] is a ratchet: each listed file may keep at most
//     its listed number of violations, so the count can shrink but never grow.
//
// The walkers fail the test on a file they cannot parse, so a rule never passes
// on code it did not read.
package archtest

import (
	"cmp"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// Violation is one place where a rule is broken.
type Violation struct {
	// File is the file's path, conventionally relative to the module root
	// with forward slashes, as [RelPath] returns it. Allowlists are keyed by it.
	File string
	// Line is the 1-based line, or 0 when the violation is about the whole file.
	Line    int
	Message string
}

// String formats v as "  file:line — message", or "  file — message" when
// Line is 0.
func (v Violation) String() string {
	if v.Line > 0 {
		return fmt.Sprintf("  %s:%d — %s", v.File, v.Line, v.Message)
	}
	return fmt.Sprintf("  %s — %s", v.File, v.Message)
}

// ModuleRoot returns the nearest directory at or above the working directory
// that holds a go.mod. go test runs a test in its package directory, so this is
// the root of the module under test. It fails the test when there is none.
func ModuleRoot(t testing.TB) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd: %v", err)
	}
	for {
		// A Stat error means "no go.mod here, keep walking up"; the case
		// where none exists is the t.Fatal below.
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil { //swallowcheck:ignore // probe: absence is the walk-up signal
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find go.mod in the working directory or any parent")
			return ""
		}
		dir = parent
	}
}

// RelPath returns path relative to root with forward slashes, the form
// allowlists are written in on every operating system. When path cannot be made
// relative to root it returns path unchanged.
func RelPath(root, path string) string {
	rel, err := filepath.Rel(root, path) //swallowcheck:ignore // documented fallback: the path unchanged
	if err != nil {
		return path
	}
	return filepath.ToSlash(rel)
}

// SkipDir reports whether the walkers skip a directory with this name below
// the walk root. It follows the go tool: names that start with "." or "_",
// and testdata; plus vendor and node_modules.
func SkipDir(name string) bool {
	switch name {
	case "testdata", "vendor", "node_modules":
		return true
	}
	return strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_")
}

// WalkGoFiles walks dir and calls fn for every .go file that is not a
// _test.go file, parsed up to its imports (parser.ImportsOnly). Directories
// for which [SkipDir] is true are skipped, except dir itself. A file that does
// not parse fails the test with t.Errorf naming it, and the walk goes on.
func WalkGoFiles(t testing.TB, dir string, fn func(path string, fset *token.FileSet, file *ast.File)) {
	t.Helper()
	walk(t, dir, parser.ImportsOnly, fn)
}

// WalkGoFilesFullParse is [WalkGoFiles] with the whole file parsed, function
// bodies included, for rules that inspect calls, selectors or declarations.
func WalkGoFilesFullParse(t testing.TB, dir string, fn func(path string, fset *token.FileSet, file *ast.File)) {
	t.Helper()
	walk(t, dir, 0, fn)
}

func walk(t testing.TB, dir string, mode parser.Mode, fn func(string, *token.FileSet, *ast.File)) {
	t.Helper()
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			// The root is walked whatever its name: a rule pointed at a
			// testdata tree must not pass by visiting nothing.
			if path != dir && SkipDir(d.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		fset := token.NewFileSet()
		f, parseErr := parser.ParseFile(fset, path, nil, mode)
		if parseErr != nil {
			t.Errorf("archtest: cannot parse %s, so no rule can check it: %v", path, parseErr)
			return nil
		}
		fn(path, fset, f)
		return nil
	})
	if err != nil {
		t.Fatalf("archtest: walking %s: %v", dir, err)
	}
}

// sorted returns a copy of vs ordered by file, then line, then message.
func sorted(vs []Violation) []Violation {
	out := slices.Clone(vs)
	slices.SortStableFunc(out, func(a, b Violation) int {
		return cmp.Or(cmp.Compare(a.File, b.File), cmp.Compare(a.Line, b.Line), cmp.Compare(a.Message, b.Message))
	})
	return out
}

// ReportViolations reports the violations of the rule named rule, sorted by
// file and line. With none it does nothing. Otherwise it fails the test, or,
// when knownViolations is true, logs them and skips the test so a rule can land
// before its existing violations are fixed.
func ReportViolations(t testing.TB, rule string, violations []Violation, knownViolations bool) {
	t.Helper()
	if len(violations) == 0 {
		return
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "%s: %d violation(s) found:\n", rule, len(violations))
	for _, v := range sorted(violations) {
		fmt.Fprintln(&sb, v.String())
	}
	if knownViolations {
		t.Log(strings.TrimSuffix(sb.String(), "\n"))
		t.Skipf("%s: %d known violation(s) exist — unskip after refactoring (see output above)", rule, len(violations))
		return
	}
	t.Fatal(strings.TrimSuffix(sb.String(), "\n"))
}

// ReportWithFileAllowlist reports violations against a file allowlist, keyed
// by [Violation.File]. A violation in a file not in the allowlist fails the
// test. Violations in allowlisted files are logged as warnings. An allowlisted
// file with no violations is logged as a hint to remove it.
//
// To use it, list the files with pre-existing violations, call it with the
// rule's violations, and remove each file from the list once it is fixed.
func ReportWithFileAllowlist(t testing.TB, rule string, violations []Violation, allowlist map[string]bool) {
	t.Helper()

	violatedFiles := map[string]bool{}
	var newViolations, allowedViolations []Violation
	for _, v := range sorted(violations) {
		violatedFiles[v.File] = true
		if allowlist[v.File] {
			allowedViolations = append(allowedViolations, v)
		} else {
			newViolations = append(newViolations, v)
		}
	}

	for _, file := range sortedKeys(allowlist) {
		if !violatedFiles[file] {
			t.Logf("%s: allowlisted file %s has no violations — remove from allowlist to lock in the fix", rule, file)
		}
	}

	if len(allowedViolations) > 0 {
		var sb strings.Builder
		fmt.Fprintf(&sb, "%s: %d pre-existing violation(s) in allowlisted files (WARN):\n", rule, len(allowedViolations))
		for _, v := range allowedViolations {
			fmt.Fprintln(&sb, v.String())
		}
		t.Log(strings.TrimSuffix(sb.String(), "\n"))
	}
	if len(newViolations) > 0 {
		var sb strings.Builder
		fmt.Fprintf(&sb, "%s: %d NEW violation(s) not in allowlist — MUST FIX:\n", rule, len(newViolations))
		for _, v := range newViolations {
			fmt.Fprintln(&sb, v.String())
		}
		t.Fatal(strings.TrimSuffix(sb.String(), "\n"))
	}
}

// ReportWithCountAllowlist is a stricter [ReportWithFileAllowlist] keyed by
// file and the number of violations it may keep. A file not in the allowlist,
// or a file with more violations than its count, fails the test. A file with
// fewer logs a hint to lower the count. So the total can shrink but never
// grow, while a file that is listed for N violations cannot hide an N+1th.
//
// It counts and does not track instances, so fixing one violation and adding
// another in the same file passes. Review has to catch that case.
func ReportWithCountAllowlist(t testing.TB, rule string, violations []Violation, allowlist map[string]int) {
	t.Helper()

	byFile := map[string][]Violation{}
	for _, v := range sorted(violations) {
		byFile[v.File] = append(byFile[v.File], v)
	}

	var failures, warns []string
	for _, file := range sortedKeys(byFile) {
		vs := byFile[file]
		allowed, listed := allowlist[file]
		switch {
		case !listed:
			for _, v := range vs {
				failures = append(failures, v.String())
			}
		case len(vs) > allowed:
			failures = append(failures, fmt.Sprintf(
				"%s: %d violation(s) but only %d grandfathered — a new one was added. "+
					"Fix it, or if it is genuinely pre-existing bump the allowlist count to %d:",
				file, len(vs), allowed, len(vs)))
			for _, v := range vs {
				failures = append(failures, "  "+v.String())
			}
		default:
			for _, v := range vs {
				warns = append(warns, v.String())
			}
			if len(vs) < allowed {
				warns = append(warns, fmt.Sprintf(
					"  ↳ %s now has %d violation(s) (allowlist expects %d) — lower the count to lock in the fix",
					file, len(vs), allowed))
			}
		}
	}
	for _, file := range sortedKeys(allowlist) {
		if len(byFile[file]) == 0 {
			t.Logf("%s: allowlisted file %s has no violations — remove from allowlist to lock in the fix", rule, file)
		}
	}

	if len(warns) > 0 {
		t.Logf("%s: %d pre-existing violation(s) in allowlisted files (WARN):\n%s", rule, len(warns), strings.Join(warns, "\n"))
	}
	if len(failures) > 0 {
		t.Fatalf("%s: new violation(s) — MUST FIX:\n%s", rule, strings.Join(failures, "\n"))
	}
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}

// baseSelector returns the pkg.Name selector at the core of a type
// expression, unwrapping pointers, slices, arrays, channels, parentheses and
// maps (the value type first, then the key). It returns nil when there is none.
func baseSelector(expr ast.Expr) *ast.SelectorExpr {
	switch t := expr.(type) {
	case *ast.StarExpr:
		return baseSelector(t.X)
	case *ast.ArrayType:
		return baseSelector(t.Elt)
	case *ast.ChanType:
		return baseSelector(t.Value)
	case *ast.ParenExpr:
		return baseSelector(t.X)
	case *ast.MapType:
		if sel := baseSelector(t.Value); sel != nil {
			return sel
		}
		return baseSelector(t.Key)
	case *ast.SelectorExpr:
		return t
	}
	return nil
}

// TypeRefersToPackage reports whether the type expression expr names a type
// from the package imported as pkgName, such as *pkg.T, []pkg.T or
// map[string]pkg.T.
func TypeRefersToPackage(expr ast.Expr, pkgName string) bool {
	return SelectorPkg(expr) == pkgName && pkgName != ""
}

// SelectorPkg returns the package name in the type expression expr, such as
// "pkg" for []*pkg.T, or "" when expr names no imported type.
func SelectorPkg(expr ast.Expr) string {
	sel := baseSelector(expr)
	if sel == nil {
		return ""
	}
	ident, ok := sel.X.(*ast.Ident)
	if !ok {
		return ""
	}
	return ident.Name
}

// SelectorTypeName returns the type name in the type expression expr, such
// as "T" for []*pkg.T, or "" when expr names no imported type.
func SelectorTypeName(expr ast.Expr) string {
	sel := baseSelector(expr)
	if sel == nil {
		return ""
	}
	return sel.Sel.Name
}

var majorVersion = regexp.MustCompile(`^v[0-9]+$`)

// DefaultImportName returns the name an unnamed import of importPath is
// assumed to bind: the last path element, with a /vN major-version element
// skipped, a ".vN" suffix removed and a "go-" prefix removed. So
// "gopkg.in/yaml.v3" binds "yaml", "example.com/lib/v2" binds "lib" and
// "example.com/go-cmp" binds "cmp". A package whose name differs from its
// path in another way is not seen.
func DefaultImportName(importPath string) string {
	base := path.Base(importPath)
	if majorVersion.MatchString(base) {
		if dir := path.Dir(importPath); dir != "." {
			base = path.Base(dir)
		}
	}
	if i := strings.LastIndex(base, ".v"); i > 0 && majorVersion.MatchString(base[i+1:]) {
		base = base[:i]
	}
	return strings.TrimPrefix(base, "go-")
}

// ImportPathForAlias returns the import path that alias refers to among
// imports: an import named alias, or an unnamed import whose
// [DefaultImportName] is alias. It returns "" when none matches.
func ImportPathForAlias(alias string, imports []*ast.ImportSpec) string {
	for _, imp := range imports {
		// The parser accepted the literal, so Unquote fails only on an AST built by hand.
		importPath, err := strconv.Unquote(imp.Path.Value) //swallowcheck:ignore // such an import cannot be used
		if err != nil {
			continue
		}
		name := DefaultImportName(importPath)
		if imp.Name != nil {
			name = imp.Name.Name
		}
		if name == alias {
			return importPath
		}
	}
	return ""
}

// IsEmptyInterfaceType reports whether expr is any or interface{}.
func IsEmptyInterfaceType(expr ast.Expr) bool {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name == "any"
	case *ast.InterfaceType:
		return t.Methods == nil || len(t.Methods.List) == 0
	}
	return false
}
