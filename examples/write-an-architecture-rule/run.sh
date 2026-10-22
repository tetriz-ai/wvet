#!/bin/sh
# Runs three architecture rules written with archtest: first on a tree that
# follows them, then on a tree that breaks each one once.
set -eu
# shellcheck source=scripts/example-setup.sh
. "$(dirname "$0")/../../scripts/example-setup.sh"
cd "$repo"

go test -count=1 ./examples/write-an-architecture-rule/
RULE_TREE=bad go test -count=1 ./examples/write-an-architecture-rule/ || echo "go test exited with status $?"
