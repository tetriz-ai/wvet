#!/bin/sh
# Runs wvet over a module where every finding carries a directive. Only the two
# directives in the wrong place leave a report.
set -eu
# shellcheck source=scripts/example-setup.sh
. "$(dirname "$0")/../../scripts/example-setup.sh"
cp -R "$repo/examples/suppress-a-finding/testdata/sample/." .

wvet ./... 2>&1 || echo "wvet exited with status $?"
