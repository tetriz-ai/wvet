// Copyright 2026 Workloom Solutions Private Limited
// SPDX-License-Identifier: MIT

// Command wvet runs the swallowcheck, goroutinecheck and waitgroupcheck
// analyzers over Go packages.
//
// Run it on its own, or as go vet's analysis tool:
//
//	wvet ./...
//	go vet -vettool=$(command -v wvet) ./...
//
// Run on its own, it exits with status 0 when the packages are clean, 3 on a
// finding, 1 on a load or type error, and 2 for -fix, which it refuses because
// no check offers a fix.
package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"golang.org/x/tools/go/analysis/multichecker"

	"github.com/tetriz-ai/wvet/goroutinecheck"
	"github.com/tetriz-ai/wvet/swallowcheck"
	"github.com/tetriz-ai/wvet/waitgroupcheck"
)

func main() {
	if fixRequested(os.Args[1:]) {
		fmt.Fprintln(os.Stderr, "wvet: -fix is not supported: wvet's checks offer no fixes, and the driver's fix mode would hide every finding")
		os.Exit(2)
	}
	multichecker.Main(
		swallowcheck.Analyzer,
		goroutinecheck.Analyzer,
		waitgroupcheck.Analyzer,
	)
}

// fixRequested reports whether any argument turns on -fix: -fix, --fix, or
// -fix=<value> where the value parses as true. It looks at every argument, on
// purpose without knowing which flags take a value: the flag package reads
// "--" after a value flag as that flag's value, so stopping there would let
// -fix through. The cost is the harmless direction, refusing a run such as
// -tags -fix whose -fix is a flag's value. A value that does not parse is left
// to the driver, which reports it.
func fixRequested(args []string) bool {
	for _, arg := range args {
		name, value, hasValue := strings.Cut(strings.TrimLeft(arg, "-"), "=")
		if !strings.HasPrefix(arg, "-") || name != "fix" {
			continue
		}
		if !hasValue {
			return true
		}
		on, err := strconv.ParseBool(value) //swallowcheck:ignore // the driver rejects a bad value
		if err == nil && on {
			return true
		}
	}
	return false
}
