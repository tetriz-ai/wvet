# Install

The recommended route is `go install`, as the [Quick start](../README.md#quick-start) shows.
Release archives, go vet and your own multichecker are the other ways to run the checks.

## What wvet needs

wvet needs Go 1.26 or newer and the module to check. It needs no service or credential, and it
never runs your code.

- **The `go` command** must be on your `PATH` whenever wvet runs, to load your packages.
- **wvet itself** must be on your `PATH` too.
    - `go install` puts it in `$GOBIN`, or in `$(go env GOPATH)/bin` when `GOBIN` is unset.
- **The Go that builds wvet** sets the newest Go version of code it can check.
    - `go install` fetches wvet and its dependencies and builds wvet with your Go.

```sh
go install github.com/tetriz-ai/wvet/cmd/wvet@latest
```

## A release archive

Use a release archive when you would rather download a binary than build one. The
[Releases page](https://github.com/tetriz-ai/wvet/releases) has archives for Linux, macOS and
Windows, on amd64 and arm64.

1. Download the archive for your platform, and `checksums.txt` beside it. That file lists the
   archives' SHA-256 sums.
2. Check the download.

   ```sh
   sha256sum -c --ignore-missing checksums.txt
   ```

   On macOS, use `shasum -a 256 -c --ignore-missing checksums.txt` instead. On Windows, run
   `certutil -hashfile <archive> SHA256` and compare the sum with the archive's line in
   `checksums.txt`.
3. Unpack it, and put `wvet` in a directory on your `PATH`.

### Notes on the archives

- The `go` command must still be on your `PATH`.
- `go version -m wvet` names the Go the binary was built with.
- On code that needs a newer Go, wvet stops with "package requires newer Go version".
- wvet itself is not tested on macOS or Windows.
    - CI runs every test on Linux and none on macOS.
    - On Windows it runs only archtest's tests and the directive parser's.

## go vet -vettool

Use this route when CI already runs go vet. It runs the same checks and takes go vet's flags,
such as `-tags`.

```sh
go vet -vettool=$(command -v wvet) ./...
```

- It exits with status 1 for findings and errors alike.
- Its output format changes between Go versions.

## Your own multichecker

When you already build a go/analysis multichecker, add `swallowcheck.Analyzer`,
`goroutinecheck.Analyzer` and `waitgroupcheck.Analyzer` to it.

## golangci-lint

golangci-lint has no wvet plugin. Run wvet as its own step beside it.

## The helper packages and archtest

`goroutine`, `waitgroup` and `archtest` are packages of the same module. Add it to your module
with `go get`. It raises a `go` line below 1.26.0 in your `go.mod` to 1.26.0.

```sh
go get github.com/tetriz-ai/wvet@latest
```
