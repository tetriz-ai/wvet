// Copyright 2026 Workloom Solutions Private Limited
// SPDX-License-Identifier: MIT

package main_test

import (
	"bytes"
	"errors"
	"flag"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
)

var update = flag.Bool("update", false, "rewrite testdata/help.txt from the binary's output")

// wvet is the binary under test, built once by TestMain.
var wvet string

func TestMain(m *testing.M) {
	flag.Parse()
	dir, err := os.MkdirTemp("", "wvet-test-")
	if err != nil {
		panic(err)
	}
	wvet = filepath.Join(dir, "wvet")
	out, err := exec.Command("go", "build", "-o", wvet, ".").CombinedOutput()
	if err != nil {
		panic("building wvet: " + err.Error() + "\n" + string(out))
	}
	code := m.Run()
	_ = os.RemoveAll(dir)
	os.Exit(code)
}

// run runs name with args in the fixture module and returns its combined
// output and its standard output, with the module's absolute path written as
// "<mod>", and its exit code.
func run(t *testing.T, dir string, env []string, name string, args ...string) (string, string, int) {
	t.Helper()
	mod, err := filepath.Abs(filepath.Join("testdata", dir))
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(name, args...)
	cmd.Dir = mod
	// The fixture module has no dependencies: nothing is fetched and the
	// local toolchain is used.
	cmd.Env = append(os.Environ(), "GOFLAGS=-mod=mod", "GOPROXY=off", "GOTOOLCHAIN=local")
	cmd.Env = append(cmd.Env, env...)
	// Two copier goroutines write the combined buffer, so it is locked.
	var stdout bytes.Buffer
	out := &lockedBuffer{}
	cmd.Stdout = io.MultiWriter(out, &stdout)
	cmd.Stderr = out
	err = cmd.Run()
	code := 0
	var exit *exec.ExitError
	if errors.As(err, &exit) {
		code = exit.ExitCode()
	} else if err != nil {
		t.Fatalf("running %s: %v", name, err)
	}
	resolved, err := filepath.EvalSymlinks(mod)
	if err != nil {
		t.Fatal(err)
	}
	clean := func(s string) string {
		return strings.ReplaceAll(strings.ReplaceAll(s, resolved, "<mod>"), mod, "<mod>")
	}
	return clean(out.String()), clean(stdout.String()), code
}

type lockedBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *lockedBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *lockedBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

const (
	swallowed  = "error checked but never propagated, logged, or inspected (swallowed)"
	goStmt     = "go statement: a panic in this goroutine ends the process"
	waitGroup  = "sync.WaitGroup.Wait blocks forever if a Done is missed"
	typeErrMsg = `cannot use "not an int"`
)

