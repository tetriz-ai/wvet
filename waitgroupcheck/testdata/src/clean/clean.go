// Copyright 2026 Workloom Solutions Private Limited
// SPDX-License-Identifier: MIT

// Package clean holds code the analyzer must not report.
package clean

import (
	"sync"
	"time"
)

func noWaitGroup() {}

// Wait is a standalone function, not a method.
func Wait() { time.Sleep(time.Millisecond) }

func standaloneWait() { Wait() }

type dummy struct{}

func (d *dummy) Wait() { time.Sleep(time.Millisecond) }

func otherReceiver() {
	x := &dummy{}
	x.Wait()
}

// SafeWait stands in for a helper that bounds the wait.
func SafeWait(wg *sync.WaitGroup, d time.Duration) error { return nil }

func throughAHelper() {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
	}()
	_ = SafeWait(&wg, time.Second)
}

func multiNameDecl() {
	var wg, cond = &sync.WaitGroup{}, sync.NewCond(&sync.Mutex{})
	wg.Add(1)
	cond.Wait()
}

func usesWaitGroup() {
	var wg sync.WaitGroup
	wg.Add(1)
	wg.Done()
}

func sameNameOtherType() {
	wg := sync.NewCond(&sync.Mutex{})
	wg.Wait()
}

type waiter interface{ Wait() }

// throughAnInterface is a known limit: the static type is the interface.
func throughAnInterface() {
	var w waiter = &sync.WaitGroup{}
	w.Wait()
}

func ignored() {
	var wg sync.WaitGroup
	//waitgroupcheck:ignore
	wg.Wait()
}

func ignoredWithReason() {
	var wg sync.WaitGroup
	//waitgroupcheck:ignore // bounded by the caller's deadline
	wg.Wait()
}

func reasonOnTheLineAbove() {
	var wg sync.WaitGroup
	// Every goroutine here returns within the request.
	//waitgroupcheck:ignore
	wg.Wait()
}

func bareNolint() {
	var wg sync.WaitGroup
	//nolint
	wg.Wait()
}

func nolintList() {
	var wg sync.WaitGroup
	//nolint:gosec,waitgroupcheck // reason
	wg.Wait()
}

// WaitGroup is a user type with the same name; its Wait is not sync's.
type WaitGroup struct{}

func (WaitGroup) Wait() {}

func userDefinedWaitGroup() {
	var wg WaitGroup
	wg.Wait()
}

// Type parameters whose constraint allows a type that is not a WaitGroup.

type otherWaiter struct{}

func (*otherWaiter) Wait() {}

func unionWithAnotherType[W interface {
	*sync.WaitGroup | *otherWaiter
	Wait()
}](w W) {
	w.Wait()
}

func waitOnlyConstraint[W interface{ Wait() }](w W) {
	w.Wait()
}

func anotherPointerType[W interface {
	*otherWaiter
	Wait()
}](w W) {
	w.Wait()
}

// A package-level wait under a directive.
var pkgWG sync.WaitGroup

//waitgroupcheck:ignore
var waitLater = pkgWG.Wait

// multiLineStatement: the directive names the statement, which holds both waits.
func multiLineStatement(a, b *sync.WaitGroup) {
	//waitgroupcheck:ignore
	runBoth(a.Wait,
		b.Wait)
}

func runBoth(f, g func()) { f(); g() }

// A directive above var ( covers the whole declaration.
//
//waitgroupcheck:ignore
var (
	waitA = pkgWG.Wait
	waitB = pkgWG.Wait
)
