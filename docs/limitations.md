# Limitations

These limits matter before you gate a build on the checks. Each is a limit of static analysis,
of the x/tools driver or of a stated design choice, pinned by a test. An item labelled declined
is a fix the project chose not to make.

## swallowcheck limits

- **Misses, by design:** an error passed to a helper that drops it.
    - Each function is checked alone, with the closures that capture its variables.
- **Misses, by design:** an error used wrongly, such as wrapped with the wrong message.
    - Any use other than a nil comparison counts as handled.
- **Misses, declined:** an error that a deferred closure replaces on its non-nil branch.
    - An example is `defer func() { if err != nil { err = errX } }()`.
    - The check walks the function to its exit and does not see a store made in a deferred
      closure.
- **Misses, declined:** an error assigned after a capturing closure that logged it has already
  run.
    - The check counts any read inside a capturing closure as a use, whenever the closure runs.
- **Flags too much:** an error dropped on purpose for a fallback, so silence it with a directive.
    - An existence probe is one, such as `_, err := os.Stat(p); return err == nil`.
- **Flags too much:** a cleanup closure, called from inside a deferred closure, that stores into
  the named result.
    - The check walks only the function that creates the closure.

- **Flags too much:** `if ctx.Err() != nil { return }` uses cancellation as control flow without
  logging or returning the error value. The check treats `context.Context.Err` like other
  error-returning calls (`swallowcheck/testdata/src/swallowed/limitations.go`).
- **Flags too much:** expected-error assertions in tests, such as
  `if _, err := f(); err == nil { t.Fatal("expected error") }`. A nil comparison alone does
  not count as handling the error (`swallowcheck/testdata/src/swallowed/limitations_test.go`).

## goroutinecheck limits

- **Misses:** goroutines started inside other functions.
    - Examples are `time.AfterFunc`, `sync.WaitGroup.Go` (Go 1.25) and `errgroup.Group.Go`.
- **Flags too much:** a goroutine whose function recovers its own panics.

## waitgroupcheck limits

- **Misses:** a `Wait` through an interface whose value is a WaitGroup.
- **Flags too much:** a wait that is bounded some other way.
- **Flags too much:** a WaitGroup populated only with `wg.Go`. Go calls Done when the function
  returns, so there is no missing Done in that pattern. A task can still block forever, and
  the check does not track how a WaitGroup was populated
  (`waitgroupcheck/testdata/src/flagged/limitations.go`).

## Type errors and toolchains

A type error, or code that needs a newer Go than the one that built wvet, fails the run with
status 1. It does not pass silently.

- **Misses:** None known.
- **Flags too much:** None known.

Run on its own, wvet prints the type error and three lines saying swallowcheck was skipped. The
last of them is `swallowcheck: failed prerequisites`. wvet then exits with status 1, and the
other two checks still report. Under go vet, only the type error is printed.

## archtest limits

- **Misses, by design:** a swapped violation under the count ratchet.
    - It compares counts only, so fixing one violation and adding another in one file passes.
    - See [Land a rule before its violations are fixed](usage.md#land-a-rule-before-its-violations-are-fixed).
- **Flags too much:** None known.
