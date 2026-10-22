Three architecture rules written as Go tests with the archtest package, one for each reporter: a rule that fails on any violation, a rule with a file allowlist, and a ratchet whose count per file may shrink but never grow. They pass on the `fixed` tree, and fail on the `bad` tree, which breaks each rule once and which only this script runs, so a plain `go test ./...` stays green.

```sh
examples/write-an-architecture-rule/run.sh
```

```text
ok  	github.com/tetriz-ai/wvet/examples/write-an-architecture-rule	<time>
--- FAIL: TestNoPrintln (<time>)
    rule_test.go:49: no-println: 1 violation(s) found:
          store/store.go:6 — fmt.Println outside package main
--- FAIL: TestDomainLayering (<time>)
    rule_test.go:67: domain-layering: 1 pre-existing violation(s) in allowlisted files (WARN):
          domain/legacy.go — domain imports net/http
    rule_test.go:67: domain-layering: 1 NEW violation(s) not in allowlist — MUST FIX:
          domain/order.go — domain imports net/http
--- FAIL: TestPanicRatchet (<time>)
    rule_test.go:92: panic-ratchet: new violation(s) — MUST FIX:
        parse/parse.go: 2 violation(s) but only 1 grandfathered — a new one was added. Fix it, or if it is genuinely pre-existing bump the allowlist count to 2:
            parse/parse.go:6 — panic
            parse/parse.go:9 — panic
FAIL
FAIL	github.com/tetriz-ai/wvet/examples/write-an-architecture-rule	<time>
FAIL
go test exited with status 1
```
