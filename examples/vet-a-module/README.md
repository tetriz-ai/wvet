The Quick start, run against a small module that has one problem of each kind: an error that is checked and then dropped, a goroutine started with a bare go statement, and a WaitGroup waited on with no bound. wvet prints one line per finding and exits with status 3.

Run it the same way from your own module's root. If CI already runs go vet, `go vet -vettool=$(command -v wvet) ./...` runs the same checks instead, exits with status 1 and prints go vet's format.

```sh
examples/vet-a-module/run.sh
```

```text
/home/you/project/sample.go:13:16: error checked but never propagated, logged, or inspected (swallowed)
/home/you/project/sample.go:25:3: go statement: a panic in this goroutine ends the process; start it with goroutine.Run (github.com/tetriz-ai/wvet/goroutine) or github.com/sourcegraph/conc
/home/you/project/sample.go:30:2: sync.WaitGroup.Wait blocks forever if a Done is missed; use waitgroup.SafeWait (github.com/tetriz-ai/wvet/waitgroup) or github.com/sourcegraph/conc
wvet exited with status 3
```
