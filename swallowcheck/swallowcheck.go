// Copyright 2026 Workloom Solutions Private Limited
// SPDX-License-Identifier: MIT

// Package swallowcheck reports an error that is compared to nil and then never
// propagated, logged or inspected: the swallowed error.
//
// It reports an error-typed value produced by a call when it is compared to nil
// and every other use is a nil comparison. Any other use counts as handled:
// returned, passed to a call, stored, sent or panicked with. The value is
// followed through interface boxing, type changes, branches that merge it, and local
// variables, including a variable a closure captures or stores into, both for
// its uses and for the nil comparison. An error replaced on its non-nil branch
// (if err != nil { err = errX }) is not handled by the replacement. A
// replacement made in a deferred closure is not seen, so that error is not
// reported. An error never compared to nil is not reported.
//
// The report is at the call that produced the error. A //swallowcheck:ignore,
// //nolint or //nolint:swallowcheck comment on the call's line or on a nil
// comparison's line silences it. A reason may follow the directive. Files with
// the standard "// Code generated ... DO NOT EDIT." header are skipped.
//
// Each function is checked alone, so an error handed to a helper that drops it
// is not reported. An error used wrongly, such as wrapped with the wrong
// message, counts as handled.
package swallowcheck

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/buildssa"
	"golang.org/x/tools/go/ssa"

	"github.com/tetriz-ai/wvet/internal/directive"
)

// Analyzer reports error values that are checked but never propagated,
// logged, or inspected.
var Analyzer = &analysis.Analyzer{
	Name:     "swallowcheck",
	Doc:      "reports error values checked but never propagated, logged, or inspected (the swallowed error)",
	URL:      "https://pkg.go.dev/github.com/tetriz-ai/wvet/swallowcheck",
	Requires: []*analysis.Analyzer{buildssa.Analyzer},
	Run:      run,
}

var (
	errorType  = types.Universe.Lookup("error").Type()
	errorIface = mustErrorInterface()
)

// mustErrorInterface returns the underlying interface of the universe `error`
// type. The assertion is a Go-language invariant (error is always an
// interface); guard it explicitly so a future toolchain change fails loudly
// here rather than panicking on an unchecked assertion.
func mustErrorInterface() *types.Interface {
	iface, ok := errorType.Underlying().(*types.Interface)
	if !ok {
		panic(fmt.Sprintf("swallowcheck: error.Underlying() is %T, expected *types.Interface", errorType.Underlying()))
	}
	return iface
}

func run(pass *analysis.Pass) (any, error) {
	ignore := directive.Collect(pass.Fset, pass.Files, "swallowcheck", pass.ReadFile)
	generated := map[*token.File]bool{}
	for _, f := range pass.Files {
		if ast.IsGenerated(f) {
			generated[pass.Fset.File(f.Pos())] = true
		}
	}
	res, ok := pass.ResultOf[buildssa.Analyzer].(*buildssa.SSA)
	if !ok {
		return nil, fmt.Errorf("swallowcheck: unexpected result type %T for buildssa analyzer", pass.ResultOf[buildssa.Analyzer])
	}
	for _, fn := range functions(res) {
		for _, b := range fn.Blocks {
			for _, instr := range b.Instrs {
				v, ok := instr.(ssa.Value)
				if !ok || !producedByCall(v) || !isErrorTyped(v.Type()) {
					continue
				}
				if used(v, map[ssa.Value]bool{}) {
					continue
				}
				var checks []token.Pos
				nilChecks(v, map[ssa.Value]bool{}, &checks)
				if len(checks) == 0 {
					// An error neither checked nor used is unchecked, which errcheck reports. Skip it.
					continue
				}
				reportPos := callPos(v)
				if generated[pass.Fset.File(reportPos)] || ignore.OnLine(append(checks, reportPos)...) {
					continue
				}
				pass.Reportf(reportPos, "error checked but never propagated, logged, or inspected (swallowed)")
			}
		}
	}
	return nil, nil
}

// functions returns the package's source functions and the function literals
// in package-level variable initializers, which live in the package's init
// function and are not among the source functions.
func functions(res *buildssa.SSA) []*ssa.Function {
	seen := map[*ssa.Function]bool{}
	var out []*ssa.Function
	var add func(fn *ssa.Function)
	add = func(fn *ssa.Function) {
		if fn == nil || seen[fn] {
			return
		}
		seen[fn] = true
		out = append(out, fn)
		for _, anon := range fn.AnonFuncs {
			add(anon)
		}
	}
	for _, fn := range res.SrcFuncs {
		add(fn)
	}
	if init := res.Pkg.Func("init"); init != nil {
		for _, anon := range init.AnonFuncs {
			add(anon)
		}
	}
	return out
}

