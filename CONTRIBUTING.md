# Contributing

Thank you for helping. The scope is deliberately small (see the README), so please open an
issue before starting on a new check or a change to what a check reports. That way we can agree
on it before you write the code.

## Run the checks

These are the build, test and lint checks CI runs. The minimum Go version is 1.26.

```sh
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0
go install github.com/google/go-licenses/v2@v2.0.1
go build ./...
go vet ./...
go test -race -count=1 ./...
go install ./cmd/wvet && go vet -vettool="$(go env GOPATH)/bin/wvet" ./...
golangci-lint run   # golangci-lint v2.14.0, the version CI pins, builds with Go 1.26 or newer; configuration in .golangci.yml
scripts/run-examples.sh   # needs shellcheck 0.11.0, the version CI pins
.github/scripts/third-party-licenses.sh --check
```

The `go install ./cmd/wvet` line runs wvet over its own code, which must stay clean. None of the checks needs
network access beyond downloading modules.

## What the tests cover

`go test -race ./...` runs the following suites.

- The analyzer corpora: code each check must report, and code it must not.
- The directive grammar.
- The command's flags, exit codes and help text.
- The helpers, and archtest's walkers, reporters and three example rules.

`scripts/run-examples.sh` runs every example and compares its output.

## Every change needs a test

A pull request without a test is not merged. A check's behaviour is pinned by its corpus under
`testdata/src`: code it must report carries a `// want "..."` comment, and code it must not
report carries none. A change to what a check reports, a directive, a flag or an exit status is
a breaking change (see "Versioning and compatibility" in docs/reference.md), so it also needs a
`CHANGELOG.md` entry.

If you change an analyzer's name or doc, or the flags, regenerate the help golden file and
review the diff:

```sh
go test ./cmd/wvet -run TestHelp -update
```

## Third-party licenses

The release archives ship `third_party/licenses`, the license texts of Go and of every module
compiled into the binary. After a dependency change, regenerate the folder and commit it:

```sh
.github/scripts/third-party-licenses.sh
```

CI fails when the folder does not match what is compiled in. The script needs the
license-report tool, installed with `go install github.com/google/go-licenses/v2@v2.0.1`.

## Commit style

Use [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/): a short subject
such as `fix: follow a merged error to its nil check`, a blank line, then a body that says why.

## Licensing of contributions

No sign-off line and no separately signed agreement are needed. By opening a pull request you
agree that your contribution is released under the project's [LICENSE](LICENSE).
