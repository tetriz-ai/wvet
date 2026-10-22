// Copyright 2026 Workloom Solutions Private Limited
// SPDX-License-Identifier: MIT

// Package handled holds cases the analyzer must NOT flag: properly handled
// errors, intentional drops suppressed via directive, and semantic cases where
// err IS used, just used wrongly (out of scope).
package handled

import (
	"errors"
	"fmt"
	stdlog "log"
)

var errUnavailable = errors.New("backend unavailable")

func ext() (string, error)  { return "", errUnavailable }
func wrap(e error) error    { return fmt.Errorf("ctx: %w", e) }
func newErr(m string) error { return errors.New(m) }

type logger struct{}

func (logger) Error(args ...interface{}) {}

var log logger

// suppressed (intentional)

func suppressedOnCall() (string, error) {
	v, err := ext() //nolint:swallowcheck
	if err != nil {
		return "", errors.New("expired")
	}
	return v, nil
}

func suppressedOnIf() (string, error) {
	v, err := ext()
	if err != nil { //swallowcheck:ignore
		return "", errors.New("expired")
	}
	return v, nil
}

// handled

func propagate() (string, error) {
	v, err := ext()
	if err != nil {
		return "", err
	}
	return v, nil
}

func wrapInline() (string, error) {
	v, err := ext()
	if err != nil {
		return "", fmt.Errorf("ctx: %w", err)
	}
	return v, nil
}

func wrapHelper() (string, error) {
	v, err := ext()
	if err != nil {
		return "", wrap(err)
	}
	return v, nil
}

func logged() error {
	_, err := ext()
	if err != nil {
		log.Error(err)
		return newErr("expired")
	}
	return nil
}

func inspect() error {
	_, err := ext()
	if errors.Is(err, errUnavailable) {
		return newErr("unavailable")
	}
	return nil
}

func panics() {
	_, err := ext()
	if err != nil {
		panic(err)
	}
}

// semantic, out of scope (err IS used → not flagged)

func wrongMessage() (string, error) {
	v, err := ext()
	if err != nil {
		return "", fmt.Errorf("created successfully: %w", err)
	}
	return v, nil
}

func suppressedWithReason() (string, error) {
	v, err := ext() //swallowcheck:ignore // the fallback value is correct here
	if err != nil {
		return "", errors.New("expired")
	}
	return v, nil
}

func suppressedByBareNolint() (string, error) {
	v, err := ext() //nolint
	if err != nil {
		return "", errors.New("expired")
	}
	return v, nil
}

func suppressedByNolintList() (string, error) {
	v, err := ext()
	if err != nil { //nolint:gosec,swallowcheck // reason
		return "", errors.New("expired")
	}
	return v, nil
}

// mergedAndReturned: the merged error is propagated.
func mergedAndReturned(c bool) error {
	var err error
	if c {
		_, err = ext()
	} else {
		err = newErr("other")
	}
	if err != nil {
		return err
	}
	return nil
}

type concreteErr struct{}

func (*concreteErr) Error() string { return "concrete" }

func concrete() *concreteErr { return &concreteErr{} }

// boxedAndReturned: the boxed error is propagated.
func boxedAndReturned() error {
	var err error = concrete()
	if err != nil {
		return err
	}
	return nil
}

type holder struct{ err error }

// stored: the error is stored in a field.
func stored(h *holder) {
	_, err := ext()
	if err != nil {
		h.err = err
	}
}

// sent: the error is sent on a channel.
func sent(ch chan<- error) {
	_, err := ext()
	if err != nil {
		ch <- err
	}
}

// directiveOnSecondCheck: the directive sits on the second nil comparison.
func directiveOnSecondCheck() error {
	_, err := ext()
	if err == nil {
		return nil
	}
	if err != nil { //swallowcheck:ignore // a fallback error is right here
		return errors.New("expired")
	}
	return nil
}

