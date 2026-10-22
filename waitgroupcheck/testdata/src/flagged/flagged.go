// Copyright 2026 Workloom Solutions Private Limited
// SPDX-License-Identifier: MIT

// Package flagged holds waits the analyzer must report.
package flagged

import (
	"sync"

	"constraints"
)

func shortVarDecl() {
	x := &sync.WaitGroup{}
	x.Add(1)
	go func() {
		defer x.Done()
	}()
	x.Wait() // want "sync.WaitGroup.Wait blocks forever"
}

func varValue() {
	var (
		wg sync.WaitGroup
	)
	wg.Add(1)
	go func() {
		defer wg.Done()
	}()
	wg.Wait() // want "sync.WaitGroup.Wait blocks forever"
}

func varPointer() {
	var (
		wg *sync.WaitGroup
	)
	wg.Add(1)
	wg.Wait() // want "sync.WaitGroup.Wait blocks forever"
}

func varAssignedPointer() {
	var wg = &sync.WaitGroup{}
	wg.Wait() // want "sync.WaitGroup.Wait blocks forever"
}

type server struct {
	wg sync.WaitGroup
}

func (s *server) stop() {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
	}()
	s.wg.Wait() // want "sync.WaitGroup.Wait blocks forever"
}

func parameter(wg *sync.WaitGroup) {
	wg.Wait() // want "sync.WaitGroup.Wait blocks forever"
}

func valueConstructed() {
	wg := sync.WaitGroup{}
	wg.Wait() // want "sync.WaitGroup.Wait blocks forever"
}

func newConstructed() {
	wg := new(sync.WaitGroup)
	wg.Wait() // want "sync.WaitGroup.Wait blocks forever"
}

type group struct {
	sync.WaitGroup
}

func embedded() {
	var g group
	g.Wait() // want "sync.WaitGroup.Wait blocks forever"
}

func methodExpression() {
	var wg sync.WaitGroup
	(*sync.WaitGroup).Wait(&wg) // want "sync.WaitGroup.Wait blocks forever"
}

func run(f func()) { f() }

func methodValue() {
	var wg sync.WaitGroup
	run(wg.Wait) // want "sync.WaitGroup.Wait blocks forever"
}

func methodValueStored() {
	var wg sync.WaitGroup
	wait := wg.Wait // want "sync.WaitGroup.Wait blocks forever"
	wait()
}

func deferred() {
	var wg sync.WaitGroup
	defer wg.Wait() // want "sync.WaitGroup.Wait blocks forever"
}

func inAGoroutine() {
	var wg sync.WaitGroup
	done := make(chan struct{})
	go func() {
		wg.Wait() // want "sync.WaitGroup.Wait blocks forever"
		close(done)
	}()
	<-done
}

// Directives that do not suppress.

func trailingDirective() {
	var wg sync.WaitGroup
	wg.Wait() //waitgroupcheck:ignore // want "sync.WaitGroup.Wait blocks forever"
}

func ignoredIsNotIgnore() {
	var wg sync.WaitGroup
	//waitgroupcheck:ignoreall
	wg.Wait() // want "sync.WaitGroup.Wait blocks forever"
}

func unknownDirective() {
	var wg sync.WaitGroup
	//lint:ignore waitgroupcheck
	wg.Wait() // want "sync.WaitGroup.Wait blocks forever"
}

func otherAnalyzer() {
	var wg sync.WaitGroup
	//goroutinecheck:ignore
	wg.Wait() // want "sync.WaitGroup.Wait blocks forever"
}

// throughATypeParameter: the constraint allows only *sync.WaitGroup, so Wait
// is the WaitGroup's.
func throughATypeParameter[W interface {
	*sync.WaitGroup
	Wait()
}](w W) {
	w.Wait() // want "sync.WaitGroup.Wait blocks forever"
}

func callsTheTypeParameter() {
	var wg sync.WaitGroup
	throughATypeParameter(&wg)
}

