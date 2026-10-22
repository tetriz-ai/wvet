// Command fix-the-findings runs three jobs the way wvet asks: each goroutine
// starts through goroutine.Run, which recovers a panic, and the wait is
// bounded by waitgroup.SafeWait.
package main

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/tetriz-ai/wvet/goroutine"
	"github.com/tetriz-ai/wvet/waitgroup"
)

func main() {
	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		results []string
	)
	record := func(s string) {
		mu.Lock()
		defer mu.Unlock()
		results = append(results, s)
	}
	for _, name := range []string{"alpha", "beta", "gamma"} {
		wg.Add(1)
		goroutine.Run(context.Background(), func(context.Context) {
			if name == "beta" {
				panic("beta failed")
			}
			record(name + ": done")
			wg.Done() // not deferred: on a panic, onPanic calls Done after recording
		}, func(p *goroutine.PanicError) {
			record(fmt.Sprintf("%s: recovered %q", name, p.Value))
			wg.Done()
		})
	}
	if err := waitgroup.SafeWait(&wg, 5*time.Second); err != nil {
		fmt.Println(err)
		return
	}
	sort.Strings(results)
	for _, r := range results {
		fmt.Println(r)
	}
}
