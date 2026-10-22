#!/bin/sh
# A file behind a build tag is checked only when the tag is set, the same way
# go build sees it. Set the tag through GOFLAGS for wvet, or with go vet -tags.
set -eu
# shellcheck source=scripts/example-setup.sh
. "$(dirname "$0")/../../scripts/example-setup.sh"
cp -R "$repo/examples/check-tagged-files/testdata/sample/." .

echo "Default build:"
wvet ./... 2>&1 && echo "no findings"
echo "With the integration tag:"
GOFLAGS=-tags=integration wvet ./... 2>&1 || echo "wvet exited with status $?"
