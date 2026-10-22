// Copyright 2026 Workloom Solutions Private Limited
// SPDX-License-Identifier: MIT

package goroutine_test

import (
	"context"
	"fmt"

	"github.com/tetriz-ai/wvet/goroutine"
)

func ExampleRun() {
	recovered := make(chan *goroutine.PanicError)
	goroutine.Run(context.Background(), func(context.Context) {
		panic("unexpected input")
	}, func(p *goroutine.PanicError) { recovered <- p })

	fmt.Println("recovered:", (<-recovered).Value)
	// Output: recovered: unexpected input
}