// isErrorTyped matches the error interface itself and any concrete type that
// implements it.
func isErrorTyped(t types.Type) bool {
	return types.Implements(t, errorIface)
}

// producedByCall limits the check to real error origins (call results), not
// parameters or globals.
func producedByCall(v ssa.Value) bool {
	switch t := v.(type) {
	case *ssa.Call:
		return true
	case *ssa.Extract:
		_, ok := t.Tuple.(*ssa.Call)
		return ok
	}
	return false
}

func callPos(v ssa.Value) token.Pos {
	switch t := v.(type) {
	case *ssa.Call:
		return t.Pos()
	case *ssa.Extract:
		if c, ok := t.Tuple.(*ssa.Call); ok {
			return c.Pos()
		}
	}
	return v.Pos()
}

func isNilCompare(b *ssa.BinOp) bool {
	if b.Op != token.EQL && b.Op != token.NEQ {
		return false
	}
	if c, ok := b.X.(*ssa.Const); ok && c.IsNil() {
		return true
	}
	if c, ok := b.Y.(*ssa.Const); ok && c.IsNil() {
		return true
	}
	return false
}

// passThrough returns the values v flows into without being used: an
// interface boxing, a type change, a phi merge, or the loads a store into a
// local variable reaches. ok is false when v is stored somewhere it can no
// longer be followed, which counts as handled.
func passThrough(instr ssa.Instruction, v ssa.Value) (next []ssa.Value, ok bool) {
	switch r := instr.(type) {
	case *ssa.MakeInterface, *ssa.ChangeType, *ssa.ChangeInterface, *ssa.Phi:
		return []ssa.Value{r.(ssa.Value)}, true
	case *ssa.Store:
		if r.Val != v {
			return nil, false
		}
		switch addr := r.Addr.(type) {
		case *ssa.Alloc:
			return loadsReached(r, addr)
		case *ssa.FreeVar:
			return closureStoreReaches(r, addr)
		}
	}
	return nil, false
}

// boundAlloc returns the local variable a closure's free variable is bound to,
// through any nesting of closures, or nil when it is not a single local.
func boundAlloc(fv *ssa.FreeVar) *ssa.Alloc {
	fn := fv.Parent()
	idx := -1
	for i, f := range fn.FreeVars {
		if f == fv {
			idx = i
		}
	}
	refs := fn.Referrers()
	if idx < 0 || refs == nil {
		return nil
	}
	// A function literal has one MakeClosure, the one that creates it.
	for _, instr := range *refs {
		mc, ok := instr.(*ssa.MakeClosure)
		if !ok || mc.Fn != fn || idx >= len(mc.Bindings) {
			continue
		}
		switch b := mc.Bindings[idx].(type) {
		case *ssa.Alloc:
			return b
		case *ssa.FreeVar:
			return boundAlloc(b)
		}
		return nil
	}
	return nil
}

// nilProof reports which successor of the branch on cond is the one where
// one of vals is known to be nil: cond compares a value in vals, or an
// interface-to-interface copy of one, to nil. nilIdx is the index into the
// branch block's Succs.
func nilProof(cond ssa.Value, vals func(ssa.Value) bool) (nilIdx int, ok bool) {
	b, isBin := cond.(*ssa.BinOp)
	if !isBin || !isNilCompare(b) {
		return 0, false
	}
	operand := b.X
	if c, isConst := operand.(*ssa.Const); isConst && c.IsNil() {
		operand = b.Y
	}
	for {
		if vals(operand) {
			break
		}
		// A conversion between interface types is nil exactly when its source is.
		switch c := operand.(type) {
		case *ssa.ChangeInterface:
			operand = c.X
			continue
		case *ssa.ChangeType:
			if types.IsInterface(c.X.Type()) {
				operand = c.X
				continue
			}
		}
		return 0, false
	}
	if b.Op == token.EQL {
		return 0, true // v == nil is true on Succs[0]
	}
	return 1, true // v != nil is false on Succs[1]
}

