#!/bin/sh
# The README's quick start: run wvet over a module with one problem of each
# kind. wvet exits with status 3 when it reports.
set -eu
# shellcheck source=scripts/example-setup.sh
. "$(dirname "$0")/../../scripts/example-setup.sh"
cp -R "$repo/examples/vet-a-module/testdata/sample/." .

wvet ./... 2>&1 || echo "wvet exited with status $?"
