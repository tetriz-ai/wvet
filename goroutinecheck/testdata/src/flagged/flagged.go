// Copyright 2026 Workloom Solutions Private Limited
// SPDX-License-Identifier: MIT

// Package flagged holds go statements the analyzer must report.
package flagged

type server struct{}

func (server) serve() {}

func work() {}

func closure() {
	go func() error { // want "go statement"
		return nil
	}()
}

func namedFunction() {
	go work() // want "go statement"
}

func methodValue(s server) {
	go s.serve() // want "go statement"
}

func nested() {
	go func() { // want "go statement"
		go work() // want "go statement"
	}()
}

func recoversOnItsOwn() {
	go func() { // want "go statement"
		defer func() { _ = recover() }()
		work()
	}()
}

// Directives that do not suppress.

//goroutinecheck:ignore
func directiveAboveTheFunction() {
	go work() // want "go statement"
}

func trailingDirective() {
	go work() //goroutinecheck:ignore // want "go statement"
}

func directiveTrailingTheLineAbove() {
	x := 1                //goroutinecheck:ignore
	go func() { _ = x }() // want "go statement"
}

func directiveTwoLinesAbove() {
	//goroutinecheck:ignore

	go work() // want "go statement"
}

func ignoredIsNotIgnore() {
	//goroutinecheck:ignored
	go work() // want "go statement"
}

func otherAnalyzer() {
	//waitgroupcheck:ignore
	go work() // want "go statement"
}

func otherLinterOnly() {
	//nolint:gosec
	go work() // want "go statement"
}

func unknownDirective() {
	//lint:ignore goroutinecheck
	go work() // want "go statement"
}

func proseStartingWithNolint() {
	// nolint is not something we use here
	go work() // want "go statement"
}

// nestedUnderADirective: the directive names the outer go statement only.
func nestedUnderADirective() {
	//goroutinecheck:ignore
	go func() { go work() }() // want "go statement"
}