// nilOnEdge reports whether, on the edge from pred into block, v is known to
// be nil: a nil comparison of v (or of an interface copy of it) branches there
// on its nil side.
func nilOnEdge(v ssa.Value, pred, block *ssa.BasicBlock) bool {
	var branches []*ssa.If
	var collect func(x ssa.Value)
	collect = func(x ssa.Value) {
		refs := x.Referrers()
		if refs == nil {
			return
		}
		for _, instr := range *refs {
			switch r := instr.(type) {
			case *ssa.BinOp:
				if br := r.Referrers(); br != nil && isNilCompare(r) {
					for _, use := range *br {
						if branch, ok := use.(*ssa.If); ok && branch.Cond == r {
							branches = append(branches, branch)
						}
					}
				}
			case *ssa.ChangeInterface:
				collect(r)
			case *ssa.ChangeType:
				if types.IsInterface(x.Type()) {
					collect(r)
				}
			}
		}
	}
	collect(v)
	for _, branch := range branches {
		nilIdx, ok := nilProof(branch.Cond, func(x ssa.Value) bool { return x == v })
		if !ok {
			continue
		}
		from := branch.Block()
		nilSide := from.Succs[nilIdx]
		if pred == from && nilSide == block {
			return true
		}
		if len(nilSide.Preds) == 1 && nilSide.Dominates(pred) {
			return true
		}
	}
	return false
}

// onlyNilThrough reports whether v enters phi only on edges where v is known
// to be nil, so the phi never carries v as an error. An error replaced on its
// non-nil branch (if err != nil { err = errX }) merges this way. phi is a
// referrer of v, so at least one edge carries it.
func onlyNilThrough(v ssa.Value, phi *ssa.Phi) bool {
	block := phi.Block()
	for i, edge := range phi.Edges {
		if edge == v && !nilOnEdge(v, block.Preds[i], block) {
			return false
		}
	}
	return true
}

// nilChecks collects the position of every nil comparison of v, directly or
// after the pass-through ops. The visited set guards against phi cycles in loops.
func nilChecks(v ssa.Value, seen map[ssa.Value]bool, out *[]token.Pos) {
	if seen[v] {
		return
	}
	seen[v] = true
	refs := v.Referrers()
	if refs == nil {
		return
	}
	for _, instr := range *refs {
		if b, ok := instr.(*ssa.BinOp); ok {
			if isNilCompare(b) {
				*out = append(*out, b.Pos())
			}
			continue
		}
		next, ok := passThrough(instr, v)
		if !ok {
			continue
		}
		for _, n := range next {
			nilChecks(n, seen, out)
		}
	}
}

// used reports whether v has a meaningful use: anything other than a nil
// comparison, after the pass-through ops. The visited set guards against phi
// cycles in loops.
func used(v ssa.Value, seen map[ssa.Value]bool) bool {
	if seen[v] {
		return false
	}
	seen[v] = true
	refs := v.Referrers()
	if refs == nil {
		return false
	}
	for _, instr := range *refs {
		if b, ok := instr.(*ssa.BinOp); ok && isNilCompare(b) {
			continue
		}
		if phi, ok := instr.(*ssa.Phi); ok && onlyNilThrough(v, phi) {
			continue
		}
		next, ok := passThrough(instr, v)
		if !ok {
			return true
		}
		for _, n := range next {
			if used(n, seen) {
				return true
			}
		}
	}
	return false
}

// walk follows a store to addr forward from instruction start of block b. It
// returns the loads of addr that the store reaches: those on a path with no
// other store to addr in between, and not past the nil side of a branch on a
// nil comparison of such a load, where the stored error is known to be nil.
// exits reports whether a return is reached on such a path, so the value can
// leave the function.
func walk(addr ssa.Value, b *ssa.BasicBlock, start int) (loads []ssa.Value, exits bool) {
	reached := map[ssa.Value]bool{}
	// scan visits instrs, and returns the successors to follow, or none when a
	// store to addr ends the path.
	scan := func(blk *ssa.BasicBlock, instrs []ssa.Instruction) []*ssa.BasicBlock {
		for _, instr := range instrs {
			if u, ok := instr.(*ssa.UnOp); ok && u.Op == token.MUL && u.X == addr {
				loads = append(loads, u)
				reached[u] = true
			}
			if st, ok := instr.(*ssa.Store); ok && st.Addr == addr {
				return nil
			}
		}
		switch last := blk.Instrs[len(blk.Instrs)-1].(type) {
		case *ssa.Return:
			exits = true
		case *ssa.If:
			if nilIdx, ok := nilProof(last.Cond, func(x ssa.Value) bool { return reached[x] }); ok {
				return []*ssa.BasicBlock{blk.Succs[1-nilIdx]}
			}
		}
		return blk.Succs
	}
	queue := scan(b, b.Instrs[start:])
	visited := map[*ssa.BasicBlock]bool{}
	for len(queue) > 0 {
		blk := queue[0]
		queue = queue[1:]
		if visited[blk] {
			continue
		}
		visited[blk] = true
		queue = append(queue, scan(blk, blk.Instrs)...)
	}
	return loads, exits
}

