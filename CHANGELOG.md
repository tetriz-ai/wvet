# Changelog

All notable changes are recorded using [Keep a Changelog](https://keepachangelog.com/en/1.1.0/)
and [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2026-10-22

### Added

- Initial release.
- `wvet`, a Go analysis multichecker that runs standalone or through `go vet`.
- Checks for swallowed errors, bare goroutines and unbounded `sync.WaitGroup` waits.
- `goroutine.Run`, `waitgroup.SafeWait` and `waitgroup.SafeWaitContext` helpers.
- The `archtest` package for architecture rules, allowlists and count ratchets.
- Runnable examples, analyzer test corpora and documented limitations.

[Unreleased]: https://github.com/tetriz-ai/wvet/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/tetriz-ai/wvet/releases/tag/v0.1.0
