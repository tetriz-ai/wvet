// Copyright 2026 Workloom Solutions Private Limited
// SPDX-License-Identifier: MIT

package goroutine_test

import (
	"bytes"
	"context"
	"errors"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/tetriz-ai/wvet/goroutine"
)

var errBoom = errors.New("boom")

func TestRun(t *testing.T) {
	tests := []struct {
		name      string
		fn        func(context.Context)
		wantPanic bool
		wantValue any
		wantIs    error
	}{
		{"returns normally", func(context.Context) {}, false, nil, nil},
		{"panics with a string", func(context.Context) { panic("bad input") }, true, "bad input", nil},
		{"panics with an error", func(context.Context) { panic(errBoom) }, true, errBoom, errBoom},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := make(chan *goroutine.PanicError, 1)
			finished := make(chan struct{})
			goroutine.Run(context.Background(), func(ctx context.Context) {
				defer close(finished)
				tt.fn(ctx)
			}, func(p *goroutine.PanicError) { got <- p })
			select {
			case <-finished:
			case <-time.After(5 * time.Second):
				t.Fatal("fn did not run")
			}
			var p *goroutine.PanicError
			select {
			case p = <-got:
			case <-time.After(100 * time.Millisecond):
			}
			if (p != nil) != tt.wantPanic {
				t.Fatalf("onPanic called = %v, want %v", p != nil, tt.wantPanic)
			}
			if p == nil {
				return
			}
			if p.Value != tt.wantValue {
				t.Errorf("Value = %v, want %v", p.Value, tt.wantValue)
			}
			if len(p.Stack) == 0 || !strings.Contains(p.Error(), "goroutine panicked") {
				t.Errorf("Error() = %q, want the panic value and a stack", p.Error())
			}
			if tt.wantIs != nil && !errors.Is(p, tt.wantIs) {
				t.Errorf("errors.Is(p, %v) = false, want true", tt.wantIs)
			}
			if err := p.Unwrap(); tt.wantIs == nil && err != nil {
				t.Errorf("Unwrap() = %v, want nil for a non-error value", err)
			}
		})
	}
}

func TestRunPassesTheContext(t *testing.T) {
	type key struct{}
	ctx := context.WithValue(context.Background(), key{}, "value")
	got := make(chan any, 1)
	goroutine.Run(ctx, func(ctx context.Context) { got <- ctx.Value(key{}) }, nil)
	if v := <-got; v != "value" {
		t.Errorf("fn saw context value %v, want %q", v, "value")
	}
}

// With no onPanic, the panic goes to the standard logger rather than being dropped.
func TestRunNilHandlerLogs(t *testing.T) {
	var buf bytes.Buffer
	logged := make(chan struct{})
	log.SetOutput(writerFunc(func(b []byte) (int, error) {
		n, err := buf.Write(b)
		close(logged)
		return n, err
	}))
	t.Cleanup(func() { log.SetOutput(logWriter) })
	goroutine.Run(context.Background(), func(context.Context) { panic("unhandled") }, nil)
	select {
	case <-logged:
	case <-time.After(5 * time.Second):
		t.Fatal("the panic was not logged")
	}
	if !strings.Contains(buf.String(), "goroutine panicked: unhandled") {
		t.Errorf("log = %q, want the panic value", buf.String())
	}
}

var logWriter = log.Writer()

type writerFunc func([]byte) (int, error)

func (f writerFunc) Write(b []byte) (int, error) { return f(b) }

// onPanic runs after fn's own deferred calls, so a deferred Done in fn
// signals before the panic is handled.
func TestOnPanicRunsAfterDeferredCalls(t *testing.T) {
	order := make(chan string, 2)
	done := make(chan struct{})
	goroutine.Run(context.Background(), func(context.Context) {
		defer func() { order <- "fn's deferred call" }()
		panic("boom")
	}, func(*goroutine.PanicError) {
		order <- "onPanic"
		close(done)
	})
	<-done
	if first, second := <-order, <-order; first != "fn's deferred call" || second != "onPanic" {
		t.Errorf("order = %q, %q; want fn's deferred call, then onPanic", first, second)
	}
}
