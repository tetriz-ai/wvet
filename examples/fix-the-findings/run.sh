#!/bin/sh
# Runs a program that uses the two helpers wvet points at, then runs wvet on
# it to show it has no findings.
set -eu
# shellcheck source=scripts/example-setup.sh
. "$(dirname "$0")/../../scripts/example-setup.sh"
cd "$repo"

go run ./examples/fix-the-findings
wvet ./examples/fix-the-findings && echo "wvet: no findings"
