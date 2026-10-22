A file behind a build tag is checked only in a build that sets the tag, the same way go build sees it. Run wvet once per set of tags your CI builds with, with the tags in `GOFLAGS`. Under go vet, pass `-tags` to go vet.

```sh
examples/check-tagged-files/run.sh
```

```text
Default build:
no findings
With the integration tag:
/home/you/project/jobs_integration.go:7:2: go statement: a panic in this goroutine ends the process; start it with goroutine.Run (github.com/tetriz-ai/wvet/goroutine) or github.com/sourcegraph/conc
wvet exited with status 3
```