// namedResultReturned: the deferred closure logs err, and a naked return
// returns it.
func namedResultReturned() (err error) {
	defer func() {
		if err != nil {
			stdlog.Print(err)
		}
	}()
	_, err = ext()
	if err != nil {
		return
	}
	return nil
}

// capturedAndLogged: the closure passes err to a call.
func capturedAndLogged() error {
	_, err := ext()
	if err != nil {
		return errors.New("failed")
	}
	defer func() { stdlog.Print(err) }()
	return nil
}

func fill(p *error) {}

// addressPassedOn: a pointer to err goes to a call, so its fate is unknown.
func addressPassedOn() error {
	_, err := ext()
	if err != nil {
		fill(&err)
		return errors.New("failed")
	}
	return nil
}

func drop(error) {}

// handedToAHelper is not reported: each function is checked alone, so the
// helper that drops the error is not seen.
func handedToAHelper() error {
	_, err := ext()
	if err != nil {
		drop(err)
		return errors.New("failed")
	}
	return nil
}

// wrappedReplacement: the replacement wraps the original.
func wrappedReplacement() error {
	_, err := ext()
	if err != nil {
		err = fmt.Errorf("x: %w", err)
	}
	return err
}

// storedInAClosureAndReturned: the closure's store reaches the return.
func storedInAClosureAndReturned() error {
	var err error
	func() {
		_, err = ext()
	}()
	if err != nil {
		return err
	}
	return nil
}

// gotoPastTheCheck: the goto reaches the return with err unchecked.
func gotoPastTheCheck(c bool) error {
	_, err := ext()
	if c {
		goto out
	}
	if err != nil {
		err = errors.New("replaced")
	}
out:
	return err
}

// replacedWhenNil: the replacement happens on the nil branch, so a non-nil err
// is returned.
func replacedWhenNil() error {
	_, err := ext()
	if err == nil {
		err = errors.New("no result")
	}
	return err
}

// storedInAClosureReturnedInIt: the closure that stores also returns it.
func storedInAClosureReturnedInIt() func() error {
	var err error
	return func() error {
		_, err = ext()
		if err != nil {
			return err
		}
		return nil
	}
}

// gotoIntoTheNilSide: the nil side of the check is also reached by a goto,
// where err may be non-nil, so it proves nothing.
func gotoIntoTheNilSide(c bool) error {
	_, err := ext()
	if c {
		goto rest
	}
	if err != nil {
		err = errors.New("replaced")
		goto out
	}
rest:
	stdlog.Print("rest")
out:
	return err
}

type closer interface {
	Write([]byte) (int, error)
	Close() error
}

// deferredCloseIntoNamedResult: the deferred Close may replace err, and the
// named result is returned.
func deferredCloseIntoNamedResult(f closer) (err error) {
	defer func() {
		if cerr := f.Close(); cerr != nil {
			err = cerr
		}
	}()
	_, err = f.Write(nil)
	return err
}

// storedInADeferredClosure: the deferred closure stores into the named
// result, which is returned after it runs.
func storedInADeferredClosure() (err error) {
	defer func() {
		_, err = ext()
		if err != nil {
			stdlog.Print("x")
		}
	}()
	return nil
}

// Declined (a bug the author chose not to fix): the deferred closure replaces
// err on its non-nil branch, so the original error is dropped, but the
// parent's walk does not see a deferred closure's replacing store at the exit.
func namedResultReplacedByADefer() (err error) {
	defer func() {
		if err != nil {
			err = errors.New("internal")
		}
	}()
	_, err = ext()
	return err
}

// Declined (a bug the author chose not to fix): the closure that logs err runs
// before the store, so the error is dropped, but the check counts every load
// in a capturing closure as a use.
func closureRanBeforeTheStore() error {
	var err error
	f := func() { stdlog.Print(err) }
	f()
	_, err = ext()
	if err != nil {
		return errors.New("failed")
	}
	return nil
}

