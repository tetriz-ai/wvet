# Usage

How to run wvet in CI, and how to use the helpers and archtest. Every example is listed in
[examples/README.md](../examples/README.md).

## Run it in CI

1. Install a fixed version, so a new release cannot change the result.

   ```sh
   go install github.com/tetriz-ai/wvet/cmd/wvet@v0.1.0
   ```

2. Run it without `-json`. `wvet -json` exits 0 with findings, so it cannot fail the step.

   ```sh
   wvet ./...
   ```

A package that does not type-check fails the run. See
[Type errors and toolchains](limitations.md#type-errors-and-toolchains).

On an existing code base, start CI with one check and add the others as you migrate. Each check
misses some cases and flags others, so read [docs/limitations.md](limitations.md) before you
gate a build on it.

```sh
wvet -swallowcheck ./...
```

## Check tagged files

Some files build only with a tag, such as integration tests. Pass the tag in `GOFLAGS`.

```sh
GOFLAGS=-tags=integration wvet ./...
```

- A file outside the build is not checked.
- Run wvet once per set of tags your builds use.
- wvet's own `-tags` flag does nothing, so use `GOFLAGS` or `go vet -tags=integration -vettool=...`.

[examples/check-tagged-files](../examples/check-tagged-files/) runs both builds.

## Start and wait safely

These helpers fix a goroutinecheck or waitgroupcheck finding. Add the module first, as
[docs/install.md](install.md#the-helper-packages-and-archtest) says.

- **Start** the goroutine with `goroutine.Run`. A panic reaches your handler instead of ending
  the process.
- **Wait** with `waitgroup.SafeWait` for a timeout, or `waitgroup.SafeWaitContext` for a context.
- **Call `Done`** at the end of the job and in the panic handler, not in a `defer`.
    - The handler runs after the job's deferred calls.
    - So a deferred `Done` can end the wait before the panic is recorded.
- **After `SafeWait` times out,** one goroutine stays blocked on the WaitGroup.
    - It is released when the counter reaches zero.

The [README](../README.md#fix-a-goroutine-or-wait-finding) has a complete program with one job.
[examples/fix-the-findings](../examples/fix-the-findings/) runs three jobs this way and collects
their results. [examples/stop-workers-at-a-deadline](../examples/stop-workers-at-a-deadline/)
cancels workers and waits for them with `waitgroup.SafeWaitContext` under a shutdown deadline.

github.com/sourcegraph/conc is an alternative. Its `WaitGroup.Go` calls `Done` for you and
re-raises a panic from `Wait`, but it does not bound the wait.

## Write an architecture rule

An archtest rule is an ordinary Go test. Save it as a `_test.go` file in any package of your
module, and `go test` runs it. The [README](../README.md#write-an-architecture-rule) has a
complete rule, and the full API is on [pkg.go.dev](https://pkg.go.dev/github.com/tetriz-ai/wvet).

- The walkers parse files without type information, so a rule matches names, not types.
- go test prefixes each logged line with the test file and line.
- The last argument of `ReportViolations` is `knownViolations`.
    - Set it to true to log and skip while violations are known, instead of failing.

## Land a rule before its violations are fixed

Report with `ReportWithFileAllowlist` or `ReportWithCountAllowlist` instead of
`ReportViolations`.

- **`ReportWithFileAllowlist`** lets listed files keep their violations.
- **`ReportWithCountAllowlist`** is a ratchet, where each listed file keeps at most its count.
    - The count can fall but never rise.
- **The ratchet compares counts only.** Fixing one and adding one in a file passes.

[examples/write-an-architecture-rule](../examples/write-an-architecture-rule/) uses all three
reporters.
