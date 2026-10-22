// Copyright 2026 Workloom Solutions Private Limited
// SPDX-License-Identifier: MIT

// Package goroutine starts goroutines that recover from a panic instead of
// ending the process.
package goroutine

import (
	"context"
	"fmt"
	"log"
	"runtime/debug"
)

// PanicError is a panic recovered from a goroutine started by [Run].
type PanicError struct {
	// Value is the value passed to panic.
	Value any
	// Stack is the goroutine's stack at the panic.
	Stack []byte
}

// Error returns the panic value and the stack.
func (p *PanicError) Error() string {
	return fmt.Sprintf("goroutine panicked: %v\n%s", p.Value, p.Stack)
}

// Unwrap returns the panic value when it is an error, so errors.Is and
// errors.As see through a PanicError.
func (p *PanicError) Unwrap() error {
	err, _ := p.Value.(error)
	return err
}

// Run starts fn(ctx) in a new goroutine and returns at once. If fn panics, the
// panic is recovered and passed to onPanic in that goroutine. When onPanic is
// nil, the panic is written to the standard logger instead, so it is never
// dropped silently. onPanic runs after fn's own deferred calls, so a Done that
// fn defers can end a wait before onPanic sees the panic.
//
// Run does not wait for fn or cancel it. fn should return when ctx is done.
func Run(ctx context.Context, fn func(ctx context.Context), onPanic func(*PanicError)) {
	//goroutinecheck:ignore // this is the recovering helper the check points at
	go func() {
		defer func() {
			v := recover()
			if v == nil {
				return
			}
			p := &PanicError{Value: v, Stack: debug.Stack()}
			if onPanic == nil {
				log.Print(p)
				return
			}
			onPanic(p)
		}()
		fn(ctx)
	}()
}
