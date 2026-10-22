# shellcheck shell=sh
# Copyright 2026 Workloom Solutions Private Limited
# SPDX-License-Identifier: MIT
#
# Sourced by every examples/*/run.sh. It builds wvet from this clone, puts it
# first on PATH, and moves into an empty project directory inside a temporary
# directory that is removed on exit.
# Nothing an example runs reaches the network:
#
#   - GOPROXY=off, and the sample modules have no dependencies.
#   - GOTOOLCHAIN is the toolchain that built wvet, so go switches to no other.
#   - GOFLAGS is cleared, and GOWORK=off, so a caller's settings do not leak in.
#   - HOME is an empty temporary directory. Go's build and module caches are
#     the caller's, found before HOME moves, so nothing is rebuilt from scratch.

repo=$(cd "$(dirname "$0")/../.." && pwd)
work=$(mktemp -d "${TMPDIR:-/tmp}/wvet-example.XXXXXX")
trap 'rm -rf "$work"' EXIT
trap 'exit 1' HUP INT TERM

GOCACHE=$(go env GOCACHE)
toolchain=$(cd "$repo" && go env GOVERSION)
GOMODCACHE=$(go env GOMODCACHE)
mkdir "$work/bin" "$work/home" "$work/project"
(cd "$repo" && go build -o "$work/bin/wvet" ./cmd/wvet)

PATH=$work/bin:$PATH
HOME=$work/home
GOFLAGS=
GOWORK=off
GOPROXY=off
GOTOOLCHAIN=$toolchain
export PATH HOME GOCACHE GOMODCACHE GOFLAGS GOWORK GOPROXY GOTOOLCHAIN
cd "$work/project" || exit 1
