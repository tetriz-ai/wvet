// Command stop-workers-at-a-deadline shuts down workers started with
// goroutine.Run: it cancels the context they were given, then waits for them
// with waitgroup.SafeWaitContext under a shutdown deadline. One worker ignores
// its context, so the wait gives up at the deadline instead of hanging.
package main

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/tetriz-ai/wvet/goroutine"
	"github.com/tetriz-ai/wvet/waitgroup"
)

func main() {
	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		stopped = map[string]bool{}
	)
	markStopped := func(name string) {
		mu.Lock()
		defer mu.Unlock()
		stopped[name] = true
	}
	// stuck is never closed: the exporter ignores its context and waits here.
	stuck := make(chan struct{})

	run, stop := context.WithCancel(context.Background())
	workers := map[string]func(ctx context.Context){
		"poller":   func(ctx context.Context) { <-ctx.Done() },
		"indexer":  func(ctx context.Context) { <-ctx.Done() },
		"exporter": func(context.Context) { <-stuck },
	}
	for name, work := range workers {
		wg.Add(1)
		goroutine.Run(run, func(ctx context.Context) {
			work(ctx)
			markStopped(name)
			wg.Done()
		}, func(p *goroutine.PanicError) {
			fmt.Printf("%s: recovered %v\n", name, p.Value)
			wg.Done()
		})
	}

	stop()
	shutdown, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err := waitgroup.SafeWaitContext(shutdown, &wg)
	switch {
	case err == nil:
		fmt.Println("shutdown: every worker stopped")
	case errors.Is(err, context.DeadlineExceeded):
		fmt.Println("shutdown: gave up at the deadline:", err)
	default:
		fmt.Println("shutdown:", err)
	}

	mu.Lock()
	defer mu.Unlock()
	var done, running []string
	for name := range workers {
		if stopped[name] {
			done = append(done, name)
		} else {
			running = append(running, name)
		}
	}
	sort.Strings(done)
	sort.Strings(running)
	fmt.Println("stopped:", strings.Join(done, ", "))
	fmt.Println("still running:", strings.Join(running, ", "))
}
