#!/bin/sh
# Runs a program that cancels its workers and waits for them under a shutdown
# deadline, then runs wvet on it to show it has no findings.
set -eu
# shellcheck source=scripts/example-setup.sh
. "$(dirname "$0")/../../scripts/example-setup.sh"
cd "$repo"

go run ./examples/stop-workers-at-a-deadline
wvet ./examples/stop-workers-at-a-deadline && echo "wvet: no findings"