// storedInOneClosureLoggedInAnother: one closure stores, another logs.
func storedInOneClosureLoggedInAnother() {
	var err error
	set := func() { _, err = ext() }
	logIt := func() { stdlog.Print(err) }
	set()
	if err != nil {
		logIt()
	}
}

// killedOnOnePath: one branch replaces err, the other returns it.
func killedOnOnePath(c bool) (err error) {
	defer func() { _ = err }()
	_, err = ext()
	if c {
		err = errors.New("replaced")
	}
	return err
}

// panicsInAClosure: the closure panics with the error it stored.
func panicsInAClosure() {
	var err error
	func() {
		_, err = ext()
		if err != nil {
			panic(err)
		}
	}()
}

// retried: the error leads to a recursive retry, then is returned.
func retried(n int) error {
	_, err := ext()
	if err != nil && n > 0 {
		return retried(n - 1)
	}
	return err
}

// deferredCloseTwoReturns: the deferred Close may replace the named result,
// which both returns read.
func deferredCloseTwoReturns(f closer, c bool) (err error) {
	defer func() {
		if cerr := f.Close(); cerr != nil {
			err = cerr
		}
	}()
	if c {
		return nil
	}
	_, err = f.Write(nil)
	return err
}

// deferredCloseInALoop: a function literal per pass, each deferred.
func deferredCloseInALoop(fs []closer) (err error) {
	for _, f := range fs {
		defer func() {
			if cerr := f.Close(); cerr != nil {
				err = cerr
			}
		}()
	}
	_, err = ext()
	return err
}

// storedInBothBranches: either closure's store reaches the return.
func storedInBothBranches(c bool) error {
	var err error
	a := func() { _, err = ext() }
	b := func() { _, err = ext() }
	if c {
		a()
	} else {
		b()
	}
	if err != nil {
		return err
	}
	return nil
}

// afterEarlyReturns: the closure runs after two early returns.
func afterEarlyReturns(x, y bool) error {
	var err error
	load := func() { _, err = ext() }
	if x {
		return nil
	}
	if y {
		return nil
	}
	load()
	if err != nil {
		return err
	}
	return nil
}

type group struct{}

func (group) Go(f func() error) { _ = f() }

// passedToAGoMethod: an errgroup-style Go runs the closure, which returns the error.
func passedToAGoMethod(g group) {
	g.Go(func() error {
		_, err := ext()
		if err != nil {
			return err
		}
		return nil
	})
}

// recursiveVisit: a closure that calls itself through a variable.
func recursiveVisit(n int) error {
	var err error
	var visit func(int)
	visit = func(i int) {
		if i == 0 {
			_, err = ext()
			return
		}
		visit(i - 1)
	}
	visit(n)
	if err != nil {
		return err
	}
	return nil
}

// deferredCleanupIntoNamedResult: a deferred local cleanup sets the result.
func deferredCleanupIntoNamedResult() (err error) {
	cleanup := func() error { _, e := ext(); return e }
	defer func() {
		if cerr := cleanup(); cerr != nil {
			err = cerr
		}
	}()
	return nil
}

// joinedInADefer: the deferred closure joins the errors.
func joinedInADefer() (err error) {
	defer func() {
		_, cerr := ext()
		err = errors.Join(err, cerr)
	}()
	_, err = ext()
	return err
}

// calledBeforeAndAfterAStore: the second call's error is returned.
func calledBeforeAndAfterAStore() error {
	var err error
	load := func() { _, err = ext() }
	load()
	_, err = ext()
	load()
	if err != nil {
		return err
	}
	return nil
}

// calledOnlyInsideAnotherClosure: the outer closure returns the error.
func calledOnlyInsideAnotherClosure() func() error {
	var err error
	load := func() { _, err = ext() }
	return func() error {
		load()
		if err != nil {
			return err
		}
		return nil
	}
}
