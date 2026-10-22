// Copyright 2026 Workloom Solutions Private Limited
// SPDX-License-Identifier: MIT

// Package swallowed holds cases the analyzer MUST flag.
package swallowed

import (
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
)

var (
	errUnavailable = errors.New("backend unavailable")
	ErrNope        = errors.New("nope")
)

func ext() (string, error)    { return "", errUnavailable }
func ext2() (string, error)   { return "", errUnavailable }
func getT[T any]() (T, error) { var z T; return z, errUnavailable }

func returnNew() (string, error) {
	v, err := ext() // want "swallowed"
	if err != nil {
		return "", errors.New("expired")
	}
	return v, nil
}

func returnErrorfNoWrap() (string, error) {
	v, err := ext() // want "swallowed"
	if err != nil {
		return "", fmt.Errorf("expired")
	}
	return v, nil
}

func blankAssignDecoy() (string, error) {
	v, err := ext() // want "swallowed"
	if err != nil {
		_ = err
		return "", errors.New("expired")
	}
	return v, nil
}

func differentVarName() (string, error) {
	v, e := ext() // want "swallowed"
	if e != nil {
		return "", errors.New("expired")
	}
	return v, nil
}

func inlineInit() error {
	if _, err := ext(); err != nil { // want "swallowed"
		return errors.New("expired")
	}
	return nil
}

func returnSentinel() error {
	_, err := ext() // want "swallowed"
	if err != nil {
		return ErrNope
	}
	return nil
}

func returnNilAfterCheck() error {
	_, err := ext() // want "swallowed"
	if err != nil {
		return nil
	}
	return nil
}

func overwriteThenSwallow() error {
	_, err := ext() // first result overwritten unchecked — errcheck's domain, not flagged here
	_, err = ext2() // want "swallowed"
	if err != nil {
		return errors.New("x")
	}
	return nil
}

func generic[T any]() (T, error) {
	v, err := getT[T]() // want "swallowed"
	if err != nil {
		var z T
		return z, errors.New("expired")
	}
	return v, nil
}

type concreteErr struct{}

func (*concreteErr) Error() string { return "concrete" }

func concrete() *concreteErr { return &concreteErr{} }

// mergedThroughPhi: the error from either branch reaches the nil check only
// through the phi that merges the two.
func mergedThroughPhi(c bool) error {
	var err error
	if c {
		_, err = ext() // want "swallowed"
	} else {
		_, err = ext2() // want "swallowed"
	}
	if err != nil {
		return errors.New("expired")
	}
	return nil
}

// boxedIntoInterface: a concrete error pointer is stored in an error variable,
// then checked and dropped.
func boxedIntoInterface() error {
	var err error = concrete() // want "swallowed"
	if err != nil {
		return errors.New("expired")
	}
	return nil
}

func inClosure() func() error {
	return func() error {
		_, err := ext() // want "swallowed"
		if err != nil {
			return errors.New("expired")
		}
		return nil
	}
}

// Directives that do not suppress.

func ignoredIsNotIgnore() error {
	_, err := ext() //swallowcheck:ignored // want "swallowed"
	if err != nil {
		return errors.New("expired")
	}
	return nil
}

func otherLinterOnly() error {
	_, err := ext() //nolint:gosec // want "swallowed"
	if err != nil {
		return errors.New("expired")
	}
	return nil
}

func otherAnalyzerIgnore() error {
	_, err := ext() //goroutinecheck:ignore // want "swallowed"
	if err != nil {
		return errors.New("expired")
	}
	return nil
}

func directiveOnLineAbove() error {
	//swallowcheck:ignore
	_, err := ext() // want "swallowed"
	if err != nil {
		return errors.New("expired")
	}
	return nil
}

// Captured and address-taken variables.