// directiveAboveAFuncLiteral: the directive names the call statement; the
// waits inside the literal are statements of their own.
func directiveAboveAFuncLiteral(a, b *sync.WaitGroup) {
	//waitgroupcheck:ignore
	func() { a.Wait(); b.Wait() }() // want "sync.WaitGroup.Wait blocks forever" "sync.WaitGroup.Wait blocks forever"
}

// Type parameters whose constraint allows only *sync.WaitGroup.

type onlyWaitGroup interface{ *sync.WaitGroup }

func throughAnEmbeddedConstraint[W interface {
	onlyWaitGroup
	Wait()
}](w W) {
	w.Wait() // want "sync.WaitGroup.Wait blocks forever"
}

func throughAnInlineConstraint[W interface {
	interface{ *sync.WaitGroup }
	Wait()
}](w W) {
	w.Wait() // want "sync.WaitGroup.Wait blocks forever"
}

func throughATildeTerm[W interface {
	~*sync.WaitGroup
	Wait()
}](w W) {
	w.Wait() // want "sync.WaitGroup.Wait blocks forever"
}

type pool[W interface {
	*sync.WaitGroup
	Wait()
}] struct{ w W }

// stop: the receiver's type parameter.
func (p pool[W]) stop() {
	p.w.Wait() // want "sync.WaitGroup.Wait blocks forever"
}

func typeParamMethodValue[W interface {
	*sync.WaitGroup
	Wait()
}](w W) func() {
	return w.Wait // want "sync.WaitGroup.Wait blocks forever"
}

// A package-level wait with no directive.
var pkgWG sync.WaitGroup

var waitLater = pkgWG.Wait // want "sync.WaitGroup.Wait blocks forever"

// Directive placements that do not reach the wait.

func directiveAboveAnIf(wg *sync.WaitGroup, c bool) {
	//waitgroupcheck:ignore
	if c {
		wg.Wait() // want "sync.WaitGroup.Wait blocks forever"
	}
}

func directiveAboveAFor(wg *sync.WaitGroup) {
	//waitgroupcheck:ignore
	for range 1 {
		wg.Wait() // want "sync.WaitGroup.Wait blocks forever"
	}
}

func directiveAboveABlock(wg *sync.WaitGroup) {
	//waitgroupcheck:ignore
	{
		wg.Wait() // want "sync.WaitGroup.Wait blocks forever"
	}
}

func directiveAboveASelect(wg *sync.WaitGroup, ch chan int) {
	//waitgroupcheck:ignore
	select {
	case <-ch:
		wg.Wait() // want "sync.WaitGroup.Wait blocks forever"
	}
}

func commentLineBetween(wg *sync.WaitGroup) {
	//waitgroupcheck:ignore
	// a note in between
	wg.Wait() // want "sync.WaitGroup.Wait blocks forever"
}

func blankLineBetween(wg *sync.WaitGroup) {
	//waitgroupcheck:ignore

	wg.Wait() // want "sync.WaitGroup.Wait blocks forever"
}

// More constraint shapes that allow only *sync.WaitGroup.

type waiter interface{ Wait() }

type level1 interface{ *sync.WaitGroup }

type level2 interface{ level1 }

type level3 interface {
	level2
	Wait()
}

// oneElementRestricts: waiter allows anything with Wait, onlyWaitGroup
// narrows the set to *sync.WaitGroup.
func oneElementRestricts[W interface {
	waiter
	onlyWaitGroup
}](w W) {
	w.Wait() // want "sync.WaitGroup.Wait blocks forever"
}

func throughThreeLevels[W level3](w W) {
	w.Wait() // want "sync.WaitGroup.Wait blocks forever"
}

func throughALevelAndComparable[T level3, U interface {
	level3
	comparable
}](u U) {
	u.Wait() // want "sync.WaitGroup.Wait blocks forever"
}

func throughAnotherPackage[W constraints.OnlyWaitGroup](w W) {
	w.Wait() // want "sync.WaitGroup.Wait blocks forever"
}

// A directive above one spec inside var ( ... ) does not apply.
var (
	//waitgroupcheck:ignore
	waitInBlock = pkgWG.Wait // want "sync.WaitGroup.Wait blocks forever"
)
