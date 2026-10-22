Every finding in this module carries a directive, and only the two in the wrong place leave a report. swallowcheck reads a directive on the call's line or on the `if err != nil` line. goroutinecheck and waitgroupcheck read only a comment alone on the line directly above the statement, so a trailing directive, or one with a blank line before the statement, does not count.

```sh
examples/suppress-a-finding/run.sh
```

```text
/home/you/project/sample.go:38:2: go statement: a panic in this goroutine ends the process; start it with goroutine.Run (github.com/tetriz-ai/wvet/goroutine) or github.com/sourcegraph/conc
/home/you/project/sample.go:51:2: sync.WaitGroup.Wait blocks forever if a Done is missed; use waitgroup.SafeWait (github.com/tetriz-ai/wvet/waitgroup) or github.com/sourcegraph/conc
wvet exited with status 3
```
