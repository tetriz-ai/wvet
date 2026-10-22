// Copyright 2026 Workloom Solutions Private Limited
// SPDX-License-Identifier: MIT

package main_test

import (
	"os"
	"regexp"
	"slices"
	"strings"
	"testing"
)

// TestQuickStartMatchesExample keeps the README's Quick start and
// examples/vet-a-module the same program. Every command in every Quick start
// sh block runs in the example's run.sh, and the Quick start's output is the
// example's. A line only one side can have is listed with its reason, and must
// exist on its own side, so a stale entry fails too.
func TestQuickStartMatchesExample(t *testing.T) {
	readme := section(t, read(t, "../../README.md"), "## Quick start")
	var readmeCmds []string
	for _, b := range blocks(readme, "sh") {
		readmeCmds = append(readmeCmds, lines(b)...)
	}
	if len(readmeCmds) == 0 {
		t.Fatal("the Quick start has no sh block")
	}
	// The Quick start shows the paths a clone in /home/you prints. The
	// example shows its temporary project directory as /home/you/project.
	readmeOut := lines(strings.ReplaceAll(block(t, readme, "text", 0),
		"/home/you/wvet/examples/vet-a-module/testdata/sample/", "/home/you/project/"))
	runSh := read(t, "../../examples/vet-a-module/run.sh")
	example := read(t, "../../examples/vet-a-module/README.md")
	exampleOut := lines(block(t, example, "text", 0))

	onlyInReadme := map[string]string{
		"go install github.com/tetriz-ai/wvet/cmd/wvet@latest": "the example builds wvet from the clone instead",
		"git clone https://github.com/tetriz-ai/wvet":          "the example runs inside the clone",
		"cd wvet/examples/vet-a-module/testdata/sample":        "the example copies the sample module into a temporary directory",
	}
	for _, cmd := range readmeCmds {
		if reason, ok := onlyInReadme[cmd]; ok {
			t.Logf("skipped %q: %s", cmd, reason)
			continue
		}
		if !slices.ContainsFunc(lines(runSh), func(l string) bool { return strings.HasPrefix(l, cmd+" ") || l == cmd }) {
			t.Errorf("Quick start command %q is not a command in examples/vet-a-module/run.sh", cmd)
		}
	}
	for cmd := range onlyInReadme {
		if !slices.Contains(readmeCmds, cmd) {
			t.Errorf("stale skip: %q is no longer in the Quick start", cmd)
		}
	}

	onlyInExample := map[string]string{
		"wvet exited with status 3": "the script prints the exit status; the Quick start states it in prose",
	}
	var got []string
	for _, l := range exampleOut {
		if _, ok := onlyInExample[l]; !ok {
			got = append(got, l)
		}
	}
	for l := range onlyInExample {
		if !slices.Contains(exampleOut, l) {
			t.Errorf("stale skip: %q is no longer in the example's output", l)
		}
	}
	if !slices.Equal(readmeOut, got) {
		t.Errorf("Quick start output differs from examples/vet-a-module:\nREADME:\n%s\nexample:\n%s", strings.Join(readmeOut, "\n"), strings.Join(got, "\n"))
	}
}

func read(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// section returns text from the heading to the next heading of the same level.
func section(t *testing.T, text, heading string) string {
	t.Helper()
	i := strings.Index(text, "\n"+heading+"\n")
	if i < 0 {
		t.Fatalf("no %q heading", heading)
	}
	rest := text[i+1+len(heading):]
	if j := strings.Index(rest, "\n## "); j >= 0 {
		rest = rest[:j]
	}
	return rest
}

// block returns the body of the n-th fenced block with the given language.
func block(t *testing.T, text, lang string, n int) string {
	t.Helper()
	all := blocks(text, lang)
	if len(all) <= n {
		t.Fatalf("no %s block number %d", lang, n)
	}
	return all[n]
}

// blocks returns the bodies of every fenced block with the given language. A
// fence may be indented, as it is under a numbered step.
func blocks(text, lang string) []string {
	parts := strings.Split(text, "```"+lang+"\n")
	var out []string
	for _, p := range parts[1:] {
		body, _, _ := strings.Cut(p, "```")
		out = append(out, body)
	}
	return out
}

// lines splits text into its non-empty lines, comment lines dropped.
func lines(text string) []string {
	var out []string
	for _, l := range strings.Split(text, "\n") {
		if l = strings.TrimSpace(l); l != "" && !strings.HasPrefix(l, "#") {
			out = append(out, l)
		}
	}
	return out
}

// TestReadmeLinksEveryExample keeps the README's Quick start and Common tasks
// in step with examples/: every example folder is linked from one of them,
// and every example they link is a folder.
func TestReadmeLinksEveryExample(t *testing.T) {
	readme := read(t, "../../README.md")
	tasks := section(t, readme, "## Quick start") + section(t, readme, "## Common tasks")
	linked := map[string]bool{}
	for _, m := range regexp.MustCompile(`\]\(examples/([a-z0-9-]+)/\)`).FindAllStringSubmatch(tasks, -1) {
		linked[m[1]] = true
	}
	entries, err := os.ReadDir("../../examples")
	if err != nil {
		t.Fatal(err)
	}
	present := map[string]bool{}
	for _, e := range entries {
		if e.IsDir() {
			present[e.Name()] = true
			if !linked[e.Name()] {
				t.Errorf("examples/%s/ is not linked from the README's Quick start or Common tasks", e.Name())
			}
		}
	}
	for name := range linked {
		if !present[name] {
			t.Errorf("the README links examples/%s/, which does not exist", name)
		}
	}
}
