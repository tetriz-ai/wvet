// Copyright 2026 Workloom Solutions Private Limited
// SPDX-License-Identifier: MIT

// Package goroutinecheck reports go statements.
//
// A panic in a goroutine that does not recover it ends the whole process. The
// check reports every go statement, so each goroutine starts through a helper
// that recovers, such as Run in github.com/tetriz-ai/wvet/goroutine or a pool
// from github.com/sourcegraph/conc. It does not look for goroutine leaks, and it
// does not look inside the function a go statement starts: a goroutine that
// recovers on its own is reported too.
//
// Files with the standard "// Code generated ... DO NOT EDIT." header are
// skipped. A //goroutinecheck:ignore, //nolint or //nolint:goroutinecheck
// comment on its own line directly above the go statement suppresses the
// report for that statement only, not for a go statement nested in it or
// following it on the same line. The directive may be followed by whitespace
// and a reason ("// reason"). A trailing comment on the statement's line does
// not count.
package goroutinecheck

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"

	"github.com/tetriz-ai/wvet/internal/directive"
)

// Analyzer reports go statements.
var Analyzer = &analysis.Analyzer{
	Name: "goroutinecheck",
	Doc:  "reports go statements, whose goroutine ends the process if it panics without recovering",
	URL:  "https://pkg.go.dev/github.com/tetriz-ai/wvet/goroutinecheck",
	// The check is syntactic, so a type error elsewhere in the package does
	// not stop it from reporting.
	RunDespiteErrors: true,
	Run:              run,
}

const message = "go statement: a panic in this goroutine ends the process; " +
	"start it with goroutine.Run (github.com/tetriz-ai/wvet/goroutine) or github.com/sourcegraph/conc"

func run(pass *analysis.Pass) (any, error) {
	ignore := directive.Collect(pass.Fset, pass.Files, "goroutinecheck", pass.ReadFile)
	for _, f := range pass.Files {
		if ast.IsGenerated(f) {
			continue
		}
		ast.Inspect(f, func(n ast.Node) bool {
			stmt, ok := n.(*ast.GoStmt)
			if ok && !ignore.StatementSuppressed(stmt) {
				pass.Reportf(stmt.Pos(), message)
			}
			return true
		})
	}
	return nil, nil
}