// runsAt returns where a closure runs in the function that creates it: at
// each call or go statement that calls it or takes it as an argument, and at
// each rundefers for a deferred one. A closure that escapes otherwise (stored,
// for one) may run at any time after its creation, so its creation is the one
// place returned.
func runsAt(mc *ssa.MakeClosure) []ssa.Instruction {
	var at []ssa.Instruction
	for _, use := range *mc.Referrers() {
		call, ok := use.(ssa.CallInstruction)
		if !ok {
			return []ssa.Instruction{mc}
		}
		if _, isDefer := call.(*ssa.Defer); !isDefer {
			at = append(at, call)
			continue
		}
		for _, b := range mc.Parent().Blocks {
			for _, instr := range b.Instrs {
				if _, ok := instr.(*ssa.RunDefers); ok {
					at = append(at, instr)
				}
			}
		}
	}
	return at
}

// after returns the index just past instr in its block.
func after(instr ssa.Instruction) int {
	b := instr.Block()
	for i, x := range b.Instrs {
		if x == instr {
			return i + 1
		}
	}
	return len(b.Instrs)
}

// loadsReached returns the loads from alloc that store can reach: in the
// function, along walk's paths, plus every load through a closure that
// captures alloc, since a closure may run at any time. ok is false when
// alloc's address escapes: passed to a call, stored, or converted, so the
// stored value's fate is unknown.
func loadsReached(store *ssa.Store, alloc *ssa.Alloc) (loads []ssa.Value, ok bool) {
	if !captured(alloc, nil, &loads) {
		return nil, false
	}
	direct, _ := walk(alloc, store.Block(), after(store))
	return append(loads, direct...), true
}

// closureStoreReaches returns the loads a closure's store to a captured
// variable can reach. Inside the closure they are walk's. If the value can
// leave the closure, they also include every load in the other closures that
// capture the variable, and, in the function that owns it, the loads walk
// reaches from each place the closure runs. ok is false when the variable is not a
// single local or its address escapes.
func closureStoreReaches(store *ssa.Store, fv *ssa.FreeVar) (loads []ssa.Value, ok bool) {
	alloc := boundAlloc(fv)
	if alloc == nil {
		return nil, false
	}
	var closureLoads []ssa.Value
	if !captured(alloc, nil, &closureLoads) {
		return nil, false
	}
	inside, exits := walk(fv, store.Block(), after(store))
	if !exits {
		return inside, true
	}
	loads = inside
	for _, l := range closureLoads {
		if l.(ssa.Instruction).Parent() != store.Parent() {
			loads = append(loads, l)
		}
	}
	if refs := alloc.Referrers(); refs != nil {
		for _, instr := range *refs {
			if mc, isClosure := instr.(*ssa.MakeClosure); isClosure {
				for _, at := range runsAt(mc) {
					fromParent, _ := walk(alloc, at.Block(), after(at))
					loads = append(loads, fromParent...)
				}
			}
		}
	}
	return loads, true
}

// captured checks every use of addr, a local variable or a closure's free
// variable bound to it. Loads and stores are expected. A closure that captures
// addr is followed into its free variable, and that free variable's loads are
// appended to loads. Any other use is an escape, and captured returns false.
// fv is the free variable being checked, or nil for the local itself.
func captured(addr ssa.Value, fv *ssa.FreeVar, loads *[]ssa.Value) bool {
	refs := addr.Referrers()
	if refs == nil {
		return true
	}
	for _, instr := range *refs {
		switch r := instr.(type) {
		case *ssa.UnOp:
			if r.Op != token.MUL {
				return false
			}
			if fv != nil {
				*loads = append(*loads, r)
			}
		case *ssa.Store:
			if r.Addr != addr || r.Val == addr {
				return false
			}
		case *ssa.MakeClosure:
			fn, ok := r.Fn.(*ssa.Function)
			if !ok {
				return false
			}
			for i, bound := range r.Bindings {
				if bound != addr || i >= len(fn.FreeVars) {
					continue
				}
				if !captured(fn.FreeVars[i], fn.FreeVars[i], loads) {
					return false
				}
			}
		case *ssa.DebugRef:
		default:
			return false
		}
	}
	return true
}