func TestCommand(t *testing.T) {
	tests := []struct {
		name     string
		dir      string // under testdata; "mod" when empty
		env      []string
		cmd      []string // "wvet" is replaced by the built binary
		wantCode int
		want     []string // each must appear in the output
		wantNot  []string // none may appear
		inOrder  bool     // want appears in this order
		stdout   *string  // when set, standard output must equal it
	}{
		{
			name: "clean package", cmd: []string{"wvet", "./clean/..."}, wantCode: 0,
			wantNot: []string{swallowed, goStmt, waitGroup},
		},
		{
			name: "one finding per analyzer", cmd: []string{"wvet", "./bad/..."}, wantCode: 3,
			want: []string{"<mod>/bad/bad.go:16:16: " + swallowed, "<mod>/bad/bad.go:25:2: " + goStmt, "<mod>/bad/bad.go:30:2: " + waitGroup},
		},
		{
			name: "select one analyzer", cmd: []string{"wvet", "-goroutinecheck", "./bad/..."}, wantCode: 3,
			want: []string{goStmt}, wantNot: []string{swallowed, waitGroup},
		},
		{
			name: "disable one analyzer", cmd: []string{"wvet", "-swallowcheck=false", "./bad/..."}, wantCode: 3,
			want: []string{goStmt, waitGroup}, wantNot: []string{swallowed},
		},
		{
			name: "as go vet's tool", cmd: []string{"go", "vet", "-vettool=wvet", "./bad/..."}, wantCode: 1,
			want: []string{"bad/bad.go:16:16: " + swallowed, "bad/bad.go:25:2: " + goStmt, "bad/bad.go:30:2: " + waitGroup},
		},
		{
			name: "tagged file outside the build is not seen", cmd: []string{"wvet", "./tagged/..."}, wantCode: 0,
			wantNot: []string{waitGroup},
		},
		{
			name: "tagged file with GOFLAGS tags", env: []string{"GOFLAGS=-mod=mod -tags=e2e"}, cmd: []string{"wvet", "./tagged/..."}, wantCode: 3,
			want: []string{"<mod>/tagged/drain_e2e.go:12:2: " + waitGroup},
		},
		{
			name: "tagged file with go vet -tags", cmd: []string{"go", "vet", "-tags=e2e", "-vettool=wvet", "./tagged/..."}, wantCode: 1,
			want: []string{"tagged/drain_e2e.go:12:2: " + waitGroup}, wantNot: []string{goStmt, swallowed},
		},
		{
			name: "type error, standalone: reported, swallowcheck skipped by name, the others still run", cmd: []string{"wvet", "./broken/..."}, wantCode: 1,
			want: []string{typeErrMsg, "swallowcheck: failed prerequisites", goStmt},
		},
		{
			name: "type error, go vet: reported, run fails", cmd: []string{"go", "vet", "-vettool=wvet", "./broken/..."}, wantCode: 1,
			want: []string{typeErrMsg}, wantNot: []string{goStmt},
		},
		{
			name: "findings go to standard error", cmd: []string{"wvet", "./bad/..."}, wantCode: 3,
			want: []string{goStmt}, stdout: new(string),
		},
		{
			name: "output is grouped by check, then by position", cmd: []string{"wvet", "./order/..."}, wantCode: 3,
			want:    []string{"order/order.go:11:2: " + goStmt, "order/order.go:13:2: " + goStmt, "order/order_test.go:9:2: " + goStmt, "order/order.go:12:2: " + waitGroup},
			inOrder: true,
		},
		{
			name: "-test=false skips test files", cmd: []string{"wvet", "-test=false", "./order/..."}, wantCode: 3,
			want: []string{"order/order.go:11:2: " + goStmt}, wantNot: []string{"order_test.go"},
		},
		{
			name: "-json prints findings as data and exits 0", cmd: []string{"wvet", "-json", "./order/..."}, wantCode: 0,
			want: []string{`"goroutinecheck"`, `"waitgroupcheck"`, `order.go:11:2`},
		},
		{
			name: "-c prints the offending line", cmd: []string{"wvet", "-c", "0", "-goroutinecheck", "./bad/..."}, wantCode: 3,
			want: []string{"25\t\tgo work()"},
		},
		{
			name: "-fix is rejected", cmd: []string{"wvet", "-fix", "./bad/..."}, wantCode: 2,
			want: []string{"wvet: -fix is not supported"},
		},
		{
			name: "-fix=true is rejected", cmd: []string{"wvet", "-fix=true", "./bad/..."}, wantCode: 2,
			want: []string{"wvet: -fix is not supported"},
		},
		{
			name: "--fix is rejected", cmd: []string{"wvet", "--fix", "./bad/..."}, wantCode: 2,
			want: []string{"wvet: -fix is not supported"},
		},
		{
			name: "-fix=1 is rejected", cmd: []string{"wvet", "-fix=1", "./bad/..."}, wantCode: 2,
			want: []string{"wvet: -fix is not supported"},
		},
		{
			name: "-c 1 then -fix", cmd: []string{"wvet", "-c", "1", "-fix", "./bad/..."}, wantCode: 2,
			want: []string{"wvet: -fix is not supported"},
		},
		{
			name: "-tags x then -fix", cmd: []string{"wvet", "-tags", "x", "-fix", "./bad/..."}, wantCode: 2,
			want: []string{"wvet: -fix is not supported"},
		},
		{
			name: "a pattern before -fix", cmd: []string{"wvet", "./bad/...", "-fix"}, wantCode: 2,
			want: []string{"wvet: -fix is not supported"},
		},
		{
			name: "-fix=T", cmd: []string{"wvet", "-fix=T", "./bad/..."}, wantCode: 2,
			want: []string{"wvet: -fix is not supported"},
		},
		{
			name: "-fix=TRUE", cmd: []string{"wvet", "-fix=TRUE", "./bad/..."}, wantCode: 2,
			want: []string{"wvet: -fix is not supported"},
		},
		{
			name: "-fix with -json", cmd: []string{"wvet", "-fix", "-json", "./bad/..."}, wantCode: 2,
			want: []string{"wvet: -fix is not supported"},
		},
		{
			name: "-fix=0 is off", cmd: []string{"wvet", "-fix=0", "./bad/..."}, wantCode: 3,
			want: []string{goStmt},
		},
		{
			name: "-fix=yes is the driver's error", cmd: []string{"wvet", "-fix=yes", "./bad/..."}, wantCode: 2,
			want: []string{"invalid boolean value"}, wantNot: []string{"wvet: -fix is not supported"},
		},
		{
			name: "-fix after -- is refused", cmd: []string{"wvet", "-goroutinecheck", "./bad/...", "--", "-fix"}, wantCode: 2,
			want: []string{"wvet: -fix is not supported"},
		},
		{
			name: "-tags -- -fix is refused", cmd: []string{"wvet", "-tags", "--", "-fix", "./bad/..."}, wantCode: 2,
			want: []string{"wvet: -fix is not supported"},
		},
		{
			name: "-debug -- -fix is refused", cmd: []string{"wvet", "-debug", "--", "-fix", "./bad/..."}, wantCode: 2,
			want: []string{"wvet: -fix is not supported"},
		},
		{
			name: "-- -fix is refused", cmd: []string{"wvet", "--", "-fix", "./bad/..."}, wantCode: 2,
			want: []string{"wvet: -fix is not supported"},
		},
		{
			name: "-debug f -fix is refused", cmd: []string{"wvet", "-debug", "f", "-fix", "./bad/..."}, wantCode: 2,
			want: []string{"wvet: -fix is not supported"},
		},
		{
			name: "-fix -- ./bad/... is refused", cmd: []string{"wvet", "-fix", "--", "./bad/..."}, wantCode: 2,
			want: []string{"wvet: -fix is not supported"},
		},
		{
			name: "-fix=t is refused", cmd: []string{"wvet", "-fix=t", "./bad/..."}, wantCode: 2,
			want: []string{"wvet: -fix is not supported"},
		},
		{
			name: "-fix=True is refused", cmd: []string{"wvet", "-fix=True", "./bad/..."}, wantCode: 2,
			want: []string{"wvet: -fix is not supported"},
		},
		{
			name: "-fix -diff is refused", cmd: []string{"wvet", "-fix", "-diff", "./bad/..."}, wantCode: 2,
			want: []string{"wvet: -fix is not supported"},
		},
		{
			name: "---fix is refused", cmd: []string{"wvet", "---fix", "./bad/..."}, wantCode: 2,
			want: []string{"wvet: -fix is not supported"},
		},
		{
			name: "-tags -fix is refused", cmd: []string{"wvet", "-tags", "-fix", "./bad/..."}, wantCode: 2,
			want: []string{"wvet: -fix is not supported"},
		},
		{
			name: "-fix x.cfg, the go vet path is refused", cmd: []string{"wvet", "-fix", "x.cfg"}, wantCode: 2,
			want: []string{"wvet: -fix is not supported"},
		},
		{
			name: "-cpuprofile f -fix is refused", cmd: []string{"wvet", "-cpuprofile", "PROFILE", "-fix", "./bad/..."}, wantCode: 2,
			want: []string{"wvet: -fix is not supported"},
		},
		{
			name: "GOFLAGS=-fix does not reach the driver's fix mode", env: []string{"GOFLAGS=-mod=mod -fix"}, cmd: []string{"wvet", "./bad/..."}, wantCode: 3,
			want: []string{goStmt},
		},
		{
			name: "-- before the patterns", cmd: []string{"wvet", "-goroutinecheck", "--", "./bad/..."}, wantCode: 3,
			want: []string{goStmt},
		},
		{
			name: "-json on a broken package exits 1", cmd: []string{"wvet", "-json", "./broken/..."}, wantCode: 1,
			want: []string{typeErrMsg},
		},
		{
			name: "-fix=false is the default", cmd: []string{"wvet", "-fix=false", "./bad/..."}, wantCode: 3,
			want: []string{goStmt},
		},
		{
			name: "-diff alone still reports", cmd: []string{"wvet", "-diff", "./bad/..."}, wantCode: 3,
			want: []string{goStmt, waitGroup, swallowed},
		},
		{
			name: "help for one check", cmd: []string{"wvet", "help", "goroutinecheck"}, wantCode: 0,
			want: []string{"goroutinecheck: reports go statements"},
		},
		{
			name: "module newer than the toolchain: the run fails, not silently", dir: "newer", cmd: []string{"wvet", "./..."}, wantCode: 1,
			want: []string{"requires go >= 1.99"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := make([]string, len(tt.cmd)-1)
			for i, a := range tt.cmd[1:] {
				args[i] = strings.Replace(a, "-vettool=wvet", "-vettool="+wvet, 1)
				if a == "PROFILE" {
					args[i] = filepath.Join(t.TempDir(), "cpu.prof")
				}
			}
			name := tt.cmd[0]
			if name == "wvet" {
				name = wvet
			}
			dir := tt.dir
			if dir == "" {
				dir = "mod"
			}
			out, stdout, code := run(t, dir, tt.env, name, args...)
			if code != tt.wantCode {
				t.Errorf("exit code %d, want %d\n%s", code, tt.wantCode, out)
			}
			for _, w := range tt.want {
				if !strings.Contains(out, w) {
					t.Errorf("output does not contain %q:\n%s", w, out)
				}
			}
			if tt.inOrder {
				last := -1
				for _, w := range tt.want {
					i := strings.Index(out, w)
					if i <= last {
						t.Errorf("%q is out of order:\n%s", w, out)
					}
					last = i
				}
			}
			if tt.stdout != nil && stdout != *tt.stdout {
				t.Errorf("standard output = %q, want %q", stdout, *tt.stdout)
			}
			for _, w := range tt.wantNot {
				if strings.Contains(out, w) {
					t.Errorf("output contains %q:\n%s", w, out)
				}
			}
		})
	}
}

// The help text is pinned, so a change to an analyzer's name or doc, or to
// the flags, shows up in review.
func TestHelp(t *testing.T) {
	out, _, code := run(t, "mod", nil, wvet, "help")
	if code != 0 {
		t.Fatalf("wvet help exited %d:\n%s", code, out)
	}
	out = regexp.MustCompile(`(?m)^\S*wvet\b`).ReplaceAllString(out, "wvet")
	out = strings.ReplaceAll(out, filepath.Base(wvet)+" ", "wvet ")
	golden := filepath.Join("testdata", "help.txt")
	if *update {
		if err := os.WriteFile(golden, []byte(out), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatal(err)
	}
	if out != string(want) {
		t.Errorf("wvet help differs from %s (rerun with -update to accept):\n%s", golden, out)
	}
}
