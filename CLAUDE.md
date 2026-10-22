## What this is

`wvet`, a go/analysis multichecker: swallowcheck (an error compared to nil, then dropped),
goroutinecheck (every `go` statement), waitgroupcheck (every `sync.WaitGroup` Wait), the helpers
they point at, and `archtest` for rules written as Go tests. It finds no leaks or unchecked errors.

## Layout

- `cmd/wvet/`: `main.go`. `main_test.go` runs the binary on `testdata/`. `readme_test.go` ties
  the README to `examples/`. `docs/`: reference, limitations, install and usage.
- `swallowcheck/`, `goroutinecheck/`, `waitgroupcheck/`: one analyzer each, tested by
  `analysistest` on the corpora in `testdata/src/<case>/`. `internal/directive/`: suppression.
- `goroutine/`, `waitgroup/`: the helpers. `archtest/`: rule helpers, toy rules in `rules_test.go`.
- `examples/`: one folder per task and the index `examples/README.md`, run by
  `scripts/run-examples.sh` with `scripts/example-setup.sh`. CI is `.github/workflows/ci.yml`.

## Commands

The minimum Go is 1.26.0. Run the floor and the newest CI Go with `GOTOOLCHAIN=go1.26.0` and
`GOTOOLCHAIN=go1.27.3` in front of a command. If your go env file (`go env GOENV`) sets GOPRIVATE,
GOFLAGS or GOPROXY, run everything with `GOENV=off`. Install the pinned tools once, with
`$(go env GOPATH)/bin` (or your `GOBIN`) on `PATH`. golangci-lint v2.14.0 builds with Go 1.26 or
newer. Its cache is `golangci-lint` in your user cache dir, moved by `GOLANGCI_LINT_CACHE`.

```sh
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0
go install github.com/google/go-licenses/v2@v2.0.1
go build ./...
go vet ./...
go test -race -count=1 ./...
go install ./cmd/wvet && go vet -vettool="$(go env GOPATH)/bin/wvet" ./...
golangci-lint run   # golangci-lint v2.14.0, the version CI pins; configuration in .golangci.yml
scripts/run-examples.sh   # needs shellcheck 0.11.0, the version CI pins
.github/scripts/third-party-licenses.sh --check
```

With a custom `GOBIN`, self-vet with `go vet -vettool="$(command -v wvet)" ./...`. CI-only jobs:
govulncheck, archtest on Windows, a Go 1.26.0 wvet on Go 1.27 code and goreleaser.

## Contract to keep

- **What each check reports, and does not,** is its corpus: a `// want` line must be reported,
  every other line must not.
- **Directives are exact:** `//<check>:ignore`, `//nolint`, `//nolint:<list>`, then nothing or a
  `//` reason (`TestSuppresses` in `internal/directive/directive_test.go`). An own-line
  directive silences only the first statement on the next line (`TestSetPlacement`, corpora).
- **Failure is loud:** exit 3 on a finding, 1 on a load or type error (`TestCommand` in
  `cmd/wvet/main_test.go`). The repo vets clean with its own wvet.
- **The walkers never pass on code they did not read** (`archtest/archtest_test.go`).
- **`docs/limitations.md` holds only limits of static analysis, the x/tools driver and stated
  design choices,** each pinned by a test. A bug is fixed, not listed there.

## Making a change

- A behaviour change needs a test that fails before it and passes after (corpora for an
  analyzer, table tests elsewhere) and a `CHANGELOG.md` line under Unreleased. A new example
  needs no changelog line.
- An analyzer name, doc or flag change: `go test ./cmd/wvet -run TestHelp -update`, then review.
  A dependency change reruns `.github/scripts/third-party-licenses.sh` and updates NOTICE.
- Fix lint findings in the code. `.golangci.yml` only exempts G204 and G304 in tests.
- **A new example** is `examples/<task-in-words>/` with an executable `run.sh` and a `README.md`
  (one paragraph, an `sh` block, a `text` block last). Copy one with its `# shellcheck source=`
  line. Link it from `examples/README.md` and from the root README's Quick start or a Common
  tasks H3: an example for a task already there joins that H3, and one for a new task gets a new
  H3 or a link from `docs/usage.md`. `TestReadmeLinksEveryExample` fails when a folder is missing.
  Example code must pass wvet itself, and the self-vet covers `examples/`.
- The setup builds wvet onto `PATH`, sets `$repo`, and moves into an empty project directory,
  with `GOPROXY=off`. The runner runs the `sh` block from the repo root and compares standard
  output with the last block, writing that directory as `/home/you/project` and durations as
  `<time>`. wvet writes findings to standard error, so add `2>&1`. A clean run prints nothing,
  so echo a status line, as `fix-the-findings` does. Use standalone `wvet`: go vet's format varies.
- An example with a failing variant keeps the default `go test ./...` green and runs the
  failing variant only from its `run.sh`, as `write-an-architecture-rule` does with `RULE_TREE`.

## Do not

- Do not raise the `go` line in `go.mod`: golang.org/x/tools v0.51.0 sets the 1.26.0 floor.
- Do not add a license header under `examples/` or to a `// Code generated` fixture.
- Do not edit `cmd/wvet/testdata/help.txt` by hand.
