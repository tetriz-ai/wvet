// Copyright 2026 Workloom Solutions Private Limited
// SPDX-License-Identifier: MIT

// Package waitgroupcheck reports uses of sync.WaitGroup's Wait method.
//
// A plain Wait blocks forever if a goroutine never calls Done: a forgotten
// Done, a panic before Done, or a goroutine that hangs. The check points at a
// wait with a bound, such as SafeWait in github.com/tetriz-ai/wvet/waitgroup.
//
// It reports a selector whose resolved method is (*sync.WaitGroup).Wait, read
// through go/types from the method's receiver type, not from a variable name.
// These forms all count:
//
//   - a local variable, a struct field (s.wg.Wait()) or a *sync.WaitGroup parameter
//   - a value built with sync.WaitGroup{} or new(sync.WaitGroup)
//   - an aliased sync import, or an embedded sync.WaitGroup (promoted Wait)
//   - the method expression (*sync.WaitGroup).Wait(&wg)
//   - the method value passed on without a call, as in run(wg.Wait)
//
// A Wait through a type parameter whose constraint allows only sync.WaitGroup
// or *sync.WaitGroup counts too. A Wait on any other type, such as sync.Cond or
// errgroup.Group, is not reported. Nor is a Wait through an interface value,
// whose dynamic type the check cannot see.
//
// Files with the standard "// Code generated ... DO NOT EDIT." header are
// skipped. A //waitgroupcheck:ignore, //nolint or //nolint:waitgroupcheck
// comment on its own line directly above the statement suppresses the report
// for that statement only, not for one nested in it or following it on the
// same line. The directive may be followed by whitespace and a reason
// ("// reason"). A trailing comment on the Wait's line does not count.
package waitgroupcheck

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"

	"github.com/tetriz-ai/wvet/internal/directive"
)

// Analyzer reports uses of (*sync.WaitGroup).Wait.
var Analyzer = &analysis.Analyzer{
	Name: "waitgroupcheck",
	Doc:  "reports sync.WaitGroup.Wait, which blocks forever if a Done is missed",
	URL:  "https://pkg.go.dev/github.com/tetriz-ai/wvet/waitgroupcheck",
	// Type information for the well-typed parts of a package still resolves
	// the Wait selections, so a type error elsewhere does not stop the check.
	RunDespiteErrors: true,
	Run:              run,
}

const message = "sync.WaitGroup.Wait blocks forever if a Done is missed; " +
	"use waitgroup.SafeWait (github.com/tetriz-ai/wvet/waitgroup) or github.com/sourcegraph/conc"

func run(pass *analysis.Pass) (any, error) {
	ignore := directive.Collect(pass.Fset, pass.Files, "waitgroupcheck", pass.ReadFile)
	for _, f := range pass.Files {
		if ast.IsGenerated(f) {
			continue
		}
		// stmts is the stack of statements enclosing the node being visited.
		// decl is the top-level declaration being visited, the fallback for a
		// match outside any statement, such as var waitFn = wg.Wait.
		var stmts []ast.Stmt
		var decl ast.Decl
		var visit func(n ast.Node) bool
		visit = func(n ast.Node) bool {
			if stmt, ok := n.(ast.Stmt); ok {
				stmts = append(stmts, stmt)
				ast.Inspect(stmt, func(c ast.Node) bool {
					if c == stmt {
						return true
					}
					if c == nil {
						return false
					}
					return visit(c)
				})
				stmts = stmts[:len(stmts)-1]
				return false
			}
			sel, ok := n.(*ast.SelectorExpr)
			if ok && sel.Sel.Name == "Wait" && isSyncWaitGroupWait(pass.TypesInfo, sel) {
				var enclosing ast.Node = decl
				if len(stmts) > 0 {
					enclosing = stmts[len(stmts)-1]
				}
				if !ignore.Suppressed(enclosing) {
					pass.Reportf(sel.Pos(), message)
				}
			}
			return true
		}
		for _, d := range f.Decls {
			decl = d
			ast.Inspect(d, func(n ast.Node) bool {
				if n == nil {
					return false
				}
				return visit(n)
			})
		}
	}
	return nil, nil
}

// isSyncWaitGroupWait reports whether sel selects the (*sync.WaitGroup).Wait
// method, as a method value (wg.Wait, called or not) or a method expression
// ((*sync.WaitGroup).Wait). Only the resolved method's receiver type decides.
func isSyncWaitGroupWait(info *types.Info, sel *ast.SelectorExpr) bool {
	if info == nil {
		return false
	}
	selection := info.Selections[sel]
	if selection == nil {
		return false
	}
	switch selection.Kind() {
	case types.MethodVal, types.MethodExpr:
	default:
		return false
	}
	if tp, ok := selection.Recv().(*types.TypeParam); ok {
		return onlySyncWaitGroup(tp)
	}
	fn, ok := selection.Obj().(*types.Func)
	if !ok {
		return false
	}
	sig, ok := fn.Type().(*types.Signature)
	if !ok || sig.Recv() == nil {
		return false
	}
	return isSyncWaitGroup(sig.Recv().Type())
}

// onlySyncWaitGroup reports whether every type a type parameter's constraint
// allows is sync.WaitGroup or *sync.WaitGroup, so its Wait is the WaitGroup's.
func onlySyncWaitGroup(tp *types.TypeParam) bool {
	return restrictsToWaitGroup(tp.Constraint())
}

// restrictsToWaitGroup reports whether the constraint t allows only
// WaitGroup types. Its type set is the intersection of its embedded elements,
// so one element that allows only WaitGroup types is enough.
func restrictsToWaitGroup(t types.Type) bool {
	iface, ok := t.Underlying().(*types.Interface)
	if !ok {
		return false
	}
	for i := range iface.NumEmbeddeds() {
		if elementOnlyWaitGroup(iface.EmbeddedType(i)) {
			return true
		}
	}
	return false
}

// elementOnlyWaitGroup reports whether one element of a constraint, a type,
// a union or an embedded interface, allows only WaitGroup types.
func elementOnlyWaitGroup(e types.Type) bool {
	if u, ok := e.(*types.Union); ok {
		for j := range u.Len() {
			if !waitGroupTerm(u.Term(j).Type(), u.Term(j).Tilde()) {
				return false
			}
		}
		return u.Len() > 0
	}
	if types.IsInterface(e) {
		return restrictsToWaitGroup(e)
	}
	return waitGroupTerm(e, false)
}

// waitGroupTerm reports whether a type term allows only WaitGroup types with
// a Wait method. ~*sync.WaitGroup counts: the other types it allows are defined
// pointer types, which have no methods.
func waitGroupTerm(t types.Type, tilde bool) bool {
	if tilde {
		ptr, ok := t.(*types.Pointer)
		return ok && isSyncWaitGroup(ptr)
	}
	return isSyncWaitGroup(t)
}

// isSyncWaitGroup reports whether t is sync.WaitGroup or *sync.WaitGroup.
func isSyncWaitGroup(t types.Type) bool {
	if ptr, ok := t.(*types.Pointer); ok {
		t = ptr.Elem()
	}
	named, ok := t.(*types.Named)
	if !ok {
		return false
	}
	obj := named.Obj()
	return obj != nil && obj.Pkg() != nil && obj.Pkg().Path() == "sync" && obj.Name() == "WaitGroup"
}
