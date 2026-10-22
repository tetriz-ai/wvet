# Reference

The exact rules for the `wvet` command, each check, the directives and each package. Every entry
point is part of the one module, `github.com/tetriz-ai/wvet`. Signatures and doc comments are on
[pkg.go.dev](https://pkg.go.dev/github.com/tetriz-ai/wvet).

## wvet

| Run | Exit status | Output |
|---|---|---|
| `wvet [flags] <packages>` | 0 clean, 3 findings, 1 a load or type error, 2 for `-fix` | `file:line:col: message` on standard error, absolute paths, grouped by check, then by position |
| `go vet -vettool=$(command -v wvet) <packages>` | 0 clean, 1 findings or errors | go vet's own format, which depends on the Go version |
| `wvet help`, `wvet help <check>` | 0 | The checks and the driver's flags, or one check's one-line summary |

- **Packages** are go package patterns.
- **Test files** are checked too, unless `-test=false`.
- **JSON:** with `-json`, findings exit 0. Gate CI on a run without it.
- **A load or type error** exits 1, with or without `-json`. See
  [Type errors and toolchains](limitations.md#type-errors-and-toolchains).
- **Build tags** come from `GOFLAGS` for wvet and from `-tags` for go vet.
- **wvet's own `-tags`** flag does nothing.
- **Generated files** carry the standard `// Code generated ... DO NOT EDIT.` header.
  All three checks skip them.
- **Analyzers:** `swallowcheck.Analyzer`, `goroutinecheck.Analyzer` and
  `waitgroupcheck.Analyzer` plug into your own multichecker.

### Flags

| Flag | Does |
|---|---|
| `-<check>` | Runs only that check, such as `wvet -swallowcheck ./...` |
| `-<check>=false` | Skips that check |
| `-test=false` | Skips test files |
| `-json` | Prints findings as data, and exits 0 with findings |
| `-c N` | Prints N lines of context |
| `-fix` | Refused with exit status 2, wherever it appears |
| `-diff` | Prints the findings and no diff |
| `-V=full` | Prints the build ID go vet uses to cache results, not a release version |

wvet runs on the x/tools analysis driver, so `wvet help` opens with the driver's general text
and lists the driver's flags. The part about wvet starts at "Registered analyzers".

No check suggests a fix. wvet refuses `-fix` because the driver's fix mode prints no finding and
exits 0.

`go version -m $(command -v wvet)` prints the module version.

## swallowcheck rules

swallowcheck reports an error-typed value that a call produced, that is compared to nil at least
once, and whose every other use is a nil comparison. It follows the value through local
variables, closures that capture it, interface conversions, type changes and branches that merge
it. The finding sits on the call that produced the error.

- **Handled** means any other use: returned, passed to a call, stored, sent, panicked with.
- **Replaced:** an error overwritten on its non-nil branch is reported.
    - An example is `if err != nil { err = errX }`.
    - It is reported also when a closure captures `err` or stores into it.
    - It is not reported when that closure is deferred. See
      [swallowcheck limits](limitations.md#swallowcheck-limits).
    - Wrapping the error in the replacement counts as handled.
- **Interface copies:** a nil check on a copy, such as `var a any = err`, checks `err`.
- **Not reported:** an error never compared to nil, which is errcheck's domain.
- **Not reported:** `%w` and `errors.Is` hygiene. Linters such as errorlint and wrapcheck cover
  those.
- **Scope:** every function, including function literals in package-level variable
  initializers.

## goroutinecheck rules

- **Reports** every `go` statement, at the `go` keyword, whatever it runs.
- **Not reported:** goroutine leaks, which the check does not look for.

## waitgroupcheck rules

waitgroupcheck reports every selector that resolves to `(*sync.WaitGroup).Wait`, called or
passed as a value, however the WaitGroup is reached. Each of these counts.

- a local, a field, a parameter or an embedded WaitGroup, under any import name
- the method expression `(*sync.WaitGroup).Wait(&wg)`
- a type parameter whose constraint allows only a WaitGroup
- the same through an embedded interface or a `~*sync.WaitGroup` term

A `Wait` on any other type, such as `sync.Cond` or `errgroup.Group`, is not reported.

## Directives

| Directive | Silences |
|---|---|
| `//<check>:ignore` | That check |
| `//nolint` | Every check |
| `//nolint:<check>,<other>` | The checks listed |

- **Reason:** after the directive comes nothing, or a reason that starts with `//`.
    - An example is `//nolint:swallowcheck // a miss means the default`.
- **Exact:** `//swallowcheck:ignored` and `//nolint:other` silence nothing.
    - Nor does prose that starts with the word, such as `// nolint is not used here`.
- **Placement for swallowcheck:** on the line of the call or of the nil comparison.
- **Placement for goroutinecheck and waitgroupcheck:** alone on the line directly above the
  statement.
- **Anywhere else,** a directive silences nothing.
- **One statement:** under goroutinecheck and waitgroupcheck, a directive silences only that
  statement.
    - It does not reach a statement nested in it or following it on the line.
- **Package-level var:** a directive above a package-level `var` silences that declaration.
    - One above `var (` silences the whole block.
    - Above one spec inside the block, it does not apply.
- **Labels and cases:** a labeled statement or a `case` clause is the statement on its line.
    - A directive above it does not reach the statement inside, which is still reported.

[examples/suppress-a-finding](../examples/suppress-a-finding/) shows each placement that counts
and two that do not.

## goroutine

- `Run(ctx, fn, onPanic)` starts `fn(ctx)` in a new goroutine and returns at once.
- A panic in `fn` is recovered and passed to `onPanic` as a `*PanicError`.
- With a nil `onPanic`, the panic goes to the standard logger.
- `PanicError` holds the panic value and stack.
- When the panic value is an error, `errors.Is` and `errors.As` reach it through `PanicError`.

## waitgroup

| Call | Returns |
|---|---|
| `SafeWait(wg, timeout)` | nil when the counter reaches zero, `ErrTimeout` when the timeout passes first or is not positive |
| `SafeWaitContext(ctx, wg)` | nil when the counter reaches zero, `ctx.Err()` when ctx is done first or already done |

- **At the deadline:** a counter that reaches zero as the deadline passes returns nil.
- **A nil WaitGroup** panics in the caller's goroutine, as `wg.Wait()` would.
- **After a timeout,** one goroutine stays blocked on the WaitGroup until its counter reaches
  zero.

## archtest rules

archtest ships no rules of its own. It is the walking and reporting a rule needs.

| Function | Does |
|---|---|
| `WalkGoFiles`, `WalkGoFilesFullParse` | Call your function for each non-test `.go` file, parsed to its imports or in full |
| `ReportViolations` | Fail on any violation, or log and skip while violations are known |
| `ReportWithFileAllowlist` | Fail on a violation in a file not allowlisted, log the rest |
| `ReportWithCountAllowlist` | Fail when a file has more violations than its count |
| `ModuleRoot`, `RelPath` | Find the module root, make a path relative with forward slashes |
| `Violation` | Holds one violation's file, line and message |
| `SkipDir` | Tells which directories the walkers skip |
| `ImportPathForAlias`, `DefaultImportName` | Resolve the name a file uses for an import |
| `SelectorPkg`, `SelectorTypeName`, `TypeRefersToPackage`, `IsEmptyInterfaceType` | Read type expressions |

- **Skipped directories:** below the root, names starting with `.` or `_`.
    - Also `testdata`, `vendor` and `node_modules`.
    - The root is always walked.
- **No type information:** the walkers parse files without it, so a rule matches names.
- **Unparseable files** fail the test with the file named, and the walk goes on.
- **Output** is sorted by file, then line.
- **Import names:** `ImportPathForAlias` matches the name an import binds.
    - A blank import answers to `_`, and a dot import to `.`.
- **Type expressions** are matched as written, without type information.
    - `IsEmptyInterfaceType` matches `any` and `interface{}`, not an interface that only embeds
      one.

## Versioning and compatibility

wvet follows semantic versioning. A new finding on unchanged code counts as a breaking change.
Before v1.0, a minor release may change a check's rules, a message or the archtest API, and
[CHANGELOG.md](../CHANGELOG.md) says so.

The minimum Go version is the one in `go.mod`, Go 1.26.
