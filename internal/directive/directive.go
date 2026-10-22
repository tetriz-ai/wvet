// Copyright 2026 Workloom Solutions Private Limited
// SPDX-License-Identifier: MIT

// Package directive reads the suppression comments the wvet analyzers share.
//
// A directive is a line comment whose first word is one of:
//
//	//<analyzer>:ignore
//	//nolint                 (every analyzer)
//	//nolint:<a>,<b>,...     (the analyzers listed)
//
// Whitespace may separate "//" from the word. After the word comes nothing, or
// whitespace and a reason that starts with "//", such as "// reason". Where a directive
// must sit to take effect is up to each analyzer; [Set] answers both rules.
package directive

import (
	"go/ast"
	"go/token"
	"strings"
)

// Suppresses reports whether text, a comment as [ast.Comment.Text] holds it,
// is a directive for the named analyzer.
func Suppresses(text, analyzer string) bool {
	body, ok := strings.CutPrefix(text, "//")
	if !ok {
		return false
	}
	body = strings.TrimSpace(body)
	word, rest := body, ""
	if i := strings.IndexAny(body, " \t"); i >= 0 {
		word, rest = body[:i], strings.TrimSpace(body[i:])
	}
	// After the word comes nothing, or a reason that starts with "//". Prose that
	// merely starts with "nolint" is not a directive.
	if rest != "" && !strings.HasPrefix(rest, "//") {
		return false
	}
	if word == analyzer+":ignore" || word == "nolint" {
		return true
	}
	list, ok := strings.CutPrefix(word, "nolint:")
	if !ok {
		return false
	}
	for _, name := range strings.Split(list, ",") {
		if name == analyzer {
			return true
		}
	}
	return false
}

type lineKey struct {
	file string
	line int
}

// Set holds the directives for one analyzer across a package's files.
type Set struct {
	fset *token.FileSet
	// onLine holds every line that carries a directive.
	onLine map[lineKey]bool
	// ownLine holds the lines whose directive is the first thing on the line.
	ownLine map[lineKey]bool
	// first holds, for each line, the statement or declaration that starts
	// there first: the outermost one, since an enclosing node starts before
	// what it holds.
	first map[lineKey]ast.Node
}

// Collect finds every directive for analyzer in files. readFile returns a
// file's source; it decides whether a directive starts its own line. A file
// it cannot read keeps its directives for [Set.OnLine] only.
func Collect(fset *token.FileSet, files []*ast.File, analyzer string, readFile func(string) ([]byte, error)) *Set {
	s := &Set{fset: fset, onLine: map[lineKey]bool{}, ownLine: map[lineKey]bool{}, first: map[lineKey]ast.Node{}}
	for _, f := range files {
		ast.Inspect(f, func(n ast.Node) bool {
			switch n.(type) {
			case ast.Stmt, ast.Decl:
			default:
				return true
			}
			p := fset.PositionFor(n.Pos(), false)
			key := lineKey{p.Filename, p.Line}
			if prev, ok := s.first[key]; !ok || n.Pos() < prev.Pos() {
				s.first[key] = n
			}
			return true
		})
		var src []byte
		for _, group := range f.Comments {
			for _, c := range group.List {
				if !Suppresses(c.Text, analyzer) {
					continue
				}
				p := fset.PositionFor(c.Slash, false)
				key := lineKey{p.Filename, p.Line}
				s.onLine[key] = true
				if src == nil {
					data, err := readFile(p.Filename) //swallowcheck:ignore // unreadable source: the directive still counts for OnLine
					if err != nil {
						continue
					}
					src = data
				}
				if startsLine(src, p.Offset) {
					s.ownLine[key] = true
				}
			}
		}
	}
	return s
}

// startsLine reports whether only spaces and tabs precede offset on its line.
func startsLine(src []byte, offset int) bool {
	if offset > len(src) {
		return false
	}
	for i := offset - 1; i >= 0 && src[i] != '\n'; i-- {
		if src[i] != ' ' && src[i] != '\t' {
			return false
		}
	}
	return true
}

// OnLine reports whether any of positions sits on a line that carries a
// directive, wherever on that line the directive is.
func (s *Set) OnLine(positions ...token.Pos) bool {
	for _, pos := range positions {
		if !pos.IsValid() {
			continue
		}
		p := s.fset.PositionFor(pos, false)
		if s.onLine[lineKey{p.Filename, p.Line}] {
			return true
		}
	}
	return false
}

// StatementSuppressed reports whether stmt is the statement a directive on
// the line directly above names: the directive is alone on its line, and stmt
// is the first statement that starts on the next line. A statement nested in it,
// or one that follows it on the same line, is not suppressed. A labeled
// statement or a case clause is the statement that starts its line, so a
// directive above one does not reach the statement inside it.
func (s *Set) StatementSuppressed(stmt ast.Stmt) bool {
	return s.Suppressed(stmt)
}

// Suppressed is StatementSuppressed for a statement or a declaration, such
// as a package-level var. A directive above var ( covers the whole block, and
// one above a single spec inside the block names no statement or declaration.
func (s *Set) Suppressed(n ast.Node) bool {
	if n == nil {
		return false
	}
	p := s.fset.PositionFor(n.Pos(), false)
	return s.first[lineKey{p.Filename, p.Line}] == n && s.OwnLineAbove(n.Pos())
}

// OwnLineAbove reports whether the line directly above pos holds a directive
// as a comment on its own line.
func (s *Set) OwnLineAbove(pos token.Pos) bool {
	p := s.fset.PositionFor(pos, false)
	return s.ownLine[lineKey{p.Filename, p.Line - 1}]
}