// namedResultInDefer: the deferred closure only compares err to nil, and the
// returned value is a new error.
func namedResultInDefer() (err error) {
	defer func() {
		if err != nil {
			log.Print("rollback")
		}
	}()
	_, err = ext() // want "swallowed"
	if err != nil {
		return errors.New("failed")
	}
	return nil
}

// capturedAndDiscarded: the closure reads err and discards it.
func capturedAndDiscarded() error {
	_, err := ext() // want "swallowed"
	if err != nil {
		return errors.New("failed")
	}
	defer func() { _ = err }()
	return nil
}

// addressTaken: a pointer to err is taken but never used to read it.
func addressTaken() error {
	_, err := ext() // want "swallowed"
	p := &err
	if err != nil {
		return errors.New("failed")
	}
	_ = p
	return nil
}

// A function literal in a package-level variable initializer.
var initialized = func() int {
	_, err := ext() // want "swallowed"
	if err != nil {
		return 0
	}
	return 1
}()

// Comparison forms and flows.

func nilOnTheLeft() error {
	_, err := ext() // want "swallowed"
	if nil != err {
		return ErrNope
	}
	return nil
}

func equalsNil() error {
	_, err := ext() // want "swallowed"
	if err == nil {
		return nil
	}
	return ErrNope
}

type namedError error

// throughChangeType: converted to a named interface type, then checked.
func throughChangeType() error {
	_, err := ext() // want "swallowed"
	e := namedError(err)
	if e != nil {
		return ErrNope
	}
	return nil
}

// throughChangeInterface: converted to any, then checked.
func throughChangeInterface() error {
	_, err := ext() // want "swallowed"
	var a any = err
	if a != nil {
		return ErrNope
	}
	return nil
}

// loopCarried: the error is carried round the loop by a phi.
func loopCarried(n int) error {
	var err error
	for i := 0; i < n; i++ {
		if err != nil {
			break
		}
		if i%2 == 0 {
			_, err = ext() // want "swallowed"
		}
	}
	return nil
}

// existenceProbe is reported: the error is only compared to nil. Silence an
// intended probe with a directive.
func existenceProbe(p string) bool {
	_, err := os.Stat(p) // want "swallowed"
	return err == nil
}

// Replaced on the non-nil branch: the merged value is only ever the
// replacement or a nil error.

func replaced() error {
	_, err := ext() // want "swallowed"
	if err != nil {
		err = ErrNope
	}
	return err
}

func replacedNamed() (err error) {
	_, err = ext() // want "swallowed"
	if err != nil {
		err = ErrNope
	}
	return
}

// replacedBehindADefer: err is captured, so it lives in a variable. The
// non-nil branch stores over it, and the nil branch carries only nil.
func replacedBehindADefer() (err error) {
	defer func() { _ = err == nil }()
	_, err = ext() // want "swallowed"
	if err != nil {
		err = ErrNope
	}
	return err
}

// Stored by a closure into a variable it captured.

func storedInAClosure() error {
	var err error
	func() {
		_, err = ext() // want "swallowed"
	}()
	if err != nil {
		return ErrNope
	}
	return nil
}

func storedInOnceDo(once *sync.Once) error {
	var err error
	once.Do(func() {
		_, err = ext() // want "swallowed"
	})
	if err != nil {
		return ErrNope
	}
	return nil
}

func storedInAGoroutine(done chan struct{}) error {
	var err error
	go func() {
		_, err = ext() // want "swallowed"
		close(done)
	}()
	<-done
	if err != nil {
		return ErrNope
	}
	return nil
}

// Replaced while captured.

// namedResultReplacedInDefer: a deferred closure reads err, and the non-nil
// branch replaces it before the return.
func namedResultReplacedInDefer() (err error) {
	defer func() {
		if err != nil {
			log.Print("rollback")
		}
	}()
	_, err = ext() // want "swallowed"
	if err != nil {
		err = ErrNope
	}
	return err
}

