#!/bin/sh
# Copyright 2026 Workloom Solutions Private Limited
# SPDX-License-Identifier: MIT
#
# Writes the license texts of everything compiled into the wvet binary to
# third_party/licenses: the modules for every release target (Linux, macOS
# and Windows) and the Go distribution's own LICENSE. The release archives
# ship the folder.
#
#   .github/scripts/third-party-licenses.sh          regenerate the folder
#   .github/scripts/third-party-licenses.sh --check  fail if it is out of date
#
# Needs: go, and go-licenses v2 (go install github.com/google/go-licenses/v2@v2.0.1),
# found on PATH or in $(go env GOPATH)/bin, where go install puts it.
set -eu
root=$(git rev-parse --show-toplevel)
cd "$root"
golicenses=$(command -v go-licenses || echo "$(go env GOPATH)/bin/go-licenses")
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
mkdir -p "$tmp/all/go"
for os in linux darwin windows; do
  GOOS=$os GOROOT=$(go env GOROOT) "$golicenses" save ./cmd/wvet \
    --ignore github.com/tetriz-ai/wvet --save_path="$tmp/$os" 2>/dev/null
  # go-licenses writes read-only files; make the merge target writable first.
  chmod -R u+w "$tmp/all"
  cp -R "$tmp/$os/." "$tmp/all/"
done
cp "$(go env GOROOT)/LICENSE" "$tmp/all/go/LICENSE"
chmod -R u+w,go-w "$tmp/all"
find "$tmp/all" -type f -exec chmod 0644 {} +
if [ "${1:-}" = "--check" ]; then
  if diff -r "$tmp/all" third_party/licenses; then
    echo "third_party/licenses is up to date"
  else
    echo "third_party/licenses is out of date: run .github/scripts/third-party-licenses.sh" >&2
    exit 1
  fi
else
  rm -rf third_party/licenses
  mkdir -p third_party
  cp -R "$tmp/all" third_party/licenses
  find third_party/licenses -type f | sort
fi