// storedInOnceDoReplacedInParent: the closure stores, the parent replaces.
func storedInOnceDoReplacedInParent(once *sync.Once) error {
	var err error
	once.Do(func() {
		_, err = ext() // want "swallowed"
	})
	if err != nil {
		err = ErrNope
	}
	return err
}

// storedAndReplacedInAClosure: the closure stores and replaces, the parent returns.
func storedAndReplacedInAClosure() error {
	var err error
	func() {
		_, err = ext() // want "swallowed"
		if err != nil {
			err = ErrNope
		}
	}()
	return err
}

// storedInANestedClosure: the store is two closures deep.
func storedInANestedClosure() error {
	var err error
	func() {
		func() {
			_, err = ext() // want "swallowed"
		}()
	}()
	if err != nil {
		return ErrNope
	}
	return nil
}

// More replacement shapes.

func replacedInElseIf(c bool) error {
	_, err := ext() // want "swallowed"
	if c {
		return nil
	} else if err != nil {
		err = ErrNope
	}
	return err
}

func replacedInverted() error {
	_, err := ext() // want "swallowed"
	if err == nil {
		log.Print("ok")
	} else {
		err = ErrNope
	}
	return err
}

// replacedAfterABoxedCheck: a copy converted to any is nil exactly when err is.
func replacedAfterABoxedCheck() error {
	_, err := ext() // want "swallowed"
	var a any = err
	if a != nil {
		err = ErrNope
	}
	return err
}

// storedInAClosureCalledAfterAStore: the closure runs where it is called, after
// the parent's own store, so its error reaches the second check.
func storedInAClosureCalledAfterAStore() (int, error) {
	var c int
	var err error
	load := func() {
		c, err = cfgValue() // want "swallowed"
	}
	_, err = ext()
	if err != nil {
		return 0, err
	}
	load()
	if err != nil {
		return 0, nil
	}
	return c, nil
}

func cfgValue() (int, error) { return 0, errUnavailable }

// replacedAfterANamedInterfaceCheck: a copy converted to a named error type
// is nil exactly when err is.
func replacedAfterANamedInterfaceCheck() error {
	_, err := ext() // want "swallowed"
	e := namedError(err)
	if e != nil {
		err = ErrNope
	}
	return err
}

// passedToOnceDoAfterAStore: once.Do runs the closure where it is called.
func passedToOnceDoAfterAStore(once *sync.Once) error {
	var err error
	load := func() {
		_, err = ext() // want "swallowed"
	}
	_, err = ext2()
	if err != nil {
		return err
	}
	once.Do(load)
	if err != nil {
		return ErrNope
	}
	return nil
}

// startedWithGoAfterAStore: the go statement is where the closure runs.
func startedWithGoAfterAStore(done chan struct{}) error {
	var err error
	g := func() {
		_, err = ext() // want "swallowed"
		close(done)
	}
	_, err = ext2()
	if err != nil {
		return err
	}
	go g()
	<-done
	if err != nil {
		return ErrNope
	}
	return nil
}

// storedInALoopClosure: the closure runs on each pass.
func storedInALoopClosure(n int) error {
	var err error
	step := func() {
		_, err = ext() // want "swallowed"
	}
	for range n {
		step()
		if err != nil {
			return ErrNope
		}
	}
	return nil
}

// deferredStoreIntoAnUnnamedLocal: the deferred closure's store is never read.
func deferredStoreIntoAnUnnamedLocal() error {
	var err error
	defer func() {
		_, err = ext() // want "swallowed"
		if err != nil {
			log.Print("cleanup failed")
		}
	}()
	return nil
}

// Flags too much: a cleanup closure called inside a deferred closure stores
// into the named result, which is returned, but the check walks only the
// function that creates the closure.
func cleanupInsideADefer() (err error) {
	cleanup := func() {
		_, err = ext() // want "swallowed"
		if err != nil {
			log.Print("cleanup failed")
		}
	}
	defer func() { cleanup() }()
	return nil
}
